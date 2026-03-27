"use client";

import { useMemo } from "react";
import {
  formatAbsorptionBiasLabel,
  formatFactorName,
  formatIcebergBiasLabel,
  formatMicrostructureEventTypeLabel,
  formatSignalAction,
  formatSweepLabel,
  formatTrendLabel,
  formatTrendBiasLabel,
} from "@/lib/uiLabels";
import { useMarketStore } from "@/store/marketStore";

export default function AIAnalysisPanel() {
  const { signal, signalTimeline, structure, liquidity, orderFlow, indicator, microstructureEvents = [] } =
    useMarketStore();

  const positives = useMemo(
    () => (signal?.factors ?? []).filter((factor) => factor.score > 0).sort((a, b) => b.score - a.score),
    [signal?.factors],
  );
  const negatives = useMemo(
    () => (signal?.factors ?? []).filter((factor) => factor.score < 0).sort((a, b) => a.score - b.score),
    [signal?.factors],
  );
  const latestTimeline = signalTimeline.slice(-4).reverse();
  const latestMicroEvents = (
    microstructureEvents.length > 0 ? microstructureEvents : (orderFlow?.microstructure_events ?? [])
  )
    .slice(-4)
    .reverse();

  const executionBias = useMemo(() => {
    if (!signal) {
      return "暂无分析";
    }

    if (signal.signal === "BUY") {
      return "等待回踩确认后偏向多头执行";
    }
    if (signal.signal === "SELL") {
      return "等待反弹衰竭后偏向空头执行";
    }
    return "当前更适合等待结构突破再执行";
  }, [signal]);

  return (
    <section className="ai-analysis" data-surface="analysis-deck" data-testid="ai-analysis-panel">
      <div className="ai-analysis__header">
        <div className="ai-analysis__copy">
          <p className="ai-analysis__eyebrow">AI 分析</p>
          <h3 className="ai-analysis__title">决策备忘</h3>
          <p className="ai-analysis__description">
            {signal?.explain ?? "当前还没有可用的 AI 分析结果。"}
          </p>
        </div>

        <div className="ai-analysis__summary">
          <SummaryPill
            label="方向"
            value={formatTrendBiasLabel(signal?.trend_bias)}
            tone={biasTone(signal?.trend_bias)}
          />
          <SummaryPill
            label="R/R"
            value={signal ? signal.risk_reward.toFixed(2) : "-"}
            tone="neutral"
          />
        </div>
      </div>

      <div className="ai-analysis__workspace">
        <div className="ai-analysis__primary">
          <PlaybookStrip
            rows={[
              {
                label: "进场区间",
                value: signal ? signal.entry_price.toFixed(2) : "-",
                detail: signal ? "围绕当前方向等待确认后再执行。" : "等待新的信号快照。",
              },
              {
                label: "失效位",
                value: signal ? signal.stop_loss.toFixed(2) : "-",
                detail: structure?.trend === "uptrend" ? "上行结构失守即降级。" : "结构未确认前避免追单。",
              },
              {
                label: "目标位",
                value: signal ? signal.target_price.toFixed(2) : "-",
                detail: liquidity?.sweep_type
                  ? `重点观察 ${formatSweepLabel(liquidity.sweep_type)} 后续是否延续。`
                  : "关注流动性回收后的方向选择。",
              },
            ]}
          />

          <InsightBlock
            title="多头驱动"
            emptyText="当前没有明显的多头共振因子。"
            tone="accent"
            factors={positives.slice(0, 4).map((factor) => ({
              name: formatFactorName(factor.name),
              detail: factor.reason,
              score: factor.score,
            }))}
          />

          <InsightBlock
            title="风险因子"
            emptyText="当前没有显著的反向风险因子。"
            tone="danger"
            factors={negatives.slice(0, 4).map((factor) => ({
              name: formatFactorName(factor.name),
              detail: factor.reason,
              score: factor.score,
            }))}
          />
        </div>

        <div className="ai-analysis__secondary" data-testid="ai-analysis-sequences">
          <ContextPanel
            title="执行计划"
            rows={[
              { label: "执行倾向", value: executionBias },
              { label: "趋势", value: formatTrendLabel(structure?.trend) },
              { label: "扫流动性", value: formatSweepLabel(liquidity?.sweep_type) },
              { label: "吸收", value: formatAbsorptionBiasLabel(orderFlow?.absorption_bias) },
              { label: "冰山", value: formatIcebergBiasLabel(orderFlow?.iceberg_bias) },
              { label: "RSI", value: indicator ? indicator.rsi.toFixed(2) : "-" },
            ]}
          />

          <ContextPanel
            title="近期信号序列"
            rows={
              latestTimeline.length > 0
                ? latestTimeline.map((point) => ({
                    label: `${formatSignalTime(point.open_time)} ${formatSignalAction(point.signal)}`,
                    value: `评分 ${point.score} / 置信度 ${point.confidence}%`,
                  }))
                : [{ label: "历史信号", value: "暂无历史信号" }]
            }
          />

          <ContextPanel
            title="微结构序列"
            rows={
              latestMicroEvents.length > 0
                ? latestMicroEvents.map((event) => ({
                    label: `${formatMicrostructureEventTypeLabel(event.type)} ${formatEventTime(event.trade_time)}`,
                    value: `${formatTrendBiasLabel(event.bias)} / ${event.detail}`,
                  }))
                : [{ label: "微结构", value: "暂无微结构事件" }]
            }
          />
        </div>
      </div>
    </section>
  );
}

