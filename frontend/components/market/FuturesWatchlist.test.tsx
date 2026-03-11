import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { useMarketStore } from "@/store/marketStore";
import { marketApi } from "@/services/apiClient";
import { buildMockMarketStoreState } from "@/test/fixtures/marketStore";
import { buildMockMarketSnapshot } from "@/test/fixtures/marketSnapshot";
import FuturesWatchlist from "./FuturesWatchlist";

vi.mock("@/store/marketStore", () => ({
  useMarketStore: vi.fn(),
}));

vi.mock("@/services/apiClient", () => ({
  marketApi: {
    getMarketSnapshot: vi.fn(),
  },
}));

const mockedUseMarketStore = vi.mocked(useMarketStore);
const mockedMarketApi = vi.mocked(marketApi);

describe("FuturesWatchlist", () => {
  it("renders BTC/ETH/SOL watchlist cards and allows switching symbol", async () => {
    const setSymbol = vi.fn();
    mockedUseMarketStore.mockReturnValue(
      buildMockMarketStoreState({
        setSymbol,
      }) as ReturnType<typeof useMarketStore>,
    );

    mockedMarketApi.getMarketSnapshot.mockImplementation(async (symbol) => {
      const snapshot = buildMockMarketSnapshot(symbol, "1h", 24);
      return snapshot;
    });

    render(<FuturesWatchlist />);

    await waitFor(() => {
      expect(screen.getByText("BTCUSDT")).toBeInTheDocument();
      expect(screen.getByText("ETHUSDT")).toBeInTheDocument();
      expect(screen.getByText("SOLUSDT")).toBeInTheDocument();
    });

    expect(mockedMarketApi.getMarketSnapshot).toHaveBeenCalledTimes(3);
    expect(mockedMarketApi.getMarketSnapshot).toHaveBeenNthCalledWith(1, "BTCUSDT", "1h", 24);
    expect(mockedMarketApi.getMarketSnapshot).toHaveBeenNthCalledWith(2, "ETHUSDT", "1h", 24);
    expect(mockedMarketApi.getMarketSnapshot).toHaveBeenNthCalledWith(3, "SOLUSDT", "1h", 24);

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "切换到 SOLUSDT" }));
    expect(setSymbol).toHaveBeenCalledWith("SOLUSDT");
  });

  it("shows unavailable futures state when futures metrics are missing", async () => {
    mockedUseMarketStore.mockReturnValue(buildMockMarketStoreState() as ReturnType<typeof useMarketStore>);
    mockedMarketApi.getMarketSnapshot.mockImplementation(async (symbol) => {
      const snapshot = buildMockMarketSnapshot(symbol, "1h", 24);
      if (symbol === "ETHUSDT") {
        snapshot.futures = {
          ...snapshot.futures,
          available: false,
          reason: "Futures metrics unavailable",
        };
      }
      return snapshot;
    });

    render(<FuturesWatchlist />);

    await waitFor(() => {
      expect(screen.getByText("ETHUSDT")).toBeInTheDocument();
    });

    expect(screen.getByText("Futures metrics unavailable")).toBeInTheDocument();
  });
});
