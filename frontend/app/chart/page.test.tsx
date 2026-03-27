import { render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/components/market/MarketSnapshotLoader", () => ({
  default: () => <div data-testid="market-snapshot-loader" />,
}));
vi.mock("@/components/layout/TradingWorkspaceHero", () => ({
  default: () => <section data-testid="trading-workspace-hero">TradingWorkspaceHero</section>,
}));
vi.mock("@/components/chart/KlineChart", () => ({
  default: () => <section data-testid="kline-chart">KlineChart</section>,
}));
vi.mock("@/components/chart/ChartInsightRail", () => ({
  default: () => <section data-testid="chart-insight-rail">ChartInsightRail</section>,
}));
vi.mock("@/components/trading/ChartPositionCalcButton", () => ({
  default: () => <section data-testid="chart-position-calc-button">ChartPositionCalcButton</section>,
}));

import ChartPage from "./page";

describe("ChartPage", () => {
  it("renders a chart-first workspace with overview band and insight rail", () => {
    render(<ChartPage />);

    expect(screen.getByTestId("chart-command-page")).toBeInTheDocument();
    expect(screen.getByTestId("chart-overview-band")).toBeInTheDocument();
    expect(screen.getByTestId("chart-primary-workspace")).toBeInTheDocument();
    expect(screen.getByTestId("chart-insight-side-rail")).toBeInTheDocument();
    expect(screen.getByTestId("chart-action-strip")).toBeInTheDocument();

    expect(within(screen.getByTestId("chart-overview-band")).getByTestId("trading-workspace-hero")).toBeInTheDocument();
    expect(within(screen.getByTestId("chart-primary-workspace")).getByTestId("kline-chart")).toBeInTheDocument();
    expect(within(screen.getByTestId("chart-insight-side-rail")).getByTestId("chart-insight-rail")).toBeInTheDocument();
    expect(within(screen.getByTestId("chart-action-strip")).getByTestId("chart-position-calc-button")).toBeInTheDocument();
  });
});
