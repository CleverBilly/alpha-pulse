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
	ID              uint64    `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	Symbol          string    `gorm:"size:20;index;not null;comment:交易对代码，如 BTCUSDT" json:"symbol"`
	RSI             float64   `gorm:"column:rsi;type:decimal(10,2);not null;comment:相对强弱指标 RSI" json:"rsi"`
	MACD            float64   `gorm:"column:macd;type:decimal(10,4);not null;comment:MACD 快线值" json:"macd"`
	MACDSignal      float64   `gorm:"column:macd_signal;type:decimal(10,4);not null;comment:MACD Signal 线值" json:"macd_signal"`
	MACDHistogram   float64   `gorm:"column:macd_histogram;type:decimal(10,4);not null;comment:MACD 柱体值" json:"macd_histogram"`
	EMA20           float64   `gorm:"column:ema20;type:decimal(18,8);not null;comment:20 周期 EMA" json:"ema20"`
	EMA50           float64   `gorm:"column:ema50;type:decimal(18,8);not null;comment:50 周期 EMA" json:"ema50"`
	ATR             float64   `gorm:"column:atr;type:decimal(18,8);not null;comment:平均真实波幅 ATR" json:"atr"`
	BollingerUpper  float64   `gorm:"column:bollinger_upper;type:decimal(18,8);not null;comment:布林带上轨" json:"bollinger_upper"`
	BollingerMiddle float64   `gorm:"column:bollinger_middle;type:decimal(18,8);not null;comment:布林带中轨" json:"bollinger_middle"`
	BollingerLower  float64   `gorm:"column:bollinger_lower;type:decimal(18,8);not null;comment:布林带下轨" json:"bollinger_lower"`
	VWAP            float64   `gorm:"column:vwap;type:decimal(18,8);not null;comment:成交量加权平均价 VWAP" json:"vwap"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime;comment:指标计算入库时间" json:"created_at"`
}

// TableName 指定数据表名。
func (Indicator) TableName() string {
	return "indicators"
}

// TableComment 返回数据表注释。
func (Indicator) TableComment() string {
	return "技术指标计算快照，记录 RSI、MACD、EMA、ATR、布林带和 VWAP"
}
