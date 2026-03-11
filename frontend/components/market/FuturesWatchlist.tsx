"use client";

import { useEffect, useState } from "react";
import { Card, Tag } from "antd";
import { buildDashboardDecision } from "@/components/dashboard/dashboardViewModel";
import { marketApi } from "@/services/apiClient";
import { useMarketStore } from "@/store/marketStore";
import { MARKET_SYMBOLS } from "@/types/market";
import type { MarketSnapshot } from "@/types/snapshot";

const WATCHLIST_INTERVAL = "1h";
const WATCHLIST_LIMIT = 24;
const WATCHLIST_REFRESH_INTERVAL_MS = 15_000;

type WatchlistStatus = "loading" | "ready" | "error";

type WatchlistItem = {
  symbol: string;
  status: WatchlistStatus;
  snapshot: MarketSnapshot | null;
  error: string | null;
};

export default function FuturesWatchlist() {
  const { symbol: activeSymbol, setSymbol } = useMarketStore();
  const [items, setItems] = useState<WatchlistItem[]>(
    MARKET_SYMBOLS.map((symbol) => ({
      symbol,
      status: "loading",
      snapshot: null,
      error: null,
    })),
  );

  useEffect(() => {
    let active = true;

    const loadSnapshots = async () => {
      const nextItems = await Promise.all(
        MARKET_SYMBOLS.map(async (symbol) => {
          try {
            const snapshot = await marketApi.getMarketSnapshot(symbol, WATCHLIST_INTERVAL, WATCHLIST_LIMIT);
            return {
              symbol,
              status: "ready" as const,
              snapshot,
              error: null,
            };
          } catch (error) {
            return {
              symbol,
              status: "error" as const,
              snapshot: null,
              error: formatError(error),
            };
          }
        }),
      );

      if (active) {
        setItems(nextItems);
      }
    };

    void loadSnapshots();
    const timer = window.setInterval(() => {
      void loadSnapshots();
    }, WATCHLIST_REFRESH_INTERVAL_MS);

    return () => {
      active = false;
      window.clearInterval(timer);
    };
  }, []);

  return (
    <section aria-label="Futures Watchlist" className="space-y-4">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
        <div className="max-w-3xl">
          <p className="text-xs font-semibold uppercase tracking-[0.24em] text-sky-700">Futures Watchlist</p>
          <h2 className="mt-3 text-[30px] font-semibold leading-tight tracking-[-0.04em] text-slate-950">
            BTC / ETH / SOL 的 1h 主方向与期货因子概览
          </h2>
          <p className="mt-3 text-[15px] leading-7 text-slate-600">
            这条 watchlist 固定用 1h 快照做主判断，优先告诉你哪边更强、Futures 因子是否支持，以及要不要切到该标的细看。
          </p>
        </div>
        <Tag color="geekblue">1h Direction Core</Tag>
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        {items.map((item) => (
          <WatchlistCard
            key={item.symbol}
            item={item}
            active={item.symbol === activeSymbol}
            onSelect={() => setSymbol(item.symbol)}
          />
        ))}
      </div>
    </section>
  );
}

