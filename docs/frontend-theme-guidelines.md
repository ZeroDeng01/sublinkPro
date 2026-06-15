# Frontend Theme Guidelines

## Overview

All frontend UI changes involving colors, surfaces, shadows, borders, icons, dialogs, panels, preview cards, overlays, notices, chips, rich text, or status colors must be checked in both light and dark modes, not only in the mode currently visible.

## Theme Source of Truth

- Prefer existing tokens from `theme.vars?.palette` and `theme.palette`.
- Prefer `background.default`, `background.paper`, `divider`, `text.primary`, and `text.secondary` for surfaces and text hierarchy.
- Use existing semantic palette colors for emphasis or state: `primary`, `secondary`, `info`, `success`, `warning`, and `error`.
- Express opacity and elevation with `alpha()` or existing helpers like `withAlpha()` instead of inventing a parallel color system.

## Implementation Rules

- **Outer cards, dialog bodies, and main panels** should generally use `background.paper`; in dark mode, light gradients, soft borders, and subtle inner highlights are acceptable, but avoid large bright slabs.
- **Nested panels, bottom strips, preview footers, and notice blocks** should generally use `background.default` or lightly tinted semantic surfaces; do not create white-on-white layering.
- **Main values, titles, and key interactive text** should rely on `text.primary` and font weight; in dark mode, a higher alpha is acceptable, but do not default everything to pure white.
- **Secondary descriptions, helper text, percentages, and hints** should use `text.secondary` or a lower-alpha `text.primary`, rather than introducing ad-hoc gray constants.
- **High-frequency interactive elements** such as icon buttons, header actions, notification entry points, and theme toggles must be checked in light/dark across default, hover, active/selected, and disabled states.
- **Rich text, markdown, code blocks, lists, and footer overlays**: do not stop at fixing the container background; also verify body text, list items, code, links, markers, and footer text contrast.

## Explicit Anti-Patterns

- Do not hardcode `#fff`, `white`, `#000`, or `black` for large surfaces unless it is truly brand/icon-specific and cannot be expressed with theme tokens.
- Do not solve dark-mode readability by turning all text pure white.
- Do not create multiple unrelated `isDark ? ... : ...` magic-number surface styles for equivalent UI layers within one page.
- Do not fix only one component instance while equivalent previews, dialogs, footers, or chips elsewhere still use outdated semantics.

## Recommended Workflow

1. First look for nearby helpers, reusable components, or similar `sx` patterns before deciding to add a small shared helper.
2. If the same theme problem appears in 2 or more files, prefer a small focused helper over copy-pasting color expressions.
3. In your delivery summary, explicitly state which light/dark scenarios were checked, which surface/text/icon layers changed, and which related layers were inspected but needed no modification.

## Theme Coverage Expectations

### Scope

Theme adaptation does not stop after fixing the one visible component; if the change affects a page, module, shared helper, theme override, or reusable component, inspect all related pages, dialogs, drawers, preview cards, nested panels, footers, overlays, and mobile entry points in that feature flow.

### Shared Infrastructure

If you change shared theme infrastructure such as `webs/src/themes/*`, `webs/src/utils/colorUtils.js`, shared `sx` helpers, or common dialog/card/panel styles, treat the default blast radius as every page reusing that pattern rather than only the page touched by the current task.

### Variant Coverage

If a module has desktop and mobile variants, dialog and fullscreen mobile modes, drawer and inline-panel variants, or both list-card and detail-panel entry points, theme work must cover the full set rather than only the desktop or primary path.

### High-Risk Components

Components using `useMediaQuery`, `fullScreen`, `Drawer`, `Popover`, `Menu`, `Tooltip`, `Collapse`, `Tabs`, `Stepper`, `ReactMarkdown`, `code` / `pre`, custom footer overlays, or sticky action bars should be treated as high-risk theme-adaptation areas by default.

### Complete Structure Checks

