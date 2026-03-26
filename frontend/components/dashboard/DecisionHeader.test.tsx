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
    expect(screen.getAllByText("强偏多").length).toBeGreaterThan(0);
    expect(screen.getAllByText("A 级可跟踪").length).toBeGreaterThan(0);
    expect(screen.getByText("可控风险")).toBeInTheDocument();
    expect(screen.getByText("4h 强偏多")).toBeInTheDocument();
    expect(screen.getByText("1h 强偏多")).toBeInTheDocument();
    expect(screen.getByText("5m 强偏多")).toBeInTheDocument();
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

    await user.selectOptions(screen.getByLabelText("标的"), "ETHUSDT");
    expect(setSymbol).toHaveBeenCalledWith("ETHUSDT");

    await user.click(screen.getByRole("button", { name: "4h" }));
    expect(setIntervalType).toHaveBeenCalledWith("4h");
  });

  it("exposes dedicated quote and workspace regions for the cockpit layout", () => {
    mockedUseMarketStore.mockReturnValue(
      buildMockMarketStoreState() as ReturnType<typeof useMarketStore>,
    );

    render(<DecisionHeader />);

    expect(screen.getByRole("region", { name: "市场报价" })).toBeInTheDocument();
    expect(screen.getByRole("region", { name: "交易工作台控件" })).toBeInTheDocument();
  });
});
