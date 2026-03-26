# UI Redesign — Light Cockpit Shell Design Spec

## Goal

将当前偏向 Ant Design Pro 的深色后台壳层，纠正为参考图中的浅色玻璃态交易驾驶舱：左侧为半透明浅色导航，右侧为大圆角内容画布，仪表盘顶部保持“判断 + 报价 + 控件”一体化布局，整体观感更接近专业交易终端而不是通用后台。

---

## 1. 问题定义

- **现状问题**：当前实现严格照着错误的 ProLayout 方案落地，导致出现深色侧边栏、后台式 header、过于规整的管理台结构，和目标图完全不是同一套设计语言。
- **真实目标**：以用户提供的参考图为唯一视觉基准，保留现有业务数据流和页面结构，但把全局壳层、仪表盘头部、辅助面板样式统一到浅色玻璃态工作台。
- **实现边界**：不改动 API、Zustand store、业务计算逻辑；优先重用现有 `DecisionHeader`、`KlineChart`、`ExecutionPanel`、`EvidenceRail` 等组件，只修正结构和视觉表达。

---

## 2. 视觉基准

### 2.1 页面骨架

- 整体背景：浅灰蓝雾化渐变，允许轻微 radial glow，不使用深色后台。
- 左侧导航：固定宽度浅色半透明侧栏，带柔和描边、模糊和阴影。
- 主内容：右侧为独立的大圆角“内容画布”，带薄边框、轻阴影和很浅的蓝绿渐变。
- 组件表面：统一使用白色到近白色的玻璃感 surface，减少纯实色块。

### 2.2 关键视觉参数

| 属性 | 值 |
|------|---|
| 页面背景 | `linear-gradient(180deg, #f7fafc 0%, #eef3f8 100%)` + 轻微 radial glow |
| 侧栏背景 | `rgba(255,255,255,0.72)` |
| 内容画布背景 | `linear-gradient(180deg, rgba(255,255,255,0.9), rgba(244,248,252,0.9))` |
| 主色 | `#1b7f79` |
| 深色文案 | `#172033` |
| 次级文案 | `#64748b` |
| 通用大圆角 | `28px` 到 `36px` |
| 芯片圆角 | `999px` |
| 主内容区边框 | `1px solid rgba(255,255,255,0.72)` |

### 2.3 交互观感

- 活跃菜单项为深青色实底，不使用后台风格蓝色高亮条。
- 折叠按钮为悬浮小圆形按钮，贴在侧栏右边缘。
- 底部状态区固定在侧栏底部，展示信号 badge 与告警按钮。
- 页面不再展示后台式顶栏 header，当前页面的信息留在内容区自身表达。

---

## 3. 信息架构

### 3.1 左侧导航

导航保持现有页面入口，但按参考图分组和语气呈现：

- 主功能
  - `/dashboard` 驾驶舱
  - `/chart` 图表
  - `/review` 复盘
  - `/market` 市场
- 交易
  - `/auto-trading` 自动交易
  - `/alerts` 持仓记录

### 3.2 侧栏底部

- `SignalStatusBadge` 作为绿色/红色/灰色圆角 pill，展示 `BUY · 82%` 这类状态。
- `AlertCenter` 以白色圆角按钮呈现，保留现有告警抽屉逻辑，不再放入顶栏右上角。

### 3.3 主内容区

- 所有页面内容包裹在统一的 `cockpit-shell__canvas` 中。
- `dashboard` 顶部保留当前 `DecisionHeader` 作为首屏大卡。
- 图表、执行方案、仓位信息、证据链继续留在内容区，不迁移到全局壳层。

---

## 4. 组件架构

### 4.1 `frontend/components/layout/ProAppShell.tsx`

保留文件名以减少调用面变化，但组件职责改为“自定义 Cockpit Shell”，不再依赖 `@ant-design/pro-components`：

- 自行渲染 sidebar、collapse trigger、内容画布和 footer dock
- 使用 `usePathname()` 判定激活菜单
- 使用 `useState + localStorage` 记住折叠状态
- 登录页继续 bypass shell

### 4.2 `frontend/components/layout/APLogo.tsx`

- 保留现有 Logo 基础结构
- 微调为更贴近参考图的深青底、白字、配合品牌标题与副标题

### 4.3 `frontend/components/layout/SignalStatusBadge.tsx`

- 调整为浅色界面可读的 pill 颜色
- 作为侧栏底部状态组件复用

### 4.4 `frontend/components/dashboard/DecisionHeader.tsx`

保持现有业务文案与数据来源，但细节调整为参考图结构：

- 左侧：eyebrow、决策标签、大标题、摘要、meta 卡片、原因 chips、时间框架 chips
- 右侧：小号置信度卡、主报价卡、标的与刷新控件卡、周期按钮排
- 让顶部大卡的留白、边框、圆角与目标图一致

### 4.5 `frontend/components/trading/PositionCalculator.tsx`

- 从深色调试面板改成白色玻璃面板
- 保留计算逻辑不变
- 视觉上与 `ExecutionPanel`、`KlineChart` 同一层级

---

## 5. CSS 策略

- 新增并使用 `.cockpit-shell__*` 命名空间管理全局壳层。
- 保留现有 `.surface-panel`、`.dashboard-*`、`.terminal-hero__*` 体系，在此基础上微调，而不是再次引入一套第三方布局系统。
- 删除或弃用仅为 ProLayout 服务的样式思路；不再新增 `ant-pro-*` 定制规则。

---

## 6. 测试策略

| 测试项 | 方法 |
|------|---|
| 登录页不包 shell | Vitest：`/login` 路由下只渲染 children |
| 侧栏折叠持久化 | Vitest：toggle 后 `localStorage` 写入 `true/false` |
| 活跃菜单渲染 | Vitest：当前 pathname 对应菜单项带 active class |
| 底部状态区存在 | Vitest：`SignalStatusBadge` 与告警按钮可见 |
| Dashboard 顶部关键控件保留 | Vitest：标题、刷新、标的选择、周期切换仍可工作 |
| Dashboard 主路径可见 | Playwright：`/dashboard` 可见“当前判断 / K 线图 / 执行方案 / 证据链” |

---

## 7. 风险与规避

| 风险 | 规避 |
|------|------|
| 直接替换 ProLayout 导致现有测试失效 | 先写针对新壳层的失败测试，再逐步实现 |
| 侧栏重构影响登录页 | 保留登录页 bypass，并加单测覆盖 |
| 仪表盘样式改动过大影响业务内容可读性 | 优先保留 DOM 结构和文案，只调整布局与 surface 样式 |
| 仓位计算器视觉切换影响现有行为 | 仅改容器与展示样式，不碰计算逻辑与副作用流程 |
