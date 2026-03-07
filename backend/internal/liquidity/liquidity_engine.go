package liquidity

import (
	"encoding/json"
	"errors"
	"math"
	"sort"
	"time"

	"alpha-pulse/backend/models"
	binancepkg "alpha-pulse/backend/pkg/binance"
)

const (
	historyLimit        = 80
	minimumRequired     = 25
	clusterWindow       = 24
	equalLevelLookback  = 36
	orderBookDepthLimit = 20
	orderBookCluster    = 4
	imbalanceLevels     = 8
	sweepTolerance      = 0.001
	depthSweepTolerance = 0.0008
	equalLevelTolerance = 0.0012
)

// Engine 负责流动性区域识别。
type Engine struct {
	historyLimit int
}

// NewEngine 创建流动性引擎。
func NewEngine() *Engine {
	return &Engine{historyLimit: historyLimit}
}

// HistoryLimit 返回流动性分析建议使用的历史 K 线数量。
func (e *Engine) HistoryLimit() int {
	return e.historyLimit
}

// MinimumRequired 返回流动性分析所需的最小 K 线数量。
func (e *Engine) MinimumRequired() int {
	return minimumRequired
}

// OrderBookDepthLimit 返回盘口分析建议使用的深度档位数量。
func (e *Engine) OrderBookDepthLimit() int {
	return orderBookDepthLimit
}

// Analyze 基于历史高低点和扫损回收行为识别流动性区域。
func (e *Engine) Analyze(symbol string, klines []models.Kline) (models.Liquidity, error) {
	if len(klines) < e.MinimumRequired() {
		return models.Liquidity{}, errors.New("not enough klines to analyze liquidity")
	}

	sortedKlines := sortKlinesAscending(klines)
	buyLiquidity, sellLiquidity, sweepType := analyzeLiquidityFromKlines(sortedKlines)
	equalHigh, equalHighTouches := detectEqualHigh(sortedKlines)
	equalLow, equalLowTouches := detectEqualLow(sortedKlines)
	stopClusters := buildStopClusters(
		buyLiquidity,
		sellLiquidity,
		equalHigh,
		equalLow,
		float64(equalHighTouches),
		float64(equalLowTouches),
		0,
		0,
	)
	latest := sortedKlines[len(sortedKlines)-1]
	createdAt := latest.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	return models.Liquidity{
		Symbol:             symbol,
		BuyLiquidity:       roundFloat(buyLiquidity, 8),
		SellLiquidity:      roundFloat(sellLiquidity, 8),
		SweepType:          sweepType,
		OrderBookImbalance: 0,
		DataSource:         "kline",
		EqualHigh:          roundFloat(equalHigh, 8),
		EqualLow:           roundFloat(equalLow, 8),
		StopClusters:       stopClusters,
		CreatedAt:          createdAt,
	}, nil
}

