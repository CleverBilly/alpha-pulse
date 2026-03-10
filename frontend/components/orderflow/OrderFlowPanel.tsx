"use client";

import { Button, Card, Tag, Typography } from "antd";
import { useMarketStore } from "@/store/marketStore";

export default function OrderFlowPanel() {
  const { orderFlow, microstructureEvents = [], refreshDashboard } = useMarketStore();
  const largeTrades = orderFlow?.large_trades ?? [];
  const microEvents =
    microstructureEvents.length > 0 ? microstructureEvents : (orderFlow?.microstructure_events ?? []);

  return (
    <section>
      <Card
        variant="borderless"
        className="surface-card surface-card--paper"
      >
        <div className="mb-5 flex items-center justify-between gap-3">
          <Typography.Title level={3} className="!mb-0 !text-[24px] !tracking-[-0.03em]">
            Order Flow
          </Typography.Title>
          <Button
            onClick={() => {
              void refreshDashboard(true);
            }}
            className="!rounded-2xl !border-slate-200 !bg-white/80"
          >
            更新
          </Button>
        </div>

        {orderFlow ? (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-3 text-sm">
              <Metric label="Buy Volume" value={orderFlow.buy_volume} accent="emerald" />
              <Metric label="Sell Volume" value={orderFlow.sell_volume} accent="rose" />
              <Metric label="Delta" value={orderFlow.delta} accent="teal" />
              <Metric label="CVD" value={orderFlow.cvd} />
              <Metric label="Large Buy" value={orderFlow.buy_large_trade_notional} accent="emerald" />
              <Metric label="Large Sell" value={orderFlow.sell_large_trade_notional} accent="rose" />
              <Metric label="Absorption" value={orderFlow.absorption_strength} digits={3} accent="amber" />
              <Metric label="Iceberg" value={orderFlow.iceberg_strength} digits={3} accent="violet" />
            </div>

            <div className="rounded-[26px] border border-slate-100 bg-slate-50/80 p-4">
              <div className="mb-3 flex flex-wrap items-center gap-2 text-xs">
                <ToneTag label={`Source: ${orderFlow.data_source || "unknown"}`} color="blue" />
                <ToneTag label={`Absorption: ${orderFlow.absorption_bias || "none"}`} color={flowTone(orderFlow.absorption_bias)} />
                <ToneTag label={`Iceberg: ${orderFlow.iceberg_bias || "none"}`} color={flowTone(orderFlow.iceberg_bias)} />
              </div>
              {largeTrades.length > 0 ? (
                <div className="space-y-2">
                  {largeTrades.slice(-4).map((trade) => (
                    <div
                      key={`${trade.trade_time}-${trade.side}-${trade.price}`}
                      className="flex items-center justify-between rounded-[20px] border border-slate-100 bg-white px-3 py-3 text-xs shadow-[0_10px_24px_rgba(32,42,63,0.04)]"
                    >
                      <span className={trade.side === "buy" ? "text-positive" : "text-negative"}>
                        {trade.side.toUpperCase()}
                      </span>
                      <span>{trade.price.toFixed(2)}</span>
                      <span>{trade.notional.toFixed(0)} USDT</span>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-xs text-muted">暂无大单事件</p>
              )}
            </div>

            <div className="rounded-[26px] border border-slate-100 bg-slate-50/80 p-4">
              <div className="mb-3 flex items-center justify-between gap-3">
                <h4 className="text-sm font-semibold text-slate-900">Microstructure Events</h4>
                <span className="text-xs text-slate-500">
                  {microEvents.length} items
                </span>
              </div>
              {microEvents.length > 0 ? (
                <div className="space-y-2">
                  {microEvents.slice(-4).reverse().map((event) => (
                    <div
                      key={`${event.trade_time}-${event.type}-${event.bias}`}
                      className="rounded-[22px] border border-slate-100 bg-white px-3 py-3 shadow-[0_10px_24px_rgba(32,42,63,0.04)]"
                    >
                      <div className="flex items-center justify-between gap-3">
                        <div className="flex items-center gap-2">
                          <span className={`rounded-full px-2 py-0.5 text-[11px] font-semibold ${eventTone(event.bias)}`}>
                            {event.bias}
                          </span>
                          <span className="text-sm font-semibold text-slate-900">
                            {formatEventType(event.type)}
                          </span>
                        </div>
                        <span className="text-xs text-slate-500">
                          {event.score > 0 ? `+${event.score}` : event.score}
                        </span>
                      </div>
                      <p className="mt-2 text-xs leading-5 text-slate-600">{event.detail}</p>
                      <div className="mt-2 flex items-center justify-between text-[11px] text-slate-500">
                        <span>price {event.price.toFixed(2)}</span>
                        <span>strength {event.strength.toFixed(2)}</span>
                        <span>{formatEventTime(event.trade_time)}</span>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-xs text-muted">暂无微结构事件</p>
              )}
            </div>
          </div>
        ) : (
          <p className="text-sm text-muted">暂无订单流数据</p>
        )}
      </Card>
    </section>
  );
}

function Metric({
  label,
  value,
  digits = 2,
  accent,
}: {
  label: string;
  value: number;
  digits?: number;
  accent?: "emerald" | "rose" | "teal" | "amber" | "violet";
}) {
  const tone =
    accent === "emerald"
      ? "border-emerald-100 bg-emerald-50/60"
      : accent === "rose"
        ? "border-rose-100 bg-rose-50/60"
        : accent === "teal"
          ? "border-teal-100 bg-teal-50/60"
          : accent === "amber"
            ? "border-amber-100 bg-amber-50/60"
            : accent === "violet"
              ? "border-violet-100 bg-violet-50/60"
              : "border-slate-100 bg-white/76";

  return (
    <div className={`rounded-[22px] border p-3 shadow-[0_10px_24px_rgba(32,42,63,0.04)] ${tone}`}>
      <p className="text-[11px] font-semibold uppercase tracking-[0.16em] text-muted">{label}</p>
      <p className="mt-2 font-semibold text-slate-900">{value.toFixed(digits)}</p>
    </div>
  );
}

function ToneTag({ label, color }: { label: string; color?: string }) {
  return <Tag color={color}>{label}</Tag>;
}

function eventTone(bias: string) {
  if (bias === "bullish") {
    return "bg-emerald-50 text-emerald-700";
  }
  if (bias === "bearish") {
    return "bg-rose-50 text-rose-700";
  }
  return "bg-slate-100 text-slate-700";
}

function flowTone(value?: string) {
  if (value === "bullish") {
    return "success";
  }
  if (value === "bearish") {
    return "error";
  }
  return undefined;
}

function formatEventType(value: string) {
  return value
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function formatEventTime(timestamp: number) {
  return new Date(timestamp).toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}
