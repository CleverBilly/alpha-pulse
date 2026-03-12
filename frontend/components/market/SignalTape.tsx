"use client";

import { Progress, Tag } from "antd";
import { formatSignalAction, formatTrendLabel } from "@/lib/uiLabels";
import { useMarketStore } from "@/store/marketStore";

export default function SignalTape() {
  const { signalTimeline, signal, structure } = useMarketStore();

  const points = signalTimeline.length > 0 ? signalTimeline.slice(-8).reverse() : [];
  const featuredScore = signal?.score ?? points[0]?.score ?? 0;
  const featuredConfidence = signal?.confidence ?? points[0]?.confidence ?? 0;

  return (
    <section className="surface-panel surface-panel--dark text-slate-100">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.14em] text-sky-300">
            信号序列
          </p>
          <h3 className="mt-2 text-xl font-semibold">近期决策</h3>
        </div>
        {signal ? (
          <span className={`rounded-full px-3 py-1 text-xs font-semibold ${signalTone(signal.signal)}`}>
            {formatSignalAction(signal.signal)} {signal.confidence.toFixed(0)}%
          </span>
        ) : null}
      </div>

      <div className="mt-5 rounded-[26px] border border-white/10 bg-white/6 p-5 backdrop-blur">
        <div className="flex flex-wrap items-center gap-2">
          <Tag color="cyan">实时判断</Tag>
          <Tag color={structureTone(structure?.trend)}>{formatTrendLabel(structure?.trend)}</Tag>
          <Tag color="gold">评分 {featuredScore}</Tag>
        </div>

        <div className="mt-4 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <p className={`text-[13px] font-semibold uppercase tracking-[0.12em] ${signalLabelTone(signal?.signal ?? "NEUTRAL")}`}>
              {formatSignalAction(signal?.signal)}
            </p>
            <p className="mt-2 text-3xl font-semibold tracking-[-0.04em] text-white">
              {featuredConfidence.toFixed(0)}%
            </p>
            <p className="mt-2 text-sm text-slate-300">
              {signal
                ? `当前执行偏向${formatSignalAction(signal.signal)}，关注 ${signal.entry_price.toFixed(2)} 一线。`
                : "当前还没有可用的执行建议。"}
            </p>
          </div>

          <div className="grid grid-cols-3 gap-3 text-sm lg:min-w-[360px]">
            <TapeMetric label="进场位" value={signal ? signal.entry_price.toFixed(2) : "-"} />
            <TapeMetric label="目标位" value={signal ? signal.target_price.toFixed(2) : "-"} />
            <TapeMetric label="止损位" value={signal ? signal.stop_loss.toFixed(2) : "-"} />
          </div>
        </div>

        <Progress
          percent={Math.max(0, Math.min(100, featuredConfidence))}
          showInfo={false}
          strokeColor="#38bdf8"
          railColor="rgba(255,255,255,0.1)"
          className="!mt-4"
        />
      </div>

      <div className="mt-5 space-y-3">
        {points.map((point, index) => (
          <div key={`${point.open_time}-${point.id}`} className="relative pl-6">
            {index !== points.length - 1 ? (
              <span className="absolute left-[7px] top-9 h-[calc(100%-0.25rem)] w-px bg-white/10" />
            ) : null}
            <span className={`absolute left-0 top-5 h-4 w-4 rounded-full ring-4 ring-[#101826] ${signalDotTone(point.signal)}`} />
            <div className="grid grid-cols-1 gap-3 rounded-[22px] border border-white/10 bg-white/5 px-4 py-3 md:grid-cols-[0.8fr_1fr_1fr]">
              <div>
                <p className={`text-sm font-semibold ${signalLabelTone(point.signal)}`}>{formatSignalAction(point.signal)}</p>
                <p className="mt-1 text-xs text-slate-400">{formatTime(point.open_time)}</p>
              </div>
              <div className="text-sm">
                <p className="text-slate-400">进场位</p>
                <p className="mt-1 font-medium text-white">{point.entry_price.toFixed(2)}</p>
              </div>
              <div className="text-sm md:text-right">
                <p className="text-slate-400">评分 / 置信度</p>
                <p className="mt-1 font-medium text-white">
                  {point.score} / {point.confidence}%
                </p>
              </div>
            </div>
          </div>
        ))}

        {points.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-white/15 bg-white/5 px-4 py-6 text-sm text-slate-400">
            暂无历史信号
          </div>
        ) : null}
      </div>
    </section>
  );
}

function TapeMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-[18px] border border-white/10 bg-white/6 px-3 py-3">
      <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-400">{label}</p>
      <p className="mt-2 text-base font-semibold text-white">{value}</p>
    </div>
  );
}

function signalTone(action: string) {
  if (action === "BUY") {
    return "bg-emerald-500/15 text-emerald-300";
  }
  if (action === "SELL") {
    return "bg-rose-500/15 text-rose-300";
  }
  return "bg-slate-500/15 text-slate-300";
}

function signalLabelTone(action: string) {
  if (action === "BUY") {
    return "text-emerald-300";
  }
  if (action === "SELL") {
    return "text-rose-300";
  }
  return "text-slate-200";
}

function signalDotTone(action: string) {
  if (action === "BUY") {
    return "bg-emerald-400";
  }
  if (action === "SELL") {
    return "bg-rose-400";
  }
  return "bg-slate-300";
}

function structureTone(trend?: string) {
  if (trend === "uptrend") {
    return "success";
  }
  if (trend === "downtrend") {
    return "error";
  }
  return "default";
}

function formatTime(timestamp: number) {
  return new Date(timestamp).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}
