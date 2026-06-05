import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { alpha, useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import ClickAwayListener from '@mui/material/ClickAwayListener';
import Paper from '@mui/material/Paper';
import Popper from '@mui/material/Popper';
import Stack from '@mui/material/Stack';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import List from '@mui/material/List';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';

import TranslateIcon from '@mui/icons-material/Translate';
import CheckIcon from '@mui/icons-material/Check';

import { LANGUAGE_OPTIONS, normalizeLanguage } from 'i18n/locales';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getHeaderPopoverTokens, getHeaderTriggerTokens } from '../headerPopoverTokens';
import MainCard from 'ui-component/cards/MainCard';
import Transitions from 'ui-component/extended/Transitions';

export default function LanguageSection() {
  const theme = useTheme();
  const downMD = useMediaQuery(theme.breakpoints.down('md'));
  const { isDark } = useResolvedColorScheme();
  const { t, i18n } = useTranslation();

  const [open, setOpen] = useState(false);
  const anchorRef = useRef(null);

  const currentLanguage = normalizeLanguage(i18n.resolvedLanguage || i18n.language);
  const currentLangLabel = t(`language.${currentLanguage}`);

  const accentColor = theme.palette.primary.main;
  const {
    popoverSurface,
    popoverSurfaceAccent,
    popoverBorder,
    popoverInsetShadow,
    primaryText,
    secondaryText,
    mutedText,
    listItemHover,
    selectedSurface,
    selectedHoverSurface
  } = getHeaderPopoverTokens(theme, isDark);
  const { triggerColor, triggerSurface, triggerBorder, activeColor, activeSurface, activeBorder } = getHeaderTriggerTokens(
    theme,
    isDark,
    accentColor,
    {
      lightSurfaceAlpha: 0.14,
      lightHoverAlpha: 0.22
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

  const handleLanguageChange = (lng) => {
    i18n.changeLanguage(normalizeLanguage(lng));
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
        <Tooltip title={currentLangLabel}>
          <Avatar
            ref={anchorRef}
            variant="rounded"
            aria-controls={open ? 'language-menu' : undefined}
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
              '&:hover, &[aria-controls="language-menu"]': {
                color: activeColor,
                background: activeSurface,
                borderColor: activeBorder
              }
            }}
          >
            <TranslateIcon fontSize="small" />
          </Avatar>
        </Tooltip>
      </Box>

      <Popper
        id="language-menu"
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
                      minWidth: 280,
                      maxWidth: 320,
                      bgcolor: popoverSurface,
                      backgroundImage: popoverSurfaceAccent,
                      border: '1px solid',
                      borderColor: popoverBorder,
                      boxShadow: popoverInsetShadow
                    }}
                  >
                    <Stack sx={{ p: 2.5, gap: 1 }}>
                      <Box sx={{ mb: 1 }}>
                        <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
                          {t('language.title')}
                        </Typography>
                        <Typography variant="body2" sx={{ mt: 0.5, color: mutedText }}>
                          {t('language.description')}
                        </Typography>
                      </Box>
                      <List component="nav" sx={{ p: 0 }}>
                        {LANGUAGE_OPTIONS.map((item) => (
                          <ListItemButton
                            key={item.value}
                            selected={currentLanguage === item.value}
                            onClick={() => handleLanguageChange(item.value)}
                            sx={{
                              borderRadius: 1.5,
                              mb: 0.5,
                              color: primaryText,
                              border: '1px solid transparent',
                              '&:hover': {
                                bgcolor: listItemHover,
                                borderColor: alpha(theme.palette.primary.main, isDark ? 0.22 : 0.12)
                              },
                              '&.Mui-selected': {
                                bgcolor: selectedSurface,
                                borderColor: alpha(theme.palette.primary.main, isDark ? 0.36 : 0.2),
                                color: primaryText
                              },
                              '&.Mui-selected:hover': {
                                bgcolor: selectedHoverSurface
                              }
                            }}
                          >
                            <ListItemText
                              primary={
                                <Typography
                                  variant="body1"
                                  sx={{ color: primaryText, fontWeight: currentLanguage === item.value ? 700 : 500 }}
                                >
                                  {t(`language.${item.value}`)}
                                </Typography>
                              }
                              secondary={
                                <Typography variant="caption" sx={{ display: 'block', mt: 0.25, color: secondaryText }}>
                                  {t(`language.${item.value}Desc`)}
                                </Typography>
                              }
                            />
                            {currentLanguage === item.value && <CheckIcon color="primary" fontSize="small" />}
                          </ListItemButton>
                        ))}
                      </List>
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
