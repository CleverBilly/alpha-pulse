"use client";

import { useMemo, useState } from "react";
import { useMarketStore } from "@/store/marketStore";
import {
  IndicatorSeriesPoint,
  Kline,
  Liquidity,
  LiquiditySeriesPoint,
  OrderFlowMicrostructureEvent,
  Structure,
  StructureEvent,
  StructureSeriesPoint,
} from "@/types/market";
import { Signal, SignalTimelinePoint } from "@/types/signal";

const CHART_WIDTH = 960;
const CHART_HEIGHT = 360;
const PADDING_TOP = 24;
const PADDING_RIGHT = 118;
const PADDING_BOTTOM = 36;
const PADDING_LEFT = 18;
const PRIMARY_MICROSTRUCTURE_TYPES = ["absorption", "iceberg", "aggression_burst"] as const;
const SECONDARY_MICROSTRUCTURE_LAYERS = [
  { key: "initiative_shift", label: "Initiative Shift", types: ["initiative_shift"] },
  { key: "large_trade_cluster", label: "Large Trade Cluster", types: ["large_trade_cluster"] },
  {
    key: "reload_exhaustion",
    label: "Reload / Exhaustion",
    types: ["iceberg_reload", "initiative_exhaustion"],
  },
  {
    key: "failed_auction",
    label: "Failed Auction",
    types: ["failed_auction", "failed_auction_high_reject", "failed_auction_low_reclaim"],
  },
  {
    key: "order_book_migration",
    label: "Order Book Migration",
    types: ["order_book_migration", "order_book_migration_layered", "order_book_migration_accelerated"],
  },
  {
    key: "composite_patterns",
    label: "Composite Patterns",
    types: [
      "auction_trap_reversal",
      "liquidity_ladder_breakout",
      "migration_auction_flip",
      "absorption_reload_continuation",
      "exhaustion_migration_reversal",
    ],
  },
  { key: "microstructure_confluence", label: "Microstructure Confluence", types: ["microstructure_confluence"] },
] as const;

type SecondaryLayerKey = (typeof SECONDARY_MICROSTRUCTURE_LAYERS)[number]["key"];

const DEFAULT_SECONDARY_LAYER_STATE: Record<SecondaryLayerKey, boolean> = {
  initiative_shift: false,
  large_trade_cluster: false,
  reload_exhaustion: false,
  failed_auction: false,
  order_book_migration: false,
  composite_patterns: false,
  microstructure_confluence: false,
};

