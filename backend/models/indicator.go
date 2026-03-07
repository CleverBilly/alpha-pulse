package models

import "time"

// IndicatorSeriesPoint 描述单根 K 线上对应的技术指标序列点。
type IndicatorSeriesPoint struct {
	OpenTime        int64   `json:"open_time"`
	RSI             float64 `json:"rsi"`
	MACD            float64 `json:"macd"`
	MACDSignal      float64 `json:"macd_signal"`
	MACDHistogram   float64 `json:"macd_histogram"`
	EMA20           float64 `json:"ema20"`
	EMA50           float64 `json:"ema50"`
	ATR             float64 `json:"atr"`
	BollingerUpper  float64 `json:"bollinger_upper"`
	BollingerMiddle float64 `json:"bollinger_middle"`
	BollingerLower  float64 `json:"bollinger_lower"`
	VWAP            float64 `json:"vwap"`
}

// Indicator 对应 indicators 表，保存技术指标计算结果。
type Indicator struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol          string    `gorm:"size:20;index;not null" json:"symbol"`
	RSI             float64   `gorm:"column:rsi;type:decimal(10,2);not null" json:"rsi"`
	MACD            float64   `gorm:"column:macd;type:decimal(10,4);not null" json:"macd"`
	MACDSignal      float64   `gorm:"column:macd_signal;type:decimal(10,4);not null" json:"macd_signal"`
	MACDHistogram   float64   `gorm:"column:macd_histogram;type:decimal(10,4);not null" json:"macd_histogram"`
	EMA20           float64   `gorm:"column:ema20;type:decimal(18,8);not null" json:"ema20"`
	EMA50           float64   `gorm:"column:ema50;type:decimal(18,8);not null" json:"ema50"`
	ATR             float64   `gorm:"column:atr;type:decimal(18,8);not null" json:"atr"`
	BollingerUpper  float64   `gorm:"column:bollinger_upper;type:decimal(18,8);not null" json:"bollinger_upper"`
	BollingerMiddle float64   `gorm:"column:bollinger_middle;type:decimal(18,8);not null" json:"bollinger_middle"`
	BollingerLower  float64   `gorm:"column:bollinger_lower;type:decimal(18,8);not null" json:"bollinger_lower"`
	VWAP            float64   `gorm:"column:vwap;type:decimal(18,8);not null" json:"vwap"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 指定数据表名。
func (Indicator) TableName() string {
	return "indicators"
}
