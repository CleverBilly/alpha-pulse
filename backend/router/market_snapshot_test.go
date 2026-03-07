package router_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"alpha-pulse/backend/internal/ai"
	"alpha-pulse/backend/internal/collector"
	"alpha-pulse/backend/internal/handler"
	"alpha-pulse/backend/internal/indicator"
	"alpha-pulse/backend/internal/liquidity"
	"alpha-pulse/backend/internal/orderflow"
	"alpha-pulse/backend/internal/service"
	signalengine "alpha-pulse/backend/internal/signal"
	structureengine "alpha-pulse/backend/internal/structure"
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/pkg/binance"
	"alpha-pulse/backend/repository"
	routerpkg "alpha-pulse/backend/router"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// apiEnvelope 对应项目统一响应结构。
type apiEnvelope[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func TestMarketSnapshotEndpointReturnsAggregatedPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	r := newTestRouter(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/market-snapshot?symbol=BTCUSDT&interval=5m&limit=24", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload apiEnvelope[service.MarketSnapshot]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if payload.Code != 0 {
		t.Fatalf("unexpected business code: got=%d message=%s", payload.Code, payload.Message)
	}

	snapshot := payload.Data
	if snapshot.Price.Symbol != "BTCUSDT" {
		t.Fatalf("unexpected price symbol: got=%s", snapshot.Price.Symbol)
	}
	if snapshot.Price.Price <= 0 {
		t.Fatalf("price should be positive: got=%f", snapshot.Price.Price)
	}
	if len(snapshot.Klines) != 24 {
		t.Fatalf("unexpected kline length: got=%d want=24", len(snapshot.Klines))
	}
	assertAscendingOpenTime(t, snapshot.Klines)
	for _, item := range snapshot.Klines {
		if item.Symbol != "BTCUSDT" {
			t.Fatalf("unexpected kline symbol: got=%s", item.Symbol)
		}
		if item.IntervalType != "5m" {
			t.Fatalf("unexpected kline interval: got=%s", item.IntervalType)
		}
	}

	latest := snapshot.Klines[len(snapshot.Klines)-1]
	if diff := math.Abs(snapshot.Signal.EntryPrice - latest.ClosePrice); diff > 0.0001 {
		t.Fatalf("signal entry price should match latest close: diff=%f entry=%f close=%f", diff, snapshot.Signal.EntryPrice, latest.ClosePrice)
	}
	if snapshot.Indicator.EMA20 <= 0 || snapshot.Indicator.EMA50 <= 0 || snapshot.Indicator.VWAP <= 0 {
		t.Fatalf("indicator values should be positive: %+v", snapshot.Indicator)
	}
	if snapshot.Indicator.BollingerUpper <= snapshot.Indicator.BollingerLower {
		t.Fatalf("bollinger bands are invalid: upper=%f lower=%f", snapshot.Indicator.BollingerUpper, snapshot.Indicator.BollingerLower)
	}
	if len(snapshot.IndicatorSeries) != 24 {
		t.Fatalf("unexpected indicator series length: got=%d want=24", len(snapshot.IndicatorSeries))
	}
	assertAscendingIndicatorSeriesOpenTime(t, snapshot.IndicatorSeries)
	latestIndicatorPoint := snapshot.IndicatorSeries[len(snapshot.IndicatorSeries)-1]
	if math.Abs(latestIndicatorPoint.EMA20-snapshot.Indicator.EMA20) > 0.0001 {
		t.Fatalf("latest indicator series EMA20 should match snapshot indicator: series=%f indicator=%f", latestIndicatorPoint.EMA20, snapshot.Indicator.EMA20)
	}
	if math.Abs(latestIndicatorPoint.VWAP-snapshot.Indicator.VWAP) > 0.0001 {
		t.Fatalf("latest indicator series VWAP should match snapshot indicator: series=%f indicator=%f", latestIndicatorPoint.VWAP, snapshot.Indicator.VWAP)
	}
	if snapshot.OrderFlow.BuyVolume <= 0 || snapshot.OrderFlow.SellVolume <= 0 {
		t.Fatalf("order flow volumes should be positive: %+v", snapshot.OrderFlow)
	}
	if snapshot.OrderFlow.DataSource == "" {
		t.Fatal("order flow data source should not be empty")
	}
	if snapshot.OrderFlow.IntervalType != "5m" {
		t.Fatalf("unexpected order flow interval: got=%s", snapshot.OrderFlow.IntervalType)
	}
	if snapshot.OrderFlow.OpenTime != latest.OpenTime {
		t.Fatalf("order flow open time should match latest kline: got=%d want=%d", snapshot.OrderFlow.OpenTime, latest.OpenTime)
	}
	if len(snapshot.OrderFlow.MicrostructureEvents) == 0 {
		t.Fatal("order flow microstructure events should not be empty")
	}
	if len(snapshot.MicrostructureEvents) == 0 {
		t.Fatal("snapshot microstructure history should not be empty")
	}
	assertAscendingMicrostructureTradeTime(t, snapshot.MicrostructureEvents)
	if snapshot.Structure.Trend == "" {
		t.Fatal("structure trend should not be empty")
	}
	if len(snapshot.Structure.Events) == 0 {
		t.Fatal("structure events should not be empty")
	}
	if len(snapshot.StructureSeries) != 24 {
		t.Fatalf("unexpected structure series length: got=%d want=24", len(snapshot.StructureSeries))
	}
	if len(snapshot.LiquiditySeries) != 24 {
		t.Fatalf("unexpected liquidity series length: got=%d want=24", len(snapshot.LiquiditySeries))
	}
	assertAscendingSeriesOpenTime(t, snapshot.StructureSeries)
	assertAscendingLiquiditySeriesOpenTime(t, snapshot.LiquiditySeries)
	if snapshot.Liquidity.SweepType == "" {
		t.Fatal("liquidity sweep type should not be empty")
	}
	if len(snapshot.Liquidity.StopClusters) == 0 {
		t.Fatal("liquidity stop clusters should not be empty")
	}
	if snapshot.Signal.Action == "" {
		t.Fatal("signal action should not be empty")
	}
	if len(snapshot.SignalTimeline) == 0 {
		t.Fatal("signal timeline should not be empty")
	}
	assertAscendingSignalTimelineOpenTime(t, snapshot.SignalTimeline)
	latestSignalPoint := snapshot.SignalTimeline[len(snapshot.SignalTimeline)-1]
	if latestSignalPoint.Signal != snapshot.Signal.Action {
		t.Fatalf("latest timeline signal should match snapshot signal: timeline=%s snapshot=%s", latestSignalPoint.Signal, snapshot.Signal.Action)
	}
	if latestSignalPoint.IntervalType != "5m" {
		t.Fatalf("unexpected signal timeline interval: got=%s", latestSignalPoint.IntervalType)
	}
	if len(snapshot.Signal.Factors) < 5 {
		t.Fatalf("signal factors are too few: got=%d", len(snapshot.Signal.Factors))
	}
	if strings.TrimSpace(snapshot.Signal.Explain) == "" {
		t.Fatal("signal explain should not be empty")
	}
	if snapshot.Signal.Confidence <= 0 {
		t.Fatalf("signal confidence should be positive: got=%d", snapshot.Signal.Confidence)
	}
}

