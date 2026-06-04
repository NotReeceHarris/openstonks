package stocks

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type Repository struct {
	conn *pgx.Conn
}

func NewRepository(conn *pgx.Conn) *Repository {
	return &Repository{conn: conn}
}

// UpsertLivePrice updates live_prices only when the incoming timestamp is newer.
// Returns true if the row was actually updated (i.e. the price changed).
func (r *Repository) UpsertLivePrice(ctx context.Context, symbol string, price float64, updatedAt time.Time, source string) (updated bool, err error) {
	tag, err := r.conn.Exec(ctx, `
		INSERT INTO live_prices (symbol, price, updated_at, source)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (symbol) DO UPDATE
			SET price      = EXCLUDED.price,
			    updated_at = EXCLUDED.updated_at,
			    source     = EXCLUDED.source
			WHERE EXCLUDED.updated_at > live_prices.updated_at
	`, symbol, price, updatedAt, source)
	if err != nil {
		return false, fmt.Errorf("upsert live price %s: %w", symbol, err)
	}
	return tag.RowsAffected() > 0, nil
}

// InsertHistory records a price into the history table.
func (r *Repository) InsertHistory(ctx context.Context, symbol string, price float64, t time.Time, source string) error {
	_, err := r.conn.Exec(ctx, `
		INSERT INTO price_history (time, symbol, price, source)
		VALUES ($1, $2, $3, $4)
	`, t, symbol, price, source)
	if err != nil {
		return fmt.Errorf("insert history %s: %w", symbol, err)
	}
	return nil
}
