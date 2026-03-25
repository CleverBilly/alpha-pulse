# Trading Decision Loop Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为单用户实盘辅助构建交易决策闭环：声音告警 + 图表三线叠加 + 仓位计算器（快线 Week 1）；信号结果追踪 + Review 复盘 + 胜率统计面板（慢线 Week 2-3）。

**Architecture:** 快线以纯前端或极小后端改动为原则，核心是拆分 KlineChart.tsx（2145行）并新增 SignalOverlayLayer。慢线在 alert_records 加结果字段，新增 OutcomeTrackerService 在调度器中定期结算信号，并在前端 Review 页展示 K 线上下文和胜率统计。

**Tech Stack:** Go 1.21 + Gin + GORM + MySQL（后端），Next.js 14 App Router + TypeScript + Zustand + Ant Design + SVG（前端），Vitest（前端单测），go test（后端单测）

**Spec:** `docs/superpowers/specs/2026-03-25-trading-decision-loop-design.md`

**Dependency order:** Task 1→2→3（快线顺序）；Task 4（KlineChart拆分）必须在 Task 5 之前；Task 7→8→9（慢线S1顺序）；Task 10→11（S2顺序）；Task 12→13（S3顺序）。慢线各组之间无依赖可并行。

---

## Task 1: F2 — 调度间隔降至 15s

**Files:**
- Modify: `backend/.env.example`
- Modify: `CLAUDE.md`

- [ ] **Step 1: 更新 .env.example**

将 `SCHEDULER_INTERVAL_SECONDS` 默认值改为 15：

```
SCHEDULER_INTERVAL_SECONDS=15
```

在 `backend/.env.example` 中找到该行并修改（若不存在则追加）。

- [ ] **Step 2: 更新 CLAUDE.md 配置表**

在 CLAUDE.md 的"关键配置"表中将 `SCHEDULER_INTERVAL_SECONDS` 的默认值说明改为 `15`（原为 `60`）。

- [ ] **Step 3: 验证 jobs.go 无需改动**

阅读 `backend/internal/scheduler/jobs.go`，确认 `NewJobs` 接收 `interval time.Duration` 且在 `interval <= 0` 时才用默认值 1 分钟。配置驱动已完整，无需代码改动。

- [ ] **Step 4: Commit**

```bash
git add backend/.env.example CLAUDE.md
git commit -m "feat: lower default scheduler interval to 15s"
```

---

## Task 2: F1 — 声音告警后端（model + service + handler）

**Files:**
- Modify: `backend/models/alert_preference.go`
- Modify: `backend/internal/service/alert_preferences.go`
- Modify: `backend/internal/handler/alert_handler.go`

- [ ] **Step 1: 给 AlertPreference 模型加 SoundEnabled 字段**

在 `backend/models/alert_preference.go` 的 `AlertPreference` struct 中，在 `WatchedSymbols` 字段之前插入：

```go
SoundEnabled bool `gorm:"column:sound_enabled;not null;default:false;comment:是否启用声音告警" json:"sound_enabled"`
```

- [ ] **Step 2: 给 AlertPreferences service 类型加 SoundEnabled**

在 `backend/internal/service/alert_preferences.go` 的 `AlertPreferences` struct 中新增：

```go
SoundEnabled bool `json:"sound_enabled"`
```

更新 `defaultAlertPreferences`：
```go
SoundEnabled: false,
```

更新 `sanitizeAlertPreferences` 返回值：
```go
SoundEnabled: input.SoundEnabled,
```

更新 `projectAlertPreferences`：
```go
SoundEnabled: prefs.SoundEnabled,
```

更新 `hydrateAlertPreferences` 传入 `sanitizeAlertPreferences` 的参数：
```go
SoundEnabled: record.SoundEnabled,
```

- [ ] **Step 3: 更新 alert_handler.go request 结构体和映射**

在 `alertPreferencesRequest` struct 中新增：
```go
SoundEnabled bool `json:"sound_enabled"`
```

在 `UpdateAlertPreferences` 方法的 `service.AlertPreferences{...}` 映射中新增：
```go
SoundEnabled: request.SoundEnabled,
```

- [ ] **Step 4: 编译验证**

```bash
cd backend && go build ./...
```

期望：零编译错误。

- [ ] **Step 5: Commit**

```bash
git add backend/models/alert_preference.go \
        backend/internal/service/alert_preferences.go \
        backend/internal/handler/alert_handler.go
git commit -m "feat: add sound_enabled to alert preferences"
```

---

## Task 3: F1 — 声音告警前端（AlertCenter + AlertConfigPanel）

**Files:**
- Modify: `frontend/types/alert.ts`
- Modify: `frontend/components/alerts/AlertCenter.tsx`
- Modify: `frontend/components/alerts/AlertConfigPanel.tsx`

- [ ] **Step 1: 更新 AlertPreferences 类型**

在 `frontend/types/alert.ts` 的 `AlertPreferences` interface 中新增：
```typescript
sound_enabled: boolean;
```

- [ ] **Step 2: 在 AlertCenter.tsx 中实现声音告警**

读取 `frontend/components/alerts/AlertCenter.tsx`，在 `notifyBrowser` 调用处附近添加声音告警逻辑。

在组件顶部（`export default function AlertCenter()`函数体内，useState 之后）添加 AudioContext ref：

```typescript
const audioCtxRef = useRef<AudioContext | null>(null);
const soundEnabledRef = useRef(false);
```

在 `useEffect(() => { browserEnabledRef.current = preferences?.browser_enabled ?? true; }, [...])` 之后添加：

```typescript
useEffect(() => {
  soundEnabledRef.current = preferences?.sound_enabled ?? false;
}, [preferences?.sound_enabled]);
```

添加 `playSoundAlert` 函数（在组件函数体内，return 之前）：

```typescript
const playSoundAlert = () => {
  if (!soundEnabledRef.current) return;
  try {
    if (!audioCtxRef.current) {
      audioCtxRef.current = new AudioContext();
    }
    const ctx = audioCtxRef.current;
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.type = "sine";
    osc.frequency.setValueAtTime(880, ctx.currentTime);
    gain.gain.setValueAtTime(0.3, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.4);
    osc.start(ctx.currentTime);
    osc.stop(ctx.currentTime + 0.4);
  } catch {
    // AudioContext 不可用时静默降级
  }
};
```

在铃铛按钮的 `onClick` 中初始化 AudioContext（找到 `setOpen(!open)` 的按钮），在 `setOpen(!open)` 之前添加：
```typescript
if (!audioCtxRef.current) {
  try { audioCtxRef.current = new AudioContext(); } catch { /* ignore */ }
}
```

在检测到新 alert 的循环中（`newItems.forEach` 块内，`notifyBrowser` 调用之后）添加：
```typescript
playSoundAlert();
```

> 注意：`playSoundAlert` 只调用一次（不对每条 alert 各调一次），因为 `newItems` 可能有多条，一次提示音足够。将 `playSoundAlert()` 移到 `newItems.forEach` 循环外、`if (newItems.length > 0)` 判断内。

- [ ] **Step 3: 在 AlertConfigPanel.tsx 中添加声音开关 UI**

读取 `frontend/components/alerts/AlertConfigPanel.tsx`，找到"浏览器"推送的 `Switch` 控件所在区域，在其正下方（同一 `Form.Item` 容器风格）添加声音开关：

```tsx
<div style={{ display: "flex", alignItems: "center", gap: 8, marginTop: 8 }}>
  <Switch
    checked={current?.sound_enabled ?? false}
    onChange={(checked) => setDraft((d) => d ? { ...d, sound_enabled: checked } : d)}
    size="small"
  />
  <span style={{ fontSize: 13 }}>声音提示</span>
  <span style={{ color: "#888", fontSize: 12 }}>新告警到达时播放提示音</span>
</div>
```

- [ ] **Step 4: 运行前端测试**

```bash
cd frontend && npm test -- --run
```

期望：全部通过（无新增失败）。

- [ ] **Step 5: Commit**

```bash
git add frontend/types/alert.ts \
        frontend/components/alerts/AlertCenter.tsx \
        frontend/components/alerts/AlertConfigPanel.tsx
git commit -m "feat: add sound alert to alert center"
```

---

## Task 4: F3 前置 — KlineChart.tsx 拆分

**Files:**
- Modify: `frontend/components/chart/KlineChart.tsx`（主容器，目标 ≤420 行）
- Create: `frontend/components/chart/KlineCandleLayer.tsx`
- Create: `frontend/components/chart/StructureLiquidityLayer.tsx`
- Create: `frontend/components/chart/SignalOverlayLayer.tsx`（不含三线，Task 5 再加）

> **重要：** 本 Task 只拆分，不增加任何新功能。拆分后行为与原版完全相同。拆分完成后立即运行 `npm test` 验证无回归。