func TestMarketSnapshotEndpointMaintainsJSONContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	r := newTestRouter(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/market-snapshot?symbol=BTCUSDT&interval=5m&limit=24", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode raw response failed: %v", err)
	}

	assertJSONKeys(t, payload, "code", "message", "data")

	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data to be object, got=%T", payload["data"])
	}
	assertJSONKeys(
		t,
		data,
		"price",
		"klines",
		"indicator",
		"indicator_series",
		"orderflow",
		"microstructure_events",
		"structure",
		"structure_series",
		"liquidity",
		"liquidity_series",
		"signal",
		"signal_timeline",
	)

	price, ok := data["price"].(map[string]any)
	if !ok {
		t.Fatalf("expected price to be object, got=%T", data["price"])
	}
	assertJSONKeys(t, price, "symbol", "price", "time")

	orderFlow, ok := data["orderflow"].(map[string]any)
	if !ok {
		t.Fatalf("expected orderflow to be object, got=%T", data["orderflow"])
	}
	assertJSONKeys(
		t,
		orderFlow,
		"buy_volume",
		"sell_volume",
		"delta",
		"cvd",
		"absorption_bias",
		"absorption_strength",
		"iceberg_bias",
		"iceberg_strength",
		"data_source",
		"microstructure_events",
	)

	liquidity, ok := data["liquidity"].(map[string]any)
	if !ok {
		t.Fatalf("expected liquidity to be object, got=%T", data["liquidity"])
	}
	assertJSONKeys(
		t,
		liquidity,
		"buy_liquidity",
		"sell_liquidity",
		"sweep_type",
		"order_book_imbalance",
		"data_source",
		"equal_high",
		"equal_low",
		"stop_clusters",
		"wall_levels",
	)

	wallLevels, ok := liquidity["wall_levels"].([]any)
	if !ok {
		t.Fatalf("expected non-empty liquidity.wall_levels array, got=%T", liquidity["wall_levels"])
	}
	if len(wallLevels) > 0 {
		firstWall, ok := wallLevels[0].(map[string]any)
		if !ok {
			t.Fatalf("expected first wall level to be object, got=%T", wallLevels[0])
		}
		assertJSONKeys(
			t,
			firstWall,
			"label",
			"kind",
			"side",
			"layer",
			"price",
			"quantity",
			"notional",
			"distance_bps",
			"strength",
		)
	}

	microEvents, ok := data["microstructure_events"].([]any)
	if !ok || len(microEvents) == 0 {
		t.Fatalf("expected non-empty microstructure_events array, got=%T", data["microstructure_events"])
	}
	firstMicro, ok := microEvents[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first microstructure event to be object, got=%T", microEvents[0])
	}
	assertJSONKeys(t, firstMicro, "type", "bias", "score", "strength", "price", "trade_time", "detail")

	signalTimeline, ok := data["signal_timeline"].([]any)
	if !ok || len(signalTimeline) == 0 {
		t.Fatalf("expected non-empty signal_timeline array, got=%T", data["signal_timeline"])
	}
	firstSignal, ok := signalTimeline[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first signal timeline point to be object, got=%T", signalTimeline[0])
	}
	assertJSONKeys(t, firstSignal, "interval_type", "open_time", "signal", "score", "confidence", "entry_price", "stop_loss", "target_price")
}

