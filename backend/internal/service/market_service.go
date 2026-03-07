package service

import (
	"context"
	"strings"
	"time"

	"alpha-pulse/backend/internal/collector"
	"alpha-pulse/backend/internal/indicator"
	"alpha-pulse/backend/internal/liquidity"
	"alpha-pulse/backend/internal/observability"
	"alpha-pulse/backend/internal/orderflow"
	structureengine "alpha-pulse/backend/internal/structure"
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
	"gorm.io/gorm"
)

// MarketService 聚合行情相关能力。
type MarketService struct {
	db               *gorm.DB
	collector        *collector.BinanceCollector
	indicatorEngine  *indicator.Engine
	orderFlowEngine  *orderflow.Engine
	structureEngine  *structureengine.Engine
	liquidityEngine  *liquidity.Engine
	klineRepo        *repository.KlineRepository
	aggTradeRepo     *repository.AggTradeRepository
	orderBookRepo    *repository.OrderBookSnapshotRepository
	indicatorRepo    *repository.IndicatorRepository
	microEventRepo   *repository.MicrostructureEventRepository
	analysisCache    MarketSnapshotCache
	analysisCacheTTL time.Duration
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

// SetAnalysisCache 为高频分析视图接口配置缓存。
func (s *MarketService) SetAnalysisCache(cache MarketSnapshotCache, ttl time.Duration) {
	s.analysisCache = cache
	s.analysisCacheTTL = ttl
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
	interval = normalizeInterval(interval)
	kline, err := s.collector.GetKline(symbol, interval)
	if err != nil {
		return models.Kline{}, err
	}

	if err := s.klineRepo.Create(&kline); err != nil {
		return models.Kline{}, err
	}
	invalidateAllSymbolCacheScopes(s.analysisCache, symbol, allCacheScopes()...)

	return kline, nil
}

// GetIndicators 获取并落库技术指标。
func (s *MarketService) GetIndicators(symbol, interval string) (models.Indicator, error) {
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
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
	invalidateAllSymbolCacheScopes(s.analysisCache, symbol, allCacheScopes()...)

	return result, nil
}

// GetIndicatorSeries 获取指标时间序列视图。
func (s *MarketService) GetIndicatorSeries(symbol, interval string, limit int) (IndicatorSeriesResult, error) {
	return s.GetIndicatorSeriesWithRefresh(symbol, interval, limit, false)
}

// GetIndicatorSeriesWithRefresh 获取指标时间序列视图，并可显式绕过缓存。
func (s *MarketService) GetIndicatorSeriesWithRefresh(symbol, interval string, limit int, refresh bool) (IndicatorSeriesResult, error) {
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
	limit = clampInt(limit, 1, 120)

	cacheStartedAt := time.Now()
	if refresh {
		invalidateAllSymbolCacheScopes(s.analysisCache, symbol, allCacheScopes()...)
		logServiceDuration("market_service", "indicator_series.cache_read", symbol, interval, limit, cacheStartedAt, "refresh", "", observability.Bool("refresh", true))
	} else {
		if cached, ok, err := s.getCachedIndicatorSeries(symbol, interval, limit); err == nil && ok {
			logServiceDuration("market_service", "indicator_series.cache_read", symbol, interval, limit, cacheStartedAt, "hit", "", observability.String("source", "cache"))
			return cached, nil
		} else if err != nil {
			logServiceDuration("market_service", "indicator_series.cache_read", symbol, interval, limit, cacheStartedAt, "error", err.Error(), observability.String("source", "cache"))
		} else {
			logServiceDuration("market_service", "indicator_series.cache_read", symbol, interval, limit, cacheStartedAt, "miss", "", observability.String("source", "cache"))
		}
	}

	buildStartedAt := time.Now()
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
		logServiceDuration("market_service", "indicator_series.build", symbol, interval, limit, buildStartedAt, "error", err.Error())
		return IndicatorSeriesResult{}, err
	}

	points, err := buildIndicatorSeries(s.indicatorEngine, symbol, klines, limit)
	if err != nil {
		logServiceDuration("market_service", "indicator_series.build", symbol, interval, limit, buildStartedAt, "error", err.Error())
		return IndicatorSeriesResult{}, err
	}

	result := IndicatorSeriesResult{
		Symbol:   symbol,
		Interval: interval,
		Points:   points,
	}
	if err := s.setCachedIndicatorSeries(symbol, interval, limit, result); err != nil {
		logServiceDuration("market_service", "indicator_series.cache_write", symbol, interval, limit, time.Now(), "error", err.Error())
	}
	logServiceDuration("market_service", "indicator_series.build", symbol, interval, limit, buildStartedAt, "ok", "", observability.Int("points", len(points)))
	return result, nil
}

