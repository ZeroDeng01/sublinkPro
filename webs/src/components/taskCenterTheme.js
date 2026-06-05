import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

export const TASK_CLUSTER_ACCENT = '#6366f1';

export const TASK_TYPE_META = {
  speed_test: { label: 'Speed test', labelKey: 'tasks.type.speedTest', color: '#10b981' },
  sub_update: { label: 'Subscription update', labelKey: 'tasks.type.subUpdate', color: '#6366f1' },
  tag_rule: { label: 'Tag rule', labelKey: 'tasks.type.tagRule', color: '#f59e0b' },
  db_migration: { label: 'Database migration', labelKey: 'tasks.type.dbMigration', color: '#0284c7' }
};

export const TASK_TRIGGER_META = {
  manual: { label: 'Manual', labelKey: 'tasks.trigger.manual', color: '#8b5cf6' },
  scheduled: { label: 'Scheduled', labelKey: 'tasks.trigger.scheduled', color: '#06b6d4' },
  airport_update: { label: 'Airport update', labelKey: 'tasks.trigger.airportUpdate', color: '#0ea5e9' }
};

const withTranslatedLabel = (meta, t) => ({
  ...meta,
  label: t && meta.labelKey ? t(meta.labelKey, meta.label) : meta.label
});

export const getTaskTypeMeta = (type, t) => withTranslatedLabel(TASK_TYPE_META[type] || TASK_TYPE_META.speed_test, t);

export const getTaskTriggerMeta = (trigger, t) => withTranslatedLabel(TASK_TRIGGER_META[trigger] || TASK_TRIGGER_META.manual, t);

export const getTaskStatusMeta = (theme, status, t) => {
  const statusMap = {
    pending: { label: 'Pending', labelKey: 'tasks.status.pending', color: theme.palette.warning.main },
    running: { label: 'Running', labelKey: 'tasks.status.running', color: theme.palette.primary.main },
    completed: { label: 'Completed', labelKey: 'tasks.status.completed', color: theme.palette.success.main },
    cancelled: { label: 'Cancelled', labelKey: 'tasks.status.cancelled', color: theme.palette.warning.main },
    cancelling: { label: 'Stopping', labelKey: 'tasks.status.cancelling', color: theme.palette.warning.main },
    error: { label: 'Failed', labelKey: 'tasks.status.error', color: theme.palette.error.main }
  };

  return withTranslatedLabel(statusMap[status] || statusMap.pending, t);
};

export const getTaskCenterTokens = (theme, isDark) => {
  const surfaceTokens = getSurfaceTokens(theme, isDark);
  const textTokens = getReadableTextTokens(theme, isDark);
  const { palette } = surfaceTokens;
  const shellSurface = isDark ? withAlpha(palette.background.default, 0.985) : palette.background.paper;
  const shellOverlay = isDark ? withAlpha(palette.background.paper, 0.18) : withAlpha(TASK_CLUSTER_ACCENT, 0.035);

  return {
    ...surfaceTokens,
    ...textTokens,
    palette,
    shellSurface,
    shellSurfaceGradient: `linear-gradient(180deg, ${shellOverlay} 0%, ${shellSurface} 100%)`,
    floatingSurface: isDark ? withAlpha(palette.background.default, 0.975) : withAlpha(palette.background.paper, 0.985),
    sectionSurface: isDark ? withAlpha(palette.background.default, 0.9) : withAlpha(palette.background.paper, 0.98),
    nestedInteractiveSurface: isDark ? withAlpha(palette.background.paper, 0.36) : withAlpha(palette.background.default, 0.82),
    mutedSectionSurface: isDark ? withAlpha(palette.background.default, 0.78) : withAlpha(palette.background.default, 0.9),
    strongBorder: isDark ? withAlpha(palette.divider, 0.86) : withAlpha(palette.divider, 0.92),
    softBorder: isDark ? withAlpha(palette.divider, 0.72) : withAlpha(palette.divider, 0.72),
    tableHeaderSurface: isDark ? withAlpha(palette.background.paper, 0.5) : withAlpha(TASK_CLUSTER_ACCENT, 0.03),
    rowHoverSurface: isDark ? withAlpha(TASK_CLUSTER_ACCENT, 0.12) : withAlpha(TASK_CLUSTER_ACCENT, 0.04),
    insetHighlight: isDark ? `inset 0 1px 0 ${withAlpha(theme.palette.common.white, 0.04)}` : 'none'
  };
};

export const getTaskShellSx = (theme, tokens, accentColor = TASK_CLUSTER_ACCENT, options = {}) => {
  const { floating = false, interactive = true } = options;
  const surfaceColor = floating ? tokens.floatingSurface : tokens.shellSurface;
  const overlayColor = tokens.isDark
    ? withAlpha(tokens.palette.background.paper, floating ? 0.22 : 0.18)
    : withAlpha(accentColor, floating ? 0.05 : 0.035);

  return {
    position: 'relative',
    bgcolor: surfaceColor,
    backgroundImage: `linear-gradient(180deg, ${overlayColor} 0%, ${surfaceColor} 100%)`,
    border: '1px solid',
    borderColor: tokens.isDark ? tokens.strongBorder : withAlpha(accentColor, 0.16),
    boxShadow: tokens.isDark
      ? `0 16px 36px ${withAlpha(theme.palette.common.black, floating ? 0.28 : 0.2)}, ${tokens.insetHighlight}`
      : floating
        ? theme.shadows[8]
        : `0 8px 24px ${withAlpha(theme.palette.common.black, 0.08)}`,
    transition: 'border-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease, background-color 0.2s ease',
    ...(interactive
      ? {
          '&:hover': {
            borderColor: withAlpha(accentColor, tokens.isDark ? 0.3 : 0.22),
            boxShadow: tokens.isDark
              ? `0 20px 40px ${withAlpha(theme.palette.common.black, floating ? 0.3 : 0.24)}, inset 0 1px 0 ${withAlpha(
                  theme.palette.common.white,
                  0.05
                )}`
              : `0 14px 28px ${withAlpha(theme.palette.common.black, 0.1)}`
          }
        }
      : {})
  };
};

