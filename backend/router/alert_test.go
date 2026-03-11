package router_test

import (
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
