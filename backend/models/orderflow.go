package models

import "time"

// OrderFlowLargeTrade 描述单笔大额成交事件。
type OrderFlowLargeTrade struct {
	AggTradeID int64   `json:"agg_trade_id,omitempty"`
	Side       string  `json:"side"`
	Price      float64 `json:"price"`
	Quantity   float64 `json:"quantity"`
	Notional   float64 `json:"notional"`
	TradeTime  int64   `json:"trade_time"`
}

// OrderFlowMicrostructureEvent 描述最近成交流中识别到的微结构事件。
type OrderFlowMicrostructureEvent struct {
	Type      string  `json:"type"`
	Bias      string  `json:"bias"`
	Score     int     `json:"score"`
	Strength  float64 `json:"strength"`
	Price     float64 `json:"price"`
	TradeTime int64   `json:"trade_time"`
	Detail    string  `json:"detail"`
}

// OrderFlow 对应 orderflow 表，保存订单流分析结果。
type OrderFlow struct {
	ID                     uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol                 string    `gorm:"size:20;index;not null" json:"symbol"`
	IntervalType           string    `gorm:"column:interval_type;size:10;index;not null;default:'1m'" json:"interval_type"`
	OpenTime               int64     `gorm:"column:open_time;index;not null;default:0" json:"open_time"`
	BuyVolume              float64   `gorm:"column:buy_volume;type:decimal(18,8);not null" json:"buy_volume"`
	SellVolume             float64   `gorm:"column:sell_volume;type:decimal(18,8);not null" json:"sell_volume"`
	Delta                  float64   `gorm:"column:delta;type:decimal(18,8);not null" json:"delta"`
	CVD                    float64   `gorm:"column:cvd;type:decimal(18,8);not null" json:"cvd"`
	BuyLargeTradeCount     int       `gorm:"column:buy_large_trade_count;not null;default:0" json:"buy_large_trade_count"`
	SellLargeTradeCount    int       `gorm:"column:sell_large_trade_count;not null;default:0" json:"sell_large_trade_count"`
	BuyLargeTradeNotional  float64   `gorm:"column:buy_large_trade_notional;type:decimal(24,8);not null;default:0" json:"buy_large_trade_notional"`
	SellLargeTradeNotional float64   `gorm:"column:sell_large_trade_notional;type:decimal(24,8);not null;default:0" json:"sell_large_trade_notional"`
	LargeTradeDelta        float64   `gorm:"column:large_trade_delta;type:decimal(24,8);not null;default:0" json:"large_trade_delta"`
	AbsorptionBias         string    `gorm:"column:absorption_bias;size:20;not null;default:'none'" json:"absorption_bias"`
	AbsorptionStrength     float64   `gorm:"column:absorption_strength;type:decimal(12,6);not null;default:0" json:"absorption_strength"`
	IcebergBias            string    `gorm:"column:iceberg_bias;size:20;not null;default:'none'" json:"iceberg_bias"`
	IcebergStrength        float64   `gorm:"column:iceberg_strength;type:decimal(12,6);not null;default:0" json:"iceberg_strength"`
	DataSource             string    `gorm:"column:data_source;size:20;not null;default:'kline'" json:"data_source"`
	CreatedAt              time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	// LargeTrades 不落库，用于返回最近识别到的大额成交事件。
	LargeTrades []OrderFlowLargeTrade `gorm:"-" json:"large_trades,omitempty"`

	// MicrostructureEvents 不落库，用于返回最近微结构事件序列。
	MicrostructureEvents []OrderFlowMicrostructureEvent `gorm:"-" json:"microstructure_events,omitempty"`
}

// TableName 指定数据表名。
func (OrderFlow) TableName() string {
	return "orderflow"
}
