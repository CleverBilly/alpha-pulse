'use client';

import type { ReactNode } from 'react';
import { useState } from 'react';
import { ProLayout, type MenuDataItem } from '@ant-design/pro-components';
import {
  DashboardOutlined,
  LineChartOutlined,
  ThunderboltOutlined,
  BarChartOutlined,
  RobotOutlined,
  UnorderedListOutlined,
} from '@ant-design/icons';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import AlertCenter from '@/components/alerts/AlertCenter';
import APLogo from './APLogo';
import SignalStatusBadge from './SignalStatusBadge';

const MENU_ROUTES = [
  { path: '/dashboard', name: '驾驶舱', icon: <DashboardOutlined /> },
  { path: '/chart',     name: '图表',   icon: <LineChartOutlined /> },
  { path: '/review',    name: '复盘',   icon: <ThunderboltOutlined /> },
  { path: '/market',    name: '市场',   icon: <BarChartOutlined /> },
  {
    name: '交易',
    icon: <UnorderedListOutlined />,
    children: [
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

  const isAuthRoute = pathname === '/login';
  if (isAuthRoute) return <>{children}</>;

  const handleCollapse = (val: boolean) => {
    setCollapsed(val);
    localStorage.setItem('sidebar-collapsed', String(val));
  };

  return (
    <ProLayout
      layout="side"
      navTheme="realDark"
      colorPrimary="#0f766e"
      siderWidth={220}
      collapsed={collapsed}
      onCollapse={handleCollapse}
      fixSiderbar
      title="Alpha Pulse"
      logo={<APLogo />}
      route={{ routes: MENU_ROUTES }}
      location={{ pathname: pathname ?? '/dashboard' }}
      onMenuHeaderClick={() => router.push('/dashboard')}
      actionsRender={() => [
        <SignalStatusBadge key="signal" />,
        <AlertCenter key="alert" />,
      ]}
      menuItemRender={(item: MenuDataItem, dom: ReactNode) => (
        <Link href={(item as MenuDataItem & { path?: string }).path ?? '/'}>{dom}</Link>
      )}
      contentStyle={{ padding: 0, background: '#f0f2f5' }}
    >
      {children}
    </ProLayout>
  );
}
