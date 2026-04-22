import { alpha } from '@mui/material/styles';

export default function Tooltip(theme) {
  const tooltipBg = alpha(theme.palette.grey[800], 0.96);
  const tooltipColor = theme.palette.common.white;

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
          border: '1px solid transparent',
          boxShadow: theme.shadows[8],
          fontSize: '0.75rem',
          lineHeight: 1.4,
          fontWeight: 500,
          padding: '8px 10px',
          maxWidth: 320,
          ...theme.applyStyles('dark', {
            backgroundColor: alpha(theme.palette.background.default, 0.98),
            color: theme.palette.text.primary,
            border: `1px solid ${alpha(theme.palette.divider, 0.75)}`,
            boxShadow: `0 8px 24px ${alpha(theme.palette.common.black, 0.32)}`
          })
        },
        arrow: {
          color: tooltipBg,
          '&::before': {
            border: undefined
          },
          ...theme.applyStyles('dark', {
            color: alpha(theme.palette.background.default, 0.98),
            '&::before': {
              border: `1px solid ${alpha(theme.palette.divider, 0.75)}`
            }
          })
        },
        popper: {
          zIndex: theme.zIndex.tooltip
        }
      }
    }
  };
}
