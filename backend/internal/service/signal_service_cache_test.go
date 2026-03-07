package service

import (
	"context"
	"errors"
	"fmt"
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

type memorySnapshotCache struct {
	mu       sync.Mutex
	values   map[string][]byte
	getCalls int
	setCalls int
	delCalls int
	setKeys  []string
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
	delete(c.values, key)
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
		repository.NewMicrostructureEventRepository(db),
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
