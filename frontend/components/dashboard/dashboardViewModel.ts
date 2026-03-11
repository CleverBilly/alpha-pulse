import type {
  FuturesSnapshot,
  Liquidity,
  OrderFlow,
  OrderFlowMicrostructureEvent,
  PriceTicker,
  Structure,
} from "@/types/market";
import type { MarketSnapshot } from "@/types/snapshot";
import type { Signal } from "@/types/signal";

export type DashboardDecisionState =
  | "strong-bullish"
  | "bullish"
  | "neutral"
  | "bearish"
  | "strong-bearish"
  | "invalid";

export type DashboardTone = "positive" | "negative" | "neutral" | "warning";

export interface DashboardDecision {
  state: DashboardDecisionState;
  tone: DashboardTone;
  verdict: string;
  summary: string;
  reasons: string[];
  confidence: number;
  riskLabel: string;
  tradable: boolean;
  tradeabilityLabel: string;
  timeframeLabels: string[];
}

export interface ExecutionSetup {
  status: "ready" | "unavailable";
  tone: DashboardTone;
  biasLabel: string;
  reason: string;
  entryLow: number;
  entryHigh: number;
  stopLoss: number;
  target: number;
  riskReward: number;
  trigger: string;
  caution: string;
}

export interface EvidenceMetric {
  label: string;
  value: string;
}

export interface EvidenceCardView {
  id: "orderflow" | "liquidity" | "structure";
  title: string;
  href: string;
  ctaLabel: string;
  status: "ready" | "unavailable";
  tone: DashboardTone;
  summary: string;
  metrics: EvidenceMetric[];
}

type DecisionInput = {
  signal?: Signal | null;
  structure?: Structure | null;
  liquidity?: Liquidity | null;
  orderFlow?: OrderFlow | null;
};

type SetupInput = DecisionInput & {
  microstructureEvents?: OrderFlowMicrostructureEvent[] | null;
  price?: PriceTicker | null;
};

type EvidenceInput = {
  structure?: Structure | null;
  liquidity?: Liquidity | null;
  orderFlow?: OrderFlow | null;
  microstructureEvents?: OrderFlowMicrostructureEvent[] | null;
};

type DirectionCopilotInput = {
  macroSnapshot?: MarketSnapshot | null;
  biasSnapshot?: MarketSnapshot | null;
  triggerSnapshot?: MarketSnapshot | null;
};

export function buildDashboardDecision({
  signal,
  structure,
  liquidity,
  orderFlow,
}: DecisionInput): DashboardDecision {
  if (!signal) {
    return {
      state: "invalid",
      tone: "warning",
      verdict: "当前不建议执行",
      summary: "信号快照尚未准备好，先等待下一次 market snapshot 同步。",
      reasons: ["等待快照同步完成", "当前结论可信度不足"],
      confidence: 0,
      riskLabel: "高风险",
      tradable: false,
      tradeabilityLabel: "等待同步",
      timeframeLabels: [],
    };
  }

  const confidence = clamp(signal.confidence, 0, 100);
  const state = resolveDecisionState(signal.score, confidence);
  const reasons = pickReasons(signal, structure, liquidity, orderFlow);

  return {
    state,
    tone: resolveTone(state),
    verdict: resolveVerdict(state),
    summary: signal.explain?.trim() || reasons.join("；"),
    reasons,
    confidence,
    riskLabel: resolveRiskLabel(state, confidence, signal.risk_reward),
    tradable: state !== "invalid" && confidence >= 45,
    tradeabilityLabel: state !== "invalid" && confidence >= 45 ? "可继续观察" : "等待同步",
    timeframeLabels: [],
  };
}

