import { withAlpha } from '../../../utils/colorUtils';

export function getSubscriptionNameChipSx(theme) {
  const palette = theme.vars?.palette || theme.palette;

  return [
    {
      color: palette.primary.dark,
      bgcolor: withAlpha(palette.primary.main, 0.1),
      border: '1px solid',
      borderColor: withAlpha(palette.primary.main, 0.2),
      fontWeight: 600,
      '& .MuiChip-label': {
        color: 'inherit'
      }
    },
    theme.applyStyles('dark', {
      color: palette.primary.light,
      bgcolor: withAlpha(palette.primary.main, 0.16),
      borderColor: withAlpha(palette.primary.main, 0.34)
    })
  ];
}
