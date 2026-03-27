# Full Frontend Command Center Redesign Design Spec

## Goal

将 Alpha Pulse 前端从“多个业务卡片拼接的后台界面”重构为统一的浅色专业指挥台。整站保留现有业务数据流、路由和主要交互能力，但统一公共壳层、页面骨架、视觉语言和模块类型，让 `dashboard / market / chart / review / alerts / auto-trading` 全部属于同一套系统。

> 本设计文档覆盖并扩展 `2026-03-26-ui-redesign-design.md`。旧文档聚焦壳层与 dashboard 修正；本次方案升级为全站级重构。

---

## 1. Problem Statement

- 当前前端已经从 ProLayout 切回自定义壳层，但整体仍明显带有“卡片堆叠”感。
- 各页面虽然共享部分组件，但没有共享同一套信息层级：有的先看 Hero，有的先看列表，有的直接上卡片，导致整站不成体系。
- 同类模块在不同页面里边距、标题密度、边框、按钮气质都不一致，看起来像若干独立页面，而不是同一个交易工作台。
- `chart`、`market`、`review`、`dashboard` 都在承载“分析与决策”工作流，但主次关系不一致，用户会频繁在重复的 Hero 与零散的卡片之间切换注意力。
- `alerts` 仍主要以抽屉能力存在，`auto-trading` 是占位页，整站的“控制台”概念并不完整。

### Non-goals

- 不修改后端 API、鉴权策略、Zustand store 的核心状态结构。
- 不重做业务计算逻辑，不改变图表渲染引擎。
- 不为了视觉重构引入新的大型 UI 框架或状态管理方案。

---

## 2. Design Principles

### 2.1 Command Center Over Card Grid

- 页面应表现为“指挥台”，而不是标准 SaaS 后台。
- 版面秩序优先通过统一轴线、模块关系和信息节奏建立，而不是继续增加卡片数量。

### 2.2 Light Professional Tone

- 采用浅色专业基调，强调可读性、稳定性和长期使用舒适度。
- 只保留少数强调色位点，不让颜色承担结构职责。

### 2.3 Shared Skeleton, Different Roles

- 所有主要业务页共用同一套版面语言：
  - 顶部总览带
  - 中央主工作区
  - 侧向轨道或底部证据带
- 每页有自己的角色，但必须属于同一骨架系统。

### 2.4 Refine, Don’t Replatform

- 优先通过组件重组、样式原语和页面结构重排完成重构。
- 尽量保留已有业务组件与数据来源，只重塑它们在页面中的表达方式。

---

## 3. Visual Language

### 3.1 Palette

| Token | Value | Usage |
| --- | --- | --- |
| `--cc-bg` | `#f3f4f0` | 页面外层背景 |
| `--cc-bg-accent` | `#edf3ef` | 背景渐变/高光 |
| `--cc-panel` | `rgba(255, 255, 255, 0.84)` | 主模块背景 |
| `--cc-panel-strong` | `rgba(255, 255, 255, 0.96)` | 高优先级模块背景 |
| `--cc-line` | `rgba(18, 32, 28, 0.08)` | 描边 |
| `--cc-ink` | `#17211d` | 主文字 |
| `--cc-muted` | `#67756f` | 次级文字 |
| `--cc-accent` | `#1f7a67` | 激活态/主强调色 |
| `--cc-accent-soft` | `#e8f5ef` | 轻强调背景 |
| `--cc-warning` | `#b7791f` | 警示色 |
| `--cc-danger` | `#b54343` | 风险色 |

### 3.2 Surface Rules

- 外层背景使用低对比渐变和极轻微雾化 radial glow。
- 主内容模块以白色和近白色面板为主，弱化大面积纯灰块。
- 阴影必须短、轻、靠近，不再制造明显悬浮卡片感。
- 通过边框、背景明度和局部高光建立层次，而不是用厚重投影。

### 3.3 Radius and Spacing

