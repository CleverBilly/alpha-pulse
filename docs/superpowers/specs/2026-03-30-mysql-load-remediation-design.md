# MySQL Load Remediation Design

## Problem

The production host is a 2-core machine where `mysqld` is saturating CPU during Alpha Pulse analysis work. Investigation showed two expensive query patterns dominating the load:

1. `order_book_snapshots` lookups for "latest snapshot before time" using `symbol + event_time` filters with descending sort.
2. `agg_trades` lookups for "recent trades by symbol" using `symbol` filters with descending sort by `trade_time`.

The current indexes do not match those access patterns, so MySQL falls back to large scans plus filesort. On top of that, the liquidity series builder queries one historical order book snapshot per chart point, multiplying the number of slow queries under load.

## Goals

- Reduce production MySQL CPU load immediately.
- Eliminate the known filesort-heavy query patterns.
- Prevent fresh environments from missing the required indexes.
- Reduce query fan-out during liquidity series generation.

## Non-Goals

- Redesign the market snapshot API shape.
- Change alerting or trading behavior.
- Rework the scheduler architecture beyond what is needed for this hotspot.

## Approach

### 1. Immediate production mitigation

Apply composite indexes directly in production for the two hot tables:

- `order_book_snapshots(symbol, event_time, last_update_id, id)`
- `agg_trades(symbol, trade_time, agg_trade_id)`

This gives MySQL an index path aligned to the actual `WHERE + ORDER BY + LIMIT` access pattern and should reduce scan volume and filesort pressure immediately.

### 2. Persist the indexing fix in code

Update the GORM model definitions so new databases and auto-migrated environments create the same composite indexes. This avoids a repeat incident after redeploys or fresh setup.

### 3. Remove N-per-point order book lookups

Refactor liquidity series building so it preloads a bounded set of recent order book snapshots once, then resolves each chart point against the in-memory snapshot window. The series builder should no longer call the repository once per chart point.

## Testing Strategy

- Add repository/model regression coverage for the required composite indexes.
- Add service-level regression coverage proving liquidity series resolution does not perform one repository query per point.
- Run `go test ./...` in `backend/`.
- Re-check production query plans and active process list after the fix.

## Expected Outcome

- Production MySQL CPU falls materially from the current steady-state overload.
- Slow query log no longer shows repeated 20s+ `order_book_snapshots` lookups with large `Rows_examined`.
- Liquidity series generation becomes bounded by one preload query rather than dozens of historical lookups.
