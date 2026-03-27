import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import TradeOrderPanel from "./TradeOrderPanel";

describe("TradeOrderPanel", () => {
  it("groups orders by lifecycle state and only exposes close on open positions", () => {
    render(
      <TradeOrderPanel
        orders={[
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
          {
            id: 2,
            alert_id: "alert-2",
            symbol: "ETHUSDT",
            side: "SHORT",
            requested_qty: 0.2,
            filled_qty: 0.2,
            limit_price: 3200,
            entry_status: "filled",
            status: "open",
            source: "system",
            created_at: 1710000001000,
            closed_at: 0,
          },
          {
            id: 3,
            alert_id: "alert-3",
            symbol: "BTCUSDT",
            side: "LONG",
            requested_qty: 0.01,
            filled_qty: 0.01,
            limit_price: 65000,
            entry_status: "filled",
            status: "failed",
            source: "system",
            close_reason: "protective order failed",
            created_at: 1710000002000,
            closed_at: 1710000003000,
          },
          {
            id: 4,
            alert_id: "alert-4",
            symbol: "BTCUSDT",
            side: "LONG",
            requested_qty: 0.01,
            filled_qty: 0.01,
            limit_price: 65000,
            entry_status: "filled",
            status: "closed",
            source: "manual",
            created_at: 1710000004000,
            closed_at: 1710000005000,
          },
        ]}
        onClose={vi.fn()}
      />,
    );

    expect(screen.getByText("挂单中")).toBeInTheDocument();
    expect(screen.getByText("持仓中")).toBeInTheDocument();
    expect(screen.getByText("失败记录")).toBeInTheDocument();
    expect(screen.getByText("已收口")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "手动平仓" })).toBeInTheDocument();
    expect(screen.getByText("protective order failed")).toBeInTheDocument();
  });
});
