package models

// TradeOrder 对应 trade_orders 表，保存自动交易与手动持仓的审计记录。
type TradeOrder struct {
	ID             uint64  `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	AlertID        string  `gorm:"column:alert_id;size:96;index;not null;default:'';comment:关联 alert id" json:"alert_id"`
	Symbol         string  `gorm:"column:symbol;size:20;index;not null;comment:交易对代码" json:"symbol"`
	Side           string  `gorm:"column:side;size:8;not null;default:'';comment:LONG/SHORT" json:"side"`
	RequestedQty   float64 `gorm:"column:requested_qty;type:decimal(18,8);not null;default:0;comment:请求开仓数量" json:"requested_qty"`
	FilledQty      float64 `gorm:"column:filled_qty;type:decimal(18,8);not null;default:0;comment:实际成交数量" json:"filled_qty"`
	EntryOrderID   string  `gorm:"column:entry_order_id;size:64;not null;default:'';comment:Binance 开仓订单 ID" json:"entry_order_id"`
	EntryOrderType string  `gorm:"column:entry_order_type;size:24;not null;default:'LIMIT';comment:开仓订单类型" json:"entry_order_type"`
	LimitPrice     float64 `gorm:"column:limit_price;type:decimal(18,8);not null;default:0;comment:限价开仓价格" json:"limit_price"`
	FilledPrice    float64 `gorm:"column:filled_price;type:decimal(18,8);not null;default:0;comment:实际成交均价" json:"filled_price"`
	StopLoss       float64 `gorm:"column:stop_loss;type:decimal(18,8);not null;default:0;comment:止损价" json:"stop_loss"`
	TargetPrice    float64 `gorm:"column:target_price;type:decimal(18,8);not null;default:0;comment:止盈价" json:"target_price"`
	EntryStatus    string  `gorm:"column:entry_status;size:24;index;not null;default:'pending_fill';comment:挂单阶段状态" json:"entry_status"`
	Status         string  `gorm:"column:status;size:24;index;not null;default:'pending_fill';comment:订单生命周期状态" json:"status"`
	EntryExpiresAt int64   `gorm:"column:entry_expires_at;not null;default:0;comment:挂单过期时间 Unix ms" json:"entry_expires_at"`
	Source         string  `gorm:"column:source;size:24;index;not null;default:'system';comment:system/manual" json:"source"`
	CloseOrderID   string  `gorm:"column:close_order_id;size:64;not null;default:'';comment:人工或兜底平仓订单 ID" json:"close_order_id"`
	CloseReason    string  `gorm:"column:close_reason;size:255;not null;default:'';comment:平仓或失败原因" json:"close_reason"`
	CreatedAtUnixMs int64  `gorm:"column:created_at;index;not null;default:0;comment:创建时间 Unix ms" json:"created_at"`
	ClosedAt       int64   `gorm:"column:closed_at;not null;default:0;comment:平仓时间 Unix ms" json:"closed_at"`
}

// TableName 指定数据表名。
func (TradeOrder) TableName() string {
	return "trade_orders"
}

// TableComment 返回数据表注释。
func (TradeOrder) TableComment() string {
	return "自动交易订单审计表，保存限价挂单、持仓和收口状态"
}
