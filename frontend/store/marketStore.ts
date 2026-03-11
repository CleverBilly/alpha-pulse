"use client";

import { create } from "zustand";
import { marketApi } from "@/services/apiClient";
import {
  Indicator,
  IndicatorSeriesPoint,
  Kline,
  Liquidity,
  LiquiditySeriesPoint,
  FuturesSnapshot,
  MarketInterval,
  OrderFlow,
  OrderFlowMicrostructureEvent,
  PriceTicker,
  Structure,
  StructureSeriesPoint,
} from "@/types/market";
import { MarketSnapshot } from "@/types/snapshot";
import { Signal, SignalTimelinePoint } from "@/types/signal";

export type MarketTransportMode = "idle" | "websocket" | "polling";
export type MarketStreamStatus = "idle" | "connecting" | "live" | "fallback" | "error";

interface MarketState {
  symbol: string;
  interval: MarketInterval;
  lastUpdatedAt: number | null;
  lastRefreshMode: "cache" | "force" | null;
  transportMode: MarketTransportMode;
  streamStatus: MarketStreamStatus;
  streamError: string | null;
  price: PriceTicker | null;
  futures: FuturesSnapshot | null;
  kline: Kline | null;
  klines: Kline[];
  indicator: Indicator | null;
  indicatorSeries: IndicatorSeriesPoint[];
  orderFlow: OrderFlow | null;
  microstructureEvents: OrderFlowMicrostructureEvent[];
  structure: Structure | null;
  structureSeries: StructureSeriesPoint[];
  liquidity: Liquidity | null;
  liquiditySeries: LiquiditySeriesPoint[];
  signal: Signal | null;
  signalTimeline: SignalTimelinePoint[];
  loading: boolean;
  error: string | null;
  applySnapshot: (snapshot: MarketSnapshot, options?: { refresh?: boolean; transport?: MarketTransportMode }) => void;
  setStreamState: (status: MarketStreamStatus, transport?: MarketTransportMode, error?: string | null) => void;
  setSymbol: (symbol: string) => void;
  setIntervalType: (interval: MarketInterval) => void;
  refreshPrice: () => Promise<void>;
  refreshKline: () => Promise<void>;
  refreshIndicators: () => Promise<void>;
  refreshOrderFlow: () => Promise<void>;
  refreshStructure: () => Promise<void>;
  refreshLiquidity: () => Promise<void>;
  refreshDashboard: (refresh?: boolean) => Promise<void>;
}

export const useMarketStore = create<MarketState>((set, get) => {
  // 按 "symbol:interval" 键追踪进行中的 refreshDashboard 请求，防止并发重复触发。
  const pendingRefreshMap = new Map<string, Promise<void>>();

  return {
  symbol: "BTCUSDT",
  interval: "1m",
  lastUpdatedAt: null,
  lastRefreshMode: null,
  transportMode: "idle",
  streamStatus: "idle",
  streamError: null,
  price: null,
  futures: null,
  kline: null,
  klines: [],
  indicator: null,
  indicatorSeries: [],
  orderFlow: null,
  microstructureEvents: [],
  structure: null,
  structureSeries: [],
  liquidity: null,
  liquiditySeries: [],
  signal: null,
  signalTimeline: [],
  loading: false,
  error: null,

  applySnapshot: (snapshot, options) => {
    const latestKline = snapshot.klines[snapshot.klines.length - 1] ?? null;
    set((state) => ({
      lastUpdatedAt: snapshot.price?.time ?? Date.now(),
      lastRefreshMode: options?.refresh ? "force" : "cache",
      transportMode: options?.transport ?? state.transportMode,
      streamError: null,
      price: snapshot.price,
      futures: snapshot.futures,
      kline: latestKline,
      klines: snapshot.klines,
      indicator: snapshot.indicator,
      indicatorSeries: snapshot.indicator_series,
      orderFlow: snapshot.orderflow,
      microstructureEvents: snapshot.microstructure_events,
      structure: snapshot.structure,
      structureSeries: snapshot.structure_series,
      liquidity: snapshot.liquidity,
      liquiditySeries: snapshot.liquidity_series,
      signal: snapshot.signal,
      signalTimeline: snapshot.signal_timeline,
      loading: false,
      error: null,
    }));
  },

  setStreamState: (status, transport, streamError = null) => {
    set((state) => ({
      streamStatus: status,
      transportMode: transport ?? state.transportMode,
      streamError,
    }));
  },

  setSymbol: (symbol: string) => {
    set({ symbol });
  },

  setIntervalType: (interval: MarketInterval) => {
    set({ interval });
  },

  refreshPrice: async () => {
    await get().refreshDashboard(true);
  },

  refreshKline: async () => {
    await get().refreshDashboard(true);
  },

  refreshIndicators: async () => {
    await get().refreshDashboard(true);
  },

  refreshOrderFlow: async () => {
    await get().refreshDashboard(true);
  },

  refreshStructure: async () => {
    await get().refreshDashboard(true);
  },

  refreshLiquidity: async () => {
    await get().refreshDashboard(true);
  },

  refreshDashboard: async (refresh = false) => {
    const { symbol, interval } = get();
    const key = `${symbol}:${interval}`;

    const existing = pendingRefreshMap.get(key);
    if (existing && !refresh) {
      return existing;
    }

    // 使用 let + 明确赋值断言，允许 finally 块通过闭包引用 promise 自身做幂等清理。
    // finally 仅在异步完成后执行，此时 promise 已赋值，类型安全。
    // eslint-disable-next-line prefer-const
    let promise!: Promise<void>;
    promise = (async () => {
      try {
        set({ loading: true, error: null });
        const snapshot = await marketApi.getMarketSnapshot(symbol, interval, 48, refresh);
        get().applySnapshot(snapshot, { refresh });
      } catch (error) {
        set({ loading: false, error: formatError(error) });
      } finally {
        if (pendingRefreshMap.get(key) === promise) {
          pendingRefreshMap.delete(key);
        }
      }
    })();

    pendingRefreshMap.set(key, promise);
    return promise;
  },
  };
});

function formatError(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  return "unknown error";
}
