import { afterEach, describe, expect, it, vi } from "vitest";

describe("marketApi websocket stream url", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
    vi.resetModules();
  });

  it("uses the current site origin for market snapshot streaming instead of the HTTP api base", async () => {
    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", "http://127.0.0.1:8080/api");

    const { marketApi } = await import("./apiClient");
    const expectedOrigin = window.location.origin.replace(/^http/, "ws");

    expect(marketApi.createMarketSnapshotStreamUrl("BTCUSDT", "15m", 48)).toBe(
      `${expectedOrigin}/api/market-snapshot/stream?symbol=BTCUSDT&interval=15m&limit=48`,
    );
  });

  it("prefers an explicit market stream base url when provided", async () => {
    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", "/api");
    vi.stubEnv("NEXT_PUBLIC_MARKET_STREAM_BASE_URL", "http://127.0.0.1:8080");

    const { marketApi } = await import("./apiClient");

    expect(marketApi.createMarketSnapshotStreamUrl("BTCUSDT", "15m", 48)).toBe(
      "ws://127.0.0.1:8080/api/market-snapshot/stream?symbol=BTCUSDT&interval=15m&limit=48",
    );
  });
});
