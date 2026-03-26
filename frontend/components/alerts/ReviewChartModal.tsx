"use client";

import { useEffect, useState } from "react";
import { Badge, Modal, Spin, Tag, Typography } from "antd";
import { marketApi } from "@/services/apiClient";
import type { AlertEvent, AlertOutcome } from "@/types/alert";
import type { Kline, MarketInterval } from "@/types/market";
import KlineChart from "@/components/chart/KlineChart";
import type { ActiveSignal } from "@/components/chart/SignalOverlayLayer";
import { isLongDirection } from "@/utils/alertUtils";

interface ReviewChartModalProps {
  event: AlertEvent | null;
  open: boolean;
  onClose: () => void;
}

const OUTCOME_COLOR: Record<AlertOutcome, string> = {
  target_hit: "success",
  stop_hit: "error",
  pending: "processing",
  expired: "default",
};

const OUTCOME_LABEL: Record<AlertOutcome, string> = {
  target_hit: "命中目标",
  stop_hit: "触发止损",
  pending: "观察中",
  expired: "已过期",
};

export default function ReviewChartModal({ event, open, onClose }: ReviewChartModalProps) {
  const [klines, setKlines] = useState<Kline[]>([]);
  const [loading, setLoading] = useState(false);
  const [noData, setNoData] = useState(false);

  useEffect(() => {
    if (!open || !event) return;
    let active = true;
    setLoading(true);
    setNoData(false);
    setKlines([]);

    const symbol = event.symbol;
    const interval = event.interval ?? "1h";
    const centerTs = event.created_at;

    Promise.all([
      marketApi.getKline(symbol, interval as MarketInterval, 20, { before_ts: centerTs }),
      marketApi.getKline(symbol, interval as MarketInterval, 40, { after_ts: centerTs }),
    ])
      .then(([before, after]) => {
        if (!active) return;
        const merged = [...before, ...after].sort((a, b) => a.open_time - b.open_time);
        if (merged.length === 0) {
          setNoData(true);
        } else {
          setKlines(merged);
        }
      })
      .catch(() => {
        if (active) setNoData(true);
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [open, event]);

  const activeSignal: ActiveSignal | null =
    event && event.entry_price > 0
      ? {
          entryPrice: event.entry_price,
          stopLoss: event.stop_loss,
          targetPrice: event.target_price,
          direction: isLongDirection(event.verdict, event.tradeability_label) ? "long" : "short",
        }
      : null;

  const outcome = event?.outcome as AlertOutcome | undefined;

  return (
    <Modal
      open={open}
      onCancel={onClose}
      footer={null}
      width={1040}
      title={
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <span>
            复盘 — {event?.symbol} {event?.interval?.toUpperCase()}
          </span>
          {outcome && (
            <Badge
              status={OUTCOME_COLOR[outcome] as "success" | "error" | "processing" | "default" | "warning"}
              text={OUTCOME_LABEL[outcome]}
            />
          )}
          {event?.confidence != null && <Tag>{event.confidence}% 置信度</Tag>}
        </div>
      }
    >
      {loading && (
        <div style={{ textAlign: "center", padding: 40 }}>
          <Spin />
        </div>
      )}
      {!loading && noData && (
        <Typography.Text type="secondary">数据不足，无法还原当时走势。</Typography.Text>
      )}
      {!loading && !noData && klines.length > 0 && (
        <KlineChart
          historicalMode={{ klines, symbol: event?.symbol ?? "", interval: event?.interval ?? "1h" }}
          activeSignal={activeSignal}
        />
      )}
    </Modal>
  );
}
