package orderflow

import (
	"errors"
	"math"
	"sort"
	"time"

	"alpha-pulse/backend/models"
)

const (
	historyLimit         = 60
	minimumRequired      = 20
	tradeHistoryLimit    = 250
	tradeMinimumRequired = 60
	aggregationWindow    = 6
	largeTradeThreshold  = 100000.0
	maxLargeTradeEvents  = 8
	maxMicroEvents       = 8
	icebergBucketPct     = 0.00025
	icebergPriceDriftPct = 0.0015
	aggressionWindow     = 20
)

// Engine 负责订单流分析。
type Engine struct {
	historyLimit int
}

// NewEngine 创建订单流引擎。
func NewEngine() *Engine {
	return &Engine{historyLimit: historyLimit}
}

// HistoryLimit 返回订单流分析建议使用的历史 K 线数量。
func (e *Engine) HistoryLimit() int {
	return e.historyLimit
}

// MinimumRequired 返回订单流分析所需的最小 K 线数量。
func (e *Engine) MinimumRequired() int {
	return minimumRequired
}

// TradeHistoryLimit 返回真实成交分析建议使用的历史成交数量。
func (e *Engine) TradeHistoryLimit() int {
	return tradeHistoryLimit
}

// TradeMinimumRequired 返回真实成交分析所需的最小样本数。
func (e *Engine) TradeMinimumRequired() int {
	return tradeMinimumRequired
}

// Analyze 基于历史 K 线估算主动买卖量、Delta 和 CVD。
func (e *Engine) Analyze(symbol string, klines []models.Kline) (models.OrderFlow, error) {
	if len(klines) < e.MinimumRequired() {
		return models.OrderFlow{}, errors.New("not enough klines to analyze order flow")
	}

	sortedKlines := sortKlinesAscending(klines)
	windowStart := maxInt(len(sortedKlines)-aggregationWindow, 0)

	buyVolume := 0.0
	sellVolume := 0.0
	cvd := 0.0

	for index, kline := range sortedKlines {
		buyRatio := estimateBuyRatio(kline)
		candleBuyVolume := kline.Volume * buyRatio
		candleSellVolume := math.Max(kline.Volume-candleBuyVolume, 0)
		candleDelta := candleBuyVolume - candleSellVolume

		cvd += candleDelta
		if index >= windowStart {
			buyVolume += candleBuyVolume
			sellVolume += candleSellVolume
		}
	}

	latest := sortedKlines[len(sortedKlines)-1]
	createdAt := latest.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	return models.OrderFlow{
		Symbol:               symbol,
		BuyVolume:            roundFloat(buyVolume, 8),
		SellVolume:           roundFloat(sellVolume, 8),
		Delta:                roundFloat(buyVolume-sellVolume, 8),
		CVD:                  roundFloat(cvd, 8),
		AbsorptionBias:       "none",
		AbsorptionStrength:   0,
		IcebergBias:          "none",
		IcebergStrength:      0,
		DataSource:           "kline",
		LargeTrades:          nil,
		MicrostructureEvents: nil,
		CreatedAt:            createdAt,
	}, nil
}

