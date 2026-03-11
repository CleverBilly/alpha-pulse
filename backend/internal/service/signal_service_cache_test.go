package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"alpha-pulse/backend/internal/ai"
	"alpha-pulse/backend/internal/collector"
	"alpha-pulse/backend/internal/indicator"
	"alpha-pulse/backend/internal/liquidity"
	"alpha-pulse/backend/internal/orderflow"
	signalengine "alpha-pulse/backend/internal/signal"
	structureengine "alpha-pulse/backend/internal/structure"
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/pkg/binance"
	"alpha-pulse/backend/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetMarketSnapshotUsesCacheAndSkipsDuplicatePersistence(t *testing.T) {
	db := newServiceTestDB(t)
	cache := &memorySnapshotCache{
		values: make(map[string][]byte),
	}

	service := newTestSignalService(t, db, cache, 10*time.Second)

	first, err := service.GetMarketSnapshot("BTCUSDT", "5m", 24)
	if err != nil {
		t.Fatalf("first GetMarketSnapshot failed: %v", err)
	}

	assertServiceCount(t, db, &models.Indicator{}, "symbol = ?", 1, "BTCUSDT")
	assertServiceCount(t, db, &models.OrderFlow{}, "symbol = ?", 1, "BTCUSDT")
	assertServiceCount(t, db, &models.Structure{}, "symbol = ?", 1, "BTCUSDT")
	assertServiceCount(t, db, &models.Liquidity{}, "symbol = ?", 1, "BTCUSDT")
	assertServiceCount(t, db, &models.Signal{}, "symbol = ?", 1, "BTCUSDT")

	second, err := service.GetMarketSnapshot("BTCUSDT", "5m", 24)
	if err != nil {
		t.Fatalf("second GetMarketSnapshot failed: %v", err)
	}

	if cache.setCalls != 1 {
		t.Fatalf("unexpected cache set count: got=%d want=1", cache.setCalls)
	}
	if cache.getCalls < 2 {
		t.Fatalf("cache should be consulted for both calls: got=%d", cache.getCalls)
	}
	if first.Price.Time != second.Price.Time {
		t.Fatalf("cached snapshot should preserve first response timestamp: first=%d second=%d", first.Price.Time, second.Price.Time)
	}

	assertServiceCount(t, db, &models.Indicator{}, "symbol = ?", 1, "BTCUSDT")
	assertServiceCount(t, db, &models.OrderFlow{}, "symbol = ?", 1, "BTCUSDT")
	assertServiceCount(t, db, &models.Structure{}, "symbol = ?", 1, "BTCUSDT")
	assertServiceCount(t, db, &models.Liquidity{}, "symbol = ?", 1, "BTCUSDT")
	assertServiceCount(t, db, &models.Signal{}, "symbol = ?", 1, "BTCUSDT")
}

func TestGetSignalTimelineUsesCache(t *testing.T) {
	db := newServiceTestDB(t)
	snapshotCache := &memorySnapshotCache{
		values: make(map[string][]byte),
	}
	service := newTestSignalService(t, db, nil, 0)
	service.SetViewCache(snapshotCache, 15*time.Second)

	first, err := service.GetSignalTimeline("BTCUSDT", "1m", 16)
	if err != nil {
		t.Fatalf("first GetSignalTimeline failed: %v", err)
	}
	if len(first.Points) == 0 {
		t.Fatal("expected signal timeline points to be generated")
	}

	second, err := service.GetSignalTimeline("BTCUSDT", "1m", 16)
	if err != nil {
		t.Fatalf("second GetSignalTimeline failed: %v", err)
	}
	if len(second.Points) != len(first.Points) {
		t.Fatalf("expected cached timeline to preserve point count: first=%d second=%d", len(first.Points), len(second.Points))
	}
	if countKeysWithPrefix(snapshotCache.setKeys, "alpha-pulse:signal-timeline:") != 1 {
		t.Fatalf("expected one signal timeline cache write, got keys=%v", snapshotCache.setKeys)
	}
	assertServiceCount(t, db, &models.Signal{}, "symbol = ? AND interval_type = ?", 1, "BTCUSDT", "1m")
}

