package models

import "time"

// AggTrade 对应 agg_trades 表，保存 Binance 聚合成交原始数据。
type AggTrade struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol           string    `gorm:"size:20;not null;uniqueIndex:idx_agg_trade_symbol_trade_id,priority:1;index" json:"symbol"`
	AggTradeID       int64     `gorm:"column:agg_trade_id;not null;uniqueIndex:idx_agg_trade_symbol_trade_id,priority:2" json:"agg_trade_id"`
	Price            float64   `gorm:"type:decimal(18,8);not null" json:"price"`
	Quantity         float64   `gorm:"type:decimal(24,8);not null" json:"quantity"`
	QuoteQuantity    float64   `gorm:"column:quote_quantity;type:decimal(24,8);not null" json:"quote_quantity"`
	FirstTradeID     int64     `gorm:"column:first_trade_id;not null" json:"first_trade_id"`
	LastTradeID      int64     `gorm:"column:last_trade_id;not null" json:"last_trade_id"`
	TradeTime        int64     `gorm:"column:trade_time;not null;index" json:"trade_time"`
	IsBuyerMaker     bool      `gorm:"column:is_buyer_maker;not null" json:"is_buyer_maker"`
	IsBestPriceMatch bool      `gorm:"column:is_best_price_match;not null" json:"is_best_price_match"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 指定数据表名。
func (AggTrade) TableName() string {
	return "agg_trades"
}
