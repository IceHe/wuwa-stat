package api

import "time"

type messageResponse struct {
	Message string `json:"message"`
}

type authMeResponse struct {
	UserID      int64    `json:"user_id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type listResponse[T any] struct {
	Data        []T `json:"data"`
	Total       int `json:"total"`
	PageSize    int `json:"page_size"`
	CurrentPage int `json:"current_page"`
}

type tacetRecordInput struct {
	Date        string `json:"date"`
	PlayerID    string `json:"player_id"`
	GoldTubes   int    `json:"gold_tubes"`
	PurpleTubes int    `json:"purple_tubes"`
	ClaimCount  int    `json:"claim_count"`
	SolaLevel   int    `json:"sola_level"`
}

type tacetBatchCreate struct {
	TacetRecords []tacetRecordInput `json:"tacet_records"`
}

type tacetRecordResponse struct {
	ID              int64     `json:"id"`
	Date            string    `json:"date"`
	PlayerID        string    `json:"player_id"`
	GoldTubes       int       `json:"gold_tubes"`
	PurpleTubes     int       `json:"purple_tubes"`
	ClaimCount      int       `json:"claim_count"`
	SolaLevel       int       `json:"sola_level"`
	CreatedByUserID *int64    `json:"created_by_user_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type statsResponse struct {
	TotalRecords     int     `json:"total_records"`
	TotalClaimCount  int     `json:"total_claim_count"`
	TotalGoldTubes   int     `json:"total_gold_tubes"`
	TotalPurpleTubes int     `json:"total_purple_tubes"`
	AvgGoldTubes     float64 `json:"avg_gold_tubes"`
	AvgPurpleTubes   float64 `json:"avg_purple_tubes"`
	PlayerCount      int     `json:"player_count"`
}

type dropCombination struct {
	GoldTubes   int     `json:"gold_tubes"`
	PurpleTubes int     `json:"purple_tubes"`
	ClaimCount  int     `json:"claim_count"`
	Experience  int     `json:"experience"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
}

type solaLevelStats struct {
	SolaLevel     int               `json:"sola_level"`
	Combinations  []dropCombination `json:"combinations"`
	TotalCount    int               `json:"total_count"`
	AvgExperience float64           `json:"avg_experience"`
}

type detailedStatsResponse struct {
	LevelStats []solaLevelStats `json:"level_stats"`
}

type ascensionRecordInput struct {
	Date      string `json:"date"`
	PlayerID  string `json:"player_id"`
	SolaLevel int    `json:"sola_level"`
	DropCount int    `json:"drop_count"`
}

type ascensionBatchCreate struct {
	AscensionRecords []ascensionRecordInput `json:"ascension_records"`
}

type ascensionRecordResponse struct {
	ID              int64     `json:"id"`
	Date            string    `json:"date"`
	PlayerID        string    `json:"player_id"`
	SolaLevel       int       `json:"sola_level"`
	DropCount       int       `json:"drop_count"`
	CreatedByUserID *int64    `json:"created_by_user_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type ascensionDropCombination struct {
	DropCount  int     `json:"drop_count"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type ascensionSolaLevelStats struct {
	SolaLevel    int                        `json:"sola_level"`
	Combinations []ascensionDropCombination `json:"combinations"`
	TotalCount   int                        `json:"total_count"`
	AvgDropCount float64                    `json:"avg_drop_count"`
}

type ascensionDetailedStatsResponse struct {
	LevelStats []ascensionSolaLevelStats `json:"level_stats"`
}

type resonanceRecordInput struct {
	Date       string `json:"date"`
	PlayerID   string `json:"player_id"`
	SolaLevel  int    `json:"sola_level"`
	ClaimCount int    `json:"claim_count"`
	Gold       int    `json:"gold"`
	Purple     int    `json:"purple"`
	Blue       int    `json:"blue"`
	Green      int    `json:"green"`
}

type resonanceBatchCreate struct {
	ResonanceRecords []resonanceRecordInput `json:"resonance_records"`
}

type resonanceRecordResponse struct {
	ID              int64     `json:"id"`
	Date            string    `json:"date"`
	PlayerID        string    `json:"player_id"`
	SolaLevel       int       `json:"sola_level"`
	ClaimCount      int       `json:"claim_count"`
	Gold            int       `json:"gold"`
	Purple          int       `json:"purple"`
	Blue            int       `json:"blue"`
	Green           int       `json:"green"`
	CreatedByUserID *int64    `json:"created_by_user_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type resonanceDropCombination struct {
	ClaimCount int     `json:"claim_count"`
	Gold       int     `json:"gold"`
	Purple     int     `json:"purple"`
	Blue       int     `json:"blue"`
	Green      int     `json:"green"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type resonanceSolaLevelStats struct {
	SolaLevel       int                        `json:"sola_level"`
	Combinations    []resonanceDropCombination `json:"combinations"`
	TotalCount      int                        `json:"total_count"`
	TotalClaimCount int                        `json:"total_claim_count"`
	TotalGold       int                        `json:"total_gold"`
	TotalPurple     int                        `json:"total_purple"`
	TotalBlue       int                        `json:"total_blue"`
	TotalGreen      int                        `json:"total_green"`
	AvgGold         float64                    `json:"avg_gold"`
	AvgPurple       float64                    `json:"avg_purple"`
	AvgBlue         float64                    `json:"avg_blue"`
	AvgGreen        float64                    `json:"avg_green"`
}

type resonanceDetailedStatsResponse struct {
	LevelStats []resonanceSolaLevelStats `json:"level_stats"`
}
