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

### 前置步骤

```bash
npm install @ant-design/pro-components@^2.8.10
npm ls antd  # 确认无 peer dep 冲突，antd 必须是 5.x
```

### 3.1 新增文件

| 文件 | 职责 | 备注 |
|------|------|------|
| `frontend/components/layout/ProAppShell.tsx` | 新 Shell：`ProLayout` + 菜单定义 + Header 右侧插槽 | **必须** `'use client'` |
| `frontend/components/layout/SignalStatusBadge.tsx` | Header 右侧信号 badge（从 marketStore 读取） | `'use client'` |
| `frontend/components/layout/APLogo.tsx` | 侧边栏 Logo（28px 深绿圆角方块 + "AP" 文字） | |
| `frontend/app/auto-trading/page.tsx` | 自动交易占位页面（功能暂未实现，显示"开发中"） | |

### 3.2 修改文件

| 文件 | 变更 |
|------|------|
| `frontend/components/layout/AppShell.tsx` | **直接删除**，同时将所有 import 引用更新为 `ProAppShell`（共 1 处：`frontend/app/layout.tsx`） |
| `frontend/components/providers/AntdThemeProvider.tsx` | `borderRadius: 8`，`borderRadiusLG: 10` |
| `frontend/styles/globals.css` | 仅删除 `.app-shell__*` 和 `.page-hero__*` 规则；`.terminal-hero__*` **本次不动**，后续单独清理 |
| `frontend/package.json` | 添加 `@ant-design/pro-components` 依赖 |

### 3.3 不改动文件

所有业务组件、页面文件、Next.js routing、Zustand store、API client、类型定义。

---

## 4. 菜单结构

菜单只包含**已存在的路由**（`/dashboard`、`/chart`、`/review`、`/market`、`/alerts`）。`/auto-trading` 作为占位页面新建（见 3.1），菜单中显示但页面内容为"开发中"。

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
'use client';  // 必须，因为使用了 usePathname、useState、useRouter

// collapsed 状态用 localStorage 持久化，刷新后保留用户选择
const [collapsed, setCollapsed] = useState<boolean>(() => {
  if (typeof window === 'undefined') return false;
  return localStorage.getItem('sidebar-collapsed') === 'true';
});

const handleCollapse = (val: boolean) => {
  setCollapsed(val);
  localStorage.setItem('sidebar-collapsed', String(val));
};

// ...

<ProLayout
  layout="side"
  navTheme="realDark"          // 深色侧边栏
  colorPrimary="#0f766e"
  siderWidth={220}
  collapsed={collapsed}
  onCollapse={handleCollapse}
  collapsedWidth={48}
  fixSiderbar
  headerHeight={52}
  title="Alpha Pulse"
  logo={<APLogo />}           // frontend/components/layout/APLogo.tsx
  route={{ routes: menuItems }}
  location={{ pathname: usePathname() }}   // usePathname() 在 'use client' 组件中安全调用
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
- AlertBell：通过 `/api/alerts?limit=20` 读取未读告警数（**无 alertStore**，直接 fetch）。未读数 = alerts 中 `read: false` 的条目数。保留现有 AlertBell 组件逻辑，只是将其放入 ProLayout 的 `actionsRender` 插槽。

---

## 7. TradingWorkspaceHero / PageHero 处理

当前每个页面顶部有 `TradingWorkspaceHero`（标的选择器、周期按钮、状态芯片）。这些控件**保持位置不变**，作为各页面内容区的第一个组件渲染，不迁移到 Header 或侧边栏。原因：不同页面有不同的标的/周期配置，放在全局 Shell 中会耦合所有页面状态。

---

## 8. CSS 清理策略

删除以下前缀的所有 CSS 规则：

- `.app-shell__*`（AppShell 删除后这些类不再被引用）
- `.page-hero__*`（PageHero 组件不再使用）

**本次不动**：
- `.terminal-hero__*`（业务组件内部仍有引用，后续单独清理）
- `.surface-panel`、`.surface-card`（业务组件使用）
- 所有 `@keyframes`、工具类、Tailwind 相关声明

**操作方法**：删除前先 `grep -r "app-shell__" frontend/` 和 `grep -r "page-hero__" frontend/` 确认无残留引用，再批量删除。

---

## 9. 测试计划

| 测试项 | 方法 |
|--------|------|
| 菜单导航正确跳转 | Playwright e2e：点击每个菜单项，验证 pathname |
| 侧边栏折叠/展开 | Playwright：点击折叠按钮，验证 `.ant-pro-sider` 宽度变化 |
| 折叠状态 localStorage 持久化 | Playwright：折叠后刷新页面，验证侧边栏仍为折叠态 |
| 信号 badge 渲染 | Vitest：mock marketStore，验证 badge 文字和颜色 |
| 现有业务组件不受影响 | 运行 `npm test`，全部现有 18 个测试仍通过 |
| 响应式（屏宽 < 768px） | Playwright：viewport 缩小，验证侧边栏自动折叠 |
| 视觉回归 | Playwright screenshot：对比迁移前后 `/dashboard` 截图，确认内容区业务组件无位移 |

---

## 10. 风险与规避

| 风险 | 规避 |
|------|------|
| ProLayout CSS 与 globals.css 冲突 | 迁移后先跑视觉回归截图对比 |
| Next.js App Router 中 `usePathname` SSR 问题 | 用 `'use client'` 指令包裹 ProAppShell |
| `@ant-design/pro-components` 与现有 antd 版本冲突 | 安装后立即运行 `npm ls antd` 检查 peer deps |
| globals.css 删除导致业务组件样式缺失 | 删前逐类 grep 确认无引用再删 |
