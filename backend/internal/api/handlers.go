package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"wuwa/stat/backend/internal/config"
)

type API struct {
	db   *sql.DB
	cfg  config.Config
	auth *authValidator
}

func New(database *sql.DB, cfg config.Config) *API {
	return &API{
		db:   database,
		cfg:  cfg,
		auth: newAuthValidator(cfg),
	}
}

func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleRoot)
	mux.HandleFunc("/health", a.handleHealth)
	mux.HandleFunc("/api/auth/me", a.withView(a.handleAuthMe))
	mux.HandleFunc("/api/tacet_records", a.handleTacetRecords)
	mux.HandleFunc("/api/tacet_records/", a.handleTacetRecordByID)
	mux.HandleFunc("/api/stats", a.withView(a.handleTacetStats))
	mux.HandleFunc("/api/detailed-stats", a.withView(a.handleTacetDetailedStats))
	mux.HandleFunc("/api/player-ids", a.withView(a.handleTacetPlayerIDs))
	mux.HandleFunc("/api/ascension-records", a.handleAscensionRecords)
	mux.HandleFunc("/api/ascension-records/", a.handleAscensionRecordByID)
	mux.HandleFunc("/api/ascension-detailed-stats", a.withView(a.handleAscensionDetailedStats))
	mux.HandleFunc("/api/ascension-player-ids", a.withView(a.handleAscensionPlayerIDs))
	mux.HandleFunc("/api/resonance-records", a.handleResonanceRecords)
	mux.HandleFunc("/api/resonance-records/", a.handleResonanceRecordByID)
	mux.HandleFunc("/api/resonance-detailed-stats", a.withView(a.handleResonanceDetailedStats))
	mux.HandleFunc("/api/resonance-player-ids", a.withView(a.handleResonancePlayerIDs))
	return a.withCORS(mux)
}

func (a *API) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == a.cfg.FrontendURL {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Token")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *API) withView(next func(http.ResponseWriter, *http.Request, []string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		permissions, err := a.auth.requireView(r)
		if err != nil {
			writeError(w, err.Status, err.Detail)
			return
		}
		next(w, r, permissions)
	}
}

func (a *API) withEdit(next func(http.ResponseWriter, *http.Request, []string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		permissions, err := a.auth.requireEdit(r)
		if err != nil {
			writeError(w, err.Status, err.Detail)
			return
		}
		next(w, r, permissions)
	}
}

func (a *API) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" || r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, messageResponse{Message: "鸣潮产出统计 API"})
}

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/health" || r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) handleAuthMe(w http.ResponseWriter, r *http.Request, permissions []string) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, authMeResponse{Permissions: permissions})
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

func (a *API) createTacetRecords(w http.ResponseWriter, r *http.Request, _ []string) {
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
			INSERT INTO tacet_records (date, player_id, gold_tubes, purple_tubes, claim_count, sola_level)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, date::text, player_id, gold_tubes, purple_tubes, claim_count, sola_level, created_at
		`, record.Date, record.PlayerID, record.GoldTubes, record.PurpleTubes, record.ClaimCount, record.SolaLevel).
			Scan(&created.ID, &created.Date, &created.PlayerID, &created.GoldTubes, &created.PurpleTubes, &created.ClaimCount, &created.SolaLevel, &created.CreatedAt)
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

func (a *API) getTacetRecords(w http.ResponseWriter, r *http.Request, _ []string) {
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

	dataQuery := "SELECT id, date::text, player_id, gold_tubes, purple_tubes, claim_count, sola_level, created_at FROM tacet_records" +
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
		if err := rows.Scan(&record.ID, &record.Date, &record.PlayerID, &record.GoldTubes, &record.PurpleTubes, &record.ClaimCount, &record.SolaLevel, &record.CreatedAt); err != nil {
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

func (a *API) handleTacetStats(w http.ResponseWriter, r *http.Request, _ []string) {
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

func (a *API) handleTacetDetailedStats(w http.ResponseWriter, r *http.Request, _ []string) {
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

func (a *API) handleTacetPlayerIDs(w http.ResponseWriter, r *http.Request, _ []string) {
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

func (a *API) deleteTacetRecord(w http.ResponseWriter, r *http.Request, _ []string) {
	recordID, err := parseIDFromPath(r.URL.Path, "/api/tacet_records/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "记录 ID 无效")
		return
	}

	deleted, err := deleteByID(r.Context(), a.db, "tacet_records", recordID)
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

func (a *API) createAscensionRecords(w http.ResponseWriter, r *http.Request, _ []string) {
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
			INSERT INTO ascension_records (date, player_id, sola_level, drop_count)
			VALUES ($1, $2, $3, $4)
			RETURNING id, date::text, player_id, sola_level, drop_count, created_at
		`, record.Date, record.PlayerID, record.SolaLevel, record.DropCount).
			Scan(&created.ID, &created.Date, &created.PlayerID, &created.SolaLevel, &created.DropCount, &created.CreatedAt)
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

