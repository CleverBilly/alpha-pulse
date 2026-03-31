package service

import (
	"context"
	"errors"
	"time"

	"alpha-pulse/backend/internal/ai"
	"alpha-pulse/backend/internal/collector"
	"alpha-pulse/backend/internal/indicator"
	"alpha-pulse/backend/internal/liquidity"
	"alpha-pulse/backend/internal/observability"
	"alpha-pulse/backend/internal/orderflow"
	signalengine "alpha-pulse/backend/internal/signal"
	structureengine "alpha-pulse/backend/internal/structure"
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
	"gorm.io/gorm"
)

// MarketPrice 定义快照中的价格数据。
type MarketPrice struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Time   int64   `json:"time"`
}

// MarketSnapshot 定义聚合市场快照。
type MarketSnapshot struct {
	Price                MarketPrice                   `json:"price"`
	Futures              FuturesSnapshot               `json:"futures"`
	Klines               []models.Kline                `json:"klines"`
	Indicator            models.Indicator              `json:"indicator"`
	IndicatorSeries      []models.IndicatorSeriesPoint `json:"indicator_series"`
	OrderFlow            models.OrderFlow              `json:"orderflow"`
	MicrostructureEvents []models.MicrostructureEvent  `json:"microstructure_events"`
	Structure            models.Structure              `json:"structure"`
	StructureSeries      []StructureSeriesPoint        `json:"structure_series"`
	Liquidity            models.Liquidity              `json:"liquidity"`
	LiquiditySeries      []LiquiditySeriesPoint        `json:"liquidity_series"`
	Signal               models.Signal                 `json:"signal"`
	SignalTimeline       []models.SignalTimelinePoint  `json:"signal_timeline"`
}

// SignalResult 定义 /api/signal 返回结构。
type SignalResult struct {
	Signal    models.Signal    `json:"signal"`
	Indicator models.Indicator `json:"indicator"`
	OrderFlow models.OrderFlow `json:"orderflow"`
	Structure models.Structure `json:"structure"`
	Liquidity models.Liquidity `json:"liquidity"`
}

// SignalService 聚合信号生成链路。
type SignalService struct {
	db              *gorm.DB
	collector       *collector.BinanceCollector
	indicatorEngine *indicator.Engine
	orderFlowEngine *orderflow.Engine
	structureEngine *structureengine.Engine
	liquidityEngine *liquidity.Engine
	signalEngine    *signalengine.Engine
	explainEngine   *ai.Engine
	klineRepo       *repository.KlineRepository
	aggTradeRepo    *repository.AggTradeRepository
	orderBookRepo   *repository.OrderBookSnapshotRepository
	indicatorRepo   *repository.IndicatorRepository
	signalRepo      *repository.SignalRepository
	largeTradeRepo  *repository.LargeTradeEventRepository
	microEventRepo  *repository.MicrostructureEventRepository
	featureRepo     *repository.FeatureSnapshotRepository
	snapshotCache   MarketSnapshotCache
	snapshotTTL     time.Duration
	viewCache       MarketSnapshotCache
	viewCacheTTL    time.Duration
}

// NewSignalService 创建 SignalService。
func NewSignalService(
	db *gorm.DB,
	collector *collector.BinanceCollector,
	indicatorEngine *indicator.Engine,
	orderFlowEngine *orderflow.Engine,
	structureEngine *structureengine.Engine,
	liquidityEngine *liquidity.Engine,
	signalEngine *signalengine.Engine,
	explainEngine *ai.Engine,
	klineRepo *repository.KlineRepository,
	aggTradeRepo *repository.AggTradeRepository,
	orderBookRepo *repository.OrderBookSnapshotRepository,
	indicatorRepo *repository.IndicatorRepository,
	signalRepo *repository.SignalRepository,
	largeTradeRepo *repository.LargeTradeEventRepository,
	microEventRepo *repository.MicrostructureEventRepository,
	featureRepo *repository.FeatureSnapshotRepository,
	snapshotCache MarketSnapshotCache,
	snapshotTTL time.Duration,
) *SignalService {
	return &SignalService{
		db:              db,
		collector:       collector,
		indicatorEngine: indicatorEngine,
		orderFlowEngine: orderFlowEngine,
		structureEngine: structureEngine,
		liquidityEngine: liquidityEngine,
		signalEngine:    signalEngine,
		explainEngine:   explainEngine,
		klineRepo:       klineRepo,
		aggTradeRepo:    aggTradeRepo,
		orderBookRepo:   orderBookRepo,
		indicatorRepo:   indicatorRepo,
		signalRepo:      signalRepo,
		largeTradeRepo:  largeTradeRepo,
		microEventRepo:  microEventRepo,
		featureRepo:     featureRepo,
		snapshotCache:   snapshotCache,
		snapshotTTL:     snapshotTTL,
	}
}

