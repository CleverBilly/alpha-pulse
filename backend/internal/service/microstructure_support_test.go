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

	if !containsMergedEvent(merged, "auction_trap_reversal", "bearish") {
		t.Fatalf("expected bearish auction trap reversal, got %#v", merged)
	}
	if !containsMergedEvent(merged, "liquidity_ladder_breakout", "bullish") {
		t.Fatalf("expected bullish liquidity ladder breakout, got %#v", merged)
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
