package service

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"alpha-pulse/backend/internal/indicator"
	"alpha-pulse/backend/internal/liquidity"
	structureengine "alpha-pulse/backend/internal/structure"
	"alpha-pulse/backend/models"
)

type orderBookSeriesRepository interface {
	GetSeriesWindow(symbol string, startTime, endTime int64) ([]models.OrderBookSnapshot, error)
}

const maxOrderBookSeriesSpanMillis int64 = 4 * 60 * 60 * 1000

// buildIndicatorSeries 基于滚动窗口生成指标时间序列，供图表直接使用后端计算结果。
func buildIndicatorSeries(
	engine *indicator.Engine,
	symbol string,
	allKlines []models.Kline,
	limit int,
) ([]models.IndicatorSeriesPoint, error) {
	if engine == nil {
		return nil, errors.New("indicator engine is nil")
	}

	sortedKlines := sortServiceKlinesAscending(allKlines)
	if len(sortedKlines) < engine.MinimumRequired() {
		return nil, indicatorInsufficientHistoryError(engine.MinimumRequired(), len(sortedKlines))
	}

	start := len(sortedKlines) - clampInt(limit, 1, len(sortedKlines))
	points := make([]models.IndicatorSeriesPoint, 0, len(sortedKlines)-start)

	for index := start; index < len(sortedKlines); index++ {
		window := takeLastKlines(sortedKlines[:index+1], engine.HistoryLimit())
		result, err := engine.Calculate(symbol, window)
		if err != nil {
			return nil, err
		}

		points = append(points, models.IndicatorSeriesPoint{
			OpenTime:        sortedKlines[index].OpenTime,
			RSI:             result.RSI,
			MACD:            result.MACD,
			MACDSignal:      result.MACDSignal,
			MACDHistogram:   result.MACDHistogram,
			EMA20:           result.EMA20,
			EMA50:           result.EMA50,
			ATR:             result.ATR,
			BollingerUpper:  result.BollingerUpper,
			BollingerMiddle: result.BollingerMiddle,
			BollingerLower:  result.BollingerLower,
			VWAP:            result.VWAP,
		})
	}

	return points, nil
}

// buildStructureSeries 基于滚动窗口生成结构时间序列，供图表绘制动态支撑阻力。
func buildStructureSeries(
	engine *structureengine.Engine,
	symbol string,
	allKlines []models.Kline,
	limit int,
) ([]StructureSeriesPoint, error) {
	if engine == nil {
		return nil, errors.New("structure engine is nil")
	}

	sortedKlines := sortServiceKlinesAscending(allKlines)
	if len(sortedKlines) < engine.MinimumRequired() {
		return nil, indicatorInsufficientHistoryError(engine.MinimumRequired(), len(sortedKlines))
	}

	start := len(sortedKlines) - clampInt(limit, 1, len(sortedKlines))
	points := make([]StructureSeriesPoint, 0, len(sortedKlines)-start)

	for index := start; index < len(sortedKlines); index++ {
		window := takeLastKlines(sortedKlines[:index+1], engine.HistoryLimit())
		result, err := engine.Analyze(symbol, window)
		if err != nil {
			return nil, err
		}

		points = append(points, StructureSeriesPoint{
			OpenTime:           sortedKlines[index].OpenTime,
			Trend:              result.Trend,
			PrimaryTier:        result.PrimaryTier,
			Support:            result.Support,
			Resistance:         result.Resistance,
			InternalSupport:    result.InternalSupport,
			InternalResistance: result.InternalResistance,
			ExternalSupport:    result.ExternalSupport,
			ExternalResistance: result.ExternalResistance,
			BOS:                result.BOS,
			Choch:              result.Choch,
			EventLabels:        eventLabelsAtTime(result.Events, sortedKlines[index].OpenTime),
			EventTags:          eventTagsAtTime(result.Events, sortedKlines[index].OpenTime),
		})
	}

	return points, nil
}

