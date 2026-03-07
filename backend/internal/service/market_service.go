package service

import (
	"strings"
	"time"

	"alpha-pulse/backend/internal/collector"
	"alpha-pulse/backend/internal/indicator"
	"alpha-pulse/backend/internal/liquidity"
	"alpha-pulse/backend/internal/orderflow"
	structureengine "alpha-pulse/backend/internal/structure"
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
	"gorm.io/gorm"
)

// MarketService 聚合行情相关能力。
type MarketService struct {
	db              *gorm.DB
	collector       *collector.BinanceCollector
	indicatorEngine *indicator.Engine
	orderFlowEngine *orderflow.Engine
	structureEngine *structureengine.Engine
	liquidityEngine *liquidity.Engine
	klineRepo       *repository.KlineRepository
	aggTradeRepo    *repository.AggTradeRepository
	orderBookRepo   *repository.OrderBookSnapshotRepository
	indicatorRepo   *repository.IndicatorRepository
	microEventRepo  *repository.MicrostructureEventRepository
}

// NewMarketService 创建 MarketService。
func NewMarketService(
	db *gorm.DB,
	collector *collector.BinanceCollector,
	indicatorEngine *indicator.Engine,
	orderFlowEngine *orderflow.Engine,
	structureEngine *structureengine.Engine,
	liquidityEngine *liquidity.Engine,
	klineRepo *repository.KlineRepository,
	aggTradeRepo *repository.AggTradeRepository,
	orderBookRepo *repository.OrderBookSnapshotRepository,
	indicatorRepo *repository.IndicatorRepository,
	microEventRepo *repository.MicrostructureEventRepository,
) *MarketService {
	return &MarketService{
		db:              db,
		collector:       collector,
		indicatorEngine: indicatorEngine,
		orderFlowEngine: orderFlowEngine,
		structureEngine: structureEngine,
		liquidityEngine: liquidityEngine,
		klineRepo:       klineRepo,
		aggTradeRepo:    aggTradeRepo,
		orderBookRepo:   orderBookRepo,
		indicatorRepo:   indicatorRepo,
		microEventRepo:  microEventRepo,
	}
}

// GetPrice 获取实时价格。
func (s *MarketService) GetPrice(symbol string) (map[string]any, error) {
	symbol = normalizeSymbol(symbol)
	price, err := s.collector.GetPrice(symbol)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"symbol": symbol,
		"price":  price,
		"time":   time.Now().UnixMilli(),
	}, nil
}

// GetKline 获取并落库最新 K 线。
func (s *MarketService) GetKline(symbol, interval string) (models.Kline, error) {
	symbol = normalizeSymbol(symbol)
	kline, err := s.collector.GetKline(symbol, interval)
	if err != nil {
		return models.Kline{}, err
	}

	if err := s.klineRepo.Create(&kline); err != nil {
		return models.Kline{}, err
	}

	return kline, nil
}

// GetIndicators 获取并落库技术指标。
func (s *MarketService) GetIndicators(symbol, interval string) (models.Indicator, error) {
	symbol = normalizeSymbol(symbol)
	klines, err := loadAnalysisKlines(s.collector, s.indicatorEngine, s.klineRepo, symbol, interval)
	if err != nil {
		return models.Indicator{}, err
	}

	result, err := s.indicatorEngine.Calculate(symbol, klines)
	if err != nil {
		return models.Indicator{}, err
	}

	if err := s.indicatorRepo.Create(&result); err != nil {
		return models.Indicator{}, err
	}

	return result, nil
}

// GetIndicatorSeries 获取指标时间序列视图。
func (s *MarketService) GetIndicatorSeries(symbol, interval string, limit int) (IndicatorSeriesResult, error) {
	symbol = normalizeSymbol(symbol)
	limit = clampInt(limit, 1, 120)
	requiredLimit := maxInt(limit+s.indicatorEngine.MinimumRequired()-1, s.indicatorEngine.HistoryLimit())

	klines, err := loadAnalysisKlinesWithLimit(
		s.collector,
		s.klineRepo,
		symbol,
		interval,
		requiredLimit,
		s.indicatorEngine.MinimumRequired(),
	)
	if err != nil {
		return IndicatorSeriesResult{}, err
	}

	points, err := buildIndicatorSeries(s.indicatorEngine, symbol, klines, limit)
	if err != nil {
		return IndicatorSeriesResult{}, err
	}

	return IndicatorSeriesResult{
		Symbol:   symbol,
		Interval: interval,
		Points:   points,
	}, nil
}

