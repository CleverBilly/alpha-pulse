import { expect, test } from "@playwright/test";
import { mockMarketSnapshotApi } from "./support/mockMarketApi";

test("dashboard renders cockpit layers and supports 4h interval", async ({ page }) => {
  const controller = await mockMarketSnapshotApi(page);
  await page.goto("/dashboard");
  await expect.poll(() => controller.getRequestCount()).toBeGreaterThan(0);

  await expect(page.getByText("当前判断")).toBeVisible();
  await expect(page.getByText("强偏多").first()).toBeVisible();
  await expect(page.getByRole("heading", { name: "Execution Setup" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Evidence Chain" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Kline Chart" })).toBeVisible();
  await expect(page.getByText("Entry Zone").first()).toBeVisible();
  await expect(page.getByRole("link", { name: "查看信号深页" })).toBeVisible();
  await expect(page.getByRole("link", { name: "查看市场深页" })).toBeVisible();
  await expect(page.getByRole("link", { name: "查看图表深页" })).toBeVisible();

  await page.getByRole("button", { name: "4h" }).click();
  await expect(page.getByRole("button", { name: "4h" })).toHaveAttribute("aria-pressed", "true");
});
