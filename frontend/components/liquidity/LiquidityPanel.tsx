"use client";

import { useMemo } from "react";
import { LiquidityWallLevel } from "@/types/market";
import { useMarketStore } from "@/store/marketStore";

export default function LiquidityPanel() {
  const { liquidity, price, refreshDashboard } = useMarketStore();
  const askWalls = useMemo(
    () => (liquidity?.wall_levels ?? []).filter((wall) => wall.side === "ask"),
    [liquidity?.wall_levels],
  );
  const bidWalls = useMemo(
    () => (liquidity?.wall_levels ?? []).filter((wall) => wall.side === "bid"),
    [liquidity?.wall_levels],
  );

  return (
    <section className="rounded-[28px] border border-slate-200/80 bg-[linear-gradient(180deg,#ffffff_0%,#f8fafc_100%)] p-6 shadow-panel">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">Liquidity</p>
          <h3 className="mt-2 text-xl font-semibold text-slate-900">Wall Map</h3>
        </div>
        <button
          onClick={() => {
            void refreshDashboard();
          }}
          className="rounded-full border border-slate-200 bg-white px-3 py-1.5 text-sm font-medium text-slate-700"
        >
          更新
        </button>
      </div>

      {liquidity ? (
        <>
          <div className="mt-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            <MetricCard label="Buy Liquidity" value={formatPrice(liquidity.buy_liquidity)} />
            <MetricCard label="Sell Liquidity" value={formatPrice(liquidity.sell_liquidity)} />
            <MetricCard label="Imbalance" value={formatSigned(liquidity.order_book_imbalance, 2)} />
            <MetricCard label="Sweep" value={formatSweepType(liquidity.sweep_type)} accent />
          </div>

          <div className="mt-6 grid gap-4 lg:grid-cols-2">
            <WallColumn
              title="Ask Wall Map"
              subtitle="上方卖单墙与潜在流动性回收区"
              walls={askWalls}
              currentPrice={price?.price}
              tone="border-rose-100 bg-[linear-gradient(180deg,rgba(255,241,242,0.7)_0%,rgba(255,255,255,0.98)_100%)]"
            />
            <WallColumn
              title="Bid Wall Map"
              subtitle="下方买单墙与被动承接带"
              walls={bidWalls}
              currentPrice={price?.price}
              tone="border-emerald-100 bg-[linear-gradient(180deg,rgba(236,253,245,0.7)_0%,rgba(255,255,255,0.98)_100%)]"
            />
          </div>

          <div className="mt-6 rounded-2xl border border-slate-100 bg-slate-50 px-4 py-4">
            <div className="flex flex-wrap items-center gap-2">
              <Chip label={`Source ${liquidity.data_source}`} />
              <Chip label={`Equal High ${formatPrice(liquidity.equal_high)}`} />
              <Chip label={`Equal Low ${formatPrice(liquidity.equal_low)}`} />
              <Chip label={`${(liquidity.stop_clusters ?? []).length} stop clusters`} />
            </div>
          </div>
        </>
      ) : (
        <p className="mt-5 text-sm text-slate-500">暂无流动性数据</p>
      )}
    </section>
  );
}

function WallColumn({
  title,
  subtitle,
  walls,
  currentPrice,
  tone,
}: {
  title: string;
  subtitle: string;
  walls: LiquidityWallLevel[];
  currentPrice?: number | null;
  tone: string;
}) {
  return (
    <div className={`rounded-[24px] border p-4 ${tone}`}>
      <div>
        <p className="text-sm font-semibold text-slate-900">{title}</p>
        <p className="mt-1 text-xs text-slate-500">{subtitle}</p>
      </div>

      <div className="mt-4 space-y-3">
        {walls.map((wall) => (
          <div key={`${wall.side}-${wall.layer}-${wall.price}`} className="rounded-2xl border border-white bg-white px-4 py-3">
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="text-sm font-semibold text-slate-900">{wall.label}</p>
                <p className="mt-1 text-[11px] uppercase tracking-[0.18em] text-slate-400">
                  {formatLayer(wall.layer)} layer
                </p>
              </div>
              <div className="text-right">
                <p className="text-sm font-semibold text-slate-900">{formatPrice(wall.price)}</p>
                <p className="mt-1 text-xs text-slate-500">{formatPctDistance(currentPrice, wall.price)}</p>
              </div>
            </div>

            <div className="mt-3 grid grid-cols-3 gap-2 text-xs">
              <MiniStat label="Notional" value={formatCompactNotional(wall.notional)} />
              <MiniStat label="Distance" value={`${wall.distance_bps.toFixed(1)} bps`} />
              <MiniStat label="Strength" value={wall.strength.toFixed(2)} />
            </div>
          </div>
        ))}

        {walls.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-slate-200 bg-white/70 px-4 py-6 text-sm text-slate-500">
            当前数据源未提供订单簿分层墙位
          </div>
        ) : null}
      </div>
    </div>
  );
}

function MetricCard({
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

function MiniStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-slate-100 bg-slate-50 px-3 py-2">
      <p className="text-[11px] uppercase tracking-[0.16em] text-slate-400">{label}</p>
      <p className="mt-1 font-semibold text-slate-900">{value}</p>
    </div>
  );
}

function Chip({ label }: { label: string }) {
  return <span className="rounded-full border border-slate-200 bg-white px-3 py-1 text-xs text-slate-600">{label}</span>;
}

function formatPrice(value?: number | null) {
  return typeof value === "number" && Number.isFinite(value) && value > 0 ? value.toFixed(2) : "-";
}

function formatSigned(value: number, decimals: number) {
  if (!Number.isFinite(value)) {
    return "-";
  }
  const prefix = value > 0 ? "+" : "";
  return `${prefix}${value.toFixed(decimals)}`;
}

function formatSweepType(value: string) {
  if (!value || value === "none") {
    return "No Sweep";
  }
  return value
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function formatLayer(value: string) {
  if (!value) {
    return "Unknown";
  }
  return value.charAt(0).toUpperCase() + value.slice(1);
}

function formatCompactNotional(value: number) {
  if (!Number.isFinite(value) || value <= 0) {
    return "-";
  }
  if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`;
  }
  if (value >= 1_000) {
    return `${(value / 1_000).toFixed(1)}K`;
  }
  return value.toFixed(0);
}

function formatPctDistance(currentPrice?: number | null, target?: number | null) {
  if (
    typeof currentPrice !== "number" ||
    typeof target !== "number" ||
    !Number.isFinite(currentPrice) ||
    !Number.isFinite(target) ||
    currentPrice <= 0 ||
    target <= 0
  ) {
    return "-";
  }

  const pct = ((target - currentPrice) / currentPrice) * 100;
  const prefix = pct > 0 ? "+" : "";
  return `${prefix}${pct.toFixed(2)}%`;
}
