import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';

// material-ui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import ClickAwayListener from '@mui/material/ClickAwayListener';
import Divider from '@mui/material/Divider';
import Paper from '@mui/material/Paper';
import Popper from '@mui/material/Popper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Tooltip from '@mui/material/Tooltip';
import Button from '@mui/material/Button';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import Transitions from 'ui-component/extended/Transitions';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getHeaderPopoverTokens, getHeaderTriggerTokens } from '../headerPopoverTokens';

import { donationConfig, affiliateRecommendationConfig } from 'config/donation';

export default function DonationSection() {
  const theme = useTheme();
  const { t } = useTranslation();
  const downMD = useMediaQuery(theme.breakpoints.down('md'));
  const { isDark } = useResolvedColorScheme();
  const [open, setOpen] = useState(false);
  const anchorRef = useRef(null);
  const accentColor = theme.palette[donationConfig.headerIconColor].main;

  const { popoverSurface, popoverSurfaceAccent, popoverBorder, popoverInsetShadow } = getHeaderPopoverTokens(theme, isDark);
  const { triggerColor, triggerSurface, triggerBorder, activeSurface, activeBorder } = getHeaderTriggerTokens(theme, isDark, accentColor, {
    lightSurfaceAlpha: 0.12,
    lightHoverAlpha: 0.22,
    activeColor: accentColor
  });

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
      anchorRef.current.focus();
    }
    prevOpen.current = open;
  }, [open]);

  return (
    <>
      <Box sx={{ ml: { xs: 1, md: 2 } }}>
        <Tooltip title={t('donation.tooltip')}>
          <Avatar
            variant="rounded"
            sx={{
              ...theme.typography.commonAvatar,
              ...theme.typography.mediumAvatar,
              transition: 'all .2s ease-in-out',
              color: triggerColor,
              background: triggerSurface,
              border: '1px solid',
              borderColor: triggerBorder,
              position: 'relative',
              '&:hover, &[aria-controls="menu-list-grow"]': {
                color: triggerColor,
                background: activeSurface,
                borderColor: activeBorder
              }
            }}
            ref={anchorRef}
            aria-controls={open ? 'menu-list-grow' : undefined}
            aria-haspopup="true"
            onClick={handleToggle}
          >
            {donationConfig.headerIcon}
          </Avatar>
        </Tooltip>
      </Box>
      <Popper
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
                      bgcolor: popoverSurface,
                      backgroundImage: popoverSurfaceAccent,
                      border: '1px solid',
                      borderColor: popoverBorder,
                      boxShadow: popoverInsetShadow
                    }}
                  >
                    <Stack sx={{ gap: 2, p: 2 }}>
                      <Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'center' }}>
                        <Typography variant="subtitle1">{t('donation.title')}</Typography>
                      </Stack>
                      <Divider />
                      <Stack spacing={1.5}>
                        {donationConfig.links.map((item, index) => (
                          <Button
                            key={index}
                            variant="outlined"
                            color={item.color}
                            startIcon={item.icon}
                            href={item.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            fullWidth
                            sx={{
                              justifyContent: 'flex-start',
                              px: 2,
                              py: 1.5,
                              borderRadius: 2,
                              transition: 'all 0.2s',
                              '&:hover': {
                                transform: 'translateY(-2px)'
                              }
                            }}
                          >
                            {t(`donation.links.${item.id}`, item.title)}
                          </Button>
                        ))}
                      </Stack>

                      {affiliateRecommendationConfig && affiliateRecommendationConfig.items.length > 0 && (
                        <>
                          <Divider />
                          <Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'center' }}>
                            <Typography variant="subtitle2" color="text.secondary">
                              {t('affiliate.title')}
                            </Typography>
                          </Stack>
                          <Stack spacing={1.5}>
                            {affiliateRecommendationConfig.items.map((item, index) => (
                              <Tooltip
                                key={index}
                                title={[
                                  t(`affiliate.items.${item.id}.description`, item.description),
                                  ...(item.highlights || []).map((_, highlightIndex) =>
                                    t(`affiliate.items.${item.id}.highlights.${highlightIndex}`, item.highlights[highlightIndex])
                                  )
                                ].join(' · ')}
                                placement="left"
                                arrow
                              >
                                <Button
                                  variant="outlined"
                                  color={item.color || 'primary'}
                                  startIcon={item.icon}
                                  href={item.url}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  fullWidth
                                  sx={{
                                    justifyContent: 'flex-start',
                                    px: 2,
                                    py: 1,
                                    borderRadius: 2,
                                    transition: 'all 0.2s',
                                    '&:hover': {
                                      transform: 'translateY(-2px)'
                                    }
                                  }}
                                >
                                  {t(`affiliate.items.${item.id}.title`, item.title)} {' · '}
                                  {t(`affiliate.items.${item.id}.ctaLabel`, item.ctaLabel)}
                                </Button>
                              </Tooltip>
                            ))}
                            <Typography
                              variant="caption"
                              sx={{ color: 'text.disabled', textAlign: 'center', fontSize: '0.65rem', mt: 0.5, px: 1 }}
                            >
                              {t('affiliate.disclaimer')}
                            </Typography>
                          </Stack>
                        </>
                      )}
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
