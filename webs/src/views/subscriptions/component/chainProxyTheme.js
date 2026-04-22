import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from '../../../utils/colorUtils';

export function getChainProxyThemeTokens(theme, isDark) {
  const surfaceTokens = getSurfaceTokens(theme, isDark);
  const textTokens = getReadableTextTokens(theme, isDark);
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = surfaceTokens;
  const getSemanticSurface = (color, darkOpacity = 0.16, lightOpacity = 0.08) => withAlpha(color, isDark ? darkOpacity : lightOpacity);
  const getSemanticBorder = (color, darkOpacity = 0.34, lightOpacity = 0.2) => withAlpha(color, isDark ? darkOpacity : lightOpacity);

  return {
    ...surfaceTokens,
    ...textTokens,
    dialogSurface,
    dialogSurfaceGradient,
    mutedPanelSurface,
    nestedPanelSurface,
    panelBorder,
    containerSurface: isDark ? withAlpha(palette.background.paper, 0.3) : withAlpha(dialogSurface, 0.98),
    elevatedSurface: isDark ? withAlpha(palette.background.default, 0.92) : withAlpha(palette.background.paper, 0.98),
    fieldSurface: isDark ? withAlpha(palette.background.paper, 0.18) : withAlpha(palette.background.paper, 0.92),
    fieldSurfaceActive: isDark ? withAlpha(palette.background.paper, 0.28) : palette.background.paper,
    insetHighlight: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none',
    cardShadow: isDark ? 'none' : theme.shadows[1],
    panelShadow: isDark ? 'none' : `0 12px 28px ${withAlpha(palette.common.black, 0.08)}`,
    softBorder: isDark ? withAlpha(palette.divider, 0.64) : withAlpha(palette.divider, 0.82),
    subtleBorder: isDark ? withAlpha(palette.divider, 0.48) : withAlpha(palette.divider, 0.72),
    primarySurface: getSemanticSurface(palette.primary.main),
    secondarySurface: getSemanticSurface(palette.secondary.main),
    infoSurface: getSemanticSurface(palette.info.main),
    successSurface: getSemanticSurface(palette.success.main),
    warningSurface: getSemanticSurface(palette.warning.main),
    errorSurface: withAlpha(palette.error.main, isDark ? 0.14 : 0.08),
    primarySoftBorder: withAlpha(palette.primary.main, isDark ? 0.34 : 0.2),
    primaryStrongBorder: withAlpha(palette.primary.main, isDark ? 0.46 : 0.28),
    secondarySoftBorder: getSemanticBorder(palette.secondary.main),
    infoSoftBorder: getSemanticBorder(palette.info.main),
    successSoftBorder: getSemanticBorder(palette.success.main),
    warningSoftBorder: getSemanticBorder(palette.warning.main),
    errorSoftBorder: getSemanticBorder(palette.error.main, 0.3, 0.18),
    hoverSurface: withAlpha(palette.primary.main, isDark ? 0.12 : 0.06),
    selectedSurface: withAlpha(palette.primary.main, isDark ? 0.18 : 0.1),
    warningSoftText: isDark ? withAlpha(palette.warning.light || palette.warning.main, 0.94) : palette.warning.dark,
    disabledSurface: withAlpha(palette.text.secondary, isDark ? 0.12 : 0.08),
    disabledBorder: withAlpha(palette.text.secondary, isDark ? 0.24 : 0.22),
    coveredSurface: withAlpha(palette.warning.main, isDark ? 0.18 : 0.12),
    coveredBorder: withAlpha(palette.warning.main, isDark ? 0.38 : 0.34),
    canvasBaseSurface: isDark ? withAlpha(palette.background.default, 0.84) : palette.background.paper,
    canvasElevatedSurface: isDark ? withAlpha(palette.background.paper, 0.3) : withAlpha(palette.background.paper, 0.98),
    canvasGrid: isDark ? withAlpha(palette.divider, 0.12) : withAlpha(palette.divider, 0.38),
    canvasHoverSurface: isDark ? withAlpha(palette.background.paper, 0.22) : withAlpha(palette.background.default, 0.92),
    canvasShadow: withAlpha(palette.text.primary, isDark ? 0.24 : 0.12),
    canvasChipSurface: isDark ? withAlpha(palette.background.paper, 0.22) : withAlpha(palette.background.default, 0.92),
    tableHeaderSurface: isDark ? withAlpha(palette.background.default, 0.92) : withAlpha(palette.background.default, 0.9)
  };
}

