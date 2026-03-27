# Full Frontend Command Center Redesign Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 Alpha Pulse 前端一次性重构为统一的浅色专业指挥台，让公共壳层、关键页面骨架和共享模块都属于同一套命令中心视觉系统。

**Architecture:** 先建立共享布局原语与视觉 token，再在不改变现有业务数据流的前提下，重排 `dashboard`、`market`、`chart`、`review/signals`、`alerts`、`auto-trading` 的页面结构，并统一告警入口、状态模块和控制区语义。实现坚持 TDD：先写结构级失败测试，再做最小实现，最后补全样式与页面集成验证。

**Tech Stack:** Next.js 14 App Router, React 18, TypeScript, Tailwind utilities, global CSS, Ant Design 5, Vitest, Testing Library, Playwright

---

## Chunk 1: Design System and Shell Foundation

### Task 1: Lock the new command-center shell in tests

**Files:**
- Modify: `frontend/components/layout/ProAppShell.test.tsx`
- Modify: `frontend/components/layout/SignalStatusBadge.test.tsx`

- [ ] **Step 1: Write failing tests for the new shell semantics**

Add assertions for:
- shell root exposes the command-center structure
- left rail navigation remains visible
- bottom dock contains signal status and alert entry
- collapsed state is persisted
- login route still bypasses shell

- [ ] **Step 2: Run focused tests and verify RED**

Run: `cd frontend && npm test -- ProAppShell SignalStatusBadge`

Expected: FAIL because current structure and classes do not satisfy the new command-center assertions.

- [ ] **Step 3: Implement the minimal shell updates**

**Files:**
- Modify: `frontend/components/layout/ProAppShell.tsx`
- Modify: `frontend/components/layout/APLogo.tsx`
- Modify: `frontend/components/layout/SignalStatusBadge.tsx`
- Modify: `frontend/styles/globals.css`

Introduce:
- command-center shell structure and naming
- unified left rail / canvas / dock semantics
- light command-center surface tokens

- [ ] **Step 4: Run focused tests and verify GREEN**

Run: `cd frontend && npm test -- ProAppShell SignalStatusBadge`

Expected: PASS.

### Task 2: Add shared page layout primitives

**Files:**
- Create: `frontend/components/layout/CommandPage.tsx`
- Create: `frontend/components/layout/OverviewBand.tsx`
- Create: `frontend/components/layout/CommandPanel.tsx`
- Create: `frontend/components/layout/RailPanel.tsx`
- Modify: `frontend/components/layout/TradingWorkspaceHero.tsx`
- Modify: `frontend/components/layout/TradingWorkspaceHero.tsx`

- [ ] **Step 1: Write failing tests for shared layout wrappers**

Create or extend tests to assert:
- overview content renders in a dedicated band
- command panels expose consistent heading/body slots
- rail panels render optional status area

- [ ] **Step 2: Run focused tests and verify RED**

Run: `cd frontend && npm test -- TradingWorkspaceHero`

Expected: FAIL because the old hero layout is still page-specific and not built on the new shared primitives.

- [ ] **Step 3: Implement the minimal primitives and refactor hero**

Build lightweight wrapper components and move `TradingWorkspaceHero` onto them without changing store behavior.

- [ ] **Step 4: Run focused tests and verify GREEN**

Run: `cd frontend && npm test -- TradingWorkspaceHero`

Expected: PASS.

---

## Chunk 2: Dashboard as the Master Command Page

### Task 3: Rebuild dashboard overview and primary workspace structure

**Files:**
- Modify: `frontend/components/dashboard/DecisionHeader.test.tsx`
- Modify: `frontend/app/dashboard/page.tsx`
- Modify: `frontend/components/dashboard/DecisionHeader.tsx`
- Modify: `frontend/components/dashboard/ExecutionPanel.tsx`
- Modify: `frontend/components/dashboard/EvidenceRail.tsx`
- Modify: `frontend/styles/globals.css`

- [ ] **Step 1: Write failing tests for the new dashboard information hierarchy**

Add assertions for:
- overview band exists and contains direction summary
- chart region is the primary workspace area
- execution tools render in a dedicated side rail
- evidence content is grouped in a bottom strip

- [ ] **Step 2: Run focused tests and verify RED**

Run: `cd frontend && npm test -- DecisionHeader ExecutionPanel EvidenceRail`

