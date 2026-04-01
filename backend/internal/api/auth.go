package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"wuwa/stat/backend/internal/config"
)

const (
	authInvalidDetail     = "Token 无效或已过期"
	authForbiddenDetail   = "权限不足"
	authUnavailableDetail = "鉴权服务不可用"
)

type authValidator struct {
	client *http.Client
	cfg    config.Config
}

type authError struct {
	Status int
	Detail string
}

func (e *authError) Error() string {
	return e.Detail
}

func newAuthValidator(cfg config.Config) *authValidator {
	timeout := time.Duration(cfg.AuthServiceTimeoutSeconds * float64(time.Second))
	return &authValidator{
		client: &http.Client{Timeout: timeout},
		cfg:    cfg,
	}
}

func (v *authValidator) requireView(r *http.Request) ([]string, *authError) {
	return v.requirePermission(r, "view")
}

func (v *authValidator) requireEdit(r *http.Request) ([]string, *authError) {
	return v.requirePermission(r, "edit")
}

func (v *authValidator) requirePermission(r *http.Request, required string) ([]string, *authError) {
	token := extractToken(r)
	if token == "" {
		return nil, &authError{Status: http.StatusUnauthorized, Detail: authInvalidDetail}
	}
	return v.validateToken(r.Context(), token, required)
}

func extractToken(r *http.Request) string {
	// Prefer standard Bearer token, with X-Token kept for compatibility with existing clients.
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if authorization != "" {
		parts := strings.SplitN(authorization, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && strings.TrimSpace(parts[1]) != "" {
			return strings.TrimSpace(parts[1])
		}
	}

	xToken := strings.TrimSpace(r.Header.Get("X-Token"))
	if xToken != "" {
		return xToken
	}

	return ""
}

func (v *authValidator) validateToken(ctx context.Context, token string, required string) ([]string, *authError) {
	payload := map[string]string{"token": token}
	if required != "" {
		// Passing the required permission allows the auth service to short-circuit forbidden checks.
		payload["permission"] = required
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}

	url := strings.TrimRight(v.cfg.AuthServiceURL, "/") + "/api/validate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}
	req.Header.Set("Content-Type", "application/json")

	started := time.Now()
	resp, err := v.client.Do(req)
	if err != nil {
		log.Printf("auth service request failed: permission=%s token_fp=%s url=%s elapsed_ms=%d error=%v", required, tokenFingerprint(token), url, time.Since(started).Milliseconds(), err)
		return nil, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		log.Printf("auth service upstream error: permission=%s token_fp=%s status_code=%d elapsed_ms=%d", required, tokenFingerprint(token), resp.StatusCode, time.Since(started).Milliseconds())
		return nil, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, &authError{Status: http.StatusUnauthorized, Detail: authInvalidDetail}
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, &authError{Status: http.StatusForbidden, Detail: authForbiddenDetail}
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("auth service unexpected status: permission=%s token_fp=%s status_code=%d elapsed_ms=%d", required, tokenFingerprint(token), resp.StatusCode, time.Since(started).Milliseconds())
		return nil, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}

	var result struct {
		Valid       bool     `json:"valid"`
		Reason      string   `json:"reason"`
		Permissions []string `json:"permissions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("auth service invalid json: permission=%s token_fp=%s elapsed_ms=%d error=%v", required, tokenFingerprint(token), time.Since(started).Milliseconds(), err)
		return nil, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}

	if !result.Valid {
		if strings.EqualFold(result.Reason, "forbidden") {
			return nil, &authError{Status: http.StatusForbidden, Detail: authForbiddenDetail}
		}
		return nil, &authError{Status: http.StatusUnauthorized, Detail: authInvalidDetail}
	}

	if required != "" && !hasPermission(result.Permissions, required) {
		return nil, &authError{Status: http.StatusForbidden, Detail: authForbiddenDetail}
	}

	return result.Permissions, nil
}

func hasPermission(permissions []string, required string) bool {
	for _, permission := range permissions {
		if permission == "manage" || permission == required {
			return true
		}
	}
	return false
}

func tokenFingerprint(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])[:12]
}

func asAuthError(err error) *authError {
	var authErr *authError
	if errors.As(err, &authErr) {
		return authErr
	}
	return nil
}
