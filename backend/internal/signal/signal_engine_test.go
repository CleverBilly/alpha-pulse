package signal

import (
	"strconv"
	"strings"
	"testing"

	"alpha-pulse/backend/models"
)

func TestGenerateReturnsBuySignalForBullishConfluence(t *testing.T) {
	engine := NewEngine()

	result := engine.Generate(
		"BTCUSDT",
		65000,
		models.Indicator{
			RSI:             62,
			MACD:            180,
			MACDSignal:      120,
			MACDHistogram:   60,
			EMA20:           64850,
			EMA50:           64200,
			ATR:             420,
			BollingerUpper:  65800,
			BollingerMiddle: 64920,
			BollingerLower:  64040,
			VWAP:            64680,
		},
		models.OrderFlow{
			BuyVolume:              1400,
			SellVolume:             900,
			Delta:                  500,
			CVD:                    2200,
			BuyLargeTradeCount:     7,
			SellLargeTradeCount:    2,
			BuyLargeTradeNotional:  880000,
			SellLargeTradeNotional: 250000,
			LargeTradeDelta:        630000,
			AbsorptionBias:         "buy_absorption",
			AbsorptionStrength:     0.62,
			IcebergBias:            "buy_iceberg",
			IcebergStrength:        0.48,
			DataSource:             "agg_trade",
			MicrostructureEvents:   bullishMicrostructureEvents(),
		},
		models.Structure{
			Trend:      "uptrend",
			Support:    64600,
			Resistance: 66500,
			BOS:        true,
			Choch:      false,
		},
		models.Liquidity{
			BuyLiquidity:       64650,
			SellLiquidity:      66100,
			SweepType:          "sell_sweep",
			OrderBookImbalance: 0.18,
			DataSource:         "orderbook",
			EqualLow:           64820,
			StopClusters: []models.LiquidityCluster{
				{Kind: "buy_stop_cluster", Strength: 3.8},
			},
		},
	)

	if result.Action != "BUY" {
		t.Fatalf("expected BUY, got %s", result.Action)
	}
	if result.Score < buyThreshold {
		t.Fatalf("expected buy score >= %d, got %d", buyThreshold, result.Score)
	}
	if result.Confidence < 60 {
		t.Fatalf("expected confidence >= 60, got %d", result.Confidence)
	}
	if len(result.Factors) != 7 {
		t.Fatalf("expected 7 factors, got %d", len(result.Factors))
	}
	if result.RiskReward <= 1 {
		t.Fatalf("expected risk reward > 1, got %f", result.RiskReward)
	}
}

func TestGenerateOrderFlowFactorIncludesMicrostructureReason(t *testing.T) {
	engine := NewEngine()

	result := engine.Generate(
		"BTCUSDT",
		64800,
		models.Indicator{
			RSI:             56,
			MACD:            48,
			MACDSignal:      32,
			MACDHistogram:   16,
			EMA20:           64650,
			EMA50:           64280,
			ATR:             280,
			BollingerUpper:  65240,
			BollingerMiddle: 64720,
			BollingerLower:  64210,
			VWAP:            64590,
		},
		models.OrderFlow{
			BuyVolume:              920,
			SellVolume:             980,
			Delta:                  -60,
			CVD:                    120,
			BuyLargeTradeNotional:  420000,
			SellLargeTradeNotional: 380000,
			LargeTradeDelta:        40000,
			AbsorptionBias:         "buy_absorption",
			AbsorptionStrength:     0.71,
			IcebergBias:            "buy_iceberg",
			IcebergStrength:        0.66,
			DataSource:             "agg_trade",
			MicrostructureEvents: []models.OrderFlowMicrostructureEvent{
				{
					Type:      "absorption",
					Bias:      "bullish",
					Score:     5,
					Strength:  0.71,
					Price:     64795,
					TradeTime: 1741300000000,
					Detail:    "卖压被持续吸收，价格未继续下破",
				},
				{
					Type:      "initiative_shift",
					Bias:      "bullish",
					Score:     4,
					Strength:  0.32,
					Price:     64800,
					TradeTime: 1741300060000,
					Detail:    "买方主动性较前半段明显增强",
				},
			},
		},
		models.Structure{
			Trend:      "uptrend",
			Support:    64580,
			Resistance: 65350,
			BOS:        true,
		},
		models.Liquidity{
			BuyLiquidity:       64620,
			SellLiquidity:      65210,
			SweepType:          "sell_sweep",
			OrderBookImbalance: 0.16,
			DataSource:         "orderbook",
			EqualLow:           64790,
			StopClusters: []models.LiquidityCluster{
				{Kind: "buy_stop_cluster", Strength: 3.6},
			},
		},
	)

	orderFlowFactor := findFactor(result.Factors, "orderflow")
	if orderFlowFactor == nil {
		t.Fatal("expected to find orderflow factor")
	}
	if orderFlowFactor.Score <= 0 {
		t.Fatalf("expected positive orderflow factor score, got %d", orderFlowFactor.Score)
	}
	if !containsText(orderFlowFactor.Reason, "吸收") {
		t.Fatalf("expected orderflow reason to mention absorption, got %s", orderFlowFactor.Reason)
	}

	microstructureFactor := findFactor(result.Factors, "microstructure")
	if microstructureFactor == nil {
		t.Fatal("expected to find microstructure factor")
	}
	if microstructureFactor.Score <= 0 {
		t.Fatalf("expected positive microstructure factor score, got %d", microstructureFactor.Score)
	}
	if !containsText(microstructureFactor.Reason, "主动性") && !containsText(microstructureFactor.Reason, "sweep") {
		t.Fatalf("expected microstructure reason to mention event sequence or sweep, got %s", microstructureFactor.Reason)
	}
}

