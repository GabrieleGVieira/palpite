package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
)

func TestCreateFriendRequestRejectsSelf(t *testing.T) {
	uc := NewFriendsUsecaseWithStore(&fakeFriendshipStore{}, fakeSocialPublisher{})

	_, err := uc.CreateRequest(context.Background(), "user-1", "user-1")

	if !apperrors.IsValidation(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCreateFriendRequestRequiresExistingAddressee(t *testing.T) {
	uc := NewFriendsUsecaseWithStore(&fakeFriendshipStore{}, fakeSocialPublisher{})

	_, err := uc.CreateRequest(context.Background(), "user-1", "user-2")

	if !apperrors.IsNotFound(err) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestCreateFriendRequestRejectsDuplicatePending(t *testing.T) {
	store := &fakeFriendshipStore{
		existingUsers: map[string]bool{"user-2": true},
		friendship: domain.Friendship{
			ID:              "friendship-1",
			RequesterUserID: "user-1",
			AddresseeUserID: "user-2",
			Status:          domain.FriendshipStatusPending,
		},
	}
	uc := NewFriendsUsecaseWithStore(store, fakeSocialPublisher{})

	_, err := uc.CreateRequest(context.Background(), "user-1", "user-2")

	if !apperrors.IsConflict(err) {
		t.Fatalf("expected conflict error, got %v", err)
	}
	if store.created {
		t.Fatal("expected no friendship creation")
	}
}

func TestAcceptRequiresAddressee(t *testing.T) {
	store := &fakeFriendshipStore{
		friendship: domain.Friendship{
			ID:              "friendship-1",
			RequesterUserID: "user-1",
			AddresseeUserID: "user-2",
			Status:          domain.FriendshipStatusPending,
		},
	}
	uc := NewFriendsUsecaseWithStore(store, fakeSocialPublisher{})

	_, err := uc.Accept(context.Background(), "user-1", "friendship-1")

	if !apperrors.IsForbidden(err) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestAcceptPendingRequest(t *testing.T) {
	store := &fakeFriendshipStore{
		friendship: domain.Friendship{
			ID:              "friendship-1",
			RequesterUserID: "user-1",
			AddresseeUserID: "user-2",
			Status:          domain.FriendshipStatusPending,
		},
	}
	events := &recordingSocialPublisher{}
	uc := NewFriendsUsecaseWithStore(store, events)

	friendship, err := uc.Accept(context.Background(), "user-2", "friendship-1")

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if friendship.Status != domain.FriendshipStatusAccepted {
		t.Fatalf("expected accepted status, got %s", friendship.Status)
	}
	if len(events.events) != 1 || events.events[0].Type != domain.SocialEventFriendRequestAccepted {
		t.Fatalf("expected accepted social event, got %#v", events.events)
	}
}

func TestDeleteRequiresParticipant(t *testing.T) {
	store := &fakeFriendshipStore{
		friendship: domain.Friendship{
			ID:              "friendship-1",
			RequesterUserID: "user-1",
			AddresseeUserID: "user-2",
			Status:          domain.FriendshipStatusAccepted,
		},
	}
	uc := NewFriendsUsecaseWithStore(store, fakeSocialPublisher{})

	err := uc.Delete(context.Background(), "user-3", "friendship-1")

	if !apperrors.IsForbidden(err) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

type fakeFriendshipStore struct {
	created       bool
	existingUsers map[string]bool
	friendship    domain.Friendship
}

func (store *fakeFriendshipStore) Accept(_ context.Context, friendshipID string) (domain.Friendship, error) {
	if store.friendship.ID != friendshipID {
		return domain.Friendship{}, repositories.ErrNotFound
	}
	store.friendship.Status = domain.FriendshipStatusAccepted
	return store.friendship, nil
}

func (store *fakeFriendshipStore) CreateRequest(_ context.Context, requesterUserID string, addresseeUserID string) (domain.Friendship, error) {
	store.created = true
	store.friendship = domain.Friendship{
		ID:              "friendship-created",
		RequesterUserID: requesterUserID,
		AddresseeUserID: addresseeUserID,
		Status:          domain.FriendshipStatusPending,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	return store.friendship, nil
}

func (store *fakeFriendshipStore) Decline(_ context.Context, friendshipID string) (domain.Friendship, error) {
	if store.friendship.ID != friendshipID {
		return domain.Friendship{}, repositories.ErrNotFound
	}
	store.friendship.Status = domain.FriendshipStatusDeclined
	return store.friendship, nil
}

func (store *fakeFriendshipStore) Delete(_ context.Context, friendshipID string) error {
	if store.friendship.ID != friendshipID {
		return repositories.ErrNotFound
	}
	store.friendship = domain.Friendship{}
	return nil
}

func (store *fakeFriendshipStore) GetByID(_ context.Context, friendshipID string) (domain.Friendship, error) {
	if store.friendship.ID != friendshipID {
		return domain.Friendship{}, repositories.ErrNotFound
	}
	return store.friendship, nil
}

func (store *fakeFriendshipStore) GetFriendship(_ context.Context, _ string, _ string) (domain.Friendship, error) {
	if store.friendship.ID == "" {
		return domain.Friendship{}, repositories.ErrNotFound
	}
	return store.friendship, nil
}

func (store *fakeFriendshipStore) ListFriends(_ context.Context, _ string) ([]dto.FriendResponse, error) {
	return nil, nil
}

func (store *fakeFriendshipStore) ListPendingRequests(_ context.Context, _ string) ([]dto.PendingFriendRequestResponse, error) {
	return nil, nil
}

func (store *fakeFriendshipStore) PublicProfile(_ context.Context, _ string) (dto.PublicProfileResponse, error) {
	return dto.PublicProfileResponse{}, nil
}

func (store *fakeFriendshipStore) SearchUsers(_ context.Context, _ string, _ string, _ int) ([]dto.UserSearchResponse, error) {
	return nil, nil
}

func (store *fakeFriendshipStore) UserExists(_ context.Context, userID string) (bool, error) {
	if store.existingUsers == nil {
		return false, nil
	}
	return store.existingUsers[userID], nil
}

type fakeSocialPublisher struct{}

func (fakeSocialPublisher) Publish(context.Context, domain.SocialEvent) error {
	return nil
}

type recordingSocialPublisher struct {
	events []domain.SocialEvent
}

func (publisher *recordingSocialPublisher) Publish(_ context.Context, event domain.SocialEvent) error {
	publisher.events = append(publisher.events, event)
	return nil
}

var _ FriendshipStore = (*fakeFriendshipStore)(nil)
