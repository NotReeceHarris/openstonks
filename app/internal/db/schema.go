package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate(ctx context.Context, conn *pgxpool.Pool) error {
	_, err := conn.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE`)
	if err != nil {
		return fmt.Errorf("enable timescaledb extension: %w", err)
	}

	// One row per symbol — only updated when the incoming timestamp is newer.
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS live_prices (
			symbol     TEXT        PRIMARY KEY,
			price      NUMERIC     NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL,
			source     TEXT        NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create live_prices table: %w", err)
	}

	// Full price history — a row is inserted every time live_prices actually changes.
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS price_history (
			time   TIMESTAMPTZ NOT NULL,
			symbol TEXT        NOT NULL,
			price  NUMERIC     NOT NULL,
			source TEXT        NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create price_history table: %w", err)
	}

	_, err = conn.Exec(ctx, `
		SELECT create_hypertable('price_history', 'time', if_not_exists => TRUE)
	`)
	if err != nil {
		return fmt.Errorf("create price_history hypertable: %w", err)
	}

	log.Println("schema ready")
	return nil
}
