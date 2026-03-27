import { render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/components/market/MarketSnapshotLoader", () => ({
  default: () => <div data-testid="market-snapshot-loader" />,
}));
vi.mock("@/components/layout/TradingWorkspaceHero", () => ({
  default: () => <section data-testid="trading-workspace-hero">TradingWorkspaceHero</section>,
}));
vi.mock("@/components/alerts/WinRatePanel", () => ({
  default: () => <section data-testid="win-rate-panel">WinRatePanel</section>,
}));
vi.mock("@/components/alerts/AlertHistoryBoard", () => ({
  default: () => <section data-testid="alert-history-board">AlertHistoryBoard</section>,
}));
vi.mock("@/components/signal/SignalCard", () => ({
  default: () => <section data-testid="signal-card">SignalCard</section>,
}));
vi.mock("@/components/analysis/AIAnalysisPanel", () => ({
  default: () => <section data-testid="ai-analysis-panel">AIAnalysisPanel</section>,
}));

import ReviewWorkspace from "./ReviewWorkspace";

describe("ReviewWorkspace", () => {
  it("renders overview, history strip, and live context workspace", () => {
    render(<ReviewWorkspace />);

    expect(screen.getByTestId("review-command-page")).toBeInTheDocument();
    expect(screen.getByTestId("review-overview-band")).toBeInTheDocument();
    expect(screen.getByTestId("review-history-strip")).toBeInTheDocument();
    expect(screen.getByTestId("review-context-workspace")).toBeInTheDocument();

    expect(within(screen.getByTestId("review-overview-band")).getByTestId("trading-workspace-hero")).toBeInTheDocument();
    expect(within(screen.getByTestId("review-history-strip")).getByTestId("win-rate-panel")).toBeInTheDocument();
    expect(within(screen.getByTestId("review-history-strip")).getByTestId("alert-history-board")).toBeInTheDocument();
    expect(within(screen.getByTestId("review-context-workspace")).getByTestId("signal-card")).toBeInTheDocument();
    expect(within(screen.getByTestId("review-context-workspace")).getByTestId("ai-analysis-panel")).toBeInTheDocument();
  });
});