- 页面大容器：`28px` 到 `32px`
- 一级模块：`24px`
- 二级模块与按钮：`16px` 到 `20px`
- 状态 chip：`999px`
- 页面主间距统一使用 `24px / 32px`
- 模块内部垂直节奏优先使用 `12px / 16px / 20px`

### 3.4 Typography

- 页面标题更大、更稳，承担“工作台模式切换”的角色。
- 模块标题统一采用高权重、低装饰的交易终端式层级。
- 标签、状态字段、度量项采用一致的小号 uppercase 风格，避免组件各自发挥。
- 数字、价格、比例等关键指标以更紧凑的字距与更稳定的字号关系呈现。

---

## 4. Information Architecture

### 4.1 Global Shell

`ProAppShell` 保留文件名，但语义升级为全站 Command Center Shell：

- 左侧固定“指挥轨道”：品牌、导航、系统状态、告警入口、折叠动作
- 右侧“主画布”：不再只是简单 children 容器，而是承载统一内容网格的画布
- 全站所有业务页在画布内遵守同一套宽度和分区规则

### 4.2 Shared Page Skeleton

每个页面都由以下元素组成：

1. **Overview Band**
   - 页面 eyebrow、主标题、简述、全局状态、关键快捷指标
2. **Primary Workspace**
   - 页面最重要的工作区域，必须抢到最高视觉优先级
3. **Side Rail / Evidence Strip**
   - 次级信息放入右侧轨道或底部证据区，避免在主区竞争焦点

### 4.3 Module Taxonomy

所有业务模块都要归入以下类型之一：

- `command-panel`：高优先级主任务模块
- `overview-strip`：顶层摘要模块
- `rail-panel`：侧向轨道模块
- `evidence-panel`：证据、上下文、回顾模块
- `status-panel`：状态、摘要、指标模块
- `action-panel`：刷新、执行、筛选、配置模块

这套命名会在 CSS 和结构中落地，替代目前“各组件自己像一张卡片”的状态。

---

## 5. Page-by-Page Design

### 5.1 Dashboard

**Role:** 总指挥页

- 顶部改为紧凑但有重量的决策总览带，集中展示方向、置信度、报价、刷新状态、时间框架。
- 中央主工作区保留大图表，但增强为真正的页面主屏。
- 右侧固定操作轨道展示执行方案和仓位计算，不再像普通右列卡片。
- 底部证据带统一承接推理理由、结构确认与辅助信号。

### 5.2 Market

**Role:** 情报台

- 观察列表从普通列表卡改造成更像“雷达面板”的筛选区。
- 顶部总览带承接方向、多周期、触发逻辑和流动性摘要。
- 中央双工作面：左侧聚焦关键位与概览，右侧聚焦 signal tape、微结构和时序变化。
- 订单流与流动性分析移入底部成对的证据/诊断区。

### 5.3 Chart

**Role:** 深度作业台

- 图表必须成为绝对主角，因此顶部 Hero 更克制，避免和图表抢视觉焦点。
- 中央图表区扩大，右侧洞察轨道保持窄而稳定。
- 所有辅助说明、图层状态、定位动作都收束进图表周边，不再做成多块抢眼面板。

### 5.4 Review / Signals

**Role:** 复盘与信号回看台

- `review` 与 `/signals` 继续共用同一工作区组件，但视觉和信息结构更明确。
- 顶部先给复盘摘要，再展示胜率、告警历史、实时信号上下文。
- 历史事件流与 AI 分析形成双工作区关系，表现出“记录 -> 分析”的顺序。

### 5.5 Alerts

**Role:** 告警中枢

- 保留抽屉作为全局快捷入口。
- 新增独立页面承接告警概览、事件流、通知策略和配置工作区。
- 视觉上属于同一套控制台，而不是孤立的配置面板。

### 5.6 Auto Trading

**Role:** 执行中枢预留页

