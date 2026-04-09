package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func queryPlayerIDs(ctx context.Context, database *sql.DB, table string) ([]string, error) {
	query := fmt.Sprintf("SELECT DISTINCT player_id FROM %s ORDER BY player_id", table)
	rows, err := database.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playerIDs []string
	for rows.Next() {
		var playerID string
		if err := rows.Scan(&playerID); err != nil {
			return nil, err
		}
		playerIDs = append(playerIDs, playerID)
	}
	return playerIDs, nil
}

func deleteByID(ctx context.Context, database *sql.DB, table string, id int64, auth authContext) (bool, *authError, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var createdByUserID sql.NullInt64
	err := database.QueryRowContext(ctx, fmt.Sprintf("SELECT created_by_user_id FROM %s WHERE id = $1", table), id).Scan(&createdByUserID)
	if err == sql.ErrNoRows {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}

	if !hasExactPermission(auth.Permissions, "manage") {
		if !createdByUserID.Valid || createdByUserID.Int64 != auth.UserID {
			return false, &authError{Status: http.StatusForbidden, Detail: authForbiddenDetail}, nil
		}
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", table)
	result, err := database.ExecContext(ctx, query, id)
	if err != nil {
		return false, nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, nil, err
	}

	return rowsAffected > 0, nil, nil
}

type filterBuilder struct {
	clauses []string
	args    []any
}

func buildCommonFilters(playerID, startDate, endDate string, solaLevel *int) (filterBuilder, error) {
	// Build parameterized SQL conditions shared by list and stats APIs.
	builder := filterBuilder{}

	if playerID != "" {
		builder.add("player_id = %s", playerID)
	}
	if startDate != "" {
		if _, err := time.Parse("2006-01-02", startDate); err != nil {
			return filterBuilder{}, fmt.Errorf("start_date 参数无效")
		}
		builder.add("date >= %s", startDate)
	}
	if endDate != "" {
		if _, err := time.Parse("2006-01-02", endDate); err != nil {
			return filterBuilder{}, fmt.Errorf("end_date 参数无效")
		}
		builder.add("date <= %s", endDate)
	}
	if solaLevel != nil && *solaLevel > 0 {
		builder.add("sola_level = %s", *solaLevel)
	}

	return builder, nil
}

func (b *filterBuilder) add(clause string, value any) {
	index := len(b.args) + 1
	b.clauses = append(b.clauses, fmt.Sprintf(clause, fmt.Sprintf("$%d", index)))
	b.args = append(b.args, value)
}

func (b filterBuilder) whereClause() string {
	if len(b.clauses) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(b.clauses, " AND ")
}

func parseIDFromPath(path, prefix string) (int64, error) {
	raw := strings.TrimPrefix(path, prefix)
	raw = strings.Trim(raw, "/")
	if raw == "" {
		return 0, fmt.Errorf("missing id")
	}
	return strconv.ParseInt(raw, 10, 64)
}

func parseOptionalInt(value string) (*int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseIntWithDefault(value string, fallback int) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("json encode failed: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, detail string) {
	writeJSON(w, status, map[string]string{"detail": detail})
}

func readJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("请求体无效")
	}
	return nil
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

func roundTo(value float64, digits int) float64 {
	pow := math.Pow(10, float64(digits))
	return math.Round(value*pow) / pow
}

func mapKeys[V any](items map[int]V) []int {
	keys := make([]int, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	return keys
}
