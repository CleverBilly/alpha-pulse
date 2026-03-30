package models

import "time"

// AggTrade 对应 agg_trades 表，保存 Binance 聚合成交原始数据。
type AggTrade struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	Symbol           string    `gorm:"size:20;not null;comment:交易对代码，如 BTCUSDT;uniqueIndex:idx_agg_trade_symbol_trade_id,priority:1;index;index:idx_agg_trade_symbol_time_lookup,priority:1" json:"symbol"`
	AggTradeID       int64     `gorm:"column:agg_trade_id;not null;comment:Binance 聚合成交 ID;uniqueIndex:idx_agg_trade_symbol_trade_id,priority:2;index:idx_agg_trade_symbol_time_lookup,priority:3" json:"agg_trade_id"`
	Price            float64   `gorm:"type:decimal(18,8);not null;comment:成交价" json:"price"`
	Quantity         float64   `gorm:"type:decimal(24,8);not null;comment:成交数量，基础资产数量" json:"quantity"`
	QuoteQuantity    float64   `gorm:"column:quote_quantity;type:decimal(24,8);not null;comment:成交额，计价资产数量" json:"quote_quantity"`
	FirstTradeID     int64     `gorm:"column:first_trade_id;not null;comment:聚合成交覆盖的首笔原始成交 ID" json:"first_trade_id"`
	LastTradeID      int64     `gorm:"column:last_trade_id;not null;comment:聚合成交覆盖的末笔原始成交 ID" json:"last_trade_id"`
	TradeTime        int64     `gorm:"column:trade_time;not null;index;index:idx_agg_trade_symbol_time_lookup,priority:2;comment:成交时间，Unix 毫秒" json:"trade_time"`
	IsBuyerMaker     bool      `gorm:"column:is_buyer_maker;not null;comment:买方是否为挂单方" json:"is_buyer_maker"`
	IsBestPriceMatch bool      `gorm:"column:is_best_price_match;not null;comment:是否最佳价格成交" json:"is_best_price_match"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime;comment:入库时间" json:"created_at"`
}

// TableName 指定数据表名。
func (AggTrade) TableName() string {
	return "agg_trades"
}

// TableComment 返回数据表注释。
func (AggTrade) TableComment() string {
	return "Binance 聚合成交原始数据，为订单流和微结构分析提供真实成交输入"
}
