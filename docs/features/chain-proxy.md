English | [简体中文](chain-proxy.zh-CN.md)

# Chain Proxy

This is one of SublinkPro's **featured capabilities**. With proxy chaining, you can build traffic forwarding paths from an “entry proxy” to a “landing proxy”.

---

## 🔌 What Is Chain Proxy?

In simple terms, your traffic first goes through “entry node A”, then forwards to “landing node B”, and finally reaches the target website.

**Traffic path**: `Client -> entry node A -> landing node B -> target website`

---

## 🛠️ Core Benefits and Scenarios

| Scenario | Solution | Benefit |
|:---|:---|:---|
| **Rescue blocked nodes** | Your landing VPS IP is blocked and cannot connect directly | Use an airport node as the entry relay to restore VPS connectivity |
| **Clean up IP reputation** | Airport nodes are fast but have dirty IPs, such as frequent Google CAPTCHA | Use an airport node as entry -> niche VPS with clean IP as landing, keeping both speed and cleaner reputation |
| **Improve network path** | Direct route to landing node is poor, such as non CN2 routes | Use CN2 or dedicated route nodes as entry to reduce latency and packet loss |

---

## ⚙️ Powerful Configuration

SublinkPro provides highly flexible chain proxy configuration.

### 1. Native Clash Dialer-Proxy support

The generated Clash config automatically includes the `dialer-proxy` field, using the client core for traffic forwarding with **no performance loss** and no extra server side relay program.

### 2. Visual configuration flow

Flowchart based and intuitive UI design lets you clearly see each chain.

### 3. Flexible entry selection

- **Specific node**: choose one exact node as the entry, such as "Hong Kong Dedicated 01".
- **Dynamic policy group**: choose a policy group, such as "Auto Select" or "Load Balance", as the entry for dynamic high availability.

### 4. Conditional node selection

Both **intermediate nodes** and **target nodes** support dynamic condition based filtering. This is useful when entry quality and landing quality need separate control.

| Condition type | Supported fields |
|:---|:---|
| Basic fields | Node name, country, protocol, tags, latency, speed |
| IP quality fields | Fraud score, IP type, residential attribute |
| Unlock fields | Unlock Provider, unlock status, unlock keyword, unlock summary |

> [!TIP]
> `IP type` values are “native IP / broadcast IP / untested”. `Residential attribute` values are “residential IP / data center IP / untested”.
>
> Unlock status is read from backend provided status metadata. After adding an unlock checker, Provider options in chain proxy conditions update automatically.

---

## 💡 Usage Tip

> [!TIP]
> **Tip**: Chain proxy is mainly for **subscription conversion**. You can configure rules so all nodes in a subscription connect through another “front proxy”, instantly turning a normal subscription into a relayed subscription.

---

## Configuration Flow

1. Open the subscription configuration page.
2. Find the “Chain Proxy” configuration area.
3. Set the entry node, supporting specific node or policy group.
4. Set landing node rules, such as matching by tag or country.
5. Save. The subscription link will include chain proxy config automatically.

> [!NOTE]
> Chain proxy rules are **saved per subscription and applied during subscription generation**. They are not global runtime proxy rules. Template proxy groups come from the Clash template associated with the subscription. Final generated nodes implement the chain through Clash `dialer-proxy`.

### Common examples

| Goal | Intermediate node condition | Target node condition |
|:---|:---|:---|
| Improve entry availability | Latency `< 200ms`, speed `> 5MB/s` | No limit |
| Prefer clean landing IPs | No limit | Fraud score `<= 30`, IP type `= Native IP` |
| Residential landing scenario | No limit | Residential attribute `= Residential IP` |
| Exclude nodes that need retest | IP type `!= Untested` | IP type `!= Untested` |
| Prefer OpenAI landing | No limit | Unlock Provider `= openai`, unlock status `= Available` |
| Prefer unlocked landing | No limit | Unlock status `= Available` or `= Reachable` |
