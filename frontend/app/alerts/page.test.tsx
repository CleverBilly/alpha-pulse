import { render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/components/alerts/AlertCenter", () => ({
  default: () => <section data-testid="alert-center">AlertCenter</section>,
}));
vi.mock("@/components/alerts/AlertHistoryBoard", () => ({
  default: () => <section data-testid="alert-history-board">AlertHistoryBoard</section>,
}));

import AlertsPage from "./page";

describe("AlertsPage", () => {
  it("renders overview, control surface, and history workspace", () => {
    render(<AlertsPage />);

    expect(screen.getByTestId("alerts-command-page")).toBeInTheDocument();
    expect(screen.getByTestId("alerts-overview-band")).toBeInTheDocument();
    expect(screen.getByTestId("alerts-status-band")).toBeInTheDocument();
    expect(screen.getByTestId("alerts-watch-desk")).toBeInTheDocument();
    expect(screen.getByTestId("alerts-history-rail")).toBeInTheDocument();

    expect(within(screen.getByTestId("alerts-watch-desk")).getByTestId("alert-center")).toBeInTheDocument();
    expect(within(screen.getByTestId("alerts-history-rail")).getByTestId("alert-history-board")).toBeInTheDocument();
  });
});
