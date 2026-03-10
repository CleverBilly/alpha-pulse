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
	ID                     uint64    `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	Symbol                 string    `gorm:"size:20;index;not null;comment:交易对代码，如 BTCUSDT" json:"symbol"`
	IntervalType           string    `gorm:"column:interval_type;size:10;index;not null;default:'1m';comment:订单流分析周期" json:"interval_type"`
	OpenTime               int64     `gorm:"column:open_time;index;not null;default:0;comment:对齐的 K 线起始时间，Unix 毫秒" json:"open_time"`
	BuyVolume              float64   `gorm:"column:buy_volume;type:decimal(18,8);not null;comment:主动买入成交量" json:"buy_volume"`
	SellVolume             float64   `gorm:"column:sell_volume;type:decimal(18,8);not null;comment:主动卖出成交量" json:"sell_volume"`
	Delta                  float64   `gorm:"column:delta;type:decimal(18,8);not null;comment:买卖量差值" json:"delta"`
	CVD                    float64   `gorm:"column:cvd;type:decimal(18,8);not null;comment:累积成交量差 CVD" json:"cvd"`
	BuyLargeTradeCount     int       `gorm:"column:buy_large_trade_count;not null;default:0;comment:买方大单数量" json:"buy_large_trade_count"`
	SellLargeTradeCount    int       `gorm:"column:sell_large_trade_count;not null;default:0;comment:卖方大单数量" json:"sell_large_trade_count"`
	BuyLargeTradeNotional  float64   `gorm:"column:buy_large_trade_notional;type:decimal(24,8);not null;default:0;comment:买方大单成交额" json:"buy_large_trade_notional"`
	SellLargeTradeNotional float64   `gorm:"column:sell_large_trade_notional;type:decimal(24,8);not null;default:0;comment:卖方大单成交额" json:"sell_large_trade_notional"`
	LargeTradeDelta        float64   `gorm:"column:large_trade_delta;type:decimal(24,8);not null;default:0;comment:大单成交额差值" json:"large_trade_delta"`
	AbsorptionBias         string    `gorm:"column:absorption_bias;size:20;not null;default:'none';comment:吸收行为方向" json:"absorption_bias"`
	AbsorptionStrength     float64   `gorm:"column:absorption_strength;type:decimal(12,6);not null;default:0;comment:吸收行为强度" json:"absorption_strength"`
	IcebergBias            string    `gorm:"column:iceberg_bias;size:20;not null;default:'none';comment:冰山单方向" json:"iceberg_bias"`
	IcebergStrength        float64   `gorm:"column:iceberg_strength;type:decimal(12,6);not null;default:0;comment:冰山单强度" json:"iceberg_strength"`
	DataSource             string    `gorm:"column:data_source;size:20;not null;default:'kline';comment:分析数据来源，如 agg_trade 或 kline" json:"data_source"`
	CreatedAt              time.Time `gorm:"column:created_at;autoCreateTime;comment:订单流分析入库时间" json:"created_at"`

	// LargeTrades 不落库，用于返回最近识别到的大额成交事件。
	LargeTrades []OrderFlowLargeTrade `gorm:"-" json:"large_trades,omitempty"`

	// MicrostructureEvents 不落库，用于返回最近微结构事件序列。
	MicrostructureEvents []OrderFlowMicrostructureEvent `gorm:"-" json:"microstructure_events,omitempty"`
}

// TableName 指定数据表名。
func (OrderFlow) TableName() string {
	return "orderflow"
}

// TableComment 返回数据表注释。
func (OrderFlow) TableComment() string {
	return "订单流分析快照，记录主动买卖量、大单统计、吸收和冰山单特征"
}
