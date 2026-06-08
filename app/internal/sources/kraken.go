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

// krakenPairs maps base symbol to Kraken's pair name. BNB is not listed on Kraken.
var krakenPairs = map[string]string{
	"BTC":  "XXBTZUSD",
	"ETH":  "XETHZUSD",
	"SOL":  "SOLUSD",
	"XRP":  "XXRPZUSD",
	"USDC": "USDCUSD",
	"ADA":  "ADAUSD",
	"AVAX": "AVAXUSD",
	"DOGE": "XDGUSD",
	"TRX":  "TRXUSD",
}

type krakenSource struct{ client *http.Client }

func NewKraken() LivePricer {
	return &krakenSource{client: &http.Client{Timeout: 10 * time.Second}}
}

func (k *krakenSource) Name() string { return "kraken" }

func (k *krakenSource) FetchLive(ctx context.Context, symbols []string) ([]LiveQuote, error) {
	pairToSymbol := map[string]string{}
	var pairs []string
	for _, s := range symbols {
		pair, ok := krakenPairs[s]
		if !ok {
			continue
		}
		pairs = append(pairs, pair)
		pairToSymbol[pair] = s
	}
	if len(pairs) == 0 {
		return nil, nil
	}

	url := fmt.Sprintf("https://api.kraken.com/0/public/Ticker?pair=%s", strings.Join(pairs, ","))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := k.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kraken: status %d", resp.StatusCode)
	}

	var body struct {
		Error  []string `json:"error"`
		Result map[string]struct {
			C []string `json:"c"` // last trade: [price, lot volume]
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if len(body.Error) > 0 {
		return nil, fmt.Errorf("kraken: %s", strings.Join(body.Error, ", "))
	}

	now := time.Now().UTC()
	results := make([]LiveQuote, 0, len(body.Result))
	for pair, ticker := range body.Result {
		sym, ok := pairToSymbol[pair]
		if !ok {
			continue
		}
		if len(ticker.C) == 0 {
			continue
		}
		price, err := strconv.ParseFloat(ticker.C[0], 64)
		if err != nil || price == 0 {
			continue
		}
		results = append(results, LiveQuote{
			Symbol:    sym,
			Price:     price,
			UpdatedAt: now,
			Source:    "kraken",
		})
	}
	return results, nil
}
