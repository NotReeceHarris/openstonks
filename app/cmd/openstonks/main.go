package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"openstonks/internal/db"
	"openstonks/internal/stocks"
)

func main() {

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	ctx := context.Background()

	conn, err := db.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close(ctx)

	if err := db.Migrate(ctx, conn); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	repo := stocks.NewRepository(conn)

	if err := repo.InsertSamples(ctx); err != nil {
		log.Fatalf("failed to insert sample data: %v", err)
	}

	prices, err := repo.All(ctx)
	if err != nil {
		log.Fatalf("failed to query data: %v", err)
	}

	fmt.Println()
	fmt.Printf("%-6s  %-6s  %-10s  %s\n", "ID", "SYMBOL", "PRICE", "RECORDED_AT")
	fmt.Println("------  ------  ----------  --------------------------")
	for _, p := range prices {
		fmt.Printf("%-6d  %-6s  %-10.2f  %s\n", p.ID, p.Symbol, p.Price, p.RecordedAt.Format(time.RFC3339))
	}
	fmt.Printf("\n%d rows returned.\n", len(prices))
}
