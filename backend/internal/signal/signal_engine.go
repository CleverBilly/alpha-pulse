package signal

import (
	"fmt"
	"math"
	"strings"
	"time"

	"alpha-pulse/backend/models"
)

const (
	buyThreshold  = 35
	sellThreshold = -35
	maxAbsScore   = 100
)

// Engine 负责综合多模块结果生成交易信号。
type Engine struct{}

// NewEngine 创建信号引擎。
func NewEngine() *Engine {
	return &Engine{}
}

// Generate 根据多因子评分模型生成最终交易信号。
func (e *Engine) Generate(
	symbol string,
	price float64,
	indicator models.Indicator,
	orderFlow models.OrderFlow,
	structure models.Structure,
	liquidity models.Liquidity,
) models.Signal {
	factors := []models.SignalFactor{
		e.scoreTrend(price, indicator),
		e.scoreMomentum(price, indicator),
		e.scoreOrderFlow(orderFlow),
		e.scoreStructure(price, structure),
		e.scoreLiquidity(price, liquidity),
		e.scoreMicrostructure(price, orderFlow, liquidity),
	}

	directionalScore := sumFactorScores(factors)
	volatilityFactor := e.scoreVolatility(price, indicator, directionalScore)
	factors = append(factors, volatilityFactor)

	score := clampInt(sumFactorScores(factors), -maxAbsScore, maxAbsScore)
	action := resolveAction(score)
	stopLoss, targetPrice := buildRiskTargets(action, price, indicator.ATR, structure)
	riskReward := calculateRiskReward(price, stopLoss, targetPrice)
	confidence := calculateConfidence(score, factors, riskReward)

	return models.Signal{
		Symbol:      symbol,
		Action:      action,
		Score:       score,
		Confidence:  confidence,
		EntryPrice:  roundFloat(price, 8),
		StopLoss:    roundFloat(stopLoss, 8),
		TargetPrice: roundFloat(targetPrice, 8),
		CreatedAt:   time.Now(),
		Factors:     factors,
		RiskReward:  roundFloat(riskReward, 2),
		TrendBias:   deriveTrendBias(score),
	}
}

func (e *Engine) scoreMicrostructure(price float64, orderFlow models.OrderFlow, liquidity models.Liquidity) models.SignalFactor {
	score := 0
	reasons := make([]string, 0, 10)

	sequenceScore, sequenceReasons := scoreMicrostructureSequence(orderFlow.MicrostructureEvents)
	score += sequenceScore
	reasons = append(reasons, sequenceReasons...)

	switch {
	case orderFlow.AbsorptionBias == "buy_absorption" && liquidity.SweepType == "sell_sweep":
		score += 4
		reasons = append(reasons, "买方吸收与 sell-side sweep 共振，存在流动性回收后的反推")
	case orderFlow.AbsorptionBias == "sell_absorption" && liquidity.SweepType == "buy_sweep":
		score -= 4
		reasons = append(reasons, "卖方吸收与 buy-side sweep 共振，存在扫流动性后的回落")
	}

	switch {
	case orderFlow.IcebergBias == "buy_iceberg" && liquidity.OrderBookImbalance >= 0.10:
		score += 3
		reasons = append(reasons, fmt.Sprintf("隐藏买单与盘口失衡 %.2f 同向，承接更可信", liquidity.OrderBookImbalance))
	case orderFlow.IcebergBias == "sell_iceberg" && liquidity.OrderBookImbalance <= -0.10:
		score -= 3
		reasons = append(reasons, fmt.Sprintf("隐藏卖单与盘口失衡 %.2f 同向，压制更可信", liquidity.OrderBookImbalance))
	}

	if price > 0 {
		if liquidity.EqualLow > 0 &&
			math.Abs(price-liquidity.EqualLow)/price <= 0.004 &&
			(orderFlow.AbsorptionBias == "buy_absorption" || orderFlow.BuyLargeTradeCount > orderFlow.SellLargeTradeCount) {
			score += 2
			reasons = append(reasons, "价格贴近 equal low，且下方出现吸收/大单承接")
		}
		if liquidity.EqualHigh > 0 &&
			math.Abs(price-liquidity.EqualHigh)/price <= 0.004 &&
			(orderFlow.AbsorptionBias == "sell_absorption" || orderFlow.SellLargeTradeCount > orderFlow.BuyLargeTradeCount) {
			score -= 2
			reasons = append(reasons, "价格贴近 equal high，且上方出现吸收/大单压制")
		}
	}

	buyClusterStrength, sellClusterStrength := strongestClusterStrengths(liquidity.StopClusters)
	switch {
	case buyClusterStrength >= 3 && liquidity.OrderBookImbalance >= 0.08:
		score += 2
		reasons = append(reasons, fmt.Sprintf("下方止损/流动性簇强度 %.2f，且盘口买盘更厚", buyClusterStrength))
	case sellClusterStrength >= 3 && liquidity.OrderBookImbalance <= -0.08:
		score -= 2
		reasons = append(reasons, fmt.Sprintf("上方止损/流动性簇强度 %.2f，且盘口卖盘更厚", sellClusterStrength))
	}

	largeTradeTotal := orderFlow.BuyLargeTradeNotional + orderFlow.SellLargeTradeNotional
	if largeTradeTotal > 0 {
		largeTradeRatio := orderFlow.LargeTradeDelta / largeTradeTotal
		switch {
		case liquidity.SweepType == "sell_sweep" && largeTradeRatio >= 0.18:
			score += 2
			reasons = append(reasons, "sell-side sweep 后出现大单净流入，反转确认度提高")
		case liquidity.SweepType == "buy_sweep" && largeTradeRatio <= -0.18:
			score -= 2
			reasons = append(reasons, "buy-side sweep 后出现大单净流出，回落确认度提高")
		}
	}

	if orderFlow.DataSource == "agg_trade" && liquidity.DataSource == "orderbook" {
		reasons = append(reasons, "微结构因子来自真实成交与实时盘口联合验证")
	}

	return newFactor(
		"microstructure",
		"Microstructure",
		"microstructure",
		clampInt(score, -18, 18),
		fallbackReason(reasons, "微结构层面暂未形成额外优势"),
	)
}