// GetSignal 生成并返回完整信号分析结果。
func (s *SignalService) GetSignal(symbol, interval string) (SignalResult, error) {
	snapshot, err := s.buildMarketSnapshot(symbol, interval, 1, true)
	if err != nil {
		return SignalResult{}, err
	}

	return SignalResult{
		Signal:    snapshot.Signal,
		Indicator: snapshot.Indicator,
		OrderFlow: snapshot.OrderFlow,
		Structure: snapshot.Structure,
		Liquidity: snapshot.Liquidity,
	}, nil
}

// GetMarketSnapshot 返回聚合行情快照，供前端一次性拉取。
func (s *SignalService) GetMarketSnapshot(symbol, interval string, limit int) (MarketSnapshot, error) {
	return s.GetMarketSnapshotWithRefresh(symbol, interval, limit, false)
}

// GetMarketSnapshotWithRefresh 返回聚合行情快照，并可显式绕过缓存。
func (s *SignalService) GetMarketSnapshotWithRefresh(symbol, interval string, limit int, refresh bool) (MarketSnapshot, error) {
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
	limit = clampInt(limit, 1, 120)
	cacheStartedAt := time.Now()
	if refresh {
		invalidateAllSymbolCacheScopes(s.snapshotCache, symbol, allCacheScopes()...)
		invalidateAllSymbolCacheScopes(s.viewCache, symbol, allCacheScopes()...)
		logServiceDuration("signal_service", "market_snapshot.cache_read", symbol, interval, limit, cacheStartedAt, "refresh", "", observability.Bool("refresh", true))
	} else {
		if cached, ok, err := s.getCachedMarketSnapshot(symbol, interval, limit); err == nil && ok {
			logServiceDuration("signal_service", "market_snapshot.cache_read", symbol, interval, limit, cacheStartedAt, "hit", "", observability.String("source", "cache"))
			return cached, nil
		} else if err != nil {
			logServiceDuration("signal_service", "market_snapshot.cache_read", symbol, interval, limit, cacheStartedAt, "error", err.Error(), observability.String("source", "cache"))
		} else {
			logServiceDuration("signal_service", "market_snapshot.cache_read", symbol, interval, limit, cacheStartedAt, "miss", "", observability.String("source", "cache"))
		}
	}

	buildStartedAt := time.Now()
	snapshot, err := s.buildMarketSnapshot(symbol, interval, limit, true)
	if err != nil {
		logServiceDuration("signal_service", "market_snapshot.build", symbol, interval, limit, buildStartedAt, "error", err.Error())
		return MarketSnapshot{}, err
	}

	if err := s.setCachedMarketSnapshot(symbol, interval, limit, snapshot); err != nil {
		logServiceDuration("signal_service", "market_snapshot.cache_write", symbol, interval, limit, time.Now(), "error", err.Error())
	}
	logServiceDuration(
		"signal_service",
		"market_snapshot.build",
		symbol,
		interval,
		limit,
		buildStartedAt,
		"ok",
		"",
		observability.Int("klines", len(snapshot.Klines)),
		observability.Int("micro_events", len(snapshot.MicrostructureEvents)),
		observability.Int("signal_points", len(snapshot.SignalTimeline)),
	)
	return snapshot, nil
}

// GetLatestSignal 查询最新信号；若无历史则即时生成一条。
func (s *SignalService) GetLatestSignal(symbol string) (models.Signal, error) {
	symbol = normalizeSymbol(symbol)
	result, err := s.signalRepo.GetLatest(symbol)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fresh, freshErr := s.GetSignal(symbol, "1m")
			if freshErr != nil {
				return models.Signal{}, freshErr
			}
			return fresh.Signal, nil
		}
		return models.Signal{}, err
	}

	result.Explain = s.explainEngine.Explain(result)
	return result, nil
}

