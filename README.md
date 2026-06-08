# OpenStonks

Real-time stock and cryptocurrency price ingestion into a TimescaleDB time-series database. Fetches live prices from multiple market data sources and keeps both a current snapshot and full price history per symbol.

## Architecture

```
YFinance Service ─┐
Binance API       │
Polygon API       ├──► Go Ingester ──► PostgreSQL + TimescaleDB
Alpaca API        │
Finnhub API      ─┘
```

The Go ingester polls all configured sources on a set interval, upserts the latest price per symbol into `live_prices`, and appends to `price_history` only when the price actually changes.

## Stack

| Component            | Technology                  |
|----------------------|-----------------------------|
| Ingestion service    | Go 1.25                     |
| Time-series database | PostgreSQL 16 + TimescaleDB |
| YFinance adapter     | Python (FastAPI)            |
| Containerization     | Docker + Docker Compose     |

## Prerequisites

- Docker and Docker Compose
- (Optional) API keys for Polygon, Alpaca, or Finnhub

## Quickstart

```bash
# Production
make build

# Development (with live reload)
make dev
```

Both commands start the full stack: TimescaleDB, the YFinance microservice, and the Go ingester.

## Configuration

Set environment variables in your shell or a `.env` file before running. Only `SYMBOLS` and `UPDATE_INTERVAL` are required — API keys are optional and enable additional data sources.

| Variable            | Default                        | Description                                     |
|---------------------|--------------------------------|-------------------------------------------------|
| `SYMBOLS`           | `69 major stocks + 10 crypto`  | Comma-separated list of ticker symbols to track |
| `UPDATE_INTERVAL`   | `1s`                           | How often to fetch prices (Go duration string)  |
| `POLYGON_API_KEY`   | `—`                            | Enables Polygon.io source                       |
| `ALPACA_API_KEY`    | `—`                            | Enables Alpaca source                           |
| `ALPACA_API_SECRET` | `—`                            | Required with `ALPACA_API_KEY`                  |
| `FINNHUB_API_KEY`   | `—`                            | Enables Finnhub source                          |

### Cryptocurrency

Crypto is supported out of the box via the Binance public REST API (no API key required). Any symbol ending in `USDT` in the `SYMBOLS` list is automatically routed to Binance. The default watchlist includes the top 10 cryptocurrencies by market cap:

```
BTCUSDT, ETHUSDT, BNBUSDT, SOLUSDT, XRPUSDT,
USDCUSDT, ADAUSDT, AVAXUSDT, DOGEUSDT, TRXUSDT
```

Add any other Binance USDT pair to `SYMBOLS` to track it alongside equities.

## Connecting to the Database

The database is exposed on `localhost:5432` with these credentials:

| Field    | Value        |
|----------|--------------|
| Host     | `localhost`  |
| Port     | `5432`       |
| Database | `openstonks` |
| Username | `openstonks` |
| Password | `openstonks` |

Connection string:

```
postgres://openstonks:openstonks@localhost:5432/openstonks?sslmode=disable
```

## Database Schema

**`live_prices`** — latest price per symbol (upserted on each fetch)

| Column     | Type          | Notes                                |
|------------|---------------|--------------------------------------|
| symbol     | `text`        | PRIMARY KEY                          |
| price      | `numeric`     |                                      |
| updated_at | `timestamptz` | Only updated when timestamp is newer |
| source     | `text`        | Which data source provided the price |

**`price_history`** — full audit trail (TimescaleDB hypertable, partitioned by time)

| Column | Type          | Notes                                     |
|--------|---------------|-------------------------------------------|
| time   | `timestamptz` | Partition key                             |
| symbol | `text`        |                                           |
| price  | `numeric`     | Only recorded when price actually changes |
| source | `text`        |                                           |
