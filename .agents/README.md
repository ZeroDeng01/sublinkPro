# Agent Skills System

## Overview

This repository uses a structured agent skills system for operational checklists and validation procedures. Skills are reusable, focused guides that agents can invoke for specific tasks.

## Directory Structure

```
.agents/
└── skills/
    ├── post-dev-workflow/
    │   └── SKILL.md                # MANDATORY post-development workflow
    ├── pre-commit-check/
    │   └── SKILL.md                # Pre-commit validation checklist
    ├── cross-layer-sync/
    │   └── SKILL.md                # Cross-layer synchronization guide
    ├── doc-sync-check/
    │   └── SKILL.md                # Documentation sync checklist
    ├── theme-check/
    │   └── SKILL.md                # Theme adaptation checklist
    ├── security-review/
    │   └── SKILL.md                # Security review checklist
    └── performance-check/
        └── SKILL.md                # Performance check checklist
```

## Accessing Skills

Different agent systems access skills through symlinks:

- **Claude Code**: `.claude/skills` → `../.agents/skills`
- **Codex**: `.codex/skills` → `../.agents/skills` (when added)
- **Other agents**: Create similar relative symlinks

## Available Skills

### 1. Post-Development Workflow (`post-dev-workflow/SKILL.md`) ⭐ MANDATORY

**Use when**: After completing ANY code change (backend, frontend, or both)

**CRITICAL FOR AI AGENTS**: This is the master workflow that MUST be automatically triggered after every code change. It orchestrates all validation, synchronization, and documentation phases.

**Covers**:
- Phase 1: Code validation (lint, format, build, test)
- Phase 2: Cross-layer sync check (if applicable)
- Phase 3: Documentation sync check (if applicable)
- Phase 4: Test execution (if applicable)
- Phase 5: Change summary preparation

**Exit criteria**: All phases pass, work is ready to commit

**Relationship to other skills**: This skill orchestrates and invokes the other skills (pre-commit-check, cross-layer-sync, doc-sync-check) as needed. It is the single entry point for "code complete" workflow.

### 2. Pre-Commit Check (`pre-commit-check/SKILL.md`)

**Use when**: Before every commit or PR (also invoked automatically by post-dev-workflow Phase 1)

**Covers**:
- Backend validation (gofmt, golangci-lint, go test)
- Frontend validation (yarn lint, yarn build)
- i18n validation
- Theme changes validation
- Cross-layer validation
- Git commit checklist

**Exit criteria**: All validation commands pass, staging is clean

### 3. Cross-Layer Sync (`cross-layer-sync/SKILL.md`)

**Use when**: Changes affect multiple layers (backend ↔ frontend ↔ docs) (also invoked automatically by post-dev-workflow Phase 2)

**Covers**:
- Backend → Frontend sync
- Frontend → Backend sync
- Configuration sync
- Documentation sync
- API contract updates
- Skill API reference updates

**Exit criteria**: All impacted layers synchronized, verification complete

### 4. Doc Sync Check (`doc-sync-check/SKILL.md`)

**Use when**: Code changes affect documentation (also invoked automatically by post-dev-workflow Phase 3)

**Covers**:
- Feature documentation
- Configuration documentation
- API documentation
- Installation/deployment docs
- Developer documentation
- Bilingual consistency
- Documentation map updates

**Exit criteria**: All docs updated, both languages synchronized, links verified

### 5. Theme Check (`theme-check/SKILL.md`)

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

### 6. Security Review (`security-review/SKILL.md`)

**Use when**: Making changes to authentication, authorization, sensitive data handling, or security-critical features

**Covers**:
- Authentication & authorization checks
- MFA (Multi-Factor Authentication) security
- Sensitive data handling (passwords, tokens, API keys)
- Input validation & sanitization (SQL injection, XSS, path traversal)
- Database security (parameterized queries, mass assignment)
- API security (CORS, rate limiting, error handling)
- Cryptography (strong algorithms, secure random)
- File upload/download security
- Session management
- Dependency security scanning

**Exit criteria**: All security checks pass, no sensitive data exposed, security linters pass, security implications documented

### 7. Performance Check (`performance-check/SKILL.md`)

**Use when**: Making changes that may impact performance, scalability, or resource usage

**Covers**:
- Database query optimization (N+1 problems, indexes, pagination)
- API response optimization (DTOs, compression, caching)
- Frontend rendering optimization (memo, virtualization, debouncing)
- Caching strategy (cache keys, TTL, invalidation)
- Concurrency & parallelization (goroutines, race conditions)
- Memory management (leaks, allocations, streaming)
- Network optimization (connection pooling, timeouts)
- Algorithm complexity (time/space complexity)
- Performance testing (benchmarks, load tests, Lighthouse)

**Exit criteria**: No N+1 queries, large lists paginated/virtualized, benchmark tests pass, performance implications documented

## How to Use Skills

### For AI Agents

**CRITICAL**: The `post-dev-workflow` skill is MANDATORY and must be automatically triggered after completing ANY code change. It is the master workflow that orchestrates all other skills.

#### Automatic Workflow (Post-Development)

After completing code changes:

1. **Automatically invoke** `.agents/skills/post-dev-workflow/SKILL.md`
2. The post-dev-workflow will **automatically orchestrate**:
   - Phase 1: Code validation (pre-commit-check)
   - Phase 2: Cross-layer sync (if applicable)
   - Phase 3: Documentation sync (if applicable)
   - Phase 4: Test execution (if applicable)
   - Phase 5: Change summary preparation
3. **Report results** to the user
4. **Only declare work complete** when all phases pass

Example flow:
```
Code changes complete
    ↓
Automatically invoke: .agents/skills/post-dev-workflow/SKILL.md
    ↓
Run Phase 1: Code Validation
    → Backend: gofmt ✅, golangci-lint ✅, go test ✅
    → Frontend: yarn lint ✅, yarn build ✅
    ↓
Run Phase 2: Cross-Layer Sync (backend API changed)
    → Frontend updated ✅
    → skill-sublinkpro/reference/api.md updated ✅
    ↓
Run Phase 3: Documentation Sync
    → docs/configuration.md + .zh-CN.md updated ✅
    ↓
Run Phase 4: Test Execution
    → Added tests for new logic ✅
    → All tests passing ✅
    ↓
Run Phase 5: Change Summary
    → Commit message prepared ✅
    ↓
Report to user: "Work is ready to commit"
```

#### Manual Skill Invocation (Special Cases)

For specialized tasks, you may also invoke individual skills directly:

```
Task: Update theme colors for dark mode
→ Invoke: .agents/skills/theme-check/SKILL.md
→ Follow: Mode coverage, device coverage, component structure checks
→ Verify: All checklist items complete
→ Document: What was checked, what patterns were used
```

But even after specialized skills, you must still run the post-dev-workflow before declaring work complete.

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
- **MANDATORY post-development workflow instruction for AI agents**

**Skills** contain:
- Operational checklists
- Validation procedures
- Step-by-step guides
- Detailed verification steps

Think of AGENTS.md as "why and what" (principles), and skills as "how" (procedures).

**The post-dev-workflow skill** bridges the two: it's mandated by AGENTS.md and implemented as a skill that orchestrates all other skills.

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
