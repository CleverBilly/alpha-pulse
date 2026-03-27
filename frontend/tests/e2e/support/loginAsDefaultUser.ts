import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { mockAuthApi, primeAuthSession } from "./mockAuthApi";

export async function loginAsDefaultUser(page: Page) {
  await mockAuthApi(page);
  await primeAuthSession(page);
  await page.goto("/dashboard");
  await expect(page).toHaveURL(/\/dashboard$/);
}
