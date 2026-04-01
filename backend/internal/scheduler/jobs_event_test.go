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
	cancel() // 确保 goroutine 在断言前退出，避免测试退出时 goroutine 泄漏

	if warmupCount.Load() < 1 {
		t.Errorf("expected WarmupSymbol to be called at least once, got %d", warmupCount.Load())
	}
}

// TestStartWithNilChannelFallsBackToPolling 验证 klineEvents=nil 时退化为纯轮询。
func TestStartWithNilChannelFallsBackToPolling(t *testing.T) {
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
		interval:      50 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// nil channel：退化为轮询模式
	go jobs.StartWithKlineEvents(ctx, nil)

	// 等待至少一次轮询（runOnce 在 Start 开头立即调用一次，加上 ticker 触发的）
	time.Sleep(150 * time.Millisecond)

	if warmupCount.Load() < 1 {
		t.Errorf("expected polling to call WarmupSymbol at least once, got %d", warmupCount.Load())
	}
}

// TestHandleKlineEventOnlyWarmsUpMatchedSymbol 验证事件只触发对应 symbol 的 warmup。
func TestHandleKlineEventOnlyWarmsUpMatchedSymbol(t *testing.T) {
	var calledWith []string
	stub := &stubMarketServiceForScheduler{
		warmupFn: func(symbol string) error {
			calledWith = append(calledWith, symbol)
			return nil
		},
	}

	jobs := &Jobs{
		marketService: stub,
		symbols:       []string{"BTCUSDT", "ETHUSDT"},
		interval:      60 * time.Second,
	}

	jobs.handleKlineEvent(collector.KlineClosedEvent{Symbol: "ETHUSDT", Interval: "5m"})

	if len(calledWith) != 1 || calledWith[0] != "ETHUSDT" {
		t.Errorf("expected WarmupSymbol called once with ETHUSDT, got %v", calledWith)
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
