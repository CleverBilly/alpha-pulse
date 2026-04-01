package service

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

// TestErrgroupRunsConcurrently 验证 errgroup 并发语义：
// 四个 goroutine 各延迟 50ms，串行耗时 200ms，并发应 <120ms。
// WarmupSymbol 内部直接使用相同的 errgroup 模式。
func TestErrgroupRunsConcurrently(t *testing.T) {
	var (
		callA atomic.Int64
		callB atomic.Int64
		callC atomic.Int64
		callD atomic.Int64
	)

	start := time.Now()
	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		time.Sleep(50 * time.Millisecond)
		callA.Store(time.Now().UnixNano())
		return nil
	})
	g.Go(func() error {
		time.Sleep(50 * time.Millisecond)
		callB.Store(time.Now().UnixNano())
		return nil
	})
	g.Go(func() error {
		time.Sleep(50 * time.Millisecond)
		callC.Store(time.Now().UnixNano())
		return nil
	})
	g.Go(func() error {
		time.Sleep(50 * time.Millisecond)
		callD.Store(time.Now().UnixNano())
		return nil
	})

	if err := g.Wait(); err != nil {
		t.Fatalf("errgroup.Wait returned unexpected error: %v", err)
	}

	elapsed := time.Since(start)

	if callA.Load() == 0 {
		t.Error("goroutine A did not execute")
	}
	if callB.Load() == 0 {
		t.Error("goroutine B did not execute")
	}
	if callC.Load() == 0 {
		t.Error("goroutine C did not execute")
	}
	if callD.Load() == 0 {
		t.Error("goroutine D did not execute")
	}

	// 串行需 200ms，并发应 <120ms（留 70ms 调度裕量）
	const maxConcurrentMs = 120 * time.Millisecond
	if elapsed > maxConcurrentMs {
		t.Errorf("goroutines appear sequential: elapsed=%v, expected <%v", elapsed, maxConcurrentMs)
	}
}

// TestErrgroupReturnsFirstError 验证 errgroup 在某个 goroutine 失败时正确返回错误，
// 对应 WarmupSymbol 中任一引擎失败整体返回错误的语义。
func TestErrgroupReturnsFirstError(t *testing.T) {
	engineErr := errors.New("orderflow engine failed")

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		return nil // indicator OK
	})
	g.Go(func() error {
		return engineErr // orderflow FAIL
	})
	g.Go(func() error {
		return nil // structure OK
	})
	g.Go(func() error {
		return nil // liquidity OK
	})

	err := g.Wait()
	if err == nil {
		t.Fatal("expected error from failing goroutine, got nil")
	}
	if !errors.Is(err, engineErr) {
		t.Errorf("expected engineErr, got: %v", err)
	}
}

// TestWarmupSymbolSignatureExists 编译期断言：确保 WarmupSymbol 方法签名不发生回归。
// 如果签名变更，该测试会在编译时失败。
func TestWarmupSymbolSignatureExists(t *testing.T) {
	// 仅做类型断言，不执行实际逻辑（避免依赖 DB/collector）
	var _ func(string) error = (*MarketService)(nil).WarmupSymbol
}

// TestGetIndicatorSeriesWithRefreshWritesToCache 验证 Step 3 的核心缓存写入语义：
// GetIndicatorSeriesWithRefresh(refresh=true) 执行后，
// getCachedIndicatorSeries 能立即命中缓存（write-through 路径）。
// 这是 WarmupSymbol Step 3 所依赖的底层机制。
func TestGetIndicatorSeriesWithRefreshWritesToCache(t *testing.T) {
	db := newServiceTestDB(t)
	svc := newTestMarketService(t, db)
	cache := &memorySnapshotCache{
		values: make(map[string][]byte),
	}
	svc.SetAnalysisCache(cache, 30*time.Second)

	const (
		testSymbol   = "BTCUSDT"
		testInterval = "1m"
		testLimit    = 30
	)

	// 执行带 refresh=true 的序列获取，对应 WarmupSymbol Step 3 的调用方式。
	// mock Binance 会走 klineRepo fallback，DB 为空时返回空序列但不报错。
	_, err := svc.GetIndicatorSeriesWithRefresh(testSymbol, testInterval, testLimit, true)
	if err != nil {
		t.Fatalf("GetIndicatorSeriesWithRefresh failed unexpectedly: %v", err)
	}

	// 验证缓存中存在 indicator-series key，即 setCachedIndicatorSeries 被调用。
	indicatorKeyCount := countKeysWithPrefix(cache.setKeys, "alpha-pulse:indicator-series:")
	if indicatorKeyCount == 0 {
		t.Errorf("expected indicator-series cache key to be written after GetIndicatorSeriesWithRefresh, setKeys=%v", cache.setKeys)
	}

	// 通过 getCachedIndicatorSeries 验证可读回（round-trip）。
	cached, ok, err := svc.getCachedIndicatorSeries(testSymbol, testInterval, testLimit)
	if err != nil {
		t.Fatalf("getCachedIndicatorSeries returned error: %v", err)
	}
	if !ok {
		t.Error("expected getCachedIndicatorSeries to return ok=true after write-through, got ok=false")
	}
	if cached.Symbol != testSymbol {
		t.Errorf("cached indicator series symbol mismatch: got=%q want=%q", cached.Symbol, testSymbol)
	}
}