func TestMarketSnapshotEndpointPersistsAnalysisRowsAndAppliesFallbackLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	r := newTestRouter(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/market-snapshot?symbol=ETHUSDT&interval=15m&limit=0", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload apiEnvelope[service.MarketSnapshot]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("unexpected business code: got=%d message=%s", payload.Code, payload.Message)
	}

	if len(payload.Data.Klines) != 48 {
		t.Fatalf("invalid limit should fallback to 48 klines: got=%d", len(payload.Data.Klines))
	}

	expectedHistory := maxInt(
		indicator.NewEngine().HistoryLimit(),
		orderflow.NewEngine().HistoryLimit(),
		structureengine.NewEngine().HistoryLimit(),
		liquidity.NewEngine().HistoryLimit(),
	)

	assertCount(t, db, &models.Kline{}, "symbol = ? AND interval_type = ?", int64(expectedHistory), "ETHUSDT", "15m")
	assertCount(t, db, &models.Indicator{}, "symbol = ?", 1, "ETHUSDT")
	assertCount(t, db, &models.OrderFlow{}, "symbol = ?", 1, "ETHUSDT")
	assertMinCount(t, db, &models.MicrostructureEvent{}, "symbol = ? AND interval_type = ?", 1, "ETHUSDT", "15m")
	assertCount(t, db, &models.Structure{}, "symbol = ?", 1, "ETHUSDT")
	assertCount(t, db, &models.Liquidity{}, "symbol = ?", 1, "ETHUSDT")
	assertCount(t, db, &models.Signal{}, "symbol = ?", 1, "ETHUSDT")
}