// AnalyzeWithOrderBook 基于历史 K 线和盘口快照识别真实流动性池与扫单行为。
func (e *Engine) AnalyzeWithOrderBook(symbol string, klines []models.Kline, snapshot models.OrderBookSnapshot) (models.Liquidity, error) {
	if len(klines) < e.MinimumRequired() {
		return models.Liquidity{}, errors.New("not enough klines to analyze liquidity")
	}
	if snapshot.Symbol == "" || (snapshot.BidsJSON == "" && snapshot.AsksJSON == "") {
		return models.Liquidity{}, errors.New("order book snapshot is empty")
	}

	sortedKlines := sortKlinesAscending(klines)
	baseBuyLiquidity, baseSellLiquidity, baseSweepType := analyzeLiquidityFromKlines(sortedKlines)
	equalHigh, equalHighTouches := detectEqualHigh(sortedKlines)
	equalLow, equalLowTouches := detectEqualLow(sortedKlines)
	latest := sortedKlines[len(sortedKlines)-1]

	bids, asks, err := parseOrderBookSnapshot(snapshot)
	if err != nil {
		return models.Liquidity{}, err
	}
	if len(bids) == 0 || len(asks) == 0 {
		return models.Liquidity{}, errors.New("order book snapshot has no valid levels")
	}

	buyLiquidity, bidClusterNotional := strongestLiquidityCluster(bids, orderBookCluster)
	sellLiquidity, askClusterNotional := strongestLiquidityCluster(asks, orderBookCluster)
	if buyLiquidity <= 0 {
		buyLiquidity = baseBuyLiquidity
	}
	if sellLiquidity <= 0 {
		sellLiquidity = baseSellLiquidity
	}

	imbalance := calculateOrderBookImbalance(bids, asks, imbalanceLevels)
	sweepType := deriveSweepTypeFromOrderBook(latest, buyLiquidity, sellLiquidity, imbalance)
	if sweepType == "none" {
		sweepType = baseSweepType
	}

	// 如果盘口聚类结果明显异常，则回退到 K 线流动性位，避免使用失真快照。
	if bidClusterNotional <= 0 {
		buyLiquidity = baseBuyLiquidity
	}
	if askClusterNotional <= 0 {
		sellLiquidity = baseSellLiquidity
	}

	createdAt := latest.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	stopClusters := buildStopClusters(
		buyLiquidity,
		sellLiquidity,
		equalHigh,
		equalLow,
		float64(equalHighTouches),
		float64(equalLowTouches),
		normalizeWallStrength(askClusterNotional),
		normalizeWallStrength(bidClusterNotional),
	)

	return models.Liquidity{
		Symbol:             symbol,
		BuyLiquidity:       roundFloat(buyLiquidity, 8),
		SellLiquidity:      roundFloat(sellLiquidity, 8),
		SweepType:          sweepType,
		OrderBookImbalance: roundFloat(imbalance, 6),
		DataSource:         "orderbook",
		EqualHigh:          roundFloat(equalHigh, 8),
		EqualLow:           roundFloat(equalLow, 8),
		StopClusters:       stopClusters,
		CreatedAt:          createdAt,
	}, nil
}

func analyzeLiquidityFromKlines(sortedKlines []models.Kline) (float64, float64, string) {
	latest := sortedKlines[len(sortedKlines)-1]
	reference := tailKlines(sortedKlines[:len(sortedKlines)-1], clusterWindow)

	buyLiquidity := averageBottomNLows(reference, 3)
	sellLiquidity := averageTopNHighs(reference, 3)
	averageReferenceVolume := averageVolume(reference)
	sweepType := "none"

	lowerSweep := latest.LowPrice < buyLiquidity*(1-sweepTolerance) &&
		latest.ClosePrice > buyLiquidity &&
		latest.Volume >= averageReferenceVolume*0.85
	upperSweep := latest.HighPrice > sellLiquidity*(1+sweepTolerance) &&
		latest.ClosePrice < sellLiquidity &&
		latest.Volume >= averageReferenceVolume*0.85

	switch {
	case lowerSweep:
		sweepType = "sell_sweep"
	case upperSweep:
		sweepType = "buy_sweep"
	}

	return buyLiquidity, sellLiquidity, sweepType
}

func deriveSweepTypeFromOrderBook(latest models.Kline, buyLiquidity, sellLiquidity, imbalance float64) string {
	switch {
	case buyLiquidity > 0 &&
		latest.LowPrice < buyLiquidity*(1-depthSweepTolerance) &&
		latest.ClosePrice >= buyLiquidity &&
		imbalance >= -0.18:
		return "sell_sweep"
	case sellLiquidity > 0 &&
		latest.HighPrice > sellLiquidity*(1+depthSweepTolerance) &&
		latest.ClosePrice <= sellLiquidity &&
		imbalance <= 0.18:
		return "buy_sweep"
	default:
		return "none"
	}
}

