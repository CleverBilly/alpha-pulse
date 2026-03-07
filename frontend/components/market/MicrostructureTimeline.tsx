"use client";

import { useMemo, useState } from "react";
import { OrderFlowMicrostructureEvent } from "@/types/market";
import { useMarketStore } from "@/store/marketStore";

const HIGH_ORDER_EVENT_TYPES = new Set([
  "failed_auction_high_reject",
  "failed_auction_low_reclaim",
  "order_book_migration_layered",
  "order_book_migration_accelerated",
  "microstructure_confluence",
]);

const TIMELINE_FILTERS = [
  { id: "all", label: "All" },
  { id: "high_order", label: "High Order" },
  { id: "auction", label: "Auction" },
  { id: "migration", label: "Migration" },
  { id: "execution", label: "Execution" },
  { id: "absorption", label: "Absorption" },
] as const;

type TimelineFilterId = (typeof TIMELINE_FILTERS)[number]["id"];

export default function MicrostructureTimeline() {
  const { microstructureEvents = [], orderFlow } = useMarketStore();
  const [activeFilter, setActiveFilter] = useState<TimelineFilterId>("all");
  const events = useMemo(
    () =>
      (microstructureEvents.length > 0 ? microstructureEvents : (orderFlow?.microstructure_events ?? []))
        .slice(-8)
        .reverse(),
    [microstructureEvents, orderFlow?.microstructure_events],
  );
  const visibleEvents = useMemo(
    () => events.filter((event) => matchesFilter(event, activeFilter)),
    [activeFilter, events],
  );
  const summary = useMemo(() => buildSummary(visibleEvents, events.length), [visibleEvents, events.length]);

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
          {visibleEvents.length} / {events.length} visible
        </span>
      </div>

      <div className="mt-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <SummaryCard label="Visible" value={`${summary.visibleCount} / ${summary.totalCount}`} />
        <SummaryCard label="Net Score" value={formatSigned(summary.netScore)} />
        <SummaryCard label="Bias Tilt" value={`${summary.bullishCount}B / ${summary.bearishCount}S`} />
        <SummaryCard label="High Order" value={summary.highOrderCount.toString()} accent />
      </div>

      <div className="mt-5 flex flex-wrap gap-2">
        {TIMELINE_FILTERS.map((filter) => {
          const count = events.filter((event) => matchesFilter(event, filter.id)).length;
          const isActive = filter.id === activeFilter;

          return (
            <button
              key={filter.id}
              type="button"
              onClick={() => setActiveFilter(filter.id)}
              className={`inline-flex items-center gap-2 rounded-full border px-3 py-1.5 text-xs font-semibold transition ${
                isActive
                  ? "border-slate-900 bg-slate-900 text-white"
                  : "border-slate-200 bg-white text-slate-600 hover:border-slate-300 hover:text-slate-900"
              }`}
            >
              {filter.label}
              <span
                className={`rounded-full px-2 py-0.5 text-[10px] ${
                  isActive ? "bg-white/15 text-white" : "bg-slate-100 text-slate-500"
                }`}
              >
                {count}
              </span>
            </button>
          );
        })}
      </div>

      <div className="mt-6 space-y-4">
        {visibleEvents.map((event, index) => {
          const meta = getEventMeta(event.type);
          const isHighOrder = HIGH_ORDER_EVENT_TYPES.has(event.type);

          return (
            <div key={`${event.trade_time}-${event.type}-${event.bias}`} className="relative pl-7">
              {index !== visibleEvents.length - 1 ? (
                <span className="absolute left-[7px] top-9 h-[calc(100%-0.5rem)] w-px bg-slate-200" />
              ) : null}
              <span
                className={`absolute left-0 top-6 h-4 w-4 rounded-full ring-4 ring-white ${eventMarkerTone(
                  event.bias,
                  isHighOrder,
                )}`}
              />

              <article
                className={`rounded-[24px] border px-4 py-4 shadow-[0_18px_40px_rgba(15,23,42,0.04)] ${
                  isHighOrder
                    ? "border-amber-200/80 bg-[linear-gradient(135deg,rgba(255,251,235,0.98)_0%,rgba(255,255,255,0.96)_55%,rgba(255,247,237,0.98)_100%)]"
                    : "border-slate-100 bg-white"
                }`}
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="flex flex-wrap items-center gap-2">
                    <span
                      className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${familyTone(
                        meta.family,
                      )}`}
                    >
                      {meta.familyLabel}
                    </span>
                    {isHighOrder ? (
                      <span className="rounded-full border border-amber-300/80 bg-amber-100/80 px-2.5 py-1 text-[11px] font-semibold text-amber-800">
                        High Order
                      </span>
                    ) : null}
                    <span className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${eventTone(event.bias)}`}>
                      {formatBias(event.bias)}
                    </span>
                  </div>
                  <span className="text-xs font-medium text-slate-500">{formatEventTime(event.trade_time)}</span>
                </div>

                <div className="mt-3 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div className="min-w-0">
                    <p className="text-sm font-semibold text-slate-900">{meta.label}</p>
                    <p className="mt-2 text-sm leading-6 text-slate-600">{event.detail}</p>
                  </div>
                  <div className="inline-flex shrink-0 items-center rounded-2xl bg-slate-950 px-3 py-1.5 text-xs font-semibold text-white">
                    {impactLabel(event, isHighOrder)}
                  </div>
                </div>

                <div className="mt-4 grid gap-2 text-xs text-slate-500 sm:grid-cols-4">
                  <StatPill label="Score" value={formatSigned(event.score)} />
                  <StatPill label="Strength" value={event.strength.toFixed(2)} />
                  <StatPill label="Price" value={event.price.toFixed(2)} />
                  <StatPill label="Type" value={meta.shortLabel} />
                </div>
              </article>
            </div>
          );
        })}

        {visibleEvents.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-6 text-sm text-slate-500">
            当前筛选条件下暂无微结构时间线数据
          </div>
        ) : null}
      </div>
    </section>
  );
}

