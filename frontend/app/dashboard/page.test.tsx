import { render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/components/market/MarketSnapshotLoader", () => ({
  default: () => <div data-testid="market-snapshot-loader" />,
}));

vi.mock("@/components/dashboard/DecisionHeader", () => ({
  default: () => <section data-testid="decision-header">DecisionHeader</section>,
}));

vi.mock("@/components/chart/KlineChart", () => ({
  default: () => <section data-testid="kline-chart">KlineChart</section>,
}));

vi.mock("@/components/dashboard/ExecutionPanel", () => ({
  default: () => <section data-testid="execution-panel">ExecutionPanel</section>,
}));

vi.mock("@/components/trading/PositionCalculator", () => ({
  default: () => <section data-testid="position-calculator-panel">PositionCalculator</section>,
}));

vi.mock("@/components/dashboard/EvidenceRail", () => ({
  default: () => <section data-testid="evidence-rail">EvidenceRail</section>,
}));

import DashboardPage from "./page";

describe("DashboardPage", () => {
  it("renders the overview band, primary workspace, side rail, and evidence strip", () => {
    render(<DashboardPage />);

    expect(screen.getByTestId("dashboard-command-page")).toBeInTheDocument();
    expect(screen.getByTestId("dashboard-overview-band")).toBeInTheDocument();
    expect(screen.getByTestId("dashboard-primary-workspace")).toBeInTheDocument();
    expect(screen.getByTestId("dashboard-support-stack")).toBeInTheDocument();
    expect(screen.getByTestId("dashboard-chart-surface")).toBeInTheDocument();
    expect(screen.getByTestId("dashboard-side-rail")).toBeInTheDocument();
    expect(screen.getByTestId("dashboard-evidence-surface")).toBeInTheDocument();

    expect(within(screen.getByTestId("dashboard-overview-band")).getByTestId("decision-header")).toBeInTheDocument();
    expect(within(screen.getByTestId("dashboard-chart-surface")).getByTestId("kline-chart")).toBeInTheDocument();
    expect(within(screen.getByTestId("dashboard-support-stack")).getByTestId("dashboard-side-rail")).toBeInTheDocument();
    expect(within(screen.getByTestId("dashboard-support-stack")).getByTestId("dashboard-evidence-surface")).toBeInTheDocument();
    expect(within(screen.getByTestId("dashboard-side-rail")).getByTestId("execution-panel")).toBeInTheDocument();
    expect(within(screen.getByTestId("dashboard-side-rail")).getByTestId("position-calculator-panel")).toBeInTheDocument();
    expect(within(screen.getByTestId("dashboard-evidence-surface")).getByTestId("evidence-rail")).toBeInTheDocument();
  });
});
