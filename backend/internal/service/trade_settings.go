package service

import (
	"fmt"
	"strings"

	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
)

const tradeSettingSingletonKey = "default"

// TradeSettings 表示自动交易运行时配置。
type TradeSettings struct {
	AutoExecuteEnabled  bool     `json:"auto_execute_enabled"`
	AllowedSymbols      []string `json:"allowed_symbols"`
	RiskPct             float64  `json:"risk_pct"`
	MinRiskReward       float64  `json:"min_risk_reward"`
	EntryTimeoutSeconds int      `json:"entry_timeout_seconds"`
	MaxOpenPositions    int      `json:"max_open_positions"`
	SyncEnabled         bool     `json:"sync_enabled"`
	UpdatedBy           string   `json:"updated_by"`
}

func sanitizeTradeSettings(input TradeSettings, allowedSymbols []string) (TradeSettings, error) {
	symbols := filterTradeSymbols(input.AllowedSymbols, allowedSymbols)
	if len(symbols) == 0 {
		return TradeSettings{}, fmt.Errorf("at least one allowed symbol is required")
	}
	if input.RiskPct <= 0 || input.RiskPct > 10 {
		return TradeSettings{}, fmt.Errorf("risk_pct must be within (0, 10]")
	}
	if input.MinRiskReward < 1 {
		return TradeSettings{}, fmt.Errorf("min_risk_reward must be >= 1")
	}
	if input.EntryTimeoutSeconds <= 0 {
		return TradeSettings{}, fmt.Errorf("entry_timeout_seconds must be > 0")
	}
	if input.MaxOpenPositions <= 0 {
		return TradeSettings{}, fmt.Errorf("max_open_positions must be > 0")
	}

	return TradeSettings{
		AutoExecuteEnabled:  input.AutoExecuteEnabled,
		AllowedSymbols:      symbols,
		RiskPct:             input.RiskPct,
		MinRiskReward:       input.MinRiskReward,
		EntryTimeoutSeconds: input.EntryTimeoutSeconds,
		MaxOpenPositions:    input.MaxOpenPositions,
		SyncEnabled:         input.SyncEnabled,
		UpdatedBy:           strings.TrimSpace(input.UpdatedBy),
	}, nil
}

func filterTradeSymbols(candidate []string, allowed []string) []string {
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

func projectTradeSettings(settings TradeSettings) models.TradeSetting {
	return models.TradeSetting{
		SingletonKey:        tradeSettingSingletonKey,
		AutoExecuteEnabled:  settings.AutoExecuteEnabled,
		AllowedSymbols:      strings.Join(normalizeAlertSymbols(settings.AllowedSymbols), ","),
		RiskPct:             settings.RiskPct,
		MinRiskReward:       settings.MinRiskReward,
		EntryTimeoutSeconds: settings.EntryTimeoutSeconds,
		MaxOpenPositions:    settings.MaxOpenPositions,
		SyncEnabled:         settings.SyncEnabled,
		UpdatedBy:           settings.UpdatedBy,
	}
}

func defaultTradeSettings(symbols []string) TradeSettings {
	return TradeSettings{
		AutoExecuteEnabled:  false,
		AllowedSymbols:      append([]string(nil), normalizeAlertSymbols(symbols)...),
		RiskPct:             2,
		MinRiskReward:       1,
		EntryTimeoutSeconds: 45,
		MaxOpenPositions:    1,
		SyncEnabled:         true,
	}
}

func hydrateTradeSettings(record models.TradeSetting, allowedSymbols []string) TradeSettings {
	settings, err := sanitizeTradeSettings(TradeSettings{
		AutoExecuteEnabled:  record.AutoExecuteEnabled,
		AllowedSymbols:      strings.Split(record.AllowedSymbols, ","),
		RiskPct:             record.RiskPct,
		MinRiskReward:       record.MinRiskReward,
		EntryTimeoutSeconds: record.EntryTimeoutSeconds,
		MaxOpenPositions:    record.MaxOpenPositions,
		SyncEnabled:         record.SyncEnabled,
		UpdatedBy:           record.UpdatedBy,
	}, allowedSymbols)
	if err != nil {
		return defaultTradeSettings(allowedSymbols)
	}
	return settings
}

func loadTradeSettings(repo *repository.TradeSettingRepository, allowedSymbols []string) TradeSettings {
	defaults := defaultTradeSettings(allowedSymbols)
	if repo == nil {
		return defaults
	}

	record, err := repo.GetDefault()
	if err != nil {
		return defaults
	}
	return hydrateTradeSettings(record, allowedSymbols)
}

func (t TradeSettings) WatchesSymbol(symbol string) bool {
	target := strings.ToUpper(strings.TrimSpace(symbol))
	for _, item := range normalizeAlertSymbols(t.AllowedSymbols) {
		if item == target {
			return true
		}
	}
	return false
}
