import type { Page } from "@playwright/test";
import { buildMockMarketSnapshot } from "../../../test/fixtures/marketSnapshot";
import type { MarketInterval } from "../../../types/market";

interface SnapshotRequestContext {
  symbol: string;
  interval: MarketInterval;
  limit: number;
  requestCount: number;
}

interface SnapshotResponseOverride {
  status?: number;
  code?: number;
  message?: string;
  data?: unknown;
}

interface MockMarketSnapshotOptions {
  delayMs?: number;
  disableWebSocket?: boolean;
  onRequest?: (context: SnapshotRequestContext) => void;
  resolver?: (context: SnapshotRequestContext) => SnapshotResponseOverride | Promise<SnapshotResponseOverride>;
}

export async function mockMarketSnapshotApi(page: Page, options: MockMarketSnapshotOptions = {}) {
  let requestCount = 0;

  if (options.disableWebSocket ?? true) {
    await page.routeWebSocket(/wss?:\/\/(127\.0\.0\.1|localhost):8080\/api\/market-snapshot\/stream.*/, (ws) => {
      void ws.close({
        code: 1013,
        reason: "E2E mocked HTTP snapshot",
      });
    });
  }

  await page.route(/http:\/\/(127\.0\.0\.1|localhost):8080\/api\/market-snapshot.*/, async (route) => {
    const url = new URL(route.request().url());
    const symbol = url.searchParams.get("symbol") ?? "BTCUSDT";
    const interval = (url.searchParams.get("interval") ?? "1m") as MarketInterval;
    const rawLimit = Number(url.searchParams.get("limit") ?? "48");
    const limit = Number.isFinite(rawLimit) ? rawLimit : 48;
    requestCount++;

    const context = {
      symbol,
      interval,
      limit,
      requestCount,
    } satisfies SnapshotRequestContext;

    options.onRequest?.(context);
    if (options.delayMs && options.delayMs > 0) {
      await delay(options.delayMs);
    }

    const resolved = options.resolver ? await options.resolver(context) : {};
    const status = resolved.status ?? 200;
    const code = resolved.code ?? 0;
    const message = resolved.message ?? (status >= 400 ? "request failed" : "success");
    const data = resolved.data ?? buildMockMarketSnapshot(symbol, interval, limit);

    await route.fulfill({
      status,
      contentType: "application/json",
      body: JSON.stringify({
        code,
        message,
        data,
      }),
    });
  });

  return {
    getRequestCount: () => requestCount,
  };
}

function delay(timeoutMs: number) {
  return new Promise((resolve) => setTimeout(resolve, timeoutMs));
}
