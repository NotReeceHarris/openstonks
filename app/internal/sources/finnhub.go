package sources

import (
	"context"
	"time"

	finnhub "github.com/Finnhub-Stock-API/finnhub-go/v2"
)

type finnhubClient struct {
	client *finnhub.DefaultApiService
}

func NewFinnhub(apiKey string) LivePricer {
	cfg := finnhub.NewConfiguration()
	cfg.AddDefaultHeader("X-Finnhub-Token", apiKey)
	return &finnhubClient{client: finnhub.NewAPIClient(cfg).DefaultApi}
}

func (c *finnhubClient) Name() string { return "finnhub" }

func (c *finnhubClient) FetchLive(ctx context.Context, symbols []string) ([]LiveQuote, error) {
	results := make([]LiveQuote, 0, len(symbols))

	for _, symbol := range symbols {
		q, _, err := c.client.Quote(ctx).Symbol(symbol).Execute()
		if err != nil {
			continue
		}
		price := q.GetC()
		if price == 0 {
			continue
		}
		results = append(results, LiveQuote{
			Symbol:    symbol,
			Price:     float64(price),
			UpdatedAt: time.Now().UTC(),
			Source:    "finnhub",
		})
	}
	return results, nil
}
