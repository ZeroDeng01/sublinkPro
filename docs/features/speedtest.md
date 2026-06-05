English | [简体中文](speedtest.zh-CN.md)

# Speed Test System

SublinkPro provides professional node speed testing with a scientific method for accurate and reliable results.

---

## 🔬 Technical Features

| Feature | Description |
|:---|:---|
| **Separated two stage tests** | Latency and download speed tests run separately, each using the most suitable test URL and method |
| **Smart latency measurement** | Supports UnifiedDelay mode, with optional inclusion or exclusion of handshake time. The system sends two requests and uses the second result |
| **Independent concurrency settings** | Latency and speed tests can use different concurrency values to balance efficiency and accuracy |
| **Automatic status marking** | After tests complete, node latency status and speed status are updated automatically |
| **IP quality checks** | During speed tests, the system can also check exit IP fraud score, IP type, and residential attribute |
| **Unlock checks** | During node checks, the system can also check streaming / AI service availability regions |
| **Real time progress** | Progress and state are shown in real time during tests, with task panel support |
| **Traffic statistics** | Traffic consumed by each tested node is recorded, and task totals are shown when complete |

---

## 📊 Speed Test Status Categories

| Color | Latency threshold | Speed threshold |
|:---|:---|:---|
| 🟢 Green | < 200ms | >= 5 MB/s |
| 🟡 Yellow | 200-500ms | 1-5 MB/s |
| 🔴 Red | >= 500ms or timeout/failure | < 1 MB/s or failure |
| ⚪ Gray | Untested | Untested |

---

## 🌐 IP Quality Checks

After enabling **IP quality check** in “Check Settings”, the system performs extra IP quality API requests during the speed test flow and adds quality information for the node exit IP.

The current `ippure` style API returns fields such as `fraudScore`, `isBroadcast`, and `isResidential`. The frontend and filtering logic interpret them as follows.

### Result semantics

| Field | API response | Frontend display |
|:---|:---|:---|
| **IP type** | `isBroadcast = true` | Broadcast IP |
| **IP type** | `isBroadcast = false` | Native IP |
| **IP type** | Field missing or not checked in this run | Untested |
| **Residential attribute** | `isResidential = true` | Residential IP |
| **Residential attribute** | `isResidential = false` | Data center IP |
| **Residential attribute** | Field missing or not checked in this run | Untested |
| **Fraud score** | `fraudScore = 0-100` | Lower is better |

> [!IMPORTANT]
> **Untested is not false**: When the API does not return `isBroadcast` or `isResidential`, the system displays “Untested”. It does not incorrectly classify the node as “Native IP” or “Data center IP”.

### Fraud score levels

| Score range | Rating | Description |
|:---|:---|:---|
| `0 - 10` | Excellent | Very low risk |
| `11 - 30` | Great | Lower risk |
| `31 - 50` | Good | Usable normally |
| `51 - 70` | Medium | Judge with your business scenario |
| `71 - 89` | Poor | Higher risk |
| `90+` | Very poor | Very high risk |

### Where results can be used

- **Node Management**: view fraud score, IP type, and residential attribute directly, and filter by all three dimensions.
- **Subscription filtering**: set max fraud score, IP type, and residential attribute filters while editing subscriptions.
- **Node rename rules**: use `$FraudScore`, `$IpType`, and `$Residential` in rename rules.
- **Automatic tags / chain proxy**: after condition fields are added, rules can match directly on IP quality results.

---

## 🌍 Unlock Checks

After enabling **unlock check** in “Check Settings”, the system runs selected Provider probes during the node check flow.

Built in Providers currently include:

- Netflix
- Disney+
- YouTube Premium
- OpenAI
- Gemini
- Claude

Unlock results are written back to node information and displayed in the node list, node details, task progress panel, and Task Center.

> [!TIP]
> Unlock checks use an independent Provider Checker + registry architecture, so adding platforms later does not require rewriting the whole speed test task.
>
> Provider selectors, unlock status options, and unlock condition schemas are delivered by the backend, so adding a normal checker does not require manually adding options across multiple frontend pages.

For detailed architecture and extension instructions, see:

- [🌍 Unlock checks](unlock-check.md)

---

## 🔧 Speed Test Principle and Traffic Calculation

> [!IMPORTANT]
> **Core principle**: Speed tests use “timed download”. During the configured timeout, the system downloads as much data as possible, then calculates speed from actual downloaded bytes and elapsed time.

**Speed formula**:

```text
Speed (MB/s) = actual downloaded bytes ÷ 1024 ÷ 1024 ÷ actual elapsed time, seconds
```

### Common questions

| Symptom | Reason |
|:---|:---|
| Test file is 5MB, but less than 5MB is downloaded | The node is slow and did not finish the download within the timeout |
| Some nodes download 500KB, others download 5MB | Fast nodes finish before timeout, slow nodes do not |
| Traffic consumption shows a few hundred KB | Node speed is about 100KB/s × timeout in seconds |

### Example, assuming a 5 second timeout

| Node speed | Actual download | Calculated result |
|:---|:---|:---|
| 10 MB/s | 5 MB, finished early | Speed ≈ 10 MB/s |
| 1 MB/s | 5 MB | Speed = 1 MB/s |
| 0.2 MB/s | 1 MB | Speed = 0.2 MB/s |
| 0.1 MB/s | 500 KB | Speed = 0.1 MB/s |

---

## ⚙️ Parameter Recommendations

| Parameter | Description | Recommended value |
|:---|:---|:---|
| **Speed test timeout** | Maximum duration for each node speed test | 5-10 seconds, balancing speed and accuracy |
| **Speed test file URL** | File URL used for download speed testing | Recommended size ≥ timeout × expected max speed |
| **Latency test URL** | URL used for latency testing | Use HTTP 204 or a small file, such as Cloudflare generate_204 |
| **Latency concurrency** | Number of nodes tested for latency at the same time | 10-50, can be increased as needed |
| **Speed test concurrency** | Number of nodes tested for speed at the same time | 1-3, to avoid bandwidth contention |

> [!TIP]
> **Speed test file size recommendation**: If timeout is 5 seconds and the expected fastest node is 20 MB/s, the test file should be at least 100MB. Recommended Cloudflare speed test URL: `https://speed.cloudflare.com/__down?bytes=100000000`, 100MB.
>
> ⚠️ **Note**: Larger files and longer timeouts may consume more traffic.

> [!WARNING]
> **Traffic consumption warning**: Each speed test consumes real bandwidth. 100 nodes × 5MB test file = up to 500MB traffic. Slow nodes consume less, but fast nodes may download the full file.
