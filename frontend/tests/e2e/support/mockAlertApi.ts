import type { Page } from "@playwright/test";

interface MockAlertApiOptions {
  items?: unknown[];
  generated?: number;
  preferences?: unknown;
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
  const preferencesBody = {
    code: 0,
    message: "ok",
    data:
      options.preferences ?? {
        feishu_enabled: true,
        browser_enabled: true,
        setup_ready_enabled: true,
        direction_shift_enabled: true,
        no_trade_enabled: true,
        minimum_confidence: 55,
        quiet_hours_enabled: false,
        quiet_hours_start: 0,
        quiet_hours_end: 8,
        symbols: ["BTCUSDT", "ETHUSDT", "SOLUSDT"],
        available_symbols: ["BTCUSDT", "ETHUSDT", "SOLUSDT"],
      },
  };

  await page.route(/http:\/\/(127\.0\.0\.1|localhost):8080\/api\/alerts(\?.*)?$/, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(body),
    });
  });

  await page.route(/http:\/\/(127\.0\.0\.1|localhost):8080\/api\/alerts\/history(\?.*)?$/, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(body),
    });
  });

  await page.route(/http:\/\/(127\.0\.0\.1|localhost):8080\/api\/alerts\/preferences(\?.*)?$/, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(preferencesBody),
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
