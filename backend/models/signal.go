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
	ID           uint64    `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	Symbol       string    `gorm:"size:20;index;not null;comment:交易对代码，如 BTCUSDT" json:"symbol"`
	IntervalType string    `gorm:"column:interval_type;size:10;index;not null;default:'1m';comment:信号所属周期" json:"interval_type"`
	OpenTime     int64     `gorm:"column:open_time;index;not null;default:0;comment:对齐的 K 线起始时间，Unix 毫秒" json:"open_time"`
	Action       string    `gorm:"column:signal;size:20;not null;comment:综合动作信号，如 BUY、SELL、NEUTRAL" json:"signal"`
	Score        int       `gorm:"not null;comment:综合评分，范围 -100 到 +100" json:"score"`
	Confidence   int       `gorm:"not null;comment:信号置信度百分比" json:"confidence"`
	EntryPrice   float64   `gorm:"column:entry_price;type:decimal(18,8);not null;comment:建议入场价" json:"entry_price"`
	StopLoss     float64   `gorm:"column:stop_loss;type:decimal(18,8);not null;comment:建议止损价" json:"stop_loss"`
	TargetPrice  float64   `gorm:"column:target_price;type:decimal(18,8);not null;comment:建议止盈价" json:"target_price"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;comment:信号生成时间" json:"created_at"`

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

// TableComment 返回数据表注释。
func (Signal) TableComment() string {
	return "多因子交易信号快照，记录动作、分数、置信度和建议价位"
}