// GetOrderFlow 获取并落库订单流结果。
func (s *MarketService) GetOrderFlow(symbol, interval string) (models.OrderFlow, error) {
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
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
	if err := enrichOrderFlowMicrostructureWithOrderBook(s.orderFlowEngine, s.orderBookRepo, symbol, &result); err != nil {
		return models.OrderFlow{}, err
	}
	if err := s.db.Create(&result).Error; err != nil {
		return models.OrderFlow{}, err
	}
	if err := persistMicrostructureEvents(s.microEventRepo, result); err != nil {
		return models.OrderFlow{}, err
	}
	invalidateAllSymbolCacheScopes(s.analysisCache, symbol, allCacheScopes()...)
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
	interval = normalizeInterval(interval)
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
	invalidateAllSymbolCacheScopes(s.analysisCache, symbol, allCacheScopes()...)
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
	interval = normalizeInterval(interval)
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
	invalidateAllSymbolCacheScopes(s.analysisCache, symbol, allCacheScopes()...)
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
	return s.GetLiquiditySeriesWithRefresh(symbol, interval, limit, false)
}

// GetLiquiditySeriesWithRefresh 获取流动性时间序列视图，并可显式绕过缓存。
func (s *MarketService) GetLiquiditySeriesWithRefresh(symbol, interval string, limit int, refresh bool) (LiquiditySeriesResult, error) {
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
	limit = clampInt(limit, 1, 120)

	cacheStartedAt := time.Now()
	if refresh {
		invalidateAllSymbolCacheScopes(s.analysisCache, symbol, allCacheScopes()...)
		logServiceDuration("market_service", "liquidity_series.cache_read", symbol, interval, limit, cacheStartedAt, "refresh", "", observability.Bool("refresh", true))
	} else {
		if cached, ok, err := s.getCachedLiquiditySeries(symbol, interval, limit); err == nil && ok {
			logServiceDuration("market_service", "liquidity_series.cache_read", symbol, interval, limit, cacheStartedAt, "hit", "", observability.String("source", "cache"))
			return cached, nil
		} else if err != nil {
			logServiceDuration("market_service", "liquidity_series.cache_read", symbol, interval, limit, cacheStartedAt, "error", err.Error(), observability.String("source", "cache"))
		} else {
			logServiceDuration("market_service", "liquidity_series.cache_read", symbol, interval, limit, cacheStartedAt, "miss", "", observability.String("source", "cache"))
		}
	}

	buildStartedAt := time.Now()
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
		logServiceDuration("market_service", "liquidity_series.build", symbol, interval, limit, buildStartedAt, "error", err.Error())
		return LiquiditySeriesResult{}, err
	}

	points, err := buildLiquiditySeries(s.liquidityEngine, s.orderBookRepo, symbol, interval, klines, limit)
	if err != nil {
		logServiceDuration("market_service", "liquidity_series.build", symbol, interval, limit, buildStartedAt, "error", err.Error())
		return LiquiditySeriesResult{}, err
	}

	result := LiquiditySeriesResult{
		Symbol:   symbol,
		Interval: interval,
		Points:   points,
	}
	if err := s.setCachedLiquiditySeries(symbol, interval, limit, result); err != nil {
		logServiceDuration("market_service", "liquidity_series.cache_write", symbol, interval, limit, time.Now(), "error", err.Error())
	}
	logServiceDuration("market_service", "liquidity_series.build", symbol, interval, limit, buildStartedAt, "ok", "", observability.Int("points", len(points)))
	return result, nil
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

func (s *MarketService) getCachedIndicatorSeries(symbol, interval string, limit int) (IndicatorSeriesResult, bool, error) {
	if s.analysisCache == nil || s.analysisCacheTTL <= 0 {
		return IndicatorSeriesResult{}, false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()
	return getCachedJSON[IndicatorSeriesResult](ctx, s.analysisCache, indicatorSeriesCacheKey(symbol, interval, limit))
}

func (s *MarketService) setCachedIndicatorSeries(symbol, interval string, limit int, result IndicatorSeriesResult) error {
	if s.analysisCache == nil || s.analysisCacheTTL <= 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()
	return setCachedJSON(ctx, s.analysisCache, indicatorSeriesCacheKey(symbol, interval, limit), result, s.analysisCacheTTL)
}

func (s *MarketService) getCachedLiquiditySeries(symbol, interval string, limit int) (LiquiditySeriesResult, bool, error) {
	if s.analysisCache == nil || s.analysisCacheTTL <= 0 {
		return LiquiditySeriesResult{}, false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()
	return getCachedJSON[LiquiditySeriesResult](ctx, s.analysisCache, liquiditySeriesCacheKey(symbol, interval, limit))
}

func (s *MarketService) setCachedLiquiditySeries(symbol, interval string, limit int, result LiquiditySeriesResult) error {
	if s.analysisCache == nil || s.analysisCacheTTL <= 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()
	return setCachedJSON(ctx, s.analysisCache, liquiditySeriesCacheKey(symbol, interval, limit), result, s.analysisCacheTTL)
}
