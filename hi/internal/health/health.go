package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type HealthStatus struct {
	Status string            `json:"status"` // "ok" | "degraded" | "down"
	Uptime time.Duration     `json:"uptime"`
	Checks map[string]string `json:"checks"` // component -> status
}

type HealthCheckFunc func(ctx context.Context) error

type HealthChecker struct {
	startTime time.Time
	checks    map[string]HealthCheckFunc
	mu        sync.RWMutex
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		startTime: time.Now(),
		checks:    make(map[string]HealthCheckFunc),
	}
}

func (h *HealthChecker) Register(name string, fn HealthCheckFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = fn
}

func (h *HealthChecker) Check(ctx context.Context) HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	status := HealthStatus{
		Status: "ok",
		Uptime: time.Since(h.startTime),
		Checks: make(map[string]string),
	}

	for name, fn := range h.checks {
		if err := fn(ctx); err != nil {
			status.Checks[name] = fmt.Sprintf("error: %v", err)
			if status.Status == "ok" {
				status.Status = "degraded"
			}
		} else {
			status.Checks[name] = "ok"
		}
	}

	return status
}
