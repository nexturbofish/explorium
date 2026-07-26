package turn

import (
	"context"
	"hi/internal/config"
	"hi/internal/lib"
	"hi/spec"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// ContextAssembler 在每次 RunTurn() 前动态构建 system prompt
// 结果直接注入到 schema.Message 的 system role 中
type ContextAssembler struct {
	BaseSystem  string // 基础 system prompt
	MemoryStore spec.MemoryStore
	SkillStore  spec.SkillStore
	Config      config.LimitsConfig
}

type AssembledContext struct {
	System string            // 完整的 system prompt
	Usage  schema.TokenUsage // 提示词占用的 token（策略估算）
}

// Assemble 组装系统提示词
//
//	区块                 │ 内容                 │ 目的
//
// ──────────────────────┼──────────────────────┼──────────────────────────────────────────────────────────
// 1. Pinned Memories   │ 完整  Body           │ 用户显式固定的重要记忆，始终呈现
// 2. Active Memories   │  [ID] Body  摘要格式  │ 所有活跃记忆的目录（索引视图），让模型知道有哪些记忆可用
// 3. Active Skills     │  Name: Description  │ 所有技能元数据（名称+描述），让模型知道能调用什么
// 4. Relevant Memories │ 完整  Body           │ 和当前输入相关的记忆详情
func (ca *ContextAssembler) Assemble(ctx context.Context, userInput string) (*AssembledContext, error) {
	var sb strings.Builder
	sb.WriteString(ca.BaseSystem)
	sb.WriteString("\n\n")

	// 1. 加载 pinned memories（置顶记忆片段，不受 tf-idf 搜索影响、不参与老化淘汰，每次对话都注入 system prompt）
	pinned, _ := ca.MemoryStore.ListPinned(ctx)
	if len(pinned) > 0 {
		sb.WriteString("## Pinned Memories\n")
		for _, m := range pinned {
			sb.WriteString("- ")
			sb.WriteString(m.Content)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// 2. 加载 active memory index (capped)
	// [TODO] 当前输出完整 Body 作为索引，后续 phase-04 应改为仅输出截断摘要
	//（如 Body[:120]），避免与第 4 步的 relevant memory 重复。
	allActive, _ := ca.MemoryStore.ListActive(ctx)
	cap := ca.Config.ActiveMemoryIndexCap
	if cap > 0 && len(allActive) > cap {
		allActive = allActive[:cap]
	}
	if len(allActive) > 0 {
		sb.WriteString("## Active Memories\n")
		for _, m := range allActive {
			sb.WriteString("- [")
			sb.WriteString(m.Frontmatter.ID)
			sb.WriteString("]")
			sb.WriteString(strings.ReplaceAll(m.Content, "\n", " "))
		}
		sb.WriteString("\n")
	}

	// 3. 加载 skill index（capped）
	allSkills, _ := ca.SkillStore.List(ctx)
	skillCap := ca.Config.SkillIndexCap
	if skillCap > 0 && len(allSkills) > skillCap {
		allSkills = allSkills[:skillCap]
	}
	if len(allSkills) > 0 {
		sb.WriteString("## Active Skills\n")
		for _, sk := range allSkills {
			sb.WriteString("- ")
			sb.WriteString(sk.Frontmatter.Name)
			sb.WriteString(": ")
			sb.WriteString(sk.Frontmatter.Description)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// 4. [TODO] 语义匹配相关记忆 —— 当前为简化占位，仅做子串匹配
	// 真正的 TF-IDF 语义搜索将在 phase-04 (memory/relevance.go) 中实现，
	// 届时 Assemble() 应调用 lib.TfidfSearch() 替代此处逻辑。
	//
	// 注意：当前实现中 relevant 匹配与第 2 步的 active memories 存在内容
	// 重复（这里输出完整 body，第 2 步输出摘要）。后续应改为：第 2 步仅生成
	// 索引摘要（[ID] 开头一行），第 4 步只展开命中条目的完整 body；
	// 或第 4 步直接从 active memories 中去重后输出相关内容。
	q := strings.ToLower(userInput)
	for _, m := range allActive {
		if strings.Contains(strings.ToLower(m.Content), q) {
			sb.WriteString("## Relevant Memory\n")
			sb.WriteString(m.Content)
			sb.WriteString("\n\n")
		}
	}

	return &AssembledContext{
		System: sb.String(),
		Usage:  schema.TokenUsage{PromptTokens: lib.EstimateToken(sb.String())},
	}, nil
}
