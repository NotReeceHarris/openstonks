package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func Migrate(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE`)
	if err != nil {
		return fmt.Errorf("enable timescaledb extension: %w", err)
	}

	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS candles (
			time   TIMESTAMPTZ NOT NULL,
			symbol TEXT        NOT NULL,
			open   NUMERIC     NOT NULL,
			high   NUMERIC     NOT NULL,
			low    NUMERIC     NOT NULL,
			close  NUMERIC     NOT NULL,
			volume BIGINT      NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create candles table: %w", err)
	}

	_, err = conn.Exec(ctx, `
		SELECT create_hypertable('candles', 'time', if_not_exists => TRUE)
	`)
	if err != nil {
		return fmt.Errorf("create hypertable: %w", err)
	}

	// Prevent duplicate candles for the same symbol+minute
	_, err = conn.Exec(ctx, `
		CREATE UNIQUE INDEX IF NOT EXISTS candles_symbol_time_idx ON candles (symbol, time)
	`)
	if err != nil {
		return fmt.Errorf("create unique index: %w", err)
	}

	log.Println("schema ready: candles hypertable exists")
	return nil
}