func TestAnalysisSeriesUseCache(t *testing.T) {
	db := newServiceTestDB(t)
	cache := &memorySnapshotCache{
		values: make(map[string][]byte),
	}
	marketService := newTestMarketService(t, db)
	marketService.SetAnalysisCache(cache, 20*time.Second)

	firstIndicator, err := marketService.GetIndicatorSeries("BTCUSDT", "15m", 24)
	if err != nil {
		t.Fatalf("first GetIndicatorSeries failed: %v", err)
	}
	secondIndicator, err := marketService.GetIndicatorSeries("BTCUSDT", "15m", 24)
	if err != nil {
		t.Fatalf("second GetIndicatorSeries failed: %v", err)
	}
	if len(firstIndicator.Points) == 0 || len(secondIndicator.Points) == 0 {
		t.Fatal("expected indicator series points to be present")
	}

	firstLiquidity, err := marketService.GetLiquiditySeries("BTCUSDT", "15m", 24)
	if err != nil {
		t.Fatalf("first GetLiquiditySeries failed: %v", err)
	}
	secondLiquidity, err := marketService.GetLiquiditySeries("BTCUSDT", "15m", 24)
	if err != nil {
		t.Fatalf("second GetLiquiditySeries failed: %v", err)
	}
	if len(firstLiquidity.Points) == 0 || len(secondLiquidity.Points) == 0 {
		t.Fatal("expected liquidity series points to be present")
	}

	if countKeysWithPrefix(cache.setKeys, "alpha-pulse:indicator-series:") != 1 {
		t.Fatalf("expected one indicator-series cache write, got keys=%v", cache.setKeys)
	}
	if countKeysWithPrefix(cache.setKeys, "alpha-pulse:liquidity-series:") != 1 {
		t.Fatalf("expected one liquidity-series cache write, got keys=%v", cache.setKeys)
	}
	if cache.getCalls < 4 {
		t.Fatalf("expected cache lookups for repeated series requests, got=%d", cache.getCalls)
	}
}

func TestInvalidateAllSymbolCacheScopesRemovesSupportedIntervalsOnlyForMatchingSymbol(t *testing.T) {
	cache := &memorySnapshotCache{
		values: map[string][]byte{
			marketSnapshotCacheKey("BTCUSDT", "1m", 24):   []byte("snapshot"),
			marketSnapshotCacheKey("BTCUSDT", "5m", 48):   []byte("snapshot"),
			indicatorSeriesCacheKey("BTCUSDT", "15m", 24): []byte("indicator"),
			liquiditySeriesCacheKey("BTCUSDT", "1h", 24):  []byte("liquidity"),
			signalTimelineCacheKey("BTCUSDT", "4h", 24):   []byte("timeline"),
			marketSnapshotCacheKey("ETHUSDT", "1m", 24):   []byte("other"),
			indicatorSeriesCacheKey("ETHUSDT", "15m", 24): []byte("other"),
			liquiditySeriesCacheKey("ETHUSDT", "1h", 24):  []byte("other"),
			signalTimelineCacheKey("ETHUSDT", "4h", 24):   []byte("other"),
		},
	}

	invalidateAllSymbolCacheScopes(cache, "BTCUSDT", allCacheScopes()...)

	for _, interval := range supportedCacheIntervals {
		if _, exists := cache.values[marketSnapshotCacheKey("BTCUSDT", interval, 24)]; exists && interval == "1m" {
			t.Fatalf("expected BTC market snapshot key for %s to be removed", interval)
		}
	}
	if _, exists := cache.values[indicatorSeriesCacheKey("BTCUSDT", "15m", 24)]; exists {
		t.Fatal("expected BTC indicator series key to be removed")
	}
	if _, exists := cache.values[liquiditySeriesCacheKey("BTCUSDT", "1h", 24)]; exists {
		t.Fatal("expected BTC liquidity series key to be removed")
	}
	if _, exists := cache.values[signalTimelineCacheKey("BTCUSDT", "4h", 24)]; exists {
		t.Fatal("expected BTC signal timeline key to be removed")
	}
	if _, exists := cache.values[marketSnapshotCacheKey("ETHUSDT", "1m", 24)]; !exists {
		t.Fatal("expected ETH market snapshot key to remain")
	}
	if _, exists := cache.values[indicatorSeriesCacheKey("ETHUSDT", "15m", 24)]; !exists {
		t.Fatal("expected ETH indicator series key to remain")
	}
}

