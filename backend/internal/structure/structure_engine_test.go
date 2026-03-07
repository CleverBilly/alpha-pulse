package structure

import (
	"math"
	"testing"
	"time"

	"alpha-pulse/backend/models"
)

func TestAnalyzeUptrendStructureWithSwingEvents(t *testing.T) {
	engine := NewEngine()
	klines := buildSwingUptrendKlines(engine.HistoryLimit())

	result, err := engine.Analyze("BTCUSDT", klines)
	if err != nil {
		t.Fatalf("analyze structure failed: %v", err)
	}

	if result.Trend != "uptrend" {
		t.Fatalf("expected uptrend, got %s", result.Trend)
	}
	if result.Support >= result.Resistance {
		t.Fatalf("expected support < resistance, got support=%f resistance=%f", result.Support, result.Resistance)
	}
	if len(result.Events) == 0 {
		t.Fatal("expected structure events to be present")
	}
	if !containsEventLabel(result.Events, "HH") {
		t.Fatal("expected HH event in structure events")
	}
	if !containsEventLabel(result.Events, "HL") {
		t.Fatal("expected HL event in structure events")
	}
}

func TestAnalyzeDetectsBearishChoch(t *testing.T) {
	engine := NewEngine()
	klines := buildDowntrendChochKlines(engine.HistoryLimit())

	result, err := engine.Analyze("ETHUSDT", klines)
	if err != nil {
		t.Fatalf("analyze structure failed: %v", err)
	}

	if result.Trend != "downtrend" {
		t.Fatalf("expected downtrend before reversal, got %s", result.Trend)
	}
	if !result.Choch {
		t.Fatal("expected choch to be true")
	}
	if !containsEventLabel(result.Events, "CHOCH") {
		t.Fatal("expected CHOCH event in structure events")
	}
}

func buildSwingUptrendKlines(limit int) []models.Kline {
	klines := make([]models.Kline, 0, limit)
	base := 62000.0
	start := time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC)

	for i := 0; i < limit; i++ {
		drift := float64(i) * 22
		wavePrimary := math.Sin(float64(i)/2.6) * 95
		waveSecondary := math.Cos(float64(i)/5.0) * 26
		open := base + drift + wavePrimary + waveSecondary
		close := open + 18 + math.Sin(float64(i)/3.0)*11
		high := math.Max(open, close) + 34 + math.Abs(math.Cos(float64(i)/2.1))*18
		low := math.Min(open, close) - 32 - math.Abs(math.Sin(float64(i)/2.3))*16
		volume := 90 + float64(i%12)*4

		klines = append(klines, models.Kline{
			Symbol:       "BTCUSDT",
			IntervalType: "1m",
			OpenPrice:    open,
			HighPrice:    high,
			LowPrice:     low,
			ClosePrice:   close,
			Volume:       volume,
			OpenTime:     start.Add(time.Duration(i) * time.Minute).UnixMilli(),
			CreatedAt:    start.Add(time.Duration(i) * time.Minute),
		})
	}

	return klines
}

func buildDowntrendChochKlines(limit int) []models.Kline {
	klines := make([]models.Kline, 0, limit)
	base := 3450.0
	start := time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC)

	for i := 0; i < limit-1; i++ {
		drift := -float64(i) * 6.5
		wavePrimary := math.Sin(float64(i)/2.8) * 22
		waveSecondary := math.Cos(float64(i)/5.4) * 8
		open := base + drift + wavePrimary + waveSecondary
		close := open - 5 + math.Cos(float64(i)/3.1)*4
		high := math.Max(open, close) + 12 + math.Abs(math.Sin(float64(i)/2.0))*7
		low := math.Min(open, close) - 14 - math.Abs(math.Cos(float64(i)/2.5))*8
		volume := 55 + float64(i%10)*2

		klines = append(klines, models.Kline{
			Symbol:       "ETHUSDT",
			IntervalType: "5m",
			OpenPrice:    open,
			HighPrice:    high,
			LowPrice:     low,
			ClosePrice:   close,
			Volume:       volume,
			OpenTime:     start.Add(time.Duration(i) * 5 * time.Minute).UnixMilli(),
			CreatedAt:    start.Add(time.Duration(i) * 5 * time.Minute),
		})
	}

	lastIndex := limit - 1
	priorResistance := maxRecentHigh(klines[len(klines)-12:])
	lastOpen := klines[len(klines)-1].ClosePrice - 6
	lastClose := priorResistance * 1.012
	lastHigh := lastClose + 10
	lastLow := lastOpen - 18

	klines = append(klines, models.Kline{
		Symbol:       "ETHUSDT",
		IntervalType: "5m",
		OpenPrice:    lastOpen,
		HighPrice:    lastHigh,
		LowPrice:     lastLow,
		ClosePrice:   lastClose,
		Volume:       140,
		OpenTime:     start.Add(time.Duration(lastIndex) * 5 * time.Minute).UnixMilli(),
		CreatedAt:    start.Add(time.Duration(lastIndex) * 5 * time.Minute),
	})

	return klines
}

func maxRecentHigh(klines []models.Kline) float64 {
	maxValue := 0.0
	for _, kline := range klines {
		if kline.HighPrice > maxValue {
			maxValue = kline.HighPrice
		}
	}
	return maxValue
}

func containsEventLabel(events []models.StructureEvent, label string) bool {
	for _, event := range events {
		if event.Label == label {
			return true
		}
	}
	return false
}
