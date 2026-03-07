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
import { Signal, SignalTimelinePoint } from "@/types/signal";

interface MarketState {
  symbol: string;
  interval: MarketInterval;
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
  setSymbol: (symbol: string) => void;
  setIntervalType: (interval: MarketInterval) => void;
  refreshPrice: () => Promise<void>;
  refreshKline: () => Promise<void>;
  refreshIndicators: () => Promise<void>;
  refreshOrderFlow: () => Promise<void>;
  refreshStructure: () => Promise<void>;
  refreshLiquidity: () => Promise<void>;
  refreshDashboard: () => Promise<void>;
}

export const useMarketStore = create<MarketState>((set, get) => ({
  symbol: "BTCUSDT",
  interval: "1m",
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

  setSymbol: (symbol: string) => {
    set({ symbol });
  },

  setIntervalType: (interval: MarketInterval) => {
    set({ interval });
  },

  refreshPrice: async () => {
    await get().refreshDashboard();
  },

  refreshKline: async () => {
    await get().refreshDashboard();
  },

  refreshIndicators: async () => {
    await get().refreshDashboard();
  },

  refreshOrderFlow: async () => {
    await get().refreshDashboard();
  },

  refreshStructure: async () => {
    await get().refreshDashboard();
  },

  refreshLiquidity: async () => {
    await get().refreshDashboard();
  },

  refreshDashboard: async () => {
    try {
      const { symbol, interval } = get();
      set({ loading: true, error: null });
      const snapshot = await marketApi.getMarketSnapshot(symbol, interval, 48, true);
      const latestKline = snapshot.klines[snapshot.klines.length - 1] ?? null;

      set({
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
      });
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
