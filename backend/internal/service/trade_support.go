package service

import (
	"fmt"
	"time"

	"alpha-pulse/backend/internal/collector"
	"alpha-pulse/backend/internal/observability"
	"alpha-pulse/backend/internal/orderflow"
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
)

type tradeWindowEngine interface {
	TradeHistoryLimit() int
	TradeMinimumRequired() int
}

// loadAnalysisAggTrades 优先从交易所补齐聚合成交，再从本地数据库返回有序结果。
func loadAnalysisAggTrades(
	collector *collector.BinanceCollector,
	engine tradeWindowEngine,
	repo *repository.AggTradeRepository,
	symbol string,
) ([]models.AggTrade, error) {
	return loadAnalysisAggTradesWithLimit(
		collector,
		repo,
		symbol,
		engine.TradeHistoryLimit(),
		engine.TradeMinimumRequired(),
	)
}

// loadAnalysisAggTradesWithLimit 允许调用方显式指定聚合成交窗口大小与最小样本数。
func loadAnalysisAggTradesWithLimit(
	collector *collector.BinanceCollector,
	repo *repository.AggTradeRepository,
	symbol string,
	limit, minimumRequired int,
) ([]models.AggTrade, error) {
	fetched, err := collector.GetAggTrades(symbol, limit)
	if err == nil && len(fetched) > 0 {
		if err := repo.CreateBatch(fetched); err != nil {
			return nil, err
		}
	}

	trades, dbErr := repo.GetRecent(symbol, limit)
	if dbErr != nil {
		return nil, dbErr
	}

	if len(trades) >= minimumRequired {
		return trades, nil
	}

	if err != nil {
		return nil, err
	}

	if len(fetched) >= minimumRequired {
		return fetched, nil
	}

	return nil, insufficientAggTradeError(minimumRequired, len(trades))
}

func insufficientAggTradeError(required, actual int) error {
	return aggTradeHistoryError{
		required: required,
		actual:   actual,
	}
}

type aggTradeHistoryError struct {
	required int
	actual   int
}

func (e aggTradeHistoryError) Error() string {
	return fmt.Sprintf("agg trade history is insufficient: required=%d actual=%d", e.required, e.actual)
}

// analyzeOrderFlow 优先使用真实聚合成交分析订单流，不足时自动回退到 K 线估算。
func analyzeOrderFlow(
	collector *collector.BinanceCollector,
	engine *orderflow.Engine,
	klineRepo *repository.KlineRepository,
	aggTradeRepo *repository.AggTradeRepository,
	symbol, interval string,
) (models.OrderFlow, error) {
	return analyzeOrderFlowWithKlineFallback(
		collector,
		engine,
		klineRepo,
		aggTradeRepo,
		symbol,
		interval,
		nil,
	)
}

// analyzeOrderFlowWithKlineFallback 允许调用方传入已加载好的 K 线，避免重复拉取历史窗口。
func analyzeOrderFlowWithKlineFallback(
	collector *collector.BinanceCollector,
	engine *orderflow.Engine,
	klineRepo *repository.KlineRepository,
	aggTradeRepo *repository.AggTradeRepository,
	symbol, interval string,
	preloadedKlines []models.Kline,
) (models.OrderFlow, error) {
	startedAt := time.Now()
	var tradeErr error
	if aggTradeRepo != nil {
		trades, err := loadAnalysisAggTrades(collector, engine, aggTradeRepo, symbol)
		if err == nil {
			result, analyzeErr := engine.AnalyzeAggTrades(symbol, trades)
			if analyzeErr == nil {
				observability.LogDuration(
					"service",
					"orderflow",
					startedAt,
					"ok",
					"",
					observability.String("symbol", symbol),
					observability.String("interval", interval),
					observability.String("source", "agg_trade"),
				)
				return result, nil
			}
			tradeErr = analyzeErr
		} else {
			tradeErr = err
		}
	}

	klines := preloadedKlines
	var klineErr error
	if len(klines) < engine.MinimumRequired() {
		klines, klineErr = loadAnalysisKlines(collector, engine, klineRepo, symbol, interval)
	}
	if klineErr != nil {
		if tradeErr != nil {
			observability.LogDuration(
				"service",
				"orderflow",
				startedAt,
				"error",
				fmt.Sprintf("agg_trade=%v kline=%v", tradeErr, klineErr),
				observability.String("symbol", symbol),
				observability.String("interval", interval),
				observability.String("source", "kline_fallback"),
			)
			return models.OrderFlow{}, fmt.Errorf("agg trade analyze failed: %w; kline fallback failed: %v", tradeErr, klineErr)
		}
		observability.LogDuration(
			"service",
			"orderflow",
			startedAt,
			"error",
			klineErr.Error(),
			observability.String("symbol", symbol),
			observability.String("interval", interval),
			observability.String("source", "kline_fallback"),
		)
		return models.OrderFlow{}, klineErr
	}

	result, err := engine.Analyze(symbol, klines)
	if err != nil {
		if tradeErr != nil {
			observability.LogDuration(
				"service",
				"orderflow",
				startedAt,
				"error",
				fmt.Sprintf("agg_trade=%v kline=%v", tradeErr, err),
				observability.String("symbol", symbol),
				observability.String("interval", interval),
				observability.String("source", "kline_fallback"),
			)
			return models.OrderFlow{}, fmt.Errorf("agg trade analyze failed: %w; kline analyze failed: %v", tradeErr, err)
		}
		observability.LogDuration(
			"service",
			"orderflow",
			startedAt,
			"error",
			err.Error(),
			observability.String("symbol", symbol),
			observability.String("interval", interval),
			observability.String("source", "kline_fallback"),
		)
		return models.OrderFlow{}, err
	}

	reason := ""
	if tradeErr != nil {
		reason = tradeErr.Error()
	}
	observability.LogDuration(
		"service",
		"orderflow",
		startedAt,
		"fallback",
		reason,
		observability.String("symbol", symbol),
		observability.String("interval", interval),
		observability.String("source", "kline_fallback"),
	)
	return result, nil
}
