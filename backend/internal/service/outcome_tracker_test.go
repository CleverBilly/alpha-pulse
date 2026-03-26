package service

import (
	"testing"
	"time"

	"alpha-pulse/backend/models"
)

// buildTestKline 构造测试用 K 线。
func buildTestKline(openTime int64, open, high, low, closePrice float64) models.Kline {
	return models.Kline{
		OpenTime:   openTime,
		OpenPrice:  open,
		HighPrice:  high,
		LowPrice:   low,
		ClosePrice: closePrice,
	}
}

func TestEvalOutcomeLongTargetHit(t *testing.T) {
	record := models.AlertRecord{
		DirectionState: "strong-bullish",
		EntryPrice:     100,
		StopLoss:       95,
		TargetPrice:    110,
		EventTime:      1000,
		Outcome:        "pending",
	}
	klines := []models.Kline{
		buildTestKline(1001, 101, 105, 100, 103),
		buildTestKline(1002, 103, 112, 102, 111), // high >= target
	}
	outcome, price, _ := evalOutcome(record, klines, time.UnixMilli(2000))
	if outcome != "target_hit" {
		t.Fatalf("expected target_hit, got %s", outcome)
	}
	if price != 110 {
		t.Fatalf("expected outcomePrice=110, got %f", price)
	}
}

func TestEvalOutcomeLongStopHitFirst(t *testing.T) {
	record := models.AlertRecord{
		DirectionState: "bullish",
		EntryPrice:     100,
		StopLoss:       95,
		TargetPrice:    110,
		EventTime:      1000,
		Outcome:        "pending",
	}
	// 同一根 K 线同时触达止损和目标，止损优先
	klines := []models.Kline{
		buildTestKline(1001, 100, 115, 94, 105), // low<=stop AND high>=target
	}
	outcome, _, _ := evalOutcome(record, klines, time.UnixMilli(2000))
	if outcome != "stop_hit" {
		t.Fatalf("expected stop_hit (stop priority), got %s", outcome)
	}
}

func TestEvalOutcomeExpired(t *testing.T) {
	record := models.AlertRecord{
		DirectionState: "bullish",
		EntryPrice:     100,
		StopLoss:       95,
		TargetPrice:    110,
		EventTime:      1000,
		Outcome:        "pending",
	}
	klines := []models.Kline{} // 没有 K 线
	// now = event_time + 61 分钟
	now := time.UnixMilli(1000 + 61*60*1000)
	outcome, _, _ := evalOutcome(record, klines, now)
	if outcome != "expired" {
		t.Fatalf("expected expired, got %s", outcome)
	}
}

func TestEvalOutcomeShortStopHitFirst(t *testing.T) {
	record := models.AlertRecord{
		DirectionState: "strong-bearish",
		EntryPrice:     100,
		StopLoss:       105,
		TargetPrice:    90,
		EventTime:      1000,
		Outcome:        "pending",
	}
	klines := []models.Kline{
		buildTestKline(1001, 101, 106, 89, 95), // high>=stop AND low<=target
	}
	outcome, _, _ := evalOutcome(record, klines, time.UnixMilli(2000))
	if outcome != "stop_hit" {
		t.Fatalf("expected stop_hit for short, got %s", outcome)
	}
}

func TestEvalOutcomeNeutralDirectionSkipped(t *testing.T) {
	record := models.AlertRecord{
		DirectionState: "neutral",
		EntryPrice:     100,
		StopLoss:       95,
		TargetPrice:    110,
		EventTime:      1000,
		Outcome:        "pending",
	}
	klines := []models.Kline{buildTestKline(1001, 100, 120, 80, 100)}
	outcome, _, _ := evalOutcome(record, klines, time.UnixMilli(2000))
	if outcome != "" {
		t.Fatalf("expected no tracking for neutral direction, got %s", outcome)
	}
}
