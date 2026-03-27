"use client";

import { useMemo, useState, useEffect } from "react";
import { Typography } from "antd";
import { useMarketStore } from "@/store/marketStore";
import { alertApi } from "@/services/apiClient";
import { isLongDirection } from "@/utils/alertUtils";
import {
  CHART_HEIGHT,
  CHART_WIDTH,
  ChartCoords,
  ActiveSignal,
  LegendFocusKey,
  PADDING_BOTTOM,
  PADDING_LEFT,
  PADDING_RIGHT,
  PADDING_TOP,
} from "./chartTypes";
import type { Kline } from "@/types/market";
import { LayerToggle, Legend, toggleLegendFocus } from "./KlineChartControls";
import {
  buildChartModel,
  buildLegendItems,
  buildMicrostructureTooltip,
  resolveZoneLineOpacity,
  resolveLegendFocusLabel,
} from "./chartModel";
import KlineCandleLayer from "./KlineCandleLayer";
import StructureLiquidityLayer from "./StructureLiquidityLayer";
import SignalOverlayLayer from "./SignalOverlayLayer";
import KlineInfoPanels from "./KlineInfoPanels";

interface HistoricalMode {
  klines: Kline[];
  symbol: string;
  interval: string;
}

interface KlineChartProps {
  historicalMode?: HistoricalMode;
  activeSignal?: ActiveSignal | null;
}

