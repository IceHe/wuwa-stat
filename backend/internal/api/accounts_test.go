package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"wuwa/stat/backend/internal/config"
)

func TestBuildAccountServiceAccountsURL(t *testing.T) {
	tests := []struct {
		name string
		base string
		want string
	}{
		{
			name: "host root",
			base: "http://127.0.0.1:8765",
			want: "http://127.0.0.1:8765/api/accounts?active_only=true",
		},
		{
			name: "api base",
			base: "https://mgt.icehe.life/api",
			want: "https://mgt.icehe.life/api/accounts?active_only=true",
		},
		{
			name: "preserve query",
			base: "https://mgt.icehe.life/base?foo=bar",
			want: "https://mgt.icehe.life/base/api/accounts?active_only=true&foo=bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildAccountServiceAccountsURL(tt.base)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("url mismatch:\n got: %s\nwant: %s", got, tt.want)
			}
		})
	}
}

func TestPhoneTail(t *testing.T) {
	full := "13800138000"
	short := "123"

	tests := []struct {
		name string
		in   *string
		want string
	}{
		{name: "nil", in: nil, want: ""},
		{name: "full", in: &full, want: "8000"},
		{name: "short", in: &short, want: "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := phoneTail(tt.in); got != tt.want {
				t.Fatalf("tail mismatch: got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFetchActiveAccountsFiltersAndMapsUpstreamAccounts(t *testing.T) {
	phone := "13800138000"
	inactivePhone := "13900139000"
	var sawToken bool

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/accounts" {
			t.Fatalf("path mismatch: %s", r.URL.Path)
		}
		if r.URL.Query().Get("active_only") != "true" {
			t.Fatalf("missing active_only query")
		}
		if r.Header.Get("Authorization") == "Bearer test-token" && r.Header.Get("X-Token") == "test-token" {
			sawToken = true
		}

		writeJSON(w, http.StatusOK, []upstreamAccountResponse{
			{AccountID: 2, ID: "120000002", Abbr: "B", PhoneNumber: &inactivePhone, Nickname: "inactive", IsActive: false},
			{AccountID: 1, ID: "120000001", Abbr: "A", PhoneNumber: &phone, Nickname: "active", IsActive: true},
		})
	}))
	defer upstream.Close()

	api := &API{
		cfg: config.Config{
			AccountServiceURL:            upstream.URL,
			AccountServiceTimeoutSeconds: 3,
		},
	}

	got, err := api.fetchActiveAccounts(context.Background(), "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sawToken {
		t.Fatalf("upstream did not receive token headers")
	}
	if len(got) != 1 {
		body, _ := json.Marshal(got)
		t.Fatalf("account count mismatch: %s", body)
	}
	if got[0].ID != "120000001" || got[0].Abbr != "A" || got[0].PhoneTail != "8000" || got[0].Nickname != "active" {
		body, _ := json.Marshal(got[0])
		t.Fatalf("account mapping mismatch: %s", body)
	}
}
