package structure

import (
	"errors"
	"math"
	"sort"
	"time"

	"alpha-pulse/backend/models"
)

const (
	historyLimit      = 80
	minimumRequired   = 30
	referenceWindow   = 18
	swingWindow       = 2
	breakoutTolerance = 0.0015
	pivotTolerance    = 0.0008
)

// Engine 负责市场结构分析。
type Engine struct {
	historyLimit int
}

type swingPoint struct {
	kind     string
	price    float64
	openTime int64
}

// NewEngine 创建市场结构引擎。
func NewEngine() *Engine {
	return &Engine{historyLimit: historyLimit}
}

// HistoryLimit 返回市场结构分析建议使用的历史 K 线数量。
func (e *Engine) HistoryLimit() int {
	return e.historyLimit
}

// MinimumRequired 返回市场结构分析所需的最小 K 线数量。
func (e *Engine) MinimumRequired() int {
	return minimumRequired
}

// Analyze 基于 swing point 识别趋势、支撑阻力与 HH/HL/LH/LL 结构变化。
func (e *Engine) Analyze(symbol string, klines []models.Kline) (models.Structure, error) {
	if len(klines) < e.MinimumRequired() {
		return models.Structure{}, errors.New("not enough klines to analyze structure")
	}

	sortedKlines := sortKlinesAscending(klines)
	latest := sortedKlines[len(sortedKlines)-1]
	reference := tailKlines(sortedKlines[:len(sortedKlines)-1], referenceWindow)
	supportFallback := averageBottomNLows(reference, 3)
	resistanceFallback := averageTopNHighs(reference, 3)

	swingPoints := detectSwingPoints(sortedKlines[:len(sortedKlines)-1], swingWindow)
	events, support, resistance, trend := buildStructureEvents(swingPoints)
	if support <= 0 {
		support = supportFallback
	}
	if resistance <= 0 {
		resistance = resistanceFallback
	}
	if trend == "" {
		trend = classifyTrendFallback(reference, latest)
	}

	latestClose := latest.ClosePrice
	bullishBreak := resistance > 0 && latestClose > resistance*(1+breakoutTolerance)
	bearishBreak := support > 0 && latestClose < support*(1-breakoutTolerance)

	bos := (trend == "uptrend" && bullishBreak) || (trend == "downtrend" && bearishBreak)
	choch := (trend == "uptrend" && bearishBreak) || (trend == "downtrend" && bullishBreak)

	if bos {
		events = append(events, models.StructureEvent{
			Label:    "BOS",
			Kind:     breakoutKind(bullishBreak, bearishBreak),
			Price:    roundFloat(latestClose, 8),
			OpenTime: latest.OpenTime,
		})
	}
	if choch {
		events = append(events, models.StructureEvent{
			Label:    "CHOCH",
			Kind:     breakoutKind(bullishBreak, bearishBreak),
			Price:    roundFloat(latestClose, 8),
			OpenTime: latest.OpenTime,
		})
	}

	createdAt := latest.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	return models.Structure{
		Symbol:     symbol,
		Trend:      trend,
		Support:    roundFloat(support, 8),
		Resistance: roundFloat(resistance, 8),
		BOS:        bos,
		Choch:      choch,
		CreatedAt:  createdAt,
		Events:     events,
	}, nil
}

func buildStructureEvents(points []swingPoint) ([]models.StructureEvent, float64, float64, string) {
	events := make([]models.StructureEvent, 0, len(points)+2)

	lastHigh := 0.0
	lastLow := 0.0
	lastHighLabel := ""
	lastLowLabel := ""
	hasHigh := false
	hasLow := false

	for _, point := range points {
		switch point.kind {
		case "swing_high":
			label := classifyHigh(point.price, lastHigh, hasHigh)
			hasHigh = true
			lastHigh = point.price
			if label != "" {
				lastHighLabel = label
				events = append(events, models.StructureEvent{
					Label:    label,
					Kind:     point.kind,
					Price:    roundFloat(point.price, 8),
					OpenTime: point.openTime,
				})
			}
		case "swing_low":
			label := classifyLow(point.price, lastLow, hasLow)
			hasLow = true
			lastLow = point.price
			if label != "" {
				lastLowLabel = label
				events = append(events, models.StructureEvent{
					Label:    label,
					Kind:     point.kind,
					Price:    roundFloat(point.price, 8),
					OpenTime: point.openTime,
				})
			}
		}
	}

	trend := deriveTrendFromLabels(lastHighLabel, lastLowLabel)
	return tailEvents(events, 12), lastLow, lastHigh, trend
}

