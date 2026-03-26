package service

import (
	"fmt"
	"strings"
	"time"

	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
	"gorm.io/gorm"
)

const alertPreferenceSingletonKey = "default"

type AlertPreferences struct {
	FeishuEnabled         bool     `json:"feishu_enabled"`
	BrowserEnabled        bool     `json:"browser_enabled"`
	SetupReadyEnabled     bool     `json:"setup_ready_enabled"`
	DirectionShiftEnabled bool     `json:"direction_shift_enabled"`
	NoTradeEnabled        bool     `json:"no_trade_enabled"`
	MinimumConfidence     int      `json:"minimum_confidence"`
	QuietHoursEnabled     bool     `json:"quiet_hours_enabled"`
	QuietHoursStart       int      `json:"quiet_hours_start"`
	QuietHoursEnd         int      `json:"quiet_hours_end"`
	SoundEnabled          bool     `json:"sound_enabled"`
	Symbols               []string `json:"symbols"`
	AvailableSymbols      []string `json:"available_symbols"`
}

func defaultAlertPreferences(symbols []string) AlertPreferences {
	return AlertPreferences{
		FeishuEnabled:         true,
		BrowserEnabled:        true,
		SetupReadyEnabled:     true,
		DirectionShiftEnabled: true,
		NoTradeEnabled:        true,
		MinimumConfidence:     55,
		QuietHoursEnabled:     false,
		QuietHoursStart:       0,
		QuietHoursEnd:         8,
		SoundEnabled:          false,
		Symbols:               append([]string(nil), normalizeAlertSymbols(symbols)...),
		AvailableSymbols:      append([]string(nil), normalizeAlertSymbols(symbols)...),
	}
}

func sanitizeAlertPreferences(input AlertPreferences, allowedSymbols []string) (AlertPreferences, error) {
	symbols := filterAlertSymbols(input.Symbols, allowedSymbols)
	if len(symbols) == 0 {
		symbols = normalizeAlertSymbols(allowedSymbols)
	}

	minConfidence := boundedInt(input.MinimumConfidence, 0, 100)
	start := input.QuietHoursStart
	end := input.QuietHoursEnd
	if start < 0 || start > 23 || end < 0 || end > 23 {
		return AlertPreferences{}, fmt.Errorf("quiet hours must be within 0-23")
	}
	if input.QuietHoursEnabled && start == end {
		return AlertPreferences{}, fmt.Errorf("quiet hours start and end cannot be the same")
	}

	return AlertPreferences{
		FeishuEnabled:         input.FeishuEnabled,
		BrowserEnabled:        input.BrowserEnabled,
		SetupReadyEnabled:     input.SetupReadyEnabled,
		DirectionShiftEnabled: input.DirectionShiftEnabled,
		NoTradeEnabled:        input.NoTradeEnabled,
		MinimumConfidence:     minConfidence,
		QuietHoursEnabled:     input.QuietHoursEnabled,
		QuietHoursStart:       start,
		QuietHoursEnd:         end,
		SoundEnabled:          input.SoundEnabled,
		Symbols:               symbols,
		AvailableSymbols:      append([]string(nil), normalizeAlertSymbols(allowedSymbols)...),
	}, nil
}

func filterAlertSymbols(candidate []string, allowed []string) []string {
	if len(allowed) == 0 {
		return normalizeAlertSymbols(candidate)
	}

	allowedSet := make(map[string]struct{}, len(allowed))
	for _, symbol := range normalizeAlertSymbols(allowed) {
		allowedSet[symbol] = struct{}{}
	}

	filtered := make([]string, 0, len(candidate))
	for _, symbol := range normalizeAlertSymbols(candidate) {
		if _, ok := allowedSet[symbol]; ok {
			filtered = append(filtered, symbol)
		}
	}
	return filtered
}

