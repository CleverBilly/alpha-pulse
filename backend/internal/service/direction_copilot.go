package service

import (
	"fmt"
	"math"
	"strings"

	"alpha-pulse/backend/models"
)

type DirectionDecisionState string

const (
	DirectionStateStrongBullish DirectionDecisionState = "strong-bullish"
	DirectionStateBullish       DirectionDecisionState = "bullish"
	DirectionStateNeutral       DirectionDecisionState = "neutral"
	DirectionStateBearish       DirectionDecisionState = "bearish"
	DirectionStateStrongBearish DirectionDecisionState = "strong-bearish"
	DirectionStateInvalid       DirectionDecisionState = "invalid"
)

type DirectionDecision struct {
	State             DirectionDecisionState `json:"state"`
	Verdict           string                 `json:"verdict"`
	Summary           string                 `json:"summary"`
	Reasons           []string               `json:"reasons"`
	Confidence        int                    `json:"confidence"`
	RiskLabel         string                 `json:"risk_label"`
	Tradable          bool                   `json:"tradable"`
	TradeabilityLabel string                 `json:"tradeability_label"`
	TimeframeLabels   []string               `json:"timeframe_labels"`
}

func BuildDashboardDecision(snapshot MarketSnapshot) DirectionDecision {
	signal := snapshot.Signal
	if strings.TrimSpace(signal.Action) == "" && signal.Score == 0 && signal.Confidence == 0 {
		return DirectionDecision{
			State:             DirectionStateInvalid,
			Verdict:           "当前不建议执行",
			Summary:           "信号快照尚未准备好，先等待下一次 market snapshot 同步。",
			Reasons:           []string{"等待快照同步完成", "当前结论可信度不足"},
			Confidence:        0,
			RiskLabel:         "高风险",
			Tradable:          false,
			TradeabilityLabel: "等待同步",
			TimeframeLabels:   []string{},
		}
	}

	confidence := boundedInt(signal.Confidence, 0, 100)
	state := resolveDashboardDecisionState(float64(signal.Score), float64(confidence))
	reasons := pickDecisionReasons(signal, snapshot.Structure, snapshot.Liquidity, snapshot.OrderFlow)
	if len(reasons) == 0 {
		reasons = []string{"当前结论仍需更多证据确认"}
	}

	tradable := state != DirectionStateInvalid && confidence >= 45
	return DirectionDecision{
		State:             state,
		Verdict:           resolveDecisionVerdict(state),
		Summary:           firstNonEmptyTrimmed(signal.Explain, strings.Join(reasons, "；")),
		Reasons:           reasons,
		Confidence:        confidence,
		RiskLabel:         resolveRiskLabel(state, confidence, signal.RiskReward),
		Tradable:          tradable,
		TradeabilityLabel: ternaryString(tradable, "可继续观察", "等待同步"),
		TimeframeLabels:   []string{},
	}
}

