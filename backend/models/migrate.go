package models

import "gorm.io/gorm"

// AutoMigrate 自动创建或更新数据表结构。
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Kline{},
		&AggTrade{},
		&LargeTradeEvent{},
		&OrderBookSnapshot{},
		&Indicator{},
		&OrderFlow{},
		&MicrostructureEvent{},
		&Structure{},
		&Liquidity{},
		&Signal{},
		&FeatureSnapshot{},
	)
}
