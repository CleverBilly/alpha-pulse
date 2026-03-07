package models

import "time"

// Kline 对应 kline 表，保存交易对 K 线数据。
type Kline struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol       string    `gorm:"size:20;not null;uniqueIndex:idx_kline_symbol_interval_open_time,priority:1" json:"symbol"`
	IntervalType string    `gorm:"column:interval_type;size:10;not null;uniqueIndex:idx_kline_symbol_interval_open_time,priority:2" json:"interval_type"`
	OpenPrice    float64   `gorm:"column:open_price;type:decimal(18,8);not null" json:"open_price"`
	HighPrice    float64   `gorm:"column:high_price;type:decimal(18,8);not null" json:"high_price"`
	LowPrice     float64   `gorm:"column:low_price;type:decimal(18,8);not null" json:"low_price"`
	ClosePrice   float64   `gorm:"column:close_price;type:decimal(18,8);not null" json:"close_price"`
	Volume       float64   `gorm:"type:decimal(18,8);not null" json:"volume"`
	OpenTime     int64     `gorm:"column:open_time;not null;uniqueIndex:idx_kline_symbol_interval_open_time,priority:3" json:"open_time"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 指定数据表名。
func (Kline) TableName() string {
	return "kline"
}
