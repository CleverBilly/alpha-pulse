package repository

import (
	"fmt"
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
	records := []models.AlertRecord{
		{AlertID: fmt.Sprintf("alert-%d-1", now), Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "pending", EventTime: now},
		{AlertID: fmt.Sprintf("alert-%d-2", now), Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "target_hit", EventTime: now + 1},
		{AlertID: fmt.Sprintf("alert-%d-3", now), Symbol: "ETHUSDT", Kind: "setup_ready", Outcome: "pending", EventTime: now + 2},
	}
	for _, r := range records {
		if err := db.Create(&r).Error; err != nil {
			t.Fatalf("create record %s: %v", r.AlertID, err)
		}
	}

	result, err := repo.FindPending("BTCUSDT", "setup_ready", 100)
	if err != nil {
		t.Fatalf("FindPending: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 pending record for BTCUSDT, got %d", len(result))
	}
	if result[0].Outcome != "pending" {
		t.Fatalf("expected outcome=pending, got %s", result[0].Outcome)
	}
}

func TestUpdateOutcomeDoesNotOverwriteOtherFields(t *testing.T) {
	db := newTestDB(t)
	repo := NewAlertRecordRepository(db)

	now := time.Now().UnixMilli()
	record := &models.AlertRecord{
		AlertID:    fmt.Sprintf("alert-%d", now),
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

	outcomeAt := now + 1000
	if err := repo.UpdateOutcome(record.ID, "target_hit", 110.0, outcomeAt, 2.0); err != nil {
		t.Fatalf("UpdateOutcome: %v", err)
	}

	var updated models.AlertRecord
	db.First(&updated, record.ID)
	if updated.Outcome != "target_hit" {
		t.Fatalf("expected outcome=target_hit, got %s", updated.Outcome)
	}
	if updated.EntryPrice != 100 {
		t.Fatalf("UpdateOutcome should not change EntryPrice, got %f", updated.EntryPrice)
	}
	if updated.OutcomePrice != 110.0 {
		t.Fatalf("expected outcome_price=110.0, got %f", updated.OutcomePrice)
	}
	if updated.OutcomeAt != outcomeAt {
		t.Fatalf("expected outcome_at=%d, got %d", outcomeAt, updated.OutcomeAt)
	}
	if updated.ActualRR != 2.0 {
		t.Fatalf("expected actual_rr=2.0, got %f", updated.ActualRR)
	}
}
