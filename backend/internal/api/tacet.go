package api

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

func (a *API) handleAuthMe(w http.ResponseWriter, r *http.Request, auth authContext) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	token := extractToken(r)
	user, err := a.auth.lookupUser(r.Context(), token)
	if err != nil {
		writeError(w, err.Status, err.Detail)
		return
	}

	writeJSON(w, http.StatusOK, authMeResponse{
		UserID:      auth.UserID,
		Name:        user.Name,
		Permissions: auth.Permissions,
	})
}

func (a *API) handleTacetRecords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.withEdit(a.createTacetRecords)(w, r)
	case http.MethodGet:
		a.withView(a.getTacetRecords)(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (a *API) handleTacetRecordByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		methodNotAllowed(w)
		return
	}
	a.withEdit(a.deleteTacetRecord)(w, r)
}

func (a *API) createTacetRecords(w http.ResponseWriter, r *http.Request, auth authContext) {
	var payload tacetBatchCreate
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(payload.TacetRecords) == 0 {
		writeJSON(w, http.StatusOK, []tacetRecordResponse{})
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

	records := make([]tacetRecordResponse, 0, len(payload.TacetRecords))
	for _, item := range payload.TacetRecords {
		record, err := validateTacetRecord(item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		var created tacetRecordResponse
		err = tx.QueryRowContext(ctx, `
			INSERT INTO tacet_records (date, player_id, gold_tubes, purple_tubes, claim_count, sola_level, created_by_user_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, date::text, player_id, gold_tubes, purple_tubes, claim_count, sola_level, created_by_user_id, created_at
		`, record.Date, record.PlayerID, record.GoldTubes, record.PurpleTubes, record.ClaimCount, record.SolaLevel, auth.UserID).
			Scan(&created.ID, &created.Date, &created.PlayerID, &created.GoldTubes, &created.PurpleTubes, &created.ClaimCount, &created.SolaLevel, &created.CreatedByUserID, &created.CreatedAt)
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

func (a *API) getTacetRecords(w http.ResponseWriter, r *http.Request, _ authContext) {
	playerID := strings.TrimSpace(r.URL.Query().Get("player_id"))
	startDate := strings.TrimSpace(r.URL.Query().Get("start_date"))
	endDate := strings.TrimSpace(r.URL.Query().Get("end_date"))
	solaLevel, err := parseOptionalInt(r.URL.Query().Get("sola_level"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "sola_level 参数无效")
		return
	}
	skip, err := parseIntWithDefault(r.URL.Query().Get("skip"), 0)
	if err != nil || skip < 0 {
		writeError(w, http.StatusBadRequest, "skip 参数无效")
		return
	}
	limit, err := parseIntWithDefault(r.URL.Query().Get("limit"), 20)
	if err != nil || limit < 1 || limit > 1000 {
		writeError(w, http.StatusBadRequest, "limit 参数无效")
		return
	}

	builder, err := buildCommonFilters(playerID, startDate, endDate, solaLevel)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	totalQuery := "SELECT COUNT(*) FROM tacet_records" + builder.whereClause()
	var total int
	if err := a.db.QueryRowContext(ctx, totalQuery, builder.args...).Scan(&total); err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}

	dataQuery := "SELECT id, date::text, player_id, gold_tubes, purple_tubes, claim_count, sola_level, created_by_user_id, created_at FROM tacet_records" +
		builder.whereClause() +
		fmt.Sprintf(" ORDER BY created_at DESC, id DESC OFFSET $%d LIMIT $%d", len(builder.args)+1, len(builder.args)+2)
	args := append(append([]any{}, builder.args...), skip, limit)

	rows, err := a.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	defer rows.Close()

	var records []tacetRecordResponse
	for rows.Next() {
		var record tacetRecordResponse
		if err := rows.Scan(&record.ID, &record.Date, &record.PlayerID, &record.GoldTubes, &record.PurpleTubes, &record.ClaimCount, &record.SolaLevel, &record.CreatedByUserID, &record.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "数据库操作失败")
			return
		}
		records = append(records, record)
	}

	writeJSON(w, http.StatusOK, listResponse[tacetRecordResponse]{
		Data:        records,
		Total:       total,
		PageSize:    limit,
		CurrentPage: skip/limit + 1,
	})
}

func (a *API) handleTacetStats(w http.ResponseWriter, r *http.Request, _ authContext) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	playerID := strings.TrimSpace(r.URL.Query().Get("player_id"))
	startDate := strings.TrimSpace(r.URL.Query().Get("start_date"))
	endDate := strings.TrimSpace(r.URL.Query().Get("end_date"))

	builder, err := buildCommonFilters(playerID, startDate, endDate, nil)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	query := `
		SELECT
			COUNT(*) AS total_records,
			COALESCE(SUM(claim_count), 0) AS total_claim_count,
			COALESCE(SUM(gold_tubes), 0) AS total_gold_tubes,
			COALESCE(SUM(purple_tubes), 0) AS total_purple_tubes,
			COUNT(DISTINCT player_id) AS player_count
		FROM tacet_records` + builder.whereClause()

	var resp statsResponse
	if err := a.db.QueryRowContext(ctx, query, builder.args...).Scan(&resp.TotalRecords, &resp.TotalClaimCount, &resp.TotalGoldTubes, &resp.TotalPurpleTubes, &resp.PlayerCount); err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}

	if resp.TotalClaimCount > 0 {
		resp.AvgGoldTubes = float64(resp.TotalGoldTubes) / float64(resp.TotalClaimCount)
		resp.AvgPurpleTubes = float64(resp.TotalPurpleTubes) / float64(resp.TotalClaimCount)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (a *API) handleTacetDetailedStats(w http.ResponseWriter, r *http.Request, _ authContext) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	playerID := strings.TrimSpace(r.URL.Query().Get("player_id"))
	startDate := strings.TrimSpace(r.URL.Query().Get("start_date"))
	endDate := strings.TrimSpace(r.URL.Query().Get("end_date"))

	builder, err := buildCommonFilters(playerID, startDate, endDate, nil)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	query := `
		SELECT sola_level, claim_count, gold_tubes, purple_tubes, COUNT(*) AS count
		FROM tacet_records` + builder.whereClause() + `
		GROUP BY sola_level, claim_count, gold_tubes, purple_tubes`
	rows, err := a.db.QueryContext(ctx, query, builder.args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	defer rows.Close()

	levelData := map[int]map[[2]int]int{}
	for rows.Next() {
		var solaLevel, claimCount, goldTubes, purpleTubes, count int
		if err := rows.Scan(&solaLevel, &claimCount, &goldTubes, &purpleTubes, &count); err != nil {
			writeError(w, http.StatusInternalServerError, "数据库操作失败")
			return
		}

		if _, ok := levelData[solaLevel]; !ok {
			levelData[solaLevel] = map[[2]int]int{}
		}

		for _, combo := range splitTacetCombination(solaLevel, goldTubes, purpleTubes, claimCount) {
			key := [2]int{combo[0], combo[1]}
			levelData[solaLevel][key] += count
		}
	}

	levels := mapKeys(levelData)
	sort.Sort(sort.Reverse(sort.IntSlice(levels)))

	response := detailedStatsResponse{LevelStats: make([]solaLevelStats, 0, len(levels))}
	for _, level := range levels {
		combinationData := levelData[level]
		totalCount := 0
		totalExperience := 0
		for combo, count := range combinationData {
			totalCount += count
			totalExperience += (combo[0]*5000 + combo[1]*2000) * count
		}

		type comboEntry struct {
			Gold   int
			Purple int
			Count  int
		}

		entries := make([]comboEntry, 0, len(combinationData))
		for combo, count := range combinationData {
			entries = append(entries, comboEntry{Gold: combo[0], Purple: combo[1], Count: count})
		}

		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Gold != entries[j].Gold {
				return entries[i].Gold > entries[j].Gold
			}
			return entries[i].Purple > entries[j].Purple
		})

		combinations := make([]dropCombination, 0, len(entries))
		for _, entry := range entries {
			percentage := 0.0
			if totalCount > 0 {
				percentage = roundTo(float64(entry.Count)/float64(totalCount)*100, 1)
			}
			combinations = append(combinations, dropCombination{
				ClaimCount:  1,
				GoldTubes:   entry.Gold,
				PurpleTubes: entry.Purple,
				Experience:  entry.Gold*5000 + entry.Purple*2000,
				Count:       entry.Count,
				Percentage:  percentage,
			})
		}

		avgExperience := 0.0
		if totalCount > 0 {
			avgExperience = roundTo(float64(totalExperience)/float64(totalCount), 0)
		}

		response.LevelStats = append(response.LevelStats, solaLevelStats{
			SolaLevel:     level,
			Combinations:  combinations,
			TotalCount:    totalCount,
			AvgExperience: avgExperience,
		})
	}

	writeJSON(w, http.StatusOK, response)
}

func (a *API) handleTacetPlayerIDs(w http.ResponseWriter, r *http.Request, _ authContext) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	playerIDs, err := queryPlayerIDs(r.Context(), a.db, "tacet_records")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	writeJSON(w, http.StatusOK, playerIDs)
}

func (a *API) deleteTacetRecord(w http.ResponseWriter, r *http.Request, auth authContext) {
	recordID, err := parseIDFromPath(r.URL.Path, "/api/tacet_records/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "记录 ID 无效")
		return
	}

	deleted, authErr, err := deleteByID(r.Context(), a.db, "tacet_records", recordID, auth)
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
