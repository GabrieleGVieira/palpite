package service

import (
	"context"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/goalpredictions/models"
	"github.com/gabrielevieira/palpitai/backend/internal/goalpredictions/repository"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
)

type Service struct {
	db repositories.Querier
}

func New(db repositories.Querier) Service {
	return Service{db: db}
}

func (s Service) ByMatchID(ctx context.Context, matchID string) (models.MatchGoalPrediction, error) {
	return repository.ByMatchID(ctx, s.db, matchID)
}

func (s Service) ByDateAndTeams(ctx context.Context, matchDate time.Time, homeTeamID string, awayTeamID string) (models.MatchGoalPrediction, error) {
	return repository.ByDateAndTeams(ctx, s.db, matchDate, homeTeamID, awayTeamID)
}