- [ ] **Step 1: 阅读完整 KlineChart.tsx**

完整读取 `frontend/components/chart/KlineChart.tsx`（2145 行），识别以下结构：
- 顶部：类型定义、常量、`buildChartModel` 函数
- 中部：`KlineChart` 主组件（状态、事件处理）
- SVG 渲染部分：蜡烛图层、指标线图层、结构点图层、流动性图层、信号标记图层、微结构事件图层
- 底部：图例控件

- [ ] **Step 2: 定义共享 ChartCoords 接口**

在 `KlineChart.tsx` 顶部（或新建 `frontend/components/chart/chartTypes.ts`）定义：

```typescript
export interface ChartCoords {
  toX: (index: number) => number;
  toY: (price: number) => number;
  candleWidth: number;
  chartWidth: number;
  chartHeight: number;
  paddingTop: number;
  paddingRight: number;
  paddingBottom: number;
  paddingLeft: number;
  priceMin: number;
  priceMax: number;
}
```

- [ ] **Step 3: 抽取 KlineCandleLayer**

创建 `frontend/components/chart/KlineCandleLayer.tsx`，将蜡烛实体、影线、成交量柱的 SVG 渲染逻辑从 `KlineChart.tsx` 中移入。

接口：
```typescript
interface KlineCandleLayerProps {
  klines: Kline[];
  coords: ChartCoords;
  hoveredIndex: number | null;
}
export default function KlineCandleLayer({ klines, coords, hoveredIndex }: KlineCandleLayerProps)
```

组件返回一个 `<g>` 元素（SVG group），不包含外层 `<svg>`。

- [ ] **Step 4: 抽取 StructureLiquidityLayer**

创建 `frontend/components/chart/StructureLiquidityLayer.tsx`，将结构点（BOS/CHOCH/HH/HL/LH/LL）和流动性墙区域的 SVG 渲染逻辑移入。

接口：
```typescript
interface StructureLiquidityLayerProps {
  structure: Structure | null;
  structureSeries: StructureSeriesPoint[];
  liquidity: Liquidity | null;
  liquiditySeries: LiquiditySeriesPoint[];
  coords: ChartCoords;
}
export default function StructureLiquidityLayer(props: StructureLiquidityLayerProps)
```

- [ ] **Step 5: 抽取 SignalOverlayLayer（桩版本，不含三线）**

创建 `frontend/components/chart/SignalOverlayLayer.tsx`，将信号标记点和微结构事件标记移入。三线（entry/stop/target）在 Task 5 中添加。

接口：
```typescript
export interface ActiveSignal {
  entryPrice: number;
  stopLoss: number;
  targetPrice: number;
  direction: "long" | "short";
}

interface SignalOverlayLayerProps {
  signal: Signal | null;
  signalTimeline: SignalTimelinePoint[];
  microstructureEvents: OrderFlowMicrostructureEvent[];
  visibleMicrostructureTypes: string[];
  hoveredMicrostructureMarkerKey: string | null;
  pinnedMicrostructureMarkerKey: string | null;
  coords: ChartCoords;
  activeSignal?: ActiveSignal | null; // Task 5 会用到，现在留空即可
}
export default function SignalOverlayLayer(props: SignalOverlayLayerProps)
```

- [ ] **Step 6: 更新 KlineChart.tsx 主容器组合图层**

在 `KlineChart.tsx` 中：
1. 删除移出的 SVG 渲染逻辑
2. import 三个新图层组件
3. 在 SVG 内按顺序渲染：`<KlineCandleLayer>` → `<StructureLiquidityLayer>` → `<SignalOverlayLayer>`
4. 主容器目标 ≤420 行

- [ ] **Step 7: 运行测试验证无回归**

```bash
cd frontend && npm test -- --run
```

期望：全部通过。若失败，在本 Task 内修复，不进入 Task 5。

```bash
cd frontend && npm run build
```

期望：构建成功，无 TypeScript 错误。

- [ ] **Step 8: Commit**

```bash
git add frontend/components/chart/
git commit -m "refactor: split KlineChart.tsx into focused layer components"
```

---

## Task 5: F3 — 图表三线叠加（entry/stop/target）

**Files:**
- Modify: `frontend/components/chart/SignalOverlayLayer.tsx`
- Modify: `frontend/components/chart/KlineChart.tsx`
- Create: `frontend/utils/alertUtils.ts`（`isLongDirection` 工具函数，Task 11 复用）

- [ ] **Step 1: 在 SignalOverlayLayer 中实现三线渲染**

在 `SignalOverlayLayer.tsx` 的渲染逻辑中，当 `activeSignal` 非空时，渲染三条水平虚线：

```tsx
{activeSignal && (
  <g>
    {/* Entry 线 - 绿色 */}
    {activeSignal.entryPrice >= coords.priceMin && activeSignal.entryPrice <= coords.priceMax && (
      <>
        <line
          x1={coords.paddingLeft}
          x2={coords.chartWidth - coords.paddingRight}
          y1={coords.toY(activeSignal.entryPrice)}
          y2={coords.toY(activeSignal.entryPrice)}
          stroke="#52c41a"
          strokeWidth={1}
          strokeDasharray="4 3"
          opacity={0.85}
        />
        <text
          x={coords.chartWidth - coords.paddingRight + 4}
          y={coords.toY(activeSignal.entryPrice) + 4}
          fill="#52c41a"
          fontSize={10}
        >
          {`entry ${activeSignal.entryPrice.toFixed(2)}`}
        </text>
      </>
    )}
    {/* Stop 线 - 红色 */}
    {activeSignal.stopLoss >= coords.priceMin && activeSignal.stopLoss <= coords.priceMax && (
      <>
        <line
          x1={coords.paddingLeft}
          x2={coords.chartWidth - coords.paddingRight}
          y1={coords.toY(activeSignal.stopLoss)}
          y2={coords.toY(activeSignal.stopLoss)}
          stroke="#ff4d4f"
          strokeWidth={1}
          strokeDasharray="4 3"
          opacity={0.85}
        />
        <text
          x={coords.chartWidth - coords.paddingRight + 4}
          y={coords.toY(activeSignal.stopLoss) + 4}
          fill="#ff4d4f"
          fontSize={10}
        >
          {`SL ${activeSignal.stopLoss.toFixed(2)}`}
        </text>
      </>
    )}
    {/* Target 线 - 金色 */}
    {activeSignal.targetPrice >= coords.priceMin && activeSignal.targetPrice <= coords.priceMax && (
      <>
        <line
          x1={coords.paddingLeft}
          x2={coords.chartWidth - coords.paddingRight}
          y1={coords.toY(activeSignal.targetPrice)}
          y2={coords.toY(activeSignal.targetPrice)}
          stroke="#faad14"
          strokeWidth={1}
          strokeDasharray="4 3"
          opacity={0.85}
        />
        <text
          x={coords.chartWidth - coords.paddingRight + 4}
          y={coords.toY(activeSignal.targetPrice) + 4}
          fill="#faad14"
          fontSize={10}
        >
          {`TP ${activeSignal.targetPrice.toFixed(2)}`}
        </text>
      </>
    )}
  </g>
)}
```

- [ ] **Step 2: 确认 marketStore 中当前 symbol 的字段名**

先读取 `frontend/store/marketStore.ts`，确认存储当前选中交易对的字段名（可能是 `currentSymbol`、`symbol`、`selectedSymbol` 等）。记录实际字段名，后续代码以此为准。

- [ ] **Step 3: 在 KlineChart.tsx 主容器中获取 activeSignal**

在 `KlineChart.tsx` 中，通过 `alertApi.getAlertHistory(20)` 获取最近 20 条 alert，过滤出与当前 symbol 匹配且 `kind === "setup_ready"` 的最新一条，映射为 `ActiveSignal`。

在组件内添加（字段名以 Step 2 确认为准，下方以 `currentSymbol` 为例）：

```typescript
// isLongDirection 定义在 frontend/utils/alertUtils.ts（新建），Task 11 的 ReviewChartModal 复用此函数
// export function isLongDirection(verdict?: string, tradeabilityLabel?: string): boolean {
//   const v = verdict?.toLowerCase() ?? "";
//   const t = tradeabilityLabel ?? "";
//   return v.includes("bullish") || t.includes("多");
// }

const { currentSymbol } = useMarketStore(); // 字段名以 Step 2 确认为准
const [activeSignal, setActiveSignal] = useState<ActiveSignal | null>(null);

useEffect(() => {
  let active = true;
  alertApi.getAlertHistory(20).then((feed) => {
    if (!active) return;
    const match = feed.items
      .filter((item) => item.symbol === currentSymbol && item.kind === "setup_ready" && item.entry_price > 0)
      .sort((a, b) => b.created_at - a.created_at)[0];
    if (!match) {
      setActiveSignal(null);
      return;
    }
    setActiveSignal({
      entryPrice: match.entry_price,
      stopLoss: match.stop_loss,
      targetPrice: match.target_price,
      direction: isLongDirection(match.verdict, match.tradeability_label) ? "long" : "short",
    });
  }).catch(() => { /* 静默降级 */ });
  return () => { active = false; };
}, [currentSymbol]);
```

