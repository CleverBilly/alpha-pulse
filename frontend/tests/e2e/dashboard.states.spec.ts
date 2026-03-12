import { expect, test } from "@playwright/test";
import { buildMockMarketSnapshot } from "../../test/fixtures/marketSnapshot";
import { mockAlertApi } from "./support/mockAlertApi";
import { mockMarketSnapshotApi } from "./support/mockMarketApi";

test("dashboard shows api failure state", async ({ page }) => {
  await mockAlertApi(page);
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
  await expect(page.getByRole("button", { name: "等待 setup 完整" })).toBeDisabled();
});

test("dashboard shows loading state on weak network", async ({ page }) => {
  await mockAlertApi(page);
  await mockMarketSnapshotApi(page, { delayMs: 1800 });

  await page.goto("/dashboard");
  await expect(page.getByText("加载中...").first()).toBeVisible();
  await expect(page.getByText("当前判断")).toBeVisible();
  await expect(page.getByRole("heading", { name: "执行方案" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "K 线图" })).toBeVisible();
  await expect(page.getByRole("button", { name: "等待 setup 完整" })).toBeVisible();
});

test("dashboard updates when switching symbol", async ({ page }) => {
  await mockAlertApi(page);
  const controller = await mockMarketSnapshotApi(page);

  await page.goto("/dashboard");
  await expect.poll(() => controller.getRequestCount()).toBeGreaterThan(0);
  const baseline = controller.getRequestCount();

  await page.getByLabel("Symbol").selectOption("ETHUSDT");
  await expect.poll(() => controller.getRequestCount()).toBeGreaterThan(baseline);

  await expect(page.getByText("ETHUSDT").first()).toBeVisible();
});

test("dashboard manual refresh triggers another snapshot request", async ({ page }) => {
  await mockAlertApi(page);
  const controller = await mockMarketSnapshotApi(page);

  await page.goto("/dashboard");
  await expect.poll(() => controller.getRequestCount()).toBeGreaterThan(0);
  const baseline = controller.getRequestCount();

  await page.getByRole("button", { name: /^刷新$/ }).click();
  await expect.poll(() => controller.getRequestCount()).toBeGreaterThan(baseline);
});

test("dashboard does not show fake setup when signal levels are missing", async ({ page }) => {
  await mockAlertApi(page);
  await mockMarketSnapshotApi(page, {
    resolver: ({ symbol, interval, limit }) => {
      const snapshot = buildMockMarketSnapshot(symbol, interval, limit);
      snapshot.signal.entry_price = Number.NaN;
      snapshot.signal.stop_loss = Number.NaN;
      snapshot.signal.target_price = Number.NaN;
      return {
        data: snapshot,
      };
    },
  });

  await page.goto("/dashboard");

  await expect(page.getByText("等待更完整的入场、止损和目标位后再做执行判断。")).toBeVisible();
  await expect(page.getByRole("button", { name: "等待 setup 完整" })).toBeDisabled();
});
