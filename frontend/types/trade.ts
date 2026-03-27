export type TradeOrderStatus = "pending_fill" | "open" | "closed" | "expired" | "failed";
export type TradeSource = "system" | "manual";
export type TradeSide = "LONG" | "SHORT";

export interface TradeSettings {
  trade_enabled_env: boolean;
  trade_auto_execute_env: boolean;
  allowed_symbols_env: string[];
  auto_execute_enabled: boolean;
  allowed_symbols: string[];
  risk_pct: number;
  min_risk_reward: number;
  entry_timeout_seconds: number;
  max_open_positions: number;
  sync_enabled: boolean;
  updated_by: string;
}

export interface TradeOrder {
  id: number;
  alert_id: string;
  symbol: string;
  side: TradeSide;
  requested_qty: number;
  filled_qty: number;
  entry_order_id?: string;
  entry_order_type?: string;
  limit_price: number;
  filled_price?: number;
  stop_loss?: number;
  target_price?: number;
  entry_status: string;
  status: TradeOrderStatus;
  source: TradeSource;
  close_reason?: string;
  created_at: number;
  closed_at: number;
}

export interface TradeRuntimeStatus {
  trade_enabled_env: boolean;
  trade_auto_execute_env: boolean;
  pending_fill_count: number;
  open_count: number;
}
