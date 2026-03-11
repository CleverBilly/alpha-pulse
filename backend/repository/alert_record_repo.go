package repository

import (
	"alpha-pulse/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AlertRecordRepository 封装 alert_records 表读写。
type AlertRecordRepository struct {
	db *gorm.DB
}

// NewAlertRecordRepository 创建 AlertRecordRepository。
func NewAlertRecordRepository(db *gorm.DB) *AlertRecordRepository {
	return &AlertRecordRepository{db: db}
}

// Create 写入或更新一条告警记录。
func (r *AlertRecordRepository) Create(record *models.AlertRecord) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "alert_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"symbol",
			"kind",
			"severity",
			"direction_state",
			"tradable",
			"setup_ready",
			"tradeability_label",
			"title",
			"verdict",
			"summary",
			"confidence",
			"risk_label",
			"entry_price",
			"stop_loss",
			"target_price",
			"risk_reward",
			"event_time",
			"payload_json",
			"created_at",
		}),
	}).Create(record).Error
}

// ListRecent 按时间倒序返回最近的告警记录。
func (r *AlertRecordRepository) ListRecent(limit int) ([]models.AlertRecord, error) {
	if limit <= 0 {
		limit = 20
	}

	records := make([]models.AlertRecord, 0, limit)
	err := r.db.Order("event_time DESC, id DESC").Limit(limit).Find(&records).Error
	return records, err
}

// GetLatestBySymbol 查询指定交易对最近一条告警记录。
func (r *AlertRecordRepository) GetLatestBySymbol(symbol string) (models.AlertRecord, error) {
	var record models.AlertRecord
	err := r.db.Where("symbol = ?", symbol).Order("event_time DESC, id DESC").First(&record).Error
	return record, err
}
