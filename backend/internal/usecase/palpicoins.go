package usecase

import (
	"context"
	"strings"

	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
	"github.com/gabrielevieira/palpitai/backend/internal/social"
)

var (
	ErrInsufficientBalance = apperrors.NewConflict("insufficient palpicoin balance")
	ErrChallengeForbidden  = apperrors.NewForbidden("challenge participant required")
	ErrChallengeNotPending = apperrors.NewConflict("challenge is not pending")
	ErrChallengeNotFriend  = apperrors.NewForbidden("challenge requires accepted friendship")
	ErrChallengeMatchState = apperrors.NewConflict("challenge requires scheduled match")
)

type WalletUsecase struct {
	db Datastore
}

func CanDebitWalletBalance(balance int, amount int) bool {
	return amount > 0 && balance >= amount
}

func NewWalletUsecase(db Datastore) WalletUsecase {
	return WalletUsecase{db: db}
}

func (uc WalletUsecase) Credit(ctx context.Context, userID string, amount int, transactionType domain.PalpicoinTransactionType, description string, referenceType *string, referenceID *string) (bool, error) {
	return repositories.CreditWallet(ctx, uc.db, userID, amount, transactionType, description, referenceType, referenceID)
}

func (uc WalletUsecase) Debit(ctx context.Context, userID string, amount int, transactionType domain.PalpicoinTransactionType, description string, referenceType *string, referenceID *string) (bool, error) {
	return repositories.DebitWallet(ctx, uc.db, userID, amount, transactionType, description, referenceType, referenceID)
}

func (uc WalletUsecase) Refund(ctx context.Context, userID string, amount int, description string, referenceType *string, referenceID *string) (bool, error) {
	return uc.Credit(ctx, userID, amount, domain.PalpicoinTransactionChallengeRefund, description, referenceType, referenceID)
}

func (uc WalletUsecase) GetBalance(ctx context.Context, userID string) (dto.WalletResponse, error) {
	return repositories.WalletByUserID(ctx, uc.db, userID)
}

func (uc WalletUsecase) ListTransactions(ctx context.Context, userID string, limit int, offset int) (dto.PalpicoinTransactionPageResponse, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	items, err := repositories.ListPalpicoinTransactions(ctx, uc.db, userID, limit, offset)
	return dto.PalpicoinTransactionPageResponse{Items: items, Limit: limit, Offset: offset, Notice: repositories.PalpicoinNotice}, err
}

func (uc WalletUsecase) Ranking(ctx context.Context, userID string) (dto.PalpicoinRankingResponse, error) {
	ranking, err := repositories.PalpicoinRanking(ctx, uc.db, userID)
	return dto.PalpicoinRankingResponse{Ranking: ranking, Notice: repositories.PalpicoinNotice}, err
}

type ChallengeUsecase struct {
	db     Datastore
	events social.EventPublisher
}

func NewChallengeUsecase(db Datastore) ChallengeUsecase {
	return ChallengeUsecase{db: db, events: social.LogEventPublisher{}}
}

func NewChallengeUsecaseWithEvents(db Datastore, events social.EventPublisher) ChallengeUsecase {
	if events == nil {
		events = social.LogEventPublisher{}
	}
	return ChallengeUsecase{db: db, events: events}
}

func (uc ChallengeUsecase) Create(ctx context.Context, creatorUserID string, request dto.CreateChallengeRequest) (domain.PalpicoinChallenge, error) {
	opponentID := strings.TrimSpace(request.OpponentID)
	matchID := strings.TrimSpace(request.MatchID)
	if creatorUserID == "" || opponentID == "" || matchID == "" || request.StakeAmount <= 0 {
		return domain.PalpicoinChallenge{}, apperrors.NewValidation("Informe amigo, jogo e valor.")
	}
	if creatorUserID == opponentID {
		return domain.PalpicoinChallenge{}, apperrors.NewValidation("Você não pode desafiar a si mesmo.")
	}

	var created domain.PalpicoinChallenge
	err := withTx(ctx, uc.db, func(tx repositories.Querier) error {
		isFriend, err := repositories.AcceptedFriendshipExists(ctx, tx, creatorUserID, opponentID)
		if err != nil {
			return err
		}
		if !isFriend {
			return ErrChallengeNotFriend
		}

		isScheduled, err := repositories.ScheduledMatchExists(ctx, tx, matchID)
		if err != nil {
			return err
		}
		if !isScheduled {
			return ErrChallengeMatchState
		}

		challenge, err := repositories.CreateChallenge(ctx, tx, creatorUserID, opponentID, matchID, request.StakeAmount)
		if err != nil {
			return err
		}
		refType := "challenge"
		createdRef := challenge.ID
		moved, err := repositories.DebitWallet(ctx, tx, creatorUserID, request.StakeAmount, domain.PalpicoinTransactionChallengeStake, "Aposta em desafio", &refType, &createdRef)
		if err != nil {
			return err
		}
		if !moved {
			return ErrInsufficientBalance
		}
		created = challenge
		return nil
	})
	if err != nil {
		return domain.PalpicoinChallenge{}, err
	}
	_ = uc.events.Publish(ctx, domain.SocialEvent{ActorUserID: creatorUserID, TargetID: created.ID, TargetType: "palpicoin_challenge", Type: domain.SocialEventChallengeCreated})
	return created, nil
}