Dialogs and overlays must be checked as full structures: triggers, paper, title, content, actions, backdrop, scroll containers, nested panels, close buttons, footer actions, and fullscreen mobile variants for `Dialog` / `Drawer` / `Popover` / `Menu` / `Tooltip`; changing only the container background is not enough.

### Reference Patterns

When a mature dark/light adaptation pattern already exists in the repo, reuse its surface layering, border opacity, state colors, and mobile handling instead of inventing a new local rule set. Current primary reference clusters include:

- Node preview: `NodePreviewDialog.jsx`, `NodePreviewCard.jsx`, `NodePreviewDetailsPanel.jsx`
- Chain proxy: `ChainProxyDialog.jsx`, `ChainPreviewDialog.jsx`, `MobileChainBuilder.jsx`, `ConditionBuilder.jsx`
- Common UI: `layout/MainLayout/Header/NotificationSection/index.jsx`, `views/dashboard/Default/index.jsx`
- Airport management: `views/airports/index.jsx`, `AirportMobileList.jsx`, `AirportFormDialog.jsx`, `AirportBatchEditDialog.jsx`, `components/CronExpressionGenerator.jsx`

## High-Risk Misses When Migrating from Light to Dark

- Only the outer page container was updated, while inner header strips, footer strips, nested panels, stat blocks, detail blocks, empty states, skeletons, or preview footers were left behind.
- Text color was updated, but borders, shadows, dividers, hover, active, selected, disabled, focus-visible states, and scroll-container edges were not.
- The main page was fixed, but the same module's dialogs, drawers, popovers, menus, tooltips, fullscreen mobile dialogs, bottom menus, or secondary detail entry points were not.
- Desktop UI was fixed, but mobile cards, responsive collapsed layouts, safe areas, sticky footers, horizontal scrolling, tap targets, and icon-button visibility were not checked.
- Containers were updated without also fixing `ReactMarkdown`, `code`, `pre`, list markers, links, tables, chips, badges, tags, status notices, and notification overlays.
- State colors kept light-mode tints, making success / warning / error / info backgrounds indistinct, text muddy, or borders dirty in dark mode.
- Icons, secondary text, percentages, captions, placeholders, helper text, and disabled copy were not adjusted with the main text hierarchy, resulting in readable titles but washed-out supporting content.
- DataGrid, Autocomplete, Select, ToggleButtonGroup, Tabs, hovered list rows, search bars, filter bars, and form controls were only fixed in their default state without validating popups, selected states, and focus states.

## Dark-Mode Migration Principles

- Prefer the repo's established pattern: start with `const palette = theme.vars?.palette || theme.palette;`, then derive local semantic variables such as `dialogSurface`, `mutedPanelSurface`, `nestedPanelSurface`, and `panelBorder` instead of scattering many `isDark ? ... : ...` expressions across JSX.
- In dark mode, prioritize semantic surface layering rather than replacing light backgrounds with black. Outer containers should usually use `background.paper`, while inner panels, footers, notices, and muted sections should prefer `background.default` or low-opacity semantic tinting.
- Express borders, dividers, and elevation with `divider`, `alpha()`, `withAlpha()`, and subtle inset highlights instead of inventing a new gray scale in each component.
- Main text, emphasis text, secondary text, placeholders, disabled copy, icons, and state colors must preserve stable hierarchy; do not hide hierarchy problems by simply brightening everything.
- State colors should generally use the combination of semantic text, low-opacity backgrounds, and light borders rather than large high-saturation solid fills in dark mode.
- When the same issue appears in multiple files, extract a small shared helper or reuse an existing override instead of copying the same dark-mode magic numbers into multiple modules.

## Theme Migration Playbook

When performing theme adaptation in this repo, follow this sequence by default and do not skip steps:

1. **Define blast radius**: Determine whether the change affects a page, a module, a shared component, a theme override, or a global theme helper; if it is one of the latter two, default to checking every page reusing that pattern.

