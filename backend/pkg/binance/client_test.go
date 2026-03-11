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

func TestGetFuturesMarketDataReturnsMockDataWhenFallbackEnabled(t *testing.T) {
	client := NewClient("", "", 20*time.Millisecond)
	client.SetHTTPClient(NewFailingHTTPClient(errors.New("forced sdk failure")))
	client.SetMockFallbackEnabled(true)

	result, err := client.GetFuturesMarketData("SOLUSDT")
	if err != nil {
		t.Fatalf("expected futures mock fallback to succeed, got err=%v", err)
	}
	if result.Symbol != "SOLUSDT" {
		t.Fatalf("unexpected symbol: got=%s", result.Symbol)
	}
	if result.MarkPrice <= 0 || result.IndexPrice <= 0 || result.OpenInterest <= 0 {
		t.Fatalf("expected futures metrics to be positive, got=%+v", result)
	}
	if result.Source != "mock" {
		t.Fatalf("expected mock source, got=%s", result.Source)
	}
}

func TestGetFuturesMarketDataReturnsErrorWhenFallbackDisabled(t *testing.T) {
	client := NewClient("", "", 20*time.Millisecond)
	client.SetHTTPClient(NewFailingHTTPClient(errors.New("forced sdk failure")))
	client.SetMockFallbackEnabled(false)

	result, err := client.GetFuturesMarketData("BTCUSDT")
	if err == nil {
		t.Fatalf("expected sdk failure when futures mock fallback is disabled, got result=%+v", result)
	}
}
