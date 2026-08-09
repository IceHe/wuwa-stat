package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (a *API) notifyDashboardEnergyRevision() error {
	endpoint, err := buildAccountServiceAPIURL(a.cfg.AccountServiceURL, "internal", "dashboard", "energy-revision")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	token := strings.TrimSpace(a.cfg.DashboardEnergyNotifyToken)
	if token != "" {
		req.Header.Set("X-Notify-Token", token)
	}

	resp, err := accountServiceHTTPClient(a).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	var payload struct {
		Detail string `json:"detail"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	if payload.Detail != "" {
		return fmt.Errorf("dashboard energy revision notify failed: status=%d detail=%s", resp.StatusCode, payload.Detail)
	}
	return fmt.Errorf("dashboard energy revision notify failed: status=%d", resp.StatusCode)
}
