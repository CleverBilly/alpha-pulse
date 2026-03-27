import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("frontend api proxy route", () => {
  beforeEach(() => {
    vi.resetModules();
    vi.stubEnv("API_PROXY_TARGET", "http://backend:8080");
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it("forwards GET requests to the backend api target", async () => {
    const fetchSpy = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: {
          "content-type": "application/json",
        },
      }),
    );
    vi.stubGlobal("fetch", fetchSpy);

    const { GET } = await import("./route");
    const response = await GET(new Request("http://localhost/api/trade-settings?limit=24"), {
      params: { path: ["trade-settings"] },
    });

    expect(fetchSpy).toHaveBeenCalledTimes(1);
    expect(String(fetchSpy.mock.calls[0]?.[0])).toBe("http://backend:8080/api/trade-settings?limit=24");
    expect(response.status).toBe(200);
  });

  it("forwards non-GET methods with body and cookies", async () => {
    const fetchSpy = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ closed: true }), {
        status: 200,
        headers: {
          "content-type": "application/json",
        },
      }),
    );
    vi.stubGlobal("fetch", fetchSpy);

    const { POST } = await import("./route");
    const response = await POST(
      new Request("http://localhost/api/trades/42/close", {
        method: "POST",
        headers: {
          cookie: "alpha_pulse_session=test-token",
          "content-type": "application/json",
        },
        body: JSON.stringify({ reason: "manual" }),
      }),
      {
        params: { path: ["trades", "42", "close"] },
      },
    );

    expect(fetchSpy).toHaveBeenCalledTimes(1);
    expect(String(fetchSpy.mock.calls[0]?.[0])).toBe("http://backend:8080/api/trades/42/close");
    expect(fetchSpy).toHaveBeenCalledWith(
      expect.any(URL),
      expect.objectContaining({
        method: "POST",
        headers: expect.any(Headers),
      }),
    );
    expect(response.status).toBe(200);
  });
});
