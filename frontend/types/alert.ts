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
