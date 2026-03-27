# Auto Trading Live Execution Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build real Binance Futures auto trading with limit-entry execution, runtime trade settings, order lifecycle tracking, and a visual auto-trading control page.

**Architecture:** Keep alert generation intact, then add a direct execution path from `setup_ready` into an `AutoTradeCoordinator`, a `TradeExecutorService`, and a separate `TradeRuntime` for entry watching and position sync. Persist runtime settings and order state in MySQL, expose them via authenticated trade APIs, and drive the frontend control page from those APIs.

**Tech Stack:** Go, Gin, GORM, go-binance futures SDK, MySQL, Next.js App Router, React, TypeScript, Ant Design, Vitest, Testing Library

---

## Chunk 1: Spec, Models, and Configuration Contracts

### Task 1: Add the new trade models and config contracts

**Files:**
- Create: `backend/models/trade_setting.go`
- Create: `backend/models/trade_order.go`
- Modify: `backend/models/migrate.go`
- Modify: `backend/config/config.go`
- Modify: `backend/config/config_test.go`

- [ ] **Step 1: Write the failing tests**

Add config tests that assert:
- `TRADE_ENABLED` defaults to `false`
- `TRADE_AUTO_EXECUTE` defaults to `false`
- `TRADE_ALLOWED_SYMBOLS` defaults to configured market symbols
- trade watcher / sync intervals get safe defaults if absent

- [ ] **Step 2: Run the focused config test to verify it fails**

Run: `cd backend && go test ./config -run TestLoadTrade`
Expected: FAIL because trade config fields do not exist yet.

- [ ] **Step 3: Implement the minimal config and model changes**

Add:
- trade fields on `config.Config`
- env loading helpers for trade settings
- `TradeSetting` and `TradeOrder` models
- migration registration

- [ ] **Step 4: Run the focused config test**

Run: `cd backend && go test ./config -run TestLoadTrade`
Expected: PASS

## Chunk 2: Repositories and Trade Settings Service

### Task 2: Add persistence and validation for runtime trade settings

**Files:**
- Create: `backend/repository/trade_setting_repo.go`
- Create: `backend/repository/trade_order_repo.go`
- Create: `backend/repository/trade_setting_repo_test.go`
- Create: `backend/repository/trade_order_repo_test.go`
- Create: `backend/internal/service/trade_settings.go`
- Create: `backend/internal/service/trade_settings_test.go`

- [ ] **Step 1: Write the failing repository and service tests**

Cover:
- singleton trade settings save / load
- order creation and open-order lookup
- allowed symbol filtering
- risk / timeout / max position validation

- [ ] **Step 2: Run the focused backend tests to verify they fail**

Run:
- `cd backend && go test ./repository -run 'TestTrade(Setting|Order)'`
- `cd backend && go test ./internal/service -run TestTradeSettings`

Expected: FAIL because the repository and service do not exist yet.

- [ ] **Step 3: Implement the minimal repositories and settings service**

Add:
- setting projection / hydration helpers
- settings sanitization
- order query helpers for `pending_fill`, `open`, history, and symbol filters

- [ ] **Step 4: Run the focused backend tests**

Run:
- `cd backend && go test ./repository -run 'TestTrade(Setting|Order)'`
- `cd backend && go test ./internal/service -run TestTradeSettings`

Expected: PASS

## Chunk 3: Binance Futures Trading Primitives

### Task 3: Extend the Binance client with real trading methods

**Files:**
- Modify: `backend/pkg/binance/client.go`
- Create: `backend/pkg/binance/client_trade_test.go`

- [ ] **Step 1: Write the failing client tests**

Cover:
- parsing futures balance
- parsing symbol filters / precision
- building a limit order request
- building stop-loss / take-profit requests
- parsing order status and positions

- [ ] **Step 2: Run the focused client tests to verify they fail**

Run: `cd backend && go test ./pkg/binance -run TestFuturesTrade`
Expected: FAIL because the trade methods do not exist yet.

- [ ] **Step 3: Implement the minimal trade methods**

Add:
- balance / leverage / exchange info helpers
- limit order placement
- order status lookup
- cancel order
- protective order placement
- close-position market order
- position fetch

- [ ] **Step 4: Run the focused client tests**

Run: `cd backend && go test ./pkg/binance -run TestFuturesTrade`
Expected: PASS

## Chunk 4: Coordinator, Executor, and Runtime

### Task 4: Add the trade execution state machine

**Files:**
- Create: `backend/internal/service/auto_trade_coordinator.go`
- Create: `backend/internal/service/trade_executor.go`
- Create: `backend/internal/service/trade_runtime.go`
- Create: `backend/internal/service/auto_trade_coordinator_test.go`
- Create: `backend/internal/service/trade_executor_test.go`
- Create: `backend/internal/service/trade_runtime_test.go`
- Modify: `backend/internal/service/alert_service.go`
- Modify: `backend/internal/scheduler/jobs.go`

- [ ] **Step 1: Write the failing service tests**

Cover:
- `setup_ready` only triggers execution when both env and runtime settings allow it
- duplicate open position is rejected
- limit order creates `pending_fill`
- watcher promotes `pending_fill -> open` after fill and adds protective orders
- watcher promotes `pending_fill -> expired` after timeout
- protective order failure triggers forced close and `failed`
- position sync creates `manual` orders and closes local orders when Binance position disappears

- [ ] **Step 2: Run the focused service tests to verify they fail**

Run:
- `cd backend && go test ./internal/service -run TestAutoTradeCoordinator`
- `cd backend && go test ./internal/service -run TestTradeExecutor`
- `cd backend && go test ./internal/service -run TestTradeRuntime`

