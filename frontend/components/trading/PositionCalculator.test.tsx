import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

vi.mock("@/services/apiClient", () => ({
  alertApi: {
    getAlertHistory: vi.fn(),
  },
}));

vi.mock("@/store/marketStore", () => ({
  useMarketStore: () => ({ symbol: "BTCUSDT" }),
}));

import { alertApi } from "@/services/apiClient";
import { calcPosition } from "./PositionCalculator";
import PositionCalculator from "./PositionCalculator";

const mockedAlertApi = vi.mocked(alertApi);

describe("calcPosition", () => {
  it("computes position size correctly", () => {
    // stopDist% = |100-95|/100 = 5%, size = 10000*1%/5% = 2000
    const result = calcPosition({ balance: 10000, riskPct: 1, entry: 100, stop: 95 });
    expect(result).not.toBeNull();
    expect(result!.positionSize).toBeCloseTo(2000);
    expect(result!.maxLoss).toBeCloseTo(100);
  });

  it("returns null when stop equals entry", () => {
    const result = calcPosition({ balance: 10000, riskPct: 1, entry: 100, stop: 100 });
    expect(result).toBeNull();
  });

  it("flags warning when margin exceeds balance", () => {
    // stopDist% = 0.1%, size = 1000*1%/0.1% = 10000 > 1000, leverage=1 → margin=10000>1000
    const result = calcPosition({ balance: 1000, riskPct: 1, entry: 100, stop: 99.9, leverage: 1 });
    expect(result).not.toBeNull();
    expect(result!.exceedsBalance).toBe(true);
  });

  it("leverage reduces margin required", () => {
    // size=2000, leverage=10 → margin=200
    const result = calcPosition({ balance: 10000, riskPct: 1, entry: 100, stop: 95, leverage: 10 });
    expect(result).not.toBeNull();
    expect(result!.positionSize).toBeCloseTo(2000);
    expect(result!.marginRequired).toBeCloseTo(200);
    expect(result!.exceedsBalance).toBe(false);
  });

  it("computes R:R when target is provided", () => {
    const result = calcPosition({ balance: 10000, riskPct: 1, entry: 100, stop: 95, target: 110 });
    expect(result).not.toBeNull();
    expect(result!.rr).not.toBeNull();
    // maxProfit = 2000*(10/100)=200, maxLoss=100, rr=2
    expect(result!.rr!).toBeCloseTo(2);
  });

  it("returns null rr when target is not provided", () => {
    const result = calcPosition({ balance: 10000, riskPct: 1, entry: 100, stop: 95 });
    expect(result).not.toBeNull();
    expect(result!.rr).toBeNull();
    expect(result!.maxProfit).toBeNull();
  });

  it("returns null when riskPct is zero", () => {
    const result = calcPosition({ balance: 10000, riskPct: 0, entry: 100, stop: 95 });
    expect(result).toBeNull();
  });

  it("returns null when balance is zero", () => {
    const result = calcPosition({ balance: 0, riskPct: 1, entry: 100, stop: 95 });
    expect(result).toBeNull();
  });

  it("infers long direction when stop < entry", () => {
    const result = calcPosition({ balance: 10000, riskPct: 1, entry: 100, stop: 95 });
    expect(result).not.toBeNull();
    expect(result!.direction).toBe("long");
  });

  it("infers short direction when stop > entry", () => {
    const result = calcPosition({ balance: 10000, riskPct: 1, entry: 100, stop: 105 });
    expect(result).not.toBeNull();
    expect(result!.direction).toBe("short");
  });
});

describe("PositionCalculator render", () => {
  beforeEach(() => {
    window.localStorage.clear();
    mockedAlertApi.getAlertHistory.mockResolvedValue({ items: [], generated: 0 });
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

  it("renders without crashing", () => {
    render(<PositionCalculator />);
    expect(screen.getByText("仓位计算器")).toBeTruthy();
  });

  it("renders as a cockpit surface region", () => {
    render(<PositionCalculator />);

    expect(screen.getByTestId("position-calculator-panel")).toHaveAttribute("data-panel-role", "action");
    expect(screen.getByRole("region", { name: "仓位计算器" })).toBeInTheDocument();
    expect(screen.getByText("账户余额 (USDT)")).toBeInTheDocument();
    expect(screen.getByText("杠杆倍数 ×")).toBeInTheDocument();
  });

  it("renders direction toggle buttons", () => {
    render(<PositionCalculator />);
    expect(screen.getByTitle("做多（买入）")).toBeInTheDocument();
    expect(screen.getByTitle("做空（卖出）")).toBeInTheDocument();
  });

  it("shows error tag when stop equals entry", async () => {
    const user = userEvent.setup();
    render(<PositionCalculator />);

    const inputs = screen.getAllByRole("spinbutton");
    // inputs[0]=balance, inputs[1]=riskPct, inputs[2]=leverage, inputs[3]=entry, inputs[4]=stop, inputs[5]=target
    await user.clear(inputs[3]);
    await user.type(inputs[3], "100");
    await user.clear(inputs[4]);
    await user.type(inputs[4], "100");

    expect(await screen.findByText("止损价不能等于进场价")).toBeTruthy();
  });
});
