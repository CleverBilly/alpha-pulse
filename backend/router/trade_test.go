package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"alpha-pulse/backend/internal/handler"
	"alpha-pulse/backend/internal/service"
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestTradeSettingsEndpointSupportsReadWrite(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	r := newTradeTestRouter(t, db, service.TradeStaticConfig{
		Enabled:        true,
		AutoExecute:    true,
		AllowedSymbols: []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"},
	})

	putBody := []byte(`{"auto_execute_enabled":true,"allowed_symbols":["BTCUSDT","ETHUSDT"],"risk_pct":2.5,"min_risk_reward":1.6,"entry_timeout_seconds":45,"max_open_positions":2,"sync_enabled":true,"updated_by":"tester"}`)
	putReq := httptest.NewRequest(http.MethodPut, "/api/trade-settings", bytes.NewReader(putBody))
	putReq.Header.Set("Content-Type", "application/json")
	putRec := httptest.NewRecorder()
	r.ServeHTTP(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("unexpected put status: got=%d body=%s", putRec.Code, putRec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/trade-settings", nil)
	getRec := httptest.NewRecorder()
	r.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("unexpected get status: got=%d body=%s", getRec.Code, getRec.Body.String())
	}

	var payload apiEnvelope[tradeSettingsPayload]
	if err := json.NewDecoder(getRec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode trade settings payload failed: %v", err)
	}
	if !payload.Data.TradeEnabledEnv || !payload.Data.AutoExecuteEnabled {
		t.Fatalf("unexpected trade settings payload: %+v", payload.Data)
	}
	if len(payload.Data.AllowedSymbols) != 2 || payload.Data.AllowedSymbols[1] != "ETHUSDT" {
		t.Fatalf("unexpected allowed symbols: %+v", payload.Data.AllowedSymbols)
	}
}

func TestTradeRuntimeAndListEndpointsReturnOrderSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	seedTradeOrder(t, db, models.TradeOrder{
		AlertID:         "alert-1",
		Symbol:          "BTCUSDT",
		Side:            "LONG",
		RequestedQty:    0.01,
		EntryStatus:     "pending_fill",
		Status:          "pending_fill",
		Source:          "system",
		CreatedAtUnixMs: 1710000000000,
	})
	seedTradeOrder(t, db, models.TradeOrder{
		AlertID:         "alert-2",
		Symbol:          "BTCUSDT",
		Side:            "LONG",
		RequestedQty:    0.01,
		FilledQty:       0.01,
		EntryStatus:     "filled",
		Status:          "open",
		Source:          "system",
		CreatedAtUnixMs: 1710000001000,
	})

	r := newTradeTestRouter(t, db, service.TradeStaticConfig{
		Enabled:        true,
		AutoExecute:    true,
		AllowedSymbols: []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"},
	})

	listReq := httptest.NewRequest(http.MethodGet, "/api/trades?symbol=BTCUSDT", nil)
	listRec := httptest.NewRecorder()
	r.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("unexpected list status: got=%d body=%s", listRec.Code, listRec.Body.String())
	}

	var listPayload apiEnvelope[[]models.TradeOrder]
	if err := json.NewDecoder(listRec.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode trade list failed: %v", err)
	}
	if len(listPayload.Data) != 2 {
		t.Fatalf("unexpected trade list length: %d", len(listPayload.Data))
	}

	runtimeReq := httptest.NewRequest(http.MethodGet, "/api/trades/runtime", nil)
	runtimeRec := httptest.NewRecorder()
	r.ServeHTTP(runtimeRec, runtimeReq)
	if runtimeRec.Code != http.StatusOK {
		t.Fatalf("unexpected runtime status: got=%d body=%s", runtimeRec.Code, runtimeRec.Body.String())
	}

	var runtimePayload apiEnvelope[tradeRuntimePayload]
	if err := json.NewDecoder(runtimeRec.Body).Decode(&runtimePayload); err != nil {
		t.Fatalf("decode trade runtime payload failed: %v", err)
	}
	if runtimePayload.Data.PendingFillCount != 1 || runtimePayload.Data.OpenCount != 1 {
		t.Fatalf("unexpected runtime summary: %+v", runtimePayload.Data)
	}
}

