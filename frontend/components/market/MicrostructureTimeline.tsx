"use client";

import { useMemo, useState } from "react";
import { Tag } from "antd";
import { formatBiasLabel, formatMicrostructureEventTypeLabel } from "@/lib/uiLabels";
import { OrderFlowMicrostructureEvent } from "@/types/market";
import { useMarketStore } from "@/store/marketStore";

const HIGH_ORDER_EVENT_TYPES = new Set([
  "auction_trap_reversal",
  "absorption_reload_continuation",
  "exhaustion_migration_reversal",
  "failed_auction_high_reject",
  "failed_auction_low_reclaim",
  "iceberg_reload",
  "initiative_exhaustion",
  "liquidity_ladder_breakout",
  "migration_auction_flip",
  "order_book_migration_layered",
  "order_book_migration_accelerated",
  "microstructure_confluence",
]);

const TIMELINE_FILTERS = [
  { id: "all", label: "全部" },
  { id: "high_order", label: "高阶事件" },
  { id: "auction", label: "拍卖" },
  { id: "migration", label: "迁移" },
  { id: "execution", label: "执行" },
  { id: "absorption", label: "吸收" },
  { id: "composite", label: "复合" },
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
  const dominantFamily = useMemo(() => resolveDominantFamily(visibleEvents), [visibleEvents]);
  const averageStrength = useMemo(
    () =>
      visibleEvents.length > 0
        ? visibleEvents.reduce((sum, event) => sum + event.strength, 0) / visibleEvents.length
        : 0,
    [visibleEvents],
  );

  return (
    <section className="surface-panel surface-panel--warm">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.14em] text-amber-700">
            微结构
          </p>
          <h3 className="mt-2 text-xl font-semibold text-slate-900">微结构时间线</h3>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-slate-600">
            这里把盘口吸收、迁移、执行切换和高阶共振事件串成一条可读的盘中叙事链。
          </p>
        </div>
        <span className="rounded-full border border-slate-200 bg-white px-3 py-1 text-xs text-slate-600">
          可见 {visibleEvents.length} / {events.length}
        </span>
      </div>

      <div className="mt-5 grid gap-4 lg:grid-cols-[1.15fr_0.85fr]">
        <div className="rounded-[24px] border border-slate-200/70 bg-white/80 px-4 py-4">
          <div className="flex flex-wrap items-center gap-2">
            <Tag color="gold">{dominantFamily.label}</Tag>
            <Tag color={summary.netScore >= 0 ? "success" : "error"}>
              净分 {formatSigned(summary.netScore)}
            </Tag>
            <Tag color="blue">平均强度 {averageStrength.toFixed(2)}</Tag>
          </div>
          <p className="mt-3 text-sm leading-7 text-slate-600">
            {dominantFamily.copy(summary, averageStrength)}
          </p>
        </div>

        <div className="grid gap-3 sm:grid-cols-2">
          <SummaryCard label="可见事件" value={`${summary.visibleCount} / ${summary.totalCount}`} />
          <SummaryCard label="净分" value={formatSigned(summary.netScore)} />
          <SummaryCard label="方向倾斜" value={`${summary.bullishCount} 多 / ${summary.bearishCount} 空`} />
          <SummaryCard label="高阶事件" value={summary.highOrderCount.toString()} accent />
        </div>
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
                        高阶事件
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
                  <StatPill label="评分" value={formatSigned(event.score)} />
                  <StatPill label="强度" value={event.strength.toFixed(2)} />
                  <StatPill label="价格" value={event.price.toFixed(2)} />
                  <StatPill label="类型" value={meta.shortLabel} />
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
      <p className="text-[11px] uppercase tracking-[0.12em] text-slate-400">{label}</p>
      <p className="mt-2 text-lg font-semibold text-slate-900">{value}</p>
    </div>
  );
}

