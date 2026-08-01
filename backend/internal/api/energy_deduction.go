package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const energyInsufficientDetail = "体力不足以扣除，是否只记录不扣体力？"

var errEnergyInsufficient = errors.New("energy insufficient")

type energyDeductionTarget struct {
	PlayerID string
	Cost     int
}

type deleteEnergyRecord struct {
	PlayerID        string
	ClaimCount      int
	CreatedByUserID sql.NullInt64
	CreatedAt       time.Time
}

func (a *API) deductEnergyForPlayers(ctx context.Context, token string, costs map[string]int) error {
	targets := make([]energyDeductionTarget, 0, len(costs))
	for playerID, cost := range costs {
		playerID = strings.TrimSpace(playerID)
		if playerID == "" || cost <= 0 {
			continue
		}

		account, found, err := a.fetchAccountByID(ctx, token, playerID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}

		if account.CurrentWaveplate+account.CurrentWaveplateCrystal < cost {
			return errEnergyInsufficient
		}

		targets = append(targets, energyDeductionTarget{
			PlayerID: playerID,
			Cost:     cost,
		})
	}

	for _, target := range targets {
		for _, cost := range splitEnergySpendCost(target.Cost) {
			if err := a.spendAccountEnergy(ctx, token, target.PlayerID, cost); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *API) fetchAccountByID(ctx context.Context, token string, playerID string) (upstreamAccountResponse, bool, error) {
	endpoint, err := buildAccountServiceAPIURL(a.cfg.AccountServiceURL, "accounts", "by-id", playerID)
	if err != nil {
		return upstreamAccountResponse{}, false, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return upstreamAccountResponse{}, false, err
	}
	setAccountServiceHeaders(req, token)

	resp, err := accountServiceHTTPClient(a).Do(req)
	if err != nil {
		return upstreamAccountResponse{}, false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return upstreamAccountResponse{}, false, nil
	case http.StatusUnauthorized:
		return upstreamAccountResponse{}, false, &authError{Status: http.StatusUnauthorized, Detail: authInvalidDetail}
	case http.StatusForbidden:
		return upstreamAccountResponse{}, false, &authError{Status: http.StatusForbidden, Detail: authForbiddenDetail}
	default:
		return upstreamAccountResponse{}, false, fmt.Errorf("account service unexpected status: %d", resp.StatusCode)
	}

	var account upstreamAccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return upstreamAccountResponse{}, false, err
	}
	return account, true, nil
}

func (a *API) spendAccountEnergy(ctx context.Context, token string, playerID string, cost int) error {
	endpoint, err := buildAccountServiceAPIURL(a.cfg.AccountServiceURL, "accounts", "by-id", playerID, "energy", "spend")
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]int{"cost": cost})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	setAccountServiceHeaders(req, token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := accountServiceHTTPClient(a).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	case http.StatusUnauthorized:
		return &authError{Status: http.StatusUnauthorized, Detail: authInvalidDetail}
	case http.StatusForbidden:
		return &authError{Status: http.StatusForbidden, Detail: authForbiddenDetail}
	case http.StatusBadRequest:
		var payload struct {
			Detail string `json:"detail"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&payload)
		if strings.Contains(strings.ToLower(payload.Detail), "not enough waveplate") {
			return errEnergyInsufficient
		}
		return fmt.Errorf("account service bad request: %s", payload.Detail)
	default:
		return fmt.Errorf("account service unexpected status: %d", resp.StatusCode)
	}
}

func (a *API) refundEnergyForPlayers(ctx context.Context, token string, costs map[string]int) error {
	for playerID, cost := range costs {
		playerID = strings.TrimSpace(playerID)
		if playerID == "" || cost <= 0 {
			continue
		}

		_, found, err := a.fetchAccountByID(ctx, token, playerID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}

		for _, amount := range splitEnergyGainAmount(cost) {
			if err := a.gainAccountEnergy(ctx, token, playerID, amount); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *API) deleteRecordWithEnergyRefund(
	ctx context.Context,
	database *sql.DB,
	table string,
	id int64,
	auth authContext,
	token string,
	selectQuery string,
	scan func(*sql.Row) (deleteEnergyRecord, error),
	refundCost func(deleteEnergyRecord) int,
) (bool, *authError, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return false, nil, err
	}
	defer tx.Rollback()

	record, err := scan(tx.QueryRowContext(ctx, selectQuery, id))
	if err == sql.ErrNoRows {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}

	if !hasExactPermission(auth.Permissions, "manage") {
		if !record.CreatedByUserID.Valid || record.CreatedByUserID.Int64 != auth.UserID {
			return false, &authError{Status: http.StatusForbidden, Detail: authForbiddenDetail}, nil
		}
	}

	result, err := tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = $1", table), id)
	if err != nil {
		return false, nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, nil, err
	}
	if rowsAffected == 0 {
		return false, nil, nil
	}

	if refund := refundCost(record); refund > 0 {
		if err := a.refundEnergyForPlayers(ctx, token, map[string]int{record.PlayerID: refund}); err != nil {
			if authErr := asAuthError(err); authErr != nil {
				return false, authErr, nil
			}
			return false, nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return false, nil, err
	}

	return true, nil, nil
}

func (a *API) gainAccountEnergy(ctx context.Context, token string, playerID string, amount int) error {
	endpoint, err := buildAccountServiceAPIURL(a.cfg.AccountServiceURL, "accounts", "by-id", playerID, "energy", "gain")
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]int{"amount": amount})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	setAccountServiceHeaders(req, token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := accountServiceHTTPClient(a).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	case http.StatusUnauthorized:
		return &authError{Status: http.StatusUnauthorized, Detail: authInvalidDetail}
	case http.StatusForbidden:
		return &authError{Status: http.StatusForbidden, Detail: authForbiddenDetail}
	case http.StatusBadRequest:
		var payload struct {
			Detail string `json:"detail"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&payload)
		return fmt.Errorf("account service bad request: %s", payload.Detail)
	default:
		return fmt.Errorf("account service unexpected status: %d", resp.StatusCode)
	}
}

