package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

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

func TestDeductEnergyForPlayersSplitsSpendCosts(t *testing.T) {
	var spentCosts []int

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/accounts/by-id/120000001":
			writeJSON(w, http.StatusOK, upstreamAccountResponse{
				AccountID:               1,
				ID:                      "120000001",
				IsActive:                true,
				CurrentWaveplate:        180,
				CurrentWaveplateCrystal: 0,
			})
		case "/api/accounts/by-id/120000001/energy/spend":
			var payload struct {
				Cost int `json:"cost"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("invalid spend payload: %v", err)
			}
			spentCosts = append(spentCosts, payload.Cost)
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	api := &API{
		cfg: config.Config{
			AccountServiceURL:            upstream.URL,
			AccountServiceTimeoutSeconds: 3,
		},
	}

	err := api.deductEnergyForPlayers(context.Background(), "test-token", map[string]int{"120000001": 180})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []int{120, 60}
	if !reflect.DeepEqual(spentCosts, want) {
		t.Fatalf("spend costs mismatch: got %v, want %v", spentCosts, want)
	}
}

func TestDeductEnergyForPlayersReturnsInsufficientBeforeSpend(t *testing.T) {
	var spendCalled bool

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/accounts/by-id/120000001":
			writeJSON(w, http.StatusOK, upstreamAccountResponse{
				AccountID:               1,
				ID:                      "120000001",
				IsActive:                true,
				CurrentWaveplate:        40,
				CurrentWaveplateCrystal: 0,
			})
		case "/api/accounts/by-id/120000001/energy/spend":
			spendCalled = true
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	api := &API{
		cfg: config.Config{
			AccountServiceURL:            upstream.URL,
			AccountServiceTimeoutSeconds: 3,
		},
	}

	err := api.deductEnergyForPlayers(context.Background(), "test-token", map[string]int{"120000001": 60})
	if !errors.Is(err, errEnergyInsufficient) {
		t.Fatalf("expected insufficient error, got %v", err)
	}
	if spendCalled {
		t.Fatalf("spend endpoint should not be called when precheck is insufficient")
	}
}

func TestRefundEnergyForPlayersSplitsGainAmounts(t *testing.T) {
	var gainedAmounts []int

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/accounts/by-id/120000001":
			writeJSON(w, http.StatusOK, upstreamAccountResponse{
				AccountID: 1,
				ID:        "120000001",
				IsActive:  true,
			})
		case "/api/accounts/by-id/120000001/energy/gain":
			var payload struct {
				Amount int `json:"amount"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("invalid gain payload: %v", err)
			}
			gainedAmounts = append(gainedAmounts, payload.Amount)
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	api := &API{
		cfg: config.Config{
			AccountServiceURL:            upstream.URL,
			AccountServiceTimeoutSeconds: 3,
		},
	}

	err := api.refundEnergyForPlayers(context.Background(), "test-token", map[string]int{"120000001": 80})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []int{40, 40}
	if !reflect.DeepEqual(gainedAmounts, want) {
		t.Fatalf("gain amounts mismatch: got %v, want %v", gainedAmounts, want)
	}
}

func TestRecentRecordEnergyCostOnlyRefundsWithinOneHour(t *testing.T) {
	recent := deleteEnergyRecord{ClaimCount: 2, CreatedAt: time.Now().Add(-30 * time.Minute)}
	if got := recentRecordEnergyCost(recent, 40); got != 80 {
		t.Fatalf("recent refund mismatch: got %d, want 80", got)
	}

	old := deleteEnergyRecord{ClaimCount: 2, CreatedAt: time.Now().Add(-61 * time.Minute)}
	if got := recentRecordEnergyCost(old, 40); got != 0 {
		t.Fatalf("old refund mismatch: got %d, want 0", got)
	}
}
