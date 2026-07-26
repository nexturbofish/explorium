package lib

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter 是 Markdown 文件开头的 YAML 元数据块，用 `---` 包裹：
//
// ---
// name: my-skill
// description: Does X
// triggers: ["analyze", "review"]
// ---

func ParseFrontmatter(content string) (map[string]any, string, error) {
	const sep = "---\n"
	if !strings.HasPrefix(content, sep) {
		return nil, content, nil
	}
	rest := strings.TrimPrefix(content, sep)
	before, after, found := strings.Cut(rest, sep)
	if !found {
		return nil, content, fmt.Errorf("unclosed frontmatter")
	}
	fm := make(map[string]any)
	if err := yaml.Unmarshal([]byte(before), &fm); err != nil {
		return nil, content, fmt.Errorf("parse yaml: %w", err)
	}
	return fm, after, nil
}

func WriteFrontmatter(w io.Writer, fm map[string]any, body string) error {
	if _, err := io.WriteString(w, "---\n"); err != nil {
		return err
	}
	enc := yaml.NewEncoder(w)
	if err := enc.Encode(fm); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}
	_, err := io.WriteString(w, "---\n"+body)
	return err
}

func AtomicWriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "*.tmp")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// EstimateToken 处略估算 token 数（4字符≈1 token）
func EstimateToken(text string) int {
	return len([]rune(text)) / 4
}