func (e *Engine) scoreTrend(price float64, indicator models.Indicator) models.SignalFactor {
	score := 0
	reasons := make([]string, 0, 3)

	if price > indicator.VWAP {
		score += 6
		reasons = append(reasons, "价格位于 VWAP 上方")
	} else if price < indicator.VWAP {
		score -= 6
		reasons = append(reasons, "价格位于 VWAP 下方")
	}

	if indicator.EMA20 > indicator.EMA50 {
		score += 12
		reasons = append(reasons, "EMA20 高于 EMA50")
	} else if indicator.EMA20 < indicator.EMA50 {
		score -= 12
		reasons = append(reasons, "EMA20 低于 EMA50")
	}

	if price > indicator.EMA20 && indicator.EMA20 > indicator.EMA50 {
		score += 7
		reasons = append(reasons, "价格站稳短中期均线之上")
	} else if price < indicator.EMA20 && indicator.EMA20 < indicator.EMA50 {
		score -= 7
		reasons = append(reasons, "价格运行在短中期均线之下")
	}

	return newFactor(
		"trend",
		"Trend",
		"indicator",
		clampInt(score, -25, 25),
		fallbackReason(reasons, "均线与 VWAP 尚未形成明确方向"),
	)
}

func (e *Engine) scoreMomentum(price float64, indicator models.Indicator) models.SignalFactor {
	score := 0
	reasons := make([]string, 0, 4)

	switch {
	case indicator.RSI >= 55 && indicator.RSI <= 70:
		score += 6
		reasons = append(reasons, fmt.Sprintf("RSI %.2f 处于健康强势区间", indicator.RSI))
	case indicator.RSI > 70:
		score += 2
		reasons = append(reasons, fmt.Sprintf("RSI %.2f 偏热但仍保持强势", indicator.RSI))
	case indicator.RSI >= 45 && indicator.RSI < 55:
		score += 1
		reasons = append(reasons, fmt.Sprintf("RSI %.2f 中性偏多", indicator.RSI))
	case indicator.RSI >= 30 && indicator.RSI < 45:
		score -= 6
		reasons = append(reasons, fmt.Sprintf("RSI %.2f 落入弱势区", indicator.RSI))
	default:
		score -= 2
		reasons = append(reasons, fmt.Sprintf("RSI %.2f 处于超跌区，动量仍偏弱", indicator.RSI))
	}

	if indicator.MACD > indicator.MACDSignal {
		score += 8
		reasons = append(reasons, "MACD 位于信号线之上")
	} else if indicator.MACD < indicator.MACDSignal {
		score -= 8
		reasons = append(reasons, "MACD 位于信号线之下")
	}

	if indicator.MACDHistogram > 0 {
		score += 4
		reasons = append(reasons, "MACD Histogram 为正")
	} else if indicator.MACDHistogram < 0 {
		score -= 4
		reasons = append(reasons, "MACD Histogram 为负")
	}

	if price > indicator.BollingerUpper {
		score -= 3
		reasons = append(reasons, "价格高于布林上轨，短线存在回吐风险")
	} else if price < indicator.BollingerLower {
		score += 3
		reasons = append(reasons, "价格低于布林下轨，存在超跌反弹空间")
	} else if price > indicator.BollingerMiddle {
		score += 2
		reasons = append(reasons, "价格位于布林中轨上方")
	} else if price < indicator.BollingerMiddle {
		score -= 2
		reasons = append(reasons, "价格位于布林中轨下方")
	}

	return newFactor(
		"momentum",
		"Momentum",
		"indicator",
		clampInt(score, -25, 25),
		fallbackReason(reasons, "动量指标未给出清晰倾向"),
	)
}