Expected: FAIL because current dashboard pieces are not grouped into the new hierarchy.

- [ ] **Step 3: Implement the minimal dashboard restructuring**

Keep business data and interactions intact while rewrapping the dashboard with the new shared layout primitives.

- [ ] **Step 4: Run focused tests and verify GREEN**

Run: `cd frontend && npm test -- DecisionHeader ExecutionPanel EvidenceRail`

Expected: PASS.

### Task 4: Bring the position workflow into the dashboard side rail

**Files:**
- Modify: `frontend/components/trading/PositionCalculator.test.tsx`
- Modify: `frontend/components/trading/PositionCalculator.tsx`
- Modify: `frontend/app/dashboard/page.tsx`
- Modify: `frontend/styles/globals.css`

- [ ] **Step 1: Write failing tests for side-rail calculator presentation**

Assert the calculator still exposes core inputs/outputs while participating in the side-rail module structure.

- [ ] **Step 2: Run focused tests and verify RED**

Run: `cd frontend && npm test -- PositionCalculator`

Expected: FAIL because the calculator is still styled as a standalone card.

- [ ] **Step 3: Implement the minimal calculator restyling**

Apply the new action-panel semantics without changing calculation logic.

- [ ] **Step 4: Run focused tests and verify GREEN**

Run: `cd frontend && npm test -- PositionCalculator`

Expected: PASS.

---

## Chunk 3: Market and Chart Page Alignment

### Task 5: Rebuild the market page as an intelligence desk

**Files:**
- Modify: `frontend/app/market/page.tsx`
- Modify: `frontend/components/market/FuturesWatchlist.tsx`
- Modify: `frontend/components/market/MarketOverviewBoard.tsx`
- Modify: `frontend/components/market/MarketLevelsBoard.tsx`
- Modify: `frontend/components/market/SignalTape.tsx`
- Modify: `frontend/components/market/MicrostructureTimeline.tsx`
- Modify: `frontend/components/orderflow/OrderFlowPanel.tsx`
- Modify: `frontend/components/liquidity/LiquidityPanel.tsx`
- Modify: `frontend/components/market/FuturesWatchlist.test.tsx`
- Modify: `frontend/components/market/MicrostructureTimeline.test.tsx`
- Modify: `frontend/components/orderflow/OrderFlowPanel.test.tsx`
- Modify: `frontend/components/liquidity/LiquidityPanel.test.tsx`
- Modify: `frontend/styles/globals.css`

- [ ] **Step 1: Write failing tests for the market page grouping**

Cover:
- radar-style watchlist area
- paired central workspaces
- lower evidence/diagnostic zone

- [ ] **Step 2: Run focused tests and verify RED**

Run: `cd frontend && npm test -- FuturesWatchlist MicrostructureTimeline OrderFlowPanel LiquidityPanel`

Expected: FAIL because the old grid stacking does not match the new structure.

- [ ] **Step 3: Implement the minimal market restructuring**

Move existing modules into the intelligence-desk layout while keeping all live data behavior.

- [ ] **Step 4: Run focused tests and verify GREEN**

Run: `cd frontend && npm test -- FuturesWatchlist MicrostructureTimeline OrderFlowPanel LiquidityPanel`

Expected: PASS.

### Task 6: Rebuild the chart page as a deep-work canvas

**Files:**
- Modify: `frontend/app/chart/page.tsx`
- Modify: `frontend/components/chart/ChartInsightRail.tsx`
- Modify: `frontend/components/chart/KlineChart.tsx`
- Modify: `frontend/components/chart/KlineChart.test.tsx`
- Modify: `frontend/styles/globals.css`

- [ ] **Step 1: Write failing tests for chart-first composition**

Assert:
- chart region remains primary and dominant
- side insight rail stays visible
- top controls are present but visually secondary

- [ ] **Step 2: Run focused tests and verify RED**

Run: `cd frontend && npm test -- KlineChart`

Expected: FAIL because the current page composition still treats the hero as the dominant block.

- [ ] **Step 3: Implement the minimal chart layout revision**

Refactor page structure and supporting styles to emphasize the chart workspace.

- [ ] **Step 4: Run focused tests and verify GREEN**

Run: `cd frontend && npm test -- KlineChart`

Expected: PASS.

---

## Chunk 4: Review, Alerts, and Auto-Trading Completion

