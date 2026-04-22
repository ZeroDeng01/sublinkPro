import { withAlpha } from 'utils/colorUtils';

export function getSurfaceTokens(theme, isDark) {
  const palette = theme.vars?.palette || theme.palette;
  const dialogSurface = isDark ? withAlpha(palette.background.default, 0.96) : palette.background.paper;

  return {
    isDark,
    palette,
    dialogSurface,
    dialogSurfaceGradient: isDark
      ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.16)} 0%, ${dialogSurface} 100%)`
      : 'none',
    mutedPanelSurface: isDark ? withAlpha(palette.background.default, 0.84) : palette.background.default,
    nestedPanelSurface: isDark ? withAlpha(palette.background.paper, 0.42) : palette.background.paper,
    panelBorder: isDark ? withAlpha(palette.divider, 0.82) : withAlpha(palette.divider, 0.9)
  };
}

export function getReadableTextTokens(theme, isDark) {
  const palette = theme.vars?.palette || theme.palette;
  const darkText = palette.text?.dark || theme.palette.common.white;

  return {
    isDark,
    palette,
    darkText,
    primaryText: isDark ? withAlpha(darkText, 0.94) : palette.text.primary,
    secondaryText: isDark ? withAlpha(darkText, 0.82) : palette.text.secondary,
    tertiaryText: isDark ? withAlpha(darkText, 0.68) : withAlpha(palette.text.primary, 0.68)
  };
}
