package config

import (
	"os"
	"strconv"
)

// Config 定义后端运行所需配置。
type Config struct {
	AppPort          string
	MySQLDSN         string
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
	SnapshotCacheTTL int
	AnalysisCacheTTL int
	BinanceAPIKey    string
	BinanceSecretKey string
}

// Load 从环境变量加载配置。
func Load() Config {
	return Config{
		AppPort:          getEnv("APP_PORT", "8080"),
		MySQLDSN:         getEnv("MYSQL_DSN", "root:root@tcp(localhost:3306)/alpha_pulse?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:          getEnvAsInt("REDIS_DB", 0),
		SnapshotCacheTTL: getEnvAsInt("MARKET_SNAPSHOT_CACHE_TTL", 5),
		AnalysisCacheTTL: getEnvAsInt("ANALYSIS_VIEW_CACHE_TTL", 15),
		BinanceAPIKey:    getEnv("BINANCE_API_KEY", ""),
		BinanceSecretKey: getEnv("BINANCE_SECRET_KEY", ""),
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
