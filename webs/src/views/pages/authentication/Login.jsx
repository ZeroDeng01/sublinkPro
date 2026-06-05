import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';

import useMediaQuery from '@mui/material/useMediaQuery';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Tooltip from '@mui/material/Tooltip';
import { alpha, useColorScheme, useTheme } from '@mui/material/styles';

import DarkModeOutlinedIcon from '@mui/icons-material/DarkModeOutlined';
import LightModeOutlinedIcon from '@mui/icons-material/LightModeOutlined';
import SettingsBrightnessOutlinedIcon from '@mui/icons-material/SettingsBrightnessOutlined';
import TranslateIcon from '@mui/icons-material/Translate';
import CheckIcon from '@mui/icons-material/Check';

// project imports
import AuthWrapper1 from './AuthWrapper1';
import AuthCardWrapper from './AuthCardWrapper';

import Logo from 'ui-component/Logo';
import AuthFooter from 'ui-component/cards/AuthFooter';
import AuthLogin from '../auth-forms/AuthLogin';
import { DEFAULT_THEME_MODE } from 'config';
import { LANGUAGE_OPTIONS, normalizeLanguage } from 'i18n/locales';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

const themeModeOptions = [
  { value: 'system', icon: SettingsBrightnessOutlinedIcon },
  { value: 'light', icon: LightModeOutlinedIcon },
  { value: 'dark', icon: DarkModeOutlinedIcon }
];

