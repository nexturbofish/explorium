package guard

import (
	"context"
	"fmt"
	"hi/internal/permission"
	"strings"
	"time"
)

const (
	confirmReasonOutsideWorkspace = "writing to path outside workspace"
	confirmReasonUnsafeCommand    = "command not in safe list"
	confirmReasonInternalNetwork  = "request to internal/private network"
	confirmReasonMultiDelete      = "deleting more than 5 files"
	confirmReasonHighTokens       = "estimated token usage exceeds 100K"
)

// PolicyEngine combines PermissionChecker + RateLimiter + ConfirmationQueue.
type PolicyEngine struct {
	permissions *permission.PermissionChecker
	rateLimiter *RateLimiter
	confirm     *ConfirmationQueue
	audit       *AuditLog
}

type PolicyResult struct {
	Allowed     bool
	DenyReason  string
	NeedConfirm bool
	ConfirmReq  *ConfirmationRequest
	WaitTime    time.Duration
}

func NewPolicyEngine(permissions *permission.PermissionChecker, rateLimiter *RateLimiter, confirm *ConfirmationQueue, audit *AuditLog) *PolicyEngine {
	return &PolicyEngine{
		permissions: permissions,
		rateLimiter: rateLimiter,
		confirm:     confirm,
		audit:       audit,
	}
}

func (e *PolicyEngine) Evaluate(ctx context.Context, tool string, args map[string]any) PolicyResult {
	start := time.Now()
	result := PolicyResult{Allowed: true}

	// 1. Rate limit check
	allowed, wait := e.rateLimiter.Allow(tool)
	if !allowed {
		result.Allowed = false
		result.DenyReason = fmt.Sprintf("rate limited, retry in %v", wait)
		result.WaitTime = wait
		e.log(tool, args, result, start)
		return result
	}

	// 2. Permission check (deny takes precedence)
	if e.permissions != nil {
		permResult := e.permissions.Check(tool)
		if permResult == permission.PermissionDeny {
			result.Allowed = false
			result.DenyReason = "tool denied by policy"
			e.log(tool, args, result, start)
			return result
		}
	}

	// 3. Confirmation needed?
	if e.needsConfirmation(tool, args) {
		req := e.confirm.Enqueue(tool, args, e.confirmReason(tool, args))
		result.NeedConfirm = true
		result.ConfirmReq = req
	}

	e.log(tool, args, result, start)
	return result
}

// Wait blocks until the confirmation request is approved, denied, or timed out.
func (e *PolicyEngine) Wait(req *ConfirmationRequest) (bool, error) {
	if e.confirm == nil {
		return true, nil
	}
	return e.confirm.Wait(req)
}

// AuditToolResult logs the outcome of a tool execution (success or failure).
// This is called after Execute, separate from Evaluate's policy audit.
func (e *PolicyEngine) AuditToolResult(tool string, args map[string]any, execErr error) {
	if e.audit == nil {
		return
	}
	entry := AuditEntry{
		Timestamp: time.Now(),
		Tool:      tool,
		Args:      args,
		Allowed:   execErr == nil,
	}
	if execErr != nil {
		entry.DenyReason = execErr.Error()
	}
	e.audit.Log(entry)
}

func (e *PolicyEngine) log(tool string, args map[string]any, result PolicyResult, start time.Time) {
	if e.audit == nil {
		return
	}
	e.audit.Log(AuditEntry{
		Timestamp:  start,
		Tool:       tool,
		Args:       args,
		Allowed:    result.Allowed,
		DenyReason: result.DenyReason,
		Duration:   time.Since(start),
	})
}

func (e *PolicyEngine) needsConfirmation(tool string, args map[string]any) bool {
	if e.confirm == nil {
		return false
	}
	switch tool {
	case "write":
		if path, ok := args["path"].(string); ok {
			return isPathOutsideWorkspace(path)
		}
	case "bash":
		if cmd, ok := args["command"].(string); ok {
			return !isSafeCommand(cmd)
		}
	case "web_fetch":
		return isPrivateTarget(args)
	case "edit":
		if files, ok := args["files"].([]any); ok && len(files) > 5 {
			return true
		}
	case "*":
		if tokens, ok := args["estimated_tokens"].(float64); ok && tokens > 100000 {
			return true
		}
	}
	return false
}

func (e *PolicyEngine) confirmReason(tool string, args map[string]any) string {
	switch tool {
	case "write":
		if path, ok := args["path"].(string); ok && isPathOutsideWorkspace(path) {
			return confirmReasonOutsideWorkspace + ": " + path
		}
	case "bash":
		if cmd, ok := args["command"].(string); ok && !isSafeCommand(cmd) {
			return confirmReasonUnsafeCommand + ": " + cmd
		}
	case "web_fetch":
		return confirmReasonInternalNetwork
	case "edit":
		if files, ok := args["files"].([]any); ok && len(files) > 5 {
			return confirmReasonMultiDelete
		}
	case "*":
		return confirmReasonHighTokens
	}
	return "unknown reason"
}

// isPathOutsideWorkspace checks if a path looks like it's outside typical safe directories.
// Simple heuristic: paths starting with /etc/, /usr/, /System, /Windows are flagged.
func isPathOutsideWorkspace(path string) bool {
	lower := strings.ToLower(path)
	for _, prefix := range []string{"/etc/", "/usr/", "/system", "/windows", "c:\\windows", "c:\\program files"} {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}

// isSafeCommand checks if a bash command is in the safe list.
func isSafeCommand(cmd string) bool {
	trimmed := strings.TrimSpace(cmd)
	safePrefixes := []string{
		"ls", "cat", "head", "tail", "echo", "pwd", "whoami",
		"grep", "find", "wc", "sort", "uniq", "cut", "tr",
		"cd", "mkdir", "cp", "mv", "rm", "chmod", "chown",
		"git", "go", "npm", "node", "python", "pip",
		"docker", "make", "curl", "wget",
	}
	for _, prefix := range safePrefixes {
		if trimmed == prefix || strings.HasPrefix(trimmed, prefix+" ") {
			return true
		}
	}
	return false
}
