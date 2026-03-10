package service

import (
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
)

func persistLargeTradeEvents(
	repo *repository.LargeTradeEventRepository,
	orderFlow models.OrderFlow,
) error {
	events := projectLargeTradeEvents(orderFlow)
	if repo == nil || len(events) == 0 {
		return nil
	}

	return repo.CreateBatch(events)
}

func projectLargeTradeEvents(orderFlow models.OrderFlow) []models.LargeTradeEvent {
	if len(orderFlow.LargeTrades) == 0 || orderFlow.Symbol == "" {
		return nil
	}

	events := make([]models.LargeTradeEvent, 0, len(orderFlow.LargeTrades))
	for _, trade := range orderFlow.LargeTrades {
		events = append(events, models.LargeTradeEvent{
			OrderFlowID:  orderFlow.ID,
			Symbol:       orderFlow.Symbol,
			AggTradeID:   trade.AggTradeID,
			IntervalType: orderFlow.IntervalType,
			OpenTime:     orderFlow.OpenTime,
			Side:         trade.Side,
			Price:        trade.Price,
			Quantity:     trade.Quantity,
			Notional:     trade.Notional,
			TradeTime:    trade.TradeTime,
			CreatedAt:    orderFlow.CreatedAt,
		})
	}

	return events
}
