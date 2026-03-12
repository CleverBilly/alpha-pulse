"use client";

import { useMemo } from "react";
import { Card, Progress, Tag, Typography } from "antd";
import { formatLiquiditySideDominance, formatSweepLabel } from "@/lib/uiLabels";
import { LiquidityWallEvolution, LiquidityWallLevel, LiquidityWallStrengthBand } from "@/types/market";
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
  const askBands = useMemo(
    () =>
      [...(liquidity?.wall_strength_bands ?? [])]
        .filter((band) => band.side === "ask")
        .sort((left, right) => left.lower_distance_bps - right.lower_distance_bps),
    [liquidity?.wall_strength_bands],
  );
  const bidBands = useMemo(
    () =>
      [...(liquidity?.wall_strength_bands ?? [])]
        .filter((band) => band.side === "bid")
        .sort((left, right) => left.lower_distance_bps - right.lower_distance_bps),
    [liquidity?.wall_strength_bands],
  );
  const wallEvolution = useMemo(
    () => [...(liquidity?.wall_evolution ?? [])].sort((left, right) => intervalRank(left.interval) - intervalRank(right.interval)),
    [liquidity?.wall_evolution],
  );

  return (
    <section>
      <Card
        variant="borderless"
        className="surface-card surface-card--paper"
      >
        <div className="flex items-center justify-between gap-3">
          <div>
            <Typography.Title level={3} className="!mb-0 !text-[24px] !tracking-[-0.03em]">
              流动性
            </Typography.Title>
            <p className="mt-2 text-xs font-semibold uppercase tracking-[0.14em] text-slate-500">墙位分布</p>
          </div>
          <button
            onClick={() => {
              void refreshDashboard(true);
            }}
            className="rounded-full border border-slate-200 bg-white/80 px-4 py-2 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:text-slate-950"
          >
            更新
          </button>
        </div>

        {liquidity ? (
          <>
            <div className="mt-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
              <MetricCard label="买方流动性" value={formatPrice(liquidity.buy_liquidity)} />
              <MetricCard label="卖方流动性" value={formatPrice(liquidity.sell_liquidity)} />
              <MetricCard label="盘口失衡" value={formatSigned(liquidity.order_book_imbalance, 2)} />
              <MetricCard label="扫流动性" value={formatSweepType(liquidity.sweep_type)} accent />
            </div>

            <div className="mt-6 grid gap-4 lg:grid-cols-2">
              <WallColumn
                title="卖墙分布"
                subtitle="上方卖单墙与潜在流动性回收区"
                walls={askWalls}
                currentPrice={price?.price}
                tone="border-rose-100 bg-[linear-gradient(180deg,rgba(255,241,242,0.7)_0%,rgba(255,255,255,0.98)_100%)]"
              />
              <WallColumn
                title="买墙分布"
                subtitle="下方买单墙与被动承接带"
                walls={bidWalls}
                currentPrice={price?.price}
                tone="border-emerald-100 bg-[linear-gradient(180deg,rgba(236,253,245,0.7)_0%,rgba(255,255,255,0.98)_100%)]"
              />
            </div>

            <div className="mt-6 grid gap-4 lg:grid-cols-2">
              <BandColumn
                title="卖墙热度带"
                subtitle="按距离分带聚合卖单墙热度"
                bands={askBands}
                tone="border-rose-100 bg-[linear-gradient(180deg,rgba(255,241,242,0.7)_0%,rgba(255,255,255,0.98)_100%)]"
                strokeColor="#f43f5e"
              />
              <BandColumn
                title="买墙热度带"
                subtitle="按距离分带聚合买单墙热度"
                bands={bidBands}
                tone="border-emerald-100 bg-[linear-gradient(180deg,rgba(236,253,245,0.7)_0%,rgba(255,255,255,0.98)_100%)]"
                strokeColor="#10b981"
              />
            </div>

            <div className="mt-6 rounded-[24px] border border-slate-200 bg-white px-4 py-4">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-sm font-semibold text-slate-900">跨周期墙位演化</p>
                  <p className="mt-1 text-xs text-slate-500">观察 1m 到 4h 的主导 wall、强度变化和距离迁移。</p>
                </div>
                <Tag>{wallEvolution.length} 个周期</Tag>
              </div>

              <div className="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                {wallEvolution.map((point) => (
                  <EvolutionCard key={point.interval} point={point} />
                ))}
                {wallEvolution.length === 0 ? (
                  <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-6 text-sm text-slate-500">
                    当前数据源未提供跨周期 wall 演化概览
                  </div>
                ) : null}
              </div>
            </div>

            <div className="mt-6 rounded-2xl border border-slate-100 bg-slate-50 px-4 py-4">
              <div className="flex flex-wrap items-center gap-2">
                <Chip label={`数据源 ${liquidity.data_source}`} />
                <Chip label={`等高点 ${formatPrice(liquidity.equal_high)}`} />
                <Chip label={`等低点 ${formatPrice(liquidity.equal_low)}`} />
                <Chip label={`${(liquidity.stop_clusters ?? []).length} 个止损簇`} />
              </div>
            </div>
          </>
        ) : (
          <p className="mt-5 text-sm text-slate-500">暂无流动性数据</p>
        )}
      </Card>
    </section>
  );
}

