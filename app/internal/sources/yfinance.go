package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type yfinanceClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewYFinance(baseURL string) PriceSource {
	return &yfinanceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *yfinanceClient) Name() string { return "yfinance" }

func (c *yfinanceClient) FetchCandles(ctx context.Context, symbols []string) ([]Candle, error) {
	endpoint := fmt.Sprintf("%s/candles?symbols=%s", c.baseURL, url.QueryEscape(strings.Join(symbols, ",")))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yfinance service returned %d", resp.StatusCode)
	}

	var body []struct {
		Symbol string  `json:"symbol"`
		Time   string  `json:"time"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume int64   `json:"volume"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	candles := make([]Candle, 0, len(body))
	for _, c := range body {
		t, err := time.Parse(time.RFC3339, c.Time)
		if err != nil {
			continue
		}
		candles = append(candles, Candle{
			Symbol: c.Symbol,
			Time:   t,
			Open:   c.Open,
			High:   c.High,
			Low:    c.Low,
			Close:  c.Close,
			Volume: c.Volume,
			Source: "yfinance",
		})
	}

	return candles, nil
}
