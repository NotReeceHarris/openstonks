#!/usr/bin/env python3
"""
Real-time desktop visualizer for OpenStonks.

Usage:
    python visualize.py [SYMBOL1 SYMBOL2 ...]

Environment variables:
    DATABASE_URL   (default: postgres://openstonks:openstonks@localhost:5432/openstonks)
    POLL_INTERVAL  seconds between refreshes (default: 10)
"""

import os
import sys
import psycopg2
import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
import matplotlib.animation as animation

DATABASE_URL = os.environ.get(
    "DATABASE_URL",
    "postgres://openstonks:openstonks@localhost:5432/openstonks",
)
POLL_INTERVAL_S = int(os.environ.get("POLL_INTERVAL", "10"))


def get_conn():
    return psycopg2.connect(DATABASE_URL)


def fetch_symbols(conn):
    with conn.cursor() as cur:
        cur.execute("SELECT DISTINCT symbol FROM candles ORDER BY symbol")
        return [r[0] for r in cur.fetchall()]


def fetch_candles(conn, symbol, limit=100) -> pd.DataFrame:
    with conn.cursor() as cur:
        cur.execute(
            """
            SELECT time, open, high, low, close, volume
            FROM candles
            WHERE symbol = %s
            ORDER BY time DESC
            LIMIT %s
            """,
            (symbol, limit),
        )
        rows = cur.fetchall()

    if not rows:
        return pd.DataFrame()

    rows.reverse()
    df = pd.DataFrame(rows, columns=["time", "Open", "High", "Low", "Close", "Volume"])
    df["time"] = pd.to_datetime(df["time"], utc=True)
    df = df.set_index("time")
    return df


def main():
    symbols = [s.upper() for s in sys.argv[1:]]

    print(f"Connecting to {DATABASE_URL} ...")
    conn = get_conn()
    print("Connected.")

    if not symbols:
        symbols = fetch_symbols(conn)
    if not symbols:
        print("No symbols found in the database yet. Is the ingester running?")
        sys.exit(1)

    print(f"Plotting: {', '.join(symbols)}")

    cols = min(3, len(symbols))
    rows_n = (len(symbols) + cols - 1) // cols
    fig, axes = plt.subplots(rows_n, cols, figsize=(6 * cols, 4 * rows_n), squeeze=False)
    fig.suptitle("OpenStonks — live", fontsize=13)

    ax_list = [axes[r][c] for r in range(rows_n) for c in range(cols)]
    for ax in ax_list[len(symbols):]:
        ax.set_visible(False)

    lines, fills = [], []
    for ax in ax_list[:len(symbols)]:
        line, = ax.plot([], [], linewidth=1.5)
        fill = ax.fill_between([], [], alpha=0)
        lines.append(line)
        fills.append(fill)

    def update(_frame):
        nonlocal fills
        for i, symbol in enumerate(symbols):
            ax = ax_list[i]
            df = fetch_candles(conn, symbol)
            if df.empty:
                continue

            times = df.index.to_pydatetime()
            prices = df["Close"].tolist()

            lines[i].set_data(times, prices)
            fills[i].remove()
            fills[i] = ax.fill_between(times, prices, alpha=0.08, color=lines[i].get_color())

            ax.set_xlim(times[0], times[-1])
            lo, hi = min(prices), max(prices)
            margin = (hi - lo) * 0.05 or 1
            ax.set_ylim(lo - margin, hi + margin)
            ax.set_title(f"{symbol}  ${prices[-1]:.2f}", fontsize=10, pad=4)
            ax.xaxis.set_major_formatter(mdates.DateFormatter("%H:%M"))
            ax.tick_params(labelsize=7)
            fig.autofmt_xdate()

        fig.tight_layout(rect=[0, 0, 1, 0.96])
        return lines + fills

    ani = animation.FuncAnimation(  # noqa: F841
        fig, update, interval=POLL_INTERVAL_S * 1000, cache_frame_data=False
    )
    plt.show()
    conn.close()


if __name__ == "__main__":
    main()
