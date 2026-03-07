"use client";

import { useMemo } from "react";
import { useMarketStore } from "@/store/marketStore";

export default function MarketOverviewBoard() {
  const { price, indicator, orderFlow, structure, liquidity, signal } = useMarketStore();

  const regime = useMemo(() => {
    const trend = structure?.trend ?? "range";
    const score = signal?.score ?? 0;
    const confidence = signal?.confidence ?? 0;
    const rsi = indicator?.rsi ?? 50;
    const imbalance = liquidity?.order_book_imbalance ?? 0;
    const delta = orderFlow?.delta ?? 0;

    const headline =
      trend === "uptrend"
        ? "趋势延续偏多"
        : trend === "downtrend"
          ? "趋势延续偏空"
          : "区间整理";

    const stance =
      score >= 35
        ? "顺势做多优先"
        : score <= -35
          ? "顺势做空优先"
          : "等待确认";

    const liquidityPressure =
      imbalance >= 0.12
        ? "买盘墙更厚"
        : imbalance <= -0.12
          ? "卖盘墙更厚"
          : "盘口相对平衡";

    const flowBias =
      delta > 0
        ? "主动买盘占优"
        : delta < 0
          ? "主动卖盘占优"
          : "主动力度均衡";

    const momentum =
      rsi >= 60 ? "动量强势" : rsi <= 40 ? "动量偏弱" : "动量中性";

    return {
      headline,
      stance,
      liquidityPressure,
      flowBias,
      momentum,
      confidence,
    };
  }, [indicator?.rsi, liquidity?.order_book_imbalance, orderFlow?.delta, signal?.confidence, signal?.score, structure?.trend]);

  return (
    <section className="rounded-[28px] border border-slate-200/80 bg-[linear-gradient(135deg,#ffffff_0%,#eff6ff_48%,#ecfeff_100%)] p-6 shadow-panel">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div className="max-w-2xl">
          <p className="text-xs font-semibold uppercase tracking-[0.24em] text-sky-700">
            Market Overview
          </p>
          <h2 className="mt-2 text-2xl font-semibold text-slate-900">{regime.headline}</h2>
          <p className="mt-2 text-sm leading-6 text-slate-600">
            当前价格 {formatPrice(price?.price)}，信号倾向为 {regime.stance}。{regime.momentum}，{regime.flowBias}，
            {regime.liquidityPressure}。
          </p>
        </div>

        <div className="rounded-2xl border border-white/70 bg-white/80 px-4 py-3 backdrop-blur">
          <p className="text-xs text-slate-500">Confidence</p>
          <p className="mt-1 text-3xl font-semibold text-slate-900">{regime.confidence.toFixed(0)}%</p>
        </div>
      </div>

      <div className="mt-6 grid grid-cols-2 gap-3 xl:grid-cols-4">
        <StatCard label="Price" value={formatPrice(price?.price)} accent="text-sky-700" />
        <StatCard label="Signal" value={signal?.signal ?? "NEUTRAL"} accent={signalTone(signal?.signal).text} />
        <StatCard label="Trend" value={structure?.trend ?? "range"} accent={trendTone(structure?.trend).text} />
        <StatCard label="Momentum" value={`${(indicator?.rsi ?? 0).toFixed(1)} RSI`} accent="text-violet-700" />
        <StatCard label="Delta" value={formatCompact(orderFlow?.delta)} accent="text-emerald-700" />
        <StatCard label="CVD" value={formatCompact(orderFlow?.cvd)} accent="text-slate-700" />
        <StatCard
          label="Order Book"
          value={formatSigned(liquidity?.order_book_imbalance, 3)}
          accent="text-amber-700"
        />
        <StatCard label="Sweep" value={liquidity?.sweep_type || "none"} accent="text-rose-700" />
      </div>
    </section>
  );
}

function StatCard({
  label,
  value,
  accent,
}: {
  label: string;
  value: string;
  accent: string;
}) {
  return (
    <div className="rounded-2xl border border-slate-200/70 bg-white/75 p-4 backdrop-blur">
      <p className="text-xs uppercase tracking-[0.16em] text-slate-500">{label}</p>
      <p className={`mt-2 text-lg font-semibold ${accent}`}>{value}</p>
    </div>
  );
}

function formatPrice(value?: number | null) {
  return typeof value === "number" && Number.isFinite(value) ? `$${value.toFixed(2)}` : "-";
}

function formatCompact(value?: number | null) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return "-";
  }
  return new Intl.NumberFormat("zh-CN", {
    notation: "compact",
    maximumFractionDigits: 2,
  }).format(value);
}

function formatSigned(value?: number | null, digits = 2) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return "-";
  }
  const prefix = value > 0 ? "+" : "";
  return `${prefix}${value.toFixed(digits)}`;
}

function signalTone(action?: string) {
  if (action === "BUY") {
    return { text: "text-emerald-700" };
  }
  if (action === "SELL") {
    return { text: "text-rose-700" };
  }
  return { text: "text-slate-700" };
}

function trendTone(trend?: string) {
  if (trend === "uptrend") {
    return { text: "text-emerald-700" };
  }
  if (trend === "downtrend") {
    return { text: "text-rose-700" };
  }
  return { text: "text-slate-700" };
}
