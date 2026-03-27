"use client";

import { Button, Tag } from "antd";
import { HistoryOutlined } from "@ant-design/icons";
import type { AlertEvent } from "@/types/alert";

interface AlertEventCardProps {
  item: AlertEvent;
  onReview?: (event: AlertEvent) => void;
  variant?: "card" | "row";
}

export default function AlertEventCard({ item, onReview, variant = "card" }: AlertEventCardProps) {
  const rowMode = variant === "row";

  return (
    <article
      className={rowMode
        ? "alert-event-card alert-event-card--row"
        : `alert-event-card alert-event-card--card rounded-3xl border px-4 py-4 shadow-[0_12px_30px_rgba(32,42,63,0.05)] ${resolveCardTone(item.severity)}`}
      data-alert-severity={item.severity}
    >
      <div className="alert-event-card__head">
        <div>
          <div className="alert-event-card__meta-row">
            <span className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">{item.symbol}</span>
            <Tag color={resolveSeverityColor(item.severity)}>{resolveSeverityLabel(item.severity)}</Tag>
            <Tag color={item.tradeability_label === "No-Trade" || item.tradeability_label === "禁止交易" ? "warning" : "success"}>
              {item.tradeability_label}
            </Tag>
          </div>
          <h3 className="alert-event-card__title">{item.title}</h3>
        </div>
        <div className="alert-event-card__actions">
          <span className="alert-event-card__time">{formatAlertTime(item.created_at)}</span>
          {onReview && (
            <Button
              size="small"
              icon={<HistoryOutlined />}
              onClick={() => onReview(item)}
              style={{ marginLeft: 8 }}
            >
              复盘
            </Button>
          )}
        </div>
      </div>

      <p className="alert-event-card__summary">
        {item.verdict} · {item.summary}
      </p>

      <div className="alert-event-card__timeframes">
        {item.timeframe_labels.map((label) => (
          <span
            key={label}
            className="rounded-full border border-slate-200 bg-white/85 px-3 py-1 text-[12px] font-medium text-slate-700"
          >
            {label}
          </span>
        ))}
      </div>

      <div className="alert-event-card__metrics">
        <Metric label="置信度" value={`${item.confidence}%`} variant={variant} />
        <Metric label="风险" value={item.risk_label} variant={variant} />
        <Metric label="进场位" value={formatLevel(item.entry_price)} variant={variant} />
        <Metric label="目标位" value={formatLevel(item.target_price)} variant={variant} />
      </div>

      {item.reasons.length > 0 ? <p className="alert-event-card__reasons">原因：{item.reasons.join("；")}</p> : null}

      {item.deliveries.length > 0 && !rowMode ? (
        <div className="mt-3 flex flex-wrap gap-2">
          {item.deliveries.map((delivery) => (
            <Tag
              key={`${item.id}-${delivery.channel}`}
              color={delivery.status === "sent" ? "success" : delivery.status === "failed" ? "error" : "default"}
            >
              {delivery.channel}:{formatDeliveryStatus(delivery.status)}
            </Tag>
          ))}
        </div>
      ) : null}
    </article>
  );
}

function Metric({ label, value, variant }: { label: string; value: string; variant: "card" | "row" }) {
  return (
    <div className={variant === "row" ? "alert-event-card__metric alert-event-card__metric--row" : "rounded-2xl border border-white/80 bg-white/80 px-3 py-3"}>
      <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">{label}</p>
      <p className="mt-2 text-sm font-semibold text-slate-950">{value}</p>
    </div>
  );
}

export function resolveSeverityColor(severity: string) {
  if (severity === "A") {
    return "gold";
  }
  if (severity === "warning") {
    return "warning";
  }
  return "blue";
}

export function resolveSeverityLabel(severity: string) {
  if (severity === "A") {
    return "A 级";
  }
  if (severity === "warning") {
    return "禁止交易";
  }
  return "方向";
}

export function resolveCardTone(severity: string) {
  if (severity === "A") {
    return "border-emerald-200 bg-emerald-50/60";
  }
  if (severity === "warning") {
    return "border-amber-200 bg-amber-50/70";
  }
  return "border-sky-200 bg-sky-50/55";
}

export function formatAlertTime(timestamp: number) {
  if (!Number.isFinite(timestamp) || timestamp <= 0) {
    return "--";
  }
  return new Date(timestamp).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatLevel(value: number) {
  if (!Number.isFinite(value) || value <= 0) {
    return "--";
  }
  return value >= 1000 ? value.toFixed(2) : value.toFixed(3);
}

function formatDeliveryStatus(status: string) {
  if (status === "sent") {
    return "已发送";
  }
  if (status === "failed") {
    return "发送失败";
  }
  return "待发送";
}
