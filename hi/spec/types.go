package spec

import (
	"time"

	"github.com/cloudwego/eino/schema"
)

// ============ Session ============

type SessionMeta struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Model     string    `json:"model"`
	Provider  string    `json:"provider"`
}

type Session struct {
	Meta             SessionMeta
	Messages         []*schema.Message
	TotalInputToken  int
	TotalOutputToken int
}

// ============ Memory ============

type MemoryScope string

const (
	MemoryScopeUser    MemoryScope = "user"
	MemoryScopeProject MemoryScope = "project"
)

type MemorySource string

const (
	MemorySourceReflection MemorySource = "reflection"
	MemorySourceUser       MemorySource = "user"
	MemorySourceImported   MemorySource = "imported"
)

type MemoryConfidence int

const (
	MemoryConfidenceLow    MemoryConfidence = 1
	MemoryConfidenceMedium MemoryConfidence = 2
	MemoryConfidenceHigh   MemoryConfidence = 3
)

type MemoryFrontmatter struct {
	ID          string           `yaml:"id"`
	CreatedAt   time.Time        `yaml:"created_at"`
	Scope       MemoryScope      `yaml:"scope"`
	Source      MemorySource     `yaml:"source"`
	Confidence  MemoryConfidence `yaml:"confidence"`
	Tags        []string         `yaml:"tags,omitempty"`
	Zone        string           `yaml:"zone"`
	Pinned      bool             `yaml:"pinned,omitempty"`
	Supersedes  []string         `yaml:"supersedes,omitempty"`
	AccessedAt  time.Time        `yaml:"accessed_at,omitempty"`
	AccessCount int              `yaml:"access_count,omitempty"`
}

type LoadedMemory struct {
	Frontmatter MemoryFrontmatter
	Content     string
	Path        string
}

// ============ Skill ============

type SkillScope string

const (
	SkillScopeUser    SkillScope = "user"
	SkillScopeProject SkillScope = "project"
)

type SkillFrontmatter struct {
	Name         string     `yaml:"name"`
	Description  string     `yaml:"description"`
	Triggers     []string   `yaml:"triggers,omitempty"`
	AlwaysActive bool       `yaml:"always_active,omitempty"`
	Version      string     `yaml:"version,omitempty"`
	Author       string     `yaml:"author"`
	Scope        SkillScope `yaml:"scope"`
}

type LoadedSkill struct {
	Frontmatter SkillFrontmatter
	Body        string
	Path        string
	Slug        string
}

// ============ Reflection ============

type SkillCandidate struct {
	Name        string   `json:"name"`
	Description string   `json:"descripton"`
	Triggers    []string `json:"triggers"`
	Body        string   `json:"body"`
	Confidence  string   `json:"confidence"` // low, medium, high
	Rationale   string   `json:"rationale"`  // why this skill was proposed
}

type MemoryCandidate struct {
	Content    string   `json:"content"`
	Tags       []string `json:"tags"`
	Zone       string   `json:"zone"`
	Source     string   `json:"source"`
	Confidence string   `json:"confidence"`
	Supersedes []string `json:"supersedes,omitempty"`
	Rationale  string   `json:"rationale"` // why this skill was proposed
}

type ReflectionOutput struct {
	Skills    []SkillCandidate    `json:"skills"`
	Memories  []MemoryCandidate   `json:"memorise"`
	Conflicts []ConflictCandidate `json:"conflicts"`
	Summary   string              `json:"summary"`
}

type ConflictCandidate struct {
	ExistingID  string `json:"existing_id"`
	NewContent  string `json:"new_content"`
	Explanation string `json:"explanation"`
	Resolution  string `json:"resolution"`
}

// ============ Channel ============

type ChannelType string

const (
	ChannelFeishu   ChannelType = "feishu"
	ChannelWechat   ChannelType = "wechat"
	ChannelTelegram ChannelType = "telegram"
)

type ChannelMessage struct {
	ID        string
	Channel   ChannelType
	UserID    string
	Text      string
	Timestamp time.Time
	Raw       any // 原始协议消息
}

type ChannelConfig struct {
	Type    ChannelType       `json:"type"`
	Enabled bool              `json:"enabled"`
	Options map[string]string `json:"options"`
}

// ============ Session ============

// ============ Session ============

// ============ Session ============

// ============ Session ============
