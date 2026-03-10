"use client";

import { useMemo } from "react";
import { Card, Tag, Typography } from "antd";
import { useMarketStore } from "@/store/marketStore";

export default function MarketLevelsBoard() {
  const { price, structure, liquidity, signal } = useMarketStore();

  const levelRows = useMemo(
    () => [
      { label: "Support", value: structure?.support, tone: "bg-emerald-50 text-emerald-700 border-emerald-100" },
      { label: "Resistance", value: structure?.resistance, tone: "bg-rose-50 text-rose-700 border-rose-100" },
      { label: "Buy Liquidity", value: liquidity?.buy_liquidity, tone: "bg-teal-50 text-teal-700 border-teal-100" },
      { label: "Sell Liquidity", value: liquidity?.sell_liquidity, tone: "bg-orange-50 text-orange-700 border-orange-100" },
      { label: "Equal High", value: liquidity?.equal_high, tone: "bg-amber-50 text-amber-700 border-amber-100" },
      { label: "Equal Low", value: liquidity?.equal_low, tone: "bg-cyan-50 text-cyan-700 border-cyan-100" },
      { label: "Signal Entry", value: signal?.entry_price, tone: "bg-sky-50 text-sky-700 border-sky-100" },
      { label: "Signal Target", value: signal?.target_price, tone: "bg-violet-50 text-violet-700 border-violet-100" },
    ],
    [
      liquidity?.buy_liquidity,
      liquidity?.equal_high,
      liquidity?.equal_low,
      liquidity?.sell_liquidity,
      signal?.entry_price,
      signal?.target_price,
      structure?.resistance,
      structure?.support,
    ],
  );

  return (
    <section>
      <Card
        variant="borderless"
        className="surface-card surface-card--paper"
      >
        <div className="flex items-end justify-between gap-3">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.14em] text-slate-500">
              Key Levels
            </p>
            <Typography.Title level={3} className="!mb-0 !mt-3 !text-[24px] !tracking-[-0.03em]">
              Price Ladder
            </Typography.Title>
          </div>
          <Tag>Current {formatPrice(price?.price)}</Tag>
        </div>

        <div className="mt-5 space-y-3">
          {levelRows.map((row) => (
            <div
              key={row.label}
              className={`grid grid-cols-[1.2fr_1fr_0.9fr] items-center gap-3 rounded-[22px] border px-4 py-3 ${row.tone}`}
            >
              <span className="text-sm font-medium">{row.label}</span>
              <span className="text-right text-sm font-semibold">{formatPrice(row.value)}</span>
              <span className="text-right text-xs">
                {formatDistance(price?.price, row.value)}
              </span>
            </div>
          ))}
        </div>

        <div className="mt-6 rounded-[24px] border border-slate-100 bg-slate-50 p-4">
          <h4 className="text-sm font-semibold text-slate-900">Stop Clusters</h4>
          <div className="mt-3 space-y-2">
            {(liquidity?.stop_clusters ?? []).slice(0, 6).map((cluster) => (
              <div
                key={`${cluster.kind}-${cluster.price}`}
                className="flex items-center justify-between rounded-[18px] border border-white bg-white px-3 py-2 text-sm shadow-[0_10px_24px_rgba(32,42,63,0.04)]"
              >
                <span className="text-slate-700">{cluster.label}</span>
                <span className="font-medium text-slate-900">{cluster.price.toFixed(2)}</span>
                <span className="text-xs text-slate-500">strength {cluster.strength.toFixed(2)}</span>
              </div>
            ))}
            {(liquidity?.stop_clusters ?? []).length === 0 ? (
              <p className="text-sm text-slate-500">暂无止损簇</p>
            ) : null}
          </div>
        </div>
      </Card>
    </section>
  );
}

function formatPrice(value?: number | null) {
  return typeof value === "number" && Number.isFinite(value) && value > 0 ? value.toFixed(2) : "-";
}

function formatDistance(current?: number | null, target?: number | null) {
  if (
    typeof current !== "number" ||
    typeof target !== "number" ||
    !Number.isFinite(current) ||
    !Number.isFinite(target) ||
    current <= 0 ||
    target <= 0
  ) {
    return "-";
  }

  const pct = ((target - current) / current) * 100;
  const prefix = pct > 0 ? "+" : "";
  return `${prefix}${pct.toFixed(2)}%`;
}
