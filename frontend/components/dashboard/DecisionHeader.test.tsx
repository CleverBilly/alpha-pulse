import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import DecisionHeader from "@/components/dashboard/DecisionHeader";
import { useMarketStore } from "@/store/marketStore";
import { buildMockMarketStoreState } from "../../test/fixtures/marketStore";

vi.mock("@/store/marketStore", () => ({
  useMarketStore: vi.fn(),
}));

const mockedUseMarketStore = vi.mocked(useMarketStore);

describe("DecisionHeader", () => {
  it("renders verdict, risk and top reasons", () => {
    mockedUseMarketStore.mockReturnValue(
      buildMockMarketStoreState() as ReturnType<typeof useMarketStore>,
    );

    render(<DecisionHeader />);

    expect(screen.getByText("当前判断")).toBeInTheDocument();
    expect(screen.getAllByText("强偏多")).toHaveLength(2);
    expect(screen.getByText("可控风险")).toBeInTheDocument();
    expect(screen.getByText("EMA20 高于 EMA50，价格位于 VWAP 上方")).toBeInTheDocument();
    expect(screen.getByText("Delta 为正，大单净流入持续扩张")).toBeInTheDocument();
  });

  it("triggers refresh and symbol or interval changes", async () => {
    const user = userEvent.setup();
    const refreshDashboard = vi.fn().mockResolvedValue(undefined);
    const setSymbol = vi.fn();
    const setIntervalType = vi.fn();

    mockedUseMarketStore.mockReturnValue(
      buildMockMarketStoreState({
        refreshDashboard,
        setSymbol,
        setIntervalType,
      }) as ReturnType<typeof useMarketStore>,
    );

    render(<DecisionHeader />);

    await user.click(screen.getByRole("button", { name: "刷新" }));
    expect(refreshDashboard).toHaveBeenCalledWith(true);

    await user.selectOptions(screen.getByLabelText("Symbol"), "ETHUSDT");
    expect(setSymbol).toHaveBeenCalledWith("ETHUSDT");

    await user.click(screen.getByRole("button", { name: "4h" }));
    expect(setIntervalType).toHaveBeenCalledWith("4h");
  });
});