function LoginFloatingControls() {
  const theme = useTheme();
  const { mode, setMode } = useColorScheme();
  const { isDark } = useResolvedColorScheme();
  const { t, i18n } = useTranslation();
  const selectedMode = mode || DEFAULT_THEME_MODE;
  const currentLanguage = normalizeLanguage(i18n.resolvedLanguage || i18n.language);

  const switchTokens = useMemo(
    () => ({
      surface: isDark ? alpha(theme.palette.common.white, 0.045) : alpha(theme.palette.background.default, 0.88),
      border: isDark ? alpha(theme.palette.common.white, 0.12) : alpha(theme.palette.divider, 0.72),
      shadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.05)}` : `0 12px 30px ${alpha(theme.palette.grey[500], 0.12)}`,
      text: isDark ? alpha(theme.palette.common.white, 0.78) : theme.palette.text.secondary,
      hoverText: isDark ? alpha(theme.palette.common.white, 0.96) : theme.palette.text.primary,
      hover: isDark ? alpha(theme.palette.common.white, 0.08) : alpha(theme.palette.primary.main, 0.08),
      selected: isDark ? alpha(theme.palette.primary.main, 0.2) : alpha(theme.palette.primary.main, 0.1),
      selectedBorder: alpha(theme.palette.primary.main, isDark ? 0.54 : 0.28),
      selectedText: isDark ? theme.palette.primary.light : theme.palette.primary.main
    }),
    [isDark, theme]
  );

  const controlGroupSx = {
    p: 0.375,
    borderRadius: 999,
    display: 'flex',
    gap: 0.375,
    bgcolor: switchTokens.surface,
    border: '1px solid',
    borderColor: switchTokens.border,
    boxShadow: switchTokens.shadow,
    backdropFilter: 'blur(16px)',
    WebkitBackdropFilter: 'blur(16px)'
  };

  const controlButtonSx = (selected) => ({
    minWidth: { xs: 36, sm: 74 },
    height: 34,
    px: { xs: 0, sm: 1 },
    border: '1px solid',
    borderColor: selected ? switchTokens.selectedBorder : 'transparent',
    borderRadius: 999,
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 0.75,
    color: selected ? switchTokens.selectedText : switchTokens.text,
    bgcolor: selected ? switchTokens.selected : 'transparent',
    cursor: 'pointer',
    font: 'inherit',
    transition: theme.transitions.create(['background-color', 'border-color', 'color', 'transform'], {
      duration: theme.transitions.duration.shorter
    }),
    '&:hover': {
      bgcolor: selected ? switchTokens.selected : switchTokens.hover,
      color: selected ? switchTokens.selectedText : switchTokens.hoverText
    },
    '&:focus-visible': {
      outline: `2px solid ${alpha(theme.palette.primary.main, 0.48)}`,
      outlineOffset: 2
    },
    '&:active': {
      transform: 'scale(0.97)'
    }
  });

  return (
    <Box
      sx={{
        position: 'absolute',
        top: { xs: 16, sm: 24, md: 32 },
        right: { xs: 16, sm: 24, md: 32 },
        zIndex: 1200,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'flex-end',
        gap: 1
      }}
    >
      <Box component="nav" aria-label={t('language.title')} sx={controlGroupSx}>
        {LANGUAGE_OPTIONS.map((item) => {
          const selected = currentLanguage === item.value;

          return (
            <Tooltip key={item.value} title={t(`language.${item.value}`)}>
              <Box
                component="button"
                type="button"
                aria-pressed={selected}
                aria-label={t(`language.${item.value}`)}
                onClick={() => i18n.changeLanguage(normalizeLanguage(item.value))}
                sx={controlButtonSx(selected)}
              >
                {selected ? <CheckIcon fontSize="small" /> : <TranslateIcon fontSize="small" />}
                <Typography
                  component="span"
                  variant="caption"
                  sx={{ display: { xs: 'none', sm: 'inline' }, fontWeight: selected ? 700 : 600, lineHeight: 1 }}
                >
                  {t(`language.${item.value}Short`)}
                </Typography>
              </Box>
            </Tooltip>
          );
        })}
      </Box>

      <Box component="nav" aria-label={t('theme.title')} sx={controlGroupSx}>
        {themeModeOptions.map((item) => {
          const selected = selectedMode === item.value;
          const Icon = item.icon;

          return (
            <Tooltip key={item.value} title={t(`theme.${item.value}`)}>
              <Box
                component="button"
                type="button"
                aria-pressed={selected}
                aria-label={t(`theme.${item.value}`)}
                onClick={() => setMode(item.value)}
                sx={controlButtonSx(selected)}
              >
                <Icon fontSize="small" />
                <Typography
                  component="span"
                  variant="caption"
                  sx={{ display: { xs: 'none', sm: 'inline' }, fontWeight: selected ? 700 : 600, lineHeight: 1 }}
                >
                  {t(`theme.${item.value}Short`)}
                </Typography>
              </Box>
            </Tooltip>
          );
        })}
      </Box>
    </Box>
  );
}

// ================================|| 登录页面 ||================================ //

export default function Login() {
  const downMD = useMediaQuery((theme) => theme.breakpoints.down('md'));
  const { t } = useTranslation();

  return (
    <AuthWrapper1>
      <LoginFloatingControls />
      <Stack sx={{ justifyContent: 'flex-end', minHeight: '100vh' }}>
        <Stack sx={{ justifyContent: 'center', alignItems: 'center', minHeight: 'calc(100vh - 68px)' }}>
          <Box sx={{ m: { xs: 1, sm: 3 }, mb: 0 }}>
            <AuthCardWrapper>
              <Stack sx={{ alignItems: 'center', justifyContent: 'center', gap: 2 }}>
                <Box sx={{ mb: 3 }}>
                  <Logo />
                </Box>
                <Stack sx={{ alignItems: 'center', justifyContent: 'center', gap: 1 }}>
                  <Typography variant={downMD ? 'h3' : 'h2'} sx={{ color: 'secondary.main' }}>
                    {t('auth.page.welcome')}
                  </Typography>
                  <Typography variant="caption" sx={{ fontSize: '16px', textAlign: { xs: 'center', md: 'inherit' } }}>
                    {t('auth.page.subtitle')}
                  </Typography>
                </Stack>
                <Box sx={{ width: 1 }}>
                  <AuthLogin />
                </Box>
              </Stack>
            </AuthCardWrapper>
          </Box>
        </Stack>
        <Box sx={{ px: 3, my: 3 }}>
          <AuthFooter />
        </Box>
      </Stack>
    </AuthWrapper1>
  );
}
