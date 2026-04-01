package repository

import (
	"testing"

	"alpha-pulse/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSignalConfigTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.SignalConfig{}); err != nil {
		t.Fatalf("migrate signal_config: %v", err)
	}
	return db
}

func TestSignalConfigUpsertAndGet(t *testing.T) {
	db := newSignalConfigTestDB(t)
	repo := NewSignalConfigRepository(db)

	if err := repo.Upsert(models.SignalConfig{
		Symbol: "BTCUSDT", Interval: "1m", Key: "buy_threshold", Value: "40",
	}); err != nil {
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

func TestSignalConfigUpsertUpdate(t *testing.T) {
	db := newSignalConfigTestDB(t)
	repo := NewSignalConfigRepository(db)

	// 第一次插入
	if err := repo.Upsert(models.SignalConfig{
		Symbol: "BTCUSDT", Interval: "1m", Key: "buy_threshold", Value: "40",
	}); err != nil {
		t.Fatalf("first Upsert: %v", err)
	}

	// 第二次用同一 key 更新
	if err := repo.Upsert(models.SignalConfig{
		Symbol: "BTCUSDT", Interval: "1m", Key: "buy_threshold", Value: "50",
	}); err != nil {
		t.Fatalf("second Upsert: %v", err)
	}

	configs, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].Value != "50" {
		t.Errorf("expected value=50 after update, got %s", configs[0].Value)
	}
	if configs[0].UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set after upsert")
	}
}

func TestSignalConfigGetBySymbolInterval(t *testing.T) {
	db := newSignalConfigTestDB(t)
	repo := NewSignalConfigRepository(db)

	// 插入多个配置
	if err := repo.Upsert(models.SignalConfig{
		Symbol: "BTCUSDT", Interval: "1m", Key: "buy_threshold", Value: "40",
	}); err != nil {
		t.Fatalf("upsert 1: %v", err)
	}
	if err := repo.Upsert(models.SignalConfig{
		Symbol: "BTCUSDT", Interval: "5m", Key: "buy_threshold", Value: "45",
	}); err != nil {
		t.Fatalf("upsert 2: %v", err)
	}
	if err := repo.Upsert(models.SignalConfig{
		Symbol: "ETHUSDT", Interval: "1m", Key: "buy_threshold", Value: "35",
	}); err != nil {
		t.Fatalf("upsert 3: %v", err)
	}

	// 查询特定 symbol + interval
	configs, err := repo.GetBySymbolInterval("BTCUSDT", "1m")
	if err != nil {
		t.Fatalf("GetBySymbolInterval: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config for BTCUSDT/1m, got %d", len(configs))
	}
	if configs[0].Value != "40" {
		t.Errorf("expected value=40, got %s", configs[0].Value)
	}
}
