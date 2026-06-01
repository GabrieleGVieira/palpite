package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/gabrielevieira/palpitai/backend/internal/usecase"
)

type WalletService interface {
	GetBalance(ctx context.Context, userID string) (dto.WalletResponse, error)
	ListTransactions(ctx context.Context, userID string, limit int, offset int) (dto.PalpicoinTransactionPageResponse, error)
	Ranking(ctx context.Context, userID string) (dto.PalpicoinRankingResponse, error)
}

type ChallengeService interface {
	Accept(ctx context.Context, userID string, challengeID string) (domain.PalpicoinChallenge, error)
	Cancel(ctx context.Context, userID string, challengeID string) (domain.PalpicoinChallenge, error)
	Create(ctx context.Context, creatorUserID string, request dto.CreateChallengeRequest) (domain.PalpicoinChallenge, error)
	Decline(ctx context.Context, userID string, challengeID string) (domain.PalpicoinChallenge, error)
	Get(ctx context.Context, userID string, challengeID string) (dto.ChallengeResponse, error)
	List(ctx context.Context, userID string, status string, challengeType string) (dto.ChallengeListResponse, error)
}

func GetWalletHandler(cfg config.Config, wallet WalletService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}
		response, err := wallet.GetBalance(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Não foi possível carregar sua carteira.")
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func ListWalletTransactionsHandler(cfg config.Config, wallet WalletService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}
		response, err := wallet.ListTransactions(r.Context(), userID, queryInt(r, "limit", 20), queryInt(r, "offset", 0))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Não foi possível carregar o histórico.")
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func PalpicoinRankingHandler(cfg config.Config, wallet WalletService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}
		response, err := wallet.Ranking(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Não foi possível carregar o ranking.")
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func CreateChallengeHandler(cfg config.Config, challenges ChallengeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}
		var request dto.CreateChallengeRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "JSON invalido.")
			return
		}
		challenge, err := challenges.Create(r.Context(), userID, request)
		if err != nil {
			writeChallengeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, challenge)
	}
}

func AcceptChallengeHandler(cfg config.Config, challenges ChallengeService) http.HandlerFunc {
	return challengeActionHandler(cfg, challenges.Accept)
}

func DeclineChallengeHandler(cfg config.Config, challenges ChallengeService) http.HandlerFunc {
	return challengeActionHandler(cfg, challenges.Decline)
}

func CancelChallengeHandler(cfg config.Config, challenges ChallengeService) http.HandlerFunc {
	return challengeActionHandler(cfg, challenges.Cancel)
}

func challengeActionHandler(cfg config.Config, action func(context.Context, string, string) (domain.PalpicoinChallenge, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}
		challenge, err := action(r.Context(), userID, r.PathValue("id"))
		if err != nil {
			writeChallengeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, challenge)
	}
}

func ListChallengesHandler(cfg config.Config, challenges ChallengeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}
		response, err := challenges.List(r.Context(), userID, r.URL.Query().Get("status"), r.URL.Query().Get("type"))
		if err != nil {
			writeChallengeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func GetChallengeHandler(cfg config.Config, challenges ChallengeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}
		response, err := challenges.Get(r.Context(), userID, r.PathValue("id"))
		if err != nil {
			writeChallengeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func writeChallengeError(w http.ResponseWriter, err error) {
	switch {
	case apperrors.IsValidation(err):
		writeError(w, http.StatusBadRequest, err.Error())
	case apperrors.IsForbidden(err):
		writeError(w, http.StatusForbidden, err.Error())
	case apperrors.IsConflict(err):
		if err == usecase.ErrInsufficientBalance {
			writeError(w, http.StatusConflict, "Saldo insuficiente de Palpicoins.")
			return
		}
		writeError(w, http.StatusConflict, err.Error())
	case apperrors.IsNotFound(err):
		writeError(w, http.StatusNotFound, "Desafio não encontrado.")
	default:
		writeError(w, http.StatusInternalServerError, "Não foi possível processar o desafio.")
	}
}
