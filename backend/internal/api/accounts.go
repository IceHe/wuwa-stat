package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type upstreamAccountResponse struct {
	AccountID   int     `json:"account_id"`
	ID          string  `json:"id"`
	Abbr        string  `json:"abbr"`
	PhoneNumber *string `json:"phone_number"`
	Nickname    string  `json:"nickname"`
	IsActive    bool    `json:"is_active"`
}

func (a *API) handleActiveAccounts(w http.ResponseWriter, r *http.Request, _ authContext) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	accounts, err := a.fetchActiveAccounts(r.Context(), extractToken(r))
	if err != nil {
		if authErr := asAuthError(err); authErr != nil {
			writeError(w, authErr.Status, authErr.Detail)
			return
		}
		log.Printf("account service request failed: %v", err)
		writeError(w, http.StatusServiceUnavailable, "账号服务不可用")
		return
	}

	writeJSON(w, http.StatusOK, accounts)
}

func (a *API) fetchActiveAccounts(ctx context.Context, token string) ([]activeAccountResponse, error) {
	endpoint, err := buildAccountServiceAccountsURL(a.cfg.AccountServiceURL)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(a.cfg.AccountServiceTimeoutSeconds * float64(time.Second))
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-Token", token)
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return nil, &authError{Status: http.StatusUnauthorized, Detail: authInvalidDetail}
	case http.StatusForbidden:
		return nil, &authError{Status: http.StatusForbidden, Detail: authForbiddenDetail}
	default:
		return nil, fmt.Errorf("account service unexpected status: %d", resp.StatusCode)
	}

	var upstream []upstreamAccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&upstream); err != nil {
		return nil, err
	}

	accounts := make([]activeAccountResponse, 0, len(upstream))
	for _, account := range upstream {
		if !account.IsActive {
			continue
		}
		accounts = append(accounts, activeAccountResponse{
			AccountID: account.AccountID,
			ID:        strings.TrimSpace(account.ID),
			Abbr:      strings.TrimSpace(account.Abbr),
			PhoneTail: phoneTail(account.PhoneNumber),
			Nickname:  strings.TrimSpace(account.Nickname),
			IsActive:  account.IsActive,
		})
	}

	sort.SliceStable(accounts, func(i, j int) bool {
		if accounts[i].Abbr != accounts[j].Abbr {
			return accounts[i].Abbr < accounts[j].Abbr
		}
		if accounts[i].AccountID != accounts[j].AccountID {
			return accounts[i].AccountID < accounts[j].AccountID
		}
		return accounts[i].ID < accounts[j].ID
	})

	return accounts, nil
}

func buildAccountServiceAccountsURL(base string) (string, error) {
	base = strings.TrimSpace(base)
	if base == "" {
		return "", fmt.Errorf("account service url is empty")
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("account service url is invalid")
	}

	basePath := strings.TrimRight(parsed.Path, "/")
	if strings.HasSuffix(basePath, "/api") {
		parsed.Path = basePath + "/accounts"
	} else {
		parsed.Path = basePath + "/api/accounts"
	}
	parsed.RawPath = ""

	query := parsed.Query()
	query.Set("active_only", "true")
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

func phoneTail(phoneNumber *string) string {
	if phoneNumber == nil {
		return ""
	}
	phone := strings.TrimSpace(*phoneNumber)
	if phone == "" {
		return ""
	}
	if len(phone) <= 4 {
		return phone
	}
	return phone[len(phone)-4:]
}
