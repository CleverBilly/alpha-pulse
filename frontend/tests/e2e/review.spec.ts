import { expect, test } from "@playwright/test";
import { mockAlertApi } from "./support/mockAlertApi";
import { mockMarketSnapshotApi } from "./support/mockMarketApi";

test("review route shows alert replay and live signal context", async ({ page }) => {
  await mockAlertApi(page);
  await mockMarketSnapshotApi(page);
  await page.goto("/review");

  await expect(page.getByRole("heading", { name: "告警复盘看板" })).toBeVisible();
  await expect(page.getByText("实时信号上下文")).toBeVisible();
  await expect(page.getByText("决策备忘")).toBeVisible();
});
