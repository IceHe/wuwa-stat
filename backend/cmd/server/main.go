package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"wuwa/stat/backend/internal/api"
	"wuwa/stat/backend/internal/config"
	"wuwa/stat/backend/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := db.EnsureSchema(ctx, database); err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.New(database, cfg).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("backend listening on :%s", cfg.Port)
	log.Fatal(server.ListenAndServe())
}
