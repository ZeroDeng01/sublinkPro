English | [简体中文](template-ai.zh-CN.md)

# AI Template Editing

SublinkPro provides an **AI assisted edit-session workflow** for the template editor. Describe the change you want, let AI propose structured edit operations, then review the server-built preview before accepting it into the normal editor.

This feature keeps the user in control. AI doesn't return a full replacement template as its output. It returns precise `replace`, `insert`, and `delete` operations. The server applies those operations to the current template, validates the preview, and shows you a read-only diff.

---

## ✨ Core Features

| Feature | Description |
|:---|:---|
| **🧠 Natural language template edits** | Type what you want to change, and AI proposes structured operations against the current template |
| **🪟 Floating command bar in editor** | The AI instruction box floats above the code editor without interrupting normal editing |
| **🔍 Read-only preview diff** | Review the original template and server-built candidate side by side before accepting anything |
| **✅ Accept into editor** | Accepting a preview copies the candidate into the editor. It doesn't save the template automatically |
| **⚠️ Validation warnings** | Validation errors block acceptance. Warnings highlight items to review but don't require extra confirmation |
| **🧹 Discard preview** | Discard an edit session when the preview isn't useful or has expired |
| **⚙️ System settings integration** | Template AI uses `Settings -> AI Assistant` as its configuration entry |

---

## 🧭 Suitable Use Cases

AI template editing is useful for:

- Adding new proxy groups, rule sections, or comments to an existing template
- Improving readability without changing the whole structure
- Making targeted edits to a long template without asking the model to rewrite the whole file
- Reviewing exact changes in a diff before the editor content changes

It is better for “**editing an existing template through reviewed operations**” than for replacing your review process.

---

## 🚀 What to Prepare First

To use Template AI, finish AI Assistant setup first.

Open:

`Settings -> AI Assistant`

At minimum, configure:

- **Base URL**
- **Model name**
- **API Key**

Choose the **Interface type** that matches your provider:

- **Responses API** calls `/responses`.
- **Chat Completions API** calls `/chat/completions`.

Optional **Max Tokens** defaults to `400000`. If you enter `0` on the settings page, the server default is used.

> [!TIP]
> The AI feature in the template editor uses the system level configuration saved in `Settings -> AI Assistant`. Configure it and test the connection first, then return to the template editor.

If the template page shows:

- `AI Assistant is currently unavailable`
- `AI settings are incomplete`

Open:

`Settings -> AI Assistant`

Check that AI is enabled and that Base URL, model name, and API Key are filled correctly.

Also confirm that the AI service behind the Base URL supports the selected interface type.

---

## 📝 Usage Flow

### 1. Open the template editor

From the template management page, you can:

- Create a new template
- Edit an existing template

Prepare a base template in the editor before entering AI instructions for more stable results.

### 2. Enter AI instructions

Use natural language in the floating AI command bar above the editor to describe the change you want.

Examples:

- `Keep the existing structure and add an auto select policy group for Hong Kong nodes`
- `Do not remove comments. Help me make the rules section clearer`
- `Only replace the DNS comment block. Leave proxy groups unchanged`
- `Delete all USA node entries from every proxy group`

### 3. Start an edit session

After clicking **Generate**, the system creates a short-lived edit session through:

- `POST /api/v1/template/ai/edit-sessions/stream`

The request includes the current template text, file name, category, rule source, proxy options, include-all setting, and your prompt. The model receives enough context to propose targeted operations, especially for long templates, but the server remains the source of truth for the preview.

The model returns an operation list only. Supported operations in v1 are:

- `replace`: exact `oldString` is replaced by `newString`
- `insert`: `newString` is inserted `before` or `after` an exact `anchor`
- `delete`: exact `oldString` is removed

`replace` and `delete` can also include optional `match`. When `match` is omitted, the default is `unique`: the exact target must appear once, and duplicates still fail with `PATCH_AMBIGUOUS_MATCH`. When the user explicitly asks for all/every occurrences, such as “delete all USA node entries”, the model may use `"match":"all"` so the server applies the exact `oldString` to every occurrence. If no occurrence exists, the preview fails with `PATCH_NO_MATCH`. `insert` does not support `"match":"all"`; insertion anchors must remain exact and unique.

v1 applies operations to exact text targets in the current template. The server materializes the candidate from those operations, then validates the candidate before the preview becomes ready. It does not use fuzzy matching, regular expressions, YAML paths, AST edits, or line-number-only edits.

### 4. Review the server preview

The server applies the operations atomically. If one operation can't be applied, no preview is accepted as ready.

When the preview is ready, the editor enters **diff mode** automatically. This view is read-only. The left side shows the original base text. The right side shows the server-built candidate.

### 5. Accept or discard

If validation has errors, acceptance is blocked. If validation has warnings only, you can still accept the preview after reviewing them. Warnings are attention metadata, not an extra confirmation gate.

Accepting the preview copies the candidate into the editor. It doesn't save the template to disk. Use the normal template save action, including the usual save confirmation when shown, after any final edits.

You can accept more than one AI preview before saving. After the first accept, the next edit session is based on the editor text that already includes the unsaved accepted change. When accepting that next preview, the client sends the current editor text as `currentText` so the server can prove the editor base still matches the session base. `currentText` is not saved, and it is not the candidate output. It is only base proof for safe consecutive accepts.

Discarding a preview ends the edit session without changing the editor content.

---

## 🔍 How to Review AI Results

The template editor supports two main views.

### Edit mode

- Used to edit the current template content directly
- Suitable for final manual adjustments
- Templates can be saved only in edit mode

### Diff mode

- Left side shows the original template snapshot for the edit session
- Right side shows the server-built candidate preview
- Suitable for reviewing changes like code diffs
- Read-only by design, so save remains unavailable in this view

