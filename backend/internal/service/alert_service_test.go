package service

import (
	"context"
	"testing"
	"time"

	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
)

func TestAlertServiceGeneratesSetupReadyAlertOnce(t *testing.T) {
	fetcher := &stubDirectionFetcher{
		snapshots: map[string]MarketSnapshot{
			"BTCUSDT:4h":  buildDirectionTestSnapshot("BTCUSDT", "4h", 62, 78, "BUY", "uptrend"),
			"BTCUSDT:1h":  buildDirectionTestSnapshot("BTCUSDT", "1h", 58, 74, "BUY", "uptrend"),
			"BTCUSDT:15m": buildDirectionTestSnapshot("BTCUSDT", "15m", 56, 70, "BUY", "uptrend"),
			"BTCUSDT:5m":  buildDirectionTestSnapshot("BTCUSDT", "5m", 54, 68, "BUY", "uptrend"),
		},
	}
	service := NewAlertService(fetcher, nil, nil, []string{"BTCUSDT"}, 10)
	macro := fetcher.snapshots["BTCUSDT:4h"]
	bias := fetcher.snapshots["BTCUSDT:1h"]
	trigger := fetcher.snapshots["BTCUSDT:15m"]
	execution := fetcher.snapshots["BTCUSDT:5m"]
	decision := BuildDirectionCopilotDecision(&macro, &bias, &trigger, &execution)
	if !decision.Tradable {
		t.Fatalf("expected seeded decision to be tradable, got=%+v", decision)
	}

	events, err := service.EvaluateAll(context.Background(), false)
	if err != nil {
		t.Fatalf("evaluate all failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 alert event, got=%d", len(events))
	}
	if events[0].Kind != "setup_ready" {
		t.Fatalf("expected setup_ready alert, got=%s", events[0].Kind)
	}
	if events[0].TradeabilityLabel != "A 级可跟踪" {
		t.Fatalf("unexpected tradeability label: %s", events[0].TradeabilityLabel)
	}

	repeated, err := service.EvaluateAll(context.Background(), false)
	if err != nil {
		t.Fatalf("repeat evaluate failed: %v", err)
	}
	if len(repeated) != 0 {
		t.Fatalf("expected duplicate state to generate no alert, got=%d", len(repeated))
	}
}

func TestAlertServiceInvokesAutoTradeCoordinatorForSetupReady(t *testing.T) {
	fetcher := &stubDirectionFetcher{
		snapshots: map[string]MarketSnapshot{
			"BTCUSDT:4h":  buildDirectionTestSnapshot("BTCUSDT", "4h", 62, 78, "BUY", "uptrend"),
			"BTCUSDT:1h":  buildDirectionTestSnapshot("BTCUSDT", "1h", 58, 74, "BUY", "uptrend"),
			"BTCUSDT:15m": buildDirectionTestSnapshot("BTCUSDT", "15m", 56, 70, "BUY", "uptrend"),
			"BTCUSDT:5m":  buildDirectionTestSnapshot("BTCUSDT", "5m", 54, 68, "BUY", "uptrend"),
		},
	}
	autoTrader := &stubAutoTradeCoordinator{}
	service := NewAlertService(fetcher, nil, nil, []string{"BTCUSDT"}, 10)
	service.SetAutoTradeCoordinator(autoTrader)

	events, err := service.EvaluateAll(context.Background(), false)
	if err != nil {
		t.Fatalf("evaluate all failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 alert event, got=%d", len(events))
	}
	if autoTrader.calls != 1 {
		t.Fatalf("expected auto trader to be called once, got %d", autoTrader.calls)
	}
	if autoTrader.lastEvent.Kind != "setup_ready" {
		t.Fatalf("unexpected auto trader event: %+v", autoTrader.lastEvent)
	}
}

