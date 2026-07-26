package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SkillEventType 技能事件的类型
type SkillEventType string

const (
	SkillInvoked   SkillEventType = "invoked"   // 技能被激活
	SkillCompleted SkillEventType = "completed" // 技能执行完成
	SkillFailed    SkillEventType = "failed"    // 技能执行失败
	SkillInstalled SkillEventType = "installed" // 技能被安装
	SkillDeleted   SkillEventType = "deleted"   // 技能被删除
)

// SkillEvent 技能事件记录
type SkillEvent struct {
	Time    time.Time      `json:"time"`
	Event   SkillEventType `json:"event"`
	Skill   string         `json:"skill"`
	Slug    string         `json:"slug,omitempty"`
	Score   float64        `json:"score,omitempty"`
	Trigger string         `json:"trigger,omitempty"` // 触发词
	Error   string         `json:"error,omitempty"`
}

// SkillStats 聚合后的技能统计数据
type SkillStats struct {
	Name             string    `json:"name"`
	Invoked          int       `json:"invoked"`
	Completed        int       `json:"completed"`
	Failed           int       `json:"failed"`
	AvgScore         float64   `json:"avg_score"`
	LastUsed         time.Time `json:"last_used"`
}

// StatsRecorder 记录技能事件，输出到 JSONL 文件
type StatsRecorder struct {
	dir string
	mu  sync.Mutex
	f   *os.File
}

// NewStatsRecorder 创建记录器，日志写入 rootDir/skill_stats.jsonl
func NewStatsRecorder(rootDir string) (*StatsRecorder, error) {
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return nil, fmt.Errorf("create stats dir: %w", err)
	}
	path := filepath.Join(rootDir, "skill_stats.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open stats file: %w", err)
	}
	return &StatsRecorder{dir: rootDir, f: f}, nil
}

// Record 记录一条技能事件
func (r *StatsRecorder) Record(ctx context.Context, event SkillEventType, slug, trigger string, score float64, errStr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e := SkillEvent{
		Time:    time.Now(),
		Event:   event,
		Slug:    slug,
		Score:   score,
		Trigger: trigger,
		Error:   errStr,
	}
	data, _ := json.Marshal(e)
	data = append(data, '\n')
	r.f.Write(data)
}

// RecordInvoked 快速记录技能被调用
func (r *StatsRecorder) RecordInvoked(slug, trigger string) {
	r.Record(context.Background(), SkillInvoked, slug, trigger, 0, "")
}

// RecordCompleted 快速记录技能完成
func (r *StatsRecorder) RecordCompleted(slug string, score float64) {
	r.Record(context.Background(), SkillCompleted, slug, "", score, "")
}

// RecordFailed 快速记录技能失败
func (r *StatsRecorder) RecordFailed(slug string, errStr string) {
	r.Record(context.Background(), SkillFailed, slug, "", 0, errStr)
}

// Aggregate 从 JSONL 读取并聚合所有技能统计
func (r *StatsRecorder) Aggregate(ctx context.Context) (map[string]*SkillStats, error) {
	r.mu.Lock()
	r.f.Sync()
	r.mu.Unlock()

	path := filepath.Join(r.dir, "skill_stats.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]*SkillStats{}, nil
		}
		return nil, fmt.Errorf("read stats: %w", err)
	}

	stats := make(map[string]*SkillStats)
	var totalScore map[string]float64
	var countScore map[string]int

	lines := 0
	for len(data) > 0 {
		var e SkillEvent
		end := 0
		for end < len(data) && data[end] != '\n' {
			end++
		}
		line := data[:end]
		if end < len(data) {
			data = data[end+1:]
		} else {
			data = nil
		}
		lines++

		if len(line) == 0 {
			continue
		}
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}

		s, ok := stats[e.Slug]
		if !ok {
			s = &SkillStats{Name: e.Slug}
			stats[e.Slug] = s
			totalScore[e.Slug] = 0
			countScore[e.Slug] = 0
		}
		if totalScore == nil {
			totalScore = make(map[string]float64)
		}
		if countScore == nil {
			countScore = make(map[string]int)
		}

		s.LastUsed = e.Time

		switch e.Event {
		case SkillInvoked:
			s.Invoked++
		case SkillCompleted:
			s.Completed++
			totalScore[e.Slug] += e.Score
			countScore[e.Slug]++
		case SkillFailed:
			s.Failed++
		}
	}

	for slug, s := range stats {
		if countScore[slug] > 0 {
			s.AvgScore = totalScore[slug] / float64(countScore[slug])
		}
	}

	return stats, nil
}

// Close 关闭记录器
func (r *StatsRecorder) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.f != nil {
		return r.f.Close()
	}
	return nil
}
