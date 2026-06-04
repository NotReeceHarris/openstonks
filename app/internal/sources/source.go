package sources

import (
	"context"
	"time"
)

// Candle is a single OHLCV bar for a symbol.
type Candle struct {
	Symbol string
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
	Source string
}

// PriceSource is the interface every data source must implement.
type PriceSource interface {
	Name() string
	// FetchCandles fetches 1-minute OHLCV candles for a batch of symbols.
	FetchCandles(ctx context.Context, symbols []string) ([]Candle, error)
}
