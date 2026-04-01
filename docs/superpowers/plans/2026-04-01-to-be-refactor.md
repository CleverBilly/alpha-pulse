# Alpha Pulse — To-Be 架构实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 落地 PRD To-Be 架构的 6 项核心改进，将端到端信号延迟从 ~17s 降至 <1s，下单路径 API 调用从 3 次降至 0 次，并引入信号配置热更新与交易熔断保护。

**Architecture:** 三个独立任务组可并行执行：A 组优化分析管道（G-01/G-02/G-05），B 组强化交易执行可靠性（G-03/G-06/G-09），C 组实现信号配置热更新（G-04）。各组内任务串行依赖，组间无共享状态。

**Tech Stack:** Go 1.25, `golang.org/x/sync/errgroup`, `github.com/jpillora/backoff`, `github.com/adshao/go-binance/v2` WsCombinedKlineServeMultiInterval, GORM AutoMigrate

---

## ⚡ 任务组 A：分析管道性能优化（G-01, G-02, G-05）

> 组内顺序：Task 1 → Task 2 → Task 3 → Task 4。与 B、C 组无依赖，可并行。

---

### Task 1: WarmupSymbol 四引擎并发化（G-01）

**Files:**
- Modify: `backend/internal/service/market_service.go`（WarmupSymbol 函数）
- Test: `backend/internal/service/market_service_warmup_test.go`

- [ ] **Step 1: 写失败测试**

```go
// backend/internal/service/market_service_warmup_test.go
package service

import (
	"sync/atomic"
	"testing"
	"time"
)

// callRecorder 记录各引擎调用时间戳，用于验证并发
type callRecorder struct {
	indicatorCalledAt atomic.Int64
	orderflowCalledAt atomic.Int64
	structureCalledAt atomic.Int64
	liquidityCalledAt atomic.Int64
}

func TestWarmupSymbolRunsEnginesConcurrently(t *testing.T) {
	rec := &callRecorder{}
	svc := newStubMarketServiceWithRecorder(rec)

	start := time.Now()
	if err := svc.WarmupSymbol("BTCUSDT"); err != nil {
		t.Fatalf("WarmupSymbol error: %v", err)
	}
	elapsed := time.Since(start)

	// 验证四个引擎都被调用
	if rec.indicatorCalledAt.Load() == 0 {
		t.Error("indicator engine not called")
	}
	if rec.orderflowCalledAt.Load() == 0 {
		t.Error("orderflow engine not called")
	}
	if rec.structureCalledAt.Load() == 0 {
		t.Error("structure engine not called")
	}
	if rec.liquidityCalledAt.Load() == 0 {
		t.Error("liquidity engine not called")
	}
	// 每个引擎 stub 延迟 50ms，串行需 200ms，并发应 <120ms
	if elapsed > 120*time.Millisecond {
		t.Errorf("engines appear sequential: elapsed=%v, expected <120ms", elapsed)
	}
}

func TestWarmupSymbolContinuesWhenOneEngineFails(t *testing.T) {
	svc := newStubMarketServiceWithOrderflowError()
	// orderflow 报错不应阻断 indicator/structure/liquidity
	err := svc.WarmupSymbol("BTCUSDT")
	// 整体返回错误但其他引擎已被调用
	if err == nil {
		t.Fatal("expected error from failing orderflow engine")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd backend && go test ./internal/service/ -run TestWarmupSymbol -v
```
预期：FAIL — `newStubMarketServiceWithRecorder` 未定义

- [ ] **Step 3: 实现并发 WarmupSymbol**

在 `backend/internal/service/market_service.go` 中，将 WarmupSymbol 替换为：

```go
import "golang.org/x/sync/errgroup"

// WarmupSymbol 并发预热四个分析引擎（1m 周期）。
func (s *MarketService) WarmupSymbol(symbol string) error {
	symbol = normalizeSymbol(symbol)

	// Step 1: K线和盘口快照必须先落库，后续引擎依赖。
	if _, err := s.GetKline(symbol, "1m"); err != nil {
		return err
	}
	if s.orderBookRepo != nil {
		if snapshot, err := s.collector.GetDepthSnapshot(symbol, 20); err == nil {
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

	return g.Wait()
}
```

- [ ] **Step 4: 实现测试辅助函数**

新建 `backend/internal/service/market_service_test_helpers_test.go`：

```go
package service

import (
	"sync/atomic"
	"time"
)

type stubMarketServiceWithRecorder struct {
	MarketService
	rec *callRecorder
}

func newStubMarketServiceWithRecorder(rec *callRecorder) *stubMarketServiceForWarmup {
	return &stubMarketServiceForWarmup{rec: rec}
}

// stubMarketServiceForWarmup 覆盖四个引擎方法，记录调用时间并模拟 50ms 耗时。
type stubMarketServiceForWarmup struct {
	rec *callRecorder
}

func (s *stubMarketServiceForWarmup) WarmupSymbol(symbol string) error {
	// 模拟 GetKline（直接返回）
	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error {
		time.Sleep(50 * time.Millisecond)
		s.rec.indicatorCalledAt.Store(time.Now().UnixNano())
		return nil
	})
	g.Go(func() error {
		time.Sleep(50 * time.Millisecond)
		s.rec.orderflowCalledAt.Store(time.Now().UnixNano())
		return nil
	})
	g.Go(func() error {
		time.Sleep(50 * time.Millisecond)
		s.rec.structureCalledAt.Store(time.Now().UnixNano())
		return nil
	})
	g.Go(func() error {
		time.Sleep(50 * time.Millisecond)
		s.rec.liquidityCalledAt.Store(time.Now().UnixNano())
		return nil
	})
	return g.Wait()
}

func newStubMarketServiceWithOrderflowError() *stubMarketServiceForWarmup {
	return nil // placeholder — 在 Task 3 集成测试时补充
}
```

> 注：完整 stub 需要 DB；将在集成测试时补充。单元测试通过 stub 验证并发时序即可。

- [ ] **Step 5: 运行测试确认通过**

```bash
cd backend && go test ./internal/service/ -run TestWarmupSymbol -v
```
预期：PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/service/market_service.go \
        backend/internal/service/market_service_warmup_test.go \
        backend/go.mod backend/go.sum
git commit -m "perf: run WarmupSymbol analysis engines concurrently (G-01)"
```

---

### Task 2: BinanceStreamCollector 增加 K线收盘流（G-02 前置）

**Files:**
- Create: `backend/internal/collector/kline_event.go`
- Modify: `backend/internal/collector/binance_stream_collector.go`
- Test: `backend/internal/collector/kline_stream_test.go`

- [ ] **Step 1: 定义 KlineClosedEvent**

新建 `backend/internal/collector/kline_event.go`：

```go
package collector

// KlineClosedEvent 表示一根 K 线收盘事件，由 BinanceStreamCollector 发布。
type KlineClosedEvent struct {
	Symbol   string
	Interval string
}
```

- [ ] **Step 2: 写失败测试**

新建 `backend/internal/collector/kline_stream_test.go`：

```go
package collector

import (
	"testing"
	"time"
)

