# Documentation Synchronization Rules

Complete reference for documentation synchronization requirements in Phase 3 of the post-development workflow.

---

## When Documentation Sync is Required

Documentation synchronization is **mandatory** when changes affect:
- User-visible behavior
- API endpoints or contracts
- Configuration options
- Deployment procedures
- Developer workflows
- Architecture or design

---

## Documentation Types and Triggers

### User Documentation

| Change Type | Documentation to Update |
|---|---|
| New feature | `README.md` + `.zh-CN.md`, `docs/user/features/*.md` + `.zh-CN.md` |
| Feature behavior changed | Relevant `docs/user/features/*.md` + `.zh-CN.md` |
| Feature removed | Remove from docs, add deprecation/migration guide |
| UI changed | Update screenshots in both language versions |
| Installation method changed | `docs/user/installation.md` + `.zh-CN.md` |
| Security recommendation changed | `docs/user/security-guidelines.md` + `.zh-CN.md` |

### API Documentation

| Change Type | Documentation to Update |
|---|---|
| New `/api/v1/*` endpoint | `skills/sublinkpro/reference/api.md` |
| New `/c/*` subscription endpoint | `skills/sublinkpro/reference/api.md` |
| Endpoint path changed | `skills/sublinkpro/reference/api.md` |
| Request/response structure changed | `skills/sublinkpro/reference/api.md` - Update examples |
| Authentication changed | `skills/sublinkpro/reference/api.md` - Update auth section |
| Error codes changed | `skills/sublinkpro/reference/api.md` - Update error reference |

**Note**: `skill-sublinkpro` is an AI agent skill that consumes the REST API. Keep API reference accurate for AI agent interactions.

### Configuration Documentation

| Change Type | Documentation to Update |
|---|---|
| New config option | `docs/user/configuration.md` + `.zh-CN.md`, example configs, `skills/sublinkpro/reference/deploy.md` |
| Config option changed | `docs/user/configuration.md` + `.zh-CN.md`, migration guide |
| Environment variable added/changed | `docs/user/configuration.md` + `.zh-CN.md`, Dockerfile, docker-compose.yml |
| Default value changed | `docs/user/configuration.md` + `.zh-CN.md`, release notes (breaking change) |
| Config precedence changed | `docs/user/configuration.md` + `.zh-CN.md` |

### Deployment Documentation

| Change Type | Documentation to Update |
|---|---|
| Docker setup changed | `docs/user/installation.md` + `.zh-CN.md`, `skills/sublinkpro/reference/deploy.md` |
| docker-compose config changed | `docs/user/installation.md` + `.zh-CN.md`, example docker-compose.yml |
| One-line install script changed | `docs/user/installation.md` + `.zh-CN.md` |
| Port/path changed | `docs/user/installation.md` + `.zh-CN.md`, `docs/user/configuration.md` + `.zh-CN.md` |
| System requirements changed | `docs/user/installation.md` + `.zh-CN.md` |
| Update procedure changed | `docs/user/installation.md` + `.zh-CN.md` |

### Developer Documentation

| Change Type | Documentation to Update |
|---|---|
| Build process changed | `docs/development/getting-started.md` + `.zh-CN.md`, `docs/development/build-and-deployment.md` + `.zh-CN.md` |
| Development setup changed | `docs/development/getting-started.md` + `.zh-CN.md` |
| Testing requirements changed | `docs/development/getting-started.md` + `.zh-CN.md`, `CONTRIBUTING.md` + `.zh-CN.md` |
| Code standards changed | `docs/development/getting-started.md` + `.zh-CN.md` |
| PR process changed | `CONTRIBUTING.md` + `.zh-CN.md` |
| New protocol extension pattern | `docs/development/getting-started.md` + `.zh-CN.md`, `docs/development/practical-recipes.md` + `.zh-CN.md` |

### Architecture Documentation

| Change Type | Documentation to Update |
|---|---|
| Architecture changed | `AGENTS.md` |
| New subsystem added | `AGENTS.md` - Update section 9 (High-Value Entry Points) |
| Tech stack changed | `AGENTS.md` - Update section 3 (Tech Stack) |
| mihomo integration changed | `AGENTS.md` - Update section 8 (mihomo Integration) |
| Workflow changed | `.agents/skills/*/SKILL.md` |