将 `activeSignal` 传入 `<SignalOverlayLayer activeSignal={activeSignal} />`。

- [ ] **Step 3: 运行测试**

```bash
cd frontend && npm test -- --run && npm run build
```

期望：通过。

- [ ] **Step 4: Commit**

```bash
git add frontend/components/chart/SignalOverlayLayer.tsx \
        frontend/components/chart/KlineChart.tsx
git commit -m "feat: add entry/stop/target signal overlay to chart"
```

---

## Task 6: F4 — 仓位计算器

**Files:**
- Create: `frontend/components/trading/PositionCalculator.tsx`
- Create: `frontend/components/trading/PositionCalculator.test.tsx`
- Modify: `frontend/app/dashboard/page.tsx`
- Modify: `frontend/app/chart/page.tsx`（浮动 Drawer 入口）

- [ ] **Step 1: 写失败测试**

创建 `frontend/components/trading/PositionCalculator.test.tsx`：

```typescript
import { describe, it, expect } from "vitest";

// 独立计算函数测试（从 PositionCalculator 中导出）
import { calcPosition } from "./PositionCalculator";

describe("calcPosition", () => {
  it("computes position size correctly", () => {
    const result = calcPosition({ balance: 10000, riskPct: 1, entry: 100, stop: 95 });
    // stopDist% = |100-95|/100 = 5%, size = 10000*1%/5% = 2000
    expect(result.positionSize).toBeCloseTo(2000);
    expect(result.maxLoss).toBeCloseTo(100);
  });

  it("returns null when stop equals entry", () => {
    const result = calcPosition({ balance: 10000, riskPct: 1, entry: 100, stop: 100 });
    expect(result).toBeNull();
  });

  it("flags warning when position exceeds balance", () => {
    const result = calcPosition({ balance: 1000, riskPct: 1, entry: 100, stop: 99.9 });
    // stopDist% = 0.1%, size = 1000*1%/0.1% = 10000 > 1000
    expect(result?.exceedsBalance).toBe(true);
  });
});

// React 渲染测试（确保错误状态正确展示）
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import PositionCalculator from "./PositionCalculator";

// mock apiClient 和 store 避免网络/Zustand 依赖
vi.mock("@/services/apiClient", () => ({ alertApi: { getAlertHistory: vi.fn().mockResolvedValue({ items: [] }) } }));
vi.mock("@/store/marketStore", () => ({ useMarketStore: () => ({ currentSymbol: "BTCUSDT" }) }));

describe("PositionCalculator render", () => {
  it("shows error tag when stop equals entry", async () => {
    render(<PositionCalculator />);
    // 等待组件 hydrate 后手动设置相同的进场价和止损价
    const inputs = screen.getAllByRole("spinbutton");
    // 进场价 = 止损价 = 100，触发错误
    await userEvent.clear(inputs[2]); // 进场价 index
    await userEvent.type(inputs[2], "100");
    await userEvent.clear(inputs[3]); // 止损价 index
    await userEvent.type(inputs[3], "100");
    expect(await screen.findByText("止损价不能等于进场价")).toBeTruthy();
  });
});
```

- [ ] **Step 2: 运行，确认失败**

```bash
cd frontend && npm test -- PositionCalculator --run
```

期望：FAIL（模块不存在）。

- [ ] **Step 3: 实现 PositionCalculator**

创建 `frontend/components/trading/PositionCalculator.tsx`：

```tsx
"use client";

import { useEffect, useState } from "react";
import { Button, Divider, InputNumber, Tag, Tooltip } from "antd";
import { alertApi } from "@/services/apiClient";
import { useMarketStore } from "@/store/marketStore";

const LS_BALANCE_KEY = "alpha-pulse:pos-calc-balance";
const LS_RISK_KEY = "alpha-pulse:pos-calc-risk";

export interface CalcInput {
  balance: number;
  riskPct: number;
  entry: number;
  stop: number;
  target?: number;
}

export interface CalcResult {
  stopDistPct: number;
  positionSize: number;
  maxLoss: number;
  maxProfit: number | null;
  rr: number | null;
  exceedsBalance: boolean;
}

export function calcPosition(input: CalcInput): CalcResult | null {
  const { balance, riskPct, entry, stop, target } = input;
  if (entry <= 0 || Math.abs(entry - stop) < 1e-10) return null;
  const stopDistPct = Math.abs(entry - stop) / entry;
  const positionSize = (balance * (riskPct / 100)) / stopDistPct;
  const maxLoss = balance * (riskPct / 100);
  const maxProfit = target ? positionSize * (Math.abs(target - entry) / entry) : null;
  const rr = maxProfit !== null ? maxProfit / maxLoss : null;
  return { stopDistPct: stopDistPct * 100, positionSize, maxLoss, maxProfit, rr, exceedsBalance: positionSize > balance };
}

export default function PositionCalculator() {
  const { currentSymbol } = useMarketStore(); // 字段名以实际 store 为准
  const [balance, setBalance] = useState<number>(() => {
    if (typeof window === "undefined") return 10000;
    return Number(localStorage.getItem(LS_BALANCE_KEY) ?? "10000");
  });
  const [riskPct, setRiskPct] = useState<number>(() => {
    if (typeof window === "undefined") return 1;
    return Number(localStorage.getItem(LS_RISK_KEY) ?? "1");
  });
  const [entry, setEntry] = useState<number>(0);
  const [stop, setStop] = useState<number>(0);
  const [target, setTarget] = useState<number>(0);

  useEffect(() => { localStorage.setItem(LS_BALANCE_KEY, String(balance)); }, [balance]);
  useEffect(() => { localStorage.setItem(LS_RISK_KEY, String(riskPct)); }, [riskPct]);

  // 自动填入最新 setup_ready 信号
  useEffect(() => {
    let active = true;
    alertApi.getAlertHistory(20).then((feed) => {
      if (!active) return;
      const match = feed.items
        .filter((i) => i.symbol === currentSymbol && i.kind === "setup_ready" && i.entry_price > 0)
        .sort((a, b) => b.created_at - a.created_at)[0];
      if (match) {
        setEntry(match.entry_price);
        setStop(match.stop_loss);
        setTarget(match.target_price);
      }
    }).catch(() => {});
    return () => { active = false; };
  }, [currentSymbol]);

  const result = entry > 0 && stop > 0 ? calcPosition({ balance, riskPct, entry, stop, target: target > 0 ? target : undefined }) : null;

  return (
    <div style={{ padding: 16, background: "#141414", border: "1px solid #303030", borderRadius: 6 }}>
      <div style={{ fontWeight: 600, marginBottom: 12, color: "#e8e8e8" }}>仓位计算器</div>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 8 }}>
        <label style={{ fontSize: 12, color: "#888" }}>账户余额 (USDT)</label>
        <InputNumber value={balance} onChange={(v) => v !== null && setBalance(v)} min={0} style={{ width: "100%" }} />
        <label style={{ fontSize: 12, color: "#888" }}>风险比例 %</label>
        <InputNumber value={riskPct} onChange={(v) => v !== null && setRiskPct(v)} min={0.1} max={10} step={0.5} style={{ width: "100%" }} />
        <label style={{ fontSize: 12, color: "#888" }}>进场价</label>
        <InputNumber value={entry || undefined} onChange={(v) => v !== null && setEntry(v)} min={0} style={{ width: "100%" }} />
        <label style={{ fontSize: 12, color: "#888" }}>止损价</label>
        <InputNumber value={stop || undefined} onChange={(v) => v !== null && setStop(v)} min={0} style={{ width: "100%" }} />
        <label style={{ fontSize: 12, color: "#888" }}>目标价（展示）</label>
        <InputNumber value={target || undefined} onChange={(v) => v !== null && setTarget(v)} min={0} style={{ width: "100%" }} />
      </div>
      <Divider style={{ margin: "12px 0" }} />
      {result === null && entry > 0 && (
        <Tag color="error">止损价不能等于进场价</Tag>
      )}
      {result && (
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 6, fontSize: 13 }}>
          <span style={{ color: "#888" }}>止损距离</span>
          <span>{result.stopDistPct.toFixed(2)}%</span>
          <span style={{ color: "#888" }}>建议仓位</span>
          <span style={{ color: result.exceedsBalance ? "#ff7875" : "#52c41a", fontWeight: 600 }}>
            {result.positionSize.toFixed(0)} USDT
            {result.exceedsBalance && <Tooltip title="超过账户余额"><Tag color="warning" style={{ marginLeft: 4 }}>超额</Tag></Tooltip>}
          </span>
          <span style={{ color: "#888" }}>最大亏损</span>
          <span style={{ color: "#ff7875" }}>-{result.maxLoss.toFixed(0)} USDT</span>
          {result.maxProfit !== null && (
            <>
              <span style={{ color: "#888" }}>预期盈利</span>
              <span style={{ color: "#52c41a" }}>+{result.maxProfit.toFixed(0)} USDT</span>
              <span style={{ color: "#888" }}>R:R</span>
              <span>{result.rr?.toFixed(2)}</span>
            </>
          )}
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd frontend && npm test -- PositionCalculator --run
```