func (s *SignalService) buildMarketSnapshot(symbol, interval string, limit int, persist bool) (MarketSnapshot, error) {
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
	chartLimit := clampInt(limit, 1, 120)
	minimumRequired := maxInt(
		s.indicatorEngine.MinimumRequired(),
		s.orderFlowEngine.MinimumRequired(),
		s.structureEngine.MinimumRequired(),
		s.liquidityEngine.MinimumRequired(),
	)
	requiredLimit := maxInt(
		chartLimit+minimumRequired-1,
		s.indicatorEngine.HistoryLimit(),
		s.orderFlowEngine.HistoryLimit(),
		s.structureEngine.HistoryLimit(),
		s.liquidityEngine.HistoryLimit(),
	)

	stageStartedAt := time.Now()
	allKlines, err := loadAnalysisKlinesWithLimit(
		s.collector,
		s.klineRepo,
		symbol,
		interval,
		requiredLimit,
		minimumRequired,
	)
	if err != nil {
		logServiceDuration("signal_service", "market_snapshot.load_klines", symbol, interval, chartLimit, stageStartedAt, "error", err.Error())
		return MarketSnapshot{}, err
	}
	logServiceDuration("signal_service", "market_snapshot.load_klines", symbol, interval, chartLimit, stageStartedAt, "ok", "", observability.Int("samples", len(allKlines)))

	latestKline := allKlines[len(allKlines)-1]
	chartKlines := takeLastKlines(allKlines, chartLimit)

	stageStartedAt = time.Now()
	priceValue, err := s.collector.GetPrice(symbol)
	if err != nil {
		priceValue = latestKline.ClosePrice
	}
	price := MarketPrice{
		Symbol: symbol,
		Price:  priceValue,
		Time:   time.Now().UnixMilli(),
	}
	priceStatus := "ok"
	priceReason := ""
	if err != nil {
		priceStatus = "fallback"
		priceReason = err.Error()
	}
	logServiceDuration("signal_service", "market_snapshot.price", symbol, interval, chartLimit, stageStartedAt, priceStatus, priceReason)

	stageStartedAt = time.Now()
	futuresSnapshot := buildFuturesSnapshot(s.collector, symbol)
	futuresStatus := "ok"
	if !futuresSnapshot.Available {
		futuresStatus = "fallback"
	}
	logServiceDuration(
		"signal_service",
		"market_snapshot.futures",
		symbol,
		interval,
		chartLimit,
		stageStartedAt,
		futuresStatus,
		futuresSnapshot.Reason,
		observability.String("source", futuresSnapshot.Source),
	)

	stageStartedAt = time.Now()
	indicatorResult, err := s.indicatorEngine.Calculate(symbol, allKlines)
	if err != nil {
		logServiceDuration("signal_service", "market_snapshot.indicator", symbol, interval, chartLimit, stageStartedAt, "error", err.Error())
		return MarketSnapshot{}, err
	}
	logServiceDuration("signal_service", "market_snapshot.indicator", symbol, interval, chartLimit, stageStartedAt, "ok", "")

	stageStartedAt = time.Now()
	indicatorSeries, err := buildIndicatorSeries(s.indicatorEngine, symbol, allKlines, chartLimit)
	if err != nil {
		logServiceDuration("signal_service", "market_snapshot.indicator_series", symbol, interval, chartLimit, stageStartedAt, "error", err.Error())
		return MarketSnapshot{}, err
	}
	logServiceDuration("signal_service", "market_snapshot.indicator_series", symbol, interval, chartLimit, stageStartedAt, "ok", "", observability.Int("points", len(indicatorSeries)))

	stageStartedAt = time.Now()
	orderFlowResult, err := analyzeOrderFlowWithKlineFallback(
		s.collector,
		s.orderFlowEngine,
		s.klineRepo,
		s.aggTradeRepo,
		symbol,
		interval,
		allKlines,
	)
	if err != nil {
		logServiceDuration("signal_service", "market_snapshot.orderflow", symbol, interval, chartLimit, stageStartedAt, "error", err.Error())
		return MarketSnapshot{}, err
	}
	orderFlowResult.IntervalType = interval
	orderFlowResult.OpenTime = latestKline.OpenTime
	logServiceDuration("signal_service", "market_snapshot.orderflow", symbol, interval, chartLimit, stageStartedAt, "ok", "", observability.String("source", orderFlowResult.DataSource))

	stageStartedAt = time.Now()
	structureResult, err := s.structureEngine.Analyze(symbol, allKlines)
	if err != nil {
		logServiceDuration("signal_service", "market_snapshot.structure", symbol, interval, chartLimit, stageStartedAt, "error", err.Error())
		return MarketSnapshot{}, err
	}
	logServiceDuration("signal_service", "market_snapshot.structure", symbol, interval, chartLimit, stageStartedAt, "ok", "")

	stageStartedAt = time.Now()
	structureSeries, err := buildStructureSeries(s.structureEngine, symbol, allKlines, chartLimit)
	if err != nil {
		logServiceDuration("signal_service", "market_snapshot.structure_series", symbol, interval, chartLimit, stageStartedAt, "error", err.Error())
		return MarketSnapshot{}, err
	}
	logServiceDuration("signal_service", "market_snapshot.structure_series", symbol, interval, chartLimit, stageStartedAt, "ok", "", observability.Int("points", len(structureSeries)))

	stageStartedAt = time.Now()
	liquidityResult, err := analyzeLiquidityWithKlineFallback(
		s.collector,
		s.liquidityEngine,
		s.klineRepo,
		s.orderBookRepo,
		symbol,
		interval,
		allKlines,
	)
	if err != nil {
		logServiceDuration("signal_service", "market_snapshot.liquidity", symbol, interval, chartLimit, stageStartedAt, "error", err.Error())
		return MarketSnapshot{}, err
	}
	logServiceDuration("signal_service", "market_snapshot.liquidity", symbol, interval, chartLimit, stageStartedAt, "ok", "", observability.String("source", liquidityResult.DataSource))

	stageStartedAt = time.Now()
	liquiditySeries, err := buildLiquiditySeries(
		s.liquidityEngine,
		s.orderBookRepo,
		symbol,
		interval,
		allKlines,
		chartLimit,
	)
	if err != nil {
		logServiceDuration("signal_service", "market_snapshot.liquidity_series", symbol, interval, chartLimit, stageStartedAt, "error", err.Error())
		return MarketSnapshot{}, err
	}
	logServiceDuration("signal_service", "market_snapshot.liquidity_series", symbol, interval, chartLimit, stageStartedAt, "ok", "", observability.Int("points", len(liquiditySeries)))

	stageStartedAt = time.Now()
	enrichLiquidityDepthContext(
		s.collector,
		s.liquidityEngine,
		s.klineRepo,
		s.orderBookRepo,
		symbol,
		interval,
		latestKline.ClosePrice,
		&liquidityResult,
		liquiditySeries,
	)
	logServiceDuration(
		"signal_service",
		"market_snapshot.wall_evolution",
		symbol,
		interval,
		chartLimit,
		stageStartedAt,
		"ok",
		"",
		observability.Int("bands", len(liquidityResult.WallStrengthBands)),
		observability.Int("intervals", len(liquidityResult.WallEvolution)),
	)

	stageStartedAt = time.Now()
	if err := enrichOrderFlowMicrostructureWithOrderBook(
		s.orderFlowEngine,
		s.orderBookRepo,
		symbol,
		&orderFlowResult,
	); err != nil {
		logServiceDuration("signal_service", "market_snapshot.micro_enrich", symbol, interval, chartLimit, stageStartedAt, "error", err.Error())
		return MarketSnapshot{}, err
	}
	logServiceDuration("signal_service", "market_snapshot.micro_enrich", symbol, interval, chartLimit, stageStartedAt, "ok", "", observability.Int("events", len(orderFlowResult.MicrostructureEvents)))

	stageStartedAt = time.Now()
	signalResult := s.signalEngine.Generate(
		symbol,
		latestKline.ClosePrice,
		indicatorResult,
		orderFlowResult,
		structureResult,
		liquidityResult,
	)
	signalResult.IntervalType = interval
	signalResult.OpenTime = latestKline.OpenTime
	signalResult.Explain = s.explainEngine.Explain(signalResult)
	logServiceDuration("signal_service", "market_snapshot.signal", symbol, interval, chartLimit, stageStartedAt, "ok", "", observability.Int("score", signalResult.Score))

	if persist {
		stageStartedAt = time.Now()
		// 将全部写库操作合并为单个事务，固定加锁顺序，消除并发死锁（MySQL 1213）。
		// 死锁时自动重试最多 3 次（指数退避 50/100/200ms）。
		if err := persistSnapshotResults(
			s.db,
			&indicatorResult,
			&orderFlowResult,
			&structureResult,
			&liquidityResult,
			&signalResult,
		); err != nil {
			logServiceDuration("signal_service", "market_snapshot.persist", symbol, interval, chartLimit, stageStartedAt, "error", err.Error())
			return MarketSnapshot{}, err
		}
		// 只清当前 interval 的 signal-timeline：新信号写入后该视图已过期。
		invalidateCacheScopes(s.viewCache, symbol, interval, cacheScopeSignalTimeline)
		logServiceDuration("signal_service", "market_snapshot.persist", symbol, interval, chartLimit, stageStartedAt, "ok", "")
	}

	stageStartedAt = time.Now()
	signalTimeline, err := s.loadSignalTimeline(symbol, interval, chartLimit)
	if err != nil {
		logServiceDuration("signal_service", "market_snapshot.signal_timeline", symbol, interval, chartLimit, stageStartedAt, "error", err.Error())
		return MarketSnapshot{}, err
	}
	if len(signalTimeline) == 0 {
		signalTimeline = compactSignalTimeline([]models.Signal{signalResult}, chartLimit)
	}
	logServiceDuration("signal_service", "market_snapshot.signal_timeline", symbol, interval, chartLimit, stageStartedAt, "ok", "", observability.Int("points", len(signalTimeline)))

	stageStartedAt = time.Now()
	microstructureEvents, err := s.loadSnapshotMicrostructureEvents(symbol, interval, chartKlines, orderFlowResult)
	if err != nil {
		logServiceDuration("signal_service", "market_snapshot.micro_history", symbol, interval, chartLimit, stageStartedAt, "error", err.Error())
		return MarketSnapshot{}, err
	}
	logServiceDuration("signal_service", "market_snapshot.micro_history", symbol, interval, chartLimit, stageStartedAt, "ok", "", observability.Int("events", len(microstructureEvents)))

	snapshot := MarketSnapshot{
		Price:                price,
		Futures:              futuresSnapshot,
		Klines:               chartKlines,
		Indicator:            indicatorResult,
		IndicatorSeries:      indicatorSeries,
		OrderFlow:            orderFlowResult,
		MicrostructureEvents: microstructureEvents,
		Structure:            structureResult,
		StructureSeries:      structureSeries,
		Liquidity:            liquidityResult,
		LiquiditySeries:      liquiditySeries,
		Signal:               signalResult,
		SignalTimeline:       signalTimeline,
	}
	if persist {
		// 将 feature_snapshot 单独 upsert（数据在 snapshot 组装后才完整）。
		stageStartedAt = time.Now()
		if err := persistFeatureSnapshot(s.featureRepo, snapshot); err != nil {
			logServiceDuration("signal_service", "market_snapshot.feature_snapshot", symbol, interval, chartLimit, stageStartedAt, "error", err.Error())
			// feature_snapshot 写失败不阻断响应，仅记录日志。
		} else {
			logServiceDuration("signal_service", "market_snapshot.feature_snapshot", symbol, interval, chartLimit, stageStartedAt, "ok", "", observability.String("version", featureSnapshotVersion))
		}
	}

	return snapshot, nil
}

