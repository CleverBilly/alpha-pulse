package models

import "time"

// SignalConfig 存储信号引擎可热更新的阈值和权重。
// key 示例：buy_threshold, sell_threshold, orderflow_weight, trend_weight
type SignalConfig struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	Symbol    string    `gorm:"column:symbol;size:20;not null;uniqueIndex:idx_signal_config,priority:1;comment:交易对" json:"symbol"`
	Interval  string    `gorm:"column:interval;size:10;not null;uniqueIndex:idx_signal_config,priority:2;comment:时间周期" json:"interval"`
	Key       string    `gorm:"column:key;size:64;not null;uniqueIndex:idx_signal_config,priority:3;comment:配置键" json:"key"`
	Value     string    `gorm:"column:value;size:128;not null;comment:配置值，存为字符串" json:"value"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定数据表名。
func (SignalConfig) TableName() string {
	return "signal_configs"
}

// TableComment 返回数据表注释。
func (SignalConfig) TableComment() string {
	return "信号引擎可热更新的阈值和权重配置表"
}
