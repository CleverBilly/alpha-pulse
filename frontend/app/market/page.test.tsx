import { render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/components/market/MarketSnapshotLoader", () => ({
  default: () => <div data-testid="market-snapshot-loader" />,
}));
vi.mock("@/components/market/FuturesWatchlist", () => ({
  default: () => <section data-testid="futures-watchlist">FuturesWatchlist</section>,
}));
vi.mock("@/components/layout/TradingWorkspaceHero", () => ({
  default: () => <section data-testid="trading-workspace-hero">TradingWorkspaceHero</section>,
}));
vi.mock("@/components/market/MarketOverviewBoard", () => ({
  default: () => <section data-testid="market-overview-board">MarketOverviewBoard</section>,
}));
vi.mock("@/components/market/MarketLevelsBoard", () => ({
  default: () => <section data-testid="market-levels-board">MarketLevelsBoard</section>,
}));
vi.mock("@/components/market/SignalTape", () => ({
  default: () => <section data-testid="signal-tape">SignalTape</section>,
}));
vi.mock("@/components/market/MicrostructureTimeline", () => ({
  default: () => <section data-testid="microstructure-timeline">MicrostructureTimeline</section>,
}));
vi.mock("@/components/orderflow/OrderFlowPanel", () => ({
  default: () => <section data-testid="orderflow-panel">OrderFlowPanel</section>,
}));
vi.mock("@/components/liquidity/LiquidityPanel", () => ({
  default: () => <section data-testid="liquidity-panel">LiquidityPanel</section>,
}));

import MarketPage from "./page";

describe("MarketPage", () => {
  it("renders the radar watchlist, central workspaces, and diagnostics strip", () => {
    render(<MarketPage />);

    expect(screen.getByTestId("market-command-page")).toBeInTheDocument();
    expect(screen.getByTestId("market-watchlist-band")).toBeInTheDocument();
    expect(screen.getByTestId("market-overview-band")).toBeInTheDocument();
    expect(screen.getByTestId("market-primary-grid")).toBeInTheDocument();
    expect(screen.getByTestId("market-secondary-rail")).toBeInTheDocument();
    expect(screen.getByTestId("market-diagnostics-strip")).toBeInTheDocument();

    expect(within(screen.getByTestId("market-watchlist-band")).getByTestId("futures-watchlist")).toBeInTheDocument();
    expect(within(screen.getByTestId("market-overview-band")).getByTestId("trading-workspace-hero")).toBeInTheDocument();
    expect(within(screen.getByTestId("market-overview-band")).getByTestId("market-overview-board")).toBeInTheDocument();
    expect(within(screen.getByTestId("market-primary-grid")).getByTestId("market-levels-board")).toBeInTheDocument();
    expect(within(screen.getByTestId("market-secondary-rail")).getByTestId("signal-tape")).toBeInTheDocument();
    expect(within(screen.getByTestId("market-secondary-rail")).getByTestId("microstructure-timeline")).toBeInTheDocument();
    expect(within(screen.getByTestId("market-diagnostics-strip")).getByTestId("orderflow-panel")).toBeInTheDocument();
    expect(within(screen.getByTestId("market-diagnostics-strip")).getByTestId("liquidity-panel")).toBeInTheDocument();
  });
});
