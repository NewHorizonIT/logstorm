package middleware

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type serviceLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time // For cleanup purposes
}

type Manager struct {
	mu       sync.RWMutex
	limiters map[string]*serviceLimiter
	rate     rate.Limit
	burst    int
	ttl      time.Duration
}

func NewManager(rate rate.Limit, burst int, ttl time.Duration) *Manager {
	m := &Manager{
		limiters: make(map[string]*serviceLimiter),
		rate:     rate,
		burst:    burst,
		ttl:      ttl,
	}

	go m.cleanupLoop()
	return m
}

func (m *Manager) Allow(service string) bool {
	m.mu.RLock()
	sl, exists := m.limiters[service]
	m.mu.RUnlock()

	if !exists {
		m.mu.Lock()
		sl = &serviceLimiter{
			limiter:  rate.NewLimiter(m.rate, m.burst),
			lastSeen: time.Now(),
		}
		m.limiters[service] = sl
		m.mu.Unlock()
	}

	sl.lastSeen = time.Now()
	return sl.limiter.Allow()
}

func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		for k, v := range m.limiters {
			if time.Since(v.lastSeen) > m.ttl {
				delete(m.limiters, k)
			}
		}
		m.mu.Unlock()
	}
}
