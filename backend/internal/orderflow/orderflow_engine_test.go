package orderflow

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"alpha-pulse/backend/models"
	binancepkg "alpha-pulse/backend/pkg/binance"
)

func TestAnalyzeBullishKlines(t *testing.T) {
	engine := NewEngine()
	klines := buildBullishKlines(engine.HistoryLimit())

	result, err := engine.Analyze("BTCUSDT", klines)
	if err != nil {
		t.Fatalf("analyze order flow failed: %v", err)
	}

	if result.BuyVolume <= result.SellVolume {
		t.Fatalf("expected buy volume > sell volume, got buy=%f sell=%f", result.BuyVolume, result.SellVolume)
	}
	if result.Delta <= 0 {
		t.Fatalf("expected positive delta, got %f", result.Delta)
	}
	if result.CVD <= 0 {
		t.Fatalf("expected positive cvd, got %f", result.CVD)
	}
}

func TestAnalyzeAggTradesBullishFlow(t *testing.T) {
	engine := NewEngine()
	trades := buildBullishAggTrades(engine.TradeHistoryLimit())

	result, err := engine.AnalyzeAggTrades("BTCUSDT", trades)
	if err != nil {
		t.Fatalf("analyze agg trades failed: %v", err)
	}

	if result.BuyVolume <= result.SellVolume {
		t.Fatalf("expected buy volume > sell volume, got buy=%f sell=%f", result.BuyVolume, result.SellVolume)
	}
	if result.Delta <= 0 {
		t.Fatalf("expected positive delta, got %f", result.Delta)
	}
	if result.CVD <= 0 {
		t.Fatalf("expected positive cvd, got %f", result.CVD)
	}
	if result.DataSource != "agg_trade" {
		t.Fatalf("expected agg_trade data source, got %s", result.DataSource)
	}
	if result.BuyLargeTradeCount == 0 {
		t.Fatal("expected at least one buy large trade")
	}
	if result.BuyLargeTradeNotional <= result.SellLargeTradeNotional {
		t.Fatalf(
			"expected buy large trade notional > sell large trade notional, got buy=%f sell=%f",
			result.BuyLargeTradeNotional,
			result.SellLargeTradeNotional,
		)
	}
	if len(result.LargeTrades) == 0 {
		t.Fatal("expected large trade events to be present")
	}
}

func TestAnalyzeAggTradesDetectsMicrostructureSignals(t *testing.T) {
	engine := NewEngine()
	trades := buildAbsorptionIcebergAggTrades(engine.TradeMinimumRequired())

	result, err := engine.AnalyzeAggTrades("BTCUSDT", trades)
	if err != nil {
		t.Fatalf("analyze agg trades failed: %v", err)
	}

	if result.AbsorptionBias != "buy_absorption" {
		t.Fatalf("expected buy_absorption, got %s", result.AbsorptionBias)
	}
	if result.AbsorptionStrength <= 0 {
		t.Fatalf("expected positive absorption strength, got %f", result.AbsorptionStrength)
	}
	if result.IcebergBias != "buy_iceberg" {
		t.Fatalf("expected buy_iceberg, got %s", result.IcebergBias)
	}
	if result.IcebergStrength <= 0 {
		t.Fatalf("expected positive iceberg strength, got %f", result.IcebergStrength)
	}
	if len(result.MicrostructureEvents) < 3 {
		t.Fatalf("expected at least 3 microstructure events, got %d", len(result.MicrostructureEvents))
	}
	if !hasMicrostructureEvent(result.MicrostructureEvents, "absorption", "bullish") {
		t.Fatalf("expected bullish absorption event, got %#v", result.MicrostructureEvents)
	}
	if !hasMicrostructureEvent(result.MicrostructureEvents, "iceberg", "bullish") {
		t.Fatalf("expected bullish iceberg event, got %#v", result.MicrostructureEvents)
	}
	if !hasMicrostructureEvent(result.MicrostructureEvents, "large_trade_cluster", "bullish") {
		t.Fatalf("expected bullish large trade cluster event, got %#v", result.MicrostructureEvents)
	}
	if !hasMicrostructureEvent(result.MicrostructureEvents, "continuous_absorption", "bullish") {
		t.Fatalf("expected bullish continuous absorption event, got %#v", result.MicrostructureEvents)
	}
}

