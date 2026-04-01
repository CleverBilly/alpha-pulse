package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	signalpkg "alpha-pulse/backend/internal/signal"
	"alpha-pulse/backend/models"
	"github.com/gin-gonic/gin"
)

// stubSignalConfigRepo 实现 SignalConfigRepo 接口，用于测试注入。
type stubSignalConfigRepo struct {
	upserted []models.SignalConfig
}

func (r *stubSignalConfigRepo) Upsert(cfg models.SignalConfig) error {
	r.upserted = append(r.upserted, cfg)
	return nil
}

func (r *stubSignalConfigRepo) GetAll() ([]models.SignalConfig, error) {
	return r.upserted, nil
}

func newStubDBConfigProvider() *signalpkg.DBConfigProvider {
	return signalpkg.NewDBConfigProvider(nil)
}

func newStubSignalConfigRepo() *stubSignalConfigRepo {
	return &stubSignalConfigRepo{}
}

func TestPatchSignalConfigUpdatesProviderAndDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := newStubDBConfigProvider()
	repo := newStubSignalConfigRepo()
	h := NewSignalConfigHandler(provider, repo)

	r := gin.New()
	r.PATCH("/api/signal-configs", h.Patch)

	body, _ := json.Marshal(map[string]string{
		"symbol":   "BTCUSDT",
		"interval": "1m",
		"key":      "buy_threshold",
		"value":    "45",
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/signal-configs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	// 验证内存 provider 已被热更新
	got := provider.GetInt("BTCUSDT", "1m", "buy_threshold", 35)
	if got != 45 {
		t.Errorf("expected provider buy_threshold=45, got %d", got)
	}
	// 验证 DB stub 已持久化
	if len(repo.upserted) != 1 {
		t.Errorf("expected 1 upserted record, got %d", len(repo.upserted))
	}
}

func TestPatchSignalConfigMissingField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := newStubDBConfigProvider()
	repo := newStubSignalConfigRepo()
	h := NewSignalConfigHandler(provider, repo)

	r := gin.New()
	r.PATCH("/api/signal-configs", h.Patch)

	// 缺少 value 字段
	body, _ := json.Marshal(map[string]string{
		"symbol":   "BTCUSDT",
		"interval": "1m",
		"key":      "buy_threshold",
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/signal-configs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListSignalConfigsReturnsAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := newStubDBConfigProvider()
	repo := newStubSignalConfigRepo()
	repo.upserted = []models.SignalConfig{
		{Symbol: "BTCUSDT", Interval: "1m", Key: "buy_threshold", Value: "40"},
	}
	h := NewSignalConfigHandler(provider, repo)

	r := gin.New()
	r.GET("/api/signal-configs", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/signal-configs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
