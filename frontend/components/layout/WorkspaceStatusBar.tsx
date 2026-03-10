"use client";

import { useMemo } from "react";
import { Tag } from "antd";
import { useMarketStore } from "@/store/marketStore";

export default function WorkspaceStatusBar() {
  const {
    symbol,
    interval,
    signal,
    structure,
    liquidity,
    loading,
    error,
    transportMode,
    streamStatus,
    streamError,
    lastUpdatedAt,
    lastRefreshMode,
  } = useMarketStore();

  const statusLabel =
    streamStatus === "live" && transportMode === "websocket"
      ? "Realtime feed live"
      : streamStatus === "connecting"
        ? "Realtime feed connecting"
        : streamStatus === "fallback"
          ? "Polling fallback active"
          : loading
            ? "Syncing snapshot"
            : error
              ? "Attention required"
              : "Snapshot live";
  const statusTone =
    streamStatus === "live"
      ? "success"
      : streamStatus === "connecting"
        ? "processing"
        : streamStatus === "fallback" || streamStatus === "error" || error
          ? "warning"
          : loading
            ? "processing"
            : "success";
  const updatedLabel = useMemo(() => formatUpdateLabel(lastUpdatedAt), [lastUpdatedAt]);
  const regimeLabel = useMemo(() => {
    if (structure?.trend === "uptrend") {
      return "Uptrend regime";
    }
    if (structure?.trend === "downtrend") {
      return "Downtrend regime";
    }
    return "Range regime";
  }, [structure?.trend]);

  return (
    <section className="surface-panel surface-panel--status">
      <div className="status-strip">
        <div className="status-strip__lead">
          <div className="status-bar__eyebrow">Workspace Status</div>
          <div className="status-strip__headline">
            <h2 className="status-strip__title">{statusLabel}</h2>
            <Tag color={statusTone}>{loading ? "Loading" : error ? "Issue" : "Healthy"}</Tag>
          </div>
          <p className="status-strip__description">
            {error || streamError
              ? error || streamError
              : `当前监控 ${symbol} ${interval}，${regimeLabel}，信号倾向 ${signal?.signal ?? "NEUTRAL"}。`}
          </p>
        </div>

        <div className="status-strip__metrics">
          <MetricPill label="Symbol" value={symbol} />
          <MetricPill label="Interval" value={interval} />
          <MetricPill label="Feed" value={formatFeedLabel(streamStatus, transportMode)} />
          <MetricPill label="Signal" value={signal?.signal ?? "NEUTRAL"} />
          <MetricPill label="Trend" value={structure?.trend ?? "range"} />
          <MetricPill label="Sweep" value={liquidity?.sweep_type || "none"} />
          <MetricPill
            label="Refresh"
            value={
              lastRefreshMode === "force"
                ? "Force rebuild"
                : lastRefreshMode === "cache"
                  ? "Cache-backed"
                  : "Waiting"
            }
          />
          <MetricPill label="Updated" value={updatedLabel} />
        </div>
      </div>
    </section>
  );
}

function MetricPill({
  label,
  value,
}: {
  label: string;
  value: string;
}) {
  return (
    <div className="status-strip__metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function formatUpdateLabel(timestamp: number | null) {
  if (!timestamp || !Number.isFinite(timestamp)) {
    return "Not synced yet";
  }

  return new Date(timestamp).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function formatFeedLabel(
  status: "idle" | "connecting" | "live" | "fallback" | "error",
  transport: "idle" | "websocket" | "polling",
) {
  if (status === "live" && transport === "websocket") {
    return "WebSocket live";
  }
  if (status === "connecting") {
    return "WebSocket connecting";
  }
  if (transport === "polling") {
    return "HTTP polling";
  }
  if (status === "error") {
    return "Stream issue";
  }
  return "Waiting";
}
