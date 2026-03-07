package models

import "time"

// SignalFactor 描述单个因子对最终信号的贡献。
type SignalFactor struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Score   int    `json:"score"`
	Bias    string `json:"bias"`
	Reason  string `json:"reason"`
	Section string `json:"section"`
}

// Signal 对应 signals 表，保存综合交易信号。
type Signal struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol       string    `gorm:"size:20;index;not null" json:"symbol"`
	IntervalType string    `gorm:"column:interval_type;size:10;index;not null;default:'1m'" json:"interval_type"`
	OpenTime     int64     `gorm:"column:open_time;index;not null;default:0" json:"open_time"`
	Action       string    `gorm:"column:signal;size:20;not null" json:"signal"`
	Score        int       `gorm:"not null" json:"score"`
	Confidence   int       `gorm:"not null" json:"confidence"`
	EntryPrice   float64   `gorm:"column:entry_price;type:decimal(18,8);not null" json:"entry_price"`
	StopLoss     float64   `gorm:"column:stop_loss;type:decimal(18,8);not null" json:"stop_loss"`
	TargetPrice  float64   `gorm:"column:target_price;type:decimal(18,8);not null" json:"target_price"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	// Explain 不落库，用于返回 AI 解释文本。
	Explain string `gorm:"-" json:"explain,omitempty"`

	// Factors 不落库，用于返回多因子评分明细。
	Factors []SignalFactor `gorm:"-" json:"factors,omitempty"`

	// RiskReward 不落库，用于返回当前信号的盈亏比。
	RiskReward float64 `gorm:"-" json:"risk_reward,omitempty"`

	// TrendBias 不落库，用于返回当前信号的总体方向偏向。
	TrendBias string `gorm:"-" json:"trend_bias,omitempty"`
}

// SignalTimelinePoint 描述图表所需的历史信号点。
type SignalTimelinePoint struct {
	ID           uint64  `json:"id"`
	Symbol       string  `json:"symbol"`
	IntervalType string  `json:"interval_type"`
	OpenTime     int64   `json:"open_time"`
	Signal       string  `json:"signal"`
	Score        int     `json:"score"`
	Confidence   int     `json:"confidence"`
	EntryPrice   float64 `json:"entry_price"`
	StopLoss     float64 `json:"stop_loss"`
	TargetPrice  float64 `json:"target_price"`
}

// TableName 指定数据表名。
func (Signal) TableName() string {
	return "signals"
}
