package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL               string
	FrontendURL               string
	AuthServiceURL            string
	AuthServiceTimeoutSeconds float64
	Port                      string
}

func Load() (Config, error) {
	loadEnvFile(".env")
	loadEnvFile(filepath.Join("backend", ".env"))

	cfg := Config{
		DatabaseURL:               strings.TrimSpace(os.Getenv("DATABASE_URL")),
		FrontendURL:               getEnv("FRONTEND_URL", "http://localhost:5173"),
		AuthServiceURL:            getEnv("AUTH_SERVICE_URL", "http://127.0.0.1:8080"),
		AuthServiceTimeoutSeconds: getEnvFloat("AUTH_SERVICE_TIMEOUT_SECONDS", 3),
		Port:                      getEnv("PORT", "8000"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvFloat(key string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}

	return parsed
}

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)

		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		_ = os.Setenv(key, value)
	}
}