func TestMarketStructureEventsEndpointReturnsEventStream(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	r := newTestRouter(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/market-structure-events?symbol=BTCUSDT&interval=5m", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload apiEnvelope[service.StructureEventsResult]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("unexpected business code: got=%d message=%s", payload.Code, payload.Message)
	}

	result := payload.Data
	if result.Symbol != "BTCUSDT" {
		t.Fatalf("unexpected symbol: got=%s", result.Symbol)
	}
	if result.Interval != "5m" {
		t.Fatalf("unexpected interval: got=%s", result.Interval)
	}
	if result.Trend == "" {
		t.Fatal("structure trend should not be empty")
	}
	if len(result.Events) == 0 {
		t.Fatal("structure events should not be empty")
	}
}

func TestMarketStructureSeriesEndpointReturnsSeriesPoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	r := newTestRouter(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/market-structure-series?symbol=BTCUSDT&interval=5m&limit=18", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload apiEnvelope[service.StructureSeriesResult]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("unexpected business code: got=%d message=%s", payload.Code, payload.Message)
	}

	result := payload.Data
	if result.Symbol != "BTCUSDT" {
		t.Fatalf("unexpected symbol: got=%s", result.Symbol)
	}
	if result.Interval != "5m" {
		t.Fatalf("unexpected interval: got=%s", result.Interval)
	}
	if len(result.Points) != 18 {
		t.Fatalf("unexpected point length: got=%d want=18", len(result.Points))
	}
	assertAscendingSeriesOpenTime(t, result.Points)
	if result.Points[len(result.Points)-1].Trend == "" {
		t.Fatal("latest structure trend should not be empty")
	}
}

func TestIndicatorSeriesEndpointReturnsSeriesPoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	r := newTestRouter(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/indicator-series?symbol=BTCUSDT&interval=5m&limit=20", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload apiEnvelope[service.IndicatorSeriesResult]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("unexpected business code: got=%d message=%s", payload.Code, payload.Message)
	}

	result := payload.Data
	if result.Symbol != "BTCUSDT" {
		t.Fatalf("unexpected symbol: got=%s", result.Symbol)
	}
	if result.Interval != "5m" {
		t.Fatalf("unexpected interval: got=%s", result.Interval)
	}
	if len(result.Points) != 20 {
		t.Fatalf("unexpected point length: got=%d want=20", len(result.Points))
	}
	assertAscendingIndicatorSeriesOpenTime(t, result.Points)
	latest := result.Points[len(result.Points)-1]
	if latest.EMA20 <= 0 || latest.EMA50 <= 0 || latest.VWAP <= 0 {
		t.Fatalf("latest indicator series point should be positive: %+v", latest)
	}
}

func TestLiquidityMapEndpointReturnsLiquidityClusters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	r := newTestRouter(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/liquidity-map?symbol=ETHUSDT&interval=15m", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload apiEnvelope[service.LiquidityMapResult]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("unexpected business code: got=%d message=%s", payload.Code, payload.Message)
	}

	result := payload.Data
	if result.Symbol != "ETHUSDT" {
		t.Fatalf("unexpected symbol: got=%s", result.Symbol)
	}
	if result.Interval != "15m" {
		t.Fatalf("unexpected interval: got=%s", result.Interval)
	}
	if result.BuyLiquidity <= 0 || result.SellLiquidity <= 0 {
		t.Fatalf("liquidity levels should be positive: %+v", result)
	}
	if len(result.StopClusters) == 0 {
		t.Fatal("liquidity stop clusters should not be empty")
	}
	if result.WallLevels == nil {
		t.Fatal("expected wall levels field to be initialized")
	}
	if len(result.WallLevels) > 0 && (result.WallLevels[0].Side == "" || result.WallLevels[0].Layer == "") {
		t.Fatalf("expected wall level side/layer to be populated: %+v", result.WallLevels[0])
	}
}

