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

func NewYFinance(baseURL string) LivePricer {
	return &yfinanceClient{
		baseURL: baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *yfinanceClient) Name() string { return "yfinance" }

func (c *yfinanceClient) FetchLive(ctx context.Context, symbols []string) ([]LiveQuote, error) {
	if len(symbols) == 0 {
		return nil, nil
	}

	endpoint := fmt.Sprintf("%s/live?symbols=%s", c.baseURL, url.QueryEscape(strings.Join(symbols, ",")))

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
		Symbol    string  `json:"symbol"`
		Price     float64 `json:"price"`
		UpdatedAt string  `json:"updated_at"`
		Source    string  `json:"source"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	quotes := make([]LiveQuote, 0, len(body))
	for _, q := range body {
		t, err := time.Parse(time.RFC3339, q.UpdatedAt)
		if err != nil {
			t = time.Now().UTC()
		}
		quotes = append(quotes, LiveQuote{
			Symbol:    q.Symbol,
			Price:     q.Price,
			UpdatedAt: t,
			Source:    q.Source,
		})
	}
	return quotes, nil
}
