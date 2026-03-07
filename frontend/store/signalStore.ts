"use client";

import { create } from "zustand";
import { signalApi } from "@/services/apiClient";
import { MarketInterval } from "@/types/market";
import { SignalBundle } from "@/types/signal";

interface SignalState {
  symbol: string;
  interval: MarketInterval;
  signalBundle: SignalBundle | null;
  loading: boolean;
  error: string | null;
  setSymbol: (symbol: string) => void;
  setIntervalType: (interval: MarketInterval) => void;
  refreshSignal: (symbol?: string) => Promise<void>;
}

export const useSignalStore = create<SignalState>((set, get) => ({
  symbol: "BTCUSDT",
  interval: "1m",
  signalBundle: null,
  loading: false,
  error: null,

  setSymbol: (symbol: string) => {
    set({ symbol });
    void get().refreshSignal(symbol);
  },

  setIntervalType: (interval: MarketInterval) => {
    set({ interval });
    void get().refreshSignal();
  },

  refreshSignal: async (symbol?: string) => {
    try {
      set({ loading: true, error: null });
      const targetSymbol = symbol ?? get().symbol;
      const { interval } = get();
      const data = await signalApi.getSignal(targetSymbol, interval);
      set({ signalBundle: data, loading: false });
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
