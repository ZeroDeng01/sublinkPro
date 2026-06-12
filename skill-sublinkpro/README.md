# SublinkPro AI Skill

English | [简体中文](README.zh-CN.md)

An AI-powered interface for [SublinkPro](https://github.com/ZeroDeng01/sublinkPro) — manage proxy nodes, subscriptions, and shares through natural language.

## What is this?

This skill lets you control your SublinkPro instance by talking to an AI assistant instead of clicking through the web UI. Ask the assistant to "add this node" or "create a subscription" and it calls the API for you.

**It's guided by default.** You don't need to know the API, the parameter names, or what already exists in your instance. Just say what you want — even something vague like "I want to add a node" — and the assistant walks you through it step by step: it looks up your existing groups/subscriptions/nodes so you can pick from a list, asks only for what's needed, confirms before making any change, and reports the result in plain language. Not sure where to start? Just ask **"what can I do?"**.

**It can also deploy SublinkPro for you.** Don't have an instance running yet? Ask the assistant to set one up and it guides you through deploying via **docker**, **docker-compose**, or the **install script** — **locally or on a remote host** — then helps you finish first-run setup. So the skill covers both standing up SublinkPro and operating it afterward.

It is a portable, vendor-neutral skill: the `SKILL.md` + helper-script format works with any AI agent or client that supports skills (for example Claude Code, but not limited to it). Under the hood it just drives the SublinkPro REST API, so anything that can run the helper script can use it.

## Installation

### Step 0 — Get the skill

There are two ways to obtain this `skill-sublinkpro/` directory; pick whichever fits you:

- **From a running SublinkPro instance (easiest if you already have one).** Sign in to the web UI and click the **AI skill download** icon in the top bar (next to the donation icon). You'll get `skill-sublinkpro.zip` — unzip it and you have the directory. The bundled copy always matches your server's version.
- **From GitHub (for first-timers who don't have SublinkPro yet).** Download/clone the `skill-sublinkpro/` directory from the repository. You can then *use the skill itself to deploy SublinkPro* (see "It can also deploy SublinkPro for you" above) — handy if you're starting from nothing.

In short: **no instance yet → GitHub** (then let the skill deploy one); **already have an instance → just download it from the web UI**.

### Step 1 — Install it

1. **Copy it to your AI agent's skills directory.** The exact path depends on your tool. For Claude Code:
   ```bash
   cp -r skill-sublinkpro ~/.claude/skills/sublinkpro
   # or symlink if you cloned the repo
   ```
   For other agents, place the directory wherever that tool loads skills from.

2. **Set environment variables:**
   ```bash
   export SUBLINK_BASE_URL=http://localhost:8000
   export SUBLINK_API_KEY=prefix_xxx_yyy
   ```

3. **Get an API key:**
   - Open SublinkPro web UI
   - Go to **Settings → Access Keys**
   - Click **Create**
   - Copy the key (shown only once)

## Usage Examples

You don't need to know the API or the parameters. Just say what you want — or ask **"what can I do?"** for a menu — and the assistant guides you step by step, looking up your existing groups, subscriptions, and nodes so you can pick from a list instead of typing, and confirming before it changes anything.

- "I want to add a node" → it walks you through it
- "Add this node: `vless://...` and name it 'HK-01'"
- "Help me build a new subscription"
- "Show me all my subscriptions"
- "Create a share link for the 'US Servers' subscription"
- "Import nodes from this airport: https://..."
- "Edit the 'ACL4SSR' template to block ads"
- "Show dashboard stats"

## What's Included

- **`SKILL.md`** — skill definition (the [Agent Skills](https://code.claude.com/docs/en/skills) open format, usable by any compatible AI agent)
- **`reference/api.md`** — Complete API reference (every endpoint with curl examples)
- **`reference/deploy.md`** — Deployment command catalog (docker / compose / script)
- **`scripts/sublink.py`** — *optional* Python helper that wraps auth, content-type, and error checking (only used if Python 3 is present; the API is normally called with curl)

## Requirements

- **curl** (preinstalled on virtually every macOS, Windows 10+, and Linux system) — the default way the API is called
- A running SublinkPro instance and an API key (created in the web UI)
- Python 3 is **optional** — only needed if you prefer the `scripts/sublink.py` convenience wrapper (stdlib only, no pip dependencies)

## License

Same as SublinkPro (check the main repo).