export function buildDirectionCopilotDecision({
  macroSnapshot,
  biasSnapshot,
  triggerSnapshot,
}: DirectionCopilotInput): DashboardDecision {
  if (!macroSnapshot || !biasSnapshot || !triggerSnapshot) {
    return {
      state: "invalid",
      tone: "warning",
      verdict: "当前禁止交易",
      summary: "方向引擎还没拿齐 4h / 1h / 15m 快照，先等待同步完成。",
      reasons: ["等待 4h / 1h / 15m 同步", "当前多周期证据不足"],
      confidence: 0,
      riskLabel: "高风险",
      tradable: false,
      tradeabilityLabel: "No-Trade",
      timeframeLabels: [],
    };
  }

  const macroDecision = buildDashboardDecision(fromSnapshot(macroSnapshot));
  const biasDecision = buildDashboardDecision(fromSnapshot(biasSnapshot));
  const triggerDecision = buildDashboardDecision(fromSnapshot(triggerSnapshot));
  const macroBias = directionToNumeric(macroDecision.state);
  const biasBias = directionToNumeric(biasDecision.state);
  const triggerBias = directionToNumeric(triggerDecision.state);
  const timeframeLabels = [
    `4h ${macroDecision.verdict}`,
    `1h ${biasDecision.verdict}`,
    `15m ${triggerDecision.verdict}`,
  ];

  if (Math.abs(biasBias) === 0 || biasDecision.confidence < 55) {
    return buildNoTradeDecision({
      summary: "1h 主判断还不够明确，先等主周期把方向走出来。",
      reasons: ["1h 主周期置信度不足", ...takeTopReasons(biasDecision.reasons)],
      confidence: biasDecision.confidence,
      timeframeLabels,
    });
  }

  if (macroBias !== 0 && macroBias !== biasBias) {
    return buildNoTradeDecision({
      summary: "4h 与 1h 方向互相打架，当前属于逆大级别风险区。",
      reasons: ["4h 与 1h 方向冲突", ...takeTopReasons([macroDecision.reasons[0], biasDecision.reasons[0]])],
      confidence: weightedConfidence(macroDecision.confidence, biasDecision.confidence, triggerDecision.confidence),
      timeframeLabels,
    });
  }

  if (triggerBias !== 0 && triggerBias !== biasBias) {
    return buildNoTradeDecision({
      summary: "15m 触发还没和 1h 主方向对齐，先不要提前动手。",
      reasons: ["15m 触发未确认", ...takeTopReasons([triggerDecision.reasons[0], biasDecision.reasons[0]])],
      confidence: weightedConfidence(macroDecision.confidence, biasDecision.confidence, triggerDecision.confidence),
      timeframeLabels,
    });
  }

  const crowdingReason = resolveCrowdingReason(biasSnapshot.futures, biasBias);
  if (crowdingReason) {
    return buildNoTradeDecision({
      summary: crowdingReason,
      reasons: ["Futures 因子过度拥挤", ...takeTopReasons([formatFuturesReason(biasSnapshot.futures, biasBias), biasDecision.reasons[0]])],
      confidence: weightedConfidence(macroDecision.confidence, biasDecision.confidence, triggerDecision.confidence),
      timeframeLabels,
    });
  }

  const weightedBias = biasBias * 1.35 + macroBias * 0.8 + triggerBias * 0.55 + futuresSupportScore(biasSnapshot.futures, biasBias);
  const confidence = weightedConfidence(macroDecision.confidence, biasDecision.confidence, triggerDecision.confidence);
  const state = resolveDirectionalState(weightedBias, confidence);
  const reasons = takeTopReasons([
    macroDecision.reasons[0],
    biasDecision.reasons[0],
    triggerDecision.reasons[0],
    formatFuturesReason(biasSnapshot.futures, biasBias),
  ]);

  return {
    state,
    tone: resolveTone(state),
    verdict: resolveVerdict(state),
    summary: buildAlignedSummary(state, biasBias, macroDecision, triggerDecision, biasSnapshot.futures),
    reasons,
    confidence,
    riskLabel: confidence >= 72 && Math.abs(weightedBias) >= 2.7 ? "可控风险" : "中风险",
    tradable: true,
    tradeabilityLabel: "A 级可跟踪",
    timeframeLabels,
  };
}

