package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"openstonks/internal/db"
	"openstonks/internal/ingestion"
	"openstonks/internal/sources"
	"openstonks/internal/stocks"
)

func main() {
	databaseURL := mustEnv("DATABASE_URL")
	yfinanceURL := mustEnv("YFINANCE_URL")

	stockSymbols := parseSymbols(os.Getenv("STOCK_SYMBOLS"))
	cryptoSymbols := parseSymbols(os.Getenv("CRYPTO_SYMBOLS"))
	if len(stockSymbols) == 0 && len(cryptoSymbols) == 0 {
		log.Fatal("at least one of STOCK_SYMBOLS or CRYPTO_SYMBOLS must be set")
	}
	log.Printf("tracking %d stocks, %d crypto", len(stockSymbols), len(cryptoSymbols))

	liveInterval := 10 * time.Second
	if v := os.Getenv("LIVE_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			liveInterval = d
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conn, err := db.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close()

	if err := db.Migrate(ctx, conn); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	repo := stocks.NewRepository(conn)

	stockPricers := []sources.LivePricer{
		sources.NewYFinance(yfinanceURL),
	}
	if k := os.Getenv("POLYGON_API_KEY"); k != "" {
		stockPricers = append(stockPricers, sources.NewPolygon(k))
	}
	if k, s := os.Getenv("ALPACA_API_KEY"), os.Getenv("ALPACA_API_SECRET"); k != "" && s != "" {
		stockPricers = append(stockPricers, sources.NewAlpaca(k, s))
	}
	if k := os.Getenv("FINNHUB_API_KEY"); k != "" {
		stockPricers = append(stockPricers, sources.NewFinnhub(k))
	}

	cryptoPricers := []sources.LivePricer{
		sources.NewBinance(),
		sources.NewCoinGecko(),
		sources.NewCoinCap(),
		sources.NewKraken(),
	}
	if k := os.Getenv("COINMARKETCAP_API_KEY"); k != "" {
		cryptoPricers = append(cryptoPricers, sources.NewCoinMarketCap(k))
	}

	var wg sync.WaitGroup

	if len(stockSymbols) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ingestion.NewLive(stockPricers, repo, stockSymbols, liveInterval).Run(ctx)
		}()
	}

	if len(cryptoSymbols) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ingestion.NewLive(cryptoPricers, repo, cryptoSymbols, liveInterval).Run(ctx)
		}()
	}

	wg.Wait()
	log.Println("shutdown complete")
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}

func parseSymbols(csv string) []string {
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(strings.ToUpper(p)); s != "" {
			out = append(out, s)
		}
	}
	return out
}
