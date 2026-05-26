package service

import (
	"context"

	explanationmodels "github.com/gabrielevieira/palpitai/backend/internal/explanations/models"
	explanationrepo "github.com/gabrielevieira/palpitai/backend/internal/explanations/repository"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
)

type PredictionReadService struct {
	explanations explanationrepo.Repository
}

func NewPredictionReadService(db repositories.Querier) PredictionReadService {
	return PredictionReadService{explanations: explanationrepo.New(db)}
}

func (s PredictionReadService) ExplanationByMatchID(ctx context.Context, matchID string, promptVersion string) (explanationmodels.PredictionExplanation, error) {
	return s.explanations.FindByMatchID(ctx, matchID, promptVersion)
}
