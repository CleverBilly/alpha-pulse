package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

const defaultAccountCacheTTL = 30 * time.Second
const accountCacheHardExpiry = 60 * time.Second

// AccountStateCache 缓存账户余额、杠杆和交易规则，避免下单时同步调用 Binance API。
// 后台每隔 interval 自动刷新；读路径 <1ms，替代原来 ~500ms 的同步调用。
type AccountStateCache struct {
	mu          sync.RWMutex
	client      TradeClient
	symbols     []string
	ttl         time.Duration
	balance     float64
	leverage    map[string]int
	rules       map[string]FuturesSymbolRules
	refreshedAt time.Time
}

// NewAccountStateCache 创建账户状态缓存。symbols 列表用于按交易对拉取杠杆和规则。
func NewAccountStateCache(client TradeClient, symbols []string) *AccountStateCache {
	return &AccountStateCache{
		client:   client,
		symbols:  symbols,
		ttl:      defaultAccountCacheTTL,
		leverage: make(map[string]int),
		rules:    make(map[string]FuturesSymbolRules),
	}
}

// Refresh 从 Binance API 同步拉取账户余额、各 symbol 杠杆和交易规则，并原子写入缓存。
func (c *AccountStateCache) Refresh() error {
	balance, err := c.client.GetFuturesBalance()
	if err != nil {
		return fmt.Errorf("get balance: %w", err)
	}

	leverage := make(map[string]int, len(c.symbols))
	rules := make(map[string]FuturesSymbolRules, len(c.symbols))
	for _, symbol := range c.symbols {
		lev, levErr := c.client.GetFuturesLeverage(symbol)
		if levErr != nil {
			return fmt.Errorf("get leverage for %s: %w", symbol, levErr)
		}
		leverage[symbol] = lev

		r, ruleErr := c.client.GetFuturesSymbolRules(symbol)
		if ruleErr != nil {
			return fmt.Errorf("get rules for %s: %w", symbol, ruleErr)
		}
		rules[symbol] = r
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.balance = balance
	c.leverage = leverage
	c.rules = rules
	c.refreshedAt = time.Now()
	return nil
}

// IsStale 返回缓存是否已超过 TTL（未初始化也视为 stale）。
func (c *AccountStateCache) IsStale() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.refreshedAt.IsZero() || time.Since(c.refreshedAt) > c.ttl
}

// GetBalance 返回缓存余额。若缓存从未初始化或已超过硬过期阈值（60s），返回错误。
func (c *AccountStateCache) GetBalance() (float64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.refreshedAt.IsZero() {
		return 0, fmt.Errorf("account state cache not yet initialized")
	}
	if time.Since(c.refreshedAt) > accountCacheHardExpiry {
		return 0, fmt.Errorf("account state cache hard-expired (>60s)")
	}
	return c.balance, nil
}

// GetLeverage 返回指定 symbol 的缓存杠杆。symbol 未缓存、缓存未初始化或已超过硬过期阈值（60s）时返回错误。
func (c *AccountStateCache) GetLeverage(symbol string) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.refreshedAt.IsZero() {
		return 0, fmt.Errorf("account state cache not yet initialized")
	}
	if time.Since(c.refreshedAt) > accountCacheHardExpiry {
		return 0, fmt.Errorf("account state cache hard-expired (>60s)")
	}
	lev, ok := c.leverage[symbol]
	if !ok {
		return 0, fmt.Errorf("no leverage cached for symbol %s", symbol)
	}
	return lev, nil
}

// GetRules 返回指定 symbol 的缓存交易规则。symbol 未缓存、缓存未初始化或已超过硬过期阈值（60s）时返回错误。
func (c *AccountStateCache) GetRules(symbol string) (FuturesSymbolRules, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.refreshedAt.IsZero() {
		return FuturesSymbolRules{}, fmt.Errorf("account state cache not yet initialized")
	}
	if time.Since(c.refreshedAt) > accountCacheHardExpiry {
		return FuturesSymbolRules{}, fmt.Errorf("account state cache hard-expired (>60s)")
	}
	r, ok := c.rules[symbol]
	if !ok {
		return FuturesSymbolRules{}, fmt.Errorf("no rules cached for symbol %s", symbol)
	}
	return r, nil
}

// StartBackgroundRefresh 启动后台刷新 goroutine，每隔 interval 调用一次 Refresh。
// 会立即返回；调用方无需加 go。
func (c *AccountStateCache) StartBackgroundRefresh(ctx context.Context, interval time.Duration) {
	go func() {
		if err := c.Refresh(); err != nil {
			log.Printf("account state cache initial refresh failed: %v", err)
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.Refresh(); err != nil {
					log.Printf("account state cache refresh failed: %v", err)
				}
			}
		}
	}()
}
