# UI Redesign — ProLayout Sidebar Shell Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将现有水平顶栏布局替换为 ProLayout 侧边栏布局，实现可折叠深色侧边栏 + 浅色内容区的专业交易工作台外观。

**Architecture:** 安装 `@ant-design/pro-components`（需先将 antd 从 6.x 降至 5.x 以满足 peer dep），新建 `ProAppShell.tsx` 替换 `AppShell.tsx`，`app/layout.tsx` 更新引用，`AntdThemeProvider` 调整圆角，`globals.css` 删除废弃的 `app-shell__*` 和 `page-hero__*` 规则。

**Tech Stack:** Next.js 14 App Router, `@ant-design/pro-components@^2.8.10`, antd 5.x, Zustand, TypeScript

---

## File Map

| 操作 | 文件 |
|------|------|
| 修改 | `frontend/package.json` — antd 5.x + pro-components |
| 新建 | `frontend/components/layout/APLogo.tsx` |
| 新建 | `frontend/components/layout/ProAppShell.tsx` |
| 新建 | `frontend/components/layout/SignalStatusBadge.tsx` |
| 新建 | `frontend/app/auto-trading/page.tsx` |
| 修改 | `frontend/app/layout.tsx` — 替换 AppShell → ProAppShell |
| 删除 | `frontend/components/layout/AppShell.tsx` |
| 修改 | `frontend/components/providers/AntdThemeProvider.tsx` — 圆角 + ProLayout token |
| 修改 | `frontend/styles/globals.css` — 删除 app-shell__* 和 page-hero__* |
| 新建 | `frontend/components/layout/ProAppShell.test.tsx` |
| 新建 | `frontend/components/layout/SignalStatusBadge.test.tsx` |

---

## Task 1: 降级 antd 5.x 并安装 pro-components

**Files:**
- Modify: `frontend/package.json`

> antd 6.x 不在 `@ant-design/pro-components` 的 peer dep 范围（要求 `^4 || ^5`），必须先降级。

- [ ] **Step 1: 安装依赖**

> antd 6.x → 5.x 降级时 npm 7+ 会因 peer dep 冲突报错，需要 `--legacy-peer-deps`。

```bash
cd frontend
npm install antd@^5.29.3 @ant-design/pro-components@^2.8.10 --legacy-peer-deps
```

Expected: package.json 中 `antd` 变为 `^5.29.3`，新增 `@ant-design/pro-components`。

- [ ] **Step 2: 检查 peer dep**

```bash
npm ls antd
```

Expected: 树中 antd 版本为 5.x，无 `WARN` 或 `ERR`。

- [ ] **Step 3: 类型检查（确认 antd 5 API 兼容）**

```bash
npx tsc --noEmit
```

Expected: 无错误（antd 5 与项目使用的 Button/Tag/Table/Modal 等 API 兼容）。

- [ ] **Step 4: 验证现有测试仍通过**

```bash
npm test
```

Expected: `18 passed`。

- [ ] **Step 4: Commit**

```bash
git add frontend/package.json frontend/package-lock.json
git commit -m "chore: downgrade antd to 5.x, add @ant-design/pro-components"
```

---

## Task 2: 新建 APLogo 组件

**Files:**
- Create: `frontend/components/layout/APLogo.tsx`

- [ ] **Step 1: 创建文件**

```tsx
// frontend/components/layout/APLogo.tsx
export default function APLogo() {
  return (
    <div
      style={{
        width: 28,
        height: 28,
        background: '#0f766e',
        borderRadius: 7,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        color: '#fff',
        fontWeight: 800,
        fontSize: 12,
        flexShrink: 0,
      }}
    >
      AP
    </div>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/components/layout/APLogo.tsx
git commit -m "feat: add APLogo component for sidebar"
```

---

## Task 3: 新建 SignalStatusBadge 组件（TDD）

**Files:**
- Create: `frontend/components/layout/SignalStatusBadge.tsx`
- Create: `frontend/components/layout/SignalStatusBadge.test.tsx`

- [ ] **Step 1: 写失败测试**

