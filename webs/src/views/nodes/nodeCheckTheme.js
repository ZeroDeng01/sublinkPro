import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

export function getNodeCheckStrategyThemeTokens(theme, isDark) {
  const surfaceTokens = getSurfaceTokens(theme, isDark);
  const textTokens = getReadableTextTokens(theme, isDark);
  const { palette, nestedPanelSurface } = surfaceTokens;

  return {
    ...surfaceTokens,
    ...textTokens,
    cardSurface: isDark ? withAlpha(palette.background.paper, 0.34) : withAlpha(palette.background.paper, 0.98),
    cardEnabledBorder: withAlpha(palette.success.main, isDark ? 0.5 : 0.3),
    cardIdleBorder: isDark ? withAlpha(palette.divider, 0.82) : withAlpha(palette.divider, 0.9),
    cardHoverShadow: isDark ? `0 4px 20px ${withAlpha(palette.common.black, 0.3)}` : `0 4px 20px ${withAlpha(palette.common.black, 0.1)}`,
    headerSurface: isDark ? withAlpha(palette.background.paper, 0.2) : nestedPanelSurface,
    actionSurface: isDark ? withAlpha(palette.background.default, 0.9) : withAlpha(palette.background.default, 0.78),
    emptyStateSurface: isDark ? withAlpha(palette.background.paper, 0.34) : withAlpha(palette.background.default, 0.72),
    sectionSurface: isDark ? withAlpha(palette.background.paper, 0.36) : nestedPanelSurface,
    sectionHeaderSurface: isDark ? withAlpha(palette.background.default, 0.84) : withAlpha(palette.background.default, 0.72),
    sectionHoverSurface: isDark
      ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.2)} 0%, ${withAlpha(palette.primary.main, 0.08)} 100%)`
      : withAlpha(palette.primary.main, 0.04),
    listRowHoverBackground: isDark
      ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.28)} 0%, ${withAlpha(palette.primary.main, 0.05)} 100%)`
      : withAlpha(palette.primary.main, 0.04),
    listRowSelectedBackground: isDark
      ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.42)} 0%, ${withAlpha(palette.primary.main, 0.12)} 100%)`
      : withAlpha(palette.primary.main, 0.08),
    listRowSelectedHoverBackground: isDark
      ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.48)} 0%, ${withAlpha(palette.primary.main, 0.16)} 100%)`
      : withAlpha(palette.primary.main, 0.11),
    listRowSelectedShadow: isDark
      ? `inset 0 0 0 1px ${withAlpha(palette.primary.main, 0.24)}, inset 0 1px 0 ${withAlpha(palette.common.white, 0.05)}`
      : `inset 0 0 0 1px ${withAlpha(palette.primary.main, 0.14)}`,
    closeButtonHoverSurface: withAlpha(palette.primary.main, isDark ? 0.14 : 0.08),
    primaryActionHoverSurface: withAlpha(palette.primary.main, isDark ? 0.14 : 0.08),
    successActionHoverSurface: withAlpha(palette.success.main, isDark ? 0.16 : 0.08),
    errorActionHoverSurface: withAlpha(palette.error.main, isDark ? 0.16 : 0.08),
    secondaryIconOpacity: isDark ? 0.82 : 0.64
  };
}

export function getNodeCheckStrategyChipSx(themeTokens, tone = 'neutral') {
  const { palette, isDark, darkText, primaryText } = themeTokens;
  const toneMap = {
    success: {
      accent: palette.success.main,
      text: isDark ? palette.success.light : palette.success.dark
    },
    info: {
      accent: palette.info.main,
      text: isDark ? palette.info.light : palette.info.dark
    },
    warning: {
      accent: palette.warning.main,
      text: isDark ? palette.warning.light : palette.warning.dark
    },
    neutral: {
      accent: palette.grey[500],
      text: isDark ? withAlpha(darkText, 0.92) : withAlpha(primaryText, 0.82)
    }
  };

  const { accent, text } = toneMap[tone] || toneMap.neutral;

  return {
    height: 22,
    borderRadius: 1.5,
    flexShrink: 0,
    fontSize: '0.72rem',
    fontWeight: 700,
    letterSpacing: '0.01em',
    color: text,
    backgroundColor: withAlpha(accent, isDark ? (tone === 'neutral' ? 0.24 : 0.18) : tone === 'neutral' ? 0.08 : 0.1),
    border: `1px solid ${withAlpha(accent, isDark ? (tone === 'neutral' ? 0.34 : 0.32) : tone === 'neutral' ? 0.16 : 0.2)}`,
    boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.05)}` : 'none',
    '& .MuiChip-label': {
      px: 1,
      py: 0
    },
    '& .MuiChip-icon': {
      color: 'inherit',
      fontSize: '0.82rem',
      ml: 0.75,
      mr: -0.25
    }
  };
}