func (e *Engine) scoreOrderFlow(orderFlow models.OrderFlow) models.SignalFactor {
	totalVolume := orderFlow.BuyVolume + orderFlow.SellVolume
	if totalVolume <= 0 {
		return newFactor("orderflow", "Order Flow", "flow", 0, "主动买卖量不足，暂不计入订单流方向")
	}

	score := 0
	reasons := make([]string, 0, 7)
	deltaRatio := orderFlow.Delta / totalVolume

	switch {
	case deltaRatio >= 0.12:
		score += 10
		reasons = append(reasons, fmt.Sprintf("Delta 比例 %.2f%%，主动买盘显著占优", deltaRatio*100))
	case deltaRatio >= 0.05:
		score += 6
		reasons = append(reasons, fmt.Sprintf("Delta 比例 %.2f%%，主动买盘占优", deltaRatio*100))
	case deltaRatio >= 0.02:
		score += 3
		reasons = append(reasons, fmt.Sprintf("Delta 比例 %.2f%%，买盘略强", deltaRatio*100))
	case deltaRatio <= -0.12:
		score -= 10
		reasons = append(reasons, fmt.Sprintf("Delta 比例 %.2f%%，主动卖盘显著占优", deltaRatio*100))
	case deltaRatio <= -0.05:
		score -= 6
		reasons = append(reasons, fmt.Sprintf("Delta 比例 %.2f%%，主动卖盘占优", deltaRatio*100))
	case deltaRatio <= -0.02:
		score -= 3
		reasons = append(reasons, fmt.Sprintf("Delta 比例 %.2f%%，卖盘略强", deltaRatio*100))
	default:
		reasons = append(reasons, "Delta 接近平衡")
	}

	if orderFlow.CVD > 0 {
		score += 4
		reasons = append(reasons, "CVD 为正，累计买盘延续")
	} else if orderFlow.CVD < 0 {
		score -= 4
		reasons = append(reasons, "CVD 为负，累计卖盘延续")
	}

	if orderFlow.BuyVolume > orderFlow.SellVolume*1.15 {
		score += 2
		reasons = append(reasons, "买方成交量明显大于卖方")
	} else if orderFlow.SellVolume > orderFlow.BuyVolume*1.15 {
		score -= 2
		reasons = append(reasons, "卖方成交量明显大于买方")
	}

	largeTradeTotal := orderFlow.BuyLargeTradeNotional + orderFlow.SellLargeTradeNotional
	if largeTradeTotal > 0 {
		largeTradeRatio := orderFlow.LargeTradeDelta / largeTradeTotal
		switch {
		case largeTradeRatio >= 0.20:
			score += 5
			reasons = append(reasons, fmt.Sprintf("大单净流入 %.2f%%，机构买盘偏强", largeTradeRatio*100))
		case largeTradeRatio >= 0.08:
			score += 3
			reasons = append(reasons, fmt.Sprintf("大单净流入 %.2f%%，买方大单占优", largeTradeRatio*100))
		case largeTradeRatio <= -0.20:
			score -= 5
			reasons = append(reasons, fmt.Sprintf("大单净流出 %.2f%%，机构卖盘偏强", largeTradeRatio*100))
		case largeTradeRatio <= -0.08:
			score -= 3
			reasons = append(reasons, fmt.Sprintf("大单净流出 %.2f%%，卖方大单占优", largeTradeRatio*100))
		}
	}

	switch orderFlow.AbsorptionBias {
	case "buy_absorption":
		score += 4
		reasons = append(reasons, fmt.Sprintf("出现买方吸收，强度 %.2f", orderFlow.AbsorptionStrength))
	case "sell_absorption":
		score -= 4
		reasons = append(reasons, fmt.Sprintf("出现卖方吸收，强度 %.2f", orderFlow.AbsorptionStrength))
	}

	switch orderFlow.IcebergBias {
	case "buy_iceberg":
		score += 3
		reasons = append(reasons, fmt.Sprintf("疑似隐藏买单承接，强度 %.2f", orderFlow.IcebergStrength))
	case "sell_iceberg":
		score -= 3
		reasons = append(reasons, fmt.Sprintf("疑似隐藏卖单压制，强度 %.2f", orderFlow.IcebergStrength))
	}

	if orderFlow.DataSource == "agg_trade" {
		reasons = append(reasons, "订单流结果来自真实 aggTrade 成交流")
	}

	return newFactor(
		"orderflow",
		"Order Flow",
		"flow",
		clampInt(score, -26, 26),
		fallbackReason(reasons, "订单流未形成优势方向"),
	)
}