func TestGetMarketSnapshotWithRefreshBypassesCacheAndRepopulatesCurrentKey(t *testing.T) {
	db := newServiceTestDB(t)
	cache := &memorySnapshotCache{
		values: make(map[string][]byte),
	}
	service := newTestSignalService(t, db, cache, 10*time.Second)
	service.SetViewCache(cache, 10*time.Second)

	first, err := service.GetMarketSnapshot("BTCUSDT", "5m", 24)
	if err != nil {
		t.Fatalf("initial GetMarketSnapshot failed: %v", err)
	}

	cache.values[indicatorSeriesCacheKey("BTCUSDT", "1m", 24)] = []byte("stale-indicator")
	cache.values[liquiditySeriesCacheKey("BTCUSDT", "15m", 24)] = []byte("stale-liquidity")
	cache.values[signalTimelineCacheKey("BTCUSDT", "4h", 24)] = []byte("stale-timeline")
	cache.values[marketSnapshotCacheKey("ETHUSDT", "5m", 24)] = []byte("other-symbol")

	second, err := service.GetMarketSnapshotWithRefresh("BTCUSDT", "5m", 24, true)
	if err != nil {
		t.Fatalf("refresh GetMarketSnapshot failed: %v", err)
	}

	if second.Price.Time == first.Price.Time {
		t.Fatalf("expected refresh to rebuild snapshot timestamp: first=%d second=%d", first.Price.Time, second.Price.Time)
	}
	if _, exists := cache.values[indicatorSeriesCacheKey("BTCUSDT", "1m", 24)]; exists {
		t.Fatal("expected BTC indicator series cache to be invalidated on refresh")
	}
	if _, exists := cache.values[liquiditySeriesCacheKey("BTCUSDT", "15m", 24)]; exists {
		t.Fatal("expected BTC liquidity series cache to be invalidated on refresh")
	}
	if _, exists := cache.values[signalTimelineCacheKey("BTCUSDT", "4h", 24)]; exists {
		t.Fatal("expected BTC signal timeline cache to be invalidated on refresh")
	}
	if _, exists := cache.values[marketSnapshotCacheKey("ETHUSDT", "5m", 24)]; !exists {
		t.Fatal("expected other symbol cache key to remain after refresh")
	}
	if _, exists := cache.values[marketSnapshotCacheKey("BTCUSDT", "5m", 24)]; !exists {
		t.Fatal("expected refreshed market snapshot key to be repopulated")
	}
}

func TestGetSignalDoesNotEvictSnapshotCache(t *testing.T) {
	db := newServiceTestDB(t)
	cache := &memorySnapshotCache{
		values: make(map[string][]byte),
	}
	svc := newTestSignalService(t, db, cache, 10*time.Second)
	svc.SetViewCache(cache, 10*time.Second)

	// 首次调用写入 snapshotCache
	first, err := svc.GetMarketSnapshot("BTCUSDT", "1m", 24)
	if err != nil {
		t.Fatalf("GetMarketSnapshot failed: %v", err)
	}

	snapshotKey := marketSnapshotCacheKey("BTCUSDT", "1m", 24)
	if _, exists := cache.values[snapshotKey]; !exists {
		t.Fatal("expected snapshot key to exist after GetMarketSnapshot")
	}
	setCalls := cache.setCalls

	// 模拟调度器调用 GetSignal（不通过 GetMarketSnapshotWithRefresh）
	if _, err := svc.GetSignal("BTCUSDT", "1m"); err != nil {
		t.Fatalf("GetSignal failed: %v", err)
	}

	// snapshotCache 不应被 GetSignal 清除
	if _, exists := cache.values[snapshotKey]; !exists {
		t.Fatal("GetSignal must not evict snapshotCache — subsequent frontend requests would always miss")
	}

	// 下一次 GetMarketSnapshot 应命中缓存，不触发新一轮构建和写入
	second, err := svc.GetMarketSnapshot("BTCUSDT", "1m", 24)
	if err != nil {
		t.Fatalf("second GetMarketSnapshot failed: %v", err)
	}
	if first.Price.Time != second.Price.Time {
		t.Fatalf("expected cache hit after GetSignal: Price.Time mismatch first=%d second=%d", first.Price.Time, second.Price.Time)
	}
	if cache.setCalls != setCalls {
		t.Fatalf("expected no additional cache writes after GetSignal: setCalls got=%d want=%d", cache.setCalls, setCalls)
	}
}


func TestGetMarketSnapshotEmitsUnifiedDurationLogs(t *testing.T) {
	db := newServiceTestDB(t)
	cache := &memorySnapshotCache{
		values: make(map[string][]byte),
	}
	service := newTestSignalService(t, db, cache, 10*time.Second)
	service.SetViewCache(cache, 10*time.Second)

	logs := captureLogs(t, func() {
		if _, err := service.GetMarketSnapshotWithRefresh("BTCUSDT", "5m", 24, true); err != nil {
			t.Fatalf("GetMarketSnapshotWithRefresh failed: %v", err)
		}
	})

	for _, token := range []string{
		"component=collector stage=klines",
		"component=service stage=orderflow",
		"source=kline_fallback",
		"component=signal_service stage=market_snapshot.load_klines",
		"component=signal_service stage=market_snapshot.signal_timeline",
		"duration=",
	} {
		if !strings.Contains(logs, token) {
			t.Fatalf("expected log output to contain %q, got logs=%s", token, logs)
		}
	}
}

