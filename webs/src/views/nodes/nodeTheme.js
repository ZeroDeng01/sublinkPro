import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

export function getNodeThemeTokens(theme, isDark) {
  const surfaceTokens = getSurfaceTokens(theme, isDark);
  const textTokens = getReadableTextTokens(theme, isDark);
  const { palette } = surfaceTokens;

  return {
    ...surfaceTokens,
    ...textTokens,
    palette,
    shellSurface: isDark ? withAlpha(palette.background.default, 0.985) : palette.background.paper,
    shellSurfaceGradient: `linear-gradient(180deg, ${
      isDark ? withAlpha(palette.background.paper, 0.18) : withAlpha(palette.primary.main, 0.03)
    } 0%, ${isDark ? withAlpha(palette.background.default, 0.985) : palette.background.paper} 100%)`,
    cardSurface: isDark ? withAlpha(palette.background.paper, 0.34) : withAlpha(palette.background.paper, 0.98),
    elevatedSurface: isDark ? withAlpha(palette.background.default, 0.9) : withAlpha(palette.background.paper, 0.96),
    toolbarSurface: isDark ? withAlpha(palette.background.default, 0.9) : withAlpha(palette.background.default, 0.76),
    fieldSurface: isDark ? withAlpha(palette.background.paper, 0.18) : withAlpha(palette.background.paper, 0.92),
    fieldSurfaceActive: isDark ? withAlpha(palette.background.paper, 0.28) : palette.background.paper,
    strongBorder: isDark ? withAlpha(palette.divider, 0.84) : withAlpha(palette.divider, 0.9),
    softBorder: isDark ? withAlpha(palette.divider, 0.66) : withAlpha(palette.divider, 0.78),
    subtleBorder: isDark ? withAlpha(palette.divider, 0.5) : withAlpha(palette.divider, 0.64),
    hoverSurface: withAlpha(palette.primary.main, isDark ? 0.12 : 0.05),
    selectedSurface: isDark
      ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.42)} 0%, ${withAlpha(palette.primary.main, 0.14)} 100%)`
      : withAlpha(palette.primary.main, 0.08),
    selectedHoverSurface: isDark
      ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.48)} 0%, ${withAlpha(palette.primary.main, 0.18)} 100%)`
      : withAlpha(palette.primary.main, 0.11),
    selectedBorder: withAlpha(palette.primary.main, isDark ? 0.34 : 0.18),
    insetHighlight: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none'
  };
}

export function getNodeDialogPaperSx(theme, tokens, accentColor = tokens.palette.primary.main) {
  return {
    borderRadius: 3,
    overflow: 'hidden',
    border: '1px solid',
    borderColor: tokens.isDark ? tokens.strongBorder : withAlpha(accentColor, 0.16),
    bgcolor: tokens.dialogSurface,
    backgroundImage: tokens.dialogSurfaceGradient,
    boxShadow: tokens.isDark
      ? `0 20px 44px ${withAlpha(theme.palette.common.black, 0.28)}, ${tokens.insetHighlight}`
      : `0 16px 36px ${withAlpha(theme.palette.common.black, 0.12)}`
  };
}

