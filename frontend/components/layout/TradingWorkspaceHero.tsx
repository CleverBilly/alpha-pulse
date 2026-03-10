"use client";

import { Tag } from "antd";
import { MARKET_INTERVALS } from "@/types/market";
import { useMarketStore } from "@/store/marketStore";

type TradingWorkspaceMetric = {
  label: string;
  value: string;
};

export default function TradingWorkspaceHero({
  eyebrow,
  title,
  description,
  metrics,
}: {
  eyebrow: string;
  title: string;
  description: string;
  metrics: TradingWorkspaceMetric[];
}) {
  const {
    symbol,
    interval,
    price,
    signal,
    structure,
    liquidity,
    loading,
    error,
    transportMode,
    streamStatus,
    streamError,
    lastUpdatedAt,
    lastRefreshMode,
    refreshDashboard,
    setSymbol,
    setIntervalType,
  } = useMarketStore();

  const transportLabel = formatFeed(streamStatus, transportMode);
  const transportTone = streamStatus === "live" ? "success" : streamStatus === "connecting" ? "processing" : "default";
  const issue = error || streamError;

  return (
    <section className="terminal-hero surface-panel surface-panel--control">
      <div className="terminal-hero__main">
        <div className="terminal-hero__copy">
          <div className="terminal-hero__eyebrow-row">
            <p className="terminal-hero__eyebrow">{eyebrow}</p>
            <Tag color={transportTone}>{transportLabel}</Tag>
            <Tag color={signalTone(signal?.signal)}>{signal?.signal ?? "NEUTRAL"}</Tag>
          </div>

          <h1 className="terminal-hero__title">{title}</h1>
          <p className="terminal-hero__description">
            {issue || description}
          </p>

          <div className="terminal-hero__statusline">
            <StatusChip label="Trend" value={structure?.trend ?? "range"} />
            <StatusChip label="Sweep" value={liquidity?.sweep_type || "none"} />
            <StatusChip label="Refresh" value={formatRefreshMode(lastRefreshMode)} />
            <StatusChip label="Updated" value={formatUpdated(lastUpdatedAt)} />
          </div>
        </div>

        <div className="terminal-hero__quote">
          <div className="terminal-hero__quote-head">
            <span>{symbol}</span>
            <strong>{interval}</strong>
          </div>
          <div className="terminal-hero__price">{loading && !price ? "..." : `$${price?.price.toFixed(2) ?? "-"}`}</div>
          <div className="terminal-hero__quote-sub">
            <span>{transportLabel}</span>
            <span>{signal?.trend_bias ?? "neutral"}</span>
          </div>
        </div>
      </div>

      <div className="terminal-hero__workspace">
        <div className="terminal-hero__controls">
          <div className="terminal-hero__control-box">
            <label htmlFor="terminal-symbol-select" className="terminal-hero__control-label">
              Symbol
            </label>
            <div className="terminal-hero__control-row">
              <select
                id="terminal-symbol-select"
                value={symbol}
                onChange={(e) => setSymbol(e.target.value)}
                className="terminal-hero__select"
              >
                <option value="BTCUSDT">BTCUSDT</option>
                <option value="ETHUSDT">ETHUSDT</option>
              </select>
              <button
                type="button"
                onClick={() => {
                  void refreshDashboard(true);
                }}
                className="terminal-hero__refresh"
              >
                刷新
              </button>
            </div>
          </div>

          <div className="terminal-hero__intervals">
            {MARKET_INTERVALS.map((item) => {
              const active = item === interval;
              return (
                <button
                  key={item}
                  type="button"
                  aria-pressed={active}
                  onClick={() => setIntervalType(item)}
                  className={`terminal-hero__interval ${active ? "terminal-hero__interval--active" : ""}`}
                >
                  {item}
                </button>
              );
            })}
          </div>
        </div>

        <div className="terminal-hero__metrics">
          {metrics.map((metric) => (
            <div key={metric.label} className="terminal-hero__metric">
              <span>{metric.label}</span>
              <strong>{metric.value}</strong>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function StatusChip({ label, value }: { label: string; value: string }) {
  return (
    <div className="terminal-hero__status-chip">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function formatFeed(
  status: "idle" | "connecting" | "live" | "fallback" | "error",
  transport: "idle" | "websocket" | "polling",
) {
  if (status === "live" && transport === "websocket") {
    return "Realtime WS";
  }
  if (status === "connecting") {
    return "Connecting";
  }
  if (transport === "polling") {
    return "HTTP Polling";
  }
  if (status === "error") {
    return "Stream Issue";
  }
  return "Waiting";
}

function signalTone(signal?: string) {
  if (signal === "BUY") {
    return "success";
  }
  if (signal === "SELL") {
    return "error";
  }
  return "default";
}

function formatRefreshMode(mode: "cache" | "force" | null) {
  if (mode === "force") {
    return "Force";
  }
  if (mode === "cache") {
    return "Cache";
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
