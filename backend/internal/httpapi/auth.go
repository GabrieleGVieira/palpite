package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
)

var errUnauthorized = errors.New("unauthorized")

type supabaseUserResponse = dto.SupabaseUserResponse

func userIDFromRequest(r *http.Request, cfg config.Config) (string, error) {
	header := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || strings.TrimSpace(token) == "" {
		return "", errUnauthorized
	}

	return userIDFromToken(r, cfg, token)
}

func userIDAndDisplayNameFromRequest(r *http.Request, cfg config.Config) (string, string, error) {
	header := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || strings.TrimSpace(token) == "" {
		return "", "", errUnauthorized
	}

	user, err := userFromToken(r, cfg, token)
	if err != nil {
		return "", "", err
	}

	return user.ID, userDisplayName(user), nil
}

func userIDFromToken(r *http.Request, cfg config.Config, token string) (string, error) {
	user, err := userFromToken(r, cfg, token)
	if err != nil {
		return "", err
	}

	return user.ID, nil
}

func userFromToken(r *http.Request, cfg config.Config, token string) (supabaseUserResponse, error) {
	if strings.TrimSpace(token) == "" {
		return supabaseUserResponse{}, errUnauthorized
	}

	if strings.TrimSpace(cfg.SupabaseURL) == "" || strings.TrimSpace(cfg.SupabaseKey) == "" {
		return supabaseUserResponse{}, errUnauthorized
	}

	authURL, err := url.JoinPath(cfg.SupabaseURL, "/auth/v1/user")
	if err != nil {
		return supabaseUserResponse{}, errUnauthorized
	}

	request, err := http.NewRequestWithContext(r.Context(), http.MethodGet, authURL, nil)
	if err != nil {
		return supabaseUserResponse{}, errUnauthorized
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("apikey", cfg.SupabaseKey)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return supabaseUserResponse{}, errUnauthorized
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return supabaseUserResponse{}, errUnauthorized
	}

	var user supabaseUserResponse
	if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
		return supabaseUserResponse{}, errUnauthorized
	}

	if strings.TrimSpace(user.ID) == "" {
		return supabaseUserResponse{}, errUnauthorized
	}

	return user, nil
}

func userDisplayName(user supabaseUserResponse) string {
	name := strings.TrimSpace(user.UserMetadata.FullName)
	if name != "" {
		return name
	}

	email := strings.TrimSpace(user.Email)
	if email != "" {
		return email
	}

	return user.ID
}
