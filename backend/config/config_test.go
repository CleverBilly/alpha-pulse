package config

import (
	"reflect"
	"testing"
)

func TestLoadDefaultsToDevMode(t *testing.T) {
	clearConfigEnv(t)

	cfg := Load()

	if cfg.AppMode != ModeDev {
		t.Fatalf("expected default mode %s, got %s", ModeDev, cfg.AppMode)
	}
	if cfg.GinMode != "debug" {
		t.Fatalf("expected debug gin mode, got %s", cfg.GinMode)
	}
	if !cfg.EnableAutoMigrate {
		t.Fatal("expected auto migrate to be enabled in dev mode")
	}
	if !cfg.EnableRedisCache {
		t.Fatal("expected redis cache to be enabled in dev mode")
	}
	if !cfg.EnableStreamCollector {
		t.Fatal("expected stream collector to be enabled in dev mode")
	}
	if !cfg.EnableScheduler {
		t.Fatal("expected scheduler to be enabled in dev mode")
	}
	if !cfg.AllowMockBinanceData {
		t.Fatal("expected mock binance data to be enabled in dev mode")
	}
	if cfg.SchedulerIntervalSeconds != 60 {
		t.Fatalf("expected default scheduler interval 60s, got %d", cfg.SchedulerIntervalSeconds)
	}
	if !reflect.DeepEqual(cfg.MarketSymbols, []string{"BTCUSDT", "ETHUSDT"}) {
		t.Fatalf("unexpected default market symbols: %#v", cfg.MarketSymbols)
	}
}

func TestLoadProdModeDefaults(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("APP_MODE", "prod")

	cfg := Load()

	if cfg.AppMode != ModeProd {
		t.Fatalf("expected mode %s, got %s", ModeProd, cfg.AppMode)
	}
	if cfg.GinMode != "release" {
		t.Fatalf("expected release gin mode, got %s", cfg.GinMode)
	}
	if cfg.EnableAutoMigrate {
		t.Fatal("expected auto migrate to be disabled in prod mode")
	}
	if cfg.AllowMockBinanceData {
		t.Fatal("expected mock binance data to be disabled in prod mode")
	}
	if !cfg.EnableStreamCollector {
		t.Fatal("expected stream collector to remain enabled in prod mode")
	}
	if !cfg.EnableScheduler {
		t.Fatal("expected scheduler to remain enabled in prod mode")
	}
}

func TestLoadHonorsModeOverrides(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("APP_ENV", "test")
	t.Setenv("GIN_MODE", "debug")
	t.Setenv("AUTO_MIGRATE", "true")
	t.Setenv("ENABLE_REDIS_CACHE", "true")
	t.Setenv("ENABLE_STREAM_COLLECTOR", "yes")
	t.Setenv("ENABLE_SCHEDULER", "1")
	t.Setenv("ALLOW_MOCK_BINANCE_DATA", "false")
	t.Setenv("MARKET_SYMBOLS", " btcusdt , solusdt,ETHUSDT,solusdt ")
	t.Setenv("SCHEDULER_INTERVAL_SECONDS", "15")

	cfg := Load()

	if cfg.AppMode != ModeTest {
		t.Fatalf("expected mode %s, got %s", ModeTest, cfg.AppMode)
	}
	if cfg.GinMode != "debug" {
		t.Fatalf("expected overridden gin mode debug, got %s", cfg.GinMode)
	}
	if !cfg.EnableAutoMigrate || !cfg.EnableRedisCache || !cfg.EnableStreamCollector || !cfg.EnableScheduler {
		t.Fatalf("expected boolean overrides to be honored: %+v", cfg)
	}
	if cfg.AllowMockBinanceData {
		t.Fatal("expected allow mock binance override to disable mock data")
	}
	if cfg.SchedulerIntervalSeconds != 15 {
		t.Fatalf("expected scheduler interval override 15, got %d", cfg.SchedulerIntervalSeconds)
	}

	expectedSymbols := []string{"BTCUSDT", "SOLUSDT", "ETHUSDT"}
	if !reflect.DeepEqual(cfg.MarketSymbols, expectedSymbols) {
		t.Fatalf("unexpected market symbols: got=%#v want=%#v", cfg.MarketSymbols, expectedSymbols)
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"APP_MODE",
		"APP_ENV",
		"GIN_MODE",
		"APP_PORT",
		"MYSQL_DSN",
		"REDIS_ADDR",
		"REDIS_PASSWORD",
		"REDIS_DB",
		"MARKET_SNAPSHOT_CACHE_TTL",
		"ANALYSIS_VIEW_CACHE_TTL",
		"BINANCE_API_KEY",
		"BINANCE_SECRET_KEY",
		"MARKET_SYMBOLS",
		"AUTO_MIGRATE",
		"ENABLE_REDIS_CACHE",
		"ENABLE_STREAM_COLLECTOR",
		"ENABLE_SCHEDULER",
		"ALLOW_MOCK_BINANCE_DATA",
		"SCHEDULER_INTERVAL_SECONDS",
	}
	for _, key := range keys {
		t.Setenv(key, "")
	}
}