Expected: FAIL because the services and hooks do not exist yet.

- [ ] **Step 3: Implement the minimal execution services**

Add:
- coordinator guards
- quantity calculation
- limit-entry execution
- runtime watch loops as testable single-run methods
- alert hook for `setup_ready`
- runtime invocation point alongside existing scheduler startup

- [ ] **Step 4: Run the focused service tests**

Run:
- `cd backend && go test ./internal/service -run TestAutoTradeCoordinator`
- `cd backend && go test ./internal/service -run TestTradeExecutor`
- `cd backend && go test ./internal/service -run TestTradeRuntime`

Expected: PASS

## Chunk 5: Trade APIs and Server Wiring

### Task 5: Expose trade settings, runtime, list, and close endpoints

**Files:**
- Create: `backend/internal/handler/trade_handler.go`
- Create: `backend/router/trade_test.go`
- Modify: `backend/router/router.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Write the failing router tests**

Cover:
- authenticated `GET /api/trade-settings`
- authenticated `PUT /api/trade-settings`
- authenticated `GET /api/trades`
- authenticated `GET /api/trades/runtime`
- authenticated `POST /api/trades/:id/close`
- `TRADE_ENABLED=false` blocks write actions with `403`

- [ ] **Step 2: Run the focused router test to verify it fails**

Run: `cd backend && go test ./router -run TestTrade`
Expected: FAIL because the trade routes do not exist yet.

- [ ] **Step 3: Implement the minimal handler and wiring**

Add:
- request / response DTOs
- handler methods
- router registration
- server wiring for repos, services, handlers, and runtime startup

- [ ] **Step 4: Run the focused router test**

Run: `cd backend && go test ./router -run TestTrade`
Expected: PASS

## Chunk 6: Frontend Types and API Client

### Task 6: Add frontend trade data contracts

**Files:**
- Create: `frontend/types/trade.ts`
- Modify: `frontend/services/apiClient.ts`
- Create: `frontend/services/apiClient.trade.test.ts`

- [ ] **Step 1: Write the failing frontend contract tests**

Cover:
- trade settings fetch / update wrappers
- trade list and runtime wrappers
- close wrapper

- [ ] **Step 2: Run the focused frontend test to verify it fails**

Run: `cd frontend && npm test -- services/apiClient.trade.test.ts`
Expected: FAIL because trade API wrappers do not exist yet.

- [ ] **Step 3: Implement the minimal frontend contracts**

Add:
- trade types
- `tradeApi` methods

- [ ] **Step 4: Run the focused frontend test**

Run: `cd frontend && npm test -- services/apiClient.trade.test.ts`
Expected: PASS

## Chunk 7: Auto Trading Control Page and Order Panel

### Task 7: Replace the placeholder auto-trading page with the real control surface

**Files:**
- Create: `frontend/components/trading/TradeSettingsPanel.tsx`
- Create: `frontend/components/trading/TradeRuntimeBand.tsx`
- Create: `frontend/components/trading/TradeOrderPanel.tsx`
- Create: `frontend/components/trading/TradeOrderPanel.test.tsx`
- Modify: `frontend/app/auto-trading/page.tsx`
- Modify: `frontend/app/auto-trading/page.test.tsx`
- Modify: `frontend/styles/globals.css`

- [ ] **Step 1: Write the failing page and panel tests**

Cover:
- runtime status band renders
- configuration controls render and save
- order sections render `pending_fill`, `open`, `failed`, and `closed`
- pending and open rows surface the right actions

- [ ] **Step 2: Run the focused frontend tests to verify they fail**

Run:
- `cd frontend && npm test -- app/auto-trading/page.test.tsx`
- `cd frontend && npm test -- components/trading/TradeOrderPanel.test.tsx`

Expected: FAIL because the real control page does not exist yet.

- [ ] **Step 3: Implement the minimal auto-trading control page**

Add:
- data loading and save flow
- status band
- settings panel
- order panel
- command-center styling for the new surface

- [ ] **Step 4: Run the focused frontend tests**

Run:
- `cd frontend && npm test -- app/auto-trading/page.test.tsx`
- `cd frontend && npm test -- components/trading/TradeOrderPanel.test.tsx`

Expected: PASS

## Chunk 8: Review Workspace Trade Visibility

### Task 8: Surface the trade panel in review mode

**Files:**
- Modify: `frontend/components/review/ReviewWorkspace.tsx`
- Modify: `frontend/components/review/ReviewWorkspace.test.tsx`

- [ ] **Step 1: Write the failing review workspace test**

Add an assertion that the review workspace renders the trade order panel region.

- [ ] **Step 2: Run the focused review test to verify it fails**

Run: `cd frontend && npm test -- components/review/ReviewWorkspace.test.tsx`
Expected: FAIL because the trade panel is not mounted yet.

- [ ] **Step 3: Implement the minimal review composition**

Mount the order panel in the review workspace without disturbing existing signal / alert context.

- [ ] **Step 4: Run the focused review test**

Run: `cd frontend && npm test -- components/review/ReviewWorkspace.test.tsx`
Expected: PASS

## Chunk 9: Final Verification

### Task 9: Run the full verification set

**Files:**
- Modify: none expected beyond implementation files above

- [ ] **Step 1: Run all backend tests**

Run: `cd backend && go test ./...`
Expected: PASS

- [ ] **Step 2: Run all frontend tests**

Run: `cd frontend && npm test`
Expected: PASS

- [ ] **Step 3: Run lint and build**

Run:
- `cd frontend && npm run lint`
- `cd frontend && npm run build`

Expected: PASS
