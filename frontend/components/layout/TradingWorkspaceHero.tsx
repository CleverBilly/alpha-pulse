"use client";

import { formatSignalAction, formatSweepLabel, formatTrendBiasLabel, formatTrendLabel } from "@/lib/uiLabels";
import { MARKET_INTERVALS, MARKET_SYMBOLS } from "@/types/market";
import { useMarketStore } from "@/store/marketStore";
import CommandPage from "./CommandPage";
import OverviewBand from "./OverviewBand";
import CommandPanel from "./CommandPanel";
import RailPanel from "./RailPanel";

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
  const issue = error || streamError;
  const signalLabel = formatSignalAction(signal?.signal);

  return (
    <CommandPage className="terminal-hero surface-panel surface-panel--control">
      <OverviewBand className="terminal-hero__main" data-testid="overview-band">
        <div className="terminal-hero__copy">
          <div className="terminal-hero__signal-strip">
            <p className="terminal-hero__eyebrow">{eyebrow}</p>
            <div className="terminal-hero__signal-tags">
              <SignalTag label="链路" value={transportLabel} tone={transportTone(streamStatus, transportMode)} />
              <SignalTag label="方向" value={signalLabel} tone={signalTone(signal?.signal)} />
            </div>
          </div>

          <h1 className="terminal-hero__title">{title}</h1>
          <p className="terminal-hero__description">
            {issue || description}
          </p>

          <div className="terminal-hero__statusline" data-testid="overview-status-strip">
            <StatusChip label="趋势" value={formatTrendLabel(structure?.trend)} />
            <StatusChip label="扫流动性" value={formatSweepLabel(liquidity?.sweep_type)} />
            <StatusChip label="刷新" value={formatRefreshMode(lastRefreshMode)} />
            <StatusChip label="更新时间" value={formatUpdated(lastUpdatedAt)} />
          </div>
        </div>

        <CommandPanel
          className="terminal-hero__quote"
          data-testid="overview-band-quote"
          data-surface="instrument"
          variant="quote"
        >
          <div className="terminal-hero__quote-kicker">
            <span>主仪表</span>
            <span>{symbol}</span>
          </div>
          <div className="terminal-hero__quote-head">
            <strong>{interval}</strong>
            <span>{formatTrendBiasLabel(signal?.trend_bias)}偏向</span>
          </div>
          <div className="terminal-hero__quote-price-row">
            <div className="terminal-hero__price">{loading && !price ? "..." : `$${price?.price.toFixed(2) ?? "-"}`}</div>
            <div className="terminal-hero__quote-readings">
              <QuoteReading label="链路" value={transportLabel} />
              <QuoteReading label="刷新" value={formatRefreshMode(lastRefreshMode)} />
            </div>
          </div>
        </CommandPanel>
      </OverviewBand>

      <div className="terminal-hero__workspace">
        <CommandPanel
          className="terminal-hero__controls-panel"
          data-testid="command-panel-controls"
          data-surface="console"
          variant="control"
        >
          <div className="terminal-hero__controls-head">
            <div>
              <p className="terminal-hero__controls-eyebrow">控制台</p>
              <h2 className="terminal-hero__controls-title">标的与周期</h2>
            </div>
            <span className="terminal-hero__controls-sub">快速切换执行上下文</span>
          </div>

          <div className="terminal-hero__control-box">
            <label htmlFor="terminal-symbol-select" className="terminal-hero__control-label">
              标的
            </label>
            <div className="terminal-hero__control-row">
              <select
                id="terminal-symbol-select"
                value={symbol}
                onChange={(e) => setSymbol(e.target.value)}
                className="terminal-hero__select"
              >
                {MARKET_SYMBOLS.map((item) => (
                  <option key={item} value={item}>
                    {item}
                  </option>
                ))}
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
        </CommandPanel>

        <RailPanel className="terminal-hero__metrics" data-testid="command-panel-metrics" data-surface="rail">
          {metrics.map((metric) => (
            <div key={metric.label} className="terminal-hero__metric">
              <span>{metric.label}</span>
              <strong>{metric.value}</strong>
            </div>
          ))}
        </RailPanel>
      </div>
    </CommandPage>
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

function SignalTag({
  label,
  value,
  tone,
}: {
  label: string;
  value: string;
  tone: "accent" | "neutral" | "danger";
}) {
  return (
    <div className={`terminal-hero__signal-tag terminal-hero__signal-tag--${tone}`}>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function QuoteReading({ label, value }: { label: string; value: string }) {
  return (
    <div className="terminal-hero__quote-reading">
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
    return "实时 WS";
  }
  if (status === "connecting") {
    return "连接中";
  }
  if (transport === "polling") {
    return "HTTP 轮询";
  }
  if (status === "error") {
    return "推流异常";
  }
  return "等待中";
}

function signalTone(signal?: string): "accent" | "neutral" | "danger" {
  if (signal === "BUY") {
    return "accent";
  }
  if (signal === "SELL") {
    return "danger";
  }
  return "neutral";
}

function transportTone(
  status: "idle" | "connecting" | "live" | "fallback" | "error",
  transport: "idle" | "websocket" | "polling",
): "accent" | "neutral" | "danger" {
  if (status === "live" && transport === "websocket") {
    return "accent";
  }
  if (status === "error") {
    return "danger";
  }
  return "neutral";
}

function formatRefreshMode(mode: "cache" | "force" | null) {
  if (mode === "force") {
    return "强制";
  }
  if (mode === "cache") {
    return "缓存";
  }
  return "等待中";
}

function formatUpdated(timestamp: number | null) {
  if (!timestamp || !Number.isFinite(timestamp)) {
    return "未同步";
  }

  return new Date(timestamp).toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}
