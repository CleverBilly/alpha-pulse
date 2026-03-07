package service

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// MarketSnapshotCache 抽象 market-snapshot 缓存读写，便于运行时使用 Redis、测试时使用内存替身。
type MarketSnapshotCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
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