export default function KlineChart() {
  const {
    klines,
    indicator,
    indicatorSeries,
    microstructureEvents = [],
    structure,
    structureSeries,
    liquidity,
    liquiditySeries,
    signal,
    signalTimeline,
    loading,
    error,
    refreshDashboard,
  } = useMarketStore();
  const [enabledLayers, setEnabledLayers] = useState<Record<SecondaryLayerKey, boolean>>(DEFAULT_SECONDARY_LAYER_STATE);
  const [activeMicrostructureMarkerKey, setActiveMicrostructureMarkerKey] = useState<string | null>(null);
  const visibleKlines = klines.slice(-48);
  const visibleMicrostructureTypes = useMemo(() => {
    const types: string[] = [...PRIMARY_MICROSTRUCTURE_TYPES];
    for (const layer of SECONDARY_MICROSTRUCTURE_LAYERS) {
      if (enabledLayers[layer.key]) {
        types.push(...layer.types);
      }
    }
    return types;
  }, [enabledLayers]);

  const chart = useMemo(
    () =>
      buildChartModel(
        visibleKlines,
        indicatorSeries,
        microstructureEvents,
        visibleMicrostructureTypes,
        structure,
        structureSeries,
        liquidity,
        liquiditySeries,
        signal,
        signalTimeline,
      ),
    [
      visibleKlines,
      indicatorSeries,
      microstructureEvents,
      visibleMicrostructureTypes,
      structure,
      structureSeries,
      liquidity,
      liquiditySeries,
      signal,
      signalTimeline,
    ],
  );
  const activeMicrostructureMarker = useMemo(
    () =>
      chart.microstructureMarkers.find((marker) => marker.key === activeMicrostructureMarkerKey) ?? null,
    [activeMicrostructureMarkerKey, chart.microstructureMarkers],
  );
  const latestKline = visibleKlines[visibleKlines.length - 1] ?? null;
  const microstructureTooltip = activeMicrostructureMarker
    ? buildMicrostructureTooltip(activeMicrostructureMarker)
    : null;

  return (
    <section className="rounded-2xl bg-panel p-5 shadow-panel">
      <div className="mb-4 flex items-center justify-between gap-3">
        <div>
          <h3 className="text-lg font-semibold">Kline Chart</h3>
          <p className="text-sm text-muted">
            48 根 K 线，叠加结构点、动态支撑阻力、流动性轨迹、信号位与多指标
          </p>
        </div>
        <button
          onClick={() => {
            void refreshDashboard(true);
          }}
          className="rounded-lg border border-slate-200 px-3 py-1 text-sm text-slate-700"
        >
          更新K线
        </button>
      </div>

      {loading && visibleKlines.length === 0 ? <p className="text-sm text-muted">加载中...</p> : null}
      {error ? <p className="text-sm text-negative">{error}</p> : null}
      {!loading && !error && visibleKlines.length === 0 ? (
        <p className="text-sm text-muted">暂无 K 线数据</p>
      ) : null}

      {visibleKlines.length > 0 ? (
        <div className="space-y-4">
          <div className="flex flex-wrap items-center gap-2 text-xs">
            <span className="rounded-full border border-slate-200 bg-white px-3 py-1 text-slate-500">
              Core Layers: ABS / ICE / AGR
            </span>
            {SECONDARY_MICROSTRUCTURE_LAYERS.map((layer) => (
              <LayerToggle
                key={layer.key}
                label={layer.label}
                active={enabledLayers[layer.key]}
                onClick={() => {
                  setEnabledLayers((value) => ({ ...value, [layer.key]: !value[layer.key] }));
                  setActiveMicrostructureMarkerKey(null);
                }}
              />
            ))}
          </div>

          <div className="overflow-hidden rounded-xl border border-slate-100 bg-[linear-gradient(180deg,#ffffff_0%,#f8fbff_100%)] p-3">
            <svg viewBox={`0 0 ${CHART_WIDTH} ${CHART_HEIGHT}`} className="h-[360px] w-full">
              <defs>
                <linearGradient id="chart-bg" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="rgba(14,165,233,0.08)" />
                  <stop offset="100%" stopColor="rgba(255,255,255,0)" />
                </linearGradient>
              </defs>

              <rect x="0" y="0" width={CHART_WIDTH} height={CHART_HEIGHT} fill="url(#chart-bg)" />

              {chart.gridLines.map((line, index) => (
                <g key={`grid-${index}`}>
                  <line
                    x1={PADDING_LEFT}
                    y1={line.y}
                    x2={CHART_WIDTH - PADDING_RIGHT}
                    y2={line.y}
                    stroke="rgba(148,163,184,0.22)"
                    strokeDasharray="4 6"
                  />
                  <text
                    x={CHART_WIDTH - PADDING_RIGHT + 8}
                    y={line.y + 4}
                    fontSize="11"
                    fill="#64748b"
                  >
                    {line.value.toFixed(2)}
                  </text>
                </g>
              ))}

              {chart.zoneLines.map((line) => (
                <g key={`${line.label}-${line.y}`}>
                  <line
                    x1={PADDING_LEFT}
                    y1={line.y}
                    x2={CHART_WIDTH - PADDING_RIGHT}
                    y2={line.y}
                    stroke={line.color}
                    strokeWidth="1.3"
                    strokeDasharray={line.dasharray}
                    opacity="0.72"
                  />
                  <rect
                    x={CHART_WIDTH - PADDING_RIGHT + 8}
                    y={line.y - 9}
                    width={line.label.length * 7.2 + 18}
                    height={18}
                    rx={9}
                    fill={line.labelBackground}
                    opacity="0.92"
                  />
                  <text
                    x={CHART_WIDTH - PADDING_RIGHT + 16}
                    y={line.y + 4}
                    fontSize="10"
                    fill={line.labelColor}
                  >
                    {line.label}
                  </text>
                </g>
              ))}

              {chart.series.supportTrack.map((points, index) => (
                <polyline
                  key={`support-track-${index}`}
                  points={points}
                  fill="none"
                  stroke="#047857"
                  strokeWidth="1.8"
                  opacity="0.8"
                />
              ))}
              {chart.series.internalSupportTrack.map((points, index) => (
                <polyline
                  key={`internal-support-track-${index}`}
                  points={points}
                  fill="none"
                  stroke="#0f766e"
                  strokeWidth="1.2"
                  strokeDasharray="4 4"
                  opacity="0.45"
                />
              ))}
              {chart.series.resistanceTrack.map((points, index) => (
                <polyline
                  key={`resistance-track-${index}`}
                  points={points}
                  fill="none"
                  stroke="#be123c"
                  strokeWidth="1.8"
                  opacity="0.8"
                />
              ))}
              {chart.series.internalResistanceTrack.map((points, index) => (
                <polyline
                  key={`internal-resistance-track-${index}`}
                  points={points}
                  fill="none"
                  stroke="#fb7185"
                  strokeWidth="1.2"
                  strokeDasharray="4 4"
                  opacity="0.45"
                />
              ))}
              {chart.series.buyLiquidityTrack.map((points, index) => (
                <polyline
                  key={`buy-liquidity-track-${index}`}
                  points={points}
                  fill="none"
                  stroke="#0f766e"
                  strokeWidth="1.4"
                  strokeDasharray="7 5"
                  opacity="0.7"
                />
              ))}
              {chart.series.sellLiquidityTrack.map((points, index) => (
                <polyline
                  key={`sell-liquidity-track-${index}`}
                  points={points}
                  fill="none"
                  stroke="#ea580c"
                  strokeWidth="1.4"
                  strokeDasharray="7 5"
                  opacity="0.7"
                />
              ))}
              {chart.series.equalHighTrack.map((points, index) => (
                <polyline
                  key={`equal-high-track-${index}`}
                  points={points}
                  fill="none"
                  stroke="#b45309"
                  strokeWidth="1.2"
                  strokeDasharray="3 5"
                  opacity="0.6"
                />
              ))}
              {chart.series.equalLowTrack.map((points, index) => (
                <polyline
                  key={`equal-low-track-${index}`}
                  points={points}
                  fill="none"
                  stroke="#155e75"
                  strokeWidth="1.2"
                  strokeDasharray="3 5"
                  opacity="0.6"
                />
              ))}

              {chart.series.bollingerUpper.map((points, index) => (
                <polyline
                  key={`bb-upper-${index}`}
                  points={points}
                  fill="none"
                  stroke="#38bdf8"
                  strokeWidth="1.5"
                  strokeDasharray="4 4"
                />
              ))}
              {chart.series.bollingerMiddle.map((points, index) => (
                <polyline
                  key={`bb-middle-${index}`}
                  points={points}
                  fill="none"
                  stroke="#64748b"
                  strokeWidth="1.3"
                />
              ))}
              {chart.series.bollingerLower.map((points, index) => (
                <polyline
                  key={`bb-lower-${index}`}
                  points={points}
                  fill="none"
                  stroke="#7dd3fc"
                  strokeWidth="1.5"
                  strokeDasharray="4 4"
                />
              ))}
              {chart.series.vwap.map((points, index) => (
                <polyline
                  key={`vwap-${index}`}
                  points={points}
                  fill="none"
                  stroke="#f59e0b"
                  strokeWidth="1.8"
                />
              ))}
              {chart.series.ema20.map((points, index) => (
                <polyline
                  key={`ema20-${index}`}
                  points={points}
                  fill="none"
                  stroke="#10b981"
                  strokeWidth="2"
                />
              ))}
              {chart.series.ema50.map((points, index) => (
                <polyline
                  key={`ema50-${index}`}
                  points={points}
                  fill="none"
                  stroke="#f43f5e"
                  strokeWidth="2"
                />
              ))}

              {chart.candles.map((candle, index) => (
                <g key={`${candle.openTime}-${index}`}>
                  <line
                    x1={candle.x}
                    y1={candle.highY}
                    x2={candle.x}
                    y2={candle.lowY}
                    stroke={candle.color}
                    strokeWidth="1.5"
                  />
                  <rect
                    x={candle.x - candle.bodyWidth / 2}
                    y={candle.bodyY}
                    width={candle.bodyWidth}
                    height={candle.bodyHeight}
                    rx="2"
                    fill={candle.color}
                    opacity="0.92"
                  />
                </g>
              ))}

              {chart.structureMarkers.map((marker) => (
                <g key={`${marker.label}-${marker.openTime}-${marker.x}`}>
                  <circle
                    cx={marker.x}
                    cy={marker.y}
                    r={marker.tier === "internal" ? 3.8 : 4.6}
                    fill={marker.color}
                    stroke="#ffffff"
                    strokeWidth={marker.tier === "internal" ? 1.2 : 1.4}
                    opacity={marker.tier === "internal" ? 0.82 : 1}
                  />
                  <rect
                    x={marker.x - marker.label.length * 3.6 - 7}
                    y={marker.labelY - 9}
                    width={marker.label.length * 7.2 + 14}
                    height={16}
                    rx={8}
                    fill={marker.labelBackground}
                    opacity={marker.tier === "internal" ? 0.88 : 1}
                  />
                  <text
                    x={marker.x}
                    y={marker.labelY + 3}
                    textAnchor="middle"
                    fontSize={marker.tier === "internal" ? "8.8" : "9.5"}
                    fill={marker.labelColor}
                    fontWeight="600"
                  >
                    {marker.label}
                  </text>
                </g>
              ))}

              {chart.microstructureMarkers.map((marker) => (
                <g
                  key={marker.key}
                  role="button"
                  tabIndex={0}
                  aria-label={`Micro ${marker.label} ${formatMicrostructureEventType(marker.type)}`}
                  className="cursor-pointer"
                  onMouseEnter={() => setActiveMicrostructureMarkerKey(marker.key)}
                  onMouseLeave={() => setActiveMicrostructureMarkerKey((value) => (value === marker.key ? null : value))}
                  onFocus={() => setActiveMicrostructureMarkerKey(marker.key)}
                  onBlur={() => setActiveMicrostructureMarkerKey((value) => (value === marker.key ? null : value))}
                  onClick={() => setActiveMicrostructureMarkerKey(marker.key)}
                >
                  <rect
                    x={marker.x - 5.2}
                    y={marker.y - 5.2}
                    width={10.4}
                    height={10.4}
                    rx={2.4}
                    fill={marker.color}
                    stroke="#ffffff"
                    strokeWidth="1.3"
                    transform={`rotate(45 ${marker.x} ${marker.y})`}
                  />
                  <rect
                    x={marker.x - 18}
                    y={marker.labelY - 9}
                    width={36}
                    height={16}
                    rx={8}
                    fill={marker.labelBackground}
                  />
                  <text
                    x={marker.x}
                    y={marker.labelY + 3}
                    textAnchor="middle"
                    fontSize="9.5"
                    fill={marker.labelColor}
                    fontWeight="700"
                  >
                    {marker.label}
                  </text>
                </g>
              ))}

              {microstructureTooltip ? (
                <g pointerEvents="none" aria-label="Microstructure Tooltip">
                  <rect
                    x={microstructureTooltip.x}
                    y={microstructureTooltip.y}
                    width={microstructureTooltip.width}
                    height={microstructureTooltip.height}
                    rx={14}
                    fill="rgba(15,23,42,0.94)"
                    stroke="rgba(148,163,184,0.24)"
                  />
                  {microstructureTooltip.lines.map((line, index) => (
                    <text
                      key={`${line}-${index}`}
                      x={microstructureTooltip.x + 14}
                      y={microstructureTooltip.y + 20 + index * 14}
                      fontSize={index === 0 ? "11.5" : "10.5"}
                      fill={index === 0 ? "#f8fafc" : "#cbd5e1"}
                      fontWeight={index === 0 ? "700" : "500"}
                    >
                      {line}
                    </text>
                  ))}
                </g>
              ) : null}

              {chart.signalMarkers.map((marker) => (
                <g key={`${marker.label}-${marker.x}-${marker.y}`}>
                  <circle
                    cx={marker.x}
                    cy={marker.y}
                    r={5.6}
                    fill={marker.color}
                    stroke="#ffffff"
                    strokeWidth="1.6"
                  />
                  <rect
                    x={marker.x - marker.label.length * 3.6 - 10}
                    y={marker.labelY - 10}
                    width={marker.label.length * 7.2 + 20}
                    height={18}
                    rx={9}
                    fill={marker.labelBackground}
                  />
                  <text
                    x={marker.x}
                    y={marker.labelY + 3}
                    textAnchor="middle"
                    fontSize="9.5"
                    fill={marker.labelColor}
                    fontWeight="700"
                  >
                    {marker.label}
                  </text>
                </g>
              ))}

              {chart.timeLabels.map((label) => (
                <g key={label.x}>
                  <line
                    x1={label.x}
                    y1={CHART_HEIGHT - PADDING_BOTTOM + 4}
                    x2={label.x}
                    y2={CHART_HEIGHT - PADDING_BOTTOM + 10}
                    stroke="rgba(100,116,139,0.6)"
                  />
                  <text
                    x={label.x}
                    y={CHART_HEIGHT - 10}
                    textAnchor="middle"
                    fontSize="11"
                    fill="#64748b"
                  >
                    {label.label}
                  </text>
                </g>
              ))}
            </svg>
          </div>

          <div className="flex flex-wrap gap-2 text-xs">
            <Legend label="EMA20" value={indicator?.ema20} color="bg-emerald-500" />
            <Legend label="EMA50" value={indicator?.ema50} color="bg-rose-500" />
            <Legend label="VWAP" value={indicator?.vwap} color="bg-amber-500" />
            <Legend label="BB Upper" value={indicator?.bollinger_upper} color="bg-sky-500" />
            <Legend label="BB Mid" value={indicator?.bollinger_middle} color="bg-slate-500" />
            <Legend label="BB Lower" value={indicator?.bollinger_lower} color="bg-cyan-400" />
            <Legend label="Support" value={structure?.support} color="bg-emerald-700" />
            <Legend label="Resistance" value={structure?.resistance} color="bg-rose-700" />
            <Legend label="Int Support" value={structure?.internal_support} color="bg-teal-500" />
            <Legend label="Int Resistance" value={structure?.internal_resistance} color="bg-pink-400" />
            <Legend label="Buy Liquidity" value={liquidity?.buy_liquidity} color="bg-teal-600" />
            <Legend label="Sell Liquidity" value={liquidity?.sell_liquidity} color="bg-orange-500" />
            <Legend label="Equal High" value={liquidity?.equal_high} color="bg-amber-700" />
            <Legend label="Equal Low" value={liquidity?.equal_low} color="bg-cyan-700" />
            <Legend label="Micro Events" value={chart.microstructureMarkers.length} color="bg-fuchsia-600" />
            {signal?.signal !== "NEUTRAL" ? (
              <>
                <Legend label="Signal Entry" value={signal?.entry_price} color="bg-sky-700" />
                <Legend label="Signal Target" value={signal?.target_price} color="bg-violet-700" />
                <Legend label="Signal Stop" value={signal?.stop_loss} color="bg-rose-800" />
              </>
            ) : null}
          </div>

          {latestKline ? (
            <div className="grid grid-cols-2 gap-3 text-sm md:grid-cols-4 xl:grid-cols-6">
              <Metric label="Open" value={latestKline.open_price} />
              <Metric label="High" value={latestKline.high_price} />
              <Metric label="Low" value={latestKline.low_price} />
              <Metric label="Close" value={latestKline.close_price} />
              <Metric label="Volume" value={latestKline.volume} digits={4} />
              {indicator ? <Metric label="VWAP" value={indicator.vwap} /> : null}
              {indicator ? <Metric label="EMA20" value={indicator.ema20} /> : null}
              {indicator ? <Metric label="EMA50" value={indicator.ema50} /> : null}
              {structure ? <Metric label="Trend" value={structure.trend} /> : null}
              {structure ? <Metric label="Struct Tier" value={structure.primary_tier || "internal"} /> : null}
              {structure ? <Metric label="Events" value={String(structure.events.length)} /> : null}
              {structure?.internal_support ? <Metric label="Int Support" value={structure.internal_support} /> : null}
              {structure?.internal_resistance ? <Metric label="Int Resist" value={structure.internal_resistance} /> : null}
              <Metric label="Micro Events" value={String(chart.microstructureMarkers.length)} />
              {liquidity ? <Metric label="Sweep" value={liquidity.sweep_type || "none"} /> : null}
              {liquidity ? <Metric label="OB Imb." value={liquidity.order_book_imbalance} digits={3} /> : null}
              {signal ? <Metric label="Signal" value={signal.signal} /> : null}
              {signal ? <Metric label="Entry" value={signal.entry_price} /> : null}
              {signal ? <Metric label="Target" value={signal.target_price} /> : null}
              {signal ? <Metric label="Stop" value={signal.stop_loss} /> : null}
            </div>
          ) : null}
        </div>
      ) : null}
    </section>
  );
}

