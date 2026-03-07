import { expect, test } from "@playwright/test";
import { mockMarketSnapshotApi } from "./support/mockMarketApi";

test("dashboard shows api failure state", async ({ page }) => {
  await mockMarketSnapshotApi(page, {
    resolver: () => ({
      status: 500,
      code: 500,
      message: "upstream unavailable",
      data: null,
    }),
  });

  await page.goto("/dashboard");
  await expect(page.getByText("API request failed: 500").first()).toBeVisible();
});

test("dashboard shows loading state on weak network", async ({ page }) => {
  await mockMarketSnapshotApi(page, { delayMs: 1800 });

  await page.goto("/dashboard");
  await expect(page.getByText("加载中...").first()).toBeVisible();
  await expect(page.getByRole("heading", { name: "Kline Chart" })).toBeVisible();
  await expect(page.getByText("Signal Entry").first()).toBeVisible();
});

test("dashboard updates when switching symbol", async ({ page }) => {
  const controller = await mockMarketSnapshotApi(page);

  await page.goto("/dashboard");
  await expect.poll(() => controller.getRequestCount()).toBeGreaterThan(0);
  const baseline = controller.getRequestCount();

  await page.locator("select").first().selectOption("ETHUSDT");
  await expect.poll(() => controller.getRequestCount()).toBeGreaterThan(baseline);

  const signalSection = page.locator("section").filter({
    has: page.getByRole("heading", { name: "Signal" }),
  }).first();
  await expect(signalSection.getByText("ETHUSDT")).toBeVisible();
});

test("dashboard manual refresh triggers another snapshot request", async ({ page }) => {
  const controller = await mockMarketSnapshotApi(page);

  await page.goto("/dashboard");
  await expect.poll(() => controller.getRequestCount()).toBeGreaterThan(0);
  const baseline = controller.getRequestCount();

  await page.getByRole("button", { name: /^刷新$/ }).click();
  await expect.poll(() => controller.getRequestCount()).toBeGreaterThan(baseline);
});
