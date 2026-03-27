import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { alertApi } from "@/services/apiClient";
import { useMarketStore } from "@/store/marketStore";
import AlertHistoryBoard from "./AlertHistoryBoard";

vi.mock("@/services/apiClient", () => ({
  alertApi: {
    getAlerts: vi.fn(),
    getAlertHistory: vi.fn(),
    getAlertPreferences: vi.fn(),
    updateAlertPreferences: vi.fn(),
    refreshAlerts: vi.fn(),
  },
}));

const mockedAlertApi = vi.mocked(alertApi);

describe("AlertHistoryBoard", () => {
  beforeEach(() => {
    useMarketStore.setState({ symbol: "BTCUSDT" });
    vi.spyOn(window, "getComputedStyle").mockImplementation(
      () =>
        ({
          getPropertyValue: () => "",
        }) as unknown as CSSStyleDeclaration,
    );
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("loads history and filters by current symbol", async () => {
    mockedAlertApi.getAlertHistory.mockResolvedValue({
      items: [buildMockAlert("BTCUSDT"), buildMockAlert("ETHUSDT")],
      generated: 0,
    });
    mockedAlertApi.refreshAlerts.mockResolvedValue({
      items: [buildMockAlert("BTCUSDT")],
      generated: 1,
    });

    render(<AlertHistoryBoard />);

    await waitFor(() => {
      expect(mockedAlertApi.getAlertHistory).toHaveBeenCalledWith(60);
    });

    expect(screen.getByTestId("alert-history-rail")).toBeInTheDocument();
    expect(screen.getByTestId("alert-history-summary")).toBeInTheDocument();
    expect(screen.getByText("BTCUSDT A 级机会已就绪")).toBeInTheDocument();
    expect(screen.queryByText("ETHUSDT A 级机会已就绪")).not.toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "全部历史" }));
    expect(screen.getByText("ETHUSDT A 级机会已就绪")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /重新评估告警/ }));
    await waitFor(() => {
      expect(mockedAlertApi.refreshAlerts).toHaveBeenCalledWith(20);
    });
  });
});

function buildMockAlert(symbol: string) {
  return {
    id: `${symbol}-alert-1`,
    symbol,
    kind: "setup_ready",
    severity: "A",
    title: `${symbol} A 级机会已就绪`,
    verdict: "强偏多",
    tradeability_label: "A 级可跟踪",
    summary: "4h 与 1h 已经对齐，15m 触发也站在同一边。",
    reasons: ["趋势因子主导当前方向。", "订单流与结构保持一致。"],
    timeframe_labels: ["4h 强偏多", "1h 强偏多", "15m 强偏多"],
    confidence: 74,
    risk_label: "可控风险",
    entry_price: 65200,
    stop_loss: 64880,
    target_price: 65880,
    risk_reward: 2.1,
    created_at: 1710000000000,
    deliveries: [],
  };
}