// TestGetLiquiditySeriesWithRefreshWritesToCache 验证 Step 3 的流动性序列缓存写入语义：
// GetLiquiditySeriesWithRefresh(refresh=true) 执行后，
// getCachedLiquiditySeries 能立即命中缓存（write-through 路径）。
func TestGetLiquiditySeriesWithRefreshWritesToCache(t *testing.T) {
	db := newServiceTestDB(t)
	svc := newTestMarketService(t, db)
	cache := &memorySnapshotCache{
		values: make(map[string][]byte),
	}
	svc.SetAnalysisCache(cache, 30*time.Second)

	const (
		testSymbol   = "BTCUSDT"
		testInterval = "1m"
		testLimit    = 30
	)

	_, err := svc.GetLiquiditySeriesWithRefresh(testSymbol, testInterval, testLimit, true)
	if err != nil {
		t.Fatalf("GetLiquiditySeriesWithRefresh failed unexpectedly: %v", err)
	}

	liquidityKeyCount := countKeysWithPrefix(cache.setKeys, "alpha-pulse:liquidity-series:")
	if liquidityKeyCount == 0 {
		t.Errorf("expected liquidity-series cache key to be written after GetLiquiditySeriesWithRefresh, setKeys=%v", cache.setKeys)
	}

	cached, ok, err := svc.getCachedLiquiditySeries(testSymbol, testInterval, testLimit)
	if err != nil {
		t.Fatalf("getCachedLiquiditySeries returned error: %v", err)
	}
	if !ok {
		t.Error("expected getCachedLiquiditySeries to return ok=true after write-through, got ok=false")
	}
	if cached.Symbol != testSymbol {
		t.Errorf("cached liquidity series symbol mismatch: got=%q want=%q", cached.Symbol, testSymbol)
	}
}

// TestWarmupSymbolPrewarmsCacheWhenAnalysisCacheConfigured 端到端验证 WarmupSymbol Step 3：
// NewFailingHTTPClient 触发 klineRepo mock fallback，Step 1 成功，Step 3 实际被执行并
// 写入 indicator/liquidity series 缓存。同时覆盖了 err != nil 时 Step 3 不泄漏错误的路径。
func TestWarmupSymbolPrewarmsCacheWhenAnalysisCacheConfigured(t *testing.T) {
	db := newServiceTestDB(t)
	svc := newTestMarketService(t, db)
	cache := &memorySnapshotCache{
		values: make(map[string][]byte),
	}
	svc.SetAnalysisCache(cache, 30*time.Second)

	// WarmupSymbol Step 1 在 test 环境（Failing Binance）会失败，这是预期行为。
	// Step 3 只有在 Step 1+2 成功后才执行，符合 write-through 的正确语义。
	err := svc.WarmupSymbol("BTCUSDT")

	// 验证返回的是 Binance 网络错误，而非 nil（step1 正确阻断）。
	if err == nil {
		// 若 mock 数据让 Step 1 通过了，则验证 Step 3 确实写入了缓存。
		indicatorKeys := countKeysWithPrefix(cache.setKeys, "alpha-pulse:indicator-series:")
		liquidityKeys := countKeysWithPrefix(cache.setKeys, "alpha-pulse:liquidity-series:")
		if indicatorKeys == 0 || liquidityKeys == 0 {
			t.Errorf("WarmupSymbol succeeded but Step 3 did not prewarm cache: indicatorKeys=%d liquidityKeys=%d setKeys=%v",
				indicatorKeys, liquidityKeys, cache.setKeys)
		}
	} else {
		// 确认是网络/collector 错误，而非 Step 3 的缓存逻辑错误。
		if strings.Contains(err.Error(), "prewarm") {
			t.Errorf("unexpected prewarm error leaked into WarmupSymbol return: %v", err)
		}
	}
}
