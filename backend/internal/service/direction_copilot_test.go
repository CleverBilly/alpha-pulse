package service

import "testing"

func TestBuildDirectionCopilotDecisionReturnsNoTradeWhen5mConflicts(t *testing.T) {
	macro := buildDirectionTestSnapshot("BTCUSDT", "4h", 62, 78, "BUY", "uptrend")
	bias := buildDirectionTestSnapshot("BTCUSDT", "1h", 58, 74, "BUY", "uptrend")
	trigger := buildDirectionTestSnapshot("BTCUSDT", "15m", 56, 70, "BUY", "uptrend")
	execution := buildDirectionTestSnapshot("BTCUSDT", "5m", -48, 72, "SELL", "downtrend")

	decision := BuildDirectionCopilotDecision(&macro, &bias, &trigger, &execution)
	if decision.Tradable {
		t.Fatalf("expected 5m conflict to downgrade tradability, got=%+v", decision)
	}
	if decision.TradeabilityLabel != "No-Trade" {
		t.Fatalf("expected no-trade label, got=%s", decision.TradeabilityLabel)
	}
	if len(decision.TimeframeLabels) != 4 || decision.TimeframeLabels[3] != "5m 偏空" {
		t.Fatalf("expected 5m timeframe label, got=%v", decision.TimeframeLabels)
	}
}

func TestDeriveLiquidationProxyBuildsLongSqueezeZone(t *testing.T) {
	snapshot := deriveLiquidationProxy(FuturesSnapshot{
		Available:         true,
		Symbol:            "BTCUSDT",
		MarkPrice:         65000,
		BasisBps:          8.6,
		FundingRate:       0.00031,
		OpenInterestValue: 18_000_000_000,
		LongShortRatio:    1.19,
	})

	if snapshot.LiquidationPressure != "long-squeeze" {
		t.Fatalf("expected long-squeeze pressure, got=%s", snapshot.LiquidationPressure)
	}
	if snapshot.LongLiquidationZoneLow <= 0 || snapshot.LongLiquidationZoneHigh <= 0 {
		t.Fatalf("expected populated long liquidation zone, got=%+v", snapshot)
	}
	if snapshot.LongLiquidationZoneHigh >= snapshot.MarkPrice {
		t.Fatalf("expected long liquidation zone below mark price, got=%+v", snapshot)
	}
	if snapshot.LiquidationSummary == "" {
		t.Fatal("expected liquidation summary to be populated")
	}
}
