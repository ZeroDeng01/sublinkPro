import { withAlpha } from '../../../utils/colorUtils';

export function getSubscriptionNameChipSx(theme) {
  const palette = theme.vars?.palette || theme.palette;
  const isDark = theme.palette.mode === 'dark';

  return {
    color: isDark ? palette.primary.light : palette.primary.dark,
    bgcolor: withAlpha(palette.primary.main, isDark ? 0.16 : 0.1),
    border: '1px solid',
    borderColor: withAlpha(palette.primary.main, isDark ? 0.34 : 0.2),
    fontWeight: 600,
    '& .MuiChip-label': {
      color: 'inherit'
    }
  };
}
