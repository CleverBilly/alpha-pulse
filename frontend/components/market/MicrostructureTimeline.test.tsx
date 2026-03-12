import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import MicrostructureTimeline from "@/components/market/MicrostructureTimeline";
import { useMarketStore } from "@/store/marketStore";
import { buildMockMarketStoreState } from "../../test/fixtures/marketStore";

vi.mock("@/store/marketStore", () => ({
  useMarketStore: vi.fn(),
}));

const mockedUseMarketStore = vi.mocked(useMarketStore);

describe("MicrostructureTimeline", () => {
  it("renders recent microstructure events from snapshot data", () => {
    mockedUseMarketStore.mockReturnValue(
      buildMockMarketStoreState() as ReturnType<typeof useMarketStore>,
    );

    render(<MicrostructureTimeline />);

    expect(screen.getByText("微结构时间线")).toBeInTheDocument();
    expect(screen.getByText("净分")).toBeInTheDocument();
    expect(screen.getAllByText("高阶事件").length).toBeGreaterThan(0);
    expect(screen.getByText("主动性衰竭")).toBeInTheDocument();
    expect(screen.getByText("冰山回补")).toBeInTheDocument();
    expect(screen.getByText("微结构共振")).toBeInTheDocument();
    expect(screen.getByText("衰竭迁移反转")).toBeInTheDocument();
    expect(screen.getByText(/卖压被持续吸收/)).toBeInTheDocument();
    expect(screen.getAllByText("+5").length).toBeGreaterThan(0);
  });

  it("filters timeline events by family", async () => {
    const user = userEvent.setup();
    mockedUseMarketStore.mockReturnValue(
      buildMockMarketStoreState() as ReturnType<typeof useMarketStore>,
    );

    render(<MicrostructureTimeline />);

    await user.click(screen.getAllByRole("button", { name: /复合/ })[0]);

    expect(screen.getByText(/下方失败拍卖后挂单墙上移确认反转/)).toBeInTheDocument();
    expect(screen.getByText(/吸收与补单维持后出现买方延续推进/)).toBeInTheDocument();
    expect(screen.getByText(/挂单墙迁移与主动买盘同向推进/)).toBeInTheDocument();
    expect(screen.getByText(/主动卖盘耗尽后挂单墙上移确认反转/)).toBeInTheDocument();
    expect(screen.queryByText(/卖压被持续吸收/)).not.toBeInTheDocument();
    expect(screen.queryByText("主动成交衰竭")).not.toBeInTheDocument();
    expect(screen.getByText("可见 4 / 8")).toBeInTheDocument();
  });
});
