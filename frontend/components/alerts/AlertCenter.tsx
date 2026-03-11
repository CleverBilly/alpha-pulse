"use client";

import { useEffect, useRef, useState } from "react";
import { BellOutlined, NotificationOutlined, SyncOutlined } from "@ant-design/icons";
import { Badge, Button, Drawer, Empty, Spin, Tag } from "antd";
import { alertApi } from "@/services/apiClient";
import type { AlertEvent } from "@/types/alert";

const ALERT_POLL_INTERVAL_MS = 15_000;
const ALERT_LIMIT = 20;
const LAST_SEEN_STORAGE_KEY = "alpha-pulse:last-seen-alert-at";

type BrowserPermissionState = "unsupported" | "default" | "granted" | "denied";

export default function AlertCenter() {
  const [open, setOpen] = useState(false);
  const [alerts, setAlerts] = useState<AlertEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [permission, setPermission] = useState<BrowserPermissionState>("unsupported");
  const [lastSeenAt, setLastSeenAt] = useState(0);
  const initializedRef = useRef(false);
  const notifiedIdsRef = useRef<Set<string>>(new Set());

  useEffect(() => {
    setPermission(readBrowserPermission());
    if (typeof window !== "undefined") {
      const stored = Number(window.localStorage.getItem(LAST_SEEN_STORAGE_KEY) ?? "0");
      if (Number.isFinite(stored) && stored > 0) {
        setLastSeenAt(stored);
      }
    }

    let active = true;

    const loadAlerts = async (mode: "initial" | "poll" | "manual") => {
      try {
        if (mode === "initial") {
          setLoading(true);
        }
        if (mode === "manual") {
          setRefreshing(true);
        }

        const feed = mode === "manual" ? await alertApi.refreshAlerts(ALERT_LIMIT) : await alertApi.getAlerts(ALERT_LIMIT);
        if (!active) {
          return;
        }

        setAlerts(feed.items);
        setError(null);

        if (!initializedRef.current) {
          initializedRef.current = true;
          feed.items.forEach((item) => {
            notifiedIdsRef.current.add(item.id);
          });
          return;
        }

        const newItems = feed.items.filter((item) => !notifiedIdsRef.current.has(item.id));
        newItems.forEach((item) => {
          notifiedIdsRef.current.add(item.id);
          notifyBrowser(item, permission);
        });
      } catch (loadError) {
        if (!active) {
          return;
        }
        setError(formatError(loadError));
      } finally {
        if (!active) {
          return;
        }
        setLoading(false);
        setRefreshing(false);
      }
    };

    void loadAlerts("initial");
    const timer = window.setInterval(() => {
      void loadAlerts("poll");
    }, ALERT_POLL_INTERVAL_MS);

    return () => {
      active = false;
      window.clearInterval(timer);
    };
  }, [permission]);

  useEffect(() => {
    if (!open || alerts.length === 0) {
      return;
    }

    const latestTimestamp = Math.max(...alerts.map((item) => item.created_at || 0), 0);
    if (latestTimestamp <= lastSeenAt) {
      return;
    }

    setLastSeenAt(latestTimestamp);
    if (typeof window !== "undefined") {
      window.localStorage.setItem(LAST_SEEN_STORAGE_KEY, String(latestTimestamp));
    }
  }, [alerts, lastSeenAt, open]);

  const unreadCount = alerts.filter((item) => item.created_at > lastSeenAt).length;

  return (
    <>
      <Badge count={unreadCount} size="small" offset={[-2, 4]}>
        <Button
          type="default"
          icon={<BellOutlined />}
          onClick={() => setOpen(true)}
          className="!rounded-full !border-slate-200 !bg-white/85"
          aria-label="Open alert center"
        >
          Alerts
        </Button>
      </Badge>

      <Drawer
        title="Alert Center"
        placement="right"
        open={open}
        onClose={() => setOpen(false)}
        size="default"
        className="app-shell__drawer"
      >
        <div className="space-y-4">
          <div className="flex flex-wrap items-center gap-2">
            <Tag color="gold">A 级 setup</Tag>
            <Tag color={permissionTagColor(permission)}>{permissionLabel(permission)}</Tag>
            {error ? <Tag color="error">{error}</Tag> : null}
          </div>

          <div className="flex flex-wrap gap-2">
            <Button
              icon={<SyncOutlined spin={refreshing} />}
              onClick={async () => {
                setRefreshing(true);
                try {
                  const feed = await alertApi.refreshAlerts(ALERT_LIMIT);
                  setAlerts(feed.items);
                  setError(null);
                  feed.items.forEach((item) => {
                    if (!notifiedIdsRef.current.has(item.id)) {
                      notifiedIdsRef.current.add(item.id);
                      notifyBrowser(item, permission);
                    }
                  });
                } catch (refreshError) {
                  setError(formatError(refreshError));
                } finally {
                  setRefreshing(false);
                }
              }}
            >
              立即检查
            </Button>
            {permission !== "granted" && permission !== "unsupported" ? (
              <Button
                icon={<NotificationOutlined />}
                onClick={async () => {
                  const nextPermission = await requestBrowserPermission();
                  setPermission(nextPermission);
                }}
              >
                开启浏览器通知
              </Button>
            ) : null}
          </div>

          {loading ? (
            <div className="flex min-h-[160px] items-center justify-center rounded-3xl border border-slate-200 bg-slate-50/80">
              <Spin />
            </div>
          ) : alerts.length === 0 ? (
            <div className="rounded-3xl border border-dashed border-slate-200 bg-slate-50/80 px-6 py-10">
              <Empty description="最近还没有新的方向告警" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            </div>
          ) : (
            <div className="space-y-3">
              {alerts.map((item) => (
                <article
                  key={item.id}
                  className={`rounded-3xl border px-4 py-4 shadow-[0_12px_30px_rgba(32,42,63,0.05)] ${resolveCardTone(item.severity)}`}
                >
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div>
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">
                          {item.symbol}
                        </span>
                        <Tag color={resolveSeverityColor(item.severity)}>{resolveSeverityLabel(item.severity)}</Tag>
                        <Tag color={item.tradeability_label === "No-Trade" ? "warning" : "success"}>
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
                    <Metric label="Entry" value={formatLevel(item.entry_price)} />
                    <Metric label="Target" value={formatLevel(item.target_price)} />
                  </div>

                  {item.reasons.length > 0 ? (
                    <p className="mt-3 text-sm leading-6 text-slate-600">原因：{item.reasons.join("；")}</p>
                  ) : null}

                  {item.deliveries.length > 0 ? (
                    <div className="mt-3 flex flex-wrap gap-2">
                      {item.deliveries.map((delivery) => (
                        <Tag key={`${item.id}-${delivery.channel}`} color={delivery.status === "sent" ? "success" : delivery.status === "failed" ? "error" : "default"}>
                          {delivery.channel}:{delivery.status}
                        </Tag>
                      ))}
                    </div>
                  ) : null}
                </article>
              ))}
            </div>
          )}
        </div>
      </Drawer>
    </>
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

function notifyBrowser(item: AlertEvent, permission: BrowserPermissionState) {
  if (permission !== "granted" || typeof window === "undefined" || typeof Notification !== "function") {
    return;
  }

  const notification = new Notification(item.title, {
    body: `${item.verdict} · ${item.summary}`,
    tag: item.id,
  });
  notification.onclick = () => {
    window.focus();
    window.location.assign("/dashboard");
  };
}

function readBrowserPermission(): BrowserPermissionState {
  if (typeof window === "undefined" || typeof Notification !== "function") {
    return "unsupported";
  }
  if (Notification.permission === "granted") {
    return "granted";
  }
  if (Notification.permission === "denied") {
    return "denied";
  }
  return "default";
}

async function requestBrowserPermission(): Promise<BrowserPermissionState> {
  if (typeof window === "undefined" || typeof Notification !== "function") {
    return "unsupported";
  }
  const result = await Notification.requestPermission();
  if (result === "granted") {
    return "granted";
  }
  if (result === "denied") {
    return "denied";
  }
  return "default";
}

function resolveSeverityColor(severity: string) {
  if (severity === "A") {
    return "gold";
  }
  if (severity === "warning") {
    return "warning";
  }
  return "blue";
}

function resolveSeverityLabel(severity: string) {
  if (severity === "A") {
    return "A 级";
  }
  if (severity === "warning") {
    return "No-Trade";
  }
  return "Direction";
}

function resolveCardTone(severity: string) {
  if (severity === "A") {
    return "border-emerald-200 bg-emerald-50/60";
  }
  if (severity === "warning") {
    return "border-amber-200 bg-amber-50/70";
  }
  return "border-sky-200 bg-sky-50/55";
}

function permissionTagColor(permission: BrowserPermissionState) {
  if (permission === "granted") {
    return "success";
  }
  if (permission === "denied") {
    return "error";
  }
  if (permission === "default") {
    return "processing";
  }
  return "default";
}

function permissionLabel(permission: BrowserPermissionState) {
  if (permission === "granted") {
    return "浏览器通知已开启";
  }
  if (permission === "denied") {
    return "浏览器通知被拒绝";
  }
  if (permission === "default") {
    return "浏览器通知未授权";
  }
  return "浏览器通知不可用";
}

function formatAlertTime(timestamp: number) {
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

function formatLevel(value: number) {
  if (!Number.isFinite(value) || value <= 0) {
    return "--";
  }
  return value >= 1000 ? value.toFixed(2) : value.toFixed(3);
}

function formatError(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return "请求 alert center 失败";
}
