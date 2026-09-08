package ai

import (
	"context"
	"sync"
	"time"
)

type chatRequestKey struct {
	sessionID int64
	requestID string
}

type chatRequestState struct {
	cancel  context.CancelFunc
	expires time.Time
}

// ChatRequestRegistry belongs to one application handler. Its zero value is ready
// to use. Finished and early-cancelled IDs are retained briefly so that a delayed
// request cannot resurrect generation or cancel a later request in the session.
type ChatRequestRegistry struct {
	mu       sync.Mutex
	requests map[chatRequestKey]*chatRequestState
}

const chatRequestRetention = 5 * time.Minute

func (r *ChatRequestRegistry) pruneLocked() {
	if r.requests == nil {
		r.requests = make(map[chatRequestKey]*chatRequestState)
	}
	now := time.Now()
	for key, state := range r.requests {
		if state.cancel == nil && now.After(state.expires) {
			delete(r.requests, key)
		}
	}
	// Bound terminal tombstones without ever evicting a running request.
	for len(r.requests) >= 1024 {
		var oldestKey chatRequestKey
		var oldest time.Time
		found := false
		for key, state := range r.requests {
			if state.cancel == nil && (!found || state.expires.Before(oldest)) {
				oldestKey, oldest, found = key, state.expires, true
			}
		}
		if !found {
			break
		}
		delete(r.requests, oldestKey)
	}
}

// Begin registers a unique logical request. An empty request ID preserves legacy
// callers, which can still cancel through the HTTP request context.
func (r *ChatRequestRegistry) Begin(parent context.Context, sessionID int64, requestID string) (context.Context, func()) {
	ctx, cancel := context.WithCancel(parent)
	if requestID == "" {
		return ctx, cancel
	}
	key := chatRequestKey{sessionID, requestID}
	r.mu.Lock()
	r.pruneLocked()
	if _, exists := r.requests[key]; exists {
		r.mu.Unlock()
		cancel()
		return ctx, func() {}
	}
	state := &chatRequestState{cancel: cancel}
	r.requests[key] = state
	r.mu.Unlock()
	return ctx, func() {
		r.mu.Lock()
		if r.requests[key] == state {
			state.cancel = nil
			state.expires = time.Now().Add(chatRequestRetention)
		}
		r.mu.Unlock()
		cancel()
	}
}

// Cancel is idempotent and remembers a cancellation that arrives before Begin.
func (r *ChatRequestRegistry) Cancel(sessionID int64, requestID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pruneLocked()
	key := chatRequestKey{sessionID, requestID}
	if state, exists := r.requests[key]; exists {
		if state.cancel != nil {
			state.cancel()
		}
		return
	}
	r.requests[key] = &chatRequestState{expires: time.Now().Add(chatRequestRetention)}
}

// RunIfActive serializes history/response completion with explicit cancellation.
// If completion wins the lock, the result is already committed; a later cancel
// remains a successful no-op. The caller must keep work bounded to local writes.
func (r *ChatRequestRegistry) RunIfActive(ctx context.Context, work func()) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ctx.Err() != nil {
		return false
	}
	work()
	return true
}
