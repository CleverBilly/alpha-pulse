package service

import (
	"context"
	"testing"

	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTradeServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.TradeSetting{}, &models.TradeOrder{}); err != nil {
		t.Fatalf("migrate trade models: %v", err)
	}
	return db
}

func TestAutoTradeCoordinatorSkipsExecutionWhenSymbolNotAllowed(t *testing.T) {
	db := newTradeServiceTestDB(t)
	settingsRepo := repository.NewTradeSettingRepository(db)
	orderRepo := repository.NewTradeOrderRepository(db)
	client := &stubTradeClient{}
	executor := NewTradeExecutorService(client, orderRepo)
	coordinator := NewAutoTradeCoordinator(TradeStaticConfig{
		Enabled:        true,
		AutoExecute:    true,
		AllowedSymbols: []string{"BTCUSDT"},
	}, settingsRepo, orderRepo, executor)

	if err := settingsRepo.Save(&models.TradeSetting{
		SingletonKey:        "default",
		AutoExecuteEnabled:  true,
		AllowedSymbols:      "BTCUSDT",
		RiskPct:             2,
		MinRiskReward:       1.2,
		EntryTimeoutSeconds: 45,
		MaxOpenPositions:    1,
		SyncEnabled:         true,
	}); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	err := coordinator.HandleEvent(context.Background(), AlertEvent{
		ID:          "alert-1",
		Symbol:      "ETHUSDT",
		Kind:        "setup_ready",
		EntryPrice:  3200,
		StopLoss:    3150,
		TargetPrice: 3300,
		RiskReward:  2.0,
	})
	if err != nil {
		t.Fatalf("handle event: %v", err)
	}
	if client.placeLimitOrderCalls != 0 {
		t.Fatalf("expected coordinator to skip execution, got %d place calls", client.placeLimitOrderCalls)
	}
}

func TestAutoTradeCoordinatorExecutesAllowedSetupReadyEvent(t *testing.T) {
	db := newTradeServiceTestDB(t)
	settingsRepo := repository.NewTradeSettingRepository(db)
	orderRepo := repository.NewTradeOrderRepository(db)
	client := &stubTradeClient{
		balance:  500,
		leverage: 20,
		rules: FuturesSymbolRules{
			Symbol:            "BTCUSDT",
			QuantityPrecision: 3,
			MinQty:            0.001,
			StepSize:          0.001,
			TickSize:          0.1,
		},
		placeLimitOrderResult: FuturesOrder{OrderID: "entry-1", Status: "NEW"},
	}
	executor := NewTradeExecutorService(client, orderRepo)
	coordinator := NewAutoTradeCoordinator(TradeStaticConfig{
		Enabled:        true,
		AutoExecute:    true,
		AllowedSymbols: []string{"BTCUSDT", "ETHUSDT"},
	}, settingsRepo, orderRepo, executor)

	if err := settingsRepo.Save(&models.TradeSetting{
		SingletonKey:        "default",
		AutoExecuteEnabled:  true,
		AllowedSymbols:      "BTCUSDT",
		RiskPct:             2,
		MinRiskReward:       1.2,
		EntryTimeoutSeconds: 45,
		MaxOpenPositions:    1,
		SyncEnabled:         true,
	}); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	err := coordinator.HandleEvent(context.Background(), AlertEvent{
		ID:          "alert-1",
		Symbol:      "BTCUSDT",
		Kind:        "setup_ready",
		EntryPrice:  65000,
		StopLoss:    64800,
		TargetPrice: 65600,
		RiskReward:  3.0,
	})
	if err != nil {
		t.Fatalf("handle event: %v", err)
	}
	if client.placeLimitOrderCalls != 1 {
		t.Fatalf("expected 1 place limit call, got %d", client.placeLimitOrderCalls)
	}
}
