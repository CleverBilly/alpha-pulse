import {
  Indicator,
  IndicatorSeriesPoint,
  Kline,
  FuturesSnapshot,
  Liquidity,
  LiquiditySeriesPoint,
  OrderFlowMicrostructureEvent,
  OrderFlow,
  PriceTicker,
  Structure,
  StructureSeriesPoint,
} from "@/types/market";
import { Signal } from "@/types/signal";
import { SignalTimelinePoint } from "@/types/signal";

export interface MarketSnapshot {
  price: PriceTicker;
  futures: FuturesSnapshot;
  klines: Kline[];
  indicator: Indicator;
  indicator_series: IndicatorSeriesPoint[];
  orderflow: OrderFlow;
  microstructure_events: OrderFlowMicrostructureEvent[];
  structure: Structure;
  structure_series: StructureSeriesPoint[];
  liquidity: Liquidity;
  liquidity_series: LiquiditySeriesPoint[];
  signal: Signal;
  signal_timeline: SignalTimelinePoint[];
}

export interface MarketSnapshotStreamMessage {
  type: "snapshot" | "error";
  symbol: string;
  interval: string;
  limit: number;
  sent_at: number;
  data?: MarketSnapshot;
  error?: string;
}
