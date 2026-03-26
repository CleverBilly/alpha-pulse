package repository

import (
	"alpha-pulse/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// KlineRepository 封装 kline 表读写。
type KlineRepository struct {
	db *gorm.DB
}

// NewKlineRepository 创建 KlineRepository。
func NewKlineRepository(db *gorm.DB) *KlineRepository {
	return &KlineRepository{db: db}
}

// Create 写入一条 K 线；若同一交易对、周期和开盘时间已存在则更新。
func (r *KlineRepository) Create(kline *models.Kline) error {
	return r.db.Clauses(klineUpsertClause()).Create(kline).Error
}

// CreateBatch 批量写入 K 线；已存在的数据会被最新值覆盖。
func (r *KlineRepository) CreateBatch(klines []models.Kline) error {
	if len(klines) == 0 {
		return nil
	}

	return r.db.Clauses(klineUpsertClause()).Create(&klines).Error
}

// GetLatest 查询指定交易对和周期的最新 K 线。
func (r *KlineRepository) GetLatest(symbol, interval string) (models.Kline, error) {
	var kline models.Kline
	err := r.db.
		Where("symbol = ? AND interval_type = ?", symbol, interval).
		Order("open_time DESC").
		First(&kline).Error
	return kline, err
}

// GetRecent 查询指定交易对和周期的最近 N 根 K 线，并按时间升序返回。
func (r *KlineRepository) GetRecent(symbol, interval string, limit int) ([]models.Kline, error) {
	var klines []models.Kline
	err := r.db.
		Where("symbol = ? AND interval_type = ?", symbol, interval).
		Order("open_time DESC").
		Limit(limit).
		Find(&klines).Error
	if err != nil {
		return nil, err
	}

	// 查询时用倒序拿最新数据，返回前恢复为升序，便于指标引擎直接消费。
	for left, right := 0, len(klines)-1; left < right; left, right = left+1, right-1 {
		klines[left], klines[right] = klines[right], klines[left]
	}

	return klines, nil
}

func klineUpsertClause() clause.OnConflict {
	return clause.OnConflict{
		Columns: []clause.Column{
			{Name: "symbol"},
			{Name: "interval_type"},
			{Name: "open_time"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"open_price",
			"high_price",
			"low_price",
			"close_price",
			"volume",
			"created_at",
		}),
	}
}

// FindAfter 返回指定时间点（afterMs，Unix 毫秒）之后的 K 线，按时间升序。
func (r *KlineRepository) FindAfter(symbol, interval string, afterMs int64, limit int) ([]models.Kline, error) {
	if limit <= 0 {
		limit = 60
	}
	klines := make([]models.Kline, 0, limit)
	err := r.db.
		Where("symbol = ? AND interval_type = ? AND open_time > ?", symbol, interval, afterMs).
		Order("open_time ASC").
		Limit(limit).
		Find(&klines).Error
	return klines, err
}

// FindBefore 返回指定时间点（beforeMs，Unix 毫秒）之前的最近 N 根 K 线，按时间升序返回。
func (r *KlineRepository) FindBefore(symbol, interval string, beforeMs int64, limit int) ([]models.Kline, error) {
	if limit <= 0 {
		limit = 20
	}
	klines := make([]models.Kline, 0, limit)
	err := r.db.
		Where("symbol = ? AND interval_type = ? AND open_time < ?", symbol, interval, beforeMs).
		Order("open_time DESC").
		Limit(limit).
		Find(&klines).Error
	if err != nil {
		return nil, err
	}
	// 倒序查最新的，返回时恢复升序
	for left, right := 0, len(klines)-1; left < right; left, right = left+1, right-1 {
		klines[left], klines[right] = klines[right], klines[left]
	}
	return klines, nil
}
