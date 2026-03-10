package models

import "time"

// LargeTradeEvent 对应 large_trade_events 表，保存可回放的大单事件。
type LargeTradeEvent struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderFlowID  uint64    `gorm:"column:orderflow_id;index;not null;default:0" json:"orderflow_id"`
	Symbol       string    `gorm:"size:20;index;not null;uniqueIndex:idx_large_trade_event_unique,priority:1" json:"symbol"`
	AggTradeID   int64     `gorm:"column:agg_trade_id;index;not null;uniqueIndex:idx_large_trade_event_unique,priority:2" json:"agg_trade_id"`
	IntervalType string    `gorm:"column:interval_type;size:10;index;not null;default:'1m'" json:"interval_type"`
	OpenTime     int64     `gorm:"column:open_time;index;not null;default:0" json:"open_time"`
	Side         string    `gorm:"size:10;not null" json:"side"`
	Price        float64   `gorm:"type:decimal(18,8);not null;default:0" json:"price"`
	Quantity     float64   `gorm:"type:decimal(24,8);not null;default:0" json:"quantity"`
	Notional     float64   `gorm:"type:decimal(24,8);index;not null;default:0" json:"notional"`
	TradeTime    int64     `gorm:"column:trade_time;index;not null;default:0" json:"trade_time"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 指定数据表名。
func (LargeTradeEvent) TableName() string {
	return "large_trade_events"
}