If you only want to judge whether AI made the right change, start with **diff mode**.

If you decide to accept the result, click **Accept preview** to copy the candidate into the editor.

---

## ✅ Accept, Discard, and Save

### Accept preview

Click the **check button** in the floating command bar to accept the current preview.

After accepting:

- Editor content becomes the preview candidate
- You can continue manual editing
- You can generate and accept another AI preview before saving
- You can save the template normally afterward
- The accept action itself doesn't persist the template

When multiple previews are accepted before a save, each accept still checks that the session base is current. The editor text sent as `currentText` proves that base when it matches the session. If it is missing or doesn't match, the server checks the saved template file instead. If the editor no longer matches the session base and the saved file changed too, accept is blocked with `AI_EDIT_STALE_BASE` so an old preview can't overwrite newer work.

### Validation warnings

Warnings may appear when the preview is usable but needs extra attention, such as a rule-source warning. They tell you what to review before you accept or save.

Warnings don't block accept. Only validation errors prevent accepting the preview into the editor.

### Discard preview

Use discard when the preview isn't useful or you want to start over. Discarding clears the edit session state and doesn't change the editor content.

> [!IMPORTANT]
> AI edits are not written into templates automatically. A preview enters the editor only after you accept it, and the template is saved only through the normal save action.

---

## 📡 API Contract Summary

The edit-session API is available under `/api/v1/template/ai` and requires authentication.

Final v1 routes:

- `POST /api/v1/template/ai/edit-sessions/stream`: create an edit session and stream operation generation, server patching, preview validation, and preview readiness
- `GET /api/v1/template/ai/edit-sessions/:sessionId`: read the current session preview state
- `POST /api/v1/template/ai/edit-sessions/:sessionId/accept`: accept a ready preview and return the candidate text for the editor. Clients should include optional `currentText` with the current editor text as base proof, especially after previous unsaved accepts
- `POST /api/v1/template/ai/edit-sessions/:sessionId/discard`: discard a session

Streaming events:

- `template.edit.session.created`
- `template.edit.model.delta`
- `template.edit.operations.ready`
- `template.edit.preview.validating`
- `template.edit.preview.ready`
- `template.edit.warning`
- `template.edit.error`
- `template.edit.completed`

Sessions are short-lived. The v1 TTL is `15m`, cleanup runs periodically, and sessions are in memory only. They are not persistent edit history. If a session expires, start a new one.

The old full-template generation contract is removed. The model output must not be a full-template `candidateText`. Any `candidateText` field in the new flow is the server-materialized preview returned to the editor after operations were applied and validated.

---

## 💡 Practical Tips

### 1. Give AI clear boundaries first

Instead of simply saying “optimize this”, describe:

- What should be kept
- What should be added
- Which sections should not be touched

Example:

`Keep the existing comments and node placeholder structure. Only add one manual switching policy group for Japan.`

### 2. Ask for targeted edits on long templates

For long templates, ask for one focused edit at a time:

1. Adjust policy groups first.
2. Then organize rule sections.
3. Finally improve comments or naming.

The system targets structured operations against exact text, which is safer than asking the model to rewrite the whole file.

### 3. Compare before accepting

Recommended order:

1. Enter instruction.
2. Generate an edit session preview from operation based model output.
3. Review the server validated, read-only diff.
4. Review any warnings shown with the preview.
5. Accept the preview into the editor.
6. Optionally generate and accept another preview from the updated editor text.
7. Manually adjust and use the normal template save action.

This is the safest flow and matches the feature design.

---

## 🛠️ FAQ

### Generation reports “AI Assistant is currently unavailable”

The AI Assistant is not enabled in the current system.

Open:

`Settings -> AI Assistant`

Enable AI Assistant and save settings.

### Generation reports “AI settings are incomplete”

Usually at least one of these settings is missing:

- Base URL
- Model name
- API Key

Open:

`Settings -> AI Assistant`

Complete configuration, then return to generation.

### Why can't I save the diff result directly?

Because **diff mode** is a read-only review view, not the final editing state.

Click **Accept preview** first to copy candidate content into the editor, then switch to edit mode and save.

### Why does an edit session expire?

Edit sessions are short-lived server-owned previews. They expire after `15m` and are not saved as history.

If a session expires, generate a new preview from the latest editor content.

### Why is acceptance blocked?

Acceptance can be blocked when:

- The session expired or was discarded
- The preview isn't ready
- The base template changed since the session was created
- The editor text doesn't match the session base and the saved template file changed too
- Validation returned errors

### Why do I see `PATCH_AMBIGUOUS_MATCH`?

The default operation mode is `unique`. If the same exact target appears more than once and the instruction didn't clearly ask for every occurrence, the server rejects the operation so AI doesn't silently edit unintended sections.

For requests like `Delete all USA node entries`, the model can use explicit `"match":"all"` on `replace` or `delete` so every exact occurrence is changed atomically. If the preview still fails, regenerate with wording such as `all`, `every`, `全部`, or `所有`, and include the exact visible target text when possible.

### Can AI rewrite the whole template for me?

The new contract doesn't ask AI to return a full replacement template. AI returns edit operations, and the server builds the candidate preview. This reduces token waste and makes review more reliable.

---

## 🔐 Security and Usage Advice

- Treat the preview as a candidate, not the final answer.
- Review diffs first after generation, then decide whether to accept.
- Read warnings as review guidance. They don't block accept.
- Keep original versions of important templates and adjust gradually.
- If the result goes in the wrong direction, discard it and regenerate with a clearer instruction.

> [!TIP]
> The ideal use is not “let AI produce the final template in one shot”. It is “let AI propose exact edits, then quickly review, accept, and fine tune them yourself”. This is both efficient and safer.
