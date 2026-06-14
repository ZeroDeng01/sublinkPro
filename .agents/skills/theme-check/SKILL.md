---
name: theme-check
description: "Theme adaptation checklist for light/dark mode and UI component changes. Use when modifying colors, surfaces, dialogs, or theme infrastructure. Not for simple text changes."
version: "1.0.0"
author: "SublinkPro Team"
user-invocable: true
---

# Theme Check Skill

Comprehensive checklist for frontend theme adaptation changes.

## When to use this skill

Use this skill when:
- Modifying colors, surfaces, shadows, borders, icons, dialogs, panels, preview cards, overlays, notices, chips, rich text, or status colors
- Changing theme infrastructure (`webs/src/themes/*`, `webs/src/utils/colorUtils.js`)
- Updating shared `sx` helpers or common dialog/card/panel styles
- Working on any UI component that affects light/dark mode presentation

## Prerequisites

Before running this checklist:
1. Read `docs/frontend-theme-guidelines.md` for complete theme rules
2. Identify the blast radius: page, module, shared component, or theme infrastructure
3. Find reference patterns in the codebase (NodePreview*, Chain*, Airport* clusters)

## Checklist

### 1. Mode Coverage
- [ ] Checked both light and dark modes (not just visual validation in current mode)
- [ ] Verified with actual theme toggle, not just screenshots

### 2. Device Coverage
- [ ] Checked desktop view
- [ ] Checked mobile view
- [ ] Verified responsive breakpoints (`useMediaQuery`)
- [ ] Tested fullscreen dialogs on mobile
- [ ] Checked safe-area handling

### 3. Component Structure Coverage
- [ ] Page body
- [ ] Dialogs (`Dialog` component)
- [ ] Drawers (`Drawer` component)
- [ ] Popovers (`Popover` component)
- [ ] Menus (`Menu` component)
- [ ] Tooltips (`Tooltip` component)
- [ ] Preview cards
- [ ] Detail panels
- [ ] Footer overlays
- [ ] Empty states
- [ ] Skeletons
- [ ] Notification/overlay entry points

### 4. Text Hierarchy
- [ ] Primary text (`text.primary`)
- [ ] Secondary text (`text.secondary`)
- [ ] Icons
- [ ] Chips
- [ ] Badges
- [ ] Links
- [ ] Code blocks (`code`, `pre`)
- [ ] Placeholders
- [ ] Helper text
- [ ] Disabled copy
- [ ] State colors (success/warning/error/info)

### 5. Interactive States
- [ ] Default state
- [ ] Hover state
- [ ] Active/selected state
- [ ] Disabled state
- [ ] Focus-visible state
- [ ] Expanded/collapsed state

### 6. Surface Layering
- [ ] Outer containers use `background.paper`
- [ ] Nested panels use `background.default` or tinted surfaces
- [ ] No white-on-white layering
- [ ] Borders use `divider` or `alpha()`/`withAlpha()`
- [ ] No hardcoded `#fff`, `white`, `#000`, `black` for large surfaces

### 7. Component-Specific Checks

#### Rich Content
- [ ] `ReactMarkdown` content
- [ ] Code blocks
- [ ] List markers
- [ ] Tables
- [ ] Footer overlays

#### Form Controls
- [ ] DataGrid
- [ ] Autocomplete (including popups)
- [ ] Select (including dropdowns)
- [ ] ToggleButtonGroup
- [ ] Tabs
- [ ] Search bars
- [ ] Filter bars
- [ ] Focused states

### 8. Mobile-Specific
- [ ] Responsive collapsed layouts
- [ ] Sticky footers
- [ ] Horizontal scrolling areas
- [ ] Tap target sizes
- [ ] Icon button visibility
- [ ] Bottom action bars
- [ ] Safe-area padding

### 9. Code Quality
- [ ] No `isDark ? ... : ...` magic numbers scattered in JSX
- [ ] Used semantic variables (`dialogSurface`, `mutedPanelSurface`, etc.)
- [ ] Reused existing helpers or created focused helper (not copy-paste)
- [ ] No new dark-mode magic numbers in multiple files

### 10. Reference Patterns Used
Document which existing patterns were followed:
- [ ] NodePreview cluster (`NodePreviewDialog.jsx`, `NodePreviewCard.jsx`, `NodePreviewDetailsPanel.jsx`)
- [ ] Chain proxy cluster (`ChainProxyDialog.jsx`, `ChainPreviewDialog.jsx`, `MobileChainBuilder.jsx`, `ConditionBuilder.jsx`)
- [ ] Common UI (`layout/MainLayout/Header/NotificationSection/index.jsx`, `views/dashboard/Default/index.jsx`)
- [ ] Airport management cluster (`views/airports/index.jsx`, `AirportMobileList.jsx`, `AirportFormDialog.jsx`, `AirportBatchEditDialog.jsx`, `components/CronExpressionGenerator.jsx`)

## Delivery Requirements

Before marking complete, the change summary must include:

1. **Coverage statement**: Which pages/modules/overlays/mobile entry points were checked
2. **Reference patterns**: Which existing repo patterns were reused
3. **Inspection results**: Which similar instances were checked but did not need changes
4. **Mode verification**: Explicit confirmation of light/dark testing method (not just screenshots)

## Anti-patterns to avoid

- ❌ Only fixing the currently visible component
- ❌ Only checking desktop without mobile
- ❌ Only checking default state without hover/active/disabled
- ❌ Only updating container background without inner text/borders/icons
- ❌ Claiming "theme-adapted" when dialogs/mobile variants still use old semantics

## When to mark as partial coverage

If you do NOT have time to check all component variants (desktop, mobile, dialog, drawer, overlay, details panel), mark the work as **partial coverage** rather than claiming full theme adaptation.

## Exit criteria

✅ Can exit when:
- All relevant checklist items are verified
- Both light and dark modes tested
- Desktop and mobile variants checked
- All interactive states verified
- Delivery summary documents coverage

❌ Cannot exit when:
- Only current page fixed but same module's dialogs/overlays unverified
- Container fixed but inner panels/text/icons still unreadable
- Desktop fixed but mobile unchecked
- Only default state checked without interactive states
