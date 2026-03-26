import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { mockAuthApi } from "./mockAuthApi";

export async function loginAsDefaultUser(page: Page) {
  await mockAuthApi(page);
  await page.goto("/login");
  await page.getByLabel("用户名").fill("alpha-admin");
  await page.getByLabel("密码").fill("alpha-pass");
  await page.getByRole("button", { name: "登录" }).click();
  await expect(page).toHaveURL(/\/dashboard$/);
}
