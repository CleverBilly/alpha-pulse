"use client";

import { useMarketStore } from "@/store/marketStore";

export default function SignalCard() {
  const { signal, indicator, loading, error, refreshDashboard } = useMarketStore();

  const action = signal?.signal ?? "NEUTRAL";
  const badgeClass =
    action === "BUY"
      ? "bg-positive/10 text-positive"
      : action === "SELL"
        ? "bg-negative/10 text-negative"
        : "bg-slate-200 text-slate-700";

  return (
    <section className="rounded-2xl bg-panel p-5 shadow-panel">
      <div className="mb-4 flex items-center justify-between gap-3">
        <h3 className="text-lg font-semibold">Signal</h3>
        <button
          onClick={() => {
            void refreshDashboard();
          }}
          className="rounded-lg border border-slate-200 px-3 py-1 text-sm"
        >
          刷新信号
        </button>
      </div>

      {loading && !signal ? <p className="text-sm text-muted">加载中...</p> : null}
      {error ? <p className="text-sm text-negative">{error}</p> : null}
      {!loading && !error && !signal ? <p className="text-sm text-muted">暂无信号数据</p> : null}

      {signal ? (
        <div className="space-y-4">
          <div className="flex items-center justify-between gap-3">
            <span className={`inline-block rounded-full px-3 py-1 text-xs font-semibold ${badgeClass}`}>
              {action}
            </span>
            <span className="text-xs text-muted">{signal.symbol}</span>
          </div>

          <p className="text-sm leading-6 text-muted">{signal.explain}</p>

          <div className="grid grid-cols-2 gap-3 text-sm xl:grid-cols-4">
            <Data label="Score" value={signal.score.toString()} />
            <Data label="Confidence" value={`${signal.confidence.toFixed(0)}%`} />
            <Data label="Entry" value={formatNumber(signal.entry_price)} />
            <Data label="Target" value={formatNumber(signal.target_price)} />
            <Data label="Stop" value={formatNumber(signal.stop_loss)} />
            <Data label="R/R" value={signal.risk_reward.toFixed(2)} />
            <Data label="Trend Bias" value={signal.trend_bias || "neutral"} />
            <Data label="RSI" value={indicator ? indicator.rsi.toFixed(2) : "-"} />
          </div>

          <div className="rounded-xl border border-slate-100 bg-slate-50/80 p-4">
            <div className="mb-3 flex items-center justify-between gap-3">
              <h4 className="text-sm font-semibold">Factor Breakdown</h4>
              <span className="text-xs text-muted">{signal.factors.length} factors</span>
            </div>
            <div className="space-y-2">
              {signal.factors.map((factor) => (
                <FactorRow
                  key={factor.key}
                  name={factor.name}
                  score={factor.score}
                  reason={factor.reason}
                />
              ))}
            </div>
          </div>
        </div>
      ) : null}
    </section>
  );
}

function Data({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-slate-100 bg-slate-50 p-3">
      <p className="text-xs text-muted">{label}</p>
      <p className="mt-1 font-semibold">{value}</p>
    </div>
  );
}

function FactorRow({
  name,
  score,
  reason,
}: {
  name: string;
  score: number;
  reason: string;
}) {
  const tone =
    score > 0
      ? "bg-positive/10 text-positive border-positive/20"
      : score < 0
        ? "bg-negative/10 text-negative border-negative/20"
        : "bg-slate-100 text-slate-700 border-slate-200";

  return (
    <div className="rounded-xl border border-slate-100 bg-white p-3">
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm font-semibold">{name}</p>
        <span className={`rounded-full border px-2 py-0.5 text-xs font-semibold ${tone}`}>
          {score > 0 ? `+${score}` : score}
        </span>
      </div>
      <p className="mt-2 text-xs leading-5 text-muted">{reason}</p>
    </div>
  );
}

function formatNumber(value: number): string {
  return Number.isFinite(value) ? value.toFixed(2) : "-";
}
