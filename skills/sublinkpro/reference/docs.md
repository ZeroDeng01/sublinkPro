# SublinkPro Documentation Map (online manual)

This turns the skill into an **online instruction manual** for SublinkPro: when the
user asks *how something works*, *how to configure it*, *what a feature does*, or
*how to operate the project itself* (not a live API call), the agent answers from
the project's official docs.

**The docs are NOT bundled into this skill.** They live in the GitHub repo and are
fetched on demand, so answers always reflect the current published documentation.
Use your web-fetch capability to read the relevant file, then answer in the user's
language — summarize and quote the doc, don't invent details.

## How to use this map

1. Match the user's question to a topic in the table below.
2. Fetch that doc's raw URL (see URL pattern). Prefer the language matching the
   user's question: Chinese → the `*.zh-CN.md` variant; English → the base `.md`.
3. Read it and answer. If the answer spans multiple docs, fetch each as needed.
4. Cite the human-readable GitHub page (the `blob` URL) so the user can open it too.
5. If you can't fetch (offline / no web access), tell the user and give them the
   GitHub link to read it themselves — never fabricate the content.

## URL patterns

- **Raw (fetch this to read content):**
  `https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/main/<path>`
- **Human page (give this to the user):**
  `https://github.com/ZeroDeng01/sublinkPro/blob/main/<path>`
- **Repo root:** `https://github.com/ZeroDeng01/sublinkPro`

Every doc is maintained in **both languages**: English at `<path>.md`, Simplified
Chinese at the same path with a `.zh-CN.md` suffix (e.g. `docs/user/installation.md` →
`docs/user/installation.zh-CN.md`). The table lists the English `<path>`; swap the
suffix for Chinese.

## Topic → document map

### Overview & getting started
| User asks about | Doc path (`<path>`) |
|---|---|
| What the project is, feature overview, quick start, screenshots | `README.md` |
| Documentation index and separate guides for users, developers, and coding agents | `docs/README.md` |

### Install, configure, deploy
| User asks about | Doc path |
|---|---|
| Install methods (Docker, Docker Compose, one-line script), updates and removal | `docs/user/installation.md` |
| Environment variables, command-line flags, config precedence, CAPTCHA modes | `docs/user/configuration.md` |
| Build process, production build, Docker build, CI/CD, troubleshooting | `docs/development/build-and-deployment.md` |
| Security best practices, default credentials, sensitive config, MFA, Docker security | `docs/user/security-guidelines.md` |

> For deploying *through the skill* (guided, conversational), also see this skill's
> own `reference/deploy.md`. Use `docs/user/installation.md` + `docs/user/configuration.md`
> when the user wants to read the canonical project documentation.

### Feature guides
| User asks about | Doc path |
|---|---|
| Smart tag system — rule-based auto-tagging, mutually-exclusive groups, IP quality conditions | `docs/user/features/tags.md` |
| Speed-test system — two-stage tests, latency/speed, IP quality & unlock checks, tuning | `docs/user/features/speedtest.md` |
| Unlock checks — streaming & AI availability, Provider architecture, extensions | `docs/user/features/unlock-check.md` |
| Chain proxy — Dialer-Proxy, condition-based node selection, config flow | `docs/user/features/chain-proxy.md` |
| AI template editing, operation sessions, server preview validation, read-only diff review, validation warnings, accept into editor, normal save | `docs/user/features/template-ai.md` |
| Airport management — import, scheduled updates, traffic monitoring | `docs/user/features/airport.md` |
| Subscription sharing — multiple links, expiration policies, access stats | `docs/user/features/subscription-share.md` |
| Host management — domain mappings, DNS, CDN preferred IPs | `docs/user/features/host.md` |
| Cloudflare Tunnel — create tunnel, token, public access | `docs/user/features/cloudflare-tunnel.md` |
| System backup and WebDAV — configure, upload, list, restore, security boundaries | `docs/user/features/backup.md` |
| Proxy gateway — SOCKS5/HTTP/Mixed listeners, username routing, independent accounts, sticky sessions, curl examples, and monitoring | `docs/user/features/proxy-gateway.md` |
| Telegram Bot — command list, setup | `docs/user/features/telegram-bot.md` |
| Multi-factor auth (MFA) — TOTP setup, recovery codes, emergency reset | `docs/user/features/mfa.md` |
| Script support — node filtering, content post-processing, function reference | `docs/user/features/script-support.md` |

### For developers
| User asks about | Doc path |
|---|---|
| Project structure, local dev, scheduled-task dev, commenting standards, testing standards | `docs/development/getting-started.md` |
| Complete development workflow — task understanding to commit, mandatory post-dev validation | `docs/development/development-workflow.md` |
| Build process, production build, Docker build, CI/CD pipeline, build troubleshooting | `docs/development/build-and-deployment.md` |
| Common development patterns — add backend feature, scheduled task, mihomo changes, deployment changes | `docs/development/practical-recipes.md` |
| Protocol extension guide (add a protocol, register capabilities, field metadata) | `docs/development/getting-started.md` (section "Protocol Extension Guide") |
| Internationalization (i18n) contract and how translations are maintained | `docs/development/internationalization.md` |
| Frontend theme adaptation rules (light/dark mode, surface layering, coverage) | `docs/development/frontend-theme-guidelines.md` |
| Security best practices, credentials, sensitive config, MFA, incident response | `docs/user/security-guidelines.md` |
| Contribution workflow, branch conventions, PR process, cross-layer sync requirements | `CONTRIBUTING.md` |
| Code of Conduct | `CODE_OF_CONDUCT.md` |
| Architectural guidance, tech stack, mihomo integration, cross-layer contracts | `AGENTS.md` |

## Notes & guardrails

- This map links to the most common topics. If a fetch 404s, the doc may
  have been renamed/moved — open the live [documentation index](https://github.com/ZeroDeng01/sublinkPro/blob/main/docs/README.md)
  to find the current path, and tell the user it moved. Don't guess a new path.
- These docs describe the **project** (install, features, configuration). For
  *operating a live instance* via the REST API, use `reference/api.md`. For
  *deploying* an instance step by step, use `reference/deploy.md`.
- Always answer in the user's language; fetch the matching language variant when
  available.
