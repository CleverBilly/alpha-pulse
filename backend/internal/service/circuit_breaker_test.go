package service

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreakerOpensAfterConsecutiveFailures(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	err := errors.New("api error")
	for i := 0; i < 3; i++ {
		cb.RecordFailure(err)
	}

	if !cb.IsOpen() {
		t.Error("expected circuit to be open after 3 consecutive failures")
	}
}

func TestCircuitBreakerResetsOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	cb.RecordFailure(errors.New("err"))
	cb.RecordFailure(errors.New("err"))
	cb.RecordSuccess()

	if cb.IsOpen() {
		t.Error("expected circuit to be closed after success")
	}
}

func TestCircuitBreakerRecoversAfterCooldown(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	cb.RecordFailure(errors.New("err"))
	cb.RecordFailure(errors.New("err"))

	if !cb.IsOpen() {
		t.Fatal("expected open circuit")
	}

	time.Sleep(60 * time.Millisecond)

	if cb.IsOpen() {
		t.Error("expected circuit to be closed after cooldown")
	}
}
