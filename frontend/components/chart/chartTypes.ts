export const CHART_WIDTH = 960;
export const CHART_HEIGHT = 360;
export const PADDING_TOP = 24;
export const PADDING_RIGHT = 118;
export const PADDING_BOTTOM = 36;
export const PADDING_LEFT = 18;

export interface ChartCoords {
  toX: (index: number) => number;
  toY: (price: number) => number;
  candleWidth: number;
  chartWidth: number;
  chartHeight: number;
  paddingTop: number;
  paddingRight: number;
  paddingBottom: number;
  paddingLeft: number;
  priceMin: number;
  priceMax: number;
}

export interface CandleShape {
  openTime: number;
  x: number;
  highY: number;
  lowY: number;
  closeY: number;
  bodyY: number;
  bodyHeight: number;
  bodyWidth: number;
  hitBoxWidth: number;
  color: string;
}

export interface GridLine {
  value: number;
  y: number;
}

export interface TimeLabel {
  x: number;
  label: string;
}

export interface StructureMarker {
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

export interface ZoneDraft {
  label: string;
  value?: number;
  color: string;
  dasharray: string;
  labelBackground: string;
  labelColor: string;
}

export interface ZoneLine extends ZoneDraft {
  value: number;
  y: number;
}

export interface SignalMarker {
  label: string;
  x: number;
  y: number;
  labelY: number;
  color: string;
  labelColor: string;
  labelBackground: string;
}

export interface MicrostructureMarker {
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

export interface MicrostructureTooltip {
  x: number;
  y: number;
  width: number;
  height: number;
  lines: string[];
}

export interface ChartSeries {
  supportTrack: string[];
  resistanceTrack: string[];
  internalSupportTrack: string[];
  internalResistanceTrack: string[];
  buyLiquidityTrack: string[];
  sellLiquidityTrack: string[];
  equalHighTrack: string[];
  equalLowTrack: string[];
  ema20: string[];
  ema50: string[];
  vwap: string[];
  bollingerUpper: string[];
  bollingerMiddle: string[];
  bollingerLower: string[];
}

export interface ChartModel {
  candles: CandleShape[];
  gridLines: GridLine[];
  timeLabels: TimeLabel[];
  structureMarkers: StructureMarker[];
  microstructureMarkers: MicrostructureMarker[];
  signalMarkers: SignalMarker[];
  zoneLines: ZoneLine[];
  series: ChartSeries;
}

export interface ActiveSignal {
  entryPrice: number;
  stopLoss: number;
  targetPrice: number;
  direction: "long" | "short";
}

export type LegendFocusKey = "indicators" | "structure" | "liquidity" | "signal" | "micro";
