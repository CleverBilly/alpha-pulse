package models

import "time"

// AlertRecord 对应 alert_records 表，保存方向告警的结构化历史与完整载荷。
type AlertRecord struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	AlertID           string    `gorm:"column:alert_id;size:96;uniqueIndex;not null;comment:告警事件唯一 ID" json:"alert_id"`
	Symbol            string    `gorm:"size:20;index;not null;comment:交易对代码，如 BTCUSDT" json:"symbol"`
	Kind              string    `gorm:"size:32;index;not null;comment:告警类型，如 setup_ready/no_trade" json:"kind"`
	Severity          string    `gorm:"size:24;index;not null;comment:告警级别，如 A/warning/info" json:"severity"`
	DirectionState    string    `gorm:"column:direction_state;size:32;index;not null;default:'invalid';comment:方向状态快照" json:"direction_state"`
	Tradable          bool      `gorm:"column:tradable;index;not null;default:false;comment:当时是否允许跟踪" json:"tradable"`
	SetupReady        bool      `gorm:"column:setup_ready;index;not null;default:false;comment:当时 setup 是否完整" json:"setup_ready"`
	TradeabilityLabel string    `gorm:"column:tradeability_label;size:32;not null;default:'';comment:交易性标签" json:"tradeability_label"`
	Title             string    `gorm:"size:128;not null;default:'';comment:告警标题" json:"title"`
	Verdict           string    `gorm:"size:32;not null;default:'';comment:方向结论" json:"verdict"`
	Summary           string    `gorm:"type:text;not null;comment:告警摘要" json:"summary"`
	Confidence        int       `gorm:"not null;default:0;comment:置信度百分比" json:"confidence"`
	RiskLabel         string    `gorm:"column:risk_label;size:24;not null;default:'';comment:风险标签" json:"risk_label"`
	EntryPrice        float64   `gorm:"column:entry_price;type:decimal(18,8);not null;default:0;comment:候选进场价" json:"entry_price"`
	StopLoss          float64   `gorm:"column:stop_loss;type:decimal(18,8);not null;default:0;comment:失效位" json:"stop_loss"`
	TargetPrice       float64   `gorm:"column:target_price;type:decimal(18,8);not null;default:0;comment:目标位" json:"target_price"`
	RiskReward        float64   `gorm:"column:risk_reward;type:decimal(10,4);not null;default:0;comment:盈亏比" json:"risk_reward"`
	EventTime         int64     `gorm:"column:event_time;index;not null;default:0;comment:告警产生时间，Unix 毫秒" json:"event_time"`
	PayloadJSON       string    `gorm:"column:payload_json;type:longtext;not null;comment:完整 alert event JSON" json:"payload_json"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime;comment:保存时间" json:"created_at"`
}

// TableName 指定数据表名。
func (AlertRecord) TableName() string {
	return "alert_records"
}

// TableComment 返回数据表注释。
func (AlertRecord) TableComment() string {
	return "方向告警归档表，保存 setup/no-trade/history 事件以供复盘和审计"
}
