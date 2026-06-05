English | [简体中文](tags.zh-CN.md)

# Smart Tag System

SublinkPro's automatic tag system is one of its most powerful features. It upgrades node management from “manual editing” to “rule driven”.

---

## 💡 Core Benefits

| Benefit | Description |
|:---|:---|
| **🚫 No code filtering** | After tag rules apply tags to nodes, subscriptions can use tags for complex filtering with **no code or scripts required** |
| **🔄 Dynamic automation** | Tags update automatically with node state, for example high latency nodes are removed from a “low latency” tag after tests |
| **📊 Multi dimensional classification** | Combine conditions such as latency, speed, country, protocol, source, node name, and IP quality |
| **🎛️ Flexible operators** | Supports equals, contains, greater than, less than, regex match, and more |
| **♻️ Mutually exclusive groups** | Tags in the same exclusive group replace each other automatically, keeping node classification unique and clear |

---

## 📋 Example Use Cases

```text
Case 1: Grade by latency
├── Rule: latency < 100ms -> tag “⚡Fastest”
├── Rule: latency 100-300ms -> tag “✅Normal”
└── Rule: latency > 300ms -> tag “🐌Slow”

Case 2: Classify by region
├── Rule: country = Hong Kong -> tag “🇭🇰Hong Kong”
├── Rule: country = Japan -> tag “🇯🇵Japan”
└── Rule: country = United States -> tag “🇺🇸United States”

Case 3: Layer by IP quality
├── Rule: residential attribute = residential IP -> tag “🏠Residential”
├── Rule: IP type = native IP -> tag “🌐Native”
└── Rule: fraud score <= 30 -> tag “✅Low risk”

Case 4: Subscription filtering, no code
├── Subscription A: tag allowlist = “⚡Fastest” “🇭🇰Hong Kong” -> only fast Hong Kong nodes
└── Subscription B: tag blocklist = “🐌Slow” -> exclude all slow nodes
```

---

## Mutually Exclusive Tag Groups

> [!TIP]
> **Use case**: Create “Excellent”, “Good”, and “Poor” tags in the same group “Speed rating”. After speed tests, a node keeps only the latest rating, preventing tag buildup.

Exclusive groups ensure that only one tag from the same group remains on a node. When a new tag matches, old tags in the same group are removed automatically.

**Suitable for:**

- Speed rating, excellent / good / poor
- Latency level, fastest / normal / slow
- Stability rating, stable / unstable

---

## Tag Rule Configuration

On the “Tag Management” page, each tag can have automatic matching rules.

### Supported condition fields

| Field | Description |
|:---|:---|
| Node name | Node display name |
| Country/region | Node landing location |
| Protocol type | ss/ssr/vmess/vless/trojan and others |
| Latency, ms | Node latency test result |
| Speed, MB/s | Node speed test result |
| Fraud score | 0-100, lower is lower risk |
| IP type | Native IP / broadcast IP / untested |
| Residential attribute | Residential IP / data center IP / untested |
| Unlock Provider | Service for the node's primary unlock result, such as Netflix / OpenAI / Gemini |
| Unlock status | Available / partial / reachable / restricted / unsupported / unknown / error / untested |
| Unlock keyword | Fuzzy match Provider, status, region, reason, and details in unlock results |
| Unlock summary | Compact node unlock summary, useful for string matching |
| Source airport | Airport subscription that owns the node |

> [!NOTE]
> Values for `Unlock Provider` and `Unlock status` are delivered dynamically by the backend. After adding an unlock checker, related dropdown options in tag rules update automatically without adding frontend enums manually.

### Supported operators

| Operator | Description |
|:---|:---|
| Equals | Exact match |
| Not equals | Exclude exact match |
| Contains | Partial string match |
| Does not contain | Exclude partial match |
| Greater than | Numeric comparison |
| Less than | Numeric comparison |
| Regex match | Use a regular expression |
| Greater than or equal | Numeric comparison |
| Less than or equal | Numeric comparison |

> [!TIP]
> **Enum field advice**: `IP type` and `Residential attribute` are enum fields. Prefer equals / not equals. `Fraud score` is better suited for greater than, less than, greater than or equal, and less than or equal.

### Typical rule examples

| Target tag | Recommended condition |
|:---|:---|
| `Low risk` | Fraud score `<= 30` |
| `Residential` | Residential attribute `= Residential IP` |
| `Native` | IP type `= Native IP` |
| `Needs retest` | IP type `= Untested` |
| `OpenAI available` | Unlock Provider `= openai` and unlock status `= available` |
| `AI reachable` | Unlock keyword `contains Gemini` or `contains OpenAI` |

---

## Use Tags in Subscriptions

When creating a subscription, you can set:

- **Tag allowlist**: include only nodes with specified tags
- **Tag blocklist**: exclude nodes with specified tags

Allowlist and blocklist can be combined. The system applies the allowlist first, then excludes blocklisted nodes.
