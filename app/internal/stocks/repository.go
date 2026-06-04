package stocks

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type Candle struct {
	Time   time.Time
	Symbol string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

type Repository struct {
	conn *pgx.Conn
}

func NewRepository(conn *pgx.Conn) *Repository {
	return &Repository{conn: conn}
}

func (r *Repository) InsertCandle(ctx context.Context, c Candle) error {
	_, err := r.conn.Exec(ctx, `
		INSERT INTO candles (time, symbol, open, high, low, close, volume)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (symbol, time) DO NOTHING
	`, c.Time, c.Symbol, c.Open, c.High, c.Low, c.Close, c.Volume)
	if err != nil {
		return fmt.Errorf("insert candle %s @ %v: %w", c.Symbol, c.Time, err)
	}
	return nil
}
