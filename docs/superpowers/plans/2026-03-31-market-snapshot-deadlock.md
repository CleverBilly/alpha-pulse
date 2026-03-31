# Market Snapshot Deadlock Mitigation Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stop transient MySQL deadlocks from surfacing as 500 responses on `/api/market-snapshot` while reducing duplicate concurrent rebuild work.

**Architecture:** Keep the current market snapshot assembly pipeline, but add two protections around it: coalesce concurrent builds for the same `(symbol, interval, limit)` key inside the process, and route all write-heavy persistence hotspots through a shared retry helper for retryable lock errors. Also normalize projected event ordering before batch upserts so overlapping writes acquire row locks in a consistent order.

**Tech Stack:** Go 1.22, Gin, GORM, MySQL, existing backend service/repository test suite.

---

## Chunk 1: Reproduce the concurrency contract in tests

### Task 1: Add tests for retryable write errors

**Files:**
- Modify: `backend/internal/service/persist_support.go`
- Test: `backend/internal/service/persist_support_test.go`

- [ ] **Step 1: Write the failing tests**

Add tests that prove:
- deadlock-like errors are classified as retryable
- lock wait timeout errors are classified as retryable
- non-lock errors are not retried

- [ ] **Step 2: Run the targeted test to verify it fails**

Run: `cd backend && go test ./internal/service -run 'Test(RetryableDBWriteError|WithRetryableWrite)' -count=1`
Expected: FAIL because the helper does not exist yet.

- [ ] **Step 3: Write the minimal helper implementation**

Add a shared helper that retries a closure for retryable DB write errors with bounded exponential backoff.

- [ ] **Step 4: Run the targeted test to verify it passes**

Run: `cd backend && go test ./internal/service -run 'Test(RetryableDBWriteError|WithRetryableWrite)' -count=1`
Expected: PASS

### Task 2: Add tests for stable event persistence ordering

**Files:**
- Modify: `backend/internal/service/large_trade_support.go`
- Modify: `backend/internal/service/microstructure_support.go`
- Test: `backend/internal/service/persistence_projection_test.go`

- [ ] **Step 1: Write the failing tests**

Add tests that verify projected large-trade and microstructure event slices are emitted in deterministic ascending order for overlapping writes.

- [ ] **Step 2: Run the targeted test to verify it fails**

Run: `cd backend && go test ./internal/service -run 'TestProject(LargeTrade|Microstructure)Events' -count=1`
Expected: FAIL because projection currently preserves caller order.

- [ ] **Step 3: Write the minimal implementation**

Sort projected events by their unique conflict keys before returning them.

- [ ] **Step 4: Run the targeted test to verify it passes**

Run: `cd backend && go test ./internal/service -run 'TestProject(LargeTrade|Microstructure)Events' -count=1`
Expected: PASS

## Chunk 2: Coalesce duplicate snapshot builds

### Task 3: Add tests for same-key concurrent snapshot dedupe

**Files:**
- Modify: `backend/internal/service/signal_service.go`
- Test: `backend/internal/service/signal_service_cache_test.go`

- [ ] **Step 1: Write the failing test**

Add a test that issues concurrent `GetMarketSnapshot` calls for the same key with cache disabled and asserts the expensive build path only executes once.

- [ ] **Step 2: Run the targeted test to verify it fails**

Run: `cd backend && go test ./internal/service -run TestGetMarketSnapshotCoalescesConcurrentBuilds -count=1`
Expected: FAIL because concurrent callers currently rebuild independently.

- [ ] **Step 3: Write the minimal implementation**

Add a per-key in-process build coordinator to `SignalService` and route `GetMarketSnapshotWithRefresh` through it.

- [ ] **Step 4: Run the targeted test to verify it passes**

Run: `cd backend && go test ./internal/service -run TestGetMarketSnapshotCoalescesConcurrentBuilds -count=1`
Expected: PASS

## Chunk 3: Wire retries into write hotspots and verify behavior

### Task 4: Apply retry helper to snapshot, kline, and agg-trade persistence

**Files:**
- Modify: `backend/internal/service/persist_support.go`
- Modify: `backend/internal/service/indicator_support.go`
- Modify: `backend/internal/service/trade_support.go`
- Modify: `backend/internal/service/feature_snapshot_support.go`

- [ ] **Step 1: Write the failing integration-level assertions if new coverage is needed**

Extend tests only where required to prove the helper is used in the relevant path.

- [ ] **Step 2: Run the targeted tests to verify they fail**

Run: `cd backend && go test ./internal/service -run 'Test(RetryableDBWriteError|WithRetryableWrite|GetMarketSnapshotCoalescesConcurrentBuilds|Project(LargeTrade|Microstructure)Events)' -count=1`
Expected: FAIL before the wiring is complete.

- [ ] **Step 3: Write the minimal implementation**

Use the retry helper around:
- `persistSnapshotResults`
- `repo.CreateBatch(fetched)` in kline loading
- `repo.CreateBatch(fetched)` in agg-trade loading
- feature snapshot archival upsert

- [ ] **Step 4: Run targeted tests to verify they pass**

Run: `cd backend && go test ./internal/service -run 'Test(RetryableDBWriteError|WithRetryableWrite|GetMarketSnapshotCoalescesConcurrentBuilds|Project(LargeTrade|Microstructure)Events)' -count=1`
Expected: PASS

### Task 5: Run broader backend verification

**Files:**
- Modify: `backend/internal/service/*.go`
- Modify: `backend/internal/service/*_test.go`

- [ ] **Step 1: Run focused backend package tests**

Run: `cd backend && go test ./internal/service ./router -count=1`
Expected: PASS

- [ ] **Step 2: Run the full backend suite if focused tests are green**

Run: `cd backend && go test ./...`
Expected: PASS

- [ ] **Step 3: Review diff for scope control**

Run: `git diff -- backend/internal/service backend/router docs/superpowers/plans/2026-03-31-market-snapshot-deadlock.md`
Expected: only deadlock mitigation and related tests are changed.
