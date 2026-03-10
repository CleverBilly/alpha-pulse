package repository

import (
	"alpha-pulse/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FeatureSnapshotRepository 封装 feature_snapshots 表读写。
type FeatureSnapshotRepository struct {
	db *gorm.DB
}

// NewFeatureSnapshotRepository 创建 FeatureSnapshotRepository。
func NewFeatureSnapshotRepository(db *gorm.DB) *FeatureSnapshotRepository {
	return &FeatureSnapshotRepository{db: db}
}

// Create 写入或更新一条特征快照。
func (r *FeatureSnapshotRepository) Create(snapshot *models.FeatureSnapshot) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "symbol"},
			{Name: "interval_type"},
			{Name: "open_time"},
			{Name: "snapshot_source"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"feature_version",
			"price",
			"signal_action",
			"signal_score",
			"signal_confidence",
			"snapshot_json",
			"created_at",
		}),
	}).Create(snapshot).Error
}

// GetLatestByInterval 查询指定交易对和周期的最新特征快照。
func (r *FeatureSnapshotRepository) GetLatestByInterval(symbol, interval string) (models.FeatureSnapshot, error) {
	var snapshot models.FeatureSnapshot
	err := r.db.
		Where("symbol = ? AND interval_type = ?", symbol, interval).
		Order("open_time DESC, id DESC").
		First(&snapshot).Error
	return snapshot, err
}
