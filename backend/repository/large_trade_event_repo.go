package repository

import (
	"alpha-pulse/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LargeTradeEventRepository 封装 large_trade_events 表读写。
type LargeTradeEventRepository struct {
	db *gorm.DB
}

// NewLargeTradeEventRepository 创建 LargeTradeEventRepository。
func NewLargeTradeEventRepository(db *gorm.DB) *LargeTradeEventRepository {
	return &LargeTradeEventRepository{db: db}
}

// CreateBatch 批量写入大单事件；同一 symbol + agg_trade_id 冲突时执行更新。
func (r *LargeTradeEventRepository) CreateBatch(events []models.LargeTradeEvent) error {
	if len(events) == 0 {
		return nil
	}

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "symbol"},
			{Name: "agg_trade_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"orderflow_id",
			"interval_type",
			"open_time",
			"side",
			"price",
			"quantity",
			"notional",
			"trade_time",
			"created_at",
		}),
	}).Create(&events).Error
}

// GetRecent 查询指定交易对最近 N 条大单事件，并按时间升序返回。
func (r *LargeTradeEventRepository) GetRecent(symbol string, limit int) ([]models.LargeTradeEvent, error) {
	if limit <= 0 {
		limit = 20
	}

	events := make([]models.LargeTradeEvent, 0, limit)
	err := r.db.
		Where("symbol = ?", symbol).
		Order("trade_time DESC, agg_trade_id DESC").
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