function Metric({
  label,
  value,
  digits = 2,
}: {
  label: string;
  value: number | string;
  digits?: number;
}) {
  const display = typeof value === "number" ? value.toFixed(digits) : value;

  return (
    <div className="rounded-lg border border-slate-100 bg-slate-50 p-3">
      <p className="text-xs text-muted">{label}</p>
      <p className="mt-1 font-semibold">{display}</p>
    </div>
  );
}

function Legend({
  label,
  value,
  color,
}: {
  label: string;
  value?: number;
  color: string;
}) {
  return (
    <div className="flex items-center gap-2 rounded-full border border-slate-200 bg-white px-3 py-1">
      <span className={`h-2.5 w-2.5 rounded-full ${color}`} />
      <span className="font-medium text-slate-700">{label}</span>
      <span className="text-slate-500">{typeof value === "number" && value > 0 ? value.toFixed(2) : "-"}</span>
    </div>
  );
}

function LayerToggle({
  label,
  active,
  onClick,
}: {
  label: string;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={`rounded-full border px-3 py-1 font-medium transition ${
        active
          ? "border-slate-900 bg-slate-900 text-white"
          : "border-slate-200 bg-white text-slate-600 hover:border-slate-300"
      }`}
    >
      {label}
    </button>
  );
}

function buildChartModel(
  klines: Kline[],
  indicatorSeries: IndicatorSeriesPoint[],
  microstructureEvents: OrderFlowMicrostructureEvent[],
  visibleMicrostructureTypes: string[],
  structure: Structure | null,
  structureSeries: StructureSeriesPoint[],
  liquidity: Liquidity | null,
  liquiditySeries: LiquiditySeriesPoint[],
  signal: Signal | null,
  signalTimeline: SignalTimelinePoint[],
) {
  if (klines.length === 0) {
    return {
      candles: [] as CandleShape[],
      gridLines: [] as GridLine[],
      structureMarkers: [] as StructureMarker[],
      microstructureMarkers: [] as MicrostructureMarker[],
      signalMarkers: [] as SignalMarker[],
      zoneLines: [] as ZoneLine[],
      series: emptySeries(),
      timeLabels: [] as TimeLabel[],
    };
  }

  const plotWidth = CHART_WIDTH - PADDING_LEFT - PADDING_RIGHT;
  const plotHeight = CHART_HEIGHT - PADDING_TOP - PADDING_BOTTOM;
  const slotWidth = plotWidth / klines.length;
  const bodyWidth = clamp(slotWidth * 0.62, 4, 14);

  const ema20 = buildAlignedSeriesValues(
    klines,
    indicatorSeries,
    (point) => point.open_time,
    (point) => point.ema20,
  );
  const ema50 = buildAlignedSeriesValues(
    klines,
    indicatorSeries,
    (point) => point.open_time,
    (point) => point.ema50,
  );
  const bollingerUpper = buildAlignedSeriesValues(
    klines,
    indicatorSeries,
    (point) => point.open_time,
    (point) => point.bollinger_upper,
  );
  const bollingerMiddle = buildAlignedSeriesValues(
    klines,
    indicatorSeries,
    (point) => point.open_time,
    (point) => point.bollinger_middle,
  );
  const bollingerLower = buildAlignedSeriesValues(
    klines,
    indicatorSeries,
    (point) => point.open_time,
    (point) => point.bollinger_lower,
  );
  const vwap = buildAlignedSeriesValues(
    klines,
    indicatorSeries,
    (point) => point.open_time,
    (point) => point.vwap,
  );
  const supportTrack = buildAlignedSeriesValues(
    klines,
    structureSeries,
    (point) => point.open_time,
    (point) => point.support,
  );
  const internalSupportTrack = buildAlignedSeriesValues(
    klines,
    structureSeries,
    (point) => point.open_time,
    (point) => point.internal_support ?? null,
  );
  const internalResistanceTrack = buildAlignedSeriesValues(
    klines,
    structureSeries,
    (point) => point.open_time,
    (point) => point.internal_resistance ?? null,
  );
  const resistanceTrack = buildAlignedSeriesValues(
    klines,
    structureSeries,
    (point) => point.open_time,
    (point) => point.resistance,
  );
  const buyLiquidityTrack = buildAlignedSeriesValues(
    klines,
    liquiditySeries,
    (point) => point.open_time,
    (point) => point.buy_liquidity,
  );
  const sellLiquidityTrack = buildAlignedSeriesValues(
    klines,
    liquiditySeries,
    (point) => point.open_time,
    (point) => point.sell_liquidity,
  );
  const equalHighTrack = buildAlignedSeriesValues(
    klines,
    liquiditySeries,
    (point) => point.open_time,
    (point) => point.equal_high,
  );
  const equalLowTrack = buildAlignedSeriesValues(
    klines,
    liquiditySeries,
    (point) => point.open_time,
    (point) => point.equal_low,
  );

  const priceCandidates = [
    ...klines.flatMap((item) => [item.high_price, item.low_price]),
    ...collectDefinedValues(ema20),
    ...collectDefinedValues(ema50),
    ...collectDefinedValues(bollingerUpper),
    ...collectDefinedValues(bollingerMiddle),
    ...collectDefinedValues(bollingerLower),
    ...collectDefinedValues(vwap),
    ...collectDefinedValues(supportTrack),
    ...collectDefinedValues(resistanceTrack),
    ...collectDefinedValues(internalSupportTrack),
    ...collectDefinedValues(internalResistanceTrack),
    ...collectDefinedValues(buyLiquidityTrack),
    ...collectDefinedValues(sellLiquidityTrack),
    ...collectDefinedValues(equalHighTrack),
    ...collectDefinedValues(equalLowTrack),
    ...collectZoneValues(structure, liquidity),
    ...collectSignalValues(signal),
  ];

  const rawMin = Math.min(...priceCandidates);
  const rawMax = Math.max(...priceCandidates);
  const padding = Math.max((rawMax - rawMin) * 0.08, rawMax * 0.002, 1);
  const priceMin = rawMin - padding;
  const priceMax = rawMax + padding;
  const priceRange = Math.max(priceMax - priceMin, 1);

  const priceToY = (value: number) =>
    PADDING_TOP + ((priceMax - value) / priceRange) * plotHeight;
  const xForIndex = (index: number) => PADDING_LEFT + slotWidth * index + slotWidth / 2;

  const candles = klines.map((item, index) => {
    const x = xForIndex(index);
    const openY = priceToY(item.open_price);
    const closeY = priceToY(item.close_price);
    return {
      openTime: item.open_time,
      x,
      highY: priceToY(item.high_price),
      lowY: priceToY(item.low_price),
      bodyY: Math.min(openY, closeY),
      bodyHeight: Math.max(Math.abs(closeY - openY), 2),
      bodyWidth,
      color: item.close_price >= item.open_price ? "#10b981" : "#f43f5e",
    };
  });

  const gridLines = Array.from({ length: 5 }, (_, index) => {
    const ratio = index / 4;
    const value = priceMax - priceRange * ratio;
    return {
      value,
      y: priceToY(value),
    };
  });

  const timeLabelIndexes = Array.from(
    new Set([0, Math.floor((klines.length - 1) / 2), klines.length - 1]),
  );
  const timeLabels = timeLabelIndexes.map((index) => ({
    x: xForIndex(index),
    label: formatKlineTime(klines[index].open_time),
  }));

  return {
    candles,
    gridLines,
    timeLabels,
    structureMarkers: buildStructureMarkers(structure?.events ?? [], klines, xForIndex, priceToY),
    microstructureMarkers: buildMicrostructureMarkers(
      microstructureEvents,
      visibleMicrostructureTypes,
      klines,
      xForIndex,
      priceToY,
    ),
    signalMarkers: buildSignalMarkers(signal, signalTimeline, klines, xForIndex, priceToY),
    zoneLines: buildZoneLines(structure, liquidity, signal, priceToY),
    series: {
      supportTrack: buildLineSegments(supportTrack, xForIndex, priceToY),
      resistanceTrack: buildLineSegments(resistanceTrack, xForIndex, priceToY),
      internalSupportTrack: buildLineSegments(internalSupportTrack, xForIndex, priceToY),
      internalResistanceTrack: buildLineSegments(internalResistanceTrack, xForIndex, priceToY),
      buyLiquidityTrack: buildLineSegments(buyLiquidityTrack, xForIndex, priceToY),
      sellLiquidityTrack: buildLineSegments(sellLiquidityTrack, xForIndex, priceToY),
      equalHighTrack: buildLineSegments(equalHighTrack, xForIndex, priceToY),
      equalLowTrack: buildLineSegments(equalLowTrack, xForIndex, priceToY),
      ema20: buildLineSegments(ema20, xForIndex, priceToY),
      ema50: buildLineSegments(ema50, xForIndex, priceToY),
      vwap: buildLineSegments(vwap, xForIndex, priceToY),
      bollingerUpper: buildLineSegments(bollingerUpper, xForIndex, priceToY),
      bollingerMiddle: buildLineSegments(bollingerMiddle, xForIndex, priceToY),
      bollingerLower: buildLineSegments(bollingerLower, xForIndex, priceToY),
    },
  };
}

