English | [简体中文](template-ai.zh-CN.md)

# AI Template Editing

SublinkPro provides an **AI assisted rewrite workflow** for the template editor. Describe the changes you want in natural language, let the system generate a candidate draft, then review, apply, or roll back through edit and diff views.

This feature emphasizes **human visibility, human confirmation, and rollback**. AI produces a candidate. You decide whether it is written into the template.

---

## ✨ Core Features

| Feature | Description |
|:---|:---|
| **🧠 Natural language template edits** | Type what you want to change, and AI generates a candidate draft based on the current template |
| **🪟 Floating command bar in editor** | The AI instruction box floats above the code editor without interrupting normal editing |
| **🔍 Edit / diff views** | Switch between normal editing and side by side diff review, like reviewing code changes |
| **✅ Local apply** | AI drafts never overwrite current content automatically. They enter the editor only after you click “Apply” |
| **↩️ Local rollback** | If the AI result is not satisfactory, roll back to the editor content from before the latest apply |
| **⚙️ System settings integration** | Template AI uses `Settings -> AI Assistant` as its generation configuration entry |

---

## 🧭 Suitable Use Cases

AI template editing is useful for:

- Adding new proxy groups, rule sections, or comments to an existing template
- Improving readability without changing the whole structure
- Turning an informal idea into a reviewable template draft
- Generating first, then comparing, to reduce manual YAML / INI editing cost

It is better for “**rewriting from an existing template**” than for replacing your review process.

---

## 🚀 What to Prepare First

To use Template AI, finish AI Assistant setup first.

Open:

`Settings -> AI Assistant`

At minimum, configure:

- **Base URL**
- **Model name**
- **API Key**

Optional **Max Tokens** defaults to `400000`. If you enter `0` on the settings page, the server default is used.

> [!IMPORTANT]
> The current AI Assistant **only supports services that provide a `/responses` endpoint**. If a service supports only `/chat/completions` or another compatibility layer, but not `/responses`, it cannot be used for AI template editing or connection tests in this project.

> [!TIP]
> The AI feature in the template editor uses the system level configuration saved in `Settings -> AI Assistant`. Configure it and test the connection first, then return to the template editor to generate drafts.

If the template page shows:

- `AI Assistant is currently unavailable`
- `AI settings are incomplete`

Open:

`Settings -> AI Assistant`

Check that AI is enabled and that Base URL, model name, and API Key are filled correctly.

Also confirm that the AI service behind the Base URL actually supports the `/responses` endpoint. Interfaces that only provide traditional `chat/completions` are currently unavailable.

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
- `Make this template more suitable for a normal Clash use case`

### 3. Generate a candidate draft

After clicking **Generate**, the system uses:

- Current template content
- Current category, such as Clash / Surge
- Current rule source and related configuration
- Your AI instruction

It generates a candidate draft.

After generation succeeds, the editor enters **diff mode** automatically so you can review changes directly.

> [!NOTE]
> During generation, the AI command input is temporarily disabled to prevent instruction changes during the same generation and avoid inconsistent state.

---

## 🔍 How to Review AI Results

The template editor supports two main views.

### Edit mode

- Used to edit the current template content directly
- Suitable for final manual adjustments
- Templates can be saved only in edit mode

### Diff mode

- Left side shows the original template before generation
- Right side shows the AI generated candidate draft
- Suitable for quickly reviewing changes like code diffs

If you only want to judge whether AI made the right change, start with **diff mode**.

If you decide to accept the generated result, click **Apply** to write the draft back into the editor.

---

## ✅ Apply and Rollback

### Apply

Click the **check button** in the floating command bar to write the current AI candidate draft into the editor.

After applying:

- Editor content becomes the AI draft
- You can continue manual editing
- You can save the template normally afterward

### Rollback

If you just applied an AI draft and want to return to the content from before applying, click the **rollback button**.

After rollback:

- Only the editor content from before the latest local apply is restored
- Other templates are unaffected
- This is useful for quick trial and error

> [!IMPORTANT]
> AI drafts are not written into templates automatically after generation. Candidate content enters the editor only after you click “Apply”.

---

## 💡 Practical Tips

### 1. Give AI clear boundaries first

Instead of simply saying “optimize this”, describe:

- What should be kept
- What should be added
- Which sections should not be touched

Example:

`Keep the existing comments and node placeholder structure. Only add one manual switching policy group for Japan.`

### 2. Split complex templates into steps

For long templates, split the request into multiple generations:

1. Adjust policy groups first.
2. Then organize rule sections.
3. Finally improve comments or naming.

This makes review easier and rollback safer.

### 3. Compare before applying

Recommended order:

1. Enter instruction.
2. Generate.
3. Review differences in diff mode.
4. Apply to the editor.
5. Manually adjust and save.

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

Because **diff mode** is a read only review view, not the final editing state.

Click **Apply** first to write candidate content back into the editor, then switch to edit mode and save.

### Why does a draft become invalid?

If you change any of these after generation:

- Template body
- File name
- Category
- Rule source
- Proxy download related configuration

The current AI draft may no longer match the latest context and should be generated again.

### Why is the rollback button unavailable?

Rollback applies only to the **latest local apply of an AI draft**.

If you have not clicked “Apply”, or if the local snapshot has been cleared, there is nothing to roll back.

---

## 🔐 Security and Usage Advice

- Treat AI output as a “candidate”, not the final answer.
- Review diffs first after generation, then decide whether to apply.
- Keep original versions of important templates and adjust gradually.
- If the result goes in the wrong direction, change the instruction and regenerate instead of stacking edits on a bad draft.

> [!TIP]
> The ideal use is not “let AI generate the final template in one shot”. It is “let AI produce a strong draft, then quickly review, apply, and fine tune it yourself”. This is both efficient and safer.