export function getChainProxyCanvasCssVars(tokens) {
  return {
    '--canvas-bg': tokens.containerSurface,
    '--canvas-surface': tokens.canvasBaseSurface,
    '--canvas-surface-strong': tokens.canvasElevatedSurface,
    '--canvas-panel-surface': tokens.nestedPanelSurface,
    '--canvas-border': tokens.softBorder,
    '--canvas-grid': tokens.canvasGrid,
    '--canvas-muted': tokens.secondaryText,
    '--canvas-text': tokens.primaryText,
    '--canvas-text-soft': tokens.tertiaryText,
    '--canvas-hover': tokens.canvasHoverSurface,
    '--canvas-shadow': tokens.canvasShadow,
    '--canvas-primary-soft': tokens.primarySurface,
    '--canvas-primary-border': tokens.primarySoftBorder,
    '--canvas-primary-strong': tokens.palette.primary.main,
    '--canvas-secondary-soft': tokens.secondarySurface,
    '--canvas-secondary-border': tokens.secondarySoftBorder,
    '--canvas-secondary-strong': tokens.palette.secondary.main,
    '--canvas-warning-soft': tokens.warningSurface,
    '--canvas-warning-border': tokens.warningSoftBorder,
    '--canvas-warning-strong': tokens.palette.warning.main,
    '--canvas-warning-main': tokens.palette.warning.main,
    '--canvas-success-soft': tokens.successSurface,
    '--canvas-success-border': tokens.successSoftBorder,
    '--canvas-success-strong': tokens.palette.success.main,
    '--canvas-disabled-soft': tokens.disabledSurface,
    '--canvas-disabled-border': tokens.disabledBorder,
    '--canvas-covered-soft': tokens.coveredSurface,
    '--canvas-covered-border': tokens.coveredBorder,
    '--canvas-chip-bg': tokens.canvasChipSurface,
    '--canvas-handle': tokens.palette.primary.main,
    '--canvas-panel-header': tokens.tableHeaderSurface,
    '--canvas-primary-main': tokens.palette.primary.main,
    '--canvas-shadow-soft': tokens.primarySoftBorder,
    '--canvas-info-soft': tokens.infoSurface,
    '--canvas-info-border': tokens.infoSoftBorder
  };
}

export function getChainProxyFieldControlSx(tokens, accentColor = tokens.palette.primary.main) {
  return {
    '& .MuiInputLabel-root': { color: tokens.secondaryText },
    '& .MuiInputLabel-root.Mui-focused': { color: accentColor },
    '& .MuiOutlinedInput-root': {
      backgroundColor: tokens.fieldSurface,
      boxShadow: tokens.insetHighlight,
      '& .MuiOutlinedInput-notchedOutline': { borderColor: tokens.softBorder },
      '&:hover .MuiOutlinedInput-notchedOutline': {
        borderColor: withAlpha(accentColor, tokens.isDark ? 0.42 : 0.28)
      },
      '&.Mui-focused': {
        backgroundColor: tokens.fieldSurfaceActive
      },
      '&.Mui-focused .MuiOutlinedInput-notchedOutline': { borderColor: accentColor }
    },
    '& .MuiInputBase-input': { color: tokens.primaryText },
    '& .MuiSelect-select': { color: tokens.primaryText },
    '& .MuiSelect-icon': { color: tokens.secondaryText },
    '& .MuiAutocomplete-popupIndicator, & .MuiAutocomplete-clearIndicator': {
      color: tokens.secondaryText
    },
    '& .MuiChip-root': {
      backgroundColor: tokens.fieldSurface,
      border: `1px solid ${tokens.subtleBorder}`,
      color: tokens.primaryText
    },
    '& .MuiFormHelperText-root': { color: tokens.tertiaryText }
  };
}

export function getChainProxyToggleButtonGroupSx(tokens, accentColor = tokens.palette.primary.main) {
  return {
    '& .MuiToggleButton-root': {
      color: tokens.secondaryText,
      backgroundColor: tokens.fieldSurface,
      borderColor: tokens.softBorder,
      boxShadow: tokens.insetHighlight,
      '&:hover': {
        backgroundColor: withAlpha(accentColor, tokens.isDark ? 0.12 : 0.06),
        borderColor: withAlpha(accentColor, tokens.isDark ? 0.34 : 0.22)
      },
      '&.Mui-selected, &.Mui-selected:hover': {
        color: accentColor,
        backgroundColor: withAlpha(accentColor, tokens.isDark ? 0.18 : 0.1),
        borderColor: withAlpha(accentColor, tokens.isDark ? 0.46 : 0.28)
      }
    }
  };
}

export function getChainProxyIconButtonSx(tokens, accentColor = tokens.palette.primary.main) {
  return {
    color: tokens.secondaryText,
    border: `1px solid ${tokens.subtleBorder}`,
    backgroundColor: tokens.fieldSurface,
    boxShadow: tokens.insetHighlight,
    '&:hover': {
      color: accentColor,
      backgroundColor: withAlpha(accentColor, tokens.isDark ? 0.12 : 0.06),
      borderColor: withAlpha(accentColor, tokens.isDark ? 0.34 : 0.22)
    }
  };
}