func TestGenerateMicrostructureConfluenceAddsBonus(t *testing.T) {
	engine := NewEngine()

	base := engine.Generate(
		"BTCUSDT",
		64980,
		models.Indicator{
			RSI:             59,
			MACD:            84,
			MACDSignal:      58,
			MACDHistogram:   26,
			EMA20:           64860,
			EMA50:           64420,
			ATR:             360,
			BollingerUpper:  65520,
			BollingerMiddle: 64910,
			BollingerLower:  64280,
			VWAP:            64740,
		},
		models.OrderFlow{
			BuyVolume:              1200,
			SellVolume:             860,
			Delta:                  340,
			CVD:                    1720,
			BuyLargeTradeNotional:  780000,
			SellLargeTradeNotional: 240000,
			LargeTradeDelta:        540000,
			AbsorptionBias:         "buy_absorption",
			AbsorptionStrength:     0.63,
			IcebergBias:            "buy_iceberg",
			IcebergStrength:        0.41,
			DataSource:             "agg_trade",
			MicrostructureEvents: []models.OrderFlowMicrostructureEvent{
				{
					Type:      "large_trade_cluster",
					Bias:      "bullish",
					Score:     3,
					Strength:  0.74,
					Price:     64960,
					TradeTime: 1741300000000,
					Detail:    "连续卖方大单被市场吸收，承接强度提升",
				},
				{
					Type:      "initiative_shift",
					Bias:      "bullish",
					Score:     2,
					Strength:  0.31,
					Price:     64980,
					TradeTime: 1741300060000,
					Detail:    "买方主动性较前半段明显增强",
				},
			},
		},
		models.Structure{Trend: "uptrend", Support: 64720, Resistance: 65480, BOS: true},
		models.Liquidity{
			BuyLiquidity:       64780,
			SellLiquidity:      65360,
			SweepType:          "sell_sweep",
			OrderBookImbalance: 0.14,
			DataSource:         "orderbook",
		},
	)

	withConfluence := engine.Generate(
		"BTCUSDT",
		64980,
		models.Indicator{
			RSI:             59,
			MACD:            84,
			MACDSignal:      58,
			MACDHistogram:   26,
			EMA20:           64860,
			EMA50:           64420,
			ATR:             360,
			BollingerUpper:  65520,
			BollingerMiddle: 64910,
			BollingerLower:  64280,
			VWAP:            64740,
		},
		models.OrderFlow{
			BuyVolume:              1200,
			SellVolume:             860,
			Delta:                  340,
			CVD:                    1720,
			BuyLargeTradeNotional:  780000,
			SellLargeTradeNotional: 240000,
			LargeTradeDelta:        540000,
			AbsorptionBias:         "buy_absorption",
			AbsorptionStrength:     0.63,
			IcebergBias:            "buy_iceberg",
			IcebergStrength:        0.41,
			DataSource:             "agg_trade",
			MicrostructureEvents: []models.OrderFlowMicrostructureEvent{
				{
					Type:      "large_trade_cluster",
					Bias:      "bullish",
					Score:     3,
					Strength:  0.74,
					Price:     64960,
					TradeTime: 1741300000000,
					Detail:    "连续卖方大单被市场吸收，承接强度提升",
				},
				{
					Type:      "initiative_shift",
					Bias:      "bullish",
					Score:     2,
					Strength:  0.31,
					Price:     64980,
					TradeTime: 1741300060000,
					Detail:    "买方主动性较前半段明显增强",
				},
				{
					Type:      "microstructure_confluence",
					Bias:      "bullish",
					Score:     7,
					Strength:  0.82,
					Price:     64982,
					TradeTime: 1741300120000,
					Detail:    "高阶微结构共振：large_trade_cluster + initiative_shift",
				},
			},
		},
		models.Structure{Trend: "uptrend", Support: 64720, Resistance: 65480, BOS: true},
		models.Liquidity{
			BuyLiquidity:       64780,
			SellLiquidity:      65360,
			SweepType:          "sell_sweep",
			OrderBookImbalance: 0.14,
			DataSource:         "orderbook",
		},
	)

	baseFactor := findFactor(base.Factors, "microstructure")
	withConfluenceFactor := findFactor(withConfluence.Factors, "microstructure")
	if baseFactor == nil || withConfluenceFactor == nil {
		t.Fatal("expected microstructure factors to exist")
	}
	if withConfluenceFactor.Score <= baseFactor.Score {
		t.Fatalf("expected confluence factor to increase score: base=%d with=%d", baseFactor.Score, withConfluenceFactor.Score)
	}
	if !containsText(withConfluenceFactor.Reason, "共振") {
		t.Fatalf("expected confluence factor reason to mention 共振, got %s", withConfluenceFactor.Reason)
	}
}

