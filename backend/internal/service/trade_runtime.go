package service

import (
	"context"
	"strings"
	"time"

	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
)

// TradeRuntime 负责盯单和持仓同步。
type TradeRuntime struct {
	client   TradeClient
	orderRepo *repository.TradeOrderRepository
	now      func() time.Time
}

// NewTradeRuntime 创建 TradeRuntime。
func NewTradeRuntime(client TradeClient, orderRepo *repository.TradeOrderRepository) *TradeRuntime {
	return &TradeRuntime{
		client:   client,
		orderRepo: orderRepo,
		now:      time.Now,
	}
}

// ReconcilePendingEntries 推进 pending_fill 订单状态。
func (r *TradeRuntime) ReconcilePendingEntries(ctx context.Context) error {
	orders, err := r.orderRepo.FindPendingFill(100)
	if err != nil {
		return err
	}

	for i := range orders {
		order := orders[i]
		status, err := r.client.GetFuturesOrder(order.Symbol, order.EntryOrderID)
		if err != nil {
			return err
		}

		if strings.EqualFold(status.Status, "FILLED") {
			if _, err := r.client.PlaceFuturesProtectionOrder(order.Symbol, oppositeTradeSide(order.Side), "STOP_MARKET", order.StopLoss); err != nil {
				return r.markFailed(order, err.Error())
			}
			if _, err := r.client.PlaceFuturesProtectionOrder(order.Symbol, oppositeTradeSide(order.Side), "TAKE_PROFIT_MARKET", order.TargetPrice); err != nil {
				return r.markFailed(order, err.Error())
			}

			order.EntryStatus = "filled"
			order.Status = "open"
			order.FilledQty = status.FilledQty
			order.FilledPrice = status.FilledPrice
			if order.FilledQty == 0 {
				order.FilledQty = order.RequestedQty
			}
			if err := r.orderRepo.Save(&order); err != nil {
				return err
			}
			continue
		}

		if r.now().UnixMilli() >= order.EntryExpiresAt {
			if err := r.client.CancelFuturesOrder(order.Symbol, order.EntryOrderID); err != nil {
				return err
			}
			order.EntryStatus = "expired"
			order.Status = "expired"
			order.ClosedAt = r.now().UnixMilli()
			if err := r.orderRepo.Save(&order); err != nil {
				return err
			}
		}
	}

	return nil
}

// SyncPositions 对齐交易所持仓与本地状态。
func (r *TradeRuntime) SyncPositions(ctx context.Context) error {
	positions, err := r.client.GetFuturesPositions()
	if err != nil {
		return err
	}

	remote := make(map[string]FuturesPosition, len(positions))
	for _, position := range positions {
		remote[position.Symbol] = position

		existing, err := r.orderRepo.FindOpen(position.Symbol)
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			continue
		}

		manual := models.TradeOrder{
			Symbol:          position.Symbol,
			Side:            position.Side,
			RequestedQty:    position.Qty,
			FilledQty:       position.Qty,
			FilledPrice:     position.EntryPrice,
			EntryStatus:     "filled",
			Status:          "open",
			Source:          "manual",
			CreatedAtUnixMs: r.now().UnixMilli(),
		}
		if err := r.orderRepo.Create(&manual); err != nil {
			return err
		}
	}

	openOrders, err := r.orderRepo.FindAllOpen()
	if err != nil {
		return err
	}
	for i := range openOrders {
		order := openOrders[i]
		if _, ok := remote[order.Symbol]; ok {
			continue
		}
		order.Status = "closed"
		order.ClosedAt = r.now().UnixMilli()
		if err := r.orderRepo.Save(&order); err != nil {
			return err
		}
	}

	return nil
}

func (r *TradeRuntime) markFailed(order models.TradeOrder, reason string) error {
	if _, err := r.client.CloseFuturesPosition(order.Symbol, oppositeTradeSide(order.Side), order.RequestedQty); err != nil {
		return err
	}
	order.Status = "failed"
	order.CloseReason = reason
	order.ClosedAt = r.now().UnixMilli()
	return r.orderRepo.Save(&order)
}

func oppositeTradeSide(side string) string {
	if strings.EqualFold(side, "LONG") {
		return "SHORT"
	}
	return "LONG"
}
