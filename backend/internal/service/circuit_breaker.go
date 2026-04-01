package service

import (
	"sync"
	"time"
)

// CircuitBreaker 简单计数式熔断器：连续失败 threshold 次后打开，冷却后自动关闭。
type CircuitBreaker struct {
	mu               sync.Mutex
	threshold        int
	cooldown         time.Duration
	consecutiveFails int
	openedAt         time.Time
}

// NewCircuitBreaker 创建熔断器。threshold=连续失败多少次触发熔断；cooldown=熔断冷却时长。
func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{threshold: threshold, cooldown: cooldown}
}

// IsOpen 返回熔断器是否处于打开（熔断）状态。冷却期结束后自动关闭。
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.openedAt.IsZero() {
		return false
	}
	if time.Since(cb.openedAt) >= cb.cooldown {
		cb.openedAt = time.Time{} // 自动关闭
		cb.consecutiveFails = 0
		return false
	}
	return true
}

// RecordFailure 记录一次失败，达到阈值时打开熔断器。已打开时不再累加计数。
func (cb *CircuitBreaker) RecordFailure(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if !cb.openedAt.IsZero() {
		return // 熔断已打开，避免计数无限累加
	}
	cb.consecutiveFails++
	if cb.consecutiveFails >= cb.threshold {
		cb.openedAt = time.Now()
	}
}

// RecordSuccess 记录成功，重置连续失败计数。
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.consecutiveFails = 0
	cb.openedAt = time.Time{}
}
