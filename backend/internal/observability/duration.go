package observability

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// Field 表示一条结构化日志字段。
type Field struct {
	Key   string
	Value any
}

// String 构造字符串字段。
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int 构造整数字段。
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 构造 64 位整数字段。
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float 构造浮点数字段。
func Float(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool 构造布尔字段。
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// LogDuration 输出统一的耗时日志。
func LogDuration(component, stage string, startedAt time.Time, status, reason string, fields ...Field) {
	if status == "" {
		status = "ok"
	}

	payload := []Field{
		String("component", component),
		String("stage", stage),
		String("duration", time.Since(startedAt).String()),
		String("status", status),
	}
	if strings.TrimSpace(reason) != "" {
		payload = append(payload, String("reason", reason))
	}
	payload = append(payload, fields...)

	log.Printf(render(payload...))
}

// Log 输出非耗时结构化日志。
func Log(component, stage string, fields ...Field) {
	payload := []Field{
		String("component", component),
		String("stage", stage),
	}
	payload = append(payload, fields...)
	log.Printf(render(payload...))
}

func render(fields ...Field) string {
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		if strings.TrimSpace(field.Key) == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%v", field.Key, field.Value))
	}
	return strings.Join(parts, " ")
}
