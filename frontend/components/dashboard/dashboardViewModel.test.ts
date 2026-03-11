import { describe, expect, it } from "vitest";
import { buildMockMarketSnapshot } from "../../test/fixtures/marketSnapshot";
import {
  buildDashboardDecision,
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
