# UI Redesign Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让前端仪表盘和全局壳层对齐用户提供的浅色玻璃态驾驶舱设计，而不是继续沿用错误的 ProLayout 深色侧栏方案。

**Architecture:** 保留现有页面、数据流和大部分业务组件，移除 `ProLayout` 依赖式壳层实现，改为自定义 `ProAppShell` 侧栏 + 内容画布；通过 TDD 先锁定壳层结构、折叠状态和仪表盘关键控件，再完成样式与组件修正。

**Tech Stack:** Next.js 14 App Router, React 18, TypeScript, Vitest, Testing Library, Playwright, Ant Design 5

---

## Chunk 1: Shell Correction

### Task 1: 为新的 Cockpit Shell 补失败测试

**Files:**
- Modify: `frontend/components/layout/ProAppShell.test.tsx`

- [ ] **Step 1: 写失败测试，描述新壳层结构**

覆盖以下行为：
- 当前页菜单项高亮
- 折叠按钮点击后写入 `localStorage`
- 侧栏底部显示 `SignalStatusBadge` 和告警按钮
- `/login` 不渲染全局壳层

- [ ] **Step 2: 运行测试确认失败**

Run: `cd frontend && npm test -- ProAppShell`

Expected: 现有基于 `@ant-design/pro-components` 的 mock 测试不能满足新结构断言，测试失败。

- [ ] **Step 3: 实现最小通过版本**

Modify:
- `frontend/components/layout/ProAppShell.tsx`

实现自定义 sidebar / canvas 结构，保留 `usePathname`、`localStorage` 折叠持久化和登录页 bypass。

- [ ] **Step 4: 再跑测试确认通过**

Run: `cd frontend && npm test -- ProAppShell`

Expected: 所有 `ProAppShell` 测试通过。

### Task 2: 调整壳层配套样式与状态组件

**Files:**
- Modify: `frontend/components/layout/SignalStatusBadge.tsx`
- Modify: `frontend/components/layout/APLogo.tsx`
- Modify: `frontend/styles/globals.css`

- [ ] **Step 1: 写失败测试锁定 badge 呈现不回归**

必要时扩展 `frontend/components/layout/SignalStatusBadge.test.tsx`，断言默认态与置信度文本仍正确。

- [ ] **Step 2: 运行测试确认失败**

Run: `cd frontend && npm test -- SignalStatusBadge`

- [ ] **Step 3: 实现浅色壳层可读的 brand / badge / shell CSS**

新增：
- `.cockpit-shell__*`
- `.cockpit-nav__*`
- `.cockpit-dock__*`

并同步微调 `APLogo` 与 `SignalStatusBadge` 的视觉表达。

- [ ] **Step 4: 跑测试确认通过**

Run: `cd frontend && npm test -- SignalStatusBadge ProAppShell`

Expected: 对应测试通过。

---

## Chunk 2: Dashboard Alignment

### Task 3: 为仪表盘顶部布局增加失败测试

**Files:**
- Modify: `frontend/components/dashboard/DecisionHeader.test.tsx`

- [ ] **Step 1: 写失败测试**

新增断言：
- 置信度卡存在
- 报价卡展示 symbol 与 interval
- 标的选择器和刷新按钮仍可操作
- 多周期按钮区仍保留

- [ ] **Step 2: 运行测试确认失败**

Run: `cd frontend && npm test -- DecisionHeader`

- [ ] **Step 3: 实现最小修正**

Modify:
- `frontend/components/dashboard/DecisionHeader.tsx`
- `frontend/styles/globals.css`

让顶部卡片布局、间距、边框、按钮组更接近设计稿，但不修改数据逻辑。

- [ ] **Step 4: 跑测试确认通过**

Run: `cd frontend && npm test -- DecisionHeader`

### Task 4: 修正执行与仓位面板样式

**Files:**
- Modify: `frontend/components/dashboard/ExecutionPanel.tsx`
- Modify: `frontend/components/trading/PositionCalculator.tsx`
- Modify: `frontend/components/trading/PositionCalculator.test.tsx`
- Modify: `frontend/styles/globals.css`

- [ ] **Step 1: 先写失败测试**

为 `PositionCalculator` 增加结构断言，确保标题、输入区、计算结果仍可见。

- [ ] **Step 2: 运行测试确认失败**

Run: `cd frontend && npm test -- PositionCalculator ExecutionPanel`

- [ ] **Step 3: 实现视觉修正**

将 `PositionCalculator` 从深色调试容器改为与 `ExecutionPanel` 同语义的浅色玻璃面板，必要时补充 className 取代内联深色样式。

- [ ] **Step 4: 跑测试确认通过**

Run: `cd frontend && npm test -- PositionCalculator ExecutionPanel`

---

## Chunk 3: Integration Verification

### Task 5: 更新页面级验证并清理错误设计残留

**Files:**
- Modify: `frontend/tests/e2e/dashboard.spec.ts`
- Modify: `frontend/tests/e2e/dashboard.states.spec.ts`
- Modify: `frontend/app/layout.tsx`（仅当壳层组件命名或结构需要）

- [ ] **Step 1: 调整 e2e 断言到新壳层语义**

新增或更新断言：
- 驾驶舱菜单可见
- 当前判断、K 线图、执行方案、证据链仍存在

- [ ] **Step 2: 运行针对性单测**

Run: `cd frontend && npm test -- ProAppShell SignalStatusBadge DecisionHeader ExecutionPanel PositionCalculator`

Expected: 全部通过。

- [ ] **Step 3: 运行完整单测**

Run: `cd frontend && npm test`

Expected: 全部通过。

- [ ] **Step 4: 运行构建验证**

Run: `cd frontend && npm run build`

Expected: Next.js 生产构建成功。

- [ ] **Step 5: 运行关键 e2e**

Run: `cd frontend && npm run test:e2e -- dashboard.spec.ts`

Expected: dashboard 关键路径通过；若环境限制导致无法跑完，需在交付说明中明确说明。
