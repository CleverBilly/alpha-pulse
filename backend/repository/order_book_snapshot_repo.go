package repository

import (
	"alpha-pulse/backend/models"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OrderBookSnapshotRepository 封装 order_book_snapshots 表读写。
type OrderBookSnapshotRepository struct {
	db *gorm.DB
}

// NewOrderBookSnapshotRepository 创建 OrderBookSnapshotRepository。
func NewOrderBookSnapshotRepository(db *gorm.DB) *OrderBookSnapshotRepository {
	return &OrderBookSnapshotRepository{db: db}
}

// Create 写入一条盘口快照；若 update_id 相同则更新。
func (r *OrderBookSnapshotRepository) Create(snapshot *models.OrderBookSnapshot) error {
	return r.db.Clauses(orderBookSnapshotUpsertClause()).Create(snapshot).Error
}

// GetLatest 查询指定交易对的最新盘口快照。
func (r *OrderBookSnapshotRepository) GetLatest(symbol string) (models.OrderBookSnapshot, error) {
	var snapshot models.OrderBookSnapshot
	err := r.db.
		Where("symbol = ?", symbol).
		Order("event_time DESC, last_update_id DESC").
		First(&snapshot).Error
	return snapshot, err
}

// GetLatestBefore 查询指定时间点之前最近的一条盘口快照。
func (r *OrderBookSnapshotRepository) GetLatestBefore(symbol string, eventTime int64) (models.OrderBookSnapshot, error) {
	var snapshot models.OrderBookSnapshot
	err := r.db.
		Where("symbol = ? AND event_time <= ?", symbol, eventTime).
		Order("event_time DESC, last_update_id DESC").
		First(&snapshot).Error
	return snapshot, err
}

// GetRecent 查询指定交易对最近 N 条盘口快照，并按事件时间升序返回。
func (r *OrderBookSnapshotRepository) GetRecent(symbol string, limit int) ([]models.OrderBookSnapshot, error) {
	if limit <= 0 {
		limit = 8
	}

	snapshots := make([]models.OrderBookSnapshot, 0, limit)
	err := r.db.
		Where("symbol = ?", symbol).
		Order("event_time DESC, last_update_id DESC").
		Limit(limit).
		Find(&snapshots).Error
	if err != nil {
		return nil, err
	}

	for left, right := 0, len(snapshots)-1; left < right; left, right = left+1, right-1 {
		snapshots[left], snapshots[right] = snapshots[right], snapshots[left]
	}
	return snapshots, nil
}

// GetSeriesWindow 查询流动性时间序列所需的盘口快照窗口。
// 返回值会额外包含一条早于 startTime 的最近快照，便于后续在内存中回放到每个时间点。
func (r *OrderBookSnapshotRepository) GetSeriesWindow(symbol string, startTime, endTime int64) ([]models.OrderBookSnapshot, error) {
	if endTime <= 0 {
		return nil, nil
	}

	snapshots := make([]models.OrderBookSnapshot, 0, 128)
	if startTime > 0 {
		var leading models.OrderBookSnapshot
		err := r.db.
			Where("symbol = ? AND event_time < ?", symbol, startTime).
			Order("event_time DESC, last_update_id DESC, id DESC").
			Limit(1).
			First(&leading).Error
		switch {
		case err == nil:
			snapshots = append(snapshots, leading)
		case errors.Is(err, gorm.ErrRecordNotFound):
			// 没有前导快照时直接继续查窗口。
		default:
			return nil, err
		}
	}

	window := make([]models.OrderBookSnapshot, 0, 128)
	query := r.db.Where("symbol = ? AND event_time <= ?", symbol, endTime)
	if startTime > 0 {
		query = query.Where("event_time >= ?", startTime)
	}
	if err := query.
		Order("event_time ASC, last_update_id ASC, id ASC").
		Find(&window).Error; err != nil {
		return nil, err
	}

	snapshots = append(snapshots, window...)
	return snapshots, nil
}

func orderBookSnapshotUpsertClause() clause.OnConflict {
	return clause.OnConflict{
		Columns: []clause.Column{
			{Name: "symbol"},
			{Name: "last_update_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"depth_level",
			"bids_json",
			"asks_json",
			"best_bid_price",
			"best_ask_price",
			"spread",
			"event_time",
			"created_at",
		}),
	}
}
