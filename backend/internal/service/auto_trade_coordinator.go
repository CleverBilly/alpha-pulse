package service

import (
	"context"

	"alpha-pulse/backend/repository"
)

// AutoTradeCoordinator 负责把 setup_ready 告警转为自动执行动作。
type AutoTradeCoordinator struct {
	static      TradeStaticConfig
	settingsRepo *repository.TradeSettingRepository
	orderRepo   *repository.TradeOrderRepository
	executor    *TradeExecutorService
}

// NewAutoTradeCoordinator 创建 AutoTradeCoordinator。
func NewAutoTradeCoordinator(
	static TradeStaticConfig,
	settingsRepo *repository.TradeSettingRepository,
	orderRepo *repository.TradeOrderRepository,
	executor *TradeExecutorService,
) *AutoTradeCoordinator {
	return &AutoTradeCoordinator{
		static:      static,
		settingsRepo: settingsRepo,
		orderRepo:   orderRepo,
		executor:    executor,
	}
}

// HandleEvent 在 setup_ready 场景下触发自动下单。
func (c *AutoTradeCoordinator) HandleEvent(ctx context.Context, event AlertEvent) error {
	if event.Kind != "setup_ready" || !c.static.Enabled || !c.static.AutoExecute {
		return nil
	}

	settings := loadTradeSettings(c.settingsRepo, c.static.AllowedSymbols)
	if !settings.AutoExecuteEnabled || !settings.WatchesSymbol(event.Symbol) {
		return nil
	}

	openOrders, err := c.orderRepo.FindOpen(event.Symbol)
	if err != nil {
		return err
	}
	if len(openOrders) >= settings.MaxOpenPositions {
		return nil
	}

	_, err = c.executor.ExecuteLimitEntry(ctx, event, settings)
	return err
}
