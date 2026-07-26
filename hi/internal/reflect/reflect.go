package reflect

import (
	"context"
	"hi/internal/lib"
	"hi/spec"
	"os"

	"github.com/cloudwego/eino/components/model"
)

type ReflectConfig struct {
	Model                model.ToolCallingChatModel
	MemoryStore          spec.MemoryStore
	SkillStore           spec.SkillStore
	DeferredQueue        *DeferredQueue
	MinTurns             int
	AutoAcceptMemories   bool
	AutoAcceptConfidence string // low medium high
	logStore             *reflectLogStore
}

type ReflectOptions struct {
	Mode       string // full micro focussed
	FocusTopic string // focused mode 使用
}

// Reflect runs one reflection pass. Returns the output even when empty.
func Reflect(ctx context.Context, session *spec.Session, config ReflectConfig, opts ReflectOptions) (*spec.ReflectionOutput, error) {
	if len(session.Messages) < config.MinTurns*2 {
		return &spec.ReflectionOutput{}, nil
	}

	// Initialize log store on first use
	if config.logStore == nil {
		home, _ := os.UserHomeDir()
		config.logStore = newReflectLogStore(home)
	}

	recentOutcomes := config.logStore.OutcomesSummary(10)

	var output *spec.ReflectionOutput
	var err error
	switch opts.Mode {
	case "micro":
		output, err = MicroReflect(ctx, config.Model, session, recentOutcomes)
	case "focused":
		output, err = FocusedReflect(ctx, session, config.Model, opts.FocusTopic, recentOutcomes)
	default:
		output, err = FullReflect(ctx, session, config.Model, recentOutcomes)
	}
	if err != nil {
		return nil, err
	}

	// Auto-accept eligible memories
	if config.AutoAcceptMemories {
		threshold := lib.ConfidenceToFloat(config.AutoAcceptConfidence)
		var kept []spec.MemoryCandidate
		for _, m := range output.Memories {
			if lib.ConfidenceToFloat(m.Confidence) >= threshold {
				if err := config.MemoryStore.Save(memCandidateToLoaded(&m)); err == nil {
					config.logStore.Append(ReflectLogEntry{
						Kind:   CandidateMemory,
						Action: ActionAutoAuto,
						Label:  m.Content,
					})
				}
			} else {
				kept = append(kept, m)
			}
		}
		output.Memories = kept
	}

	// Deferred queue
	if config.DeferredQueue != nil {
		for _, m := range output.Memories {
			config.DeferredQueue.Add(&DeferredCandidate{
				Type:  "memory",
				Data:  m,
				Score: lib.ConfidenceToFloat(m.Confidence),
			})
		}
		for _, s := range output.Skills {
			config.DeferredQueue.Add(&DeferredCandidate{
				Type:  "skill",
				Data:  s,
				Score: lib.ConfidenceToFloat(s.Confidence),
			})
		}
	}

	// Log reflection outcomes
	label := opts.Mode + "-reflection completed"
	if opts.Mode == "focused" && opts.FocusTopic != "" {
		label = "focused-reflection on " + opts.FocusTopic
	}
	config.logStore.Append(ReflectLogEntry{
		Kind:   CandidateSkill,
		Action: ActionDefer,
		Label:  label,
	})

	return output, nil
}

// memCandidateToLoaded 将 MemoryCandidate 转为 MemoryStore.LoadedMemory 用于保存。
func memCandidateToLoaded(m *spec.MemoryCandidate) *spec.LoadedMemory {
	return &spec.LoadedMemory{
		Content: m.Content,
		Frontmatter: spec.MemoryFrontmatter{
			Tags:   m.Tags,
			Zone:   m.Zone,
			Source: spec.MemorySource(m.Source),
		},
	}
}
