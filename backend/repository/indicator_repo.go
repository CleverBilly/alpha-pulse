package repository

import (
	"alpha-pulse/backend/models"
	"gorm.io/gorm"
)

// IndicatorRepository 封装 indicators 表读写。
type IndicatorRepository struct {
	db *gorm.DB
}

// NewIndicatorRepository 创建 IndicatorRepository。
func NewIndicatorRepository(db *gorm.DB) *IndicatorRepository {
	return &IndicatorRepository{db: db}
}

// Create 写入或更新一条指标记录。
func (r *IndicatorRepository) Create(indicator *models.Indicator) error {
	return r.db.Clauses(IndicatorUpsertClause()).Create(indicator).Error
}

// GetLatest 查询指定交易对的最新指标。
func (r *IndicatorRepository) GetLatest(symbol string) (models.Indicator, error) {
	var indicator models.Indicator
	err := r.db.
		Where("symbol = ?", symbol).
		Order("id DESC").
		First(&indicator).Error
	return indicator, err
}

// GetLatestByInterval 查询指定交易对和周期的最新指标。
func (r *IndicatorRepository) GetLatestByInterval(symbol, interval string) (models.Indicator, error) {
	var indicator models.Indicator
	err := r.db.
		Where("symbol = ? AND interval_type = ?", symbol, interval).
		Order("open_time DESC, id DESC").
		First(&indicator).Error
	return indicator, err
}
