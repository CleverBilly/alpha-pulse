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

    expect(screen.getByText("Kline Chart")).toBeInTheDocument();
    expect(screen.getAllByText("Signal Entry").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Buy Liquidity").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Int Support").length).toBeGreaterThan(0);
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
    expect(screen.getByText("Events")).toBeInTheDocument();
    expect(screen.getByText("Struct Tier")).toBeInTheDocument();
    expect(screen.getByText("uptrend")).toBeInTheDocument();
    expect(screen.getByText("external")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Initiative Shift" }));
    await user.click(screen.getByRole("button", { name: "Large Trade Cluster" }));
    await user.click(screen.getByRole("button", { name: "Reload / Exhaustion" }));
    await user.click(screen.getByRole("button", { name: "Failed Auction" }));
    await user.click(screen.getByRole("button", { name: "Order Book Migration" }));
    await user.click(screen.getByRole("button", { name: "Composite Patterns" }));
    await user.click(screen.getByRole("button", { name: "Microstructure Confluence" }));

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

    await user.hover(screen.getByRole("button", { name: "Micro ABS Absorption" }));
    expect(screen.getByText("Absorption · ABS")).toBeInTheDocument();
    expect(screen.getByText("BULLISH | score +5")).toBeInTheDocument();
    expect(screen.getAllByText("卖压被持续吸收，价格未继续下破").length).toBeGreaterThan(0);

    await user.click(screen.getByRole("button", { name: /EMA20/i }));
    expect(screen.getByText("Indicator Stack")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Micro ABS Absorption" }));
    expect(screen.getByText("Pinned")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "更新K线" }));
    expect(refreshDashboard).toHaveBeenCalledTimes(1);
    expect(refreshDashboard).toHaveBeenCalledWith(true);
  });
});
