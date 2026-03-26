import {
  IndicatorSeriesPoint,
  Kline,
  Liquidity,
  LiquiditySeriesPoint,
  OrderFlowMicrostructureEvent,
  Structure,
  StructureSeriesPoint,
} from "@/types/market";
import { Signal, SignalTimelinePoint } from "@/types/signal";
import {
  CHART_HEIGHT,
  CHART_WIDTH,
  CandleShape,
  type ChartModel,
  type ChartSeries,
  GridLine,
  MicrostructureMarker,
  PADDING_BOTTOM,
  PADDING_LEFT,
  PADDING_RIGHT,
  PADDING_TOP,
  SignalMarker,
  StructureMarker,
  TimeLabel,
  ZoneLine,
} from "./chartTypes";
import { clamp, formatKlineTime } from "./chartHelpers";
import {
  buildMicrostructureMarkers,
  buildSignalMarkers,
  buildStructureMarkers,
  buildZoneLines,
} from "./chartMarkers";

export function buildChartModel(
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
): ChartModel {
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
      closeY,
      bodyY: Math.min(openY, closeY),
      bodyHeight: Math.max(Math.abs(closeY - openY), 2),
      bodyWidth,
      hitBoxWidth: Math.max(slotWidth, 14),
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

function emptySeries(): ChartSeries {
  return {
    supportTrack: [],
    resistanceTrack: [],
    internalSupportTrack: [],
    internalResistanceTrack: [],
    buyLiquidityTrack: [],
    sellLiquidityTrack: [],
    equalHighTrack: [],
    equalLowTrack: [],
    ema20: [],
    ema50: [],
    vwap: [],
    bollingerUpper: [],
    bollingerMiddle: [],
    bollingerLower: [],
  };
}

// Re-export helpers for backward-compatible imports
export {
  buildLegendItems,
  buildMicrostructureTooltip,
  clamp,
  formatCandleAriaLabel,
  formatKlineTime,
  formatMicrostructureEventType,
  resolveLegendFocusLabel,
  resolveMicrostructureMarkerLabel,
  resolveZoneLineOpacity,
} from "./chartHelpers";
export type { LegendItem } from "./chartHelpers";