// AnalyzeAggTrades 基于真实聚合成交计算主动买卖量、Delta 和 CVD。
func (e *Engine) AnalyzeAggTrades(symbol string, trades []models.AggTrade) (models.OrderFlow, error) {
	if len(trades) < e.TradeMinimumRequired() {
		return models.OrderFlow{}, errors.New("not enough agg trades to analyze order flow")
	}

	sortedTrades := sortAggTradesAscending(trades)
	buyVolume := 0.0
	sellVolume := 0.0
	cvd := 0.0
	buyLargeTradeCount := 0
	sellLargeTradeCount := 0
	buyLargeTradeNotional := 0.0
	sellLargeTradeNotional := 0.0
	largeTrades := make([]models.OrderFlowLargeTrade, 0, maxLargeTradeEvents)

	for _, trade := range sortedTrades {
		quantity := math.Max(trade.Quantity, 0)
		notional := math.Max(trade.QuoteQuantity, 0)
		if trade.IsBuyerMaker {
			sellVolume += quantity
			cvd -= quantity
			if notional >= largeTradeThreshold {
				sellLargeTradeCount++
				sellLargeTradeNotional += notional
				largeTrades = appendLargeTradeEvent(largeTrades, trade, "sell")
			}
		} else {
			buyVolume += quantity
			cvd += quantity
			if notional >= largeTradeThreshold {
				buyLargeTradeCount++
				buyLargeTradeNotional += notional
				largeTrades = appendLargeTradeEvent(largeTrades, trade, "buy")
			}
		}
	}

	absorptionBias, absorptionStrength := detectAbsorption(sortedTrades, buyVolume, sellVolume)
	icebergBias, icebergStrength := detectIceberg(sortedTrades)
	microstructureEvents := buildMicrostructureEvents(
		sortedTrades,
		absorptionBias,
		absorptionStrength,
		icebergBias,
		icebergStrength,
		buyLargeTradeCount,
		sellLargeTradeCount,
		buyLargeTradeNotional,
		sellLargeTradeNotional,
	)

	latest := sortedTrades[len(sortedTrades)-1]
	createdAt := latest.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	return models.OrderFlow{
		Symbol:                 symbol,
		BuyVolume:              roundFloat(buyVolume, 8),
		SellVolume:             roundFloat(sellVolume, 8),
		Delta:                  roundFloat(buyVolume-sellVolume, 8),
		CVD:                    roundFloat(cvd, 8),
		BuyLargeTradeCount:     buyLargeTradeCount,
		SellLargeTradeCount:    sellLargeTradeCount,
		BuyLargeTradeNotional:  roundFloat(buyLargeTradeNotional, 8),
		SellLargeTradeNotional: roundFloat(sellLargeTradeNotional, 8),
		LargeTradeDelta:        roundFloat(buyLargeTradeNotional-sellLargeTradeNotional, 8),
		AbsorptionBias:         absorptionBias,
		AbsorptionStrength:     roundFloat(absorptionStrength, 6),
		IcebergBias:            icebergBias,
		IcebergStrength:        roundFloat(icebergStrength, 6),
		DataSource:             "agg_trade",
		LargeTrades:            largeTrades,
		MicrostructureEvents:   microstructureEvents,
		CreatedAt:              createdAt,
	}, nil
}

func appendLargeTradeEvent(events []models.OrderFlowLargeTrade, trade models.AggTrade, side string) []models.OrderFlowLargeTrade {
	event := models.OrderFlowLargeTrade{
		Side:      side,
		Price:     roundFloat(trade.Price, 8),
		Quantity:  roundFloat(trade.Quantity, 8),
		Notional:  roundFloat(trade.QuoteQuantity, 8),
		TradeTime: trade.TradeTime,
	}

	events = append(events, event)
	if len(events) <= maxLargeTradeEvents {
		return events
	}
	return events[len(events)-maxLargeTradeEvents:]
}

func detectAbsorption(trades []models.AggTrade, buyVolume, sellVolume float64) (string, float64) {
	if len(trades) < 2 || buyVolume <= 0 || sellVolume <= 0 {
		return "none", 0
	}

	startPrice := trades[0].Price
	endPrice := trades[len(trades)-1].Price
	if startPrice <= 0 || endPrice <= 0 {
		return "none", 0
	}

	priceMovePct := (endPrice - startPrice) / startPrice
	switch {
	case sellVolume > buyVolume*1.12 && priceMovePct >= -0.0006:
		pressure := sellVolume / buyVolume
		strength := clamp((pressure-1)*0.55+maxFloat(priceMovePct, 0)*80, 0, 1)
		return "buy_absorption", strength
	case buyVolume > sellVolume*1.12 && priceMovePct <= 0.0006:
		pressure := buyVolume / sellVolume
		strength := clamp((pressure-1)*0.55+maxFloat(-priceMovePct, 0)*80, 0, 1)
		return "sell_absorption", strength
	default:
		return "none", 0
	}
}

