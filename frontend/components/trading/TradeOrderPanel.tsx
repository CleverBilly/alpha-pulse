import CommandPanel from "@/components/layout/CommandPanel";
import type { TradeOrder } from "@/types/trade";

interface TradeOrderPanelProps {
  closingOrderId?: number | null;
  loading?: boolean;
  onClose: (orderId: number) => void;
  orders: TradeOrder[];
}

const GROUPS = [
  { key: "pending_fill", title: "挂单中" },
  { key: "open", title: "持仓中" },
  { key: "failed", title: "失败记录" },
  { key: "settled", title: "已收口" },
] as const;

export default function TradeOrderPanel({
  closingOrderId = null,
  loading = false,
  onClose,
  orders,
}: TradeOrderPanelProps) {
  const groupedOrders = {
    pending_fill: orders.filter((order) => order.status === "pending_fill"),
    open: orders.filter((order) => order.status === "open"),
    failed: orders.filter((order) => order.status === "failed"),
    settled: orders.filter((order) => order.status === "closed" || order.status === "expired"),
  };

  return (
    <CommandPanel
      className="trade-order-panel"
      data-surface="console"
      data-testid="trade-order-panel"
      variant="control"
    >
      <div className="trade-order-panel__header">
        <div>
          <p className="trade-order-panel__eyebrow">订单状态</p>
          <h2 className="trade-order-panel__title">挂单、持仓、失败与收口统一归档</h2>
        </div>
        <span className="trade-order-panel__count">{orders.length} 条记录</span>
      </div>

      <p className="trade-order-panel__description">
        自动单和 Binance 侧识别出的手动单会统一进入这套状态面板，便于你确认限价单有没有成交、保护单是否挂上，以及仓位最终如何收口。
      </p>

      <div className="trade-order-panel__lanes">
        {GROUPS.map((group) => {
          const items = [...groupedOrders[group.key]].sort((a, b) => b.created_at - a.created_at);
          return (
            <section key={group.key} className="trade-order-panel__lane" aria-label={group.title}>
              <div className="trade-order-panel__lane-head">
                <h3>{group.title}</h3>
                <span>{items.length}</span>
              </div>

              <div className="trade-order-panel__lane-body">
                {loading ? (
                  <div className="trade-order-panel__empty">正在同步交易记录…</div>
                ) : items.length === 0 ? (
                  <div className="trade-order-panel__empty">当前没有对应记录。</div>
                ) : (
                  items.map((order) => (
                    <article key={order.id} className="trade-order-panel__card">
                      <div className="trade-order-panel__card-head">
                        <div>
                          <p>{order.symbol}</p>
                          <strong>{formatSide(order.side)}</strong>
                        </div>
                        <span className={`trade-order-panel__source trade-order-panel__source--${order.source}`}>
                          {order.source === "manual" ? "手动单" : "系统单"}
                        </span>
                      </div>

                      <div className="trade-order-panel__metrics">
                        <Metric label="限价" value={formatNumber(order.limit_price)} />
                        <Metric label="数量" value={formatNumber(order.requested_qty, 4)} />
                        <Metric label="成交" value={order.filled_qty > 0 ? formatNumber(order.filled_qty, 4) : "--"} />
                        <Metric label="状态" value={formatStatus(order.status)} />
                      </div>

                      <div className="trade-order-panel__meta">
                        <span>创建于 {formatDate(order.created_at)}</span>
                        {order.closed_at > 0 ? <span>收口于 {formatDate(order.closed_at)}</span> : null}
                      </div>

                      {order.close_reason ? <p className="trade-order-panel__reason">{order.close_reason}</p> : null}

                      {order.status === "open" ? (
                        <button
                          type="button"
                          className="trade-order-panel__close"
                          onClick={() => onClose(order.id)}
                          disabled={closingOrderId === order.id}
                        >
                          {closingOrderId === order.id ? "平仓中…" : "手动平仓"}
                        </button>
                      ) : null}
                    </article>
                  ))
                )}
              </div>
            </section>
          );
        })}
      </div>
    </CommandPanel>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="trade-order-panel__metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function formatSide(side: string) {
  return side === "LONG" ? "做多" : "做空";
}

function formatStatus(status: string) {
  switch (status) {
    case "pending_fill":
      return "等待成交";
    case "open":
      return "已建仓";
    case "failed":
      return "执行失败";
    case "expired":
      return "超时撤单";
    case "closed":
      return "已平仓";
    default:
      return status;
  }
}

function formatNumber(value?: number, digits = 2) {
  if (value == null || Number.isNaN(value)) {
    return "--";
  }
  return value.toFixed(digits);
}

function formatDate(timestamp: number) {
  if (!timestamp) {
    return "--";
  }
  return new Intl.DateTimeFormat("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(timestamp));
}