---

## Bilingual Requirements

**All user-facing and developer-facing documentation must be maintained in both English and Chinese.**

### Language Versions

- English: `*.md`
- Chinese: `*.zh-CN.md`

### Consistency Requirements

- [ ] Both versions updated in the same commit
- [ ] Content is semantically equivalent (not just machine-translated)
- [ ] Code examples identical in both versions
- [ ] Screenshots show appropriate language version
- [ ] Links work in both versions
- [ ] Language switch links present and working

### Language Switch Links

Each bilingual document should have language switch links at the top:

**English version**:
```markdown
[中文](./README.zh-CN.md) | English
```

**Chinese version**:
```markdown
中文 | [English](./README.md)
```

---

## Documentation Checklist

### Before Updating Docs

- [ ] Identify all affected documentation files
- [ ] Check if both language versions exist
- [ ] Read current documentation to understand context
- [ ] Identify outdated information to remove
- [ ] Identify new information to add

### While Updating Docs

- [ ] Update English version first (or Chinese if more natural)
- [ ] Update corresponding language version immediately
- [ ] Keep content semantically equivalent
- [ ] Update code examples with tested, working code
- [ ] Update screenshots if UI changed
- [ ] Add/update links to related documentation
- [ ] Check for broken links

### After Updating Docs

- [ ] Verify both language versions updated
- [ ] Verify links work
- [ ] Verify code examples are accurate
- [ ] Verify screenshots match current UI
- [ ] Update documentation map if new docs added

---

## Documentation Map Updates

When adding new documentation files, update the documentation map:

**In `AGENTS.md` section 10 (Documentation Map)**:
- [ ] Add new file to appropriate category (user/developer/agent docs)
- [ ] Add brief description of file content

**In `skills/sublinkpro/reference/docs.md`**:
- [ ] Add new file with path and description
- [ ] Update category if needed

---

## Code Examples in Documentation

### Requirements

- [ ] Code examples are tested and working
- [ ] Code examples match current API/syntax
- [ ] Code examples show realistic use cases
- [ ] Code examples include necessary imports/setup
- [ ] Code examples include error handling (if relevant)

### Format

Use markdown code blocks with language identifier:

````markdown
```bash
# Shell commands
docker run -d -p 8000:8000 sublinkpro/sublinkpro
```

```go
// Go code
func Example() {
    // Implementation
}
```

```javascript
// JavaScript code
import { api } from './api';
```
````

### Testing Code Examples

Check shell examples with `bash -n` or `sh -n`, using the interpreter named in the example. Run examples affected by the change from their documented working directory with test configuration. For Go and frontend examples, build or test through the repository entry point that exercises the example. Use a disposable environment for examples that change data or deployment state.

---

## Special Documentation Cases

### API Reference (`skills/sublinkpro/reference/api.md`)

When updating API reference:

- [ ] Document endpoint path
- [ ] Document HTTP method
- [ ] Document authentication requirements
- [ ] Document request parameters (path, query, body)
- [ ] Document request body schema (if applicable)
- [ ] Document response status codes
- [ ] Document response body schema
- [ ] Document error responses
- [ ] Provide curl example
- [ ] Provide response example

**Template**:
````markdown
#### POST /api/v1/example

**Authentication**: Required (X-API-Key)

**Request Body**:
```json
{
  "field": "value"
}
```

**Response** (200 OK):
```json
{
  "code": 200,
  "msg": "Success",
  "data": {"field": "value"}
}
```

**Errors**:
- 400: Invalid request parameters
- 401: Authentication failed
- 500: Internal server error

**Example**:
```bash
curl -X POST http://localhost:8000/api/v1/example \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"field":"value"}'
```
````

### Configuration Reference (`docs/user/configuration.md`)

When updating configuration reference:

- [ ] Document option name
- [ ] Document environment variable name
- [ ] Document data type
- [ ] Document default value
- [ ] Document valid values/range
- [ ] Document when option is required vs optional
- [ ] Document precedence (CLI > env > file > database > default)
- [ ] Provide example usage

