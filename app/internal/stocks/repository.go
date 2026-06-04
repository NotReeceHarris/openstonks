package stocks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type Price struct {
	ID         int
	Symbol     string
	Price      float64
	RecordedAt time.Time
}

type Repository struct {
	conn *pgx.Conn
}

func NewRepository(conn *pgx.Conn) *Repository {
	return &Repository{conn: conn}
}

func (r *Repository) InsertSamples(ctx context.Context) error {
	now := time.Now().UTC()
	samples := []Price{
		{Symbol: "AAPL", Price: 189.45, RecordedAt: now.Add(-3 * time.Hour)},
		{Symbol: "AAPL", Price: 190.10, RecordedAt: now.Add(-2 * time.Hour)},
		{Symbol: "AAPL", Price: 191.30, RecordedAt: now.Add(-1 * time.Hour)},
		{Symbol: "GOOG", Price: 175.20, RecordedAt: now.Add(-3 * time.Hour)},
		{Symbol: "GOOG", Price: 176.00, RecordedAt: now.Add(-2 * time.Hour)},
		{Symbol: "GOOG", Price: 174.80, RecordedAt: now.Add(-1 * time.Hour)},
		{Symbol: "MSFT", Price: 415.60, RecordedAt: now.Add(-3 * time.Hour)},
		{Symbol: "MSFT", Price: 417.90, RecordedAt: now.Add(-2 * time.Hour)},
		{Symbol: "MSFT", Price: 418.50, RecordedAt: now.Add(-1 * time.Hour)},
	}

	for _, s := range samples {
		_, err := r.conn.Exec(ctx,
			`INSERT INTO stock_prices (symbol, price, recorded_at) VALUES ($1, $2, $3)`,
			s.Symbol, s.Price, s.RecordedAt,
		)
		if err != nil {
			return fmt.Errorf("insert %s @ %v: %w", s.Symbol, s.RecordedAt, err)
		}
	}

	log.Printf("inserted %d sample rows", len(samples))
	return nil
}

func (r *Repository) All(ctx context.Context) ([]Price, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT id, symbol, price, recorded_at
		FROM stock_prices
		ORDER BY symbol, recorded_at
	`)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var prices []Price
	for rows.Next() {
		var p Price
		if err := rows.Scan(&p.ID, &p.Symbol, &p.Price, &p.RecordedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		prices = append(prices, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return prices, nil
}
