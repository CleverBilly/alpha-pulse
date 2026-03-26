import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import WinRatePanel from "./WinRatePanel";
import { alertApi } from "@/services/apiClient";

vi.mock("@/services/apiClient", () => ({
  alertApi: {
    getAlertStats: vi.fn(),
  },
}));

const mockStats = {
  symbol: "BTCUSDT",
  total: 10,
  target_hit: 6,
  stop_hit: 3,
  pending: 1,
  expired: 0,
  win_rate: 66.7,
  avg_rr: 1.8,
  sample_size_label: "近 10 条",
};

describe("WinRatePanel", () => {
  beforeEach(() => {
    vi.mocked(alertApi.getAlertStats).mockResolvedValue(mockStats);

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

  it("shows error message when API call fails", async () => {
    vi.mocked(alertApi.getAlertStats).mockRejectedValueOnce(new Error("network"));
    render(<WinRatePanel symbols={["BTCUSDT"]} />);
    await waitFor(() => {
      expect(screen.getByText(/加载失败/)).toBeInTheDocument();
    });
  });

  it("refetches data when limit button is clicked", async () => {
    render(<WinRatePanel symbols={["BTCUSDT"]} />);
    await waitFor(() =>
      expect(alertApi.getAlertStats).toHaveBeenCalledWith("BTCUSDT", 50),
    );

    fireEvent.click(screen.getByRole("button", { name: "近20条" }));
    await waitFor(() =>
      expect(alertApi.getAlertStats).toHaveBeenCalledWith("BTCUSDT", 20),
    );
  });

  it("clears error message on subsequent successful fetch", async () => {
    vi.mocked(alertApi.getAlertStats)
      .mockRejectedValueOnce(new Error("network"))
      .mockResolvedValue(mockStats);

    render(<WinRatePanel symbols={["BTCUSDT"]} />);
    await waitFor(() => expect(screen.getByText(/加载失败/)).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "近20条" }));
    await waitFor(() =>
      expect(screen.queryByText(/加载失败/)).not.toBeInTheDocument(),
    );
  });
});
