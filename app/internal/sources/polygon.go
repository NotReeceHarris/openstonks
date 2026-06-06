package sources

import (
	"context"
	"time"

	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
)

type polygonClient struct {
	client *polygon.Client
}

func NewPolygon(apiKey string) LivePricer {
	return &polygonClient{client: polygon.New(apiKey)}
}

func (c *polygonClient) Name() string { return "polygon" }

func (c *polygonClient) FetchLive(ctx context.Context, symbols []string) ([]LiveQuote, error) {
	quotes := make([]LiveQuote, 0, len(symbols))

	for _, symbol := range symbols {
		params := models.GetTickerSnapshotParams{
			Ticker: symbol,
		}
		snap, err := c.client.GetTickerSnapshot(ctx, &params)
		if err != nil {
			continue
		}

		price := snap.Snapshot.Day.Close
		if price == 0 {
			price = snap.Snapshot.LastTrade.Price
		}
		if price == 0 {
			continue
		}

		quotes = append(quotes, LiveQuote{
			Symbol:    symbol,
			Price:     price,
			UpdatedAt: time.Now().UTC(),
			Source:    "polygon",
		})
	}

	return quotes, nil
}
