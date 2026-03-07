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

    expect(screen.getByText("Order Flow")).toBeInTheDocument();
    expect(screen.getByText("Microstructure Events")).toBeInTheDocument();
    expect(screen.getAllByText("Absorption").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Iceberg").length).toBeGreaterThan(0);
    expect(screen.getByText(/卖压被持续吸收/)).toBeInTheDocument();
    expect(screen.getByText(/连续卖方大单被市场吸收/)).toBeInTheDocument();
  });
});
