package skills

// 内置技能名称 -- 不可删除，启动时自动恢复
var BundledSkills = []string{"memory-palace", "skill-creator", "find-skills"}

const (
	MaxInstallFiles      = 50        // 单词安装最多文件数
	MaxInstallFileBytes  = 100 << 10 // 单文件最大 100 KB
	MaxInstallTotalBytes = 5 << 20   // 总计最大 5 MB
)

// InstallOutcome 安装结果
type InstallOutcome struct {
	Name         string
	Description  string
	ResolvedRef  string   // 实际使用的 git ref
	FilesWritten []string // 相对于技能目录的路径
	TotalBytes   int64
}

// DeleteOutcome 删除结果
type DeleteOutcome struct {
	Name         string
	FilesRemoved int
}
