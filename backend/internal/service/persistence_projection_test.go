package service

import (
	"encoding/json"
	"testing"

	"alpha-pulse/backend/models"
)

func TestProjectLargeTradeEventsPreservesReplayContext(t *testing.T) {
	orderFlow := models.OrderFlow{
		ID:           18,
		Symbol:       "BTCUSDT",
		IntervalType: "5m",
		OpenTime:     1741300200000,
		LargeTrades: []models.OrderFlowLargeTrade{
			{
				AggTradeID: 91001,
				Side:       "buy",
				Price:      64210.25,
				Quantity:   2.4,
				Notional:   154104.6,
				TradeTime:  1741300187000,
			},
		},
	}

	events := projectLargeTradeEvents(orderFlow)
	if len(events) != 1 {
		t.Fatalf("expected 1 large trade event, got %d", len(events))
	}

	event := events[0]
	if event.AggTradeID != 91001 {
		t.Fatalf("expected agg trade id to be preserved, got %d", event.AggTradeID)
	}
	if event.OrderFlowID != orderFlow.ID || event.IntervalType != orderFlow.IntervalType || event.OpenTime != orderFlow.OpenTime {
		t.Fatalf("expected orderflow replay context to be preserved, got %+v", event)
	}
}

func TestProjectFeatureSnapshotSerializesAuditPayload(t *testing.T) {
	snapshot := MarketSnapshot{
		Price: MarketPrice{
			Symbol: "ETHUSDT",
			Price:  3245.4,
			Time:   1741300200000,
		},
		Klines: []models.Kline{
			{
				Symbol:       "ETHUSDT",
				IntervalType: "15m",
				OpenTime:     1741300200000,
				ClosePrice:   3245.4,
			},
		},
		OrderFlow: models.OrderFlow{
			Symbol:       "ETHUSDT",
			IntervalType: "15m",
			OpenTime:     1741300200000,
		},
		Signal: models.Signal{
			Symbol:       "ETHUSDT",
			IntervalType: "15m",
			OpenTime:     1741300200000,
			Action:       "BUY",
			Score:        61,
			Confidence:   74,
		},
	}

	record, err := projectFeatureSnapshot(snapshot)
	if err != nil {
		t.Fatalf("project feature snapshot failed: %v", err)
	}
	if record.Symbol != "ETHUSDT" || record.IntervalType != "15m" || record.OpenTime != 1741300200000 {
		t.Fatalf("unexpected feature snapshot header: %+v", record)
	}
	if record.SignalAction != "BUY" || record.SignalScore != 61 || record.SignalConfidence != 74 {
		t.Fatalf("unexpected feature snapshot signal summary: %+v", record)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(record.SnapshotJSON), &payload); err != nil {
		t.Fatalf("unmarshal snapshot json failed: %v", err)
	}
	if payload["price"] == nil || payload["signal"] == nil || payload["orderflow"] == nil {
		t.Fatalf("snapshot json should preserve aggregate payload, got=%v", payload)
	}
}