// GetOrderFlow 获取并落库订单流结果。
func (s *MarketService) GetOrderFlow(symbol, interval string) (models.OrderFlow, error) {
	symbol = normalizeSymbol(symbol)
	result, err := analyzeOrderFlow(
		s.collector,
		s.orderFlowEngine,
		s.klineRepo,
		s.aggTradeRepo,
		symbol,
		interval,
	)
	if err != nil {
		return models.OrderFlow{}, err
	}
	result.IntervalType = interval

	latestKline, latestErr := s.klineRepo.GetLatest(symbol, interval)
	if latestErr != nil {
		freshKline, fetchErr := s.GetKline(symbol, interval)
		if fetchErr == nil {
			latestKline = freshKline
			latestErr = nil
		}
	}
	if latestErr == nil {
		result.IntervalType = interval
		result.OpenTime = latestKline.OpenTime
	}
	if err := s.db.Create(&result).Error; err != nil {
		return models.OrderFlow{}, err
	}
	if err := persistMicrostructureEvents(s.microEventRepo, result); err != nil {
		return models.OrderFlow{}, err
	}
	return result, nil
}

// GetMicrostructureEvents 获取最近的微结构事件历史。
func (s *MarketService) GetMicrostructureEvents(symbol, interval string, limit int) (MicrostructureEventsResult, error) {
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
	limit = clampInt(limit, 1, 80)

	events, err := s.microEventRepo.GetRecentByInterval(symbol, interval, limit)
	if err != nil {
		return MicrostructureEventsResult{}, err
	}
	if len(events) == 0 {
		if _, err := s.GetOrderFlow(symbol, interval); err != nil {
			return MicrostructureEventsResult{}, err
		}
		events, err = s.microEventRepo.GetRecentByInterval(symbol, interval, limit)
		if err != nil {
			return MicrostructureEventsResult{}, err
		}
	}

	return MicrostructureEventsResult{
		Symbol:   symbol,
		Interval: interval,
		Events:   events,
	}, nil
}

// GetStructure 获取并落库市场结构结果。
func (s *MarketService) GetStructure(symbol, interval string) (models.Structure, error) {
	symbol = normalizeSymbol(symbol)
	klines, err := loadAnalysisKlines(s.collector, s.structureEngine, s.klineRepo, symbol, interval)
	if err != nil {
		return models.Structure{}, err
	}

	result, err := s.structureEngine.Analyze(symbol, klines)
	if err != nil {
		return models.Structure{}, err
	}
	if err := s.db.Create(&result).Error; err != nil {
		return models.Structure{}, err
	}
	return result, nil
}

// GetStructureEvents 获取结构事件专用视图。
func (s *MarketService) GetStructureEvents(symbol, interval string) (StructureEventsResult, error) {
	symbol = normalizeSymbol(symbol)
	result, err := s.GetStructure(symbol, interval)
	if err != nil {
		return StructureEventsResult{}, err
	}

	return StructureEventsResult{
		Symbol:     symbol,
		Interval:   interval,
		Trend:      result.Trend,
		Support:    result.Support,
		Resistance: result.Resistance,
		BOS:        result.BOS,
		Choch:      result.Choch,
		Events:     result.Events,
	}, nil
}