func detectEqualHigh(sortedKlines []models.Kline) (float64, int) {
	reference := tailKlines(sortedKlines[:len(sortedKlines)-1], equalLevelLookback)
	values := make([]float64, 0, len(reference))
	for _, kline := range reference {
		values = append(values, kline.HighPrice)
	}
	return detectEqualPriceLevel(values, true)
}

func detectEqualLow(sortedKlines []models.Kline) (float64, int) {
	reference := tailKlines(sortedKlines[:len(sortedKlines)-1], equalLevelLookback)
	values := make([]float64, 0, len(reference))
	for _, kline := range reference {
		values = append(values, kline.LowPrice)
	}
	return detectEqualPriceLevel(values, false)
}

func detectEqualPriceLevel(values []float64, preferHigher bool) (float64, int) {
	if len(values) < 2 {
		return 0, 0
	}

	sortedValues := make([]float64, 0, len(values))
	for _, value := range values {
		if value > 0 {
			sortedValues = append(sortedValues, value)
		}
	}
	if len(sortedValues) < 2 {
		return 0, 0
	}

	sort.Float64s(sortedValues)
	bestPrice := 0.0
	bestCount := 0
	currentValues := []float64{sortedValues[0]}

	for index := 1; index < len(sortedValues); index++ {
		candidate := sortedValues[index]
		clusterMean := averageFloatSlice(currentValues)
		if math.Abs(candidate-clusterMean)/math.Max(clusterMean, 1) <= equalLevelTolerance {
			currentValues = append(currentValues, candidate)
			continue
		}

		bestPrice, bestCount = chooseBetterEqualLevel(bestPrice, bestCount, currentValues, preferHigher)
		currentValues = []float64{candidate}
	}
	bestPrice, bestCount = chooseBetterEqualLevel(bestPrice, bestCount, currentValues, preferHigher)

	if bestCount < 2 {
		return 0, 0
	}
	return bestPrice, bestCount
}

func chooseBetterEqualLevel(bestPrice float64, bestCount int, cluster []float64, preferHigher bool) (float64, int) {
	if len(cluster) < 2 {
		return bestPrice, bestCount
	}

	clusterPrice := averageFloatSlice(cluster)
	clusterCount := len(cluster)
	if clusterCount > bestCount {
		return clusterPrice, clusterCount
	}
	if clusterCount < bestCount {
		return bestPrice, bestCount
	}
	if preferHigher && clusterPrice > bestPrice {
		return clusterPrice, clusterCount
	}
	if !preferHigher && (bestPrice == 0 || clusterPrice < bestPrice) {
		return clusterPrice, clusterCount
	}
	return bestPrice, bestCount
}

func buildStopClusters(
	buyLiquidity, sellLiquidity, equalHigh, equalLow float64,
	equalHighStrength, equalLowStrength, askWallStrength, bidWallStrength float64,
) []models.LiquidityCluster {
	clusters := make([]models.LiquidityCluster, 0, 4)

	if equalHigh > 0 {
		clusters = append(clusters, models.LiquidityCluster{
			Label:    "Equal High Stops",
			Kind:     "sell_stop_cluster",
			Price:    roundFloat(equalHigh, 8),
			Strength: roundFloat(equalHighStrength, 2),
		})
	}
	if equalLow > 0 {
		clusters = append(clusters, models.LiquidityCluster{
			Label:    "Equal Low Stops",
			Kind:     "buy_stop_cluster",
			Price:    roundFloat(equalLow, 8),
			Strength: roundFloat(equalLowStrength, 2),
		})
	}
	if sellLiquidity > 0 && !isNearLevel(sellLiquidity, equalHigh) {
		clusters = append(clusters, models.LiquidityCluster{
			Label:    "Ask Wall",
			Kind:     "sell_liquidity_wall",
			Price:    roundFloat(sellLiquidity, 8),
			Strength: roundFloat(math.Max(askWallStrength, 1), 2),
		})
	}
	if buyLiquidity > 0 && !isNearLevel(buyLiquidity, equalLow) {
		clusters = append(clusters, models.LiquidityCluster{
			Label:    "Bid Wall",
			Kind:     "buy_liquidity_wall",
			Price:    roundFloat(buyLiquidity, 8),
			Strength: roundFloat(math.Max(bidWallStrength, 1), 2),
		})
	}
	return clusters
}

