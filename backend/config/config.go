package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	ModeDev  = "dev"
	ModeTest = "test"
	ModeProd = "prod"
)

// Config 定义后端运行所需配置。
type Config struct {
	AppMode          string
	GinMode          string
	AppPort          string
	MySQLDSN         string
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
	SnapshotCacheTTL int
	AnalysisCacheTTL int
	BinanceAPIKey    string
	BinanceSecretKey string
	MarketSymbols    []string

	EnableAutoMigrate        bool
	EnableRedisCache         bool
	EnableStreamCollector    bool
	EnableScheduler          bool
	AllowMockBinanceData     bool
	SchedulerIntervalSeconds int
}

// Load 从环境变量加载配置。
func Load() Config {
	loadDotEnvFiles()
	mode := normalizeMode(firstNonEmptyEnv("APP_MODE", "APP_ENV"))
	defaults := defaultsForMode(mode)
	ginMode := normalizeGinMode(getEnv("GIN_MODE", defaults.ginMode), defaults.ginMode)
	schedulerInterval := getEnvAsInt("SCHEDULER_INTERVAL_SECONDS", 60)
	if schedulerInterval <= 0 {
		schedulerInterval = 60
	}

	return Config{
		AppMode:          mode,
		GinMode:          ginMode,
		AppPort:          getEnv("APP_PORT", "8080"),
		MySQLDSN:         getEnv("MYSQL_DSN", "root:root@tcp(localhost:3306)/alpha_pulse?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:          getEnvAsInt("REDIS_DB", 0),
		SnapshotCacheTTL: getEnvAsInt("MARKET_SNAPSHOT_CACHE_TTL", 5),
		AnalysisCacheTTL: getEnvAsInt("ANALYSIS_VIEW_CACHE_TTL", 15),
		BinanceAPIKey:    getEnv("BINANCE_API_KEY", ""),
		BinanceSecretKey: getEnv("BINANCE_SECRET_KEY", ""),
		MarketSymbols:    getEnvAsCSV("MARKET_SYMBOLS", []string{"BTCUSDT", "ETHUSDT"}),

		EnableAutoMigrate:        getEnvAsBool("AUTO_MIGRATE", defaults.autoMigrate),
		EnableRedisCache:         getEnvAsBool("ENABLE_REDIS_CACHE", defaults.enableRedisCache),
		EnableStreamCollector:    getEnvAsBool("ENABLE_STREAM_COLLECTOR", defaults.enableStreamCollector),
		EnableScheduler:          getEnvAsBool("ENABLE_SCHEDULER", defaults.enableScheduler),
		AllowMockBinanceData:     getEnvAsBool("ALLOW_MOCK_BINANCE_DATA", defaults.allowMockBinanceData),
		SchedulerIntervalSeconds: schedulerInterval,
	}
}

func loadDotEnvFiles() {
	candidates := []string{
		".env.local",
		".env",
		"backend/.env.local",
		"backend/.env",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			values, readErr := godotenv.Read(path)
			if readErr != nil {
				continue
			}
			for key, value := range values {
				if strings.TrimSpace(os.Getenv(key)) == "" {
					_ = os.Setenv(key, value)
				}
			}
		}
	}
}

type modeDefaults struct {
	ginMode               string
	autoMigrate           bool
	enableRedisCache      bool
	enableStreamCollector bool
	enableScheduler       bool
	allowMockBinanceData  bool
}

func defaultsForMode(mode string) modeDefaults {
	switch mode {
	case ModeProd:
		return modeDefaults{
			ginMode:               "release",
			autoMigrate:           false,
			enableRedisCache:      true,
			enableStreamCollector: true,
			enableScheduler:       true,
			allowMockBinanceData:  false,
		}
	case ModeTest:
		return modeDefaults{
			ginMode:               "test",
			autoMigrate:           false,
			enableRedisCache:      false,
			enableStreamCollector: false,
			enableScheduler:       false,
			allowMockBinanceData:  true,
		}
	default:
		return modeDefaults{
			ginMode:               "debug",
			autoMigrate:           true,
			enableRedisCache:      true,
			enableStreamCollector: true,
			enableScheduler:       true,
			allowMockBinanceData:  true,
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}

func getEnvAsBool(key string, defaultValue bool) bool {
	value := strings.TrimSpace(strings.ToLower(getEnv(key, "")))
	if value == "" {
		return defaultValue
	}

	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return defaultValue
	}
}

func getEnvAsCSV(key string, defaultValue []string) []string {
	value := getEnv(key, "")
	if strings.TrimSpace(value) == "" {
		return append([]string(nil), defaultValue...)
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		normalized := strings.ToUpper(strings.TrimSpace(part))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	if len(result) == 0 {
		return append([]string(nil), defaultValue...)
	}
	return result
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func normalizeMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ModeProd, "production":
		return ModeProd
	case ModeTest, "testing":
		return ModeTest
	default:
		return ModeDev
	}
}

func normalizeGinMode(value string, fallback string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return "debug"
	case "release":
		return "release"
	case "test":
		return "test"
	default:
		return fallback
	}
}
