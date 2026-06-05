English | [简体中文](host.zh-CN.md)

# Host Mapping Management

Custom Host mappings work like a system hosts file, letting you assign fixed IP addresses to specific domains.

---

## Core Features

| Feature | Description |
|:---|:---|
| **Global effect** | Host mappings automatically apply to speed tests, subscription imports, Telegram Bot operations, and other proxy related flows |
| **Subscription address replacement** | Subscriptions can enable “replace server address with Host”, automatically replacing node domains with configured Host IPs |
| **DNS servers** | Supports custom DNS servers, including DoH and UDP, with optional proxy node resolution |
| **Speed test persistence** | DNS results from speed tests can be saved as Host mappings to avoid repeated resolution |
| **Text edit mode** | Supports hosts file like batch editing, one line per entry: `domain IP # note` |
| **Expiration management** | Automatically created Hosts can have expiration times and be cleaned up after expiry |
| **Pin** | Important Hosts can be pinned so they are not removed by expiration cleanup |

---

## 📝 Use Cases

```text
Case 1: CDN preferred IPs
├── Map Cloudflare node domains to preferred IPs
└── Save the best resolved result automatically after speed tests

Case 2: DNS pollution bypass
├── Manually set the correct IP for polluted domains
└── Use the correct resolution in all proxy operations

Case 3: Batch import
├── Edit many mappings at once in text mode
└── Format: example.com 192.168.1.1 # note
```

---

## Speed Test Host Persistence

> [!TIP]
> **Speed test Host persistence**: When enabled, speed tests automatically save proxy node domains and their resolved IPs into the Host table. Use it with expiration settings to avoid unbounded Host data growth.

---

## DNS Proxy Settings

> [!NOTE]
> **DNS proxy settings**: Configure whether DNS resolution goes through a proxy. You can auto select an available proxy or manually specify a proxy node. This is useful for resolving blocked domains.

---

## Text Edit Format

Use this format in text mode batch editing:

```text
# This is a comment line
example.com 192.168.1.1
cdn.example.com 192.168.1.2 # note
```

Each line is one record: `domain IP [# note]`.