function buildMicrostructureMarkers(
  events: OrderFlowMicrostructureEvent[],
  visibleTypes: string[],
  klines: Kline[],
  xForIndex: (index: number) => number,
  priceToY: (value: number) => number,
): MicrostructureMarker[] {
  if (events.length === 0 || klines.length === 0) {
    return [];
  }

  const visibleTypeSet = new Set(visibleTypes);
  const relevantEvents = events.filter((event) => visibleTypeSet.has(event.type)).slice(-18);
  if (relevantEvents.length === 0) {
    return [];
  }

  const stackByCandle = new Map<number, number>();

  return relevantEvents
    .map((event) => {
      const index = findKlineIndexByTradeTime(klines, event.trade_time);
      if (index < 0) {
        return null;
      }

      const stack = stackByCandle.get(index) ?? 0;
      stackByCandle.set(index, stack + 1);

      const tone = resolveMicrostructureMarkerTone(event.type, event.bias);
      const y = priceToY(event.price > 0 ? event.price : klines[index].close_price);
      const labelOffset = 16 + stack * 12;
      const horizontalOffset = stack % 2 === 0 ? -4 : 4;

      return {
        key: `${event.type}-${event.trade_time}-${event.price}-${stack}`,
        label: resolveMicrostructureMarkerLabel(event.type),
        type: event.type,
        bias: event.bias,
        score: event.score,
        strength: event.strength,
        price: event.price,
        detail: event.detail,
        tradeTime: event.trade_time,
        x: xForIndex(index) + horizontalOffset,
        y,
        labelY: y + (event.bias === "bearish" ? labelOffset : -labelOffset),
        color: tone.color,
        labelColor: tone.labelColor,
        labelBackground: tone.labelBackground,
      };
    })
    .filter((marker): marker is MicrostructureMarker => marker !== null);
}