func TestLiquiditySeriesEndpointReturnsSeriesPoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	r := newTestRouter(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/liquidity-series?symbol=ETHUSDT&interval=15m&limit=16", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload apiEnvelope[service.LiquiditySeriesResult]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("unexpected business code: got=%d message=%s", payload.Code, payload.Message)
	}

	result := payload.Data
	if result.Symbol != "ETHUSDT" {
		t.Fatalf("unexpected symbol: got=%s", result.Symbol)
	}
	if result.Interval != "15m" {
		t.Fatalf("unexpected interval: got=%s", result.Interval)
	}
	if len(result.Points) != 16 {
		t.Fatalf("unexpected point length: got=%d want=16", len(result.Points))
	}
	assertAscendingLiquiditySeriesOpenTime(t, result.Points)
	if result.Points[len(result.Points)-1].BuyLiquidity <= 0 {
		t.Fatal("latest buy liquidity should be positive")
	}
}

func TestSignalTimelineEndpointReturnsSignalHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	r := newTestRouter(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/signal-timeline?symbol=BTCUSDT&interval=5m&limit=12", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload apiEnvelope[service.SignalTimelineResult]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("unexpected business code: got=%d message=%s", payload.Code, payload.Message)
	}

	result := payload.Data
	if result.Symbol != "BTCUSDT" {
		t.Fatalf("unexpected symbol: got=%s", result.Symbol)
	}
	if result.Interval != "5m" {
		t.Fatalf("unexpected interval: got=%s", result.Interval)
	}
	if len(result.Points) == 0 {
		t.Fatal("signal timeline should not be empty")
	}
	assertAscendingSignalTimelineOpenTime(t, result.Points)
	latest := result.Points[len(result.Points)-1]
	if latest.IntervalType != "5m" {
		t.Fatalf("unexpected timeline point interval: got=%s", latest.IntervalType)
	}
	if latest.Signal == "" {
		t.Fatal("latest timeline signal should not be empty")
	}
}

func TestMicrostructureEventsEndpointReturnsHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	r := newTestRouter(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/microstructure-events?symbol=BTCUSDT&interval=5m&limit=6", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload apiEnvelope[service.MicrostructureEventsResult]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("unexpected business code: got=%d message=%s", payload.Code, payload.Message)
	}

	result := payload.Data
	if result.Symbol != "BTCUSDT" {
		t.Fatalf("unexpected symbol: got=%s", result.Symbol)
	}
	if result.Interval != "5m" {
		t.Fatalf("unexpected interval: got=%s", result.Interval)
	}
	if len(result.Events) == 0 {
		t.Fatal("microstructure events should not be empty")
	}
	assertAscendingMicrostructureTradeTime(t, result.Events)
	if result.Events[len(result.Events)-1].EventType == "" {
		t.Fatal("latest microstructure event type should not be empty")
	}
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", sanitizeName(t.Name()))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db failed: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	return db
}

func newTestRouter(t *testing.T, db *gorm.DB) *gin.Engine {
	t.Helper()

	binanceClient := binance.NewClient("", "", 20*time.Millisecond)
	binanceClient.SetHTTPClient(binance.NewFailingHTTPClient(errors.New("test forces sdk fallback path")))
	binanceCollector := collector.NewBinanceCollector(binanceClient)
	indicatorEngine := indicator.NewEngine()
	orderFlowEngine := orderflow.NewEngine()
	structureEngine := structureengine.NewEngine()
	liquidityEngine := liquidity.NewEngine()
	signalEngine := signalengine.NewEngine()
	explainEngine := ai.NewEngine()

	klineRepo := repository.NewKlineRepository(db)
	aggTradeRepo := repository.NewAggTradeRepository(db)
	orderBookRepo := repository.NewOrderBookSnapshotRepository(db)
	indicatorRepo := repository.NewIndicatorRepository(db)
	signalRepo := repository.NewSignalRepository(db)
	microEventRepo := repository.NewMicrostructureEventRepository(db)

	marketService := service.NewMarketService(
		db,
		binanceCollector,
		indicatorEngine,
		orderFlowEngine,
		structureEngine,
		liquidityEngine,
		klineRepo,
		aggTradeRepo,
		orderBookRepo,
		indicatorRepo,
		microEventRepo,
	)
	signalService := service.NewSignalService(
		db,
		binanceCollector,
		indicatorEngine,
		orderFlowEngine,
		structureEngine,
		liquidityEngine,
		signalEngine,
		explainEngine,
		klineRepo,
		aggTradeRepo,
		orderBookRepo,
		indicatorRepo,
		signalRepo,
		microEventRepo,
		nil,
		0,
	)

	marketHandler := handler.NewMarketHandler(marketService, signalService)
	signalHandler := handler.NewSignalHandler(signalService)

	return routerpkg.SetupRouter(routerpkg.HandlerSet{
		Market: marketHandler,
		Signal: signalHandler,
	})
}

