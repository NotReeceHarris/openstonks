from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import yfinance as yf
import pandas as pd
from datetime import datetime, timezone
from typing import List

app = FastAPI()


class LiveQuote(BaseModel):
    symbol: str
    price: float
    updated_at: str
    source: str


@app.get("/live", response_model=List[LiveQuote])
def get_live(symbols: str):
    symbol_list = [s.strip().upper() for s in symbols.split(",") if s.strip()]
    if not symbol_list:
        raise HTTPException(status_code=400, detail="no symbols provided")

    now = datetime.now(timezone.utc).isoformat()
    results = []

    try:
        data = yf.download(
            tickers=symbol_list,
            period="1d",
            interval="1m",
            progress=False,
            auto_adjust=False,
        )

        if data.empty:
            return results

        # yfinance returns MultiIndex columns for multiple tickers,
        # flat columns for a single ticker.
        if isinstance(data.columns, pd.MultiIndex):
            close = data["Close"]
            for symbol in symbol_list:
                if symbol not in close.columns:
                    continue
                series = close[symbol].dropna()
                if series.empty:
                    continue
                price = float(series.iloc[-1])
                if price:
                    results.append(LiveQuote(
                        symbol=symbol,
                        price=price,
                        updated_at=now,
                        source="yfinance",
                    ))
        else:
            series = data["Close"].dropna()
            if not series.empty and len(symbol_list) == 1:
                price = float(series.iloc[-1])
                if price:
                    results.append(LiveQuote(
                        symbol=symbol_list[0],
                        price=price,
                        updated_at=now,
                        source="yfinance",
                    ))
    except Exception:
        pass

    return results


@app.get("/health")
def health():
    return {"status": "ok"}
