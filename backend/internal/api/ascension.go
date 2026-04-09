package api

import (
	"context"
	"database/sql"
	"net/http"
	"sort"
	"time"
)

func (a *API) handleAscensionRecords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.withEdit(a.createAscensionRecords)(w, r)
	case http.MethodGet:
		a.withView(a.getAscensionRecords)(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (a *API) handleAscensionRecordByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		methodNotAllowed(w)
		return
	}
	a.withEdit(a.deleteAscensionRecord)(w, r)
}

func (a *API) createAscensionRecords(w http.ResponseWriter, r *http.Request, auth authContext) {
	var payload ascensionBatchCreate
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(payload.AscensionRecords) == 0 {
		writeJSON(w, http.StatusOK, []ascensionRecordResponse{})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	defer tx.Rollback()

	records := make([]ascensionRecordResponse, 0, len(payload.AscensionRecords))
	for _, item := range payload.AscensionRecords {
		record, err := validateAscensionRecord(item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		var created ascensionRecordResponse
		err = tx.QueryRowContext(ctx, `
			INSERT INTO ascension_records (date, player_id, sola_level, drop_count, created_by_user_id)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, date::text, player_id, sola_level, drop_count, created_by_user_id, created_at
		`, record.Date, record.PlayerID, record.SolaLevel, record.DropCount, auth.UserID).
			Scan(&created.ID, &created.Date, &created.PlayerID, &created.SolaLevel, &created.DropCount, &created.CreatedByUserID, &created.CreatedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "数据库操作失败")
			return
		}
		records = append(records, created)
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}

	writeJSON(w, http.StatusOK, records)
}

func (a *API) getAscensionRecords(w http.ResponseWriter, r *http.Request, _ authContext) {
	params, err := buildListQueryParams(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := queryListRecords(r.Context(), a.db, "ascension_records", "id, date::text, player_id, sola_level, drop_count, created_by_user_id, created_at", params, func(rows *sql.Rows) (ascensionRecordResponse, error) {
		var record ascensionRecordResponse
		err := rows.Scan(&record.ID, &record.Date, &record.PlayerID, &record.SolaLevel, &record.DropCount, &record.CreatedByUserID, &record.CreatedAt)
		return record, err
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (a *API) handleAscensionDetailedStats(w http.ResponseWriter, r *http.Request, _ authContext) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	builder, err := buildFilterBuilderFromRequest(r, false)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	query := `
		SELECT sola_level, drop_count, COUNT(*) AS count
		FROM ascension_records` + builder.whereClause() + `
		GROUP BY sola_level, drop_count`

	rows, err := a.db.QueryContext(ctx, query, builder.args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	defer rows.Close()

	type entry struct {
		DropCount int
		Count     int
	}

	levelData := map[int][]entry{}
	for rows.Next() {
		var solaLevel, dropCount, count int
		if err := rows.Scan(&solaLevel, &dropCount, &count); err != nil {
			writeError(w, http.StatusInternalServerError, "数据库操作失败")
			return
		}
		levelData[solaLevel] = append(levelData[solaLevel], entry{DropCount: dropCount, Count: count})
	}

	levels := mapKeys(levelData)
	sort.Sort(sort.Reverse(sort.IntSlice(levels)))

	response := ascensionDetailedStatsResponse{LevelStats: make([]ascensionSolaLevelStats, 0, len(levels))}
	for _, level := range levels {
		entries := levelData[level]
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].DropCount > entries[j].DropCount
		})

		totalCount := 0
		totalDropCount := 0
		combinations := make([]ascensionDropCombination, 0, len(entries))
		for _, item := range entries {
			totalCount += item.Count
			totalDropCount += item.DropCount * item.Count
		}

		for _, item := range entries {
			percentage := 0.0
			if totalCount > 0 {
				percentage = roundTo(float64(item.Count)/float64(totalCount)*100, 1)
			}
			combinations = append(combinations, ascensionDropCombination{
				DropCount:  item.DropCount,
				Count:      item.Count,
				Percentage: percentage,
			})
		}

		avgDropCount := 0.0
		if totalCount > 0 {
			avgDropCount = roundTo(float64(totalDropCount)/float64(totalCount), 2)
		}

		response.LevelStats = append(response.LevelStats, ascensionSolaLevelStats{
			SolaLevel:    level,
			Combinations: combinations,
			TotalCount:   totalCount,
			AvgDropCount: avgDropCount,
		})
	}

	writeJSON(w, http.StatusOK, response)
}

func (a *API) handleAscensionPlayerIDs(w http.ResponseWriter, r *http.Request, _ authContext) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	playerIDs, err := queryPlayerIDs(r.Context(), a.db, "ascension_records")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	writeJSON(w, http.StatusOK, playerIDs)
}

func (a *API) deleteAscensionRecord(w http.ResponseWriter, r *http.Request, auth authContext) {
	recordID, err := parseIDFromPath(r.URL.Path, "/api/ascension-records/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "记录 ID 无效")
		return
	}

	deleted, authErr, err := deleteByID(r.Context(), a.db, "ascension_records", recordID, auth)
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
