# MySQL Load Remediation Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reduce Alpha Pulse production MySQL load by fixing missing composite indexes and removing per-point historical order book queries from liquidity series generation.

**Architecture:** First mitigate production by adding the composite indexes directly on the live database. Then encode the same index definitions in the GORM models and refactor liquidity series building to preload a snapshot window once and resolve point lookups in memory. Verification covers query plans, backend tests, and production process list checks.

**Tech Stack:** Go, GORM, MySQL, Bash, PM2

---

## Chunk 1: Production Mitigation

### Task 1: Capture pre-change production evidence

**Files:**
- Modify: none
- Test: none

- [ ] **Step 1: Record current production query pressure**

Run:

```bash
ssh root@43.156.163.236
mysql -ualpha-pulse -p'***' -D 'alpha-pulse' -e "SHOW FULL PROCESSLIST;"
mysql -ualpha-pulse -p'***' -D 'alpha-pulse' -e "EXPLAIN SELECT * FROM order_book_snapshots WHERE symbol='BTCUSDT' AND event_time <= 1774828800000 ORDER BY event_time DESC, last_update_id DESC, id LIMIT 1; EXPLAIN SELECT * FROM agg_trades WHERE symbol='BTCUSDT' ORDER BY trade_time DESC, agg_trade_id DESC LIMIT 250;"
```

Expected: evidence of executing queries and `Using filesort`.

- [ ] **Step 2: Record machine load**

Run:

```bash
ssh root@43.156.163.236
uptime
top -b -n 1 | head -n 20
```

Expected: high load with `mysqld` dominating CPU.

### Task 2: Apply the hot indexes in production

**Files:**
- Modify: production MySQL schema only
- Test: query plan re-check

- [ ] **Step 1: Add the order book lookup index**

Run:

```sql
ALTER TABLE order_book_snapshots
ADD INDEX idx_order_book_symbol_event_update_lookup (symbol, event_time, last_update_id, id);
```

- [ ] **Step 2: Add the aggregate trade lookup index**

Run:

```sql
ALTER TABLE agg_trades
ADD INDEX idx_agg_trade_symbol_time_lookup (symbol, trade_time, agg_trade_id);
```

- [ ] **Step 3: Re-run EXPLAIN and process list**

Run the same pre-change commands.

Expected: no `Using filesort` on the hot path or materially fewer rows examined.

## Chunk 2: Code Fixes

### Task 3: Write failing backend tests

**Files:**
- Create: `backend/internal/service/analysis_series_test.go`
- Modify: `backend/models/order_book_snapshot.go`
- Modify: `backend/models/agg_trade.go`
- Test: `backend/internal/service/analysis_series_test.go`

- [ ] **Step 1: Add a failing test for bulk snapshot resolution**

Write a test that builds liquidity series over multiple points and asserts the repository is queried once for a bounded snapshot window, not once per point.

- [ ] **Step 2: Run the focused test and confirm failure**

Run:

```bash
cd backend
go test ./internal/service -run TestBuildLiquiditySeriesPreloadsSnapshotsOnce
```

Expected: FAIL because the current implementation calls point-by-point lookup logic.

### Task 4: Implement the minimal service refactor

**Files:**
- Modify: `backend/repository/order_book_snapshot_repo.go`
- Modify: `backend/internal/service/analysis_series.go`
- Test: `backend/internal/service/analysis_series_test.go`

- [ ] **Step 1: Add a repository method to load recent snapshots up to a max event time**

- [ ] **Step 2: Refactor liquidity series building to preload snapshots once**

- [ ] **Step 3: Resolve each point against the preloaded slice in memory**

- [ ] **Step 4: Re-run the focused service test**

Run:

```bash
cd backend
go test ./internal/service -run TestBuildLiquiditySeriesPreloadsSnapshotsOnce
```

Expected: PASS.

### Task 5: Persist composite indexes in the models

**Files:**
- Modify: `backend/models/order_book_snapshot.go`
- Modify: `backend/models/agg_trade.go`
- Test: model tag assertions in `backend/internal/service/analysis_series_test.go` or a focused repository/model test

- [ ] **Step 1: Add composite lookup index tags to both models**

- [ ] **Step 2: Add or extend a test that asserts the model tags contain the lookup indexes**

- [ ] **Step 3: Run the focused test**

Run:

```bash
cd backend
go test ./internal/service -run 'TestBuildLiquiditySeriesPreloadsSnapshotsOnce|TestHotLookupIndexesAreDeclared'
```

Expected: PASS.

## Chunk 3: Full Verification

### Task 6: Run backend verification

**Files:**
- Modify: none
- Test: backend suite

- [ ] **Step 1: Run all backend tests**

Run:

```bash
cd backend
go test ./...
```

Expected: PASS.

### Task 7: Re-verify production after deploy

**Files:**
- Modify: none
- Test: production diagnostics

- [ ] **Step 1: Deploy the backend update**

Run:

```bash
cd /www/wwwroot/alpha-pulse
git pull origin main
bash deploy.sh
```

- [ ] **Step 2: Check production load and query health**

Run:

```bash
ssh root@43.156.163.236
uptime
top -b -n 1 | head -n 20
mysql -ualpha-pulse -p'***' -D 'alpha-pulse' -e "SHOW FULL PROCESSLIST;"
tail -n 40 /www/server/data/mysql-slow.log
```

Expected: materially lower MySQL CPU pressure and no fresh bursts of the original slow queries.
