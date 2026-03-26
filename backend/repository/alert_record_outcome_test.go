package repository

import (
	"testing"
	"time"

	"alpha-pulse/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.AlertRecord{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestFindPendingReturnsOnlyPendingRecords(t *testing.T) {
	db := newTestDB(t)
	repo := NewAlertRecordRepository(db)

	now := time.Now().UnixMilli()
	_ = db.Create(&models.AlertRecord{Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "pending", EventTime: now}).Error
	_ = db.Create(&models.AlertRecord{Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "target_hit", EventTime: now + 1}).Error
	_ = db.Create(&models.AlertRecord{Symbol: "ETHUSDT", Kind: "setup_ready", Outcome: "pending", EventTime: now + 2}).Error

	records, err := repo.FindPending("BTCUSDT", 100)
	if err != nil {
		t.Fatalf("FindPending: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 pending record for BTCUSDT, got %d", len(records))
	}
	if records[0].Outcome != "pending" {
		t.Fatalf("expected outcome=pending, got %s", records[0].Outcome)
	}
}

func TestUpdateOutcomeDoesNotOverwriteOtherFields(t *testing.T) {
	db := newTestDB(t)
	repo := NewAlertRecordRepository(db)

	now := time.Now().UnixMilli()
	record := &models.AlertRecord{
		Symbol:     "BTCUSDT",
		Kind:       "setup_ready",
		Outcome:    "pending",
		EventTime:  now,
		EntryPrice: 100,
		StopLoss:   95,
	}
	if err := db.Create(record).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.UpdateOutcome(record.ID, "target_hit", 110.0, now+1000, 2.0); err != nil {
		t.Fatalf("UpdateOutcome: %v", err)
	}

	var updated models.AlertRecord
	db.First(&updated, record.ID)
	if updated.Outcome != "target_hit" {
		t.Fatalf("expected target_hit, got %s", updated.Outcome)
	}
	if updated.EntryPrice != 100 {
		t.Fatalf("UpdateOutcome should not change EntryPrice, got %f", updated.EntryPrice)
	}
	if updated.ActualRR != 2.0 {
		t.Fatalf("expected actual_rr=2.0, got %f", updated.ActualRR)
	}
}
