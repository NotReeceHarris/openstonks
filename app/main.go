package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	ctx := context.Background()

	conn := waitForDB(ctx, databaseURL)
	defer conn.Close(ctx)

	if err := createSchema(ctx, conn); err != nil {
		log.Fatalf("failed to create schema: %v", err)
	}

	if err := insertSampleData(ctx, conn); err != nil {
		log.Fatalf("failed to insert sample data: %v", err)
	}

	if err := queryAndPrint(ctx, conn); err != nil {
		log.Fatalf("failed to query data: %v", err)
	}
}

// waitForDB retries connecting to the database until it succeeds or times out.
func waitForDB(ctx context.Context, databaseURL string) *pgx.Conn {
	const maxAttempts = 30
	const retryDelay = 2 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		conn, err := pgx.Connect(ctx, databaseURL)
		if err == nil {
			// Verify the connection is live.
			if pingErr := conn.Ping(ctx); pingErr == nil {
				log.Printf("connected to database on attempt %d", attempt)
				return conn
			}
			conn.Close(ctx)
		}
		log.Printf("attempt %d/%d: database not ready (%v), retrying in %s...",
			attempt, maxAttempts, err, retryDelay)
		time.Sleep(retryDelay)
	}
	log.Fatal("could not connect to database after maximum attempts")
	return nil // unreachable
}

// createSchema creates the stock_prices table and converts it to a TimescaleDB hypertable.
func createSchema(ctx context.Context, conn *pgx.Conn) error {
	// Enable the TimescaleDB extension if not already enabled.
	_, err := conn.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE`)
	if err != nil {
		return fmt.Errorf("enable timescaledb extension: %w", err)
	}

	// Create the table.
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

	// Convert to a TimescaleDB hypertable on recorded_at.
	// create_hypertable returns an error if the table is already a hypertable,
	// so we use if_not_exists => TRUE to make it idempotent.
	_, err = conn.Exec(ctx, `
		SELECT create_hypertable('stock_prices', 'recorded_at', if_not_exists => TRUE)
	`)
	if err != nil {
		return fmt.Errorf("create hypertable: %w", err)
	}

	log.Println("schema ready: stock_prices hypertable exists")
	return nil
}

// insertSampleData inserts a small set of stock price rows.
func insertSampleData(ctx context.Context, conn *pgx.Conn) error {
	type row struct {
		symbol     string
		price      float64
		recordedAt time.Time
	}

	now := time.Now().UTC()
	samples := []row{
		{"AAPL", 189.45, now.Add(-3 * time.Hour)},
		{"AAPL", 190.10, now.Add(-2 * time.Hour)},
		{"AAPL", 191.30, now.Add(-1 * time.Hour)},
		{"GOOG", 175.20, now.Add(-3 * time.Hour)},
		{"GOOG", 176.00, now.Add(-2 * time.Hour)},
		{"GOOG", 174.80, now.Add(-1 * time.Hour)},
		{"MSFT", 415.60, now.Add(-3 * time.Hour)},
		{"MSFT", 417.90, now.Add(-2 * time.Hour)},
		{"MSFT", 418.50, now.Add(-1 * time.Hour)},
	}

	for _, s := range samples {
		_, err := conn.Exec(ctx,
			`INSERT INTO stock_prices (symbol, price, recorded_at) VALUES ($1, $2, $3)`,
			s.symbol, s.price, s.recordedAt,
		)
		if err != nil {
			return fmt.Errorf("insert %s @ %v: %w", s.symbol, s.recordedAt, err)
		}
	}

	log.Printf("inserted %d sample rows", len(samples))
	return nil
}

// queryAndPrint fetches all rows ordered by symbol and time, then prints them.
func queryAndPrint(ctx context.Context, conn *pgx.Conn) error {
	rows, err := conn.Query(ctx, `
		SELECT id, symbol, price, recorded_at
		FROM stock_prices
		ORDER BY symbol, recorded_at
	`)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	fmt.Println()
	fmt.Printf("%-6s  %-6s  %-10s  %s\n", "ID", "SYMBOL", "PRICE", "RECORDED_AT")
	fmt.Println("------  ------  ----------  --------------------------")

	count := 0
	for rows.Next() {
		var id int
		var symbol string
		var price float64
		var recordedAt time.Time

		if err := rows.Scan(&id, &symbol, &price, &recordedAt); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		fmt.Printf("%-6d  %-6s  %-10.2f  %s\n", id, symbol, price, recordedAt.Format(time.RFC3339))
		count++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows iteration: %w", err)
	}

	fmt.Printf("\n%d rows returned.\n", count)
	return nil
}
