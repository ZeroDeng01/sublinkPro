# Agent Skills System

## Overview

This repository uses a structured agent skills system for operational checklists and validation procedures. Skills are reusable, focused guides that agents can invoke for specific tasks.

## Directory Structure

```
.agents/
└── skills/
    ├── theme-check/
    │   └── SKILL.md                # Theme adaptation checklist
    ├── cross-layer-sync/
    │   └── SKILL.md                # Cross-layer synchronization guide
    ├── pre-commit-check/
    │   └── SKILL.md                # Pre-commit validation checklist
    └── doc-sync-check/
        └── SKILL.md                # Documentation sync checklist
```

## Accessing Skills

Different agent systems access skills through symlinks:

- **Claude Code**: `.claude/skills` → `../.agents/skills`
- **Codex**: `.codex/skills` → `../.agents/skills` (when added)
- **Other agents**: Create similar relative symlinks

## Available Skills

### 1. Theme Check (`theme-check/SKILL.md`)

**Use when**: Modifying any UI colors, surfaces, dialogs, panels, or theme infrastructure

**Covers**:
- Light/dark mode verification
- Desktop/mobile coverage
- Component structure checks
- Text hierarchy validation
- Interactive states
- Surface layering
- Reference pattern usage

**Exit criteria**: Both modes tested, all variants checked, delivery summary complete

### 2. Cross-Layer Sync (`cross-layer-sync/SKILL.md`)

**Use when**: Changes affect multiple layers (backend ↔ frontend ↔ docs)

**Covers**:
- Backend → Frontend sync
- Frontend → Backend sync
- Configuration sync
- Documentation sync
- API contract updates
- Skill API reference updates

**Exit criteria**: All impacted layers synchronized, verification complete

### 3. Pre-Commit Check (`pre-commit-check/SKILL.md`)

**Use when**: Before every commit or PR

**Covers**:
- Backend validation (gofmt, golangci-lint, go test)
- Frontend validation (yarn lint, yarn build)
- i18n validation
- Theme changes validation
- Cross-layer validation
- Git commit checklist

**Exit criteria**: All validation commands pass, staging is clean

### 4. Doc Sync Check (`doc-sync-check/SKILL.md`)

**Use when**: Code changes affect documentation

**Covers**:
- Feature documentation
- Configuration documentation
- API documentation
- Installation/deployment docs
- Developer documentation
- Bilingual consistency
- Documentation map updates

**Exit criteria**: All docs updated, both languages synchronized, links verified

## How to Use Skills

### For AI Agents

When working on a task that matches a skill's domain:

1. **Identify the relevant skill** based on the task type
2. **Read the skill file** to understand requirements
3. **Follow the checklist** systematically
4. **Document completion** in your change summary

Example flow:
```
Task: Update theme colors for dark mode
→ Invoke: .agents/skills/theme-check/SKILL.md
→ Follow: Mode coverage, device coverage, component structure checks
→ Verify: All checklist items complete
→ Document: What was checked, what patterns were used
```

### For Human Contributors

Skills are also useful for manual reviews:

1. Find the relevant skill in `.agents/skills/`
2. Use it as a checklist before committing
3. Ensure all items are addressed

## Skill Structure

Each skill follows this format:

```markdown
# [Skill Name]

## When to use this skill
Clear trigger conditions

## Prerequisites
What to know or prepare before starting

## Checklist
Detailed, actionable items with [ ] checkboxes

## Delivery Requirements
What must be documented

## Anti-patterns to avoid
Common mistakes

## Exit criteria
When you can mark the work complete
```

## Adding New Skills

To add a new skill:

1. **Create directory**: `.agents/skills/your-skill-name/`
2. **Create SKILL.md**: Follow the structure above
3. **Document triggers**: Clear "when to use" section
4. **Provide checklist**: Actionable items with checkboxes
5. **Define exit criteria**: Clear completion conditions
6. **Update this README**: Add to "Available Skills" section

## Skills vs AGENTS.md

**AGENTS.md** contains:
- Architectural principles
- Core guidelines
- Project structure
- Key integration points
- Source of truth hierarchy

**Skills** contain:
- Operational checklists
- Validation procedures
- Step-by-step guides
- Detailed verification steps

Think of AGENTS.md as "why and what" (principles), and skills as "how" (procedures).

## Benefits

1. **Focused**: Each skill addresses one specific concern
2. **Reusable**: Same checklist for every similar task
3. **Maintainable**: Update one skill without touching core docs
4. **Discoverable**: Clear naming makes skills easy to find
5. **Portable**: Skills work across different agent systems

## Related Documentation

- **AGENTS.md**: Core architectural guidelines
- **CONTRIBUTING.md**: General contribution workflow
- **docs/development.md**: Development setup and standards
- **docs/frontend-theme-guidelines.md**: Detailed theme rules (referenced by theme-check skill)