func detectIceberg(trades []models.AggTrade) (string, float64) {
	if len(trades) < 3 {
		return "none", 0
	}

	startPrice := trades[0].Price
	endPrice := trades[len(trades)-1].Price
	if startPrice <= 0 || endPrice <= 0 {
		return "none", 0
	}
	priceDriftPct := math.Abs(endPrice-startPrice) / startPrice
	if priceDriftPct > icebergPriceDriftPct {
		return "none", 0
	}

	bucketSize := math.Max(startPrice*icebergBucketPct, 0.5)
	buyBuckets := make(map[float64]bucketStats)
	sellBuckets := make(map[float64]bucketStats)

	for _, trade := range trades {
		if trade.QuoteQuantity < largeTradeThreshold {
			continue
		}
		bucketPrice := roundFloat(math.Round(trade.Price/bucketSize)*bucketSize, 4)
		if trade.IsBuyerMaker {
			stats := sellBuckets[bucketPrice]
			stats.count++
			stats.notional += trade.QuoteQuantity
			sellBuckets[bucketPrice] = stats
			continue
		}
		stats := buyBuckets[bucketPrice]
		stats.count++
		stats.notional += trade.QuoteQuantity
		buyBuckets[bucketPrice] = stats
	}

	bestBuyPrice, bestBuy := bestBucket(buyBuckets)
	bestSellPrice, bestSell := bestBucket(sellBuckets)

	switch {
	case bestSell.count >= 3 && endPrice >= bestSellPrice*(1-0.001):
		strength := clamp(float64(bestSell.count-2)*0.2+bestSell.notional/(largeTradeThreshold*6), 0, 1)
		return "buy_iceberg", strength
	case bestBuy.count >= 3 && endPrice <= bestBuyPrice*(1+0.001):
		strength := clamp(float64(bestBuy.count-2)*0.2+bestBuy.notional/(largeTradeThreshold*6), 0, 1)
		return "sell_iceberg", strength
	default:
		return "none", 0
	}
}

func buildMicrostructureEvents(
	trades []models.AggTrade,
	absorptionBias string,
	absorptionStrength float64,
	icebergBias string,
	icebergStrength float64,
	buyLargeTradeCount, sellLargeTradeCount int,
	buyLargeTradeNotional, sellLargeTradeNotional float64,
) []models.OrderFlowMicrostructureEvent {
	if len(trades) == 0 {
		return nil
	}

	events := make([]models.OrderFlowMicrostructureEvent, 0, maxMicroEvents)
	latestTrade := trades[len(trades)-1]

	if event, ok := buildAbsorptionEvent(latestTrade, absorptionBias, absorptionStrength); ok {
		events = append(events, event)
	}
	if event, ok := buildIcebergEvent(latestTrade, icebergBias, icebergStrength); ok {
		events = append(events, event)
	}

	events = append(events, detectAggressionBursts(trades)...)

	if event, ok := detectInitiativeShift(trades); ok {
		events = append(events, event)
	}
	if event, ok := detectLargeTradeCluster(
		trades,
		absorptionBias,
		buyLargeTradeCount,
		sellLargeTradeCount,
		buyLargeTradeNotional,
		sellLargeTradeNotional,
	); ok {
		events = append(events, event)
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].TradeTime == events[j].TradeTime {
			if events[i].Score == events[j].Score {
				return events[i].Type < events[j].Type
			}
			return events[i].Score < events[j].Score
		}
		return events[i].TradeTime < events[j].TradeTime
	})

	if len(events) <= maxMicroEvents {
		return events
	}
	return events[len(events)-maxMicroEvents:]
}

