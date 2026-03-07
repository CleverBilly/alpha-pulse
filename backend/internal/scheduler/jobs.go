package scheduler

import (
	"context"
	"log"
	"time"

	"alpha-pulse/backend/internal/service"
)

// Jobs 管理后台定时任务。
type Jobs struct {
	marketService *service.MarketService
	signalService *service.SignalService
}

// NewJobs 创建任务调度器。
func NewJobs(marketService *service.MarketService, signalService *service.SignalService) *Jobs {
	return &Jobs{marketService: marketService, signalService: signalService}
}

// Start 启动定时任务，每分钟刷新一次核心分析数据。
func (j *Jobs) Start(ctx context.Context) {
	j.runOnce()

	ticker := time.NewTicker(1 * time.Minute)
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
	symbols := []string{"BTCUSDT", "ETHUSDT"}
	for _, symbol := range symbols {
		if err := j.marketService.WarmupSymbol(symbol); err != nil {
			log.Printf("warmup failed for %s: %v", symbol, err)
			continue
		}
		if _, err := j.signalService.GetSignal(symbol, "1m"); err != nil {
			log.Printf("signal generation failed for %s: %v", symbol, err)
		}
	}
}
