import { expect, test } from "@playwright/test";
import { mockAlertApi } from "./support/mockAlertApi";
import { mockMarketSnapshotApi } from "./support/mockMarketApi";

test("signals page shows replay board and live signal context", async ({ page }) => {
  await mockAlertApi(page, {
    items: [
      {
        id: "alert-1",
        symbol: "BTCUSDT",
        kind: "setup_ready",
        severity: "A",
        title: "BTCUSDT A 级 setup 已就绪",
        verdict: "强偏多",
        tradeability_label: "A 级可跟踪",
        summary: "4h 与 1h 已经对齐，15m 触发也站在同一边。",
        reasons: ["趋势因子主导当前方向。"],
        timeframe_labels: ["4h 强偏多", "1h 强偏多", "15m 强偏多"],
        confidence: 74,
        risk_label: "可控风险",
        entry_price: 65200,
        stop_loss: 64880,
        target_price: 65880,
        risk_reward: 2.1,
        created_at: 1710000000000,
        deliveries: [],
      },
    ],
  });
  await mockMarketSnapshotApi(page);
  await page.goto("/signals");

  await expect(page.getByRole("heading", { name: "Alert Review Board" })).toBeVisible();
  await expect(page.getByText("BTCUSDT A 级 setup 已就绪")).toBeVisible();
  await expect(page.getByText("Live Signal Context")).toBeVisible();
  await expect(page.getByText("Decision Memo")).toBeVisible();
  await expect(page.getByText("Factor Breakdown")).toBeVisible();
  await expect(page.getByText("Bullish Drivers")).toBeVisible();
  await expect(page.getByText("Recent Signal Tape")).toBeVisible();
  await expect(page.getByText("Microstructure Tape")).toBeVisible();
  await expect(page.getByText("最近微结构事件连续偏多，买方主动性增强")).toBeVisible();
});