func (e *Engine) scoreStructure(price float64, structure models.Structure) models.SignalFactor {
	score := 0
	reasons := make([]string, 0, 4)

	switch structure.Trend {
	case "uptrend":
		score += 10
		reasons = append(reasons, "市场结构处于上升趋势")
	case "downtrend":
		score -= 10
		reasons = append(reasons, "市场结构处于下降趋势")
	default:
		reasons = append(reasons, "市场结构处于震荡区间")
	}

	if structure.BOS {
		switch structure.Trend {
		case "uptrend":
			score += 6
			reasons = append(reasons, "发生 BOS，趋势延续概率提升")
		case "downtrend":
			score -= 6
			reasons = append(reasons, "发生 BOS，空头结构延续")
		}
	}

	if structure.Choch {
		if score > 0 {
			score -= 5
		} else if score < 0 {
			score += 5
		}
		reasons = append(reasons, "出现 CHOCH，结构连续性下降")
	}

	if price > 0 {
		if structure.Trend == "uptrend" && structure.Support > 0 && math.Abs(price-structure.Support)/price <= 0.006 {
			score += 4
			reasons = append(reasons, "价格靠近支撑位，上涨结构具备承接")
		}
		if structure.Trend == "downtrend" && structure.Resistance > 0 && math.Abs(price-structure.Resistance)/price <= 0.006 {
			score -= 4
			reasons = append(reasons, "价格靠近阻力位，下跌结构具备压制")
		}
	}

	return newFactor(
		"structure",
		"Structure",
		"structure",
		clampInt(score, -20, 20),
		fallbackReason(reasons, "结构因子暂无额外增量"),
	)
}

func (e *Engine) scoreLiquidity(price float64, liquidity models.Liquidity) models.SignalFactor {
	score := 0
	reasons := make([]string, 0, 4)

	switch liquidity.SweepType {
	case "sell_sweep":
		score += 8
		reasons = append(reasons, "出现 sell-side liquidity sweep，偏向多头反推")
	case "buy_sweep":
		score -= 8
		reasons = append(reasons, "出现 buy-side liquidity sweep，偏向空头回落")
	default:
		reasons = append(reasons, "暂未识别明确流动性扫单方向")
	}

	if price > 0 && liquidity.BuyLiquidity > 0 && math.Abs(price-liquidity.BuyLiquidity)/price <= 0.004 {
		score += 2
		reasons = append(reasons, "价格贴近买方流动性池")
	}
	if price > 0 && liquidity.SellLiquidity > 0 && math.Abs(price-liquidity.SellLiquidity)/price <= 0.004 {
		score -= 2
		reasons = append(reasons, "价格贴近卖方流动性池")
	}

	switch {
	case liquidity.OrderBookImbalance >= 0.10:
		score += 3
		reasons = append(reasons, fmt.Sprintf("盘口失衡 %.2f，买盘挂单更厚", liquidity.OrderBookImbalance))
	case liquidity.OrderBookImbalance <= -0.10:
		score -= 3
		reasons = append(reasons, fmt.Sprintf("盘口失衡 %.2f，卖盘挂单更厚", liquidity.OrderBookImbalance))
	}

	if liquidity.DataSource == "orderbook" {
		reasons = append(reasons, "流动性结果来自实时盘口快照")
	}

	return newFactor(
		"liquidity",
		"Liquidity",
		"liquidity",
		clampInt(score, -15, 15),
		fallbackReason(reasons, "流动性因子中性"),
	)
}

