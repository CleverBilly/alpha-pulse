import type { Page } from "@playwright/test";

interface MockAlertApiOptions {
  items?: unknown[];
  generated?: number;
}

export async function mockAlertApi(page: Page, options: MockAlertApiOptions = {}) {
  const body = {
    code: 0,
    message: "ok",
    data: {
      items: options.items ?? [],
      generated: options.generated ?? 0,
    },
  };

  await page.route(/http:\/\/(127\.0\.0\.1|localhost):8080\/api\/alerts(\?.*)?$/, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(body),
    });
  });

  await page.route(/http:\/\/(127\.0\.0\.1|localhost):8080\/api\/alerts\/refresh(\?.*)?$/, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(body),
    });
  });
}