function buildStructureMarkers(
  events: StructureEvent[],
  klines: Kline[],
  xForIndex: (index: number) => number,
  priceToY: (value: number) => number,
): StructureMarker[] {
  if (events.length === 0 || klines.length === 0) {
    return [];
  }

  const indexByOpenTime = new Map<number, number>();
  klines.forEach((kline, index) => {
    indexByOpenTime.set(kline.open_time, index);
  });

  const markers: StructureMarker[] = [];
  for (const event of events) {
    const index = indexByOpenTime.get(event.open_time);
    if (typeof index !== "number") {
      continue;
    }

    const tone = resolveStructureMarkerTone(event.label, event.kind, event.tier);
    const y = priceToY(event.price);

    markers.push({
      label: resolveStructureMarkerLabel(event.label, event.tier),
      openTime: event.open_time,
      tier: event.tier,
      x: xForIndex(index),
      y,
      labelY: y + resolveStructureMarkerOffset(event.kind, event.tier),
      color: tone.color,
      labelColor: tone.labelColor,
      labelBackground: tone.labelBackground,
    });
  }

  return markers;
}

function buildZoneLines(
  structure: Structure | null,
  liquidity: Liquidity | null,
  signal: Signal | null,
  priceToY: (value: number) => number,
): ZoneLine[] {
  const lines: ZoneLine[] = [];

  addZoneLine(lines, {
    label: "Support",
    value: structure?.support,
    color: "#047857",
    dasharray: "0",
    labelBackground: "#d1fae5",
    labelColor: "#065f46",
  });
  addZoneLine(lines, {
    label: "Resistance",
    value: structure?.resistance,
    color: "#be123c",
    dasharray: "0",
    labelBackground: "#ffe4e6",
    labelColor: "#9f1239",
  });
  addZoneLine(lines, {
    label: "Int Support",
    value: structure?.internal_support,
    color: "#0f766e",
    dasharray: "4 4",
    labelBackground: "#ccfbf1",
    labelColor: "#115e59",
  });
  addZoneLine(lines, {
    label: "Int Resistance",
    value: structure?.internal_resistance,
    color: "#fb7185",
    dasharray: "4 4",
    labelBackground: "#ffe4e6",
    labelColor: "#9f1239",
  });
  addZoneLine(lines, {
    label: "Buy Liquidity",
    value: liquidity?.buy_liquidity,
    color: "#0f766e",
    dasharray: "6 4",
    labelBackground: "#ccfbf1",
    labelColor: "#115e59",
  });
  addZoneLine(lines, {
    label: "Sell Liquidity",
    value: liquidity?.sell_liquidity,
    color: "#ea580c",
    dasharray: "6 4",
    labelBackground: "#ffedd5",
    labelColor: "#9a3412",
  });
  addZoneLine(lines, {
    label: "Equal High",
    value: liquidity?.equal_high,
    color: "#b45309",
    dasharray: "2 4",
    labelBackground: "#fef3c7",
    labelColor: "#92400e",
  });
  addZoneLine(lines, {
    label: "Equal Low",
    value: liquidity?.equal_low,
    color: "#0f766e",
    dasharray: "2 4",
    labelBackground: "#cffafe",
    labelColor: "#155e75",
  });

  const stopClusters = liquidity?.stop_clusters ?? [];
  stopClusters.slice(0, 4).forEach((cluster) => {
    addZoneLine(lines, {
      label: cluster.label,
      value: cluster.price,
      color: cluster.kind.includes("sell") ? "#9a3412" : "#0f766e",
      dasharray: "10 6",
      labelBackground: cluster.kind.includes("sell") ? "#ffedd5" : "#ccfbf1",
      labelColor: cluster.kind.includes("sell") ? "#9a3412" : "#115e59",
    });
  });

  if (signal && signal.signal !== "NEUTRAL") {
    addZoneLine(lines, {
      label: "Signal Entry",
      value: signal.entry_price,
      color: "#1d4ed8",
      dasharray: "0",
      labelBackground: "#dbeafe",
      labelColor: "#1e3a8a",
    });
    addZoneLine(lines, {
      label: "Signal Target",
      value: signal.target_price,
      color: "#6d28d9",
      dasharray: "8 4",
      labelBackground: "#ede9fe",
      labelColor: "#5b21b6",
    });
    addZoneLine(lines, {
      label: "Signal Stop",
      value: signal.stop_loss,
      color: "#be123c",
      dasharray: "8 4",
      labelBackground: "#ffe4e6",
      labelColor: "#9f1239",
    });
  }

  return lines
    .filter((line) => typeof line.value === "number" && line.value > 0)
    .map((line) => ({
      ...line,
      y: priceToY(line.value),
    }));
}

