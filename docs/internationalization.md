English | [简体中文](internationalization.zh-CN.md)

# Internationalization Contract

This document defines the Wave 1 implementation contract for SublinkPro internationalization. It is for contributors wiring i18n into the existing React/Vite/MUI frontend and Go/Gin backend.

---

## Scope

- Initial supported locales are `zh-CN` and `en-US`.
- `zh-CN` is the source language for much of the current UI. `en-US` must be complete before the rollout is considered usable.
- New user-visible frontend text must be translated in both initial locales in the same change.
- Future locales must be added through the same namespace layout, detection rules, MUI mapping, and formatting helpers described here.

SublinkPro doesn't use URL path or subdomain language routing. `SUBLINK_WEB_BASE_PATH` is already a routing concern for the Web UI, so locale selection must not add another path prefix or host based routing layer.

---

## Frontend Contract

Frontend i18n belongs under `webs/src/`. Keep translation resources close to the frontend app, not in the backend or runtime data directories.

Use namespaces by feature or stable UI boundary. A practical starting layout is:

```text
webs/src/i18n/
├── index.js
├── locales.js
└── locales/
    ├── en-US.json
    └── zh-CN.json
```

The Wave 1 resource files are one JSON file per locale. If later translation files grow too large, split them by namespace in a deliberate follow-up and update this contract at the same time.

Namespace rules:

- `common` is for shared actions, statuses, validation, empty states, and generic labels.
- Feature namespaces match page or module names, such as `airports`, `subscriptions`, `dashboard`, `settings`, and `tasks`.
- Don't create one namespace per component unless the component is reused across unrelated features.
- Keep keys stable when copy changes. A key names the meaning, not the current wording.

Key naming rules:

- Use dot paths with lower camel case segments, such as `actions.save`, `fields.username`, and `errors.loginFailed`.
- Prefer semantic names over English text as keys.
- Use named interpolation values, such as `taskCount: "{{count}} tasks"`.
- Use pluralization support for counts instead of manual string assembly.
- Avoid concatenating translated fragments in JSX. Translate full phrases and sentences.

---

## Language State

Language detection must follow this order:

1. User preference saved in local storage.
2. Browser language when it matches a supported locale or a supported base language.
3. `zh-CN` as the default fallback.

Persist the explicit user choice in local storage. The selected language should survive refresh, sign out, and sign in on the same browser. Don't store the language in backend runtime directories or server session state for Wave 1.

When matching browser languages, normalize aliases conservatively:

- `zh`, `zh-CN`, `zh-Hans` map to `zh-CN`.
- `en`, `en-US`, `en-GB`, and other English variants map to `en-US`.

Unsupported languages fall back to `zh-CN` until their translation resources and MUI locale mapping are added.

---

## MUI Locale And Formatting

MUI locale configuration must be selected from the active app locale:

| App locale | MUI locale |
|:---|:---|
| `zh-CN` | `zhCN` |
| `en-US` | `enUS` |

Keep this mapping in one frontend helper, for example `webs/src/i18n/muiLocale.js`, so future locales are added in one place.

Dates, times, relative time, numbers, percentages, and byte or traffic values must be formatted through shared helpers that accept the active locale. Don't hardcode Chinese or English date order in feature components.

Formatting rules:

- Use `Intl.DateTimeFormat` and `Intl.NumberFormat` where possible.
- Keep protocol values, node names, domains, file paths, API keys, and template syntax unchanged.
- Keep units translatable when they are visible UI text.
- For logs and debug identifiers, preserve machine-readable values and translate only surrounding labels when needed.

---

## Backend Contract

The backend should stay compatible with existing API clients while allowing localized frontend display.

API responses that already expose `msg` or `code` must keep those fields stable unless the endpoint contract is intentionally changed and documented. New or revised responses may add `i18nKey` and optional interpolation data so the frontend can render localized text.

Recommended response shape for user-facing errors or task results:

```json
{
  "code": 400,
  "msg": "Invalid request",
  "i18nKey": "errors.invalidRequest",
  "i18nParams": {
    "field": "name"
  }
}
```

Compatibility rules:

- `msg` remains a fallback for older clients and diagnostics.
- `code` remains the programmatic status or business code.
- `i18nKey` is optional at first, then can become the preferred frontend display source as coverage grows.
- `i18nParams` must contain plain JSON values only.
- Don't translate protocol output, subscription content, template variables, machine logs, IDs, tokens, or values consumed by external clients.

Backend-owned human text is still allowed when the backend is the product surface, such as CLI help, startup errors, logs, webhook payloads, Telegram Bot replies, generated task messages, and exported files. When that text is shown inside the Web UI, prefer `i18nKey` plus params if the UI needs localization.

---

## Adding A Locale

To add a future locale:

1. Add the locale code to the supported locale list.
2. Add a complete resource file under `webs/src/i18n/locales/<locale>.json` with the same key structure as `zh-CN` and `en-US`.
3. Add browser language normalization if needed.
4. Add the MUI locale mapping.
5. Confirm date, time, number, percentage, and traffic formatting for that locale.
6. Update both English and Simplified Chinese docs if behavior or contributor guidance changes.

Do not add a locale unless all user-visible strings in the initial namespace set have translations or a documented rollout plan.

---

## Documentation Rule

Documentation changes for i18n must be made in English and Simplified Chinese pairs. If `docs/internationalization.md` changes, update `docs/internationalization.zh-CN.md` in the same work, and keep language switch links valid.

The same rule applies to future README or feature documentation updates: English canonical docs and matching `*.zh-CN.md` files must stay aligned.
