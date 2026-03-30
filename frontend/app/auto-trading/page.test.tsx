import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import AutoTradingPage from "./page";

const initialSettings = {
  trade_enabled_env: true,
  trade_auto_execute_env: true,
  allowed_symbols_env: ["BTCUSDT", "ETHUSDT", "SOLUSDT"],
  auto_execute_enabled: true,
  allowed_symbols: ["BTCUSDT", "ETHUSDT"],
  risk_pct: 2.5,
  min_risk_reward: 1.6,
  entry_timeout_seconds: 45,
  max_open_positions: 2,
  sync_enabled: true,
  updated_by: "tester",
};

const initialRuntime = {
  trade_enabled_env: true,
  trade_auto_execute_env: true,
  pending_fill_count: 1,
  open_count: 2,
};

const initialOrders = [
  {
    id: 1,
    alert_id: "alert-1",
    symbol: "BTCUSDT",
    side: "LONG",
    requested_qty: 0.01,
    filled_qty: 0,
    limit_price: 65000,
    entry_status: "pending_fill",
    status: "pending_fill",
    source: "system",
    created_at: 1710000000000,
    closed_at: 0,
  },
];

const savedSettings = {
  ...initialSettings,
  allowed_symbols: ["BTCUSDT"],
  risk_pct: 3.2,
  min_risk_reward: 1.8,
  entry_timeout_seconds: 61,
  max_open_positions: 3,
  sync_enabled: false,
  updated_by: "fresh-save",
};

const { mockTradeApi } = vi.hoisted(() => ({
  mockTradeApi: {
    getSettings: vi.fn(),
    getRuntime: vi.fn(),
    list: vi.fn(),
    updateSettings: vi.fn(),
    close: vi.fn(),
  },
}));

vi.mock("@/services/apiClient", () => ({
  tradeApi: mockTradeApi,
}));

describe("AutoTradingPage", () => {
  beforeEach(() => {
    mockTradeApi.getSettings.mockReset().mockResolvedValue(initialSettings);
    mockTradeApi.getRuntime.mockReset().mockResolvedValue(initialRuntime);
    mockTradeApi.list.mockReset().mockResolvedValue(initialOrders);
    mockTradeApi.updateSettings.mockReset().mockResolvedValue(savedSettings);
    mockTradeApi.close.mockReset();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it("renders the live control page with runtime band, config panel, and order board", async () => {
    render(<AutoTradingPage />);

    expect(screen.getByTestId("auto-trading-command-page")).toBeInTheDocument();
    expect(screen.getByTestId("auto-trading-overview-band")).toBeInTheDocument();
    expect(screen.getByTestId("auto-trading-workspace")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "自动交易指挥台" })).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByTestId("trade-runtime-band")).toBeInTheDocument();
    });

    expect(screen.getByTestId("trade-settings-panel")).toBeInTheDocument();
    expect(screen.getByTestId("trade-order-panel")).toBeInTheDocument();
    expect(screen.getByText("系统交易权限")).toBeInTheDocument();
    expect(screen.getByText("允许标的")).toBeInTheDocument();
  });

  it("keeps saved trade settings visible when an older poll resolves later", async () => {
    const staleSettings = deferred<typeof initialSettings>();
    const staleRuntime = deferred<typeof initialRuntime>();
    const staleOrders = deferred<typeof initialOrders>();
    const intervalHandlers: Array<() => void> = [];

    vi.spyOn(window, "setInterval").mockImplementation((handler: TimerHandler) => {
      if (typeof handler === "function") {
        intervalHandlers.push(handler as () => void);
      }
      return 1 as unknown as number;
    });
    vi.spyOn(window, "clearInterval").mockImplementation(() => {});

    mockTradeApi.getSettings
      .mockResolvedValueOnce(initialSettings)
      .mockImplementationOnce(() => staleSettings.promise)
      .mockResolvedValueOnce(savedSettings);
    mockTradeApi.getRuntime
      .mockResolvedValueOnce(initialRuntime)
      .mockImplementationOnce(() => staleRuntime.promise)
      .mockResolvedValueOnce(initialRuntime);
    mockTradeApi.list
      .mockResolvedValueOnce(initialOrders)
      .mockImplementationOnce(() => staleOrders.promise)
      .mockResolvedValueOnce(initialOrders);

    render(<AutoTradingPage />);

    const updatedByInput = await screen.findByLabelText("更新人");
    expect(updatedByInput).toHaveValue("tester");
    expect(intervalHandlers.length).toBeGreaterThan(0);
    for (const handler of intervalHandlers) {
      handler();
    }
    await waitFor(() => {
      expect(mockTradeApi.getSettings).toHaveBeenCalledTimes(2);
    });

    fireEvent.change(updatedByInput, { target: { value: "fresh-save" } });
    fireEvent.click(getCheckboxWithinLabel("持仓同步"));
    fireEvent.click(getCheckboxWithinLabel("ETHUSDT"));
    fireEvent.change(screen.getByLabelText("风险比例 %"), { target: { value: "3.2" } });
    fireEvent.change(screen.getByLabelText("最低盈亏比"), { target: { value: "1.8" } });
    fireEvent.change(screen.getByLabelText("限价单超时秒数"), { target: { value: "61" } });
    fireEvent.change(screen.getByLabelText("最大持仓数"), { target: { value: "3" } });
    fireEvent.click(screen.getByRole("button", { name: "保存配置" }));
    await act(async () => {
      await Promise.resolve();
    });

    await waitFor(() => {
      expect(mockTradeApi.updateSettings).toHaveBeenCalledWith({
        auto_execute_enabled: true,
        allowed_symbols: ["BTCUSDT"],
        risk_pct: 3.2,
        min_risk_reward: 1.8,
        entry_timeout_seconds: 61,
        max_open_positions: 3,
        sync_enabled: false,
        updated_by: "fresh-save",
      });
    });

    await waitFor(() => {
      expect(within(screen.getByTestId("trade-runtime-band")).queryByText("BTCUSDT / ETHUSDT")).not.toBeInTheDocument();
    });

    await act(async () => {
      staleSettings.resolve(initialSettings);
      staleRuntime.resolve(initialRuntime);
      staleOrders.resolve(initialOrders);
      await Promise.resolve();
    });

    expect(within(screen.getByTestId("trade-runtime-band")).queryByText("BTCUSDT / ETHUSDT")).not.toBeInTheDocument();
    expect(screen.getByLabelText("更新人")).toHaveValue("fresh-save");
    expect(screen.getByLabelText("风险比例 %")).toHaveValue(3.2);
    expect(getCheckboxWithinLabel("ETHUSDT")).not.toBeChecked();
    expect(getCheckboxWithinLabel("持仓同步")).not.toBeChecked();
  });
});

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((innerResolve) => {
    resolve = innerResolve;
  });
  return { promise, resolve };
}

function getCheckboxWithinLabel(text: string) {
  const label = screen.getByText(text).closest("label");
  if (!label) {
    throw new Error(`label not found for ${text}`);
  }
  const input = label.querySelector("input[type='checkbox']");
  if (!(input instanceof HTMLInputElement)) {
    throw new Error(`checkbox not found for ${text}`);
  }
  return input;
}
