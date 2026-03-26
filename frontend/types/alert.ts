export interface AlertDelivery {
  channel: string;
  status: string;
  detail?: string;
  sent_at?: number;
}

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
