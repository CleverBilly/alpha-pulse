import { expect, test } from "@playwright/test";
import { mockAlertApi } from "./support/mockAlertApi";
import { mockMarketSnapshotApi } from "./support/mockMarketApi";

test("market page shows overview, price ladder and signal tape", async ({ page }) => {
  await mockAlertApi(page);
  await mockMarketSnapshotApi(page);
  await page.goto("/market");

  await expect(page.getByText("Market Overview")).toBeVisible();
  await expect(page.getByText("Price Ladder")).toBeVisible();
  await expect(page.getByText("Signal Tape")).toBeVisible();
  await expect(page.getByText("Microstructure Timeline")).toBeVisible();
  await expect(page.getByRole("heading", { name: "Order Flow" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Liquidity" })).toBeVisible();
  await expect(page.getByText("Microstructure Events")).toBeVisible();
  await expect(page.getByText("Initiative Exhaustion").first()).toBeVisible();
  await expect(page.getByRole("heading", { name: "Stop Clusters" })).toBeVisible();
});
