package service

import (
	"math"

	"alpha-pulse/backend/internal/collector"
	"alpha-pulse/backend/internal/liquidity"
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
)

const wallEvolutionSeriesLimit = 4

func enrichLiquidityDepthContext(
	collector *collector.BinanceCollector,
	engine *liquidity.Engine,
	klineRepo *repository.KlineRepository,
	orderBookRepo *repository.OrderBookSnapshotRepository,
	symbol, currentInterval string,
	currentPrice float64,
	result *models.Liquidity,
	currentSeries []LiquiditySeriesPoint,
) {
	if result == nil {
		return
	}
	if result.WallLevels == nil {
		result.WallLevels = []models.LiquidityWallLevel{}
	}
	if result.WallStrengthBands == nil {
		result.WallStrengthBands = []models.LiquidityWallStrengthBand{}
	}

	evolution := buildLiquidityWallEvolution(
		collector,
		engine,
		klineRepo,
		orderBookRepo,
		symbol,
		currentInterval,
		resolveLiquidityReferencePrice(currentPrice, *result),
		currentSeries,
	)
	if evolution == nil {
		evolution = []models.LiquidityWallEvolution{}
	}
	result.WallEvolution = evolution
}

func buildLiquidityWallEvolution(
	collector *collector.BinanceCollector,
	engine *liquidity.Engine,
	klineRepo *repository.KlineRepository,
	orderBookRepo *repository.OrderBookSnapshotRepository,
	symbol, currentInterval string,
	currentPrice float64,
	currentSeries []LiquiditySeriesPoint,
) []models.LiquidityWallEvolution {
	if engine == nil || klineRepo == nil || symbol == "" {
		return nil
	}

	requiredLimit := maxInt(engine.HistoryLimit(), engine.MinimumRequired()+wallEvolutionSeriesLimit-1)
	evolution := make([]models.LiquidityWallEvolution, 0, len(supportedCacheIntervals))

	for _, interval := range orderedLiquidityEvolutionIntervals(currentInterval) {
		points := currentSeries
		if interval != currentInterval || len(points) == 0 {
			klines, err := loadAnalysisKlinesWithLimit(
				collector,
				klineRepo,
				symbol,
				interval,
				requiredLimit,
				engine.MinimumRequired(),
			)
			if err != nil {
				continue
			}

			series, err := buildLiquiditySeries(engine, orderBookRepo, symbol, interval, klines, wallEvolutionSeriesLimit)
			if err != nil {
				continue
			}
			points = series
		}

		point, ok := summarizeLiquidityWallEvolution(interval, currentPrice, points)
		if !ok {
			continue
		}
		evolution = append(evolution, point)
	}

	return evolution
}

func orderedLiquidityEvolutionIntervals(currentInterval string) []string {
	ordered := make([]string, 0, len(supportedCacheIntervals)+1)
	seen := make(map[string]struct{}, len(supportedCacheIntervals)+1)

	if currentInterval != "" {
		ordered = append(ordered, currentInterval)
		seen[currentInterval] = struct{}{}
	}

	for _, interval := range supportedCacheIntervals {
		if _, ok := seen[interval]; ok {
			continue
		}
		ordered = append(ordered, interval)
		seen[interval] = struct{}{}
	}

	return ordered
}

func summarizeLiquidityWallEvolution(
	interval string,
	currentPrice float64,
	points []LiquiditySeriesPoint,
) (models.LiquidityWallEvolution, bool) {
	if len(points) == 0 {
		return models.LiquidityWallEvolution{}, false
	}

	latest := points[len(points)-1]
	baseline := points[0]

	return models.LiquidityWallEvolution{
		Interval:            interval,
		BuyLiquidity:        roundServiceFloat(latest.BuyLiquidity, 8),
		SellLiquidity:       roundServiceFloat(latest.SellLiquidity, 8),
		BuyDistanceBps:      roundServiceFloat(distanceBps(currentPrice, latest.BuyLiquidity), 2),
		SellDistanceBps:     roundServiceFloat(distanceBps(currentPrice, latest.SellLiquidity), 2),
		BuyClusterStrength:  roundServiceFloat(latest.BuyClusterStrength, 2),
		SellClusterStrength: roundServiceFloat(latest.SellClusterStrength, 2),
		BuyStrengthDelta:    roundServiceFloat(latest.BuyClusterStrength-baseline.BuyClusterStrength, 2),
		SellStrengthDelta:   roundServiceFloat(latest.SellClusterStrength-baseline.SellClusterStrength, 2),
		OrderBookImbalance:  roundServiceFloat(latest.OrderBookImbalance, 4),
		SweepType:           latest.SweepType,
		DataSource:          latest.DataSource,
		DominantSide:        dominantWallSide(latest),
	}, true
}

func dominantWallSide(point LiquiditySeriesPoint) string {
	switch {
	case point.BuyClusterStrength >= point.SellClusterStrength+0.25 || point.OrderBookImbalance >= 0.08:
		return "bid"
	case point.SellClusterStrength >= point.BuyClusterStrength+0.25 || point.OrderBookImbalance <= -0.08:
		return "ask"
	default:
		return "balanced"
	}
}

func resolveLiquidityReferencePrice(currentPrice float64, liquidity models.Liquidity) float64 {
	if currentPrice > 0 {
		return currentPrice
	}
	switch {
	case liquidity.BuyLiquidity > 0 && liquidity.SellLiquidity > 0:
		return (liquidity.BuyLiquidity + liquidity.SellLiquidity) / 2
	case liquidity.BuyLiquidity > 0:
		return liquidity.BuyLiquidity
	default:
		return liquidity.SellLiquidity
	}
}

func distanceBps(referencePrice, targetPrice float64) float64 {
	if referencePrice <= 0 || targetPrice <= 0 {
		return 0
	}
	return math.Abs(targetPrice-referencePrice) / referencePrice * 10000
}

func roundServiceFloat(value float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	if pow == 0 {
		return value
	}
	return math.Round(value*pow) / pow
}
