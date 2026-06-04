package ingestion

import (
	"context"
	"log"
	"time"

	"openstonks/internal/sources"
	"openstonks/internal/stocks"
)

const batchSize = 50

type LiveIngester struct {
	sources  []sources.LivePricer
	repo     *stocks.Repository
	symbols  []string
	interval time.Duration
}

func NewLive(pricers []sources.LivePricer, repo *stocks.Repository, symbols []string, interval time.Duration) *LiveIngester {
	return &LiveIngester{
		sources:  pricers,
		repo:     repo,
		symbols:  symbols,
		interval: interval,
	}
}

func (li *LiveIngester) Run(ctx context.Context) {
	names := make([]string, len(li.sources))
	for i, s := range li.sources {
		names[i] = s.Name()
	}
	log.Printf("live ingester starting: sources=%v symbols=%d interval=%s", names, len(li.symbols), li.interval)

	li.fetchAndStore(ctx)

	ticker := time.NewTicker(li.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			li.fetchAndStore(ctx)
		}
	}
}

func (li *LiveIngester) fetchAndStore(ctx context.Context) {
	for _, src := range li.sources {
		for i := 0; i < len(li.symbols); i += batchSize {
			end := i + batchSize
			if end > len(li.symbols) {
				end = len(li.symbols)
			}
			batch := li.symbols[i:end]

			quotes, err := src.FetchLive(ctx, batch)
			if err != nil {
				log.Printf("[%s] fetch: %v", src.Name(), err)
				continue
			}

			for _, q := range quotes {
				updated, err := li.repo.UpsertLivePrice(ctx, q.Symbol, q.Price, q.UpdatedAt, q.Source)
				if err != nil {
					log.Printf("[%s] upsert %s: %v", src.Name(), q.Symbol, err)
					continue
				}
				if updated {
					if err := li.repo.InsertHistory(ctx, q.Symbol, q.Price, q.UpdatedAt, q.Source); err != nil {
						log.Printf("[%s] history %s: %v", src.Name(), q.Symbol, err)
					}
				}
			}
		}
	}
}
