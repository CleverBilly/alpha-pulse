"use client";

import type { ReactNode } from "react";
import { useMemo, useState } from "react";
import {
  BarChartOutlined,
  DashboardOutlined,
  SyncOutlined,
  LineChartOutlined,
  MenuOutlined,
  ThunderboltOutlined,
} from "@ant-design/icons";
import { Button, Drawer, Layout, Menu, Tag, Typography, Grid, type MenuProps } from "antd";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import AlertCenter from "@/components/alerts/AlertCenter";
import { formatSignalAction } from "@/lib/uiLabels";
import { authApi } from "@/services/apiClient";
import { useMarketStore } from "@/store/marketStore";

const NAV_ITEMS = [
  {
    key: "/dashboard",
    icon: <DashboardOutlined />,
    label: "驾驶舱",
  },
  {
    key: "/chart",
    icon: <LineChartOutlined />,
    label: "图表",
  },
  {
    key: "/review",
    icon: <ThunderboltOutlined />,
    label: "复盘",
  },
  {
    key: "/market",
    icon: <BarChartOutlined />,
    label: "市场",
  },
] as const;

export default function AppShell({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const screens = Grid.useBreakpoint();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);
  const selectedKey = resolveSelectedKey(pathname);
  const authEnabled = process.env.NEXT_PUBLIC_AUTH_ENABLED === "true";
  const isAuthRoute = pathname === "/login";
  const { symbol, interval, signal, loading, lastRefreshMode, lastUpdatedAt, transportMode, streamStatus } =
    useMarketStore();
  const selectedItem = useMemo(
    () => NAV_ITEMS?.find((item) => String(item?.key ?? "") === selectedKey) ?? NAV_ITEMS?.[0],
    [selectedKey],
  );
  const menuItems: MenuProps["items"] = NAV_ITEMS.map((item) => ({
    key: item.key,
    icon: item.icon,
    label: item.label,
  }));

  if (isAuthRoute) {
    return <>{children}</>;
  }

  const menu = (
    <Menu
      mode={screens.md ? "horizontal" : "inline"}
      selectedKeys={[selectedKey]}
      items={menuItems}
      className="app-shell__menu"
      onClick={({ key }) => {
        setDrawerOpen(false);
        router.push(String(key));
      }}
    />
  );

  return (
    <Layout className="app-shell">
      <Layout.Header className="app-shell__header">
        <div className="app-shell__header-inner">
          <Link href="/dashboard" className="app-shell__brand">
            <span className="app-shell__brand-mark">AP</span>
            <span>
              <strong>Alpha Pulse</strong>
              <small>合约方向副驾驶</small>
            </span>
          </Link>

          {screens.md ? menu : null}

          {screens.md ? (
            <div className="app-shell__status">
              <span className="app-shell__status-chip app-shell__status-chip--muted">
                {String(selectedItem?.label ?? "驾驶舱")}
              </span>
              <span className="app-shell__status-chip">
                {symbol} · {interval}
              </span>
              <span className={`app-shell__status-chip ${resolveSignalClassName(signal?.signal)}`}>
                {formatSignalAction(signal?.signal)}
              </span>
              {screens.xl ? (
                <span className="app-shell__status-chip app-shell__status-chip--dim">
                  {loading ? <SyncOutlined spin /> : null}
                  {formatSnapshotMeta(lastRefreshMode, lastUpdatedAt, transportMode, streamStatus)}
                </span>
              ) : null}
              <AlertCenter />
              {authEnabled ? (
                <Button
                  type="default"
                  onClick={async () => {
                    setLoggingOut(true);
                    try {
                      await authApi.logout();
                    } finally {
                      setLoggingOut(false);
                      router.push("/login");
                      router.refresh();
                    }
                  }}
                  className="!rounded-full !border-slate-200 !bg-white/85"
                >
                  {loggingOut ? "退出中..." : "退出登录"}
                </Button>
              ) : null}
            </div>
          ) : (
            <div className="flex items-center gap-2">
              <AlertCenter />
              <Button
                type="default"
                icon={<MenuOutlined />}
                aria-label="打开导航"
                onClick={() => setDrawerOpen(true)}
                className="!border-white/40 !bg-white/75 !backdrop-blur"
              />
            </div>
          )}
        </div>
      </Layout.Header>

      <Layout.Content className="app-shell__content">
        <div className="app-shell__backdrop" />
        <div className="app-shell__container">{children}</div>
      </Layout.Content>

      <nav className="app-shell__quicknav" aria-label="快捷导航">
        {NAV_ITEMS.map((item) => {
          const key = String(item?.key ?? "");
          const active = key === selectedKey;
          return (
            <button
              key={key}
              type="button"
              aria-current={active ? "page" : undefined}
              className={`app-shell__quicknav-item ${active ? "app-shell__quicknav-item--active" : ""}`}
              onClick={() => router.push(key)}
            >
              <span className="app-shell__quicknav-icon">{item?.icon}</span>
              <span>{item?.label}</span>
            </button>
          );
        })}
      </nav>

      <Drawer
        placement="right"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        closable={false}
        size="default"
        className="app-shell__drawer"
      >
        <div className="app-shell__drawer-head">
          <Typography.Text className="!font-semibold !text-slate-900">页面导航</Typography.Text>
          <Tag color="cyan">V2.0</Tag>
        </div>
        {menu}
      </Drawer>
    </Layout>
  );
}

function resolveSelectedKey(pathname: string | null) {
  if (!pathname || pathname === "/") {
    return "/dashboard";
  }

  if (pathname === "/signals" || pathname.startsWith("/signals/")) {
    return "/review";
  }

  const match = NAV_ITEMS?.find((item) => {
    const key = String(item?.key ?? "");
    return pathname === key || pathname.startsWith(`${key}/`);
  });

  return String(match?.key ?? "/dashboard");
}

function formatSnapshotMeta(
  mode: "cache" | "force" | null,
  updatedAt: number | null,
  transport: "idle" | "websocket" | "polling",
  status: "idle" | "connecting" | "live" | "fallback" | "error",
) {
  const modeLabel = mode === "force" ? "强制" : mode === "cache" ? "缓存" : "空闲";
  const transportLabel =
    status === "live" && transport === "websocket"
      ? "实时 WS"
      : transport === "polling"
        ? "HTTP 轮询"
        : status === "connecting"
          ? "连接中"
          : "空闲";
  if (!updatedAt || !Number.isFinite(updatedAt)) {
    return `${transportLabel} · 等待中`;
  }

  const timeLabel = new Date(updatedAt).toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
  });
  return `${transportLabel} ${modeLabel} · ${timeLabel}`;
}

function resolveSignalClassName(signal?: string) {
  if (signal === "BUY") {
    return "app-shell__status-chip--positive";
  }
  if (signal === "SELL") {
    return "app-shell__status-chip--negative";
  }
  return "app-shell__status-chip--neutral";
}
