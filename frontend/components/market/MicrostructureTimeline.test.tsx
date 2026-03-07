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

    expect(screen.getByText("Microstructure Timeline")).toBeInTheDocument();
    expect(screen.getByText("Net Score")).toBeInTheDocument();
    expect(screen.getAllByText("High Order").length).toBeGreaterThan(0);
    expect(screen.getByText("Initiative Shift")).toBeInTheDocument();
    expect(screen.getByText("Microstructure Confluence")).toBeInTheDocument();
    expect(screen.getByText("Large Trade Cluster")).toBeInTheDocument();
    expect(screen.getByText(/卖压被持续吸收/)).toBeInTheDocument();
    expect(screen.getAllByText("+5").length).toBeGreaterThan(0);
  });

  it("filters timeline events by family", async () => {
    const user = userEvent.setup();
    mockedUseMarketStore.mockReturnValue(
      buildMockMarketStoreState() as ReturnType<typeof useMarketStore>,
    );

    render(<MicrostructureTimeline />);

    await user.click(screen.getAllByRole("button", { name: /Migration/i })[0]);

    expect(screen.getByText(/买方挂单墙连续多层上移/)).toBeInTheDocument();
    expect(screen.queryByText(/卖压被持续吸收/)).not.toBeInTheDocument();
    expect(screen.queryByText("Initiative Shift")).not.toBeInTheDocument();
    expect(screen.getByText("1 / 8 visible")).toBeInTheDocument();
  });
});
