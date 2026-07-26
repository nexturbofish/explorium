package skills

import (
	"context"
	"fmt"
	"hi/internal/lib"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// InstallFromGitHub 从 GitHub 仓库安装技能。
// 默认使用 main 分支, 下载 SKILL.md 及其附件的目录结构。
func InstallFromGitHub(ctx context.Context, slug, destDir string) (*InstallOutcome, error) {
	owner, repo, path, skillName, err := parseGitHubSlug(slug)
	if err != nil {
		return nil, err
	}

	// 检查是否与 bundled skill 重名
	for _, b := range BundledSkills {
		if strings.EqualFold(skillName, b) {
			return nil, fmt.Errorf("skill name %q is reserved for bundled skills", skillName)
		}
	}

	ref := "main"
	basePath := path
	if basePath == "" {
		basePath = "."
	}

	// 获取 SKILL.md
	skillPath := filepath.Join(basePath, "SKILL.md")
	data, err := fetchRawGitHub(ctx, owner, repo, ref, skillPath)
	if err != nil {
		return nil, fmt.Errorf("fetch SKILL.md: %w", err)
	}

	// 解析并校验 frontmatter
	fm, body, err := lib.ParseFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("invalid frontmatter in SKILL.md: %w", err)
	}
	if lib.GetString(fm, "name") == "" {
		return nil, fmt.Errorf("SKILL.md frontmatter missing required field: name")
	}
	if lib.GetString(fm, "description") == "" {
		return nil, fmt.Errorf("SKILL.md frontmatter missing required field: description")
	}

	// 创建目标目录
	skillDir := filepath.Join(destDir, filepath.Clean(skillName))
	// 防目录穿越
	if !strings.HasPrefix(filepath.Clean(skillDir), filepath.Clean(destDir)+string(filepath.Separator)) && skillDir != filepath.Clean(destDir) {
		return nil, fmt.Errorf("skill name %q escapes base directory", skillName)
	}
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return nil, fmt.Errorf("create skill dir: %w", err)
	}

	outcome := &InstallOutcome{
		Name:        skillName,
		Description: lib.GetString(fm, "description"),
		ResolvedRef: ref,
		TotalBytes:  int64(len(data)),
	}

	// 事务写入: 先写 temp 再重命名
	tmpFile, err := os.CreateTemp(skillDir, "SKILL.md.tmp.*")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	if err := lib.WriteFrontmatter(tmpFile, fm, body); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return nil, fmt.Errorf("write SKILL.md: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("close temp file: %w", err)
	}

	finalPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("rename %s -> %s: %w", tmpPath, finalPath, err)
	}

	outcome.FilesWritten = []string{"SKILL.md"}
	return outcome, nil
}

// DeleteSkill 删除已安装的技能。拒绝删除 bundled skills。
func DeleteSkill(ctx context.Context, rootDir, skillName string) (*DeleteOutcome, error) {
	for _, b := range BundledSkills {
		if strings.EqualFold(skillName, b) {
			return nil, fmt.Errorf("skill %q is a bundled skill and cannot be deleted", skillName)
		}
	}

	skillDir := filepath.Join(rootDir, skillName)
	info, err := os.Stat(skillDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("skill %q not found", skillName)
		}
		return nil, fmt.Errorf("stat %s: %w", skillDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a skill directory", skillDir)
	}

	// 删除后统计文件数
	var filesRemoved int
	err = filepath.WalkDir(skillDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			filesRemoved++
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", skillDir, err)
	}

	if err := os.RemoveAll(skillDir); err != nil {
		return nil, fmt.Errorf("remove %s: %w", skillDir, err)
	}

	return &DeleteOutcome{
		Name:         skillName,
		FilesRemoved: filesRemoved,
	}, nil
}

// parseGitHubSlug 解析两种 slug 格式:
//
//	"github:owner/repo/path" — GitHub Contents API 路径
//	"owner/repo@skill-name"  — 简写格式
func parseGitHubSlug(slug string) (owner, repo, path, skillName string, err error) {
	if strings.HasPrefix(slug, "github:") {
		// github:owner/repo/path -> owner, repo, path
		rest := strings.TrimPrefix(slug, "github:")
		parts := strings.SplitN(rest, "/", 3)
		if len(parts) < 2 {
			return "", "", "", "", fmt.Errorf("invalid github slug %q: expected github:owner/repo/path", slug)
		}
		owner, repo = parts[0], parts[1]
		if len(parts) == 3 {
			path = parts[2]
		}
		skillName = repo
		if path != "" {
			skillName = filepath.Base(path)
		}
		return
	}
	// owner/repo@skill-name
	at := strings.LastIndex(slug, "@")
	if at < 1 {
		return "", "", "", "", fmt.Errorf("invalid slug %q: expected owner/repo@slug or github:owner/repo/path", slug)
	}
	repoPart := slug[:at]
	skillName = slug[at+1:]
	parts := strings.SplitN(repoPart, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", "", fmt.Errorf("invalid repo part %q in slug %q", repoPart, slug)
	}
	return parts[0], parts[1], "", skillName, nil
}

// fetchRawGitHub 从 raw.githubusercontent.com 下载文件
func fetchRawGitHub(ctx context.Context, owner, repo, ref, filePath string) ([]byte, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, ref, filePath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: HTTP %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxInstallFileBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", url, err)
	}
	if len(body) > MaxInstallFileBytes {
		return nil, fmt.Errorf("file %s exceeds %d bytes", filePath, MaxInstallFileBytes)
	}
	return body, nil
}