func (a *API) getAscensionRecords(w http.ResponseWriter, r *http.Request, _ []string) {
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

	totalQuery := "SELECT COUNT(*) FROM ascension_records" + builder.whereClause()
	var total int
	if err := a.db.QueryRowContext(ctx, totalQuery, builder.args...).Scan(&total); err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}

	dataQuery := "SELECT id, date::text, player_id, sola_level, drop_count, created_at FROM ascension_records" +
		builder.whereClause() +
		fmt.Sprintf(" ORDER BY created_at DESC, id DESC OFFSET $%d LIMIT $%d", len(builder.args)+1, len(builder.args)+2)
	args := append(append([]any{}, builder.args...), skip, limit)

	rows, err := a.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	defer rows.Close()

	var records []ascensionRecordResponse
	for rows.Next() {
		var record ascensionRecordResponse
		if err := rows.Scan(&record.ID, &record.Date, &record.PlayerID, &record.SolaLevel, &record.DropCount, &record.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "数据库操作失败")
			return
		}
		records = append(records, record)
	}

	writeJSON(w, http.StatusOK, listResponse[ascensionRecordResponse]{
		Data:        records,
		Total:       total,
		PageSize:    limit,
		CurrentPage: skip/limit + 1,
	})
}

func (a *API) handleAscensionDetailedStats(w http.ResponseWriter, r *http.Request, _ []string) {
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

func (a *API) handleAscensionPlayerIDs(w http.ResponseWriter, r *http.Request, _ []string) {
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

func (a *API) deleteAscensionRecord(w http.ResponseWriter, r *http.Request, _ []string) {
	recordID, err := parseIDFromPath(r.URL.Path, "/api/ascension-records/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "记录 ID 无效")
		return
	}

	deleted, err := deleteByID(r.Context(), a.db, "ascension_records", recordID)
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

func (a *API) handleResonanceRecords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.withEdit(a.createResonanceRecords)(w, r)
	case http.MethodGet:
		a.withView(a.getResonanceRecords)(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (a *API) handleResonanceRecordByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		methodNotAllowed(w)
		return
	}
	a.withEdit(a.deleteResonanceRecord)(w, r)
}

func (a *API) createResonanceRecords(w http.ResponseWriter, r *http.Request, _ []string) {
	var payload resonanceBatchCreate
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(payload.ResonanceRecords) == 0 {
		writeJSON(w, http.StatusOK, []resonanceRecordResponse{})
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

	records := make([]resonanceRecordResponse, 0, len(payload.ResonanceRecords))
	for _, item := range payload.ResonanceRecords {
		record, err := validateResonanceRecord(item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		var created resonanceRecordResponse
		err = tx.QueryRowContext(ctx, `
			INSERT INTO resonance_records (date, player_id, sola_level, gold, purple, blue, green)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, date::text, player_id, sola_level, gold, purple, blue, green, created_at
		`, record.Date, record.PlayerID, record.SolaLevel, record.Gold, record.Purple, record.Blue, record.Green).
			Scan(&created.ID, &created.Date, &created.PlayerID, &created.SolaLevel, &created.Gold, &created.Purple, &created.Blue, &created.Green, &created.CreatedAt)
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

func (a *API) getResonanceRecords(w http.ResponseWriter, r *http.Request, _ []string) {
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

	totalQuery := "SELECT COUNT(*) FROM resonance_records" + builder.whereClause()
	var total int
	if err := a.db.QueryRowContext(ctx, totalQuery, builder.args...).Scan(&total); err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}

	dataQuery := "SELECT id, date::text, player_id, sola_level, gold, purple, blue, green, created_at FROM resonance_records" +
		builder.whereClause() +
		fmt.Sprintf(" ORDER BY created_at DESC, id DESC OFFSET $%d LIMIT $%d", len(builder.args)+1, len(builder.args)+2)
	args := append(append([]any{}, builder.args...), skip, limit)

	rows, err := a.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	defer rows.Close()

	var records []resonanceRecordResponse
	for rows.Next() {
		var record resonanceRecordResponse
		if err := rows.Scan(&record.ID, &record.Date, &record.PlayerID, &record.SolaLevel, &record.Gold, &record.Purple, &record.Blue, &record.Green, &record.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "数据库操作失败")
			return
		}
		records = append(records, record)
	}

	writeJSON(w, http.StatusOK, listResponse[resonanceRecordResponse]{
		Data:        records,
		Total:       total,
		PageSize:    limit,
		CurrentPage: skip/limit + 1,
	})
}

func (a *API) handleResonanceDetailedStats(w http.ResponseWriter, r *http.Request, _ []string) {
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
		SELECT sola_level, gold, purple, blue, green, COUNT(*) AS count
		FROM resonance_records` + builder.whereClause() + `
		GROUP BY sola_level, gold, purple, blue, green`

	rows, err := a.db.QueryContext(ctx, query, builder.args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	defer rows.Close()

	type entry struct {
		Gold   int
		Purple int
		Blue   int
		Green  int
		Count  int
	}

	levelData := map[int][]entry{}
	for rows.Next() {
		var solaLevel int
		var item entry
		if err := rows.Scan(&solaLevel, &item.Gold, &item.Purple, &item.Blue, &item.Green, &item.Count); err != nil {
			writeError(w, http.StatusInternalServerError, "数据库操作失败")
			return
		}
		levelData[solaLevel] = append(levelData[solaLevel], item)
	}

	levels := mapKeys(levelData)
	sort.Sort(sort.Reverse(sort.IntSlice(levels)))

	response := resonanceDetailedStatsResponse{LevelStats: make([]resonanceSolaLevelStats, 0, len(levels))}
	for _, level := range levels {
		entries := levelData[level]
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Gold != entries[j].Gold {
				return entries[i].Gold > entries[j].Gold
			}
			if entries[i].Purple != entries[j].Purple {
				return entries[i].Purple > entries[j].Purple
			}
			if entries[i].Blue != entries[j].Blue {
				return entries[i].Blue > entries[j].Blue
			}
			return entries[i].Green > entries[j].Green
		})

		totalCount := 0
		totalGold := 0
		totalPurple := 0
		totalBlue := 0
		totalGreen := 0
		for _, item := range entries {
			totalCount += item.Count
			totalGold += item.Gold * item.Count
			totalPurple += item.Purple * item.Count
			totalBlue += item.Blue * item.Count
			totalGreen += item.Green * item.Count
		}

		combinations := make([]resonanceDropCombination, 0, len(entries))
		for _, item := range entries {
			percentage := 0.0
			if totalCount > 0 {
				percentage = roundTo(float64(item.Count)/float64(totalCount)*100, 1)
			}
			combinations = append(combinations, resonanceDropCombination{
				Gold:       item.Gold,
				Purple:     item.Purple,
				Blue:       item.Blue,
				Green:      item.Green,
				Count:      item.Count,
				Percentage: percentage,
			})
		}

		levelStats := resonanceSolaLevelStats{
			SolaLevel:    level,
			Combinations: combinations,
			TotalCount:   totalCount,
		}
		if totalCount > 0 {
			levelStats.AvgGold = roundTo(float64(totalGold)/float64(totalCount), 2)
			levelStats.AvgPurple = roundTo(float64(totalPurple)/float64(totalCount), 2)
			levelStats.AvgBlue = roundTo(float64(totalBlue)/float64(totalCount), 2)
			levelStats.AvgGreen = roundTo(float64(totalGreen)/float64(totalCount), 2)
		}

		response.LevelStats = append(response.LevelStats, levelStats)
	}

	writeJSON(w, http.StatusOK, response)
}

func (a *API) handleResonancePlayerIDs(w http.ResponseWriter, r *http.Request, _ []string) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	playerIDs, err := queryPlayerIDs(r.Context(), a.db, "resonance_records")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	writeJSON(w, http.StatusOK, playerIDs)
}

func (a *API) deleteResonanceRecord(w http.ResponseWriter, r *http.Request, _ []string) {
	recordID, err := parseIDFromPath(r.URL.Path, "/api/resonance-records/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "记录 ID 无效")
		return
	}

	deleted, err := deleteByID(r.Context(), a.db, "resonance_records", recordID)
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

func deleteByID(ctx context.Context, database *sql.DB, table string, id int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", table)
	result, err := database.ExecContext(ctx, query, id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

type filterBuilder struct {
	clauses []string
	args    []any
}

func buildCommonFilters(playerID, startDate, endDate string, solaLevel *int) (filterBuilder, error) {
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

func validateTacetRecord(input tacetRecordInput) (tacetRecordInput, error) {
	if err := validateDate(input.Date); err != nil {
		return tacetRecordInput{}, err
	}
	input.PlayerID = strings.TrimSpace(input.PlayerID)
	if input.PlayerID == "" {
		return tacetRecordInput{}, fmt.Errorf("player_id 不能为空")
	}
	if input.GoldTubes < 0 || input.PurpleTubes < 0 {
		return tacetRecordInput{}, fmt.Errorf("掉落数量不能为负数")
	}
	if input.ClaimCount == 0 {
		input.ClaimCount = 1
	}
	if input.ClaimCount < 1 || input.ClaimCount > 2 {
		return tacetRecordInput{}, fmt.Errorf("claim_count 必须为 1 或 2")
	}
	if input.SolaLevel == 0 {
		input.SolaLevel = 8
	}
	if input.SolaLevel < 1 {
		return tacetRecordInput{}, fmt.Errorf("sola_level 必须大于 0")
	}
	return input, nil
}

func validateAscensionRecord(input ascensionRecordInput) (ascensionRecordInput, error) {
	if err := validateDate(input.Date); err != nil {
		return ascensionRecordInput{}, err
	}
	input.PlayerID = strings.TrimSpace(input.PlayerID)
	if input.PlayerID == "" {
		return ascensionRecordInput{}, fmt.Errorf("player_id 不能为空")
	}
	if input.DropCount < 0 {
		return ascensionRecordInput{}, fmt.Errorf("drop_count 不能为负数")
	}
	if input.SolaLevel == 0 {
		input.SolaLevel = 8
	}
	if input.SolaLevel < 1 {
		return ascensionRecordInput{}, fmt.Errorf("sola_level 必须大于 0")
	}
	return input, nil
}

func validateResonanceRecord(input resonanceRecordInput) (resonanceRecordInput, error) {
	if err := validateDate(input.Date); err != nil {
		return resonanceRecordInput{}, err
	}
	input.PlayerID = strings.TrimSpace(input.PlayerID)
	if input.PlayerID == "" {
		return resonanceRecordInput{}, fmt.Errorf("player_id 不能为空")
	}
	if input.Gold < 0 || input.Purple < 0 || input.Blue < 0 || input.Green < 0 {
		return resonanceRecordInput{}, fmt.Errorf("掉落数量不能为负数")
	}
	if input.SolaLevel == 0 {
		input.SolaLevel = 8
	}
	if input.SolaLevel < 1 {
		return resonanceRecordInput{}, fmt.Errorf("sola_level 必须大于 0")
	}
	return input, nil
}

func validateDate(value string) error {
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return fmt.Errorf("date 参数无效")
	}
	return nil
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

var tacetSingleCombos = map[int][][2]int{
	8: {{4, 4}, {3, 4}},
	7: {{4, 4}, {4, 3}, {3, 4}, {3, 3}},
	6: {{4, 4}, {4, 3}, {3, 4}, {3, 3}},
	5: {{3, 6}, {3, 5}, {2, 6}, {2, 5}},
}

func splitTacetCombination(solaLevel, goldTubes, purpleTubes, claimCount int) [][2]int {
	if claimCount <= 1 {
		return [][2]int{{goldTubes, purpleTubes}}
	}

	combos := tacetSingleCombos[solaLevel]
	var matching [][2][2]int
	for _, left := range combos {
		for _, right := range combos {
			if left[0]+right[0] == goldTubes && left[1]+right[1] == purpleTubes {
				pair := [2][2]int{left, right}
				if lessCombo(pair[0], pair[1]) {
					pair[0], pair[1] = pair[1], pair[0]
				}
				matching = append(matching, pair)
			}
		}
	}

	if len(matching) == 0 {
		return [][2]int{{goldTubes, purpleTubes}}
	}

	sort.Slice(matching, func(i, j int) bool {
		if matching[i][0][0] != matching[j][0][0] {
			return matching[i][0][0] > matching[j][0][0]
		}
		if matching[i][0][1] != matching[j][0][1] {
			return matching[i][0][1] > matching[j][0][1]
		}
		if matching[i][1][0] != matching[j][1][0] {
			return matching[i][1][0] > matching[j][1][0]
		}
		return matching[i][1][1] > matching[j][1][1]
	})

	return [][2]int{matching[0][0], matching[0][1]}
}

func lessCombo(left, right [2]int) bool {
	if left[0] != right[0] {
		return left[0] < right[0]
	}
	return left[1] < right[1]
}

func mapKeys[V any](items map[int]V) []int {
	keys := make([]int, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	return keys
}
