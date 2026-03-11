package repository

import (
	"alpha-pulse/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

// AlertPreferenceRepository 封装 alert_preferences 表读写。
type AlertPreferenceRepository struct {
	db *gorm.DB
}

// NewAlertPreferenceRepository 创建 AlertPreferenceRepository。
func NewAlertPreferenceRepository(db *gorm.DB) *AlertPreferenceRepository {
	return &AlertPreferenceRepository{db: db}
}

// GetDefault 查询默认单用户偏好。
func (r *AlertPreferenceRepository) GetDefault() (models.AlertPreference, error) {
	var record models.AlertPreference
	err := r.db.Where("singleton_key = ?", "default").First(&record).Error
	return record, err
}

// Save 保存默认单用户偏好。
func (r *AlertPreferenceRepository) Save(record *models.AlertPreference) error {
	now := time.Now()
	values := map[string]any{
		"singleton_key":           record.SingletonKey,
		"feishu_enabled":          record.FeishuEnabled,
		"browser_enabled":         record.BrowserEnabled,
		"setup_ready_enabled":     record.SetupReadyEnabled,
		"direction_shift_enabled": record.DirectionShiftEnabled,
		"no_trade_enabled":        record.NoTradeEnabled,
		"minimum_confidence":      record.MinimumConfidence,
		"quiet_hours_enabled":     record.QuietHoursEnabled,
		"quiet_hours_start":       record.QuietHoursStart,
		"quiet_hours_end":         record.QuietHoursEnd,
		"watched_symbols":         record.WatchedSymbols,
		"created_at":              now,
		"updated_at":              now,
	}
	return r.db.Model(&models.AlertPreference{}).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "singleton_key"}},
		DoUpdates: clause.Assignments(map[string]any{
			"feishu_enabled":          record.FeishuEnabled,
			"browser_enabled":         record.BrowserEnabled,
			"setup_ready_enabled":     record.SetupReadyEnabled,
			"direction_shift_enabled": record.DirectionShiftEnabled,
			"no_trade_enabled":        record.NoTradeEnabled,
			"minimum_confidence":      record.MinimumConfidence,
			"quiet_hours_enabled":     record.QuietHoursEnabled,
			"quiet_hours_start":       record.QuietHoursStart,
			"quiet_hours_end":         record.QuietHoursEnd,
			"watched_symbols":         record.WatchedSymbols,
			"updated_at":              now,
		}),
	}).Create(values).Error
}