func assertAscendingOpenTime(t *testing.T, klines []models.Kline) {
	t.Helper()

	for index := 1; index < len(klines); index++ {
		if klines[index-1].OpenTime > klines[index].OpenTime {
			t.Fatalf(
				"klines should be sorted ascending by open_time: prev=%d current=%d index=%d",
				klines[index-1].OpenTime,
				klines[index].OpenTime,
				index,
			)
		}
	}
}

func assertAscendingSeriesOpenTime(t *testing.T, points []service.StructureSeriesPoint) {
	t.Helper()
	for index := 1; index < len(points); index++ {
		if points[index].OpenTime <= points[index-1].OpenTime {
			t.Fatalf("structure series should be sorted ascending: prev=%d current=%d", points[index-1].OpenTime, points[index].OpenTime)
		}
	}
}

func assertAscendingIndicatorSeriesOpenTime(t *testing.T, points []models.IndicatorSeriesPoint) {
	t.Helper()
	for index := 1; index < len(points); index++ {
		if points[index].OpenTime <= points[index-1].OpenTime {
			t.Fatalf("indicator series should be sorted ascending: prev=%d current=%d", points[index-1].OpenTime, points[index].OpenTime)
		}
	}
}

func assertAscendingSignalTimelineOpenTime(t *testing.T, points []models.SignalTimelinePoint) {
	t.Helper()
	for index := 1; index < len(points); index++ {
		if points[index].OpenTime <= points[index-1].OpenTime {
			t.Fatalf("signal timeline should be sorted ascending: prev=%d current=%d", points[index-1].OpenTime, points[index].OpenTime)
		}
	}
}

func assertAscendingMicrostructureTradeTime(t *testing.T, events []models.MicrostructureEvent) {
	t.Helper()
	for index := 1; index < len(events); index++ {
		if events[index].TradeTime < events[index-1].TradeTime {
			t.Fatalf("microstructure events should be sorted ascending: prev=%d current=%d", events[index-1].TradeTime, events[index].TradeTime)
		}
	}
}

func assertAscendingLiquiditySeriesOpenTime(t *testing.T, points []service.LiquiditySeriesPoint) {
	t.Helper()
	for index := 1; index < len(points); index++ {
		if points[index].OpenTime <= points[index-1].OpenTime {
			t.Fatalf("liquidity series should be sorted ascending: prev=%d current=%d", points[index-1].OpenTime, points[index].OpenTime)
		}
	}
}

func assertCount(t *testing.T, db *gorm.DB, model any, query string, expected int64, args ...any) {
	t.Helper()

	var count int64
	if err := db.Model(model).Where(query, args...).Count(&count).Error; err != nil {
		t.Fatalf("count model failed: %v", err)
	}
	if count != expected {
		t.Fatalf("unexpected count for %T: got=%d want=%d", model, count, expected)
	}
}

func assertMinCount(t *testing.T, db *gorm.DB, model any, query string, minimum int64, args ...any) {
	t.Helper()

	var count int64
	if err := db.Model(model).Where(query, args...).Count(&count).Error; err != nil {
		t.Fatalf("count model failed: %v", err)
	}
	if count < minimum {
		t.Fatalf("unexpected count for %T: got=%d minimum=%d", model, count, minimum)
	}
}

func sanitizeName(name string) string {
	replacer := strings.NewReplacer("/", "_", " ", "_", "=", "_", "?", "_")
	return replacer.Replace(name)
}

func maxInt(values ...int) int {
	maxValue := values[0]
	for _, value := range values[1:] {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}

func assertJSONKeys(t *testing.T, object map[string]any, keys ...string) {
	t.Helper()

	for _, key := range keys {
		if _, exists := object[key]; !exists {
			t.Fatalf("expected JSON object to contain key %q, object=%v", key, object)
		}
	}
}
