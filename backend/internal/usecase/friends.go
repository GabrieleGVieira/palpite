package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
	"github.com/gabrielevieira/palpitai/backend/internal/social"
)

var (
	ErrFriendshipNotParticipant = apperrors.NewForbidden("friendship participant required")
	ErrFriendshipRecipientOnly  = apperrors.NewForbidden("friendship recipient required")
	ErrFriendshipNotPending     = apperrors.NewConflict("friendship is not pending")
)

type FriendshipStore interface {
	Accept(ctx context.Context, friendshipID string) (domain.Friendship, error)
	CreateRequest(ctx context.Context, requesterUserID string, addresseeUserID string) (domain.Friendship, error)
	Decline(ctx context.Context, friendshipID string) (domain.Friendship, error)
	Delete(ctx context.Context, friendshipID string) error
	GetByID(ctx context.Context, friendshipID string) (domain.Friendship, error)
	GetFriendship(ctx context.Context, userID string, otherUserID string) (domain.Friendship, error)
	ListFriends(ctx context.Context, userID string) ([]dto.FriendResponse, error)
	ListPendingRequests(ctx context.Context, userID string) ([]dto.PendingFriendRequestResponse, error)
	PublicProfile(ctx context.Context, requesterUserID string, userID string) (dto.PublicProfileResponse, error)
	SearchUsers(ctx context.Context, requesterUserID string, query string, limit int) ([]dto.UserSearchResponse, error)
	UserExists(ctx context.Context, userID string) (bool, error)
}

type FriendsUsecase struct {
	db     Datastore
	events social.EventPublisher
	store  FriendshipStore
}

func NewFriendsUsecase(db Datastore) FriendsUsecase {
	return FriendsUsecase{
		db:     db,
		events: social.LogEventPublisher{},
		store:  repositories.NewFriendshipRepository(db),
	}
}

func NewFriendsUsecaseWithStore(store FriendshipStore, events social.EventPublisher) FriendsUsecase {
	if events == nil {
		events = social.LogEventPublisher{}
	}
	return FriendsUsecase{events: events, store: store}
}

func (uc FriendsUsecase) CreateRequest(ctx context.Context, requesterUserID string, addresseeUserID string) (domain.Friendship, error) {
	requesterUserID = strings.TrimSpace(requesterUserID)
	addresseeUserID = strings.TrimSpace(addresseeUserID)
	if requesterUserID == "" || addresseeUserID == "" {
		return domain.Friendship{}, apperrors.NewValidation("Informe o usuario.")
	}
	if requesterUserID == addresseeUserID {
		return domain.Friendship{}, apperrors.NewValidation("Você não pode adicionar a si mesmo.")
	}

	exists, err := uc.store.UserExists(ctx, addresseeUserID)
	if err != nil {
		return domain.Friendship{}, err
	}
	if !exists {
		return domain.Friendship{}, apperrors.NewNotFound("user not found")
	}

	current, err := uc.store.GetFriendship(ctx, requesterUserID, addresseeUserID)
	if err != nil && !errors.Is(err, repositories.ErrNotFound) {
		return domain.Friendship{}, err
	}
	if err == nil {
		switch current.Status {
		case domain.FriendshipStatusAccepted:
			return domain.Friendship{}, apperrors.NewConflict("friendship already accepted")
		case domain.FriendshipStatusPending:
			return domain.Friendship{}, apperrors.NewConflict("friendship request already pending")
		case domain.FriendshipStatusBlocked:
			return domain.Friendship{}, apperrors.NewConflict("friendship blocked")
		}
	}

	friendship, err := uc.store.CreateRequest(ctx, requesterUserID, addresseeUserID)
	if err != nil {
		if repositories.IsUniqueViolation(err) {
			return domain.Friendship{}, apperrors.NewConflict("friendship already exists")
		}
		return domain.Friendship{}, err
	}

	_ = uc.events.Publish(ctx, domain.SocialEvent{
		ActorUserID: requesterUserID,
		TargetID:    friendship.ID,
		TargetType:  "friendship",
		Type:        domain.SocialEventFriendRequestSent,
	})

	return friendship, nil
}

