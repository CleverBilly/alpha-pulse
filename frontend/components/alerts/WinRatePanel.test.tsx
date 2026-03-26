import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import WinRatePanel from "./WinRatePanel";
import * as apiClient from "@/services/apiClient";

vi.mock("@/services/apiClient", () => ({
  alertApi: {
    getAlertStats: vi.fn().mockResolvedValue({
      symbol: "BTCUSDT",
      total: 10,
      target_hit: 6,
      stop_hit: 3,
      pending: 1,
      expired: 0,
      win_rate: 66.7,
      avg_rr: 1.8,
      sample_size_label: "近 10 条",
    }),
  },
}));

describe("WinRatePanel", () => {
  beforeEach(() => {
    vi.spyOn(window, "getComputedStyle").mockImplementation(
      () =>
        ({
          getPropertyValue: () => "",
        }) as CSSStyleDeclaration,
    );
    Object.defineProperty(window, "matchMedia", {
      writable: true,
      value: vi.fn().mockImplementation((query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders win rate percentage", async () => {
    render(<WinRatePanel symbols={["BTCUSDT"]} />);
    // Ant Design Statistic splits integer and decimal into separate spans;
    // verify both parts are present in the document.
    const intPart = await screen.findByText("66", { selector: ".ant-statistic-content-value-int" });
    const decPart = screen.getByText(".7", { selector: ".ant-statistic-content-value-decimal" });
    expect(intPart).toBeTruthy();
    expect(decPart).toBeTruthy();
  });
});