func TestGenerateLiquidityLadderBreakoutAddsBonus(t *testing.T) {
	engine := NewEngine()

	base := engine.Generate(
		"BTCUSDT",
		64980,
		models.Indicator{
			RSI:             58,
			MACD:            18,
			MACDSignal:      12,
			MACDHistogram:   6,
			EMA20:           64820,
			EMA50:           64420,
			ATR:             360,
			BollingerUpper:  65520,
			BollingerMiddle: 64910,
			BollingerLower:  64280,
			VWAP:            64740,
		},
		models.OrderFlow{
			BuyVolume:              1200,
			SellVolume:             860,
			Delta:                  340,
			CVD:                    1720,
			BuyLargeTradeNotional:  780000,
			SellLargeTradeNotional: 240000,
			LargeTradeDelta:        540000,
			AbsorptionBias:         "buy_absorption",
			AbsorptionStrength:     0.63,
			IcebergBias:            "buy_iceberg",
			IcebergStrength:        0.41,
			DataSource:             "agg_trade",
			MicrostructureEvents: []models.OrderFlowMicrostructureEvent{
				{
					Type:      "order_book_migration_layered",
					Bias:      "bullish",
					Score:     2,
					Strength:  0.71,
					Price:     64940,
					TradeTime: 1741300000000,
					Detail:    "买方挂单墙连续多层上移",
				},
				{
					Type:      "initiative_shift",
					Bias:      "bullish",
					Score:     1,
					Strength:  0.46,
					Price:     64970,
					TradeTime: 1741300060000,
					Detail:    "买方主动性较前半段明显增强",
				},
			},
		},
		models.Structure{Trend: "uptrend", Support: 64720, Resistance: 65480, BOS: true},
		models.Liquidity{
			BuyLiquidity:       64780,
			SellLiquidity:      65360,
			OrderBookImbalance: 0.09,
			DataSource:         "orderbook",
		},
	)
	withComposite := engine.Generate(
		"BTCUSDT",
		64980,
		models.Indicator{
			RSI:             58,
			MACD:            18,
			MACDSignal:      12,
			MACDHistogram:   6,
			EMA20:           64820,
			EMA50:           64420,
			ATR:             360,
			BollingerUpper:  65520,
			BollingerMiddle: 64910,
			BollingerLower:  64280,
			VWAP:            64740,
		},
		models.OrderFlow{
			BuyVolume:              1200,
			SellVolume:             860,
			Delta:                  340,
			CVD:                    1720,
			BuyLargeTradeNotional:  780000,
			SellLargeTradeNotional: 240000,
			LargeTradeDelta:        540000,
			AbsorptionBias:         "buy_absorption",
			AbsorptionStrength:     0.63,
			IcebergBias:            "buy_iceberg",
			IcebergStrength:        0.41,
			DataSource:             "agg_trade",
			MicrostructureEvents: []models.OrderFlowMicrostructureEvent{
				{
					Type:      "order_book_migration_layered",
					Bias:      "bullish",
					Score:     2,
					Strength:  0.71,
					Price:     64940,
					TradeTime: 1741300000000,
					Detail:    "买方挂单墙连续多层上移",
				},
				{
					Type:      "initiative_shift",
					Bias:      "bullish",
					Score:     1,
					Strength:  0.46,
					Price:     64970,
					TradeTime: 1741300060000,
					Detail:    "买方主动性较前半段明显增强",
				},
				{
					Type:      "liquidity_ladder_breakout",
					Bias:      "bullish",
					Score:     2,
					Strength:  0.8,
					Price:     64982,
					TradeTime: 1741300120000,
					Detail:    "挂单墙迁移与主动买盘同向推进：order_book_migration_layered + initiative_shift",
				},
			},
		},
		models.Structure{Trend: "uptrend", Support: 64720, Resistance: 65480, BOS: true},
		models.Liquidity{
			BuyLiquidity:       64780,
			SellLiquidity:      65360,
			OrderBookImbalance: 0.09,
			DataSource:         "orderbook",
		},
	)

	baseFactor := findFactor(base.Factors, "microstructure")
	withCompositeFactor := findFactor(withComposite.Factors, "microstructure")
	if baseFactor == nil || withCompositeFactor == nil {
		t.Fatal("expected microstructure factors to exist")
	}
	if withCompositeFactor.Score <= baseFactor.Score {
		t.Fatalf("expected liquidity ladder breakout factor to increase score: base=%d with=%d", baseFactor.Score, withCompositeFactor.Score)
	}
	if !containsText(withCompositeFactor.Reason, "挂单墙迁移") {
		t.Fatalf("expected factor reason to mention 挂单墙迁移, got %s", withCompositeFactor.Reason)
	}
}

