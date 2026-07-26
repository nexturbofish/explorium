package lib

import (
	"fmt"
	"path/filepath"
	"strings"
)

func SafePath(workspace, target string) (string, error) {
	base, err := filepath.Abs(workspace)
	if err != nil {
		return "", fmt.Errorf("safe path: bad workspace: %q: %w", workspace, err)
	}
	clean := filepath.Join(base, target)
	clean, err = filepath.Abs(clean)
	if err != nil {
		return "", fmt.Errorf("safe path: bad target %q: %w", clean, err)
	}
	prefix := base + string(filepath.Separator)
	if !strings.HasPrefix(clean, prefix) && clean != base {
		return "", fmt.Errorf("path traversal detected: %s", target)
	}
	return clean, nil
}
