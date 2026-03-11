import {
  Indicator,
  IndicatorSeriesResult,
  Kline,
  Liquidity,
  LiquidityMapResult,
  MarketInterval,
  OrderFlow,
  PriceTicker,
  Structure,
  StructureSeriesResult,
  LiquiditySeriesResult,
  MicrostructureEventsResult,
} from "@/types/market";
import { MarketSnapshot } from "@/types/snapshot";
import { SignalBundle, SignalTimelineResult } from "@/types/signal";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api";
const AUTH_ENABLED = process.env.NEXT_PUBLIC_AUTH_ENABLED === "true";

interface ApiEnvelope<T> {
  code: number;
  message: string;
  data: T;
}

interface RequestOptions extends RequestInit {
  redirectOnUnauthorized?: boolean;
}

export interface AuthSessionState {
  enabled: boolean;
  authenticated: boolean;
  username?: string;
}

function buildApiUrl(path: string) {
  const fallbackOrigin = typeof window !== "undefined" ? window.location.origin : "http://localhost";
  const base = new URL(API_BASE_URL, fallbackOrigin);
  const [pathname, search = ""] = path.split("?");
  const url = new URL(base.toString());
  url.pathname = `${url.pathname.replace(/\/$/, "")}${pathname.startsWith("/") ? pathname : `/${pathname}`}`;
  url.search = search;
  return url;
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { redirectOnUnauthorized = true, headers, ...init } = options;
  const response = await fetch(buildApiUrl(path).toString(), {
    cache: "no-store",
    credentials: "include",
    headers,
    ...init,
  });

  if (response.status === 401) {
    if (redirectOnUnauthorized && AUTH_ENABLED && typeof window !== "undefined" && window.location.pathname !== "/login") {
      window.location.assign("/login");
    }
    throw new Error("authentication required");
  }

  if (!response.ok) {
    throw new Error(`API request failed: ${response.status}`);
  }

  const payload = (await response.json()) as ApiEnvelope<T>;
  if (payload.code !== 0) {
    throw new Error(payload.message || "API error");
  }

  return payload.data;
}

function buildWebSocketUrl(path: string) {
  const url = buildApiUrl(path);
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  return url.toString();
}

export const marketApi = {
  getPrice(symbol: string) {
    return request<PriceTicker>(`/price?symbol=${symbol}`);
  },
  getKline(symbol: string, interval: MarketInterval = "1m") {
    return request<Kline>(`/kline?symbol=${symbol}&interval=${interval}`);
  },
  getIndicators(symbol: string, interval: MarketInterval = "1m") {
    return request<Indicator>(`/indicators?symbol=${symbol}&interval=${interval}`);
  },
  getIndicatorSeries(symbol: string, interval: MarketInterval = "1m", limit = 48, refresh = false) {
    return request<IndicatorSeriesResult>(
      `/indicator-series?symbol=${symbol}&interval=${interval}&limit=${limit}${refresh ? "&refresh=1" : ""}`,
    );
  },
  getOrderFlow(symbol: string, interval: MarketInterval = "1m") {
    return request<OrderFlow>(`/orderflow?symbol=${symbol}&interval=${interval}`);
  },
  getMicrostructureEvents(symbol: string, interval: MarketInterval = "1m", limit = 20) {
    return request<MicrostructureEventsResult>(
      `/microstructure-events?symbol=${symbol}&interval=${interval}&limit=${limit}`,
    );
  },
  getStructure(symbol: string, interval: MarketInterval = "1m") {
    return request<Structure>(`/structure?symbol=${symbol}&interval=${interval}`);
  },
  getStructureEvents(symbol: string, interval: MarketInterval = "1m") {
    return request<Structure>(`/market-structure-events?symbol=${symbol}&interval=${interval}`);
  },
  getStructureSeries(symbol: string, interval: MarketInterval = "1m", limit = 48) {
    return request<StructureSeriesResult>(
      `/market-structure-series?symbol=${symbol}&interval=${interval}&limit=${limit}`,
    );
  },
  getLiquidity(symbol: string, interval: MarketInterval = "1m") {
    return request<Liquidity>(`/liquidity?symbol=${symbol}&interval=${interval}`);
  },
  getLiquidityMap(symbol: string, interval: MarketInterval = "1m") {
    return request<LiquidityMapResult>(`/liquidity-map?symbol=${symbol}&interval=${interval}`);
  },
  getLiquiditySeries(symbol: string, interval: MarketInterval = "1m", limit = 48, refresh = false) {
    return request<LiquiditySeriesResult>(
      `/liquidity-series?symbol=${symbol}&interval=${interval}&limit=${limit}${refresh ? "&refresh=1" : ""}`,
    );
  },
  getMarketSnapshot(symbol: string, interval: MarketInterval = "1m", limit = 48, refresh = false) {
    return request<MarketSnapshot>(
      `/market-snapshot?symbol=${symbol}&interval=${interval}&limit=${limit}${refresh ? "&refresh=1" : ""}`,
    );
  },
  createMarketSnapshotStreamUrl(symbol: string, interval: MarketInterval = "1m", limit = 48) {
    return buildWebSocketUrl(`/market-snapshot/stream?symbol=${symbol}&interval=${interval}&limit=${limit}`);
  },
  openMarketSnapshotStream(symbol: string, interval: MarketInterval = "1m", limit = 48) {
    return new WebSocket(this.createMarketSnapshotStreamUrl(symbol, interval, limit));
  },
};

export const signalApi = {
  getSignal(symbol: string, interval: MarketInterval = "1m") {
    return request<SignalBundle>(`/signal?symbol=${symbol}&interval=${interval}`);
  },
  getSignalTimeline(symbol: string, interval: MarketInterval = "1m", limit = 48, refresh = false) {
    return request<SignalTimelineResult>(
      `/signal-timeline?symbol=${symbol}&interval=${interval}&limit=${limit}${refresh ? "&refresh=1" : ""}`,
    );
  },
};

export const authApi = {
  login(username: string, password: string) {
    return request<AuthSessionState>("/auth/login", {
      method: "POST",
      redirectOnUnauthorized: false,
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        username,
        password,
      }),
    });
  },
  logout() {
    return request<AuthSessionState>("/auth/logout", {
      method: "POST",
      redirectOnUnauthorized: false,
    });
  },
  getSession() {
    return request<AuthSessionState>("/auth/session", {
      redirectOnUnauthorized: false,
    });
  },
};