func TestScoreMicrostructureSequenceReloadAndContinuationPatternsAddBonus(t *testing.T) {
	baseScore, _ := scoreMicrostructureSequence([]models.OrderFlowMicrostructureEvent{
		{
			Type:      "iceberg",
			Bias:      "bullish",
			Score:     2,
			Strength:  0.34,
			Price:     64940,
			TradeTime: 1741300000000,
			Detail:    "同价带重复出现隐藏买单承接",
		},
		{
			Type:      "large_trade_cluster",
			Bias:      "bullish",
			Score:     2,
			Strength:  0.38,
			Price:     64966,
			TradeTime: 1741300060000,
			Detail:    "连续卖方大单被市场吸收，承接强度提升",
		},
	})
	withPatternsScore, withPatternReasons := scoreMicrostructureSequence([]models.OrderFlowMicrostructureEvent{
		{
			Type:      "iceberg",
			Bias:      "bullish",
			Score:     2,
			Strength:  0.34,
			Price:     64940,
			TradeTime: 1741300000000,
			Detail:    "同价带重复出现隐藏买单承接",
		},
		{
			Type:      "large_trade_cluster",
			Bias:      "bullish",
			Score:     2,
			Strength:  0.38,
			Price:     64966,
			TradeTime: 1741300060000,
			Detail:    "连续卖方大单被市场吸收，承接强度提升",
		},
		{
			Type:      "iceberg_reload",
			Bias:      "bullish",
			Score:     5,
			Strength:  0.77,
			Price:     64972,
			TradeTime: 1741300120000,
			Detail:    "隐藏买单反复补单并维持承接：iceberg + large_trade_cluster",
		},
		{
			Type:      "absorption_reload_continuation",
			Bias:      "bullish",
			Score:     6,
			Strength:  0.84,
			Price:     64986,
			TradeTime: 1741300180000,
			Detail:    "吸收与补单维持后出现买方延续推进：iceberg_reload + large_trade_cluster",
		},
	})

	if withPatternsScore <= baseScore {
		t.Fatalf("expected reload continuation patterns to increase sequence score: base=%d with=%d", baseScore, withPatternsScore)
	}
	if !containsText(strings.Join(withPatternReasons, "；"), "补单") {
		t.Fatalf("expected sequence reasons to mention 补单, got %#v", withPatternReasons)
	}
}

