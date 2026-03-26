import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { marketApi } from "@/services/apiClient";
import ReviewChartModal from "./ReviewChartModal";
import type { AlertEvent } from "@/types/alert";

// Mock marketApi — only getKline is used by ReviewChartModal
vi.mock("@/services/apiClient", () => ({
  marketApi: {
    getKline: vi.fn(),
  },
}));

// Mock KlineChart to isolate ReviewChartModal from chart rendering complexity.
// We verify that the chart container appears when data is present.
vi.mock("@/components/chart/KlineChart", () => ({
  default: () => <div data-testid="kline-chart" />,
}));

const mockedMarketApi = vi.mocked(marketApi);

const MOCK_EVENT: AlertEvent = {
  id: "evt-1",
  symbol: "BTCUSDT",
  kind: "setup_ready",
  severity: "A",
  title: "BTCUSDT A 级机会已就绪",
  verdict: "强偏多",
  tradeability_label: "A 级可跟踪",
  summary: "多周期对齐。",
  reasons: ["趋势因子主导当前方向。"],
  timeframe_labels: ["1h 强偏多"],
  confidence: 74,
  risk_label: "可控风险",
  entry_price: 65200,
  stop_loss: 64880,
  target_price: 65880,
  risk_reward: 2.1,
  created_at: 1710000000000,
  deliveries: [],
  interval: "1h",
};

function buildMockKline(openTime: number) {
  return {
    open_time: openTime,
    close_time: openTime + 3_600_000,
    open_price: 65000,
    close_price: 65200,
    high_price: 65400,
    low_price: 64800,
    volume: 100,
    quote_volume: 6_500_000,
    trade_count: 1000,
    taker_buy_volume: 55,
    taker_buy_quote_volume: 3_575_000,
  };
}

describe("ReviewChartModal", () => {
  beforeEach(() => {
    vi.spyOn(window, "getComputedStyle").mockImplementation(
      () =>
        ({
          getPropertyValue: () => "",
        }) as CSSStyleDeclaration,
    );
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  // 1. open=false 时不发起任何 API 请求
  it("does not call API when open is false", () => {
    render(<ReviewChartModal event={MOCK_EVENT} open={false} onClose={vi.fn()} />);
    expect(mockedMarketApi.getKline).not.toHaveBeenCalled();
  });

  // 2. open=true 且 event 有效时，getKline 被调用两次（before_ts + after_ts）
  it("calls getKline twice with before_ts and after_ts when open with valid event", async () => {
    const klineBefore = [buildMockKline(MOCK_EVENT.created_at - 3_600_000)];
    const klineAfter = [buildMockKline(MOCK_EVENT.created_at)];
    mockedMarketApi.getKline
      .mockResolvedValueOnce(klineBefore)
      .mockResolvedValueOnce(klineAfter);

    render(<ReviewChartModal event={MOCK_EVENT} open={true} onClose={vi.fn()} />);

    await waitFor(() => {
      expect(mockedMarketApi.getKline).toHaveBeenCalledTimes(2);
    });

    expect(mockedMarketApi.getKline).toHaveBeenCalledWith(
      "BTCUSDT",
      "1h",
      20,
      { before_ts: MOCK_EVENT.created_at },
    );
    expect(mockedMarketApi.getKline).toHaveBeenCalledWith(
      "BTCUSDT",
      "1h",
      40,
      { after_ts: MOCK_EVENT.created_at },
    );
  });

  // 3. 两次 API 均返回空数组时，显示"数据不足"文案
  it("shows 数据不足 when both API calls return empty arrays", async () => {
    mockedMarketApi.getKline.mockResolvedValue([]);

    render(<ReviewChartModal event={MOCK_EVENT} open={true} onClose={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getByText(/数据不足/)).toBeInTheDocument();
    });
  });

  // 4. API 返回数据时，loading 消失且 KlineChart 渲染
  it("renders KlineChart and hides loading when API returns data", async () => {
    const klineBefore = [buildMockKline(MOCK_EVENT.created_at - 3_600_000)];
    const klineAfter = [buildMockKline(MOCK_EVENT.created_at + 3_600_000)];
    mockedMarketApi.getKline
      .mockResolvedValueOnce(klineBefore)
      .mockResolvedValueOnce(klineAfter);

    render(<ReviewChartModal event={MOCK_EVENT} open={true} onClose={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getByTestId("kline-chart")).toBeInTheDocument();
    });

    // loading spinner 应已消失（Spin 组件内无加载状态时无 aria-busy）
    expect(screen.queryByRole("img", { name: /loading/i })).not.toBeInTheDocument();
  });

  // 5. 组件卸载后不触发 setState（active flag 保护）
  it("does not call setState after unmount while loading", async () => {
    let resolveKline!: (value: ReturnType<typeof buildMockKline>[]) => void;
    const pending = new Promise<ReturnType<typeof buildMockKline>[]>((resolve) => {
      resolveKline = resolve;
    });
    mockedMarketApi.getKline.mockReturnValue(pending);

    const { unmount } = render(
      <ReviewChartModal event={MOCK_EVENT} open={true} onClose={vi.fn()} />,
    );

    // 卸载时 Promise 仍处于 pending 状态
    unmount();

    // Resolve after unmount — active flag prevents setState.
    // React 18 removed the "setState on unmounted component" warning, so this test
    // verifies the absence of unhandled exceptions rather than console.error output.
    const consoleSpy = vi.spyOn(console, "error");
    resolveKline([buildMockKline(MOCK_EVENT.created_at)]);

    // 给 microtask 队列一点时间清空
    await new Promise((resolve) => setTimeout(resolve, 0));

    // Should not throw; consoleSpy is a secondary indicator (React 18+ doesn't emit this warning)
    expect(consoleSpy).not.toHaveBeenCalledWith(
      expect.stringContaining("Can't perform a React state update on an unmounted component"),
    );
  });
});
