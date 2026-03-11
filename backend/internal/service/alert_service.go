package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
	"gorm.io/gorm"
)

const (
	alertMacroInterval     = "4h"
	alertBiasInterval      = "1h"
	alertTriggerInterval   = "15m"
	alertExecutionInterval = "5m"
	alertSnapshotLimit     = 48
)

type DirectionSnapshotFetcher interface {
	GetMarketSnapshotWithRefresh(symbol, interval string, limit int, refresh bool) (MarketSnapshot, error)
}

type AlertDelivery struct {
	Channel string `json:"channel"`
	Status  string `json:"status"`
	Detail  string `json:"detail,omitempty"`
	SentAt  int64  `json:"sent_at,omitempty"`
}

type AlertEvent struct {
	ID                string          `json:"id"`
	Symbol            string          `json:"symbol"`
	Kind              string          `json:"kind"`
	Severity          string          `json:"severity"`
	Title             string          `json:"title"`
	Verdict           string          `json:"verdict"`
	TradeabilityLabel string          `json:"tradeability_label"`
	Summary           string          `json:"summary"`
	Reasons           []string        `json:"reasons"`
	TimeframeLabels   []string        `json:"timeframe_labels"`
	Confidence        int             `json:"confidence"`
	RiskLabel         string          `json:"risk_label"`
	EntryPrice        float64         `json:"entry_price"`
	StopLoss          float64         `json:"stop_loss"`
	TargetPrice       float64         `json:"target_price"`
	RiskReward        float64         `json:"risk_reward"`
	CreatedAt         int64           `json:"created_at"`
	Deliveries        []AlertDelivery `json:"deliveries"`
}

type AlertFeed struct {
	Items     []AlertEvent `json:"items"`
	Generated int          `json:"generated"`
}

type AlertNotifier interface {
	Channel() string
	Notify(ctx context.Context, event AlertEvent) AlertDelivery
}

type AlertService struct {
	fetcher        DirectionSnapshotFetcher
	repo           *repository.AlertRecordRepository
	preferenceRepo *repository.AlertPreferenceRepository
	symbols        []string
	historyLimit   int
	notifiers      []AlertNotifier
	now            func() time.Time

	mu            sync.RWMutex
	recent        []AlertEvent
	stateBySymbol map[string]alertState
}

type alertState struct {
	State      DirectionDecisionState
	Tradable   bool
	SetupReady bool
}

func NewAlertService(fetcher DirectionSnapshotFetcher, repo *repository.AlertRecordRepository, preferenceRepo *repository.AlertPreferenceRepository, symbols []string, historyLimit int, notifiers ...AlertNotifier) *AlertService {
	normalized := normalizeAlertSymbols(symbols)
	if len(normalized) == 0 {
		normalized = []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}
	}
	if historyLimit <= 0 {
		historyLimit = 40
	}

	return &AlertService{
		fetcher:        fetcher,
		repo:           repo,
		preferenceRepo: preferenceRepo,
		symbols:        normalized,
		historyLimit:   historyLimit,
		notifiers:      notifiers,
		now:            time.Now,
		recent:         make([]AlertEvent, 0, historyLimit),
		stateBySymbol:  make(map[string]alertState, len(normalized)),
	}
}

