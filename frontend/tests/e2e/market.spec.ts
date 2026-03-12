import { expect, test } from "@playwright/test";
import { mockAlertApi } from "./support/mockAlertApi";
import { mockMarketSnapshotApi } from "./support/mockMarketApi";

test("market page shows overview, price ladder and signal tape", async ({ page }) => {
  await mockAlertApi(page);
  await mockMarketSnapshotApi(page);
  await page.goto("/market");

  await expect(page.getByText("市场总览")).toBeVisible();
  await expect(page.getByText("价格阶梯")).toBeVisible();
  await expect(page.getByText("信号序列")).toBeVisible();
  await expect(page.getByText("微结构时间线")).toBeVisible();
  await expect(page.getByRole("heading", { name: "订单流" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "流动性" })).toBeVisible();
  await expect(page.getByText("微结构事件")).toBeVisible();
  await expect(page.getByText("主动性衰竭").first()).toBeVisible();
  await expect(page.getByRole("heading", { name: "止损簇" })).toBeVisible();
});
