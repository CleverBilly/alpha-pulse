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
import { usePathname, useRouter } from 'next/navigation';
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
  const router = useRouter();
  const [collapsed, setCollapsed] = useState<boolean>(readCollapsed);

  if (pathname === '/login') return <>{children}</>;

  const handleCollapse = () => {
    const next = !collapsed;
    setCollapsed(next);
    localStorage.setItem('sidebar-collapsed', String(next));
  };

  const handleLogout = async () => {
    await authApi.logout();
    router.push('/login');
  };

  return (
    <div className="cockpit-shell">
      {/* 侧边栏 */}
      <aside className={`cockpit-shell__sider${collapsed ? ' cockpit-shell__sider--collapsed' : ''}`}>
        {/* Logo 区 */}
        <div className="cockpit-shell__logo">
          <APLogo />
          {!collapsed && (
            <div className="cockpit-shell__brand">
              <span className="cockpit-shell__brand-title">Alpha Pulse</span>
              <span className="cockpit-shell__brand-sub">交易驾驶舱</span>
            </div>
          )}
        </div>

        {/* 导航菜单 */}
        <nav className="cockpit-shell__nav">
          {NAV_GROUPS.map((group) => (
            <div key={group.label} className="cockpit-shell__nav-group">
              {!collapsed && (
                <span className="cockpit-shell__nav-group-label">{group.label}</span>
              )}
              {group.items.map((item) => {
                const active = pathname === item.path || pathname.startsWith(item.path + '/');
                return (
                  <Link
                    key={item.path}
                    href={item.path}
                    className={`cockpit-shell__nav-item${active ? ' cockpit-shell__nav-item--active' : ''}`}
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

        {/* 底部状态区 */}
        <div className="cockpit-shell__footer">
          <SignalStatusBadge collapsed={collapsed} />
          <AlertCenter />
          <button
            className="cockpit-shell__logout-btn"
            onClick={handleLogout}
            title="退出登录"
          >
            <LogoutOutlined />
            {!collapsed && <span>退出登录</span>}
          </button>
        </div>

        {/* 折叠按钮 */}
        <button className="cockpit-shell__collapse-btn" onClick={handleCollapse} aria-label="折叠侧栏">
          {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
        </button>
      </aside>

      {/* 内容区 */}
      <main className="cockpit-shell__content">
        {children}
      </main>
    </div>
  );
}
