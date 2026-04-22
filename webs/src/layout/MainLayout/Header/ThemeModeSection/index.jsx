import { useEffect, useRef, useState } from 'react';

import { useColorScheme, useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import ClickAwayListener from '@mui/material/ClickAwayListener';
import Paper from '@mui/material/Paper';
import Popper from '@mui/material/Popper';
import Stack from '@mui/material/Stack';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

import DarkModeOutlinedIcon from '@mui/icons-material/DarkModeOutlined';
import LightModeOutlinedIcon from '@mui/icons-material/LightModeOutlined';
import SettingsBrightnessOutlinedIcon from '@mui/icons-material/SettingsBrightnessOutlined';

import { DEFAULT_THEME_MODE } from 'config';
import ThemeModeSelector from 'components/ThemeModeSelector';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getHeaderPopoverTokens, getHeaderTriggerTokens } from '../headerPopoverTokens';
import MainCard from 'ui-component/cards/MainCard';
import Transitions from 'ui-component/extended/Transitions';

const modeMeta = {
  system: {
    label: '跟随系统',
    icon: SettingsBrightnessOutlinedIcon,
    colorKey: 'info'
  },
  light: {
    label: '浅色模式',
    icon: LightModeOutlinedIcon,
    colorKey: 'warning'
  },
  dark: {
    label: '深色模式',
    icon: DarkModeOutlinedIcon,
    colorKey: 'secondary'
  }
};

export default function ThemeModeSection() {
  const theme = useTheme();
  const downMD = useMediaQuery(theme.breakpoints.down('md'));
  const { mode } = useColorScheme();
  const { isDark } = useResolvedColorScheme();

  const [open, setOpen] = useState(false);
  const anchorRef = useRef(null);

  const selectedMode = mode || DEFAULT_THEME_MODE;
  const currentMeta = modeMeta[selectedMode] || modeMeta.system;
  const Icon = currentMeta.icon;
  const paletteColor = theme.palette[currentMeta.colorKey];
  const accentColor = isDark
    ? paletteColor.main
    : currentMeta.colorKey === 'warning'
      ? theme.palette.warning.dark
      : paletteColor.dark || paletteColor.main;
  const { popoverSurface, popoverSurfaceAccent, popoverBorder, popoverInsetShadow, mutedText } = getHeaderPopoverTokens(theme, isDark);
  const { triggerColor, triggerSurface, triggerBorder, activeColor, activeSurface, activeBorder } = getHeaderTriggerTokens(
    theme,
    isDark,
    accentColor,
    {
      lightSurfaceAlpha: 0.14,
      lightHoverAlpha: 0.22,
      triggerColor: accentColor
    }
  );

  const handleToggle = () => {
    setOpen((prevOpen) => !prevOpen);
  };

  const handleClose = (event) => {
    if (anchorRef.current && anchorRef.current.contains(event.target)) {
      return;
    }
    setOpen(false);
  };

  const prevOpen = useRef(open);

  useEffect(() => {
    if (prevOpen.current === true && open === false) {
      anchorRef.current?.focus();
    }
    prevOpen.current = open;
  }, [open]);

  return (
    <>
      <Box sx={{ ml: 2 }}>
        <Tooltip title={`主题模式：${currentMeta.label}`}>
          <Avatar
            ref={anchorRef}
            variant="rounded"
            aria-controls={open ? 'theme-mode-menu' : undefined}
            aria-haspopup="true"
            onClick={handleToggle}
            sx={{
              ...theme.typography.commonAvatar,
              ...theme.typography.mediumAvatar,
              transition: 'all .2s ease-in-out',
              color: triggerColor,
              background: triggerSurface,
              border: '1px solid',
              borderColor: triggerBorder,
              '&:hover, &[aria-controls="theme-mode-menu"]': {
                color: activeColor,
                background: activeSurface,
                borderColor: activeBorder
              }
            }}
          >
            <Icon fontSize="small" />
          </Avatar>
        </Tooltip>
      </Box>

      <Popper
        id="theme-mode-menu"
        placement={downMD ? 'bottom' : 'bottom-end'}
        open={open}
        anchorEl={anchorRef.current}
        role={undefined}
        transition
        disablePortal
        modifiers={[{ name: 'offset', options: { offset: [downMD ? 5 : 0, 20] } }]}
      >
        {({ TransitionProps }) => (
          <ClickAwayListener onClickAway={handleClose}>
            <Transitions position={downMD ? 'top' : 'top-right'} in={open} {...TransitionProps}>
              <Paper sx={{ bgcolor: 'transparent' }}>
                {open && (
                  <MainCard
                    border={false}
                    elevation={0}
                    content={false}
                    boxShadow
                    shadow={isDark ? 'none' : theme.shadows[12]}
                    sx={{
                      minWidth: 320,
                      maxWidth: 360,
                      bgcolor: popoverSurface,
                      backgroundImage: popoverSurfaceAccent,
                      border: '1px solid',
                      borderColor: popoverBorder,
                      boxShadow: popoverInsetShadow
                    }}
                  >
                    <Stack sx={{ p: 2.5, gap: 2 }}>
                      <Box>
                        <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
                          主题模式
                        </Typography>
                        <Typography variant="body2" sx={{ mt: 0.5, color: mutedText }}>
                          当前为 {currentMeta.label}，可立即切换整个界面的配色方案。
                        </Typography>
                      </Box>
                      <ThemeModeSelector locale="zh" title={null} />
                    </Stack>
                  </MainCard>
                )}
              </Paper>
            </Transitions>
          </ClickAwayListener>
        )}
      </Popper>
    </>
  );
}
