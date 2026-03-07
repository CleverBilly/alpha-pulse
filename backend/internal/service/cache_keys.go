package service

import (
	"fmt"
	"strconv"
)

var supportedCacheIntervals = []string{"1m", "5m", "15m", "1h", "4h"}

func marketSnapshotCacheKey(symbol, interval string, limit int) string {
	return fmt.Sprintf("%s%d", marketSnapshotCachePrefix(symbol, interval), limit)
}

func marketSnapshotCachePrefix(symbol, interval string) string {
	return fmt.Sprintf("alpha-pulse:market-snapshot:v3:%s:%s:", symbol, interval)
}

func indicatorSeriesCacheKey(symbol, interval string, limit int) string {
	return indicatorSeriesCachePrefix(symbol, interval) + strconv.Itoa(limit)
}

func indicatorSeriesCachePrefix(symbol, interval string) string {
	return "alpha-pulse:indicator-series:v1:" + symbol + ":" + interval + ":"
}

func liquiditySeriesCacheKey(symbol, interval string, limit int) string {
	return liquiditySeriesCachePrefix(symbol, interval) + strconv.Itoa(limit)
}

func liquiditySeriesCachePrefix(symbol, interval string) string {
	return "alpha-pulse:liquidity-series:v1:" + symbol + ":" + interval + ":"
}

func signalTimelineCacheKey(symbol, interval string, limit int) string {
	return signalTimelineCachePrefix(symbol, interval) + strconv.Itoa(limit)
}

func signalTimelineCachePrefix(symbol, interval string) string {
	return fmt.Sprintf("alpha-pulse:signal-timeline:v1:%s:%s:", symbol, interval)
}
