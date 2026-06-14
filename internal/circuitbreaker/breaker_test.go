package circuitbreaker

import (
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errUnavailable = status.Error(codes.Unavailable, "backend down")

func TestBreakerClosedPassesThrough(t *testing.T) {
	b := New("svc", 3, time.Second)
	called := 0
	for range 5 {
		b.Do(func() error { called++; return nil }) //nolint:errcheck
	}
	if called != 5 {
		t.Fatalf("expected 5 calls, got %d", called)
	}
}

func TestBreakerOpensAfterThreshold(t *testing.T) {
	b := New("svc", 3, time.Second)
	for range 3 {
		b.Do(func() error { return errUnavailable }) //nolint:errcheck
	}
	err := b.Do(func() error { return nil })
	if !errors.Is(err, ErrOpen) {
		t.Fatalf("expected ErrOpen after threshold, got %v", err)
	}
}

func TestBreakerResetsOnSuccess(t *testing.T) {
	b := New("svc", 3, time.Second)
	// two failures — not yet at threshold
	for range 2 {
		b.Do(func() error { return errUnavailable }) //nolint:errcheck
	}
	// success resets the failure counter
	b.Do(func() error { return nil }) //nolint:errcheck
	// two more failures still below threshold
	for range 2 {
		b.Do(func() error { return errUnavailable }) //nolint:errcheck
	}
	err := b.Do(func() error { return nil })
	if errors.Is(err, ErrOpen) {
		t.Fatal("breaker should still be closed after reset")
	}
}

func TestBreakerHalfOpenAllowsProbe(t *testing.T) {
	b := New("svc", 3, 50*time.Millisecond)
	for range 3 {
		b.Do(func() error { return errUnavailable }) //nolint:errcheck
	}
	// wait for timeout → half-open
	time.Sleep(60 * time.Millisecond)
	// probe succeeds → back to closed
	err := b.Do(func() error { return nil })
	if err != nil {
		t.Fatalf("probe call should succeed, got %v", err)
	}
	// next call should also pass (closed)
	err = b.Do(func() error { return nil })
	if err != nil {
		t.Fatalf("post-recovery call should succeed, got %v", err)
	}
}

func TestBreakerAppErrorsDoNotTrip(t *testing.T) {
	b := New("svc", 3, time.Second)
	appErr := status.Error(codes.NotFound, "not found")
	for range 10 {
		b.Do(func() error { return appErr }) //nolint:errcheck
	}
	// breaker must stay closed; NotFound is a valid app response
	err := b.Do(func() error { return nil })
	if errors.Is(err, ErrOpen) {
		t.Fatal("application errors must not trip the circuit breaker")
	}
}
