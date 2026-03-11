package models

import "time"

// AlertPreference 对应 alert_preferences 表，保存单用户告警偏好。
type AlertPreference struct {
	ID                    uint64    `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	SingletonKey          string    `gorm:"column:singleton_key;size:32;uniqueIndex;not null;default:'default';comment:单用户配置固定键" json:"singleton_key"`
	FeishuEnabled         bool      `gorm:"column:feishu_enabled;not null;default:true;comment:是否启用飞书推送" json:"feishu_enabled"`
	BrowserEnabled        bool      `gorm:"column:browser_enabled;not null;default:true;comment:是否启用浏览器通知" json:"browser_enabled"`
	SetupReadyEnabled     bool      `gorm:"column:setup_ready_enabled;not null;default:true;comment:是否启用 setup_ready 事件" json:"setup_ready_enabled"`
	DirectionShiftEnabled bool      `gorm:"column:direction_shift_enabled;not null;default:true;comment:是否启用 direction_shift 事件" json:"direction_shift_enabled"`
	NoTradeEnabled        bool      `gorm:"column:no_trade_enabled;not null;default:true;comment:是否启用 no_trade 事件" json:"no_trade_enabled"`
	MinimumConfidence     int       `gorm:"column:minimum_confidence;not null;default:55;comment:最小置信度阈值" json:"minimum_confidence"`
	QuietHoursEnabled     bool      `gorm:"column:quiet_hours_enabled;not null;default:false;comment:是否启用静默时段" json:"quiet_hours_enabled"`
	QuietHoursStart       int       `gorm:"column:quiet_hours_start;not null;default:0;comment:静默开始小时，0-23" json:"quiet_hours_start"`
	QuietHoursEnd         int       `gorm:"column:quiet_hours_end;not null;default:8;comment:静默结束小时，0-23" json:"quiet_hours_end"`
	WatchedSymbols        string    `gorm:"column:watched_symbols;type:text;not null;comment:关注标的 CSV" json:"watched_symbols"`
	CreatedAt             time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定数据表名。
func (AlertPreference) TableName() string {
	return "alert_preferences"
}

// TableComment 返回数据表注释。
func (AlertPreference) TableComment() string {
	return "单用户告警偏好表，保存飞书、浏览器、静默时段和事件过滤设置"
}
