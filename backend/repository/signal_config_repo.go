package repository

import (
	"alpha-pulse/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SignalConfigRepository 封装 signal_configs 表读写。
type SignalConfigRepository struct {
	db *gorm.DB
}

// NewSignalConfigRepository 创建 SignalConfigRepository。
func NewSignalConfigRepository(db *gorm.DB) *SignalConfigRepository {
	return &SignalConfigRepository{db: db}
}

// GetAll 返回所有信号配置。
func (r *SignalConfigRepository) GetAll() ([]models.SignalConfig, error) {
	var configs []models.SignalConfig
	return configs, r.db.Find(&configs).Error
}

// GetBySymbolInterval 按 symbol + interval 查询所有配置。
func (r *SignalConfigRepository) GetBySymbolInterval(symbol, interval string) ([]models.SignalConfig, error) {
	var configs []models.SignalConfig
	return configs, r.db.Where("symbol = ? AND interval = ?", symbol, interval).Find(&configs).Error
}

// Upsert 插入或更新（按 symbol+interval+key 唯一键）。
func (r *SignalConfigRepository) Upsert(cfg models.SignalConfig) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "symbol"}, {Name: "interval"}, {Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&cfg).Error
}
