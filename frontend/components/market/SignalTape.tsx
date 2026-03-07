"use client";

import { useMarketStore } from "@/store/marketStore";

export default function SignalTape() {
  const { signalTimeline, signal } = useMarketStore();

  const points = signalTimeline.length > 0 ? signalTimeline.slice(-8).reverse() : [];

  return (
    <section className="rounded-[28px] border border-slate-200/80 bg-[linear-gradient(180deg,#0f172a_0%,#111827_100%)] p-6 text-slate-100 shadow-panel">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.22em] text-sky-300">
            Signal Tape
          </p>
          <h3 className="mt-2 text-xl font-semibold">Recent Decisions</h3>
        </div>
        {signal ? (
          <span className={`rounded-full px-3 py-1 text-xs font-semibold ${signalTone(signal.signal)}`}>
            {signal.signal} {signal.confidence.toFixed(0)}%
          </span>
        ) : null}
      </div>

      <div className="mt-5 space-y-3">
        {points.map((point) => (
          <div
            key={`${point.open_time}-${point.id}`}
            className="grid grid-cols-[0.7fr_1fr_1fr] items-center gap-3 rounded-2xl border border-white/10 bg-white/5 px-4 py-3"
          >
            <div>
              <p className={`text-sm font-semibold ${signalLabelTone(point.signal)}`}>{point.signal}</p>
              <p className="mt-1 text-xs text-slate-400">{formatTime(point.open_time)}</p>
            </div>
            <div className="text-sm">
              <p className="text-slate-400">Entry</p>
              <p className="mt-1 font-medium text-white">{point.entry_price.toFixed(2)}</p>
            </div>
            <div className="text-right text-sm">
              <p className="text-slate-400">Score / Conf.</p>
              <p className="mt-1 font-medium text-white">
                {point.score} / {point.confidence}%
              </p>
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

function formatTime(timestamp: number) {
  return new Date(timestamp).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}