func BuildDirectionCopilotDecision(macroSnapshot, biasSnapshot, triggerSnapshot, executionSnapshot *MarketSnapshot) DirectionDecision {
	if macroSnapshot == nil || biasSnapshot == nil || triggerSnapshot == nil || executionSnapshot == nil {
		return buildNoTradeDecision(
			"方向引擎还没拿齐 4h / 1h / 15m / 5m 快照，先等待同步完成。",
			[]string{"等待 4h / 1h / 15m / 5m 同步", "当前多周期证据不足"},
			0,
			[]string{},
		)
	}

	macroDecision := BuildDashboardDecision(*macroSnapshot)
	biasDecision := BuildDashboardDecision(*biasSnapshot)
	triggerDecision := BuildDashboardDecision(*triggerSnapshot)
	executionDecision := BuildDashboardDecision(*executionSnapshot)
	macroBias := directionToNumeric(macroDecision.State)
	biasBias := directionToNumeric(biasDecision.State)
	triggerBias := directionToNumeric(triggerDecision.State)
	executionBias := directionToNumeric(executionDecision.State)
	timeframeLabels := []string{
		fmt.Sprintf("4h %s", macroDecision.Verdict),
		fmt.Sprintf("1h %s", biasDecision.Verdict),
		fmt.Sprintf("15m %s", triggerDecision.Verdict),
		fmt.Sprintf("5m %s", executionDecision.Verdict),
	}

	if math.Abs(float64(biasBias)) == 0 || biasDecision.Confidence < 55 {
		return buildNoTradeDecision(
			"1h 主判断还不够明确，先等主周期把方向走出来。",
			append([]string{"1h 主周期置信度不足"}, takeTopReasons(biasDecision.Reasons)...),
			biasDecision.Confidence,
			timeframeLabels,
		)
	}

	if directionsConflict(macroBias, biasBias) {
		return buildNoTradeDecision(
			"4h 与 1h 方向互相打架，当前属于逆大级别风险区。",
			append([]string{"4h 与 1h 方向冲突"}, takeTopReasons([]string{firstReason(macroDecision.Reasons), firstReason(biasDecision.Reasons)})...),
			weightedConfidence(macroDecision.Confidence, biasDecision.Confidence, triggerDecision.Confidence, executionDecision.Confidence),
			timeframeLabels,
		)
	}

	if directionsConflict(triggerBias, biasBias) {
		return buildNoTradeDecision(
			"15m 触发还没和 1h 主方向对齐，先不要提前动手。",
			append([]string{"15m 触发未确认"}, takeTopReasons([]string{firstReason(triggerDecision.Reasons), firstReason(biasDecision.Reasons)})...),
			weightedConfidence(macroDecision.Confidence, biasDecision.Confidence, triggerDecision.Confidence, executionDecision.Confidence),
			timeframeLabels,
		)
	}

	if directionsConflict(executionBias, triggerBias) {
		return buildNoTradeDecision(
			"5m 执行触发开始反着 15m 走，先别抢最后一脚。",
			append([]string{"5m 执行触发逆着 15m"}, takeTopReasons([]string{firstReason(executionDecision.Reasons), firstReason(triggerDecision.Reasons)})...),
			weightedConfidence(macroDecision.Confidence, biasDecision.Confidence, triggerDecision.Confidence, executionDecision.Confidence),
			timeframeLabels,
		)
	}

	if crowdingReason := resolveCrowdingReason(biasSnapshot.Futures, biasBias); crowdingReason != "" {
		return buildNoTradeDecision(
			crowdingReason,
			append([]string{"Futures 因子过度拥挤"}, takeTopReasons([]string{formatFuturesReason(biasSnapshot.Futures, biasBias), firstReason(biasDecision.Reasons)})...),
			weightedConfidence(macroDecision.Confidence, biasDecision.Confidence, triggerDecision.Confidence, executionDecision.Confidence),
			timeframeLabels,
		)
	}

	weightedBias := float64(biasBias)*1.35 +
		float64(macroBias)*0.85 +
		float64(triggerBias)*0.55 +
		float64(executionBias)*0.35 +
		futuresSupportScore(biasSnapshot.Futures, biasBias)
	confidence := weightedConfidence(macroDecision.Confidence, biasDecision.Confidence, triggerDecision.Confidence, executionDecision.Confidence)
	state := resolveDirectionalState(weightedBias, confidence)
	reasons := takeTopReasons([]string{
		firstReason(macroDecision.Reasons),
		firstReason(biasDecision.Reasons),
		firstReason(triggerDecision.Reasons),
		firstReason(executionDecision.Reasons),
		formatFuturesReason(biasSnapshot.Futures, biasBias),
	})

	riskLabel := "中风险"
	if confidence >= 72 && math.Abs(weightedBias) >= 2.7 {
		riskLabel = "可控风险"
	}

	return DirectionDecision{
		State:             state,
		Verdict:           resolveDecisionVerdict(state),
		Summary:           buildAlignedSummary(state, biasBias, macroDecision, triggerDecision, executionDecision, biasSnapshot.Futures),
		Reasons:           reasons,
		Confidence:        confidence,
		RiskLabel:         riskLabel,
		Tradable:          true,
		TradeabilityLabel: "A 级可跟踪",
		TimeframeLabels:   timeframeLabels,
	}
}

func HasExecutableSetup(snapshot MarketSnapshot) bool {
	return isFinitePositive(snapshot.Signal.EntryPrice) &&
		isFinitePositive(snapshot.Signal.StopLoss) &&
		isFinitePositive(snapshot.Signal.TargetPrice)
}

func buildNoTradeDecision(summary string, reasons []string, confidence int, timeframeLabels []string) DirectionDecision {
	return DirectionDecision{
		State:             DirectionStateInvalid,
		Verdict:           "当前禁止交易",
		Summary:           summary,
		Reasons:           takeTopReasons(reasons),
		Confidence:        boundedInt(confidence, 0, 100),
		RiskLabel:         "高风险",
		Tradable:          false,
		TradeabilityLabel: "No-Trade",
		TimeframeLabels:   timeframeLabels,
	}
}

