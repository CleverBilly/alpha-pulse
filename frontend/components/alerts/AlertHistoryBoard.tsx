"use client";

import { useEffect, useMemo, useState } from "react";
import { HistoryOutlined, SyncOutlined } from "@ant-design/icons";
import { Button, Card, Empty, Spin, Tag } from "antd";
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
    <section>
      <Card variant="borderless" className="surface-card surface-card--paper">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.14em] text-cyan-700">复盘</p>
            <h2 className="mt-3 text-[28px] font-semibold tracking-[-0.04em] text-slate-950">告警复盘看板</h2>
            <p className="mt-2 max-w-2xl text-sm leading-7 text-slate-600">
              把 A 级机会、禁止交易和方向切换落成可回看的历史，先看少数真正值得处理的机会，再回看它们为什么出现。
            </p>
          </div>

          <div className="flex flex-wrap gap-2">
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

        <div className="mt-5 flex flex-wrap items-center gap-2">
          <Tag color="processing">{scope === "current" ? `过滤 ${symbol}` : "过滤全部标的"}</Tag>
          <Tag color="gold">最近 {HISTORY_LIMIT} 条记录</Tag>
          {error ? <Tag color="error">{error}</Tag> : null}
        </div>

        <div className="mt-6 grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
          <SummaryCard label="历史告警" value={String(summary.total)} detail="当前过滤范围内" />
          <SummaryCard label="A 级机会" value={String(summary.setupReady)} detail="优先关注的机会" />
          <SummaryCard label="禁止交易" value={String(summary.noTrade)} detail="被系统挡掉的时刻" />
          <SummaryCard label="最近一条" value={summary.latestLabel} detail="可直接回看原因链" />
        </div>

        <div className="mt-6">
          {loading ? (
            <div className="flex min-h-[220px] items-center justify-center rounded-[28px] border border-slate-200 bg-slate-50/75">
              <Spin indicator={<HistoryOutlined spin />} />
            </div>
          ) : filteredAlerts.length === 0 ? (
            <div className="rounded-[28px] border border-dashed border-slate-200 bg-slate-50/80 px-6 py-12">
              <Empty
                description={scope === "current" ? `还没有 ${symbol} 的复盘记录` : "还没有可用的告警历史"}
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            </div>
          ) : (
            <div className="space-y-3">
              {filteredAlerts.map((item) => (
                <AlertEventCard key={item.id} item={item} onReview={setReviewEvent} />
              ))}
            </div>
          )}
        </div>
      </Card>

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
    <div className="rounded-[24px] border border-slate-200 bg-white/88 px-4 py-4 shadow-[0_12px_30px_rgba(32,42,63,0.04)]">
      <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">{label}</p>
      <p className="mt-3 text-lg font-semibold tracking-[-0.03em] text-slate-950">{value}</p>
      <p className="mt-2 text-xs leading-5 text-slate-500">{detail}</p>
    </div>
  );
}

function formatError(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return "请求告警历史失败";
}
