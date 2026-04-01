package collector

import (
	"testing"
	"time"

	binancesdk "github.com/adshao/go-binance/v2"
)

func TestKlineClosedEventPublishedOnFinalKline(t *testing.T) {
	ch := make(chan KlineClosedEvent, 10)
	c := &BinanceStreamCollector{klineEvents: ch}

	event := &binancesdk.WsKlineEvent{
		Symbol: "BTCUSDT",
		Kline: binancesdk.WsKline{
			Interval: "1m",
			IsFinal:  true,
		},
	}
	c.handleKlineEvent(event)

	select {
	case got := <-ch:
		if got.Symbol != "BTCUSDT" || got.Interval != "1m" {
			t.Errorf("unexpected event: %+v", got)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("expected KlineClosedEvent not received")
	}
}

func TestKlineClosedEventNotPublishedOnNonFinalKline(t *testing.T) {
	ch := make(chan KlineClosedEvent, 10)
	c := &BinanceStreamCollector{klineEvents: ch}

	event := &binancesdk.WsKlineEvent{
		Symbol: "BTCUSDT",
		Kline: binancesdk.WsKline{
			Interval: "1m",
			IsFinal:  false,
		},
	}
	c.handleKlineEvent(event)

	select {
	case got := <-ch:
		t.Errorf("unexpected event published: %+v", got)
	case <-time.After(30 * time.Millisecond):
		// correct — no event
	}
}
