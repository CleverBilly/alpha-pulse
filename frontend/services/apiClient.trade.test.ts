import { afterEach, describe, expect, it, vi } from "vitest";

describe("tradeApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.resetModules();
  });

  it("loads trade settings", async () => {
    const fetchSpy = vi.fn().mockResolvedValue(mockResponse({
      trade_enabled_env: true,
      trade_auto_execute_env: true,
      auto_execute_enabled: true,
      allowed_symbols: ["BTCUSDT", "ETHUSDT"],
      risk_pct: 2.5,
      min_risk_reward: 1.6,
      entry_timeout_seconds: 45,
      max_open_positions: 2,
      sync_enabled: true,
      updated_by: "tester",
    }));
    vi.stubGlobal("fetch", fetchSpy);

    const { tradeApi } = await import("./apiClient");
    const result = await tradeApi.getSettings();

    expect(fetchSpy).toHaveBeenCalledWith(
      expect.stringContaining("/trade-settings"),
      expect.objectContaining({
        credentials: "include",
      }),
    );
    expect(result.allowed_symbols).toEqual(["BTCUSDT", "ETHUSDT"]);
  });

  it("updates trade settings", async () => {
    const fetchSpy = vi.fn().mockResolvedValue(mockResponse({
      trade_enabled_env: true,
      trade_auto_execute_env: true,
      auto_execute_enabled: true,
      allowed_symbols: ["BTCUSDT"],
      risk_pct: 2,
      min_risk_reward: 1.2,
      entry_timeout_seconds: 45,
      max_open_positions: 1,
      sync_enabled: true,
      updated_by: "tester",
    }));
    vi.stubGlobal("fetch", fetchSpy);

    const { tradeApi } = await import("./apiClient");
    await tradeApi.updateSettings({
      auto_execute_enabled: true,
      allowed_symbols: ["BTCUSDT"],
      risk_pct: 2,
      min_risk_reward: 1.2,
      entry_timeout_seconds: 45,
      max_open_positions: 1,
      sync_enabled: true,
      updated_by: "tester",
    });

    expect(fetchSpy).toHaveBeenCalledWith(
      expect.stringContaining("/trade-settings"),
      expect.objectContaining({
        method: "PUT",
      }),
    );
  });

  it("lists trade orders and loads runtime", async () => {
    const fetchSpy = vi
      .fn()
      .mockResolvedValueOnce(mockResponse([
        {
          id: 1,
          alert_id: "alert-1",
          symbol: "BTCUSDT",
          side: "LONG",
          requested_qty: 0.01,
          filled_qty: 0,
          limit_price: 65000,
          entry_status: "pending_fill",
          status: "pending_fill",
          source: "system",
          created_at: 1710000000000,
          closed_at: 0,
        },
      ]))
      .mockResolvedValueOnce(mockResponse({
        trade_enabled_env: true,
        trade_auto_execute_env: true,
        pending_fill_count: 1,
        open_count: 0,
      }));
    vi.stubGlobal("fetch", fetchSpy);

    const { tradeApi } = await import("./apiClient");
    const orders = await tradeApi.list({ symbol: "BTCUSDT" });
    const runtime = await tradeApi.getRuntime();

    expect(orders).toHaveLength(1);
    expect(runtime.pending_fill_count).toBe(1);
  });

  it("closes a trade order", async () => {
    const fetchSpy = vi.fn().mockResolvedValue(mockResponse({ closed: true }));
    vi.stubGlobal("fetch", fetchSpy);

    const { tradeApi } = await import("./apiClient");
    await tradeApi.close(42);

    expect(fetchSpy).toHaveBeenCalledWith(
      expect.stringContaining("/trades/42/close"),
      expect.objectContaining({
        method: "POST",
      }),
    );
  });
});

function mockResponse(data: unknown) {
  return {
    ok: true,
    status: 200,
    json: async () => ({
      code: 0,
      message: "ok",
      data,
    }),
  };
}
