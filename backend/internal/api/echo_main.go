package api

import (
	"context"
	"database/sql"
	"net/http"
	"sort"
	"time"
)

func (a *API) handleEchoMainRecords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.withEdit(a.createEchoMainRecords)(w, r)
	case http.MethodGet:
		a.withView(a.getEchoMainRecords)(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (a *API) handleEchoMainRecordByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		methodNotAllowed(w)
		return
	}
	a.withEdit(a.deleteEchoMainRecord)(w, r)
}

func (a *API) createEchoMainRecords(w http.ResponseWriter, r *http.Request, auth authContext) {
	var payload echoMainBatchCreate
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(payload.Records) == 0 {
		writeJSON(w, http.StatusOK, []echoMainRecordResponse{})
		return
	}

	validated := make([]echoMainRecordInput, 0, len(payload.Records))
	energyCosts := map[string]int{}
	for _, item := range payload.Records {
		record, err := validateEchoMainRecord(item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		validated = append(validated, record)
		addEnergyDeduction(energyCosts, record.PlayerID, 60)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	defer tx.Rollback()

	records := make([]echoMainRecordResponse, 0, len(validated))
	for _, record := range validated {
		var created echoMainRecordResponse
		err = tx.QueryRowContext(ctx, `
			INSERT INTO echo_main_records (date, player_id, sola_level, c3_main_stat, c1_main_stat, tacet_domain, echo_set, created_by_user_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id, date::text, player_id, sola_level, c3_main_stat, c1_main_stat, tacet_domain, echo_set, created_by_user_id, created_at
		`, record.Date, record.PlayerID, record.SolaLevel, record.C3MainStat, record.C1MainStat, record.TacetDomain, record.EchoSet, auth.UserID).
			Scan(&created.ID, &created.Date, &created.PlayerID, &created.SolaLevel, &created.C3MainStat, &created.C1MainStat, &created.TacetDomain, &created.EchoSet, &created.CreatedByUserID, &created.CreatedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "数据库操作失败")
			return
		}
		records = append(records, created)
	}

	if !payload.SkipEnergyDeduction {
		if err := a.deductEnergyForPlayers(ctx, extractToken(r), energyCosts); err != nil {
			if writeEnergyDeductionError(w, err) {
				return
			}
			writeError(w, http.StatusServiceUnavailable, "账号服务不可用")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (a *API) getEchoMainRecords(w http.ResponseWriter, r *http.Request, _ authContext) {
	params, err := buildListQueryParams(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := queryListRecords(r.Context(), a.db, "echo_main_records", "id, date::text, player_id, sola_level, c3_main_stat, c1_main_stat, tacet_domain, echo_set, created_by_user_id, created_at", params, func(rows *sql.Rows) (echoMainRecordResponse, error) {
		var record echoMainRecordResponse
		err := rows.Scan(&record.ID, &record.Date, &record.PlayerID, &record.SolaLevel, &record.C3MainStat, &record.C1MainStat, &record.TacetDomain, &record.EchoSet, &record.CreatedByUserID, &record.CreatedAt)
		return record, err
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *API) handleEchoMainStats(w http.ResponseWriter, r *http.Request, _ authContext) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	builder, err := buildFilterBuilderFromRequest(r, true)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	query := `SELECT date::text, player_id, sola_level, c3_main_stat, c1_main_stat, tacet_domain, echo_set, COUNT(*)
		FROM echo_main_records` + builder.whereClause() + `
		GROUP BY date, player_id, sola_level, c3_main_stat, c1_main_stat, tacet_domain, echo_set`
	rows, err := a.db.QueryContext(ctx, query, builder.args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	defer rows.Close()

	details := make([]echoMainDetailStat, 0)
	type summaryKey struct{ dimension, value string }
	summaryCounts := map[summaryKey]int{}
	total := 0
	for rows.Next() {
		var item echoMainDetailStat
		if err := rows.Scan(&item.Date, &item.PlayerID, &item.SolaLevel, &item.C3MainStat, &item.C1MainStat, &item.TacetDomain, &item.EchoSet, &item.Count); err != nil {
			writeError(w, http.StatusInternalServerError, "数据库操作失败")
			return
		}
		details = append(details, item)
		total += item.Count
		summaryCounts[summaryKey{"C3", item.C3MainStat}] += item.Count
		summaryCounts[summaryKey{"C1", item.C1MainStat}] += item.Count
		summaryCounts[summaryKey{"无音区", item.TacetDomain}] += item.Count
		summaryCounts[summaryKey{"声骸套装", item.EchoSet}] += item.Count
	}
	sort.Slice(details, func(i, j int) bool {
		if details[i].Date != details[j].Date {
			return details[i].Date > details[j].Date
		}
		if details[i].PlayerID != details[j].PlayerID {
			return details[i].PlayerID < details[j].PlayerID
		}
		if details[i].SolaLevel != details[j].SolaLevel {
			return details[i].SolaLevel > details[j].SolaLevel
		}
		return details[i].Count > details[j].Count
	})

	summary := make([]echoMainSummaryStat, 0, len(summaryCounts))
	for key, count := range summaryCounts {
		percentage := 0.0
		if total > 0 {
			percentage = roundTo(float64(count)/float64(total)*100, 1)
		}
		summary = append(summary, echoMainSummaryStat{Dimension: key.dimension, Value: key.value, Count: count, Percentage: percentage})
	}
	sort.Slice(summary, func(i, j int) bool {
		if summary[i].Dimension != summary[j].Dimension {
			return summary[i].Dimension < summary[j].Dimension
		}
		return summary[i].Count > summary[j].Count
	})
	writeJSON(w, http.StatusOK, echoMainStatsResponse{Details: details, Summary: summary})
}

func (a *API) handleEchoMainPlayerIDs(w http.ResponseWriter, r *http.Request, _ authContext) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	playerIDs, err := queryPlayerIDs(r.Context(), a.db, "echo_main_records")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	writeJSON(w, http.StatusOK, playerIDs)
}

func (a *API) deleteEchoMainRecord(w http.ResponseWriter, r *http.Request, auth authContext) {
	recordID, err := parseIDFromPath(r.URL.Path, "/api/echo-main-records/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "记录 ID 无效")
		return
	}
	deleted, authErr, err := a.deleteRecordWithEnergyRefund(
		r.Context(), a.db, "echo_main_records", recordID, auth, extractToken(r),
		`SELECT player_id, 1, created_by_user_id, created_at FROM echo_main_records WHERE id = $1 FOR UPDATE`,
		func(row *sql.Row) (deleteEnergyRecord, error) {
			var record deleteEnergyRecord
			err := row.Scan(&record.PlayerID, &record.ClaimCount, &record.CreatedByUserID, &record.CreatedAt)
			return record, err
		},
		func(record deleteEnergyRecord) int { return recentRecordEnergyCost(record, 60) },
	)
	if authErr != nil {
		writeError(w, authErr.Status, authErr.Detail)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	if !deleted {
		writeError(w, http.StatusNotFound, "记录不存在")
		return
	}
	writeJSON(w, http.StatusOK, messageResponse{Message: "删除成功"})
}
