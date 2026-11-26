// project imports
import { withAlpha } from 'utils/colorUtils';

// ===============================||  OVERRIDES - CHIP  ||=============================== //

export default function Chip(theme) {
  return {
    MuiChip: {
      defaultProps: {
        color: 'primary',
        variant: 'light'
      },
      styleOverrides: {
        root: {
          variants: [
            {
              props: { variant: 'light' },
              style: ({ ownerState, theme }) => {
                const colorKey = ownerState.color;
                const palette = theme.vars.palette;
                const paletteColor = palette[colorKey];

                if (!paletteColor) return {};

                const lightText = colorKey === 'warning' || colorKey === 'success' ? paletteColor.dark : paletteColor.main;
                const lightBackgroundOpacity = { error: 0.25, success: 0.5 }[colorKey];
                const lightBackground = lightBackgroundOpacity ? withAlpha(paletteColor.light, lightBackgroundOpacity) : paletteColor.light;
                const darkText =
                  { primary: palette.primary[200], secondary: palette.secondary[200], warning: palette.warning.main }[colorKey] ||
                  paletteColor.light;

                return {
                  '--Chip-light-text': lightText,
                  '--Chip-light-background': lightBackground,
                  '--Chip-light-interactive-text': paletteColor.light,
                  '--Chip-light-interactive-background': paletteColor.dark,
                  '--Chip-light-focus-outline': 'none',
                  color: 'var(--Chip-light-text)',
                  backgroundColor: 'var(--Chip-light-background)',
                  '&.MuiChip-clickable': {
                    '&:hover': {
                      color: 'var(--Chip-light-interactive-text)',
                      backgroundColor: 'var(--Chip-light-interactive-background)'
                    }
                  },
                  '&.Mui-focusVisible': {
                    backgroundColor: 'var(--Chip-light-interactive-background)',
                    outline: 'var(--Chip-light-focus-outline)',
                    outlineOffset: 2
                  },
                  // Keep color declarations at their usual specificity so local sx can override them.
                  ...theme.applyStyles('dark', {
                    '--Chip-light-text': darkText,
                    '--Chip-light-background': withAlpha(paletteColor.main, 0.14),
                    '--Chip-light-interactive-text': darkText,
                    '--Chip-light-interactive-background': withAlpha(paletteColor.main, 0.22),
                    '--Chip-light-focus-outline': `2px solid ${darkText}`
                  })
                };
              }
            },
            {
              props: { variant: 'outlined', color: 'warning' },
              style: {
                borderColor: theme.vars.palette.warning.dark,
                color: theme.vars.palette.warning.dark
              }
            },
            {
              props: { variant: 'outlined', color: 'success' },
              style: {
                borderColor: theme.vars.palette.success.dark,
                color: theme.vars.palette.success.dark
              }
            }
          ],
          '&.MuiChip-deletable .MuiChip-deleteIcon': {
            color: 'inherit'
          }
        }
      }
    }
  };
}
