package service

import (
	"context"
	"log"
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
	"golang.org/x/sync/errgroup"
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
	largeTradeRepo   *repository.LargeTradeEventRepository
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
	largeTradeRepo *repository.LargeTradeEventRepository,
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
		largeTradeRepo:  largeTradeRepo,
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
	// 新 K 线影响所有指标序列和流动性序列，清两个 scope。
	invalidateAllSymbolCacheScopes(s.analysisCache, symbol, cacheScopeIndicatorSeries, cacheScopeLiquiditySeries)

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
	latestKline := klines[len(klines)-1]
	result.IntervalType = interval
	result.OpenTime = latestKline.OpenTime

	if err := s.indicatorRepo.Create(&result); err != nil {
		return models.Indicator{}, err
	}
	// 新指标结果只影响指标序列视图。
	invalidateAllSymbolCacheScopes(s.analysisCache, symbol, cacheScopeIndicatorSeries)

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
		// 仅清当前 interval 的指标序列，不影响其他 interval 或不相关的 scope。
		invalidateCacheScopes(s.analysisCache, symbol, interval, cacheScopeIndicatorSeries)
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
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		// orderflow 必须先写（autoIncrement 回填 ID，后续从表依赖）
		if err := tx.Clauses(repository.OrderFlowUpsertClause()).Create(&result).Error; err != nil {
			return err
		}
		if err := hydrateOrderFlowID(tx, &result); err != nil {
			return err
		}
		// large_trade_events（ON DUPLICATE KEY UPDATE）
		if largeEvents := projectLargeTradeEvents(result); len(largeEvents) > 0 {
			if err := persistLargeTradeEventsTx(tx, largeEvents); err != nil {
				return err
			}
		}
		// microstructure_events（INSERT IGNORE）
		if microEvents := projectMicrostructureEvents(result); len(microEvents) > 0 {
			if err := persistMicrostructureEventsTx(tx, microEvents); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return models.OrderFlow{}, err
	}
	// 订单流结果不影响 indicator-series 或 liquidity-series，无需清 analysisCache。
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
	result.IntervalType = interval
	result.OpenTime = klines[len(klines)-1].OpenTime
	if err := s.db.Clauses(repository.StructureUpsertClause()).Create(&result).Error; err != nil {
		return models.Structure{}, err
	}
	// 结构分析结果不影响 indicator-series 或 liquidity-series，无需清 analysisCache。
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
		Symbol:             symbol,
		Interval:           interval,
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
		Events:             result.Events,
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
	enrichLiquidityDepthContext(
		s.collector,
		s.liquidityEngine,
		s.klineRepo,
		s.orderBookRepo,
		symbol,
		interval,
		0,
		&result,
		nil,
	)
	latestKline, latestErr := s.klineRepo.GetLatest(symbol, interval)
	if latestErr == nil {
		result.IntervalType = interval
		result.OpenTime = latestKline.OpenTime
	}
	if err := s.db.Clauses(repository.LiquidityUpsertClause()).Create(&result).Error; err != nil {
		return models.Liquidity{}, err
	}
	// 新流动性结果只影响流动性序列视图。
	invalidateAllSymbolCacheScopes(s.analysisCache, symbol, cacheScopeLiquiditySeries)
	return result, nil
}

// GetLiquidityMap 获取流动性图谱专用视图。
func (s *MarketService) GetLiquidityMap(symbol, interval string) (LiquidityMapResult, error) {
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
	result, err := s.GetLiquidity(symbol, interval)
	if err != nil {
		return LiquidityMapResult{}, err
	}
	wallLevels := result.WallLevels
	if wallLevels == nil {
		wallLevels = []models.LiquidityWallLevel{}
	}
	wallStrengthBands := result.WallStrengthBands
	if wallStrengthBands == nil {
		wallStrengthBands = []models.LiquidityWallStrengthBand{}
	}
	wallEvolution := result.WallEvolution
	if wallEvolution == nil {
		wallEvolution = []models.LiquidityWallEvolution{}
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
		WallLevels:         wallLevels,
		WallStrengthBands:  wallStrengthBands,
		WallEvolution:      wallEvolution,
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
		// 仅清当前 interval 的流动性序列，不影响其他 interval 或不相关的 scope。
		invalidateCacheScopes(s.analysisCache, symbol, interval, cacheScopeLiquiditySeries)
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

// GetKlineBefore 返回指定时间点之前的 N 根 K 线（升序）。
func (s *MarketService) GetKlineBefore(symbol, interval string, beforeMs int64, limit int) ([]models.Kline, error) {
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
	return s.klineRepo.FindBefore(symbol, interval, beforeMs, limit)
}

// GetKlineAfter 返回指定时间点之后的 N 根 K 线（升序）。
func (s *MarketService) GetKlineAfter(symbol, interval string, afterMs int64, limit int) ([]models.Kline, error) {
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
	return s.klineRepo.FindAfter(symbol, interval, afterMs, limit)
}

// WarmupSymbol 并发预热四个分析引擎（1m 周期）。
func (s *MarketService) WarmupSymbol(symbol string) error {
	symbol = normalizeSymbol(symbol)

	// Step 1: K线和盘口快照必须先落库，后续引擎依赖。
	if _, err := s.GetKline(symbol, "1m"); err != nil {
		return err
	}
	if s.orderBookRepo != nil {
		if snapshot, err := s.collector.GetDepthSnapshot(symbol, 20); err == nil {
			// 盘口快照落库采用 best-effort 策略：失败不阻断四引擎并发预热，
			// 引擎在盘口数据缺失时有降级逻辑。
			_ = s.orderBookRepo.Create(&snapshot)
		}
	}

	// Step 2: 四个分析引擎并发执行，任一失败整体返回错误。
	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		_, err := s.GetIndicators(symbol, "1m")
		return err
	})
	g.Go(func() error {
		_, err := s.GetOrderFlow(symbol, "1m")
		return err
	})
	g.Go(func() error {
		_, err := s.GetStructure(symbol, "1m")
		return err
	})
	g.Go(func() error {
		_, err := s.GetLiquidity(symbol, "1m")
		return err
	})

	if err := g.Wait(); err != nil {
		return err
	}

	// Step 3: 主动预热分析序列缓存（write-through）。
	// 四引擎写库完成后立即重建序列缓存，后续请求 100% 命中，消除 DB 压力峰值（G-05）。
	// 预热失败只记录日志，不影响主流程返回值。
	if s.analysisCache != nil && s.analysisCacheTTL > 0 {
		const prewarmInterval = "1m"
		const prewarmLimit = 30
		if _, err := s.GetIndicatorSeriesWithRefresh(symbol, prewarmInterval, prewarmLimit, true); err != nil {
			log.Printf("prewarm indicator series failed: symbol=%s err=%v", symbol, err)
		}
		if _, err := s.GetLiquiditySeriesWithRefresh(symbol, prewarmInterval, prewarmLimit, true); err != nil {
			log.Printf("prewarm liquidity series failed: symbol=%s err=%v", symbol, err)
		}
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
