package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
)

var errUnauthorized = errors.New("unauthorized")

type supabaseUserResponse struct {
	ID string `json:"id"`
}

func userIDFromRequest(r *http.Request, cfg config.Config) (string, error) {
	header := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || strings.TrimSpace(token) == "" {
		return "", errUnauthorized
	}

	return userIDFromToken(r, cfg, token)
}

func userIDFromToken(r *http.Request, cfg config.Config, token string) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", errUnauthorized
	}

	if strings.TrimSpace(cfg.SupabaseURL) == "" || strings.TrimSpace(cfg.SupabaseKey) == "" {
		return "", errUnauthorized
	}

	authURL, err := url.JoinPath(cfg.SupabaseURL, "/auth/v1/user")
	if err != nil {
		return "", errUnauthorized
	}

	request, err := http.NewRequestWithContext(r.Context(), http.MethodGet, authURL, nil)
	if err != nil {
		return "", errUnauthorized
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("apikey", cfg.SupabaseKey)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", errUnauthorized
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", errUnauthorized
	}

	var user supabaseUserResponse
	if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
		return "", errUnauthorized
	}

	if strings.TrimSpace(user.ID) == "" {
		return "", errUnauthorized
	}

	return user.ID, nil
}
