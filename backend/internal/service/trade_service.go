package service

import (
	"context"
	"errors"

	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
)

var ErrTradeDisabled = errors.New("auto trading is disabled")

// TradeSettingsView 表示 trade-settings 接口载荷。
type TradeSettingsView struct {
	TradeEnabledEnv     bool     `json:"trade_enabled_env"`
	TradeAutoExecuteEnv bool     `json:"trade_auto_execute_env"`
	AllowedSymbolsEnv   []string `json:"allowed_symbols_env"`
	AutoExecuteEnabled  bool     `json:"auto_execute_enabled"`
	AllowedSymbols      []string `json:"allowed_symbols"`
	RiskPct             float64  `json:"risk_pct"`
	MinRiskReward       float64  `json:"min_risk_reward"`
	EntryTimeoutSeconds int      `json:"entry_timeout_seconds"`
	MaxOpenPositions    int      `json:"max_open_positions"`
	SyncEnabled         bool     `json:"sync_enabled"`
	UpdatedBy           string   `json:"updated_by"`
}

// TradeRuntimeStatus 表示交易运行时摘要。
type TradeRuntimeStatus struct {
	TradeEnabledEnv  bool `json:"trade_enabled_env"`
	AutoExecuteEnv   bool `json:"trade_auto_execute_env"`
	PendingFillCount int  `json:"pending_fill_count"`
	OpenCount        int  `json:"open_count"`
}

// TradeService 封装交易配置和订单查询能力。
type TradeService struct {
	static      TradeStaticConfig
	settingsRepo *repository.TradeSettingRepository
	orderRepo   *repository.TradeOrderRepository
	executor    *TradeExecutorService
	runtime     *TradeRuntime
}

// NewTradeService 创建 TradeService。
func NewTradeService(
	static TradeStaticConfig,
	settingsRepo *repository.TradeSettingRepository,
	orderRepo *repository.TradeOrderRepository,
	executor *TradeExecutorService,
	runtime *TradeRuntime,
) *TradeService {
	return &TradeService{
		static:      static,
		settingsRepo: settingsRepo,
		orderRepo:   orderRepo,
		executor:    executor,
		runtime:     runtime,
	}
}

// GetSettings 返回当前交易配置视图。
func (s *TradeService) GetSettings() TradeSettingsView {
	settings := loadTradeSettings(s.settingsRepo, s.static.AllowedSymbols)
	return TradeSettingsView{
		TradeEnabledEnv:     s.static.Enabled,
		TradeAutoExecuteEnv: s.static.AutoExecute,
		AllowedSymbolsEnv:   append([]string(nil), s.static.AllowedSymbols...),
		AutoExecuteEnabled:  settings.AutoExecuteEnabled,
		AllowedSymbols:      settings.AllowedSymbols,
		RiskPct:             settings.RiskPct,
		MinRiskReward:       settings.MinRiskReward,
		EntryTimeoutSeconds: settings.EntryTimeoutSeconds,
		MaxOpenPositions:    settings.MaxOpenPositions,
		SyncEnabled:         settings.SyncEnabled,
		UpdatedBy:           settings.UpdatedBy,
	}
}

// UpdateSettings 更新运行时交易配置。
func (s *TradeService) UpdateSettings(input TradeSettings) (TradeSettingsView, error) {
	if !s.static.Enabled {
		return TradeSettingsView{}, ErrTradeDisabled
	}

	settings, err := sanitizeTradeSettings(input, s.static.AllowedSymbols)
	if err != nil {
		return TradeSettingsView{}, err
	}
	record := projectTradeSettings(settings)
	if err := s.settingsRepo.Save(&record); err != nil {
		return TradeSettingsView{}, err
	}
	return s.GetSettings(), nil
}

// ListOrders 返回订单列表。
func (s *TradeService) ListOrders(limit int, symbol, status, source string) ([]models.TradeOrder, error) {
	return s.orderRepo.List(limit, symbol, status, source)
}

// GetRuntime 返回 runtime 摘要。
func (s *TradeService) GetRuntime() (TradeRuntimeStatus, error) {
	pending, err := s.orderRepo.FindPendingFill(1000)
	if err != nil {
		return TradeRuntimeStatus{}, err
	}
	openOrders, err := s.orderRepo.FindAllOpen()
	if err != nil {
		return TradeRuntimeStatus{}, err
	}
	return TradeRuntimeStatus{
		TradeEnabledEnv:  s.static.Enabled,
		AutoExecuteEnv:   s.static.AutoExecute,
		PendingFillCount: len(pending),
		OpenCount:        len(openOrders),
	}, nil
}

// CloseOrder 人工平仓指定订单。
func (s *TradeService) CloseOrder(ctx context.Context, id uint64) error {
	if !s.static.Enabled {
		return ErrTradeDisabled
	}
	return s.executor.CloseOrder(ctx, id)
}
