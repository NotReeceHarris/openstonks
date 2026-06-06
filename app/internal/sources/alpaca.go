package sources

import (
	"context"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

type alpacaClient struct {
	client *marketdata.Client
}

func NewAlpaca(apiKey, apiSecret string) LivePricer {
	return &alpacaClient{
		client: marketdata.NewClient(marketdata.ClientOpts{
			APIKey:    apiKey,
			APISecret: apiSecret,
		}),
	}
}

func (c *alpacaClient) Name() string { return "alpaca" }

func (c *alpacaClient) FetchLive(ctx context.Context, symbols []string) ([]LiveQuote, error) {
	quotes, err := c.client.GetLatestQuotes(symbols, marketdata.GetLatestQuoteRequest{})
	if err != nil {
		return nil, err
	}

	results := make([]LiveQuote, 0, len(quotes))
	for symbol, q := range quotes {
		price := (q.AskPrice + q.BidPrice) / 2
		if price == 0 {
			continue
		}
		results = append(results, LiveQuote{
			Symbol:    symbol,
			Price:     price,
			UpdatedAt: q.Timestamp.UTC(),
			Source:    "alpaca",
		})
	}
	return results, nil
}
