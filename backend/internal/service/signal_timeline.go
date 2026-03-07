package service

import (
	"context"
	"sort"
	"time"

	"alpha-pulse/backend/internal/observability"
	"alpha-pulse/backend/models"
	"gorm.io/gorm"
)

// GetSignalTimeline 获取指定交易对与周期的历史信号时间线。
func (s *SignalService) GetSignalTimeline(symbol, interval string, limit int) (SignalTimelineResult, error) {
	return s.GetSignalTimelineWithRefresh(symbol, interval, limit, false)
}

// GetSignalTimelineWithRefresh 获取指定交易对与周期的历史信号时间线，并可显式绕过缓存。
func (s *SignalService) GetSignalTimelineWithRefresh(symbol, interval string, limit int, refresh bool) (SignalTimelineResult, error) {
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
	limit = clampInt(limit, 1, 120)

	cacheStartedAt := time.Now()
	if refresh {
		invalidateAllSymbolCacheScopes(s.viewCache, symbol, allCacheScopes()...)
		invalidateAllSymbolCacheScopes(s.snapshotCache, symbol, allCacheScopes()...)
		logServiceDuration("signal_service", "signal_timeline.cache_read", symbol, interval, limit, cacheStartedAt, "refresh", "", observability.Bool("refresh", true))
	} else {
		if cached, ok, err := s.getCachedSignalTimeline(symbol, interval, limit); err == nil && ok {
			logServiceDuration("signal_service", "signal_timeline.cache_read", symbol, interval, limit, cacheStartedAt, "hit", "", observability.String("source", "cache"))
			return cached, nil
		} else if err != nil {
			logServiceDuration("signal_service", "signal_timeline.cache_read", symbol, interval, limit, cacheStartedAt, "error", err.Error(), observability.String("source", "cache"))
		} else {
			logServiceDuration("signal_service", "signal_timeline.cache_read", symbol, interval, limit, cacheStartedAt, "miss", "", observability.String("source", "cache"))
		}
	}

	buildStartedAt := time.Now()
	points, err := s.loadSignalTimeline(symbol, interval, limit)
	if err != nil {
		logServiceDuration("signal_service", "signal_timeline.build", symbol, interval, limit, buildStartedAt, "error", err.Error())
		return SignalTimelineResult{}, err
	}
	if len(points) == 0 {
		if _, err := s.GetSignal(symbol, interval); err != nil {
			logServiceDuration("signal_service", "signal_timeline.build", symbol, interval, limit, buildStartedAt, "error", err.Error())
			return SignalTimelineResult{}, err
		}
		points, err = s.loadSignalTimeline(symbol, interval, limit)
		if err != nil {
			logServiceDuration("signal_service", "signal_timeline.build", symbol, interval, limit, buildStartedAt, "error", err.Error())
			return SignalTimelineResult{}, err
		}
	}

	result := SignalTimelineResult{
		Symbol:   symbol,
		Interval: interval,
		Points:   points,
	}
	if err := s.setCachedSignalTimeline(symbol, interval, limit, result); err != nil {
		logServiceDuration("signal_service", "signal_timeline.cache_write", symbol, interval, limit, time.Now(), "error", err.Error())
	}
	logServiceDuration("signal_service", "signal_timeline.build", symbol, interval, limit, buildStartedAt, "ok", "", observability.Int("points", len(points)))
	return result, nil
}

func (s *SignalService) loadSignalTimeline(symbol, interval string, limit int) ([]models.SignalTimelinePoint, error) {
	if s.signalRepo == nil {
		return nil, nil
	}

	fetchLimit := maxInt(limit*4, 20)
	signals, err := s.signalRepo.GetRecentByInterval(symbol, interval, fetchLimit)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return compactSignalTimeline(signals, limit), nil
}

func compactSignalTimeline(signals []models.Signal, limit int) []models.SignalTimelinePoint {
	if len(signals) == 0 || limit <= 0 {
		return nil
	}

	deduped := make([]models.SignalTimelinePoint, 0, minInt(len(signals), limit))
	seen := make(map[int64]struct{}, len(signals))

	for _, signal := range signals {
		openTime := signal.OpenTime
		if openTime <= 0 && !signal.CreatedAt.IsZero() {
			openTime = signal.CreatedAt.UnixMilli()
		}
		if openTime <= 0 {
			continue
		}
		if _, exists := seen[openTime]; exists {
			continue
		}
		seen[openTime] = struct{}{}

		deduped = append(deduped, models.SignalTimelinePoint{
			ID:           signal.ID,
			Symbol:       signal.Symbol,
			IntervalType: signal.IntervalType,
			OpenTime:     openTime,
			Signal:       signal.Action,
			Score:        signal.Score,
			Confidence:   signal.Confidence,
			EntryPrice:   signal.EntryPrice,
			StopLoss:     signal.StopLoss,
			TargetPrice:  signal.TargetPrice,
		})

		if len(deduped) >= limit {
			break
		}
	}

	sort.Slice(deduped, func(i, j int) bool {
		if deduped[i].OpenTime == deduped[j].OpenTime {
			return deduped[i].ID < deduped[j].ID
		}
		return deduped[i].OpenTime < deduped[j].OpenTime
	})

	return deduped
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func (s *SignalService) getCachedSignalTimeline(symbol, interval string, limit int) (SignalTimelineResult, bool, error) {
	if s.viewCache == nil || s.viewCacheTTL <= 0 {
		return SignalTimelineResult{}, false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()
	return getCachedJSON[SignalTimelineResult](ctx, s.viewCache, signalTimelineCacheKey(symbol, interval, limit))
}

func (s *SignalService) setCachedSignalTimeline(symbol, interval string, limit int, result SignalTimelineResult) error {
	if s.viewCache == nil || s.viewCacheTTL <= 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()
	return setCachedJSON(ctx, s.viewCache, signalTimelineCacheKey(symbol, interval, limit), result, s.viewCacheTTL)
}