期望：PASS。

- [ ] **Step 5: 挂载到 Dashboard**

读取 `frontend/app/dashboard/page.tsx`，在合适的右侧面板区域 import 并渲染 `<PositionCalculator />`。

- [ ] **Step 5b: 挂载到 Chart 页（浮动按钮）**

读取 `frontend/app/chart/page.tsx`，在右下角区域添加浮动按钮或 Drawer 触发，点击展开 `<PositionCalculator />`：

```tsx
import { useState } from "react";
import { Button, Drawer } from "antd";
import { CalculatorOutlined } from "@ant-design/icons";
import PositionCalculator from "@/components/trading/PositionCalculator";

// 在 Chart 页 return 内底部
const [calcOpen, setCalcOpen] = useState(false);
// ...
<div style={{ position: "fixed", right: 24, bottom: 80, zIndex: 100 }}>
  <Button
    type="primary"
    shape="circle"
    icon={<CalculatorOutlined />}
    size="large"
    onClick={() => setCalcOpen(true)}
  />
</div>
<Drawer
  title="仓位计算器"
  open={calcOpen}
  onClose={() => setCalcOpen(false)}
  width={340}
  placement="right"
>
  <PositionCalculator />
</Drawer>
```

- [ ] **Step 6: 运行完整前端测试**

```bash
cd frontend && npm test -- --run && npm run lint
```

期望：全部通过。

- [ ] **Step 7: Commit**

```bash
git add frontend/components/trading/ \
        frontend/app/dashboard/page.tsx \
        frontend/app/chart/page.tsx
git commit -m "feat: add position calculator component"
```

---

## Task 7: S1 — alert_records schema + repo 方法

**Files:**
- Modify: `backend/models/alert_record.go`
- Modify: `backend/repository/alert_record_repo.go`
- Modify: `backend/repository/kline_repo.go`

- [ ] **Step 1: 写失败测试 — FindPending 和 UpdateOutcome**

创建 `backend/repository/alert_record_outcome_test.go`（`package repository`）。

使用 SQLite in-memory 数据库（`gorm.io/driver/sqlite`）构造最小测试 DB：

```go
package repository

import (
    "testing"
    "time"

    "alpha-pulse/backend/models"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("open sqlite: %v", err)
    }
    if err := db.AutoMigrate(&models.AlertRecord{}); err != nil {
        t.Fatalf("migrate: %v", err)
    }
    return db
}

func TestFindPendingReturnsOnlyPendingRecords(t *testing.T) {
    db := newTestDB(t)
    repo := NewAlertRecordRepository(db)

    now := time.Now().UnixMilli()
    _ = db.Create(&models.AlertRecord{Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "pending", EventTime: now}).Error
    _ = db.Create(&models.AlertRecord{Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "target_hit", EventTime: now + 1}).Error
    _ = db.Create(&models.AlertRecord{Symbol: "ETHUSDT", Kind: "setup_ready", Outcome: "pending", EventTime: now + 2}).Error

    records, err := repo.FindPending("BTCUSDT", 100)
    if err != nil {
        t.Fatalf("FindPending: %v", err)
    }
    if len(records) != 1 {
        t.Fatalf("expected 1 pending record for BTCUSDT, got %d", len(records))
    }
    if records[0].Outcome != "pending" {
        t.Fatalf("expected outcome=pending, got %s", records[0].Outcome)
    }
}

func TestUpdateOutcomeDoesNotOverwriteOtherFields(t *testing.T) {
    db := newTestDB(t)
    repo := NewAlertRecordRepository(db)

    now := time.Now().UnixMilli()
    record := &models.AlertRecord{
        Symbol:    "BTCUSDT",
        Kind:      "setup_ready",
        Outcome:   "pending",
        EventTime: now,
        EntryPrice: 100,
        StopLoss:   95,
    }
    if err := db.Create(record).Error; err != nil {
        t.Fatalf("create: %v", err)
    }

    if err := repo.UpdateOutcome(record.ID, "target_hit", 110.0, now+1000, 2.0); err != nil {
        t.Fatalf("UpdateOutcome: %v", err)
    }

    var updated models.AlertRecord
    db.First(&updated, record.ID)
    if updated.Outcome != "target_hit" {
        t.Fatalf("expected target_hit, got %s", updated.Outcome)
    }
    if updated.EntryPrice != 100 {
        t.Fatalf("UpdateOutcome should not change EntryPrice, got %f", updated.EntryPrice)
    }
    if updated.ActualRR != 2.0 {
        t.Fatalf("expected actual_rr=2.0, got %f", updated.ActualRR)
    }
}
```

- [ ] **Step 1b: 运行，确认失败**

```bash
cd backend && go test ./repository/... -run TestFindPending -v
```

期望：FAIL（`FindPending` / `UpdateOutcome` 未定义）。如果 SQLite driver 未安装，先运行：
```bash
go get gorm.io/driver/sqlite
```

- [ ] **Step 2: 给 AlertRecord 新增字段**

在 `backend/models/alert_record.go` 的 `AlertRecord` struct 中，在 `EventTime` 字段之后新增：

```go
Interval     string  `gorm:"column:interval;size:8;not null;default:'1h';comment:触发告警时的参考周期" json:"interval"`
Outcome      string  `gorm:"column:outcome;size:24;index;not null;default:'pending';comment:pending/target_hit/stop_hit/expired" json:"outcome"`
OutcomePrice float64 `gorm:"column:outcome_price;type:decimal(18,8);not null;default:0;comment:结果触发价格" json:"outcome_price"`
OutcomeAt    int64   `gorm:"column:outcome_at;not null;default:0;comment:结果触发时间 Unix ms" json:"outcome_at"`
ActualRR     float64 `gorm:"column:actual_rr;type:decimal(10,4);not null;default:0;comment:实际盈亏比" json:"actual_rr"`
```

- [ ] **Step 3: 更新 alert_record_repo.go 的 upsert 列表**

确认 `backend/repository/alert_record_repo.go` 的 `Create` 方法中 `DoUpdates` 的列列表**不包含** `interval`、`outcome`、`outcome_price`、`outcome_at`、`actual_rr`。这 5 个字段不应出现在 upsert 更新列表中（它们已在上面的 default 值中处理，且 outcome 需要被 OutcomeTracker 单独维护）。

当前列表仅更新：`symbol, kind, severity, direction_state, tradable, setup_ready, tradeability_label, title, verdict, summary, confidence, risk_label, entry_price, stop_loss, target_price, risk_reward, event_time, payload_json, created_at`。

新增两个方法：

```go
// FindPending 返回指定标的中 outcome 为 pending 的记录（按 event_time 升序）。
func (r *AlertRecordRepository) FindPending(symbol string, limit int) ([]models.AlertRecord, error) {
    if limit <= 0 {
        limit = 100
    }
    records := make([]models.AlertRecord, 0, limit)
    err := r.db.Where("symbol = ? AND outcome = ?", symbol, "pending").
        Order("event_time ASC").
        Limit(limit).
        Find(&records).Error
    return records, err
}

// UpdateOutcome 更新单条记录的结果字段，不影响其他字段。
func (r *AlertRecordRepository) UpdateOutcome(id uint64, outcome string, outcomePrice float64, outcomeAt int64, actualRR float64) error {
    return r.db.Model(&models.AlertRecord{}).
        Where("id = ?", id).
        Updates(map[string]any{
            "outcome":       outcome,
            "outcome_price": outcomePrice,
            "outcome_at":    outcomeAt,
            "actual_rr":     actualRR,
        }).Error
}
```

- [ ] **Step 4: 给 kline_repo.go 新增 FindAfter 和 FindBefore**

在 `backend/repository/kline_repo.go` 末尾添加：

