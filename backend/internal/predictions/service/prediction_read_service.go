package service

import (
	"context"

	explanationmodels "github.com/gabrielevieira/palpitai/backend/internal/explanations/models"
	explanationrepo "github.com/gabrielevieira/palpitai/backend/internal/explanations/repository"
	"github.com/gabrielevieira/palpitai/backend/internal/goalpredictions/models"
	goalrepo "github.com/gabrielevieira/palpitai/backend/internal/goalpredictions/repository"
	predictionmodels "github.com/gabrielevieira/palpitai/backend/internal/predictions/models"
	predictionrepo "github.com/gabrielevieira/palpitai/backend/internal/predictions/repository"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
)

type PredictionReadService struct {
	db           repositories.Querier
	explanations explanationrepo.Repository
}

func NewPredictionReadService(db repositories.Querier) PredictionReadService {
	return PredictionReadService{db: db, explanations: explanationrepo.New(db)}
}

func (s PredictionReadService) MatchPredictionByMatchID(ctx context.Context, matchID string) (predictionmodels.MatchPrediction, error) {
	return predictionrepo.ByMatchID(ctx, s.db, matchID)
}

func (s PredictionReadService) GoalPredictionByMatchID(ctx context.Context, matchID string) (models.MatchGoalPrediction, error) {
	return goalrepo.ByMatchID(ctx, s.db, matchID)
}

func (s PredictionReadService) ExplanationByMatchID(ctx context.Context, matchID string, promptVersion string) (explanationmodels.PredictionExplanation, error) {
	return s.explanations.FindByMatchID(ctx, matchID, promptVersion)
}
