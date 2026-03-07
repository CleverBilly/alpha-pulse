import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import SignalCard from "@/components/signal/SignalCard";
import { useMarketStore } from "@/store/marketStore";
import { buildMockMarketStoreState } from "../../test/fixtures/marketStore";

vi.mock("@/store/marketStore", () => ({
  useMarketStore: vi.fn(),
}));

const mockedUseMarketStore = vi.mocked(useMarketStore);

describe("SignalCard", () => {
  it("renders signal metrics and factor breakdown", async () => {
    const refreshDashboard = vi.fn();

    mockedUseMarketStore.mockReturnValue({
      ...buildMockMarketStoreState({ refreshDashboard }),
    } as ReturnType<typeof useMarketStore>);

    render(<SignalCard />);

    expect(screen.getByText("Signal")).toBeInTheDocument();
    expect(screen.getByText("BUY")).toBeInTheDocument();
    expect(screen.getByText("BTCUSDT")).toBeInTheDocument();
    expect(screen.getByText("当前多头信号由趋势、订单流与微结构事件序列共振驱动。")).toBeInTheDocument();
    expect(screen.getByText("Factor Breakdown")).toBeInTheDocument();
    expect(screen.getByText("Microstructure")).toBeInTheDocument();
    expect(screen.getByText("最近微结构事件连续偏多，买方主动性增强")).toBeInTheDocument();
    expect(screen.getByText("74%")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "刷新信号" }));
    expect(refreshDashboard).toHaveBeenCalledTimes(1);
  });
});
