package service

import (
	"errors"
	"fmt"
	"time"

	"alpha-pulse/backend/internal/collector"
	"alpha-pulse/backend/internal/liquidity"
	"alpha-pulse/backend/internal/observability"
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

	startedAt := time.Now()
	var fetchErr error
	if repo != nil {
		snapshot, err := collector.GetDepthSnapshot(symbol, depthLimit)
		if err == nil {
			if persistErr := repo.Create(&snapshot); persistErr != nil {
				observability.LogDuration(
					"service",
					"orderbook_snapshot",
					startedAt,
					"error",
					persistErr.Error(),
					observability.String("symbol", symbol),
					observability.String("source", "exchange"),
					observability.Int("limit", depthLimit),
				)
				return models.OrderBookSnapshot{}, persistErr
			}
			observability.LogDuration(
				"service",
				"orderbook_snapshot",
				startedAt,
				"ok",
				"",
				observability.String("symbol", symbol),
				observability.String("source", "exchange"),
				observability.Int("limit", depthLimit),
			)
			return snapshot, nil
		}
		fetchErr = err

		cached, cacheErr := repo.GetLatest(symbol)
		if cacheErr == nil {
			observability.LogDuration(
				"service",
				"orderbook_snapshot",
				startedAt,
				"fallback",
				fetchErr.Error(),
				observability.String("symbol", symbol),
				observability.String("source", "repo"),
				observability.Int("limit", depthLimit),
			)
			return cached, nil
		}
		if !errors.Is(cacheErr, gorm.ErrRecordNotFound) {
			observability.LogDuration(
				"service",
				"orderbook_snapshot",
				startedAt,
				"error",
				cacheErr.Error(),
				observability.String("symbol", symbol),
				observability.String("source", "repo"),
				observability.Int("limit", depthLimit),
			)
			return models.OrderBookSnapshot{}, cacheErr
		}
		observability.LogDuration(
			"service",
			"orderbook_snapshot",
			startedAt,
			"error",
			fetchErr.Error(),
			observability.String("symbol", symbol),
			observability.String("source", "exchange"),
			observability.Int("limit", depthLimit),
		)
		return models.OrderBookSnapshot{}, fetchErr
	}

	snapshot, err := collector.GetDepthSnapshot(symbol, depthLimit)
	if err != nil {
		observability.LogDuration(
			"service",
			"orderbook_snapshot",
			startedAt,
			"error",
			err.Error(),
			observability.String("symbol", symbol),
			observability.String("source", "exchange"),
			observability.Int("limit", depthLimit),
		)
		return models.OrderBookSnapshot{}, err
	}
	observability.LogDuration(
		"service",
		"orderbook_snapshot",
		startedAt,
		"ok",
		"",
		observability.String("symbol", symbol),
		observability.String("source", "exchange"),
		observability.Int("limit", depthLimit),
	)
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
	startedAt := time.Now()
	klines := preloadedKlines
	var klineErr error
	if len(klines) < engine.MinimumRequired() {
		klines, klineErr = loadAnalysisKlines(collector, engine, klineRepo, symbol, interval)
	}
	if klineErr != nil {
		observability.LogDuration(
			"service",
			"liquidity",
			startedAt,
			"error",
			klineErr.Error(),
			observability.String("symbol", symbol),
			observability.String("interval", interval),
			observability.String("source", "kline"),
		)
		return models.Liquidity{}, klineErr
	}

	var orderBookErr error
	if orderBookRepo != nil {
		snapshot, err := loadLatestOrderBookSnapshot(collector, orderBookRepo, symbol, engine.OrderBookDepthLimit())
		if err == nil {
			result, analyzeErr := engine.AnalyzeWithOrderBook(symbol, klines, snapshot)
			if analyzeErr == nil {
				observability.LogDuration(
					"service",
					"liquidity",
					startedAt,
					"ok",
					"",
					observability.String("symbol", symbol),
					observability.String("interval", interval),
					observability.String("source", "orderbook"),
				)
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
			observability.LogDuration(
				"service",
				"liquidity",
				startedAt,
				"error",
				fmt.Sprintf("orderbook=%v kline=%v", orderBookErr, err),
				observability.String("symbol", symbol),
				observability.String("interval", interval),
				observability.String("source", "kline_fallback"),
			)
			return models.Liquidity{}, fmt.Errorf("order book analyze failed: %w; kline fallback failed: %v", orderBookErr, err)
		}
		observability.LogDuration(
			"service",
			"liquidity",
			startedAt,
			"error",
			err.Error(),
			observability.String("symbol", symbol),
			observability.String("interval", interval),
			observability.String("source", "kline"),
		)
		return models.Liquidity{}, err
	}

	reason := ""
	if orderBookErr != nil {
		reason = orderBookErr.Error()
	}
	observability.LogDuration(
		"service",
		"liquidity",
		startedAt,
		"fallback",
		reason,
		observability.String("symbol", symbol),
		observability.String("interval", interval),
		observability.String("source", "kline_fallback"),
	)
	return result, nil
}