export function getNodePanelSx(theme, tokens, accentColor = tokens.palette.primary.main, options = {}) {
  const { compact = false, interactive = false, selected = false } = options;
  const surface = compact ? tokens.nestedPanelSurface : tokens.cardSurface;
  const overlayColor = selected
    ? withAlpha(accentColor, tokens.isDark ? 0.16 : 0.08)
    : tokens.isDark
      ? withAlpha(tokens.palette.background.paper, compact ? 0.14 : 0.18)
      : withAlpha(accentColor, compact ? 0.03 : 0.04);

  return {
    bgcolor: surface,
    backgroundImage: `linear-gradient(180deg, ${overlayColor} 0%, ${surface} 100%)`,
    border: '1px solid',
    borderColor: selected ? withAlpha(accentColor, tokens.isDark ? 0.34 : 0.18) : tokens.softBorder,
    boxShadow: tokens.isDark
      ? `0 10px 24px ${withAlpha(theme.palette.common.black, compact ? 0.12 : 0.16)}, inset 0 1px 0 ${withAlpha(
          theme.palette.common.white,
          compact ? 0.03 : 0.04
        )}`
      : `0 6px 18px ${withAlpha(theme.palette.common.black, compact ? 0.04 : 0.06)}`,
    transition: 'border-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease, background-color 0.2s ease',
    ...(interactive
      ? {
          '&:hover': {
            borderColor: withAlpha(accentColor, tokens.isDark ? 0.28 : 0.18),
            boxShadow: tokens.isDark
              ? `0 16px 32px ${withAlpha(theme.palette.common.black, 0.2)}, inset 0 1px 0 ${withAlpha(theme.palette.common.white, 0.05)}`
              : `0 12px 24px ${withAlpha(theme.palette.common.black, 0.08)}`,
            transform: 'translateY(-1px)'
          }
        }
      : {})
  };
}

