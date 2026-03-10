"use client";

import { Button, Card, Progress, Tag, Typography } from "antd";
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
    <section>
      <Card
        variant="borderless"
        className="surface-card surface-card--paper"
      >
        <div className="mb-5 flex items-center justify-between gap-3">
          <Typography.Title level={3} className="!mb-0 !text-[24px] !tracking-[-0.03em]">
            Signal
          </Typography.Title>
          <Button
            onClick={() => {
              void refreshDashboard(true);
            }}
            className="!rounded-2xl !border-slate-200 !bg-white/80"
          >
            刷新信号
          </Button>
        </div>

        {loading && !signal ? <p className="text-sm text-muted">加载中...</p> : null}
        {error ? <p className="text-sm text-negative">{error}</p> : null}
        {!loading && !error && !signal ? <p className="text-sm text-muted">暂无信号数据</p> : null}

        {signal ? (
          <div className="space-y-5">
            <div className="flex flex-col gap-4 rounded-[28px] border border-white/75 bg-white/76 p-5 shadow-[0_14px_38px_rgba(32,42,63,0.06)]">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="flex flex-wrap items-center gap-2">
                  <span className={`inline-block rounded-full px-3 py-1 text-xs font-semibold ${badgeClass}`}>
                    {action}
                  </span>
                  <Tag color="blue">{signal.symbol}</Tag>
                  <Tag color={signal.trend_bias === "bullish" ? "success" : signal.trend_bias === "bearish" ? "error" : undefined}>
                    {signal.trend_bias || "neutral"}
                  </Tag>
                </div>
                <div className="text-right">
                  <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">Confidence</p>
                  <p className="mt-2 text-3xl font-semibold tracking-[-0.04em] text-slate-900">
                    {signal.confidence.toFixed(0)}%
                  </p>
                </div>
              </div>

              <Progress percent={Math.max(0, Math.min(100, signal.confidence))} showInfo={false} />
              <Typography.Paragraph className="!mb-0 !text-[14px] !leading-7 !text-slate-600">
                {signal.explain}
              </Typography.Paragraph>
            </div>

            <div className="grid grid-cols-2 gap-3 text-sm xl:grid-cols-4">
              <Data label="Score" value={signal.score.toString()} accent />
              <Data label="Confidence" value={`${signal.confidence.toFixed(0)} / 100`} />
              <Data label="Entry" value={formatNumber(signal.entry_price)} />
              <Data label="Target" value={formatNumber(signal.target_price)} />
              <Data label="Stop" value={formatNumber(signal.stop_loss)} />
              <Data label="R/R" value={signal.risk_reward.toFixed(2)} />
              <Data label="Trend Bias" value={signal.trend_bias || "neutral"} />
              <Data label="RSI" value={indicator ? indicator.rsi.toFixed(2) : "-"} />
            </div>

            <div className="rounded-[28px] border border-slate-100 bg-slate-50/82 p-4">
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
      </Card>
    </section>
  );
}

function Data({
  label,
  value,
  accent = false,
}: {
  label: string;
  value: string;
  accent?: boolean;
}) {
  return (
    <div
      className={`rounded-[22px] border p-3 ${
        accent
          ? "border-teal-100 bg-[linear-gradient(180deg,rgba(240,253,250,0.95)_0%,rgba(255,255,255,0.95)_100%)]"
          : "border-slate-100 bg-white/76"
      }`}
    >
      <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-muted">{label}</p>
      <p className="mt-2 font-semibold text-slate-900">{value}</p>
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
    <div className="rounded-[22px] border border-slate-100 bg-white p-3 shadow-[0_12px_30px_rgba(32,42,63,0.04)]">
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
