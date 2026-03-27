"use client";

import { useEffect, useMemo, useState } from "react";
import { HistoryOutlined, SyncOutlined } from "@ant-design/icons";
import { Button, Empty, Spin, Tag } from "antd";
import AlertEventCard, { formatAlertTime } from "@/components/alerts/AlertEventCard";
import ReviewChartModal from "@/components/alerts/ReviewChartModal";
import { alertApi } from "@/services/apiClient";
import { useMarketStore } from "@/store/marketStore";
import type { AlertEvent } from "@/types/alert";

const HISTORY_LIMIT = 60;
const REFRESH_LIMIT = 20;

type HistoryScope = "current" | "all";

export default function AlertHistoryBoard() {
  const { symbol } = useMarketStore();
  const [scope, setScope] = useState<HistoryScope>("current");
  const [alerts, setAlerts] = useState<AlertEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [reviewEvent, setReviewEvent] = useState<AlertEvent | null>(null);

  useEffect(() => {
    let active = true;

    const load = async () => {
      try {
        setLoading(true);
        const feed = await alertApi.getAlertHistory(HISTORY_LIMIT);
        if (!active) {
          return;
        }
        setAlerts(feed.items);
        setError(null);
      } catch (loadError) {
        if (!active) {
          return;
        }
        setError(formatError(loadError));
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    };

    void load();
    return () => {
      active = false;
    };
  }, []);

  const filteredAlerts = useMemo(() => {
    if (scope === "all") {
      return alerts;
    }
    return alerts.filter((item) => item.symbol === symbol);
  }, [alerts, scope, symbol]);

  const summary = useMemo(() => {
    const items = filteredAlerts;
    const latest = items[0] ?? null;
    return {
      total: items.length,
      setupReady: items.filter((item) => item.kind === "setup_ready").length,
      noTrade: items.filter((item) => item.kind === "no_trade").length,
      latestLabel: latest ? `${latest.symbol} · ${formatAlertTime(latest.created_at)}` : "--",
    };
  }, [filteredAlerts]);

  return (
    <section className="alert-history" data-testid="alert-history-rail">
      <div className="alert-history__shell">
        <div className="alert-history__header">
          <div>
            <p className="alert-history__eyebrow">复盘轨道</p>
            <h2 className="alert-history__title">告警复盘看板</h2>
            <p className="alert-history__description">
              把 A 级机会、禁止交易和方向切换落成可回看的历史，先看少数真正值得处理的机会，再回看它们为什么出现。
            </p>
          </div>

          <div className="alert-history__actions">
            <Button
              type={scope === "current" ? "primary" : "default"}
              onClick={() => setScope("current")}
            >
              当前标的
            </Button>
            <Button
              type={scope === "all" ? "primary" : "default"}
              onClick={() => setScope("all")}
            >
              全部历史
            </Button>
            <Button
              icon={<SyncOutlined spin={refreshing} />}
              onClick={async () => {
                setRefreshing(true);
                try {
                  await alertApi.refreshAlerts(REFRESH_LIMIT);
                  const feed = await alertApi.getAlertHistory(HISTORY_LIMIT);
                  setAlerts(feed.items);
                  setError(null);
                } catch (refreshError) {
                  setError(formatError(refreshError));
                } finally {
                  setRefreshing(false);
                }
              }}
            >
              重新评估告警
            </Button>
          </div>
        </div>

        <div className="alert-history__filters">
          <Tag color="processing">{scope === "current" ? `过滤 ${symbol}` : "过滤全部标的"}</Tag>
          <Tag color="gold">最近 {HISTORY_LIMIT} 条记录</Tag>
          {error ? <Tag color="error">{error}</Tag> : null}
        </div>

        <div className="alert-history__summary" data-testid="alert-history-summary">
          <SummaryCard label="历史告警" value={String(summary.total)} detail="当前过滤范围内" />
          <SummaryCard label="A 级机会" value={String(summary.setupReady)} detail="优先关注的机会" />
          <SummaryCard label="禁止交易" value={String(summary.noTrade)} detail="被系统挡掉的时刻" />
          <SummaryCard label="最近一条" value={summary.latestLabel} detail="可直接回看原因链" />
        </div>

        <div className="alert-history__feed">
          {loading ? (
            <div className="alert-history__empty">
              <Spin indicator={<HistoryOutlined spin />} />
            </div>
          ) : filteredAlerts.length === 0 ? (
            <div className="alert-history__empty">
              <Empty
                description={scope === "current" ? `还没有 ${symbol} 的复盘记录` : "还没有可用的告警历史"}
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            </div>
          ) : (
            <div className="alert-history__flow">
              {filteredAlerts.map((item) => (
                <AlertEventCard key={item.id} item={item} onReview={setReviewEvent} variant="row" />
              ))}
            </div>
          )}
        </div>
      </div>

      <ReviewChartModal
        event={reviewEvent}
        open={!!reviewEvent}
        onClose={() => setReviewEvent(null)}
      />
    </section>
  );
}

function SummaryCard({ label, value, detail }: { label: string; value: string; detail: string }) {
  return (
    <div className="alert-history__summary-item">
      <p className="alert-history__summary-label">{label}</p>
      <p className="alert-history__summary-value">{value}</p>
      <p className="alert-history__summary-detail">{detail}</p>
    </div>
  );
}

function formatError(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return "请求告警历史失败";
}
