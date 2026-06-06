package sources

import (
	"context"
	"time"
)

type LiveQuote struct {
	Symbol    string
	Price     float64
	UpdatedAt time.Time
	Source    string
}

// LivePricer fetches the current market price for a batch of symbols.
type LivePricer interface {
	Name() string
	FetchLive(ctx context.Context, symbols []string) ([]LiveQuote, error)
}
