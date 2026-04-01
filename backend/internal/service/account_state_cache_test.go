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
