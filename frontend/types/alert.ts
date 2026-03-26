export interface AlertDelivery {
  channel: string;
  status: string;
  detail?: string;
  sent_at?: number;
}

export type AlertOutcome = "pending" | "target_hit" | "stop_hit" | "expired";

export interface AlertEvent {
  id: string;
  symbol: string;
  kind: string;
  severity: string;
  title: string;
  verdict: string;
  tradeability_label: string;
  summary: string;
  reasons: string[];
  timeframe_labels: string[];
  confidence: number;
  risk_label: string;
  entry_price: number;
  stop_loss: number;
  target_price: number;
  risk_reward: number;
  created_at: number;
  deliveries: AlertDelivery[];
  // 复盘追踪字段
  outcome?: AlertOutcome;
  outcome_price?: number;
  outcome_at?: number;
  actual_rr?: number;
  interval?: string;
}

export interface AlertStats {
  symbol: string;
  total: number;
  target_hit: number;
  stop_hit: number;
  pending: number;
  expired: number;
  win_rate: number;
  avg_rr: number;
  sample_size_label: string;
}

export interface AlertFeed {
  items: AlertEvent[];
  generated: number;
}

export interface AlertPreferences {
  feishu_enabled: boolean;
  browser_enabled: boolean;
  sound_enabled: boolean;
  setup_ready_enabled: boolean;
  direction_shift_enabled: boolean;
  no_trade_enabled: boolean;
  minimum_confidence: number;
  quiet_hours_enabled: boolean;
  quiet_hours_start: number;
  quiet_hours_end: number;
  symbols: string[];
  available_symbols: string[];
}
