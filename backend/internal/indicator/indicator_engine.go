package indicator

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"alpha-pulse/backend/models"
)

const (
	rsiPeriod        = 14
	macdFastPeriod   = 12
	macdSlowPeriod   = 26
	macdSignalPeriod = 9
	ema20Period      = 20
	ema50Period      = 50
	atrPeriod        = 14
	bollingerPeriod  = 20
	bollingerStdDev  = 2.0

	// indicatorHistoryLimit 既满足最小计算窗口，也给 EMA/MACD 留出足够预热长度。
	indicatorHistoryLimit = 120
)

// Engine 负责基于历史 K 线计算技术指标。
type Engine struct {
	historyLimit int
}

// NewEngine 创建指标引擎。
func NewEngine() *Engine {
	return &Engine{
		historyLimit: indicatorHistoryLimit,
	}
}

// HistoryLimit 返回建议拉取的历史 K 线数量。
func (e *Engine) HistoryLimit() int {
	return e.historyLimit
}

// MinimumRequired 返回指标计算所需的最小 K 线数量。
func (e *Engine) MinimumRequired() int {
	return maxInt(
		ema50Period,
		rsiPeriod+1,
		atrPeriod+1,
		bollingerPeriod,
		macdSlowPeriod+macdSignalPeriod-1,
	)
}

// Calculate 基于历史 K 线计算 RSI、MACD、EMA20、EMA50、ATR、Bollinger Bands、VWAP。
func (e *Engine) Calculate(symbol string, klines []models.Kline) (models.Indicator, error) {
	if len(klines) < e.MinimumRequired() {
		return models.Indicator{}, fmt.Errorf(
			"indicator engine requires at least %d klines, got %d",
			e.MinimumRequired(),
			len(klines),
		)
	}

	sorted := sortKlinesAscending(klines)
	closes := make([]float64, 0, len(sorted))
	highs := make([]float64, 0, len(sorted))
	lows := make([]float64, 0, len(sorted))
	for _, kline := range sorted {
		closes = append(closes, kline.ClosePrice)
		highs = append(highs, kline.HighPrice)
		lows = append(lows, kline.LowPrice)
	}

	rsi, err := calculateRSI(closes, rsiPeriod)
	if err != nil {
		return models.Indicator{}, err
	}

	ema20, err := calculateEMA(closes, ema20Period)
	if err != nil {
		return models.Indicator{}, err
	}

	ema50, err := calculateEMA(closes, ema50Period)
	if err != nil {
		return models.Indicator{}, err
	}

	macdLine, signalLine, err := calculateMACD(closes, macdFastPeriod, macdSlowPeriod, macdSignalPeriod)
	if err != nil {
		return models.Indicator{}, err
	}
	macdHistogram := macdLine - signalLine

	atr, err := calculateATR(highs, lows, closes, atrPeriod)
	if err != nil {
		return models.Indicator{}, err
	}

	bollingerUpper, bollingerMiddle, bollingerLower, err := calculateBollingerBands(closes, bollingerPeriod, bollingerStdDev)
	if err != nil {
		return models.Indicator{}, err
	}

	vwap, err := calculateVWAP(sorted)
	if err != nil {
		return models.Indicator{}, err
	}

	latest := sorted[len(sorted)-1]
	createdAt := latest.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	return models.Indicator{
		Symbol:          symbol,
		RSI:             round(rsi, 2),
		MACD:            round(macdLine, 4),
		MACDSignal:      round(signalLine, 4),
		MACDHistogram:   round(macdHistogram, 4),
		EMA20:           round(ema20, 8),
		EMA50:           round(ema50, 8),
		ATR:             round(atr, 8),
		BollingerUpper:  round(bollingerUpper, 8),
		BollingerMiddle: round(bollingerMiddle, 8),
		BollingerLower:  round(bollingerLower, 8),
		VWAP:            round(vwap, 8),
		CreatedAt:       createdAt,
	}, nil
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

func calculateRSI(closes []float64, period int) (float64, error) {
	if len(closes) < period+1 {
		return 0, errors.New("not enough closes to calculate RSI")
	}

	var gainSum float64
	var lossSum float64
	for i := 1; i <= period; i++ {
		change := closes[i] - closes[i-1]
		if change >= 0 {
			gainSum += change
		} else {
			lossSum += -change
		}
	}

	avgGain := gainSum / float64(period)
	avgLoss := lossSum / float64(period)

	for i := period + 1; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		gain := 0.0
		loss := 0.0
		if change >= 0 {
			gain = change
		} else {
			loss = -change
		}

		// 使用 Wilder 平滑算法更新平均涨跌幅。
		avgGain = ((avgGain * float64(period-1)) + gain) / float64(period)
		avgLoss = ((avgLoss * float64(period-1)) + loss) / float64(period)
	}

	if avgLoss == 0 {
		return 100, nil
	}

	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs)), nil
}

