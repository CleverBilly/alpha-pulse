package service

import (
	"context"
	"testing"
	"time"

	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
)

func TestTradeRuntimeReconcilePendingEntriesPromotesFilledOrderAndAddsProtection(t *testing.T) {
	db := newTradeServiceTestDB(t)
	orderRepo := repository.NewTradeOrderRepository(db)
	client := &stubTradeClient{
		orderStatusResult: FuturesOrder{
			OrderID:     "entry-1",
			Status:      "FILLED",
			FilledQty:   0.003,
			FilledPrice: 64990,
		},
	}
	runtime := NewTradeRuntime(client, orderRepo)
	runtime.now = func() time.Time {
		return time.UnixMilli(1710000000000)
	}

	if err := orderRepo.Create(&models.TradeOrder{
		AlertID:         "alert-1",
		Symbol:          "BTCUSDT",
		Side:            "LONG",
		RequestedQty:    0.003,
		EntryOrderID:    "entry-1",
		EntryOrderType:  "LIMIT",
		LimitPrice:      65000,
		EntryStatus:     "pending_fill",
		Status:          "pending_fill",
		EntryExpiresAt:  1710000000000 + 45_000,
		Source:          "system",
		CreatedAtUnixMs: 1710000000000,
	}); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	if err := runtime.ReconcilePendingEntries(context.Background()); err != nil {
		t.Fatalf("reconcile pending entries: %v", err)
	}

	orders, err := orderRepo.FindOpen("BTCUSDT")
	if err != nil {
		t.Fatalf("find open: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 open order, got %d", len(orders))
	}
	if orders[0].EntryStatus != "filled" {
		t.Fatalf("expected entry status filled, got %s", orders[0].EntryStatus)
	}
	if client.placeProtectionOrderCalls != 2 {
		t.Fatalf("expected 2 protection orders, got %d", client.placeProtectionOrderCalls)
	}
}

func TestTradeRuntimeReconcilePendingEntriesExpiresTimedOutOrder(t *testing.T) {
	db := newTradeServiceTestDB(t)
	orderRepo := repository.NewTradeOrderRepository(db)
	client := &stubTradeClient{
		orderStatusResult: FuturesOrder{
			OrderID: "entry-1",
			Status:  "NEW",
		},
	}
	runtime := NewTradeRuntime(client, orderRepo)
	runtime.now = func() time.Time {
		return time.UnixMilli(1710000050000)
	}

	if err := orderRepo.Create(&models.TradeOrder{
		AlertID:         "alert-1",
		Symbol:          "BTCUSDT",
		Side:            "LONG",
		RequestedQty:    0.003,
		EntryOrderID:    "entry-1",
		EntryOrderType:  "LIMIT",
		LimitPrice:      65000,
		EntryStatus:     "pending_fill",
		Status:          "pending_fill",
		EntryExpiresAt:  1710000000000 + 30_000,
		Source:          "system",
		CreatedAtUnixMs: 1710000000000,
	}); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	if err := runtime.ReconcilePendingEntries(context.Background()); err != nil {
		t.Fatalf("reconcile pending entries: %v", err)
	}

	expired, err := orderRepo.FindByAlertID("alert-1")
	if err != nil {
		t.Fatalf("find by alert id: %v", err)
	}
	if expired.Status != "expired" {
		t.Fatalf("expected expired status, got %s", expired.Status)
	}
	if client.cancelOrderCalls != 1 {
		t.Fatalf("expected 1 cancel call, got %d", client.cancelOrderCalls)
	}
}

func TestTradeRuntimeSyncPositionsCreatesManualOrderAndClosesMissingLocalPosition(t *testing.T) {
	db := newTradeServiceTestDB(t)
	orderRepo := repository.NewTradeOrderRepository(db)
	client := &stubTradeClient{
		positionsResult: []FuturesPosition{
			{
				Symbol:        "BTCUSDT",
				Side:          "LONG",
				Qty:           0.01,
				EntryPrice:    65000,
				UnrealizedPnL: 12.5,
				Leverage:      20,
			},
		},
	}
	runtime := NewTradeRuntime(client, orderRepo)
	runtime.now = func() time.Time {
		return time.UnixMilli(1710000100000)
	}

	if err := orderRepo.Create(&models.TradeOrder{
		AlertID:         "alert-old",
		Symbol:          "ETHUSDT",
		Side:            "LONG",
		RequestedQty:    0.2,
		FilledQty:       0.2,
		FilledPrice:     3200,
		EntryStatus:     "filled",
		Status:          "open",
		Source:          "system",
		CreatedAtUnixMs: 1710000000000,
	}); err != nil {
		t.Fatalf("seed open order: %v", err)
	}

	if err := runtime.SyncPositions(context.Background()); err != nil {
		t.Fatalf("sync positions: %v", err)
	}

	manual, err := orderRepo.FindBySourceAndSymbol("manual", "BTCUSDT")
	if err != nil {
		t.Fatalf("find manual order: %v", err)
	}
	if manual.Status != "open" {
		t.Fatalf("expected manual order to be open, got %s", manual.Status)
	}

	closed, err := orderRepo.FindByAlertID("alert-old")
	if err != nil {
		t.Fatalf("find closed order: %v", err)
	}
	if closed.Status != "closed" {
		t.Fatalf("expected missing local order to be closed, got %s", closed.Status)
	}
}

type stubTradeClient struct {
	balance                 float64
	leverage                int
	rules                   FuturesSymbolRules
	placeLimitOrderResult   FuturesOrder
	orderStatusResult       FuturesOrder
	positionsResult         []FuturesPosition
	placeLimitOrderCalls    int
	placeProtectionOrderCalls int
	cancelOrderCalls        int
}

func (s *stubTradeClient) GetFuturesBalance() (float64, error) { return s.balance, nil }
func (s *stubTradeClient) GetFuturesLeverage(symbol string) (int, error) { return s.leverage, nil }
func (s *stubTradeClient) GetFuturesSymbolRules(symbol string) (FuturesSymbolRules, error) {
	return s.rules, nil
}
func (s *stubTradeClient) PlaceFuturesLimitOrder(symbol, side string, qty, price float64) (FuturesOrder, error) {
	s.placeLimitOrderCalls++
	return s.placeLimitOrderResult, nil
}
func (s *stubTradeClient) GetFuturesOrder(symbol, orderID string) (FuturesOrder, error) {
	return s.orderStatusResult, nil
}
func (s *stubTradeClient) CancelFuturesOrder(symbol string, orderID string) error {
	s.cancelOrderCalls++
	return nil
}
func (s *stubTradeClient) PlaceFuturesProtectionOrder(symbol, side, orderType string, stopPrice float64) (string, error) {
	s.placeProtectionOrderCalls++
	return orderType + "-id", nil
}
func (s *stubTradeClient) CloseFuturesPosition(symbol, side string, qty float64) (string, error) {
	return "close-order", nil
}
func (s *stubTradeClient) GetFuturesPositions() ([]FuturesPosition, error) {
	return s.positionsResult, nil
}