func (e *Engine) scoreVolatility(price float64, indicator models.Indicator, directionalScore int) models.SignalFactor {
	if price <= 0 || directionalScore == 0 {
		return newFactor(
			"volatility",
			"Volatility",
			"risk",
			0,
			"方向因子尚未形成一致性，波动率暂不参与加权",
		)
	}

	atrPct := indicator.ATR / price * 100
	quality := 0
	reason := ""
	switch {
	case atrPct >= 0.25 && atrPct <= 1.80:
		quality = 6
		reason = fmt.Sprintf("ATR 占比 %.2f%%，波动率适中，利于趋势延续", atrPct)
	case atrPct > 1.80 && atrPct <= 3.00:
		quality = 2
		reason = fmt.Sprintf("ATR 占比 %.2f%%，波动放大但仍可控", atrPct)
	case atrPct < 0.10:
		quality = -4
		reason = fmt.Sprintf("ATR 占比 %.2f%%，波动过低，突破延续性不足", atrPct)
	default:
		quality = -6
		reason = fmt.Sprintf("ATR 占比 %.2f%%，波动过大，方向信号需打折", atrPct)
	}

	score := signInt(directionalScore) * quality
	if score == 0 {
		return newFactor("volatility", "Volatility", "risk", 0, reason)
	}

	return newFactor("volatility", "Volatility", "risk", score, reason)
}

func newFactor(key, name, section string, score int, reason string) models.SignalFactor {
	return models.SignalFactor{
		Key:     key,
		Name:    name,
		Score:   score,
		Bias:    scoreBias(score),
		Reason:  reason,
		Section: section,
	}
}

func resolveAction(score int) string {
	switch {
	case score >= buyThreshold:
		return "BUY"
	case score <= sellThreshold:
		return "SELL"
	default:
		return "NEUTRAL"
	}
}

func deriveTrendBias(score int) string {
	switch {
	case score >= 15:
		return "bullish"
	case score <= -15:
		return "bearish"
	default:
		return "neutral"
	}
}

func calculateConfidence(score int, factors []models.SignalFactor, riskReward float64) int {
	direction := signInt(score)
	if direction == 0 {
		raw := 42 - minInt(countOpposingBuckets(factors), 12)
		return clampInt(raw, 25, 60)
	}

	alignedCount := 0
	opposingCount := 0
	neutralCount := 0
	for _, factor := range factors {
		switch signInt(factor.Score) {
		case direction:
			alignedCount++
		case 0:
			neutralCount++
		default:
			opposingCount++
		}
	}

	base := 46 + minInt(absInt(score)/2, 28)
	alignmentBonus := minInt(alignedCount*4, 20)
	oppositionPenalty := minInt(opposingCount*5, 15)
	neutralPenalty := minInt(neutralCount*2, 8)
	riskRewardBonus := 0
	switch {
	case riskReward >= 2.0:
		riskRewardBonus = 6
	case riskReward >= 1.5:
		riskRewardBonus = 3
	}

	return clampInt(base+alignmentBonus-oppositionPenalty-neutralPenalty+riskRewardBonus, 5, 95)
}

func buildRiskTargets(action string, price, atr float64, structure models.Structure) (float64, float64) {
	if price <= 0 {
		return 0, 0
	}

	if atr <= 0 {
		atr = price * 0.01
	}

	switch action {
	case "BUY":
		stopLoss := price - (1.5 * atr)
		if structure.Support > 0 && structure.Support < price {
			stopLoss = minFloat(stopLoss, structure.Support*0.998)
		}

		targetPrice := price + (2.4 * atr)
		if structure.Resistance > price {
			targetPrice = maxFloat(targetPrice, structure.Resistance)
		}

		return maxFloat(stopLoss, price*0.80), targetPrice
	case "SELL":
		stopLoss := price + (1.5 * atr)
		if structure.Resistance > price {
			stopLoss = maxFloat(stopLoss, structure.Resistance*1.002)
		}

		targetPrice := price - (2.4 * atr)
		if structure.Support > 0 && structure.Support < price {
			targetPrice = minFloat(targetPrice, structure.Support)
		}

		return stopLoss, maxFloat(targetPrice, price*0.20)
	default:
		return price - atr, price + atr
	}
}

