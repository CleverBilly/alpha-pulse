package service

import (
	"context"
	"time"

	"alpha-pulse/backend/internal/observability"
)

type cacheScope string

const (
	cacheScopeMarketSnapshot  cacheScope = "market-snapshot"
	cacheScopeIndicatorSeries cacheScope = "indicator-series"
	cacheScopeLiquiditySeries cacheScope = "liquidity-series"
	cacheScopeSignalTimeline  cacheScope = "signal-timeline"
)

type SymbolCacheInvalidator struct {
	cache MarketSnapshotCache
}

func NewSymbolCacheInvalidator(cache MarketSnapshotCache) *SymbolCacheInvalidator {
	if cache == nil {
		return nil
	}
	return &SymbolCacheInvalidator{cache: cache}
}

func (i *SymbolCacheInvalidator) InvalidateSymbol(symbol string) {
	if i == nil || i.cache == nil {
		return
	}
	invalidateAllSymbolCacheScopes(i.cache, symbol, allCacheScopes()...)
}

func invalidateCacheScopes(cache MarketSnapshotCache, symbol, interval string, scopes ...cacheScope) {
	if cache == nil || len(scopes) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	seen := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		prefix := cachePrefixForScope(scope, symbol, interval)
		if prefix == "" {
			continue
		}
		if _, exists := seen[prefix]; exists {
			continue
		}
		seen[prefix] = struct{}{}

		startedAt := time.Now()
		if err := cache.DeletePrefix(ctx, prefix); err != nil {
			observability.LogDuration(
				"cache",
				"invalidate",
				startedAt,
				"error",
				err.Error(),
				observability.String("scope", string(scope)),
				observability.String("symbol", symbol),
				observability.String("interval", interval),
				observability.String("prefix", prefix),
			)
			continue
		}
		observability.LogDuration(
			"cache",
			"invalidate",
			startedAt,
			"ok",
			"",
			observability.String("scope", string(scope)),
			observability.String("symbol", symbol),
			observability.String("interval", interval),
			observability.String("prefix", prefix),
		)
	}
}

func invalidateAllSymbolCacheScopes(cache MarketSnapshotCache, symbol string, scopes ...cacheScope) {
	if cache == nil || len(scopes) == 0 {
		return
	}

	symbol = normalizeSymbol(symbol)
	for _, interval := range supportedCacheIntervals {
		invalidateCacheScopes(cache, symbol, interval, scopes...)
	}
}

func allCacheScopes() []cacheScope {
	return []cacheScope{
		cacheScopeMarketSnapshot,
		cacheScopeIndicatorSeries,
		cacheScopeLiquiditySeries,
		cacheScopeSignalTimeline,
	}
}

func cachePrefixForScope(scope cacheScope, symbol, interval string) string {
	switch scope {
	case cacheScopeMarketSnapshot:
		return marketSnapshotCachePrefix(symbol, interval)
	case cacheScopeIndicatorSeries:
		return indicatorSeriesCachePrefix(symbol, interval)
	case cacheScopeLiquiditySeries:
		return liquiditySeriesCachePrefix(symbol, interval)
	case cacheScopeSignalTimeline:
		return signalTimelineCachePrefix(symbol, interval)
	default:
		return ""
	}
}
