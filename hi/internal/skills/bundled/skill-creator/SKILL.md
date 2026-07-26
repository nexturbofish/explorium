---
name: skill-creator
description: Create new skills, modify existing ones, and measure skill performance. Use whenever the user wants to author a new skill, turn a recurring workflow into a reusable skill, or asks "how do I make a skill". Walks intent capture, drafting, and iteration.
always_active: false
---

# Skill Creator

This skill teaches you to author, evaluate, and iterate skills end-to-end. A "skill" lives at `~/.hi/skills/<name>/SKILL.md` and is a chunk of expertise the agent loads when a similar task shows up.

## Creating a skill

### Step 1: Capture intent

Ask:
- What task should this skill handle? Be specific.
- What should trigger this skill? (keywords, patterns)
- Should it always be active, or only on matching triggers?

### Step 2: Set frontmatter

Every SKILL.md starts with YAML frontmatter:

```
---
name: my-skill
description: Does X for Y
triggers: ["keyword1", "keyword2"]
always_active: false
version: 1.0
author: user
scope: user
---
```

| Field | Required | Purpose |
|-------|----------|---------|
| `name` | yes | Slug used for directory name and identification |
| `description` | yes | First line of discovery — appears in search results |
| `triggers` | no | Words/patterns that activate this skill |
| `always_active` | no | If true, body is always injected into system prompt |
| `version` | no | Track iterations |
| `author` | no | Who created it |
| `scope` | no | `user` or `project` |

Good descriptions are concrete:
- Bad: "Helps with coding"
- Good: "Review pull requests for security vulnerabilities in Python code"

Triggers should be terms the user would naturally say:
- `["review", "audit", "check for bugs"]` for a code review skill

### Step 3: Write the body

The body is a Markdown document that tells the agent what to do. Structure:

```
# Skill Title

Brief context — when and why this skill matters.

## When to use

Specific conditions that should activate this skill.

## Workflow

Step-by-step instructions.

## Key details

Important nuances, edge cases, or rules.

## Examples

Concrete examples of what to do and what to avoid.
```

### Step 4: Create the file

The agent uses `skill_create` tool which writes the SKILL.md atomically.

## Best practices

1. **Single responsibility** — One skill should handle one kind of task
2. **Be specific** — Concrete instructions beat vague suggestions
3. **Define triggers** — Good triggers make the skill discoverable
4. **Test** — Try the skill and refine
5. **Iterate** — Update based on real usage
