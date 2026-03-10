package models

import "time"

// LargeTradeEvent 对应 large_trade_events 表，保存可回放的大单事件。
type LargeTradeEvent struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	OrderFlowID  uint64    `gorm:"column:orderflow_id;index;not null;default:0;comment:来源订单流快照 ID" json:"orderflow_id"`
	Symbol       string    `gorm:"size:20;index;not null;comment:交易对代码，如 BTCUSDT;uniqueIndex:idx_large_trade_event_unique,priority:1" json:"symbol"`
	AggTradeID   int64     `gorm:"column:agg_trade_id;index;not null;comment:来源聚合成交 ID;uniqueIndex:idx_large_trade_event_unique,priority:2" json:"agg_trade_id"`
	IntervalType string    `gorm:"column:interval_type;size:10;index;not null;default:'1m';comment:事件所属周期" json:"interval_type"`
	OpenTime     int64     `gorm:"column:open_time;index;not null;default:0;comment:对齐的 K 线起始时间，Unix 毫秒" json:"open_time"`
	Side         string    `gorm:"size:10;not null;comment:成交方向，buy 或 sell" json:"side"`
	Price        float64   `gorm:"type:decimal(18,8);not null;default:0;comment:成交价" json:"price"`
	Quantity     float64   `gorm:"type:decimal(24,8);not null;default:0;comment:成交数量，基础资产数量" json:"quantity"`
	Notional     float64   `gorm:"type:decimal(24,8);index;not null;default:0;comment:成交额，计价资产数量" json:"notional"`
	TradeTime    int64     `gorm:"column:trade_time;index;not null;default:0;comment:成交发生时间，Unix 毫秒" json:"trade_time"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;comment:入库时间" json:"created_at"`
}

// TableName 指定数据表名。
func (LargeTradeEvent) TableName() string {
	return "large_trade_events"
}

// TableComment 返回数据表注释。
func (LargeTradeEvent) TableComment() string {
	return "大单事件持久化镜像，支持历史回放、聚类分析和时间轴重建"
}
