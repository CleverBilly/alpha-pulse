package repository

import (
	"alpha-pulse/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MicrostructureEventRepository 封装 microstructure_events 表读写。
type MicrostructureEventRepository struct {
	db *gorm.DB
}

// NewMicrostructureEventRepository 创建 MicrostructureEventRepository。
func NewMicrostructureEventRepository(db *gorm.DB) *MicrostructureEventRepository {
	return &MicrostructureEventRepository{db: db}
}

// CreateBatch 批量写入微结构事件；同一分析窗口内重复事件将忽略。
func (r *MicrostructureEventRepository) CreateBatch(events []models.MicrostructureEvent) error {
	if len(events) == 0 {
		return nil
	}

	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&events).Error
}

// GetRecentByInterval 查询指定交易对和周期的最近 N 条微结构事件，并按时间升序返回。
func (r *MicrostructureEventRepository) GetRecentByInterval(symbol, interval string, limit int) ([]models.MicrostructureEvent, error) {
	if limit <= 0 {
		limit = 20
	}

	events := make([]models.MicrostructureEvent, 0, limit)
	err := r.db.
		Where("symbol = ? AND interval_type = ?", symbol, interval).
		Order("trade_time DESC, id DESC").
		Limit(limit).
		Find(&events).Error
	if err != nil {
		return nil, err
	}

	for left, right := 0, len(events)-1; left < right; left, right = left+1, right-1 {
		events[left], events[right] = events[right], events[left]
	}

	return events, nil
}

// GetByTradeTimeRange 查询指定交易对和周期在时间窗口内的微结构事件，并按时间升序返回。
func (r *MicrostructureEventRepository) GetByTradeTimeRange(
	symbol, interval string,
	fromTradeTime, toTradeTime int64,
	limit int,
) ([]models.MicrostructureEvent, error) {
	if limit <= 0 {
		limit = 120
	}

	events := make([]models.MicrostructureEvent, 0, limit)
	query := r.db.
		Where("symbol = ? AND interval_type = ?", symbol, interval).
		Where("trade_time >= ?", fromTradeTime)
	if toTradeTime > 0 {
		query = query.Where("trade_time < ?", toTradeTime)
	}

	err := query.
		Order("trade_time ASC, id ASC").
		Limit(limit).
		Find(&events).Error
	if err != nil {
		return nil, err
	}

	return events, nil
}
