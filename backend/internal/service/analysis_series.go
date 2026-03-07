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
	"alpha-pulse/backend/repository"
	"gorm.io/gorm"
)

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
			OpenTime:    sortedKlines[index].OpenTime,
			Trend:       result.Trend,
			Support:     result.Support,
			Resistance:  result.Resistance,
			BOS:         result.BOS,
			Choch:       result.Choch,
			EventLabels: eventLabelsAtTime(result.Events, sortedKlines[index].OpenTime),
		})
	}

	return points, nil
}

// buildLiquiditySeries 基于滚动窗口生成流动性时间序列，优先使用对应时间点之前的盘口快照。
func buildLiquiditySeries(
	engine *liquidity.Engine,
	orderBookRepo *repository.OrderBookSnapshotRepository,
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

	for index := start; index < len(sortedKlines); index++ {
		window := takeLastKlines(sortedKlines[:index+1], engine.HistoryLimit())
		targetTime := sortedKlines[index].OpenTime + intervalMillis

		result, err := buildLiquidityPoint(engine, orderBookRepo, symbol, window, targetTime)
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
	orderBookRepo *repository.OrderBookSnapshotRepository,
	symbol string,
	window []models.Kline,
	targetTime int64,
) (models.Liquidity, error) {
	if orderBookRepo != nil && targetTime > 0 {
		snapshot, err := orderBookRepo.GetLatestBefore(symbol, targetTime)
		if err == nil {
			result, analyzeErr := engine.AnalyzeWithOrderBook(symbol, window, snapshot)
			if analyzeErr == nil {
				return result, nil
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Liquidity{}, err
		}
	}

	return engine.Analyze(symbol, window)
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