export const getTaskCardSx = (theme, tokens, accentColor, options = {}) => {
  const { interactive = true, compact = false } = options;
  const baseSurface = compact ? tokens.nestedInteractiveSurface : tokens.sectionSurface;
  const overlayColor = tokens.isDark
    ? withAlpha(tokens.palette.background.paper, compact ? 0.16 : 0.2)
    : withAlpha(accentColor, compact ? 0.035 : 0.045);

  return {
    bgcolor: baseSurface,
    backgroundImage: `linear-gradient(180deg, ${overlayColor} 0%, ${baseSurface} 100%)`,
    border: '1px solid',
    borderColor: tokens.isDark ? tokens.softBorder : withAlpha(accentColor, 0.14),
    boxShadow: tokens.isDark
      ? `0 10px 24px ${withAlpha(theme.palette.common.black, 0.14)}, inset 0 1px 0 ${withAlpha(theme.palette.common.white, 0.03)}`
      : `0 4px 16px ${withAlpha(theme.palette.common.black, 0.06)}`,
    transition: 'border-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease, background-color 0.2s ease',
    ...(interactive
      ? {
          '&:hover': {
            borderColor: withAlpha(accentColor, tokens.isDark ? 0.28 : 0.2),
            boxShadow: tokens.isDark
              ? `0 14px 30px ${withAlpha(theme.palette.common.black, 0.18)}, inset 0 1px 0 ${withAlpha(theme.palette.common.white, 0.04)}`
              : `0 10px 24px ${withAlpha(theme.palette.common.black, 0.08)}`,
            transform: 'translateY(-1px)'
          }
        }
      : {})
  };
};

export const getTaskIconBoxSx = (theme, tokens, accentColor, options = {}) => {
  const { size = 40, radius = 2, filled = false } = options;

  return {
    width: size,
    height: size,
    borderRadius: radius,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: filled ? accentColor : withAlpha(accentColor, tokens.isDark ? 0.18 : 0.12),
    color: filled ? theme.palette.common.white : accentColor,
    border: '1px solid',
    borderColor: filled ? withAlpha(accentColor, 0.36) : withAlpha(accentColor, tokens.isDark ? 0.32 : 0.18),
    boxShadow: filled
      ? `0 10px 18px ${withAlpha(accentColor, tokens.isDark ? 0.28 : 0.18)}`
      : `inset 0 1px 0 ${withAlpha(theme.palette.common.white, tokens.isDark ? 0.04 : 0)}`,
    flexShrink: 0
  };
};

export const getTaskChipSx = (theme, tokens, accentColor, options = {}) => {
  const { emphasis = 'soft' } = options;
  const isSolid = emphasis === 'solid';

  return {
    bgcolor: isSolid ? accentColor : withAlpha(accentColor, tokens.isDark ? 0.2 : 0.1),
    color: isSolid ? theme.palette.common.white : tokens.isDark ? withAlpha(theme.palette.common.white, 0.92) : accentColor,
    border: '1px solid',
    borderColor: isSolid ? withAlpha(accentColor, 0.4) : withAlpha(accentColor, tokens.isDark ? 0.32 : 0.2),
    fontWeight: 600,
    boxShadow: isSolid ? `0 8px 16px ${withAlpha(accentColor, tokens.isDark ? 0.22 : 0.16)}` : 'none',
    '& .MuiChip-icon': {
      color: isSolid ? theme.palette.common.white : accentColor
    },
    '&:hover': {
      bgcolor: isSolid ? accentColor : withAlpha(accentColor, tokens.isDark ? 0.26 : 0.14)
    }
  };
};

export const getTaskProgressSx = (tokens, accentColor, options = {}) => {
  const { height = 6 } = options;

  return {
    height,
    borderRadius: height / 2,
    backgroundColor: tokens.isDark ? withAlpha(tokens.palette.background.default, 0.92) : withAlpha(accentColor, 0.1),
    '& .MuiLinearProgress-bar': {
      backgroundColor: accentColor,
      borderRadius: height / 2
    }
  };
};

export const getTaskDialogPaperSx = (theme, tokens, accentColor = TASK_CLUSTER_ACCENT) => ({
  ...getTaskShellSx(theme, tokens, accentColor, { floating: true, interactive: false }),
  borderRadius: 3,
  overflow: 'hidden'
});

export const getTaskActionButtonSx = (theme, tokens, accentColor, options = {}) => {
  const { variant = 'soft' } = options;
  const isSolid = variant === 'solid';

  return {
    textTransform: 'none',
    borderRadius: 2,
    fontWeight: 600,
    color: isSolid ? theme.palette.common.white : accentColor,
    bgcolor: isSolid ? accentColor : withAlpha(accentColor, tokens.isDark ? 0.16 : 0.08),
    border: '1px solid',
    borderColor: withAlpha(accentColor, tokens.isDark ? 0.28 : 0.16),
    boxShadow: isSolid ? `0 10px 20px ${withAlpha(accentColor, tokens.isDark ? 0.24 : 0.16)}` : 'none',
    '&:hover': {
      bgcolor: isSolid ? accentColor : withAlpha(accentColor, tokens.isDark ? 0.22 : 0.12),
      borderColor: withAlpha(accentColor, tokens.isDark ? 0.36 : 0.22)
    }
  };
};
