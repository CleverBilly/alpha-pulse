"use client";

import { useEffect } from "react";
import { useMarketStore } from "@/store/marketStore";

const AUTO_REFRESH_INTERVAL_MS = 4000;

export default function MarketSnapshotLoader() {
  const symbol = useMarketStore((state) => state.symbol);
  const interval = useMarketStore((state) => state.interval);
  const refreshDashboard = useMarketStore((state) => state.refreshDashboard);

  useEffect(() => {
    void refreshDashboard();
    const timer = setInterval(() => {
      void refreshDashboard();
    }, AUTO_REFRESH_INTERVAL_MS);

    return () => clearInterval(timer);
  }, [refreshDashboard, symbol, interval]);

  return null;
}
