package repository

import "gorm.io/gorm/clause"

func IndicatorUpsertClause() clause.OnConflict {
	return clause.OnConflict{
		Columns: snapshotSeriesConflictColumns(),
		DoUpdates: clause.AssignmentColumns([]string{
			"rsi",
			"macd",
			"macd_signal",
			"macd_histogram",
			"ema20",
			"ema50",
			"atr",
			"bollinger_upper",
			"bollinger_middle",
			"bollinger_lower",
			"vwap",
			"created_at",
		}),
	}
}

func OrderFlowUpsertClause() clause.OnConflict {
	return clause.OnConflict{
		Columns: snapshotSeriesConflictColumns(),
		DoUpdates: clause.AssignmentColumns([]string{
			"buy_volume",
			"sell_volume",
			"delta",
			"cvd",
			"buy_large_trade_count",
			"sell_large_trade_count",
			"buy_large_trade_notional",
			"sell_large_trade_notional",
			"large_trade_delta",
			"absorption_bias",
			"absorption_strength",
			"iceberg_bias",
			"iceberg_strength",
			"data_source",
			"created_at",
		}),
	}
}

func StructureUpsertClause() clause.OnConflict {
	return clause.OnConflict{
		Columns: snapshotSeriesConflictColumns(),
		DoUpdates: clause.AssignmentColumns([]string{
			"trend",
			"support",
			"resistance",
			"bos",
			"choch",
			"created_at",
		}),
	}
}

func LiquidityUpsertClause() clause.OnConflict {
	return clause.OnConflict{
		Columns: snapshotSeriesConflictColumns(),
		DoUpdates: clause.AssignmentColumns([]string{
			"buy_liquidity",
			"sell_liquidity",
			"sweep_type",
			"order_book_imbalance",
			"data_source",
			"created_at",
		}),
	}
}

func SignalUpsertClause() clause.OnConflict {
	return clause.OnConflict{
		Columns: snapshotSeriesConflictColumns(),
		DoUpdates: clause.AssignmentColumns([]string{
			"signal",
			"score",
			"confidence",
			"entry_price",
			"stop_loss",
			"target_price",
			"created_at",
		}),
	}
}

func snapshotSeriesConflictColumns() []clause.Column {
	return []clause.Column{
		{Name: "symbol"},
		{Name: "interval_type"},
		{Name: "open_time"},
	}
}
