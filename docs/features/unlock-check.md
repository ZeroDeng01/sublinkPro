English | [简体中文](unlock-check.zh-CN.md)

# Unlock Checks

SublinkPro now supports **streaming / AI service availability checks** as part of the node check flow.

This feature does not start a separate task system. It runs directly inside the existing **node check / speed test profile** flow:

- Choose node range
- Run latency / speed checks
- Optionally check landing country and IP quality
- Optionally check unlock availability
- Write results directly back to node information, and display them in the frontend node list, details panel, and Task Center

---

## Current Built In Checks

First built in Providers:

- Netflix
- Disney+
- YouTube Premium
- OpenAI
- Gemini
- Claude

> [!NOTE]
> Current built in Providers try to use service level probes consistent with mainstream unlock check scripts. For example, OpenAI checks Web and iOS entries separately, Disney+ uses device, token, and region GraphQL probes, while YouTube Premium, Gemini, Claude, and Netflix read availability markers from the relevant pages or final redirects.
>
> The list above means **currently built in** Providers, not fixed frontend or rule system enums. After adding a checker, related Provider selectors and unlock conditions update automatically through backend metadata.

---

## Usage

Open:

`Node Check -> New / Edit Profile`

In the profile, enable:

- **Unlock check**
- The Provider list to check

If no Providers are selected manually, the system runs the backend registered default Provider set.

After execution, results appear in:

- Node list
- Node card
- Node details panel
- Task progress panel
- Task Center history

---

## Unlock Filter Rules

Node lists and subscription filters now support **multiple unlock filter rules**.

Each rule contains:

- `Provider`
- `Status`
- `Keyword`

Matching semantics:

- **Inside one rule**: **AND**
- **Between multiple rules**: user selectable **OR** or **AND**

That means:

- One rule: `Gemini + available + US`
  - The same unlock result must satisfy all three conditions
- Multiple rules:
  - `Gemini + available`
  - `YouTube Premium + available`
  - In OR mode: matching any one rule passes
  - In AND mode: all rules must be satisfied

If the user has not added unlock rules, **unlock filtering is disabled**.

---

## Naming and Display Advice

Unlock information can be used in node renaming, but the **compact summary** is recommended instead of expanding every platform result into the node name.

Recommended:

- `$Unlock(provider)`: output compact result for a specific Provider, for example `$Unlock(openai)` -> `Available-US`
- `$Unlock`: primary unlock summary, for example `Netflix-Available-US-+2`

Putting detailed results for many platforms into node names is not recommended, because names become too long, hard to read, and hard to search.

This is useful when you want to find:

- Nodes where “Gemini is available”
- Nodes where “Gemini is available or YouTube Premium is available”
- Nodes where “Claude is available and contains keyword US”

---

## Result Semantics

Each Provider result uses a unified structure. Core fields include:

- `provider`: Provider identifier
- `status`: check result status
- `region`: detected region, if applicable
- `reason`: failure / restriction reason
- `detail`: extra notes

Common current statuses:

| Status | Meaning |
|:---|:---|
| `available` | Clearly available |
| `partial` | Partial, such as Originals Only |
| `reachable` | Entry is reachable, but this does not mean full capability |
| `restricted` | Restricted, current region or exit is blocked |
| `unsupported` | Unsupported, current region is outside official support |
| `error` | Error, this check failed |
| `unknown` | Cannot be judged reliably |

> [!IMPORTANT]
> `reachable` and `available` are not equivalent.
>
> `available` means the Provider specific unlock marker or API availability result was detected. `reachable` only means the entry point is reachable and should not be treated as full unlock.
> Gemini checks additionally read page level feature availability markers. Only explicit availability markers are treated as `available`. If the marker is absent, the result is treated as `restricted`, preventing region limited pages that still return HTTP success from being misclassified as “available”.
> OpenAI `partial` means “Only Available with Web Browser” or “Only Available with Mobile APP” in upstream check semantics. Disney+ `partial` means that region is `Available Soon`. Netflix `partial` means Originals Only.

---

## Architecture

The feature uses a **registry + independent Checker modules + orchestrator** structure.

### 1. Scheduling layer

Node checks are still handled by the existing flow:

