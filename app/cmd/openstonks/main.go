package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
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
	watchlist := parseSymbols(mustEnv("SYMBOLS"))
	log.Printf("tracking %d symbols", len(watchlist))

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
	defer conn.Close(context.Background())

	if err := db.Migrate(ctx, conn); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	repo := stocks.NewRepository(conn)

	pricers := []sources.LivePricer{
		sources.NewYFinance(yfinanceURL),
	}
	if k := os.Getenv("POLYGON_API_KEY"); k != "" {
		pricers = append(pricers, sources.NewPolygon(k))
	}
	if k, s := os.Getenv("ALPACA_API_KEY"), os.Getenv("ALPACA_API_SECRET"); k != "" && s != "" {
		pricers = append(pricers, sources.NewAlpaca(k, s))
	}
	if k := os.Getenv("FINNHUB_API_KEY"); k != "" {
		pricers = append(pricers, sources.NewFinnhub(k))
	}

	ingestion.NewLive(pricers, repo, watchlist, liveInterval).Run(ctx)
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
