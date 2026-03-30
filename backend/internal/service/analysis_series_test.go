package service

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"alpha-pulse/backend/internal/liquidity"
	"alpha-pulse/backend/models"
	binancepkg "alpha-pulse/backend/pkg/binance"
)

type orderBookWindowRepoStub struct {
	snapshots  []models.OrderBookSnapshot
	calls      int
	lastSymbol string
	lastStart  int64
	lastEnd    int64
}

func (s *orderBookWindowRepoStub) GetSeriesWindow(symbol string, startTime, endTime int64) ([]models.OrderBookSnapshot, error) {
	s.calls++
	s.lastSymbol = symbol
	s.lastStart = startTime
	s.lastEnd = endTime

	result := make([]models.OrderBookSnapshot, len(s.snapshots))
	copy(result, s.snapshots)
	return result, nil
}

func TestBuildLiquiditySeriesPreloadsSnapshotsOnce(t *testing.T) {
	engine := liquidity.NewEngine()
	klines := buildLiquiditySeriesTestKlines(30)
	limit := 6
	start := len(klines) - limit
	intervalMillis := intervalDurationMillis("1m")

	repo := &orderBookWindowRepoStub{
		snapshots: []models.OrderBookSnapshot{
			buildLiquiditySeriesSnapshot(t, klines[start].OpenTime+intervalMillis-30_000, 1001),
			buildLiquiditySeriesSnapshot(t, klines[start+1].OpenTime+intervalMillis, 1002),
			buildLiquiditySeriesSnapshot(t, klines[start+3].OpenTime+intervalMillis, 1003),
			buildLiquiditySeriesSnapshot(t, klines[len(klines)-1].OpenTime+intervalMillis, 1004),
		},
	}

	points, err := buildLiquiditySeries(engine, repo, "BTCUSDT", "1m", klines, limit)
	if err != nil {
		t.Fatalf("buildLiquiditySeries returned error: %v", err)
	}

	if len(points) != limit {
		t.Fatalf("expected %d liquidity series points, got %d", limit, len(points))
	}
	if repo.calls != 1 {
		t.Fatalf("expected one snapshot window preload, got %d calls", repo.calls)
	}
	if repo.lastSymbol != "BTCUSDT" {
		t.Fatalf("unexpected preload symbol: got=%s", repo.lastSymbol)
	}

	expectedStart := klines[start].OpenTime + intervalMillis
	expectedEnd := klines[len(klines)-1].OpenTime + intervalMillis
	if repo.lastStart != expectedStart || repo.lastEnd != expectedEnd {
		t.Fatalf("unexpected preload window: got=(%d,%d) want=(%d,%d)", repo.lastStart, repo.lastEnd, expectedStart, expectedEnd)
	}
}

func TestBuildLiquiditySeriesSkipsWideOrderBookHistoryWindows(t *testing.T) {
	engine := liquidity.NewEngine()
	klines := buildLiquiditySeriesTestKlinesWithInterval(60, 4*60*60*1000)
	limit := 25
	repo := &orderBookWindowRepoStub{
		snapshots: []models.OrderBookSnapshot{
			buildLiquiditySeriesSnapshot(t, klines[0].OpenTime+intervalDurationMillis("4h"), 2001),
		},
	}

	points, err := buildLiquiditySeries(engine, repo, "BTCUSDT", "4h", klines, limit)
	if err != nil {
		t.Fatalf("buildLiquiditySeries returned error: %v", err)
	}
	if len(points) != limit {
		t.Fatalf("expected %d liquidity series points, got %d", limit, len(points))
	}
	if repo.calls != 0 {
		t.Fatalf("expected wide historical span to skip order book preload, got %d calls", repo.calls)
	}
}

func TestHotLookupIndexesAreDeclared(t *testing.T) {
	assertTagContains(t, reflect.TypeOf(models.OrderBookSnapshot{}), "Symbol", "idx_order_book_symbol_event_update_lookup")
	assertTagContains(t, reflect.TypeOf(models.OrderBookSnapshot{}), "EventTime", "idx_order_book_symbol_event_update_lookup")
	assertTagContains(t, reflect.TypeOf(models.OrderBookSnapshot{}), "LastUpdateID", "idx_order_book_symbol_event_update_lookup")
	assertTagContains(t, reflect.TypeOf(models.OrderBookSnapshot{}), "ID", "idx_order_book_symbol_event_update_lookup")

	assertTagContains(t, reflect.TypeOf(models.AggTrade{}), "Symbol", "idx_agg_trade_symbol_time_lookup")
	assertTagContains(t, reflect.TypeOf(models.AggTrade{}), "TradeTime", "idx_agg_trade_symbol_time_lookup")
	assertTagContains(t, reflect.TypeOf(models.AggTrade{}), "AggTradeID", "idx_agg_trade_symbol_time_lookup")
}

func assertTagContains(t *testing.T, typ reflect.Type, fieldName, want string) {
	t.Helper()

	field, ok := typ.FieldByName(fieldName)
	if !ok {
		t.Fatalf("field %s not found on %s", fieldName, typ.Name())
	}
	tag := field.Tag.Get("gorm")
	if !strings.Contains(tag, want) {
		t.Fatalf("expected gorm tag for %s.%s to contain %q, got %q", typ.Name(), fieldName, want, tag)
	}
}

func buildLiquiditySeriesTestKlines(count int) []models.Kline {
	return buildLiquiditySeriesTestKlinesWithInterval(count, 60_000)
}

func buildLiquiditySeriesTestKlinesWithInterval(count int, stepMillis int64) []models.Kline {
	klines := make([]models.Kline, 0, count)
	baseOpenTime := int64(1_774_828_800_000)
	for index := 0; index < count; index++ {
		open := 100 + float64(index)*0.3
		klines = append(klines, buildTestKline(
			baseOpenTime+int64(index)*stepMillis,
			open,
			open+1.6,
			open-1.2,
			open+0.4,
		))
	}
	return klines
}

func buildLiquiditySeriesSnapshot(t *testing.T, eventTime int64, updateID int64) models.OrderBookSnapshot {
	t.Helper()

	bids := make([]binancepkg.OrderBookLevel, 0, 8)
	asks := make([]binancepkg.OrderBookLevel, 0, 8)
	for level := 0; level < 8; level++ {
		bids = append(bids, binancepkg.OrderBookLevel{
			Price:    100 - float64(level)*0.1,
			Quantity: 5 + float64(level)*0.4,
		})
		asks = append(asks, binancepkg.OrderBookLevel{
			Price:    100.2 + float64(level)*0.1,
			Quantity: 5.2 + float64(level)*0.4,
		})
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
		ID:           uint64(updateID),
		Symbol:       "BTCUSDT",
		LastUpdateID: updateID,
		DepthLevel:   20,
		BidsJSON:     string(bidsJSON),
		AsksJSON:     string(asksJSON),
		BestBidPrice: bids[0].Price,
		BestAskPrice: asks[0].Price,
		Spread:       asks[0].Price - bids[0].Price,
		EventTime:    eventTime,
	}
}
