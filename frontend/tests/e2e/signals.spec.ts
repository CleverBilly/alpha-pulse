import { expect, test } from "@playwright/test";
import { mockMarketSnapshotApi } from "./support/mockMarketApi";

test("signals page shows signal card and ai analysis", async ({ page }) => {
  await mockMarketSnapshotApi(page);
  await page.goto("/signals");

  await expect(page.getByRole("heading", { name: "Trading Signals" })).toBeVisible();
  await expect(page.getByText("Decision Memo")).toBeVisible();
  await expect(page.getByText("Factor Breakdown")).toBeVisible();
  await expect(page.getByText("Bullish Drivers")).toBeVisible();
  await expect(page.getByText("Recent Signal Tape")).toBeVisible();
  await expect(page.getByText("Microstructure Tape")).toBeVisible();
  await expect(page.getByText("最近微结构事件连续偏多，买方主动性增强")).toBeVisible();
});
