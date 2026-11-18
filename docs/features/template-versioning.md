English | [简体中文](template-versioning.zh-CN.md)

# Template Version History

SublinkPro keeps a content history for every template. Each save creates a new immutable version, any two versions can be compared, and any version can be downloaded or restored.

---

## ✨ What You Get

| Capability | Behavior |
|:---|:---|
| **Automatic snapshots** | Creating a template or saving the editor stores a new version automatically |
| **Unlimited history** | No retention limit; versions are kept until the template is deleted |
| **Side-by-side comparison** | Compare two versions, or one version against the current template content |
| **Download** | Download any snapshot as a normal template file (`<name>.v<version><ext>`) |
| **Restore** | Restore any version. Restore is append-only: it writes the old content as the current template and records a new `restore` version, so nothing is lost |
| **Content-only restore** | Restoring changes template content only. Category, rule source, proxy settings, and other metadata stay as they are |
| **Deleted with the template** | Deleting a template removes its files, metadata, and the whole version history |

---

## 🧭 How It Works

### Version sources

| Source | Created when |
|:---|:---|
| `baseline` | The first time an existing template is saved after versioning is enabled. It stores the content that was on disk before that save |
| `add` | A new template is created |
| `save` | Template content is saved without renaming |
| `rename` | Template is saved and its file name changes |
| `restore` | A historical version is restored |

### Where snapshots live

Template content stays in `template/<name>`. Snapshots are stored next to it:

```text
template/
├── clash.yaml
└── .versions/
    └── <template-id>/
        ├── 000001.snapshot
        ├── 000002.snapshot
        └── ...
```

Only version metadata (number, hash, size, source, author, timestamp) is kept in the database. The system backup archive already includes the `template/` directory, so snapshots are included in backups automatically.

### Duplicate saves

Saving content that is identical to the latest version does not create a new version. This keeps history readable when the editor is saved without real changes.

---

## 🖥️ Using It in the UI

1. Open **Template Management**.
2. Click the **history** icon on a template row.
3. The dialog lists all versions, newest first, with size, author, source, and a **Current** marker for the version that matches the template file on disk.
4. Use the row actions to:
   - **Preview** the version in a read-only editor.
   - **Download** it as a file.
   - **Restore** it (confirmation required).
5. Select one version and click **Compare with current**, or select two versions and click **Compare selected** to open a side-by-side diff with addition/deletion counters.

---

## 📡 API Summary

All routes require authentication and live under `/api/v1/template`.

| Method | Route | Purpose |
|:---|:---|:---|
| `GET` | `/versions?filename=&page=&pageSize=` | List versions (newest first). Items include `isCurrent` |
| `GET` | `/versions/:version?filename=` | Version metadata plus `content` |
| `GET` | `/versions/:version/download?filename=` | Download the snapshot file |
| `GET` | `/versions/diff?filename=&from=&to=` | Unified diff plus both contents. `to=0` compares against the current template file |
| `POST` | `/versions/restore` | Form: `filename`, `version`, optional `comment` |

`POST /template/add` and `POST /template/update` accept an optional `comment` form field that is stored with the new version.

---

## 📝 Notes

- Version history is per template, and renaming a template keeps its history because versions are keyed by template ID.
- Template files created outside the UI (for example, copied into `template/`) get their first history entry the next time they are saved.
- Snapshots contain template content only. Proxy links and other credentials are not duplicated into version records.
- Built-in templates are written to `template/` only on a fresh installation (when the database is created for the first time). If you delete a built-in template later, the deletion is permanent: restarting the service does not restore the files.
- Deleting a template whose file was already removed from `template/` still removes its metadata and version history.
