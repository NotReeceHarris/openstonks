package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var coincapIDs = map[string]string{
	"BTC":  "bitcoin",
	"ETH":  "ethereum",
	"BNB":  "binance-coin",
	"SOL":  "solana",
	"XRP":  "xrp",
	"USDC": "usd-coin",
	"ADA":  "cardano",
	"AVAX": "avalanche",
	"DOGE": "dogecoin",
	"TRX":  "tron",
}

type coincapSource struct{ client *http.Client }

func NewCoinCap() LivePricer {
	return &coincapSource{client: &http.Client{Timeout: 10 * time.Second}}
}

func (c *coincapSource) Name() string { return "coincap" }

func (c *coincapSource) FetchLive(ctx context.Context, symbols []string) ([]LiveQuote, error) {
	idToSymbol := map[string]string{}
	var ids []string
	for _, s := range symbols {
		id, ok := coincapIDs[s]
		if !ok {
			continue
		}
		ids = append(ids, id)
		idToSymbol[id] = s
	}
	if len(ids) == 0 {
		return nil, nil
	}

	url := fmt.Sprintf("https://api.coincap.io/v2/assets?ids=%s", strings.Join(ids, ","))
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
		return nil, fmt.Errorf("coincap: status %d", resp.StatusCode)
	}

	var body struct {
		Data []struct {
			ID       string `json:"id"`
			PriceUSD string `json:"priceUsd"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	results := make([]LiveQuote, 0, len(body.Data))
	for _, asset := range body.Data {
		price, err := strconv.ParseFloat(asset.PriceUSD, 64)
		if err != nil || price == 0 {
			continue
		}
		sym, ok := idToSymbol[asset.ID]
		if !ok {
			continue
		}
		results = append(results, LiveQuote{
			Symbol:    sym,
			Price:     price,
			UpdatedAt: now,
			Source:    "coincap",
		})
	}
	return results, nil
}
