package models

import "time"

// Kline 对应 kline 表，保存交易对 K 线数据。
type Kline struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	Symbol       string    `gorm:"size:20;not null;comment:交易对代码，如 BTCUSDT;uniqueIndex:idx_kline_symbol_interval_open_time,priority:1" json:"symbol"`
	IntervalType string    `gorm:"column:interval_type;size:10;not null;comment:K 线周期，如 1m、5m、1h;uniqueIndex:idx_kline_symbol_interval_open_time,priority:2" json:"interval_type"`
	OpenPrice    float64   `gorm:"column:open_price;type:decimal(18,8);not null;comment:开盘价" json:"open_price"`
	HighPrice    float64   `gorm:"column:high_price;type:decimal(18,8);not null;comment:最高价" json:"high_price"`
	LowPrice     float64   `gorm:"column:low_price;type:decimal(18,8);not null;comment:最低价" json:"low_price"`
	ClosePrice   float64   `gorm:"column:close_price;type:decimal(18,8);not null;comment:收盘价" json:"close_price"`
	Volume       float64   `gorm:"type:decimal(18,8);not null;comment:成交量，基础资产数量" json:"volume"`
	OpenTime     int64     `gorm:"column:open_time;not null;comment:K 线起始时间，Unix 毫秒;uniqueIndex:idx_kline_symbol_interval_open_time,priority:3" json:"open_time"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;comment:入库时间" json:"created_at"`
}

// TableName 指定数据表名。
func (Kline) TableName() string {
	return "kline"
}

// TableComment 返回数据表注释。
func (Kline) TableComment() string {
	return "原始 K 线行情数据，供指标、结构、流动性和信号分析复用"
}
