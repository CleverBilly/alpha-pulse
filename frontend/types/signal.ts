import { Indicator, Liquidity, OrderFlow, Structure } from "@/types/market";

export interface SignalFactor {
  key: string;
  name: string;
  score: number;
  bias: string;
  reason: string;
  section: string;
}

export interface Signal {
  id: number;
  symbol: string;
  interval_type: string;
  open_time: number;
  signal: string;
  score: number;
  confidence: number;
  entry_price: number;
  stop_loss: number;
  target_price: number;
  risk_reward: number;
  trend_bias: string;
  explain: string;
  factors: SignalFactor[];
  created_at: string;
}

export interface SignalTimelinePoint {
  id: number;
  symbol: string;
  interval_type: string;
  open_time: number;
  signal: string;
  score: number;
  confidence: number;
  entry_price: number;
  stop_loss: number;
  target_price: number;
}

export interface SignalTimelineResult {
  symbol: string;
  interval: string;
  points: SignalTimelinePoint[];
}

export interface SignalBundle {
  signal: Signal;
  indicator: Indicator;
  orderflow: OrderFlow;
  structure: Structure;
  liquidity: Liquidity;
}
