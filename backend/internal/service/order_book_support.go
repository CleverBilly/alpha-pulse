package service

import (
	"errors"
	"fmt"

	"alpha-pulse/backend/internal/collector"
	"alpha-pulse/backend/internal/liquidity"
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
	"gorm.io/gorm"
)

// loadLatestOrderBookSnapshot 优先拉取实时盘口快照，失败时回退本地缓存。
func loadLatestOrderBookSnapshot(
	collector *collector.BinanceCollector,
	repo *repository.OrderBookSnapshotRepository,
	symbol string,
	depthLimit int,
) (models.OrderBookSnapshot, error) {
	if collector == nil {
		return models.OrderBookSnapshot{}, errors.New("binance collector is nil")
	}

	var fetchErr error
	if repo != nil {
		snapshot, err := collector.GetDepthSnapshot(symbol, depthLimit)
		if err == nil {
			if persistErr := repo.Create(&snapshot); persistErr != nil {
				return models.OrderBookSnapshot{}, persistErr
			}
			return snapshot, nil
		}
		fetchErr = err

		cached, cacheErr := repo.GetLatest(symbol)
		if cacheErr == nil {
			return cached, nil
		}
		if !errors.Is(cacheErr, gorm.ErrRecordNotFound) {
			return models.OrderBookSnapshot{}, cacheErr
		}
		return models.OrderBookSnapshot{}, fetchErr
	}

	snapshot, err := collector.GetDepthSnapshot(symbol, depthLimit)
	if err != nil {
		return models.OrderBookSnapshot{}, err
	}
	return snapshot, nil
}

// analyzeLiquidity 优先使用盘口快照识别流动性池，不足时自动回退到 K 线版本。
func analyzeLiquidity(
	collector *collector.BinanceCollector,
	engine *liquidity.Engine,
	klineRepo *repository.KlineRepository,
	orderBookRepo *repository.OrderBookSnapshotRepository,
	symbol, interval string,
) (models.Liquidity, error) {
	return analyzeLiquidityWithKlineFallback(
		collector,
		engine,
		klineRepo,
		orderBookRepo,
		symbol,
		interval,
		nil,
	)
}

// analyzeLiquidityWithKlineFallback 允许调用方传入已加载好的 K 线，避免重复拉取历史窗口。
func analyzeLiquidityWithKlineFallback(
	collector *collector.BinanceCollector,
	engine *liquidity.Engine,
	klineRepo *repository.KlineRepository,
	orderBookRepo *repository.OrderBookSnapshotRepository,
	symbol, interval string,
	preloadedKlines []models.Kline,
) (models.Liquidity, error) {
	klines := preloadedKlines
	var klineErr error
	if len(klines) < engine.MinimumRequired() {
		klines, klineErr = loadAnalysisKlines(collector, engine, klineRepo, symbol, interval)
	}
	if klineErr != nil {
		return models.Liquidity{}, klineErr
	}

	var orderBookErr error
	if orderBookRepo != nil {
		snapshot, err := loadLatestOrderBookSnapshot(collector, orderBookRepo, symbol, engine.OrderBookDepthLimit())
		if err == nil {
			result, analyzeErr := engine.AnalyzeWithOrderBook(symbol, klines, snapshot)
			if analyzeErr == nil {
				return result, nil
			}
			orderBookErr = analyzeErr
		} else {
			orderBookErr = err
		}
	}

	result, err := engine.Analyze(symbol, klines)
	if err != nil {
		if orderBookErr != nil {
			return models.Liquidity{}, fmt.Errorf("order book analyze failed: %w; kline fallback failed: %v", orderBookErr, err)
		}
		return models.Liquidity{}, err
	}

	return result, nil
}
