import {
  Indicator,
  IndicatorSeriesResult,
  Kline,
  Liquidity,
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

interface ApiEnvelope<T> {
  code: number;
  message: string;
  data: T;
}

async function request<T>(path: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    cache: "no-store",
  });

  if (!response.ok) {
    throw new Error(`API request failed: ${response.status}`);
  }

  const payload = (await response.json()) as ApiEnvelope<T>;
  if (payload.code !== 0) {
    throw new Error(payload.message || "API error");
  }

  return payload.data;
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
    return request<Liquidity>(`/liquidity-map?symbol=${symbol}&interval=${interval}`);
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
