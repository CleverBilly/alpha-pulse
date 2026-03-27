import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import AutoTradingPage from "./page";

vi.mock("@/services/apiClient", () => ({
  tradeApi: {
    getSettings: vi.fn().mockResolvedValue({
      trade_enabled_env: true,
      trade_auto_execute_env: true,
      auto_execute_enabled: true,
      allowed_symbols: ["BTCUSDT", "ETHUSDT"],
      risk_pct: 2.5,
      min_risk_reward: 1.6,
      entry_timeout_seconds: 45,
      max_open_positions: 2,
      sync_enabled: true,
      updated_by: "tester",
    }),
    getRuntime: vi.fn().mockResolvedValue({
      trade_enabled_env: true,
      trade_auto_execute_env: true,
      pending_fill_count: 1,
      open_count: 2,
    }),
    list: vi.fn().mockResolvedValue([
      {
        id: 1,
        alert_id: "alert-1",
        symbol: "BTCUSDT",
        side: "LONG",
        requested_qty: 0.01,
        filled_qty: 0,
        limit_price: 65000,
        entry_status: "pending_fill",
        status: "pending_fill",
        source: "system",
        created_at: 1710000000000,
        closed_at: 0,
      },
    ]),
    updateSettings: vi.fn().mockResolvedValue({
      trade_enabled_env: true,
      trade_auto_execute_env: true,
      auto_execute_enabled: true,
      allowed_symbols: ["BTCUSDT"],
      risk_pct: 2,
      min_risk_reward: 1.2,
      entry_timeout_seconds: 45,
      max_open_positions: 1,
      sync_enabled: true,
      updated_by: "tester",
    }),
    close: vi.fn(),
  },
}));

describe("AutoTradingPage", () => {
  it("renders the live control page with runtime band, config panel, and order board", async () => {
    render(<AutoTradingPage />);

    expect(screen.getByTestId("auto-trading-command-page")).toBeInTheDocument();
    expect(screen.getByTestId("auto-trading-overview-band")).toBeInTheDocument();
    expect(screen.getByTestId("auto-trading-workspace")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "自动交易指挥台" })).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByTestId("trade-runtime-band")).toBeInTheDocument();
    });

    expect(screen.getByTestId("trade-settings-panel")).toBeInTheDocument();
    expect(screen.getByTestId("trade-order-panel")).toBeInTheDocument();
    expect(screen.getByText("系统交易权限")).toBeInTheDocument();
    expect(screen.getByText("允许标的")).toBeInTheDocument();
  });
});