function BandColumn({
  title,
  subtitle,
  bands,
  tone,
  strokeColor,
}: {
  title: string;
  subtitle: string;
  bands: LiquidityWallStrengthBand[];
  tone: string;
  strokeColor: string;
}) {
  const maxStrength = bands.reduce((max, band) => Math.max(max, band.strength), 0);

  return (
    <div className={`rounded-[24px] border p-4 ${tone}`}>
      <div>
        <p className="text-sm font-semibold text-slate-900">{title}</p>
        <p className="mt-1 text-xs text-slate-500">{subtitle}</p>
      </div>

      <div className="mt-4 space-y-3">
        {bands.map((band) => (
          <div key={`${band.side}-${band.band}`} className="rounded-2xl border border-white bg-white px-4 py-3">
            <div className="flex items-center justify-between gap-3">
              <div>
                <p className="text-sm font-semibold text-slate-900">{band.band}</p>
                <p className="mt-1 text-xs text-slate-500">
                  {band.level_count} 层 · 主价位 {formatPrice(band.dominant_price)}
                </p>
              </div>
              <div className="text-right">
                <p className="text-sm font-semibold text-slate-900">{formatCompactNotional(band.total_notional)}</p>
                <p className="mt-1 text-xs text-slate-500">强度 {band.strength.toFixed(2)}</p>
              </div>
            </div>

            <Progress
              percent={resolveBandWidth(band.strength, maxStrength)}
              showInfo={false}
              size="small"
              strokeColor={strokeColor}
              className="!mt-3"
            />
          </div>
        ))}

        {bands.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-slate-200 bg-white/70 px-4 py-6 text-sm text-slate-500">
            当前数据源未提供 wall strength band
          </div>
        ) : null}
      </div>
    </div>
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
                <p className="text-sm font-semibold text-slate-900">{formatWallLabel(wall.label)}</p>
                <p className="mt-1 text-[11px] uppercase tracking-[0.12em] text-slate-400">
                  {formatLayer(wall.layer)}层
                </p>
              </div>
              <div className="text-right">
                <p className="text-sm font-semibold text-slate-900">{formatPrice(wall.price)}</p>
                <p className="mt-1 text-xs text-slate-500">{formatPctDistance(currentPrice, wall.price)}</p>
              </div>
            </div>

            <div className="mt-3 grid grid-cols-3 gap-2 text-xs">
              <MiniStat label="名义价值" value={formatCompactNotional(wall.notional)} />
              <MiniStat label="距离" value={`${wall.distance_bps.toFixed(1)} bps`} />
              <MiniStat label="强度" value={wall.strength.toFixed(2)} />
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

function EvolutionCard({ point }: { point: LiquidityWallEvolution }) {
  const dominantLabel = formatDominantSide(point.dominant_side);
  const dominantTone =
    point.dominant_side === "bid"
      ? "bg-emerald-100 text-emerald-700"
      : point.dominant_side === "ask"
        ? "bg-rose-100 text-rose-700"
        : "bg-slate-100 text-slate-600";

  return (
    <div className="rounded-[22px] border border-slate-100 bg-slate-50 px-4 py-4">
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm font-semibold text-slate-900">{point.interval}</p>
        <span className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${dominantTone}`}>{dominantLabel}</span>
      </div>

      <div className="mt-3 grid grid-cols-2 gap-2 text-xs">
        <MiniStat label="买盘变化" value={formatSigned(point.buy_strength_delta, 2)} />
        <MiniStat label="卖盘变化" value={formatSigned(point.sell_strength_delta, 2)} />
        <MiniStat label="买盘距离" value={`${point.buy_distance_bps.toFixed(1)} bps`} />
        <MiniStat label="卖盘距离" value={`${point.sell_distance_bps.toFixed(1)} bps`} />
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-2">
        <Chip label={`扫流动性 ${formatSweepType(point.sweep_type)}`} />
        <Chip label={`失衡 ${formatSigned(point.order_book_imbalance, 2)}`} />
        <Chip label={`数据源 ${point.data_source}`} />
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
          : "border-slate-100 bg-white/76"
      }`}
    >
      <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-400">{label}</p>
      <p className="mt-2 text-lg font-semibold text-slate-900">{value}</p>
    </div>
  );
}

function MiniStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-[18px] border border-slate-100 bg-slate-50 px-3 py-2">
      <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-400">{label}</p>
      <p className="mt-1 font-semibold text-slate-900">{value}</p>
    </div>
  );
}

function Chip({ label }: { label: string }) {
  return <Tag>{label}</Tag>;
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
    return "未见明显扫流动性";
  }
  return formatSweepLabel(value);
}

function formatLayer(value: string) {
  if (!value) {
    return "未知";
  }
  if (value === "near") {
    return "近端";
  }
  if (value === "mid") {
    return "中段";
  }
  if (value === "far") {
    return "远端";
  }
  return value;
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

function formatWallLabel(value?: string | null) {
  if (!value) {
    return "未知墙位";
  }

  return value
    .replace(/^Near Ask Wall$/i, "近端卖墙")
    .replace(/^Mid Ask Wall$/i, "中段卖墙")
    .replace(/^Far Ask Wall$/i, "远端卖墙")
    .replace(/^Near Bid Wall$/i, "近端买墙")
    .replace(/^Mid Bid Wall$/i, "中段买墙")
    .replace(/^Far Bid Wall$/i, "远端买墙");
}

function formatDominantSide(value: string) {
  return formatLiquiditySideDominance(value);
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

function resolveBandWidth(value: number, max: number) {
  if (!Number.isFinite(value) || value <= 0 || !Number.isFinite(max) || max <= 0) {
    return 0;
  }
  return Math.max(12, Math.min(100, (value / max) * 100));
}

function intervalRank(interval: string) {
  switch (interval) {
    case "1m":
      return 1;
    case "5m":
      return 2;
    case "15m":
      return 3;
    case "1h":
      return 4;
    case "4h":
      return 5;
    default:
      return 99;
  }
}