function StatPill({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-slate-100 bg-slate-50 px-3 py-2">
      <p className="text-[11px] uppercase tracking-[0.12em] text-slate-400">{label}</p>
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
    case "composite":
      return "bg-orange-100 text-orange-700";
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

function resolveDominantFamily(events: OrderFlowMicrostructureEvent[]) {
  if (events.length === 0) {
    return {
      label: "等待微结构事件",
      copy: () => "当前还没有足够的微结构事件用于形成可读叙事。",
    };
  }

  const counts = new Map<string, number>();
  for (const event of events) {
    const family = getEventMeta(event.type).familyLabel;
    counts.set(family, (counts.get(family) ?? 0) + 1);
  }

  const [label = "混合流向"] =
    [...counts.entries()].sort((left, right) => right[1] - left[1])[0] ?? [];

  return {
    label,
    copy: (
      summary: {
        netScore: number;
        bullishCount: number;
        bearishCount: number;
        highOrderCount: number;
      },
      averageStrength: number,
    ) =>
      `${label} 当前是主导叙事，净分数 ${formatSigned(summary.netScore)}，高阶事件 ${summary.highOrderCount} 个，平均强度 ${averageStrength.toFixed(
        2,
      )}。多头 ${summary.bullishCount} 个，空头 ${summary.bearishCount} 个。`,
  };
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
    familyLabel: "其他",
  } as const;
}

function impactLabel(event: OrderFlowMicrostructureEvent, isHighOrder: boolean) {
  if (isHighOrder) {
    return event.bias === "bearish" ? "高阶反转" : "高阶延续";
  }
  if (Math.abs(event.score) >= 5 || event.strength >= 0.7) {
    return "主触发";
  }
  return "辅助流";
}

function formatBias(value: string) {
  return formatBiasLabel(value);
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
  return formatMicrostructureEventTypeLabel(value);
}

const EVENT_METADATA: Record<
  string,
  {
    label: string;
    shortLabel: string;
    family: "auction" | "migration" | "execution" | "absorption" | "composite" | "confluence" | "other";
    familyLabel: string;
  }
> = {
  absorption: {
    label: "吸收",
    shortLabel: "ABS",
    family: "absorption",
    familyLabel: "吸收",
  },
  iceberg: {
    label: "冰山",
    shortLabel: "ICE",
    family: "absorption",
    familyLabel: "吸收",
  },
  iceberg_reload: {
    label: "冰山回补",
    shortLabel: "IRL",
    family: "absorption",
    familyLabel: "吸收",
  },
  aggression_burst: {
    label: "主动成交爆发",
    shortLabel: "AGR",
    family: "execution",
    familyLabel: "执行",
  },
  initiative_shift: {
    label: "主动性切换",
    shortLabel: "INI",
    family: "execution",
    familyLabel: "执行",
  },
  initiative_exhaustion: {
    label: "主动性衰竭",
    shortLabel: "IEX",
    family: "execution",
    familyLabel: "执行",
  },
  large_trade_cluster: {
    label: "大单簇",
    shortLabel: "LTC",
    family: "execution",
    familyLabel: "执行",
  },
  failed_auction: {
    label: "失败拍卖",
    shortLabel: "FAU",
    family: "auction",
    familyLabel: "拍卖",
  },
  failed_auction_high_reject: {
    label: "高位失败拍卖",
    shortLabel: "FAH",
    family: "auction",
    familyLabel: "拍卖",
  },
  failed_auction_low_reclaim: {
    label: "低位失败回收",
    shortLabel: "FAL",
    family: "auction",
    familyLabel: "拍卖",
  },
  order_book_migration: {
    label: "订单簿迁移",
    shortLabel: "OBM",
    family: "migration",
    familyLabel: "迁移",
  },
  order_book_migration_layered: {
    label: "分层订单簿迁移",
    shortLabel: "OBL",
    family: "migration",
    familyLabel: "迁移",
  },
  order_book_migration_accelerated: {
    label: "加速订单簿迁移",
    shortLabel: "OBA",
    family: "migration",
    familyLabel: "迁移",
  },
  microstructure_confluence: {
    label: "微结构共振",
    shortLabel: "MCF",
    family: "confluence",
    familyLabel: "共振",
  },
  auction_trap_reversal: {
    label: "拍卖陷阱反转",
    shortLabel: "TRP",
    family: "composite",
    familyLabel: "复合",
  },
  liquidity_ladder_breakout: {
    label: "流动性阶梯突破",
    shortLabel: "LLB",
    family: "composite",
    familyLabel: "复合",
  },
  migration_auction_flip: {
    label: "迁移拍卖翻转",
    shortLabel: "MAF",
    family: "composite",
    familyLabel: "复合",
  },
  absorption_reload_continuation: {
    label: "吸收回补延续",
    shortLabel: "ARC",
    family: "composite",
    familyLabel: "复合",
  },
  exhaustion_migration_reversal: {
    label: "衰竭迁移反转",
    shortLabel: "EMR",
    family: "composite",
    familyLabel: "复合",
  },
};
