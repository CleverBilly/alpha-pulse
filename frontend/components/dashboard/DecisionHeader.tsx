"use client";

import { formatTrendBiasLabel } from "@/lib/uiLabels";
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
    <section
      className="dashboard-decision command-panel command-panel--quote surface-panel surface-panel--control"
      aria-label="决策头部"
      data-panel-role="overview"
      data-surface="analysis-deck"
      data-testid="dashboard-overview-summary"
    >
      <div className="dashboard-decision__grid">
        <div className="dashboard-decision__main-stage">
          <div className="dashboard-decision__summary">
            <div className="dashboard-decision__signalbar" data-testid="dashboard-decision-strip">
              <p className="dashboard-decision__eyebrow">当前判断</p>
              <div className="dashboard-decision__signalset">
                <SignalCell label="判定" value={decision.verdict} tone={decision.tone} />
                <SignalCell label="执行" value={decision.tradeabilityLabel} tone={decision.tradable ? "positive" : "warning"} />
                <SignalCell label="链路" value={formatFeed(streamStatus, transportMode)} />
              </div>
            </div>

            <div className="dashboard-decision__title-row">
              <div className="dashboard-decision__title-copy">
                <h1 className="dashboard-decision__title">{decision.verdict}</h1>
                <p className="dashboard-decision__description">{issue || decision.summary}</p>
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
            </div>

            <div className="dashboard-decision__timeframes">
              {decision.timeframeLabels.map((label) => (
                <span key={label} className="dashboard-decision__timeframe">
                  {label}
                </span>
              ))}
            </div>
          </div>
        </div>

        <div className="dashboard-decision__instrument-rail" data-testid="dashboard-instrument-rail">
          <div
            className={`dashboard-decision__confidence command-panel command-panel--status dashboard-decision__confidence--${decision.tone}`}
          >
            <span>置信度</span>
            <strong>{decision.confidence.toFixed(0)}%</strong>
          </div>

          <div
            className="dashboard-decision__quote command-panel command-panel--quote"
            data-surface="instrument"
            data-testid="dashboard-quote-panel"
            role="region"
            aria-label="市场报价"
          >
            <div className="dashboard-decision__quote-kicker">
              <span>主报价板</span>
              <span>{symbol}</span>
            </div>
            <div className="dashboard-decision__quote-head">
              <strong>{interval}</strong>
              <span>{formatTrendBiasLabel(signal?.trend_bias)}偏向</span>
            </div>
            <div className="dashboard-decision__quote-body">
              <div className="dashboard-decision__price">{loading && !price ? "..." : `$${price?.price.toFixed(2) ?? "-"}`}</div>
              <div className="dashboard-decision__quote-sub">
                <QuoteCell label="链路" value={formatFeed(streamStatus, transportMode)} />
                <QuoteCell label="置信度" value={`${decision.confidence.toFixed(0)}%`} />
              </div>
            </div>
          </div>

          <div
            className="dashboard-decision__controls command-panel command-panel--control"
            data-surface="console"
            data-testid="dashboard-action-cluster"
            role="region"
            aria-label="交易工作台控件"
          >
            <div className="dashboard-decision__controls-head">
              <div>
                <p className="dashboard-decision__controls-eyebrow">控制台</p>
                <h2 className="dashboard-decision__controls-title">标的与周期</h2>
              </div>
              <span className="dashboard-decision__controls-sub">快速切换交易上下文</span>
            </div>

            <div className="dashboard-decision__control-box">
              <label htmlFor="dashboard-symbol-select" className="dashboard-decision__control-label">
                标的
              </label>
              <div className="dashboard-decision__control-row">
                <select
                  id="dashboard-symbol-select"
                  value={symbol}
                  onChange={(event) => setSymbol(event.target.value)}
                  className="dashboard-decision__select"
                  aria-label="标的"
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

            <div className="dashboard-decision__intervals" aria-label="周期切换">
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
      </div>
    </section>
  );
}

function SignalCell({
  label,
  value,
  tone = "neutral",
}: {
  label: string;
  value: string;
  tone?: DashboardTone;
}) {
  return (
    <div className={`dashboard-decision__signal-cell dashboard-decision__signal-cell--${tone}`}>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
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

function QuoteCell({ label, value }: { label: string; value: string }) {
  return (
    <div className="dashboard-decision__quote-cell">
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
  if (transport === "polling" || status === "fallback") {
    return "HTTP 轮询";
  }
  if (status === "error") {
    return "数据流异常";
  }
  return "等待同步";
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
