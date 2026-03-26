import { expect, test } from "@playwright/test";
import { mockAlertApi } from "./support/mockAlertApi";
import { loginAsDefaultUser } from "./support/loginAsDefaultUser";
import { mockMarketSnapshotApi } from "./support/mockMarketApi";

test("dashboard renders cockpit layers and supports 4h interval", async ({ page }) => {
  await mockAlertApi(page);
  const controller = await mockMarketSnapshotApi(page);
  await loginAsDefaultUser(page);
  await expect.poll(() => controller.getRequestCount()).toBeGreaterThan(0);

  await expect(page.getByText("当前判断")).toBeVisible();
  await expect(page.getByText("强偏多").first()).toBeVisible();
  await expect(page.getByRole("heading", { name: "执行方案" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "证据链" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "K 线图" })).toBeVisible();
  await expect(page.getByText("进场区间").first()).toBeVisible();
  await expect(page.getByRole("link", { name: "查看信号深页" })).toBeVisible();
  await expect(page.getByRole("link", { name: "查看市场深页" })).toBeVisible();
  await expect(page.getByRole("link", { name: "查看图表深页" })).toBeVisible();
  await expect(page.getByText("5m 强偏多").first()).toBeVisible();

  await page.getByRole("button", { name: "4h" }).click();
  await expect(page.getByRole("button", { name: "4h" })).toHaveAttribute("aria-pressed", "true");
});
