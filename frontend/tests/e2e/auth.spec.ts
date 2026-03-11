import { expect, test } from "@playwright/test";
import { mockAuthApi } from "./support/mockAuthApi";
import { mockAlertApi } from "./support/mockAlertApi";
import { mockMarketSnapshotApi } from "./support/mockMarketApi";

test("unauthenticated user is redirected to login", async ({ page }) => {
  await page.goto("/dashboard");

  await expect(page).toHaveURL(/\/login$/);
  await expect(page.getByRole("heading", { name: "登录 Alpha Pulse" })).toBeVisible();
});

test("user can log in and enter the dashboard", async ({ page }) => {
  await mockAuthApi(page);
  await mockAlertApi(page);
  await mockMarketSnapshotApi(page);

  await page.goto("/login");
  await page.getByLabel("用户名").fill("alpha-admin");
  await page.getByLabel("密码").fill("alpha-pass");
  await page.getByRole("button", { name: "登录" }).click();

  await expect(page).toHaveURL(/\/dashboard$/);
  await expect(page.getByText("当前判断")).toBeVisible();
});

test("user can log out and returns to login", async ({ page }) => {
  await mockAuthApi(page);
  await mockAlertApi(page);
  await mockMarketSnapshotApi(page);

  await page.goto("/login");
  await page.getByLabel("用户名").fill("alpha-admin");
  await page.getByLabel("密码").fill("alpha-pass");
  await page.getByRole("button", { name: "登录" }).click();

  await expect(page).toHaveURL(/\/dashboard$/);
  await page.getByRole("button", { name: "退出登录" }).click();

  await expect(page).toHaveURL(/\/login$/);
});
