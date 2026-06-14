---
name: doc-sync-check
description: "Documentation synchronization checklist for API, configuration, and feature changes. Invoked by post-dev-workflow when code affects documentation. Not for direct user invocation."
version: "1.0.0"
author: "SublinkPro Team"
user-invocable: false
---

# Documentation Sync Check Skill

Checklist for ensuring documentation stays synchronized with code changes.

## When to use this skill

Use this skill when changes affect:
- User-visible behavior or features
- API contracts or endpoints
- Configuration options or environment variables
- Deployment or installation procedures
- Developer workflows or contribution processes
- Build or validation processes

## Core Principle

Documentation must be updated **in the same PR** as the code change. Outdated docs are a maintenance burden and user pain point.

## Prerequisites

Before running this checklist:
1. Identify which documentation files might be affected
2. Read AGENTS.md section "Documentation expectations"
3. Review CONTRIBUTING.md for documentation standards

## Quick Reference: Which Docs to Update

| Change Type | Docs to Check |
|---|---|
| User-facing feature | `README.md` + `.zh-CN.md`, `docs/features/*.md` + `.zh-CN.md` |
| Configuration option | `docs/configuration.md` + `.zh-CN.md`, example configs |
| Installation/deployment | `docs/installation.md` + `.zh-CN.md`, `skill-sublinkpro/reference/deploy.md` |
| API endpoint | `skill-sublinkpro/reference/api.md` |
| Developer workflow | `docs/development.md` + `.zh-CN.md`, `CONTRIBUTING.md` + `.zh-CN.md` |
| Architecture/guidelines | `AGENTS.md` |
| New documentation | `skill-sublinkpro/reference/docs.md` (documentation map) |

## Feature Documentation Checklist

### When user-facing features change:

#### README updates
- [ ] Updated `README.md` if feature overview affected
- [ ] Updated `README.zh-CN.md` (Chinese version)
- [ ] Updated feature list if new feature added
- [ ] Updated screenshots/GIFs if UI changed significantly

#### Feature docs
- [ ] Updated relevant `docs/features/*.md` file
- [ ] Updated corresponding `.zh-CN.md` file
- [ ] Added new feature doc if new feature (both languages)
- [ ] Updated examples and usage instructions
- [ ] Verified code samples are accurate

#### Cross-references
- [ ] Updated links from README to feature docs
- [ ] Updated internal cross-references in docs
- [ ] Added new doc to `docs/` index/TOC (if applicable)

## Configuration Documentation Checklist

### When configuration options change:

#### Configuration guide
- [ ] Updated `docs/configuration.md`
- [ ] Updated `docs/configuration.zh-CN.md`
- [ ] Documented new environment variables
- [ ] Documented new config file options
- [ ] Updated config precedence order (if changed)
- [ ] Updated default values

#### Example files
- [ ] Updated `.env.example` (if env vars changed)
- [ ] Updated `config.example.yaml` (if config file changed)
- [ ] Updated `docker-compose.example.yml` (if Docker config changed)

#### Deployment docs
- [ ] Updated `skill-sublinkpro/reference/deploy.md` if deployment-related
- [ ] Updated Docker/install instructions if needed

## API Documentation Checklist

### When API endpoints change:

#### Skill API reference
- [ ] Updated `skill-sublinkpro/reference/api.md`
- [ ] Documented new endpoints
- [ ] Updated changed endpoints
- [ ] Marked deprecated endpoints
- [ ] Updated request/response examples
- [ ] Updated error codes and messages
- [ ] Updated authentication requirements

#### Skill workflows
- [ ] Updated `skill-sublinkpro/SKILL.md` if workflow affected
- [ ] Updated usage examples
- [ ] Updated error handling examples

#### Skill README
- [ ] Updated `skill-sublinkpro/README.md` if needed
- [ ] Updated `skill-sublinkpro/README.zh-CN.md`

## Installation/Deployment Documentation Checklist

### When installation or deployment changes:

#### Installation guide
- [ ] Updated `docs/installation.md`
- [ ] Updated `docs/installation.zh-CN.md`
- [ ] Updated Docker installation steps
- [ ] Updated docker-compose instructions
- [ ] Updated one-line script instructions
- [ ] Updated update/upgrade instructions

#### Deployment guide
- [ ] Updated `skill-sublinkpro/reference/deploy.md`
- [ ] Updated environment variable setup
- [ ] Updated port mappings
- [ ] Updated volume mounts
- [ ] Updated default credentials
- [ ] Updated security recommendations

#### Build process
- [ ] Updated production build instructions (if changed)
- [ ] Updated CI/CD documentation (if changed)

## Developer Documentation Checklist

### When developer workflows change:

#### Development guide
- [ ] Updated `docs/development.md`
- [ ] Updated `docs/development.zh-CN.md`
- [ ] Updated local setup instructions
- [ ] Updated validation commands
- [ ] Updated testing instructions
- [ ] Updated protocol extension guide (if protocol system changed)

#### Contributing guide
- [ ] Updated `CONTRIBUTING.md`
- [ ] Updated `CONTRIBUTING.zh-CN.md`
- [ ] Updated branch conventions
- [ ] Updated PR process
- [ ] Updated testing requirements

#### Architecture guide
- [ ] Updated `AGENTS.md` if architectural boundaries changed
- [ ] Updated project structure documentation
- [ ] Updated cross-layer sync requirements (if changed)

