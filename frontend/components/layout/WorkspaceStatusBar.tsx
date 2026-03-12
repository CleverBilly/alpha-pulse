"use client";

import { useMemo } from "react";
import { Tag } from "antd";
import { formatSignalAction, formatSweepLabel, formatTrendLabel } from "@/lib/uiLabels";
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
      ? "实时推流已连通"
      : streamStatus === "connecting"
        ? "实时推流连接中"
        : streamStatus === "fallback"
          ? "已切到轮询兜底"
          : loading
            ? "正在同步快照"
            : error
              ? "需要关注"
              : "快照正常";
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
      return "上升趋势";
    }
    if (structure?.trend === "downtrend") {
      return "下降趋势";
    }
    return "震荡区间";
  }, [structure?.trend]);

  return (
    <section className="surface-panel surface-panel--status">
      <div className="status-strip">
        <div className="status-strip__lead">
          <div className="status-bar__eyebrow">工作区状态</div>
          <div className="status-strip__headline">
            <h2 className="status-strip__title">{statusLabel}</h2>
            <Tag color={statusTone}>{loading ? "加载中" : error ? "异常" : "健康"}</Tag>
          </div>
          <p className="status-strip__description">
            {error || streamError
              ? error || streamError
              : `当前监控 ${symbol} ${interval}，${regimeLabel}，信号倾向 ${formatSignalAction(signal?.signal)}。`}
          </p>
        </div>

        <div className="status-strip__metrics">
          <MetricPill label="标的" value={symbol} />
          <MetricPill label="周期" value={interval} />
          <MetricPill label="连接" value={formatFeedLabel(streamStatus, transportMode)} />
          <MetricPill label="信号" value={formatSignalAction(signal?.signal)} />
          <MetricPill label="趋势" value={formatTrendLabel(structure?.trend)} />
          <MetricPill label="扫流动性" value={formatSweepLabel(liquidity?.sweep_type)} />
          <MetricPill
            label="刷新"
            value={
              lastRefreshMode === "force"
                ? "强制重建"
                : lastRefreshMode === "cache"
                  ? "缓存命中"
                  : "等待中"
            }
          />
          <MetricPill label="更新时间" value={updatedLabel} />
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
    return "尚未同步";
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
    return "WebSocket 已连通";
  }
  if (status === "connecting") {
    return "WebSocket 连接中";
  }
  if (transport === "polling") {
    return "HTTP 轮询";
  }
  if (status === "error") {
    return "推流异常";
  }
  return "等待中";
}