func normalizeWallStrength(notional float64) float64 {
	if notional <= 0 {
		return 0
	}
	return notional / 100000
}

func isNearLevel(left, right float64) bool {
	if left <= 0 || right <= 0 {
		return false
	}
	return math.Abs(left-right)/math.Max(math.Abs(left), 1) <= equalLevelTolerance
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

func strongestLiquidityCluster(levels []binancepkg.OrderBookLevel, window int) (float64, float64) {
	if len(levels) == 0 {
		return 0, 0
	}
	if window <= 0 {
		window = 1
	}

	bestPrice := 0.0
	bestNotional := 0.0
	for start := 0; start < len(levels); start++ {
		end := minInt(start+window, len(levels))
		totalQuantity := 0.0
		totalNotional := 0.0
		for _, level := range levels[start:end] {
			if level.Price <= 0 || level.Quantity <= 0 {
				continue
			}
			totalQuantity += level.Quantity
			totalNotional += level.Price * level.Quantity
		}
		if totalQuantity <= 0 || totalNotional <= bestNotional {
			continue
		}

		bestNotional = totalNotional
		bestPrice = totalNotional / totalQuantity
	}

	return bestPrice, bestNotional
}

func calculateOrderBookImbalance(bids, asks []binancepkg.OrderBookLevel, levels int) float64 {
	bidNotional := topOrderBookNotional(bids, levels)
	askNotional := topOrderBookNotional(asks, levels)
	totalNotional := bidNotional + askNotional
	if totalNotional <= 0 {
		return 0
	}
	return (bidNotional - askNotional) / totalNotional
}

func topOrderBookNotional(levels []binancepkg.OrderBookLevel, limit int) float64 {
	if limit <= 0 || len(levels) == 0 {
		return 0
	}
	if len(levels) < limit {
		limit = len(levels)
	}

	total := 0.0
	for _, level := range levels[:limit] {
		total += level.Price * level.Quantity
	}
	return total
}

func averageFloatSlice(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func tailKlines(klines []models.Kline, size int) []models.Kline {
	if len(klines) <= size {
		return klines
	}
	return klines[len(klines)-size:]
}

func averageTopNHighs(klines []models.Kline, n int) float64 {
	if len(klines) == 0 || n <= 0 {
		return 0
	}

	highs := make([]float64, 0, len(klines))
	for _, kline := range klines {
		highs = append(highs, kline.HighPrice)
	}
	sort.Float64s(highs)
	if len(highs) < n {
		n = len(highs)
	}

	total := 0.0
	for _, high := range highs[len(highs)-n:] {
		total += high
	}
	return total / float64(n)
}

func averageBottomNLows(klines []models.Kline, n int) float64 {
	if len(klines) == 0 || n <= 0 {
		return 0
	}

	lows := make([]float64, 0, len(klines))
	for _, kline := range klines {
		lows = append(lows, kline.LowPrice)
	}
	sort.Float64s(lows)
	if len(lows) < n {
		n = len(lows)
	}

	total := 0.0
	for _, low := range lows[:n] {
		total += low
	}
	return total / float64(n)
}

func averageVolume(klines []models.Kline) float64 {
	if len(klines) == 0 {
		return 0
	}

	total := 0.0
	for _, kline := range klines {
		total += kline.Volume
	}
	return total / float64(len(klines))
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

func roundFloat(value float64, precision int) float64 {
	pow := math.Pow10(precision)
	return math.Round(value*pow) / pow
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