func resolveDashboardDecisionState(score float64, confidence float64) DirectionDecisionState {
	if !isFiniteNumber(score) {
		return DirectionStateInvalid
	}
	if confidence < 45 || math.Abs(score) < 12 {
		return DirectionStateNeutral
	}
	if score >= 55 {
		return DirectionStateStrongBullish
	}
	if score >= 20 {
		return DirectionStateBullish
	}
	if score <= -55 {
		return DirectionStateStrongBearish
	}
	if score <= -20 {
		return DirectionStateBearish
	}
	return DirectionStateNeutral
}

func resolveDecisionVerdict(state DirectionDecisionState) string {
	switch state {
	case DirectionStateStrongBullish:
		return "强偏多"
	case DirectionStateBullish:
		return "偏多"
	case DirectionStateBearish:
		return "偏空"
	case DirectionStateStrongBearish:
		return "强偏空"
	case DirectionStateNeutral:
		return "观望"
	default:
		return "当前不建议执行"
	}
}

func resolveRiskLabel(state DirectionDecisionState, confidence int, riskReward float64) string {
	if state == DirectionStateInvalid || state == DirectionStateNeutral || confidence < 45 {
		return "高风险"
	}
	if confidence >= 72 && riskReward >= 2 {
		return "可控风险"
	}
	return "中风险"
}

func pickDecisionReasons(signal models.Signal, structure models.Structure, liquidity models.Liquidity, orderFlow models.OrderFlow) []string {
	factorReasons := make([]string, 0, len(signal.Factors))
	factors := append([]models.SignalFactor(nil), signal.Factors...)
	for len(factors) > 1 {
		swapped := false
		for index := 0; index < len(factors)-1; index++ {
			if absInt(factors[index].Score) < absInt(factors[index+1].Score) {
				factors[index], factors[index+1] = factors[index+1], factors[index]
				swapped = true
			}
		}
		if !swapped {
			break
		}
	}
	for _, factor := range factors {
		reason := strings.TrimSpace(factor.Reason)
		if reason == "" {
			continue
		}
		factorReasons = append(factorReasons, reason)
		if len(factorReasons) == 3 {
			return factorReasons
		}
	}

	extras := []string{}
	if structure.Choch {
		extras = append(extras, "结构发生 CHOCH，波段切换风险上升")
	} else if structure.BOS {
		extras = append(extras, "结构完成 BOS，方向延续更明确")
	}
	if strings.TrimSpace(liquidity.SweepType) != "" {
		extras = append(extras, fmt.Sprintf("%s 已出现，需盯紧后续回收情况", formatSweep(liquidity.SweepType)))
	}
	if orderFlow.Symbol != "" {
		if orderFlow.Delta >= 0 {
			extras = append(extras, "主动买盘仍在扩张")
		} else {
			extras = append(extras, "主动卖盘仍在扩张")
		}
	}

	combined := append(factorReasons, extras...)
	if len(combined) > 3 {
		return combined[:3]
	}
	return combined
}

func directionToNumeric(state DirectionDecisionState) int {
	switch state {
	case DirectionStateStrongBullish:
		return 2
	case DirectionStateBullish:
		return 1
	case DirectionStateStrongBearish:
		return -2
	case DirectionStateBearish:
		return -1
	default:
		return 0
	}
}

func resolveDirectionalState(weightedBias float64, confidence int) DirectionDecisionState {
	if confidence < 55 || math.Abs(weightedBias) < 0.8 {
		return DirectionStateNeutral
	}
	if weightedBias >= 2.6 {
		return DirectionStateStrongBullish
	}
	if weightedBias >= 1.1 {
		return DirectionStateBullish
	}
	if weightedBias <= -2.6 {
		return DirectionStateStrongBearish
	}
	if weightedBias <= -1.1 {
		return DirectionStateBearish
	}
	return DirectionStateNeutral
}

func directionsConflict(left int, right int) bool {
	return left != 0 && right != 0 && left*right < 0
}

func weightedConfidence(macro int, bias int, trigger int, execution int) int {
	return roundFloatToInt(clampFloat(float64(macro)*0.23+float64(bias)*0.42+float64(trigger)*0.22+float64(execution)*0.13, 0, 100))
}

func resolveCrowdingReason(futures FuturesSnapshot, direction int) string {
	if !futures.Available || direction == 0 {
		return ""
	}
	if direction > 0 && futures.LiquidationPressure == "long-squeeze" {
		return "虽然大方向仍偏多，但多头清算压力已经挤在下方，先别追高。"
	}
	if direction < 0 && futures.LiquidationPressure == "short-squeeze" {
		return "虽然大方向仍偏空，但空头清算压力已经挤在上方，先别追空。"
	}
	if direction > 0 && futures.FundingRate >= 0.00025 && futures.LongShortRatio >= 1.12 && futures.BasisBps >= 6 {
		return "虽然方向仍偏多，但 funding、basis 和 long-short ratio 同时拥挤，先别追多。"
	}
	if direction < 0 && futures.FundingRate <= -0.00025 && futures.LongShortRatio <= 0.88 && futures.BasisBps <= -6 {
		return "虽然方向仍偏空，但 funding、basis 和 long-short ratio 同时拥挤，先别追空。"
	}
	return ""
}