export function buildExecutionSetup({
  signal,
  structure,
  liquidity,
  orderFlow,
  microstructureEvents,
  price,
}: SetupInput): ExecutionSetup {
  const entry = signal?.entry_price;
  const stopLoss = signal?.stop_loss;
  const target = signal?.target_price;
  const riskReward = signal?.risk_reward;

  if (!signal || !isFinitePositive(entry) || !isFinitePositive(stopLoss) || !isFinitePositive(target)) {
    return {
      status: "unavailable",
      tone: "warning",
      biasLabel: "等待 setup",
      reason: "等待更完整的入场、止损和目标位后再做执行判断。",
      entryLow: 0,
      entryHigh: 0,
      stopLoss: 0,
      target: 0,
      riskReward: 0,
      trigger: "先确认最新快照是否完整。",
      caution: "当前不展示伪造点位。",
    };
  }

  const longBias = signal.score >= 0;
  const currentPrice = isFinitePositive(price?.price) ? Number(price?.price) : entry;
  const structureLevel = longBias
    ? firstFinite(structure?.internal_support, structure?.support, liquidity?.buy_liquidity, currentPrice)
    : firstFinite(structure?.internal_resistance, structure?.resistance, liquidity?.sell_liquidity, currentPrice);
  const span = Math.max(Math.abs(entry - structureLevel) * 0.28, currentPrice * 0.0012, 8);
  const entryLow = round(Math.min(entry, currentPrice, entry - span));
  const entryHigh = round(Math.max(entry, currentPrice, entry + span));
  const tone = longBias ? "positive" : "negative";
  const normalizedRiskReward = typeof riskReward === "number" && Number.isFinite(riskReward) ? riskReward : 0;
  const flowCue = resolveFlowCue(orderFlow, longBias);
  const structureCue = longBias
    ? structure?.trend === "uptrend"
      ? "结构不破多头节奏"
      : "结构重新收回支撑"
    : structure?.trend === "downtrend"
      ? "结构继续压制"
      : "结构反抽不过阻力";
  const trigger = longBias
    ? `等待回踩 ${formatRange(entryLow, entryHigh)} 后，确认 ${flowCue}，并保持 ${structureCue}。`
    : `等待反抽 ${formatRange(entryLow, entryHigh)} 后，确认 ${flowCue}，并保持 ${structureCue}。`;

  return {
    status: "ready",
    tone,
    biasLabel: longBias ? "顺势做多" : "顺势做空",
    reason: signal.explain?.trim() || "当前 setup 由多因子共振提供支持。",
    entryLow,
    entryHigh,
    stopLoss: round(stopLoss),
    target: round(target),
    riskReward: round(normalizedRiskReward, 2),
    trigger,
    caution: resolveCaution(signal, liquidity, microstructureEvents, longBias),
  };
}

export function buildDirectionAwareExecutionSetup(
  input: SetupInput & { decision?: DashboardDecision | null },
): ExecutionSetup {
  if (input.decision && !input.decision.tradable) {
    return {
      status: "unavailable",
      tone: "warning",
      biasLabel: "No-Trade",
      reason: input.decision.summary,
      entryLow: 0,
      entryHigh: 0,
      stopLoss: 0,
      target: 0,
      riskReward: 0,
      trigger: "等待 4h / 1h / 15m 重新对齐后再看 setup。",
      caution: "当前不展示伪造点位。",
    };
  }

  return buildExecutionSetup(input);
}

export function buildEvidenceSummary({
  structure,
  liquidity,
  orderFlow,
  microstructureEvents,
}: EvidenceInput): EvidenceCardView[] {
  return [
    buildOrderFlowEvidence(orderFlow, microstructureEvents),
    buildLiquidityEvidence(liquidity),
    buildStructureEvidence(structure, microstructureEvents),
  ];
}

function buildOrderFlowEvidence(
  orderFlow?: OrderFlow | null,
  microstructureEvents?: OrderFlowMicrostructureEvent[] | null,
): EvidenceCardView {
  if (!orderFlow) {
    return {
      id: "orderflow",
      title: "Order Flow",
      href: "/signals",
      ctaLabel: "查看信号深页",
      status: "unavailable",
      tone: "warning",
      summary: "订单流快照尚未同步，先等待大单与 delta 更新。",
      metrics: [],
    };
  }

  const lastEvent = [...(microstructureEvents ?? [])].sort((left, right) => right.trade_time - left.trade_time)[0];
  const tone = orderFlow.delta >= 0 ? "positive" : "negative";

  return {
    id: "orderflow",
    title: "Order Flow",
    href: "/signals",
    ctaLabel: "查看信号深页",
    status: "ready",
    tone,
    summary:
      orderFlow.delta >= 0
        ? `主动买盘占优，${formatAbsorption(orderFlow.absorption_bias)}，最近事件为 ${formatEventName(lastEvent?.type)}。`
        : `主动卖盘占优，${formatAbsorption(orderFlow.absorption_bias)}，最近事件为 ${formatEventName(lastEvent?.type)}。`,
    metrics: [
      { label: "Delta", value: formatSigned(orderFlow.delta, 0) },
      { label: "Large Trade", value: formatSigned(orderFlow.large_trade_delta, 0) },
    ],
  };
}

