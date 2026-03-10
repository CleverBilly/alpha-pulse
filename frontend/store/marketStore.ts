"use client";

import { create } from "zustand";
import { marketApi } from "@/services/apiClient";
import {
  Indicator,
  IndicatorSeriesPoint,
  Kline,
  Liquidity,
  LiquiditySeriesPoint,
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

export const useMarketStore = create<MarketState>((set, get) => ({
  symbol: "BTCUSDT",
  interval: "1m",
  lastUpdatedAt: null,
  lastRefreshMode: null,
  transportMode: "idle",
  streamStatus: "idle",
  streamError: null,
  price: null,
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
    try {
      const { symbol, interval } = get();
      set({ loading: true, error: null });
      const snapshot = await marketApi.getMarketSnapshot(symbol, interval, 48, refresh);
      get().applySnapshot(snapshot, { refresh });
    } catch (error) {
      set({ loading: false, error: formatError(error) });
    }
  },
}));

function formatError(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  return "unknown error";
}
