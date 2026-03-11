package service

import (
	"encoding/json"

	"alpha-pulse/backend/models"
)

func projectAlertRecord(event AlertEvent, state alertState) (models.AlertRecord, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return models.AlertRecord{}, err
	}

	return models.AlertRecord{
		AlertID:           event.ID,
		Symbol:            event.Symbol,
		Kind:              event.Kind,
		Severity:          event.Severity,
		DirectionState:    string(state.State),
		Tradable:          state.Tradable,
		SetupReady:        state.SetupReady,
		TradeabilityLabel: event.TradeabilityLabel,
		Title:             event.Title,
		Verdict:           event.Verdict,
		Summary:           event.Summary,
		Confidence:        event.Confidence,
		RiskLabel:         event.RiskLabel,
		EntryPrice:        event.EntryPrice,
		StopLoss:          event.StopLoss,
		TargetPrice:       event.TargetPrice,
		RiskReward:        event.RiskReward,
		EventTime:         event.CreatedAt,
		PayloadJSON:       string(payload),
	}, nil
}

func hydrateAlertEvent(record models.AlertRecord) AlertEvent {
	var event AlertEvent
	if err := json.Unmarshal([]byte(record.PayloadJSON), &event); err == nil && event.ID != "" {
		return event
	}

	return AlertEvent{
		ID:                record.AlertID,
		Symbol:            record.Symbol,
		Kind:              record.Kind,
		Severity:          record.Severity,
		Title:             record.Title,
		Verdict:           record.Verdict,
		TradeabilityLabel: record.TradeabilityLabel,
		Summary:           record.Summary,
		Confidence:        record.Confidence,
		RiskLabel:         record.RiskLabel,
		EntryPrice:        record.EntryPrice,
		StopLoss:          record.StopLoss,
		TargetPrice:       record.TargetPrice,
		RiskReward:        record.RiskReward,
		CreatedAt:         record.EventTime,
		Deliveries:        []AlertDelivery{},
	}
}
