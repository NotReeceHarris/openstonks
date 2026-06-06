package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	maxAttempts = 30
	retryDelay  = 2 * time.Second
)

func Connect(ctx context.Context, databaseURL string) (*pgx.Conn, error) {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		conn, err := pgx.Connect(ctx, databaseURL)
		if err == nil {
			if pingErr := conn.Ping(ctx); pingErr == nil {
				log.Printf("connected to database on attempt %d", attempt)
				return conn, nil
			}
			conn.Close(ctx)
		}
		log.Printf("attempt %d/%d: database not ready (%v), retrying in %s...",
			attempt, maxAttempts, err, retryDelay)
		time.Sleep(retryDelay)
	}
	return nil, fmt.Errorf("could not connect to database after %d attempts", maxAttempts)
}
