import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import OrderFlowPanel from "@/components/orderflow/OrderFlowPanel";
import { useMarketStore } from "@/store/marketStore";
import { buildMockMarketStoreState } from "../../test/fixtures/marketStore";

vi.mock("@/store/marketStore", () => ({
  useMarketStore: vi.fn(),
}));

const mockedUseMarketStore = vi.mocked(useMarketStore);

describe("OrderFlowPanel", () => {
  it("renders large trades and microstructure event tape", () => {
    mockedUseMarketStore.mockReturnValue(
      buildMockMarketStoreState() as ReturnType<typeof useMarketStore>,
    );

    render(<OrderFlowPanel />);

    expect(screen.getByText("订单流")).toBeInTheDocument();
    expect(screen.getByText("微结构事件")).toBeInTheDocument();
    expect(screen.getAllByText("吸收").length).toBeGreaterThan(0);
    expect(screen.getByText(/冰山：买方冰山/)).toBeInTheDocument();
    expect(screen.getByText(/卖压被持续吸收/)).toBeInTheDocument();
    expect(screen.getByText(/主动卖盘耗尽后挂单墙上移确认反转/)).toBeInTheDocument();
  });
});
