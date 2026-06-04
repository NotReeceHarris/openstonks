from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import yfinance as yf
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

    for symbol in symbol_list:
        try:
            price = yf.Ticker(symbol).fast_info.last_price
            if price is not None:
                results.append(LiveQuote(
                    symbol=symbol,
                    price=float(price),
                    updated_at=now,
                    source="yfinance",
                ))
        except Exception:
            pass

    return results


@app.get("/health")
def health():
    return {"status": "ok"}
