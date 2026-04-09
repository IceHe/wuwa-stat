package api

import (
	"database/sql"
	"net/http"

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
