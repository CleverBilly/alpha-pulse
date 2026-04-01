package service

import (
	"context"
	"errors"
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