func splitEnergySpendCost(total int) []int {
	costs := make([]int, 0)
	for total >= 120 {
		costs = append(costs, 120)
		total -= 120
	}

	switch total {
	case 80, 60, 40:
		costs = append(costs, total)
	case 0:
	default:
		costs = append(costs, total)
	}

	return costs
}

func splitEnergyGainAmount(total int) []int {
	amounts := make([]int, 0)
	if total%60 == 0 {
		for total > 0 {
			amounts = append(amounts, 60)
			total -= 60
		}
		return amounts
	}

	if total%40 == 0 {
		for total > 0 {
			amounts = append(amounts, 40)
			total -= 40
		}
		return amounts
	}

	if total > 0 {
		amounts = append(amounts, total)
	}
	return amounts
}

func addEnergyDeduction(costs map[string]int, playerID string, cost int) {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" || cost <= 0 {
		return
	}
	costs[playerID] += cost
}

func recentRecordEnergyCost(record deleteEnergyRecord, costPerClaim int) int {
	if costPerClaim <= 0 || record.ClaimCount <= 0 {
		return 0
	}
	if time.Since(record.CreatedAt) > time.Hour {
		return 0
	}
	return record.ClaimCount * costPerClaim
}

func writeEnergyDeductionError(w http.ResponseWriter, err error) bool {
	if errors.Is(err, errEnergyInsufficient) {
		writeJSON(w, http.StatusConflict, map[string]string{
			"code":   "energy_insufficient",
			"detail": energyInsufficientDetail,
		})
		return true
	}

	if authErr := asAuthError(err); authErr != nil {
		writeError(w, authErr.Status, authErr.Detail)
		return true
	}

	return false
}