- 即便仍是占位状态，也需要先具备完整的控制台框架。
- 页面先行展示策略状态、风险阈值、执行流程占位区和近期动作记录占位区，为后续功能开发提供视觉母版。

---

## 6. Component Strategy

### 6.1 Keep and Reshape

以下组件保留主要职责，但需要重塑结构与 className 语义：

- `ProAppShell`
- `TradingWorkspaceHero`
- `DecisionHeader`
- `ExecutionPanel`
- `EvidenceRail`
- `KlineChart`
- `MarketOverviewBoard`
- `MarketLevelsBoard`
- `SignalTape`
- `MicrostructureTimeline`
- `OrderFlowPanel`
- `LiquidityPanel`
- `ReviewWorkspace`
- `AlertCenter`
- `AlertHistoryBoard`
- `WinRatePanel`

### 6.2 New Shared Layout Primitives

建议新增一批轻量布局原语，统一整站结构：

- `frontend/components/layout/CommandPage.tsx`
- `frontend/components/layout/OverviewBand.tsx`
- `frontend/components/layout/CommandPanel.tsx`
- `frontend/components/layout/RailPanel.tsx`
- `frontend/components/layout/SectionHeading.tsx`

这些组件只负责布局和语义包裹，不承担业务逻辑。

### 6.3 CSS Organization

- 继续以 `globals.css` 为核心，但引入更明确的命名空间：
  - `.cockpit-shell__*`
  - `.command-page__*`
  - `.overview-band__*`
  - `.command-panel__*`
  - `.rail-panel__*`
- 保留并重构现有 `.dashboard-*`、`.terminal-hero__*` 等样式块，逐步将它们接入新原语。

---

## 7. Data Flow and Behavior

- `MarketSnapshotLoader`、Zustand store、API client 保持不变。
- 页面重构不能破坏现有 symbol、interval、refresh、stream status 等控制路径。
- 所有页面必须继续共享同一份市场快照，不允许因拆布局重新引入数据撕裂。
- 告警抽屉、登录拦截、导航激活态与折叠状态继续保留现有行为。

---

## 8. Error Handling and Responsive Behavior

### 8.1 Error/Loading States

- 空状态、加载态、错误态必须统一进入新的模块语义和视觉体系。
- 不能出现某些模块已经升级为新系统，另一些仍保留旧样式 placeholder 的情况。

### 8.2 Responsive

- Desktop：完整指挥台布局
- Tablet：中央主工作区优先，侧轨折为次级纵列
- Mobile：按工作优先级折叠为“概览 -> 主工作区 -> 次级证据区”，禁止简单的卡片瀑布

---

## 9. Testing Strategy

### Unit / Component Tests

- 为公共壳层、布局原语和页面级结构新增断言。
- 锁定关键信息层级：overview、primary workspace、side rail 是否存在。
- 保护现有关键交互：菜单高亮、折叠持久化、刷新、筛选、告警入口、图表关联控件。

### Integration / E2E

- `dashboard`：驾驶舱关键路径保持可访问
- `market`：观察列表、概览区、信号/时序区均可见
- `chart`：图表与洞察轨道可见
- `review`：历史与分析区同时存在
- 如 `alerts` 页面尚未存在，则新增并验证基本打开路径

---

## 10. Risks and Mitigations

| Risk | Mitigation |
| --- | --- |
| 全站一次性重构容易失去统一节奏 | 先落地共享布局原语和设计 token，再推进页面 |
| 视觉重构导致现有组件测试失效 | 用 TDD 先改测试，再实现布局 |
| 大量样式进入 `globals.css` 后失控 | 使用命名空间和模块类型语义收束 |
| 页面过度追求“好看”而牺牲可读性 | 保持浅色专业型原则，优先信息优先级和交易场景可读性 |
| `auto-trading` 与 `alerts` 页面空洞 | 用结构化占位区确保它们属于同一系统，而非孤立空页 |
