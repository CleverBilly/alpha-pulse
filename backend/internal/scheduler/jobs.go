package scheduler

import (
	"context"
	"log"
	"strings"
	"time"

	"alpha-pulse/backend/internal/collector"
	"alpha-pulse/backend/internal/service"
)

// MarketServiceWarmup 描述 Scheduler 依赖的 MarketService 能力子集，便于测试 stub。
type MarketServiceWarmup interface {
	WarmupSymbol(symbol string) error
}

// Jobs 管理后台定时任务。
type Jobs struct {
	marketService  MarketServiceWarmup
	signalService  *service.SignalService
	alertService   *service.AlertService
	outcomeTracker *service.OutcomeTrackerService
	symbols        []string
	interval       time.Duration
}

// NewJobs 创建任务调度器。
func NewJobs(
	marketService MarketServiceWarmup,
	signalService *service.SignalService,
	alertService *service.AlertService,
	outcomeTracker *service.OutcomeTrackerService,
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
		marketService:  marketService,
		signalService:  signalService,
		alertService:   alertService,
		outcomeTracker: outcomeTracker,
		symbols:        normalizedSymbols,
		interval:       interval,
	}
}

// StartWithKlineEvents 启动事件驱动模式：
//   - klineEvents 有新收盘事件时立即触发分析
//   - 每 j.interval 兜底触发一次全量刷新
//   - klineEvents 为 nil 时退化为纯轮询模式（等价于原 Start）
func (j *Jobs) StartWithKlineEvents(ctx context.Context, klineEvents <-chan collector.KlineClosedEvent) {
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
		case event := <-klineEvents: // nil channel 永不触发此 case
			j.handleKlineEvent(event)
		}
	}
}

// Deprecated: 使用 StartWithKlineEvents 替代。Start 等价于 StartWithKlineEvents(ctx, nil)。
func (j *Jobs) Start(ctx context.Context) {
	j.StartWithKlineEvents(ctx, nil)
}

// handleKlineEvent 在 K线收盘时触发对应 symbol/interval 的快速分析链。
func (j *Jobs) handleKlineEvent(event collector.KlineClosedEvent) {
	if err := j.marketService.WarmupSymbol(event.Symbol); err != nil {
		log.Printf("event-driven warmup failed for %s/%s: %v", event.Symbol, event.Interval, err)
		return
	}
	if j.signalService != nil {
		if _, err := j.signalService.GetSignal(event.Symbol, event.Interval); err != nil {
			log.Printf("event-driven signal failed for %s/%s: %v", event.Symbol, event.Interval, err)
		}
	}
	if j.alertService != nil {
		if _, err := j.alertService.EvaluateAll(context.Background(), true); err != nil {
			log.Printf("event-driven alert eval failed: %v", err)
		}
	}
}

func (j *Jobs) runOnce() {
	for _, symbol := range j.symbols {
		if err := j.marketService.WarmupSymbol(symbol); err != nil {
			log.Printf("warmup failed for %s: %v", symbol, err)
			continue
		}
		if j.signalService != nil {
			if _, err := j.signalService.GetSignal(symbol, "1m"); err != nil {
				log.Printf("signal generation failed for %s: %v", symbol, err)
			}
		}
	}

	if j.alertService != nil {
		if _, err := j.alertService.EvaluateAll(context.Background(), true); err != nil {
			log.Printf("alert evaluation failed: %v", err)
		}
	}

	if j.outcomeTracker != nil {
		j.outcomeTracker.TrackAll(context.Background())
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
