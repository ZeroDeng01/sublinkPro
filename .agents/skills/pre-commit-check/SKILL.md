---
name: pre-commit-check
description: Verify SublinkPro Go, frontend, documentation, and cross-layer changes before declaring work complete or preparing a Git submission.
user-invocable: true
metadata:
  author: SublinkPro Team
  version: "2.1"
---

# Pre-Commit Check

Read [AGENTS.md](../../../AGENTS.md) for the repository rules. Run this checklist before declaring work complete, before staging or committing, and before opening or updating a PR. If automatic skill loading is unavailable, read this file directly and execute the checks.

## Required checks

1. Check `git status --short`, `git diff --name-only`, and `git diff --cached --name-only`. Include untracked deliverables in the review. Determine which checks apply to the change.
2. For Go changes, use `gofmt -w` on the changed Go files and confirm `gofmt -l` lists none. Run `go build ./...`, `golangci-lint run`, and `go test ./...`. If the environment prevents the full suite from running, verify at least the relevant packages and report why the full suite could not run.
3. For frontend changes, run `yarn run lint` in `webs/`. Run `yarn run build` for changes affecting build configuration, routing, assets, or production integration.
4. For documentation changes, run `python3 .github/scripts/check_docs.py`. Verify the affected English and Chinese pages, links, and examples. Follow the [documentation checklist](../doc-sync-check/SKILL.md) for changes to behavior, APIs, configuration, or documentation paths.
5. Add tests for new behavior, update tests for intentional behavior changes, and add regression coverage for bug fixes. For cross-layer changes, exercise the normal user flow and apply the [cross-layer checklist](../cross-layer-sync/SKILL.md).
6. Fix failures and rerun the affected checks after any fixes or later edits. Record the commands run, results, and any environment limitations. See the [validation command reference](../post-dev-workflow/references/validation-commands.md) for troubleshooting.

## Hard stops

- An applicable check that fails or remains unverified blocks `git add`, `git commit`, creating or updating a PR, and declaring the work complete. Use the relevant-package fallback only for a documented environment limitation that prevents running the full Go suite. Failing tests must be resolved even if a smaller subset passes.
- Do not skip applicable checks for small patches or on request. Do not present a check as passed unless it ran successfully against the relevant changes.
- Do not stage credentials, secrets, runtime data, temporary files, unrelated changes, or unintended binaries. Preserve the user's staged work.
- Do not auto-commit. Stage only when requested and present the staged diff for user verification.

## Handoff

After the checks pass, report the validation results. If asked to stage, confirm `git status --short` and inspect `git diff --cached` before handing off. Prepare a [semantic commit message](../post-dev-workflow/references/commit-message-guide.md) of at most 72 characters in the subject. Include why the change was made, how it was validated, and any relevant issue. PRs target `dev`. See the [contribution guide](../../../CONTRIBUTING.md).
