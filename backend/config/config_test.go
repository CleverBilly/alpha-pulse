package config

import (
	"os"
	"path/filepath"
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
	if cfg.SchedulerIntervalSeconds != 15 {
		t.Fatalf("expected default scheduler interval 15s, got %d", cfg.SchedulerIntervalSeconds)
	}
	if !reflect.DeepEqual(cfg.MarketSymbols, []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}) {
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

func TestLoadReadsDotEnvFile(t *testing.T) {
	clearConfigEnv(t)

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(tempDir, ".env"),
		[]byte("APP_MODE=prod\nMYSQL_DSN=test:test@tcp(127.0.0.1:3306)/alpha_pulse_dev?charset=utf8mb4&parseTime=True&loc=Local\nENABLE_REDIS_CACHE=false\n"),
		0o600,
	); err != nil {
		t.Fatalf("write temp env failed: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(workingDir)
	})

	cfg := Load()

	if cfg.AppMode != ModeProd {
		t.Fatalf("expected .env mode %s, got %s", ModeProd, cfg.AppMode)
	}
	if cfg.MySQLDSN != "test:test@tcp(127.0.0.1:3306)/alpha_pulse_dev?charset=utf8mb4&parseTime=True&loc=Local" {
		t.Fatalf("unexpected mysql dsn from .env: %s", cfg.MySQLDSN)
	}
	if cfg.EnableRedisCache {
		t.Fatal("expected .env override to disable redis cache")
	}
}

func TestLoadAuthDefaults(t *testing.T) {
	clearConfigEnv(t)

	cfg := Load()

	if cfg.EnableSingleUserAuth {
		t.Fatal("expected single user auth to be disabled by default")
	}
	if cfg.AuthCookieName != "alpha_pulse_session" {
		t.Fatalf("unexpected default auth cookie name: %s", cfg.AuthCookieName)
	}
	if cfg.AuthSessionTTLHours != 168 {
		t.Fatalf("expected default auth session ttl 168h, got %d", cfg.AuthSessionTTLHours)
	}
	if cfg.AlertHistoryLimit != 40 {
		t.Fatalf("expected default alert history limit 40, got %d", cfg.AlertHistoryLimit)
	}
	if cfg.AlertPublicBaseURL != "" {
		t.Fatalf("expected alert public base url to be empty by default, got %s", cfg.AlertPublicBaseURL)
	}
	if cfg.FeishuBotWebhookURL != "" || cfg.FeishuBotSecret != "" {
		t.Fatalf("expected feishu config to be empty by default: %+v", cfg)
	}

	expectedOrigins := []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	if !reflect.DeepEqual(cfg.CORSAllowOrigins, expectedOrigins) {
		t.Fatalf("unexpected default cors origins: got=%#v want=%#v", cfg.CORSAllowOrigins, expectedOrigins)
	}
}

func TestLoadAuthOverrides(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("APP_MODE", "prod")
	t.Setenv("ENABLE_SINGLE_USER_AUTH", "true")
	t.Setenv("AUTH_USERNAME", "alpha-admin")
	t.Setenv("AUTH_PASSWORD_HASH", "$2a$10$mockedhashvalue")
	t.Setenv("AUTH_SESSION_SECRET", "super-secret")
	t.Setenv("AUTH_SESSION_TTL_HOURS", "24")
	t.Setenv("AUTH_COOKIE_NAME", "alpha_session")
	t.Setenv("AUTH_COOKIE_DOMAIN", ".example.com")
	t.Setenv("AUTH_COOKIE_SECURE", "false")
	t.Setenv("CORS_ALLOW_ORIGINS", "https://app.example.com,https://alpha.example.com")
	t.Setenv("ALERT_HISTORY_LIMIT", "88")
	t.Setenv("ALERT_PUBLIC_BASE_URL", "https://alpha.example.com")
	t.Setenv("FEISHU_BOT_WEBHOOK_URL", "https://open.feishu.cn/open-apis/bot/v2/hook/mock")
	t.Setenv("FEISHU_BOT_SECRET", "bot-secret")

	cfg := Load()

	if !cfg.EnableSingleUserAuth {
		t.Fatal("expected single user auth override to be enabled")
	}
	if cfg.AuthUsername != "alpha-admin" {
		t.Fatalf("unexpected auth username: %s", cfg.AuthUsername)
	}
	if cfg.AuthPasswordHash != "$2a$10$mockedhashvalue" {
		t.Fatalf("unexpected auth password hash: %s", cfg.AuthPasswordHash)
	}
	if cfg.AuthSessionSecret != "super-secret" {
		t.Fatalf("unexpected auth session secret: %s", cfg.AuthSessionSecret)
	}
	if cfg.AuthSessionTTLHours != 24 {
		t.Fatalf("unexpected auth session ttl hours: %d", cfg.AuthSessionTTLHours)
	}
	if cfg.AuthCookieName != "alpha_session" {
		t.Fatalf("unexpected auth cookie name: %s", cfg.AuthCookieName)
	}
	if cfg.AuthCookieDomain != ".example.com" {
		t.Fatalf("unexpected auth cookie domain: %s", cfg.AuthCookieDomain)
	}
	if cfg.AuthCookieSecure {
		t.Fatal("expected auth cookie secure override to disable secure cookie")
	}
	if cfg.AlertHistoryLimit != 88 {
		t.Fatalf("unexpected alert history limit: %d", cfg.AlertHistoryLimit)
	}
	if cfg.AlertPublicBaseURL != "https://alpha.example.com" {
		t.Fatalf("unexpected alert public base url: %s", cfg.AlertPublicBaseURL)
	}
	if cfg.FeishuBotWebhookURL == "" || cfg.FeishuBotSecret != "bot-secret" {
		t.Fatalf("unexpected feishu config: webhook=%s secret=%s", cfg.FeishuBotWebhookURL, cfg.FeishuBotSecret)
	}

	expectedOrigins := []string{"https://app.example.com", "https://alpha.example.com"}
	if !reflect.DeepEqual(cfg.CORSAllowOrigins, expectedOrigins) {
		t.Fatalf("unexpected cors allow origins: got=%#v want=%#v", cfg.CORSAllowOrigins, expectedOrigins)
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
		"ENABLE_SINGLE_USER_AUTH",
		"AUTH_USERNAME",
		"AUTH_PASSWORD_HASH",
		"AUTH_SESSION_SECRET",
		"AUTH_SESSION_TTL_HOURS",
		"AUTH_COOKIE_NAME",
		"AUTH_COOKIE_DOMAIN",
		"AUTH_COOKIE_SECURE",
		"CORS_ALLOW_ORIGINS",
		"ALERT_HISTORY_LIMIT",
		"ALERT_PUBLIC_BASE_URL",
		"FEISHU_BOT_WEBHOOK_URL",
		"FEISHU_BOT_SECRET",
	}
	for _, key := range keys {
		t.Setenv(key, "")
	}
}
