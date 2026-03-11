package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFeishuNotifierSendsSignedTextMessage(t *testing.T) {
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body failed: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"StatusCode":0}`))
	}))
	defer server.Close()

	notifier := NewFeishuNotifier(server.URL, "secret-value", 2*time.Second)
	notifier.now = func() time.Time {
		return time.Unix(1710000000, 0)
	}

	result := notifier.Notify(context.Background(), AlertEvent{
		Title:             "BTCUSDT A 级 setup 已就绪",
		Verdict:           "强偏多",
		TradeabilityLabel: "A 级可跟踪",
		Confidence:        74,
		RiskLabel:         "可控风险",
		Summary:           "多周期方向已经对齐。",
		TimeframeLabels:   []string{"4h 强偏多", "1h 强偏多", "15m 强偏多"},
		Reasons:           []string{"趋势因子主导当前方向。"},
		EntryPrice:        65200,
		StopLoss:          64880,
		TargetPrice:       65880,
		RiskReward:        2.1,
	})

	if result.Status != "sent" {
		t.Fatalf("expected sent status, got=%s detail=%s", result.Status, result.Detail)
	}
	if payload["msg_type"] != "text" {
		t.Fatalf("unexpected msg_type: %#v", payload["msg_type"])
	}
	if payload["timestamp"] != "1710000000" {
		t.Fatalf("unexpected timestamp: %#v", payload["timestamp"])
	}
	if payload["sign"] == "" {
		t.Fatal("expected sign to be present")
	}

	content, ok := payload["content"].(map[string]any)
	if !ok {
		t.Fatalf("expected content to be object, got=%T", payload["content"])
	}
	text, ok := content["text"].(string)
	if !ok {
		t.Fatalf("expected text content to be string, got=%T", content["text"])
	}
	if text == "" || text[:13] != "[Alpha Pulse]" {
		t.Fatalf("unexpected text content: %s", text)
	}
}

func TestFeishuNotifierSkipsWhenWebhookMissing(t *testing.T) {
	notifier := NewFeishuNotifier("", "", time.Second)

	result := notifier.Notify(context.Background(), AlertEvent{Title: "noop"})
	if result.Status != "skipped" {
		t.Fatalf("expected skipped status, got=%s", result.Status)
	}
}
