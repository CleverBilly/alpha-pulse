package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"alpha-pulse/backend/internal/ai"
	"alpha-pulse/backend/internal/collector"
	"alpha-pulse/backend/internal/indicator"
	"alpha-pulse/backend/internal/liquidity"
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
	microEventRepo  *repository.MicrostructureEventRepository
	snapshotCache   MarketSnapshotCache
	snapshotTTL     time.Duration
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
	microEventRepo *repository.MicrostructureEventRepository,
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
		microEventRepo:  microEventRepo,
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
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
	limit = clampInt(limit, 1, 120)

	if cached, ok, err := s.getCachedMarketSnapshot(symbol, interval, limit); err == nil && ok {
		return cached, nil
	}

	snapshot, err := s.buildMarketSnapshot(symbol, interval, limit, true)
	if err != nil {
		return MarketSnapshot{}, err
	}

	_ = s.setCachedMarketSnapshot(symbol, interval, limit, snapshot)
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

	allKlines, err := loadAnalysisKlinesWithLimit(
		s.collector,
		s.klineRepo,
		symbol,
		interval,
		requiredLimit,
		minimumRequired,
	)
	if err != nil {
		return MarketSnapshot{}, err
	}

	latestKline := allKlines[len(allKlines)-1]
	chartKlines := takeLastKlines(allKlines, chartLimit)

	priceValue, err := s.collector.GetPrice(symbol)
	if err != nil {
		priceValue = latestKline.ClosePrice
	}
	price := MarketPrice{
		Symbol: symbol,
		Price:  priceValue,
		Time:   time.Now().UnixMilli(),
	}

	indicatorResult, err := s.indicatorEngine.Calculate(symbol, allKlines)
	if err != nil {
		return MarketSnapshot{}, err
	}
	indicatorSeries, err := buildIndicatorSeries(s.indicatorEngine, symbol, allKlines, chartLimit)
	if err != nil {
		return MarketSnapshot{}, err
	}
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
		return MarketSnapshot{}, err
	}
	orderFlowResult.IntervalType = interval
	orderFlowResult.OpenTime = latestKline.OpenTime
	structureResult, err := s.structureEngine.Analyze(symbol, allKlines)
	if err != nil {
		return MarketSnapshot{}, err
	}
	structureSeries, err := buildStructureSeries(s.structureEngine, symbol, allKlines, chartLimit)
	if err != nil {
		return MarketSnapshot{}, err
	}
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
		return MarketSnapshot{}, err
	}
	liquiditySeries, err := buildLiquiditySeries(
		s.liquidityEngine,
		s.orderBookRepo,
		symbol,
		interval,
		allKlines,
		chartLimit,
	)
	if err != nil {
		return MarketSnapshot{}, err
	}
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

	if persist {
		if err := s.indicatorRepo.Create(&indicatorResult); err != nil {
			return MarketSnapshot{}, err
		}
		if err := s.db.Create(&orderFlowResult).Error; err != nil {
			return MarketSnapshot{}, err
		}
		if err := persistMicrostructureEvents(s.microEventRepo, orderFlowResult); err != nil {
			return MarketSnapshot{}, err
		}
		if err := s.db.Create(&structureResult).Error; err != nil {
			return MarketSnapshot{}, err
		}
		if err := s.db.Create(&liquidityResult).Error; err != nil {
			return MarketSnapshot{}, err
		}
		if err := s.signalRepo.Create(&signalResult); err != nil {
			return MarketSnapshot{}, err
		}
	}

	signalTimeline, err := s.loadSignalTimeline(symbol, interval, chartLimit)
	if err != nil {
		return MarketSnapshot{}, err
	}
	if len(signalTimeline) == 0 {
		signalTimeline = compactSignalTimeline([]models.Signal{signalResult}, chartLimit)
	}
	microstructureEvents, err := s.loadSnapshotMicrostructureEvents(symbol, interval, chartKlines, orderFlowResult)
	if err != nil {
		return MarketSnapshot{}, err
	}

	return MarketSnapshot{
		Price:                price,
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
	}, nil
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

	payload, err := s.snapshotCache.Get(ctx, marketSnapshotCacheKey(symbol, interval, limit))
	if err != nil || len(payload) == 0 {
		return MarketSnapshot{}, false, err
	}

	var snapshot MarketSnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		_ = s.snapshotCache.Delete(ctx, marketSnapshotCacheKey(symbol, interval, limit))
		return MarketSnapshot{}, false, err
	}
	return snapshot, true, nil
}

func (s *SignalService) setCachedMarketSnapshot(symbol, interval string, limit int, snapshot MarketSnapshot) error {
	if s.snapshotCache == nil || s.snapshotTTL <= 0 {
		return nil
	}

	payload, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()
	return s.snapshotCache.Set(ctx, marketSnapshotCacheKey(symbol, interval, limit), payload, s.snapshotTTL)
}

func marketSnapshotCacheKey(symbol, interval string, limit int) string {
	return fmt.Sprintf("alpha-pulse:market-snapshot:v3:%s:%s:%d", symbol, interval, limit)
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
