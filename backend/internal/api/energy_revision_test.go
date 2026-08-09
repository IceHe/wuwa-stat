package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"

	"wuwa/stat/backend/internal/config"
)

func TestDeductEnergyForPlayersNotifiesDashboardRevisionAfterSuccessfulSpend(t *testing.T) {
	var spendCosts []int
	var notifyCalls int32
	var notifyToken string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/accounts/by-id/120000001":
			assertAccountServiceToken(t, r, "test-token")
			writeJSON(w, http.StatusOK, upstreamAccountResponse{
				AccountID:               1,
				ID:                      "120000001",
				IsActive:                true,
				CurrentWaveplate:        180,
				CurrentWaveplateCrystal: 0,
			})
		case "/api/accounts/by-id/120000001/energy/spend":
			assertAccountServiceToken(t, r, "test-token")
			var payload struct {
				Cost int `json:"cost"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("invalid spend payload: %v", err)
			}
			spendCosts = append(spendCosts, payload.Cost)
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		case "/api/internal/dashboard/energy-revision":
			if r.Method != http.MethodPost {
				t.Fatalf("notify method mismatch: %s", r.Method)
			}
			notifyToken = r.Header.Get("X-Notify-Token")
			atomic.AddInt32(&notifyCalls, 1)
			writeJSON(w, http.StatusOK, map[string]uint64{"revision": 1})
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	api := &API{
		cfg: config.Config{
			AccountServiceURL:            upstream.URL,
			AccountServiceTimeoutSeconds: 3,
			DashboardEnergyNotifyToken:   "secret",
		},
	}

	err := api.deductEnergyForPlayers(context.Background(), "test-token", map[string]int{"120000001": 180})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantSpendCosts := []int{120, 60}
	if !reflect.DeepEqual(spendCosts, wantSpendCosts) {
		t.Fatalf("spend costs mismatch: got %v, want %v", spendCosts, wantSpendCosts)
	}
	if got := atomic.LoadInt32(&notifyCalls); got != 1 {
		t.Fatalf("notify calls = %d, want 1", got)
	}
	if notifyToken != "secret" {
		t.Fatalf("notify token = %q, want %q", notifyToken, "secret")
	}
}

func TestRefundEnergyForPlayersNotifiesDashboardRevisionAfterSuccessfulGain(t *testing.T) {
	var gainAmounts []int
	var notifyCalls int32
	var notifyToken string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/accounts/by-id/120000001":
			assertAccountServiceToken(t, r, "test-token")
			writeJSON(w, http.StatusOK, upstreamAccountResponse{
				AccountID: 1,
				ID:        "120000001",
				IsActive:  true,
			})
		case "/api/accounts/by-id/120000001/energy/gain":
			assertAccountServiceToken(t, r, "test-token")
			var payload struct {
				Amount int `json:"amount"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("invalid gain payload: %v", err)
			}
			gainAmounts = append(gainAmounts, payload.Amount)
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		case "/api/internal/dashboard/energy-revision":
			if r.Method != http.MethodPost {
				t.Fatalf("notify method mismatch: %s", r.Method)
			}
			notifyToken = r.Header.Get("X-Notify-Token")
			atomic.AddInt32(&notifyCalls, 1)
			writeJSON(w, http.StatusOK, map[string]uint64{"revision": 1})
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	api := &API{
		cfg: config.Config{
			AccountServiceURL:            upstream.URL,
			AccountServiceTimeoutSeconds: 3,
			DashboardEnergyNotifyToken:   "secret",
		},
	}

	err := api.refundEnergyForPlayers(context.Background(), "test-token", map[string]int{"120000001": 80})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantGainAmounts := []int{40, 40}
	if !reflect.DeepEqual(gainAmounts, wantGainAmounts) {
		t.Fatalf("gain amounts mismatch: got %v, want %v", gainAmounts, wantGainAmounts)
	}
	if got := atomic.LoadInt32(&notifyCalls); got != 1 {
		t.Fatalf("notify calls = %d, want 1", got)
	}
	if notifyToken != "secret" {
		t.Fatalf("notify token = %q, want %q", notifyToken, "secret")
	}
}

