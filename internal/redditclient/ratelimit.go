package redditclient

import "sync"

type rateLimiter struct {
	mu        sync.RWMutex
	remaining float32
	used      int
	reset     int
}

func (rl *rateLimiter) update(remaining float32, used int, reset int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.remaining = remaining
	rl.used = used
	rl.reset = reset
}

func (rl *rateLimiter) timeToWait() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	if rl.remaining <= 0.0 {
		return rl.reset
	}
	return 0
}
