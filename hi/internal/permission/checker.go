package permission

import (
	"hi/internal/config"

	"github.com/gobwas/glob"
)

type Permission int

const (
	PermissionAllow Permission = iota
	PermissionDeny
	PermissionPrompt
)

// PermissionChecker 基于 deny-takes-precedence 策略
// - Deny 列表优先于 Allow 列表
// - 支持 glob 模式匹配（使用 github.com/gobwas/glob）
// - 工具名称自动添加 "mcp:" 前缀（MCP 工具）
type PermissionChecker struct {
	allow []glob.Glob
	deny  []glob.Glob
}

func NewPermissionChecker(cfg config.PermissionsConfig) *PermissionChecker {
	allow := make([]glob.Glob, 0, len(cfg.Allow))
	for _, p := range cfg.Allow {
		if pat, err := glob.Compile(p); err == nil {
			allow = append(allow, pat)
		}
	}
	deny := make([]glob.Glob, 0, len(cfg.Deny))
	for _, p := range cfg.Deny {
		if pat, err := glob.Compile(p); err == nil {
			deny = append(deny, pat)
		}
	}
	return &PermissionChecker{allow: allow, deny: deny}
}

func (pc *PermissionChecker) Check(toolName string) Permission {
	for _, pat := range pc.deny {
		if pat.Match(toolName) {
			return PermissionDeny
		}
	}
	for _, pat := range pc.allow {
		if pat.Match(toolName) {
			return PermissionAllow
		}
	}
	return PermissionPrompt
}
