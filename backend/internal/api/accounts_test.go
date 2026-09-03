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

func TestBuildAccountServiceDashboardAccountsURL(t *testing.T) {
	tests := []struct {
		name string
		base string
		want string
	}{
		{
			name: "host root",
			base: "http://127.0.0.1:8765",
			want: "http://127.0.0.1:8765/api/dashboard/accounts",
		},
		{
			name: "api base",
			base: "https://mgt.icehe.life/api",
			want: "https://mgt.icehe.life/api/dashboard/accounts",
		},
		{
			name: "preserve query",
			base: "https://mgt.icehe.life/base?foo=bar",
			want: "https://mgt.icehe.life/base/api/dashboard/accounts?foo=bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildAccountServiceAPIURL(tt.base, "dashboard", "accounts")
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

func TestFetchActiveAccountsMapsDashboardAccounts(t *testing.T) {
	phone := "13800138000"
	var sawToken bool

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/dashboard/accounts" {
			t.Fatalf("path mismatch: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "Bearer test-token" && r.Header.Get("X-Token") == "test-token" {
			sawToken = true
		}

		writeJSON(w, http.StatusOK, []upstreamDashboardAccountResponse{
			{
				AccountID:               1,
				ID:                      "120000001",
				Abbr:                    "A",
				PhoneNumber:             &phone,
				Nickname:                "active",
				CurrentWaveplate:        180,
				CurrentWaveplateCrystal: 60,
				DailyTask:               true,
				DailyTaskStatus:         "done",
			},
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
	if got[0].CurrentWaveplate != 180 || got[0].CurrentWaveplateCrystal != 60 {
		body, _ := json.Marshal(got[0])
		t.Fatalf("account energy mismatch: %s", body)
	}
	if !got[0].IsActive || !got[0].DailyTask || got[0].DailyTaskStatus != "done" {
		body, _ := json.Marshal(got[0])
		t.Fatalf("account daily state mismatch: %s", body)
	}
}

func TestHandleAccountByIDMapsActiveDashboardAccount(t *testing.T) {
	phone := "13800138000"
	var sawToken bool
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/dashboard/accounts" {
			http.NotFound(w, r)
			return
		}
		sawToken = r.Header.Get("Authorization") == "Bearer test-token" && r.Header.Get("X-Token") == "test-token"
		writeJSON(w, http.StatusOK, []upstreamDashboardAccountResponse{{
			AccountID:               3,
			ID:                      " 120000003 ",
			Abbr:                    " C ",
			PhoneNumber:             &phone,
			Nickname:                " Rover ",
			CurrentWaveplate:        100,
			CurrentWaveplateCrystal: 40,
			DailyTask:               true,
			DailyTaskStatus:         "skipped",
		}})
	}))
	defer upstream.Close()

	api := &API{cfg: config.Config{
		AccountServiceURL:            upstream.URL,
		AccountServiceTimeoutSeconds: 3,
	}}
	req := httptest.NewRequest(http.MethodGet, "/api/accounts/by-id/120000003", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	recorder := httptest.NewRecorder()

	api.handleAccountByID(recorder, req, authContext{})

	if recorder.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if !sawToken {
		t.Fatal("upstream did not receive token headers")
	}
	var got activeAccountResponse
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if got.AccountID != 3 || got.ID != "120000003" || got.Abbr != "C" || got.PhoneTail != "8000" || got.Nickname != "Rover" {
		body, _ := json.Marshal(got)
		t.Fatalf("account mapping mismatch: %s", body)
	}
	if !got.IsActive || got.CurrentWaveplate != 100 || got.CurrentWaveplateCrystal != 40 || !got.DailyTask || got.DailyTaskStatus != "skipped" {
		body, _ := json.Marshal(got)
		t.Fatalf("account state mismatch: %s", body)
	}
}

func TestHandleAccountByIDFallsBackForInactiveAccount(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/dashboard/accounts":
			writeJSON(w, http.StatusOK, []upstreamDashboardAccountResponse{})
		case "/api/accounts/by-id/120000004":
			writeJSON(w, http.StatusOK, upstreamAccountResponse{
				AccountID: 4,
				ID:        "120000004",
				Abbr:      "D",
				IsActive:  false,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	api := &API{cfg: config.Config{
		AccountServiceURL:            upstream.URL,
		AccountServiceTimeoutSeconds: 3,
	}}
	req := httptest.NewRequest(http.MethodGet, "/api/accounts/by-id/120000004", nil)
	recorder := httptest.NewRecorder()

	api.handleAccountByID(recorder, req, authContext{})

	if recorder.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var got activeAccountResponse
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if got.IsActive || got.ID != "120000004" || got.DailyTask {
		body, _ := json.Marshal(got)
		t.Fatalf("inactive account mismatch: %s", body)
	}
}

func TestHandleAccountByIDReturnsNotFound(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/dashboard/accounts" {
			writeJSON(w, http.StatusOK, []upstreamDashboardAccountResponse{})
			return
		}
		http.NotFound(w, r)
	}))
	defer upstream.Close()

	api := &API{cfg: config.Config{
		AccountServiceURL:            upstream.URL,
		AccountServiceTimeoutSeconds: 3,
	}}
	req := httptest.NewRequest(http.MethodGet, "/api/accounts/by-id/120000099", nil)
	recorder := httptest.NewRecorder()

	api.handleAccountByID(recorder, req, authContext{})

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status mismatch: got %d, want %d; body=%s", recorder.Code, http.StatusNotFound, recorder.Body.String())
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
