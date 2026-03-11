import { expect, test } from "@playwright/test";
import { mockAlertApi } from "./support/mockAlertApi";
import { mockMarketSnapshotApi } from "./support/mockMarketApi";

test("review route shows alert replay and live signal context", async ({ page }) => {
  await mockAlertApi(page);
  await mockMarketSnapshotApi(page);
  await page.goto("/review");

  await expect(page.getByRole("heading", { name: "Alert Review Board" })).toBeVisible();
  await expect(page.getByText("Live Signal Context")).toBeVisible();
  await expect(page.getByText("Decision Memo")).toBeVisible();
});
