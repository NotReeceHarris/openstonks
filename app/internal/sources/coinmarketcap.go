package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// cmcIDs maps base symbol to CoinMarketCap coin ID to avoid duplicate-symbol ambiguity.
var cmcIDs = map[string]string{
	"BTC":  "1",
	"ETH":  "1027",
	"BNB":  "1839",
	"SOL":  "5426",
	"XRP":  "52",
	"USDC": "3408",
	"ADA":  "2010",
	"AVAX": "5805",
	"DOGE": "74",
	"TRX":  "1958",
}

type coinmarketcapSource struct {
	apiKey string
	client *http.Client
}

func NewCoinMarketCap(apiKey string) LivePricer {
	return &coinmarketcapSource{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *coinmarketcapSource) Name() string { return "coinmarketcap" }

func (c *coinmarketcapSource) FetchLive(ctx context.Context, symbols []string) ([]LiveQuote, error) {
	idToSymbol := map[string]string{}
	var ids []string
	for _, s := range symbols {
		id, ok := cmcIDs[s]
		if !ok {
			continue
		}
		ids = append(ids, id)
		idToSymbol[id] = s
	}
	if len(ids) == 0 {
		return nil, nil
	}

	url := fmt.Sprintf("https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?id=%s&convert=USD", strings.Join(ids, ","))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-CMC_PRO_API_KEY", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coinmarketcap: status %d", resp.StatusCode)
	}

	var body struct {
		Data map[string]struct {
			Quote map[string]struct {
				Price float64 `json:"price"`
			} `json:"quote"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	results := make([]LiveQuote, 0, len(body.Data))
	for id, coin := range body.Data {
		usd, ok := coin.Quote["USD"]
		if !ok || usd.Price == 0 {
			continue
		}
		sym, ok := idToSymbol[id]
		if !ok {
			continue
		}
		results = append(results, LiveQuote{
			Symbol:    sym,
			Price:     usd.Price,
			UpdatedAt: now,
			Source:    "coinmarketcap",
		})
	}
	return results, nil
}