func (uc ChallengeUsecase) Accept(ctx context.Context, userID string, challengeID string) (domain.PalpicoinChallenge, error) {
	var accepted domain.PalpicoinChallenge
	err := withTx(ctx, uc.db, func(tx repositories.Querier) error {
		challenge, err := repositories.GetChallenge(ctx, tx, strings.TrimSpace(challengeID))
		if err != nil {
			return err
		}
		if challenge.OpponentUserID != userID {
			return ErrChallengeForbidden
		}
		if challenge.Status != domain.ChallengeStatusPending {
			return ErrChallengeNotPending
		}
		refType := "challenge"
		refID := challenge.ID
		moved, err := repositories.DebitWallet(ctx, tx, userID, challenge.StakeAmount, domain.PalpicoinTransactionChallengeStake, "Entrada em desafio", &refType, &refID)
		if err != nil {
			return err
		}
		if !moved {
			return ErrInsufficientBalance
		}
		accepted, err = repositories.UpdateChallengeStatus(ctx, tx, challenge.ID, domain.ChallengeStatusAccepted)
		return err
	})
	if err != nil {
		return domain.PalpicoinChallenge{}, err
	}
	_ = uc.events.Publish(ctx, domain.SocialEvent{ActorUserID: userID, TargetID: accepted.ID, TargetType: "palpicoin_challenge", Type: domain.SocialEventChallengeAccepted})
	return accepted, nil
}

func (uc ChallengeUsecase) Decline(ctx context.Context, userID string, challengeID string) (domain.PalpicoinChallenge, error) {
	return uc.closePending(ctx, userID, challengeID, domain.ChallengeStatusDeclined, false)
}

func (uc ChallengeUsecase) Cancel(ctx context.Context, userID string, challengeID string) (domain.PalpicoinChallenge, error) {
	return uc.closePending(ctx, userID, challengeID, domain.ChallengeStatusCancelled, true)
}

func (uc ChallengeUsecase) closePending(ctx context.Context, userID string, challengeID string, status domain.ChallengeStatus, creatorOnly bool) (domain.PalpicoinChallenge, error) {
	var updated domain.PalpicoinChallenge
	err := withTx(ctx, uc.db, func(tx repositories.Querier) error {
		challenge, err := repositories.GetChallenge(ctx, tx, strings.TrimSpace(challengeID))
		if err != nil {
			return err
		}
		if creatorOnly && challenge.CreatorUserID != userID {
			return ErrChallengeForbidden
		}
		if !creatorOnly && challenge.OpponentUserID != userID {
			return ErrChallengeForbidden
		}
		if challenge.Status != domain.ChallengeStatusPending {
			return ErrChallengeNotPending
		}
		refType := "challenge_refund"
		refID := challenge.ID
		if _, err := repositories.CreditWallet(ctx, tx, challenge.CreatorUserID, challenge.StakeAmount, domain.PalpicoinTransactionChallengeRefund, "Estorno de desafio", &refType, &refID); err != nil {
			return err
		}
		updated, err = repositories.UpdateChallengeStatus(ctx, tx, challenge.ID, status)
		return err
	})
	return updated, err
}

