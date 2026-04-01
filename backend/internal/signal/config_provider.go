package signal

import (
	"strconv"
	"sync"

	"alpha-pulse/backend/models"
)

// ConfigProvider 为信号引擎提供可热更新的配置值。
type ConfigProvider interface {
	// GetInt 返回整型配置值；若未找到则返回 defaultVal。
	GetInt(symbol, interval, key string, defaultVal int) int
}

// StaticConfigProvider 使用硬编码默认值（无数据库依赖）。
type StaticConfigProvider struct{}

// GetInt 始终返回 defaultVal，用于不需要动态配置的场景。
func (p *StaticConfigProvider) GetInt(_, _, _ string, defaultVal int) int { return defaultVal }

// DBConfigProvider 从 signal_configs 表读取配置，支持内存热更新。
type DBConfigProvider struct {
	mu      sync.RWMutex
	configs map[string]string // key: "symbol/interval/key" → value
}

// NewDBConfigProvider 从已加载的配置列表初始化 DBConfigProvider。
func NewDBConfigProvider(configs []models.SignalConfig) *DBConfigProvider {
	provider := &DBConfigProvider{
		configs: make(map[string]string, len(configs)),
	}
	for _, cfg := range configs {
		provider.configs[configKey(cfg.Symbol, cfg.Interval, cfg.Key)] = cfg.Value
	}
	return provider
}

// GetInt 查找配置值并转为 int；未找到则尝试通配（symbol="*", interval="*"），
// 最终未命中时返回 defaultVal。
func (p *DBConfigProvider) GetInt(symbol, interval, key string, defaultVal int) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	rawValue, found := p.configs[configKey(symbol, interval, key)]
	if !found {
		// 尝试通配：symbol="*" 且 interval="*"
		rawValue, found = p.configs[configKey("*", "*", key)]
	}
	if !found {
		return defaultVal
	}

	parsedValue, err := strconv.Atoi(rawValue)
	if err != nil {
		return defaultVal
	}
	return parsedValue
}

// Update 热更新单条配置（无需重启服务）。
func (p *DBConfigProvider) Update(symbol, interval, key, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.configs[configKey(symbol, interval, key)] = value
}

func configKey(symbol, interval, key string) string {
	return symbol + "/" + interval + "/" + key
}
