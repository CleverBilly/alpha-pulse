package indicator

import (
	"math"
	"testing"
	"time"

	"alpha-pulse/backend/models"
)

func TestEngineCalculateRequiresEnoughKlines(t *testing.T) {
	engine := NewEngine()
	_, err := engine.Calculate("BTCUSDT", make([]models.Kline, engine.MinimumRequired()-1))
	if err == nil {
		t.Fatal("expected insufficient history error, got nil")
	}
}

func TestEngineCalculateWithTrendData(t *testing.T) {
	engine := NewEngine()
	klines := buildTrendKlines(engine.HistoryLimit())
	reversed := reverseKlines(klines)

	indicator, err := engine.Calculate("BTCUSDT", reversed)
	if err != nil {
		t.Fatalf("calculate indicators failed: %v", err)
	}

	if indicator.Symbol != "BTCUSDT" {
		t.Fatalf("unexpected symbol: %s", indicator.Symbol)
	}
	if indicator.EMA20 <= indicator.EMA50 {
		t.Fatalf("expected EMA20 > EMA50 in uptrend, got EMA20=%f EMA50=%f", indicator.EMA20, indicator.EMA50)
	}
	if indicator.RSI <= 50 {
		t.Fatalf("expected RSI to be above 50 in uptrend, got %f", indicator.RSI)
	}
	if indicator.MACD <= 0 {
		t.Fatalf("expected MACD to be positive in uptrend, got %f", indicator.MACD)
	}
	if indicator.MACDHistogram <= 0 {
		t.Fatalf("expected MACD histogram to be positive in uptrend, got %f", indicator.MACDHistogram)
	}
	if indicator.ATR <= 0 {
		t.Fatalf("expected ATR to be positive, got %f", indicator.ATR)
	}
	if indicator.BollingerUpper <= indicator.BollingerMiddle || indicator.BollingerMiddle <= indicator.BollingerLower {
		t.Fatalf(
			"expected bollinger bands ordered upper > middle > lower, got upper=%f middle=%f lower=%f",
			indicator.BollingerUpper,
			indicator.BollingerMiddle,
			indicator.BollingerLower,
		)
	}
	if indicator.VWAP <= 0 {
		t.Fatalf("expected VWAP to be positive, got %f", indicator.VWAP)
	}
}

func buildTrendKlines(limit int) []models.Kline {
	klines := make([]models.Kline, 0, limit)
	base := 60000.0
	start := time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC)

	for i := 0; i < limit; i++ {
		drift := math.Pow(float64(i), 2) * 3.2
		wave := math.Sin(float64(i)/5.0) * 120
		closePrice := base + drift + wave
		openPrice := closePrice - 25 + (math.Cos(float64(i)/4.0) * 10)
		highPrice := math.Max(openPrice, closePrice) + 45
		lowPrice := math.Min(openPrice, closePrice) - 42
		volume := 120 + float64(i)*1.8

		klines = append(klines, models.Kline{
			Symbol:       "BTCUSDT",
			IntervalType: "1m",
			OpenPrice:    openPrice,
			HighPrice:    highPrice,
			LowPrice:     lowPrice,
			ClosePrice:   closePrice,
			Volume:       volume,
			OpenTime:     start.Add(time.Duration(i) * time.Minute).UnixMilli(),
			CreatedAt:    start.Add(time.Duration(i) * time.Minute),
		})
	}

	return klines
}

func reverseKlines(klines []models.Kline) []models.Kline {
	reversed := make([]models.Kline, len(klines))
	copy(reversed, klines)

	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}

	return reversed
}
