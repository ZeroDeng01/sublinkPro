import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import FormControlLabel from '@mui/material/FormControlLabel';
import Grid from '@mui/material/Grid';
import Radio from '@mui/material/Radio';
import RadioGroup from '@mui/material/RadioGroup';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import { alpha, useColorScheme, useTheme } from '@mui/material/styles';

import DarkModeOutlinedIcon from '@mui/icons-material/DarkModeOutlined';
import LightModeOutlinedIcon from '@mui/icons-material/LightModeOutlined';
import SettingsBrightnessOutlinedIcon from '@mui/icons-material/SettingsBrightnessOutlined';

import { DEFAULT_THEME_MODE } from 'config';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { withAlpha } from 'utils/colorUtils';
import MainCard from 'ui-component/cards/MainCard';

const modeIcons = {
  system: SettingsBrightnessOutlinedIcon,
  light: LightModeOutlinedIcon,
  dark: DarkModeOutlinedIcon
};

export default function ThemeModeSelector({ title, description, sx }) {
  const theme = useTheme();
  const { mode, setMode } = useColorScheme();
  const { isDark } = useResolvedColorScheme();
  const { t } = useTranslation();
  const palette = theme.vars?.palette || theme.palette;
  const selectedMode = mode || DEFAULT_THEME_MODE;
  const resolvedTitle = title === undefined ? t('theme.title') : title;

  const options = ['system', 'light', 'dark'].map((value) => ({
    value,
    label: t(`theme.${value}`),
    description: t(`theme.${value}Desc`),
    icon: modeIcons[value]
  }));

  return (
    <Stack sx={{ gap: 2.5, ...(sx || {}) }}>
      {resolvedTitle ? <Typography variant="h5">{resolvedTitle}</Typography> : null}
      {description ? (
        <Typography variant="body2" color="text.secondary">
          {description}
        </Typography>
      ) : null}

      <RadioGroup aria-label="theme-mode" name="theme-mode" value={selectedMode} onChange={(event) => setMode(event.target.value)}>
        <Grid container spacing={1.25}>
          {options.map((item) => {
            const selected = selectedMode === item.value;
            const Icon = item.icon;

            return (
              <Grid key={item.value} size={12}>
                <MainCard
                  content={false}
                  sx={{
                    p: 0.75,
                    bgcolor: selected
                      ? isDark
                        ? alpha(theme.palette.primary.main, 0.14)
                        : alpha(theme.palette.primary.main, 0.08)
                      : isDark
                        ? withAlpha(palette.background.default, 0.92)
                        : 'background.default',
                    border: '1px solid',
                    borderColor: selected
                      ? alpha(theme.palette.primary.main, isDark ? 0.32 : 0.18)
                      : alpha(theme.palette.divider, isDark ? 0.9 : 0.72)
                  }}
                >
                  <MainCard
                    content={false}
                    border
                    sx={{
                      p: 1.75,
                      borderWidth: 1,
                      bgcolor: isDark ? withAlpha(palette.background.paper, 0.94) : 'background.paper',
                      ...(selected && { borderColor: 'primary.main' })
                    }}
                  >
                    <FormControlLabel
                      sx={{ m: 0, width: 1, alignItems: 'flex-start' }}
                      value={item.value}
                      control={<Radio sx={{ mt: 0.25 }} />}
                      label={
                        <Stack spacing={0.5} sx={{ pl: 1.5 }}>
                          <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
                            <Icon color={selected ? 'primary' : 'action'} fontSize="small" />
                            <Typography variant="h5">{item.label}</Typography>
                          </Stack>
                          <Typography variant="body2" color="text.secondary">
                            {item.description}
                          </Typography>
                        </Stack>
                      }
                    />
                  </MainCard>
                </MainCard>
              </Grid>
            );
          })}
        </Grid>
      </RadioGroup>
    </Stack>
  );
}

ThemeModeSelector.propTypes = {
  title: PropTypes.node,
  description: PropTypes.node,
  sx: PropTypes.object
};
