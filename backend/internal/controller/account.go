package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
)

type AccountDeletionUsecase interface {
	DeleteAccount(ctx context.Context, userID string) error
}

type AccountProfileUsecase interface {
	Profile(ctx context.Context, userID string) (dto.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, request dto.UpdateProfileRequest) (dto.ProfileResponse, error)
}

func GetProfileHandler(cfg config.Config, accounts AccountProfileUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		profile, err := accounts.Profile(r.Context(), userID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				writeError(w, http.StatusNotFound, "Perfil não encontrado.")
				return
			}

			writeError(w, http.StatusInternalServerError, "Não foi possível carregar o perfil.")
			return
		}

		writeJSON(w, http.StatusOK, profile)
	}
}

func UpdateProfileHandler(cfg config.Config, accounts AccountProfileUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		var request dto.UpdateProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "JSON invalido.")
			return
		}

		profile, err := accounts.UpdateProfile(r.Context(), userID, request)
		if err != nil {
			switch {
			case apperrors.IsValidation(err):
				writeError(w, http.StatusBadRequest, err.Error())
			case apperrors.IsNotFound(err):
				writeError(w, http.StatusNotFound, "Perfil não encontrado.")
			default:
				writeError(w, http.StatusInternalServerError, "Não foi possível atualizar o perfil.")
			}
			return
		}

		writeJSON(w, http.StatusOK, profile)
	}
}

func DeleteAccountHandler(cfg config.Config, accounts AccountDeletionUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			fmt.Printf("account deletion failed: %v\n", err)
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		if err := accounts.DeleteAccount(r.Context(), userID); err != nil {
			if apperrors.IsConflict(err) {
				writeError(w, http.StatusConflict, "Transfira a propriedade dos grupos que você administra antes de excluir sua conta.")
				return
			}

			fmt.Printf("account deletion failed: %v\n", err)
			slog.Error("account deletion failed", "error", err)
			writeError(w, http.StatusInternalServerError, "Não foi possível excluir a conta agora.")
			return
		}

		if err := deleteSupabaseAuthUser(r, cfg, userID); err != nil {
			fmt.Printf("account deletion failed: %v\n", err)
			slog.Error("supabase auth user deletion failed", "error", err)
		}

		slog.Info("account deletion processed")
		writeJSON(w, http.StatusOK, map[string]string{
			"message": "Conta marcada para exclusão e dados pessoais anonimizados.",
		})
	}
}

func deleteSupabaseAuthUser(r *http.Request, cfg config.Config, userID string) error {
	if strings.TrimSpace(cfg.SupabaseURL) == "" || strings.TrimSpace(cfg.SupabaseServiceRoleKey) == "" {
		return nil
	}

	endpoint, err := url.JoinPath(cfg.SupabaseURL, "/auth/v1/admin/users", userID)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(r.Context(), http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+cfg.SupabaseServiceRoleKey)
	request.Header.Set("apikey", cfg.SupabaseServiceRoleKey)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return errors.New("supabase auth delete failed")
	}

	return nil
}