function buildSignalMarkers(
  signal: Signal | null,
  signalTimeline: SignalTimelinePoint[],
  klines: Kline[],
  xForIndex: (index: number) => number,
  priceToY: (value: number) => number,
): SignalMarker[] {
  if (klines.length === 0) {
    return [];
  }

  const indexByOpenTime = new Map<number, number>();
  klines.forEach((kline, index) => {
    indexByOpenTime.set(kline.open_time, index);
  });

  const timeline =
    signalTimeline.length > 0
      ? signalTimeline
      : signal && signal.signal !== "NEUTRAL"
        ? [
            {
              id: signal.id,
              symbol: signal.symbol,
              interval_type: signal.interval_type,
              open_time: signal.open_time,
              signal: signal.signal,
              score: signal.score,
              confidence: signal.confidence,
              entry_price: signal.entry_price,
              stop_loss: signal.stop_loss,
              target_price: signal.target_price,
            } satisfies SignalTimelinePoint,
          ]
        : [];

  return timeline
    .filter((point) => point.signal !== "NEUTRAL")
    .map((point, markerIndex) => {
      const index = indexByOpenTime.get(point.open_time);
      if (typeof index !== "number") {
        return null;
      }

      const tone = resolveSignalMarkerTone(point.signal);
      const y = priceToY(point.entry_price);
      const isLatest = markerIndex === timeline.length - 1;
      return {
        label: point.signal,
        x: xForIndex(index),
        y,
        labelY: y - (isLatest ? 18 : 12),
        color: tone.color,
        labelColor: tone.labelColor,
        labelBackground: tone.labelBackground,
      };
    })
    .filter((marker): marker is SignalMarker => marker !== null);
}

function findKlineIndexByTradeTime(klines: Kline[], tradeTime: number) {
  if (klines.length === 0) {
    return -1;
  }
  if (tradeTime < klines[0].open_time) {
    return -1;
  }

  for (let index = 0; index < klines.length; index++) {
    const current = klines[index];
    const nextOpenTime = klines[index + 1]?.open_time ?? Number.POSITIVE_INFINITY;
    if (tradeTime >= current.open_time && tradeTime < nextOpenTime) {
      return index;
    }
  }

  return klines.length - 1;
}