func TestAlertServiceGeneratesNoTradeAfterTradableState(t *testing.T) {
	fetcher := &stubDirectionFetcher{
		snapshots: map[string]MarketSnapshot{
			"BTCUSDT:4h":  buildDirectionTestSnapshot("BTCUSDT", "4h", 62, 78, "BUY", "uptrend"),
			"BTCUSDT:1h":  buildDirectionTestSnapshot("BTCUSDT", "1h", 58, 74, "BUY", "uptrend"),
			"BTCUSDT:15m": buildDirectionTestSnapshot("BTCUSDT", "15m", 56, 70, "BUY", "uptrend"),
			"BTCUSDT:5m":  buildDirectionTestSnapshot("BTCUSDT", "5m", 54, 68, "BUY", "uptrend"),
		},
	}
	service := NewAlertService(fetcher, nil, nil, []string{"BTCUSDT"}, 10)

	if _, err := service.EvaluateAll(context.Background(), false); err != nil {
		t.Fatalf("seed evaluate failed: %v", err)
	}

	fetcher.snapshots["BTCUSDT:4h"] = buildDirectionTestSnapshot("BTCUSDT", "4h", -66, 72, "SELL", "downtrend")
	events, err := service.EvaluateAll(context.Background(), false)
	if err != nil {
		t.Fatalf("conflict evaluate failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 no-trade event, got=%d", len(events))
	}
	if events[0].Kind != "no_trade" {
		t.Fatalf("expected no_trade event, got=%s", events[0].Kind)
	}
	if events[0].Verdict != "当前禁止交易" {
		t.Fatalf("unexpected verdict: %s", events[0].Verdict)
	}
}

type stubDirectionFetcher struct {
	snapshots map[string]MarketSnapshot
}

type stubAutoTradeCoordinator struct {
	calls     int
	lastEvent AlertEvent
}

func (s *stubAutoTradeCoordinator) HandleEvent(ctx context.Context, event AlertEvent) error {
	s.calls++
	s.lastEvent = event
	return nil
}

func (s *stubDirectionFetcher) GetMarketSnapshotWithRefresh(symbol, interval string, limit int, refresh bool) (MarketSnapshot, error) {
	key := symbol + ":" + interval
	snapshot, ok := s.snapshots[key]
	if !ok {
		return MarketSnapshot{}, context.DeadlineExceeded
	}
	return snapshot, nil
}

func buildDirectionTestSnapshot(symbol string, interval string, score int, confidence int, action string, trend string) MarketSnapshot {
	signal := models.Signal{
		Symbol:       symbol,
		IntervalType: interval,
		Action:       action,
		Score:        score,
		Confidence:   confidence,
		EntryPrice:   65200,
		StopLoss:     64880,
		TargetPrice:  65880,
		RiskReward:   2.1,
		Explain:      "多周期方向已经对齐。",
		Factors: []models.SignalFactor{
			{
				Key:    "trend",
				Name:   "Trend",
				Score:  score / 3,
				Bias:   action,
				Reason: "趋势因子主导当前方向。",
			},
			{
				Key:    "flow",
				Name:   "Order Flow",
				Score:  score / 4,
				Bias:   action,
				Reason: "订单流与结构保持一致。",
			},
		},
	}

	return MarketSnapshot{
		Price: MarketPrice{
			Symbol: symbol,
			Price:  65320,
			Time:   1710000000000,
		},
		Futures: FuturesSnapshot{
			Available:      true,
			Symbol:         symbol,
			BasisBps:       ternaryFloat(score >= 0, 4.2, -4.2),
			FundingRate:    ternaryFloat(score >= 0, 0.00008, -0.00008),
			LongShortRatio: ternaryFloat(score >= 0, 1.05, 0.95),
		},
		Signal: signal,
		Structure: models.Structure{
			Symbol: symbol,
			Trend:  trend,
		},
		Liquidity: models.Liquidity{
			Symbol:             symbol,
			SweepType:          "sell_sweep",
			OrderBookImbalance: ternaryFloat(score >= 0, 0.12, -0.12),
		},
		OrderFlow: models.OrderFlow{
			Symbol:   symbol,
			Delta:    ternaryFloat(score >= 0, 180, -180),
			OpenTime: 1710000000000,
		},
	}
}

func ternaryFloat(condition bool, left float64, right float64) float64 {
	if condition {
		return left
	}
	return right
}

