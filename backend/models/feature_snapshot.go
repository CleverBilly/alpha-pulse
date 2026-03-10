package models

import "time"

// FeatureSnapshot 对应 feature_snapshots 表，保存离线审计与回放所需的聚合特征快照。
type FeatureSnapshot struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol           string    `gorm:"size:20;index;not null;uniqueIndex:idx_feature_snapshot_unique,priority:1" json:"symbol"`
	IntervalType     string    `gorm:"column:interval_type;size:10;index;not null;default:'1m';uniqueIndex:idx_feature_snapshot_unique,priority:2" json:"interval_type"`
	OpenTime         int64     `gorm:"column:open_time;index;not null;default:0;uniqueIndex:idx_feature_snapshot_unique,priority:3" json:"open_time"`
	SnapshotSource   string    `gorm:"column:snapshot_source;size:32;not null;default:'market_snapshot';uniqueIndex:idx_feature_snapshot_unique,priority:4" json:"snapshot_source"`
	FeatureVersion   string    `gorm:"column:feature_version;size:16;not null;default:'v1'" json:"feature_version"`
	Price            float64   `gorm:"type:decimal(18,8);not null;default:0" json:"price"`
	SignalAction     string    `gorm:"column:signal_action;size:20;index;not null;default:'NEUTRAL'" json:"signal_action"`
	SignalScore      int       `gorm:"column:signal_score;index;not null;default:0" json:"signal_score"`
	SignalConfidence int       `gorm:"column:signal_confidence;not null;default:0" json:"signal_confidence"`
	SnapshotJSON     string    `gorm:"column:snapshot_json;type:longtext;not null" json:"snapshot_json"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 指定数据表名。
func (FeatureSnapshot) TableName() string {
	return "feature_snapshots"
}
