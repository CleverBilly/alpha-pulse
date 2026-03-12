package service

import "testing"

func TestProjectAlertRecordPreservesReplayPayload(t *testing.T) {
	event := AlertEvent{
		ID:                "BTCUSDT-setup_ready-1710000000000",
		Symbol:            "BTCUSDT",
		Kind:              "setup_ready",
		Severity:          "A",
		Title:             "BTCUSDT A 级机会已就绪",
		Verdict:           "强偏多",
		TradeabilityLabel: "A 级可跟踪",
		Summary:           "4h 与 1h 已经对齐，15m 触发也站在同一边。",
		Reasons:           []string{"趋势因子主导当前方向。"},
		TimeframeLabels:   []string{"4h 强偏多", "1h 强偏多", "15m 强偏多"},
		Confidence:        74,
		RiskLabel:         "可控风险",
		EntryPrice:        65200,
		StopLoss:          64880,
		TargetPrice:       65880,
		RiskReward:        2.1,
		CreatedAt:         1710000000000,
		Deliveries: []AlertDelivery{
			{
				Channel: "feishu",
				Status:  "sent",
				SentAt:  1710000001000,
			},
		},
	}

	record, err := projectAlertRecord(event, alertState{
		State:      DirectionStateStrongBullish,
		Tradable:   true,
		SetupReady: true,
	})
	if err != nil {
		t.Fatalf("project alert record failed: %v", err)
	}
	if record.AlertID != event.ID || record.Symbol != "BTCUSDT" || record.DirectionState != string(DirectionStateStrongBullish) {
		t.Fatalf("unexpected alert record header: %+v", record)
	}

	restored := hydrateAlertEvent(record)
	if restored.ID != event.ID || restored.Symbol != event.Symbol || len(restored.Deliveries) != 1 {
		t.Fatalf("restored alert event should preserve payload: %+v", restored)
	}
}