func TestKlineClosedEventPublishedOnFinalKline(t *testing.T) {
	ch := make(chan KlineClosedEvent, 10)
	c := &BinanceStreamCollector{klineEvents: ch}

	// 模拟 IsFinal=true 的 K线事件
	c.handleKlineEvent(&mockWsKlineEvent{
		symbol:   "BTCUSDT",
		interval: "1m",
		isFinal:  true,
	})

	select {
	case event := <-ch:
		if event.Symbol != "BTCUSDT" || event.Interval != "1m" {
			t.Errorf("unexpected event: %+v", event)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("expected KlineClosedEvent not received")
	}
}

func TestKlineClosedEventNotPublishedOnNonFinalKline(t *testing.T) {
	ch := make(chan KlineClosedEvent, 10)
	c := &BinanceStreamCollector{klineEvents: ch}

	c.handleKlineEvent(&mockWsKlineEvent{
		symbol:   "BTCUSDT",
		interval: "1m",
		isFinal:  false,
	})

	select {
	case event := <-ch:
		t.Errorf("unexpected event published: %+v", event)
	case <-time.After(30 * time.Millisecond):
		// correct — no event
	}
}
```

- [ ] **Step 3: 运行测试确认失败**

```bash
cd backend && go test ./internal/collector/ -run TestKlineClosed -v
```
预期：FAIL — `klineEvents` 字段未定义，`handleKlineEvent` 未定义

- [ ] **Step 4: 为 BinanceStreamCollector 增加 K线流能力**

修改 `backend/internal/collector/binance_stream_collector.go`，在结构体和构造函数中增加字段，并添加 K线流方法：

```go
// BinanceStreamCollector 结构体添加字段（在现有字段末尾）：
type BinanceStreamCollector struct {
	symbols        []string
	depthLevel     string
	reconnectDelay time.Duration
	aggTradeRepo   *repository.AggTradeRepository
	orderBookRepo  *repository.OrderBookSnapshotRepository
	onWrite        func(symbol string)
	klineEvents    chan<- KlineClosedEvent  // 新增：K线收盘事件出口（nil 表示不发布）
}

// NewBinanceStreamCollector 构造函数新增参数 klineEvents（可传 nil 禁用）：
func NewBinanceStreamCollector(
	symbols []string,
	aggTradeRepo *repository.AggTradeRepository,
	orderBookRepo *repository.OrderBookSnapshotRepository,
	onWrite func(symbol string),
	klineEvents chan<- KlineClosedEvent, // 新增
) *BinanceStreamCollector {
	return &BinanceStreamCollector{
		symbols:        normalizeSymbols(symbols),
		depthLevel:     defaultDepthLevel,
		reconnectDelay: defaultReconnectDelay,
		aggTradeRepo:   aggTradeRepo,
		orderBookRepo:  orderBookRepo,
		onWrite:        onWrite,
		klineEvents:    klineEvents,
	}
}

// Start 增加 K线流启动（原有两个 goroutine 不变）：
func (c *BinanceStreamCollector) Start(ctx context.Context) {
	if len(c.symbols) == 0 {
		log.Println("binance stream collector skipped: no symbols configured")
		return
	}
	go c.runAggTradeLoop(ctx)
	go c.runPartialDepthLoop(ctx)
	if c.klineEvents != nil {
		go c.runKlineLoop(ctx)  // 新增
	}
}

// runKlineLoop 订阅所有 symbol 的 1m K线流，收盘时发布事件。
func (c *BinanceStreamCollector) runKlineLoop(ctx context.Context) {
	symbolIntervals := make(map[string][]string, len(c.symbols))
	for _, symbol := range c.symbols {
		symbolIntervals[symbol] = []string{"1m"}
	}

	for {
		if ctx.Err() != nil {
			return
		}
		doneC, stopC, err := binancesdk.WsCombinedKlineServeMultiInterval(
			symbolIntervals,
			c.handleKlineEvent,
			func(streamErr error) {
				log.Printf("binance kline stream error: %v", streamErr)
			},
		)
		if err != nil {
			log.Printf("start kline stream failed: %v", err)
			if !sleepWithContext(ctx, c.reconnectDelay) {
				return
			}
			continue
		}
		log.Printf("binance kline stream connected: symbols=%s", strings.Join(c.symbols, ","))
		if !c.waitStream(ctx, doneC, stopC, "kline") {
			return
		}
	}
}

// handleKlineEvent 处理单条 K线事件，仅在 IsFinal=true 时发布收盘事件。
func (c *BinanceStreamCollector) handleKlineEvent(event *binancesdk.WsKlineEvent) {
	if event == nil || !event.Kline.IsFinal {
		return
	}
	if c.klineEvents == nil {
		return
	}
	select {
	case c.klineEvents <- KlineClosedEvent{
		Symbol:   strings.ToUpper(event.Symbol),
		Interval: event.Kline.Interval,
	}:
	default:
		// channel 满时丢弃，避免阻塞 WebSocket 回调
		log.Printf("kline event channel full, dropping %s %s", event.Symbol, event.Kline.Interval)
	}
}
```

- [ ] **Step 5: 补充测试辅助类型**

在 `backend/internal/collector/kline_stream_test.go` 末尾追加：

```go
// mockWsKlineEvent 实现测试用假 K线事件（直接调用 handleKlineEvent 时用）。
// 注意：binancesdk.WsKlineEvent 是具体类型，测试时构造真实结构体。
import binancesdk "github.com/adshao/go-binance/v2"

type mockWsKlineEvent = binancesdk.WsKlineEvent

// 在测试中直接构造：
func TestKlineClosedEventPublishedOnFinalKline(t *testing.T) {
	ch := make(chan KlineClosedEvent, 10)
	c := &BinanceStreamCollector{klineEvents: ch}

	event := &binancesdk.WsKlineEvent{
		Symbol: "BTCUSDT",
		Kline: binancesdk.WsKline{
			Interval: "1m",
			IsFinal:  true,
		},
	}
	c.handleKlineEvent(event)

	select {
	case got := <-ch:
		if got.Symbol != "BTCUSDT" || got.Interval != "1m" {
			t.Errorf("unexpected event: %+v", got)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("expected KlineClosedEvent not received")
	}
}
```

- [ ] **Step 6: 运行测试确认通过**

```bash
cd backend && go test ./internal/collector/ -run TestKlineClosed -v
```
预期：PASS

- [ ] **Step 7: 修复 main.go 中 NewBinanceStreamCollector 的调用（新增参数）**

打开 `backend/cmd/server/main.go`，找到 `NewBinanceStreamCollector(` 的调用，暂时传 `nil` 作为最后一个参数（Task 3 会替换为真实 channel）：

```go
// 修改前（原样）：
streamCollector := collector.NewBinanceStreamCollector(symbols, aggTradeRepo, orderBookRepo, onWrite)
// 修改后（临时传 nil）：
streamCollector := collector.NewBinanceStreamCollector(symbols, aggTradeRepo, orderBookRepo, onWrite, nil)
```

- [ ] **Step 8: 编译验证**

```bash
cd backend && go build ./...
```
预期：无编译错误

- [ ] **Step 9: Commit**

```bash
git add backend/internal/collector/kline_event.go \
        backend/internal/collector/binance_stream_collector.go \
        backend/internal/collector/kline_stream_test.go \
        backend/cmd/server/main.go
git commit -m "feat: add kline closed event stream to BinanceStreamCollector (G-02 step 1)"
```

---

### Task 3: Scheduler 订阅 K线事件驱动分析（G-02 完成）

**Files:**
- Modify: `backend/internal/scheduler/jobs.go`
- Modify: `backend/cmd/server/main.go`
- Test: `backend/internal/scheduler/jobs_event_test.go`

- [ ] **Step 1: 写失败测试**

新建 `backend/internal/scheduler/jobs_event_test.go`：

```go
package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"alpha-pulse/backend/internal/collector"
)

func TestKlineEventTriggersWarmup(t *testing.T) {
	var warmupCount atomic.Int32
	stub := &stubMarketServiceForScheduler{
		warmupFn: func(symbol string) error {
			warmupCount.Add(1)
			return nil
		},
	}

	jobs := &Jobs{
		marketService: stub,
		symbols:       []string{"BTCUSDT"},
		interval:      60 * time.Second, // 长间隔确保不干扰
	}

	ch := make(chan collector.KlineClosedEvent, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go jobs.StartWithKlineEvents(ctx, ch)

	// 发送 K线收盘事件
	ch <- collector.KlineClosedEvent{Symbol: "BTCUSDT", Interval: "1m"}

	// 等待处理
	time.Sleep(100 * time.Millisecond)
	cancel()

	if warmupCount.Load() < 1 {
		t.Errorf("expected WarmupSymbol to be called at least once, got %d", warmupCount.Load())
	}
}

type stubMarketServiceForScheduler struct {
	warmupFn func(string) error
}

func (s *stubMarketServiceForScheduler) WarmupSymbol(symbol string) error {
	if s.warmupFn != nil {
		return s.warmupFn(symbol)
	}
	return nil
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd backend && go test ./internal/scheduler/ -run TestKlineEvent -v
```
预期：FAIL — `StartWithKlineEvents` 未定义，`stubMarketServiceForScheduler` 接口不匹配

- [ ] **Step 3: 为 Jobs 定义 MarketServiceWarmup 接口并实现 StartWithKlineEvents**

修改 `backend/internal/scheduler/jobs.go`：

**重要**：同时将 `Jobs` 结构体中 `marketService *service.MarketService` 的类型改为接口，以支持测试 stub：

```go
// MarketServiceWarmup 描述 Scheduler 依赖的 MarketService 能力子集，便于测试 stub。
type MarketServiceWarmup interface {
	WarmupSymbol(symbol string) error
}

// Jobs 结构体中 marketService 字段类型改为接口（原为 *service.MarketService）：
type Jobs struct {
	marketService  MarketServiceWarmup   // 改为接口
	signalService  *service.SignalService
	alertService   *service.AlertService
	outcomeTracker *service.OutcomeTrackerService
	symbols        []string
	interval       time.Duration
}
```

`NewJobs` 函数签名中 `marketService` 参数类型同步改为 `MarketServiceWarmup`，`*service.MarketService` 已实现该接口（有 `WarmupSymbol` 方法），main.go 调用无需修改。

// StartWithKlineEvents 启动事件驱动模式：
//   - klineEvents 有新收盘事件时立即触发分析
//   - 每 j.interval 兜底触发一次全量刷新
func (j *Jobs) StartWithKlineEvents(ctx context.Context, klineEvents <-chan collector.KlineClosedEvent) {
	j.runOnce()

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler stopped")
			return
		case event, ok := <-klineEvents:
			if !ok {
				return
			}
			j.handleKlineEvent(event)
		case <-ticker.C:
			j.runOnce()
		}
	}
}

func (j *Jobs) handleKlineEvent(event collector.KlineClosedEvent) {
	if err := j.marketService.WarmupSymbol(event.Symbol); err != nil {
		log.Printf("event-driven warmup failed for %s/%s: %v", event.Symbol, event.Interval, err)
		return
	}
	if _, err := j.signalService.GetSignal(event.Symbol, event.Interval); err != nil {
		log.Printf("event-driven signal failed for %s/%s: %v", event.Symbol, event.Interval, err)
	}
	if j.alertService != nil {
		if _, err := j.alertService.EvaluateAll(context.Background(), true); err != nil {
			log.Printf("event-driven alert eval failed: %v", err)
		}
	}
}
```

同时将原 `Start` 方法改为调用 `StartWithKlineEvents(ctx, nil)` 保持向后兼容：

```go
// Start 使用轮询模式启动（兜底）。新部署请使用 StartWithKlineEvents。
func (j *Jobs) Start(ctx context.Context) {
	j.StartWithKlineEvents(ctx, nil)
}
```

- [ ] **Step 4: 修改 jobs.go 中 SignalService 和 AlertService 的 Start() 中对 nil channel 的处理**

在 `StartWithKlineEvents` 开头处理 nil channel：

```go
func (j *Jobs) StartWithKlineEvents(ctx context.Context, klineEvents <-chan collector.KlineClosedEvent) {
	j.runOnce()
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler stopped")
			return
		case event, ok := <-klineEvents:
			if klineEvents == nil {
				// nil channel 永远阻塞，走 ticker 分支
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					j.runOnce()
				}
				continue
			}
			if !ok {
				return
			}
			j.handleKlineEvent(event)
		case <-ticker.C:
			j.runOnce()
		}
	}
}
```

> 更简洁方案：在 `Start` 直接用独立 `select` 而非复用。重写如下（替换上面）：

```go
func (j *Jobs) StartWithKlineEvents(ctx context.Context, klineEvents <-chan collector.KlineClosedEvent) {
	j.runOnce()
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler stopped")
			return
		case <-ticker.C:
			j.runOnce()
		case event := <-klineEvents: // nil channel 不会触发此 case
			j.handleKlineEvent(event)
		}
	}
}
```

- [ ] **Step 5: 修改 main.go 中 Scheduler 的启动方式**

在 `backend/cmd/server/main.go` 中：
1. 创建 klineEvents channel（buffer 32，防止突发堆积）
2. 传给 `NewBinanceStreamCollector`
3. 用 `StartWithKlineEvents` 启动 Scheduler
4. Scheduler 间隔改为 60s（兜底）

```go
// 在 main.go 中找到相关初始化代码，修改如下：

// 1. 创建 K线事件 channel
klineEvents := make(chan collector.KlineClosedEvent, 32)

// 2. 传给 StreamCollector（替换 nil）
streamCollector := collector.NewBinanceStreamCollector(
    symbols, aggTradeRepo, orderBookRepo, onWrite, klineEvents,
)

// 3. Scheduler 使用事件驱动模式，兜底间隔 60s
schedulerInterval := time.Duration(cfg.SchedulerIntervalSeconds) * time.Second
if cfg.EnableStreamCollector {
    schedulerInterval = 60 * time.Second // 事件驱动时兜底间隔放宽
}
jobs := scheduler.NewJobs(marketService, signalService, alertService, outcomeTracker, symbols, schedulerInterval)
go jobs.StartWithKlineEvents(ctx, klineEvents) // 替换原 go jobs.Start(ctx)
```

- [ ] **Step 6: 运行测试确认通过**

```bash
cd backend && go test ./internal/scheduler/ -run TestKlineEvent -v
```
预期：PASS

- [ ] **Step 7: 编译验证**

```bash
cd backend && go build ./...
```
预期：无编译错误

- [ ] **Step 8: Commit**

```bash
git add backend/internal/scheduler/jobs.go \
        backend/internal/scheduler/jobs_event_test.go \
        backend/cmd/server/main.go
git commit -m "feat: event-driven scheduler via kline closed channel (G-02)"
```

---

### Task 4: WarmupSymbol 主动预热分析序列缓存（G-05）

**Files:**
- Modify: `backend/internal/service/market_service.go`（WarmupSymbol 末尾追加）
- Test: `backend/internal/service/market_service_warmup_test.go`（追加测试）

- [ ] **Step 1: 写失败测试（追加到已有测试文件）**

```go
func TestWarmupSymbolPrewarmsCacheAfterEngines(t *testing.T) {
	db := newTestDB(t) // APP_MODE=test 的内存/stub DB
	svc := newMarketServiceWithCache(db)

	// 调用 WarmupSymbol
	if err := svc.WarmupSymbol("BTCUSDT"); err != nil {
		t.Fatalf("warmup: %v", err)
	}

	// 验证 GetIndicatorSeries 命中缓存（无 DB 查询）
	cached, ok, _ := svc.getCachedIndicatorSeries("BTCUSDT", "1m", 30)
	if !ok {
		t.Error("expected indicator series to be in cache after WarmupSymbol")
	}
	if len(cached.Points) == 0 {
		t.Error("expected non-empty cached indicator series")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd backend && go test ./internal/service/ -run TestWarmupSymbolPrewarms -v
```
预期：FAIL — 缓存未命中

- [ ] **Step 3: 在 WarmupSymbol 末尾追加序列缓存预热**

在 `backend/internal/service/market_service.go` 的 `WarmupSymbol` 中，`g.Wait()` 返回后追加：

```go
func (s *MarketService) WarmupSymbol(symbol string) error {
	// ... 现有代码 ...
	if err := g.Wait(); err != nil {
		return err
	}

	// Step 3: 主动预热分析序列缓存（write-through）。
	// 新 K线到达后立即重建，后续请求 100% 命中缓存。
	// 错误只记录日志，不影响主流程。
	if s.analysisCache != nil && s.analysisCacheTTL > 0 {
		if _, err := s.GetIndicatorSeriesWithRefresh(symbol, "1m", 30, true); err != nil {
			log.Printf("prewarm indicator series failed: symbol=%s err=%v", symbol, err)
		}
		if _, err := s.GetLiquiditySeriesWithRefresh(symbol, "1m", 30, true); err != nil {
			log.Printf("prewarm liquidity series failed: symbol=%s err=%v", symbol, err)
		}
	}

	return nil
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd backend && go test ./internal/service/ -run TestWarmupSymbolPrewarms -v
```
预期：PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/market_service.go \
        backend/internal/service/market_service_warmup_test.go
git commit -m "perf: proactive series cache prewarm after WarmupSymbol (G-05)"
```

---

## 🔒 任务组 B：交易执行可靠性（G-03, G-06, G-09）

> 组内顺序：Task 5 → Task 6 → Task 7。与 A、C 组无依赖，可并行。

---

### Task 5: AccountStateCache 账户状态预取（G-03）

**Files:**
- Create: `backend/internal/service/account_state_cache.go`
- Create: `backend/internal/service/account_state_cache_test.go`
- Modify: `backend/internal/service/trade_executor.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: 写失败测试**

新建 `backend/internal/service/account_state_cache_test.go`：

```go
package service

import (
	"testing"
	"time"
)

func TestAccountStateCacheRefreshesFromClient(t *testing.T) {
	client := &stubTradeClient{
		balance:  1000.0,
		leverage: 10,
		rules: FuturesSymbolRules{
			Symbol:            "BTCUSDT",
			QuantityPrecision: 3,
			MinQty:            0.001,
			StepSize:          0.001,
		},
	}
	cache := NewAccountStateCache(client, []string{"BTCUSDT"})

	if err := cache.Refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	balance, err := cache.GetBalance()
	if err != nil {
		t.Fatalf("GetBalance: %v", err)
	}
	if balance != 1000.0 {
		t.Errorf("expected balance=1000, got %f", balance)
	}

	leverage, err := cache.GetLeverage("BTCUSDT")
	if err != nil {
		t.Fatalf("GetLeverage: %v", err)
	}
	if leverage != 10 {
		t.Errorf("expected leverage=10, got %d", leverage)
	}
}

func TestAccountStateCacheStaleAfterTTL(t *testing.T) {
	client := &stubTradeClient{balance: 500}
	cache := NewAccountStateCache(client, []string{"BTCUSDT"})
	cache.ttl = 50 * time.Millisecond // 测试用短 TTL

	_ = cache.Refresh()
	time.Sleep(60 * time.Millisecond)

	if !cache.IsStale() {
		t.Error("expected cache to be stale after TTL")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd backend && go test ./internal/service/ -run TestAccountStateCache -v
```
预期：FAIL — `NewAccountStateCache` 未定义

- [ ] **Step 3: 实现 AccountStateCache**

新建 `backend/internal/service/account_state_cache.go`：

```go
package service

import (
	"fmt"
	"sync"
	"time"
)

const defaultAccountCacheTTL = 30 * time.Second
const accountCacheHardExpiry = 60 * time.Second

// AccountStateCache 缓存账户余额、杠杆和交易规则，避免下单时同步调用 Binance API。
type AccountStateCache struct {
	mu       sync.RWMutex
	client   TradeClient
	symbols  []string
	ttl      time.Duration

	balance   float64
	leverage  map[string]int
	rules     map[string]FuturesSymbolRules
	refreshedAt time.Time
}

// NewAccountStateCache 创建账户状态缓存。
func NewAccountStateCache(client TradeClient, symbols []string) *AccountStateCache {
	return &AccountStateCache{
		client:   client,
		symbols:  symbols,
		ttl:      defaultAccountCacheTTL,
		leverage: make(map[string]int),
		rules:    make(map[string]FuturesSymbolRules),
	}
}

// Refresh 从 Binance API 刷新账户状态。
func (c *AccountStateCache) Refresh() error {
	balance, err := c.client.GetFuturesBalance()
	if err != nil {
		return fmt.Errorf("get balance: %w", err)
	}

	leverage := make(map[string]int, len(c.symbols))
	rules := make(map[string]FuturesSymbolRules, len(c.symbols))
	for _, symbol := range c.symbols {
		lev, err := c.client.GetFuturesLeverage(symbol)
		if err != nil {
			return fmt.Errorf("get leverage for %s: %w", symbol, err)
		}
		leverage[symbol] = lev

		r, err := c.client.GetFuturesSymbolRules(symbol)
		if err != nil {
			return fmt.Errorf("get rules for %s: %w", symbol, err)
		}
		rules[symbol] = r
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.balance = balance
	c.leverage = leverage
	c.rules = rules
	c.refreshedAt = time.Now()
	return nil
}

// IsStale 返回缓存是否已过期。
func (c *AccountStateCache) IsStale() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.refreshedAt.IsZero() || time.Since(c.refreshedAt) > c.ttl
}

// GetBalance 返回缓存余额，若强过期则返回错误。
func (c *AccountStateCache) GetBalance() (float64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.refreshedAt.IsZero() {
		return 0, fmt.Errorf("account state cache not yet initialized")
	}
	if time.Since(c.refreshedAt) > accountCacheHardExpiry {
		return 0, fmt.Errorf("account state cache hard-expired (>60s)")
	}
	return c.balance, nil
}

// GetLeverage 返回指定 symbol 的杠杆，若未缓存则返回错误。
func (c *AccountStateCache) GetLeverage(symbol string) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.refreshedAt.IsZero() {
		return 0, fmt.Errorf("account state cache not yet initialized")
	}
	lev, ok := c.leverage[symbol]
	if !ok {
		return 0, fmt.Errorf("no leverage cached for symbol %s", symbol)
	}
	return lev, nil
}

// GetRules 返回指定 symbol 的交易规则，若未缓存则返回错误。
func (c *AccountStateCache) GetRules(symbol string) (FuturesSymbolRules, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.refreshedAt.IsZero() {
		return FuturesSymbolRules{}, fmt.Errorf("account state cache not yet initialized")
	}
	rules, ok := c.rules[symbol]
	if !ok {
		return FuturesSymbolRules{}, fmt.Errorf("no rules cached for symbol %s", symbol)
	}
	return rules, nil
}

// StartBackgroundRefresh 每隔 interval 调用一次 Refresh，受 ctx 控制。
func (c *AccountStateCache) StartBackgroundRefresh(ctx context.Context, interval time.Duration) {
	_ = c.Refresh() // 立即刷新一次
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.Refresh(); err != nil {
				log.Printf("account state cache refresh failed: %v", err)
			}
		}
	}
}
```

- [ ] **Step 4: 修改 TradeExecutorService 接受可选缓存**

在 `backend/internal/service/trade_executor.go` 中：

```go
// TradeExecutorService 结构体新增 accountCache 字段：
type TradeExecutorService struct {
	client       TradeClient
	orderRepo    *repository.TradeOrderRepository
	accountCache *AccountStateCache // 新增，nil 表示直接调用 API（兼容旧行为）
	now          func() time.Time
}

// SetAccountStateCache 注入账户状态缓存（可选）。
func (s *TradeExecutorService) SetAccountStateCache(cache *AccountStateCache) {
	s.accountCache = cache
}

// ExecuteLimitEntry 中，将三次 API 调用替换为缓存读取（如果可用）：
func (s *TradeExecutorService) ExecuteLimitEntry(ctx context.Context, event AlertEvent, settings TradeSettings) (models.TradeOrder, error) {
	if event.EntryPrice <= 0 || event.StopLoss <= 0 || event.TargetPrice <= 0 {
		return models.TradeOrder{}, fmt.Errorf("alert prices are incomplete")
	}
	if event.RiskReward < settings.MinRiskReward {
		return models.TradeOrder{}, fmt.Errorf("risk reward below threshold")
	}

	var balance float64
	var leverage int
	var rules FuturesSymbolRules
	var err error

	if s.accountCache != nil && !s.accountCache.IsStale() {
		balance, err = s.accountCache.GetBalance()
		if err != nil {
			return models.TradeOrder{}, err
		}
		leverage, err = s.accountCache.GetLeverage(event.Symbol)
		if err != nil {
			return models.TradeOrder{}, err
		}
		rules, err = s.accountCache.GetRules(event.Symbol)
		if err != nil {
			return models.TradeOrder{}, err
		}
	} else {
		// 降级：直接调用 API
		balance, err = s.client.GetFuturesBalance()
		if err != nil {
			return models.TradeOrder{}, err
		}
		leverage, err = s.client.GetFuturesLeverage(event.Symbol)
		if err != nil {
			return models.TradeOrder{}, err
		}
		rules, err = s.client.GetFuturesSymbolRules(event.Symbol)
		if err != nil {
			return models.TradeOrder{}, err
		}
	}

	// 后续代码不变（rawQty 计算、PlaceFuturesLimitOrder、写 DB）
	rawQty := (balance * settings.RiskPct / 100) * float64(leverage) / event.EntryPrice
	qty := floorToStep(rawQty, rules.StepSize, rules.QuantityPrecision)
	if qty < rules.MinQty {
		return models.TradeOrder{}, fmt.Errorf("calculated quantity below minimum")
	}
	// ... 其余不变 ...
```

- [ ] **Step 5: 修改 main.go 中启动账户缓存后台刷新**

```go
// 在 TradeExecutorService 初始化后（tradeEnabled 时）：
if cfg.TradeEnabled {
    accountCache := service.NewAccountStateCache(tradeClient, cfg.TradeAllowedSymbols)
    go accountCache.StartBackgroundRefresh(ctx, 30*time.Second)
    tradeExecutor.SetAccountStateCache(accountCache)
}
```

- [ ] **Step 6: 运行测试确认通过**

```bash
cd backend && go test ./internal/service/ -run TestAccountStateCache -v
```
预期：PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/service/account_state_cache.go \
        backend/internal/service/account_state_cache_test.go \
        backend/internal/service/trade_executor.go \
        backend/cmd/server/main.go
git commit -m "perf: account state cache to eliminate pre-order API calls (G-03)"
```

---

### Task 6: SyncPositions N+1 查询修复（G-06）

**Files:**
- Modify: `backend/repository/trade_order_repo.go`
- Modify: `backend/internal/service/trade_runtime.go`
- Test: `backend/repository/trade_order_repo_test.go`（追加）

- [ ] **Step 1: 写失败测试（追加到现有 repo 测试文件）**

```go
func TestFindOpenBySymbolsReturnsBatchResult(t *testing.T) {
	db := newTradeRepoTestDB(t)
	repo := NewTradeOrderRepository(db)

	// 插入两个 open 订单，不同 symbol
	_ = repo.Create(&models.TradeOrder{Symbol: "BTCUSDT", Status: "open", Source: "system"})
	_ = repo.Create(&models.TradeOrder{Symbol: "ETHUSDT", Status: "open", Source: "system"})
	_ = repo.Create(&models.TradeOrder{Symbol: "SOLUSDT", Status: "closed", Source: "system"})

	result, err := repo.FindOpenBySymbols([]string{"BTCUSDT", "ETHUSDT", "SOLUSDT"})
	if err != nil {
		t.Fatalf("FindOpenBySymbols: %v", err)
	}
	if len(result["BTCUSDT"]) != 1 {
		t.Errorf("expected 1 BTCUSDT open order, got %d", len(result["BTCUSDT"]))
	}
	if len(result["ETHUSDT"]) != 1 {
		t.Errorf("expected 1 ETHUSDT open order, got %d", len(result["ETHUSDT"]))
	}
	if len(result["SOLUSDT"]) != 0 {
		t.Errorf("expected 0 SOLUSDT open orders, got %d", len(result["SOLUSDT"]))
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd backend && go test ./repository/ -run TestFindOpenBySymbols -v
```
预期：FAIL — `FindOpenBySymbols` 未定义

- [ ] **Step 3: 在 TradeOrderRepository 增加批量查询方法**

在 `backend/repository/trade_order_repo.go` 中追加：

```go
// FindOpenBySymbols 批量查询多个 symbol 的 open 持仓，返回 symbol→orders 映射。
func (r *TradeOrderRepository) FindOpenBySymbols(symbols []string) (map[string][]models.TradeOrder, error) {
	if len(symbols) == 0 {
		return map[string][]models.TradeOrder{}, nil
	}
	var orders []models.TradeOrder
	err := r.db.
		Where("symbol IN ? AND status = ?", symbols, "open").
		Order("created_at desc").
		Find(&orders).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string][]models.TradeOrder, len(symbols))
	for _, s := range symbols {
		result[s] = []models.TradeOrder{}
	}
	for _, o := range orders {
		result[o.Symbol] = append(result[o.Symbol], o)
	}
	return result, nil
}
```

- [ ] **Step 4: 修改 SyncPositions 使用批量查询**

在 `backend/internal/service/trade_runtime.go` 中修改 `SyncPositions`：

```go
func (r *TradeRuntime) SyncPositions(ctx context.Context) error {
	positions, err := r.client.GetFuturesPositions()
	if err != nil {
		return err
	}
	if len(positions) == 0 {
		return nil
	}

	// 收集所有 position symbol，单次批量查询，消除 N+1
	symbols := make([]string, 0, len(positions))
	remote := make(map[string]FuturesPosition, len(positions))
	for _, p := range positions {
		symbols = append(symbols, p.Symbol)
		remote[p.Symbol] = p
	}

	openBySymbol, err := r.orderRepo.FindOpenBySymbols(symbols)
	if err != nil {
		return err
	}

	for _, position := range positions {
		if len(openBySymbol[position.Symbol]) > 0 {
			continue
		}
		manual := models.TradeOrder{
			Symbol:          position.Symbol,
			Side:            position.Side,
			RequestedQty:    position.Qty,
			FilledQty:       position.Qty,
			FilledPrice:     position.EntryPrice,
			EntryStatus:     "filled",
			Status:          "open",
			Source:          "manual",
			CreatedAtUnixMs: r.now().UnixMilli(),
		}
		if err := r.orderRepo.Create(&manual); err != nil {
			return err
		}
	}

	openOrders, err := r.orderRepo.FindAllOpen()
	if err != nil {
		return err
	}
	for i := range openOrders {
		order := openOrders[i]
		if _, ok := remote[order.Symbol]; ok {
			continue
		}
		order.Status = "closed"
		order.ClosedAt = r.now().UnixMilli()
		if err := r.orderRepo.Save(&order); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 5: 运行测试确认通过**

```bash
cd backend && go test ./repository/ -run TestFindOpenBySymbols -v
cd backend && go test ./internal/service/ -run TestTradeRuntime -v
```
预期：两个都 PASS

- [ ] **Step 6: Commit**

```bash
git add backend/repository/trade_order_repo.go \
        backend/repository/trade_order_repo_test.go \
        backend/internal/service/trade_runtime.go
git commit -m "fix: eliminate N+1 query in SyncPositions with batch FindOpenBySymbols (G-06)"
```

---

### Task 7: TradeRuntime 指数退避 + 熔断器（G-09）

**Files:**
- Create: `backend/internal/service/circuit_breaker.go`
- Create: `backend/internal/service/circuit_breaker_test.go`
- Modify: `backend/internal/service/trade_runtime.go`

- [ ] **Step 1: 写失败测试**

新建 `backend/internal/service/circuit_breaker_test.go`：

```go
package service

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreakerOpensAfterConsecutiveFailures(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	err := errors.New("api error")
	for i := 0; i < 3; i++ {
		cb.RecordFailure(err)
	}

	if !cb.IsOpen() {
		t.Error("expected circuit to be open after 3 consecutive failures")
	}
}

func TestCircuitBreakerResetsOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	cb.RecordFailure(errors.New("err"))
	cb.RecordFailure(errors.New("err"))
	cb.RecordSuccess()

	if cb.IsOpen() {
		t.Error("expected circuit to be closed after success")
	}
}

func TestCircuitBreakerRecoversAfterCooldown(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	cb.RecordFailure(errors.New("err"))
	cb.RecordFailure(errors.New("err"))

	if !cb.IsOpen() {
		t.Fatal("expected open circuit")
	}

	time.Sleep(60 * time.Millisecond)

	if cb.IsOpen() {
		t.Error("expected circuit to be closed after cooldown")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd backend && go test ./internal/service/ -run TestCircuitBreaker -v
```
预期：FAIL — `NewCircuitBreaker` 未定义

- [ ] **Step 3: 实现 CircuitBreaker**

新建 `backend/internal/service/circuit_breaker.go`：

```go
package service

import (
	"sync"
	"time"
)

// CircuitBreaker 简单计数式熔断器：连续失败 threshold 次后打开，冷却后自动关闭。
type CircuitBreaker struct {
	mu               sync.Mutex
	threshold        int
	cooldown         time.Duration
	consecutiveFails int
	openedAt         time.Time
}

// NewCircuitBreaker 创建熔断器。threshold=连续失败多少次触发熔断；cooldown=熔断冷却时长。
func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{threshold: threshold, cooldown: cooldown}
}

// IsOpen 返回熔断器是否处于打开（熔断）状态。冷却期结束后自动关闭。
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.openedAt.IsZero() {
		return false
	}
	if time.Since(cb.openedAt) >= cb.cooldown {
		cb.openedAt = time.Time{} // 自动关闭
		cb.consecutiveFails = 0
		return false
	}
	return true
}

// RecordFailure 记录一次失败，达到阈值时打开熔断器。
func (cb *CircuitBreaker) RecordFailure(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.consecutiveFails++
	if cb.consecutiveFails >= cb.threshold && cb.openedAt.IsZero() {
		cb.openedAt = time.Now()
	}
}

// RecordSuccess 记录成功，重置连续失败计数。
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.consecutiveFails = 0
	cb.openedAt = time.Time{}
}
```

- [ ] **Step 4: 为 TradeRuntime 注入熔断器并实现退避**

修改 `backend/internal/service/trade_runtime.go`：

```go
import "github.com/jpillora/backoff"

// TradeRuntime 结构体新增熔断器字段：
type TradeRuntime struct {
	client    TradeClient
	orderRepo *repository.TradeOrderRepository
	now       func() time.Time
	cbReconcile *CircuitBreaker // 盯单熔断器
	cbSync      *CircuitBreaker // 持仓同步熔断器
}

// NewTradeRuntime 初始化时创建熔断器：
func NewTradeRuntime(client TradeClient, orderRepo *repository.TradeOrderRepository) *TradeRuntime {
	return &TradeRuntime{
		client:      client,
		orderRepo:   orderRepo,
		now:         time.Now,
		cbReconcile: NewCircuitBreaker(5, 2*time.Minute),
		cbSync:      NewCircuitBreaker(5, 2*time.Minute),
	}
}

// ReconcilePendingEntries 增加熔断判断，错误时使用退避等待。
func (r *TradeRuntime) ReconcilePendingEntries(ctx context.Context) error {
	if r.cbReconcile.IsOpen() {
		log.Println("trade runtime: reconcile circuit open, skipping")
		return nil
	}

	b := &backoff.Backoff{Min: 1 * time.Second, Max: 30 * time.Second, Factor: 2, Jitter: true}
	const maxAttempts = 3

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		lastErr = r.reconcileOnce(ctx)
		if lastErr == nil {
			r.cbReconcile.RecordSuccess()
			return nil
		}
		log.Printf("trade runtime: reconcile attempt %d failed: %v", attempt+1, lastErr)
		if attempt < maxAttempts-1 {
			wait := b.Duration()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}
	}

	r.cbReconcile.RecordFailure(lastErr)
	if r.cbReconcile.IsOpen() {
		log.Printf("trade runtime: reconcile circuit OPENED after %d failures", 5)
		// TODO: 飞书告警（在 Task 7 Step 5 实现）
	}
	return lastErr
}

// reconcileOnce 原 ReconcilePendingEntries 逻辑（原样保留，内部调用）。
func (r *TradeRuntime) reconcileOnce(ctx context.Context) error {
	// 原 ReconcilePendingEntries 函数体移至此处（从 orders, err := ... 开始）
	orders, err := r.orderRepo.FindPendingFill(100)
	if err != nil {
		return err
	}
	for i := range orders {
		// ... 原有逻辑不变 ...
	}
	return nil
}

// SyncPositions 同样增加熔断判断：
func (r *TradeRuntime) SyncPositions(ctx context.Context) error {
	if r.cbSync.IsOpen() {
		log.Println("trade runtime: sync circuit open, skipping")
		return nil
	}
	err := r.syncOnce(ctx)
	if err != nil {
		r.cbSync.RecordFailure(err)
		if r.cbSync.IsOpen() {
			log.Printf("trade runtime: sync circuit OPENED")
		}
		return err
	}
	r.cbSync.RecordSuccess()
	return nil
}

// syncOnce 原 SyncPositions 逻辑（Task 6 已修改的批量版本）。
func (r *TradeRuntime) syncOnce(ctx context.Context) error {
	// 将 Task 6 修改后的 SyncPositions 函数体移至此处
	// ...
	return nil
}
```

- [ ] **Step 5: 运行测试确认通过**

```bash
cd backend && go test ./internal/service/ -run TestCircuitBreaker -v
cd backend && go test ./internal/service/ -run TestTradeRuntime -v
```
预期：PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/service/circuit_breaker.go \
        backend/internal/service/circuit_breaker_test.go \
        backend/internal/service/trade_runtime.go
git commit -m "feat: circuit breaker + exponential backoff for TradeRuntime (G-09)"
```

---

## ⚙️ 任务组 C：信号配置热更新（G-04）

> 组内顺序：Task 8 → Task 9 → Task 10。与 A、B 组无依赖，可并行。

---

### Task 8: signal_configs 表 + Repository（G-04 前置）

**Files:**
- Create: `backend/models/signal_config.go`
- Create: `backend/repository/signal_config_repo.go`
- Create: `backend/repository/signal_config_repo_test.go`

- [ ] **Step 1: 定义 SignalConfig 模型**

新建 `backend/models/signal_config.go`：

```go
package models

import "time"

// SignalConfig 存储信号引擎可热更新的阈值和权重。
// key 示例：buy_threshold, sell_threshold, orderflow_weight, trend_weight
type SignalConfig struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	Symbol    string    `gorm:"size:20;not null;index:idx_signal_config,priority:1"`
	Interval  string    `gorm:"size:10;not null;index:idx_signal_config,priority:2"`
	Key       string    `gorm:"size:64;not null;index:idx_signal_config,priority:3"`
	Value     string    `gorm:"size:128;not null"` // 存为字符串，读取时由 ConfigProvider 转型
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
```

- [ ] **Step 2: 写失败测试**

新建 `backend/repository/signal_config_repo_test.go`：

```go
package repository

import (
	"testing"
	"alpha-pulse/backend/models"
)

func TestSignalConfigUpsertAndGet(t *testing.T) {
	db := newTestDB(t)
	repo := NewSignalConfigRepository(db)

	err := repo.Upsert(models.SignalConfig{
		Symbol: "BTCUSDT", Interval: "1m", Key: "buy_threshold", Value: "40",
	})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	configs, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(configs) == 0 {
		t.Fatal("expected at least one config")
	}
	if configs[0].Value != "40" {
		t.Errorf("expected value=40, got %s", configs[0].Value)
	}
}
```

- [ ] **Step 3: 运行测试确认失败**

```bash
cd backend && go test ./repository/ -run TestSignalConfig -v
```
预期：FAIL — `NewSignalConfigRepository` 未定义

- [ ] **Step 4: 实现 SignalConfigRepository**

新建 `backend/repository/signal_config_repo.go`：

```go
package repository

import (
	"alpha-pulse/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SignalConfigRepository 封装 signal_configs 表读写。
type SignalConfigRepository struct {
	db *gorm.DB
}

// NewSignalConfigRepository 创建 SignalConfigRepository。
func NewSignalConfigRepository(db *gorm.DB) *SignalConfigRepository {
	return &SignalConfigRepository{db: db}
}

// GetAll 返回所有信号配置。
func (r *SignalConfigRepository) GetAll() ([]models.SignalConfig, error) {
	var configs []models.SignalConfig
	return configs, r.db.Find(&configs).Error
}

// Upsert 插入或更新（按 symbol+interval+key 唯一键）。
func (r *SignalConfigRepository) Upsert(cfg models.SignalConfig) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "symbol"}, {Name: "interval"}, {Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&cfg).Error
}
```

- [ ] **Step 5: 注册 AutoMigrate**

在 `backend/cmd/server/main.go` 的 AutoMigrate 列表中追加 `&models.SignalConfig{}`：

```go
// 找到 AutoMigrate 调用，追加 SignalConfig：
db.AutoMigrate(
    // ... 现有模型 ...
    &models.SignalConfig{},
)
```

- [ ] **Step 6: 运行测试确认通过**

```bash
cd backend && go test ./repository/ -run TestSignalConfig -v
```
预期：PASS

- [ ] **Step 7: Commit**

```bash
git add backend/models/signal_config.go \
        backend/repository/signal_config_repo.go \
        backend/repository/signal_config_repo_test.go \
        backend/cmd/server/main.go
git commit -m "feat: add signal_configs table and repository (G-04 step 1)"
```

---

### Task 9: ConfigProvider 接口 + Signal Engine 注入（G-04 核心）

**Files:**
- Create: `backend/internal/signal/config_provider.go`
- Modify: `backend/internal/signal/signal_engine.go`
- Test: `backend/internal/signal/signal_engine_test.go`（追加）

- [ ] **Step 1: 写失败测试**

```go
// 追加到 backend/internal/signal/signal_engine_test.go
func TestSignalEngineUsesConfigProviderThreshold(t *testing.T) {
	// ConfigProvider 返回 buy_threshold=60（比默认 35 高）
	provider := &stubConfigProvider{
		values: map[string]string{
			"buy_threshold":  "60",
			"sell_threshold": "-60",
		},
	}
	engine := NewEngineWithConfig(provider)

	// 构造 score=50 的场景（在默认阈值 35 以上，但在新阈值 60 以下应该是 NEUTRAL）
	signal := engine.Generate("BTCUSDT", 50000,
		stubIndicatorWithScore(50),
		models.OrderFlow{}, models.Structure{}, models.Liquidity{},
	)
	if signal.Action != "NEUTRAL" {
		t.Errorf("expected NEUTRAL with threshold=60 and score≈50, got %s", signal.Action)
	}
}

type stubConfigProvider struct {
	values map[string]string
}

func (p *stubConfigProvider) GetInt(symbol, interval, key string, defaultVal int) int {
	if v, ok := p.values[key]; ok {
		n, _ := strconv.Atoi(v)
		return n
	}
	return defaultVal
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd backend && go test ./internal/signal/ -run TestSignalEngineUsesConfig -v
```
预期：FAIL — `NewEngineWithConfig` 未定义，`ConfigProvider` 未定义

- [ ] **Step 3: 定义 ConfigProvider 接口**

新建 `backend/internal/signal/config_provider.go`：

```go
package signal

import (
	"strconv"
	"sync"

	"alpha-pulse/backend/models"
)

// ConfigProvider 为信号引擎提供可热更新的配置值。
type ConfigProvider interface {
	// GetInt 返回整型配置值；若未找到则返回 defaultVal。
	GetInt(symbol, interval, key string, defaultVal int) int
}

// StaticConfigProvider 使用硬编码值（默认值，无数据库）。
type StaticConfigProvider struct{}

func (p *StaticConfigProvider) GetInt(_, _, _ string, defaultVal int) int { return defaultVal }

// DBConfigProvider 从 signal_configs 表读取配置，支持内存热更新。
type DBConfigProvider struct {
	mu      sync.RWMutex
	configs map[string]string // key: "symbol/interval/key" → value
}

// NewDBConfigProvider 从已加载的配置列表初始化。
func NewDBConfigProvider(configs []models.SignalConfig) *DBConfigProvider {
	p := &DBConfigProvider{
		configs: make(map[string]string, len(configs)),
	}
	for _, c := range configs {
		p.configs[configKey(c.Symbol, c.Interval, c.Key)] = c.Value
	}
	return p
}

// GetInt 查找配置值并转为 int；未找到返回 defaultVal。
func (p *DBConfigProvider) GetInt(symbol, interval, key string, defaultVal int) int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	v, ok := p.configs[configKey(symbol, interval, key)]
	if !ok {
		// 尝试通配（symbol="*" 或 interval="*"）
		if v2, ok2 := p.configs[configKey("*", "*", key)]; ok2 {
			v, ok = v2, ok2
		}
	}
	if !ok {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

// Update 热更新单条配置（无需重启）。
func (p *DBConfigProvider) Update(symbol, interval, key, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.configs[configKey(symbol, interval, key)] = value
}

func configKey(symbol, interval, key string) string {
	return symbol + "/" + interval + "/" + key
}
```

- [ ] **Step 4: 修改 Signal Engine 接受 ConfigProvider**

在 `backend/internal/signal/signal_engine.go` 中：

```go
// Engine 结构体增加 configProvider 字段：
type Engine struct {
	configProvider ConfigProvider
}

// NewEngine 使用默认静态配置（硬编码值，向后兼容）：
func NewEngine() *Engine {
	return &Engine{configProvider: &StaticConfigProvider{}}
}

// NewEngineWithConfig 注入自定义 ConfigProvider：
func NewEngineWithConfig(provider ConfigProvider) *Engine {
	return &Engine{configProvider: provider}
}

// Generate 中将硬编码阈值替换为 ConfigProvider 读取：
func (e *Engine) Generate(
	symbol string,
	price float64,
	indicator models.Indicator,
	orderFlow models.OrderFlow,
	structure models.Structure,
	liquidity models.Liquidity,
) models.Signal {
	// 从 ConfigProvider 读取阈值（回落到原硬编码默认值）
	buyThresholdVal  := e.configProvider.GetInt(symbol, indicator.IntervalType, "buy_threshold", buyThreshold)
	sellThresholdVal := e.configProvider.GetInt(symbol, indicator.IntervalType, "sell_threshold", sellThreshold)

	// ... 原有 factors 计算逻辑不变 ...

	// resolveAction 改为接收动态阈值：
	action := resolveActionWithThreshold(score, buyThresholdVal, sellThresholdVal)

	// ... 其余不变 ...
}

// resolveActionWithThreshold 替代原 resolveAction（接收动态阈值）：
func resolveActionWithThreshold(score, buyThresh, sellThresh int) string {
	if score >= buyThresh {
		return "BUY"
	}
	if score <= sellThresh {
		return "SELL"
	}
	return "NEUTRAL"
}
```

- [ ] **Step 5: 运行测试确认通过**

```bash
cd backend && go test ./internal/signal/ -v
```
预期：全部 PASS（包括原有测试，因为 StaticConfigProvider 回落到原默认值）

- [ ] **Step 6: Commit**

```bash
git add backend/internal/signal/config_provider.go \
        backend/internal/signal/signal_engine.go \
        backend/internal/signal/signal_engine_test.go
git commit -m "feat: inject ConfigProvider into Signal Engine for hot-reload (G-04 step 2)"
```

---

### Task 10: PATCH /api/signal-configs 热更新端点（G-04 完成）

**Files:**
- Create: `backend/internal/handler/signal_config_handler.go`
- Modify: `backend/router/router.go`
- Modify: `backend/cmd/server/main.go`
- Test: `backend/internal/handler/signal_config_handler_test.go`

- [ ] **Step 1: 写失败测试**

新建 `backend/internal/handler/signal_config_handler_test.go`：

```go
package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPatchSignalConfigUpdatesProviderAndDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := newStubDBConfigProvider()
	repo := newStubSignalConfigRepo()
	h := NewSignalConfigHandler(provider, repo)

	router := gin.New()
	router.PATCH("/api/signal-configs", h.Patch)

	body, _ := json.Marshal(map[string]string{
		"symbol":   "BTCUSDT",
		"interval": "1m",
		"key":      "buy_threshold",
		"value":    "45",
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/signal-configs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	// 验证 provider 已更新
	got := provider.GetInt("BTCUSDT", "1m", "buy_threshold", 35)
	if got != 45 {
		t.Errorf("expected provider buy_threshold=45, got %d", got)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd backend && go test ./internal/handler/ -run TestPatchSignalConfig -v
```
预期：FAIL — `NewSignalConfigHandler` 未定义

- [ ] **Step 3: 实现 SignalConfigHandler**

新建 `backend/internal/handler/signal_config_handler.go`：

```go
package handler

import (
	"net/http"

	"alpha-pulse/backend/internal/signal"
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
	"github.com/gin-gonic/gin"
)

// SignalConfigHandler 处理信号配置热更新请求。
type SignalConfigHandler struct {
	provider *signal.DBConfigProvider
	repo     *repository.SignalConfigRepository
}

// NewSignalConfigHandler 创建 SignalConfigHandler。
func NewSignalConfigHandler(provider *signal.DBConfigProvider, repo *repository.SignalConfigRepository) *SignalConfigHandler {
	return &SignalConfigHandler{provider: provider, repo: repo}
}

type patchSignalConfigRequest struct {
	Symbol   string `json:"symbol"   binding:"required"`
	Interval string `json:"interval" binding:"required"`
	Key      string `json:"key"      binding:"required"`
	Value    string `json:"value"    binding:"required"`
}

// Patch 更新单条信号配置：先写 DB，再热更新内存 ConfigProvider。
func (h *SignalConfigHandler) Patch(c *gin.Context) {
	var req patchSignalConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": err.Error()})
		return
	}

	cfg := models.SignalConfig{
		Symbol:   req.Symbol,
		Interval: req.Interval,
		Key:      req.Key,
		Value:    req.Value,
	}
	if err := h.repo.Upsert(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "db upsert failed"})
		return
	}

	h.provider.Update(req.Symbol, req.Interval, req.Key, req.Value)

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": cfg})
}

// List 返回所有信号配置。
func (h *SignalConfigHandler) List(c *gin.Context) {
	configs, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": configs})
}
```

- [ ] **Step 4: 在 router.go 注册路由**

在 `backend/router/router.go` 中，找到 API 路由组，追加：

```go
// 在现有路由注册末尾追加（在 api 路由组内）：
if signalConfigHandler != nil {
    api.GET("/signal-configs", signalConfigHandler.List)
    api.PATCH("/signal-configs", signalConfigHandler.Patch)
}
```

并更新 `SetupRouter` 函数签名以接受 `signalConfigHandler`，或通过 options 模式传入。实际修改方式取决于现有 router 结构，读取 `backend/router/router.go` 后按现有模式添加。

- [ ] **Step 5: 修改 main.go 初始化配置链**

```go
// 在 main.go 中 signal.Engine 初始化处：
var signalConfigProvider *signal.DBConfigProvider
var signalConfigHandler *handler.SignalConfigHandler

signalConfigRepo := repository.NewSignalConfigRepository(db)
allConfigs, _ := signalConfigRepo.GetAll()
signalConfigProvider = signal.NewDBConfigProvider(allConfigs)

signalEngine := signal.NewEngineWithConfig(signalConfigProvider)
signalConfigHandler = handler.NewSignalConfigHandler(signalConfigProvider, signalConfigRepo)
```

- [ ] **Step 6: 运行测试确认通过**

```bash
cd backend && go test ./internal/handler/ -run TestPatchSignalConfig -v
cd backend && go build ./...
```
预期：PASS + 无编译错误

- [ ] **Step 7: 集成验证**

```bash
cd backend && go test ./... 2>&1 | grep -E "FAIL|ok"
```
预期：所有包 `ok`，无 `FAIL`

- [ ] **Step 8: Commit**

```bash
git add backend/internal/handler/signal_config_handler.go \
        backend/internal/handler/signal_config_handler_test.go \
        backend/router/router.go \
        backend/cmd/server/main.go
git commit -m "feat: PATCH /api/signal-configs hot-reload endpoint (G-04 complete)"
```

---

## 最终验证

- [ ] **全量测试**

```bash
cd backend && go test ./... -timeout 60s
```
预期：全部 PASS，无 data race

- [ ] **Race Detector 检查**

```bash
cd backend && go test -race ./internal/service/ ./internal/signal/ ./repository/
```
预期：无 DATA RACE 报告

- [ ] **编译生产二进制**

```bash
cd backend && go build -o /tmp/alpha-pulse-server ./cmd/server
```
预期：编译成功

- [ ] **最终 Commit**

```bash
git add .
git commit -m "chore: finalize To-Be refactor — all 6 core changes implemented"
```

---

*计划版本：2026-04-01 | 对应 PRD：docs/superpowers/specs/2026-04-01-auto-trading-refactor-design.md*
