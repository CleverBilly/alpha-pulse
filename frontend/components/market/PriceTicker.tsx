"use client";

import { MARKET_INTERVALS } from "@/types/market";
import { useMarketStore } from "@/store/marketStore";

export default function PriceTicker() {
  const {
    symbol,
    interval,
    price,
    loading,
    error,
    refreshDashboard,
    setSymbol,
    setIntervalType,
  } = useMarketStore();

  return (
    <section className="rounded-2xl bg-panel p-4 shadow-panel">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h2 className="text-lg font-semibold">Price Ticker</h2>
          <p className="text-sm text-muted">实时跟踪 BTC / ETH 价格与多周期分析</p>
        </div>

        <div className="flex flex-col gap-3 lg:items-end">
          <div className="flex flex-wrap items-center gap-3">
            <select
              value={symbol}
              onChange={(e) => setSymbol(e.target.value)}
              className="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm"
            >
              <option value="BTCUSDT">BTCUSDT</option>
              <option value="ETHUSDT">ETHUSDT</option>
            </select>
            <button
              onClick={() => {
                void refreshDashboard();
              }}
              className="rounded-lg bg-accent px-3 py-2 text-sm font-medium text-white"
            >
              刷新
            </button>
          </div>

          <div className="flex flex-wrap gap-2">
            {MARKET_INTERVALS.map((item) => {
              const active = item === interval;
              return (
                <button
                  key={item}
                  onClick={() => setIntervalType(item)}
                  className={`rounded-full px-3 py-1 text-xs font-semibold transition ${
                    active
                      ? "bg-slate-900 text-white"
                      : "border border-slate-200 bg-white text-slate-600"
                  }`}
                >
                  {item}
                </button>
              );
            })}
          </div>
        </div>
      </div>

      <div className="mt-4 flex items-end justify-between gap-4">
        <div className="text-3xl font-bold">
          {loading && !price ? "加载中..." : `$${price?.price.toFixed(2) ?? "-"}`}
        </div>
        <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-600">
          当前周期 {interval}
        </span>
      </div>

      {error ? <p className="mt-2 text-sm text-negative">{error}</p> : null}
    </section>
  );
}
