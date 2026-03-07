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

    expect(screen.getByText("Wall Map")).toBeInTheDocument();
    expect(screen.getByText("Ask Wall Map")).toBeInTheDocument();
    expect(screen.getByText("Bid Wall Map")).toBeInTheDocument();
    expect(screen.getByText("Near Ask Wall")).toBeInTheDocument();
    expect(screen.getByText("Mid Bid Wall")).toBeInTheDocument();
    expect(screen.getByText(/stop clusters/i)).toBeInTheDocument();
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
  });
});
