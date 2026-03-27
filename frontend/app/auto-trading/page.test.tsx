import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import AutoTradingPage from "./page";

describe("AutoTradingPage", () => {
  it("renders a structured command page instead of an inline placeholder", () => {
    render(<AutoTradingPage />);

    expect(screen.getByTestId("auto-trading-command-page")).toBeInTheDocument();
    expect(screen.getByTestId("auto-trading-overview-band")).toBeInTheDocument();
    expect(screen.getByTestId("auto-trading-workspace")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "自动交易指挥台" })).toBeInTheDocument();
    expect(screen.getByText("策略编排")).toBeInTheDocument();
    expect(screen.getByText("风险护栏")).toBeInTheDocument();
  });
});