function buildLiquidityEvidence(liquidity?: Liquidity | null): EvidenceCardView {
  if (!liquidity) {
    return {
      id: "liquidity",
      title: "Liquidity",
      href: "/market",
      ctaLabel: "查看市场深页",
      status: "unavailable",
      tone: "warning",
      summary: "流动性墙位暂未同步，先等待订单簿热区刷新。",
      metrics: [],
    };
  }

  const tone = liquidity.order_book_imbalance >= 0 ? "positive" : "negative";
  const summary = liquidity.sweep_type
    ? `${formatSweep(liquidity.sweep_type)} 后，${liquidity.order_book_imbalance >= 0 ? "买方承接更厚" : "卖方压制更强"}。`
    : `${liquidity.order_book_imbalance >= 0 ? "买方" : "卖方"}在近端墙位上更占优。`;

  return {
    id: "liquidity",
    title: "Liquidity",
    href: "/market",
    ctaLabel: "查看市场深页",
    status: "ready",
    tone,
    summary,
    metrics: [
      { label: "Imbalance", value: formatSigned(liquidity.order_book_imbalance, 2) },
      { label: "Sweep", value: formatSweep(liquidity.sweep_type || "none") },
    ],
  };
}

function buildStructureEvidence(
  structure?: Structure | null,
  microstructureEvents?: OrderFlowMicrostructureEvent[] | null,
): EvidenceCardView {
  if (!structure) {
    return {
      id: "structure",
      title: "Structure & Microstructure",
      href: "/chart",
      ctaLabel: "查看图表深页",
      status: "unavailable",
      tone: "warning",
      summary: "结构与微结构数据尚未准备好。",
      metrics: [],
    };
  }

  const latestEvent = [...structure.events].sort((left, right) => right.open_time - left.open_time)[0];
  const lastMicro = [...(microstructureEvents ?? [])].sort((left, right) => right.trade_time - left.trade_time)[0];
  const tone =
    structure.trend === "uptrend" ? "positive" : structure.trend === "downtrend" ? "negative" : "neutral";

  return {
    id: "structure",
    title: "Structure & Microstructure",
    href: "/chart",
    ctaLabel: "查看图表深页",
    status: "ready",
    tone,
    summary: `${formatTrend(structure.trend)}，最近结构事件 ${latestEvent?.label ?? "none"}，微结构最近出现 ${formatEventName(lastMicro?.type)}。`,
    metrics: [
      { label: "Trend", value: formatTrend(structure.trend) },
      { label: "Event", value: latestEvent?.label ?? "None" },
    ],
  };
}

function resolveDecisionState(score: number, confidence: number): DashboardDecisionState {
  if (!Number.isFinite(score)) {
    return "invalid";
  }
  if (confidence < 45 || Math.abs(score) < 12) {
    return "neutral";
  }
  if (score >= 55) {
    return "strong-bullish";
  }
  if (score >= 20) {
    return "bullish";
  }
  if (score <= -55) {
    return "strong-bearish";
  }
  if (score <= -20) {
    return "bearish";
  }
  return "neutral";
}

function resolveVerdict(state: DashboardDecisionState) {
  switch (state) {
    case "strong-bullish":
      return "强偏多";
    case "bullish":
      return "偏多";
    case "bearish":
      return "偏空";
    case "strong-bearish":
      return "强偏空";
    case "neutral":
      return "观望";
    default:
      return "当前不建议执行";
  }
}

function resolveTone(state: DashboardDecisionState): DashboardTone {
  if (state === "strong-bullish" || state === "bullish") {
    return "positive";
  }
  if (state === "strong-bearish" || state === "bearish") {
    return "negative";
  }
  if (state === "invalid") {
    return "warning";
  }
  return "neutral";
}

function resolveRiskLabel(state: DashboardDecisionState, confidence: number, riskReward: number) {
  if (state === "invalid" || state === "neutral" || confidence < 45) {
    return "高风险";
  }
  if (confidence >= 72 && riskReward >= 2) {
    return "可控风险";
  }
  return "中风险";
}

