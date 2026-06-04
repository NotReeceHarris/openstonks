package ingestion

import (
	"context"
	"log"
	"time"

	"openstonks/internal/sources"
	"openstonks/internal/stocks"
)

const batchSize = 50

type Ingester struct {
	source   sources.PriceSource
	repo     *stocks.Repository
	symbols  []string
	interval time.Duration
}

func New(source sources.PriceSource, repo *stocks.Repository, symbols []string, interval time.Duration) *Ingester {
	return &Ingester{
		source:   source,
		repo:     repo,
		symbols:  symbols,
		interval: interval,
	}
}

func (ing *Ingester) Run(ctx context.Context) {
	log.Printf("ingester starting: source=%s symbols=%d interval=%s",
		ing.source.Name(), len(ing.symbols), ing.interval)

	ing.fetchAndStore(ctx)

	ticker := time.NewTicker(ing.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ing.fetchAndStore(ctx)
		}
	}
}

func (ing *Ingester) fetchAndStore(ctx context.Context) {
	total, failed := 0, 0

	for i := 0; i < len(ing.symbols); i += batchSize {
		end := i + batchSize
		if end > len(ing.symbols) {
			end = len(ing.symbols)
		}
		batch := ing.symbols[i:end]

		candles, err := ing.source.FetchCandles(ctx, batch)
		if err != nil {
			log.Printf("fetch batch %d-%d: %v", i, end, err)
			failed += len(batch)
			continue
		}

		for _, c := range candles {
			if err := ing.repo.InsertCandle(ctx, stocks.Candle{
				Time:   c.Time,
				Symbol: c.Symbol,
				Open:   c.Open,
				High:   c.High,
				Low:    c.Low,
				Close:  c.Close,
				Volume: c.Volume,
			}); err != nil {
				log.Printf("store %s: %v", c.Symbol, err)
				failed++
			} else {
				total++
			}
		}
	}

	log.Printf("cycle complete: stored=%d failed=%d", total, failed)
}
