import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import LiquidityPanel from "@/components/liquidity/LiquidityPanel";
import { useMarketStore } from "@/store/marketStore";
import { buildMockMarketStoreState } from "../../test/fixtures/marketStore";

vi.mock("@/store/marketStore", () => ({
  useMarketStore: vi.fn(),
}));

const mockedUseMarketStore = vi.mocked(useMarketStore);

describe("LiquidityPanel", () => {
  it("renders bid and ask wall maps from liquidity snapshot", () => {
    mockedUseMarketStore.mockReturnValue(
      buildMockMarketStoreState() as ReturnType<typeof useMarketStore>,
    );

    render(<LiquidityPanel />);

    expect(screen.getByRole("heading", { name: "流动性" })).toBeInTheDocument();
    expect(screen.getByText("墙位分布")).toBeInTheDocument();
    expect(screen.getByText("卖墙分布")).toBeInTheDocument();
    expect(screen.getByText("买墙分布")).toBeInTheDocument();
    expect(screen.getByText("卖墙热度带")).toBeInTheDocument();
    expect(screen.getByText("跨周期墙位演化")).toBeInTheDocument();
    expect(screen.getByText("近端卖墙")).toBeInTheDocument();
    expect(screen.getByText("中段买墙")).toBeInTheDocument();
    expect(screen.getAllByText("0-10bps")).toHaveLength(2);
    expect(screen.getByText("买盘主导")).toBeInTheDocument();
    expect(screen.getByText(/止损簇/)).toBeInTheDocument();
  });

  it("triggers refresh from the panel action", async () => {
    const user = userEvent.setup();
    const refreshDashboard = vi.fn().mockResolvedValue(undefined);
    mockedUseMarketStore.mockReturnValue(
      buildMockMarketStoreState({
        refreshDashboard,
      }) as ReturnType<typeof useMarketStore>,
    );

    render(<LiquidityPanel />);

    await user.click(screen.getByRole("button", { name: "更新" }));

    expect(refreshDashboard).toHaveBeenCalledTimes(1);
    expect(refreshDashboard).toHaveBeenCalledWith(true);
  });
});
