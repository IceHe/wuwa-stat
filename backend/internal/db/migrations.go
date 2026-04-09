package db

import (
	"context"
	"database/sql"
	"fmt"
)

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

func ensureClaimCountColumn(ctx context.Context, database *sql.DB, table string) error {
	var columnExists bool
	if err := database.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = $1
			  AND column_name = 'claim_count'
		)
	`, table).Scan(&columnExists); err != nil {
		return err
	}

	if columnExists {
		return nil
	}

	_, err := database.ExecContext(ctx, fmt.Sprintf(`
		ALTER TABLE %s
		ADD COLUMN claim_count INTEGER NOT NULL DEFAULT 1
	`, table))
	return err
}