func futuresSupportScore(futures FuturesSnapshot, direction int) float64 {
	if !futures.Available || direction == 0 {
		return 0
	}
	if direction > 0 {
		score := 0.0
		if futures.LongShortRatio >= 1.02 {
			score += 0.2
		}
		if futures.BasisBps >= 0 {
			score += 0.15
		}
		if futures.FundingRate >= -0.00005 {
			score += 0.1
		}
		if futures.LiquidationPressure == "short-squeeze" {
			score += 0.12
		}
		return score
	}
	score := 0.0
	if futures.LongShortRatio <= 0.98 {
		score += 0.2
	}
	if futures.BasisBps <= 0 {
		score += 0.15
	}
	if futures.FundingRate <= 0.00005 {
		score += 0.1
	}
	if futures.LiquidationPressure == "long-squeeze" {
		score += 0.12
	}
	return -score
}

func formatFuturesReason(futures FuturesSnapshot, direction int) string {
	if !futures.Available || direction == 0 {
		return ""
	}
	basis := formatSignedNumber(futures.BasisBps, 1)
	funding := fmt.Sprintf("%.3f%%", futures.FundingRate*100)
	if direction > 0 {
		return fmt.Sprintf(
			"Futures 支持偏多，basis %s bps，funding %s，L/S %.2f。%s",
			basis,
			funding,
			roundFloat(futures.LongShortRatio, 2),
			strings.TrimSpace(futures.LiquidationSummary),
		)
	}
	return fmt.Sprintf(
		"Futures 支持偏空，basis %s bps，funding %s，L/S %.2f。%s",
		basis,
		funding,
		roundFloat(futures.LongShortRatio, 2),
		strings.TrimSpace(futures.LiquidationSummary),
	)
}

func buildAlignedSummary(
	state DirectionDecisionState,
	biasDirection int,
	macroDecision DirectionDecision,
	triggerDecision DirectionDecision,
	executionDecision DirectionDecision,
	futures FuturesSnapshot,
) string {
	directionLabel := "做空"
	if state == DirectionStateStrongBullish || state == DirectionStateBullish {
		directionLabel = "做多"
	}
	futuresHint := "Futures 因子暂时缺失。"
	if futures.Available {
		if biasDirection > 0 {
			futuresHint = "Futures 因子没有明显逆风。"
		} else {
			futuresHint = "Futures 因子没有明显反向挤压。"
		}
	}
	return fmt.Sprintf(
		"4h 与 1h 已经对齐，15m 触发和 5m 执行也站在同一边，当前优先考虑%s。%s / %s / %s，%s",
		directionLabel,
		macroDecision.Verdict,
		triggerDecision.Verdict,
		executionDecision.Verdict,
		futuresHint,
	)
}

func takeTopReasons(reasons []string) []string {
	seen := make(map[string]struct{}, len(reasons))
	result := make([]string, 0, 3)
	for _, reason := range reasons {
		normalized := strings.TrimSpace(reason)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
		if len(result) == 3 {
			break
		}
	}
	return result
}

func formatSweep(value string) string {
	switch value {
	case "sell_sweep":
		return "扫下方流动性"
	case "buy_sweep":
		return "扫上方流动性"
	default:
		return "未见明显 sweep"
	}
}

func formatSignedNumber(value float64, digits int) string {
	if !isFiniteNumber(value) {
		return "-"
	}
	rounded := roundFloat(value, digits)
	if rounded > 0 {
		return fmt.Sprintf("+%v", rounded)
	}
	return fmt.Sprintf("%v", rounded)
}

func roundFloat(value float64, digits int) float64 {
	factor := math.Pow10(digits)
	return math.Round(value*factor) / factor
}

func roundFloatToInt(value float64) int {
	return int(math.Round(value))
}

func clampFloat(value float64, min float64, max float64) float64 {
	return math.Min(math.Max(value, min), max)
}

func boundedInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func isFinitePositive(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value > 0
}

func isFiniteNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func firstReason(reasons []string) string {
	if len(reasons) == 0 {
		return ""
	}
	return reasons[0]
}

func firstNonEmptyTrimmed(values ...string) string {
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized != "" {
			return normalized
		}
	}
	return ""
}

func ternaryString(condition bool, left string, right string) string {
	if condition {
		return left
	}
	return right
}
