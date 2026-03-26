package service

import (
	"testing"

	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestAlertRecordRepoForService(t *testing.T) (*repository.AlertRecordRepository, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.AlertRecord{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return repository.NewAlertRecordRepository(db), db
}

func TestGetAlertStatsWinRateExcludesPendingFromDenominator(t *testing.T) {
	repo, db := newTestAlertRecordRepoForService(t)

	// 写入测试数据：2 target_hit, 1 stop_hit, 1 pending, 1 expired
	records := []models.AlertRecord{
		{AlertID: "t1", Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "target_hit", ActualRR: 2.0, EventTime: 1},
		{AlertID: "t2", Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "target_hit", ActualRR: 1.5, EventTime: 2},
		{AlertID: "t3", Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "stop_hit", ActualRR: -1.0, EventTime: 3},
		{AlertID: "t4", Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "pending", EventTime: 4},
		{AlertID: "t5", Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "expired", EventTime: 5},
	}
	for i := range records {
		if err := db.Create(&records[i]).Error; err != nil {
			t.Fatalf("create record: %v", err)
		}
	}

	svc := &AlertService{repo: repo, symbols: []string{"BTCUSDT"}}
	stats, err := svc.GetAlertStats("BTCUSDT", 50)
	if err != nil {
		t.Fatalf("GetAlertStats: %v", err)
	}

	if stats.Total != 5 {
		t.Fatalf("expected total=5, got %d", stats.Total)
	}
	if stats.Pending != 1 {
		t.Fatalf("expected pending=1, got %d", stats.Pending)
	}
	// win_rate = 2 / (2+1) * 100 ≈ 66.67
	expected := float64(2) / float64(3) * 100
	if stats.WinRate < expected-0.1 || stats.WinRate > expected+0.1 {
		t.Fatalf("expected win_rate≈%.2f, got %.2f", expected, stats.WinRate)
	}
	// avg_rr = (2.0 + 1.5 + (-1.0)) / 3 ≈ 0.83
	expectedRR := (2.0 + 1.5 + (-1.0)) / 3
	if stats.AvgRR < expectedRR-0.01 || stats.AvgRR > expectedRR+0.01 {
		t.Fatalf("expected avg_rr≈%.2f, got %.2f", expectedRR, stats.AvgRR)
	}
}