export function getNodeFieldControlSx(tokens, accentColor = tokens.palette.primary.main) {
  return {
    '& .MuiInputLabel-root': { color: tokens.secondaryText },
    '& .MuiInputLabel-root.Mui-focused': { color: accentColor },
    '& .MuiOutlinedInput-root': {
      backgroundColor: tokens.fieldSurface,
      boxShadow: tokens.insetHighlight,
      '& .MuiOutlinedInput-notchedOutline': { borderColor: tokens.softBorder },
      '&:hover .MuiOutlinedInput-notchedOutline': {
        borderColor: withAlpha(accentColor, tokens.isDark ? 0.42 : 0.24)
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
    '& .MuiFormHelperText-root': { color: tokens.tertiaryText },
    '& .MuiInputAdornment-root': { color: tokens.secondaryText },
    '& .MuiAutocomplete-tag': {
      maxWidth: '100%'
    }
  };
}

export function getNodeColorChipSx(theme, tokens, accentColor, options = {}) {
  const { emphasis = 'soft', deletable = false } = options;
  const isSolid = emphasis === 'solid';

  return {
    fontWeight: 600,
    border: '1px solid',
    borderColor: isSolid ? withAlpha(accentColor, 0.4) : withAlpha(accentColor, tokens.isDark ? 0.34 : 0.18),
    backgroundColor: isSolid ? accentColor : withAlpha(accentColor, tokens.isDark ? 0.2 : 0.1),
    color: isSolid ? theme.palette.common.white : tokens.isDark ? withAlpha(accentColor, 0.98) : accentColor,
    boxShadow: isSolid ? `0 8px 16px ${withAlpha(accentColor, tokens.isDark ? 0.22 : 0.14)}` : 'none',
    '& .MuiChip-icon': {
      color: isSolid ? theme.palette.common.white : accentColor
    },
    '& .MuiChip-deleteIcon': deletable
      ? {
          color: isSolid ? withAlpha(theme.palette.common.white, 0.74) : withAlpha(accentColor, 0.72),
          '&:hover': {
            color: isSolid ? theme.palette.common.white : accentColor
          }
        }
      : undefined,
    '&:hover': {
      backgroundColor: isSolid ? accentColor : withAlpha(accentColor, tokens.isDark ? 0.26 : 0.14)
    }
  };
}

export function getNodeTagChipSx(theme, tokens, color, options = {}) {
  return {
    ...getNodeColorChipSx(theme, tokens, color || tokens.palette.primary.main, { ...options, emphasis: 'soft' }),
    borderRadius: 1.5,
    '& .MuiChip-label': {
      px: 1,
      py: 0
    }
  };
}

export function getNodeStatusMetricSx(tokens, tone = 'default') {
  const tonePalette = {
    warning: {
      accent: tokens.palette.warning.main,
      text: tokens.isDark ? tokens.palette.warning.light || tokens.palette.warning.main : tokens.palette.warning.dark
    },
    error: {
      accent: tokens.palette.error.main,
      text: tokens.isDark ? tokens.palette.error.light || tokens.palette.error.main : tokens.palette.error.dark
    },
    success: {
      accent: tokens.palette.success.main,
      text: tokens.isDark ? tokens.palette.success.light || tokens.palette.success.main : tokens.palette.success.dark
    },
    info: {
      accent: tokens.palette.info.main,
      text: tokens.isDark ? tokens.palette.info.light || tokens.palette.info.main : tokens.palette.info.dark
    },
    default: {
      accent: tokens.palette.text.secondary,
      text: tokens.secondaryText
    }
  };

  const { accent, text } = tonePalette[tone] || tonePalette.default;

  return {
    color: text,
    bg: withAlpha(accent, tokens.isDark ? 0.16 : 0.08),
    border: withAlpha(accent, tokens.isDark ? 0.32 : 0.18)
  };
}

export function getNodeActionButtonSx(theme, tokens, accentColor, options = {}) {
  const { variant = 'soft' } = options;
  const isSolid = variant === 'solid';

  return {
    minHeight: 34,
    borderRadius: 2,
    textTransform: 'none',
    fontWeight: 600,
    whiteSpace: 'nowrap',
    boxShadow: 'none',
    color: isSolid ? theme.palette.common.white : accentColor,
    backgroundColor: isSolid ? accentColor : withAlpha(accentColor, tokens.isDark ? 0.14 : 0.06),
    border: '1px solid',
    borderColor: withAlpha(accentColor, tokens.isDark ? 0.28 : 0.16),
    '&:hover': {
      backgroundColor: isSolid ? accentColor : withAlpha(accentColor, tokens.isDark ? 0.2 : 0.1),
      borderColor: withAlpha(accentColor, tokens.isDark ? 0.38 : 0.22),
      boxShadow: 'none'
    },
    '&.Mui-disabled': {
      color: tokens.palette.action.disabled,
      borderColor: withAlpha(tokens.palette.action.disabledBackground, tokens.isDark ? 0.68 : 1),
      backgroundColor: tokens.isDark
        ? withAlpha(tokens.palette.action.disabledBackground, 0.42)
        : withAlpha(tokens.palette.action.disabledBackground, 0.92)
    }
  };
}

export function getNodeIconButtonSx(theme, tokens, accentColor = tokens.palette.primary.main, options = {}) {
  const { selected = false } = options;

  return {
    color: selected ? accentColor : tokens.secondaryText,
    border: '1px solid',
    borderColor: selected ? withAlpha(accentColor, tokens.isDark ? 0.34 : 0.2) : tokens.subtleBorder,
    backgroundColor: selected ? withAlpha(accentColor, tokens.isDark ? 0.18 : 0.08) : tokens.fieldSurface,
    boxShadow: tokens.insetHighlight,
    '&:hover': {
      color: accentColor,
      backgroundColor: withAlpha(accentColor, tokens.isDark ? 0.16 : 0.08),
      borderColor: withAlpha(accentColor, tokens.isDark ? 0.38 : 0.24),
      boxShadow: tokens.isDark
        ? `0 8px 16px ${withAlpha(theme.palette.common.black, 0.18)}, inset 0 1px 0 ${withAlpha(theme.palette.common.white, 0.05)}`
        : `0 6px 14px ${withAlpha(theme.palette.common.black, 0.08)}`
    }
  };
}

export function getNodeTableRowSx(theme, tokens, accentColor = tokens.palette.primary.main, selected = false) {
  return {
    cursor: 'pointer',
    transition: 'background-color 0.2s ease, box-shadow 0.2s ease',
    '&:hover': {
      backgroundColor: selected ? tokens.selectedHoverSurface : tokens.hoverSurface
    },
    '&.Mui-selected': {
      backgroundColor: selected ? tokens.selectedSurface : undefined,
      boxShadow: selected ? `inset 0 0 0 1px ${withAlpha(accentColor, tokens.isDark ? 0.24 : 0.12)}` : 'none'
    },
    '&.Mui-selected:hover': {
      backgroundColor: selected ? tokens.selectedHoverSurface : undefined
    }
  };
}