func TestScoreMicrostructureSequenceExhaustionMigrationReversalAddsBonus(t *testing.T) {
	baseScore, _ := scoreMicrostructureSequence([]models.OrderFlowMicrostructureEvent{
		{
			Type:      "failed_auction_low_reclaim",
			Bias:      "bullish",
			Score:     3,
			Strength:  0.41,
			Price:     64620,
			TradeTime: 1741300000000,
			Detail:    "下方失败拍卖形成强收回分型",
		},
		{
			Type:      "order_book_migration_layered",
			Bias:      "bullish",
			Score:     2,
			Strength:  0.35,
			Price:     64642,
			TradeTime: 1741300060000,
			Detail:    "买方挂单墙连续多层上移",
		},
	})
	withPatternScore, withPatternReasons := scoreMicrostructureSequence([]models.OrderFlowMicrostructureEvent{
		{
			Type:      "failed_auction_low_reclaim",
			Bias:      "bullish",
			Score:     3,
			Strength:  0.41,
			Price:     64620,
			TradeTime: 1741300000000,
			Detail:    "下方失败拍卖形成强收回分型",
		},
		{
			Type:      "order_book_migration_layered",
			Bias:      "bullish",
			Score:     2,
			Strength:  0.35,
			Price:     64642,
			TradeTime: 1741300060000,
			Detail:    "买方挂单墙连续多层上移",
		},
		{
			Type:      "initiative_exhaustion",
			Bias:      "bullish",
			Score:     5,
			Strength:  0.81,
			Price:     64655,
			TradeTime: 1741300120000,
			Detail:    "前序主动卖盘衰竭，随后出现买方收回：aggression_burst -> failed_auction_low_reclaim + absorption",
		},
		{
			Type:      "exhaustion_migration_reversal",
			Bias:      "bullish",
			Score:     7,
			Strength:  0.88,
			Price:     64678,
			TradeTime: 1741300240000,
			Detail:    "主动卖盘耗尽后挂单墙上移确认反转：initiative_exhaustion + order_book_migration_layered",
		},
	})

	if withPatternScore <= baseScore {
		t.Fatalf("expected exhaustion migration reversal to increase sequence score: base=%d with=%d", baseScore, withPatternScore)
	}
	if !containsText(strings.Join(withPatternReasons, "；"), "主动盘") && !containsText(strings.Join(withPatternReasons, "；"), "挂单墙") {
		t.Fatalf("expected sequence reasons to mention exhaustion or migration, got %#v", withPatternReasons)
	}
}

func TestGenerateReturnsSellSignalForBearishConfluence(t *testing.T) {
	engine := NewEngine()

	result := engine.Generate(
		"ETHUSDT",
		3200,
		models.Indicator{
			RSI:             38,
			MACD:            -24,
			MACDSignal:      -12,
			MACDHistogram:   -12,
			EMA20:           3190,
			EMA50:           3275,
			ATR:             48,
			BollingerUpper:  3330,
			BollingerMiddle: 3230,
			BollingerLower:  3130,
			VWAP:            3245,
		},
		models.OrderFlow{
			BuyVolume:              420,
			SellVolume:             860,
			Delta:                  -440,
			CVD:                    -1800,
			BuyLargeTradeCount:     1,
			SellLargeTradeCount:    6,
			BuyLargeTradeNotional:  120000,
			SellLargeTradeNotional: 640000,
			LargeTradeDelta:        -520000,
			AbsorptionBias:         "sell_absorption",
			AbsorptionStrength:     0.59,
			IcebergBias:            "sell_iceberg",
			IcebergStrength:        0.52,
			DataSource:             "agg_trade",
			MicrostructureEvents:   bearishMicrostructureEvents(),
		},
		models.Structure{
			Trend:      "downtrend",
			Support:    3060,
			Resistance: 3210,
			BOS:        true,
			Choch:      false,
		},
		models.Liquidity{
			BuyLiquidity:       3095,
			SellLiquidity:      3215,
			SweepType:          "buy_sweep",
			OrderBookImbalance: -0.17,
			DataSource:         "orderbook",
			EqualHigh:          3205,
			StopClusters: []models.LiquidityCluster{
				{Kind: "sell_stop_cluster", Strength: 3.5},
			},
		},
	)

	if result.Action != "SELL" {
		t.Fatalf("expected SELL, got %s", result.Action)
	}
	if result.Score > sellThreshold {
		t.Fatalf("expected sell score <= %d, got %d", sellThreshold, result.Score)
	}
	if result.TrendBias != "bearish" {
		t.Fatalf("expected bearish trend bias, got %s", result.TrendBias)
	}
}