```go
// FindAfter 返回指定时间点（afterMs，Unix 毫秒）之后的 K 线，按时间升序。
func (r *KlineRepository) FindAfter(symbol, interval string, afterMs int64, limit int) ([]models.Kline, error) {
    if limit <= 0 {
        limit = 60
    }
    klines := make([]models.Kline, 0, limit)
    err := r.db.
        Where("symbol = ? AND interval_type = ? AND open_time > ?", symbol, interval, afterMs).
        Order("open_time ASC").
        Limit(limit).
        Find(&klines).Error
    return klines, err
}

// FindBefore 返回指定时间点（beforeMs，Unix 毫秒）之前的最近 N 根 K 线，按时间升序返回。
func (r *KlineRepository) FindBefore(symbol, interval string, beforeMs int64, limit int) ([]models.Kline, error) {
    if limit <= 0 {
        limit = 20
    }
    klines := make([]models.Kline, 0, limit)
    err := r.db.
        Where("symbol = ? AND interval_type = ? AND open_time < ?", symbol, interval, beforeMs).
        Order("open_time DESC").
        Limit(limit).
        Find(&klines).Error
    if err != nil {
        return nil, err
    }
    // 倒序查最新的，返回时恢复升序
    for left, right := 0, len(klines)-1; left < right; left, right = left+1, right-1 {
        klines[left], klines[right] = klines[right], klines[left]
    }
    return klines, nil
}
```

- [ ] **Step 5: 编译验证**

```bash
cd backend && go build ./...
```

期望：零错误。

- [ ] **Step 6: 更新 alert_persistence.go 填充 Interval**

在 `backend/internal/service/alert_persistence.go` 的 `projectAlertRecord` 函数中，新增 `Interval` 字段填充：

```go
Interval: alertBiasInterval, // "1h"，即触发周期
```

（`alertBiasInterval` 常量定义在 `alert_service.go` 同包，值为 `"1h"`）

- [ ] **Step 7: Commit**

```bash
cd backend && go build ./... && cd ..
git add backend/models/alert_record.go \
        backend/repository/alert_record_repo.go \
        backend/repository/kline_repo.go \
        backend/internal/service/alert_persistence.go
git commit -m "feat(s1): add outcome fields to alert_records and repo methods"
```

---

## Task 8: S1 — OutcomeTrackerService + 单元测试

**Files:**
- Create: `backend/internal/service/outcome_tracker.go`
- Create: `backend/internal/service/outcome_tracker_test.go`

- [ ] **Step 1: 写失败测试**

创建 `backend/internal/service/outcome_tracker_test.go`（`package service`）：

```go
package service

import (
    "testing"
    "time"
    "alpha-pulse/backend/models"
)

// buildTestKline 构造测试用 K 线。
func buildTestKline(openTime int64, open, high, low, close float64) models.Kline {
    return models.Kline{
        OpenTime:   openTime,
        OpenPrice:  open,
        HighPrice:  high,
        LowPrice:   low,
        ClosePrice: close,
    }
}

func TestEvalOutcomeLongTargetHit(t *testing.T) {
    record := models.AlertRecord{
        DirectionState: "strong-bullish",
        EntryPrice:     100,
        StopLoss:       95,
        TargetPrice:    110,
        EventTime:      1000,
        Outcome:        "pending",
    }
    klines := []models.Kline{
        buildTestKline(1001, 101, 105, 100, 103),
        buildTestKline(1002, 103, 112, 102, 111), // high >= target
    }
    outcome, price, _ := evalOutcome(record, klines, time.UnixMilli(2000))
    if outcome != "target_hit" {
        t.Fatalf("expected target_hit, got %s", outcome)
    }
    if price != 110 {
        t.Fatalf("expected outcomePrice=110, got %f", price)
    }
}

func TestEvalOutcomeLongStopHitFirst(t *testing.T) {
    record := models.AlertRecord{
        DirectionState: "bullish",
        EntryPrice:     100,
        StopLoss:       95,
        TargetPrice:    110,
        EventTime:      1000,
        Outcome:        "pending",
    }
    // 同一根 K 线同时触达止损和目标，止损优先
    klines := []models.Kline{
        buildTestKline(1001, 100, 115, 94, 105), // low<=stop AND high>=target
    }
    outcome, _, _ := evalOutcome(record, klines, time.UnixMilli(2000))
    if outcome != "stop_hit" {
        t.Fatalf("expected stop_hit (stop priority), got %s", outcome)
    }
}

func TestEvalOutcomeExpired(t *testing.T) {
    record := models.AlertRecord{
        DirectionState: "bullish",
        EntryPrice:     100,
        StopLoss:       95,
        TargetPrice:    110,
        EventTime:      1000,
        Outcome:        "pending",
    }
    klines := []models.Kline{} // 没有 K 线
    // now = event_time + 61 分钟
    now := time.UnixMilli(1000 + 61*60*1000)
    outcome, _, _ := evalOutcome(record, klines, now)
    if outcome != "expired" {
        t.Fatalf("expected expired, got %s", outcome)
    }
}

func TestEvalOutcomeShortStopHitFirst(t *testing.T) {
    record := models.AlertRecord{
        DirectionState: "strong-bearish",
        EntryPrice:     100,
        StopLoss:       105,
        TargetPrice:    90,
        EventTime:      1000,
        Outcome:        "pending",
    }
    klines := []models.Kline{
        buildTestKline(1001, 101, 106, 89, 95), // high>=stop AND low<=target
    }
    outcome, _, _ := evalOutcome(record, klines, time.UnixMilli(2000))
    if outcome != "stop_hit" {
        t.Fatalf("expected stop_hit for short, got %s", outcome)
    }
}

func TestEvalOutcomeNeutralDirectionSkipped(t *testing.T) {
    record := models.AlertRecord{
        DirectionState: "neutral",
        EntryPrice:     100,
        StopLoss:       95,
        TargetPrice:    110,
        EventTime:      1000,
        Outcome:        "pending",
    }
    klines := []models.Kline{buildTestKline(1001, 100, 120, 80, 100)}
    outcome, _, _ := evalOutcome(record, klines, time.UnixMilli(2000))
    if outcome != "" {
        t.Fatalf("expected no tracking for neutral direction, got %s", outcome)
    }
}
```

- [ ] **Step 2: 运行，确认失败**

```bash
cd backend && go test ./internal/service/... -run TestEvalOutcome -v
```

期望：FAIL（`evalOutcome` 未定义）。

- [ ] **Step 3: 实现 OutcomeTrackerService**

创建 `backend/internal/service/outcome_tracker.go`：

```go
package service

import (
    "context"
    "log"
    "time"

    "alpha-pulse/backend/models"
    "alpha-pulse/backend/repository"
)

const outcomeExpiryMs = 60 * 60 * 1000 // 60 分钟固定窗口

// OutcomeTrackerService 定期结算 alert_records 中 pending 信号的结果。
type OutcomeTrackerService struct {
    alertRecordRepo *repository.AlertRecordRepository
    klineRepo       *repository.KlineRepository
    symbols         []string
}

// NewOutcomeTrackerService 创建 OutcomeTrackerService。
func NewOutcomeTrackerService(alertRecordRepo *repository.AlertRecordRepository, klineRepo *repository.KlineRepository, symbols []string) *OutcomeTrackerService {
    return &OutcomeTrackerService{
        alertRecordRepo: alertRecordRepo,
        klineRepo:       klineRepo,
        symbols:         normalizeAlertSymbols(symbols),
    }
}

// TrackAll 遍历所有标的，结算 pending 信号结果。
func (t *OutcomeTrackerService) TrackAll(ctx context.Context) {
    for _, symbol := range t.symbols {
        select {
        case <-ctx.Done():
            return
        default:
        }
        if err := t.trackSymbol(ctx, symbol); err != nil {
            log.Printf("outcome tracker failed for %s: %v", symbol, err)
        }
    }
}

func (t *OutcomeTrackerService) trackSymbol(ctx context.Context, symbol string) error {
    records, err := t.alertRecordRepo.FindPending(symbol, 100)
    if err != nil {
        return err
    }

    now := time.Now()
    for _, record := range records {
        interval := record.Interval
        if interval == "" {
            interval = "1h"
        }

        klines, err := t.klineRepo.FindAfter(symbol, interval, record.EventTime, 60)
        if err != nil {
            log.Printf("kline fetch failed for outcome tracking %s %s: %v", symbol, interval, err)
            continue
        }

        outcome, outcomePrice, outcomeAt := evalOutcome(record, klines, now)
        if outcome == "" {
            continue // 还未结算，方向不可追踪
        }

        actualRR := 0.0
        if outcome == "target_hit" && record.RiskReward > 0 {
            actualRR = record.RiskReward
        } else if outcome == "stop_hit" {
            actualRR = -1.0
        }

        if err := t.alertRecordRepo.UpdateOutcome(record.ID, outcome, outcomePrice, outcomeAt, actualRR); err != nil {
            log.Printf("outcome update failed for record %d: %v", record.ID, err)
        }
    }
    return nil
}

// evalOutcome 根据 K 线数据判断信号结果。返回 ("", 0, 0) 表示无法追踪或尚未结算。
func evalOutcome(record models.AlertRecord, klines []models.Kline, now time.Time) (outcome string, outcomePrice float64, outcomeAt int64) {
    // 判断方向
    isLong := record.DirectionState == "strong-bullish" || record.DirectionState == "bullish"
    isShort := record.DirectionState == "strong-bearish" || record.DirectionState == "bearish"
    if !isLong && !isShort {
        return "", 0, 0 // 中性方向不追踪
    }

    // 过期检查
    if now.UnixMilli()-record.EventTime > outcomeExpiryMs && len(klines) == 0 {
        return "expired", 0, now.UnixMilli()
    }

    // 逐根 K 线扫描（止损优先）
    for _, k := range klines {
        if isLong {
            if k.LowPrice <= record.StopLoss {
                return "stop_hit", record.StopLoss, k.OpenTime
            }
            if k.HighPrice >= record.TargetPrice {
                return "target_hit", record.TargetPrice, k.OpenTime
            }
        } else { // isShort
            if k.HighPrice >= record.StopLoss {
                return "stop_hit", record.StopLoss, k.OpenTime
            }
            if k.LowPrice <= record.TargetPrice {
                return "target_hit", record.TargetPrice, k.OpenTime
            }
        }
    }

    // K 线扫完未命中，检查是否过期
    if now.UnixMilli()-record.EventTime > outcomeExpiryMs {
        return "expired", 0, now.UnixMilli()
    }

    return "", 0, 0 // 仍在观察窗口内，本轮不更新（trackSymbol 检查 outcome == "" 时跳过写库）
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd backend && go test ./internal/service/... -run TestEvalOutcome -v
```

