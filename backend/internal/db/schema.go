package db

import (
	"context"
	"database/sql"
	"fmt"
)

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
			claim_count INTEGER NOT NULL DEFAULT 1,
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
	if err := ensureClaimCountColumn(ctx, database, "resonance_records"); err != nil {
		return err
	}

	return nil
}

func PrintSchemaSummary() {
	fmt.Println("表结构：")
	fmt.Println("- 表名: tacet_records")
	fmt.Println("- 字段: id, date, player_id, gold_tubes, purple_tubes, claim_count, sola_level, created_by_user_id, created_at")
	fmt.Println("- 表名: ascension_records")
	fmt.Println("- 字段: id, date, player_id, sola_level, drop_count, created_by_user_id, created_at")
	fmt.Println("- 表名: resonance_records")
	fmt.Println("- 字段: id, date, player_id, sola_level, claim_count, gold, purple, blue, green, created_by_user_id, created_at")
}