**Template**:
```markdown
#### OPTION_NAME

- **Environment Variable**: `SUBLINK_OPTION_NAME`
- **Type**: string / integer / boolean
- **Default**: `default-value`
- **Required**: Yes / No

Description of what this option does.

**Example**:
```yaml
option_name: value
```

Or:
```bash
export SUBLINK_OPTION_NAME=value
```
```

### Feature Documentation (`docs/user/features/*.md`)

When adding/updating feature documentation:

- [ ] Overview section: What the feature does
- [ ] Use cases: When to use this feature
- [ ] Setup: How to enable/configure the feature
- [ ] Usage: How to use the feature (with screenshots)
- [ ] Configuration: Related configuration options
- [ ] Troubleshooting: Common issues and solutions
- [ ] Related features: Links to related documentation

---

## When Documentation Sync Can Be Skipped

Skip Phase 3 only if:

- [ ] Pure internal refactoring (no user-visible changes)
- [ ] Bug fix restores documented behavior (no new behavior)
- [ ] Test-only changes
- [ ] Internal comments/code documentation only

**Document the skip reason** in your change summary:

```
Documentation sync: Skipped
Reason: Internal refactoring, no user-visible behavior changed.
```

---

## Documentation Quality Standards

### Writing Style

- [ ] Clear and concise
- [ ] Active voice (not passive)
- [ ] Present tense
- [ ] Second person ("you") for user docs
- [ ] Technical but accessible language
- [ ] No marketing language or hype

### Structure

- [ ] Logical flow (overview → details → examples)
- [ ] Proper heading hierarchy (h1 → h2 → h3)
- [ ] Short paragraphs (3-5 sentences max)
- [ ] Lists for enumerations
- [ ] Code blocks for commands/code
- [ ] Tables for comparisons/references

### Accuracy

- [ ] Technical details are correct
- [ ] Examples are tested and working
- [ ] Links point to correct locations
- [ ] Screenshots match current UI
- [ ] No outdated information

---

## Verification Commands

### Check Documentation Completeness

```bash
# List all .md files without .zh-CN.md counterpart
for f in docs/*.md; do
  if [[ ! -f "${f%.md}.zh-CN.md" ]]; then
    echo "Missing Chinese version: $f"
  fi
done

# List all .zh-CN.md files without .md counterpart
for f in docs/*.zh-CN.md; do
  if [[ ! -f "${f%.zh-CN.md}.md" ]]; then
    echo "Missing English version: $f"
  fi
done
```

### Check for Broken Links

```bash
# Check for broken internal links
grep -r "\[.*\](\./" docs/ | while read -r line; do
  file=$(echo "$line" | cut -d: -f1)
  link=$(echo "$line" | grep -o "(\./[^)]*)" | tr -d "()")
  dir=$(dirname "$file")
  target="$dir/$link"
  if [[ ! -f "$target" ]]; then
    echo "Broken link in $file: $link"
  fi
done
```

### Check Documentation Synchronization

```bash
# Compare English and Chinese file modification times
for f in docs/*.md; do
  zh="${f%.md}.zh-CN.md"
  if [[ -f "$zh" ]]; then
    en_time=$(stat -f %m "$f")
    zh_time=$(stat -f %m "$zh")
    if (( en_time - zh_time > 3600 )); then
      echo "Warning: $f modified but not $zh"
    fi
  fi
done
```

---

## Anti-Patterns

❌ **Updating only English docs**
- Chinese users won't see the update

❌ **Machine-translating without review**
- Technical terms may be incorrect
- Context may be lost

❌ **Documenting code without testing examples**
- Examples may not work

❌ **Adding features without documentation**
- Users won't know the feature exists

❌ **Updating code without updating API reference**
- AI agents will have outdated information

❌ **Writing marketing language in technical docs**
- Users need facts, not hype

---

## Related Skills

- `.agents/skills/cross-layer-sync/SKILL.md` - Cross-layer synchronization
- `.agents/skills/pre-commit-check/SKILL.md` - Pre-commit validation
