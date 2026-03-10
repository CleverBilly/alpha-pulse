"use client";

import { Tag, Typography } from "antd";
import { MARKET_INTERVALS } from "@/types/market";
import { useMarketStore } from "@/store/marketStore";

export default function PriceTicker() {
  const {
    symbol,
    interval,
    price,
    loading,
    error,
    transportMode,
    streamStatus,
    lastUpdatedAt,
    lastRefreshMode,
    refreshDashboard,
    setSymbol,
    setIntervalType,
  } = useMarketStore();

  return (
    <section>
      <div className="surface-panel surface-panel--control">
        <div className="flex flex-col gap-6">
          <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
            <div className="max-w-3xl">
              <div className="flex flex-wrap gap-2">
                <Tag color="cyan">Live Feed</Tag>
                <Tag color={error ? "error" : loading ? "processing" : "success"}>
                  {error ? "Sync issue" : loading ? "Syncing" : "Snapshot ready"}
                </Tag>
                <Tag color="gold">BTC / ETH</Tag>
              </div>
              <Typography.Title level={2} className="!mb-0 !mt-4 !text-[28px] !leading-tight !tracking-[-0.03em]">
                Price Ticker
              </Typography.Title>
              <Typography.Paragraph className="!mb-0 !mt-3 !max-w-2xl !text-[15px] !leading-7 !text-slate-600">
                实时跟踪 BTC / ETH 价格，联动多周期快照、信号和结构分析。手动刷新才会强制重建快照。
              </Typography.Paragraph>
            </div>

            <div className="flex w-full flex-col gap-3 lg:w-auto lg:min-w-[280px]">
              <div className="rounded-[24px] border border-white/60 bg-white/72 p-3 shadow-[0_12px_30px_rgba(32,42,63,0.07)] backdrop-blur">
                <label htmlFor="symbol-select" className="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-500">
                  Symbol
                </label>
                <div className="mt-2 flex flex-col gap-3 sm:flex-row">
                  <select
                    id="symbol-select"
                    value={symbol}
                    onChange={(e) => setSymbol(e.target.value)}
                    className="min-h-11 flex-1 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-sm text-slate-800 outline-none transition focus:border-teal-600"
                  >
                    <option value="BTCUSDT">BTCUSDT</option>
                    <option value="ETHUSDT">ETHUSDT</option>
                  </select>
                  <button
                    onClick={() => {
                      void refreshDashboard(true);
                    }}
                    className="min-h-11 rounded-2xl bg-teal-700 px-5 py-2.5 text-sm font-semibold text-white shadow-[0_12px_24px_rgba(15,118,110,0.22)] transition hover:bg-teal-800"
                  >
                    刷新
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-6">
            <TickerMetric label="Last Price" value={loading && !price ? "加载中..." : `$${price?.price.toFixed(2) ?? "-"}`} accent />
            <TickerMetric label="Current Cycle" value={interval} />
            <TickerMetric label="Active Symbol" value={symbol} />
            <TickerMetric label="Feed" value={formatFeed(streamStatus, transportMode)} />
            <TickerMetric label="Refresh Mode" value={formatRefreshMode(lastRefreshMode)} />
            <TickerMetric label="Updated" value={formatUpdated(lastUpdatedAt)} />
          </div>

          <div className="flex flex-wrap gap-2">
            {MARKET_INTERVALS.map((item) => {
              const active = item === interval;
              return (
                <button
                  key={item}
                  onClick={() => setIntervalType(item)}
                  className={`rounded-full px-4 py-2 text-sm font-semibold transition ${
                    active
                      ? "bg-slate-950 text-white shadow-[0_10px_20px_rgba(15,23,42,0.18)]"
                      : "border border-slate-200 bg-white/76 text-slate-700 hover:border-slate-300 hover:text-slate-950"
                  }`}
                >
                  {item}
                </button>
              );
            })}
          </div>

          <p className="text-sm font-medium text-slate-500">当前周期 {interval}</p>

          {error ? <p className="text-sm text-negative">{error}</p> : null}
        </div>
      </div>
    </section>
  );
}

function TickerMetric({
  label,
  value,
  accent = false,
}: {
  label: string;
  value: string;
  accent?: boolean;
}) {
  return (
    <div
      className={`rounded-[24px] border px-4 py-4 ${
        accent
          ? "border-teal-100 bg-[linear-gradient(180deg,rgba(240,253,250,0.95)_0%,rgba(255,255,255,0.95)_100%)]"
          : "border-white/70 bg-white/72"
      }`}
    >
      <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-500">{label}</p>
      <p className="mt-3 text-2xl font-semibold tracking-[-0.02em] text-slate-900">{value}</p>
    </div>
  );
}

function formatRefreshMode(mode: "cache" | "force" | null) {
  if (mode === "force") {
    return "Force rebuild";
  }
  if (mode === "cache") {
    return "Cache-backed";
  }
  return "Waiting";
}

function formatUpdated(timestamp: number | null) {
  if (!timestamp || !Number.isFinite(timestamp)) {
    return "Not synced";
  }

  return new Date(timestamp).toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function formatFeed(
  status: "idle" | "connecting" | "live" | "fallback" | "error",
  transport: "idle" | "websocket" | "polling",
) {
  if (status === "live" && transport === "websocket") {
    return "WebSocket";
  }
  if (status === "connecting") {
    return "Connecting";
  }
  if (transport === "polling") {
    return "HTTP polling";
  }
  if (status === "error") {
    return "Stream issue";
  }
  return "Waiting";
}
