package usecase

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
)

var (
	ErrMatchAlreadyStarted = apperrors.NewConflict("match already started")
	ErrMembershipRequired  = apperrors.NewForbidden("active membership required")
	ErrPaymentRequired     = apperrors.NewForbidden("confirmed payment required")
)

type PredictionUsecase struct {
	db Datastore
}

func NewPredictionUsecase(db Datastore) PredictionUsecase {
	return PredictionUsecase{db: db}
}

func (uc PredictionUsecase) ListGroupMatches(ctx context.Context, userID string, groupID string) ([]dto.MatchResponse, error) {
	return ListGroupMatches(ctx, uc.db, userID, groupID)
}

func (uc PredictionUsecase) UserTotalScore(ctx context.Context, userID string) (int, error) {
	return UserTotalScore(ctx, uc.db, userID)
}

func (uc PredictionUsecase) GroupRanking(ctx context.Context, userID string, groupID string) ([]dto.RankingEntryResponse, error) {
	return GroupRanking(ctx, uc.db, userID, groupID)
}

func (uc PredictionUsecase) SavePrediction(ctx context.Context, userID string, groupID string, matchID string, request dto.PredictionRequest) (dto.PredictionResponse, error) {
	return SavePrediction(ctx, uc.db, userID, groupID, matchID, request)
}

func (uc PredictionUsecase) SaveMatchResult(ctx context.Context, matchID string, request dto.MatchResultRequest) (int, error) {
	return SaveMatchResult(ctx, uc.db, matchID, request)
}

func (uc PredictionUsecase) MatchDetailsByID(ctx context.Context, matchID string) (domain.MatchDetails, error) {
	return MatchDetailsByID(ctx, uc.db, matchID)
}

func (uc PredictionUsecase) GroupsAffectedByMatch(ctx context.Context, matchID string) ([]domain.GroupSummary, error) {
	return GroupsAffectedByMatch(ctx, uc.db, matchID)
}

func ListGroupMatches(ctx context.Context, db Datastore, userID string, groupID string) ([]dto.MatchResponse, error) {
	if err := EnsureActiveGroupMember(ctx, db, userID, groupID); err != nil {
		return nil, err
	}

	return repositories.ListGroupMatches(ctx, db, groupID, userID)
}

func UserTotalScore(ctx context.Context, db Datastore, userID string) (int, error) {
	return repositories.UserTotalScore(ctx, db, userID)
}

func GroupRanking(ctx context.Context, db Datastore, userID string, groupID string) ([]dto.RankingEntryResponse, error) {
	if err := EnsureActiveGroupMember(ctx, db, userID, groupID); err != nil {
		return nil, err
	}

	return repositories.GroupRanking(ctx, db, groupID)
}

func MatchDetailsByID(ctx context.Context, db Datastore, matchID string) (domain.MatchDetails, error) {
	return repositories.MatchDetailsByID(ctx, db, matchID)
}

func GroupsAffectedByMatch(ctx context.Context, db Datastore, matchID string) ([]domain.GroupSummary, error) {
	return repositories.GroupsAffectedByMatch(ctx, db, matchID)
}

func SavePrediction(ctx context.Context, db Datastore, userID string, groupID string, matchID string, request dto.PredictionRequest) (dto.PredictionResponse, error) {
	if err := EnsureActiveGroupMember(ctx, db, userID, groupID); err != nil {
		return dto.PredictionResponse{}, err
	}

	canPredict, err := repositories.CanUserPredictInGroup(ctx, db, userID, groupID)
	if errors.Is(err, repositories.ErrNotFound) {
		return dto.PredictionResponse{}, ErrMembershipRequired
	}
	if err != nil {
		return dto.PredictionResponse{}, err
	}
	if !canPredict {
		return dto.PredictionResponse{}, ErrPaymentRequired
	}

	kickoffAt, err := repositories.MatchKickoffForGroup(ctx, db, groupID, matchID)
	if errors.Is(err, repositories.ErrNotFound) {
		return dto.PredictionResponse{}, ErrMembershipRequired
	}
	if err != nil {
		return dto.PredictionResponse{}, err
	}

	if !time.Now().UTC().Before(kickoffAt.UTC()) {
		return dto.PredictionResponse{}, ErrMatchAlreadyStarted
	}

	return repositories.UpsertPrediction(ctx, db, userID, groupID, matchID, request)
}

func SaveMatchResult(ctx context.Context, db Datastore, matchID string, request dto.MatchResultRequest) (int, error) {
	scoredPredictions := 0
	err := withTx(ctx, db, func(tx repositories.Querier) error {
		beforeByGroup, err := RankingSnapshotsByAffectedGroup(ctx, tx, matchID)
		if err != nil {
			return err
		}
		if err := repositories.UpdateMatchResult(ctx, tx, matchID, request); err != nil {
			return err
		}

		scoredPredictions, err = repositories.ScoreMatchPredictions(ctx, tx, matchID, request)
		if err != nil {
			return err
		}

		if err := RewardPalpicoinsForMatchPredictions(ctx, tx, matchID, request.HomeScore, request.AwayScore); err != nil {
			return err
		}
		if err := SettlePalpicoinChallengesForMatch(ctx, tx, matchID); err != nil {
			return err
		}

		return PublishMatchScoringFeedEvents(ctx, tx, matchID, request.HomeScore, request.AwayScore, beforeByGroup)
	})
	if err != nil {
		return 0, err
	}

	return scoredPredictions, nil
}

func FormatResultMessage(homeTeam string, awayTeam string, homeScore int, awayScore int) string {
	if homeTeam == "" || awayTeam == "" {
		return "Resultado final lancado"
	}

	return homeTeam + " " + strconv.Itoa(homeScore) + "x" + strconv.Itoa(awayScore) + " " + awayTeam + " - resultado final lancado"
}

func EnsureActiveGroupMember(ctx context.Context, db Datastore, userID string, groupID string) error {
	exists, err := repositories.ActiveGroupMemberExists(ctx, db, userID, groupID)
	if err != nil {
		return err
	}

	if !exists {
		return ErrMembershipRequired
	}

	return nil
}
