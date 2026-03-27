'use client';

import type { ReactNode } from 'react';
import { useState } from 'react';
import {
  DashboardOutlined,
  LineChartOutlined,
  ThunderboltOutlined,
  BarChartOutlined,
  RobotOutlined,
  UnorderedListOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  LogoutOutlined,
} from '@ant-design/icons';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import AlertCenter from '@/components/alerts/AlertCenter';
import APLogo from './APLogo';
import SignalStatusBadge from './SignalStatusBadge';
import { authApi } from '@/services/apiClient';

const NAV_GROUPS = [
  {
    label: '主功能',
    items: [
      { path: '/dashboard', name: '驾驶舱', icon: <DashboardOutlined /> },
      { path: '/chart',     name: '图表',   icon: <LineChartOutlined /> },
      { path: '/review',    name: '复盘',   icon: <ThunderboltOutlined /> },
      { path: '/market',    name: '市场',   icon: <BarChartOutlined /> },
    ],
  },
  {
    label: '交易',
    items: [
      { path: '/auto-trading', name: '自动交易', icon: <RobotOutlined /> },
      { path: '/alerts',       name: '持仓记录', icon: <UnorderedListOutlined /> },
    ],
  },
];

function readCollapsed(): boolean {
  if (typeof window === 'undefined') return false;
  return localStorage.getItem('sidebar-collapsed') === 'true';
}

export default function ProAppShell({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const currentPath = pathname ?? '/dashboard';
  const [collapsed, setCollapsed] = useState<boolean>(readCollapsed);

  if (currentPath === '/login') return <>{children}</>;

  const handleCollapse = () => {
    const next = !collapsed;
    setCollapsed(next);
    localStorage.setItem('sidebar-collapsed', String(next));
  };

  const handleLogout = async () => {
    await authApi.logout();
    window.location.assign('/login');
  };

  return (
    <div
      className="command-center-shell cockpit-shell"
      data-testid="command-center-shell"
      data-collapsed={collapsed ? 'true' : 'false'}
      data-shell-style="integrated"
    >
      <aside
        className={`command-center-shell__rail cockpit-shell__sider${collapsed ? ' cockpit-shell__sider--collapsed' : ''}`}
        data-testid="command-center-rail"
      >
        <div className="command-center-shell__logo-row cockpit-shell__logo">
          <div className="command-center-shell__brand-lockup">
            <APLogo />
            {!collapsed && (
              <div className="cockpit-shell__brand">
                <span className="cockpit-shell__brand-title">Alpha Pulse</span>
                <span className="cockpit-shell__brand-sub">合约方向驾驶舱</span>
              </div>
            )}
          </div>
          <div
            className="command-center-shell__actions cockpit-shell__shell-actions"
            data-testid="command-center-shell-actions"
          >
            <button
              type="button"
              className="cockpit-shell__chrome-btn cockpit-shell__collapse-btn"
              onClick={handleCollapse}
              aria-label={collapsed ? '展开侧边栏' : '收起侧边栏'}
              title={collapsed ? '展开侧边栏' : '收起侧边栏'}
            >
              {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            </button>
            <button
              type="button"
              className="cockpit-shell__chrome-btn cockpit-shell__logout-btn"
              aria-label="退出登录"
              title="退出登录"
              onClick={() => {
                void handleLogout();
              }}
            >
              <LogoutOutlined />
            </button>
          </div>
        </div>

        <nav className="command-center-shell__nav cockpit-shell__nav" aria-label="主导航">
          {NAV_GROUPS.map((group) => (
            <div key={group.label} className="cockpit-shell__nav-group">
              {!collapsed && (
                <span className="cockpit-shell__nav-group-label">{group.label}</span>
              )}
              {group.items.map((item) => {
                const active = currentPath === item.path || currentPath.startsWith(item.path + '/');
                return (
                  <Link
                    key={item.path}
                    href={item.path}
                    className={`cockpit-shell__nav-item${active ? ' cockpit-shell__nav-item--active' : ''}`}
                    data-active={active ? 'true' : 'false'}
                    title={collapsed ? item.name : undefined}
                  >
                    <span className="cockpit-shell__nav-icon">{item.icon}</span>
                    {!collapsed && <span className="cockpit-shell__nav-label">{item.name}</span>}
                  </Link>
                );
              })}
            </div>
          ))}
        </nav>

        <div className="cockpit-shell__footer">
          <div className="command-center-shell__dock cockpit-shell__dock" data-testid="command-center-dock">
            <SignalStatusBadge collapsed={collapsed} />
            <AlertCenter />
          </div>
        </div>

      </aside>

      <main className="command-center-shell__content cockpit-shell__content">
        <div
          className="command-center-shell__canvas cockpit-shell__canvas"
          data-testid="command-center-canvas"
          data-shell-surface="continuous"
        >
          {children}
        </div>
      </main>
    </div>
  );
}
