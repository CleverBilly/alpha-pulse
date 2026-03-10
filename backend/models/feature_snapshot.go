package models

import "time"

// FeatureSnapshot 对应 feature_snapshots 表，保存离线审计与回放所需的聚合特征快照。
type FeatureSnapshot struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	Symbol           string    `gorm:"size:20;index;not null;comment:交易对代码，如 BTCUSDT;uniqueIndex:idx_feature_snapshot_unique,priority:1" json:"symbol"`
	IntervalType     string    `gorm:"column:interval_type;size:10;index;not null;default:'1m';comment:快照所属周期;uniqueIndex:idx_feature_snapshot_unique,priority:2" json:"interval_type"`
	OpenTime         int64     `gorm:"column:open_time;index;not null;default:0;comment:对齐的 K 线起始时间，Unix 毫秒;uniqueIndex:idx_feature_snapshot_unique,priority:3" json:"open_time"`
	SnapshotSource   string    `gorm:"column:snapshot_source;size:32;not null;default:'market_snapshot';comment:快照来源标识;uniqueIndex:idx_feature_snapshot_unique,priority:4" json:"snapshot_source"`
	FeatureVersion   string    `gorm:"column:feature_version;size:16;not null;default:'v1';comment:特征快照版本号" json:"feature_version"`
	Price            float64   `gorm:"type:decimal(18,8);not null;default:0;comment:快照时刻价格" json:"price"`
	SignalAction     string    `gorm:"column:signal_action;size:20;index;not null;default:'NEUTRAL';comment:信号方向" json:"signal_action"`
	SignalScore      int       `gorm:"column:signal_score;index;not null;default:0;comment:信号分数" json:"signal_score"`
	SignalConfidence int       `gorm:"column:signal_confidence;not null;default:0;comment:信号置信度百分比" json:"signal_confidence"`
	SnapshotJSON     string    `gorm:"column:snapshot_json;type:longtext;not null;comment:完整 market snapshot JSON" json:"snapshot_json"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime;comment:保存时间" json:"created_at"`
}

// TableName 指定数据表名。
func (FeatureSnapshot) TableName() string {
	return "feature_snapshots"
}

// TableComment 返回数据表注释。
func (FeatureSnapshot) TableComment() string {
	return "聚合特征快照归档表，保存完整 market snapshot 以供审计、训练和回放"
}
