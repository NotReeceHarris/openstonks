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
		CREATE TABLE IF NOT EXISTS stock_prices (
			id          SERIAL,
			symbol      TEXT        NOT NULL,
			price       NUMERIC     NOT NULL,
			recorded_at TIMESTAMPTZ NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	_, err = conn.Exec(ctx, `
		SELECT create_hypertable('stock_prices', 'recorded_at', if_not_exists => TRUE)
	`)
	if err != nil {
		return fmt.Errorf("create hypertable: %w", err)
	}

	log.Println("schema ready: stock_prices hypertable exists")
	return nil
}