func buildAbsorptionEvent(
	latestTrade models.AggTrade,
	bias string,
	strength float64,
) (models.OrderFlowMicrostructureEvent, bool) {
	if strength <= 0 {
		return models.OrderFlowMicrostructureEvent{}, false
	}

	switch bias {
	case "buy_absorption":
		return models.OrderFlowMicrostructureEvent{
			Type:      "absorption",
			Bias:      "bullish",
			Score:     5,
			Strength:  roundFloat(strength, 6),
			Price:     roundFloat(latestTrade.Price, 8),
			TradeTime: latestTrade.TradeTime,
			Detail:    "卖压被持续吸收，价格未继续下破",
		}, true
	case "sell_absorption":
		return models.OrderFlowMicrostructureEvent{
			Type:      "absorption",
			Bias:      "bearish",
			Score:     -5,
			Strength:  roundFloat(strength, 6),
			Price:     roundFloat(latestTrade.Price, 8),
			TradeTime: latestTrade.TradeTime,
			Detail:    "买盘被持续吸收，价格未继续上破",
		}, true
	default:
		return models.OrderFlowMicrostructureEvent{}, false
	}
}

func buildIcebergEvent(
	latestTrade models.AggTrade,
	bias string,
	strength float64,
) (models.OrderFlowMicrostructureEvent, bool) {
	if strength <= 0 {
		return models.OrderFlowMicrostructureEvent{}, false
	}

	switch bias {
	case "buy_iceberg":
		return models.OrderFlowMicrostructureEvent{
			Type:      "iceberg",
			Bias:      "bullish",
			Score:     4,
			Strength:  roundFloat(strength, 6),
			Price:     roundFloat(latestTrade.Price, 8),
			TradeTime: latestTrade.TradeTime,
			Detail:    "同价带重复出现隐藏买单承接",
		}, true
	case "sell_iceberg":
		return models.OrderFlowMicrostructureEvent{
			Type:      "iceberg",
			Bias:      "bearish",
			Score:     -4,
			Strength:  roundFloat(strength, 6),
			Price:     roundFloat(latestTrade.Price, 8),
			TradeTime: latestTrade.TradeTime,
			Detail:    "同价带重复出现隐藏卖单压制",
		}, true
	default:
		return models.OrderFlowMicrostructureEvent{}, false
	}
}

func detectAggressionBursts(trades []models.AggTrade) []models.OrderFlowMicrostructureEvent {
	if len(trades) < aggressionWindow {
		return nil
	}

	events := make([]models.OrderFlowMicrostructureEvent, 0, 2)
	startIndexes := []int{maxInt(len(trades)-aggressionWindow*2, 0), maxInt(len(trades)-aggressionWindow, 0)}
	for _, start := range startIndexes {
		end := minInt(start+aggressionWindow, len(trades))
		window := trades[start:end]
		if len(window) < aggressionWindow/2 {
			continue
		}

		buyNotional := 0.0
		sellNotional := 0.0
		for _, trade := range window {
			notional := effectiveTradeNotional(trade)
			if trade.IsBuyerMaker {
				sellNotional += notional
			} else {
				buyNotional += notional
			}
		}

		total := buyNotional + sellNotional
		if total <= 0 {
			continue
		}

		deltaRatio := (buyNotional - sellNotional) / total
		latestTrade := window[len(window)-1]
		switch {
		case deltaRatio >= 0.22:
			events = append(events, models.OrderFlowMicrostructureEvent{
				Type:      "aggression_burst",
				Bias:      "bullish",
				Score:     3,
				Strength:  roundFloat(deltaRatio, 6),
				Price:     roundFloat(latestTrade.Price, 8),
				TradeTime: latestTrade.TradeTime,
				Detail:    "最近成交窗口出现主动买盘冲击",
			})
		case deltaRatio <= -0.22:
			events = append(events, models.OrderFlowMicrostructureEvent{
				Type:      "aggression_burst",
				Bias:      "bearish",
				Score:     -3,
				Strength:  roundFloat(math.Abs(deltaRatio), 6),
				Price:     roundFloat(latestTrade.Price, 8),
				TradeTime: latestTrade.TradeTime,
				Detail:    "最近成交窗口出现主动卖盘冲击",
			})
		}
	}

	return events
}

