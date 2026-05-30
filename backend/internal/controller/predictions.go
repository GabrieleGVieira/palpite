package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gabrielevieira/palpitai/backend/internal/ai"
	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	explanationmodels "github.com/gabrielevieira/palpitai/backend/internal/explanations/models"
	goalmodels "github.com/gabrielevieira/palpitai/backend/internal/goalpredictions/models"
	predictionmodels "github.com/gabrielevieira/palpitai/backend/internal/predictions/models"
	"github.com/gabrielevieira/palpitai/backend/internal/usecase"
)

type RealtimePublisher interface {
	Publish(ctx context.Context, event domain.Event)
}

type PredictionReader interface {
	ExplanationByMatchID(ctx context.Context, matchID string, promptVersion string) (explanationmodels.PredictionExplanation, error)
	GoalPredictionByMatchID(ctx context.Context, matchID string) (goalmodels.MatchGoalPrediction, error)
	MatchPredictionByMatchID(ctx context.Context, matchID string) (predictionmodels.MatchPrediction, error)
}

func UserScoreHandler(cfg config.Config, predictions usecase.PredictionUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		totalScore, err := predictions.UserTotalScore(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Não foi possivel carregar sua pontuacao.")
			return
		}

		writeJSON(w, http.StatusOK, map[string]int{
			"total_points": totalScore,
		})
	}
}

func GetMatchPredictionHandler(cfg config.Config, predictions PredictionReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := userIDFromRequest(r, cfg); err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		matchID := r.PathValue("matchID")
		prediction, err := predictions.MatchPredictionByMatchID(r.Context(), matchID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				writeError(w, http.StatusNotFound, "Previsão ainda não disponível para este jogo.")
				return
			}

			writeError(w, http.StatusInternalServerError, "Não foi possível carregar a previsão.")
			return
		}

		response := dto.MatchPredictionAIResponse{
			MatchID: predictionResponseMatchID(prediction, matchID),
			Probabilities: dto.MatchPredictionProbabilities{
				AwayWin: probabilityPercent(prediction.AwayWinProbability),
				Draw:    probabilityPercent(prediction.DrawProbability),
				HomeWin: probabilityPercent(prediction.HomeWinProbability),
			},
		}

		goalPrediction, err := predictions.GoalPredictionByMatchID(r.Context(), matchID)
		if err == nil {
			response.Goals = mapGoalPredictionResponse(goalPrediction)
			response.TopScores = mapTopScoreResponses(goalPrediction.TopScoreProbabilities)
		} else if !apperrors.IsNotFound(err) {
			writeError(w, http.StatusInternalServerError, "Não foi possível carregar a previsão.")
			return
		}

		explanation, err := predictions.ExplanationByMatchID(r.Context(), matchID, ai.PromptVersionPredictionExplanationV1)
		if err == nil {
			response.Explanation = &dto.PredictionExplanationResponse{
				BetStyle:    explanation.BetStyle,
				MainReasons: explanation.MainReasons,
				RiskAlert:   explanation.RiskAlert,
				Summary:     explanation.Summary,
				UserTip:     explanation.UserTip,
			}
		} else if !apperrors.IsNotFound(err) {
			writeError(w, http.StatusInternalServerError, "Não foi possível carregar a previsão.")
			return
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func GroupRankingHandler(cfg config.Config, predictions usecase.PredictionUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		ranking, err := predictions.GroupRanking(r.Context(), userID, r.PathValue("groupID"))
		if err != nil {
			if apperrors.IsForbidden(err) {
				writeError(w, http.StatusForbidden, "Você precisa participar deste grupo.")
				return
			}

			fmt.Printf("Error loading group ranking: %v\n", err)

			writeError(w, http.StatusInternalServerError, "Não foi possível carregar o ranking.")
			return
		}

		writeJSON(w, http.StatusOK, map[string][]dto.RankingEntryResponse{
			"ranking": ranking,
		})
	}
}

func predictionResponseMatchID(prediction predictionmodels.MatchPrediction, fallback string) string {
	if prediction.MatchID != nil {
		return *prediction.MatchID
	}
	return fallback
}

func mapGoalPredictionResponse(prediction goalmodels.MatchGoalPrediction) *dto.MatchPredictionGoalsResponse {
	var mostLikelyScore *string
	if prediction.MostLikelyHomeScore != nil && prediction.MostLikelyAwayScore != nil {
		score := fmt.Sprintf("%d x %d", *prediction.MostLikelyHomeScore, *prediction.MostLikelyAwayScore)
		mostLikelyScore = &score
	}

	return &dto.MatchPredictionGoalsResponse{
		ExpectedAwayGoals: prediction.ExpectedAwayGoals,
		ExpectedHomeGoals: prediction.ExpectedHomeGoals,
		MostLikelyScore:   mostLikelyScore,
	}
}

