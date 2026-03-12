"use client";

import { Tag } from "antd";
import type { AlertEvent } from "@/types/alert";

export default function AlertEventCard({ item }: { item: AlertEvent }) {
  return (
    <article
      className={`rounded-3xl border px-4 py-4 shadow-[0_12px_30px_rgba(32,42,63,0.05)] ${resolveCardTone(item.severity)}`}
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">{item.symbol}</span>
            <Tag color={resolveSeverityColor(item.severity)}>{resolveSeverityLabel(item.severity)}</Tag>
            <Tag color={item.tradeability_label === "No-Trade" || item.tradeability_label === "禁止交易" ? "warning" : "success"}>
              {item.tradeability_label}
            </Tag>
          </div>
          <h3 className="mt-2 text-base font-semibold tracking-[-0.02em] text-slate-950">{item.title}</h3>
        </div>
        <span className="text-xs text-slate-500">{formatAlertTime(item.created_at)}</span>
      </div>

      <p className="mt-3 text-sm leading-6 text-slate-700">
        {item.verdict} · {item.summary}
      </p>

      <div className="mt-3 flex flex-wrap gap-2">
        {item.timeframe_labels.map((label) => (
          <span
            key={label}
            className="rounded-full border border-slate-200 bg-white/85 px-3 py-1 text-[12px] font-medium text-slate-700"
          >
            {label}
          </span>
        ))}
      </div>

      <div className="mt-3 grid grid-cols-2 gap-3">
        <Metric label="置信度" value={`${item.confidence}%`} />
        <Metric label="风险" value={item.risk_label} />
        <Metric label="进场位" value={formatLevel(item.entry_price)} />
        <Metric label="目标位" value={formatLevel(item.target_price)} />
      </div>

      {item.reasons.length > 0 ? <p className="mt-3 text-sm leading-6 text-slate-600">原因：{item.reasons.join("；")}</p> : null}

      {item.deliveries.length > 0 ? (
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

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-white/80 bg-white/80 px-3 py-3">
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
