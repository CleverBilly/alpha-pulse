package repository

import (
	"alpha-pulse/backend/models"
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