function pickReasons(
  signal: Signal,
  structure?: Structure | null,
  liquidity?: Liquidity | null,
  orderFlow?: OrderFlow | null,
) {
  const factorReasons = [...signal.factors]
    .sort((left, right) => Math.abs(right.score) - Math.abs(left.score))
    .filter((factor) => factor.reason.trim().length > 0)
    .slice(0, 3)
    .map((factor) => factor.reason.trim());

  if (factorReasons.length >= 3) {
    return factorReasons;
  }

  const extras = [
    structure?.choch ? "结构发生 CHOCH，波段切换风险上升" : structure?.bos ? "结构完成 BOS，方向延续更明确" : "",
    liquidity?.sweep_type ? `${formatSweep(liquidity.sweep_type)} 已出现，需盯紧后续回收情况` : "",
    orderFlow ? `${orderFlow.delta >= 0 ? "主动买盘" : "主动卖盘"}仍在扩张` : "",
  ].filter(Boolean);

  return [...factorReasons, ...extras].slice(0, 3);
}

function resolveFlowCue(orderFlow: OrderFlow | null | undefined, longBias: boolean) {
  if (!orderFlow) {
    return longBias ? "delta 回到偏多" : "delta 保持偏空";
  }

  if (longBias) {
    return orderFlow.delta >= 0 ? "delta 继续偏多" : "卖压衰竭后重新转正";
  }
  return orderFlow.delta <= 0 ? "delta 继续偏空" : "买盘反抽未能延续";
}

function resolveCaution(
  signal: Signal,
  liquidity?: Liquidity | null,
  microstructureEvents?: OrderFlowMicrostructureEvent[] | null,
  longBias?: boolean,
) {
  const lastEvent = [...(microstructureEvents ?? [])].sort((left, right) => right.trade_time - left.trade_time)[0];
  const sweepHint = liquidity?.sweep_type ? `${formatSweep(liquidity.sweep_type)} 后不追价。` : "";
  const eventHint = lastEvent ? `留意 ${formatEventName(lastEvent.type)} 是否失效。` : "";
  const confidenceHint = signal.confidence < 65 ? "仓位宜保守。" : longBias ? "若回踩不守支撑，立即退出。" : "若反抽强穿阻力，立即退出。";
  return [sweepHint, eventHint, confidenceHint].filter(Boolean).join(" ");
}

function fromSnapshot(snapshot: MarketSnapshot): DecisionInput {
  return {
    signal: snapshot.signal,
    structure: snapshot.structure,
    liquidity: snapshot.liquidity,
    orderFlow: snapshot.orderflow,
  };
}

function buildNoTradeDecision({
  summary,
  reasons,
  confidence,
  timeframeLabels,
}: {
  summary: string;
  reasons: string[];
  confidence: number;
  timeframeLabels: string[];
}): DashboardDecision {
  return {
    state: "invalid",
    tone: "warning",
    verdict: "当前禁止交易",
    summary,
    reasons: takeTopReasons(reasons),
    confidence: clamp(confidence, 0, 100),
    riskLabel: "高风险",
    tradable: false,
    tradeabilityLabel: "No-Trade",
    timeframeLabels,
  };
}

function directionToNumeric(state: DashboardDecisionState) {
  switch (state) {
    case "strong-bullish":
      return 2;
    case "bullish":
      return 1;
    case "strong-bearish":
      return -2;
    case "bearish":
      return -1;
    default:
      return 0;
  }
}

function resolveDirectionalState(weightedBias: number, confidence: number): DashboardDecisionState {
  if (confidence < 55 || Math.abs(weightedBias) < 0.8) {
    return "neutral";
  }
  if (weightedBias >= 2.6) {
    return "strong-bullish";
  }
  if (weightedBias >= 1.1) {
    return "bullish";
  }
  if (weightedBias <= -2.6) {
    return "strong-bearish";
  }
  if (weightedBias <= -1.1) {
    return "bearish";
  }
  return "neutral";
}

function weightedConfidence(macro: number, bias: number, trigger: number) {
  return round(clamp(macro * 0.25 + bias * 0.5 + trigger * 0.25, 0, 100), 0);
}

