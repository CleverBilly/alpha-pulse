package service

import (
	"context"
	"log"
	"time"

	"alpha-pulse/backend/models"
	"alpha-pulse/backend/repository"
)

// outcomeExpiryMs 是 pending 信号的最大观察窗口：60 分钟。
const outcomeExpiryMs = 60 * 60 * 1000

// OutcomeTrackerService 定期结算 alert_records 中 pending 信号的结果。
type OutcomeTrackerService struct {
	alertRecordRepo *repository.AlertRecordRepository
	klineRepo       *repository.KlineRepository
	symbols         []string
}

// NewOutcomeTrackerService 创建 OutcomeTrackerService。
func NewOutcomeTrackerService(
	alertRecordRepo *repository.AlertRecordRepository,
	klineRepo *repository.KlineRepository,
	symbols []string,
) *OutcomeTrackerService {
	return &OutcomeTrackerService{
		alertRecordRepo: alertRecordRepo,
		klineRepo:       klineRepo,
		symbols:         normalizeAlertSymbols(symbols),
	}
}

// TrackAll 遍历所有标的，结算 pending 信号结果。
func (t *OutcomeTrackerService) TrackAll(ctx context.Context) {
	for _, symbol := range t.symbols {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if err := t.trackSymbol(ctx, symbol); err != nil {
			log.Printf("outcome tracker failed for %s: %v", symbol, err)
		}
	}
}

func (t *OutcomeTrackerService) trackSymbol(ctx context.Context, symbol string) error {
	records, err := t.alertRecordRepo.FindPending(symbol, "setup_ready", 100)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, record := range records {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		interval := record.Interval
		if interval == "" {
			interval = "1h"
		}

		klines, err := t.klineRepo.FindAfter(symbol, interval, record.EventTime, 60)
		if err != nil {
			log.Printf("kline fetch failed for outcome tracking %s %s: %v", symbol, interval, err)
			continue
		}

		outcome, outcomePrice, outcomeAt := evalOutcome(record, klines, now)
		if outcome == "" {
			continue // 仍在观察窗口内，或方向不可追踪
		}

		actualRR := 0.0
		switch outcome {
		case "target_hit":
			if record.RiskReward > 0 {
				actualRR = record.RiskReward
			}
		case "stop_hit":
			actualRR = -1.0
		}

		if err := t.alertRecordRepo.UpdateOutcome(record.ID, outcome, outcomePrice, outcomeAt, actualRR); err != nil {
			log.Printf("outcome update failed for record %d: %v", record.ID, err)
		}
	}
	return nil
}

// evalOutcome 根据后续 K 线判断信号结果。
// 返回 ("", 0, 0) 表示：方向不可追踪 或 仍在观察窗口内尚未结算。
// 止损检查优先于止盈检查，避免同一根 K 线双触时误判为盈利。
func evalOutcome(record models.AlertRecord, klines []models.Kline, now time.Time) (outcome string, outcomePrice float64, outcomeAt int64) {
	isLong := record.DirectionState == "strong-bullish" || record.DirectionState == "bullish"
	isShort := record.DirectionState == "strong-bearish" || record.DirectionState == "bearish"
	if !isLong && !isShort {
		return "", 0, 0 // 中性方向不追踪
	}

	// 逐根 K 线扫描；止损检查先于止盈，保证止损优先语义。
	for _, k := range klines {
		if isLong {
			if k.LowPrice <= record.StopLoss {
				return "stop_hit", record.StopLoss, k.OpenTime
			}
			if k.HighPrice >= record.TargetPrice {
				return "target_hit", record.TargetPrice, k.OpenTime
			}
		} else { // isShort
			if k.HighPrice >= record.StopLoss {
				return "stop_hit", record.StopLoss, k.OpenTime
			}
			if k.LowPrice <= record.TargetPrice {
				return "target_hit", record.TargetPrice, k.OpenTime
			}
		}
	}

	// K 线扫描完毕未命中，检查是否已超出观察窗口
	if now.UnixMilli()-record.EventTime > outcomeExpiryMs {
		return "expired", 0, now.UnixMilli()
	}

	return "", 0, 0 // 仍在观察窗口内
}