func loadAlertPreferences(repo *repository.AlertPreferenceRepository, symbols []string) AlertPreferences {
	defaults := defaultAlertPreferences(symbols)
	if repo == nil {
		return defaults
	}

	record, err := repo.GetDefault()
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return defaults
		}
		return defaults
	}
	return hydrateAlertPreferences(record, symbols)
}

func persistAlertPreferences(repo *repository.AlertPreferenceRepository, prefs AlertPreferences) error {
	if repo == nil {
		return nil
	}
	record := projectAlertPreferences(prefs)
	return repo.Save(&record)
}

func projectAlertPreferences(prefs AlertPreferences) models.AlertPreference {
	return models.AlertPreference{
		SingletonKey:          alertPreferenceSingletonKey,
		FeishuEnabled:         prefs.FeishuEnabled,
		BrowserEnabled:        prefs.BrowserEnabled,
		SetupReadyEnabled:     prefs.SetupReadyEnabled,
		DirectionShiftEnabled: prefs.DirectionShiftEnabled,
		NoTradeEnabled:        prefs.NoTradeEnabled,
		MinimumConfidence:     boundedInt(prefs.MinimumConfidence, 0, 100),
		QuietHoursEnabled:     prefs.QuietHoursEnabled,
		QuietHoursStart:       prefs.QuietHoursStart,
		QuietHoursEnd:         prefs.QuietHoursEnd,
		SoundEnabled:          prefs.SoundEnabled,
		WatchedSymbols:        strings.Join(normalizeAlertSymbols(prefs.Symbols), ","),
	}
}

func hydrateAlertPreferences(record models.AlertPreference, allowedSymbols []string) AlertPreferences {
	prefs, err := sanitizeAlertPreferences(AlertPreferences{
		FeishuEnabled:         record.FeishuEnabled,
		BrowserEnabled:        record.BrowserEnabled,
		SetupReadyEnabled:     record.SetupReadyEnabled,
		DirectionShiftEnabled: record.DirectionShiftEnabled,
		NoTradeEnabled:        record.NoTradeEnabled,
		MinimumConfidence:     record.MinimumConfidence,
		QuietHoursEnabled:     record.QuietHoursEnabled,
		QuietHoursStart:       record.QuietHoursStart,
		QuietHoursEnd:         record.QuietHoursEnd,
		SoundEnabled:          record.SoundEnabled,
		Symbols:               strings.Split(record.WatchedSymbols, ","),
	}, allowedSymbols)
	if err != nil {
		return defaultAlertPreferences(allowedSymbols)
	}
	return prefs
}

func (p AlertPreferences) WatchesSymbol(symbol string) bool {
	if symbol == "" {
		return false
	}
	target := strings.ToUpper(strings.TrimSpace(symbol))
	for _, item := range p.Symbols {
		if item == target {
			return true
		}
	}
	return false
}

func (p AlertPreferences) AllowsEvent(event AlertEvent) bool {
	if !p.WatchesSymbol(event.Symbol) {
		return false
	}
	switch event.Kind {
	case "setup_ready":
		if !p.SetupReadyEnabled {
			return false
		}
	case "direction_shift":
		if !p.DirectionShiftEnabled {
			return false
		}
	case "no_trade":
		if !p.NoTradeEnabled {
			return false
		}
	}
	if event.Kind != "no_trade" && event.Confidence < p.MinimumConfidence {
		return false
	}
	return true
}

func (p AlertPreferences) FeishuSuppressedAt(now time.Time) (bool, string) {
	if !p.FeishuEnabled {
		return true, "飞书推送已关闭"
	}
	if !p.QuietHoursEnabled {
		return false, ""
	}
	hour := now.In(time.Local).Hour()
	if p.QuietHoursStart < p.QuietHoursEnd {
		if hour >= p.QuietHoursStart && hour < p.QuietHoursEnd {
			return true, "静默时段内跳过飞书推送"
		}
		return false, ""
	}
	if hour >= p.QuietHoursStart || hour < p.QuietHoursEnd {
		return true, "静默时段内跳过飞书推送"
	}
	return false, ""
}