func TestDeductEnergyForPlayersSkipsNotifyWhenNoEnergyChange(t *testing.T) {
	t.Run("insufficient before spend", func(t *testing.T) {
		var notifyCalls int32

		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/accounts/by-id/120000001":
				assertAccountServiceToken(t, r, "test-token")
				writeJSON(w, http.StatusOK, upstreamAccountResponse{
					AccountID:               1,
					ID:                      "120000001",
					IsActive:                true,
					CurrentWaveplate:        40,
					CurrentWaveplateCrystal: 0,
				})
			case "/api/accounts/by-id/120000001/energy/spend":
				t.Fatalf("spend should not be called when precheck is insufficient")
			case "/api/internal/dashboard/energy-revision":
				atomic.AddInt32(&notifyCalls, 1)
				t.Fatalf("notify should not be called when precheck is insufficient")
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
		if got := atomic.LoadInt32(&notifyCalls); got != 0 {
			t.Fatalf("notify calls = %d, want 0", got)
		}
	})

	t.Run("account lookup fails", func(t *testing.T) {
		var notifyCalls int32

		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/accounts/by-id/120000001":
				assertAccountServiceToken(t, r, "test-token")
				writeJSON(w, http.StatusInternalServerError, map[string]string{"detail": "boom"})
			case "/api/accounts/by-id/120000001/energy/spend":
				t.Fatalf("spend should not be called when account lookup fails")
			case "/api/internal/dashboard/energy-revision":
				atomic.AddInt32(&notifyCalls, 1)
				t.Fatalf("notify should not be called when account lookup fails")
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

		if err := api.deductEnergyForPlayers(context.Background(), "test-token", map[string]int{"120000001": 60}); err == nil {
			t.Fatalf("expected error from account lookup failure")
		}
		if got := atomic.LoadInt32(&notifyCalls); got != 0 {
			t.Fatalf("notify calls = %d, want 0", got)
		}
	})
}

func TestDeductEnergyForPlayersIgnoresNotifyFailure(t *testing.T) {
	var spendCosts []int
	var notifyCalls int32

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/accounts/by-id/120000001":
			assertAccountServiceToken(t, r, "test-token")
			writeJSON(w, http.StatusOK, upstreamAccountResponse{
				AccountID:               1,
				ID:                      "120000001",
				IsActive:                true,
				CurrentWaveplate:        180,
				CurrentWaveplateCrystal: 0,
			})
		case "/api/accounts/by-id/120000001/energy/spend":
			assertAccountServiceToken(t, r, "test-token")
			var payload struct {
				Cost int `json:"cost"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("invalid spend payload: %v", err)
			}
			spendCosts = append(spendCosts, payload.Cost)
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		case "/api/internal/dashboard/energy-revision":
			atomic.AddInt32(&notifyCalls, 1)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"detail": "notify failed"})
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantSpendCosts := []int{60}
	if !reflect.DeepEqual(spendCosts, wantSpendCosts) {
		t.Fatalf("spend costs mismatch: got %v, want %v", spendCosts, wantSpendCosts)
	}
	if got := atomic.LoadInt32(&notifyCalls); got != 1 {
		t.Fatalf("notify calls = %d, want 1", got)
	}
}

func assertAccountServiceToken(t *testing.T, r *http.Request, token string) {
	t.Helper()

	if got := r.Header.Get("Authorization"); got != "Bearer "+token {
		t.Fatalf("authorization header mismatch: got %q, want %q", got, "Bearer "+token)
	}
	if got := r.Header.Get("X-Token"); got != token {
		t.Fatalf("x-token header mismatch: got %q, want %q", got, token)
	}
}
