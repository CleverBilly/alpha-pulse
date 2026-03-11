package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"alpha-pulse/backend/internal/service"
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestAlertRefreshEndpointReturnsRecentAlerts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	r := newTestRouter(t, db)

	req := httptest.NewRequest(http.MethodPost, "/api/alerts/refresh?limit=10", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload apiEnvelope[service.AlertFeed]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode alert feed failed: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("unexpected business code: got=%d message=%s", payload.Code, payload.Message)
	}
	if payload.Data.Generated < 0 {
		t.Fatalf("generated count should not be negative, got=%d", payload.Data.Generated)
	}
	if payload.Data.Generated > 0 && len(payload.Data.Items) == 0 {
		t.Fatal("generated alerts should be returned in recent feed")
	}
	if len(payload.Data.Items) > 0 && payload.Data.Items[0].Symbol == "" {
		t.Fatalf("unexpected alert item: %+v", payload.Data.Items[0])
	}
}

func TestAlertHistoryEndpointReturnsPersistedAlerts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	r := newTestRouter(t, db)
	seedAlertRecord(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/alerts/history?limit=10", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload apiEnvelope[service.AlertFeed]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode alert history failed: %v", err)
	}
	if len(payload.Data.Items) == 0 {
		t.Fatal("expected persisted alert history to be returned")
	}
	if payload.Data.Items[0].ID == "" || payload.Data.Items[0].CreatedAt == 0 {
		t.Fatalf("unexpected persisted alert item: %+v", payload.Data.Items[0])
	}
}

func TestAlertPreferencesEndpointSupportsReadWrite(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	r := newTestRouter(t, db)

	putBody := []byte(`{"feishu_enabled":false,"browser_enabled":true,"setup_ready_enabled":true,"direction_shift_enabled":false,"no_trade_enabled":true,"minimum_confidence":63,"quiet_hours_enabled":true,"quiet_hours_start":1,"quiet_hours_end":8,"symbols":["BTCUSDT","SOLUSDT"]}`)
	putReq := httptest.NewRequest(http.MethodPut, "/api/alerts/preferences", bytes.NewReader(putBody))
	putReq.Header.Set("Content-Type", "application/json")
	putRec := httptest.NewRecorder()
	r.ServeHTTP(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("unexpected put status: got=%d body=%s", putRec.Code, putRec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/alerts/preferences", nil)
	getRec := httptest.NewRecorder()
	r.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("unexpected get status: got=%d body=%s", getRec.Code, getRec.Body.String())
	}

	var payload apiEnvelope[service.AlertPreferences]
	if err := json.NewDecoder(getRec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode alert preferences failed: %v", err)
	}
	if payload.Data.FeishuEnabled || payload.Data.DirectionShiftEnabled || payload.Data.MinimumConfidence != 63 {
		t.Fatalf("unexpected alert preferences payload: %+v", payload.Data)
	}
	if len(payload.Data.Symbols) != 2 || payload.Data.Symbols[1] != "SOLUSDT" {
		t.Fatalf("unexpected watched symbols: %+v", payload.Data.Symbols)
	}
}

func seedAlertRecord(t *testing.T, db *gorm.DB) {
	t.Helper()

	event := service.AlertEvent{
		ID:                "BTCUSDT-setup_ready-1710000000000",
		Symbol:            "BTCUSDT",
		Kind:              "setup_ready",
		Severity:          "A",
		Title:             "BTCUSDT A 级 setup 已就绪",
		Verdict:           "强偏多",
		TradeabilityLabel: "A 级可跟踪",
		Summary:           "4h 与 1h 对齐，15m 与 5m 也站在同一边。",
		Reasons:           []string{"趋势因子主导当前方向。", "Futures 因子没有明显逆风。"},
		TimeframeLabels:   []string{"4h 强偏多", "1h 强偏多", "15m 强偏多", "5m 偏多"},
		Confidence:        74,
		RiskLabel:         "可控风险",
		EntryPrice:        65200,
		StopLoss:          64880,
		TargetPrice:       65880,
		RiskReward:        2.1,
		CreatedAt:         time.UnixMilli(1710000000000).UnixMilli(),
		Deliveries:        []service.AlertDelivery{},
	}
	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal alert payload failed: %v", err)
	}

	repo := repository.NewAlertRecordRepository(db)
	record := models.AlertRecord{
		AlertID:           event.ID,
		Symbol:            event.Symbol,
		Kind:              event.Kind,
		Severity:          event.Severity,
		DirectionState:    "strong-bullish",
		Tradable:          true,
		SetupReady:        true,
		TradeabilityLabel: event.TradeabilityLabel,
		Title:             event.Title,
		Verdict:           event.Verdict,
		Summary:           event.Summary,
		Confidence:        event.Confidence,
		RiskLabel:         event.RiskLabel,
		EntryPrice:        event.EntryPrice,
		StopLoss:          event.StopLoss,
		TargetPrice:       event.TargetPrice,
		RiskReward:        event.RiskReward,
		EventTime:         event.CreatedAt,
		PayloadJSON:       string(payload),
	}
	if err := repo.Create(&record); err != nil {
		t.Fatalf("seed alert record failed: %v", err)
	}
}
