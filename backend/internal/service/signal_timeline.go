package service

import (
	"context"
	"log"
	"sort"
	"time"

	"alpha-pulse/backend/models"
	"gorm.io/gorm"
)

// GetSignalTimeline 获取指定交易对与周期的历史信号时间线。
func (s *SignalService) GetSignalTimeline(symbol, interval string, limit int) (SignalTimelineResult, error) {
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
	limit = clampInt(limit, 1, 120)

	if cached, ok, err := s.getCachedSignalTimeline(symbol, interval, limit); err == nil && ok {
		log.Printf("signal-timeline cache hit symbol=%s interval=%s limit=%d", symbol, interval, limit)
		return cached, nil
	} else if err != nil {
		log.Printf("signal-timeline cache read failed symbol=%s interval=%s limit=%d err=%v", symbol, interval, limit, err)
	}

	points, err := s.loadSignalTimeline(symbol, interval, limit)
	if err != nil {
		return SignalTimelineResult{}, err
	}
	if len(points) == 0 {
		if _, err := s.GetSignal(symbol, interval); err != nil {
			return SignalTimelineResult{}, err
		}
		points, err = s.loadSignalTimeline(symbol, interval, limit)
		if err != nil {
			return SignalTimelineResult{}, err
		}
	}

	result := SignalTimelineResult{
		Symbol:   symbol,
		Interval: interval,
		Points:   points,
	}
	if err := s.setCachedSignalTimeline(symbol, interval, limit, result); err != nil {
		log.Printf("signal-timeline cache write failed symbol=%s interval=%s limit=%d err=%v", symbol, interval, limit, err)
	}
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
