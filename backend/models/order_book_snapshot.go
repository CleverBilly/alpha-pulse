package models

import "time"

// OrderBookSnapshot 对应 order_book_snapshots 表，保存盘口快照原始数据。
type OrderBookSnapshot struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;comment:主键 ID;index:idx_order_book_symbol_event_update_lookup,priority:4" json:"id"`
	Symbol       string    `gorm:"size:20;not null;comment:交易对代码，如 BTCUSDT;uniqueIndex:idx_order_book_symbol_update_id,priority:1;index;index:idx_order_book_symbol_event_update_lookup,priority:1" json:"symbol"`
	LastUpdateID int64     `gorm:"column:last_update_id;not null;comment:盘口快照更新 ID;uniqueIndex:idx_order_book_symbol_update_id,priority:2;index:idx_order_book_symbol_event_update_lookup,priority:3" json:"last_update_id"`
	DepthLevel   int       `gorm:"column:depth_level;not null;comment:快照包含的深度档位数" json:"depth_level"`
	BidsJSON     string    `gorm:"column:bids_json;type:longtext;not null;comment:买盘深度原始 JSON" json:"bids_json"`
	AsksJSON     string    `gorm:"column:asks_json;type:longtext;not null;comment:卖盘深度原始 JSON" json:"asks_json"`
	BestBidPrice float64   `gorm:"column:best_bid_price;type:decimal(18,8);not null;comment:最优买价" json:"best_bid_price"`
	BestAskPrice float64   `gorm:"column:best_ask_price;type:decimal(18,8);not null;comment:最优卖价" json:"best_ask_price"`
	Spread       float64   `gorm:"column:spread;type:decimal(18,8);not null;comment:买卖价差" json:"spread"`
	EventTime    int64     `gorm:"column:event_time;not null;index;index:idx_order_book_symbol_event_update_lookup,priority:2;comment:快照事件时间，Unix 毫秒" json:"event_time"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;comment:入库时间" json:"created_at"`
}

// TableName 指定数据表名。
func (OrderBookSnapshot) TableName() string {
	return "order_book_snapshots"
}

// TableComment 返回数据表注释。
func (OrderBookSnapshot) TableComment() string {
	return "订单簿深度快照原始数据，为流动性和盘口迁移分析提供输入"
}