func calculateRiskReward(entryPrice, stopLoss, targetPrice float64) float64 {
	risk := math.Abs(entryPrice - stopLoss)
	reward := math.Abs(targetPrice - entryPrice)
	if risk == 0 {
		return 0
	}
	return reward / risk
}

func fallbackReason(reasons []string, fallback string) string {
	if len(reasons) == 0 {
		return fallback
	}
	return strings.Join(reasons, "；")
}

func scoreMicrostructureSequence(events []models.OrderFlowMicrostructureEvent) (int, []string) {
	if len(events) == 0 {
		return 0, nil
	}

	recent := events
	if len(recent) > 4 {
		recent = recent[len(recent)-4:]
	}

	weights := []float64{0.65, 0.82, 1.0, 1.12}
	startWeight := len(weights) - len(recent)
	rawScore := 0.0
	bullishCount := 0
	bearishCount := 0
	reasons := make([]string, 0, len(recent)+1)

	for index, event := range recent {
		weight := weights[startWeight+index]
		rawScore += float64(event.Score) * weight
		switch signInt(event.Score) {
		case 1:
			bullishCount++
		case -1:
			bearishCount++
		}
		if len(reasons) < 3 && event.Detail != "" {
			reasons = append(reasons, event.Detail)
		}
		if event.Type == "microstructure_confluence" {
			rawScore += float64(signInt(event.Score)) * 2
			if len(reasons) < 3 {
				reasons = append(reasons, "高阶微结构事件出现同向共振")
			}
		}
		if event.Type == "auction_trap_reversal" {
			rawScore += float64(signInt(event.Score)) * 1.5
			if len(reasons) < 3 {
				reasons = append(reasons, "失败拍卖后的反转确认出现")
			}
		}
		if event.Type == "liquidity_ladder_breakout" {
			rawScore += float64(signInt(event.Score)) * 1.5
			if len(reasons) < 3 {
				reasons = append(reasons, "挂单墙迁移与主动成交同向推进")
			}
		}
	}

	switch {
	case bullishCount >= 2 && bearishCount == 0:
		rawScore += 2
		reasons = append(reasons, "最近微结构事件连续偏多")
	case bearishCount >= 2 && bullishCount == 0:
		rawScore -= 2
		reasons = append(reasons, "最近微结构事件连续偏空")
	}

	return clampInt(int(math.Round(rawScore)), -10, 10), reasons
}

func countOpposingBuckets(factors []models.SignalFactor) int {
	bullish := 0
	bearish := 0
	for _, factor := range factors {
		switch signInt(factor.Score) {
		case 1:
			bullish++
		case -1:
			bearish++
		}
	}

	if bullish == 0 || bearish == 0 {
		return 0
	}
	return minInt(bullish, bearish)
}

func sumFactorScores(factors []models.SignalFactor) int {
	total := 0
	for _, factor := range factors {
		total += factor.Score
	}
	return total
}

func scoreBias(score int) string {
	switch {
	case score > 0:
		return "bullish"
	case score < 0:
		return "bearish"
	default:
		return "neutral"
	}
}

func signInt(value int) int {
	switch {
	case value > 0:
		return 1
	case value < 0:
		return -1
	default:
		return 0
	}
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
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

func roundFloat(value float64, precision int) float64 {
	if precision < 0 {
		return value
	}

	pow := math.Pow10(precision)
	return math.Round(value*pow) / pow
}

func strongestClusterStrengths(clusters []models.LiquidityCluster) (float64, float64) {
	buyStrength := 0.0
	sellStrength := 0.0
	for _, cluster := range clusters {
		switch {
		case strings.Contains(cluster.Kind, "buy") && cluster.Strength > buyStrength:
			buyStrength = cluster.Strength
		case strings.Contains(cluster.Kind, "sell") && cluster.Strength > sellStrength:
			sellStrength = cluster.Strength
		}
	}
	return buyStrength, sellStrength
}
