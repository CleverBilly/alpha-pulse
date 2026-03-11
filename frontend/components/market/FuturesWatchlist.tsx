"use client";

import { useEffect, useState } from "react";
import { Card, Tag } from "antd";
import { buildDirectionCopilotDecision } from "@/components/dashboard/dashboardViewModel";
import { marketApi } from "@/services/apiClient";
import { useMarketStore } from "@/store/marketStore";
import { MARKET_SYMBOLS } from "@/types/market";
import type { MarketSnapshot } from "@/types/snapshot";

const WATCHLIST_INTERVALS = {
  macro: "4h",
  bias: "1h",
  trigger: "15m",
  execution: "5m",
} as const;
const WATCHLIST_LIMIT = 24;
const WATCHLIST_REFRESH_INTERVAL_MS = 15_000;

type WatchlistStatus = "loading" | "ready" | "error";

type WatchlistSnapshots = {
  macro: MarketSnapshot | null;
  bias: MarketSnapshot | null;
  trigger: MarketSnapshot | null;
  execution: MarketSnapshot | null;
};

type WatchlistItem = {
  symbol: string;
  status: WatchlistStatus;
  snapshots: WatchlistSnapshots;
  error: string | null;
};

export default function FuturesWatchlist() {
  const { symbol: activeSymbol, setSymbol } = useMarketStore();
  const [items, setItems] = useState<WatchlistItem[]>(
    MARKET_SYMBOLS.map((symbol) => ({
      symbol,
      status: "loading",
        snapshots: {
          macro: null,
          bias: null,
          trigger: null,
          execution: null,
        },
      error: null,
    })),
  );

  useEffect(() => {
    let active = true;

    const loadSnapshots = async () => {
      const nextItems = await Promise.all(
        MARKET_SYMBOLS.map(async (symbol) => {
          try {
            const [macro, bias, trigger, execution] = await Promise.all([
              marketApi.getMarketSnapshot(symbol, WATCHLIST_INTERVALS.macro, WATCHLIST_LIMIT),
              marketApi.getMarketSnapshot(symbol, WATCHLIST_INTERVALS.bias, WATCHLIST_LIMIT),
              marketApi.getMarketSnapshot(symbol, WATCHLIST_INTERVALS.trigger, WATCHLIST_LIMIT),
              marketApi.getMarketSnapshot(symbol, WATCHLIST_INTERVALS.execution, WATCHLIST_LIMIT),
            ]);
            return {
              symbol,
              status: "ready" as const,
              snapshots: {
                macro,
                bias,
                trigger,
                execution,
              },
              error: null,
            };
          } catch (error) {
            return {
              symbol,
              status: "error" as const,
              snapshots: {
                macro: null,
                bias: null,
                trigger: null,
                execution: null,
              },
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
            BTC / ETH / SOL 的 4h / 1h / 15m / 5m 多周期方向雷达
          </h2>
          <p className="mt-3 text-[15px] leading-7 text-slate-600">
            这条 watchlist 会先看 4h 大方向，再用 1h 判断主 bias，用 15m 检查触发是否跟上，最后再用 5m 判断执行是否已经拧回去，优先告诉你哪个标的能跟，哪个标的该直接 No-Trade。
          </p>
        </div>
        <Tag color="geekblue">4h / 1h / 15m / 5m Copilot</Tag>
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
  const decision = buildDirectionCopilotDecision({
    macroSnapshot: item.snapshots.macro,
    biasSnapshot: item.snapshots.bias,
    triggerSnapshot: item.snapshots.trigger,
    executionSnapshot: item.snapshots.execution,
  });
  const biasSnapshot = item.snapshots.bias;
  const futures = biasSnapshot?.futures;
  const price = biasSnapshot?.price;
  const confidence = item.status === "ready" ? `${decision.confidence.toFixed(0)}%` : "--";

  return (
    <Card variant="borderless" className="surface-card surface-card--market">
      <div className="flex flex-col gap-4">
        <div className="flex items-start justify-between gap-4">
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="text-xl font-semibold tracking-[-0.03em] text-slate-950">{item.symbol}</h3>
              {active ? <Tag color="cyan">当前盘面</Tag> : null}
              <Tag color={resolveAntTone(decision.tone)}>{decision.verdict}</Tag>
              <Tag color={decision.tradable ? "success" : "warning"}>{decision.tradeabilityLabel}</Tag>
            </div>
            <p className="mt-2 text-sm text-slate-600">{item.status === "error" ? item.error : decision.summary}</p>
          </div>
          <div className={`rounded-2xl px-4 py-3 text-right ${resolveConfidenceTone(decision.tone)}`}>
            <span className="block text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">Confidence</span>
            <strong className="text-2xl tracking-[-0.03em] text-slate-950">{confidence}</strong>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <Metric label="风险" value={decision.riskLabel} />
          <Metric label="执行" value={decision.tradeabilityLabel} />
          <Metric label="现价" value={formatPrice(price?.price)} />
          <Metric label="Mark" value={formatPrice(futures?.mark_price)} />
          <Metric label="Basis" value={formatSigned(futures?.basis_bps, 1, "bps")} />
          <Metric label="Funding" value={formatPercent(futures?.funding_rate, 3)} />
          <Metric label="L/S" value={formatNumber(futures?.long_short_ratio, 2)} />
        </div>

        <div className="flex flex-wrap gap-2">
          {decision.timeframeLabels.map((label) => (
            <span
              key={label}
              className="rounded-full border border-slate-200 bg-white/80 px-3 py-1 text-[12px] font-medium text-slate-700"
            >
              {label}
            </span>
          ))}
        </div>

        <div className="rounded-2xl border border-slate-200 bg-slate-50/90 px-4 py-3">
          <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">Direction Reason</p>
          <p className="mt-2 text-sm leading-6 text-slate-700">
            {decision.reasons.length > 0 ? decision.reasons.join(" · ") : "等待多周期信号同步。"}
          </p>
        </div>

        <div className="rounded-2xl border border-slate-200 bg-slate-50/90 px-4 py-3">
          <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">Futures Signal</p>
          <p className="mt-2 text-sm leading-6 text-slate-700">
            {futures?.available
              ? `${futures.liquidation_summary || "清算压力代理保持均衡。"} OI ${formatCompact(futures.open_interest)}，名义价值 ${formatCompactCurrency(
                  futures.open_interest_value,
                )}，多头账户占比 ${formatPercent(futures.long_account_ratio, 1)}。`
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
