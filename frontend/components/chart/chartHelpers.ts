import {
  formatBiasLabel,
  formatMicrostructureEventTypeLabel,
} from "@/lib/uiLabels";
import { Kline } from "@/types/market";
import {
  CHART_HEIGHT,
  CHART_WIDTH,
  MicrostructureMarker,
  MicrostructureTooltip,
} from "./chartTypes";

export { resolveMicrostructureMarkerLabel } from "./chartStyles";

export function formatMicrostructureEventType(type: string) {
  return formatMicrostructureEventTypeLabel(type);
}

export function buildMicrostructureTooltip(marker: MicrostructureMarker): MicrostructureTooltip {
  const detailLines = wrapTooltipText(marker.detail, 24, 2);
  const lines = [
    `${formatMicrostructureEventType(marker.type)} · ${marker.label}`,
    `${formatBiasLabel(marker.bias)} | 评分 ${marker.score > 0 ? `+${marker.score}` : marker.score}`,
    `强度 ${marker.strength.toFixed(2)} | 价格 ${marker.price.toFixed(2)}`,
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

export function resolveZoneLineOpacity(label: string, focusedLegendKey: string | null) {
  if (!focusedLegendKey) {
    return 1;
  }

  const upperLabel = label.toUpperCase();
  if (
    focusedLegendKey === "structure" &&
    (upperLabel.includes("支撑") ||
      upperLabel.includes("阻力") ||
      upperLabel.includes("SUPPORT") ||
      upperLabel.includes("RESIST") ||
      upperLabel.includes("BOS") ||
      upperLabel.includes("CHOCH"))
  ) {
    return 1;
  }
  if (
    focusedLegendKey === "liquidity" &&
    (upperLabel.includes("流动性") ||
      upperLabel.includes("等高") ||
      upperLabel.includes("等低") ||
      upperLabel.includes("LIQ") ||
      upperLabel.includes("EQH") ||
      upperLabel.includes("EQL") ||
      upperLabel.includes("SWEEP"))
  ) {
    return 1;
  }
  if (
    focusedLegendKey === "signal" &&
    (upperLabel.includes("进场") ||
      upperLabel.includes("目标") ||
      upperLabel.includes("止损") ||
      upperLabel.includes("ENTRY") ||
      upperLabel.includes("TARGET") ||
      upperLabel.includes("STOP"))
  ) {
    return 1;
  }

  return 0.18;
}

export function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

export function formatKlineTime(timestamp: number): string {
  return new Date(timestamp).toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatCandleAriaLabel(kline?: Kline) {
  if (!kline) {
    return "未知";
  }
  return `${formatKlineTime(kline.open_time)} 开 ${kline.open_price.toFixed(2)} 高 ${kline.high_price.toFixed(
    2,
  )} 低 ${kline.low_price.toFixed(2)} 收 ${kline.close_price.toFixed(2)}`;
}

export function resolveLegendFocusLabel(key: string) {
  switch (key) {
    case "indicators":
      return "指标图层";
    case "structure":
      return "结构图层";
    case "liquidity":
      return "流动性图层";
    case "signal":
      return "信号标记";
    case "micro":
      return "微结构";
    default:
      return "全部图层";
  }
}

export interface LegendItem {
  label: string;
  value?: number | string;
  color: string;
  group: "indicators" | "structure" | "liquidity" | "signal" | "micro";
}

export function buildLegendItems(
  indicator: { ema20: number; ema50: number; vwap: number; bollinger_upper: number; bollinger_middle: number; bollinger_lower: number } | null,
  structure: { support: number; resistance: number; internal_support?: number; internal_resistance?: number } | null,
  liquidity: { buy_liquidity: number; sell_liquidity: number; equal_high: number; equal_low: number } | null,
  signal: { signal: string; entry_price: number; target_price: number; stop_loss: number } | null,
  microstructureCount: number,
): LegendItem[] {
  const items: LegendItem[] = [
    { label: "EMA20", value: indicator?.ema20, color: "bg-emerald-500", group: "indicators" },
    { label: "EMA50", value: indicator?.ema50, color: "bg-rose-500", group: "indicators" },
    { label: "VWAP", value: indicator?.vwap, color: "bg-amber-500", group: "indicators" },
    { label: "布林上轨", value: indicator?.bollinger_upper, color: "bg-sky-500", group: "indicators" },
    { label: "布林中轨", value: indicator?.bollinger_middle, color: "bg-slate-500", group: "indicators" },
    { label: "布林下轨", value: indicator?.bollinger_lower, color: "bg-cyan-400", group: "indicators" },
    { label: "支撑位", value: structure?.support, color: "bg-emerald-700", group: "structure" },
    { label: "阻力位", value: structure?.resistance, color: "bg-rose-700", group: "structure" },
    { label: "内部支撑", value: structure?.internal_support, color: "bg-teal-500", group: "structure" },
    { label: "内部阻力", value: structure?.internal_resistance, color: "bg-pink-400", group: "structure" },
    { label: "买方流动性", value: liquidity?.buy_liquidity, color: "bg-teal-600", group: "liquidity" },
    { label: "卖方流动性", value: liquidity?.sell_liquidity, color: "bg-orange-500", group: "liquidity" },
    { label: "等高点", value: liquidity?.equal_high, color: "bg-amber-700", group: "liquidity" },
    { label: "等低点", value: liquidity?.equal_low, color: "bg-cyan-700", group: "liquidity" },
    { label: "微结构事件", value: String(microstructureCount), color: "bg-fuchsia-600", group: "micro" },
  ];
  if (signal && signal.signal !== "NEUTRAL") {
    items.push(
      { label: "信号进场", value: signal.entry_price, color: "bg-sky-700", group: "signal" },
      { label: "信号目标", value: signal.target_price, color: "bg-violet-700", group: "signal" },
      { label: "信号止损", value: signal.stop_loss, color: "bg-rose-800", group: "signal" },
    );
  }
  return items;
}
