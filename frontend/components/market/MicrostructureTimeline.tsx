"use client";

import { useMemo } from "react";
import { useMarketStore } from "@/store/marketStore";

export default function MicrostructureTimeline() {
  const { microstructureEvents = [], orderFlow } = useMarketStore();
  const events = useMemo(
    () =>
      (microstructureEvents.length > 0 ? microstructureEvents : (orderFlow?.microstructure_events ?? []))
        .slice(-8)
        .reverse(),
    [microstructureEvents, orderFlow?.microstructure_events],
  );

  return (
    <section className="rounded-[28px] border border-slate-200/80 bg-[linear-gradient(180deg,#fffef6_0%,#ffffff_55%,#f8fafc_100%)] p-6 shadow-panel">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.22em] text-amber-700">
            Microstructure
          </p>
          <h3 className="mt-2 text-xl font-semibold text-slate-900">Microstructure Timeline</h3>
        </div>
        <span className="rounded-full border border-slate-200 bg-white px-3 py-1 text-xs text-slate-600">
          {events.length} events
        </span>
      </div>

      <div className="mt-5 space-y-3">
        {events.map((event) => (
          <div
            key={`${event.trade_time}-${event.type}-${event.bias}`}
            className="rounded-2xl border border-slate-100 bg-white px-4 py-4"
          >
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <span className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${eventTone(event.bias)}`}>
                  {event.bias}
                </span>
                <span className="text-sm font-semibold text-slate-900">
                  {formatEventType(event.type)}
                </span>
              </div>
              <span className="text-xs font-medium text-slate-500">
                {formatEventTime(event.trade_time)}
              </span>
            </div>

            <p className="mt-3 text-sm leading-6 text-slate-600">{event.detail}</p>

            <div className="mt-3 grid grid-cols-3 gap-2 text-xs text-slate-500">
              <StatPill label="Score" value={formatSigned(event.score)} />
              <StatPill label="Strength" value={event.strength.toFixed(2)} />
              <StatPill label="Price" value={event.price.toFixed(2)} />
            </div>
          </div>
        ))}

        {events.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-6 text-sm text-slate-500">
            暂无微结构时间线数据
          </div>
        ) : null}
      </div>
    </section>
  );
}

function StatPill({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-slate-100 bg-slate-50 px-3 py-2">
      <p className="text-[11px] uppercase tracking-[0.16em] text-slate-400">{label}</p>
      <p className="mt-1 font-semibold text-slate-900">{value}</p>
    </div>
  );
}

function eventTone(bias: string) {
  if (bias === "bullish") {
    return "bg-emerald-50 text-emerald-700";
  }
  if (bias === "bearish") {
    return "bg-rose-50 text-rose-700";
  }
  return "bg-slate-100 text-slate-700";
}

function formatEventType(value: string) {
  return value
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function formatEventTime(timestamp: number) {
  return new Date(timestamp).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function formatSigned(value: number) {
  return value > 0 ? `+${value}` : value.toString();
}
