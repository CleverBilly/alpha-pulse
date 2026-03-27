package models

import "gorm.io/gorm"

// AutoMigrate 自动创建或更新数据表结构。
func AutoMigrate(db *gorm.DB) error {
	models := []any{
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
		&AlertRecord{},
		&AlertPreference{},
		&TradeSetting{},
		&TradeOrder{},
	}

	for _, model := range models {
		if err := autoMigrateWithTableComment(db, model); err != nil {
			return err
		}
	}

	return syncMySQLSchemaComments(db, models)
}
