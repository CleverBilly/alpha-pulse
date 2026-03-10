"use client";

import { useMemo } from "react";
import { Card, Typography } from "antd";
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
    <section>
      <Card
        variant="borderless"
        className="surface-card surface-card--analysis"
      >
        <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.26em] text-amber-700">
              AI Analysis
            </p>
            <Typography.Title level={3} className="!mb-0 !mt-3 !text-[28px] !tracking-[-0.04em]">
              Decision Memo
            </Typography.Title>
            <p className="mt-2 max-w-2xl text-sm leading-6 text-slate-600">
              {signal?.explain ?? "当前还没有可用的 AI 分析结果。"}
            </p>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <SummaryPill label="Bias" value={signal?.trend_bias ?? "neutral"} tone={biasTone(signal?.trend_bias)} />
            <SummaryPill
              label="R/R"
              value={signal ? signal.risk_reward.toFixed(2) : "-"}
              tone="bg-violet-50 text-violet-700 border-violet-100"
            />
          </div>
        </div>

        <div className="mt-6 grid grid-cols-1 gap-6 xl:grid-cols-[1.05fr_0.95fr]">
          <div className="space-y-6">
            <PlaybookStrip
              rows={[
                {
                  label: "Entry Window",
                  value: signal ? signal.entry_price.toFixed(2) : "-",
                  detail: signal ? "围绕当前方向等待确认后再执行。" : "等待新的信号快照。",
                },
                {
                  label: "Invalidation",
                  value: signal ? signal.stop_loss.toFixed(2) : "-",
                  detail: structure?.trend === "uptrend" ? "上行结构失守即降级。" : "结构未确认前避免追单。",
                },
                {
                  label: "Target",
                  value: signal ? signal.target_price.toFixed(2) : "-",
                  detail: liquidity?.sweep_type ? `重点观察 ${liquidity.sweep_type} 后续延续。` : "关注流动性回收后的方向选择。",
                },
              ]}
            />

            <InsightBlock
              title="Bullish Drivers"
              emptyText="当前没有明显的多头共振因子。"
              tone="bg-emerald-50 text-emerald-700 border-emerald-100"
              factors={positives.slice(0, 4).map((factor) => ({
                name: factor.name,
                detail: factor.reason,
                score: factor.score,
              }))}
            />

            <InsightBlock
              title="Risk Factors"
              emptyText="当前没有显著的反向风险因子。"
              tone="bg-rose-50 text-rose-700 border-rose-100"
              factors={negatives.slice(0, 4).map((factor) => ({
                name: factor.name,
                detail: factor.reason,
                score: factor.score,
              }))}
            />
          </div>

          <div className="space-y-6">
            <ContextPanel
              title="Execution Plan"
              rows={[
                { label: "Execution Bias", value: executionBias },
                { label: "Trend", value: structure?.trend ?? "range" },
                { label: "Sweep", value: liquidity?.sweep_type || "none" },
                { label: "Absorption", value: orderFlow?.absorption_bias || "none" },
                { label: "Iceberg", value: orderFlow?.iceberg_bias || "none" },
                { label: "RSI", value: indicator ? indicator.rsi.toFixed(2) : "-" },
              ]}
            />

            <ContextPanel
              title="Recent Signal Tape"
              rows={
                latestTimeline.length > 0
                  ? latestTimeline.map((point) => ({
                      label: `${formatSignalTime(point.open_time)} ${point.signal}`,
                      value: `score ${point.score} / conf ${point.confidence}%`,
                    }))
                  : [{ label: "History", value: "暂无历史信号" }]
              }
            />

            <ContextPanel
              title="Microstructure Tape"
              rows={
                latestMicroEvents.length > 0
                  ? latestMicroEvents.map((event) => ({
                      label: `${formatEventType(event.type)} ${formatEventTime(event.trade_time)}`,
                      value: `${event.bias} / ${event.detail}`,
                    }))
                  : [{ label: "Microstructure", value: "暂无微结构事件" }]
              }
            />
          </div>
        </div>
      </Card>
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
  tone: string;
}) {
  return (
    <div className={`rounded-2xl border px-4 py-3 ${tone}`}>
      <p className="text-[11px] uppercase tracking-[0.18em]">{label}</p>
      <p className="mt-2 text-lg font-semibold">{value}</p>
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
  tone: string;
  factors: Array<{ name: string; detail: string; score: number }>;
}) {
  return (
    <div className="rounded-[26px] border border-slate-100 bg-white/85 p-5 backdrop-blur">
      <div className="flex items-center justify-between gap-3">
        <h4 className="text-sm font-semibold text-slate-900">{title}</h4>
        <span className="text-xs text-slate-500">{factors.length} items</span>
      </div>

      <div className="mt-4 space-y-3">
        {factors.map((factor) => (
          <div key={`${factor.name}-${factor.score}`} className={`rounded-2xl border px-4 py-3 ${tone}`}>
            <div className="flex items-center justify-between gap-3">
              <span className="text-sm font-semibold">{factor.name}</span>
              <span className="rounded-full bg-white/70 px-2 py-0.5 text-xs font-semibold">
                {factor.score > 0 ? `+${factor.score}` : factor.score}
              </span>
            </div>
            <p className="mt-2 text-xs leading-5 text-slate-700">{factor.detail}</p>
          </div>
        ))}

        {factors.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500">
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
    <div className="grid gap-3 md:grid-cols-3">
      {rows.map((row) => (
        <div
          key={row.label}
          className="rounded-[24px] border border-slate-100 bg-white/82 px-4 py-4 shadow-[0_12px_30px_rgba(32,42,63,0.05)]"
        >
          <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-500">{row.label}</p>
          <p className="mt-2 text-xl font-semibold tracking-[-0.03em] text-slate-900">{row.value}</p>
          <p className="mt-2 text-xs leading-6 text-slate-600">{row.detail}</p>
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
    <div className="rounded-[26px] border border-slate-100 bg-slate-950 p-5 text-slate-100">
      <h4 className="text-sm font-semibold">{title}</h4>
      <div className="mt-4 space-y-3">
        {rows.map((row) => (
          <div
            key={`${row.label}-${row.value}`}
            className="flex items-start justify-between gap-4 rounded-2xl border border-white/10 bg-white/5 px-4 py-3"
          >
            <span className="text-xs uppercase tracking-[0.18em] text-slate-400">{row.label}</span>
            <span className="max-w-[62%] text-right text-sm text-slate-100">{row.value}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function biasTone(bias?: string) {
  if (bias === "bullish") {
    return "bg-emerald-50 text-emerald-700 border-emerald-100";
  }
  if (bias === "bearish") {
    return "bg-rose-50 text-rose-700 border-rose-100";
  }
  return "bg-slate-50 text-slate-700 border-slate-200";
}

function formatSignalTime(timestamp: number) {
  return new Date(timestamp).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatEventType(value: string) {
  return value
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function formatEventTime(timestamp: number) {
  return new Date(timestamp).toLocaleString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}
