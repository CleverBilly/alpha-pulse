package service

import "alpha-pulse/backend/models"

func (s *SignalService) loadSnapshotMicrostructureEvents(
	symbol, interval string,
	klines []models.Kline,
	currentOrderFlow models.OrderFlow,
) ([]models.MicrostructureEvent, error) {
	fallback := projectMicrostructureEvents(currentOrderFlow)
	if s.microEventRepo == nil || len(klines) == 0 {
		return fallback, nil
	}

	fromTradeTime := klines[0].OpenTime
	toTradeTime := klines[len(klines)-1].OpenTime + intervalDurationMillis(interval)
	limit := clampInt(len(klines)*8, 24, 320)

	events, err := s.microEventRepo.GetByTradeTimeRange(symbol, interval, fromTradeTime, toTradeTime, limit)
	if err != nil {
		return nil, err
	}
	if len(events) > 0 {
		return events, nil
	}

	return fallback, nil
}