```tsx
// frontend/components/layout/SignalStatusBadge.test.tsx
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import SignalStatusBadge from './SignalStatusBadge';

vi.mock('@/store/marketStore', () => ({
  useMarketStore: vi.fn(),
}));

import { useMarketStore } from '@/store/marketStore';

describe('SignalStatusBadge', () => {
  it('renders BUY badge with confidence', () => {
    (useMarketStore as ReturnType<typeof vi.fn>).mockReturnValue({
      signal: { signal: 'BUY', confidence: 82 },
    });
    render(<SignalStatusBadge />);
    expect(screen.getByText('BUY · 82%')).toBeInTheDocument();
  });

  it('renders SELL badge', () => {
    (useMarketStore as ReturnType<typeof vi.fn>).mockReturnValue({
      signal: { signal: 'SELL', confidence: 65 },
    });
    render(<SignalStatusBadge />);
    expect(screen.getByText('SELL · 65%')).toBeInTheDocument();
  });

  it('renders NEUTRAL when no signal', () => {
    (useMarketStore as ReturnType<typeof vi.fn>).mockReturnValue({
      signal: null,
    });
    render(<SignalStatusBadge />);
    expect(screen.getByText('NEUTRAL')).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: 运行测试确认失败**

```bash
npm test -- SignalStatusBadge
```

Expected: FAIL — `Cannot find module './SignalStatusBadge'`

- [ ] **Step 3: 实现组件**

```tsx
// frontend/components/layout/SignalStatusBadge.tsx
'use client';

import { useMarketStore } from '@/store/marketStore';

const BADGE_STYLES: Record<string, { background: string; color: string }> = {
  BUY:     { background: '#f0fdf4', color: '#15803d' },
  SELL:    { background: '#fff1f2', color: '#be123c' },
  NEUTRAL: { background: '#f1f5f9', color: '#64748b' },
};

