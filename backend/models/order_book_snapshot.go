package models

import "time"

// OrderBookSnapshot 对应 order_book_snapshots 表，保存盘口快照原始数据。
type OrderBookSnapshot struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol       string    `gorm:"size:20;not null;uniqueIndex:idx_order_book_symbol_update_id,priority:1;index" json:"symbol"`
	LastUpdateID int64     `gorm:"column:last_update_id;not null;uniqueIndex:idx_order_book_symbol_update_id,priority:2" json:"last_update_id"`
	DepthLevel   int       `gorm:"column:depth_level;not null" json:"depth_level"`
	BidsJSON     string    `gorm:"column:bids_json;type:longtext;not null" json:"bids_json"`
	AsksJSON     string    `gorm:"column:asks_json;type:longtext;not null" json:"asks_json"`
	BestBidPrice float64   `gorm:"column:best_bid_price;type:decimal(18,8);not null" json:"best_bid_price"`
	BestAskPrice float64   `gorm:"column:best_ask_price;type:decimal(18,8);not null" json:"best_ask_price"`
	Spread       float64   `gorm:"column:spread;type:decimal(18,8);not null" json:"spread"`
	EventTime    int64     `gorm:"column:event_time;not null;index" json:"event_time"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 指定数据表名。
func (OrderBookSnapshot) TableName() string {
	return "order_book_snapshots"
}