### Task 7: Align the review workspace with the new command-page pattern

**Files:**
- Modify: `frontend/components/review/ReviewWorkspace.tsx`
- Modify: `frontend/components/analysis/AIAnalysisPanel.tsx`
- Modify: `frontend/components/signal/SignalCard.tsx`
- Modify: `frontend/components/alerts/AlertHistoryBoard.tsx`
- Modify: `frontend/components/alerts/WinRatePanel.tsx`
- Modify: `frontend/components/analysis/AIAnalysisPanel.test.tsx`
- Modify: `frontend/components/signal/SignalCard.test.tsx`
- Modify: `frontend/components/alerts/AlertHistoryBoard.test.tsx`
- Modify: `frontend/components/alerts/WinRatePanel.test.tsx`
- Modify: `frontend/styles/globals.css`

- [ ] **Step 1: Write failing tests for review information flow**

Assert the page shows:
- overview summary
- historical review strip
- side-by-side real-time signal and AI context areas

- [ ] **Step 2: Run focused tests and verify RED**

Run: `cd frontend && npm test -- AIAnalysisPanel SignalCard AlertHistoryBoard WinRatePanel`

Expected: FAIL because the current layout is still card-stacked.

- [ ] **Step 3: Implement the minimal review restructuring**

Adopt the shared command-page pattern while preserving existing content and logic.

- [ ] **Step 4: Run focused tests and verify GREEN**

Run: `cd frontend && npm test -- AIAnalysisPanel SignalCard AlertHistoryBoard WinRatePanel`

Expected: PASS.

### Task 8: Promote alerts into a first-class page and upgrade auto-trading placeholder

**Files:**
- Create: `frontend/app/alerts/page.tsx`
- Modify: `frontend/components/alerts/AlertCenter.tsx`
- Modify: `frontend/app/auto-trading/page.tsx`
- Modify: `frontend/tests/e2e/dashboard.spec.ts`
- Modify: `frontend/tests/e2e/market.spec.ts`
- Modify: `frontend/tests/e2e/review.spec.ts`
- Create: `frontend/tests/e2e/alerts.spec.ts`
- Modify: `frontend/components/layout/ProAppShell.tsx`

- [ ] **Step 1: Write failing tests for alerts navigation and page presence**

Add assertions for:
- alerts route exists and is reachable from the shell
- auto-trading page uses the shared command-page structure instead of plain inline styles

- [ ] **Step 2: Run focused tests and verify RED**

Run: `cd frontend && npm test -- ProAppShell`

Expected: FAIL because the alerts page route and updated shell semantics do not exist yet.

- [ ] **Step 3: Implement the minimal alerts/auto-trading pages**

Create a first-class alerts page using existing alert components, keep the drawer as a quick-access entry, and replace auto-trading inline placeholder with structured modules.

- [ ] **Step 4: Run focused tests and verify GREEN**

Run: `cd frontend && npm test -- ProAppShell`

Expected: PASS.

---

## Chunk 5: Full Verification

### Task 9: Verify the full redesign end to end

**Files:**
- Modify: `frontend/tests/e2e/dashboard.states.spec.ts`
- Modify: `frontend/tests/e2e/auth.spec.ts` (only if selectors require updates)
- Modify: `frontend/styles/globals.css`

- [ ] **Step 1: Run the targeted unit test suite**

Run: `cd frontend && npm test -- ProAppShell SignalStatusBadge TradingWorkspaceHero DecisionHeader ExecutionPanel EvidenceRail PositionCalculator FuturesWatchlist MicrostructureTimeline OrderFlowPanel LiquidityPanel KlineChart AIAnalysisPanel SignalCard AlertHistoryBoard WinRatePanel`

Expected: PASS.

- [ ] **Step 2: Run the full frontend unit suite**

Run: `cd frontend && npm test`

Expected: PASS.

- [ ] **Step 3: Run lint**

Run: `cd frontend && npm run lint`

Expected: PASS.

- [ ] **Step 4: Run production build**

Run: `cd frontend && npm run build`

Expected: PASS.

- [ ] **Step 5: Run critical E2E coverage**

Run: `cd frontend && npm run test:e2e -- dashboard.spec.ts market.spec.ts review.spec.ts alerts.spec.ts`

Expected: Core pages load and key command-center regions remain accessible. If environment limitations block E2E, document the exact blocker in the handoff.