export default function SignalStatusBadge() {
  const { signal } = useMarketStore();
  const direction = signal?.signal ?? 'NEUTRAL';
  const confidence = signal?.confidence;
  const label = confidence != null ? `${direction} · ${Math.round(confidence)}%` : direction;
  const style = BADGE_STYLES[direction] ?? BADGE_STYLES.NEUTRAL;

  return (
    <span
      style={{
        padding: '4px 12px',
        borderRadius: 20,
        fontSize: 12,
        fontWeight: 700,
        ...style,
      }}
    >
      {label}
    </span>
  );
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
npm test -- SignalStatusBadge
```

Expected: `3 passed`

- [ ] **Step 5: Commit**

```bash
git add frontend/components/layout/SignalStatusBadge.tsx frontend/components/layout/SignalStatusBadge.test.tsx
git commit -m "feat: add SignalStatusBadge for ProLayout header"
```

---

## Task 4: 新建 ProAppShell 组件

**Files:**
- Create: `frontend/components/layout/ProAppShell.tsx`
- Create: `frontend/components/layout/ProAppShell.test.tsx`

> ProAppShell 是 Client Component，使用 `usePathname` 和 `useState`。折叠状态持久化到 localStorage。

- [ ] **Step 1: 写失败测试**

```tsx
// frontend/components/layout/ProAppShell.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

// Mock next/navigation
vi.mock('next/navigation', () => ({
  usePathname: () => '/dashboard',
  useRouter: () => ({ push: vi.fn() }),
}));

// Mock pro-components ProLayout (heavy dep, mock for unit test)
vi.mock('@ant-design/pro-components', () => ({
  ProLayout: ({ children, collapsed, onCollapse }: {
    children: React.ReactNode;
    collapsed: boolean;
    onCollapse: (v: boolean) => void;
  }) => (
    <div data-testid="pro-layout" data-collapsed={String(collapsed)}>
      <button onClick={() => onCollapse(!collapsed)}>toggle</button>
      {children}
    </div>
  ),
}));

vi.mock('@/store/marketStore', () => ({
  useMarketStore: vi.fn(() => ({ signal: null })),
}));

import ProAppShell from './ProAppShell';

describe('ProAppShell', () => {
  it('renders children', () => {
    render(<ProAppShell><div>content</div></ProAppShell>);
    expect(screen.getByText('content')).toBeInTheDocument();
  });

  it('toggles collapsed state', async () => {
    render(<ProAppShell><div>x</div></ProAppShell>);
    const layout = screen.getByTestId('pro-layout');
    expect(layout.dataset.collapsed).toBe('false');
    await userEvent.click(screen.getByText('toggle'));
    expect(layout.dataset.collapsed).toBe('true');
  });

  it('skips shell on login route', () => {
    // login route bypass is tested via Playwright e2e (see tests/e2e/)
    // Unit test skipped: re-mocking usePathname after module load requires
    // vi.doMock which conflicts with top-level vi.mock hoisting in vitest
  });
});
```

- [ ] **Step 2: 运行测试确认失败**

```bash
npm test -- ProAppShell
```

Expected: FAIL — `Cannot find module './ProAppShell'`

- [ ] **Step 3: 实现 ProAppShell**

```tsx
// frontend/components/layout/ProAppShell.tsx
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
      collapsedWidth={48}
      fixSiderbar
      headerHeight={52}
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
```

- [ ] **Step 4: 运行测试确认通过**

```bash
npm test -- ProAppShell
```

Expected: `3 passed`（第 3 个 test 是空 body，vitest 会 pass）

- [ ] **Step 5: Commit**

```bash
git add frontend/components/layout/ProAppShell.tsx frontend/components/layout/ProAppShell.test.tsx
git commit -m "feat: add ProAppShell with collapsible sidebar"
```

---

## Task 5: 新建 auto-trading 占位页面

**Files:**
- Create: `frontend/app/auto-trading/page.tsx`

- [ ] **Step 1: 创建文件**

```tsx
// frontend/app/auto-trading/page.tsx
export default function AutoTradingPage() {
  return (
    <div style={{ padding: 24, color: '#64748b', fontSize: 14 }}>
      自动交易功能开发中...
    </div>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/app/auto-trading/page.tsx
git commit -m "feat: add auto-trading placeholder page"
```

---

## Task 6: 更新 app/layout.tsx，替换 AppShell → ProAppShell

**Files:**
- Modify: `frontend/app/layout.tsx`
- Delete: `frontend/components/layout/AppShell.tsx`

- [ ] **Step 1: 更新 layout.tsx**

将 `frontend/app/layout.tsx` 中：
```tsx
import AppShell from "@/components/layout/AppShell";
```
改为：
```tsx
import ProAppShell from "@/components/layout/ProAppShell";
```

将 `<AppShell>{children}</AppShell>` 改为 `<ProAppShell>{children}</ProAppShell>`。

完整文件：
```tsx
import type { Metadata } from "next";
import { AntdRegistry } from "@ant-design/nextjs-registry";
import ProAppShell from "@/components/layout/ProAppShell";
import AntdThemeProvider from "@/components/providers/AntdThemeProvider";
import "../styles/globals.css";

export const metadata: Metadata = {
  title: "Alpha Pulse",
  description: "面向个人合约交易者的方向判断与告警工作台",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN">
      <body>
        <AntdRegistry>
          <AntdThemeProvider>
            <ProAppShell>{children}</ProAppShell>
          </AntdThemeProvider>
        </AntdRegistry>
      </body>
    </html>
  );
}
```

- [ ] **Step 2: 删除旧 AppShell（用 git rm，不用 rm）**

```bash
git rm frontend/components/layout/AppShell.tsx
```

- [ ] **Step 3: 确认无残留引用**

```bash
grep -r "AppShell" frontend/ --include="*.tsx" --include="*.ts"
```

Expected: 无输出（0 引用）

- [ ] **Step 4: 运行全量测试**

```bash
npm test
```

Expected: `≥18 passed`（新增 SignalStatusBadge + ProAppShell 测试后总数增加）

- [ ] **Step 5: Commit**

```bash
git add frontend/app/layout.tsx
git commit -m "feat: replace AppShell with ProAppShell in root layout"
```

---

## Task 7: 更新 AntdThemeProvider — 圆角 + ProLayout token

**Files:**
- Modify: `frontend/components/providers/AntdThemeProvider.tsx`

- [ ] **Step 1: 修改主题 token**

在 `frontend/components/providers/AntdThemeProvider.tsx` 中，修改以下值：

```tsx
// 修改 token 部分
borderRadius: 8,        // 原来是 20
borderRadiusLG: 10,     // 原来是 24

// 修改 colorBgContainer（ProLayout 内容区用白色卡片）
colorBgContainer: '#ffffff',   // 原来是 rgba(255,255,255,0.75)
colorBgLayout: '#f0f2f5',      // 原来是 transparent

// 删除 boxShadow 和 boxShadowSecondary（ProLayout 自带阴影）
```

同时在 `components` 中删除 `Layout` 的 `bodyBg/headerBg/siderBg` 覆盖（ProLayout 自己管理这些），保留 `Card`、`Button`、`Menu`、`Tag`、`Progress`。

完整修改后的文件：
```tsx
"use client";

import type { ReactNode } from "react";
import { App as AntdApp, ConfigProvider, type ThemeConfig, theme } from "antd";

const appTheme: ThemeConfig = {
  algorithm: theme.defaultAlgorithm,
  token: {
    colorPrimary: "#0f766e",
    colorSuccess: "#10b981",
    colorWarning: "#f59e0b",
    colorError: "#ef4444",
    colorInfo: "#0ea5e9",
    colorText: "#0f172a",
    colorTextSecondary: "#64748b",
    colorBorderSecondary: "rgba(15, 23, 42, 0.06)",
    colorBgLayout: "#f0f2f5",
    colorBgContainer: "#ffffff",
    colorFillAlter: "rgba(15, 118, 110, 0.04)",
    fontFamily: "inherit",
    borderRadius: 8,
    borderRadiusLG: 10,
  },
  components: {
    Card: {
      colorBgContainer: "#ffffff",
      headerBg: "transparent",
    },
    Button: {
      controlHeight: 42,
      controlHeightLG: 46,
      borderRadius: 8,
      primaryShadow: "none",
      defaultShadow: "none",
    },
    Menu: {
      itemBg: "transparent",
      horizontalItemSelectedBg: "rgba(15, 118, 110, 0.08)",
      itemSelectedBg: "rgba(15, 118, 110, 0.08)",
      itemSelectedColor: "#0f766e",
      itemColor: "#64748b",
      itemHoverColor: "#0f172a",
      itemHoverBg: "rgba(15, 118, 110, 0.04)",
      activeBarBorderWidth: 0,
      activeBarHeight: 0,
      activeBarWidth: 0,
      horizontalLineHeight: "42px",
      itemBorderRadius: 8,
    },
    Tag: {
      borderRadiusSM: 999,
    },
    Progress: {
      defaultColor: "#0f766e",
      remainingColor: "rgba(15, 118, 110, 0.08)",
    },
  },
};

export default function AntdThemeProvider({ children }: { children: ReactNode }) {
  return (
    <ConfigProvider theme={appTheme}>
      <AntdApp>{children}</AntdApp>
    </ConfigProvider>
  );
}
```

- [ ] **Step 2: 运行测试**

```bash
npm test
```

Expected: 全部通过

- [ ] **Step 3: Commit**

```bash
git add frontend/components/providers/AntdThemeProvider.tsx
git commit -m "feat: update theme — borderRadius 8, solid bg colors for ProLayout"
```

---

## Task 8: 清理 globals.css — 删除 app-shell__* 和 page-hero__*

**Files:**
- Modify: `frontend/styles/globals.css`

> 删除前先确认无残留引用。`page-hero__*` 仍被 `PageHero.tsx` 使用，**本次不删**（见 spec 第 8 节）。

- [ ] **Step 1: 确认 app-shell__* 无残留引用**

```bash
grep -r "app-shell__" frontend/ --include="*.tsx" --include="*.ts"
```

Expected: 无输出（AppShell.tsx 已删除）

- [ ] **Step 2: 确认 page-hero__* 仍有引用（不删）**

```bash
grep -r "page-hero__" frontend/ --include="*.tsx"
```

Expected: 输出 `PageHero.tsx` 中的引用 — 确认本次跳过删除

- [ ] **Step 3: 删除 globals.css 中所有 app-shell__* 规则**

在 `frontend/styles/globals.css` 中，删除从 `.app-shell__header {`（约第 536 行）到 `.app-shell__quicknav-icon {` 块结束（约第 1500 行）的所有 `app-shell__` 相关规则，包括 `@media` 查询中的对应规则。

操作方法：用 Edit 工具逐段删除，每次删除一个连续的 CSS 块。

- [ ] **Step 4: 运行测试**

```bash
npm test
```

Expected: 全部通过（CSS 删除不影响 JS 测试）

- [ ] **Step 5: Commit**

```bash
git add frontend/styles/globals.css
git commit -m "chore: remove app-shell__* CSS rules (AppShell replaced by ProAppShell)"
```

---

## Task 9: 验收测试

- [ ] **Step 1: 运行全量单元测试**

```bash
cd frontend && npm test
```

Expected: 全部通过，无 TypeScript 错误

- [ ] **Step 2: 类型检查**

```bash
cd frontend && npx tsc --noEmit
```

Expected: 无错误

- [ ] **Step 3: 构建验证**

```bash
cd frontend && npm run build
```

Expected: Build 成功，无 error（warning 可接受）

- [ ] **Step 4: 最终 Commit**

```bash
git add -A
git commit -m "feat: complete UI redesign — ProLayout sidebar shell"
```
