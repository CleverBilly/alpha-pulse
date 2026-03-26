import { ChartSeries, StructureMarker } from "./chartTypes";

interface StructureLiquidityLayerProps {
  structureMarkers: StructureMarker[];
  series: ChartSeries;
  structureOpacity: number;
  liquidityOpacity: number;
}

export default function StructureLiquidityLayer({
  structureMarkers,
  series,
  structureOpacity,
  liquidityOpacity,
}: StructureLiquidityLayerProps) {
  return (
    <g>
      <g opacity={structureOpacity}>
        {series.supportTrack.map((points, index) => (
          <polyline
            key={`support-track-${index}`}
            points={points}
            fill="none"
            stroke="#047857"
            strokeWidth="1.8"
            opacity="0.8"
          />
        ))}
        {series.internalSupportTrack.map((points, index) => (
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
        {series.resistanceTrack.map((points, index) => (
          <polyline
            key={`resistance-track-${index}`}
            points={points}
            fill="none"
            stroke="#be123c"
            strokeWidth="1.8"
            opacity="0.8"
          />
        ))}
        {series.internalResistanceTrack.map((points, index) => (
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
      </g>

      <g opacity={liquidityOpacity}>
        {series.buyLiquidityTrack.map((points, index) => (
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
        {series.sellLiquidityTrack.map((points, index) => (
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
        {series.equalHighTrack.map((points, index) => (
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
        {series.equalLowTrack.map((points, index) => (
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
      </g>

      <g opacity={structureOpacity}>
        {structureMarkers.map((marker) => (
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
      </g>
    </g>
  );
}