func (uc FriendsUsecase) Accept(ctx context.Context, userID string, friendshipID string) (domain.Friendship, error) {
	friendship, err := uc.pendingFriendshipForAddressee(ctx, userID, friendshipID)
	if err != nil {
		return domain.Friendship{}, err
	}

	friendship, err = uc.store.Accept(ctx, friendship.ID)
	if err != nil {
		return domain.Friendship{}, err
	}

	_ = uc.events.Publish(ctx, domain.SocialEvent{
		ActorUserID: userID,
		TargetID:    friendship.ID,
		TargetType:  "friendship",
		Type:        domain.SocialEventFriendRequestAccepted,
	})

	return friendship, nil
}

func (uc FriendsUsecase) Decline(ctx context.Context, userID string, friendshipID string) (domain.Friendship, error) {
	friendship, err := uc.pendingFriendshipForAddressee(ctx, userID, friendshipID)
	if err != nil {
		return domain.Friendship{}, err
	}
	return uc.store.Decline(ctx, friendship.ID)
}

func (uc FriendsUsecase) Delete(ctx context.Context, userID string, friendshipID string) error {
	friendship, err := uc.store.GetByID(ctx, strings.TrimSpace(friendshipID))
	if err != nil {
		return err
	}
	if friendship.RequesterUserID != userID && friendship.AddresseeUserID != userID {
		return ErrFriendshipNotParticipant
	}
	if uc.db != nil {
		return withTx(ctx, uc.db, func(tx repositories.Querier) error {
			if err := refundChallengesForRemovedFriendship(ctx, tx, friendship); err != nil {
				return err
			}
			if err := repositories.DeleteChallengesBetweenUsers(ctx, tx, friendship.RequesterUserID, friendship.AddresseeUserID); err != nil {
				return err
			}
			repo := repositories.NewFriendshipRepository(tx)
			return repo.Delete(ctx, friendship.ID)
		})
	}
	return uc.store.Delete(ctx, friendship.ID)
}

func (uc FriendsUsecase) ListFriends(ctx context.Context, userID string) ([]dto.FriendResponse, error) {
	return uc.store.ListFriends(ctx, userID)
}

func (uc FriendsUsecase) ListPendingRequests(ctx context.Context, userID string) ([]dto.PendingFriendRequestResponse, error) {
	return uc.store.ListPendingRequests(ctx, userID)
}

func (uc FriendsUsecase) SearchUsers(ctx context.Context, userID string, query string) ([]dto.UserSearchResponse, error) {
	return uc.store.SearchUsers(ctx, userID, strings.TrimSpace(query), 20)
}

func (uc FriendsUsecase) PublicProfile(ctx context.Context, requesterUserID string, userID string) (dto.PublicProfileResponse, error) {
	return uc.store.PublicProfile(ctx, strings.TrimSpace(requesterUserID), strings.TrimSpace(userID))
}

func refundChallengesForRemovedFriendship(ctx context.Context, db repositories.Querier, friendship domain.Friendship) error {
	challenges, err := repositories.ListRefundableChallengesBetweenUsers(ctx, db, friendship.RequesterUserID, friendship.AddresseeUserID)
	if err != nil {
		return err
	}
	refType := "challenge_refund"
	for _, challenge := range challenges {
		refID := challenge.ID
		if _, err := repositories.CreditWallet(ctx, db, challenge.CreatorUserID, challenge.StakeAmount, domain.PalpicoinTransactionChallengeRefund, "Estorno por amizade removida", &refType, &refID); err != nil {
			return err
		}
		if challenge.Status == domain.ChallengeStatusAccepted {
			if _, err := repositories.CreditWallet(ctx, db, challenge.OpponentUserID, challenge.StakeAmount, domain.PalpicoinTransactionChallengeRefund, "Estorno por amizade removida", &refType, &refID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (uc FriendsUsecase) pendingFriendshipForAddressee(ctx context.Context, userID string, friendshipID string) (domain.Friendship, error) {
	friendship, err := uc.store.GetByID(ctx, strings.TrimSpace(friendshipID))
	if err != nil {
		return domain.Friendship{}, err
	}
	if friendship.AddresseeUserID != userID {
		return domain.Friendship{}, ErrFriendshipRecipientOnly
	}
	if friendship.Status != domain.FriendshipStatusPending {
		return domain.Friendship{}, ErrFriendshipNotPending
	}
	return friendship, nil
}