func normalizeAlertSymbols(symbols []string) []string {
	if len(symbols) == 0 {
		return nil
	}

	result := make([]string, 0, len(symbols))
	seen := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func (s *AlertService) ListRecent(limit int) []AlertEvent {
	if limit <= 0 {
		limit = 20
	}

	if s.repo != nil {
		records, err := s.repo.ListRecent(limit)
		if err == nil {
			return hydrateAlertEvents(records)
		}
		log.Printf("alert recent history query failed: %v", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit > len(s.recent) {
		limit = len(s.recent)
	}
	result := make([]AlertEvent, limit)
	copy(result, s.recent[:limit])
	return result
}

func (s *AlertService) ListHistory(limit int) []AlertEvent {
	return s.ListRecent(limit)
}

func (s *AlertService) GetPreferences() AlertPreferences {
	return loadAlertPreferences(s.preferenceRepo, s.symbols)
}

func (s *AlertService) UpdatePreferences(input AlertPreferences) (AlertPreferences, error) {
	preferences, err := sanitizeAlertPreferences(input, s.symbols)
	if err != nil {
		return AlertPreferences{}, err
	}
	if err := persistAlertPreferences(s.preferenceRepo, preferences); err != nil {
		return AlertPreferences{}, err
	}
	return preferences, nil
}

func (s *AlertService) EvaluateAll(ctx context.Context, refresh bool) ([]AlertEvent, error) {
	if s.fetcher == nil {
		return nil, fmt.Errorf("alert fetcher is not configured")
	}

	preferences := s.GetPreferences()
	generated := make([]AlertEvent, 0, len(s.symbols))
	var lastErr error
	successCount := 0

	for _, symbol := range s.symbols {
		event, err := s.evaluateSymbol(ctx, symbol, refresh, preferences)
		if err != nil {
			lastErr = err
			log.Printf("alert evaluation failed for %s: %v", symbol, err)
			continue
		}
		successCount++
		if event != nil {
			generated = append(generated, *event)
		}
	}

	if successCount == 0 && lastErr != nil {
		return nil, lastErr
	}
	return generated, nil
}

func (s *AlertService) evaluateSymbol(ctx context.Context, symbol string, refresh bool, preferences AlertPreferences) (*AlertEvent, error) {
	macroSnapshot, err := s.fetcher.GetMarketSnapshotWithRefresh(symbol, alertMacroInterval, alertSnapshotLimit, refresh)
	if err != nil {
		return nil, err
	}
	biasSnapshot, err := s.fetcher.GetMarketSnapshotWithRefresh(symbol, alertBiasInterval, alertSnapshotLimit, refresh)
	if err != nil {
		return nil, err
	}
	triggerSnapshot, err := s.fetcher.GetMarketSnapshotWithRefresh(symbol, alertTriggerInterval, alertSnapshotLimit, refresh)
	if err != nil {
		return nil, err
	}
	executionSnapshot, err := s.fetcher.GetMarketSnapshotWithRefresh(symbol, alertExecutionInterval, alertSnapshotLimit, refresh)
	if err != nil {
		return nil, err
	}

	decision := BuildDirectionCopilotDecision(&macroSnapshot, &biasSnapshot, &triggerSnapshot, &executionSnapshot)
	setupReady := decision.Tradable && HasExecutableSetup(biasSnapshot)

	s.mu.Lock()
	defer s.mu.Unlock()

	previous, hasPrevious := s.stateBySymbol[symbol]
	if !hasPrevious {
		previous, hasPrevious = s.loadPersistedState(symbol)
	}
	current := alertState{
		State:      decision.State,
		Tradable:   decision.Tradable,
		SetupReady: setupReady,
	}
	s.stateBySymbol[symbol] = current

	event := s.resolveAlertEvent(symbol, decision, biasSnapshot, current, previous, hasPrevious)
	if event == nil {
		return nil, nil
	}
	if !preferences.AllowsEvent(*event) {
		return nil, nil
	}

	event.Deliveries = s.deliver(ctx, *event, preferences)
	if err := s.persistAlert(*event, current); err != nil {
		return nil, err
	}
	s.recent = append([]AlertEvent{*event}, s.recent...)
	if len(s.recent) > s.historyLimit {
		s.recent = s.recent[:s.historyLimit]
	}
	return event, nil
}

func (s *AlertService) resolveAlertEvent(
	symbol string,
	decision DirectionDecision,
	biasSnapshot MarketSnapshot,
	current alertState,
	previous alertState,
	hasPrevious bool,
) *AlertEvent {
	switch {
	case current.Tradable && current.SetupReady:
		if !hasPrevious || !previous.SetupReady || previous.State != current.State {
			return buildAlertEvent(s.now(), symbol, "setup_ready", "A", decision, biasSnapshot, "A 级 setup 已就绪")
		}
	case !current.Tradable:
		if hasPrevious && (previous.Tradable || previous.SetupReady) {
			return buildAlertEvent(s.now(), symbol, "no_trade", "warning", decision, biasSnapshot, "进入 No-Trade")
		}
	case current.Tradable && (!hasPrevious || !previous.Tradable || previous.State != current.State):
		return buildAlertEvent(s.now(), symbol, "direction_shift", "info", decision, biasSnapshot, fmt.Sprintf("方向切换为 %s", decision.Verdict))
	}

	return nil
}

func (s *AlertService) deliver(ctx context.Context, event AlertEvent, preferences AlertPreferences) []AlertDelivery {
	if len(s.notifiers) == 0 {
		return []AlertDelivery{}
	}

	deliveries := make([]AlertDelivery, 0, len(s.notifiers))
	for _, notifier := range s.notifiers {
		if notifier.Channel() == "feishu" {
			if suppressed, detail := preferences.FeishuSuppressedAt(s.now()); suppressed {
				deliveries = append(deliveries, AlertDelivery{
					Channel: notifier.Channel(),
					Status:  "skipped",
					Detail:  detail,
				})
				continue
			}
		}
		deliveries = append(deliveries, notifier.Notify(ctx, event))
	}
	return deliveries
}

func (s *AlertService) loadPersistedState(symbol string) (alertState, bool) {
	if s.repo == nil {
		return alertState{}, false
	}

	record, err := s.repo.GetLatestBySymbol(symbol)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Printf("load persisted alert state failed for %s: %v", symbol, err)
		}
		return alertState{}, false
	}

	state := alertState{
		State:      DirectionDecisionState(strings.TrimSpace(record.DirectionState)),
		Tradable:   record.Tradable,
		SetupReady: record.SetupReady,
	}
	if state.State == "" {
		state.State = DirectionStateInvalid
	}
	return state, true
}

func (s *AlertService) persistAlert(event AlertEvent, state alertState) error {
	if s.repo == nil {
		return nil
	}

	record, err := projectAlertRecord(event, state)
	if err != nil {
		return err
	}
	return s.repo.Create(&record)
}

func hydrateAlertEvents(records []models.AlertRecord) []AlertEvent {
	items := make([]AlertEvent, 0, len(records))
	for _, record := range records {
		items = append(items, hydrateAlertEvent(record))
	}
	return items
}

func buildAlertEvent(now time.Time, symbol string, kind string, severity string, decision DirectionDecision, biasSnapshot MarketSnapshot, titleSuffix string) *AlertEvent {
	return &AlertEvent{
		ID:                fmt.Sprintf("%s-%s-%d", symbol, kind, now.UnixMilli()),
		Symbol:            symbol,
		Kind:              kind,
		Severity:          severity,
		Title:             fmt.Sprintf("%s %s", symbol, titleSuffix),
		Verdict:           decision.Verdict,
		TradeabilityLabel: decision.TradeabilityLabel,
		Summary:           decision.Summary,
		Reasons:           append([]string(nil), decision.Reasons...),
		TimeframeLabels:   append([]string(nil), decision.TimeframeLabels...),
		Confidence:        decision.Confidence,
		RiskLabel:         decision.RiskLabel,
		EntryPrice:        biasSnapshot.Signal.EntryPrice,
		StopLoss:          biasSnapshot.Signal.StopLoss,
		TargetPrice:       biasSnapshot.Signal.TargetPrice,
		RiskReward:        biasSnapshot.Signal.RiskReward,
		CreatedAt:         now.UnixMilli(),
	}
}
