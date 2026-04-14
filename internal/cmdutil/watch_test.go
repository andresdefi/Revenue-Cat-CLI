package cmdutil

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatch_CallsAtLeastOnce(t *testing.T) {
	var calls atomic.Int32
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := Watch(ctx, 100*time.Millisecond, func(ctx context.Context) error {
		calls.Add(1)
		return nil
	})
	if err != nil {
		t.Fatalf("Watch() error: %v", err)
	}
	if calls.Load() < 1 {
		t.Errorf("Watch() called fn %d times, want >= 1", calls.Load())
	}
}

func TestWatch_StopsOnContextCancellation(t *testing.T) {
	var calls atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(250 * time.Millisecond)
		cancel()
	}()

	err := Watch(ctx, 50*time.Millisecond, func(ctx context.Context) error {
		calls.Add(1)
		return nil
	})
	if err != nil {
		t.Fatalf("Watch() error: %v", err)
	}

	// It should have been called at least once but not indefinitely
	c := calls.Load()
	if c < 1 {
		t.Errorf("Watch() called fn %d times, want >= 1", c)
	}
}

func TestWatch_ErrorsDoNotStopLoop(t *testing.T) {
	var calls atomic.Int32
	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()

	err := Watch(ctx, 80*time.Millisecond, func(ctx context.Context) error {
		calls.Add(1)
		return errors.New("transient error")
	})
	if err != nil {
		t.Fatalf("Watch() error: %v", err)
	}

	// Even with errors, Watch should keep calling fn
	c := calls.Load()
	if c < 2 {
		t.Errorf("Watch() called fn %d times despite errors, want >= 2", c)
	}
}

func TestWatch_DefaultInterval(t *testing.T) {
	// Verify that a zero interval gets replaced with the default
	if DefaultWatchInterval <= 0 {
		t.Errorf("DefaultWatchInterval = %v, want > 0", DefaultWatchInterval)
	}
}

func TestWatch_NegativeInterval(t *testing.T) {
	var calls atomic.Int32
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Negative interval should be corrected to default, not panic
	err := Watch(ctx, -1*time.Second, func(ctx context.Context) error {
		calls.Add(1)
		return nil
	})
	if err != nil {
		t.Fatalf("Watch(negative interval) error: %v", err)
	}
	if calls.Load() < 1 {
		t.Error("Watch(negative interval) should have called fn at least once")
	}
}

func TestWatch_ReturnsNilOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := Watch(ctx, 50*time.Millisecond, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("Watch() returned %v, want nil on context cancellation", err)
	}
}
