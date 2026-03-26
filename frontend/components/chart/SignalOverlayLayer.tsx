import type { Dispatch, SetStateAction } from "react";
import {
  ActiveSignal,
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
    </g>
  );
}
