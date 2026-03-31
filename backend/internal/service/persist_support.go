package service

import (
	"errors"
	"strings"
	"time"

	"alpha-pulse/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// persistSnapshotResults 将 buildMarketSnapshot 产出的核心结果在 **单个事务** 内写库。
//
// 死锁根因（MySQL Error 1213）：
//   - 旧实现中 indicator / orderflow / structure / liquidity / signal 各自是独立事务。
//   - 并发请求以不同顺序持有多个表的行级锁，MySQL 检测到环形等待后触发死锁。
//
// 修复方法：
//  1. 将全部写操作合并为单个事务，缩短锁持有时间窗口。
//  2. 所有并发路径按相同顺序加锁（indicators → orderflow → large_trades →
//     microstructure_events → structure → liquidity → signals），消除环形等待。
//  3. 死锁重试：最多 3 次，指数退避（50ms / 100ms / 200ms）。
//
// 成功后 ID 字段由 GORM autoIncrement 回填（indicator / orderflow / structure /
// liquidity / signal）。feature_snapshots 由调用方在 snapshot 组装后单独 upsert。
func persistSnapshotResults(
	db *gorm.DB,
	indicatorResult *models.Indicator,
	orderFlowResult *models.OrderFlow,
	structureResult *models.Structure,
	liquidityResult *models.Liquidity,
	signalResult *models.Signal,
) error {
	const maxDeadlockRetries = 3

	for attempt := 0; ; attempt++ {
		err := runPersistTx(db, indicatorResult, orderFlowResult, structureResult, liquidityResult, signalResult)
		if err == nil {
			return nil
		}
		if isDeadlock(err) && attempt < maxDeadlockRetries {
			// 指数退避：50ms → 100ms → 200ms
			time.Sleep(time.Duration(50<<attempt) * time.Millisecond)
			// 重置 autoIncrement ID，避免 GORM 把失败事务中回填的 ID 当成已有记录
			indicatorResult.ID = 0
			orderFlowResult.ID = 0
			structureResult.ID = 0
			liquidityResult.ID = 0
			signalResult.ID = 0
			continue
		}
		return err
	}
}

// runPersistTx 在单个数据库事务内按固定顺序写入所有核心快照结果。
func runPersistTx(
	db *gorm.DB,
	indicatorResult *models.Indicator,
	orderFlowResult *models.OrderFlow,
	structureResult *models.Structure,
	liquidityResult *models.Liquidity,
	signalResult *models.Signal,
) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. indicators
		if err := tx.Create(indicatorResult).Error; err != nil {
			return err
		}

		// 2. orderflow — autoIncrement 回填 orderFlowResult.ID，后续从表依赖此 ID
		if err := tx.Create(orderFlowResult).Error; err != nil {
			return err
		}

		// 3. large_trade_events（ON DUPLICATE KEY UPDATE，幂等）
		if largeEvents := projectLargeTradeEvents(*orderFlowResult); len(largeEvents) > 0 {
			if err := persistLargeTradeEventsTx(tx, largeEvents); err != nil {
				return err
			}
		}

		// 4. microstructure_events（INSERT IGNORE on unique index，幂等）
		if microEvents := projectMicrostructureEvents(*orderFlowResult); len(microEvents) > 0 {
			if err := persistMicrostructureEventsTx(tx, microEvents); err != nil {
				return err
			}
		}

		// 5. structure
		if err := tx.Create(structureResult).Error; err != nil {
			return err
		}

		// 6. liquidity
		if err := tx.Create(liquidityResult).Error; err != nil {
			return err
		}

		// 7. signals
		if err := tx.Create(signalResult).Error; err != nil {
			return err
		}

		return nil
	})
}

// persistLargeTradeEventsTx 在事务内批量写入大单事件（ON DUPLICATE KEY UPDATE）。
// 与 LargeTradeEventRepository.CreateBatch 语义相同，但使用传入的 tx。
func persistLargeTradeEventsTx(tx *gorm.DB, events []models.LargeTradeEvent) error {
	if len(events) == 0 {
		return nil
	}
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "symbol"},
			{Name: "agg_trade_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"orderflow_id",
			"interval_type",
			"open_time",
			"side",
			"price",
			"quantity",
			"notional",
			"trade_time",
			"created_at",
		}),
	}).Create(&events).Error
}

// persistMicrostructureEventsTx 在事务内批量写入微结构事件（INSERT IGNORE on unique index）。
// 与 MicrostructureEventRepository.CreateBatch 语义相同，但使用传入的 tx。
func persistMicrostructureEventsTx(tx *gorm.DB, events []models.MicrostructureEvent) error {
	if len(events) == 0 {
		return nil
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&events).Error
}

// isDeadlock 判断 err 是否为 MySQL 死锁错误（Error 1213）。
func isDeadlock(err error) bool {
	if err == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "1213") ||
		strings.Contains(msg, "Deadlock") ||
		strings.Contains(msg, "deadlock")
}
