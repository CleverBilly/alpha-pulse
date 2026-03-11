import { buildMockMarketSnapshot } from "./marketSnapshot";

export function buildMockMarketStoreState(overrides: Record<string, unknown> = {}) {
  const snapshot = buildMockMarketSnapshot();
  const latestKline = snapshot.klines[snapshot.klines.length - 1] ?? null;

  return {
    symbol: snapshot.price.symbol,
    interval: snapshot.signal.interval_type,
    lastUpdatedAt: snapshot.price.time,
    lastRefreshMode: "cache",
    transportMode: "websocket",
    streamStatus: "live",
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
    applySnapshot: () => undefined,
    setStreamState: () => undefined,
    setSymbol: () => undefined,
    setIntervalType: () => undefined,
    refreshPrice: async () => undefined,
    refreshKline: async () => undefined,
    refreshIndicators: async () => undefined,
    refreshOrderFlow: async () => undefined,
    refreshStructure: async () => undefined,
    refreshLiquidity: async () => undefined,
    refreshDashboard: async (_refresh?: boolean) => undefined,
    ...overrides,
  };
}
