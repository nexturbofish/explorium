---
name: find-skills
description: Find and install an existing skill from the open agent ecosystem. Use this skill whenever the user asks "is there a skill for X", "find a skill for X", or "install a skill that does Y". Walks the agent through searching, vetting candidates, presenting the SKILL.md for review, and confirming the install.
always_active: false
---

# Finding and Installing Skills

This skill helps you discover and install skills. A "skill" is a chunk of expertise that the agent loads on demand. Skills live at `~/.hi/skills/<name>/SKILL.md`.

## When to use

Trigger this skill when the user:
- Asks "is there a skill for X" / "find a skill for Y"
- Asks "can you do X" where X is a specialized capability (code review, testing, deployment, etc.)
- Mentions they wish they had help with a specific domain
- Wants to extend the agent with a packaged workflow

## Workflow

### Step 1: Understand the need

Identify:
1. **Domain** — testing, deployment, design, docs, etc.
2. **Specific task** — "review PRs for security", "optimize performance"
3. **Whether it's a common task** — if so, a skill probably exists; if not, suggest creating one (see the `skill-creator` skill).

### Step 2: Search

Use `web_fetch` to search for skills on GitHub or the web. Good starting points:
- Search for "awesome-claude-skills" or "agent-skills" lists
- GitHub search for `SKILL.md` files in popular repositories
- Check known collections: `vercel-labs/skills`, `anthropics/skills`

### Step 3: Vet candidates

For each candidate:
1. **Source reputation** — Official sources carry more weight than unknown authors
2. **Read the SKILL.md** — Present the full SKILL.md to the user verbatim for review

### Step 4: Present to the user

After vetting, summarize for the user:
- What the skill does
- Where it comes from
- Read the full SKILL.md aloud and wait for confirmation

### Step 5: Install

Once confirmed, call `skill_install` with the source. The `install` tool supports:
- `github:owner/repo/path` — GitHub repo with optional path
- `owner/repo@skill-name` — shorthand format

Tell the user what landed (the tool returns `files_written` list).

## Common categories

| Category | Example queries |
| --- | --- |
| Code quality | review, lint, refactor, security |
| Testing | test, e2e, unit testing |
| DevOps | docker, kubernetes, deployment |
| Documentation | readme, changelog, api-docs |
| Design | ui, ux, accessibility |

## When nothing matches

If the search comes up empty, offer to help directly and suggest using the `skill-creator` skill to author one if it's a recurring need.
