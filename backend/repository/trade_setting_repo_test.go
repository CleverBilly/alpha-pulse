package repository

import (
	"testing"

	"alpha-pulse/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTradeSettingTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.TradeSetting{}); err != nil {
		t.Fatalf("migrate trade setting: %v", err)
	}
	return db
}

func TestTradeSettingRepositorySaveAndLoadDefault(t *testing.T) {
	db := newTradeSettingTestDB(t)
	repo := NewTradeSettingRepository(db)

	record := &models.TradeSetting{
		SingletonKey:         "default",
		AutoExecuteEnabled:   true,
		AllowedSymbols:       "BTCUSDT,ETHUSDT",
		RiskPct:              2.5,
		MinRiskReward:        1.6,
		EntryTimeoutSeconds:  45,
		MaxOpenPositions:     2,
		SyncEnabled:          true,
		UpdatedBy:            "tester",
	}

	if err := repo.Save(record); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := repo.GetDefault()
	if err != nil {
		t.Fatalf("get default: %v", err)
	}
	if !loaded.AutoExecuteEnabled {
		t.Fatal("expected auto execute enabled to persist")
	}
	if loaded.AllowedSymbols != "BTCUSDT,ETHUSDT" {
		t.Fatalf("unexpected allowed symbols: %s", loaded.AllowedSymbols)
	}
	if loaded.EntryTimeoutSeconds != 45 {
		t.Fatalf("unexpected timeout: %d", loaded.EntryTimeoutSeconds)
	}
	if loaded.UpdatedBy != "tester" {
		t.Fatalf("unexpected updated_by: %s", loaded.UpdatedBy)
	}
}
