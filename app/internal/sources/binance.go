package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type binanceSource struct{ client *http.Client }

func NewBinance() LivePricer {
	return &binanceSource{client: &http.Client{Timeout: 10 * time.Second}}
}

func (b *binanceSource) Name() string { return "binance" }

func (b *binanceSource) FetchLive(ctx context.Context, symbols []string) ([]LiveQuote, error) {
	var pairs []string
	for _, s := range symbols {
		if strings.HasSuffix(s, "USDT") {
			pairs = append(pairs, s)
		}
	}
	if len(pairs) == 0 {
		return nil, nil
	}

	symbolsJSON, _ := json.Marshal(pairs)
	apiURL := "https://api.binance.com/api/v3/ticker/price?symbols=" + url.QueryEscape(string(symbolsJSON))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance: unexpected status %d", resp.StatusCode)
	}

	var tickers []struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tickers); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	results := make([]LiveQuote, 0, len(tickers))
	for _, t := range tickers {
		price, err := strconv.ParseFloat(t.Price, 64)
		if err != nil || price == 0 {
			continue
		}
		results = append(results, LiveQuote{
			Symbol:    t.Symbol,
			Price:     price,
			UpdatedAt: now,
			Source:    "binance",
		})
	}
	return results, nil
}
