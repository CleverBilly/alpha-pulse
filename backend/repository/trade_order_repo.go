package repository

import (
	"alpha-pulse/backend/models"
	"gorm.io/gorm"
)

// TradeOrderRepository 封装 trade_orders 表读写。
type TradeOrderRepository struct {
	db *gorm.DB
}

// NewTradeOrderRepository 创建 TradeOrderRepository。
func NewTradeOrderRepository(db *gorm.DB) *TradeOrderRepository {
	return &TradeOrderRepository{db: db}
}

// Create 创建交易订单。
func (r *TradeOrderRepository) Create(order *models.TradeOrder) error {
	return r.db.Create(order).Error
}

// FindPendingFill 查询待成交挂单。
func (r *TradeOrderRepository) FindPendingFill(limit int) ([]models.TradeOrder, error) {
	if limit <= 0 {
		limit = 20
	}

	var orders []models.TradeOrder
	err := r.db.
		Where("status = ?", "pending_fill").
		Order("created_at asc").
		Limit(limit).
		Find(&orders).Error
	return orders, err
}

// FindOpen 查询指定标的的 open 持仓。
func (r *TradeOrderRepository) FindOpen(symbol string) ([]models.TradeOrder, error) {
	var orders []models.TradeOrder
	err := r.db.
		Where("symbol = ? AND status = ?", symbol, "open").
		Order("created_at desc").
		Find(&orders).Error
	return orders, err
}

// FindAllOpen 查询全部 open 持仓。
func (r *TradeOrderRepository) FindAllOpen() ([]models.TradeOrder, error) {
	var orders []models.TradeOrder
	err := r.db.
		Where("status = ?", "open").
		Order("created_at desc").
		Find(&orders).Error
	return orders, err
}

// FindByAlertID 查询指定 alert_id 的订单。
func (r *TradeOrderRepository) FindByAlertID(alertID string) (models.TradeOrder, error) {
	var order models.TradeOrder
	err := r.db.Where("alert_id = ?", alertID).First(&order).Error
	return order, err
}

// FindBySourceAndSymbol 查询指定来源和标的的订单。
func (r *TradeOrderRepository) FindBySourceAndSymbol(source, symbol string) (models.TradeOrder, error) {
	var order models.TradeOrder
	err := r.db.Where("source = ? AND symbol = ?", source, symbol).Order("created_at desc").First(&order).Error
	return order, err
}

// FindByID 查询指定 ID 订单。
func (r *TradeOrderRepository) FindByID(id uint64) (models.TradeOrder, error) {
	var order models.TradeOrder
	err := r.db.First(&order, id).Error
	return order, err
}

// List 查询订单列表。
func (r *TradeOrderRepository) List(limit int, symbol, status, source string) ([]models.TradeOrder, error) {
	if limit <= 0 {
		limit = 50
	}

	query := r.db.Model(&models.TradeOrder{})
	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if source != "" {
		query = query.Where("source = ?", source)
	}

	var orders []models.TradeOrder
	err := query.Order("created_at desc").Limit(limit).Find(&orders).Error
	return orders, err
}

// Save 保存订单状态变更。
func (r *TradeOrderRepository) Save(order *models.TradeOrder) error {
	return r.db.Save(order).Error
}
