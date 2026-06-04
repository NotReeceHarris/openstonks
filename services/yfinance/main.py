from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import yfinance as yf
from datetime import datetime, timezone
from typing import List

app = FastAPI()


class Candle(BaseModel):
    symbol: str
    time: str
    open: float
    high: float
    low: float
    close: float
    volume: int


@app.get("/candles", response_model=List[Candle])
def get_candles(symbols: str):
    symbol_list = [s.strip().upper() for s in symbols.split(",") if s.strip()]
    if not symbol_list:
        raise HTTPException(status_code=400, detail="no symbols provided")

    results = []

    for symbol in symbol_list:
        try:
            df = yf.download(
                tickers=symbol,
                period="1d",
                interval="1m",
                auto_adjust=True,
                progress=False,
            )
            if df.empty:
                continue

            # Flatten MultiIndex columns that yfinance sometimes produces
            if isinstance(df.columns, type(df.columns)) and hasattr(df.columns, "droplevel"):
                try:
                    df.columns = df.columns.droplevel(1)
                except Exception:
                    pass

            for ts, row in df.iterrows():
                try:
                    results.append(Candle(
                        symbol=symbol,
                        time=ts.to_pydatetime().astimezone(timezone.utc).isoformat(),
                        open=float(row["Open"]),
                        high=float(row["High"]),
                        low=float(row["Low"]),
                        close=float(row["Close"]),
                        volume=int(row["Volume"]),
                    ))
                except Exception:
                    continue
        except Exception:
            continue

    return results


@app.get("/health")
def health():
    return {"status": "ok"}