func calculateEMA(values []float64, period int) (float64, error) {
	series, err := emaSeries(values, period)
	if err != nil {
		return 0, err
	}

	return series[len(series)-1], nil
}

func emaSeries(values []float64, period int) ([]float64, error) {
	if period <= 0 {
		return nil, errors.New("EMA period must be greater than zero")
	}
	if len(values) < period {
		return nil, fmt.Errorf("not enough values to calculate EMA%d", period)
	}

	seed := average(values[:period])
	result := []float64{seed}
	multiplier := 2.0 / float64(period+1)
	current := seed

	for _, value := range values[period:] {
		current = ((value - current) * multiplier) + current
		result = append(result, current)
	}

	return result, nil
}

func calculateMACD(closes []float64, fastPeriod, slowPeriod, signalPeriod int) (float64, float64, error) {
	if fastPeriod >= slowPeriod {
		return 0, 0, errors.New("MACD fast period must be smaller than slow period")
	}

	fastEMA, err := emaSeries(closes, fastPeriod)
	if err != nil {
		return 0, 0, err
	}

	slowEMA, err := emaSeries(closes, slowPeriod)
	if err != nil {
		return 0, 0, err
	}

	offset := slowPeriod - fastPeriod
	macdSeries := make([]float64, 0, len(slowEMA))
	for i := range slowEMA {
		macdSeries = append(macdSeries, fastEMA[i+offset]-slowEMA[i])
	}

	signalSeries, err := emaSeries(macdSeries, signalPeriod)
	if err != nil {
		return 0, 0, err
	}

	return macdSeries[len(macdSeries)-1], signalSeries[len(signalSeries)-1], nil
}

func calculateATR(highs, lows, closes []float64, period int) (float64, error) {
	if len(highs) != len(lows) || len(lows) != len(closes) {
		return 0, errors.New("ATR input slices must have the same length")
	}
	if len(closes) < period+1 {
		return 0, errors.New("not enough klines to calculate ATR")
	}

	trs := make([]float64, 0, len(closes)-1)
	for i := 1; i < len(closes); i++ {
		highLow := highs[i] - lows[i]
		highPrevClose := math.Abs(highs[i] - closes[i-1])
		lowPrevClose := math.Abs(lows[i] - closes[i-1])

		trs = append(trs, math.Max(highLow, math.Max(highPrevClose, lowPrevClose)))
	}

	atr := average(trs[:period])
	for _, tr := range trs[period:] {
		// ATR 同样使用 Wilder 平滑方式处理。
		atr = ((atr * float64(period-1)) + tr) / float64(period)
	}

	return atr, nil
}

func calculateBollingerBands(closes []float64, period int, stdDevMultiplier float64) (float64, float64, float64, error) {
	if len(closes) < period {
		return 0, 0, 0, errors.New("not enough closes to calculate Bollinger Bands")
	}

	window := closes[len(closes)-period:]
	middle := average(window)
	var variance float64
	for _, close := range window {
		diff := close - middle
		variance += diff * diff
	}

	stdDev := math.Sqrt(variance / float64(period))
	upper := middle + stdDevMultiplier*stdDev
	lower := middle - stdDevMultiplier*stdDev

	return upper, middle, lower, nil
}

func calculateVWAP(klines []models.Kline) (float64, error) {
	if len(klines) == 0 {
		return 0, errors.New("not enough klines to calculate VWAP")
	}

	totalPriceVolume := 0.0
	totalVolume := 0.0
	for _, kline := range klines {
		typicalPrice := (kline.HighPrice + kline.LowPrice + kline.ClosePrice) / 3
		totalPriceVolume += typicalPrice * kline.Volume
		totalVolume += kline.Volume
	}

	if totalVolume == 0 {
		return 0, errors.New("VWAP total volume can not be zero")
	}

	return totalPriceVolume / totalVolume, nil
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func round(value float64, precision int) float64 {
	if precision < 0 {
		return value
	}

	pow := math.Pow10(precision)
	return math.Round(value*pow) / pow
}

func maxInt(values ...int) int {
	maxValue := values[0]
	for _, value := range values[1:] {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}
