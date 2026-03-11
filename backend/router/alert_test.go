package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"alpha-pulse/backend/internal/service"
	"github.com/gin-gonic/gin"
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
	if payload.Data.Generated <= 0 {
		t.Fatalf("expected refresh to generate alerts, got=%d", payload.Data.Generated)
	}
	if len(payload.Data.Items) == 0 {
		t.Fatal("expected recent alerts to be returned")
	}
	if payload.Data.Items[0].Symbol == "" {
		t.Fatalf("unexpected alert item: %+v", payload.Data.Items[0])
	}
}

func TestAlertHistoryEndpointReturnsPersistedAlerts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	r := newTestRouter(t, db)

	refreshReq := httptest.NewRequest(http.MethodPost, "/api/alerts/refresh?limit=10", nil)
	refreshRec := httptest.NewRecorder()
	r.ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("unexpected refresh status: got=%d body=%s", refreshRec.Code, refreshRec.Body.String())
	}

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
