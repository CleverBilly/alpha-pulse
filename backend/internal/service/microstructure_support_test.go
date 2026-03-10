package service

import (
	"testing"

	"alpha-pulse/backend/models"
)

func TestMergeMicrostructureEventsDerivesCompositePatterns(t *testing.T) {
	base := []models.OrderFlowMicrostructureEvent{
		{
			Type:      "failed_auction_high_reject",
			Bias:      "bearish",
			Score:     -6,
			Strength:  0.74,
			Price:     64810,
			TradeTime: 1741300200000,
			Detail:    "上方失败拍卖形成强回落分型",
		},
		{
			Type:      "absorption",
			Bias:      "bearish",
			Score:     -5,
			Strength:  0.69,
			Price:     64785,
			TradeTime: 1741300260000,
			Detail:    "买盘被持续吸收",
		},
		{
			Type:      "iceberg",
			Bias:      "bullish",
			Score:     4,
			Strength:  0.64,
			Price:     64318,
			TradeTime: 1741300275000,
			Detail:    "同价带重复出现隐藏买单承接",
		},
		{
			Type:      "absorption",
			Bias:      "bullish",
			Score:     5,
			Strength:  0.68,
			Price:     64308,
			TradeTime: 1741300285000,
			Detail:    "卖压被持续吸收",
		},
		{
			Type:      "failed_auction_low_reclaim",
			Bias:      "bullish",
			Score:     6,
			Strength:  0.72,
			Price:     64288,
			TradeTime: 1741300300000,
			Detail:    "下方失败拍卖形成强收回分型",
		},
	}
	extra := []models.OrderFlowMicrostructureEvent{
		{
			Type:      "order_book_migration_layered",
			Bias:      "bullish",
			Score:     5,
			Strength:  0.66,
			Price:     64320,
			TradeTime: 1741300320000,
			Detail:    "买方挂单墙连续多层上移",
		},
		{
			Type:      "initiative_shift",
			Bias:      "bullish",
			Score:     4,
			Strength:  0.57,
			Price:     64355,
			TradeTime: 1741300380000,
			Detail:    "买方主动性明显增强",
		},
		{
			Type:      "aggression_burst",
			Bias:      "bullish",
			Score:     3,
			Strength:  0.48,
			Price:     64392,
			TradeTime: 1741300440000,
			Detail:    "主动买盘冲击放大",
		},
	}

	merged := mergeMicrostructureEvents(base, extra)

	if !containsMergedEvent(merged, "auction_trap_reversal", "bullish") &&
		!containsMergedEvent(merged, "auction_trap_reversal", "bearish") {
		t.Fatalf("expected auction trap reversal, got %#v", merged)
	}
	if !containsMergedEvent(merged, "liquidity_ladder_breakout", "bullish") {
		t.Fatalf("expected bullish liquidity ladder breakout, got %#v", merged)
	}
	if !containsMergedEvent(merged, "iceberg_reload", "bullish") {
		t.Fatalf("expected bullish iceberg reload, got %#v", merged)
	}
	if !containsMergedEvent(merged, "absorption_reload_continuation", "bullish") {
		t.Fatalf("expected bullish absorption reload continuation, got %#v", merged)
	}
	if !containsMergedEvent(merged, "migration_auction_flip", "bullish") {
		t.Fatalf("expected bullish migration auction flip, got %#v", merged)
	}
}

func containsMergedEvent(
	events []models.OrderFlowMicrostructureEvent,
	eventType, bias string,
) bool {
	for _, event := range events {
		if event.Type == eventType && event.Bias == bias {
			return true
		}
	}
	return false
}