func TestTradeCloseEndpointHonorsTradeEnabledGate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	order := seedTradeOrder(t, db, models.TradeOrder{
		AlertID:         "alert-2",
		Symbol:          "BTCUSDT",
		Side:            "LONG",
		RequestedQty:    0.01,
		FilledQty:       0.01,
		EntryStatus:     "filled",
		Status:          "open",
		Source:          "system",
		CreatedAtUnixMs: 1710000001000,
	})

	r := newTradeTestRouter(t, db, service.TradeStaticConfig{
		Enabled:        false,
		AutoExecute:    false,
		AllowedSymbols: []string{"BTCUSDT"},
	})

	closeReq := httptest.NewRequest(http.MethodPost, "/api/trades/"+strconv.FormatUint(order.ID, 10)+"/close", nil)
	closeRec := httptest.NewRecorder()
	r.ServeHTTP(closeRec, closeReq)
	if closeRec.Code != http.StatusForbidden {
		t.Fatalf("expected close to be forbidden, got=%d body=%s", closeRec.Code, closeRec.Body.String())
	}
}

func newTradeTestRouter(t *testing.T, db *gorm.DB, static service.TradeStaticConfig) *gin.Engine {
	t.Helper()

	settingsRepo := repository.NewTradeSettingRepository(db)
	orderRepo := repository.NewTradeOrderRepository(db)
	tradeClient := &routerTradeClientStub{}
	executor := service.NewTradeExecutorService(tradeClient, orderRepo)
	runtime := service.NewTradeRuntime(tradeClient, orderRepo)
	tradeService := service.NewTradeService(static, settingsRepo, orderRepo, executor, runtime)
	tradeHandler := handler.NewTradeHandler(tradeService)

	r := gin.New()
	api := r.Group("/api")
	api.GET("/trade-settings", tradeHandler.GetTradeSettings)
	api.PUT("/trade-settings", tradeHandler.UpdateTradeSettings)
	api.GET("/trades", tradeHandler.ListOrders)
	api.GET("/trades/runtime", tradeHandler.GetRuntime)
	api.POST("/trades/:id/close", tradeHandler.CloseOrder)
	return r
}

func seedTradeOrder(t *testing.T, db *gorm.DB, order models.TradeOrder) models.TradeOrder {
	t.Helper()
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("seed trade order failed: %v", err)
	}
	return order
}

type tradeSettingsPayload struct {
	TradeEnabledEnv    bool     `json:"trade_enabled_env"`
	AllowedSymbolsEnv  []string `json:"allowed_symbols_env"`
	AutoExecuteEnabled bool     `json:"auto_execute_enabled"`
	AllowedSymbols     []string `json:"allowed_symbols"`
}

type tradeRuntimePayload struct {
	PendingFillCount int `json:"pending_fill_count"`
	OpenCount        int `json:"open_count"`
}

type routerTradeClientStub struct{}

func (routerTradeClientStub) GetFuturesBalance() (float64, error) { return 0, nil }
func (routerTradeClientStub) GetFuturesLeverage(symbol string) (int, error) { return 0, nil }
func (routerTradeClientStub) GetFuturesSymbolRules(symbol string) (service.FuturesSymbolRules, error) {
	return service.FuturesSymbolRules{}, nil
}
func (routerTradeClientStub) PlaceFuturesLimitOrder(symbol, side string, qty, price float64) (service.FuturesOrder, error) {
	return service.FuturesOrder{}, nil
}
func (routerTradeClientStub) GetFuturesOrder(symbol, orderID string) (service.FuturesOrder, error) {
	return service.FuturesOrder{}, nil
}
func (routerTradeClientStub) CancelFuturesOrder(symbol string, orderID string) error { return nil }
func (routerTradeClientStub) PlaceFuturesProtectionOrder(symbol, side, orderType string, stopPrice float64) (string, error) {
	return "", nil
}
func (routerTradeClientStub) CloseFuturesPosition(symbol, side string, qty float64) (string, error) {
	return "close-order", nil
}
func (routerTradeClientStub) GetFuturesPositions() ([]service.FuturesPosition, error) { return nil, nil }
