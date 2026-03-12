import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import KlineChart from "@/components/chart/KlineChart";
import { useMarketStore } from "@/store/marketStore";
import { buildMockMarketStoreState } from "../../test/fixtures/marketStore";

vi.mock("@/store/marketStore", () => ({
  useMarketStore: vi.fn(),
}));

const mockedUseMarketStore = vi.mocked(useMarketStore);

describe("KlineChart", () => {
  it("renders chart overlays, toggleable microstructure layers and tooltip details", async () => {
    const refreshDashboard = vi.fn();
    const user = userEvent.setup();

    mockedUseMarketStore.mockReturnValue({
      ...buildMockMarketStoreState({ refreshDashboard }),
    } as ReturnType<typeof useMarketStore>);

    render(<KlineChart />);

    expect(screen.getByText("K 线图")).toBeInTheDocument();
    expect(screen.getAllByText("信号进场").length).toBeGreaterThan(0);
    expect(screen.getAllByText("买方流动性").length).toBeGreaterThan(0);
    expect(screen.getAllByText("内部支撑").length).toBeGreaterThan(0);
    expect(screen.getByText("HH")).toBeInTheDocument();
    expect(screen.getByText("iHL")).toBeInTheDocument();
    expect(screen.getByText("BOS")).toBeInTheDocument();
    expect(screen.getByText("ABS")).toBeInTheDocument();
    expect(screen.getByText("ICE")).toBeInTheDocument();
    expect(screen.getByText("AGR")).toBeInTheDocument();
    expect(screen.queryByText("SHF")).not.toBeInTheDocument();
    expect(screen.queryByText("LTC")).not.toBeInTheDocument();
    expect(screen.queryByText("IRL")).not.toBeInTheDocument();
    expect(screen.queryByText("IEX")).not.toBeInTheDocument();
    expect(screen.queryByText("FAH")).not.toBeInTheDocument();
    expect(screen.queryByText("OBL")).not.toBeInTheDocument();
    expect(screen.queryByText("TRP")).not.toBeInTheDocument();
    expect(screen.queryByText("LLB")).not.toBeInTheDocument();
    expect(screen.queryByText("MAF")).not.toBeInTheDocument();
    expect(screen.queryByText("ARC")).not.toBeInTheDocument();
    expect(screen.queryByText("EMR")).not.toBeInTheDocument();
    expect(screen.queryByText("MCF")).not.toBeInTheDocument();
    expect(screen.getByText("结构事件")).toBeInTheDocument();
    expect(screen.getByText("结构层级")).toBeInTheDocument();
    expect(screen.getByText("上升趋势")).toBeInTheDocument();
    expect(screen.getByText("外部")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "主动性切换" }));
    await user.click(screen.getByRole("button", { name: "大单簇" }));
    await user.click(screen.getByRole("button", { name: "回补 / 衰竭" }));
    await user.click(screen.getByRole("button", { name: "失败拍卖" }));
    await user.click(screen.getByRole("button", { name: "订单簿迁移" }));
    await user.click(screen.getByRole("button", { name: "复合形态" }));
    await user.click(screen.getByRole("button", { name: "微结构共振" }));

    expect(screen.getByText("SHF")).toBeInTheDocument();
    expect(screen.getByText("LTC")).toBeInTheDocument();
    expect(screen.getByText("IRL")).toBeInTheDocument();
    expect(screen.getByText("IEX")).toBeInTheDocument();
    expect(screen.getByText("FAH")).toBeInTheDocument();
    expect(screen.getByText("OBL")).toBeInTheDocument();
    expect(screen.getByText("TRP")).toBeInTheDocument();
    expect(screen.getByText("LLB")).toBeInTheDocument();
    expect(screen.getByText("MAF")).toBeInTheDocument();
    expect(screen.getByText("ARC")).toBeInTheDocument();
    expect(screen.getByText("EMR")).toBeInTheDocument();
    expect(screen.getByText("MCF")).toBeInTheDocument();

    await user.hover(screen.getByRole("button", { name: "微结构 ABS 吸收" }));
    expect(screen.getByText("吸收 · ABS")).toBeInTheDocument();
    expect(screen.getByText("多头 | 评分 +5")).toBeInTheDocument();
    expect(screen.getAllByText("卖压被持续吸收，价格未继续下破").length).toBeGreaterThan(0);

    await user.click(screen.getByRole("button", { name: /EMA20/i }));
    expect(screen.getByText("指标图层")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "微结构 ABS 吸收" }));
    expect(screen.getByText("已固定")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "更新K线" }));
    expect(refreshDashboard).toHaveBeenCalledTimes(1);
    expect(refreshDashboard).toHaveBeenCalledWith(true);
  });
});