#### Theme/i18n guidelines
- [ ] Updated `docs/frontend-theme-guidelines.md` + `.zh-CN.md` (if theme rules changed)
- [ ] Updated `docs/internationalization.md` + `.zh-CN.md` (if i18n rules changed)

## Bilingual Documentation Checklist

### For every documentation change:

#### Both languages
- [ ] Updated English canonical file (`.md`)
- [ ] Updated Chinese translation file (`.zh-CN.md`)
- [ ] Content semantically equivalent (not just machine-translated)
- [ ] Examples work in both language contexts

#### Language switches
- [ ] Language switch links at top of document work
- [ ] Links use correct format: `English | [简体中文](filename.zh-CN.md)`
- [ ] Chinese version links back: `[English](filename.md) | 简体中文`

#### Internal links
- [ ] Relative links work in both language files
- [ ] Cross-references point to correct language variants
- [ ] No broken links in either version

## Documentation Map Updates

### When documentation structure changes:

#### Adding new docs
- [ ] Added new doc path to `skill-sublinkpro/reference/docs.md`
- [ ] Added description/topic in the documentation map
- [ ] Added both English and Chinese variants to map

#### Renaming/moving docs
- [ ] Updated path in `skill-sublinkpro/reference/docs.md`
- [ ] Updated all references in other docs
- [ ] Verified old paths return 404 (on GitHub)

#### Removing docs
- [ ] Removed from `skill-sublinkpro/reference/docs.md`
- [ ] Checked for incoming links from other docs
- [ ] Redirected or updated incoming links

## Code Examples in Documentation

### When documenting code:

#### Accuracy
- [ ] Code examples actually work
- [ ] Commands match what exists in repo
- [ ] File paths are correct
- [ ] Configuration examples are valid

#### Completeness
- [ ] Required context is provided
- [ ] Output examples are realistic
- [ ] Error handling shown where relevant

#### Maintenance
- [ ] Examples use current API
- [ ] Examples use current configuration format
- [ ] Examples reflect current project structure

## Common Documentation Files

### Project-level
- `README.md` / `README.zh-CN.md` - Project overview
- `CONTRIBUTING.md` / `CONTRIBUTING.zh-CN.md` - How to contribute
- `CODE_OF_CONDUCT.md` / `CODE_OF_CONDUCT.zh-CN.md` - Community standards
- `AGENTS.md` - AI agent architectural guide

### Core documentation
- `docs/installation.md` / `.zh-CN.md` - Installation guide
- `docs/configuration.md` / `.zh-CN.md` - Configuration reference
- `docs/development.md` / `.zh-CN.md` - Developer guide
- `docs/internationalization.md` / `.zh-CN.md` - i18n guidelines
- `docs/frontend-theme-guidelines.md` / `.zh-CN.md` - Theme adaptation rules

### Feature documentation
- `docs/features/*.md` / `.zh-CN.md` - Feature-specific guides
- `docs/script_support.md` / `.zh-CN.md` - Script support documentation

### Skill documentation
- `skill-sublinkpro/SKILL.md` - Skill definition
- `skill-sublinkpro/README.md` / `.zh-CN.md` - Skill overview
- `skill-sublinkpro/reference/api.md` - API reference
- `skill-sublinkpro/reference/deploy.md` - Deployment guide
- `skill-sublinkpro/reference/docs.md` - Documentation map

## Verification Checklist

### Before committing:

#### Link checking
- [ ] All internal links work (test by clicking in preview)
- [ ] No broken references to code files
- [ ] No references to removed/renamed files
- [ ] Language switch links work

#### Formatting
- [ ] Markdown renders correctly
- [ ] Code blocks have correct syntax highlighting
- [ ] Tables render properly
- [ ] Lists and indentation correct

#### Consistency
- [ ] Both language versions have same structure
- [ ] Terminology consistent across docs
- [ ] Examples match actual codebase
- [ ] Commands match repo's actual commands

## When Only Documentation Changed

### Pure documentation commits:

#### No build required
- ✅ No need to run `yarn run lint` or `yarn run build`
- ✅ No need to run Go validation

#### Manual verification required
- [ ] Links work
- [ ] Both languages updated
- [ ] Command examples accurate
- [ ] No conflicts with AGENTS.md cross-layer rules

## Documentation Anti-Patterns

Avoid these common mistakes:

- ❌ Documenting commands that don't exist (`yarn test` when no test script)
- ❌ Only updating English docs
- ❌ Describing old behavior after code changed
- ❌ Copy-pasting from generic templates without adapting
- ❌ Leaving broken links
- ❌ Documenting "future features" not yet implemented
- ❌ Outdated screenshots showing old UI

## Delivery Requirements

Before marking documentation changes complete:

1. **Bilingual verification**: Both `.md` and `.zh-CN.md` updated
2. **Link verification**: All links tested and working
3. **Accuracy verification**: Examples/commands tested
4. **Consistency verification**: No conflicts with code behavior

## Exit Criteria

✅ Can exit when:
- All affected documentation files updated
- Both English and Chinese versions synchronized
- Links verified
- Examples tested
- Code behavior matches documentation

❌ Cannot exit when:
- Only one language updated
- Documentation still describes old behavior
- Broken links remain
- Examples are inaccurate
- New documentation not added to map