func TestGenerateMicrostructureFactorTurnsBearishOnAlignedPressure(t *testing.T) {
	engine := NewEngine()

	result := engine.Generate(
		"ETHUSDT",
		3210,
		models.Indicator{
			RSI:             44,
			MACD:            -18,
			MACDSignal:      -12,
			MACDHistogram:   -6,
			EMA20:           3208,
			EMA50:           3240,
			ATR:             42,
			BollingerUpper:  3290,
			BollingerMiddle: 3218,
			BollingerLower:  3148,
			VWAP:            3232,
		},
		models.OrderFlow{
			BuyVolume:              720,
			SellVolume:             1010,
			Delta:                  -290,
			CVD:                    -640,
			BuyLargeTradeCount:     2,
			SellLargeTradeCount:    8,
			BuyLargeTradeNotional:  160000,
			SellLargeTradeNotional: 780000,
			LargeTradeDelta:        -620000,
			AbsorptionBias:         "sell_absorption",
			AbsorptionStrength:     0.66,
			IcebergBias:            "sell_iceberg",
			IcebergStrength:        0.71,
			DataSource:             "agg_trade",
			MicrostructureEvents:   bearishMicrostructureEvents(),
		},
		models.Structure{
			Trend:      "downtrend",
			Support:    3130,
			Resistance: 3225,
			BOS:        true,
		},
		models.Liquidity{
			BuyLiquidity:       3145,
			SellLiquidity:      3220,
			SweepType:          "buy_sweep",
			OrderBookImbalance: -0.19,
			DataSource:         "orderbook",
			EqualHigh:          3215,
			StopClusters: []models.LiquidityCluster{
				{Kind: "sell_stop_cluster", Strength: 4.2},
			},
		},
	)

	microstructureFactor := findFactor(result.Factors, "microstructure")
	if microstructureFactor == nil {
		t.Fatal("expected to find microstructure factor")
	}
	if microstructureFactor.Score >= 0 {
		t.Fatalf("expected negative microstructure factor score, got %d", microstructureFactor.Score)
	}
}

func bullishMicrostructureEvents() []models.OrderFlowMicrostructureEvent {
	return []models.OrderFlowMicrostructureEvent{
		{
			Type:      "absorption",
			Bias:      "bullish",
			Score:     5,
			Strength:  0.62,
			Price:     64920,
			TradeTime: 1741300000000,
			Detail:    "卖压被持续吸收，价格未继续下破",
		},
		{
			Type:      "large_trade_cluster",
			Bias:      "bullish",
			Score:     4,
			Strength:  0.72,
			Price:     64950,
			TradeTime: 1741300060000,
			Detail:    "连续卖方大单被市场吸收，承接强度提升",
		},
		{
			Type:      "initiative_shift",
			Bias:      "bullish",
			Score:     4,
			Strength:  0.28,
			Price:     65000,
			TradeTime: 1741300120000,
			Detail:    "买方主动性较前半段明显增强",
		},
		{
			Type:      "microstructure_confluence",
			Bias:      "bullish",
			Score:     7,
			Strength:  0.84,
			Price:     65020,
			TradeTime: 1741300180000,
			Detail:    "高阶微结构共振：large_trade_cluster + initiative_shift",
		},
	}
}

