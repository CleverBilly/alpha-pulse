package liquidity

import (
	"encoding/json"
	"testing"
	"time"

	"alpha-pulse/backend/models"
	binancepkg "alpha-pulse/backend/pkg/binance"
)

func TestAnalyzeDetectsSellSweep(t *testing.T) {
	engine := NewEngine()
	klines := buildLiquiditySweepKlines(engine.HistoryLimit())

	result, err := engine.Analyze("BTCUSDT", klines)
	if err != nil {
		t.Fatalf("analyze liquidity failed: %v", err)
	}

	if result.SweepType != "sell_sweep" {
		t.Fatalf("expected sell_sweep, got %s", result.SweepType)
	}
	if result.BuyLiquidity <= 0 || result.SellLiquidity <= 0 {
		t.Fatalf("expected liquidity levels > 0, got buy=%f sell=%f", result.BuyLiquidity, result.SellLiquidity)
	}
	if result.DataSource != "kline" {
		t.Fatalf("expected kline data source, got %s", result.DataSource)
	}
	if result.EqualHigh <= 0 || result.EqualLow <= 0 {
		t.Fatalf("expected equal high/low to be detected, got equal_high=%f equal_low=%f", result.EqualHigh, result.EqualLow)
	}
	if len(result.StopClusters) < 2 {
		t.Fatalf("expected stop clusters to be present, got=%d", len(result.StopClusters))
	}
}

func TestAnalyzeWithOrderBookUsesDepthWalls(t *testing.T) {
	engine := NewEngine()
	klines := buildLiquiditySweepKlines(engine.HistoryLimit())
	snapshot := buildOrderBookSnapshot(t)

	result, err := engine.AnalyzeWithOrderBook("BTCUSDT", klines, snapshot)
	if err != nil {
		t.Fatalf("analyze liquidity with order book failed: %v", err)
	}

	if result.DataSource != "orderbook" {
		t.Fatalf("expected orderbook data source, got %s", result.DataSource)
	}
	if result.OrderBookImbalance <= 0 {
		t.Fatalf("expected positive order book imbalance, got %f", result.OrderBookImbalance)
	}
	if result.BuyLiquidity <= 61480 || result.BuyLiquidity >= 61520 {
		t.Fatalf("expected buy liquidity near bid wall cluster, got %f", result.BuyLiquidity)
	}
	if result.SellLiquidity <= 61580 || result.SellLiquidity >= 61620 {
		t.Fatalf("expected sell liquidity near ask wall cluster, got %f", result.SellLiquidity)
	}
	if result.SweepType != "sell_sweep" {
		t.Fatalf("expected sell_sweep from order book analysis, got %s", result.SweepType)
	}
	if result.EqualHigh <= 0 || result.EqualLow <= 0 {
		t.Fatalf("expected equal high/low to be detected, got equal_high=%f equal_low=%f", result.EqualHigh, result.EqualLow)
	}
	if len(result.StopClusters) < 2 {
		t.Fatalf("expected stop clusters from order book analysis, got=%d", len(result.StopClusters))
	}
	if !containsClusterKind(result.StopClusters, "sell_stop_cluster") {
		t.Fatal("expected sell_stop_cluster to be present")
	}
	if !containsClusterKind(result.StopClusters, "buy_stop_cluster") {
		t.Fatal("expected buy_stop_cluster to be present")
	}
}

func buildLiquiditySweepKlines(limit int) []models.Kline {
	klines := make([]models.Kline, 0, limit)
	base := 61500.0
	start := time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC)

	for i := 0; i < limit-1; i++ {
		open := base + float64(i%6)*18
		close := open + 8
		high := close + 20
		low := open - 16
		if i%8 == 2 || i%8 == 5 {
			high = base + 124
		}
		if i%9 == 1 || i%9 == 6 {
			low = base - 76
		}
		volume := 70 + float64(i%10)*3

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

	lastIndex := limit - 1
	klines = append(klines, models.Kline{
		Symbol:       "BTCUSDT",
		IntervalType: "1m",
		OpenPrice:    base + 30,
		HighPrice:    base + 72,
		LowPrice:     base - 138,
		ClosePrice:   base + 36,
		Volume:       180,
		OpenTime:     start.Add(time.Duration(lastIndex) * time.Minute).UnixMilli(),
		CreatedAt:    start.Add(time.Duration(lastIndex) * time.Minute),
	})

	return klines
}

func buildOrderBookSnapshot(t *testing.T) models.OrderBookSnapshot {
	t.Helper()

	bids := []binancepkg.OrderBookLevel{
		{Price: 61520, Quantity: 1.2},
		{Price: 61510, Quantity: 1.4},
		{Price: 61500, Quantity: 5.8},
		{Price: 61490, Quantity: 5.2},
		{Price: 61480, Quantity: 4.9},
		{Price: 61470, Quantity: 4.2},
		{Price: 61460, Quantity: 1.0},
		{Price: 61450, Quantity: 0.8},
	}
	asks := []binancepkg.OrderBookLevel{
		{Price: 61560, Quantity: 0.9},
		{Price: 61570, Quantity: 1.1},
		{Price: 61580, Quantity: 4.0},
		{Price: 61590, Quantity: 4.4},
		{Price: 61600, Quantity: 3.8},
		{Price: 61610, Quantity: 3.2},
		{Price: 61620, Quantity: 1.0},
		{Price: 61630, Quantity: 0.8},
	}

	bidsJSON, err := json.Marshal(bids)
	if err != nil {
		t.Fatalf("marshal bids failed: %v", err)
	}
	asksJSON, err := json.Marshal(asks)
	if err != nil {
		t.Fatalf("marshal asks failed: %v", err)
	}

	return models.OrderBookSnapshot{
		Symbol:       "BTCUSDT",
		LastUpdateID: 1001,
		DepthLevel:   20,
		BidsJSON:     string(bidsJSON),
		AsksJSON:     string(asksJSON),
		BestBidPrice: bids[0].Price,
		BestAskPrice: asks[0].Price,
		Spread:       asks[0].Price - bids[0].Price,
		EventTime:    time.Date(2026, 3, 6, 1, 0, 0, 0, time.UTC).UnixMilli(),
	}
}

func containsClusterKind(clusters []models.LiquidityCluster, kind string) bool {
	for _, cluster := range clusters {
		if cluster.Kind == kind {
			return true
		}
	}
	return false
}
