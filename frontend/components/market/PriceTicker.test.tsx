import React from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import PriceTicker from "@/components/market/PriceTicker";
import { useMarketStore } from "@/store/marketStore";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/store/marketStore", () => ({
  useMarketStore: vi.fn(),
}));

const mockedUseMarketStore = vi.mocked(useMarketStore);

describe("PriceTicker", () => {
  it("renders 4h interval button and triggers interval change", async () => {
    const refreshDashboard = vi.fn();
    const setSymbol = vi.fn();
    const setIntervalType = vi.fn();

    mockedUseMarketStore.mockReturnValue({
      symbol: "BTCUSDT",
      interval: "1m",
      price: { symbol: "BTCUSDT", price: 65234.12, time: 1741300000000 },
      loading: false,
      error: null,
      refreshDashboard,
      setSymbol,
      setIntervalType,
    } as ReturnType<typeof useMarketStore>);

    render(<PriceTicker />);

    const button = screen.getByRole("button", { name: "4h" });
    expect(button).toBeInTheDocument();

    await userEvent.click(button);
    expect(setIntervalType).toHaveBeenCalledWith("4h");
  });
});
