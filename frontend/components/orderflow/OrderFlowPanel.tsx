"use client";

import { useMarketStore } from "@/store/marketStore";

export default function OrderFlowPanel() {
  const { orderFlow, microstructureEvents = [], refreshDashboard } = useMarketStore();
  const largeTrades = orderFlow?.large_trades ?? [];
  const microEvents =
    microstructureEvents.length > 0 ? microstructureEvents : (orderFlow?.microstructure_events ?? []);

  return (
    <section className="rounded-2xl bg-panel p-5 shadow-panel">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-lg font-semibold">Order Flow</h3>
        <button
          onClick={() => {
            void refreshDashboard(true);
          }}
          className="rounded-lg border border-slate-200 px-3 py-1 text-sm"
        >
          更新
        </button>
      </div>

      {orderFlow ? (
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-3 text-sm">
            <Metric label="Buy Volume" value={orderFlow.buy_volume} />
            <Metric label="Sell Volume" value={orderFlow.sell_volume} />
            <Metric label="Delta" value={orderFlow.delta} />
            <Metric label="CVD" value={orderFlow.cvd} />
            <Metric label="Large Buy" value={orderFlow.buy_large_trade_notional} />
            <Metric label="Large Sell" value={orderFlow.sell_large_trade_notional} />
            <Metric label="Absorption" value={orderFlow.absorption_strength} digits={3} />
            <Metric label="Iceberg" value={orderFlow.iceberg_strength} digits={3} />
          </div>

          <div className="rounded-xl border border-slate-100 bg-slate-50/80 p-4">
            <div className="mb-2 flex flex-wrap items-center gap-2 text-xs">
              <Tag label={`Source: ${orderFlow.data_source || "unknown"}`} />
              <Tag label={`Absorption: ${orderFlow.absorption_bias || "none"}`} />
              <Tag label={`Iceberg: ${orderFlow.iceberg_bias || "none"}`} />
            </div>
            {largeTrades.length > 0 ? (
              <div className="space-y-2">
                {largeTrades.slice(-4).map((trade) => (
                  <div
                    key={`${trade.trade_time}-${trade.side}-${trade.price}`}
                    className="flex items-center justify-between rounded-lg border border-slate-100 bg-white px-3 py-2 text-xs"
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

          <div className="rounded-xl border border-slate-100 bg-slate-50/80 p-4">
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
                    className="rounded-xl border border-slate-100 bg-white px-3 py-3"
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
    </section>
  );
}

function Metric({ label, value, digits = 2 }: { label: string; value: number; digits?: number }) {
  return (
    <div className="rounded-lg border border-slate-100 bg-slate-50 p-3">
      <p className="text-xs text-muted">{label}</p>
      <p className="mt-1 font-semibold">{value.toFixed(digits)}</p>
    </div>
  );
}

function Tag({ label }: { label: string }) {
  return (
    <span className="rounded-full border border-slate-200 bg-white px-2 py-1">
      {label}
    </span>
  );
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
