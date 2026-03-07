package service

import (
	"sort"
	"strconv"

	"alpha-pulse/backend/internal/orderflow"
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

func enrichOrderFlowMicrostructureWithOrderBook(
	engine *orderflow.Engine,
	orderBookRepo *repository.OrderBookSnapshotRepository,
	symbol string,
	orderFlow *models.OrderFlow,
) error {
	if engine == nil || orderBookRepo == nil || orderFlow == nil || symbol == "" {
		return nil
	}

	snapshots, err := orderBookRepo.GetRecent(symbol, engine.OrderBookHistoryLimit())
	if err != nil {
		return err
	}
	if len(snapshots) == 0 {
		return nil
	}

	extraEvents, err := engine.AnalyzeOrderBookMicrostructure(symbol, snapshots)
	if err != nil {
		return err
	}
	orderFlow.MicrostructureEvents = mergeMicrostructureEvents(orderFlow.MicrostructureEvents, extraEvents)
	return nil
}

func mergeMicrostructureEvents(
	base []models.OrderFlowMicrostructureEvent,
	extra []models.OrderFlowMicrostructureEvent,
) []models.OrderFlowMicrostructureEvent {
	if len(extra) == 0 {
		return base
	}

	merged := make([]models.OrderFlowMicrostructureEvent, 0, len(base)+len(extra))
	seen := make(map[string]struct{}, len(base)+len(extra))

	appendUnique := func(event models.OrderFlowMicrostructureEvent) {
		key := event.Type + "|" + event.Bias + "|" + formatInt64(event.TradeTime) + "|" + formatMicroPrice(event.Price)
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		merged = append(merged, event)
	}

	for _, event := range base {
		appendUnique(event)
	}
	for _, event := range extra {
		appendUnique(event)
	}

	sort.Slice(merged, func(i, j int) bool {
		if merged[i].TradeTime == merged[j].TradeTime {
			if merged[i].Score == merged[j].Score {
				return merged[i].Type < merged[j].Type
			}
			return merged[i].Score < merged[j].Score
		}
		return merged[i].TradeTime < merged[j].TradeTime
	})

	if len(merged) <= 18 {
		return merged
	}
	return merged[len(merged)-18:]
}

func formatInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}

func formatMicroPrice(value float64) string {
	return strconv.FormatFloat(value, 'f', 4, 64)
}
