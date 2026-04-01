package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"alpha-pulse/backend/repository"
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

	rules, err := cache.GetRules("BTCUSDT")
	if err != nil {
		t.Fatalf("GetRules: %v", err)
	}
	if rules.Symbol != "BTCUSDT" {
		t.Errorf("expected rules.Symbol=BTCUSDT, got %s", rules.Symbol)
	}
	if rules.QuantityPrecision != 3 {
		t.Errorf("expected QuantityPrecision=3, got %d", rules.QuantityPrecision)
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

func TestAccountStateCacheNotStaleBeforeTTL(t *testing.T) {
	client := &stubTradeClient{balance: 500}
	cache := NewAccountStateCache(client, []string{"BTCUSDT"})

	_ = cache.Refresh()

	if cache.IsStale() {
		t.Error("expected cache to be fresh immediately after refresh")
	}
}

func TestAccountStateCacheIsStaleWhenUninitialized(t *testing.T) {
	client := &stubTradeClient{}
	cache := NewAccountStateCache(client, []string{"BTCUSDT"})

	if !cache.IsStale() {
		t.Error("expected uninitialized cache to be stale")
	}
}

func TestAccountStateCacheGetBalanceErrorsWhenUninitialized(t *testing.T) {
	client := &stubTradeClient{}
	cache := NewAccountStateCache(client, []string{"BTCUSDT"})

	_, err := cache.GetBalance()
	if err == nil {
		t.Error("expected error from uninitialized cache, got nil")
	}
}

func TestAccountStateCacheGetLeverageErrorsForUnknownSymbol(t *testing.T) {
	client := &stubTradeClient{balance: 100, leverage: 5}
	cache := NewAccountStateCache(client, []string{"BTCUSDT"})
	_ = cache.Refresh()

	_, err := cache.GetLeverage("ETHUSDT")
	if err == nil {
		t.Error("expected error for symbol not in cache, got nil")
	}
}

func TestAccountStateCacheGetRulesErrorsForUnknownSymbol(t *testing.T) {
	client := &stubTradeClient{balance: 100}
	cache := NewAccountStateCache(client, []string{"BTCUSDT"})
	_ = cache.Refresh()

	_, err := cache.GetRules("SOLUSDT")
	if err == nil {
		t.Error("expected error for symbol not in cache, got nil")
	}
}

// TestExecuteLimitEntryUsesCacheWhenFresh 验证缓存新鲜时 ExecuteLimitEntry 不调用账户状态 API。
func TestExecuteLimitEntryUsesCacheWhenFresh(t *testing.T) {
	db := newTradeServiceTestDB(t)
	orderRepo := repository.NewTradeOrderRepository(db)
	client := &stubTradeClient{
		balance:  500,
		leverage: 25,
		rules: FuturesSymbolRules{
			Symbol:            "BTCUSDT",
			QuantityPrecision: 3,
			MinQty:            0.001,
			StepSize:          0.001,
		},
		placeLimitOrderResult: FuturesOrder{
			OrderID: "limit-1",
			Status:  "NEW",
		},
	}

	// 用 client 填充缓存
	cache := NewAccountStateCache(client, []string{"BTCUSDT"})
	if err := cache.Refresh(); err != nil {
		t.Fatalf("cache.Refresh: %v", err)
	}

	// failOnAccountCallsTradeClient 的账户查询方法全部返回错误，
	// 只有 PlaceFuturesLimitOrder 等非账户方法代理到原始 client。
	failClient := &failOnAccountCallsTradeClient{delegate: client}
	executor := NewTradeExecutorService(failClient, orderRepo)
	executor.SetAccountStateCache(cache)

	_, err := executor.ExecuteLimitEntry(context.Background(), AlertEvent{
		ID:          "alert-cache-test",
		Symbol:      "BTCUSDT",
		Kind:        "setup_ready",
		EntryPrice:  65000,
		StopLoss:    64800,
		TargetPrice: 65600,
		RiskReward:  3.0,
	}, TradeSettings{
		AutoExecuteEnabled:  true,
		AllowedSymbols:      []string{"BTCUSDT"},
		RiskPct:             2,
		MinRiskReward:       1.2,
		EntryTimeoutSeconds: 45,
		MaxOpenPositions:    1,
	})
	if err != nil {
		t.Fatalf("ExecuteLimitEntry with fresh cache: %v", err)
	}
}

// failOnAccountCallsTradeClient 包装 delegate，账户状态查询方法全部返回错误，
// 用于验证缓存命中路径绕过了直接 API 调用。
type failOnAccountCallsTradeClient struct {
	delegate TradeClient
}

func (f *failOnAccountCallsTradeClient) GetFuturesBalance() (float64, error) {
	return 0, fmt.Errorf("should not call GetFuturesBalance when cache is fresh")
}
func (f *failOnAccountCallsTradeClient) GetFuturesLeverage(symbol string) (int, error) {
	return 0, fmt.Errorf("should not call GetFuturesLeverage when cache is fresh")
}
func (f *failOnAccountCallsTradeClient) GetFuturesSymbolRules(symbol string) (FuturesSymbolRules, error) {
	return FuturesSymbolRules{}, fmt.Errorf("should not call GetFuturesSymbolRules when cache is fresh")
}
func (f *failOnAccountCallsTradeClient) PlaceFuturesLimitOrder(symbol, side string, qty, price float64) (FuturesOrder, error) {
	return f.delegate.PlaceFuturesLimitOrder(symbol, side, qty, price)
}
func (f *failOnAccountCallsTradeClient) GetFuturesOrder(symbol, orderID string) (FuturesOrder, error) {
	return f.delegate.GetFuturesOrder(symbol, orderID)
}
func (f *failOnAccountCallsTradeClient) CancelFuturesOrder(symbol string, orderID string) error {
	return f.delegate.CancelFuturesOrder(symbol, orderID)
}
func (f *failOnAccountCallsTradeClient) PlaceFuturesProtectionOrder(symbol, side, orderType string, stopPrice float64) (string, error) {
	return f.delegate.PlaceFuturesProtectionOrder(symbol, side, orderType, stopPrice)
}
func (f *failOnAccountCallsTradeClient) CloseFuturesPosition(symbol, side string, qty float64) (string, error) {
	return f.delegate.CloseFuturesPosition(symbol, side, qty)
}
func (f *failOnAccountCallsTradeClient) GetFuturesPositions() ([]FuturesPosition, error) {
	return f.delegate.GetFuturesPositions()
}
