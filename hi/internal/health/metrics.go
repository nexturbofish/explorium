package health

import (
	"sync"
	"time"
)

type Metrics struct {
	ToolCalls  map[string]int64         `json:"tool_calls"`
	ToolErrors map[string]int64         `json:"tool_errors"`
	TokenUsage map[string]int64         `json:"token_usage"` // model -> tokens
	AvgLatency map[string]time.Duration `json:"avg_latency"`
	Uptime     time.Duration            `json:"uptime"`

	mu        sync.RWMutex
	startTime time.Time
	latencies map[string][]time.Duration
}

func NewMetrics() *Metrics {
	return &Metrics{
		ToolCalls:  make(map[string]int64),
		ToolErrors: make(map[string]int64),
		TokenUsage: make(map[string]int64),
		AvgLatency: make(map[string]time.Duration),
		startTime:  time.Now(),
		latencies:  make(map[string][]time.Duration),
	}
}

func (m *Metrics) RecordToolCall(tool string, err error, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ToolCalls[tool]++
	if err != nil {
		m.ToolErrors[tool]++
	}
	m.latencies[tool] = append(m.latencies[tool], latency)
	if len(m.latencies[tool]) > 100 {
		m.latencies[tool] = m.latencies[tool][1:]
	}
	var total time.Duration
	for _, l := range m.latencies[tool] {
		total += l
	}
	m.AvgLatency[tool] = total / time.Duration(len(m.latencies[tool]))
}

func (m *Metrics) RecordTokenUsage(model string, tokens int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TokenUsage[model] += int64(tokens)
}

type MetricsSnapshot struct {
	ToolCalls  map[string]int64         `json:"tool_calls"`
	ToolErrors map[string]int64         `json:"tool_errors"`
	TokenUsage map[string]int64         `json:"token_usage"`
	AvgLatency map[string]time.Duration `json:"avg_latency"`
	Uptime     time.Duration            `json:"uptime"`
}

func (m *Metrics) Snapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return MetricsSnapshot{
		ToolCalls:  copyMap(m.ToolCalls),
		ToolErrors: copyMap(m.ToolErrors),
		TokenUsage: copyMap(m.TokenUsage),
		AvgLatency: copyMapDuration(m.AvgLatency),
		Uptime:     time.Since(m.startTime),
	}
}

func copyMap(src map[string]int64) map[string]int64 {
	dst := make(map[string]int64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func copyMapDuration(src map[string]time.Duration) map[string]time.Duration {
	dst := make(map[string]time.Duration, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
