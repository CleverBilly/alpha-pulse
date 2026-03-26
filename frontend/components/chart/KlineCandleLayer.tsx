import type { Dispatch, SetStateAction } from "react";
import { Kline } from "@/types/market";
import {
  CandleShape,
  CHART_HEIGHT,
  CHART_WIDTH,
  PADDING_BOTTOM,
  PADDING_RIGHT,
  PADDING_TOP,
} from "./chartTypes";
import { clamp, formatCandleAriaLabel, formatKlineTime } from "./chartHelpers";

interface KlineCandleLayerProps {
  klines: Kline[];
  candles: CandleShape[];
  hoveredIndex: number | null;
  setHoveredIndex: Dispatch<SetStateAction<number | null>>;
  hoveredKline: Kline | null;
  hoveredCandle: CandleShape | null;
}

export default function KlineCandleLayer({
  klines,
  candles,
  hoveredIndex,
  setHoveredIndex,
  hoveredKline,
  hoveredCandle,
}: KlineCandleLayerProps) {
  return (
    <g>
      {candles.map((candle, index) => (
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
          <rect
            x={candle.x - candle.hitBoxWidth / 2}
            y={PADDING_TOP}
            width={candle.hitBoxWidth}
            height={CHART_HEIGHT - PADDING_TOP - PADDING_BOTTOM}
            fill="transparent"
            role="button"
            tabIndex={0}
            aria-label={`K线 ${formatCandleAriaLabel(klines[index])}`}
            onMouseEnter={() => setHoveredIndex(index)}
            onMouseLeave={() => setHoveredIndex((value) => (value === index ? null : value))}
            onFocus={() => setHoveredIndex(index)}
            onBlur={() => setHoveredIndex((value) => (value === index ? null : value))}
          />
        </g>
      ))}

      {hoveredCandle && hoveredKline ? (
        <g pointerEvents="none" aria-label="K线悬浮镜">
          <line
            x1={hoveredCandle.x}
            y1={PADDING_TOP}
            x2={hoveredCandle.x}
            y2={CHART_HEIGHT - PADDING_BOTTOM}
            stroke="rgba(15,118,110,0.36)"
            strokeDasharray="5 5"
          />
          <circle
            cx={hoveredCandle.x}
            cy={hoveredCandle.closeY}
            r={5}
            fill="#ffffff"
            stroke={hoveredCandle.color}
            strokeWidth="2"
          />
          <rect
            x={clamp(hoveredCandle.x - 34, 10, CHART_WIDTH - 86)}
            y={PADDING_TOP + 6}
            width={76}
            height={18}
            rx={9}
            fill="rgba(15,118,110,0.14)"
          />
          <text
            x={clamp(hoveredCandle.x + 4, 18, CHART_WIDTH - 58)}
            y={PADDING_TOP + 18}
            textAnchor="middle"
            fontSize="10"
            fill="#0f766e"
            fontWeight="700"
          >
            {formatKlineTime(hoveredKline.open_time)}
          </text>
        </g>
      ) : null}
    </g>
  );
}
