package repository

import (
	"fmt"

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
			"interval",
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

// FindPending 返回指定标的中指定类型且 outcome 为 pending 的记录（按 event_time 升序）。
func (r *AlertRecordRepository) FindPending(symbol, kind string, limit int) ([]models.AlertRecord, error) {
	if limit <= 0 {
		limit = 100
	}
	records := make([]models.AlertRecord, 0, limit)
	err := r.db.Where("symbol = ? AND kind = ? AND outcome = ?", symbol, kind, "pending").
		Order("event_time ASC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

// UpdateOutcome 更新单条记录的结果字段，不影响其他字段。
func (r *AlertRecordRepository) UpdateOutcome(id uint64, outcome string, outcomePrice float64, outcomeAt int64, actualRR float64) error {
	return r.db.Model(&models.AlertRecord{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"outcome":       outcome,
			"outcome_price": outcomePrice,
			"outcome_at":    outcomeAt,
			"actual_rr":     actualRR,
		}).Error
}

// AlertStats 告警结果统计摘要。
type AlertStats struct {
	Symbol          string  `json:"symbol"`
	Total           int     `json:"total"`
	TargetHit       int     `json:"target_hit"`
	StopHit         int     `json:"stop_hit"`
	Pending         int     `json:"pending"`
	Expired         int     `json:"expired"`
	WinRate         float64 `json:"win_rate"`
	AvgRR           float64 `json:"avg_rr"`
	SampleSizeLabel string  `json:"sample_size_label"`
}

// GetStats 统计指定标的最近 limit 条 setup_ready 信号的结果分布。
// 先用子查询拿到最近 limit 条记录的 ID，再对这批数据聚合，避免 MySQL 中
// LIMIT 与 GROUP BY 混用时的语义歧义。
func (r *AlertRecordRepository) GetStats(symbol string, limit int) (AlertStats, error) {
	if limit <= 0 {
		limit = 50
	}

	// 第一步：取最近 limit 条 setup_ready 记录的 ID。
	var recentIDs []uint64
	err := r.db.Model(&models.AlertRecord{}).
		Select("id").
		Where("symbol = ? AND kind = ?", symbol, "setup_ready").
		Order("event_time DESC").
		Limit(limit).
		Pluck("id", &recentIDs).Error
	if err != nil {
		return AlertStats{}, err
	}

	stats := AlertStats{Symbol: symbol, SampleSizeLabel: fmt.Sprintf("近 %d 条", 0)}
	if len(recentIDs) == 0 {
		return stats, nil
	}

	// 第二步：在这批 ID 上做 outcome 聚合。
	type outcomeCount struct {
		Outcome string
		Count   int
	}
	var counts []outcomeCount
	err = r.db.Model(&models.AlertRecord{}).
		Select("outcome, COUNT(*) as count").
		Where("id IN ?", recentIDs).
		Group("outcome").
		Scan(&counts).Error
	if err != nil {
		return AlertStats{}, err
	}

	for _, c := range counts {
		switch c.Outcome {
		case "target_hit":
			stats.TargetHit = c.Count
		case "stop_hit":
			stats.StopHit = c.Count
		case "pending":
			stats.Pending = c.Count
		case "expired":
			stats.Expired = c.Count
		}
		stats.Total += c.Count
	}

	decided := stats.TargetHit + stats.StopHit
	if decided > 0 {
		stats.WinRate = float64(stats.TargetHit) / float64(decided) * 100

		// 第三步：统计已结算记录的 actual_rr 均值。
		var avgResult struct{ Avg float64 }
		err = r.db.Model(&models.AlertRecord{}).
			Select("AVG(actual_rr) as avg").
			Where("id IN ? AND outcome IN ?", recentIDs, []string{"target_hit", "stop_hit"}).
			Scan(&avgResult).Error
		if err == nil {
			stats.AvgRR = avgResult.Avg
		}
	}

	stats.SampleSizeLabel = fmt.Sprintf("近 %d 条", stats.Total)
	return stats, nil
}
