package feed

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBrowserGate_LimitsConcurrency(t *testing.T) {
	gate := newBrowserGate(2)

	var active, peak int64
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			release, err := gate.acquire(context.Background())
			if err != nil {
				t.Errorf("acquire failed: %v", err)
				return
			}
			defer release()

			current := atomic.AddInt64(&active, 1)
			for {
				peakSnapshot := atomic.LoadInt64(&peak)
				if current <= peakSnapshot || atomic.CompareAndSwapInt64(&peak, peakSnapshot, current) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt64(&active, -1)
		}()
	}
	wg.Wait()

	if peak > 2 {
		t.Errorf("Expected peak concurrency <= 2, got %d", peak)
	}
}

func TestBrowserGate_RespectsContextCancellation(t *testing.T) {
	gate := newBrowserGate(1)

	release, err := gate.acquire(context.Background())
	if err != nil {
		t.Fatalf("initial acquire failed: %v", err)
	}
	defer release()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := gate.acquire(ctx)
		done <- err
	}()

	// Give the goroutine a moment to block on the full gate, then cancel.
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("acquire did not return after context cancellation")
	}
}

func TestBrowserGate_MinimumLimit(t *testing.T) {
	gate := newBrowserGate(0)

	// An invalid limit must clamp to at least one usable slot.
	release, err := gate.acquire(context.Background())
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	release()

	release, err = gate.acquire(context.Background())
	if err != nil {
		t.Fatalf("second acquire failed: %v", err)
	}
	release()
}