func detectInitiativeShift(trades []models.AggTrade) (models.OrderFlowMicrostructureEvent, bool) {
	if len(trades) < tradeMinimumRequired {
		return models.OrderFlowMicrostructureEvent{}, false
	}

	split := len(trades) / 2
	firstRatio := calculateNotionalDeltaRatio(trades[:split])
	secondRatio := calculateNotionalDeltaRatio(trades[split:])
	shift := secondRatio - firstRatio
	latestTrade := trades[len(trades)-1]

	switch {
	case shift >= 0.18:
		return models.OrderFlowMicrostructureEvent{
			Type:      "initiative_shift",
			Bias:      "bullish",
			Score:     4,
			Strength:  roundFloat(shift, 6),
			Price:     roundFloat(latestTrade.Price, 8),
			TradeTime: latestTrade.TradeTime,
			Detail:    "买方主动性较前半段明显增强",
		}, true
	case shift <= -0.18:
		return models.OrderFlowMicrostructureEvent{
			Type:      "initiative_shift",
			Bias:      "bearish",
			Score:     -4,
			Strength:  roundFloat(math.Abs(shift), 6),
			Price:     roundFloat(latestTrade.Price, 8),
			TradeTime: latestTrade.TradeTime,
			Detail:    "卖方主动性较前半段明显增强",
		}, true
	default:
		return models.OrderFlowMicrostructureEvent{}, false
	}
}

func detectLargeTradeCluster(
	trades []models.AggTrade,
	absorptionBias string,
	buyLargeTradeCount, sellLargeTradeCount int,
	buyLargeTradeNotional, sellLargeTradeNotional float64,
) (models.OrderFlowMicrostructureEvent, bool) {
	if len(trades) == 0 {
		return models.OrderFlowMicrostructureEvent{}, false
	}

	latestTrade := trades[len(trades)-1]
	lookbackStart := maxInt(len(trades)-aggressionWindow, 0)
	recentLargeBuy := 0
	recentLargeSell := 0
	for _, trade := range trades[lookbackStart:] {
		if effectiveTradeNotional(trade) < largeTradeThreshold {
			continue
		}
		if trade.IsBuyerMaker {
			recentLargeSell++
		} else {
			recentLargeBuy++
		}
	}

	totalNotional := buyLargeTradeNotional + sellLargeTradeNotional
	if totalNotional <= 0 {
		return models.OrderFlowMicrostructureEvent{}, false
	}

	switch {
	case sellLargeTradeCount >= 3 &&
		recentLargeSell >= 2 &&
		sellLargeTradeNotional >= buyLargeTradeNotional*1.35 &&
		absorptionBias == "buy_absorption":
		return models.OrderFlowMicrostructureEvent{
			Type:      "large_trade_cluster",
			Bias:      "bullish",
			Score:     4,
			Strength:  roundFloat(sellLargeTradeNotional/totalNotional, 6),
			Price:     roundFloat(latestTrade.Price, 8),
			TradeTime: latestTrade.TradeTime,
			Detail:    "连续卖方大单被市场吸收，承接强度提升",
		}, true
	case buyLargeTradeCount >= 3 &&
		recentLargeBuy >= 2 &&
		buyLargeTradeNotional >= sellLargeTradeNotional*1.35 &&
		absorptionBias == "sell_absorption":
		return models.OrderFlowMicrostructureEvent{
			Type:      "large_trade_cluster",
			Bias:      "bearish",
			Score:     -4,
			Strength:  roundFloat(buyLargeTradeNotional/totalNotional, 6),
			Price:     roundFloat(latestTrade.Price, 8),
			TradeTime: latestTrade.TradeTime,
			Detail:    "连续买方大单被市场吸收，上方抛压增强",
		}, true
	case buyLargeTradeCount >= 3 && recentLargeBuy >= 2 && buyLargeTradeNotional >= sellLargeTradeNotional*1.35:
		return models.OrderFlowMicrostructureEvent{
			Type:      "large_trade_cluster",
			Bias:      "bullish",
			Score:     4,
			Strength:  roundFloat(buyLargeTradeNotional/totalNotional, 6),
			Price:     roundFloat(latestTrade.Price, 8),
			TradeTime: latestTrade.TradeTime,
			Detail:    "最近出现连续买方大单簇，机构承接增强",
		}, true
	case sellLargeTradeCount >= 3 && recentLargeSell >= 2 && sellLargeTradeNotional >= buyLargeTradeNotional*1.35:
		return models.OrderFlowMicrostructureEvent{
			Type:      "large_trade_cluster",
			Bias:      "bearish",
			Score:     -4,
			Strength:  roundFloat(sellLargeTradeNotional/totalNotional, 6),
			Price:     roundFloat(latestTrade.Price, 8),
			TradeTime: latestTrade.TradeTime,
			Detail:    "最近出现连续卖方大单簇，机构抛压增强",
		}, true
	default:
		return models.OrderFlowMicrostructureEvent{}, false
	}
}

