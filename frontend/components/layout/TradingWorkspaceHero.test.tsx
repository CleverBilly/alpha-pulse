import { describe, expect, it, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
import TradingWorkspaceHero from "./TradingWorkspaceHero";

vi.mock("@/store/marketStore", () => ({
  useMarketStore: vi.fn(),
}));

import { useMarketStore } from "@/store/marketStore";

describe("TradingWorkspaceHero", () => {
  it("renders the shared overview band and command panels", () => {
    vi.mocked(useMarketStore).mockReturnValue({
      symbol: "BTCUSDT",
      interval: "15m",
      price: { price: 68123.45 },
      signal: { signal: "BUY", trend_bias: "BULLISH" },
      structure: { trend: "UPTREND" },
      liquidity: { sweep_type: "BUY_SIDE" },
      loading: false,
      error: null,
      transportMode: "websocket",
      streamStatus: "live",
      streamError: null,
      lastUpdatedAt: Date.UTC(2026, 2, 27, 9, 30, 0),
      lastRefreshMode: "cache",
      refreshDashboard: vi.fn(),
      setSymbol: vi.fn(),
      setIntervalType: vi.fn(),
    } as never);

    render(
      <TradingWorkspaceHero
        eyebrow="市场"
        title="命令中心总览"
        description="让总览、控制和度量成为同一块工作台。"
        metrics={[
          { label: "趋势", value: "4h / 1h" },
          { label: "执行", value: "15m / 5m" },
        ]}
      />,
    );

    expect(screen.getByTestId("overview-band")).toBeInTheDocument();
    expect(screen.getByTestId("overview-status-strip")).toBeInTheDocument();
    expect(screen.getByTestId("overview-band-quote")).toHaveAttribute("data-surface", "instrument");
    expect(screen.getByTestId("command-panel-controls")).toHaveAttribute("data-surface", "console");
    expect(screen.getByTestId("command-panel-metrics")).toHaveAttribute("data-surface", "rail");
    expect(screen.getByRole("heading", { name: "命令中心总览" })).toBeInTheDocument();
    expect(within(screen.getByTestId("overview-band-quote")).getByText("BTCUSDT")).toBeInTheDocument();
  });
});
