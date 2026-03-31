package repository

import (
	"alpha-pulse/backend/models"
	"gorm.io/gorm"
)

// SignalRepository 封装 signals 表读写。
type SignalRepository struct {
	db *gorm.DB
}

// NewSignalRepository 创建 SignalRepository。
func NewSignalRepository(db *gorm.DB) *SignalRepository {
	return &SignalRepository{db: db}
}

// Create 写入或更新一条交易信号。
func (r *SignalRepository) Create(signal *models.Signal) error {
	return r.db.Clauses(SignalUpsertClause()).Create(signal).Error
}

// GetLatest 查询指定交易对的最新信号。
func (r *SignalRepository) GetLatest(symbol string) (models.Signal, error) {
	var signal models.Signal
	err := r.db.
		Where("symbol = ?", symbol).
		Order("id DESC").
		First(&signal).Error
	return signal, err
}

// GetLatestByInterval 查询指定交易对和周期的最新信号。
func (r *SignalRepository) GetLatestByInterval(symbol, interval string) (models.Signal, error) {
	var signal models.Signal
	err := r.db.
		Where("symbol = ? AND interval_type = ?", symbol, interval).
		Order("open_time DESC, id DESC").
		First(&signal).Error
	return signal, err
}

// GetRecentByInterval 查询指定交易对和周期的最近 N 条信号。
func (r *SignalRepository) GetRecentByInterval(symbol, interval string, limit int) ([]models.Signal, error) {
	if limit <= 0 {
		limit = 20
	}

	signals := make([]models.Signal, 0, limit)
	err := r.db.
		Where("symbol = ? AND interval_type = ?", symbol, interval).
		Order("open_time DESC, id DESC").
		Limit(limit).
		Find(&signals).Error
	return signals, err
}