const PRIMARY_MICROSTRUCTURE_TYPES = ["absorption", "iceberg", "aggression_burst"] as const;
const SECONDARY_MICROSTRUCTURE_LAYERS = [
  { key: "initiative_shift", label: "主动性切换", types: ["initiative_shift"] },
  { key: "large_trade_cluster", label: "大单簇", types: ["large_trade_cluster"] },
  {
    key: "reload_exhaustion",
    label: "回补 / 衰竭",
    types: ["iceberg_reload", "initiative_exhaustion"],
  },
  {
    key: "failed_auction",
    label: "失败拍卖",
    types: ["failed_auction", "failed_auction_high_reject", "failed_auction_low_reclaim"],
  },
  {
    key: "order_book_migration",
    label: "订单簿迁移",
    types: ["order_book_migration", "order_book_migration_layered", "order_book_migration_accelerated"],
  },
  {
    key: "composite_patterns",
    label: "复合形态",
    types: [
      "auction_trap_reversal",
      "liquidity_ladder_breakout",
      "migration_auction_flip",
      "absorption_reload_continuation",
      "exhaustion_migration_reversal",
    ],
  },
  { key: "microstructure_confluence", label: "微结构共振", types: ["microstructure_confluence"] },
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

export default function KlineChart({ historicalMode, activeSignal: activeSignalProp }: KlineChartProps = {}) {
  const {
    symbol,
    klines: storeKlines,
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

  // 无条件调用 hooks，然后条件选择数据（遵守 React Hooks Rules）
  const klines = historicalMode ? historicalMode.klines : storeKlines;
  const [enabledLayers, setEnabledLayers] = useState<Record<SecondaryLayerKey, boolean>>(DEFAULT_SECONDARY_LAYER_STATE);
  const [hoveredCandleIndex, setHoveredCandleIndex] = useState<number | null>(null);
  const [hoveredMicrostructureMarkerKey, setHoveredMicrostructureMarkerKey] = useState<string | null>(null);
  const [pinnedMicrostructureMarkerKey, setPinnedMicrostructureMarkerKey] = useState<string | null>(null);
  const [focusedLegendKey, setFocusedLegendKey] = useState<LegendFocusKey | null>(null);
  const [fetchedActiveSignal, setFetchedActiveSignal] = useState<ActiveSignal | null>(null);
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

  // historicalMode 时使用外部传入的 activeSignal，否则从 API 拉取最近告警
  useEffect(() => {
    if (historicalMode) return;
    let active = true;
    alertApi
      .getAlertHistory(20)
      .then((feed) => {
        if (!active) return;
        const match = feed.items
          .filter(
            (item) =>
              item.symbol === symbol && item.kind === "setup_ready" && item.entry_price > 0,
          )
          .sort((a, b) => b.created_at - a.created_at)[0];
        if (!match) {
          setFetchedActiveSignal(null);
          return;
        }
        setFetchedActiveSignal({
          entryPrice: match.entry_price,
          stopLoss: match.stop_loss,
          targetPrice: match.target_price,
          direction: isLongDirection(match.verdict, match.tradeability_label) ? "long" : "short",
        });
      })
      .catch(() => {
        /* 静默降级，不影响主图表渲染 */
      });
    return () => {
      active = false;
    };
  }, [symbol, historicalMode]);

  // 最终使用的 activeSignal：historicalMode 时优先外部 prop，否则使用内部拉取的
  const activeSignal = historicalMode ? (activeSignalProp ?? null) : fetchedActiveSignal;

  const coords = useMemo<ChartCoords | null>(() => {
    if (visibleKlines.length === 0) return null;
    const plotHeight = CHART_HEIGHT - PADDING_TOP - PADDING_BOTTOM;
    const allPrices = visibleKlines.flatMap((k) => [k.high_price, k.low_price]);
    const rawMin = Math.min(...allPrices);
    const rawMax = Math.max(...allPrices);
    const pricePadding = Math.max((rawMax - rawMin) * 0.08, rawMax * 0.002, 1);
    const priceMin = rawMin - pricePadding;
    const priceMax = rawMax + pricePadding;
    const priceRange = Math.max(priceMax - priceMin, 1);
    const plotWidth = CHART_WIDTH - PADDING_LEFT - PADDING_RIGHT;
    const slotWidth = plotWidth / visibleKlines.length;
    return {
      toX: (index: number) => PADDING_LEFT + slotWidth * index + slotWidth / 2,
      toY: (price: number) => PADDING_TOP + ((priceMax - price) / priceRange) * plotHeight,
      candleWidth: slotWidth,
      chartWidth: CHART_WIDTH,
      chartHeight: CHART_HEIGHT,
      paddingTop: PADDING_TOP,
      paddingRight: PADDING_RIGHT,
      paddingBottom: PADDING_BOTTOM,
      paddingLeft: PADDING_LEFT,
      priceMin,
      priceMax,
    };
  }, [visibleKlines]);

  // historicalMode: indicator series not available for historical timestamps, show clean K-line only.
  // Passing empty arrays explicitly makes this intentional rather than a silent data mismatch.
  const chart = useMemo(
    () =>
      buildChartModel(
        visibleKlines,
        historicalMode ? [] : indicatorSeries,
        historicalMode ? [] : microstructureEvents,
        visibleMicrostructureTypes,
        historicalMode ? null : structure,
        historicalMode ? [] : structureSeries,
        historicalMode ? null : liquidity,
        historicalMode ? [] : liquiditySeries,
        historicalMode ? null : signal,
        historicalMode ? [] : signalTimeline,
      ),
    [
      visibleKlines,
      historicalMode,
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
      chart.microstructureMarkers.find(
        (marker) => marker.key === (pinnedMicrostructureMarkerKey ?? hoveredMicrostructureMarkerKey),
      ) ?? null,
    [hoveredMicrostructureMarkerKey, pinnedMicrostructureMarkerKey, chart.microstructureMarkers],
  );
  const hoveredKline = hoveredCandleIndex !== null ? visibleKlines[hoveredCandleIndex] ?? null : null;
  const hoveredCandle = hoveredCandleIndex !== null ? chart.candles[hoveredCandleIndex] ?? null : null;
  const latestKline = visibleKlines[visibleKlines.length - 1] ?? null;
  const currentSymbol = historicalMode?.symbol ?? symbol;
  const enabledLayerCount = Object.values(enabledLayers).filter(Boolean).length;
  const microstructureTooltip = activeMicrostructureMarker
    ? buildMicrostructureTooltip(activeMicrostructureMarker)
    : null;
  const legendOpacity = (key: LegendFocusKey) => (focusedLegendKey === null || focusedLegendKey === key ? 1 : 0.16);
  const highlightedLegendLabel = focusedLegendKey ? resolveLegendFocusLabel(focusedLegendKey) : "全部图层";

  return (
    <section className="kline-screen" data-testid="kline-main-screen">
      <div className="surface-panel surface-panel--paper kline-screen__surface">
        {!historicalMode && (
          <div className="kline-screen__header" data-testid="kline-screen-header">
            <div className="kline-screen__header-copy">
              <p className="kline-screen__eyebrow">主控屏</p>
              <Typography.Title level={3} className="!mb-0 !text-[24px] !tracking-[-0.03em]">
                K 线图
              </Typography.Title>
              <p className="kline-screen__description">
                48 根 K 线，叠加结构点、动态支撑阻力、流动性轨迹、信号位与多指标
              </p>
            </div>

            <div className="kline-screen__readouts">
              <div className="kline-screen__readout">
                <span>标的</span>
                <strong>{currentSymbol}</strong>
              </div>
              <div className="kline-screen__readout">
                <span>样本</span>
                <strong>{visibleKlines.length} 根</strong>
              </div>
              <div className="kline-screen__readout">
                <span>微结构</span>
                <strong>{chart.microstructureMarkers.length} 个</strong>
              </div>
              <button
                type="button"
                onClick={() => {
                  void refreshDashboard(true);
                }}
                className="kline-screen__refresh"
              >
                更新K线
              </button>
            </div>
          </div>
        )}

        {!historicalMode && loading && visibleKlines.length === 0 ? <p className="text-sm text-muted">加载中...</p> : null}
        {!historicalMode && error ? <p className="text-sm text-negative">{error}</p> : null}
        {!historicalMode && !loading && !error && visibleKlines.length === 0 ? (
          <p className="text-sm text-muted">暂无 K 线数据</p>
        ) : null}

        {visibleKlines.length > 0 ? (
          <div className="kline-screen__body">
            <div className="kline-screen__controls" data-testid="kline-screen-controls">
              <div className="kline-screen__controls-copy">
                <span className="kline-screen__controls-label">核心图层</span>
                <strong>ABS / ICE / AGR 常驻，次级信号按需展开</strong>
              </div>
              <div className="kline-screen__controls-row">
                <span className="kline-screen__controls-state">已展开 {enabledLayerCount} 个次级层</span>
                <div className="kline-screen__toggle-grid">
                  {SECONDARY_MICROSTRUCTURE_LAYERS.map((layer) => (
                    <LayerToggle
                      key={layer.key}
                      label={layer.label}
                      active={enabledLayers[layer.key]}
                      onClick={() => {
                        setEnabledLayers((value) => ({ ...value, [layer.key]: !value[layer.key] }));
                        setHoveredMicrostructureMarkerKey(null);
                        setPinnedMicrostructureMarkerKey(null);
                      }}
                    />
                  ))}
                </div>
              </div>
            </div>

            <div className="kline-screen__viewport-shell">
              <div className="kline-screen__viewport-topline">
                <div>
                  <p className="kline-screen__viewport-label">主屏视窗</p>
                  <p className="kline-screen__viewport-caption">当前聚焦 {highlightedLegendLabel}</p>
                </div>
                <span className="kline-screen__viewport-chip">KLINE / LIVE</span>
              </div>

              <div className="kline-screen__viewport" data-testid="kline-screen-viewport">
              <svg viewBox={`0 0 ${CHART_WIDTH} ${CHART_HEIGHT}`} className="h-[360px] w-full">
                <defs>
                  <linearGradient id="chart-bg" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="rgba(56,189,248,0.12)" />
                    <stop offset="100%" stopColor="rgba(15,23,42,0)" />
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
                    <text x={CHART_WIDTH - PADDING_RIGHT + 8} y={line.y + 4} fontSize="11" fill="#64748b">
                      {line.value.toFixed(2)}
                    </text>
                  </g>
                ))}

                {chart.zoneLines.map((line) => {
                  const opacity = resolveZoneLineOpacity(line.label, focusedLegendKey);
                  return (
                    <g key={`${line.label}-${line.y}`} opacity={opacity}>
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
                      <text x={CHART_WIDTH - PADDING_RIGHT + 16} y={line.y + 4} fontSize="10" fill={line.labelColor}>
                        {line.label}
                      </text>
                    </g>
                  );
                })}

                <StructureLiquidityLayer
                  structureMarkers={chart.structureMarkers}
                  series={chart.series}
                  structureOpacity={legendOpacity("structure")}
                  liquidityOpacity={legendOpacity("liquidity")}
                />

                <g opacity={legendOpacity("indicators")}>
                  {chart.series.bollingerUpper.map((points, index) => (
                    <polyline key={`bb-upper-${index}`} points={points} fill="none" stroke="#38bdf8" strokeWidth="1.5" strokeDasharray="4 4" />
                  ))}
                  {chart.series.bollingerMiddle.map((points, index) => (
                    <polyline key={`bb-middle-${index}`} points={points} fill="none" stroke="#64748b" strokeWidth="1.3" />
                  ))}
                  {chart.series.bollingerLower.map((points, index) => (
                    <polyline key={`bb-lower-${index}`} points={points} fill="none" stroke="#7dd3fc" strokeWidth="1.5" strokeDasharray="4 4" />
                  ))}
                  {chart.series.vwap.map((points, index) => (
                    <polyline key={`vwap-${index}`} points={points} fill="none" stroke="#f59e0b" strokeWidth="1.8" />
                  ))}
                  {chart.series.ema20.map((points, index) => (
                    <polyline key={`ema20-${index}`} points={points} fill="none" stroke="#10b981" strokeWidth="2" />
                  ))}
                  {chart.series.ema50.map((points, index) => (
                    <polyline key={`ema50-${index}`} points={points} fill="none" stroke="#f43f5e" strokeWidth="2" />
                  ))}
                </g>

                <KlineCandleLayer
                  klines={visibleKlines}
                  candles={chart.candles}
                  hoveredIndex={hoveredCandleIndex}
                  setHoveredIndex={setHoveredCandleIndex}
                  hoveredKline={hoveredKline}
                  hoveredCandle={hoveredCandle}
                />

                <SignalOverlayLayer
                  signalMarkers={chart.signalMarkers}
                  microstructureMarkers={chart.microstructureMarkers}
                  microstructureTooltip={microstructureTooltip}
                  hoveredMicrostructureMarkerKey={hoveredMicrostructureMarkerKey}
                  pinnedMicrostructureMarkerKey={pinnedMicrostructureMarkerKey}
                  setHoveredMicrostructureMarkerKey={setHoveredMicrostructureMarkerKey}
                  setPinnedMicrostructureMarkerKey={setPinnedMicrostructureMarkerKey}
                  signalOpacity={legendOpacity("signal")}
                  microOpacity={legendOpacity("micro")}
                  activeSignal={activeSignal}
                  coords={coords}
                />

                {chart.timeLabels.map((label) => (
                  <g key={label.x}>
                    <line x1={label.x} y1={CHART_HEIGHT - PADDING_BOTTOM + 4} x2={label.x} y2={CHART_HEIGHT - PADDING_BOTTOM + 10} stroke="rgba(100,116,139,0.6)" />
                    <text x={label.x} y={CHART_HEIGHT - 10} textAnchor="middle" fontSize="11" fill="#64748b">
                      {label.label}
                    </text>
                  </g>
                ))}
              </svg>
            </div>
            </div>

            <div className="kline-screen__legend-dock">
              <div>
                <p className="kline-screen__dock-label">图例聚焦</p>
                <p className="kline-screen__dock-copy">
                  当前聚焦 <span className="font-semibold text-slate-950">{highlightedLegendLabel}</span>。点击图例可以隔离图层。
                </p>
              </div>
              <button
                type="button"
                onClick={() => setFocusedLegendKey(null)}
                className="kline-screen__dock-reset"
              >
                重置视图
              </button>
            </div>

            <div className="kline-screen__legend-strip">
              {buildLegendItems(indicator, structure, liquidity, signal, chart.microstructureMarkers.length).map((item) => (
                <Legend
                  key={item.label}
                  label={item.label}
                  value={item.value}
                  color={item.color}
                  focused={focusedLegendKey === item.group}
                  dimmed={focusedLegendKey !== null && focusedLegendKey !== item.group}
                  onClick={() => toggleLegendFocus(setFocusedLegendKey, item.group)}
                />
              ))}
            </div>

            <div className="kline-screen__info-dock">
              <KlineInfoPanels
                hoveredKline={hoveredKline}
                latestKline={latestKline}
                activeMicrostructureMarker={activeMicrostructureMarker}
                pinnedMicrostructureMarkerKey={pinnedMicrostructureMarkerKey}
                setHoveredMicrostructureMarkerKey={setHoveredMicrostructureMarkerKey}
                setPinnedMicrostructureMarkerKey={setPinnedMicrostructureMarkerKey}
                indicator={indicator}
                structure={structure}
                liquidity={liquidity}
                signal={signal}
                microstructureMarkerCount={chart.microstructureMarkers.length}
              />
            </div>
          </div>
        ) : null}
      </div>
    </section>
  );
}