func (uc ChallengeUsecase) List(ctx context.Context, userID string, status string, challengeType string) (dto.ChallengeListResponse, error) {
	if challengeType == "" {
		challengeType = "all"
	}
	if challengeType != "sent" && challengeType != "received" && challengeType != "all" {
		return dto.ChallengeListResponse{}, apperrors.NewValidation("Filtro de tipo inválido.")
	}
	challenges, err := repositories.ListChallenges(ctx, uc.db, userID, strings.TrimSpace(status), challengeType)
	return dto.ChallengeListResponse{Challenges: challenges, Notice: repositories.PalpicoinNotice}, err
}

func (uc ChallengeUsecase) Get(ctx context.Context, userID string, challengeID string) (dto.ChallengeResponse, error) {
	return repositories.ChallengeDetail(ctx, uc.db, userID, strings.TrimSpace(challengeID))
}

func SettlePalpicoinChallengesForMatch(ctx context.Context, db repositories.Querier, matchID string) error {
	challenges, err := repositories.SettleAcceptedChallengesForMatch(ctx, db, matchID)
	if err != nil {
		return err
	}
	for _, challenge := range challenges {
		refType := "challenge_settlement"
		refID := challenge.ID
		for _, payout := range ChallengePayouts(challenge) {
			description := "Vitória em desafio"
			if payout.Type == domain.PalpicoinTransactionChallengeRefund {
				description = "Empate no desafio"
			}
			if _, err := repositories.CreditWallet(ctx, db, payout.UserID, payout.Amount, payout.Type, description, &refType, &refID); err != nil {
				return err
			}
		}
	}
	return nil
}

type ChallengePayout struct {
	Amount int
	Type   domain.PalpicoinTransactionType
	UserID string
}

func ChallengePayouts(challenge domain.PalpicoinChallenge) []ChallengePayout {
	if challenge.WinnerUserID == nil {
		return []ChallengePayout{
			{Amount: challenge.StakeAmount, Type: domain.PalpicoinTransactionChallengeRefund, UserID: challenge.CreatorUserID},
			{Amount: challenge.StakeAmount, Type: domain.PalpicoinTransactionChallengeRefund, UserID: challenge.OpponentUserID},
		}
	}
	return []ChallengePayout{
		{Amount: challenge.StakeAmount * 2, Type: domain.PalpicoinTransactionChallengeWin, UserID: *challenge.WinnerUserID},
	}
}

func RewardPalpicoinsForMatchPredictions(ctx context.Context, db repositories.Querier, matchID string, homeScore int, awayScore int) error {
	rows, err := db.Query(ctx, `
		select distinct on (p.user_id)
			p.user_id::text,
			p.home_score,
			p.away_score
		from predictions p
		where p.match_id = $1
		order by p.user_id, p.updated_at desc
	`, matchID)
	if err != nil {
		return err
	}
	defer rows.Close()

	refType := "match"
	refID := matchID
	for rows.Next() {
		var userID string
		var predictedHome int
		var predictedAway int
		if err := rows.Scan(&userID, &predictedHome, &predictedAway); err != nil {
			return err
		}
		for _, reward := range PalpicoinRewardsForPrediction(predictedHome, predictedAway, homeScore, awayScore) {
			if _, err := repositories.CreditWallet(ctx, db, userID, reward.Amount, reward.Type, reward.Description, &refType, &refID); err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

type PalpicoinReward struct {
	Amount      int
	Description string
	Type        domain.PalpicoinTransactionType
}

func PalpicoinRewardsForPrediction(predictedHome int, predictedAway int, finalHome int, finalAway int) []PalpicoinReward {
	rewards := []PalpicoinReward{}
	if predictedHome == finalHome && predictedAway == finalAway {
		rewards = append(rewards, PalpicoinReward{Amount: 50, Description: "Acerto de placar exato", Type: domain.PalpicoinTransactionExactScoreHit})
	}
	if sign(predictedHome-predictedAway) == sign(finalHome-finalAway) {
		amount := 10
		description := "Acerto de vencedor"
		if finalHome == finalAway {
			amount = 15
			description = "Acerto de empate"
		}
		rewards = append(rewards, PalpicoinReward{Amount: amount, Description: description, Type: domain.PalpicoinTransactionMatchWinnerHit})
	}
	return rewards
}

func sign(value int) int {
	switch {
	case value > 0:
		return 1
	case value < 0:
		return -1
	default:
		return 0
	}
}
