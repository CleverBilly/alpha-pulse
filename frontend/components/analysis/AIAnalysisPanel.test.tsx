import React from "react";
import { render, screen } from "@testing-library/react";
import AIAnalysisPanel from "@/components/analysis/AIAnalysisPanel";
import { useMarketStore } from "@/store/marketStore";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/store/marketStore", () => ({
  useMarketStore: vi.fn(),
}));

const mockedUseMarketStore = vi.mocked(useMarketStore);

describe("AIAnalysisPanel", () => {
  it("renders memo, factors and recent signal tape", () => {
    mockedUseMarketStore.mockReturnValue({
      signal: {
        id: 1,
        symbol: "BTCUSDT",
        interval_type: "1h",
        open_time: 1741300000000,
        signal: "BUY",
        score: 58,
        confidence: 74,
        entry_price: 65200,
        stop_loss: 64650,
        target_price: 66420,
        risk_reward: 2.22,
        trend_bias: "bullish",
        explain: "当前多头信号由趋势、订单流与流动性共振驱动。",
        factors: [
          {
            key: "trend",
            name: "Trend",
            score: 18,
            bias: "bullish",
            reason: "EMA20 高于 EMA50",
            section: "indicator",
          },
          {
            key: "microstructure",
            name: "Microstructure",
            score: -4,
            bias: "bearish",
            reason: "上方存在卖方吸收压制",
            section: "microstructure",
          },
        ],
        created_at: "2026-03-07T00:00:00Z",
      },
      signalTimeline: [
        {
          id: 1,
          symbol: "BTCUSDT",
          interval_type: "1h",
          open_time: 1741300000000,
          signal: "BUY",
          score: 58,
          confidence: 74,
          entry_price: 65200,
          stop_loss: 64650,
          target_price: 66420,
        },
      ],
      structure: {
        id: 1,
        symbol: "BTCUSDT",
        trend: "uptrend",
        support: 64800,
        resistance: 66400,
        bos: true,
        choch: false,
        events: [],
        created_at: "2026-03-07T00:00:00Z",
      },
      liquidity: {
        id: 1,
        symbol: "BTCUSDT",
        buy_liquidity: 64820,
        sell_liquidity: 66380,
        sweep_type: "sell_sweep",
        order_book_imbalance: 0.18,
        data_source: "orderbook",
        equal_high: 66320,
        equal_low: 64810,
        stop_clusters: [],
        wall_levels: [],
        created_at: "2026-03-07T00:00:00Z",
      },
      orderFlow: {
        id: 1,
        symbol: "BTCUSDT",
        interval_type: "1h",
        open_time: 1741300000000,
        buy_volume: 1200,
        sell_volume: 860,
        delta: 340,
        cvd: 1600,
        buy_large_trade_count: 6,
        sell_large_trade_count: 2,
        buy_large_trade_notional: 800000,
        sell_large_trade_notional: 220000,
        large_trade_delta: 580000,
        absorption_bias: "buy_absorption",
        absorption_strength: 0.7,
        iceberg_bias: "buy_iceberg",
        iceberg_strength: 0.55,
        data_source: "agg_trade",
        large_trades: [],
        microstructure_events: [
          {
            type: "absorption",
            bias: "bullish",
            score: 5,
            strength: 0.7,
            price: 65180,
            trade_time: 1741300000000,
            detail: "卖压被持续吸收，价格未继续下破",
          },
          {
            type: "initiative_shift",
            bias: "bullish",
            score: 4,
            strength: 0.31,
            price: 65210,
            trade_time: 1741300060000,
            detail: "买方主动性较前半段明显增强",
          },
        ],
        created_at: "2026-03-07T00:00:00Z",
      },
      microstructureEvents: [
        {
          type: "absorption",
          bias: "bullish",
          score: 5,
          strength: 0.7,
          price: 65180,
          trade_time: 1741300000000,
          detail: "卖压被持续吸收，价格未继续下破",
        },
      ],
      indicator: {
        id: 1,
        symbol: "BTCUSDT",
        rsi: 61.2,
        macd: 120,
        macd_signal: 96,
        macd_histogram: 24,
        ema20: 65180,
        ema50: 64620,
        atr: 410,
        bollinger_upper: 65920,
        bollinger_middle: 65110,
        bollinger_lower: 64300,
        vwap: 64980,
        created_at: "2026-03-07T00:00:00Z",
      },
    } as ReturnType<typeof useMarketStore>);

    render(<AIAnalysisPanel />);

    expect(screen.getByText("决策备忘")).toBeInTheDocument();
    expect(screen.getByText("当前多头信号由趋势、订单流与流动性共振驱动。")).toBeInTheDocument();
    expect(screen.getByText("多头驱动")).toBeInTheDocument();
    expect(screen.getByText("风险因子")).toBeInTheDocument();
    expect(screen.getByText("近期信号序列")).toBeInTheDocument();
    expect(screen.getByText("微结构序列")).toBeInTheDocument();
    expect(screen.getByText(/卖压被持续吸收/)).toBeInTheDocument();
    expect(screen.getByText(/做多/)).toBeInTheDocument();
  });
});
