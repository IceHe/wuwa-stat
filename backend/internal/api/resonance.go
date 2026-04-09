package api

import (
	"context"
	"database/sql"
	"net/http"
	"sort"
	"time"
)

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

func (a *API) createResonanceRecords(w http.ResponseWriter, r *http.Request, auth authContext) {
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
			INSERT INTO resonance_records (date, player_id, sola_level, claim_count, gold, purple, blue, green, created_by_user_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id, date::text, player_id, sola_level, claim_count, gold, purple, blue, green, created_by_user_id, created_at
		`, record.Date, record.PlayerID, record.SolaLevel, record.ClaimCount, record.Gold, record.Purple, record.Blue, record.Green, auth.UserID).
			Scan(&created.ID, &created.Date, &created.PlayerID, &created.SolaLevel, &created.ClaimCount, &created.Gold, &created.Purple, &created.Blue, &created.Green, &created.CreatedByUserID, &created.CreatedAt)
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

func (a *API) getResonanceRecords(w http.ResponseWriter, r *http.Request, _ authContext) {
	params, err := buildListQueryParams(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := queryListRecords(r.Context(), a.db, "resonance_records", "id, date::text, player_id, sola_level, claim_count, gold, purple, blue, green, created_by_user_id, created_at", params, func(rows *sql.Rows) (resonanceRecordResponse, error) {
		var record resonanceRecordResponse
		err := rows.Scan(&record.ID, &record.Date, &record.PlayerID, &record.SolaLevel, &record.ClaimCount, &record.Gold, &record.Purple, &record.Blue, &record.Green, &record.CreatedByUserID, &record.CreatedAt)
		return record, err
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (a *API) handleResonanceDetailedStats(w http.ResponseWriter, r *http.Request, _ authContext) {
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
		SELECT sola_level, claim_count, gold, purple, blue, green, COUNT(*) AS count
		FROM resonance_records` + builder.whereClause() + `
		GROUP BY sola_level, claim_count, gold, purple, blue, green`

	rows, err := a.db.QueryContext(ctx, query, builder.args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	defer rows.Close()

	type entry struct {
		ClaimCount int
		Gold       int
		Purple     int
		Blue       int
		Green      int
		Count      int
	}

	levelData := map[int]map[[5]int]int{}
	for rows.Next() {
		var solaLevel int
		var item entry
		if err := rows.Scan(&solaLevel, &item.ClaimCount, &item.Gold, &item.Purple, &item.Blue, &item.Green, &item.Count); err != nil {
			writeError(w, http.StatusInternalServerError, "数据库操作失败")
			return
		}
		item.Gold, item.Purple, item.Blue, item.Green = normalizeResonanceDropForStats(
			solaLevel,
			item.ClaimCount,
			item.Gold,
			item.Purple,
			item.Blue,
			item.Green,
		)
		if _, ok := levelData[solaLevel]; !ok {
			levelData[solaLevel] = map[[5]int]int{}
		}
		key := [5]int{item.ClaimCount, item.Gold, item.Purple, item.Blue, item.Green}
		levelData[solaLevel][key] += item.Count
	}

	levels := mapKeys(levelData)
	sort.Sort(sort.Reverse(sort.IntSlice(levels)))

	response := resonanceDetailedStatsResponse{LevelStats: make([]resonanceSolaLevelStats, 0, len(levels))}
	for _, level := range levels {
		entryMap := levelData[level]
		entries := make([]entry, 0, len(entryMap))
		for key, count := range entryMap {
			entries = append(entries, entry{
				ClaimCount: key[0],
				Gold:       key[1],
				Purple:     key[2],
				Blue:       key[3],
				Green:      key[4],
				Count:      count,
			})
		}
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].ClaimCount != entries[j].ClaimCount {
				return entries[i].ClaimCount < entries[j].ClaimCount
			}
			if entries[i].Gold != entries[j].Gold {
				return entries[i].Gold < entries[j].Gold
			}
			if entries[i].Purple != entries[j].Purple {
				return entries[i].Purple < entries[j].Purple
			}
			if entries[i].Blue != entries[j].Blue {
				return entries[i].Blue < entries[j].Blue
			}
			return entries[i].Green < entries[j].Green
		})

		totalCount := 0
		totalClaimCount := 0
		totalGold := 0
		totalPurple := 0
		totalBlue := 0
		totalGreen := 0
		for _, item := range entries {
			totalCount += item.Count
			totalClaimCount += item.ClaimCount * item.Count
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
				ClaimCount: item.ClaimCount,
				Gold:       item.Gold,
				Purple:     item.Purple,
				Blue:       item.Blue,
				Green:      item.Green,
				Count:      item.Count,
				Percentage: percentage,
			})
		}

		levelStats := resonanceSolaLevelStats{
			SolaLevel:       level,
			Combinations:    combinations,
			TotalCount:      totalCount,
			TotalClaimCount: totalClaimCount,
			TotalGold:       totalGold,
			TotalPurple:     totalPurple,
			TotalBlue:       totalBlue,
			TotalGreen:      totalGreen,
		}
		if totalClaimCount > 0 {
			levelStats.AvgGold = roundTo(float64(totalGold)/float64(totalClaimCount), 2)
			levelStats.AvgPurple = roundTo(float64(totalPurple)/float64(totalClaimCount), 2)
			levelStats.AvgBlue = roundTo(float64(totalBlue)/float64(totalClaimCount), 2)
			levelStats.AvgGreen = roundTo(float64(totalGreen)/float64(totalClaimCount), 2)
		}

		response.LevelStats = append(response.LevelStats, levelStats)
	}

	writeJSON(w, http.StatusOK, response)
}

func (a *API) handleResonancePlayerIDs(w http.ResponseWriter, r *http.Request, _ authContext) {
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

func (a *API) deleteResonanceRecord(w http.ResponseWriter, r *http.Request, auth authContext) {
	recordID, err := parseIDFromPath(r.URL.Path, "/api/resonance-records/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "记录 ID 无效")
		return
	}

	deleted, authErr, err := deleteByID(r.Context(), a.db, "resonance_records", recordID, auth)
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
