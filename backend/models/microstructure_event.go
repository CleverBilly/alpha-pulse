package models

import "time"

// MicrostructureEvent 对应 microstructure_events 表，保存持久化的微结构事件序列。
type MicrostructureEvent struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderFlowID  uint64    `gorm:"column:orderflow_id;index;not null;default:0" json:"orderflow_id"`
	Symbol       string    `gorm:"size:20;index;not null;uniqueIndex:idx_micro_event_unique,priority:1" json:"symbol"`
	IntervalType string    `gorm:"column:interval_type;size:10;index;not null;default:'1m';uniqueIndex:idx_micro_event_unique,priority:2" json:"interval_type"`
	OpenTime     int64     `gorm:"column:open_time;index;not null;default:0;uniqueIndex:idx_micro_event_unique,priority:3" json:"open_time"`
	EventType    string    `gorm:"column:event_type;size:40;not null;uniqueIndex:idx_micro_event_unique,priority:4" json:"type"`
	Bias         string    `gorm:"size:20;not null;default:'neutral';uniqueIndex:idx_micro_event_unique,priority:6" json:"bias"`
	Score        int       `gorm:"not null;default:0" json:"score"`
	Strength     float64   `gorm:"type:decimal(12,6);not null;default:0" json:"strength"`
	Price        float64   `gorm:"type:decimal(18,8);not null;default:0" json:"price"`
	TradeTime    int64     `gorm:"column:trade_time;index;not null;default:0;uniqueIndex:idx_micro_event_unique,priority:5" json:"trade_time"`
	Detail       string    `gorm:"type:text;not null" json:"detail"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 指定数据表名。
func (MicrostructureEvent) TableName() string {
	return "microstructure_events"
}
