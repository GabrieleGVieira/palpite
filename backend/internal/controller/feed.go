package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/gabrielevieira/palpitai/backend/internal/usecase"
)

func GroupFeedHandler(cfg config.Config, feed usecase.FeedUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		page := queryInt(r, "page", 1)
		pageSize := queryInt(r, "pageSize", 20)
		response, err := feed.List(r.Context(), userID, r.PathValue("groupID"), page, pageSize)
		if err != nil {
			if apperrors.IsForbidden(err) {
				writeError(w, http.StatusForbidden, "Você precisa participar deste grupo.")
				return
			}

			writeError(w, http.StatusInternalServerError, "Não foi possível carregar o feed.")
			return
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func ReactToFeedEventHandler(cfg config.Config, feed usecase.FeedUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		var request dto.FeedReactionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "JSON invalido.")
			return
		}

		err = feed.React(r.Context(), userID, r.PathValue("groupID"), r.PathValue("eventID"), request.ReactionType)
		if err != nil {
			switch {
			case apperrors.IsValidation(err):
				writeError(w, http.StatusBadRequest, "Reação inválida.")
			case apperrors.IsForbidden(err):
				writeError(w, http.StatusForbidden, "Você precisa participar deste grupo.")
			case apperrors.IsNotFound(err):
				writeError(w, http.StatusNotFound, "Evento não encontrado.")
			default:
				writeError(w, http.StatusInternalServerError, "Não foi possível reagir.")
			}
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func DeleteFeedReactionHandler(cfg config.Config, feed usecase.FeedUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		err = feed.DeleteReaction(r.Context(), userID, r.PathValue("groupID"), r.PathValue("eventID"), r.URL.Query().Get("reactionType"))
		if err != nil {
			switch {
			case apperrors.IsValidation(err):
				writeError(w, http.StatusBadRequest, "Reação inválida.")
			case apperrors.IsForbidden(err):
				writeError(w, http.StatusForbidden, "Você precisa participar deste grupo.")
			case apperrors.IsNotFound(err):
				writeError(w, http.StatusNotFound, "Evento não encontrado.")
			default:
				writeError(w, http.StatusInternalServerError, "Não foi possível remover a reação.")
			}
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func queryInt(r *http.Request, key string, fallback int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
