import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

export const getHeaderPopoverTokens = (theme, isDark) => {
  const { palette, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(theme, isDark);
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);

  return {
    palette,
    primaryText,
    emphasizedText: isDark ? withAlpha(primaryText, 0.98) : primaryText,
    secondaryText,
    mutedText: isDark ? withAlpha(primaryText, 0.74) : secondaryText,
    tertiaryText,
    popoverSurface: palette.background.paper,
    popoverSurfaceAccent: isDark
      ? `linear-gradient(180deg, ${withAlpha(theme.palette.common.white, 0.025)} 0%, transparent 100%)`
      : dialogSurfaceGradient,
    popoverBorder: panelBorder,
    popoverInsetShadow: isDark ? `inset 0 1px 0 ${withAlpha(theme.palette.common.white, 0.04)}` : undefined,
    headerSurface: isDark ? mutedPanelSurface : 'transparent',
    headerDivider: withAlpha(palette.divider, isDark ? 0.62 : 1),
    nestedSurface: isDark ? mutedPanelSurface : withAlpha(palette.background.default, 0.72),
    nestedSurfaceStrong: isDark ? nestedPanelSurface : palette.background.paper,
    nestedBorder: isDark ? panelBorder : withAlpha(theme.palette.primary.main, 0.16),
    listItemHover: withAlpha(theme.palette.primary.main, isDark ? 0.12 : 0.08),
    selectedSurface: withAlpha(theme.palette.primary.main, isDark ? 0.1 : 0.08),
    selectedHoverSurface: withAlpha(theme.palette.primary.main, isDark ? 0.14 : 0.1)
  };
};

export const getHeaderTriggerTokens = (theme, isDark, accentColor, options = {}) => {
  const palette = theme.vars?.palette || theme.palette;
  const {
    lightSurfaceAlpha = 0.12,
    darkSurfaceAlpha = 0.88,
    lightHoverAlpha = 0.22,
    darkHoverAlpha = 0.16,
    triggerColor = accentColor,
    activeColor = theme.palette.common.white
  } = options;

  return {
    triggerColor,
    triggerSurface: isDark ? withAlpha(palette.background.default, darkSurfaceAlpha) : withAlpha(accentColor, lightSurfaceAlpha),
    triggerBorder: withAlpha(accentColor, isDark ? 0.28 : 0.18),
    activeColor,
    activeSurface: isDark ? withAlpha(accentColor, darkHoverAlpha) : withAlpha(accentColor, lightHoverAlpha),
    activeBorder: withAlpha(accentColor, isDark ? 0.34 : 0.24)
  };
};