type memorySnapshotCache struct {
	mu       sync.Mutex
	values   map[string][]byte
	getCalls int
	setCalls int
	delCalls int
	setKeys  []string
	delKeys  []string
}

func (c *memorySnapshotCache) Get(_ context.Context, key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.getCalls++
	value, ok := c.values[key]
	if !ok {
		return nil, nil
	}

	cloned := make([]byte, len(value))
	copy(cloned, value)
	return cloned, nil
}

func (c *memorySnapshotCache) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.setCalls++
	cloned := make([]byte, len(value))
	copy(cloned, value)
	c.values[key] = cloned
	c.setKeys = append(c.setKeys, key)
	return nil
}

func (c *memorySnapshotCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.delCalls++
	c.delKeys = append(c.delKeys, key)
	delete(c.values, key)
	return nil
}

func (c *memorySnapshotCache) DeletePrefix(_ context.Context, prefix string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.delCalls++
	for key := range c.values {
		if strings.HasPrefix(key, prefix) {
			c.delKeys = append(c.delKeys, key)
			delete(c.values, key)
		}
	}
	return nil
}

func newTestSignalService(t *testing.T, db *gorm.DB, cache MarketSnapshotCache, ttl time.Duration) *SignalService {
	t.Helper()

	binanceClient := binance.NewClient("", "", 20*time.Millisecond)
	binanceClient.SetHTTPClient(binance.NewFailingHTTPClient(errors.New("test forces sdk fallback path")))
	binanceCollector := collector.NewBinanceCollector(binanceClient)

	return NewSignalService(
		db,
		binanceCollector,
		indicator.NewEngine(),
		orderflow.NewEngine(),
		structureengine.NewEngine(),
		liquidity.NewEngine(),
		signalengine.NewEngine(),
		ai.NewEngine(),
		repository.NewKlineRepository(db),
		repository.NewAggTradeRepository(db),
		repository.NewOrderBookSnapshotRepository(db),
		repository.NewIndicatorRepository(db),
		repository.NewSignalRepository(db),
		repository.NewLargeTradeEventRepository(db),
		repository.NewMicrostructureEventRepository(db),
		repository.NewFeatureSnapshotRepository(db),
		cache,
		ttl,
	)
}

func newTestMarketService(t *testing.T, db *gorm.DB) *MarketService {
	t.Helper()

	binanceClient := binance.NewClient("", "", 20*time.Millisecond)
	binanceClient.SetHTTPClient(binance.NewFailingHTTPClient(errors.New("test forces sdk fallback path")))
	binanceCollector := collector.NewBinanceCollector(binanceClient)

	return NewMarketService(
		db,
		binanceCollector,
		indicator.NewEngine(),
		orderflow.NewEngine(),
		structureengine.NewEngine(),
		liquidity.NewEngine(),
		repository.NewKlineRepository(db),
		repository.NewAggTradeRepository(db),
		repository.NewOrderBookSnapshotRepository(db),
		repository.NewIndicatorRepository(db),
		repository.NewLargeTradeEventRepository(db),
		repository.NewMicrostructureEventRepository(db),
	)
}

func newServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", sanitizeServiceTestName(t.Name()))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db failed: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	return db
}

func assertServiceCount(t *testing.T, db *gorm.DB, model any, query string, expected int64, args ...any) {
	t.Helper()

	var count int64
	if err := db.Model(model).Where(query, args...).Count(&count).Error; err != nil {
		t.Fatalf("count model failed: %v", err)
	}
	if count != expected {
		t.Fatalf("unexpected count for %T: got=%d want=%d", model, count, expected)
	}
}

func sanitizeServiceTestName(name string) string {
	replacer := strings.NewReplacer("/", "_", " ", "_", "=", "_", "?", "_")
	return replacer.Replace(name)
}

func countKeysWithPrefix(keys []string, prefix string) int {
	count := 0
	for _, key := range keys {
		if strings.HasPrefix(key, prefix) {
			count++
		}
	}
	return count
}

func captureLogs(t *testing.T, fn func()) string {
	t.Helper()

	var buffer bytes.Buffer
	currentWriter := log.Writer()
	currentFlags := log.Flags()
	log.SetOutput(&buffer)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(currentWriter)
		log.SetFlags(currentFlags)
	})

	fn()
	return buffer.String()
}
