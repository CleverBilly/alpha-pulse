import {
  Kline,
  Liquidity,
  OrderFlowMicrostructureEvent,
  Structure,
  StructureEvent,
} from "@/types/market";
import { Signal, SignalTimelinePoint } from "@/types/signal";
import {
  MicrostructureMarker,
  SignalMarker,
  StructureMarker,
  ZoneDraft,
  ZoneLine,
} from "./chartTypes";
import {
  resolveMicrostructureMarkerLabel,
  resolveMicrostructureMarkerTone,
  resolveSignalMarkerTone,
  resolveStructureMarkerLabel,
  resolveStructureMarkerOffset,
  resolveStructureMarkerTone,
} from "./chartStyles";

export function buildMicrostructureMarkers(
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

export function buildStructureMarkers(
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

export function buildZoneLines(
  structure: Structure | null,
  liquidity: Liquidity | null,
  signal: Signal | null,
  priceToY: (value: number) => number,
): ZoneLine[] {
  const lines: ZoneDraft[] = [];

  addZoneLine(lines, {
    label: "支撑位",
    value: structure?.support,
    color: "#047857",
    dasharray: "0",
    labelBackground: "#d1fae5",
    labelColor: "#065f46",
  });
  addZoneLine(lines, {
    label: "阻力位",
    value: structure?.resistance,
    color: "#be123c",
    dasharray: "0",
    labelBackground: "#ffe4e6",
    labelColor: "#9f1239",
  });
  addZoneLine(lines, {
    label: "内部支撑",
    value: structure?.internal_support,
    color: "#0f766e",
    dasharray: "4 4",
    labelBackground: "#ccfbf1",
    labelColor: "#115e59",
  });
  addZoneLine(lines, {
    label: "内部阻力",
    value: structure?.internal_resistance,
    color: "#fb7185",
    dasharray: "4 4",
    labelBackground: "#ffe4e6",
    labelColor: "#9f1239",
  });
  addZoneLine(lines, {
    label: "买方流动性",
    value: liquidity?.buy_liquidity,
    color: "#0f766e",
    dasharray: "6 4",
    labelBackground: "#ccfbf1",
    labelColor: "#115e59",
  });
  addZoneLine(lines, {
    label: "卖方流动性",
    value: liquidity?.sell_liquidity,
    color: "#ea580c",
    dasharray: "6 4",
    labelBackground: "#ffedd5",
    labelColor: "#9a3412",
  });
  addZoneLine(lines, {
    label: "等高点",
    value: liquidity?.equal_high,
    color: "#b45309",
    dasharray: "2 4",
    labelBackground: "#fef3c7",
    labelColor: "#92400e",
  });
  addZoneLine(lines, {
    label: "等低点",
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
      label: "信号进场",
      value: signal.entry_price,
      color: "#1d4ed8",
      dasharray: "0",
      labelBackground: "#dbeafe",
      labelColor: "#1e3a8a",
    });
    addZoneLine(lines, {
      label: "信号目标",
      value: signal.target_price,
      color: "#6d28d9",
      dasharray: "8 4",
      labelBackground: "#ede9fe",
      labelColor: "#5b21b6",
    });
    addZoneLine(lines, {
      label: "信号止损",
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
      y: priceToY(line.value!),
    })) as ZoneLine[];
}

export function buildSignalMarkers(
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