function SummaryCard({
  label,
  value,
  accent = false,
}: {
  label: string;
  value: string;
  accent?: boolean;
}) {
  return (
    <div
      className={`rounded-[22px] border px-4 py-3 ${
        accent
          ? "border-amber-200 bg-[linear-gradient(180deg,rgba(255,251,235,1)_0%,rgba(255,255,255,1)_100%)]"
          : "border-slate-100 bg-white"
      }`}
    >
      <p className="text-[11px] uppercase tracking-[0.16em] text-slate-400">{label}</p>
      <p className="mt-2 text-lg font-semibold text-slate-900">{value}</p>
    </div>
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

function familyTone(family: string) {
  switch (family) {
    case "auction":
      return "bg-rose-100 text-rose-700";
    case "migration":
      return "bg-sky-100 text-sky-700";
    case "execution":
      return "bg-violet-100 text-violet-700";
    case "absorption":
      return "bg-emerald-100 text-emerald-700";
    case "confluence":
      return "bg-amber-100 text-amber-800";
    default:
      return "bg-slate-100 text-slate-700";
  }
}

function eventMarkerTone(bias: string, isHighOrder: boolean) {
  if (isHighOrder && bias === "bullish") {
    return "bg-emerald-500";
  }
  if (isHighOrder && bias === "bearish") {
    return "bg-rose-500";
  }
  if (bias === "bullish") {
    return "bg-emerald-300";
  }
  if (bias === "bearish") {
    return "bg-rose-300";
  }
  return "bg-slate-300";
}

function matchesFilter(event: OrderFlowMicrostructureEvent, filterId: TimelineFilterId) {
  if (filterId === "all") {
    return true;
  }
  if (filterId === "high_order") {
    return HIGH_ORDER_EVENT_TYPES.has(event.type);
  }
  return getEventMeta(event.type).family === filterId;
}

function buildSummary(events: OrderFlowMicrostructureEvent[], totalCount: number) {
  return events.reduce(
    (summary, event) => ({
      totalCount,
      visibleCount: summary.visibleCount + 1,
      netScore: summary.netScore + event.score,
      bullishCount: summary.bullishCount + (event.bias === "bullish" ? 1 : 0),
      bearishCount: summary.bearishCount + (event.bias === "bearish" ? 1 : 0),
      highOrderCount: summary.highOrderCount + (HIGH_ORDER_EVENT_TYPES.has(event.type) ? 1 : 0),
    }),
    {
      totalCount,
      visibleCount: 0,
      netScore: 0,
      bullishCount: 0,
      bearishCount: 0,
      highOrderCount: 0,
    },
  );
}

function getEventMeta(type: string) {
  const meta = EVENT_METADATA[type];
  if (meta) {
    return meta;
  }

  return {
    label: formatRawEventType(type),
    shortLabel: type.toUpperCase().slice(0, 4),
    family: "other",
    familyLabel: "Other",
  } as const;
}

function impactLabel(event: OrderFlowMicrostructureEvent, isHighOrder: boolean) {
  if (isHighOrder) {
    return event.bias === "bearish" ? "High Order Reversal" : "High Order Continuation";
  }
  if (Math.abs(event.score) >= 5 || event.strength >= 0.7) {
    return "Primary Trigger";
  }
  return "Supporting Flow";
}

function formatBias(value: string) {
  if (value === "bullish") {
    return "Bullish";
  }
  if (value === "bearish") {
    return "Bearish";
  }
  return "Neutral";
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

function formatRawEventType(value: string) {
  return value
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

const EVENT_METADATA: Record<
  string,
  {
    label: string;
    shortLabel: string;
    family: "auction" | "migration" | "execution" | "absorption" | "confluence" | "other";
    familyLabel: string;
  }
> = {
  absorption: {
    label: "Absorption",
    shortLabel: "ABS",
    family: "absorption",
    familyLabel: "Absorption",
  },
  iceberg: {
    label: "Iceberg",
    shortLabel: "ICE",
    family: "absorption",
    familyLabel: "Absorption",
  },
  aggression_burst: {
    label: "Aggression Burst",
    shortLabel: "AGR",
    family: "execution",
    familyLabel: "Execution",
  },
  initiative_shift: {
    label: "Initiative Shift",
    shortLabel: "INI",
    family: "execution",
    familyLabel: "Execution",
  },
  large_trade_cluster: {
    label: "Large Trade Cluster",
    shortLabel: "LTC",
    family: "execution",
    familyLabel: "Execution",
  },
  failed_auction: {
    label: "Failed Auction",
    shortLabel: "FAU",
    family: "auction",
    familyLabel: "Auction",
  },
  failed_auction_high_reject: {
    label: "Failed Auction High Reject",
    shortLabel: "FAH",
    family: "auction",
    familyLabel: "Auction",
  },
  failed_auction_low_reclaim: {
    label: "Failed Auction Low Reclaim",
    shortLabel: "FAL",
    family: "auction",
    familyLabel: "Auction",
  },
  order_book_migration: {
    label: "Order Book Migration",
    shortLabel: "OBM",
    family: "migration",
    familyLabel: "Migration",
  },
  order_book_migration_layered: {
    label: "Order Book Migration Layered",
    shortLabel: "OBL",
    family: "migration",
    familyLabel: "Migration",
  },
  order_book_migration_accelerated: {
    label: "Order Book Migration Accelerated",
    shortLabel: "OBA",
    family: "migration",
    familyLabel: "Migration",
  },
  microstructure_confluence: {
    label: "Microstructure Confluence",
    shortLabel: "MCF",
    family: "confluence",
    familyLabel: "Confluence",
  },
};
