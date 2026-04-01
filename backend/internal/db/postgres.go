package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(databaseURL string) (*sql.DB, error) {
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	database.SetConnMaxLifetime(30 * time.Minute)
	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(5)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, err
	}

	return database, nil
}

func EnsureSchema(ctx context.Context, database *sql.DB) error {
	// Keep schema initialization idempotent so server and initdb can both call it safely.
	statements := []string{
		`CREATE TABLE IF NOT EXISTS tacet_records (
			id BIGSERIAL PRIMARY KEY,
			date DATE NOT NULL,
			player_id TEXT NOT NULL,
			gold_tubes INTEGER NOT NULL DEFAULT 0,
			purple_tubes INTEGER NOT NULL DEFAULT 0,
			claim_count INTEGER NOT NULL DEFAULT 1,
			sola_level INTEGER NOT NULL DEFAULT 8,
			created_by_user_id BIGINT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tacet_records_date ON tacet_records(date)`,
		`CREATE INDEX IF NOT EXISTS idx_tacet_records_player_id ON tacet_records(player_id)`,
		`CREATE TABLE IF NOT EXISTS ascension_records (
			id BIGSERIAL PRIMARY KEY,
			date DATE NOT NULL,
			player_id TEXT NOT NULL,
			sola_level INTEGER NOT NULL DEFAULT 8,
			drop_count INTEGER NOT NULL DEFAULT 0,
			created_by_user_id BIGINT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ascension_records_date ON ascension_records(date)`,
		`CREATE INDEX IF NOT EXISTS idx_ascension_records_player_id ON ascension_records(player_id)`,
		`CREATE TABLE IF NOT EXISTS resonance_records (
			id BIGSERIAL PRIMARY KEY,
			date DATE NOT NULL,
			player_id TEXT NOT NULL,
			sola_level INTEGER NOT NULL DEFAULT 8,
			gold INTEGER NOT NULL DEFAULT 0,
			purple INTEGER NOT NULL DEFAULT 0,
			blue INTEGER NOT NULL DEFAULT 0,
			green INTEGER NOT NULL DEFAULT 0,
			created_by_user_id BIGINT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_resonance_records_date ON resonance_records(date)`,
		`CREATE INDEX IF NOT EXISTS idx_resonance_records_player_id ON resonance_records(player_id)`,
	}

	for _, statement := range statements {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	if err := ensureTacetRecordSchema(ctx, database); err != nil {
		return err
	}
	if err := ensureCreatedByUserIDColumn(ctx, database, "tacet_records"); err != nil {
		return err
	}
	if err := ensureCreatedByUserIDColumn(ctx, database, "ascension_records"); err != nil {
		return err
	}
	if err := ensureCreatedByUserIDColumn(ctx, database, "resonance_records"); err != nil {
		return err
	}

	return nil
}

func ensureTacetRecordSchema(ctx context.Context, database *sql.DB) error {
	// Compatibility migration: older Python schema may miss claim_count and still contain reward_mode.
	// We backfill claim_count from reward_mode when possible.
	var tacetExists bool
	if err := database.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'tacet_records'
		)
	`).Scan(&tacetExists); err != nil {
		return err
	}

	if !tacetExists {
		return nil
	}

	var claimCountExists bool
	if err := database.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = 'tacet_records'
			  AND column_name = 'claim_count'
		)
	`).Scan(&claimCountExists); err != nil {
		return err
	}

	if claimCountExists {
		return nil
	}

	if _, err := database.ExecContext(ctx, `
		ALTER TABLE tacet_records
		ADD COLUMN claim_count INTEGER NOT NULL DEFAULT 1
	`); err != nil {
		return err
	}

	var rewardModeExists bool
	if err := database.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = 'tacet_records'
			  AND column_name = 'reward_mode'
		)
	`).Scan(&rewardModeExists); err != nil {
		return err
	}

	if !rewardModeExists {
		return nil
	}

	_, err := database.ExecContext(ctx, `
		UPDATE tacet_records
		SET claim_count = CASE WHEN reward_mode = 'double' THEN 2 ELSE 1 END
	`)
	return err
}

func ensureCreatedByUserIDColumn(ctx context.Context, database *sql.DB, table string) error {
	var columnExists bool
	if err := database.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = $1
			  AND column_name = 'created_by_user_id'
		)
	`, table).Scan(&columnExists); err != nil {
		return err
	}

	if columnExists {
		return nil
	}

	_, err := database.ExecContext(ctx, fmt.Sprintf(`
		ALTER TABLE %s
		ADD COLUMN created_by_user_id BIGINT
	`, table))
	return err
}

func PrintSchemaSummary() {
	fmt.Println("表结构：")
	fmt.Println("- 表名: tacet_records")
	fmt.Println("- 字段: id, date, player_id, gold_tubes, purple_tubes, claim_count, sola_level, created_by_user_id, created_at")
	fmt.Println("- 表名: ascension_records")
	fmt.Println("- 字段: id, date, player_id, sola_level, drop_count, created_by_user_id, created_at")
	fmt.Println("- 表名: resonance_records")
	fmt.Println("- 字段: id, date, player_id, sola_level, gold, purple, blue, green, created_by_user_id, created_at")
}
