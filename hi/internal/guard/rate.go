package guard

import (
	"sync"
	"time"
)

type RateLimitConfig struct {
	Tokens int           `json:"tokens" yaml:"tokens"`
	Window time.Duration `json:"window" yaml:"window"`
}

type RateLimiter struct {
	configs map[string]RateLimitConfig
	windows map[string]*slidingWindow
	mu      sync.Mutex
}

type slidingWindow struct {
	timestamps []time.Time
	config     RateLimitConfig
}

func NewRateLimiter(configs map[string]RateLimitConfig) *RateLimiter {
	if configs == nil {
		configs = make(map[string]RateLimitConfig)
	}
	if _, ok := configs["default"]; !ok {
		configs["default"] = RateLimitConfig{Tokens: 60, Window: time.Minute}
	}
	return &RateLimiter{
		configs: configs,
		windows: make(map[string]*slidingWindow),
	}
}

func (r *RateLimiter) Allow(toolName string) (bool, time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cfg, ok := r.configs[toolName]
	if !ok {
		cfg = r.configs["default"]
	}

	sw, ok := r.windows[toolName]
	if !ok {
		sw = &slidingWindow{config: cfg}
		r.windows[toolName] = sw
	}

	now := time.Now()
	cutoff := now.Add(-cfg.Window)

	i := 0
	for i < len(sw.timestamps) && sw.timestamps[i].Before(cutoff) {
		i++
	}
	sw.timestamps = sw.timestamps[i:]

	if len(sw.timestamps) >= cfg.Tokens {
		wait := cfg.Window - now.Sub(sw.timestamps[0])
		return false, wait
	}

	sw.timestamps = append(sw.timestamps, now)
	return true, 0
}

var DefaultRateLimits = map[string]RateLimitConfig{
	"default":   {Tokens: 60, Window: time.Minute},
	"bash":      {Tokens: 20, Window: time.Minute},
	"write":     {Tokens: 30, Window: time.Minute},
	"edit":      {Tokens: 30, Window: time.Minute},
	"web_fetch": {Tokens: 30, Window: time.Minute},
}
