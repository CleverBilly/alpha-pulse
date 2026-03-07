import {
  Indicator,
  IndicatorSeriesPoint,
  Kline,
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
