package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// binancePairs maps base symbol to its Binance USDT pair name.
var binancePairs = map[string]string{
	"BTC":  "BTCUSDT",
	"ETH":  "ETHUSDT",
	"BNB":  "BNBUSDT",
	"SOL":  "SOLUSDT",
	"XRP":  "XRPUSDT",
	"USDC": "USDCUSDT",
	"ADA":  "ADAUSDT",
	"AVAX": "AVAXUSDT",
	"DOGE": "DOGEUSDT",
	"TRX":  "TRXUSDT",
}

type binanceSource struct{ client *http.Client }

func NewBinance() LivePricer {
	return &binanceSource{client: &http.Client{Timeout: 10 * time.Second}}
}

func (b *binanceSource) Name() string { return "binance" }

func (b *binanceSource) FetchLive(ctx context.Context, symbols []string) ([]LiveQuote, error) {
	pairToBase := map[string]string{}
	var pairs []string
	for _, s := range symbols {
		pair, ok := binancePairs[s]
		if !ok {
			continue
		}
		pairs = append(pairs, pair)
		pairToBase[pair] = s
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
		base, ok := pairToBase[t.Symbol]
		if !ok {
			continue
		}
		results = append(results, LiveQuote{
			Symbol:    base,
			Price:     price,
			UpdatedAt: now,
			Source:    "binance",
		})
	}
	return results, nil
}
