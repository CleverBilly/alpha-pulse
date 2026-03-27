package repository

import (
	"alpha-pulse/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

// TradeSettingRepository 封装 trade_settings 表读写。
type TradeSettingRepository struct {
	db *gorm.DB
}

// NewTradeSettingRepository 创建 TradeSettingRepository。
func NewTradeSettingRepository(db *gorm.DB) *TradeSettingRepository {
	return &TradeSettingRepository{db: db}
}

// GetDefault 查询默认运行时配置。
func (r *TradeSettingRepository) GetDefault() (models.TradeSetting, error) {
	var record models.TradeSetting
	err := r.db.Where("singleton_key = ?", "default").First(&record).Error
	return record, err
}

// Save 保存默认运行时配置。
func (r *TradeSettingRepository) Save(record *models.TradeSetting) error {
	now := time.Now()
	values := map[string]any{
		"singleton_key":          record.SingletonKey,
		"auto_execute_enabled":   record.AutoExecuteEnabled,
		"allowed_symbols":        record.AllowedSymbols,
		"risk_pct":               record.RiskPct,
		"min_risk_reward":        record.MinRiskReward,
		"entry_timeout_seconds":  record.EntryTimeoutSeconds,
		"max_open_positions":     record.MaxOpenPositions,
		"sync_enabled":           record.SyncEnabled,
		"updated_by":             record.UpdatedBy,
		"created_at":             now,
		"updated_at":             now,
	}
	return r.db.Model(&models.TradeSetting{}).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "singleton_key"}},
		DoUpdates: clause.Assignments(map[string]any{
			"auto_execute_enabled":  record.AutoExecuteEnabled,
			"allowed_symbols":       record.AllowedSymbols,
			"risk_pct":              record.RiskPct,
			"min_risk_reward":       record.MinRiskReward,
			"entry_timeout_seconds": record.EntryTimeoutSeconds,
			"max_open_positions":    record.MaxOpenPositions,
			"sync_enabled":          record.SyncEnabled,
			"updated_by":            record.UpdatedBy,
			"updated_at":            now,
		}),
	}).Create(values).Error
}