func bearishMicrostructureEvents() []models.OrderFlowMicrostructureEvent {
	return []models.OrderFlowMicrostructureEvent{
		{
			Type:      "absorption",
			Bias:      "bearish",
			Score:     -5,
			Strength:  0.64,
			Price:     3212,
			TradeTime: 1741300000000,
			Detail:    "买盘被持续吸收，价格未继续上破",
		},
		{
			Type:      "large_trade_cluster",
			Bias:      "bearish",
			Score:     -4,
			Strength:  0.77,
			Price:     3210,
			TradeTime: 1741300060000,
			Detail:    "连续买方大单被市场吸收，上方抛压增强",
		},
		{
			Type:      "initiative_shift",
			Bias:      "bearish",
			Score:     -4,
			Strength:  0.31,
			Price:     3208,
			TradeTime: 1741300120000,
			Detail:    "卖方主动性较前半段明显增强",
		},
		{
			Type:      "microstructure_confluence",
			Bias:      "bearish",
			Score:     -7,
			Strength:  0.79,
			Price:     3206,
			TradeTime: 1741300180000,
			Detail:    "高阶微结构共振：large_trade_cluster + initiative_shift",
		},
	}
}

func TestGenerateReturnsNeutralWhenSignalsConflict(t *testing.T) {
	engine := NewEngine()

	result := engine.Generate(
		"BTCUSDT",
		64000,
		models.Indicator{
			RSI:             51,
			MACD:            12,
			MACDSignal:      10,
			MACDHistogram:   2,
			EMA20:           63980,
			EMA50:           63920,
			ATR:             35,
			BollingerUpper:  64450,
			BollingerMiddle: 64020,
			BollingerLower:  63590,
			VWAP:            64010,
		},
		models.OrderFlow{
			BuyVolume:  900,
			SellVolume: 980,
			Delta:      -80,
			CVD:        -120,
		},
		models.Structure{
			Trend:      "range",
			Support:    63600,
			Resistance: 64400,
			BOS:        false,
			Choch:      true,
		},
		models.Liquidity{
			BuyLiquidity:  63850,
			SellLiquidity: 64250,
			SweepType:     "",
		},
	)

	if result.Action != "NEUTRAL" {
		t.Fatalf("expected NEUTRAL, got %s", result.Action)
	}
	if result.Score <= sellThreshold || result.Score >= buyThreshold {
		t.Fatalf("expected score in neutral range, got %d", result.Score)
	}
	if result.Confidence > 60 {
		t.Fatalf("expected neutral confidence <= 60, got %d", result.Confidence)
	}
}

func findFactor(factors []models.SignalFactor, key string) *models.SignalFactor {
	for index := range factors {
		if factors[index].Key == key {
			return &factors[index]
		}
	}
	return nil
}

func containsText(source, target string) bool {
	return strings.Contains(source, target)
}

func TestSignalEngineUsesConfigProviderThreshold(t *testing.T) {
	// ConfigProvider 返回 buy_threshold=60（高于默认值 35）
	provider := &stubConfigProvider{
		values: map[string]string{
			"buy_threshold":  "60",
			"sell_threshold": "-60",
		},
	}
	engine := NewEngineWithConfig(provider)

	// score≈51（高于默认阈值 35，但低于自定义阈值 60 → 应为 NEUTRAL）
	signal := engine.Generate("BTCUSDT", 50000,
		stubIndicatorWithScore(50),
		models.OrderFlow{}, models.Structure{}, models.Liquidity{},
	)
	if signal.Action != "NEUTRAL" {
		t.Errorf("expected NEUTRAL with threshold=60 and score≈50, got %s (score=%d)", signal.Action, signal.Score)
	}
}

type stubConfigProvider struct {
	values map[string]string
}

func (p *stubConfigProvider) GetInt(_, _, key string, defaultVal int) int {
	if v, ok := p.values[key]; ok {
		n, _ := strconv.Atoi(v)
		return n
	}
	return defaultVal
}

// stubIndicatorWithScore 返回一个固定 Indicator，忽略参数，配合 price=50000 和空的
// OrderFlow/Structure/Liquidity 产生约 51 分（trend+25, momentum+20, volatility+6）。
// price=50000, EMA20=49800>EMA50=49500→trend+25; RSI=62,MACD bullish→momentum+20; ATR/price=0.8%→volatility+6
func stubIndicatorWithScore(_ int) models.Indicator {
	return models.Indicator{
		IntervalType:    "1m",
		VWAP:            49700,
		EMA20:           49800,
		EMA50:           49500,
		RSI:             62,
		MACD:            100,
		MACDSignal:      80,
		MACDHistogram:   20,
		ATR:             400,
		BollingerUpper:  51000,
		BollingerMiddle: 49900,
		BollingerLower:  48800,
	}
}
