package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
)

type FriendsService interface {
	Accept(ctx context.Context, userID string, friendshipID string) (domain.Friendship, error)
	CreateRequest(ctx context.Context, requesterUserID string, addresseeUserID string) (domain.Friendship, error)
	Decline(ctx context.Context, userID string, friendshipID string) (domain.Friendship, error)
	Delete(ctx context.Context, userID string, friendshipID string) error
	ListFriends(ctx context.Context, userID string) ([]dto.FriendResponse, error)
	ListPendingRequests(ctx context.Context, userID string) ([]dto.PendingFriendRequestResponse, error)
	PublicProfile(ctx context.Context, requesterUserID string, profileUserID string) (dto.PublicProfileResponse, error)
	SearchUsers(ctx context.Context, userID string, query string) ([]dto.UserSearchResponse, error)
}

func CreateFriendRequestHandler(cfg config.Config, friends FriendsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		var request dto.FriendRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "JSON invalido.")
			return
		}

		friendship, err := friends.CreateRequest(r.Context(), userID, request.UserID)
		if err != nil {
			writeFriendshipError(w, err, "Não foi possivel enviar a solicitacao.")
			return
		}

		writeJSON(w, http.StatusCreated, friendship)
	}
}

func AcceptFriendRequestHandler(cfg config.Config, friends FriendsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		friendship, err := friends.Accept(r.Context(), userID, r.PathValue("id"))
		if err != nil {
			writeFriendshipError(w, err, "Não foi possivel aceitar a solicitacao.")
			return
		}

		writeJSON(w, http.StatusOK, friendship)
	}
}

func DeclineFriendRequestHandler(cfg config.Config, friends FriendsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		friendship, err := friends.Decline(r.Context(), userID, r.PathValue("id"))
		if err != nil {
			writeFriendshipError(w, err, "Não foi possivel recusar a solicitacao.")
			return
		}

		writeJSON(w, http.StatusOK, friendship)
	}
}

func DeleteFriendshipHandler(cfg config.Config, friends FriendsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		if err := friends.Delete(r.Context(), userID, r.PathValue("id")); err != nil {
			writeFriendshipError(w, err, "Não foi possivel remover a amizade.")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"message": "Amizade removida."})
	}
}

func ListFriendsHandler(cfg config.Config, friends FriendsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		items, err := friends.ListFriends(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Não foi possivel listar seus amigos.")
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

func ListFriendRequestsHandler(cfg config.Config, friends FriendsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		items, err := friends.ListPendingRequests(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Não foi possivel listar as solicitacoes.")
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

func SearchUsersHandler(cfg config.Config, friends FriendsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		items, err := friends.SearchUsers(r.Context(), userID, r.URL.Query().Get("q"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Não foi possivel buscar usuarios.")
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

func PublicProfileHandler(cfg config.Config, friends FriendsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		profile, err := friends.PublicProfile(r.Context(), userID, r.PathValue("id"))
		if err != nil {
			writeFriendshipError(w, err, "Não foi possivel carregar o perfil.")
			return
		}

		writeJSON(w, http.StatusOK, profile)
	}
}

func writeFriendshipError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case apperrors.IsValidation(err):
		writeError(w, http.StatusBadRequest, err.Error())
	case apperrors.IsNotFound(err):
		writeError(w, http.StatusNotFound, "Usuario ou amizade não encontrado.")
	case apperrors.IsForbidden(err):
		writeError(w, http.StatusForbidden, "Você não pode executar esta ação.")
	case apperrors.IsConflict(err):
		writeError(w, http.StatusConflict, "Esta amizade não pode ser alterada agora.")
	default:
		writeError(w, http.StatusInternalServerError, fallback)
	}
}
