package binance

import (
	"errors"
	"testing"
	"time"
)

func TestGetKlinesReturnsMockDataWhenFallbackEnabled(t *testing.T) {
	client := NewClient("", "", 20*time.Millisecond)
	client.SetHTTPClient(NewFailingHTTPClient(errors.New("forced sdk failure")))
	client.SetMockFallbackEnabled(true)

	klines, err := client.GetKlines("BTCUSDT", "1m", 4)
	if err != nil {
		t.Fatalf("expected mock fallback to succeed, got err=%v", err)
	}
	if len(klines) != 4 {
		t.Fatalf("expected 4 mock klines, got %d", len(klines))
	}
}

func TestGetKlinesReturnsErrorWhenFallbackDisabled(t *testing.T) {
	client := NewClient("", "", 20*time.Millisecond)
	client.SetHTTPClient(NewFailingHTTPClient(errors.New("forced sdk failure")))
	client.SetMockFallbackEnabled(false)

	klines, err := client.GetKlines("BTCUSDT", "1m", 4)
	if err == nil {
		t.Fatalf("expected sdk failure when mock fallback is disabled, got klines=%#v", klines)
	}
}
