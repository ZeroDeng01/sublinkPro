import { useEffect, useRef, useState } from 'react';

import { alpha, useColorScheme, useTheme } from '@mui/material/styles';
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

  const [open, setOpen] = useState(false);
  const anchorRef = useRef(null);

  const selectedMode = mode || DEFAULT_THEME_MODE;
  const currentMeta = modeMeta[selectedMode] || modeMeta.system;
  const Icon = currentMeta.icon;
  const paletteColor = theme.palette[currentMeta.colorKey];
  const iconColor =
    theme.palette.mode === 'dark'
      ? paletteColor.main
      : currentMeta.colorKey === 'warning'
        ? theme.palette.warning.dark
        : paletteColor.dark || paletteColor.main;

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
              color: iconColor,
              background: alpha(paletteColor.main, theme.palette.mode === 'dark' ? 0.22 : 0.14),
              border: '1px solid',
              borderColor: alpha(iconColor, theme.palette.mode === 'dark' ? 0.3 : 0.2),
              '&:hover, &[aria-controls="theme-mode-menu"]': {
                color: theme.palette.common.white,
                background: iconColor,
                borderColor: iconColor
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
                    shadow={theme.palette.mode === 'dark' ? 'none' : theme.shadows[12]}
                    sx={{
                      minWidth: 320,
                      maxWidth: 360,
                      bgcolor: theme.palette.mode === 'dark' ? alpha(theme.palette.background.default, 0.96) : 'background.paper',
                      border: '1px solid',
                      borderColor: alpha(theme.palette.divider, theme.palette.mode === 'dark' ? 0.9 : 0.72)
                    }}
                  >
                    <Stack sx={{ p: 2.5, gap: 2 }}>
                      <Box>
                        <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
                          主题模式
                        </Typography>
                        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
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