2. **Find reference implementation before writing code**: First reuse the surface layering and state treatment from `NodePreview*`, `Chain*`, `MobileChainBuilder.jsx`, `ConditionBuilder.jsx`, `NotificationSection`, `dashboard/Default`, and the airport-management cluster (`airports/index.jsx`, `AirportMobileList.jsx`, `AirportFormDialog.jsx`, `AirportBatchEditDialog.jsx`, `CronExpressionGenerator.jsx`), then decide whether a new helper is actually needed.

3. **Fix semantic layering before tweaking local colors**: Handle outer surfaces, inner panels, footer/header strips, borders, dividers, text hierarchy, state colors, and icons before touching local gradients, shadows, or decorative polish.

4. **Check branch paths when changing main path**: After desktop, continue through mobile, fullscreen dialogs, drawers, popovers, menus, tooltips, collapse states, empty states, skeletons, footer overlays, markdown/code/pre, hovered lists, and focused form controls.

5. **Document coverage before delivery**: Explicitly state which pages/modules/overlays/mobile entry points were covered, which existing repo patterns were reused, and which similar instances were checked but did not need changes.

## Dialog Content Spacing Pattern (MANDATORY)

**Problem**: Content in dialogs often appears cramped against the title area, creating poor visual hierarchy.

**Solution**: Use margin-top on the first child container instead of padding-top on DialogContent.

### Correct Pattern

```jsx
<DialogContent sx={{ px: 2.5, bgcolor: dialogSurface }}>
  <Stack spacing={3} sx={{ mt: 2 }}>
    {/* Dialog content */}
  </Stack>
</DialogContent>
```

### Incorrect Patterns (DO NOT USE)

```jsx
// ❌ Using pt on DialogContent - gets overridden by MUI defaults
<DialogContent sx={{ px: 2.5, pt: 3, bgcolor: dialogSurface }}>
  <Stack spacing={3}>
    {/* Content appears cramped */}
  </Stack>
</DialogContent>

// ❌ Using py on DialogContent - inconsistent with project patterns
<DialogContent sx={{ px: 2.5, py: 3, bgcolor: dialogSurface }}>
  <Stack spacing={3}>
    {/* Content appears cramped */}
  </Stack>
</DialogContent>
```

### Reference Examples

Good examples to follow:
- `ShareBatchCreateDialog.jsx` - Uses `<Stack spacing={3} sx={{ mt: 1 }}>`
- `ShareBatchUpdateDialog.jsx` - Uses `<Stack spacing={3} sx={{ mt: 1 }}>`
- `SubscriptionFormDialog.jsx` - Uses proper margin-top patterns

### Guidelines

1. **Always use `mt` on the first child container** (Stack, Box, etc.) inside DialogContent
2. **Do not use `pt` or `py` on DialogContent** to create title spacing - it will be overridden
3. **Standard margin-top values**:
   - `mt: 1` (8px) - Minimal spacing for compact dialogs
   - `mt: 2` (16px) - Standard spacing for most dialogs (recommended)
   - `mt: 3` (24px) - Extra spacing for important dialogs with lots of content
4. **Keep horizontal padding consistent**: Use `px: 2.5` on DialogContent (matches DialogTitle)
5. **Bottom spacing**: Usually handled by DialogContent's default bottom padding - no need to override

If you do not have time to inspect a module's desktop, mobile, dialog/drawer, overlay, and detail-panel variants together, do not claim the module is "theme-adapted"; mark it explicitly as partial coverage instead.

## Theme-Related Definition of Done

- Check both light and dark modes.
- Check relevant hover / active / disabled or expanded states.
- Do not ship a state where the container is fixed but inner panels, footers, markdown, percentages, or icons remain too bright or too dim to read.
- Do not ship theme work in a half-finished state where the current page looks correct but dialogs, mobile variants, or overlays in the same module still use the old semantics.
- Whenever the change touches shared theme infrastructure, reusable components, or module-level visual semantics, the delivery summary must explicitly list the covered pages, overlays, mobile entry points, and the related instances that were checked but did not require changes.