function resolveCrowdingReason(futures: FuturesSnapshot | undefined, direction: number) {
  if (!futures?.available || direction === 0) {
    return "";
  }

  if (direction > 0 && futures.funding_rate >= 0.00025 && futures.long_short_ratio >= 1.12 && futures.basis_bps >= 6) {
    return "虽然方向仍偏多，但 funding、basis 和 long-short ratio 同时拥挤，先别追多。";
  }

  if (direction < 0 && futures.funding_rate <= -0.00025 && futures.long_short_ratio <= 0.88 && futures.basis_bps <= -6) {
    return "虽然方向仍偏空，但 funding、basis 和 long-short ratio 同时拥挤，先别追空。";
  }

  return "";
}

function futuresSupportScore(futures: FuturesSnapshot | undefined, direction: number) {
  if (!futures?.available || direction === 0) {
    return 0;
  }

  if (direction > 0) {
    let score = 0;
    if (futures.long_short_ratio >= 1.02) {
      score += 0.2;
    }
    if (futures.basis_bps >= 0) {
      score += 0.15;
    }
    if (futures.funding_rate >= -0.00005) {
      score += 0.1;
    }
    return score;
  }

  let score = 0;
  if (futures.long_short_ratio <= 0.98) {
    score += 0.2;
  }
  if (futures.basis_bps <= 0) {
    score += 0.15;
  }
  if (futures.funding_rate <= 0.00005) {
    score += 0.1;
  }
  return -score;
}

function formatFuturesReason(futures: FuturesSnapshot | undefined, direction: number) {
  if (!futures?.available || direction === 0) {
    return "";
  }

  const basis = formatSigned(futures.basis_bps, 1);
  const funding = `${(futures.funding_rate * 100).toFixed(3)}%`;
  if (direction > 0) {
    return `Futures 支持偏多，basis ${basis} bps，funding ${funding}，L/S ${round(futures.long_short_ratio, 2)}。`;
  }
  return `Futures 支持偏空，basis ${basis} bps，funding ${funding}，L/S ${round(futures.long_short_ratio, 2)}。`;
}

function buildAlignedSummary(
  state: DashboardDecisionState,
  biasDirection: number,
  macroDecision: DashboardDecision,
  triggerDecision: DashboardDecision,
  futures: FuturesSnapshot | undefined,
) {
  const directionLabel = state === "strong-bullish" || state === "bullish" ? "做多" : "做空";
  const futuresHint = futures?.available
    ? `Futures 因子 ${biasDirection > 0 ? "没有明显逆风" : "没有明显反向挤压"}。`
    : "Futures 因子暂时缺失。";
  return `4h 与 1h 已经对齐，15m 触发也站在同一边，当前优先考虑${directionLabel}。${macroDecision.verdict} / ${triggerDecision.verdict}，${futuresHint}`;
}

function takeTopReasons(reasons: Array<string | undefined>) {
  const normalized = reasons.filter((reason): reason is string => Boolean(reason && reason.trim()));
  return [...new Set(normalized)].slice(0, 3);
}

function formatSweep(value: string) {
  if (value === "sell_sweep") {
    return "扫下方流动性";
  }
  if (value === "buy_sweep") {
    return "扫上方流动性";
  }
  return "未见明显 sweep";
}

function formatAbsorption(value?: string) {
  if (value === "buy_absorption") {
    return "买方吸收仍在";
  }
  if (value === "sell_absorption") {
    return "卖方吸收仍在";
  }
  return "吸收信号中性";
}

function formatEventName(value?: string) {
  if (!value) {
    return "none";
  }
  return value
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function formatTrend(trend?: string) {
  if (trend === "uptrend") {
    return "Uptrend";
  }
  if (trend === "downtrend") {
    return "Downtrend";
  }
  return "Range";
}

function formatSigned(value: number, digits: number) {
  if (!Number.isFinite(value)) {
    return "-";
  }
  const rounded = round(value, digits);
  return rounded > 0 ? `+${rounded}` : `${rounded}`;
}

function formatRange(left: number, right: number) {
  return `${round(left)} - ${round(right)}`;
}

function round(value: number, digits = 2) {
  const factor = 10 ** digits;
  return Math.round(value * factor) / factor;
}

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max);
}

function isFinitePositive(value: number | null | undefined): value is number {
  return typeof value === "number" && Number.isFinite(value) && value > 0;
}

function firstFinite(...values: Array<number | null | undefined>) {
  for (const value of values) {
    if (isFinitePositive(value)) {
      return value;
    }
  }
  return 0;
}