func detectSwingPoints(klines []models.Kline, window int) []swingPoint {
	if len(klines) < window*2+1 {
		return nil
	}

	points := make([]swingPoint, 0, len(klines)/3)
	for index := window; index < len(klines)-window; index++ {
		current := klines[index]
		isHigh := true
		isLow := true

		for offset := 1; offset <= window; offset++ {
			left := klines[index-offset]
			right := klines[index+offset]

			if current.HighPrice <= left.HighPrice || current.HighPrice <= right.HighPrice {
				isHigh = false
			}
			if current.LowPrice >= left.LowPrice || current.LowPrice >= right.LowPrice {
				isLow = false
			}

			if !isHigh && !isLow {
				break
			}
		}

		if isHigh {
			points = append(points, swingPoint{
				kind:     "swing_high",
				price:    current.HighPrice,
				openTime: current.OpenTime,
			})
		}
		if isLow {
			points = append(points, swingPoint{
				kind:     "swing_low",
				price:    current.LowPrice,
				openTime: current.OpenTime,
			})
		}
	}

	sort.Slice(points, func(i, j int) bool {
		if points[i].openTime == points[j].openTime {
			return points[i].kind < points[j].kind
		}
		return points[i].openTime < points[j].openTime
	})

	return points
}

func classifyHigh(price, previous float64, hasPrevious bool) string {
	if !hasPrevious || previous <= 0 {
		return ""
	}

	switch {
	case price > previous*(1+pivotTolerance):
		return "HH"
	case price < previous*(1-pivotTolerance):
		return "LH"
	default:
		return ""
	}
}

func classifyLow(price, previous float64, hasPrevious bool) string {
	if !hasPrevious || previous <= 0 {
		return ""
	}

	switch {
	case price > previous*(1+pivotTolerance):
		return "HL"
	case price < previous*(1-pivotTolerance):
		return "LL"
	default:
		return ""
	}
}

func deriveTrendFromLabels(lastHighLabel, lastLowLabel string) string {
	switch {
	case lastHighLabel == "HH" && lastLowLabel == "HL":
		return "uptrend"
	case lastHighLabel == "LH" && lastLowLabel == "LL":
		return "downtrend"
	default:
		return ""
	}
}

func classifyTrendFallback(reference []models.Kline, latest models.Kline) string {
	if len(reference) == 0 {
		return "range"
	}

	support := averageBottomNLows(reference, minInt(len(reference), 3))
	resistance := averageTopNHighs(reference, minInt(len(reference), 3))
	switch {
	case latest.ClosePrice > resistance*(1-breakoutTolerance):
		return "uptrend"
	case latest.ClosePrice < support*(1+breakoutTolerance):
		return "downtrend"
	default:
		return "range"
	}
}

func breakoutKind(bullishBreak, bearishBreak bool) string {
	switch {
	case bullishBreak:
		return "bullish_break"
	case bearishBreak:
		return "bearish_break"
	default:
		return "break"
	}
}

func tailEvents(events []models.StructureEvent, size int) []models.StructureEvent {
	if len(events) <= size {
		return events
	}
	return events[len(events)-size:]
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

	sum := 0.0
	for _, value := range highs[len(highs)-n:] {
		sum += value
	}
	return sum / float64(n)
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

	sum := 0.0
	for _, value := range lows[:n] {
		sum += value
	}
	return sum / float64(n)
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
