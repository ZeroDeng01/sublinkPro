import { withAlpha } from 'utils/colorUtils';

export const getSidebarNavTokens = (theme, isDark) => {
  const palette = theme.vars?.palette || theme.palette;
  const accent = theme.palette.secondary.main;

  return {
    accent,
    itemText: palette.text.primary,
    mutedText: palette.text.secondary,
    hoverSurface: withAlpha(accent, isDark ? 0.12 : 0.08),
    hoverBorder: withAlpha(accent, isDark ? 0.16 : 0.1),
    selectedSurface: withAlpha(accent, isDark ? 0.1 : 0.08),
    selectedHoverSurface: withAlpha(accent, isDark ? 0.14 : 0.1),
    selectedBorder: withAlpha(accent, isDark ? 0.2 : 0.16),
    iconSurface: isDark ? withAlpha(palette.background.default, 0.72) : 'transparent',
    iconHoverSurface: isDark ? withAlpha(accent, 0.12) : withAlpha(accent, 0.14),
    iconActiveSurface: isDark ? withAlpha(accent, 0.14) : withAlpha(accent, 0.16),
    iconBorder: withAlpha(accent, isDark ? 0.18 : 0.12),
    collapsedLevelOneInset: 0.75,
    collapsedLevelOneTileSize: 40
  };
};