// SetViewCache 为 signal-timeline 等视图接口配置缓存。
func (s *SignalService) SetViewCache(cache MarketSnapshotCache, ttl time.Duration) {
	s.viewCache = cache
	s.viewCacheTTL = ttl
}

func takeLastKlines(klines []models.Kline, limit int) []models.Kline {
	if len(klines) <= limit {
		return klines
	}
	return klines[len(klines)-limit:]
}

func (s *SignalService) getCachedMarketSnapshot(symbol, interval string, limit int) (MarketSnapshot, bool, error) {
	if s.snapshotCache == nil || s.snapshotTTL <= 0 {
		return MarketSnapshot{}, false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()

	return getCachedJSON[MarketSnapshot](ctx, s.snapshotCache, marketSnapshotCacheKey(symbol, interval, limit))
}

func (s *SignalService) setCachedMarketSnapshot(symbol, interval string, limit int, snapshot MarketSnapshot) error {
	if s.snapshotCache == nil || s.snapshotTTL <= 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()
	return setCachedJSON(ctx, s.snapshotCache, marketSnapshotCacheKey(symbol, interval, limit), snapshot, s.snapshotTTL)
}

func maxInt(values ...int) int {
	maxValue := values[0]
	for _, value := range values[1:] {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func normalizeInterval(interval string) string {
	if interval == "" {
		return "1m"
	}
	return interval
}
