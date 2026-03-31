package models

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type snapshotSeriesIndexSpec struct {
	model     any
	table     string
	indexName string
	columns   []string
}

func ensureSnapshotSeriesUniqueIndexes(db *gorm.DB) error {
	specs := []snapshotSeriesIndexSpec{
		{
			model:     &Indicator{},
			table:     "indicators",
			indexName: "idx_indicator_series_unique",
			columns:   []string{"symbol", "interval_type", "open_time"},
		},
		{
			model:     &OrderFlow{},
			table:     "orderflow",
			indexName: "idx_orderflow_series_unique",
			columns:   []string{"symbol", "interval_type", "open_time"},
		},
		{
			model:     &Structure{},
			table:     "structure",
			indexName: "idx_structure_series_unique",
			columns:   []string{"symbol", "interval_type", "open_time"},
		},
		{
			model:     &Liquidity{},
			table:     "liquidity",
			indexName: "idx_liquidity_series_unique",
			columns:   []string{"symbol", "interval_type", "open_time"},
		},
		{
			model:     &Signal{},
			table:     "signals",
			indexName: "idx_signal_series_unique",
			columns:   []string{"symbol", "interval_type", "open_time"},
		},
	}

	for _, spec := range specs {
		if err := deleteSnapshotSeriesDuplicates(db, spec.table, spec.columns); err != nil {
			return err
		}
		if db.Migrator().HasIndex(spec.model, spec.indexName) {
			continue
		}

		statement := fmt.Sprintf(
			"CREATE UNIQUE INDEX %s ON %s (%s)",
			spec.indexName,
			spec.table,
			strings.Join(spec.columns, ", "),
		)
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}

	return nil
}

func deleteSnapshotSeriesDuplicates(db *gorm.DB, table string, columns []string) error {
	partitionBy := strings.Join(columns, ", ")
	statement := fmt.Sprintf(`
DELETE FROM %s
WHERE id IN (
	SELECT id FROM (
		SELECT id,
			ROW_NUMBER() OVER (PARTITION BY %s ORDER BY id DESC) AS rn
		FROM %s
	) ranked
	WHERE rn > 1
)`, table, partitionBy, table)

	return db.Exec(statement).Error
}