期望：全部 PASS。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/outcome_tracker.go \
        backend/internal/service/outcome_tracker_test.go
git commit -m "feat(s1): implement OutcomeTrackerService with stop-priority logic"
```

---

## Task 9: S1 — 调度器集成 + main.go 依赖注入

**Files:**
- Modify: `backend/internal/scheduler/jobs.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: 更新 Jobs 结构体和 NewJobs**

在 `backend/internal/scheduler/jobs.go` 中：

将 `Jobs` struct 新增字段：
```go
outcomeTracker *service.OutcomeTrackerService
```

将 `NewJobs` 新增参数（在 `alertService` 之后）：
```go
outcomeTracker *service.OutcomeTrackerService,
```

在 `return &Jobs{...}` 中新增：
```go
outcomeTracker: outcomeTracker,
```

在 `runOnce()` 末尾追加：
```go
if j.outcomeTracker != nil {
    j.outcomeTracker.TrackAll(context.Background())
}
```

在文件顶部 import 中确认有 `"context"`（已有则无需改动）。

- [ ] **Step 2: 更新 main.go 依赖注入**

在 `backend/cmd/server/main.go` 中，在 `alertService` 构建之后、`scheduler.NewJobs` 调用之前插入：

```go
outcomeTracker := service.NewOutcomeTrackerService(
    alertRecordRepo,
    klineRepo,
    cfg.MarketSymbols,
)
```

将 `scheduler.NewJobs(...)` 调用更新，在 `alertService` 参数之后新增 `outcomeTracker`：

```go
jobs := scheduler.NewJobs(
    marketService,
    signalService,
    alertService,
    outcomeTracker,   // 新增
    cfg.MarketSymbols,
    time.Duration(cfg.SchedulerIntervalSeconds)*time.Second,
)
```

- [ ] **Step 3: 编译 + 运行完整后端测试**

```bash
cd backend && go build ./... && go test ./...
```

期望：全部通过。

- [ ] **Step 4: Commit**

```bash
git add backend/internal/scheduler/jobs.go backend/cmd/server/main.go
git commit -m "feat(s1): integrate OutcomeTracker into scheduler"
```

---

## Task 10: S2 — 后端 kline before_ts/after_ts + AlertStats

**Files:**
- Modify: `backend/internal/service/market_service.go`
- Modify: `backend/internal/handler/market_handler.go`
- Modify: `backend/internal/service/alert_service.go`
- Modify: `backend/internal/handler/alert_handler.go`
- Modify: `backend/router/router.go`

- [ ] **Step 1: market_service.go 新增 GetKlineBefore/After**

阅读 `backend/internal/service/market_service.go` 中 `GetKline` 相关方法，新增：

```go
// GetKlineBefore 返回指定时间点之前的 N 根 K 线（升序）。
func (s *MarketService) GetKlineBefore(symbol, interval string, beforeMs int64, limit int) ([]models.Kline, error) {
    return s.klineRepo.FindBefore(symbol, interval, beforeMs, limit)
}

// GetKlineAfter 返回指定时间点之后的 N 根 K 线（升序）。
func (s *MarketService) GetKlineAfter(symbol, interval string, afterMs int64, limit int) ([]models.Kline, error) {
    return s.klineRepo.FindAfter(symbol, interval, afterMs, limit)
}
```

- [ ] **Step 2: market_handler.go GetKline 支持 before_ts/after_ts**

阅读 `backend/internal/handler/market_handler.go` 中 `GetKline` handler，在解析查询参数时新增：

```go
beforeTs := parseOptionalInt64(c.Query("before_ts"))
afterTs  := parseOptionalInt64(c.Query("after_ts"))
```

当 `beforeTs > 0` 时调用 `marketService.GetKlineBefore`，当 `afterTs > 0` 时调用 `marketService.GetKlineAfter`，否则走原有 `GetKline` 逻辑。

新增辅助函数（在 handler 包的工具文件中，如 `handler_utils.go` 或直接在文件末尾）：
```go
func parseOptionalInt64(s string) int64 {
    if s == "" { return 0 }
    v, err := strconv.ParseInt(s, 10, 64)
    if err != nil { return 0 }
    return v
}
```

- [ ] **Step 3: alert_service.go 新增 GetAlertStats**

在 `backend/internal/service/alert_service.go` 中添加 `AlertStats` 类型和 `GetAlertStats` 方法：

```go
type AlertStats struct {
    Symbol          string  `json:"symbol"`
    Total           int     `json:"total"`
    TargetHit       int     `json:"target_hit"`
    StopHit         int     `json:"stop_hit"`
    Pending         int     `json:"pending"`
    Expired         int     `json:"expired"`
    WinRate         float64 `json:"win_rate"`
    AvgRR           float64 `json:"avg_rr"`
    SampleSizeLabel string  `json:"sample_size_label"`
}

func (s *AlertService) GetAlertStats(symbol string, limit int) AlertStats {
    if limit <= 0 {
        limit = 50
    }
    if s.repo == nil {
        return AlertStats{Symbol: symbol}
    }
    records, err := s.repo.GetStatsBySymbol(symbol, limit)
    if err != nil {
        log.Printf("alert stats query failed: %v", err)
        return AlertStats{Symbol: symbol}
    }
    stats := AlertStats{Symbol: symbol, Total: len(records)}
    rrSum := 0.0
    rrCount := 0
    for _, r := range records {
        switch r.Outcome {
        case "target_hit":
            stats.TargetHit++
            if r.ActualRR > 0 {
                rrSum += r.ActualRR
                rrCount++
            }
        case "stop_hit":
            stats.StopHit++
            if r.ActualRR != 0 {
                rrSum += r.ActualRR
                rrCount++
            }
        case "expired":
            stats.Expired++
        default:
            stats.Pending++
        }
    }
    decided := stats.TargetHit + stats.StopHit
    if decided > 0 {
        stats.WinRate = float64(stats.TargetHit) / float64(decided) * 100
    }
    if rrCount > 0 {
        stats.AvgRR = rrSum / float64(rrCount)
    }
    stats.SampleSizeLabel = fmt.Sprintf("近 %d 条", stats.Total)
    return stats
}
```

- [ ] **Step 4: alert_record_repo.go 新增 GetStatsBySymbol**

在 `backend/repository/alert_record_repo.go` 新增：

```go
// GetStatsBySymbol 返回指定标的最近 N 条 setup_ready 记录用于胜率统计。
// 注意：不过滤 outcome，让上层 Go 代码按 outcome 字段分类计数，
// 确保 pending 字段能被正确统计（与 spec 中 win_rate 分母排除 pending 的逻辑一致）。
func (r *AlertRecordRepository) GetStatsBySymbol(symbol string, limit int) ([]models.AlertRecord, error) {
    if limit <= 0 {
        limit = 50
    }
    records := make([]models.AlertRecord, 0, limit)
    q := r.db.Select("id, outcome, actual_rr").
        Where("kind = ?", "setup_ready")
    if symbol != "" {
        q = q.Where("symbol = ?", symbol)
    }
    err := q.Order("event_time DESC").Limit(limit).Find(&records).Error
    return records, err
}
```

- [ ] **Step 5: alert_handler.go 新增 GetAlertStats**

在 `backend/internal/handler/alert_handler.go` 新增：

```go
func (h *AlertHandler) GetAlertStats(c *gin.Context) {
    symbol := strings.ToUpper(strings.TrimSpace(c.Query("symbol")))
    limit := parseLimit(c.DefaultQuery("limit", "50"), 50)
    stats := h.alertService.GetAlertStats(symbol, limit)
    c.JSON(http.StatusOK, utils.Success(stats))
}
```

