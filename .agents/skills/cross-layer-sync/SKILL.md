---
name: cross-layer-sync
description: "Cross-layer synchronization guide for backend, frontend, and documentation changes. Invoked by post-dev-workflow when changes affect multiple layers. Not for direct user invocation."
version: "1.0.0"
author: "SublinkPro Team"
user-invocable: false
---

# Cross-Layer Synchronization Check Skill

Checklist for ensuring all impacted layers are updated together when changes affect multiple parts of the application.

## When to use this skill

Use this skill when changes affect:
- API contracts, routes, request/response structures
- Frontend-backend data flow
- Configuration semantics
- Documentation that reflects runtime behavior
- User-facing text or error messages
- Deployment or build processes

## Core Principle

**Changes must be atomic across layers**: When a change affects multiple layers (backend, frontend, docs, skill API reference), all impacted layers must be updated in the same PR/commit.

## Prerequisites

Before running this checklist:
1. Identify which layers your change touches
2. Read AGENTS.md section 9 "Cross-layer synchronization"
3. Review CONTRIBUTING.md for cross-layer sync requirements

## Backend → Frontend Sync Checklist

### When backend changes affect:

#### API contracts
- [ ] Updated route paths in frontend (`webs/src/api/`)
- [ ] Updated request payload structures
- [ ] Updated response handling
- [ ] Updated error handling
- [ ] Updated TypeScript types (if applicable)

#### Fields
- [ ] Updated field names in frontend requests
- [ ] Updated field validation
- [ ] Updated form components (`webs/src/views/`)
- [ ] Updated display logic

#### Permissions
- [ ] Updated permission checks in frontend
- [ ] Updated UI visibility logic
- [ ] Updated route guards

#### Response structures
- [ ] Updated response parsing
- [ ] Updated state management
- [ ] Updated display components
- [ ] Updated loading/error states

#### Task results or notifications
- [ ] Updated task result rendering
- [ ] Updated notification display
- [ ] Added i18n keys if needed (`webs/src/i18n/locales/`)

#### Skill API (if `/api/v1/*` or `/c/*` changed)
- [ ] Updated `skill-sublinkpro/reference/api.md` endpoint documentation
- [ ] Updated `skill-sublinkpro/SKILL.md` workflow descriptions (if affected)
- [ ] Updated request/response examples
- [ ] Updated error handling examples

## Frontend → Backend Sync Checklist

### When frontend changes affect:

#### API dependencies
- [ ] Verified backend endpoint still matches
- [ ] Verified request payload structure
- [ ] Verified authentication requirements

#### Field semantics
- [ ] Verified backend accepts new field values
- [ ] Verified backend validation rules
- [ ] Verified database schema supports changes

#### Page flows
- [ ] Verified backend supports new flow sequence
- [ ] Verified state transitions are valid
- [ ] Verified all required data is available

#### i18n keys
- [ ] Backend provides `i18nKey` + `i18nParams` (if Web UI display)
- [ ] Both `zh-CN` and `en-US` translations added

## Configuration → All Layers Sync Checklist

### When configuration changes affect:

#### Environment variables
- [ ] Updated `docs/configuration.md` + `.zh-CN.md`
- [ ] Updated `.env.example` or `config.example.yaml`
- [ ] Updated Docker configuration examples
- [ ] Updated deployment documentation (`skill-sublinkpro/reference/deploy.md`)
- [ ] Updated code that reads the variable

#### Default values
- [ ] Updated code
- [ ] Updated documentation
- [ ] Updated example files

#### Config precedence
- [ ] Updated `docs/configuration.md`
- [ ] Verified new precedence order works correctly

## Documentation Sync Checklist

### When behavior changes require doc updates:

#### User-facing features
- [ ] Updated `README.md` + `README.zh-CN.md` (if feature overview affected)
- [ ] Updated relevant `docs/features/*.md` + `.zh-CN.md`
- [ ] Updated screenshots/examples if needed

#### Developer workflows
- [ ] Updated `docs/development.md` + `.zh-CN.md`
- [ ] Updated `CONTRIBUTING.md` + `.zh-CN.md` (if contribution flow affected)

#### Configuration or deployment
- [ ] Updated `docs/configuration.md` + `.zh-CN.md`
- [ ] Updated `docs/installation.md` + `.zh-CN.md`
- [ ] Updated `skill-sublinkpro/reference/deploy.md`

#### API contracts
- [ ] Updated `skill-sublinkpro/reference/api.md`
- [ ] Updated API examples and error codes

#### Documentation map
- [ ] Updated `skill-sublinkpro/reference/docs.md` if new docs added or paths changed

### Bilingual consistency
- [ ] Updated both English and `*.zh-CN.md` versions
- [ ] Verified language switch links work
- [ ] Verified relative links are consistent

## Verification Checklist

### Code verification
- [ ] **Frontend**: Ran `yarn run lint` (in `webs/` directory)
- [ ] **Frontend build**: Ran `yarn run build` (if routing, assets, or build affected)
- [ ] **Backend format**: Ran `gofmt -w <changed-files>`
- [ ] **Backend lint**: Ran `golangci-lint run`
- [ ] **Backend tests**: Ran relevant `go test ./...`

### Manual verification
- [ ] Tested the change end-to-end in local environment
- [ ] Verified both light and dark modes (if UI change)
- [ ] Verified desktop and mobile (if UI change)
- [ ] Checked browser console for errors
- [ ] Verified API responses match frontend expectations

## When one layer doesn't need changes

If you've checked an impacted layer and confirmed it truly doesn't need changes, document this in your change summary:

**Example**: "Checked frontend API layer (`webs/src/api/subscriptions.js`) - no changes needed because new backend field is optional and frontend uses default behavior."

## Common Cross-Layer Change Patterns

### Adding a new API endpoint
1. Backend: Add route, handler, validation
2. Frontend: Add API client function (`webs/src/api/`)
3. Frontend: Add UI component/view (`webs/src/views/`)
4. i18n: Add translations for UI text
5. Docs: Update `skill-sublinkpro/reference/api.md`
6. Tests: Add backend handler tests

### Changing a data model field
1. Backend: Update model, migration, validation
2. Backend: Update API handlers
3. Frontend: Update API calls
4. Frontend: Update display components
5. Frontend: Update forms/inputs
6. Docs: Update feature documentation

### Modifying configuration behavior
1. Backend: Update config parsing
2. Backend: Update runtime behavior
3. Docs: Update `docs/configuration.md` + `.zh-CN.md`
4. Docs: Update example configs
5. Docs: Update `skill-sublinkpro/reference/deploy.md`
6. Tests: Add config validation tests

## Delivery Requirements

Before marking complete, the change summary must include:

1. **Layers changed**: List which layers were modified
2. **Layers checked**: List which layers were inspected but didn't need changes
3. **Justification**: For unchanged layers, briefly explain why no sync was needed
4. **Verification**: List which validation commands were run

## Anti-patterns to avoid

- ❌ Changing backend API without updating frontend
- ❌ Changing frontend behavior without verifying backend support
- ❌ Updating code without updating documentation
- ❌ Updating only English docs without Chinese versions
- ❌ Claiming "no sync needed" without actually checking

## Exit criteria

✅ Can exit when:
- All impacted layers are synchronized
- Verification commands pass
- Documentation is updated (both languages)
- Change summary documents what was checked

❌ Cannot exit when:
- Code changed but docs still describe old behavior
- Backend changed but frontend still uses old contract
- Frontend changed but backend compatibility not verified
- One language docs updated but not the other