func TestAnalyzeAggTradesDetectsFailedAuction(t *testing.T) {
	engine := NewEngine()
	trades := buildFailedAuctionAggTrades(engine.TradeMinimumRequired())

	result, err := engine.AnalyzeAggTrades("BTCUSDT", trades)
	if err != nil {
		t.Fatalf("analyze agg trades failed: %v", err)
	}

	if !hasMicrostructureEvent(result.MicrostructureEvents, "failed_auction", "bearish") {
		t.Fatalf("expected bearish failed auction event, got %#v", result.MicrostructureEvents)
	}
}

func TestAnalyzeOrderBookMicrostructureDetectsMigration(t *testing.T) {
	engine := NewEngine()
	snapshots := buildMigrationOrderBookSnapshots(t)

	events, err := engine.AnalyzeOrderBookMicrostructure("BTCUSDT", snapshots)
	if err != nil {
		t.Fatalf("analyze order book microstructure failed: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("expected order book migration event")
	}
	if !hasMicrostructureEvent(events, "order_book_migration", "bullish") {
		t.Fatalf("expected bullish order book migration event, got %#v", events)
	}
}

func hasMicrostructureEvent(
	events []models.OrderFlowMicrostructureEvent,
	eventType, bias string,
) bool {
	for _, event := range events {
		if event.Type == eventType && event.Bias == bias {
			return true
		}
	}
	return false
}

func buildBullishKlines(limit int) []models.Kline {
	klines := make([]models.Kline, 0, limit)
	base := 63000.0
	start := time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC)

	for i := 0; i < limit; i++ {
		move := float64(i)*55 + math.Sin(float64(i)/6.0)*80
		open := base + move
		close := open + 45 + math.Cos(float64(i)/5.0)*8
		high := close + 28
		low := open - 22
		volume := 80 + float64(i)*2.4

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

func buildBullishAggTrades(limit int) []models.AggTrade {
	trades := make([]models.AggTrade, 0, limit)
	start := time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC)
	price := 63200.0

	for i := 0; i < limit; i++ {
		tradePrice := price + float64(i)*3.2 + math.Sin(float64(i)/9.0)*12
		quantity := 0.18 + math.Mod(float64(i), 7)*0.03
		if i%14 == 0 {
			quantity = 2.4 + math.Mod(float64(i), 5)*0.22
		}
		isBuyerMaker := i%5 == 0

		trades = append(trades, models.AggTrade{
			Symbol:           "BTCUSDT",
			AggTradeID:       int64(i + 1),
			Price:            tradePrice,
			Quantity:         quantity,
			QuoteQuantity:    tradePrice * quantity,
			FirstTradeID:     int64(i*2 + 1),
			LastTradeID:      int64(i*2 + 2),
			TradeTime:        start.Add(time.Duration(i) * 2 * time.Second).UnixMilli(),
			IsBuyerMaker:     isBuyerMaker,
			IsBestPriceMatch: true,
			CreatedAt:        start.Add(time.Duration(i) * 2 * time.Second),
		})
	}

	return trades
}

func buildAbsorptionIcebergAggTrades(limit int) []models.AggTrade {
	trades := make([]models.AggTrade, 0, limit)
	start := time.Date(2026, 3, 6, 1, 0, 0, 0, time.UTC)
	basePrice := 64000.0

	for i := 0; i < limit; i++ {
		price := basePrice + math.Sin(float64(i)/8.0)*18
		quantity := 0.45 + math.Mod(float64(i), 5)*0.05
		isBuyerMaker := true

		// 构造反复在同一价带出现的卖出大单，同时价格保持稳定，模拟买方吸收/买方冰山单。
		if i%9 == 0 || i%9 == 3 || i%9 == 6 {
			price = basePrice + 4
			quantity = 2.05 + math.Mod(float64(i), 3)*0.08
		}
		if i%10 == 1 {
			isBuyerMaker = false
			quantity = 0.22
			price = basePrice + 10
		}

		trades = append(trades, models.AggTrade{
			Symbol:           "BTCUSDT",
			AggTradeID:       int64(i + 1000),
			Price:            price,
			Quantity:         quantity,
			QuoteQuantity:    price * quantity,
			FirstTradeID:     int64(i*2 + 100),
			LastTradeID:      int64(i*2 + 101),
			TradeTime:        start.Add(time.Duration(i) * 1500 * time.Millisecond).UnixMilli(),
			IsBuyerMaker:     isBuyerMaker,
			IsBestPriceMatch: true,
			CreatedAt:        start.Add(time.Duration(i) * 1500 * time.Millisecond),
		})
	}

	return trades
}

func buildFailedAuctionAggTrades(limit int) []models.AggTrade {
	trades := make([]models.AggTrade, 0, limit)
	start := time.Date(2026, 3, 6, 2, 0, 0, 0, time.UTC)
	basePrice := 64600.0
	extensionPrices := []float64{64612, 64618, 64624, 64630, 64652, 64678, 64705, 64688, 64655, 64624, 64598, 64596}

	prefixCount := maxInt(limit-len(extensionPrices), 24)
	for i := 0; i < prefixCount; i++ {
		price := basePrice + math.Sin(float64(i)/4.0)*8
		quantity := 0.34 + math.Mod(float64(i), 4)*0.03
		trades = append(trades, models.AggTrade{
			Symbol:           "BTCUSDT",
			AggTradeID:       int64(i + 2000),
			Price:            price,
			Quantity:         quantity,
			QuoteQuantity:    price * quantity,
			FirstTradeID:     int64(i*2 + 500),
			LastTradeID:      int64(i*2 + 501),
			TradeTime:        start.Add(time.Duration(i) * 1500 * time.Millisecond).UnixMilli(),
			IsBuyerMaker:     i%3 == 0,
			IsBestPriceMatch: true,
			CreatedAt:        start.Add(time.Duration(i) * 1500 * time.Millisecond),
		})
	}

	for i, price := range extensionPrices {
		quantity := 0.24
		isBuyerMaker := i >= 7
		if i < 7 {
			quantity = 1.05 + float64(i)*0.04
		}
		trades = append(trades, models.AggTrade{
			Symbol:           "BTCUSDT",
			AggTradeID:       int64(prefixCount + i + 3000),
			Price:            price,
			Quantity:         quantity,
			QuoteQuantity:    price * quantity,
			FirstTradeID:     int64((prefixCount+i)*2 + 800),
			LastTradeID:      int64((prefixCount+i)*2 + 801),
			TradeTime:        start.Add(time.Duration(prefixCount+i) * 1500 * time.Millisecond).UnixMilli(),
			IsBuyerMaker:     isBuyerMaker,
			IsBestPriceMatch: true,
			CreatedAt:        start.Add(time.Duration(prefixCount+i) * 1500 * time.Millisecond),
		})
	}

	return trades
}

func buildMigrationOrderBookSnapshots(t *testing.T) []models.OrderBookSnapshot {
	t.Helper()

	baseTime := time.Date(2026, 3, 6, 3, 0, 0, 0, time.UTC)
	snapshots := make([]models.OrderBookSnapshot, 0, 6)
	for i := 0; i < 6; i++ {
		bids := []binancepkg.OrderBookLevel{
			{Price: 64500 + float64(i)*7, Quantity: 1.4},
			{Price: 64496 + float64(i)*7, Quantity: 2.2},
			{Price: 64492 + float64(i)*7, Quantity: 7.1 + float64(i)*0.4},
			{Price: 64488 + float64(i)*7, Quantity: 4.5},
		}
		asks := []binancepkg.OrderBookLevel{
			{Price: 64520 + float64(i), Quantity: 1.0},
			{Price: 64524 + float64(i), Quantity: 2.0},
			{Price: 64528 + float64(i), Quantity: 6.1},
			{Price: 64532 + float64(i), Quantity: 4.1},
		}
		bidsJSON, err := json.Marshal(bids)
		if err != nil {
			t.Fatalf("marshal bids failed: %v", err)
		}
		asksJSON, err := json.Marshal(asks)
		if err != nil {
			t.Fatalf("marshal asks failed: %v", err)
		}

		snapshots = append(snapshots, models.OrderBookSnapshot{
			Symbol:       "BTCUSDT",
			LastUpdateID: int64(8000 + i),
			DepthLevel:   20,
			BidsJSON:     string(bidsJSON),
			AsksJSON:     string(asksJSON),
			BestBidPrice: bids[0].Price,
			BestAskPrice: asks[0].Price,
			Spread:       asks[0].Price - bids[0].Price,
			EventTime:    baseTime.Add(time.Duration(i) * 2 * time.Second).UnixMilli(),
			CreatedAt:    baseTime.Add(time.Duration(i) * 2 * time.Second),
		})
	}
	return snapshots
}