在文件顶部 import 中确认有 `"strings"`。

- [ ] **Step 6: router.go 注册路由**

在 `backend/router/router.go` 的 alert 路由组中新增：

```go
protected.GET("/alerts/stats", handlers.Alert.GetAlertStats)
```

- [ ] **Step 7: 编译 + 测试**

```bash
cd backend && go build ./... && go test ./...
```

期望：通过。

- [ ] **Step 8: Commit**

```bash
git add backend/internal/service/market_service.go \
        backend/internal/handler/market_handler.go \
        backend/internal/service/alert_service.go \
        backend/repository/alert_record_repo.go \
        backend/internal/handler/alert_handler.go \
        backend/router/router.go
git commit -m "feat(s2/s3): add kline before_ts/after_ts and alert stats endpoint"
```

---

## Task 11: S2 — ReviewChartModal 前端

**Files:**
- Modify: `frontend/types/alert.ts`
- Modify: `frontend/services/apiClient.ts`
- Modify: `frontend/components/alerts/AlertEventCard.tsx`
- Create: `frontend/components/alerts/ReviewChartModal.tsx`
- Modify: `frontend/components/chart/KlineChart.tsx`
- Modify: `frontend/utils/alertUtils.ts`（添加 `isLongDirection` 实际实现，Task 5 已预留注释版本）

- [ ] **Step 1: 更新类型**

在 `frontend/types/alert.ts` 新增：

```typescript
export type AlertOutcome = "pending" | "target_hit" | "stop_hit" | "expired";

// AlertEvent 新增（可选，兼容旧数据）
// 在 AlertEvent interface 中追加：
//   outcome?: AlertOutcome;
//   outcome_price?: number;
//   outcome_at?: number;
//   actual_rr?: number;
//   interval?: string;

export interface AlertStats {
  symbol: string;
  total: number;
  target_hit: number;
  stop_hit: number;
  pending: number;
  expired: number;
  win_rate: number;
  avg_rr: number;
  sample_size_label: string;
}
```

在 `AlertEvent` interface 中追加可选字段：
```typescript
outcome?: AlertOutcome;
outcome_price?: number;
outcome_at?: number;
actual_rr?: number;
interval?: string;
```

- [ ] **Step 2: 更新 apiClient.ts**

在 `alertApi` 中新增：
```typescript
getAlertStats(symbol?: string, limit = 50) {
  const params = new URLSearchParams();
  if (symbol) params.set("symbol", symbol);
  params.set("limit", String(limit));
  return request<AlertStats>(`/alerts/stats?${params.toString()}`);
},
```

在现有 kline 请求函数中（找到 `marketApi.getKline` 或类似方法），支持可选 `before_ts` / `after_ts` 参数。阅读 `apiClient.ts` 确认现有 kline 方法签名，按以下方式扩展：

```typescript
getKline(symbol: string, interval: string, limit = 48, opts?: { before_ts?: number; after_ts?: number }) {
  const params = new URLSearchParams({ symbol, interval, limit: String(limit) });
  if (opts?.before_ts) params.set("before_ts", String(opts.before_ts));
  if (opts?.after_ts) params.set("after_ts", String(opts.after_ts));
  return request<Kline[]>(`/kline?${params.toString()}`);
},
```

- [ ] **Step 3: AlertEventCard 新增复盘按钮**

阅读 `frontend/components/alerts/AlertEventCard.tsx`，在告警卡片右下角新增"复盘"按钮：

```tsx
<Button
  size="small"
  icon={<HistoryOutlined />}
  onClick={() => onReview?.(event)}
  style={{ marginLeft: 8 }}
>
  复盘
</Button>
```

在 `AlertEventCard` 的 props 中新增：
```typescript
onReview?: (event: AlertEvent) => void;
```

- [ ] **Step 4: 创建 ReviewChartModal**

创建 `frontend/components/alerts/ReviewChartModal.tsx`：

```tsx
"use client";

import { useEffect, useState } from "react";
import { Badge, Modal, Spin, Tag, Typography } from "antd";
import { marketApi } from "@/services/apiClient";
import type { AlertEvent, AlertOutcome } from "@/types/alert";
import type { Kline } from "@/types/market";
import KlineChart from "@/components/chart/KlineChart";
import type { ActiveSignal } from "@/components/chart/SignalOverlayLayer";
import { isLongDirection } from "@/utils/alertUtils";

interface ReviewChartModalProps {
  event: AlertEvent | null;
  open: boolean;
  onClose: () => void;
}

const OUTCOME_COLOR: Record<AlertOutcome, string> = {
  target_hit: "success",
  stop_hit: "error",
  pending: "processing",
  expired: "default",
};

const OUTCOME_LABEL: Record<AlertOutcome, string> = {
  target_hit: "命中目标",
  stop_hit: "触发止损",
  pending: "观察中",
  expired: "已过期",
};

export default function ReviewChartModal({ event, open, onClose }: ReviewChartModalProps) {
  const [klines, setKlines] = useState<Kline[]>([]);
  const [loading, setLoading] = useState(false);
  const [noData, setNoData] = useState(false);

  useEffect(() => {
    if (!open || !event) return;
    let active = true;
    setLoading(true);
    setNoData(false);

    const symbol = event.symbol;
    const interval = event.interval ?? "1h";
    const centerTs = event.created_at;

    Promise.all([
      marketApi.getKline(symbol, interval, 20, { before_ts: centerTs }),
      marketApi.getKline(symbol, interval, 40, { after_ts: centerTs }),
    ]).then(([before, after]) => {
      if (!active) return;
      const merged = [...before, ...after].sort((a, b) => a.open_time - b.open_time);
      if (merged.length === 0) {
        setNoData(true);
      } else {
        setKlines(merged);
      }
    }).catch(() => {
      if (active) setNoData(true);
    }).finally(() => {
      if (active) setLoading(false);
    });

    return () => { active = false; };
  }, [open, event]);

  const activeSignal: ActiveSignal | null = event && event.entry_price > 0
    ? {
        entryPrice: event.entry_price,
        stopLoss: event.stop_loss,
        targetPrice: event.target_price,
        direction: isLongDirection(event.verdict, event.tradeability_label) ? "long" : "short",
      }
    : null;

  const outcome = event?.outcome as AlertOutcome | undefined;

  return (
    <Modal
      open={open}
      onCancel={onClose}
      footer={null}
      width={1040}
      title={
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <span>复盘 — {event?.symbol} {event?.interval?.toUpperCase()}</span>
          {outcome && <Badge status={OUTCOME_COLOR[outcome] as any} text={OUTCOME_LABEL[outcome]} />}
          {event?.confidence != null && <Tag>{event.confidence}% 置信度</Tag>}
        </div>
      }
    >
      {loading && <div style={{ textAlign: "center", padding: 40 }}><Spin /></div>}
      {!loading && noData && (
        <Typography.Text type="secondary">数据不足，无法还原当时走势。</Typography.Text>
      )}
      {!loading && !noData && klines.length > 0 && (
        <KlineChart
          historicalMode={{ klines, symbol: event?.symbol ?? "", interval: event?.interval ?? "1h" }}
          activeSignal={activeSignal}
        />
      )}
    </Modal>
  );
}
```

- [ ] **Step 5: KlineChart.tsx 支持 historicalMode prop**

在 `KlineChart.tsx` 主容器中新增 prop：

```typescript
interface KlineChartProps {
  historicalMode?: {
    klines: Kline[];
    symbol: string;
    interval: string;
  };
  activeSignal?: ActiveSignal | null;
}
```

当 `historicalMode` 存在时，使用传入的 `klines` 数据替代 Zustand store 的数据，同时跳过 store 的订阅（`useMarketStore` 调用需要条件保护）。

将 `activeSignal` prop 传递给 `SignalOverlayLayer`（覆盖内部的 fetch 逻辑）。

> **注意：** historicalMode 下不调用 `refreshDashboard`，不触发 store 更新。

- [ ] **Step 6: 更新 Review 页调用 AlertEventCard 的地方，传入 onReview**

阅读 `frontend/app/review/page.tsx`，找到渲染 `AlertEventCard` 的代码，添加 `onReview` 回调以打开 `ReviewChartModal`：

```tsx
const [reviewEvent, setReviewEvent] = useState<AlertEvent | null>(null);
// ...
<AlertEventCard event={item} onReview={setReviewEvent} />
// ...
<ReviewChartModal
  event={reviewEvent}
  open={!!reviewEvent}
  onClose={() => setReviewEvent(null)}
/>
```

- [ ] **Step 7: 运行测试 + 构建**

```bash
cd frontend && npm test -- --run && npm run build
```

期望：通过。

