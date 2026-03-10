package models

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type tableCommenter interface {
	TableComment() string
}

type schemaCommentMetadata struct {
	Table        string
	TableComment string
	Columns      map[string]string
}

type mysqlFullColumn struct {
	Field     string         `gorm:"column:field"`
	Type      string         `gorm:"column:type"`
	Collation sql.NullString `gorm:"column:collation"`
	Null      string         `gorm:"column:is_nullable"`
	Default   sql.NullString `gorm:"column:default_value"`
	Extra     sql.NullString `gorm:"column:extra"`
	Comment   sql.NullString `gorm:"column:comment"`
}

func autoMigrateWithTableComment(db *gorm.DB, model any) error {
	if db != nil && db.Dialector.Name() == "mysql" {
		if commenter, ok := model.(tableCommenter); ok {
			comment := strings.TrimSpace(commenter.TableComment())
			if comment != "" {
				return db.Set("gorm:table_options", fmt.Sprintf("COMMENT=%s", quoteMySQLString(comment))).AutoMigrate(model)
			}
		}
	}

	return db.AutoMigrate(model)
}

func syncMySQLSchemaComments(db *gorm.DB, models []any) error {
	if db == nil || db.Dialector.Name() != "mysql" {
		return nil
	}

	for _, model := range models {
		metadata, err := collectSchemaCommentMetadata(db, model)
		if err != nil {
			return err
		}

		if metadata.TableComment != "" {
			currentComment, err := lookupMySQLTableComment(db, metadata.Table)
			if err != nil {
				return err
			}
			if currentComment != metadata.TableComment {
				statement := fmt.Sprintf(
					"ALTER TABLE %s COMMENT = %s",
					quoteMySQLIdentifier(metadata.Table),
					quoteMySQLString(metadata.TableComment),
				)
				if err := db.Exec(statement).Error; err != nil {
					return fmt.Errorf("set table comment for %s failed: %w", metadata.Table, err)
				}
			}
		}

		for column, comment := range metadata.Columns {
			columnMeta, err := lookupMySQLColumn(db, metadata.Table, column)
			if err != nil {
				return err
			}
			if columnMeta.Comment.Valid && columnMeta.Comment.String == comment {
				continue
			}

			statement := fmt.Sprintf(
				"ALTER TABLE %s MODIFY COLUMN %s %s",
				quoteMySQLIdentifier(metadata.Table),
				quoteMySQLIdentifier(column),
				buildMySQLColumnDefinition(columnMeta, comment),
			)
			if err := db.Exec(statement).Error; err != nil {
				return fmt.Errorf("set column comment for %s.%s failed: %w", metadata.Table, column, err)
			}
		}
	}

	return nil
}

func collectSchemaCommentMetadata(db *gorm.DB, model any) (schemaCommentMetadata, error) {
	parsed, err := schema.Parse(model, &sync.Map{}, db.NamingStrategy)
	if err != nil {
		return schemaCommentMetadata{}, fmt.Errorf("parse schema failed: %w", err)
	}

	metadata := schemaCommentMetadata{
		Table:   parsed.Table,
		Columns: make(map[string]string),
	}

	if commenter, ok := model.(tableCommenter); ok {
		metadata.TableComment = strings.TrimSpace(commenter.TableComment())
	}

	for _, field := range parsed.Fields {
		if field.DBName == "" || field.IgnoreMigration {
			continue
		}

		comment := strings.TrimSpace(field.TagSettings["COMMENT"])
		if comment == "" {
			continue
		}
		metadata.Columns[field.DBName] = comment
	}

	return metadata, nil
}

func lookupMySQLTableComment(db *gorm.DB, table string) (string, error) {
	var comment sql.NullString
	if err := db.Raw(
		"SELECT TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?",
		table,
	).Scan(&comment).Error; err != nil {
		return "", fmt.Errorf("query table comment for %s failed: %w", table, err)
	}

	return comment.String, nil
}

func lookupMySQLColumn(db *gorm.DB, table, column string) (mysqlFullColumn, error) {
	var result mysqlFullColumn
	if err := db.Raw(
		`SELECT
			COLUMN_NAME AS field,
			COLUMN_TYPE AS type,
			COLLATION_NAME AS collation,
			IS_NULLABLE AS is_nullable,
			COLUMN_DEFAULT AS default_value,
			EXTRA AS extra,
			COLUMN_COMMENT AS comment
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			AND COLUMN_NAME = ?`,
		table,
		column,
	).Scan(&result).Error; err != nil {
		return mysqlFullColumn{}, fmt.Errorf("query column metadata for %s.%s failed: %w", table, column, err)
	}
	if result.Field == "" {
		return mysqlFullColumn{}, fmt.Errorf("column metadata not found for %s.%s", table, column)
	}

	return result, nil
}

func buildMySQLColumnDefinition(column mysqlFullColumn, comment string) string {
	parts := []string{column.Type}

	if column.Collation.Valid && isTextColumnType(column.Type) {
		parts = append(parts, "COLLATE", column.Collation.String)
	}

	if strings.EqualFold(column.Null, "NO") {
		parts = append(parts, "NOT NULL")
	} else {
		parts = append(parts, "NULL")
	}

	if column.Default.Valid {
		parts = append(parts, "DEFAULT", formatMySQLDefaultValue(column.Type, column.Default.String))
	}

	if column.Extra.Valid && strings.TrimSpace(column.Extra.String) != "" {
		parts = append(parts, column.Extra.String)
	}

	parts = append(parts, "COMMENT", quoteMySQLString(comment))
	return strings.Join(parts, " ")
}

func formatMySQLDefaultValue(columnType, value string) string {
	trimmed := strings.TrimSpace(value)
	upper := strings.ToUpper(trimmed)

	switch {
	case upper == "NULL":
		return "NULL"
	case strings.HasPrefix(upper, "CURRENT_TIMESTAMP"):
		return trimmed
	case isNumericColumnType(columnType):
		return trimmed
	default:
		return quoteMySQLString(trimmed)
	}
}

func isTextColumnType(columnType string) bool {
	lowerType := strings.ToLower(columnType)
	textPrefixes := []string{
		"char",
		"varchar",
		"text",
		"tinytext",
		"mediumtext",
		"longtext",
		"enum",
		"set",
		"json",
	}

	for _, prefix := range textPrefixes {
		if strings.HasPrefix(lowerType, prefix) {
			return true
		}
	}

	return false
}

func isNumericColumnType(columnType string) bool {
	lowerType := strings.ToLower(columnType)
	numericPrefixes := []string{
		"tinyint",
		"smallint",
		"mediumint",
		"int",
		"bigint",
		"decimal",
		"numeric",
		"float",
		"double",
		"real",
		"bit",
	}

	for _, prefix := range numericPrefixes {
		if strings.HasPrefix(lowerType, prefix) {
			return true
		}
	}

	return false
}

func quoteMySQLIdentifier(value string) string {
	return "`" + strings.ReplaceAll(value, "`", "``") + "`"
}

func quoteMySQLString(value string) string {
	escaped := strings.ReplaceAll(value, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "'", "''")
	return "'" + escaped + "'"
}
