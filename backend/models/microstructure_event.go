package models

import "time"

// MicrostructureEvent 对应 microstructure_events 表，保存持久化的微结构事件序列。
type MicrostructureEvent struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	OrderFlowID  uint64    `gorm:"column:orderflow_id;index;not null;default:0;comment:来源订单流快照 ID" json:"orderflow_id"`
	Symbol       string    `gorm:"size:20;index;not null;comment:交易对代码，如 BTCUSDT;uniqueIndex:idx_micro_event_unique,priority:1" json:"symbol"`
	IntervalType string    `gorm:"column:interval_type;size:10;index;not null;default:'1m';comment:事件所属周期;uniqueIndex:idx_micro_event_unique,priority:2" json:"interval_type"`
	OpenTime     int64     `gorm:"column:open_time;index;not null;default:0;comment:对齐的 K 线起始时间，Unix 毫秒;uniqueIndex:idx_micro_event_unique,priority:3" json:"open_time"`
	EventType    string    `gorm:"column:event_type;size:40;not null;comment:微结构事件类型;uniqueIndex:idx_micro_event_unique,priority:4" json:"type"`
	Bias         string    `gorm:"size:20;not null;default:'neutral';comment:事件方向偏向;uniqueIndex:idx_micro_event_unique,priority:6" json:"bias"`
	Score        int       `gorm:"not null;default:0;comment:事件评分" json:"score"`
	Strength     float64   `gorm:"type:decimal(12,6);not null;default:0;comment:事件强度" json:"strength"`
	Price        float64   `gorm:"type:decimal(18,8);not null;default:0;comment:事件参考价" json:"price"`
	TradeTime    int64     `gorm:"column:trade_time;index;not null;default:0;comment:事件发生时间，Unix 毫秒;uniqueIndex:idx_micro_event_unique,priority:5" json:"trade_time"`
	Detail       string    `gorm:"type:text;not null;comment:事件详细说明" json:"detail"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;comment:入库时间" json:"created_at"`
}

// TableName 指定数据表名。
func (MicrostructureEvent) TableName() string {
	return "microstructure_events"
}

// TableComment 返回数据表注释。
func (MicrostructureEvent) TableComment() string {
	return "微结构事件历史序列，用于时间轴展示、图表标注和后续回放"
}
