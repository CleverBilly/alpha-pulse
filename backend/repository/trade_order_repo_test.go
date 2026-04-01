package repository

import (
	"testing"
	"time"

	"alpha-pulse/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTradeOrderTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.TradeOrder{}); err != nil {
		t.Fatalf("migrate trade order: %v", err)
	}
	return db
}

func TestTradeOrderRepositoryFindOpenAndPendingFill(t *testing.T) {
	db := newTradeOrderTestDB(t)
	repo := NewTradeOrderRepository(db)
	now := time.Now().UnixMilli()

	orders := []models.TradeOrder{
		{
			AlertID:          "alert-1",
			Symbol:           "BTCUSDT",
			Side:             "LONG",
			RequestedQty:     0.01,
			LimitPrice:       65000,
			EntryOrderType:   "LIMIT",
			EntryStatus:      "pending_fill",
			Status:           "pending_fill",
			EntryExpiresAt:   now + 30_000,
			Source:           "system",
			CreatedAtUnixMs:  now,
		},
		{
			AlertID:          "alert-2",
			Symbol:           "BTCUSDT",
			Side:             "LONG",
			RequestedQty:     0.02,
			FilledQty:        0.02,
			FilledPrice:      64990,
			EntryOrderType:   "LIMIT",
			EntryStatus:      "filled",
			Status:           "open",
			Source:           "system",
			CreatedAtUnixMs:  now + 1,
		},
		{
			AlertID:          "alert-3",
			Symbol:           "ETHUSDT",
			Side:             "SHORT",
			RequestedQty:     0.2,
			FilledQty:        0.2,
			EntryOrderType:   "LIMIT",
			EntryStatus:      "filled",
			Status:           "closed",
			Source:           "manual",
			CreatedAtUnixMs:  now + 2,
		},
	}
	for i := range orders {
		if err := repo.Create(&orders[i]); err != nil {
			t.Fatalf("create order %s: %v", orders[i].AlertID, err)
		}
	}

	pending, err := repo.FindPendingFill(10)
	if err != nil {
		t.Fatalf("find pending fill: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending_fill order, got %d", len(pending))
	}
	if pending[0].AlertID != "alert-1" {
		t.Fatalf("unexpected pending order: %+v", pending[0])
	}

	openOrders, err := repo.FindOpen("BTCUSDT")
	if err != nil {
		t.Fatalf("find open: %v", err)
	}
	if len(openOrders) != 1 {
		t.Fatalf("expected 1 open BTC order, got %d", len(openOrders))
	}
	if openOrders[0].AlertID != "alert-2" {
		t.Fatalf("unexpected open order: %+v", openOrders[0])
	}
}

func TestFindOpenBySymbolsReturnsBatchResult(t *testing.T) {
	db := newTradeOrderTestDB(t)
	repo := NewTradeOrderRepository(db)

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
