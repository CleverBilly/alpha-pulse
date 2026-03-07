import { expect, test } from "@playwright/test";
import { mockMarketSnapshotApi } from "./support/mockMarketApi";

test("dashboard renders chart, signal and supports 4h interval", async ({ page }) => {
  await mockMarketSnapshotApi(page);
  await page.goto("/dashboard");

  await expect(page.getByRole("heading", { name: "Kline Chart" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Signal" })).toBeVisible();
  await expect(page.getByText("Signal Entry").first()).toBeVisible();
  await expect(page.getByText("Buy Liquidity").first()).toBeVisible();
  await expect(page.getByText("ABS").first()).toBeVisible();
  await expect(page.getByText("ICE").first()).toBeVisible();
  await expect(page.getByRole("button", { name: "Initiative Shift" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Large Trade Cluster" })).toBeVisible();

  await page.getByRole("button", { name: "Initiative Shift" }).click();
  await page.getByRole("button", { name: "Large Trade Cluster" }).click();
  await expect(page.getByText("SHF").first()).toBeVisible();
  await expect(page.getByText("LTC").first()).toBeVisible();

  await page.getByRole("button", { name: "Micro ABS Absorption" }).focus();
  const tooltip = page.getByLabel("Microstructure Tooltip");
  await expect(tooltip.getByText("Absorption · ABS")).toBeVisible();
  await expect(tooltip.getByText("卖压被持续吸收，价格未继续下破")).toBeVisible();

  await page.getByRole("button", { name: "4h" }).click();
  await expect(page.getByText("当前周期 4h")).toBeVisible();
});
