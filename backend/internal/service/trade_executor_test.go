package service

import (
	"context"
	"testing"

	"alpha-pulse/backend/repository"
)

func TestTradeExecutorExecuteLimitEntryCreatesPendingFillOrder(t *testing.T) {
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
			TickSize:          0.1,
		},
		placeLimitOrderResult: FuturesOrder{
			OrderID: "limit-1",
			Status:  "NEW",
		},
	}
	executor := NewTradeExecutorService(client, orderRepo)

	order, err := executor.ExecuteLimitEntry(context.Background(), AlertEvent{
		ID:          "alert-1",
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
		SyncEnabled:         true,
	})
	if err != nil {
		t.Fatalf("execute limit entry: %v", err)
	}
	if order.Status != "pending_fill" {
		t.Fatalf("expected pending_fill status, got %s", order.Status)
	}
	if order.EntryOrderID != "limit-1" {
		t.Fatalf("unexpected entry order id: %s", order.EntryOrderID)
	}
	if order.RequestedQty != 0.003 {
		t.Fatalf("unexpected requested qty: %f", order.RequestedQty)
	}
}
