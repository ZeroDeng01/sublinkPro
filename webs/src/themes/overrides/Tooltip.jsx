import { alpha } from '@mui/material/styles';

export default function Tooltip(theme) {
  const isDark = theme.palette.mode === 'dark';
  const tooltipBg = isDark ? alpha(theme.palette.background.default, 0.98) : alpha(theme.palette.grey[800], 0.96);
  const tooltipColor = isDark ? theme.palette.text.primary : theme.palette.common.white;

  return {
    MuiTooltip: {
      defaultProps: {
        arrow: true
      },
      styleOverrides: {
        tooltip: {
          backgroundColor: tooltipBg,
          color: tooltipColor,
          borderRadius: 8,
          border: `1px solid ${isDark ? alpha(theme.palette.divider, 0.75) : 'transparent'}`,
          boxShadow: isDark ? `0 8px 24px ${alpha(theme.palette.common.black, 0.32)}` : theme.shadows[8],
          fontSize: '0.75rem',
          lineHeight: 1.4,
          fontWeight: 500,
          padding: '8px 10px',
          maxWidth: 320
        },
        arrow: {
          color: tooltipBg,
          '&::before': {
            border: isDark ? `1px solid ${alpha(theme.palette.divider, 0.75)}` : undefined
          }
        },
        popper: {
          zIndex: theme.zIndex.tooltip
        }
      }
    }
  };
}
