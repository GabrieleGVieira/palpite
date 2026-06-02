package google

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
)

const directoryScope = "https://www.googleapis.com/auth/admin.directory.group.member"

type GroupMemberAdder interface {
	AddMember(ctx context.Context, email string) error
}

type DirectoryGroupMemberAdder struct {
	serviceAccountEmail string
	privateKey          *rsa.PrivateKey
	groupEmail          string
	delegatedAdminEmail string
	httpClient          *http.Client
}

type LocalGroupMemberAdder struct {
	groupEmail string
	logger     *slog.Logger
}

func NewGroupMemberAdder(cfg config.Config, logger *slog.Logger) GroupMemberAdder {
	if logger == nil {
		logger = slog.Default()
	}

	if strings.TrimSpace(cfg.GoogleServiceAccountEmail) == "" ||
		strings.TrimSpace(cfg.GooglePrivateKey) == "" ||
		strings.TrimSpace(cfg.GoogleGroupEmail) == "" {
		return LocalGroupMemberAdder{
			groupEmail: strings.TrimSpace(cfg.GoogleGroupEmail),
			logger:     logger,
		}
	}

	key, err := parsePrivateKey(cfg.GooglePrivateKey)
	if err != nil {
		logger.Error("google group private key invalid, using local fallback", "error", err)
		return LocalGroupMemberAdder{
			groupEmail: strings.TrimSpace(cfg.GoogleGroupEmail),
			logger:     logger,
		}
	}

	return DirectoryGroupMemberAdder{
		serviceAccountEmail: strings.TrimSpace(cfg.GoogleServiceAccountEmail),
		privateKey:          key,
		groupEmail:          strings.TrimSpace(cfg.GoogleGroupEmail),
		delegatedAdminEmail: strings.TrimSpace(cfg.GoogleWorkspaceDelegatedAdmin),
		httpClient:          &http.Client{Timeout: 12 * time.Second},
	}
}

func (adder DirectoryGroupMemberAdder) AddMember(ctx context.Context, email string) error {
	token, err := adder.accessToken(ctx)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]string{
		"email": email,
		"role":  "MEMBER",
	})
	if err != nil {
		return err
	}

	endpoint := "https://admin.googleapis.com/admin/directory/v1/groups/" +
		url.PathEscape(adder.groupEmail) + "/members"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	response, err := adder.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK || response.StatusCode == http.StatusCreated {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
	if response.StatusCode == http.StatusConflict || strings.Contains(string(body), "Member already exists") {
		return nil
	}

	return fmt.Errorf("directory api returned %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
}

func (adder DirectoryGroupMemberAdder) accessToken(ctx context.Context) (string, error) {
	assertion, err := adder.jwtAssertion()
	if err != nil {
		return "", err
	}

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", assertion)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := adder.httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 4096))
	if err != nil {
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("oauth token request returned %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	if payload.AccessToken == "" {
		return "", errors.New("oauth token response missing access_token")
	}

	return payload.AccessToken, nil
}

func (adder DirectoryGroupMemberAdder) jwtAssertion() (string, error) {
	now := time.Now()
	header := map[string]string{
		"alg": "RS256",
		"typ": "JWT",
	}
	claims := map[string]any{
		"iss":   adder.serviceAccountEmail,
		"scope": directoryScope,
		"aud":   "https://oauth2.googleapis.com/token",
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	}
	if adder.delegatedAdminEmail != "" {
		claims["sub"] = adder.delegatedAdminEmail
	}

	encodedHeader, err := encodeJWTPart(header)
	if err != nil {
		return "", err
	}
	encodedClaims, err := encodeJWTPart(claims)
	if err != nil {
		return "", err
	}

	signingInput := encodedHeader + "." + encodedClaims
	hash := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, adder.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (adder LocalGroupMemberAdder) AddMember(_ context.Context, email string) error {
	adder.logger.Info("google group local fallback accepted beta tester", "email", email, "group", adder.groupEmail)
	return nil
}

func parsePrivateKey(value string) (*rsa.PrivateKey, error) {
	normalized := strings.ReplaceAll(strings.TrimSpace(value), `\n`, "\n")
	block, _ := pem.Decode([]byte(normalized))
	if block == nil {
		return nil, errors.New("missing pem block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("private key is not rsa")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func encodeJWTPart(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(payload), nil
}
