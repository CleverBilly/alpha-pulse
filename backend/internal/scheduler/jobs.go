package scheduler

import (
	"context"
	"log"
	"strings"
	"time"

	"alpha-pulse/backend/internal/service"
)

// Jobs 管理后台定时任务。
type Jobs struct {
	marketService *service.MarketService
	signalService *service.SignalService
	alertService  *service.AlertService
	symbols       []string
	interval      time.Duration
}

// NewJobs 创建任务调度器。
func NewJobs(
	marketService *service.MarketService,
	signalService *service.SignalService,
	alertService *service.AlertService,
	symbols []string,
	interval time.Duration,
) *Jobs {
	normalizedSymbols := normalizeSymbols(symbols)
	if len(normalizedSymbols) == 0 {
		normalizedSymbols = []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}
	}
	if interval <= 0 {
		interval = 1 * time.Minute
	}

	return &Jobs{
		marketService: marketService,
		signalService: signalService,
		alertService:  alertService,
		symbols:       normalizedSymbols,
		interval:      interval,
	}
}

// Start 启动定时任务，按配置间隔刷新核心分析数据。
func (j *Jobs) Start(ctx context.Context) {
	j.runOnce()

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler stopped")
			return
		case <-ticker.C:
			j.runOnce()
		}
	}
}

func (j *Jobs) runOnce() {
	for _, symbol := range j.symbols {
		if err := j.marketService.WarmupSymbol(symbol); err != nil {
			log.Printf("warmup failed for %s: %v", symbol, err)
			continue
		}
		if _, err := j.signalService.GetSignal(symbol, "1m"); err != nil {
			log.Printf("signal generation failed for %s: %v", symbol, err)
		}
	}

	if j.alertService != nil {
		if _, err := j.alertService.EvaluateAll(context.Background(), true); err != nil {
			log.Printf("alert evaluation failed: %v", err)
		}
	}
}

func normalizeSymbols(symbols []string) []string {
	if len(symbols) == 0 {
		return nil
	}

	result := make([]string, 0, len(symbols))
	seen := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	return result
}