func TestAlertServicePersistsAndReloadsHistoryState(t *testing.T) {
	db := newServiceTestDB(t)
	repo := repository.NewAlertRecordRepository(db)
	fetcher := &stubDirectionFetcher{
		snapshots: map[string]MarketSnapshot{
			"BTCUSDT:4h":  buildDirectionTestSnapshot("BTCUSDT", "4h", 62, 78, "BUY", "uptrend"),
			"BTCUSDT:1h":  buildDirectionTestSnapshot("BTCUSDT", "1h", 58, 74, "BUY", "uptrend"),
			"BTCUSDT:15m": buildDirectionTestSnapshot("BTCUSDT", "15m", 56, 70, "BUY", "uptrend"),
			"BTCUSDT:5m":  buildDirectionTestSnapshot("BTCUSDT", "5m", 54, 68, "BUY", "uptrend"),
		},
	}

	first := NewAlertService(fetcher, repo, nil, []string{"BTCUSDT"}, 10)
	generated, err := first.EvaluateAll(context.Background(), false)
	if err != nil {
		t.Fatalf("seed evaluate failed: %v", err)
	}
	if len(generated) != 1 {
		t.Fatalf("expected 1 persisted alert, got=%d", len(generated))
	}

	assertServiceCount(t, db, &models.AlertRecord{}, "symbol = ?", 1, "BTCUSDT")

	second := NewAlertService(fetcher, repo, nil, []string{"BTCUSDT"}, 10)
	repeated, err := second.EvaluateAll(context.Background(), false)
	if err != nil {
		t.Fatalf("repeat evaluate failed: %v", err)
	}
	if len(repeated) != 0 {
		t.Fatalf("expected persisted state to suppress duplicate alert, got=%d", len(repeated))
	}

	history := second.ListRecent(5)
	if len(history) != 1 {
		t.Fatalf("expected persisted history to be readable, got=%d", len(history))
	}
	if history[0].Kind != "setup_ready" {
		t.Fatalf("unexpected persisted history item: %+v", history[0])
	}
}

func TestAlertServiceAppliesPreferencesAndQuietHours(t *testing.T) {
	db := newServiceTestDB(t)
	preferenceRepo := repository.NewAlertPreferenceRepository(db)
	fetcher := &stubDirectionFetcher{
		snapshots: map[string]MarketSnapshot{
			"BTCUSDT:4h":  buildDirectionTestSnapshot("BTCUSDT", "4h", 62, 78, "BUY", "uptrend"),
			"BTCUSDT:1h":  buildDirectionTestSnapshot("BTCUSDT", "1h", 58, 74, "BUY", "uptrend"),
			"BTCUSDT:15m": buildDirectionTestSnapshot("BTCUSDT", "15m", 56, 70, "BUY", "uptrend"),
			"BTCUSDT:5m":  buildDirectionTestSnapshot("BTCUSDT", "5m", 54, 68, "BUY", "uptrend"),
		},
	}
	notifier := &stubAlertNotifier{channel: "feishu"}
	service := NewAlertService(fetcher, nil, preferenceRepo, []string{"BTCUSDT"}, 10, notifier)
	service.now = func() time.Time {
		return time.Date(2026, time.March, 11, 2, 0, 0, 0, time.Local)
	}

	_, err := service.UpdatePreferences(AlertPreferences{
		FeishuEnabled:         true,
		BrowserEnabled:        true,
		SetupReadyEnabled:     true,
		DirectionShiftEnabled: false,
		NoTradeEnabled:        true,
		MinimumConfidence:     80,
		QuietHoursEnabled:     true,
		QuietHoursStart:       1,
		QuietHoursEnd:         8,
		Symbols:               []string{"BTCUSDT"},
	})
	if err != nil {
		t.Fatalf("update preferences failed: %v", err)
	}

	events, err := service.EvaluateAll(context.Background(), false)
	if err != nil {
		t.Fatalf("evaluate all failed: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected minimum confidence to suppress setup event, got=%d", len(events))
	}

	_, err = service.UpdatePreferences(AlertPreferences{
		FeishuEnabled:         true,
		BrowserEnabled:        true,
		SetupReadyEnabled:     true,
		DirectionShiftEnabled: false,
		NoTradeEnabled:        true,
		MinimumConfidence:     55,
		QuietHoursEnabled:     true,
		QuietHoursStart:       1,
		QuietHoursEnd:         8,
		Symbols:               []string{"BTCUSDT"},
	})
	if err != nil {
		t.Fatalf("update preferences failed: %v", err)
	}

	service.stateBySymbol = map[string]alertState{}
	events, err = service.EvaluateAll(context.Background(), false)
	if err != nil {
		t.Fatalf("evaluate all failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 setup event after lowering threshold, got=%d", len(events))
	}
	if len(events[0].Deliveries) != 1 || events[0].Deliveries[0].Status != "skipped" {
		t.Fatalf("expected quiet hours to skip feishu delivery, got=%+v", events[0].Deliveries)
	}
	if notifier.calls != 0 {
		t.Fatalf("expected notifier not to be called during quiet hours, got=%d", notifier.calls)
	}
}

type stubAlertNotifier struct {
	channel string
	calls   int
}

func (s *stubAlertNotifier) Channel() string {
	return s.channel
}

func (s *stubAlertNotifier) Notify(ctx context.Context, event AlertEvent) AlertDelivery {
	s.calls++
	return AlertDelivery{
		Channel: s.channel,
		Status:  "sent",
		SentAt:  1710000000000,
	}
}
