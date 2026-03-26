import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { alertApi } from "@/services/apiClient";
import AlertCenter from "./AlertCenter";

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

describe("AlertCenter", () => {
  beforeEach(() => {
    window.localStorage.clear();
    vi.spyOn(window, "getComputedStyle").mockImplementation(
      () =>
        ({
          getPropertyValue: () => "",
        }) as CSSStyleDeclaration,
    );
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("renders alert feed in the drawer", async () => {
    stubNotification("default");
    mockedAlertApi.getAlertPreferences.mockResolvedValue(buildPreferences());
    mockedAlertApi.getAlerts.mockResolvedValue({
      items: [buildMockAlert()],
      generated: 0,
    });
    mockedAlertApi.refreshAlerts.mockResolvedValue({
      items: [buildMockAlert()],
      generated: 1,
    });

    render(<AlertCenter />);

    await waitFor(() => {
      expect(mockedAlertApi.getAlerts).toHaveBeenCalledWith(20);
    });

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "打开告警中心" }));

    expect(screen.getByText("告警中心")).toBeInTheDocument();
    expect(screen.getByText("BTCUSDT A 级机会已就绪")).toBeInTheDocument();
    expect(screen.getByText("浏览器通知未授权")).toBeInTheDocument();
  });

  it("refreshes alerts and sends browser notification when granted", async () => {
    const requestPermission = vi.fn().mockResolvedValue("granted");
    const notificationSpy = stubNotification("granted", requestPermission);

    mockedAlertApi.getAlertPreferences.mockResolvedValue(buildPreferences());
    mockedAlertApi.getAlerts.mockResolvedValue({
      items: [],
      generated: 0,
    });
    mockedAlertApi.refreshAlerts.mockResolvedValue({
      items: [buildMockAlert()],
      generated: 1,
    });

    render(<AlertCenter />);
    await waitFor(() => {
      expect(mockedAlertApi.getAlerts).toHaveBeenCalledWith(20);
    });

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "打开告警中心" }));
    await user.click(screen.getByRole("button", { name: /立即检查/ }));

    await waitFor(() => {
      expect(mockedAlertApi.refreshAlerts).toHaveBeenCalledWith(20);
      expect(notificationSpy).toHaveBeenCalledWith(
        "BTCUSDT A 级机会已就绪",
        expect.objectContaining({
          tag: "alert-1",
        }),
      );
    });

    expect(requestPermission).not.toHaveBeenCalled();
  });

  it("opens config center and saves alert preferences", async () => {
    mockedAlertApi.getAlertPreferences.mockResolvedValue(buildPreferences());
    mockedAlertApi.updateAlertPreferences.mockImplementation(async (payload) => payload);
    mockedAlertApi.getAlerts.mockResolvedValue({
      items: [],
      generated: 0,
    });
    mockedAlertApi.refreshAlerts.mockResolvedValue({
      items: [],
      generated: 0,
    });

    render(<AlertCenter />);
    await waitFor(() => {
      expect(mockedAlertApi.getAlertPreferences).toHaveBeenCalled();
    });

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "打开告警中心" }));
    await user.click(screen.getByRole("button", { name: /配置中心/ }));

    expect(screen.getByText("推送渠道")).toBeInTheDocument();
    await user.click(screen.getAllByRole("switch")[0]);
    await user.click(screen.getByRole("button", { name: "保存配置" }));

    await waitFor(() => {
      expect(mockedAlertApi.updateAlertPreferences).toHaveBeenCalled();
    });
  });
});

function buildMockAlert() {
  return {
    id: "alert-1",
    symbol: "BTCUSDT",
    kind: "setup_ready",
    severity: "A",
    title: "BTCUSDT A 级机会已就绪",
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
    deliveries: [
      {
        channel: "feishu",
        status: "sent",
        sent_at: 1710000000000,
      },
    ],
  };
}

function stubNotification(permission: NotificationPermission, requestPermission = vi.fn()) {
  const notificationSpy = vi.fn();

  class MockNotification {
    static permission = permission;
    static requestPermission = requestPermission;
    onclick: (() => void) | null = null;

    constructor(title: string, options?: NotificationOptions) {
      notificationSpy(title, options);
    }
  }

  vi.stubGlobal("Notification", MockNotification);
  return notificationSpy;
}

function buildPreferences() {
  return {
    feishu_enabled: true,
    browser_enabled: true,
    setup_ready_enabled: true,
    direction_shift_enabled: true,
    no_trade_enabled: true,
    minimum_confidence: 55,
    quiet_hours_enabled: false,
    quiet_hours_start: 0,
    quiet_hours_end: 8,
    sound_enabled: false,
    symbols: ["BTCUSDT", "ETHUSDT"],
    available_symbols: ["BTCUSDT", "ETHUSDT", "SOLUSDT"],
  };
}
