package orderflow

import (
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"
	"time"

	"alpha-pulse/backend/models"
	binancepkg "alpha-pulse/backend/pkg/binance"
)

const (
	historyLimit          = 60
	minimumRequired       = 20
	tradeHistoryLimit     = 250
	tradeMinimumRequired  = 60
	aggregationWindow     = 6
	largeTradeThreshold   = 100000.0
	maxLargeTradeEvents   = 8
	maxMicroEvents        = 14
	icebergBucketPct      = 0.00025
	icebergPriceDriftPct  = 0.0015
	aggressionWindow      = 20
	orderBookHistoryLimit = 6
	orderBookWallWindow   = 3
)

// Engine 负责订单流分析。
type Engine struct {
	historyLimit int
}

type wallState struct {
	bestBid     float64
	bestAsk     float64
	bidWall     float64
	askWall     float64
	bidNotional float64
	askNotional float64
	eventTime   int64
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

// OrderBookHistoryLimit 返回识别盘口挂单迁移时所需的盘口快照数量。
func (e *Engine) OrderBookHistoryLimit() int {
	return orderBookHistoryLimit
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

// AnalyzeOrderBookMicrostructure 基于盘口快照识别更高阶微结构事件。
func (e *Engine) AnalyzeOrderBookMicrostructure(_ string, snapshots []models.OrderBookSnapshot) ([]models.OrderFlowMicrostructureEvent, error) {
	if len(snapshots) < 3 {
		return nil, nil
	}

	sortedSnapshots := sortOrderBookSnapshotsAscending(snapshots)
	return detectOrderBookMigrationEvents(sortedSnapshots), nil
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
	if event, ok := detectContinuousAbsorption(trades); ok {
		events = append(events, event)
	}
	events = append(events, detectFailedAuctionEvents(trades)...)

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
	if event, ok := detectMicrostructureConfluence(events); ok {
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

func detectContinuousAbsorption(trades []models.AggTrade) (models.OrderFlowMicrostructureEvent, bool) {
	if len(trades) < 36 {
		return models.OrderFlowMicrostructureEvent{}, false
	}

	windowSize := maxInt(len(trades)/3, 12)
	recent := trades[maxInt(len(trades)-windowSize*3, 0):]
	if len(recent) < windowSize*2 {
		return models.OrderFlowMicrostructureEvent{}, false
	}

	bullishWindows := 0
	bearishWindows := 0
	bullishStrength := 0.0
	bearishStrength := 0.0
	for start := 0; start < len(recent); start += windowSize {
		end := minInt(start+windowSize, len(recent))
		window := recent[start:end]
		if len(window) < 8 {
			continue
		}

		buyVolume, sellVolume := sumTradeVolumes(window)
		bias, strength := detectAbsorption(window, buyVolume, sellVolume)
		switch bias {
		case "buy_absorption":
			bullishWindows++
			bullishStrength += strength
		case "sell_absorption":
			bearishWindows++
			bearishStrength += strength
		}
	}

	latestTrade := trades[len(trades)-1]
	switch {
	case bullishWindows >= 2 && bearishWindows == 0:
		avgStrength := bullishStrength / float64(bullishWindows)
		return models.OrderFlowMicrostructureEvent{
			Type:      "continuous_absorption",
			Bias:      "bullish",
			Score:     6,
			Strength:  roundFloat(clamp(avgStrength+float64(bullishWindows-2)*0.12, 0, 1), 6),
			Price:     roundFloat(latestTrade.Price, 8),
			TradeTime: latestTrade.TradeTime,
			Detail:    "最近多个成交窗口连续出现买方吸收，卖压反复被承接",
		}, true
	case bearishWindows >= 2 && bullishWindows == 0:
		avgStrength := bearishStrength / float64(bearishWindows)
		return models.OrderFlowMicrostructureEvent{
			Type:      "continuous_absorption",
			Bias:      "bearish",
			Score:     -6,
			Strength:  roundFloat(clamp(avgStrength+float64(bearishWindows-2)*0.12, 0, 1), 6),
			Price:     roundFloat(latestTrade.Price, 8),
			TradeTime: latestTrade.TradeTime,
			Detail:    "最近多个成交窗口连续出现卖方吸收，买盘多次被压制",
		}, true
	default:
		return models.OrderFlowMicrostructureEvent{}, false
	}
}

func detectFailedAuctionEvents(trades []models.AggTrade) []models.OrderFlowMicrostructureEvent {
	if len(trades) < 24 {
		return nil
	}

	recent := trades[maxInt(len(trades)-24, 0):]
	split := len(recent) * 2 / 3
	if split < 8 || split >= len(recent) {
		return nil
	}

	reference := recent[:split]
	probe := recent[split:]
	priorHigh, priorLow := priceBounds(reference)
	probeHigh, probeLow := priceBounds(probe)
	endPrice := recent[len(recent)-1].Price
	latestTrade := recent[len(recent)-1]
	probeBuyNotional, probeSellNotional := sumTradeNotionalBySide(probe)
	events := make([]models.OrderFlowMicrostructureEvent, 0, 2)

	upsideBreak := priorHigh > 0 && probeHigh > priorHigh*1.00035
	upsideRejected := upsideBreak && endPrice <= priorHigh*1.00005 && endPrice < probeHigh*0.99955
	if upsideRejected && probeBuyNotional > probeSellNotional*1.08 {
		rejectionStrength := clamp((probeHigh-endPrice)/math.Max(priorHigh, 1)*1200, 0, 1)
		events = append(events, models.OrderFlowMicrostructureEvent{
			Type:      "failed_auction",
			Bias:      "bearish",
			Score:     -5,
			Strength:  roundFloat(rejectionStrength, 6),
			Price:     roundFloat(probeHigh, 8),
			TradeTime: latestTrade.TradeTime,
			Detail:    "上方拍卖突破失败，价格重新回到前高下方",
		})
		reclaimDepth := (priorHigh - endPrice) / math.Max(priorHigh, 1)
		if reclaimDepth >= 0.00035 || probeSellNotional >= probeBuyNotional*0.72 {
			events = append(events, models.OrderFlowMicrostructureEvent{
				Type:      "failed_auction_high_reject",
				Bias:      "bearish",
				Score:     -6,
				Strength:  roundFloat(clamp(rejectionStrength+reclaimDepth*800, 0, 1), 6),
				Price:     roundFloat(probeHigh, 8),
				TradeTime: latestTrade.TradeTime,
				Detail:    "上方失败拍卖形成强回落分型，突破后迅速被压回区间内",
			})
		}
	}

	downsideBreak := priorLow > 0 && probeLow < priorLow*0.99965
	downsideRejected := downsideBreak && endPrice >= priorLow*0.99995 && endPrice > probeLow*1.00045
	if downsideRejected && probeSellNotional > probeBuyNotional*1.08 {
		rejectionStrength := clamp((endPrice-probeLow)/math.Max(priorLow, 1)*1200, 0, 1)
		events = append(events, models.OrderFlowMicrostructureEvent{
			Type:      "failed_auction",
			Bias:      "bullish",
			Score:     5,
			Strength:  roundFloat(rejectionStrength, 6),
			Price:     roundFloat(probeLow, 8),
			TradeTime: latestTrade.TradeTime,
			Detail:    "下方拍卖突破失败，价格重新回到前低上方",
		})
		reclaimDepth := (endPrice - priorLow) / math.Max(priorLow, 1)
		if reclaimDepth >= 0.00035 || probeBuyNotional >= probeSellNotional*0.72 {
			events = append(events, models.OrderFlowMicrostructureEvent{
				Type:      "failed_auction_low_reclaim",
				Bias:      "bullish",
				Score:     6,
				Strength:  roundFloat(clamp(rejectionStrength+reclaimDepth*800, 0, 1), 6),
				Price:     roundFloat(probeLow, 8),
				TradeTime: latestTrade.TradeTime,
				Detail:    "下方失败拍卖形成强收回分型，跌破后快速收回前低上方",
			})
		}
	}

	return events
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

func detectMicrostructureConfluence(events []models.OrderFlowMicrostructureEvent) (models.OrderFlowMicrostructureEvent, bool) {
	if len(events) < 2 {
		return models.OrderFlowMicrostructureEvent{}, false
	}

	recent := events
	if len(recent) > 6 {
		recent = recent[len(recent)-6:]
	}

	bullishScore := 0.0
	bearishScore := 0.0
	bullishTypes := make([]string, 0, len(recent))
	bearishTypes := make([]string, 0, len(recent))
	seenBullish := map[string]struct{}{}
	seenBearish := map[string]struct{}{}
	latestEvent := recent[len(recent)-1]

	for _, event := range recent {
		if !isHighOrderMicrostructureEvent(event.Type) {
			continue
		}
		weightedScore := math.Abs(float64(event.Score)) * (1 + event.Strength*0.45)
		switch event.Bias {
		case "bullish":
			bullishScore += weightedScore
			if _, exists := seenBullish[event.Type]; !exists {
				seenBullish[event.Type] = struct{}{}
				bullishTypes = append(bullishTypes, event.Type)
			}
		case "bearish":
			bearishScore += weightedScore
			if _, exists := seenBearish[event.Type]; !exists {
				seenBearish[event.Type] = struct{}{}
				bearishTypes = append(bearishTypes, event.Type)
			}
		}
	}

	switch {
	case len(bullishTypes) >= 2 && bullishScore >= 11:
		return models.OrderFlowMicrostructureEvent{
			Type:      "microstructure_confluence",
			Bias:      "bullish",
			Score:     7,
			Strength:  roundFloat(clamp(bullishScore/18, 0, 1), 6),
			Price:     roundFloat(latestEvent.Price, 8),
			TradeTime: latestEvent.TradeTime,
			Detail:    "高阶微结构共振：" + strings.Join(bullishTypes, " + "),
		}, true
	case len(bearishTypes) >= 2 && bearishScore >= 11:
		return models.OrderFlowMicrostructureEvent{
			Type:      "microstructure_confluence",
			Bias:      "bearish",
			Score:     -7,
			Strength:  roundFloat(clamp(bearishScore/18, 0, 1), 6),
			Price:     roundFloat(latestEvent.Price, 8),
			TradeTime: latestEvent.TradeTime,
			Detail:    "高阶微结构共振：" + strings.Join(bearishTypes, " + "),
		}, true
	default:
		return models.OrderFlowMicrostructureEvent{}, false
	}
}

// DeriveCompositeMicrostructureEvents 基于已合并的微结构事件流识别跨来源组合模式。
func DeriveCompositeMicrostructureEvents(
	events []models.OrderFlowMicrostructureEvent,
) []models.OrderFlowMicrostructureEvent {
	if len(events) < 2 {
		return nil
	}

	sorted := make([]models.OrderFlowMicrostructureEvent, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].TradeTime == sorted[j].TradeTime {
			if sorted[i].Score == sorted[j].Score {
				return sorted[i].Type < sorted[j].Type
			}
			return sorted[i].Score < sorted[j].Score
		}
		return sorted[i].TradeTime < sorted[j].TradeTime
	})

	recent := sorted
	if len(recent) > 8 {
		recent = recent[len(recent)-8:]
	}

	derived := make([]models.OrderFlowMicrostructureEvent, 0, 2)
	if event, ok := detectAuctionTrapReversal(recent); ok {
		derived = append(derived, event)
	}
	if event, ok := detectLiquidityLadderBreakout(recent); ok {
		derived = append(derived, event)
	}
	return derived
}

func detectAuctionTrapReversal(events []models.OrderFlowMicrostructureEvent) (models.OrderFlowMicrostructureEvent, bool) {
	bullishEvent, bullishOK := buildCompositeCandidate(
		events,
		"bullish",
		auctionTrapBaseTypes,
		auctionTrapConfirmTypes,
		"auction_trap_reversal",
		6,
		9.5,
		15,
		"失败拍卖后出现同向承接确认",
	)
	bearishEvent, bearishOK := buildCompositeCandidate(
		events,
		"bearish",
		auctionTrapBaseTypes,
		auctionTrapConfirmTypes,
		"auction_trap_reversal",
		-6,
		9.5,
		15,
		"失败拍卖后出现同向抛压确认",
	)
	return selectStrongerCompositeCandidate(bullishEvent, bullishOK, bearishEvent, bearishOK)
}

func detectLiquidityLadderBreakout(events []models.OrderFlowMicrostructureEvent) (models.OrderFlowMicrostructureEvent, bool) {
	bullishEvent, bullishOK := buildCompositeCandidate(
		events,
		"bullish",
		migrationBaseTypes,
		executionDriveTypes,
		"liquidity_ladder_breakout",
		6,
		10.5,
		16,
		"挂单墙迁移与主动买盘同向推进",
	)
	bearishEvent, bearishOK := buildCompositeCandidate(
		events,
		"bearish",
		migrationBaseTypes,
		executionDriveTypes,
		"liquidity_ladder_breakout",
		-6,
		10.5,
		16,
		"挂单墙迁移与主动卖盘同向推进",
	)
	return selectStrongerCompositeCandidate(bullishEvent, bullishOK, bearishEvent, bearishOK)
}

func buildCompositeCandidate(
	events []models.OrderFlowMicrostructureEvent,
	bias string,
	baseTypes, confirmTypes map[string]struct{},
	eventType string,
	score int,
	minScore, normalizeBy float64,
	prefix string,
) (models.OrderFlowMicrostructureEvent, bool) {
	baseMatches, baseScore := collectCompositeMatches(events, bias, baseTypes)
	confirmMatches, confirmScore := collectCompositeMatches(events, bias, confirmTypes)
	if len(baseMatches) == 0 || len(confirmMatches) == 0 {
		return models.OrderFlowMicrostructureEvent{}, false
	}

	totalScore := baseScore + confirmScore
	if totalScore < minScore {
		return models.OrderFlowMicrostructureEvent{}, false
	}

	combinedTypes := append([]string{}, baseMatches...)
	for _, match := range confirmMatches {
		if !containsString(combinedTypes, match) {
			combinedTypes = append(combinedTypes, match)
		}
	}

	latest := latestCompositeEvent(events, bias, combinedTypes)
	return models.OrderFlowMicrostructureEvent{
		Type:      eventType,
		Bias:      bias,
		Score:     score,
		Strength:  roundFloat(clamp(totalScore/normalizeBy, 0, 1), 6),
		Price:     roundFloat(latest.Price, 8),
		TradeTime: latest.TradeTime,
		Detail:    prefix + "：" + strings.Join(combinedTypes, " + "),
	}, true
}

func collectCompositeMatches(
	events []models.OrderFlowMicrostructureEvent,
	bias string,
	allowedTypes map[string]struct{},
) ([]string, float64) {
	matches := make([]string, 0, len(events))
	seen := make(map[string]struct{}, len(events))
	totalScore := 0.0
	for _, event := range events {
		if event.Bias != bias {
			continue
		}
		if _, ok := allowedTypes[event.Type]; !ok {
			continue
		}
		if _, exists := seen[event.Type]; exists {
			continue
		}
		seen[event.Type] = struct{}{}
		matches = append(matches, event.Type)
		totalScore += math.Abs(float64(event.Score)) * (1 + event.Strength*0.4)
	}
	return matches, totalScore
}

func latestCompositeEvent(
	events []models.OrderFlowMicrostructureEvent,
	bias string,
	types []string,
) models.OrderFlowMicrostructureEvent {
	latest := events[len(events)-1]
	for _, event := range events {
		if event.Bias != bias || !containsString(types, event.Type) {
			continue
		}
		if event.TradeTime >= latest.TradeTime {
			latest = event
		}
	}
	return latest
}

func selectStrongerCompositeCandidate(
	left models.OrderFlowMicrostructureEvent,
	leftOK bool,
	right models.OrderFlowMicrostructureEvent,
	rightOK bool,
) (models.OrderFlowMicrostructureEvent, bool) {
	switch {
	case leftOK && !rightOK:
		return left, true
	case !leftOK && rightOK:
		return right, true
	case !leftOK && !rightOK:
		return models.OrderFlowMicrostructureEvent{}, false
	case left.Strength > right.Strength:
		return left, true
	case right.Strength > left.Strength:
		return right, true
	case left.TradeTime >= right.TradeTime:
		return left, true
	default:
		return right, true
	}
}

func isHighOrderMicrostructureEvent(eventType string) bool {
	switch eventType {
	case "continuous_absorption",
		"auction_trap_reversal",
		"failed_auction",
		"failed_auction_high_reject",
		"failed_auction_low_reclaim",
		"initiative_shift",
		"liquidity_ladder_breakout",
		"large_trade_cluster",
		"order_book_migration",
		"order_book_migration_layered",
		"order_book_migration_accelerated":
		return true
	default:
		return false
	}
}

var auctionTrapBaseTypes = map[string]struct{}{
	"failed_auction":             {},
	"failed_auction_high_reject": {},
	"failed_auction_low_reclaim": {},
}

var auctionTrapConfirmTypes = map[string]struct{}{
	"absorption":            {},
	"continuous_absorption": {},
	"iceberg":               {},
	"large_trade_cluster":   {},
}

var migrationBaseTypes = map[string]struct{}{
	"order_book_migration":             {},
	"order_book_migration_layered":     {},
	"order_book_migration_accelerated": {},
}

var executionDriveTypes = map[string]struct{}{
	"aggression_burst":    {},
	"initiative_shift":    {},
	"large_trade_cluster": {},
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
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

func sumTradeVolumes(trades []models.AggTrade) (float64, float64) {
	buyVolume := 0.0
	sellVolume := 0.0
	for _, trade := range trades {
		quantity := math.Max(trade.Quantity, 0)
		if trade.IsBuyerMaker {
			sellVolume += quantity
		} else {
			buyVolume += quantity
		}
	}
	return buyVolume, sellVolume
}

func sumTradeNotionalBySide(trades []models.AggTrade) (float64, float64) {
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
	return buyNotional, sellNotional
}

func priceBounds(trades []models.AggTrade) (float64, float64) {
	if len(trades) == 0 {
		return 0, 0
	}

	high := trades[0].Price
	low := trades[0].Price
	for _, trade := range trades[1:] {
		if trade.Price > high {
			high = trade.Price
		}
		if trade.Price < low {
			low = trade.Price
		}
	}
	return high, low
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

func sortOrderBookSnapshotsAscending(snapshots []models.OrderBookSnapshot) []models.OrderBookSnapshot {
	sorted := make([]models.OrderBookSnapshot, len(snapshots))
	copy(sorted, snapshots)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].EventTime == sorted[j].EventTime {
			return sorted[i].LastUpdateID < sorted[j].LastUpdateID
		}
		return sorted[i].EventTime < sorted[j].EventTime
	})
	return sorted
}

func detectOrderBookMigrationEvents(snapshots []models.OrderBookSnapshot) []models.OrderFlowMicrostructureEvent {
	if len(snapshots) < 3 {
		return nil
	}

	states := make([]wallState, 0, len(snapshots))
	for _, snapshot := range snapshots {
		bids, asks, err := parseOrderBookSnapshot(snapshot)
		if err != nil || len(bids) == 0 || len(asks) == 0 {
			continue
		}

		bidWallPrice, bidWallNotional := strongestDepthWall(bids, orderBookWallWindow)
		askWallPrice, askWallNotional := strongestDepthWall(asks, orderBookWallWindow)
		if bidWallPrice <= 0 || askWallPrice <= 0 {
			continue
		}

		states = append(states, wallState{
			bestBid:     snapshot.BestBidPrice,
			bestAsk:     snapshot.BestAskPrice,
			bidWall:     bidWallPrice,
			askWall:     askWallPrice,
			bidNotional: bidWallNotional,
			askNotional: askWallNotional,
			eventTime:   snapshot.EventTime,
		})
	}
	if len(states) < 3 {
		return nil
	}

	first := states[0]
	latest := states[len(states)-1]
	avgBidNotional := 0.0
	avgAskNotional := 0.0
	for _, state := range states {
		avgBidNotional += state.bidNotional
		avgAskNotional += state.askNotional
	}
	avgBidNotional /= float64(len(states))
	avgAskNotional /= float64(len(states))

	bidShiftPct := (latest.bidWall - first.bidWall) / math.Max(first.bidWall, 1)
	askShiftPct := (latest.askWall - first.askWall) / math.Max(first.askWall, 1)
	bestBidShiftPct := (latest.bestBid - first.bestBid) / math.Max(first.bestBid, 1)
	bestAskShiftPct := (latest.bestAsk - first.bestAsk) / math.Max(first.bestAsk, 1)
	events := make([]models.OrderFlowMicrostructureEvent, 0, 3)

	switch {
	case bidShiftPct >= 0.00045 &&
		bestBidShiftPct >= 0.00025 &&
		latest.bidNotional >= avgBidNotional*0.9 &&
		askShiftPct >= -0.00015:
		baseStrength := clamp(bidShiftPct*1200+latest.bidNotional/math.Max(avgBidNotional, 1)-1, 0, 1)
		events = append(events, models.OrderFlowMicrostructureEvent{
			Type:      "order_book_migration",
			Bias:      "bullish",
			Score:     4,
			Strength:  roundFloat(baseStrength, 6),
			Price:     roundFloat(latest.bidWall, 8),
			TradeTime: latest.eventTime,
			Detail:    "买方挂单墙持续上移，盘口承接价格带被主动抬高",
		})
		layeredSteps := countLayeredBidShifts(states)
		if layeredSteps >= 2 {
			events = append(events, models.OrderFlowMicrostructureEvent{
				Type:      "order_book_migration_layered",
				Bias:      "bullish",
				Score:     5,
				Strength:  roundFloat(clamp(baseStrength+float64(layeredSteps)*0.12, 0, 1), 6),
				Price:     roundFloat(latest.bidWall, 8),
				TradeTime: latest.eventTime,
				Detail:    "买方挂单墙连续多层上移，分层承接同步抬高",
			})
		}
		if isBullishMigrationAccelerating(states, avgBidNotional) {
			events = append(events, models.OrderFlowMicrostructureEvent{
				Type:      "order_book_migration_accelerated",
				Bias:      "bullish",
				Score:     6,
				Strength:  roundFloat(clamp(baseStrength+0.2, 0, 1), 6),
				Price:     roundFloat(latest.bidWall, 8),
				TradeTime: latest.eventTime,
				Detail:    "买方挂单迁移出现加速，best bid 与承接墙同步快速上抬",
			})
		}
	case askShiftPct <= -0.00045 &&
		bestAskShiftPct <= -0.00025 &&
		latest.askNotional >= avgAskNotional*0.9 &&
		bidShiftPct <= 0.00015:
		baseStrength := clamp(math.Abs(askShiftPct)*1200+latest.askNotional/math.Max(avgAskNotional, 1)-1, 0, 1)
		events = append(events, models.OrderFlowMicrostructureEvent{
			Type:      "order_book_migration",
			Bias:      "bearish",
			Score:     -4,
			Strength:  roundFloat(baseStrength, 6),
			Price:     roundFloat(latest.askWall, 8),
			TradeTime: latest.eventTime,
			Detail:    "卖方挂单墙持续下移，上方压单主动向现价靠拢",
		})
		layeredSteps := countLayeredAskShifts(states)
		if layeredSteps >= 2 {
			events = append(events, models.OrderFlowMicrostructureEvent{
				Type:      "order_book_migration_layered",
				Bias:      "bearish",
				Score:     -5,
				Strength:  roundFloat(clamp(baseStrength+float64(layeredSteps)*0.12, 0, 1), 6),
				Price:     roundFloat(latest.askWall, 8),
				TradeTime: latest.eventTime,
				Detail:    "卖方挂单墙连续多层下移，分层压制同步靠近现价",
			})
		}
		if isBearishMigrationAccelerating(states, avgAskNotional) {
			events = append(events, models.OrderFlowMicrostructureEvent{
				Type:      "order_book_migration_accelerated",
				Bias:      "bearish",
				Score:     -6,
				Strength:  roundFloat(clamp(baseStrength+0.2, 0, 1), 6),
				Price:     roundFloat(latest.askWall, 8),
				TradeTime: latest.eventTime,
				Detail:    "卖方挂单迁移出现加速，best ask 与压单墙同步快速下压",
			})
		}
	default:
		return nil
	}

	return events
}

func countLayeredBidShifts(states []wallState) int {
	count := 0
	for index := 1; index < len(states); index++ {
		prev := states[index-1]
		current := states[index]
		if current.bidWall > prev.bidWall*1.00012 && current.bidNotional >= prev.bidNotional*0.88 {
			count++
		}
	}
	return count
}

func countLayeredAskShifts(states []wallState) int {
	count := 0
	for index := 1; index < len(states); index++ {
		prev := states[index-1]
		current := states[index]
		if current.askWall < prev.askWall*0.99988 && current.askNotional >= prev.askNotional*0.88 {
			count++
		}
	}
	return count
}

func isBullishMigrationAccelerating(states []wallState, avgBidNotional float64) bool {
	if len(states) < 3 {
		return false
	}
	prev := states[len(states)-2]
	latest := states[len(states)-1]
	return latest.bidWall > prev.bidWall*1.00025 &&
		latest.bestBid > prev.bestBid*1.00018 &&
		latest.bidNotional >= math.Max(avgBidNotional, 1)*1.02
}

func isBearishMigrationAccelerating(states []wallState, avgAskNotional float64) bool {
	if len(states) < 3 {
		return false
	}
	prev := states[len(states)-2]
	latest := states[len(states)-1]
	return latest.askWall < prev.askWall*0.99975 &&
		latest.bestAsk < prev.bestAsk*0.99982 &&
		latest.askNotional >= math.Max(avgAskNotional, 1)*1.02
}

func parseOrderBookSnapshot(snapshot models.OrderBookSnapshot) ([]binancepkg.OrderBookLevel, []binancepkg.OrderBookLevel, error) {
	var bids []binancepkg.OrderBookLevel
	if err := json.Unmarshal([]byte(snapshot.BidsJSON), &bids); err != nil {
		return nil, nil, err
	}

	var asks []binancepkg.OrderBookLevel
	if err := json.Unmarshal([]byte(snapshot.AsksJSON), &asks); err != nil {
		return nil, nil, err
	}

	return bids, asks, nil
}

func strongestDepthWall(levels []binancepkg.OrderBookLevel, window int) (float64, float64) {
	if len(levels) == 0 {
		return 0, 0
	}

	limit := minInt(window, len(levels))
	bestPrice := 0.0
	bestNotional := 0.0
	for index := 0; index < limit; index++ {
		notional := levels[index].Price * levels[index].Quantity
		if notional > bestNotional {
			bestPrice = levels[index].Price
			bestNotional = notional
		}
	}
	return bestPrice, bestNotional
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
