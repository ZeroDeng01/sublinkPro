---
name: post-dev-workflow
description: Validate completed SublinkPro code changes, synchronize affected layers and documentation, and prepare the changes for review.
user-invocable: false
metadata:
  author: SublinkPro Team
  version: "2.1"
---

# Post-Development Workflow

Run this workflow after implementing code changes. Resolve failures before handing off a change as complete.

1. Identify the affected layers and the normal user flow you need to exercise. Compare the diff with [AGENTS.md](../../../AGENTS.md).
2. Run the checks in the [pre-commit skill](../pre-commit-check/SKILL.md) for the changed layers. Validate the behavior and add or update tests when key logic changes.
3. Apply the [cross-layer checklist](../cross-layer-sync/SKILL.md) for API, permission, UI contract, or configuration changes. Open the [theme](../theme-check/SKILL.md), [security](../security-review/SKILL.md), or [performance](../performance-check/SKILL.md) guide when that area is affected.
4. Apply the [documentation checklist](../doc-sync-check/SKILL.md). Update both languages, public API/config references when affected, and pointers to any moved documents.
5. Inspect the final diff and report the validations. If asked to stage, use the pre-commit criteria before staging only the intended files. Leave the commit for user verification.
