"use client";

import { useMemo } from "react";
import { Card, Progress, Tag, Typography } from "antd";
import { formatSignalAction, formatSweepLabel, formatTrendLabel } from "@/lib/uiLabels";
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
    <section>
      <Card
        variant="borderless"
        className="surface-card surface-card--market"
      >
        <div className="flex flex-col gap-6">
          <div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
            <div className="max-w-3xl">
              <p className="text-xs font-semibold uppercase tracking-[0.24em] text-sky-700">
                市场总览
              </p>
              <Typography.Title level={2} className="!mb-0 !mt-3 !text-[30px] !leading-tight !tracking-[-0.04em]">
                {regime.headline}
              </Typography.Title>
              <Typography.Paragraph className="!mb-0 !mt-3 !text-[15px] !leading-7 !text-slate-600">
                当前价格 {formatPrice(price?.price)}，信号倾向为 {regime.stance}。{regime.momentum}，{regime.flowBias}，
                {regime.liquidityPressure}。
              </Typography.Paragraph>
            </div>

            <div className="min-w-[250px] rounded-[28px] border border-white/70 bg-white/76 p-5 shadow-[0_14px_36px_rgba(32,42,63,0.08)]">
              <div className="flex items-center justify-between gap-3">
                <span className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">置信度</span>
                <Tag color={signalTone(signal?.signal).tag}>{formatSignalAction(signal?.signal)}</Tag>
              </div>
              <div className="mt-4 flex items-end justify-between gap-4">
                <p className="text-4xl font-semibold tracking-[-0.04em] text-slate-900">
                  {regime.confidence.toFixed(0)}%
                </p>
                <div className="text-right text-xs text-slate-500">
                  <p>订单流</p>
                  <p className="mt-1 font-medium text-slate-700">{regime.flowBias}</p>
                </div>
              </div>
              <Progress percent={Math.max(0, Math.min(100, regime.confidence))} showInfo={false} className="!mt-4" />
            </div>
          </div>

          <div className="flex flex-wrap gap-2">
            <Tag color="cyan">{regime.stance}</Tag>
            <Tag color="gold">{regime.momentum}</Tag>
            <Tag color={imbalanceTone(liquidity?.order_book_imbalance)}>{regime.liquidityPressure}</Tag>
          </div>

          <div className="grid grid-cols-2 gap-3 xl:grid-cols-4">
            <StatCard label="价格" value={formatPrice(price?.price)} accent="text-sky-700" />
            <StatCard label="信号" value={formatSignalAction(signal?.signal)} accent={signalTone(signal?.signal).text} />
            <StatCard label="趋势" value={formatTrendLabel(structure?.trend)} accent={trendTone(structure?.trend).text} />
            <StatCard label="动量" value={`${(indicator?.rsi ?? 0).toFixed(1)} RSI`} accent="text-violet-700" />
            <StatCard label="净差" value={formatCompact(orderFlow?.delta)} accent="text-emerald-700" />
            <StatCard label="CVD" value={formatCompact(orderFlow?.cvd)} accent="text-slate-700" />
            <StatCard label="盘口失衡" value={formatSigned(liquidity?.order_book_imbalance, 3)} accent="text-amber-700" />
            <StatCard label="扫流动性" value={formatSweepLabel(liquidity?.sweep_type)} accent="text-rose-700" />
          </div>
        </div>
      </Card>
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
    <div className="rounded-[24px] border border-white/70 bg-white/72 p-4 shadow-[0_12px_30px_rgba(32,42,63,0.05)] backdrop-blur">
      <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">{label}</p>
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
    return { text: "text-emerald-700", tag: "success" };
  }
  if (action === "SELL") {
    return { text: "text-rose-700", tag: "error" };
  }
  return { text: "text-slate-700", tag: undefined };
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

function imbalanceTone(value?: number | null) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return undefined;
  }
  if (value >= 0.12) {
    return "success";
  }
  if (value <= -0.12) {
    return "error";
  }
  return "gold";
}