- [ ] **Step 8: Commit**

```bash
git add frontend/types/alert.ts \
        frontend/services/apiClient.ts \
        frontend/components/alerts/AlertEventCard.tsx \
        frontend/components/alerts/ReviewChartModal.tsx \
        frontend/components/chart/KlineChart.tsx \
        frontend/app/review/page.tsx
git commit -m "feat(s2): add review chart modal with k-line context"
```

---

## Task 12: S3 — WinRatePanel 前端

**Files:**
- Create: `frontend/components/alerts/WinRatePanel.tsx`
- Create: `frontend/components/alerts/WinRatePanel.test.tsx`
- Modify: `frontend/app/review/page.tsx`

- [ ] **Step 1: 写失败测试**

创建 `frontend/components/alerts/WinRatePanel.test.tsx`：

```typescript
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import WinRatePanel from "./WinRatePanel";
import * as apiClient from "@/services/apiClient";

vi.mock("@/services/apiClient", () => ({
  alertApi: {
    getAlertStats: vi.fn().mockResolvedValue({
      symbol: "BTCUSDT",
      total: 10,
      target_hit: 6,
      stop_hit: 3,
      pending: 1,
      expired: 0,
      win_rate: 66.7,
      avg_rr: 1.8,
      sample_size_label: "近 10 条",
    }),
  },
}));

describe("WinRatePanel", () => {
  it("renders win rate percentage", async () => {
    render(<WinRatePanel symbols={["BTCUSDT"]} />);
    const cell = await screen.findByText(/66\.7/);
    expect(cell).toBeTruthy();
  });
});
```

- [ ] **Step 2: 运行，确认失败**

```bash
cd frontend && npm test -- WinRatePanel --run
```

期望：FAIL（模块不存在）。

- [ ] **Step 3: 实现 WinRatePanel**

创建 `frontend/components/alerts/WinRatePanel.tsx`：

```tsx
"use client";

import { useEffect, useState } from "react";
import { Button, Spin, Statistic, Table, Tag } from "antd";
import { alertApi } from "@/services/apiClient";
import type { AlertStats } from "@/types/alert";

const LIMIT_OPTIONS = [20, 50, 0] as const; // 0 = 全部
type Limit = (typeof LIMIT_OPTIONS)[number];

interface WinRatePanelProps {
  symbols: string[];
}

export default function WinRatePanel({ symbols }: WinRatePanelProps) {
  const [limit, setLimit] = useState<Limit>(50);
  const [stats, setStats] = useState<AlertStats[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    setLoading(true);
    Promise.all(
      symbols.map((s) => alertApi.getAlertStats(s, limit))
    ).then((results) => {
      if (active) setStats(results);
    }).finally(() => {
      if (active) setLoading(false);
    });
    return () => { active = false; };
  }, [symbols, limit]);

  const columns = [
    { title: "标的", dataIndex: "symbol", key: "symbol", render: (v: string) => <Tag>{v.replace("USDT", "")}</Tag> },
    { title: "胜率", dataIndex: "win_rate", key: "win_rate", render: (v: number) => <Statistic value={v} precision={1} suffix="%" valueStyle={{ fontSize: 14, color: v >= 55 ? "#52c41a" : "#ff7875" }} /> },
    { title: "平均 R:R", dataIndex: "avg_rr", key: "avg_rr", render: (v: number) => v > 0 ? v.toFixed(2) : "—" },
    { title: "命中/止损", key: "hits", render: (_: any, row: AlertStats) => `${row.target_hit} / ${row.stop_hit}` },
    { title: "样本", dataIndex: "sample_size_label", key: "sample_size_label" },
  ];

  return (
    <div style={{ marginBottom: 24 }}>
      <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 12 }}>
        <span style={{ fontWeight: 600 }}>历史胜率统计</span>
        {LIMIT_OPTIONS.map((l) => (
          <Button
            key={l}
            size="small"
            type={limit === l ? "primary" : "default"}
            onClick={() => setLimit(l)}
          >
            {l === 0 ? "全部" : `近${l}条`}
          </Button>
        ))}
      </div>
      {loading ? (
        <Spin />
      ) : (
        <Table
          dataSource={stats}
          columns={columns}
          rowKey="symbol"
          pagination={false}
          size="small"
        />
      )}
    </div>
  );
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd frontend && npm test -- WinRatePanel --run
```

期望：PASS。

- [ ] **Step 5: 挂载到 /review 页**

在 `frontend/app/review/page.tsx` 中，在告警列表之前添加：

```tsx
import WinRatePanel from "@/components/alerts/WinRatePanel";
// ...
<WinRatePanel symbols={["BTCUSDT", "ETHUSDT", "SOLUSDT"]} />
```

- [ ] **Step 6: 运行完整测试**

```bash
cd frontend && npm test -- --run && npm run build
```

期望：全部通过。

- [ ] **Step 7: 后端单元测试 — AlertStats 胜率公式**

在 `backend/internal/service/alert_stats_test.go`（新建）中新增：

```go
package service

import (
    "testing"
    "alpha-pulse/backend/models"
    "alpha-pulse/backend/repository"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func newTestAlertRecordRepo(t *testing.T) *repository.AlertRecordRepository {
    t.Helper()
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("open sqlite: %v", err)
    }
    if err := db.AutoMigrate(&models.AlertRecord{}); err != nil {
        t.Fatalf("migrate: %v", err)
    }
    return repository.NewAlertRecordRepository(db)
}

func TestGetAlertStatsWinRateExcludesPendingFromDenominator(t *testing.T) {
    repo := newTestAlertRecordRepo(t)

    // 写入测试数据：2 target_hit, 1 stop_hit, 1 pending, 1 expired
    records := []models.AlertRecord{
        {Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "target_hit", ActualRR: 2.0, EventTime: 1},
        {Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "target_hit", ActualRR: 1.5, EventTime: 2},
        {Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "stop_hit",   ActualRR: -1.0, EventTime: 3},
        {Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "pending",    EventTime: 4},
        {Symbol: "BTCUSDT", Kind: "setup_ready", Outcome: "expired",    EventTime: 5},
    }
    for _, r := range records {
        r := r
        if err := repo.GetDB().Create(&r).Error; err != nil {
            t.Fatalf("create record: %v", err)
        }
    }

    svc := &AlertService{repo: repo, symbols: []string{"BTCUSDT"}, now: func() interface{} { return nil }}
    stats := svc.GetAlertStats("BTCUSDT", 50)

    if stats.Total != 5 {
        t.Fatalf("expected total=5, got %d", stats.Total)
    }
    if stats.Pending != 1 {
        t.Fatalf("expected pending=1, got %d (pending must NOT be excluded from query)", stats.Pending)
    }
    // win_rate = 2 / (2+1) * 100 ≈ 66.67
    expected := float64(2) / float64(3) * 100
    if stats.WinRate < expected-0.1 || stats.WinRate > expected+0.1 {
        t.Fatalf("expected win_rate≈%.2f, got %.2f", expected, stats.WinRate)
    }
    // avg_rr = (2.0 + 1.5 + (-1.0)) / 3 ≈ 0.83
    expectedRR := (2.0 + 1.5 + (-1.0)) / 3
    if stats.AvgRR < expectedRR-0.01 || stats.AvgRR > expectedRR+0.01 {
        t.Fatalf("expected avg_rr≈%.2f, got %.2f", expectedRR, stats.AvgRR)
    }
}
```

> **注意**：`AlertService.now` 字段类型是 `func() time.Time`，直接构造时用 `now: time.Now`。
> `repo.GetDB()` 需要在 `AlertRecordRepository` 上暴露 `GetDB() *gorm.DB` 方法（如已存在则直接用，否则测试中直接用 sqlite db 变量）。

```bash
cd backend && go test ./internal/service/... -run TestGetAlertStats -v
```

期望：PASS。

- [ ] **Step 8: Commit**

```bash
git add frontend/components/alerts/WinRatePanel.tsx \
        frontend/components/alerts/WinRatePanel.test.tsx \
        frontend/app/review/page.tsx
git commit -m "feat(s3): add win rate statistics panel to review page"
```

---

## Task 13: 验收 — 全量测试 + 最终 Commit

- [ ] **Step 1: 运行后端全量测试**

```bash
cd backend && go test ./... -count=1
```

期望：全部通过。

- [ ] **Step 2: 运行前端全量测试**

```bash
cd frontend && npm test -- --run
```

期望：全部通过。

- [ ] **Step 3: 前端构建验证**

```bash
cd frontend && npm run build && npm run lint
```

期望：零报错，零 lint 警告。

- [ ] **Step 4: 回顾 spec 对照检查**

阅读 `docs/superpowers/specs/2026-03-25-trading-decision-loop-design.md` 的文件改动清单，逐一确认所有文件已实现。

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "feat: complete trading decision loop - fast & slow lanes"
```
