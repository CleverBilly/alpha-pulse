package repository

import (
	"alpha-pulse/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AggTradeRepository 封装 agg_trades 表读写。
type AggTradeRepository struct {
	db *gorm.DB
}

// NewAggTradeRepository 创建 AggTradeRepository。
func NewAggTradeRepository(db *gorm.DB) *AggTradeRepository {
	return &AggTradeRepository{db: db}
}

// Create 写入一条聚合成交；若已存在则更新为最新值。
func (r *AggTradeRepository) Create(trade *models.AggTrade) error {
	return r.db.Clauses(aggTradeUpsertClause()).Create(trade).Error
}

// CreateBatch 批量写入聚合成交；同一成交 ID 冲突时执行更新。
func (r *AggTradeRepository) CreateBatch(trades []models.AggTrade) error {
	if len(trades) == 0 {
		return nil
	}

	return r.db.Clauses(aggTradeUpsertClause()).Create(&trades).Error
}

// GetRecent 查询最近 N 条聚合成交，并按时间升序返回。
func (r *AggTradeRepository) GetRecent(symbol string, limit int) ([]models.AggTrade, error) {
	var trades []models.AggTrade
	err := r.db.
		Where("symbol = ?", symbol).
		Order("trade_time DESC, agg_trade_id DESC").
		Limit(limit).
		Find(&trades).Error
	if err != nil {
		return nil, err
	}

	for left, right := 0, len(trades)-1; left < right; left, right = left+1, right-1 {
		trades[left], trades[right] = trades[right], trades[left]
	}

	return trades, nil
}

func aggTradeUpsertClause() clause.OnConflict {
	return clause.OnConflict{
		Columns: []clause.Column{
			{Name: "symbol"},
			{Name: "agg_trade_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"price",
			"quantity",
			"quote_quantity",
			"first_trade_id",
			"last_trade_id",
			"trade_time",
			"is_buyer_maker",
			"is_best_price_match",
			"created_at",
		}),
	}
}
