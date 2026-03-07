package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// MarketSnapshotCache 抽象 market-snapshot 缓存读写，便于运行时使用 Redis、测试时使用内存替身。
type MarketSnapshotCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePrefix(ctx context.Context, prefix string) error
}

type redisMarketSnapshotCache struct {
	client redis.Cmdable
}

// NewRedisMarketSnapshotCache 创建 Redis 版 market-snapshot 缓存。
func NewRedisMarketSnapshotCache(client redis.Cmdable) MarketSnapshotCache {
	if client == nil {
		return nil
	}
	return &redisMarketSnapshotCache{client: client}
}

func (c *redisMarketSnapshotCache) Get(ctx context.Context, key string) ([]byte, error) {
	value, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return value, nil
}

func (c *redisMarketSnapshotCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *redisMarketSnapshotCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *redisMarketSnapshotCache) DeletePrefix(ctx context.Context, prefix string) error {
	if strings.TrimSpace(prefix) == "" {
		return nil
	}

	cursor := uint64(0)
	pattern := prefix + "*"
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

func getCachedJSON[T any](ctx context.Context, cache MarketSnapshotCache, key string) (T, bool, error) {
	var zero T
	if cache == nil {
		return zero, false, nil
	}

	payload, err := cache.Get(ctx, key)
	if err != nil || len(payload) == 0 {
		return zero, false, err
	}

	var value T
	if err := json.Unmarshal(payload, &value); err != nil {
		_ = cache.Delete(ctx, key)
		return zero, false, err
	}

	return value, true, nil
}

func setCachedJSON[T any](ctx context.Context, cache MarketSnapshotCache, key string, value T, ttl time.Duration) error {
	if cache == nil || ttl <= 0 {
		return nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return cache.Set(ctx, key, payload, ttl)
}
