package turn

import "hi/internal/permission"

// Re-exported from permission package for backward compatibility
// (used by turn/runner.go and external callers referencing turn.PermissionChecker).
type (
	PermissionChecker = permission.PermissionChecker
	Permission        = permission.Permission
)

const (
	PermissionAllow  = permission.PermissionAllow
	PermissionDeny   = permission.PermissionDeny
	PermissionPrompt = permission.PermissionPrompt
)
