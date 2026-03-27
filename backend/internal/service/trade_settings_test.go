package service

import (
	"reflect"
	"testing"
)

func TestSanitizeTradeSettingsFiltersSymbolsAndBoundsValues(t *testing.T) {
	settings, err := sanitizeTradeSettings(TradeSettings{
		AutoExecuteEnabled:  true,
		AllowedSymbols:      []string{"ethusdt", "BTCUSDT", "DOGEUSDT"},
		RiskPct:             2.5,
		MinRiskReward:       1.7,
		EntryTimeoutSeconds: 45,
		MaxOpenPositions:    2,
		SyncEnabled:         true,
	}, []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"})
	if err != nil {
		t.Fatalf("sanitize trade settings: %v", err)
	}

	expectedSymbols := []string{"ETHUSDT", "BTCUSDT"}
	if !reflect.DeepEqual(settings.AllowedSymbols, expectedSymbols) {
		t.Fatalf("unexpected allowed symbols: got=%#v want=%#v", settings.AllowedSymbols, expectedSymbols)
	}
	if settings.RiskPct != 2.5 {
		t.Fatalf("unexpected risk_pct: %f", settings.RiskPct)
	}
	if settings.EntryTimeoutSeconds != 45 {
		t.Fatalf("unexpected entry timeout: %d", settings.EntryTimeoutSeconds)
	}
}

func TestSanitizeTradeSettingsRejectsInvalidBounds(t *testing.T) {
	_, err := sanitizeTradeSettings(TradeSettings{
		AllowedSymbols:      []string{"BTCUSDT"},
		RiskPct:             0,
		MinRiskReward:       0.9,
		EntryTimeoutSeconds: 0,
		MaxOpenPositions:    0,
	}, []string{"BTCUSDT", "ETHUSDT"})
	if err == nil {
		t.Fatal("expected invalid trade settings to fail validation")
	}
}
