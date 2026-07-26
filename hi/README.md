# Hi — A Self-Evolving AI Agent

**Hi** is a Go-based AI agent CLI that learns from every conversation. It extracts memories (facts, preferences) and skills (reusable procedures) through reflection, persists them across sessions, and feeds them back into future interactions — making it smarter over time.

Built on [cloudwego/eino](https://github.com/cloudwego/eino), the LLM orchestration framework.

## Why "Hi"?

Simple — that's how you greet your AI agent.

## Features

- **Interactive REPL** — raw terminal mode with multi-line editing, tab completion, slash commands
- **Autonomous agent** — goal-driven execution with tool-use loop (`/run`)
- **Three-tier reflection** — micro (per-turn), full (end-of-session), focused (by topic)
- **Memory system** — persistent markdown files with YAML frontmatter, TF-IDF search, zone-based palace
- **Skills system** — reusable markdown procedures, git-based install, bundled skills
- **Tool system** — 20+ built-in tools (bash, file ops, web search, git, sub-agent) + MCP server integration
- **Guard/audit** — rate limiting, permission globs, confirmation queues, rotating audit log
- **Robust LLM parsing** — handles code fences, leading prose, truncated JSON from model output

## Architecture

```
cmd/main.go → internal/app/cli.go → CLI commands
                                        │
                              ┌─────────┴──────────┐
                              │    AppState         │
                              │  (wiring hub)       │
                              └──┬──┬──┬──┬──┬──┬──┘
                                 │  │  │  │  │  │
            ┌────────────────────┘  │  │  │  │  └──────────┐
            ▼                       ▼  ▼  ▼  ▼             ▼
       Turn Runner            Config  Memory Skills   Reflector
       (tool loop)                     Store Store   (micro/full/
            │                            │     │      focused)
            ▼                            │     │
    ┌───────────────┐                    │     │
    │  Guard/Policy  │                   │     │
    │  (rate, perm,  │                   │     │
    │   confirm,     │                   │     │
    │   audit)       │                   │     │
    └───────┬───────┘                    │     │
            ▼                            ▼     ▼
    Tool Executor                   Memory Files   Skill Files
   (builtin + MCP)                  ~/.hi/memories  ~/.hi/skills
```

### Key Abstractions

| Layer | Package | Role |
|---|---|---|
| Entry point | `cmd/main.go` | Calls `app.Run()` |
| CLI wiring | `internal/app/` | Cobra root command + subcommands |
| State hub | `internal/appstate/` | Creates all subsystems, holds references |
| Turn execution | `internal/turn/` | User input → tool loop → response |
| Memory | `internal/memory/` | CRUD, TF-IDF search, palace, relevance |
| Skills | `internal/skills/` | CRUD, search, git install, bundled |
| Reflection | `internal/reflect/` | Micro/full/focused + log store + deferred queue |
| Tools | `internal/tools/` | 20+ built-in tools + MCP adapter |
| Guard | `internal/guard/` | Policy engine, audit log, rate limiter, confirm |
| Agent | `internal/agent/` | Autonomous goal-driven loop |
| Config | `internal/config/` | YAML config at `~/.hi/config.yaml` |
| Types | `spec/` | Core interfaces and data types |

## Reflection System

The reflection system is how Hi learns. It analyzes conversations and extracts actionable knowledge.

### Three Modes

| Mode | Scope | Triggers |
|---|---|---|
| **Micro** | Last 1-3 turns | Every 3 turns, or on explicit teaching intent (keywords: "remember", "always", "prefer", Chinese equivalents) |
| **Full** | Entire session | End-of-session or manual |
| **Focused** | Topic-filtered session | On-demand with a topic |

All modes produce the same output schema:

```json
{
  "summary": "what happened",
  "skills": [{ "name": "...", "triggers": [...], "body": "...", "confidence": "low|medium|high" }],
  "memorise": [{ "content": "...", "zone": "general", "confidence": "..." }],
  "conflicts": [{ "existing_id": "...", "new_content": "...", "resolution": "keep_new|keep_old|merge" }]
}
```

### Robust JSON Parsing

`robustParse()` in `internal/reflect/parse.go` handles:
- Strips ` ```json ` code fences
- Strips leading prose (finds first `{`)
- Repairs truncated JSON (counts braces/brackets, appends missing closures)

### Feedback Loop

Reflection outcomes are logged to `~/.hi/logs/reflect.jsonl`. Recent outcomes (accept/reject/defer decisions) are injected into subsequent reflection prompts, so the model learns from past decisions.

### Commands

- `/reflect outcomes [n]` — view recent reflection log entries
- `/reflect deferred` — list deferred candidates (low-confidence proposals not yet accepted)

## Memory System

Memories persist as markdown files in `~/.hi/memories/<id>.md`:

```yaml
---
id: "uuid"
tags: ["tag1"]
zone: "general"
pinned: false
source: "reflection"  # or "user" or "imported"
created_at: "2024-01-01T00:00:00Z"
accessed_at: "2024-01-02T00:00:00Z"
access_count: 3
---
Content body here...
```

- **TF-IDF search** — Latin (whitespace) + CJK (char + bigram) tokenization
- **Memory Palace** — zone-based organization for system prompt injection
- **Pinned memories** — always included in the system prompt
- **Active window** — only recently accessed memories are included (configurable)
- **Conflict detection** — TF-IDF similarity >0.85 flags potential duplicates

### CLI Commands

- `/memory list` — list all memories
- `/memory show <id>` — show memory details
- `/memory delete <id>` — delete a memory
- `/memory pin <id>` / `/memory unpin <id>` — toggle pinned flag

## Skills System

Skills are reusable procedures stored as `~/.hi/skills/<slug>/SKILL.md`:

```yaml
---
name: "my-skill"
description: "Does X"
triggers: ["keyword1"]
always_active: false
---
## Title

Markdown instructions here.
```

- **Bundled skills** — 2 auto-installed at startup: `skill-creator` (meta-skill for creating skills) and `find-skills` (searches registry)
- **Git-based install** — clone from any GitHub repo
- **Triggered skills** — activated by keyword matching in user input
- **Always-active skills** — always included in system prompt

### CLI Commands

- `/skills list` — list installed skills
- `/skills show <slug>` — show skill details
- `/skills install <source>` — install from GitHub
- `/skills create <name>` — create a new skill
- `/skills delete <slug>` — delete a skill

## Tool System

### Built-in Tools (20+)

| Tool | Purpose |
|---|---|
| `bash` | Execute shell commands |
| `read` | Read files |
| `write` | Write files |
| `edit` | Structured find-and-replace edits |
| `glob` | File pattern search |
| `grep` | Content search |
| `git` | Git operations |
| `think` | Chain-of-thought scratchpad |
| `todo` | Task tracking |
| `web_fetch` | Fetch URLs |
| `web_search` | Web search (DuckDuckGo, Tavily, Brave) |
| `subagent` | Spawn sub-agents |
| `memory_search`, `memory_save`, `memory_delete` | Memory operations |
| `palace_zones`, `palace_read_zone`, `palace_recall` | Memory palace navigation |
| `skill_list`, `skill_read`, `skill_install`, `skill_delete`, `skill_create` | Skill operations |
| `propose_skill` | Propose a new skill from conversation |

### MCP Integration

Hi connects to external MCP servers defined in `~/.hi/mcp.json`:

```json
{
  "servers": {
    "officecli": {
      "transport": "stdio",
      "command": "officecli",
      "args": ["mcp"]
    }
  }
}
```

Supports stdio and SSE transports. MCP tools are added to the executor as fallback when no builtin matches.

### CLI Commands

- `/tools [n]` — show recent tool call audit log (same as `/audit`)
- `/mcp list` — list connected MCP servers
- `/mcp test [name]` — test a server's tool list

## Guard / Audit System

Every tool call goes through a policy pipeline:

1. **Rate limit** — sliding window counters per tool (bash: 20/min, write/edit/web: 30/min, default: 60/min)
2. **Permission check** — glob-pattern allow/deny lists (deny takes precedence)
3. **Confirmation check** — triggers for:
   - Writing outside workspace
   - Unsafe bash commands (not in prefix list)
   - Private/internal network targets
   - Multi-file delete (>5)
   - High token usage (>100K)
4. **Audit** — every decision logged to `~/.hi/logs/audit.jsonl` (10MB rotating)

### CLI Commands

- `/audit [n]` — show recent audit log entries

## Turn Runner

The core execution loop for each user input:

1. **Context assembly** — build system prompt with pinned memories, active memories, active skills, relevant memories
2. **User message** → append to session
3. **Tool-use loop** (max 20 rounds):
   - Call model → if no tool calls, respond
   - For each tool call → policy check → execute → append result
4. **Save session**

## Configuration

YAML config at `~/.hi/config.yaml`:

```yaml
default_provider: anthropic

providers:
  anthropic:
    base_url: https://api.anthropic.com
    api_key: "${ANTHROPIC_API_KEY}"
    model: claude-sonnet-4-20250514
    max_tokens: 8192

reflect:
  min_turns: 5
  auto_accept_memories: true
  auto_accept_min_confidence: medium

limits:
  model_limit: 32000
  headroom: 0.05
  keep_recent_turns: 4
  max_tool_rounds: 20
  agent_max_iterations: 50
  active_memory_index_cap: 100
  active_window_days: 7
  skill_index_cap: 500
  relevant_memory_cap: 5
  triggered_skill_cap: 5

guard:
  confirm_timeout: 120

permissions:
  allow:
    - "**/refact-go/**"
  deny: []

rate_limits:
  global_max_per_min: 60

web:
  search_engine: ddg  # ddg, tavily, brave
  tavily_api_key: ""
  brave_api_key: ""

workspace:
  root: ""
```

## Getting Started

```bash
# Build
cd hi && go build -o /tmp/hi ./cmd/main.go

# Initialize config
/tmp/hi init

# Start REPL
/tmp/hi chat

# Single question
/tmp/hi ask "what is the capital of France?"

# Autonomous agent
/tmp/hi run "research Go concurrency patterns and write a summary"

# Diagnostics
/tmp/hi doctor

# HTTP server
/tmp/hi serve
```

## Tech Stack

- **Go 1.26+** — compiled, single binary
- **cloudwego/eino v0.9** — LLM orchestration (tool-calling, streaming, composition)
- **eino-ext** — model providers (Claude, OpenAI, DeepSeek) + MCP adapter
- **mcp-go** — MCP protocol client
- **spf13/cobra** — CLI framework
- **gopkg.in/yaml.v3** — config file parsing
- **Go's built-in `log/slog`** — no third-party logging library
