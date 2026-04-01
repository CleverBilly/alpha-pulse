package service

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"alpha-pulse/backend/models"
	"alpha-pulse/backend/pkg/binance"
	"alpha-pulse/backend/repository"
)

type FuturesSymbolRules = binance.FuturesSymbolRules
type FuturesOrder = binance.FuturesOrder
type FuturesPosition = binance.FuturesPosition

// TradeClient 描述自动交易依赖的期货客户端能力。
type TradeClient interface {
	GetFuturesBalance() (float64, error)
	GetFuturesLeverage(symbol string) (int, error)
	GetFuturesSymbolRules(symbol string) (binance.FuturesSymbolRules, error)
	PlaceFuturesLimitOrder(symbol, side string, qty, price float64) (binance.FuturesOrder, error)
	GetFuturesOrder(symbol, orderID string) (binance.FuturesOrder, error)
	CancelFuturesOrder(symbol string, orderID string) error
	PlaceFuturesProtectionOrder(symbol, side, orderType string, stopPrice float64) (string, error)
	CloseFuturesPosition(symbol, side string, qty float64) (string, error)
	GetFuturesPositions() ([]binance.FuturesPosition, error)
}

// TradeStaticConfig 表示部署层静态底线配置。
type TradeStaticConfig struct {
	Enabled        bool
	AutoExecute    bool
	AllowedSymbols []string
}

// TradeExecutorService 负责执行限价开仓与兜底收口。
type TradeExecutorService struct {
	client       TradeClient
	orderRepo    *repository.TradeOrderRepository
	accountCache atomic.Pointer[AccountStateCache] // nil 时降级为直接调用 API
	now          func() time.Time
}

// NewTradeExecutorService 创建 TradeExecutorService。
func NewTradeExecutorService(client TradeClient, orderRepo *repository.TradeOrderRepository) *TradeExecutorService {
	return &TradeExecutorService{
		client:    client,
		orderRepo: orderRepo,
		now:       time.Now,
	}
}

// SetAccountStateCache 注入账户状态缓存（可选）。注入后 ExecuteLimitEntry 优先读缓存，降低下单延迟。
func (s *TradeExecutorService) SetAccountStateCache(cache *AccountStateCache) {
	s.accountCache.Store(cache)
}

// ExecuteLimitEntry 提交限价开仓并创建 pending_fill 订单。
func (s *TradeExecutorService) ExecuteLimitEntry(ctx context.Context, event AlertEvent, settings TradeSettings) (models.TradeOrder, error) {
	if event.EntryPrice <= 0 || event.StopLoss <= 0 || event.TargetPrice <= 0 {
		return models.TradeOrder{}, fmt.Errorf("alert prices are incomplete")
	}
	if event.RiskReward < settings.MinRiskReward {
		return models.TradeOrder{}, fmt.Errorf("risk reward below threshold")
	}

	var balance float64
	var leverage int
	var rules FuturesSymbolRules
	var err error

	// cache.IsStale() 和 GetBalance/Leverage/Rules 各自独立获取读锁。
	// 这里的 TOCTOU 窗口是被有意接受的：并发的 Refresh() 只会提升数据新鲜度。
	// GetBalance 的硬过期（60s）提供最终安全兜底。
	cache := s.accountCache.Load()
	if cache != nil && !cache.IsStale() {
		balance, err = cache.GetBalance()
		if err != nil {
			return models.TradeOrder{}, err
		}
		leverage, err = cache.GetLeverage(event.Symbol)
		if err != nil {
			return models.TradeOrder{}, err
		}
		rules, err = cache.GetRules(event.Symbol)
		if err != nil {
			return models.TradeOrder{}, err
		}
	} else {
		// 降级路径：cache 未注入或已过期，直接调用 Binance API。
		balance, err = s.client.GetFuturesBalance()
		if err != nil {
			return models.TradeOrder{}, err
		}
		leverage, err = s.client.GetFuturesLeverage(event.Symbol)
		if err != nil {
			return models.TradeOrder{}, err
		}
		rules, err = s.client.GetFuturesSymbolRules(event.Symbol)
		if err != nil {
			return models.TradeOrder{}, err
		}
	}

	rawQty := (balance * settings.RiskPct / 100) * float64(leverage) / event.EntryPrice
	qty := floorToStep(rawQty, rules.StepSize, rules.QuantityPrecision)
	if qty < rules.MinQty {
		return models.TradeOrder{}, fmt.Errorf("calculated quantity below minimum")
	}

	entryOrder, err := s.client.PlaceFuturesLimitOrder(event.Symbol, inferTradeSide(event), qty, event.EntryPrice)
	if err != nil {
		return models.TradeOrder{}, err
	}

	order := models.TradeOrder{
		AlertID:         event.ID,
		Symbol:          event.Symbol,
		Side:            inferTradeSide(event),
		RequestedQty:    qty,
		EntryOrderID:    entryOrder.OrderID,
		EntryOrderType:  "LIMIT",
		LimitPrice:      event.EntryPrice,
		StopLoss:        event.StopLoss,
		TargetPrice:     event.TargetPrice,
		EntryStatus:     "pending_fill",
		Status:          "pending_fill",
		EntryExpiresAt:  s.now().Add(time.Duration(settings.EntryTimeoutSeconds) * time.Second).UnixMilli(),
		Source:          "system",
		CreatedAtUnixMs: s.now().UnixMilli(),
	}
	if err := s.orderRepo.Create(&order); err != nil {
		return models.TradeOrder{}, err
	}

	return order, nil
}

// CloseOrder 执行人工平仓。
func (s *TradeExecutorService) CloseOrder(ctx context.Context, id uint64) error {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return err
	}
	if order.Status != "open" {
		return fmt.Errorf("order is not open")
	}

	qty := order.FilledQty
	if qty <= 0 {
		qty = order.RequestedQty
	}
	closeID, err := s.client.CloseFuturesPosition(order.Symbol, oppositeTradeSide(order.Side), qty)
	if err != nil {
		return err
	}
	order.Status = "closed"
	order.CloseOrderID = closeID
	order.ClosedAt = s.now().UnixMilli()
	return s.orderRepo.Save(&order)
}

func inferTradeSide(event AlertEvent) string {
	if event.EntryPrice >= event.StopLoss {
		return "LONG"
	}
	return "SHORT"
}

func floorToStep(value, step float64, precision int) float64 {
	if step <= 0 {
		step = math.Pow10(-precision)
	}
	if step <= 0 {
		return value
	}
	scaled := math.Floor(value/step) * step
	if precision <= 0 {
		return scaled
	}
	factor := math.Pow10(precision)
	return math.Floor(scaled*factor) / factor
}