function WatchlistCard({
  item,
  active,
  onSelect,
}: {
  item: WatchlistItem;
  active: boolean;
  onSelect: () => void;
}) {
  const futures = item.snapshot?.futures;
  const decision = buildDashboardDecision({
    signal: item.snapshot?.signal,
    structure: item.snapshot?.structure,
    liquidity: item.snapshot?.liquidity,
    orderFlow: item.snapshot?.orderflow,
  });

  return (
    <Card variant="borderless" className="surface-card surface-card--market">
      <div className="flex flex-col gap-4">
        <div className="flex items-start justify-between gap-4">
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="text-xl font-semibold tracking-[-0.03em] text-slate-950">{item.symbol}</h3>
              {active ? <Tag color="cyan">当前盘面</Tag> : null}
              <Tag color={resolveAntTone(decision.tone)}>{decision.verdict}</Tag>
            </div>
            <p className="mt-2 text-sm text-slate-600">{item.status === "error" ? item.error : decision.summary}</p>
          </div>
          <div className={`rounded-2xl px-4 py-3 text-right ${resolveConfidenceTone(decision.tone)}`}>
            <span className="block text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">Confidence</span>
            <strong className="text-2xl tracking-[-0.03em] text-slate-950">{decision.confidence.toFixed(0)}%</strong>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <Metric label="风险" value={decision.riskLabel} />
          <Metric label="现价" value={formatPrice(item.snapshot?.price.price)} />
          <Metric label="Mark" value={formatPrice(futures?.mark_price)} />
          <Metric label="Basis" value={formatSigned(futures?.basis_bps, 1, "bps")} />
          <Metric label="Funding" value={formatPercent(futures?.funding_rate, 3)} />
          <Metric label="L/S" value={formatNumber(futures?.long_short_ratio, 2)} />
        </div>

        <div className="rounded-2xl border border-slate-200 bg-slate-50/90 px-4 py-3">
          <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">Futures Signal</p>
          <p className="mt-2 text-sm leading-6 text-slate-700">
            {futures?.available
              ? `OI ${formatCompact(futures.open_interest)}，名义价值 ${formatCompactCurrency(futures.open_interest_value)}，多头账户占比 ${formatPercent(
                  futures.long_account_ratio,
                  1,
                )}。`
              : futures?.reason || "Futures metrics unavailable"}
          </p>
        </div>

        <button
          type="button"
          onClick={onSelect}
          className="rounded-2xl bg-slate-950 px-4 py-3 text-sm font-semibold text-white transition hover:bg-slate-800"
          aria-label={`切换到 ${item.symbol}`}
        >
          切换到 {item.symbol}
        </button>
      </div>
    </Card>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-white/70 bg-white/80 px-4 py-3 shadow-[0_12px_30px_rgba(32,42,63,0.05)]">
      <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">{label}</p>
      <p className="mt-2 text-base font-semibold text-slate-950">{value}</p>
    </div>
  );
}

function formatError(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return "请求 watchlist 失败";
}

function formatPrice(value?: number | null) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return "-";
  }
  return `$${value.toFixed(value >= 1000 ? 2 : 3)}`;
}

function formatSigned(value?: number | null, digits = 2, suffix = "") {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return "-";
  }
  const prefix = value > 0 ? "+" : "";
  return `${prefix}${value.toFixed(digits)}${suffix ? ` ${suffix}` : ""}`;
}

function formatPercent(value?: number | null, digits = 2) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return "-";
  }
  return `${(value * 100).toFixed(digits)}%`;
}

function formatNumber(value?: number | null, digits = 2) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return "-";
  }
  return value.toFixed(digits);
}

function formatCompact(value?: number | null) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return "-";
  }
  return new Intl.NumberFormat("zh-CN", {
    notation: "compact",
    maximumFractionDigits: 2,
  }).format(value);
}

function formatCompactCurrency(value?: number | null) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return "-";
  }
  return `$${new Intl.NumberFormat("zh-CN", {
    notation: "compact",
    maximumFractionDigits: 2,
  }).format(value)}`;
}

function resolveAntTone(tone: "positive" | "negative" | "neutral" | "warning") {
  if (tone === "positive") {
    return "success";
  }
  if (tone === "negative") {
    return "error";
  }
  if (tone === "warning") {
    return "warning";
  }
  return "default";
}

function resolveConfidenceTone(tone: "positive" | "negative" | "neutral" | "warning") {
  if (tone === "positive") {
    return "border border-emerald-100 bg-emerald-50/80";
  }
  if (tone === "negative") {
    return "border border-rose-100 bg-rose-50/80";
  }
  if (tone === "warning") {
    return "border border-amber-100 bg-amber-50/80";
  }
  return "border border-slate-200 bg-slate-50/80";
}