func mapTopScoreResponses(scores []goalmodels.MatchScoreProbability) []dto.PredictionScoreResponse {
	if len(scores) == 0 {
		return nil
	}

	responses := make([]dto.PredictionScoreResponse, 0, len(scores))
	for _, score := range scores {
		responses = append(responses, dto.PredictionScoreResponse{
			Probability: probabilityPercent(score.Probability),
			Score:       fmt.Sprintf("%d x %d", score.HomeScore, score.AwayScore),
		})
	}
	return responses
}

func probabilityPercent(value float64) float64 {
	return value * 100
}

func ListGroupMatchesHandler(cfg config.Config, predictions usecase.PredictionUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		matches, err := predictions.ListGroupMatches(r.Context(), userID, r.PathValue("groupID"))
		if err != nil {
			if apperrors.IsForbidden(err) {
				writeError(w, http.StatusForbidden, "Você precisa participar deste grupo.")
				return
			}

			writeError(w, http.StatusInternalServerError, "Não foi possível listar os jogos.")
			return
		}

		writeJSON(w, http.StatusOK, map[string][]dto.MatchResponse{
			"matches": matches,
		})
	}
}

func SavePredictionHandler(cfg config.Config, predictions usecase.PredictionUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		var request dto.PredictionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "JSON invalido.")
			return
		}

		if request.HomeScore < 0 || request.HomeScore > 99 || request.AwayScore < 0 || request.AwayScore > 99 {
			writeError(w, http.StatusBadRequest, "Informe placares entre 0 e 99.")
			return
		}

		prediction, err := predictions.SavePrediction(
			r.Context(),
			userID,
			r.PathValue("groupID"),
			r.PathValue("matchID"),
			request,
		)
		if err != nil {
			switch {
			case err == usecase.ErrPaymentRequired:
				writeError(w, http.StatusForbidden, "Pagamento pendente. Fale com o admin do bolão para liberar seus palpites.")
			case apperrors.IsForbidden(err):
				writeError(w, http.StatusForbidden, "Você precisa participar deste grupo.")
			case apperrors.IsConflict(err):
				writeError(w, http.StatusConflict, "O jogo já começou. Não é mais possível editar o palpite.")
			default:
				writeError(w, http.StatusInternalServerError, "Não foi possível salvar o palpite.")
			}
			return
		}

		writeJSON(w, http.StatusOK, prediction)
	}
}

func SaveMatchResultHandler(cfg config.Config, predictions usecase.PredictionUsecase, publisher RealtimePublisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := userIDFromRequest(r, cfg); err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		var request dto.MatchResultRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "JSON invalido.")
			return
		}

		if request.HomeScore < 0 || request.HomeScore > 99 || request.AwayScore < 0 || request.AwayScore > 99 {
			writeError(w, http.StatusBadRequest, "Informe placares entre 0 e 99.")
			return
		}

		scoredPredictions, err := predictions.SaveMatchResult(r.Context(), r.PathValue("matchID"), request)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Não foi possivel salvar o resultado.")
			return
		}

		if publisher != nil {
			details, _ := predictions.MatchDetailsByID(r.Context(), r.PathValue("matchID"))
			groups, _ := predictions.GroupsAffectedByMatch(r.Context(), r.PathValue("matchID"))
			resultMessage := usecase.FormatResultMessage(details.HomeTeam, details.AwayTeam, request.HomeScore, request.AwayScore)

			publisher.Publish(r.Context(), domain.Event{
				Name: "match.finished",
				Payload: map[string]any{
					"away_score": request.AwayScore,
					"away_team":  details.AwayTeam,
					"home_score": request.HomeScore,
					"home_team":  details.HomeTeam,
					"match_id":   r.PathValue("matchID"),
					"message":    resultMessage,
					"status":     "finished",
				},
				Room: "matches",
			})

			if scoredPredictions > 0 {
				for _, group := range groups {
					payload := map[string]any{
						"away_score": request.AwayScore,
						"away_team":  details.AwayTeam,
						"group_id":   group.ID,
						"group_name": group.Name,
						"home_score": request.HomeScore,
						"home_team":  details.HomeTeam,
						"match_id":   r.PathValue("matchID"),
						"message":    "Ranking do grupo " + group.Name + " atualizado",
					}

					publisher.Publish(r.Context(), domain.Event{
						Name:    "ranking.updated",
						Payload: payload,
						Room:    "rankings",
					})
					publisher.Publish(r.Context(), domain.Event{
						Name:    "ranking.updated",
						Payload: payload,
						Room:    "group:" + group.ID,
					})
				}
			}
		}

		writeJSON(w, http.StatusOK, map[string]int{
			"scored_predictions": scoredPredictions,
		})
	}
}
