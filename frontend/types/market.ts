export const MARKET_INTERVALS = ["1m", "5m", "15m", "1h", "4h"] as const;

export type MarketInterval = (typeof MARKET_INTERVALS)[number];

export interface PriceTicker {
  symbol: string;
  price: number;
  time: number;
}

export interface Kline {
  id: number;
  symbol: string;
  interval_type: string;
  open_price: number;
  high_price: number;
  low_price: number;
  close_price: number;
  volume: number;
  open_time: number;
  created_at: string;
}

export interface Indicator {
  id: number;
  symbol: string;
  rsi: number;
  macd: number;
  macd_signal: number;
  macd_histogram: number;
  ema20: number;
  ema50: number;
  atr: number;
  bollinger_upper: number;
  bollinger_middle: number;
  bollinger_lower: number;
  vwap: number;
  created_at: string;
}

export interface IndicatorSeriesPoint {
  open_time: number;
  rsi: number;
  macd: number;
  macd_signal: number;
  macd_histogram: number;
  ema20: number;
  ema50: number;
  atr: number;
  bollinger_upper: number;
  bollinger_middle: number;
  bollinger_lower: number;
  vwap: number;
}

export interface IndicatorSeriesResult {
  symbol: string;
  interval: string;
  points: IndicatorSeriesPoint[];
}

export interface OrderFlow {
  id: number;
  symbol: string;
  interval_type: string;
  open_time: number;
  buy_volume: number;
  sell_volume: number;
  delta: number;
  cvd: number;
  buy_large_trade_count: number;
  sell_large_trade_count: number;
  buy_large_trade_notional: number;
  sell_large_trade_notional: number;
  large_trade_delta: number;
  absorption_bias: string;
  absorption_strength: number;
  iceberg_bias: string;
  iceberg_strength: number;
  data_source: string;
  large_trades?: OrderFlowLargeTrade[];
  microstructure_events?: OrderFlowMicrostructureEvent[];
  created_at: string;
}

export interface OrderFlowLargeTrade {
  side: string;
  price: number;
  quantity: number;
  notional: number;
  trade_time: number;
}

export interface OrderFlowMicrostructureEvent {
  id?: number;
  symbol?: string;
  type: string;
  bias: string;
  score: number;
  strength: number;
  price: number;
  trade_time: number;
  detail: string;
  orderflow_id?: number;
  interval_type?: string;
  open_time?: number;
  created_at?: string;
}

export interface MicrostructureEventsResult {
  symbol: string;
  interval: string;
  events: OrderFlowMicrostructureEvent[];
}

export interface StructureEvent {
  label: string;
  kind: string;
  tier?: string;
  price: number;
  open_time: number;
}

export interface Structure {
  id: number;
  symbol: string;
  trend: string;
  support: number;
  resistance: number;
  primary_tier?: string;
  internal_support?: number;
  internal_resistance?: number;
  external_support?: number;
  external_resistance?: number;
  bos: boolean;
  choch: boolean;
  events: StructureEvent[];
  created_at: string;
}

export interface StructureSeriesPoint {
  open_time: number;
  trend: string;
  primary_tier?: string;
  support: number;
  resistance: number;
  internal_support?: number;
  internal_resistance?: number;
  external_support?: number;
  external_resistance?: number;
  bos: boolean;
  choch: boolean;
  event_labels: string[];
  event_tags?: string[];
}

export interface StructureSeriesResult {
  symbol: string;
  interval: string;
  points: StructureSeriesPoint[];
}

export interface LiquidityCluster {
  label: string;
  kind: string;
  price: number;
  strength: number;
}

export interface LiquidityWallLevel {
  label: string;
  kind: string;
  side: string;
  layer: string;
  price: number;
  quantity: number;
  notional: number;
  distance_bps: number;
  strength: number;
}

export interface LiquidityWallStrengthBand {
  side: string;
  band: string;
  lower_distance_bps: number;
  upper_distance_bps: number;
  level_count: number;
  total_notional: number;
  dominant_price: number;
  dominant_notional: number;
  strength: number;
}

export interface LiquidityWallEvolution {
  interval: string;
  buy_liquidity: number;
  sell_liquidity: number;
  buy_distance_bps: number;
  sell_distance_bps: number;
  buy_cluster_strength: number;
  sell_cluster_strength: number;
  buy_strength_delta: number;
  sell_strength_delta: number;
  order_book_imbalance: number;
  sweep_type: string;
  data_source: string;
  dominant_side: string;
}

export interface Liquidity {
  id: number;
  symbol: string;
  buy_liquidity: number;
  sell_liquidity: number;
  sweep_type: string;
  order_book_imbalance: number;
  data_source: string;
  equal_high: number;
  equal_low: number;
  stop_clusters: LiquidityCluster[];
  wall_levels: LiquidityWallLevel[];
  wall_strength_bands: LiquidityWallStrengthBand[];
  wall_evolution: LiquidityWallEvolution[];
  created_at: string;
}

export interface LiquiditySeriesPoint {
  open_time: number;
  buy_liquidity: number;
  sell_liquidity: number;
  sweep_type: string;
  order_book_imbalance: number;
  data_source: string;
  equal_high: number;
  equal_low: number;
  buy_cluster_strength: number;
  sell_cluster_strength: number;
}

export interface LiquiditySeriesResult {
  symbol: string;
  interval: string;
  points: LiquiditySeriesPoint[];
}

export interface LiquidityMapResult {
  symbol: string;
  interval: string;
  buy_liquidity: number;
  sell_liquidity: number;
  sweep_type: string;
  order_book_imbalance: number;
  data_source: string;
  equal_high: number;
  equal_low: number;
  stop_clusters: LiquidityCluster[];
  wall_levels: LiquidityWallLevel[];
  wall_strength_bands: LiquidityWallStrengthBand[];
  wall_evolution: LiquidityWallEvolution[];
}