// GetStructureSeries 获取结构时间序列视图。
func (s *MarketService) GetStructureSeries(symbol, interval string, limit int) (StructureSeriesResult, error) {
	symbol = normalizeSymbol(symbol)
	limit = clampInt(limit, 1, 120)
	requiredLimit := maxInt(limit+s.structureEngine.MinimumRequired()-1, s.structureEngine.HistoryLimit())

	klines, err := loadAnalysisKlinesWithLimit(
		s.collector,
		s.klineRepo,
		symbol,
		interval,
		requiredLimit,
		s.structureEngine.MinimumRequired(),
	)
	if err != nil {
		return StructureSeriesResult{}, err
	}

	points, err := buildStructureSeries(s.structureEngine, symbol, klines, limit)
	if err != nil {
		return StructureSeriesResult{}, err
	}

	return StructureSeriesResult{
		Symbol:   symbol,
		Interval: interval,
		Points:   points,
	}, nil
}

// GetLiquidity 获取并落库流动性结果。
func (s *MarketService) GetLiquidity(symbol, interval string) (models.Liquidity, error) {
	symbol = normalizeSymbol(symbol)
	result, err := analyzeLiquidity(
		s.collector,
		s.liquidityEngine,
		s.klineRepo,
		s.orderBookRepo,
		symbol,
		interval,
	)
	if err != nil {
		return models.Liquidity{}, err
	}
	if err := s.db.Create(&result).Error; err != nil {
		return models.Liquidity{}, err
	}
	return result, nil
}

// GetLiquidityMap 获取流动性图谱专用视图。
func (s *MarketService) GetLiquidityMap(symbol, interval string) (LiquidityMapResult, error) {
	symbol = normalizeSymbol(symbol)
	result, err := s.GetLiquidity(symbol, interval)
	if err != nil {
		return LiquidityMapResult{}, err
	}

	return LiquidityMapResult{
		Symbol:             symbol,
		Interval:           interval,
		BuyLiquidity:       result.BuyLiquidity,
		SellLiquidity:      result.SellLiquidity,
		SweepType:          result.SweepType,
		OrderBookImbalance: result.OrderBookImbalance,
		DataSource:         result.DataSource,
		EqualHigh:          result.EqualHigh,
		EqualLow:           result.EqualLow,
		StopClusters:       result.StopClusters,
	}, nil
}

// GetLiquiditySeries 获取流动性时间序列视图。
func (s *MarketService) GetLiquiditySeries(symbol, interval string, limit int) (LiquiditySeriesResult, error) {
	symbol = normalizeSymbol(symbol)
	limit = clampInt(limit, 1, 120)
	requiredLimit := maxInt(limit+s.liquidityEngine.MinimumRequired()-1, s.liquidityEngine.HistoryLimit())

	klines, err := loadAnalysisKlinesWithLimit(
		s.collector,
		s.klineRepo,
		symbol,
		interval,
		requiredLimit,
		s.liquidityEngine.MinimumRequired(),
	)
	if err != nil {
		return LiquiditySeriesResult{}, err
	}

	points, err := buildLiquiditySeries(s.liquidityEngine, s.orderBookRepo, symbol, interval, klines, limit)
	if err != nil {
		return LiquiditySeriesResult{}, err
	}

	return LiquiditySeriesResult{
		Symbol:   symbol,
		Interval: interval,
		Points:   points,
	}, nil
}

// WarmupSymbol 用于定时任务预热核心数据。
func (s *MarketService) WarmupSymbol(symbol string) error {
	symbol = normalizeSymbol(symbol)
	if _, err := s.GetKline(symbol, "1m"); err != nil {
		return err
	}
	if s.orderBookRepo != nil {
		if snapshot, err := s.collector.GetDepthSnapshot(symbol, 20); err == nil {
			if err := s.orderBookRepo.Create(&snapshot); err != nil {
				return err
			}
		}
	}
	if _, err := s.GetIndicators(symbol, "1m"); err != nil {
		return err
	}
	if _, err := s.GetOrderFlow(symbol, "1m"); err != nil {
		return err
	}
	if _, err := s.GetStructure(symbol, "1m"); err != nil {
		return err
	}
	if _, err := s.GetLiquidity(symbol, "1m"); err != nil {
		return err
	}
	return nil
}

func normalizeSymbol(symbol string) string {
	if strings.TrimSpace(symbol) == "" {
		return "BTCUSDT"
	}
	return strings.ToUpper(symbol)
}