// buildLiquiditySeries 基于滚动窗口生成流动性时间序列，优先使用对应时间点之前的盘口快照。
func buildLiquiditySeries(
	engine *liquidity.Engine,
	orderBookRepo orderBookSeriesRepository,
	symbol, interval string,
	allKlines []models.Kline,
	limit int,
) ([]LiquiditySeriesPoint, error) {
	if engine == nil {
		return nil, errors.New("liquidity engine is nil")
	}

	sortedKlines := sortServiceKlinesAscending(allKlines)
	if len(sortedKlines) < engine.MinimumRequired() {
		return nil, indicatorInsufficientHistoryError(engine.MinimumRequired(), len(sortedKlines))
	}

	start := len(sortedKlines) - clampInt(limit, 1, len(sortedKlines))
	points := make([]LiquiditySeriesPoint, 0, len(sortedKlines)-start)
	intervalMillis := intervalDurationMillis(interval)
	targetTimes := make([]int64, 0, len(sortedKlines)-start)

	for index := start; index < len(sortedKlines); index++ {
		targetTimes = append(targetTimes, sortedKlines[index].OpenTime+intervalMillis)
	}

	snapshots, err := loadOrderBookSeriesWindow(orderBookRepo, symbol, targetTimes)
	if err != nil {
		return nil, err
	}
	resolver := newOrderBookSnapshotResolver(snapshots)

	for pointIndex, index := 0, start; index < len(sortedKlines); index, pointIndex = index+1, pointIndex+1 {
		window := takeLastKlines(sortedKlines[:index+1], engine.HistoryLimit())
		result, err := buildLiquidityPoint(engine, symbol, window, resolver.Resolve(targetTimes[pointIndex]))
		if err != nil {
			return nil, err
		}

		buyStrength, sellStrength := clusterStrengths(result.StopClusters)
		points = append(points, LiquiditySeriesPoint{
			OpenTime:            sortedKlines[index].OpenTime,
			BuyLiquidity:        result.BuyLiquidity,
			SellLiquidity:       result.SellLiquidity,
			SweepType:           result.SweepType,
			OrderBookImbalance:  result.OrderBookImbalance,
			DataSource:          result.DataSource,
			EqualHigh:           result.EqualHigh,
			EqualLow:            result.EqualLow,
			BuyClusterStrength:  buyStrength,
			SellClusterStrength: sellStrength,
		})
	}

	return points, nil
}

func buildLiquidityPoint(
	engine *liquidity.Engine,
	symbol string,
	window []models.Kline,
	snapshot *models.OrderBookSnapshot,
) (models.Liquidity, error) {
	if snapshot != nil {
		result, analyzeErr := engine.AnalyzeWithOrderBook(symbol, window, *snapshot)
		if analyzeErr == nil {
			return result, nil
		}
	}

	return engine.Analyze(symbol, window)
}

func loadOrderBookSeriesWindow(
	orderBookRepo orderBookSeriesRepository,
	symbol string,
	targetTimes []int64,
) ([]models.OrderBookSnapshot, error) {
	if orderBookRepo == nil || len(targetTimes) == 0 {
		return nil, nil
	}
	if targetTimes[len(targetTimes)-1]-targetTimes[0] > maxOrderBookSeriesSpanMillis {
		return nil, nil
	}
	return orderBookRepo.GetSeriesWindow(symbol, targetTimes[0], targetTimes[len(targetTimes)-1])
}

type orderBookSnapshotResolver struct {
	snapshots []models.OrderBookSnapshot
	index     int
	current   *models.OrderBookSnapshot
}

func newOrderBookSnapshotResolver(snapshots []models.OrderBookSnapshot) *orderBookSnapshotResolver {
	return &orderBookSnapshotResolver{snapshots: snapshots}
}

func (r *orderBookSnapshotResolver) Resolve(targetTime int64) *models.OrderBookSnapshot {
	for r.index < len(r.snapshots) && r.snapshots[r.index].EventTime <= targetTime {
		r.current = &r.snapshots[r.index]
		r.index++
	}
	return r.current
}

func clusterStrengths(clusters []models.LiquidityCluster) (float64, float64) {
	buyStrength := 0.0
	sellStrength := 0.0
	for _, cluster := range clusters {
		if cluster.Strength <= 0 {
			continue
		}
		switch {
		case strings.Contains(cluster.Kind, "buy"):
			if cluster.Strength > buyStrength {
				buyStrength = cluster.Strength
			}
		case strings.Contains(cluster.Kind, "sell"):
			if cluster.Strength > sellStrength {
				sellStrength = cluster.Strength
			}
		}
	}
	return buyStrength, sellStrength
}

func eventLabelsAtTime(events []models.StructureEvent, openTime int64) []string {
	labels := make([]string, 0, 2)
	for _, event := range events {
		if event.OpenTime == openTime {
			labels = append(labels, event.Label)
		}
	}
	return labels
}

func eventTagsAtTime(events []models.StructureEvent, openTime int64) []string {
	tags := make([]string, 0, 2)
	for _, event := range events {
		if event.OpenTime != openTime {
			continue
		}
		if event.Tier != "" {
			tags = append(tags, event.Tier+":"+event.Label)
			continue
		}
		tags = append(tags, event.Label)
	}
	return tags
}

func sortServiceKlinesAscending(klines []models.Kline) []models.Kline {
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

func intervalDurationMillis(interval string) int64 {
	if len(interval) < 2 {
		return 60_000
	}

	unit := interval[len(interval)-1]
	value, err := strconv.Atoi(interval[:len(interval)-1])
	if err != nil || value <= 0 {
		return 60_000
	}

	switch unit {
	case 'm':
		return int64(value) * 60_000
	case 'h':
		return int64(value) * 60 * 60_000
	case 'd':
		return int64(value) * 24 * 60 * 60_000
	default:
		return 60_000
	}
}
