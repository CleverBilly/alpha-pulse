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
	fetched, fetchErr := collector.GetKlines(symbol, interval, limit)
	if fetchErr == nil && len(fetched) > 0 {
		if err := repo.CreateBatch(fetched); err != nil {
			return nil, err
		}
		// Binance 返回了完整窗口：数据刚写入 DB，直接用 fetched 省去冗余的 GetRecent 查询。
		// fetched 已按 open_time 升序排列，满足指标引擎要求。
		if len(fetched) >= minimumRequired {
			return fetched, nil
		}
	}

	// Binance 失败或返回不足最小样本：从 DB 读取（含更早的历史数据）。
	klines, dbErr := repo.GetRecent(symbol, interval, limit)
	if dbErr != nil {
		return nil, dbErr
	}

	if len(klines) >= minimumRequired {
		return klines, nil
	}

	if fetchErr != nil {
		return nil, fetchErr
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
