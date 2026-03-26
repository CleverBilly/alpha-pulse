"use client";

import { useEffect, useRef, useState } from "react";
import { BellOutlined, NotificationOutlined, SettingOutlined, SyncOutlined } from "@ant-design/icons";
import { Badge, Button, Drawer, Empty, Spin, Tag } from "antd";
import AlertConfigPanel from "@/components/alerts/AlertConfigPanel";
import AlertEventCard from "@/components/alerts/AlertEventCard";
import { alertApi } from "@/services/apiClient";
import type { AlertEvent, AlertPreferences } from "@/types/alert";

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
  const [preferences, setPreferences] = useState<AlertPreferences | null>(null);
  const [preferencesLoading, setPreferencesLoading] = useState(true);
  const [preferencesSaving, setPreferencesSaving] = useState(false);
  const [configOpen, setConfigOpen] = useState(false);
  const initializedRef = useRef(false);
  const notifiedIdsRef = useRef<Set<string>>(new Set());
  const browserEnabledRef = useRef(true);
  const audioCtxRef = useRef<AudioContext | null>(null);
  const soundEnabledRef = useRef(false);

  useEffect(() => {
    browserEnabledRef.current = preferences?.browser_enabled ?? true;
  }, [preferences?.browser_enabled]);

  useEffect(() => {
    soundEnabledRef.current = preferences?.sound_enabled ?? false;
  }, [preferences?.sound_enabled]);

  const playSoundAlert = () => {
    if (!soundEnabledRef.current) return;
    try {
      if (!audioCtxRef.current) {
        audioCtxRef.current = new AudioContext();
      }
      const ctx = audioCtxRef.current;
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.type = "sine";
      osc.frequency.setValueAtTime(880, ctx.currentTime);
      gain.gain.setValueAtTime(0.3, ctx.currentTime);
      gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.4);
      osc.start(ctx.currentTime);
      osc.stop(ctx.currentTime + 0.4);
    } catch {
      // AudioContext 不可用时静默降级
    }
  };

  useEffect(() => {
    setPermission(readBrowserPermission());
    if (typeof window !== "undefined") {
      const stored = Number(window.localStorage.getItem(LAST_SEEN_STORAGE_KEY) ?? "0");
      if (Number.isFinite(stored) && stored > 0) {
        setLastSeenAt(stored);
      }
    }

    let active = true;

    const loadPreferences = async () => {
      try {
        const nextPreferences = await alertApi.getAlertPreferences();
        if (!active) {
          return;
        }
        setPreferences(nextPreferences);
      } catch (loadError) {
        if (!active) {
          return;
        }
        setError(formatError(loadError));
      } finally {
        if (active) {
          setPreferencesLoading(false);
        }
      }
    };

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
          notifyBrowser(item, permission, browserEnabledRef.current);
        });
        if (newItems.length > 0) {
          playSoundAlert();
        }
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

    void loadPreferences();
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
          onClick={() => {
            if (!audioCtxRef.current) {
              try {
                audioCtxRef.current = new AudioContext();
                if (audioCtxRef.current.state === "suspended") {
                  void audioCtxRef.current.resume();
                }
              } catch { /* ignore */ }
            }
            setOpen(!open);
          }}
          className="!rounded-full !border-slate-200 !bg-white/85"
          aria-label="打开告警中心"
        >
          告警
        </Button>
      </Badge>

      <Drawer
        title="告警中心"
        placement="right"
        open={open}
        onClose={() => setOpen(false)}
        size="default"
        className="app-shell__drawer"
      >
        <div className="space-y-4">
          <div className="flex flex-wrap items-center gap-2">
            <Tag color="gold">A 级机会</Tag>
            <Tag color={permissionTagColor(permission)}>{permissionLabel(permission)}</Tag>
            {preferences ? (
              <>
                <Tag color={preferences.feishu_enabled ? "processing" : "default"}>
                  飞书{preferences.feishu_enabled ? "开启" : "关闭"}
                </Tag>
                <Tag color={preferences.browser_enabled ? "success" : "default"}>
                  浏览器{preferences.browser_enabled ? "开启" : "关闭"}
                </Tag>
              </>
            ) : null}
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
                      notifyBrowser(item, permission, browserEnabledRef.current);
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
            <Button icon={<SettingOutlined />} onClick={() => setConfigOpen(true)}>
              配置中心
            </Button>
            {preferences?.browser_enabled !== false && permission !== "granted" && permission !== "unsupported" ? (
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
                <AlertEventCard key={item.id} item={item} />
              ))}
            </div>
          )}
        </div>
      </Drawer>

      <AlertConfigPanel
        open={configOpen}
        loading={preferencesLoading}
        saving={preferencesSaving}
        preferences={preferences}
        browserPermission={permission}
        onClose={() => setConfigOpen(false)}
        onSave={async (next) => {
          setPreferencesSaving(true);
          try {
            const saved = await alertApi.updateAlertPreferences(next);
            setPreferences(saved);
            setError(null);
            setConfigOpen(false);
          } catch (saveError) {
            setError(formatError(saveError));
          } finally {
            setPreferencesSaving(false);
          }
        }}
      />
    </>
  );
}

function notifyBrowser(item: AlertEvent, permission: BrowserPermissionState, browserEnabled: boolean) {
  if (!browserEnabled || permission !== "granted" || typeof window === "undefined" || typeof Notification !== "function") {
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

function formatError(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return "请求告警中心失败";
}
