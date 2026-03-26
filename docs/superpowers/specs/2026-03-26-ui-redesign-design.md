# UI Redesign — ProLayout Shell Design Spec

## Goal

将现有自定义 TopBar 布局替换为基于 `@ant-design/pro-components` ProLayout 的侧边栏布局，打造专业交易工作台外观，同时零改动所有业务组件。

---

## 1. 背景与约束

- **当前问题**：水平顶栏导航随功能增加会越来越拥挤；自定义 `app-shell__*` CSS 约 400 行，维护成本高；整体配色（大量圆角、渐变）缺乏专业感。
- **约束**：保持 Next.js 14 App Router 路由不变；不引入 UmiJS；不改动任何业务组件（KlineChart、SignalCard、AlertHistoryBoard 等）；Zustand store 不变。
- **技术选型**：安装 `@ant-design/pro-components@^2.8.10`，使用其 `ProLayout` 组件作为全局 shell，不做全量 Ant Design Pro 迁移。

---

## 2. 视觉设计

| 属性 | 值 |
|------|---|
| 侧边栏背景 | `#0f172a` |
| 侧边栏宽度（展开） | 220px |
| 侧边栏宽度（折叠） | 48px |
| 内容区背景 | `#f0f2f5` |
| 卡片背景 | `#ffffff` |
| 主色 | `#0f766e`（保持现有 colorPrimary 不变） |
| Header 高度 | 52px，`#ffffff` 背景，`1px solid #e2e8f0` 下边框 |
| Header 内容 | 面包屑/页面标题 + 当前信号 badge（BUY/SELL/NEUTRAL）+ 告警铃铛 |
| 圆角 | borderRadius 从 20 → 8，borderRadiusLG 从 24 → 10 |

---

## 3. 架构

### 3.1 新增文件

| 文件 | 职责 |
|------|------|
| `frontend/components/layout/ProAppShell.tsx` | 新 Shell：`ProLayout` + 菜单定义 + Header 右侧插槽 |
| `frontend/components/layout/SignalStatusBadge.tsx` | Header 右侧信号 badge（从 marketStore 读取） |

### 3.2 修改文件

| 文件 | 变更 |
|------|------|
| `frontend/components/layout/AppShell.tsx` | 替换为 `ProAppShell` 的重导出（或直接删除并更新引用） |
| `frontend/components/providers/AntdThemeProvider.tsx` | `borderRadius: 8`，`borderRadiusLG: 10` |
| `frontend/styles/globals.css` | 删除 `app-shell__*`、`terminal-hero__*`、`page-hero__*` 相关样式（约 400 行），保留通用工具类和业务组件样式 |
| `frontend/package.json` | 添加 `@ant-design/pro-components` 依赖 |

### 3.3 不改动文件

所有业务组件、页面文件、Next.js routing、Zustand store、API client、类型定义。

---

## 4. 菜单结构

```ts
const menuItems = [
  { path: '/dashboard', name: '驾驶舱', icon: <DashboardOutlined /> },
  { path: '/chart',     name: '图表',   icon: <LineChartOutlined /> },
  { path: '/review',    name: '复盘',   icon: <ThunderboltOutlined /> },
  { path: '/market',    name: '市场',   icon: <GlobalOutlined /> },
  {
    name: '交易',
    icon: <SwapOutlined />,
    children: [
      { path: '/auto-trading', name: '自动交易', icon: <RobotOutlined /> },
      { path: '/alerts',       name: '持仓记录', icon: <UnorderedListOutlined /> },
    ],
  },
];
```

---

## 5. ProLayout 配置

```tsx
<ProLayout
  layout="side"
  navTheme="realDark"          // 深色侧边栏
  colorPrimary="#0f766e"
  siderWidth={220}
  collapsed={collapsed}
  onCollapse={setCollapsed}
  collapsedWidth={48}
  fixSiderbar
  headerHeight={52}
  title="Alpha Pulse"
  logo={<APLogo />}           // 28px 深绿圆角 logo
  route={{ routes: menuItems }}
  location={{ pathname: usePathname() }}
  onMenuHeaderClick={() => router.push('/dashboard')}
  actionsRender={() => <HeaderActions />}  // 信号badge + 告警铃铛
  menuItemRender={(item, dom) => (
    <Link href={item.path ?? '/'}>{dom}</Link>
  )}
>
  {children}
</ProLayout>
```

---

## 6. Header 右侧插槽（SignalStatusBadge + AlertBell）

- `SignalStatusBadge`：从 `marketStore` 读取 `snapshot.signal.direction` 和 `snapshot.signal.confidence`，渲染绿/红/灰 badge（`BUY · 82%` / `SELL · 65%` / `NEUTRAL`）。
- AlertBell：保留现有告警铃铛逻辑，从 `alertStore` 或通过 `/api/alerts` 读取未读数。

---

## 7. TradingWorkspaceHero / PageHero 处理

当前每个页面顶部有 `TradingWorkspaceHero`（标的选择器、周期按钮、状态芯片）。这些控件**保持位置不变**，作为各页面内容区的第一个组件渲染，不迁移到 Header 或侧边栏。原因：不同页面有不同的标的/周期配置，放在全局 Shell 中会耦合所有页面状态。

---

## 8. CSS 清理策略

删除以下前缀的所有 CSS 规则（保留行内注释标记以便 diff 审查）：

- `.app-shell__*`
- `.terminal-hero__*`（保留业务组件内部用到的子类）
- `.page-hero__*`

保留：
- `.surface-panel`、`.surface-card`（业务组件使用）
- 所有 `@keyframes`、工具类、Tailwind 相关声明

---

## 9. 测试计划

| 测试项 | 方法 |
|--------|------|
| 菜单导航正确跳转 | Playwright e2e：点击每个菜单项，验证 pathname |
| 侧边栏折叠/展开 | Playwright：点击折叠按钮，验证 `.ant-pro-sider` 宽度变化 |
| 信号 badge 渲染 | Vitest：mock marketStore，验证 badge 文字和颜色 |
| 现有业务组件不受影响 | 运行 `npm test`，全部现有测试仍通过 |
| 响应式（屏宽 < 768px） | Playwright：viewport 缩小，验证侧边栏自动折叠 |

---

## 10. 风险与规避

| 风险 | 规避 |
|------|------|
| ProLayout CSS 与 globals.css 冲突 | 迁移后先跑视觉回归截图对比 |
| Next.js App Router 中 `usePathname` SSR 问题 | 用 `'use client'` 指令包裹 ProAppShell |
| `@ant-design/pro-components` 与现有 antd 版本冲突 | 安装后立即运行 `npm ls antd` 检查 peer deps |
| globals.css 删除导致业务组件样式缺失 | 删前逐类 grep 确认无引用再删 |
