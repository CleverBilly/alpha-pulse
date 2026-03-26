import type { Dispatch, SetStateAction } from "react";
import {
  ActiveSignal,
  ChartCoords,
  MicrostructureMarker,
  MicrostructureTooltip,
  SignalMarker,
} from "./chartTypes";
import { formatMicrostructureEventType } from "./chartHelpers";

export type { ActiveSignal };

interface SignalOverlayLayerProps {
  signalMarkers: SignalMarker[];
  microstructureMarkers: MicrostructureMarker[];
  microstructureTooltip: MicrostructureTooltip | null;
  hoveredMicrostructureMarkerKey: string | null;
  pinnedMicrostructureMarkerKey: string | null;
  setHoveredMicrostructureMarkerKey: Dispatch<SetStateAction<string | null>>;
  setPinnedMicrostructureMarkerKey: Dispatch<SetStateAction<string | null>>;
  signalOpacity: number;
  microOpacity: number;
  activeSignal?: ActiveSignal | null;
  coords?: ChartCoords | null;
}

export default function SignalOverlayLayer({
  signalMarkers,
  microstructureMarkers,
  microstructureTooltip,
  hoveredMicrostructureMarkerKey,
  pinnedMicrostructureMarkerKey,
  setHoveredMicrostructureMarkerKey,
  setPinnedMicrostructureMarkerKey,
  signalOpacity,
  microOpacity,
  activeSignal,
  coords,
}: SignalOverlayLayerProps) {
  return (
    <g>
      <g opacity={microOpacity}>
        {microstructureMarkers.map((marker) => {
          const isPinned = pinnedMicrostructureMarkerKey === marker.key;
          const isHovered = hoveredMicrostructureMarkerKey === marker.key;
          return (
            <g
              key={marker.key}
              role="button"
              tabIndex={0}
              aria-label={`微结构 ${marker.label} ${formatMicrostructureEventType(marker.type)}`}
              className="cursor-pointer"
              onMouseEnter={() => setHoveredMicrostructureMarkerKey(marker.key)}
              onMouseLeave={() =>
                setHoveredMicrostructureMarkerKey((value) => (value === marker.key ? null : value))
              }
              onFocus={() => setHoveredMicrostructureMarkerKey(marker.key)}
              onBlur={() =>
                setHoveredMicrostructureMarkerKey((value) => (value === marker.key ? null : value))
              }
              onClick={() =>
                setPinnedMicrostructureMarkerKey((value) => (value === marker.key ? null : marker.key))
              }
            >
              <rect
                x={marker.x - (isPinned || isHovered ? 6.2 : 5.2)}
                y={marker.y - (isPinned || isHovered ? 6.2 : 5.2)}
                width={isPinned || isHovered ? 12.4 : 10.4}
                height={isPinned || isHovered ? 12.4 : 10.4}
                rx={2.4}
                fill={marker.color}
                stroke="#ffffff"
                strokeWidth={isPinned ? "1.8" : "1.3"}
                transform={`rotate(45 ${marker.x} ${marker.y})`}
              />
              <rect
                x={marker.x - 18}
                y={marker.labelY - 9}
                width={36}
                height={16}
                rx={8}
                fill={marker.labelBackground}
                opacity={isPinned || isHovered ? 1 : 0.88}
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
          );
        })}

        {microstructureTooltip ? (
          <g pointerEvents="none" aria-label="微结构提示">
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
      </g>

      <g opacity={signalOpacity}>
        {signalMarkers.map((marker) => (
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
      </g>

      {activeSignal && coords && (
        <g>
          {activeSignal.entryPrice >= coords.priceMin && activeSignal.entryPrice <= coords.priceMax && (
            <>
              <line
                x1={coords.paddingLeft}
                x2={coords.chartWidth - coords.paddingRight}
                y1={coords.toY(activeSignal.entryPrice)}
                y2={coords.toY(activeSignal.entryPrice)}
                stroke="#52c41a"
                strokeWidth={1}
                strokeDasharray="4 3"
                opacity={0.85}
              />
              <text
                x={coords.chartWidth - coords.paddingRight + 4}
                y={coords.toY(activeSignal.entryPrice) + 4}
                fill="#52c41a"
                fontSize={10}
              >
                {`entry ${activeSignal.entryPrice.toFixed(2)}`}
              </text>
            </>
          )}
          {activeSignal.stopLoss >= coords.priceMin && activeSignal.stopLoss <= coords.priceMax && (
            <>
              <line
                x1={coords.paddingLeft}
                x2={coords.chartWidth - coords.paddingRight}
                y1={coords.toY(activeSignal.stopLoss)}
                y2={coords.toY(activeSignal.stopLoss)}
                stroke="#ff4d4f"
                strokeWidth={1}
                strokeDasharray="4 3"
                opacity={0.85}
              />
              <text
                x={coords.chartWidth - coords.paddingRight + 4}
                y={coords.toY(activeSignal.stopLoss) + 4}
                fill="#ff4d4f"
                fontSize={10}
              >
                {`SL ${activeSignal.stopLoss.toFixed(2)}`}
              </text>
            </>
          )}
          {activeSignal.targetPrice >= coords.priceMin && activeSignal.targetPrice <= coords.priceMax && (
            <>
              <line
                x1={coords.paddingLeft}
                x2={coords.chartWidth - coords.paddingRight}
                y1={coords.toY(activeSignal.targetPrice)}
                y2={coords.toY(activeSignal.targetPrice)}
                stroke="#faad14"
                strokeWidth={1}
                strokeDasharray="4 3"
                opacity={0.85}
              />
              <text
                x={coords.chartWidth - coords.paddingRight + 4}
                y={coords.toY(activeSignal.targetPrice) + 4}
                fill="#faad14"
                fontSize={10}
              >
                {`TP ${activeSignal.targetPrice.toFixed(2)}`}
              </text>
            </>
          )}
        </g>
      )}
    </g>
  );
}
