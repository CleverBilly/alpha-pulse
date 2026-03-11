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
  it("renders BTC/ETH/SOL multi-timeframe cards and allows switching symbol", async () => {
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

    expect(mockedMarketApi.getMarketSnapshot).toHaveBeenCalledTimes(9);
    expect(mockedMarketApi.getMarketSnapshot.mock.calls).toEqual(
      expect.arrayContaining([
        ["BTCUSDT", "4h", 24],
        ["BTCUSDT", "1h", 24],
        ["BTCUSDT", "15m", 24],
        ["ETHUSDT", "4h", 24],
        ["ETHUSDT", "1h", 24],
        ["ETHUSDT", "15m", 24],
        ["SOLUSDT", "4h", 24],
        ["SOLUSDT", "1h", 24],
        ["SOLUSDT", "15m", 24],
      ]),
    );
    expect(screen.getAllByText("A 级可跟踪").length).toBeGreaterThan(0);
    expect(screen.getAllByText("4h 强偏多").length).toBeGreaterThan(0);
    expect(screen.getAllByText("1h 强偏多").length).toBeGreaterThan(0);

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "切换到 SOLUSDT" }));
    expect(setSymbol).toHaveBeenCalledWith("SOLUSDT");
  });

  it("shows no-trade when higher timeframe conflicts with the 1h bias", async () => {
    mockedUseMarketStore.mockReturnValue(buildMockMarketStoreState() as ReturnType<typeof useMarketStore>);

    mockedMarketApi.getMarketSnapshot.mockImplementation(async (symbol, interval) => {
      const snapshot = buildMockMarketSnapshot(symbol, interval, 24);
      if (symbol === "ETHUSDT" && interval === "4h") {
        snapshot.signal.signal = "SELL";
        snapshot.signal.score = -62;
        snapshot.signal.confidence = 71;
        snapshot.signal.trend_bias = "bearish";
        snapshot.signal.explain = "4h 仍在下行段，尚未完成反转";
        snapshot.signal.factors = snapshot.signal.factors.map((factor, index) => ({
          ...factor,
          bias: "bearish",
          score: index === 0 ? -16 : -Math.abs(factor.score),
          reason: index === 0 ? "4h 高一级别仍在压制反弹" : factor.reason,
        }));
        snapshot.structure.trend = "downtrend";
      }
      return snapshot;
    });

    render(<FuturesWatchlist />);

    await waitFor(() => {
      expect(screen.getByText("ETHUSDT")).toBeInTheDocument();
    });

    expect(screen.getByText("当前禁止交易")).toBeInTheDocument();
    expect(screen.getByText("4h 与 1h 方向互相打架，当前属于逆大级别风险区。")).toBeInTheDocument();
    expect(screen.getAllByText("No-Trade").length).toBeGreaterThan(0);
  });
});