function addZoneLine(lines: ZoneDraft[], draft: ZoneDraft) {
  if (typeof draft.value !== "number" || draft.value <= 0) {
    return;
  }
  const draftValue = draft.value;
  const exists = lines.some(
    (line) =>
      typeof line.value === "number" &&
      Math.abs(line.value - draftValue) / Math.max(Math.abs(line.value), 1) <= 0.0012,
  );
  if (!exists) {
    lines.push(draft);
  }
}

function collectZoneValues(structure: Structure | null, liquidity: Liquidity | null): number[] {
  const values: number[] = [];
  const candidates = [
    structure?.support,
    structure?.resistance,
    structure?.internal_support,
    structure?.internal_resistance,
    liquidity?.buy_liquidity,
    liquidity?.sell_liquidity,
    liquidity?.equal_high,
    liquidity?.equal_low,
    ...(liquidity?.stop_clusters ?? []).map((cluster) => cluster.price),
    ...(structure?.events ?? []).map((event) => event.price),
  ];

  candidates.forEach((value) => {
    if (typeof value === "number" && value > 0) {
      values.push(value);
    }
  });

  return values;
}

function collectSignalValues(signal: Signal | null): number[] {
  if (!signal || signal.signal === "NEUTRAL") {
    return [];
  }

  return [signal.entry_price, signal.target_price, signal.stop_loss].filter(
    (value) => typeof value === "number" && value > 0,
  );
}

function buildAlignedSeriesValues<T>(
  klines: Kline[],
  points: T[],
  getOpenTime: (point: T) => number,
  getValue: (point: T) => number | null | undefined,
): Array<number | null> {
  const pointMap = new Map<number, T>();
  points.forEach((point) => {
    pointMap.set(getOpenTime(point), point);
  });

  return klines.map((kline) => {
    const point = pointMap.get(kline.open_time);
    if (!point) {
      return null;
    }

    const value = getValue(point);
    if (typeof value !== "number" || Number.isNaN(value) || value <= 0) {
      return null;
    }
    return value;
  });
}

function buildLineSegments(
  values: Array<number | null>,
  xForIndex: (index: number) => number,
  priceToY: (value: number) => number,
): string[] {
  const segments: string[] = [];
  let current: string[] = [];

  values.forEach((value, index) => {
    if (value === null || Number.isNaN(value)) {
      if (current.length > 1) {
        segments.push(current.join(" "));
      }
      current = [];
      return;
    }

    current.push(`${xForIndex(index).toFixed(2)},${priceToY(value).toFixed(2)}`);
  });

  if (current.length > 1) {
    segments.push(current.join(" "));
  }

  return segments;
}

function collectDefinedValues(values: Array<number | null>): number[] {
  return values.filter((value): value is number => value !== null && !Number.isNaN(value));
}

function emptySeries() {
  return {
    supportTrack: [] as string[],
    resistanceTrack: [] as string[],
    internalSupportTrack: [] as string[],
    internalResistanceTrack: [] as string[],
    buyLiquidityTrack: [] as string[],
    sellLiquidityTrack: [] as string[],
    equalHighTrack: [] as string[],
    equalLowTrack: [] as string[],
    ema20: [] as string[],
    ema50: [] as string[],
    vwap: [] as string[],
    bollingerUpper: [] as string[],
    bollingerMiddle: [] as string[],
    bollingerLower: [] as string[],
  };
}

function resolveStructureMarkerTone(label: string, kind: string, tier?: string) {
  if (tier === "external") {
    if (label === "BOS") {
      return {
        color: "#d97706",
        labelColor: "#78350f",
        labelBackground: "#fde68a",
      };
    }
    if (label === "CHOCH") {
      return {
        color: "#1f2937",
        labelColor: "#111827",
        labelBackground: "#cbd5e1",
      };
    }
  }
  if (label === "HH" || label === "HL") {
    return {
      color: "#059669",
      labelColor: "#065f46",
      labelBackground: "#d1fae5",
    };
  }
  if (label === "LH" || label === "LL") {
    return {
      color: "#e11d48",
      labelColor: "#9f1239",
      labelBackground: "#ffe4e6",
    };
  }
  if (label === "BOS") {
    return {
      color: "#f59e0b",
      labelColor: "#92400e",
      labelBackground: "#fef3c7",
    };
  }
  if (label === "CHOCH") {
    return {
      color: "#334155",
      labelColor: "#1e293b",
      labelBackground: "#e2e8f0",
    };
  }
  if (kind === "swing_low") {
    return {
      color: "#0891b2",
      labelColor: "#155e75",
      labelBackground: "#cffafe",
    };
  }
  return {
    color: "#64748b",
    labelColor: "#334155",
    labelBackground: "#e2e8f0",
  };
}

function resolveStructureMarkerLabel(label: string, tier?: string) {
  if (tier === "internal") {
    return `i${label}`;
  }
  return label;
}

function resolveStructureMarkerOffset(kind: string, tier?: string) {
  if (kind === "swing_low") {
    return tier === "internal" ? 18 : 24;
  }
  return tier === "internal" ? -12 : -18;
}

function resolveSignalMarkerTone(action: string) {
  if (action === "BUY") {
    return {
      color: "#2563eb",
      labelColor: "#1e3a8a",
      labelBackground: "#dbeafe",
    };
  }
  return {
    color: "#be123c",
    labelColor: "#9f1239",
    labelBackground: "#ffe4e6",
  };
}

function resolveMicrostructureMarkerLabel(type: string) {
  switch (type) {
    case "absorption":
      return "ABS";
    case "iceberg":
      return "ICE";
    case "iceberg_reload":
      return "IRL";
    case "aggression_burst":
      return "AGR";
    case "initiative_shift":
      return "SHF";
    case "initiative_exhaustion":
      return "IEX";
    case "large_trade_cluster":
      return "LTC";
    case "failed_auction":
      return "FAU";
    case "failed_auction_high_reject":
      return "FAH";
    case "failed_auction_low_reclaim":
      return "FAL";
    case "order_book_migration":
      return "OBM";
    case "order_book_migration_layered":
      return "OBL";
    case "order_book_migration_accelerated":
      return "OBA";
    case "auction_trap_reversal":
      return "TRP";
    case "liquidity_ladder_breakout":
      return "LLB";
    case "migration_auction_flip":
      return "MAF";
    case "absorption_reload_continuation":
      return "ARC";
    case "exhaustion_migration_reversal":
      return "EMR";
    case "microstructure_confluence":
      return "MCF";
    default:
      return "MIC";
  }
}

