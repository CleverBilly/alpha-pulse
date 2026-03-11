"use client";

import { Tag } from "antd";
import { MARKET_INTERVALS, MARKET_SYMBOLS } from "@/types/market";
import { useMarketStore } from "@/store/marketStore";
import { buildDashboardDecision, buildDirectionCopilotDecision, type DashboardTone } from "./dashboardViewModel";

export default function DecisionHeader() {
  const {
    symbol,
    interval,
    price,
    signal,
    structure,
    liquidity,
    orderFlow,
    lastUpdatedAt,
    loading,
    error,
    transportMode,
    streamStatus,
    streamError,
    directionSnapshots,
    directionLoading,
    directionError,
    lastDirectionUpdatedAt,
    refreshDashboard,
    setSymbol,
    setIntervalType,
  } = useMarketStore();

  const fallbackDecision = buildDashboardDecision({
    signal,
    structure,
    liquidity,
    orderFlow,
  });
  const hasDirectionSnapshots = Boolean(
    directionSnapshots.macro && directionSnapshots.bias && directionSnapshots.trigger && directionSnapshots.execution,
  );
  const decision = hasDirectionSnapshots
    ? buildDirectionCopilotDecision({
        macroSnapshot: directionSnapshots.macro,
        biasSnapshot: directionSnapshots.bias,
        triggerSnapshot: directionSnapshots.trigger,
        executionSnapshot: directionSnapshots.execution,
      })
    : fallbackDecision;
  const issue = error || streamError || directionError;

  return (
    <section className="dashboard-decision surface-panel surface-panel--control" aria-label="Decision Header">
      <div className="dashboard-decision__summary">
        <div className="dashboard-decision__eyebrow-row">
          <p className="dashboard-decision__eyebrow">当前判断</p>
          <Tag color={resolveAntTone(decision.tone)}>{decision.verdict}</Tag>
          <Tag color={decision.tradable ? "success" : "warning"}>{decision.tradeabilityLabel}</Tag>
          <Tag color={streamStatus === "live" ? "success" : streamStatus === "connecting" ? "processing" : "default"}>
            {formatFeed(streamStatus, transportMode)}
          </Tag>
        </div>

        <div className="dashboard-decision__title-row">
          <div>
            <h1 className="dashboard-decision__title">{decision.verdict}</h1>
            <p className="dashboard-decision__description">{issue || decision.summary}</p>
          </div>
          <div className={`dashboard-decision__confidence dashboard-decision__confidence--${decision.tone}`}>
            <span>Confidence</span>
            <strong>{decision.confidence.toFixed(0)}%</strong>
          </div>
        </div>

        <div className="dashboard-decision__chips">
          <MetaChip label="风险" value={decision.riskLabel} tone={decision.tone} />
          <MetaChip label="执行" value={decision.tradeabilityLabel} tone={decision.tone} />
          <MetaChip label="连接" value={formatFeed(streamStatus, transportMode)} />
          <MetaChip label="更新时间" value={formatUpdated(hasDirectionSnapshots ? lastDirectionUpdatedAt : lastUpdatedAt)} />
        </div>

        <div className="dashboard-decision__reasons">
          {decision.reasons.map((reason) => (
            <span key={reason} className="dashboard-decision__reason">
              {reason}
            </span>
          ))}
          {directionLoading && !hasDirectionSnapshots ? (
            <span className="dashboard-decision__reason">方向引擎同步中</span>
          ) : null}
          {decision.timeframeLabels.map((label) => (
            <span key={label} className="dashboard-decision__reason">
              {label}
            </span>
          ))}
        </div>
      </div>

      <div className="dashboard-decision__workspace">
        <div className="dashboard-decision__quote">
          <div className="dashboard-decision__quote-head">
            <span>{symbol}</span>
            <strong>{interval}</strong>
          </div>
          <div className="dashboard-decision__price">{loading && !price ? "..." : `$${price?.price.toFixed(2) ?? "-"}`}</div>
          <p className="dashboard-decision__quote-sub">{signal?.trend_bias ?? "neutral"} bias</p>
        </div>

        <div className="dashboard-decision__controls">
          <div className="dashboard-decision__control-box">
            <label htmlFor="dashboard-symbol-select" className="dashboard-decision__control-label">
              Symbol
            </label>
            <div className="dashboard-decision__control-row">
              <select
                id="dashboard-symbol-select"
                value={symbol}
                onChange={(event) => setSymbol(event.target.value)}
                className="dashboard-decision__select"
                aria-label="Symbol"
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
                className="dashboard-decision__refresh"
              >
                刷新
              </button>
            </div>
          </div>

          <div className="dashboard-decision__intervals" aria-label="Interval switcher">
            {MARKET_INTERVALS.map((item) => {
              const active = item === interval;
              return (
                <button
                  key={item}
                  type="button"
                  aria-pressed={active}
                  onClick={() => setIntervalType(item)}
                  className={`dashboard-decision__interval ${active ? "dashboard-decision__interval--active" : ""}`}
                >
                  {item}
                </button>
              );
            })}
          </div>
        </div>
      </div>
    </section>
  );
}

function MetaChip({
  label,
  value,
  tone = "neutral",
}: {
  label: string;
  value: string;
  tone?: DashboardTone;
}) {
  return (
    <div className={`dashboard-decision__meta dashboard-decision__meta--${tone}`}>
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
  if (transport === "polling" || status === "fallback") {
    return "HTTP Polling";
  }
  if (status === "error") {
    return "Stream Issue";
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

function resolveAntTone(tone: DashboardTone) {
  if (tone === "positive") {
    return "success";
  }
  if (tone === "negative") {
    return "error";
  }
  if (tone === "warning") {
    return "warning";
  }
  return "default";
}
