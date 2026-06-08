package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxAttempts = 30
	retryDelay  = 2 * time.Second
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pool, err := pgxpool.New(ctx, databaseURL)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				log.Printf("connected to database on attempt %d", attempt)
				return pool, nil
			}
			pool.Close()
		}
		log.Printf("attempt %d/%d: database not ready (%v), retrying in %s...",
			attempt, maxAttempts, err, retryDelay)
		time.Sleep(retryDelay)
	}
	return nil, fmt.Errorf("could not connect to database after %d attempts", maxAttempts)
}
