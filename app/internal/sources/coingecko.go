package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var coingeckoIDs = map[string]string{
	"BTC":  "bitcoin",
	"ETH":  "ethereum",
	"BNB":  "binancecoin",
	"SOL":  "solana",
	"XRP":  "ripple",
	"USDC": "usd-coin",
	"ADA":  "cardano",
	"AVAX": "avalanche-2",
	"DOGE": "dogecoin",
	"TRX":  "tron",
}

type coingeckoSource struct{ client *http.Client }

func NewCoinGecko() LivePricer {
	return &coingeckoSource{client: &http.Client{Timeout: 10 * time.Second}}
}

func (c *coingeckoSource) Name() string { return "coingecko" }

func (c *coingeckoSource) FetchLive(ctx context.Context, symbols []string) ([]LiveQuote, error) {
	idToSymbol := map[string]string{}
	var ids []string
	for _, s := range symbols {
		id, ok := coingeckoIDs[s]
		if !ok {
			continue
		}
		ids = append(ids, id)
		idToSymbol[id] = s
	}
	if len(ids) == 0 {
		return nil, nil
	}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", strings.Join(ids, ","))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko: status %d", resp.StatusCode)
	}

	var data map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	results := make([]LiveQuote, 0, len(data))
	for id, prices := range data {
		price := prices["usd"]
		if price == 0 {
			continue
		}
		sym, ok := idToSymbol[id]
		if !ok {
			continue
		}
		results = append(results, LiveQuote{
			Symbol:    sym,
			Price:     price,
			UpdatedAt: now,
			Source:    "coingecko",
		})
	}
	return results, nil
}
