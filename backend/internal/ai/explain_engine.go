package ai

import (
	"fmt"
	"sort"
	"strings"

	"alpha-pulse/backend/models"
)

// Engine 负责将信号转成可读的 AI 解释文本。
type Engine struct{}

// NewEngine 创建 AI Explain Engine。
func NewEngine() *Engine {
	return &Engine{}
}

// Explain 输出对当前信号的中文解释。
func (e *Engine) Explain(signal models.Signal) string {
	factorSummary := summarizeFactors(signal.Factors, 3)

	switch signal.Action {
	case "BUY":
		return fmt.Sprintf(
			"当前为 BUY 信号，综合评分 %d，置信度 %d%%，方向偏向 %s，盈亏比 %.2f。主要驱动：%s。建议关注 %.2f 一线的回踩做多机会，止损 %.2f，目标 %.2f。",
			signal.Score,
			signal.Confidence,
			emptyAs(signal.TrendBias, "bullish"),
			signal.RiskReward,
			factorSummary,
			signal.EntryPrice,
			signal.StopLoss,
			signal.TargetPrice,
		)
	case "SELL":
		return fmt.Sprintf(
			"当前为 SELL 信号，综合评分 %d，置信度 %d%%，方向偏向 %s，盈亏比 %.2f。主要驱动：%s。若价格反弹受阻，可考虑空头策略，止损 %.2f，目标 %.2f。",
			signal.Score,
			signal.Confidence,
			emptyAs(signal.TrendBias, "bearish"),
			signal.RiskReward,
			factorSummary,
			signal.StopLoss,
			signal.TargetPrice,
		)
	default:
		return fmt.Sprintf(
			"当前为 NEUTRAL 信号，综合评分 %d，置信度 %d%%。主要观察点：%s。建议等待更明确的结构突破或订单流共振后再入场。",
			signal.Score,
			signal.Confidence,
			factorSummary,
		)
	}
}

func summarizeFactors(factors []models.SignalFactor, limit int) string {
	if len(factors) == 0 {
		return "暂无评分明细"
	}

	sorted := make([]models.SignalFactor, len(factors))
	copy(sorted, factors)
	sort.Slice(sorted, func(i, j int) bool {
		left := sorted[i].Score
		if left < 0 {
			left = -left
		}
		right := sorted[j].Score
		if right < 0 {
			right = -right
		}
		return left > right
	})

	parts := make([]string, 0, limit)
	for _, factor := range sorted {
		if factor.Score == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s(%+d)：%s", factor.Name, factor.Score, factor.Reason))
		if len(parts) == limit {
			break
		}
	}

	if len(parts) == 0 {
		return "因子暂未形成一致方向"
	}

	return strings.Join(parts, "；")
}

func emptyAs(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