function SummaryPill({
  label,
  value,
  tone,
}: {
  label: string;
  value: string;
  tone: "accent" | "neutral" | "danger";
}) {
  return (
    <div className={`ai-analysis__summary-pill ai-analysis__summary-pill--${tone}`}>
      <p>{label}</p>
      <strong>{value}</strong>
    </div>
  );
}

function InsightBlock({
  title,
  emptyText,
  tone,
  factors,
}: {
  title: string;
  emptyText: string;
  tone: "accent" | "danger";
  factors: Array<{ name: string; detail: string; score: number }>;
}) {
  return (
    <div className="ai-analysis__insight">
      <div className="ai-analysis__block-head">
        <h4>{title}</h4>
        <span>{factors.length} 条</span>
      </div>

      <div className="ai-analysis__insight-list">
        {factors.map((factor) => (
          <div key={`${factor.name}-${factor.score}`} className={`ai-analysis__factor ai-analysis__factor--${tone}`}>
            <div className="ai-analysis__factor-head">
              <span>{factor.name}</span>
              <strong>{factor.score > 0 ? `+${factor.score}` : factor.score}</strong>
            </div>
            <p>{factor.detail}</p>
          </div>
        ))}

        {factors.length === 0 ? (
          <div className="ai-analysis__empty">
            {emptyText}
          </div>
        ) : null}
      </div>
    </div>
  );
}

function PlaybookStrip({
  rows,
}: {
  rows: Array<{ label: string; value: string; detail: string }>;
}) {
  return (
    <div className="ai-analysis__playbook" data-testid="ai-analysis-playbook">
      {rows.map((row) => (
        <div key={row.label} className="ai-analysis__playbook-slot">
          <p>{row.label}</p>
          <strong>{row.value}</strong>
          <span>{row.detail}</span>
        </div>
      ))}
    </div>
  );
}

function ContextPanel({
  title,
  rows,
}: {
  title: string;
  rows: Array<{ label: string; value: string }>;
}) {
  return (
    <div className="ai-analysis__context-panel">
      <h4>{title}</h4>
      <div className="ai-analysis__context-rows">
        {rows.map((row) => (
          <div key={`${row.label}-${row.value}`} className="ai-analysis__context-row">
            <span>{row.label}</span>
            <strong>{row.value}</strong>
          </div>
        ))}
      </div>
    </div>
  );
}

function biasTone(bias?: string): "accent" | "neutral" | "danger" {
  if (bias === "bullish") {
    return "accent";
  }
  if (bias === "bearish") {
    return "danger";
  }
  return "neutral";
}

function formatSignalTime(timestamp: number) {
  return new Date(timestamp).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatEventTime(timestamp: number) {
  return new Date(timestamp).toLocaleString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}
