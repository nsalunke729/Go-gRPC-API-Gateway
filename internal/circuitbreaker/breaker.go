// Package circuitbreaker provides a three-state (closed → open → half-open) breaker
// that protects gRPC service clients from cascading failures.
package circuitbreaker

import (
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrOpen is a gRPC Unavailable error returned when the breaker is open.
// grpcToHTTP maps codes.Unavailable → 503 so callers get the right HTTP status.
var ErrOpen = status.Error(codes.Unavailable, "circuit breaker open: service temporarily unavailable")

type breakerState int

const (
	stateClosed   breakerState = iota // normal operation; calls pass through
	stateOpen                         // backend is unhealthy; calls are rejected immediately
	stateHalfOpen                     // probe: one call allowed to test recovery
)

// Breaker is a concurrency-safe circuit breaker.
type Breaker struct {
	name      string
	mu        sync.Mutex
	current   breakerState
	failures  int
	threshold int
	timeout   time.Duration
	openedAt  time.Time
}

// New returns a Breaker that opens after threshold consecutive trippable failures
// and attempts recovery after timeout.
func New(name string, threshold int, timeout time.Duration) *Breaker {
	return &Breaker{name: name, threshold: threshold, timeout: timeout}
}

// Do executes fn through the breaker.
// Returns ErrOpen immediately when the breaker is open and the timeout has not elapsed.
func (b *Breaker) Do(fn func() error) error {
	if err := b.allow(); err != nil {
		return err
	}
	err := fn()
	b.record(err)
	return err
}

func (b *Breaker) allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.current == stateOpen {
		if time.Since(b.openedAt) > b.timeout {
			b.current = stateHalfOpen
			return nil
		}
		return ErrOpen
	}
	return nil
}

func (b *Breaker) record(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !shouldTrip(err) {
		b.failures = 0
		b.current = stateClosed
		return
	}
	b.failures++
	if b.current == stateHalfOpen || b.failures >= b.threshold {
		b.current = stateOpen
		b.openedAt = time.Now()
	}
}

// shouldTrip returns true only for errors that indicate a backend service fault.
// Application-level errors (NotFound, InvalidArgument, AlreadyExists, etc.) do not
// trip the breaker — they are valid responses from a healthy service.
func shouldTrip(err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		return true // non-gRPC error; treat as infrastructure failure
	}
	switch st.Code() {
	case codes.Unavailable, codes.Internal, codes.DeadlineExceeded:
		return true
	}
	return false
}