func calculateNotionalDeltaRatio(trades []models.AggTrade) float64 {
	buyNotional := 0.0
	sellNotional := 0.0
	for _, trade := range trades {
		notional := effectiveTradeNotional(trade)
		if trade.IsBuyerMaker {
			sellNotional += notional
		} else {
			buyNotional += notional
		}
	}

	total := buyNotional + sellNotional
	if total <= 0 {
		return 0
	}
	return (buyNotional - sellNotional) / total
}

func effectiveTradeNotional(trade models.AggTrade) float64 {
	if trade.QuoteQuantity > 0 {
		return trade.QuoteQuantity
	}
	return trade.Price * trade.Quantity
}

type bucketStats struct {
	count    int
	notional float64
}

func bestBucket(buckets map[float64]bucketStats) (float64, bucketStats) {
	bestPrice := 0.0
	best := bucketStats{}
	for price, stats := range buckets {
		if stats.count > best.count || (stats.count == best.count && stats.notional > best.notional) {
			bestPrice = price
			best = stats
		}
	}
	return bestPrice, best
}

func estimateBuyRatio(kline models.Kline) float64 {
	priceRange := math.Max(kline.HighPrice-kline.LowPrice, 0.00000001)
	closeLocation := clamp((kline.ClosePrice-kline.LowPrice)/priceRange, 0.0, 1.0)

	bodyBias := 0.5
	switch {
	case kline.ClosePrice > kline.OpenPrice:
		bodyBias = clamp(0.58+((kline.ClosePrice-kline.OpenPrice)/priceRange)*0.25, 0.5, 0.92)
	case kline.ClosePrice < kline.OpenPrice:
		bodyBias = clamp(0.42-((kline.OpenPrice-kline.ClosePrice)/priceRange)*0.25, 0.08, 0.5)
	}

	lowerWick := math.Max(math.Min(kline.OpenPrice, kline.ClosePrice)-kline.LowPrice, 0)
	upperWick := math.Max(kline.HighPrice-math.Max(kline.OpenPrice, kline.ClosePrice), 0)
	wickBias := clamp((lowerWick-upperWick)/priceRange*0.08, -0.08, 0.08)

	return clamp(closeLocation*0.6+bodyBias*0.4+wickBias, 0.05, 0.95)
}

func sortKlinesAscending(klines []models.Kline) []models.Kline {
	sorted := make([]models.Kline, len(klines))
	copy(sorted, klines)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].OpenTime == sorted[j].OpenTime {
			return sorted[i].ID < sorted[j].ID
		}
		return sorted[i].OpenTime < sorted[j].OpenTime
	})
	return sorted
}

func sortAggTradesAscending(trades []models.AggTrade) []models.AggTrade {
	sorted := make([]models.AggTrade, len(trades))
	copy(sorted, trades)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].TradeTime == sorted[j].TradeTime {
			return sorted[i].AggTradeID < sorted[j].AggTradeID
		}
		return sorted[i].TradeTime < sorted[j].TradeTime
	})
	return sorted
}

func clamp(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func roundFloat(value float64, precision int) float64 {
	pow := math.Pow10(precision)
	return math.Round(value*pow) / pow
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
