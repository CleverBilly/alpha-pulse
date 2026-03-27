package models

import "time"

// TradeSetting 对应 trade_settings 表，保存自动交易运行时配置。
type TradeSetting struct {
	ID                  uint64    `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	SingletonKey        string    `gorm:"column:singleton_key;size:32;uniqueIndex;not null;default:'default';comment:单例配置固定键" json:"singleton_key"`
	AutoExecuteEnabled  bool      `gorm:"column:auto_execute_enabled;not null;default:false;comment:是否允许自动执行" json:"auto_execute_enabled"`
	AllowedSymbols      string    `gorm:"column:allowed_symbols;type:text;not null;comment:允许自动交易的标的 CSV" json:"allowed_symbols"`
	RiskPct             float64   `gorm:"column:risk_pct;type:decimal(10,4);not null;default:2;comment:单笔风险百分比" json:"risk_pct"`
	MinRiskReward       float64   `gorm:"column:min_risk_reward;type:decimal(10,4);not null;default:1;comment:最低盈亏比阈值" json:"min_risk_reward"`
	EntryTimeoutSeconds int       `gorm:"column:entry_timeout_seconds;not null;default:45;comment:限价单超时秒数" json:"entry_timeout_seconds"`
	MaxOpenPositions    int       `gorm:"column:max_open_positions;not null;default:1;comment:最大同时持仓数" json:"max_open_positions"`
	SyncEnabled         bool      `gorm:"column:sync_enabled;not null;default:true;comment:是否启用持仓同步" json:"sync_enabled"`
	UpdatedBy           string    `gorm:"column:updated_by;size:64;not null;default:'';comment:最近一次配置更新人" json:"updated_by"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定数据表名。
func (TradeSetting) TableName() string {
	return "trade_settings"
}

// TableComment 返回数据表注释。
func (TradeSetting) TableComment() string {
	return "自动交易运行时配置表，保存白名单、风控阈值和自动执行状态"
}
