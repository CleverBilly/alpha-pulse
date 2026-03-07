"use client";

import { useMarketStore } from "@/store/marketStore";

export default function LiquidityPanel() {
  const { liquidity, refreshDashboard } = useMarketStore();

  return (
    <section className="rounded-2xl bg-panel p-5 shadow-panel">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-lg font-semibold">Liquidity</h3>
        <button
          onClick={() => {
            void refreshDashboard();
          }}
          className="rounded-lg border border-slate-200 px-3 py-1 text-sm"
        >
          更新
        </button>
      </div>

      {liquidity ? (
        <div className="space-y-3 text-sm">
          <Metric label="Buy Liquidity" value={liquidity.buy_liquidity} />
          <Metric label="Sell Liquidity" value={liquidity.sell_liquidity} />
          <div className="rounded-lg border border-slate-100 bg-slate-50 p-3">
            <p className="text-xs text-muted">Sweep Type</p>
            <p className="mt-1 font-semibold">{liquidity.sweep_type}</p>
          </div>
        </div>
      ) : (
        <p className="text-sm text-muted">暂无流动性数据</p>
      )}
    </section>
  );
}

function Metric({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-lg border border-slate-100 bg-slate-50 p-3">
      <p className="text-xs text-muted">{label}</p>
      <p className="mt-1 font-semibold">{value.toFixed(2)}</p>
    </div>
  );
}
