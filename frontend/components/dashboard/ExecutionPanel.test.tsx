import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import ExecutionPanel from "@/components/dashboard/ExecutionPanel";
import { useMarketStore } from "@/store/marketStore";
import { buildMockMarketStoreState } from "../../test/fixtures/marketStore";

vi.mock("@/store/marketStore", () => ({
  useMarketStore: vi.fn(),
}));

const mockedUseMarketStore = vi.mocked(useMarketStore);

describe("ExecutionPanel", () => {
  it("renders a ready execution setup with trade levels", () => {
    mockedUseMarketStore.mockReturnValue(
      buildMockMarketStoreState() as ReturnType<typeof useMarketStore>,
    );

    render(<ExecutionPanel />);

    expect(screen.getByRole("heading", { name: "Execution Setup" })).toBeInTheDocument();
    expect(screen.getByText("顺势做多")).toBeInTheDocument();
    expect(screen.getByText("Entry Zone")).toBeInTheDocument();
    expect(screen.getByText("Stop Loss")).toBeInTheDocument();
    expect(screen.getByText("Target")).toBeInTheDocument();
    expect(screen.getByText(/等待回踩/)).toBeInTheDocument();
  });

  it("shows a disabled waiting state when levels are invalid", () => {
    const state = buildMockMarketStoreState();
    mockedUseMarketStore.mockReturnValue(
      {
        ...state,
        signal: {
          ...state.signal,
          entry_price: Number.NaN,
          stop_loss: Number.NaN,
          target_price: Number.NaN,
        },
      } as ReturnType<typeof useMarketStore>,
    );

    render(<ExecutionPanel />);

    expect(screen.getByText("等待更完整的入场、止损和目标位后再做执行判断。")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "等待 setup 完整" })).toBeDisabled();
  });
});
