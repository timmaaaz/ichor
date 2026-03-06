package auth

import (
	"sync"
	"time"
)

// Blocklist is an in-memory store of revoked JWT IDs (jti claims). It is safe
// for concurrent use. A background goroutine removes expired entries every 5
// minutes. For multi-node deployments, replace with a shared store (Redis, DB).
// Call Stop to terminate the goroutine on graceful shutdown. Stop is safe to
// call multiple times.
type Blocklist struct {
	mu      sync.RWMutex
	entries map[string]time.Time // jti → expiry
	done    chan struct{}
	once    sync.Once
}

// NewBlocklist creates a new Blocklist and starts its cleanup goroutine.
func NewBlocklist() *Blocklist {
	bl := &Blocklist{
		entries: make(map[string]time.Time),
		done:    make(chan struct{}),
	}
	go bl.cleanupLoop()
	return bl
}

// Add records a JWT ID as revoked until its expiry time. No-op for empty jti.
func (bl *Blocklist) Add(jti string, expiresAt time.Time) {
	if jti == "" {
		return
	}
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.entries[jti] = expiresAt
}

// IsRevoked reports whether the given JWT ID has been revoked and not yet expired.
func (bl *Blocklist) IsRevoked(jti string) bool {
	if jti == "" {
		return false
	}
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	exp, ok := bl.entries[jti]
	return ok && time.Now().Before(exp)
}

// Stop terminates the background cleanup goroutine. Call this during graceful
// shutdown to avoid goroutine leaks in tests and short-lived processes.
func (bl *Blocklist) Stop() {
	bl.once.Do(func() { close(bl.done) })
}

func (bl *Blocklist) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bl.mu.Lock()
			for jti, exp := range bl.entries {
				if time.Now().After(exp) {
					delete(bl.entries, jti)
				}
			}
			bl.mu.Unlock()
		case <-bl.done:
			return
		}
	}
}
