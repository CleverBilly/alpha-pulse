import { describe, expect, it } from "vitest";
import { buildMockMarketSnapshot } from "../../test/fixtures/marketSnapshot";
import {
  buildDashboardDecision,
  buildDirectionAwareExecutionSetup,
  buildDirectionCopilotDecision,
  buildEvidenceSummary,
  buildExecutionSetup,
} from "./dashboardViewModel";

describe("dashboardViewModel", () => {
  it("builds a strong bullish decision from aligned snapshot inputs", () => {
    const snapshot = buildMockMarketSnapshot();

    const decision = buildDashboardDecision({
      signal: snapshot.signal,
      structure: snapshot.structure,
      liquidity: snapshot.liquidity,
      orderFlow: snapshot.orderflow,
    });
    const setup = buildExecutionSetup({
      signal: snapshot.signal,
      structure: snapshot.structure,
      liquidity: snapshot.liquidity,
      orderFlow: snapshot.orderflow,
      microstructureEvents: snapshot.microstructure_events,
      price: snapshot.price,
    });
    const evidence = buildEvidenceSummary({
      structure: snapshot.structure,
      liquidity: snapshot.liquidity,
      orderFlow: snapshot.orderflow,
      microstructureEvents: snapshot.microstructure_events,
    });

    expect(decision.state).toBe("strong-bullish");
    expect(decision.verdict).toBe("强偏多");
    expect(decision.tradable).toBe(true);
    expect(decision.summary).toContain("趋势");
    expect(decision.reasons).toHaveLength(3);

    expect(setup.status).toBe("ready");
    expect(setup.biasLabel).toBe("顺势做多");
    expect(setup.trigger).toContain("回踩");
    expect(setup.entryHigh).toBeGreaterThanOrEqual(setup.entryLow);

    expect(evidence).toHaveLength(3);
    expect(evidence[0].status).toBe("ready");
  });

  it("builds a bearish decision when score, structure and orderflow turn lower", () => {
    const snapshot = structuredClone(buildMockMarketSnapshot());
    snapshot.signal.signal = "SELL";
    snapshot.signal.score = -64;
    snapshot.signal.confidence = 71;
    snapshot.signal.trend_bias = "bearish";
    snapshot.signal.explain = "空头由结构转弱、卖压主动性和流动性回补共同驱动。";
    snapshot.signal.entry_price = snapshot.price.price + 12;
    snapshot.signal.stop_loss = snapshot.price.price + 138;
    snapshot.signal.target_price = snapshot.price.price - 210;
    snapshot.signal.risk_reward = 1.8;
    snapshot.signal.factors = [
      {
        key: "structure",
        name: "Structure",
        score: -18,
        bias: "bearish",
        reason: "结构跌破前低，且出现 CHOCH",
        section: "structure",
      },
      {
        key: "orderflow",
        name: "Order Flow",
        score: -13,
        bias: "bearish",
        reason: "Delta 转负且卖方大单净流入抬升",
        section: "flow",
      },
      {
        key: "liquidity",
        name: "Liquidity",
        score: -8,
        bias: "bearish",
        reason: "buy-side sweep 后，上方卖墙重新增厚",
        section: "liquidity",
      },
    ];
    snapshot.structure.trend = "downtrend";
    snapshot.structure.bos = false;
    snapshot.structure.choch = true;
    snapshot.orderflow.delta = -420;
    snapshot.orderflow.large_trade_delta = -580000;
    snapshot.orderflow.absorption_bias = "sell_absorption";
    snapshot.orderflow.iceberg_bias = "sell_iceberg";
    snapshot.liquidity.sweep_type = "buy_sweep";
    snapshot.liquidity.order_book_imbalance = -0.22;

    const decision = buildDashboardDecision({
      signal: snapshot.signal,
      structure: snapshot.structure,
      liquidity: snapshot.liquidity,
      orderFlow: snapshot.orderflow,
    });
    const setup = buildExecutionSetup({
      signal: snapshot.signal,
      structure: snapshot.structure,
      liquidity: snapshot.liquidity,
      orderFlow: snapshot.orderflow,
      microstructureEvents: snapshot.microstructure_events,
      price: snapshot.price,
    });

    expect(decision.state).toBe("strong-bearish");
    expect(decision.verdict).toBe("强偏空");
    expect(decision.summary).toContain("结构");
    expect(setup.status).toBe("ready");
    expect(setup.biasLabel).toBe("顺势做空");
  });

  it("falls back to a neutral wait-and-see view for low-conviction signals", () => {
    const snapshot = structuredClone(buildMockMarketSnapshot());
    snapshot.signal.signal = "NEUTRAL";
    snapshot.signal.score = 6;
    snapshot.signal.confidence = 39;
    snapshot.signal.trend_bias = "neutral";
    snapshot.signal.explain = "信号分数接近中性，尚未形成高质量方向共振。";
    snapshot.signal.factors = [
      {
        key: "trend",
        name: "Trend",
        score: 2,
        bias: "neutral",
        reason: "趋势尚未明确",
        section: "indicator",
      },
      {
        key: "risk",
        name: "Risk",
        score: -1,
        bias: "neutral",
        reason: "需要等待确认",
        section: "risk",
      },
    ];

    const decision = buildDashboardDecision({
      signal: snapshot.signal,
      structure: snapshot.structure,
      liquidity: snapshot.liquidity,
      orderFlow: snapshot.orderflow,
    });

    expect(decision.state).toBe("neutral");
    expect(decision.verdict).toBe("观望");
    expect(decision.riskLabel).toBe("高风险");
  });

  it("builds a tradable multi-timeframe futures direction decision when 4h / 1h / 15m / 5m align", () => {
    const macro = buildMockMarketSnapshot("BTCUSDT", "4h", 48);
    const bias = buildMockMarketSnapshot("BTCUSDT", "1h", 48);
    const trigger = buildMockMarketSnapshot("BTCUSDT", "15m", 48);
    const execution = buildMockMarketSnapshot("BTCUSDT", "5m", 48);

    const decision = buildDirectionCopilotDecision({
      macroSnapshot: macro,
      biasSnapshot: bias,
      triggerSnapshot: trigger,
      executionSnapshot: execution,
    });

    expect(decision.state).toBe("strong-bullish");
    expect(decision.verdict).toBe("强偏多");
    expect(decision.tradable).toBe(true);
    expect(decision.tradeabilityLabel).toBe("A 级可跟踪");
    expect(decision.timeframeLabels).toEqual(["4h 强偏多", "1h 强偏多", "15m 强偏多", "5m 强偏多"]);
  });

  it("returns no-trade when 4h and 1h conflict or futures are overcrowded", () => {
    const macro = structuredClone(buildMockMarketSnapshot("BTCUSDT", "4h", 48));
    macro.signal.signal = "SELL";
    macro.signal.score = -62;
    macro.signal.confidence = 74;
    macro.structure.trend = "downtrend";

    const bias = structuredClone(buildMockMarketSnapshot("BTCUSDT", "1h", 48));
    bias.futures.funding_rate = 0.00031;
    bias.futures.long_short_ratio = 1.18;
    bias.futures.basis_bps = 8.4;

    const trigger = buildMockMarketSnapshot("BTCUSDT", "15m", 48);
    const execution = buildMockMarketSnapshot("BTCUSDT", "5m", 48);

    const decision = buildDirectionCopilotDecision({
      macroSnapshot: macro,
      biasSnapshot: bias,
      triggerSnapshot: trigger,
      executionSnapshot: execution,
    });
    const setup = buildDirectionAwareExecutionSetup({
      signal: bias.signal,
      structure: bias.structure,
      liquidity: bias.liquidity,
      orderFlow: bias.orderflow,
      microstructureEvents: bias.microstructure_events,
      price: bias.price,
      decision,
    });

    expect(decision.state).toBe("invalid");
    expect(decision.verdict).toBe("当前禁止交易");
    expect(decision.tradable).toBe(false);
    expect(decision.tradeabilityLabel).toBe("No-Trade");
    expect(setup.status).toBe("unavailable");
    expect(setup.reason).toContain("4h 与 1h");
  });

  it("returns no-trade when 5m execution conflicts with the 15m trigger", () => {
    const macro = buildMockMarketSnapshot("BTCUSDT", "4h", 48);
    const bias = buildMockMarketSnapshot("BTCUSDT", "1h", 48);
    const trigger = buildMockMarketSnapshot("BTCUSDT", "15m", 48);
    const execution = structuredClone(buildMockMarketSnapshot("BTCUSDT", "5m", 48));
    execution.signal.signal = "SELL";
    execution.signal.score = -46;
    execution.signal.confidence = 68;
    execution.signal.trend_bias = "bearish";
    execution.signal.explain = "5m 出现反向执行回落，先别追单。";
    execution.structure.trend = "downtrend";

    const decision = buildDirectionCopilotDecision({
      macroSnapshot: macro,
      biasSnapshot: bias,
      triggerSnapshot: trigger,
      executionSnapshot: execution,
    });

    expect(decision.state).toBe("invalid");
    expect(decision.tradeabilityLabel).toBe("No-Trade");
    expect(decision.summary).toContain("5m 执行触发开始反着 15m 走");
    expect(decision.timeframeLabels).toEqual(["4h 强偏多", "1h 强偏多", "15m 强偏多", "5m 偏空"]);
  });

  it("returns invalid execution when critical trade levels are missing", () => {
    const snapshot = structuredClone(buildMockMarketSnapshot());
    snapshot.signal.entry_price = Number.NaN;
    snapshot.signal.stop_loss = Number.NaN;
    snapshot.signal.target_price = Number.NaN;

    const decision = buildDashboardDecision({
      signal: null,
      structure: snapshot.structure,
      liquidity: snapshot.liquidity,
      orderFlow: snapshot.orderflow,
    });
    const setup = buildExecutionSetup({
      signal: snapshot.signal,
      structure: snapshot.structure,
      liquidity: snapshot.liquidity,
      orderFlow: snapshot.orderflow,
      microstructureEvents: snapshot.microstructure_events,
      price: snapshot.price,
    });
    const evidence = buildEvidenceSummary({
      structure: snapshot.structure,
      liquidity: null,
      orderFlow: null,
      microstructureEvents: [],
    });

    expect(decision.state).toBe("invalid");
    expect(decision.verdict).toBe("当前不建议执行");
    expect(setup.status).toBe("unavailable");
    expect(setup.reason).toContain("等待");
    expect(evidence.find((card) => card.id === "orderflow")?.status).toBe("unavailable");
    expect(evidence.find((card) => card.id === "liquidity")?.status).toBe("unavailable");
    expect(evidence.find((card) => card.id === "structure")?.status).toBe("ready");
  });
});
