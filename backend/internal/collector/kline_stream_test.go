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

func TestKlineClosedEventNilEventNotPanics(t *testing.T) {
	ch := make(chan KlineClosedEvent, 1)
	c := &BinanceStreamCollector{klineEvents: ch}
	c.handleKlineEvent(nil) // 不应 panic
	if len(ch) != 0 {
		t.Error("nil event should not publish")
	}
}

func TestKlineClosedEventDroppedWhenChannelFull(t *testing.T) {
	ch := make(chan KlineClosedEvent, 1)
	c := &BinanceStreamCollector{klineEvents: ch}
	event := &binancesdk.WsKlineEvent{
		Symbol: "BTCUSDT",
		Kline:  binancesdk.WsKline{Interval: "1m", IsFinal: true},
	}
	// 先填满 channel
	c.handleKlineEvent(event)
	// 再次发送，channel 满，应走 default 分支不阻塞
	c.handleKlineEvent(event)
	// 消费一个，验证第一次正常写入
	got := <-ch
	if got.Symbol != "BTCUSDT" {
		t.Errorf("unexpected: %+v", got)
	}
}