- `api/node_check.go`
- `models/node_check_profile.go`
- `services/scheduler/speedtest_config.go`
- `services/scheduler/speedtest_task.go`

These files:

- Save profile configuration
- Convert configuration into runtime parameters
- Decide when to run unlock checks
- Write results back to nodes and task results

### 2. Unlock subsystem

Unlock check core files:

- `services/unlock/registry.go`
- `services/unlock/runtime.go`
- `services/unlock/orchestrator.go`
- `services/unlock/checker_*.go`

Where:

- **registry**: registers and resolves Checkers
- **runtime**: provides shared HTTP client, timeout, landing country, and other runtime context
- **orchestrator**: selects Providers by profile, runs checks through one path, and produces result summaries
- **checker files**: each Provider maintains its own probe logic

### 3. Data layer

Node side results are stored in a unified structure:

- `models/unlock.go`
- `models/node.go`

Current nodes store:

- `unlockSummary`
- `unlockCheckAt`

This means adding Providers later does not require adding another column to `Node`.

---

## How to Add an Unlock Checker

This feature's main maintainability goal is:

> **Adding a Provider should usually require only: add checker file + register + declare provider metadata.**

### Step 1: Add checker file

Add a file under `services/unlock/`, such as:

`unlock_checker_example.go`

Implement the unified interface:

```go
type UnlockChecker interface {
    Key() string
    Aliases() []string
    Check(runtime UnlockRuntime) models.UnlockProviderResult
}
```

It is also recommended to implement metadata methods in the checker so the backend can deliver display information and rename variables automatically:

```go
type UnlockCheckerMeta interface {
    Meta() models.UnlockProviderMeta
}

type UnlockCheckerRenameMeta interface {
    RenameVariableMeta() models.UnlockRenameVariableMeta
}
```

### Step 2: Implement `Check`

Inside `Check(runtime UnlockRuntime)`:

- Use the shared proxy HTTP client from runtime
- Run the Provider's own low cost probe
- Return unified `models.UnlockProviderResult`

Don't:

- Operate on the database directly
- Depend directly on scheduler / task manager
- Handle logic for other Providers inside a checker

### Step 3: Register checker

Call registration in `init()` in the new file:

```go
func init() {
    RegisterUnlockChecker(exampleUnlockChecker{})
}
```

### Step 4: Frontend display and rule system

Normally, you don't need to manually add Provider or status enums to the frontend.

These places consume unlock information dynamically through backend metadata:

- Provider selector in node check profiles
- Unlock filter on the node page
- Unlock filter rules in subscription editing
- Unlock conditions in tag rules
- Unlock conditions in chain proxy
- `$Unlock(provider)` variable list in node renaming

In other words, **after adding a checker, downstream UI updates automatically as long as backend metadata is complete**.

Extra frontend changes are needed only when:

- You add a brand new unlock condition field, not just a Provider
- You want special visual presentation for one Provider beyond normal enum options

### Step 5: Documentation sync

After adding a Provider, update as needed:

- “Current Built In Checks” in this document, if you want docs to reflect the latest built in Providers
- Feature description in `README.md`, usually no need to list each Provider
- Development notes in `docs/development.md`, if the extension mechanism itself changes

> [!TIP]
> If you only add a normal checker without changing the extension mechanism, you usually don't need to update multiple frontend docs for enum sync, because runtime selectors read backend metadata automatically.

---

## Maintenance Constraints

To keep this subsystem maintainable, follow these rules:

1. **Each Provider owns only its own logic**.
2. **Shared logic belongs in runtime / orchestrator. Don't copy it between checkers**.
3. **Don't turn Provider dispatch back into a central switch**.
4. **Keep result structures stable and avoid repeated frontend/backend field churn**.
5. **Prefer low cost probes and avoid browser automation or high resource script dependencies**.

---

## Scope

This feature is suitable for:

- Batch filtering nodes by regional capability
- Estimating whether streaming / AI services are likely available
- Adding extra signals for subscription filtering, node operations, and tag rules
- Filtering nodes by unlock result in subscriptions
- Adding unlock summaries to node names

This feature currently does not aim for:

- Perfectly simulating real signed in business state
- 100% accuracy for every platform
- Complex anti bot bypass techniques to improve hit rate

The first version's goal is:

**Maintainable, extensible, batch runnable, and stable to display.**
