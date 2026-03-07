import { render, screen } from "@testing-library/react";
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
    expect(screen.getByText("Absorption")).toBeInTheDocument();
    expect(screen.getByText("Initiative Shift")).toBeInTheDocument();
    expect(screen.getByText("Large Trade Cluster")).toBeInTheDocument();
    expect(screen.getByText(/卖压被持续吸收/)).toBeInTheDocument();
    expect(screen.getAllByText("+5").length).toBeGreaterThan(0);
  });
});
