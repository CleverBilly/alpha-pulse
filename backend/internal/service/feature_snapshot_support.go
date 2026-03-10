package service

import (
	"encoding/json"
	"errors"

	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
)

const featureSnapshotVersion = "v1"

func persistFeatureSnapshot(
	repo *repository.FeatureSnapshotRepository,
	snapshot MarketSnapshot,
) error {
	if repo == nil {
		return nil
	}

	record, err := projectFeatureSnapshot(snapshot)
	if err != nil {
		return err
	}

	return repo.Create(&record)
}

func projectFeatureSnapshot(snapshot MarketSnapshot) (models.FeatureSnapshot, error) {
	symbol := snapshot.Price.Symbol
	if symbol == "" {
		symbol = snapshot.Signal.Symbol
	}
	if symbol == "" {
		symbol = snapshot.OrderFlow.Symbol
	}

	interval := snapshot.Signal.IntervalType
	if interval == "" {
		interval = snapshot.OrderFlow.IntervalType
	}
	if interval == "" && len(snapshot.Klines) > 0 {
		interval = snapshot.Klines[len(snapshot.Klines)-1].IntervalType
	}

	openTime := snapshot.Signal.OpenTime
	if openTime == 0 {
		openTime = snapshot.OrderFlow.OpenTime
	}
	if openTime == 0 && len(snapshot.Klines) > 0 {
		openTime = snapshot.Klines[len(snapshot.Klines)-1].OpenTime
	}

	if symbol == "" || interval == "" || openTime == 0 {
		return models.FeatureSnapshot{}, errors.New("feature snapshot missing symbol, interval, or open time")
	}

	price := snapshot.Price.Price
	if price <= 0 && len(snapshot.Klines) > 0 {
		price = snapshot.Klines[len(snapshot.Klines)-1].ClosePrice
	}

	payload, err := json.Marshal(snapshot)
	if err != nil {
		return models.FeatureSnapshot{}, err
	}

	return models.FeatureSnapshot{
		Symbol:           symbol,
		IntervalType:     interval,
		OpenTime:         openTime,
		SnapshotSource:   "market_snapshot",
		FeatureVersion:   featureSnapshotVersion,
		Price:            price,
		SignalAction:     snapshot.Signal.Action,
		SignalScore:      snapshot.Signal.Score,
		SignalConfidence: snapshot.Signal.Confidence,
		SnapshotJSON:     string(payload),
	}, nil
}
