package service

import (
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
)

func persistMicrostructureEvents(
	repo *repository.MicrostructureEventRepository,
	orderFlow models.OrderFlow,
) error {
	events := projectMicrostructureEvents(orderFlow)
	if repo == nil || len(events) == 0 {
		return nil
	}

	return repo.CreateBatch(events)
}

func projectMicrostructureEvents(orderFlow models.OrderFlow) []models.MicrostructureEvent {
	if len(orderFlow.MicrostructureEvents) == 0 || orderFlow.Symbol == "" {
		return nil
	}

	events := make([]models.MicrostructureEvent, 0, len(orderFlow.MicrostructureEvents))
	for _, event := range orderFlow.MicrostructureEvents {
		events = append(events, models.MicrostructureEvent{
			OrderFlowID:  orderFlow.ID,
			Symbol:       orderFlow.Symbol,
			IntervalType: orderFlow.IntervalType,
			OpenTime:     orderFlow.OpenTime,
			EventType:    event.Type,
			Bias:         event.Bias,
			Score:        event.Score,
			Strength:     event.Strength,
			Price:        event.Price,
			TradeTime:    event.TradeTime,
			Detail:       event.Detail,
			CreatedAt:    orderFlow.CreatedAt,
		})
	}

	return events
}

func hydrateOrderFlowMicrostructure(orderFlow *models.OrderFlow, events []models.MicrostructureEvent) {
	if orderFlow == nil || len(events) == 0 {
		return
	}

	orderFlow.MicrostructureEvents = make([]models.OrderFlowMicrostructureEvent, 0, len(events))
	for _, event := range events {
		orderFlow.MicrostructureEvents = append(orderFlow.MicrostructureEvents, models.OrderFlowMicrostructureEvent{
			Type:      event.EventType,
			Bias:      event.Bias,
			Score:     event.Score,
			Strength:  event.Strength,
			Price:     event.Price,
			TradeTime: event.TradeTime,
			Detail:    event.Detail,
		})
	}
}