function resolveMicrostructureMarkerTone(type: string, bias: string) {
  if (type === "microstructure_confluence") {
    return bias === "bearish"
      ? {
          color: "#7f1d1d",
          labelColor: "#991b1b",
          labelBackground: "#fee2e2",
        }
      : {
          color: "#92400e",
          labelColor: "#b45309",
          labelBackground: "#fef3c7",
        };
  }

  if (type === "large_trade_cluster") {
    return bias === "bearish"
      ? {
          color: "#9f1239",
          labelColor: "#881337",
          labelBackground: "#ffe4e6",
        }
      : {
          color: "#7c2d12",
          labelColor: "#9a3412",
          labelBackground: "#ffedd5",
        };
  }

  if (
    type === "failed_auction" ||
    type === "failed_auction_high_reject" ||
    type === "failed_auction_low_reclaim" ||
    type === "initiative_exhaustion"
  ) {
    return bias === "bearish"
      ? {
          color: "#be123c",
          labelColor: "#9f1239",
          labelBackground: "#ffe4e6",
        }
      : {
          color: "#0f766e",
          labelColor: "#0f766e",
          labelBackground: "#ccfbf1",
        };
  }

  if (type === "initiative_shift") {
    return bias === "bearish"
      ? {
          color: "#7c3aed",
          labelColor: "#5b21b6",
          labelBackground: "#ede9fe",
        }
      : {
          color: "#0369a1",
          labelColor: "#075985",
          labelBackground: "#e0f2fe",
        };
  }

  if (
    type === "order_book_migration" ||
    type === "order_book_migration_layered" ||
    type === "order_book_migration_accelerated"
  ) {
    return bias === "bearish"
      ? {
          color: "#6d28d9",
          labelColor: "#5b21b6",
          labelBackground: "#ede9fe",
        }
      : {
          color: "#075985",
          labelColor: "#0369a1",
          labelBackground: "#e0f2fe",
        };
  }

  if (
    type === "auction_trap_reversal" ||
    type === "liquidity_ladder_breakout" ||
    type === "migration_auction_flip" ||
    type === "absorption_reload_continuation" ||
    type === "exhaustion_migration_reversal"
  ) {
    return bias === "bearish"
      ? {
          color: "#7c2d12",
          labelColor: "#9a3412",
          labelBackground: "#ffedd5",
        }
      : {
          color: "#155e75",
          labelColor: "#0f766e",
          labelBackground: "#ccfbf1",
        };
  }

  if (type === "iceberg" || type === "iceberg_reload") {
    return bias === "bearish"
      ? {
          color: "#c2410c",
          labelColor: "#9a3412",
          labelBackground: "#ffedd5",
        }
      : {
          color: "#7c3aed",
          labelColor: "#5b21b6",
          labelBackground: "#ede9fe",
        };
  }

  if (type === "aggression_burst") {
    return bias === "bearish"
      ? {
          color: "#e11d48",
          labelColor: "#9f1239",
          labelBackground: "#ffe4e6",
        }
      : {
          color: "#0f766e",
          labelColor: "#115e59",
          labelBackground: "#ccfbf1",
        };
  }

  return bias === "bearish"
    ? {
        color: "#be123c",
        labelColor: "#9f1239",
        labelBackground: "#ffe4e6",
      }
    : {
        color: "#2563eb",
        labelColor: "#1e3a8a",
        labelBackground: "#dbeafe",
      };
}

function formatMicrostructureEventType(type: string) {
  return type
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function buildMicrostructureTooltip(marker: MicrostructureMarker): MicrostructureTooltip {
  const detailLines = wrapTooltipText(marker.detail, 24, 2);
  const lines = [
    `${formatMicrostructureEventType(marker.type)} · ${marker.label}`,
    `${marker.bias.toUpperCase()} | score ${marker.score > 0 ? `+${marker.score}` : marker.score}`,
    `strength ${marker.strength.toFixed(2)} | price ${marker.price.toFixed(2)}`,
    ...detailLines,
  ];
  const width = 226;
  const height = 14 + lines.length * 14;
  const x = clamp(marker.x + 16, 12, CHART_WIDTH - width - 12);
  const preferredY = marker.y > CHART_HEIGHT / 2 ? marker.y - height - 12 : marker.y + 14;
  const y = clamp(preferredY, 12, CHART_HEIGHT - height - 12);

  return { x, y, width, height, lines };
}

function wrapTooltipText(text: string, maxChars: number, maxLines: number) {
  const chars = Array.from(text);
  const lines: string[] = [];

  for (let index = 0; index < chars.length && lines.length < maxLines; index += maxChars) {
    const slice = chars.slice(index, index + maxChars).join("");
    lines.push(slice);
  }

  if (chars.length > maxChars * maxLines && lines.length > 0) {
    lines[lines.length - 1] = `${lines[lines.length - 1].slice(0, Math.max(maxChars - 1, 1))}…`;
  }

  return lines;
}

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

function formatKlineTime(timestamp: number): string {
  return new Date(timestamp).toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
  });
}

interface CandleShape {
  openTime: number;
  x: number;
  highY: number;
  lowY: number;
  bodyY: number;
  bodyHeight: number;
  bodyWidth: number;
  color: string;
}

interface GridLine {
  value: number;
  y: number;
}

interface TimeLabel {
  x: number;
  label: string;
}

interface StructureMarker {
  label: string;
  openTime: number;
  tier?: string;
  x: number;
  y: number;
  labelY: number;
  color: string;
  labelColor: string;
  labelBackground: string;
}

interface ZoneDraft {
  label: string;
  value?: number;
  color: string;
  dasharray: string;
  labelBackground: string;
  labelColor: string;
}

interface ZoneLine extends ZoneDraft {
  value: number;
  y: number;
}

interface SignalMarker {
  label: string;
  x: number;
  y: number;
  labelY: number;
  color: string;
  labelColor: string;
  labelBackground: string;
}

interface MicrostructureMarker {
  key: string;
  label: string;
  type: string;
  bias: string;
  score: number;
  strength: number;
  price: number;
  detail: string;
  tradeTime: number;
  x: number;
  y: number;
  labelY: number;
  color: string;
  labelColor: string;
  labelBackground: string;
}

interface MicrostructureTooltip {
  x: number;
  y: number;
  width: number;
  height: number;
  lines: string[];
}
