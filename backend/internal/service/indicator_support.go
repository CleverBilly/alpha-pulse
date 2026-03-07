package service

import (
	"fmt"

	"alpha-pulse/backend/internal/collector"
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
)

type klineWindowEngine interface {
	HistoryLimit() int
	MinimumRequired() int
}

// loadAnalysisKlines 优先从交易所补齐历史 K 线，再从本地数据库返回有序结果。
func loadAnalysisKlines(
	collector *collector.BinanceCollector,
	engine klineWindowEngine,
	repo *repository.KlineRepository,
	symbol, interval string,
) ([]models.Kline, error) {
	return loadAnalysisKlinesWithLimit(
		collector,
		repo,
		symbol,
		interval,
		engine.HistoryLimit(),
		engine.MinimumRequired(),
	)
}

// loadAnalysisKlinesWithLimit 允许调用方显式指定窗口大小与最小所需样本数。
func loadAnalysisKlinesWithLimit(
	collector *collector.BinanceCollector,
	repo *repository.KlineRepository,
	symbol, interval string,
	limit, minimumRequired int,
) ([]models.Kline, error) {
	fetched, err := collector.GetKlines(symbol, interval, limit)
	if err == nil && len(fetched) > 0 {
		if err := repo.CreateBatch(fetched); err != nil {
			return nil, err
		}
	}

	klines, dbErr := repo.GetRecent(symbol, interval, limit)
	if dbErr != nil {
		return nil, dbErr
	}

	if len(klines) >= minimumRequired {
		return klines, nil
	}

	if err != nil {
		return nil, err
	}

	if len(fetched) >= minimumRequired {
		return fetched, nil
	}

	return nil, indicatorInsufficientHistoryError(minimumRequired, len(klines))
}

func indicatorInsufficientHistoryError(required, actual int) error {
	return indicatorHistoryError{
		required: required,
		actual:   actual,
	}
}

type indicatorHistoryError struct {
	required int
	actual   int
}

func (e indicatorHistoryError) Error() string {
	return fmt.Sprintf("indicator history is insufficient: required=%d actual=%d", e.required, e.actual)
}
