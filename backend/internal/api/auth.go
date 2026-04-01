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

type authContext struct {
	UserID      int64
	Permissions []string
}

type authUser struct {
	ID          int64
	Name        string
	Permissions []string
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

func (v *authValidator) requireView(r *http.Request) (authContext, *authError) {
	return v.requirePermission(r, "view")
}

func (v *authValidator) requireEdit(r *http.Request) (authContext, *authError) {
	return v.requirePermission(r, "edit")
}

func (v *authValidator) requirePermission(r *http.Request, required string) (authContext, *authError) {
	token := extractToken(r)
	if token == "" {
		return authContext{}, &authError{Status: http.StatusUnauthorized, Detail: authInvalidDetail}
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

func (v *authValidator) validateToken(ctx context.Context, token string, required string) (authContext, *authError) {
	payload := map[string]string{"token": token}
	if required != "" {
		// Passing the required permission allows the auth service to short-circuit forbidden checks.
		payload["permission"] = required
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return authContext{}, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}

	url := strings.TrimRight(v.cfg.AuthServiceURL, "/") + "/api/validate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return authContext{}, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}
	req.Header.Set("Content-Type", "application/json")

	started := time.Now()
	resp, err := v.client.Do(req)
	if err != nil {
		log.Printf("auth service request failed: permission=%s token_fp=%s url=%s elapsed_ms=%d error=%v", required, tokenFingerprint(token), url, time.Since(started).Milliseconds(), err)
		return authContext{}, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		log.Printf("auth service upstream error: permission=%s token_fp=%s status_code=%d elapsed_ms=%d", required, tokenFingerprint(token), resp.StatusCode, time.Since(started).Milliseconds())
		return authContext{}, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return authContext{}, &authError{Status: http.StatusUnauthorized, Detail: authInvalidDetail}
	}

	if resp.StatusCode == http.StatusForbidden {
		return authContext{}, &authError{Status: http.StatusForbidden, Detail: authForbiddenDetail}
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("auth service unexpected status: permission=%s token_fp=%s status_code=%d elapsed_ms=%d", required, tokenFingerprint(token), resp.StatusCode, time.Since(started).Milliseconds())
		return authContext{}, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}

	var result struct {
		Valid       bool     `json:"valid"`
		ID          int64    `json:"id"`
		Reason      string   `json:"reason"`
		Permissions []string `json:"permissions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("auth service invalid json: permission=%s token_fp=%s elapsed_ms=%d error=%v", required, tokenFingerprint(token), time.Since(started).Milliseconds(), err)
		return authContext{}, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}

	if !result.Valid {
		if strings.EqualFold(result.Reason, "forbidden") {
			return authContext{}, &authError{Status: http.StatusForbidden, Detail: authForbiddenDetail}
		}
		return authContext{}, &authError{Status: http.StatusUnauthorized, Detail: authInvalidDetail}
	}

	if required != "" && !hasPermission(result.Permissions, required) {
		return authContext{}, &authError{Status: http.StatusForbidden, Detail: authForbiddenDetail}
	}

	if result.ID <= 0 {
		log.Printf("auth service returned invalid user id: permission=%s token_fp=%s user_id=%d", required, tokenFingerprint(token), result.ID)
		return authContext{}, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}

	return authContext{
		UserID:      result.ID,
		Permissions: result.Permissions,
	}, nil
}

func hasPermission(permissions []string, required string) bool {
	for _, permission := range permissions {
		if permission == "manage" || permission == required {
			return true
		}
	}
	return false
}

func hasExactPermission(permissions []string, target string) bool {
	for _, permission := range permissions {
		if permission == target {
			return true
		}
	}
	return false
}

func (v *authValidator) lookupUser(ctx context.Context, token string) (authUser, *authError) {
	payload := map[string]string{"token": token}
	body, err := json.Marshal(payload)
	if err != nil {
		return authUser{}, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}

	url := strings.TrimRight(v.cfg.AuthServiceURL, "/") + "/api/login"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return authUser{}, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		log.Printf("auth service user lookup failed: token_fp=%s url=%s error=%v", tokenFingerprint(token), url, err)
		return authUser{}, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return authUser{}, &authError{Status: http.StatusUnauthorized, Detail: authInvalidDetail}
	}
	if resp.StatusCode == http.StatusForbidden {
		return authUser{}, &authError{Status: http.StatusForbidden, Detail: authForbiddenDetail}
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("auth service user lookup unexpected status: token_fp=%s status_code=%d", tokenFingerprint(token), resp.StatusCode)
		return authUser{}, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}

	var result struct {
		ID          int64    `json:"id"`
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("auth service user lookup invalid json: token_fp=%s error=%v", tokenFingerprint(token), err)
		return authUser{}, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}
	if result.ID <= 0 || strings.TrimSpace(result.Name) == "" {
		log.Printf("auth service user lookup returned invalid payload: token_fp=%s user_id=%d name=%q", tokenFingerprint(token), result.ID, result.Name)
		return authUser{}, &authError{Status: http.StatusServiceUnavailable, Detail: authUnavailableDetail}
	}

	return authUser{
		ID:          result.ID,
		Name:        strings.TrimSpace(result.Name),
		Permissions: result.Permissions,
	}, nil
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
