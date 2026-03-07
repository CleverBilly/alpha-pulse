package service

import (
	"time"

	"alpha-pulse/backend/internal/observability"
)

func logServiceDuration(
	component, stage, symbol, interval string,
	limit int,
	startedAt time.Time,
	status, reason string,
	fields ...observability.Field,
) {
	base := []observability.Field{}
	if symbol != "" {
		base = append(base, observability.String("symbol", symbol))
	}
	if interval != "" {
		base = append(base, observability.String("interval", interval))
	}
	if limit > 0 {
		base = append(base, observability.Int("limit", limit))
	}
	base = append(base, fields...)
	observability.LogDuration(component, stage, startedAt, status, reason, base...)
}
