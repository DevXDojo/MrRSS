package feed

import "context"

// maxConcurrentBrowserParses bounds how many headless Chrome instances
// (chromedp) may run at the same time. Each active browser parse spawns a
// temporary Chrome process costing hundreds of MB, so the gate limits peak
// memory independently of the max_concurrent_refreshes setting.
const maxConcurrentBrowserParses = 2

// browserGate is a counting semaphore that respects context cancellation.
type browserGate struct {
	sem chan struct{}
}

func newBrowserGate(limit int) *browserGate {
	if limit < 1 {
		limit = 1
	}
	return &browserGate{sem: make(chan struct{}, limit)}
}

// acquire blocks until a slot is free or ctx is done. The returned release
// function must be called to free the slot.
func (g *browserGate) acquire(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case g.sem <- struct{}{}:
		if err := ctx.Err(); err != nil {
			<-g.sem
			return nil, err
		}
		return func() { <-g.sem }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// acquireBrowserSlot lazily initializes the gate (Fetchers built by tests may
// not have one) and acquires a browser parse slot.
func (f *Fetcher) acquireBrowserSlot(ctx context.Context) (func(), error) {
	f.mu.Lock()
	if f.browserGate == nil {
		f.browserGate = newBrowserGate(maxConcurrentBrowserParses)
	}
	gate := f.browserGate
	f.mu.Unlock()
	return gate.acquire(ctx)
}
