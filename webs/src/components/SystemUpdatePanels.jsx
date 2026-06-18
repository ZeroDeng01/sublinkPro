import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import ReactMarkdown from 'react-markdown';

import { alpha, useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import Skeleton from '@mui/material/Skeleton';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';

import RefreshIcon from '@mui/icons-material/Refresh';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import StarIcon from '@mui/icons-material/Star';
import GitHubIcon from '@mui/icons-material/GitHub';
import BugReportIcon from '@mui/icons-material/BugReport';
import FavoriteIcon from '@mui/icons-material/Favorite';

import MainCard from 'ui-component/cards/MainCard';
import { formatDateTime } from 'i18n/locales';
import { donationConfig, affiliateRecommendationConfig } from 'config/donation';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

const getReadableTokens = (theme, isDark) => {
  const readableTextTokens = getReadableTextTokens(theme, isDark);

  return {
    primaryText: readableTextTokens.primaryText,
    secondaryText: isDark ? withAlpha(readableTextTokens.primaryText, 0.84) : withAlpha(readableTextTokens.primaryText, 0.76),
    tertiaryText: isDark ? withAlpha(readableTextTokens.primaryText, 0.72) : readableTextTokens.tertiaryText
  };
};

const getReadablePrimaryTextColor = (theme, isDark) => getReadableTokens(theme, isDark).primaryText;
const getReadableSecondaryTextColor = (theme, isDark) => getReadableTokens(theme, isDark).secondaryText;

const getCalmSurface = (theme, accentColor, isDark) => {
  const { palette, dialogSurface, panelBorder } = getSurfaceTokens(theme, isDark);
  const darkSurfaceElevated = isDark ? withAlpha(palette.background.paper, 0.82) : dialogSurface;
  const calmSurfaceBackground = isDark ? `linear-gradient(180deg, ${darkSurfaceElevated} 0%, ${dialogSurface} 100%)` : 'none';

  return {
    backgroundColor: dialogSurface,
    backgroundImage: calmSurfaceBackground,
    border: `1px solid ${isDark ? panelBorder : alpha(accentColor, 0.12)}`,
    boxShadow: isDark
      ? `0 14px 34px ${alpha(theme.palette.common.black, 0.22)}, inset 0 1px 0 ${alpha(theme.palette.common.white, 0.04)}`
      : `0 1px 3px ${alpha(theme.palette.common.black, 0.06)}`,
    backdropFilter: isDark ? 'blur(10px)' : 'none',
    transition: 'border-color 0.2s ease, box-shadow 0.2s ease, background-color 0.2s ease',
    '&:hover': {
      borderColor: isDark ? alpha(accentColor, 0.24) : alpha(accentColor, 0.2),
      boxShadow: isDark
        ? `0 18px 42px ${alpha(theme.palette.common.black, 0.26)}, inset 0 1px 0 ${alpha(theme.palette.common.white, 0.06)}`
        : `0 4px 12px ${alpha(theme.palette.common.black, 0.08)}`
    }
  };
};

const getAccentChipSx = (theme, accentColor, isDark) => ({
  bgcolor: alpha(accentColor, isDark ? 0.18 : 0.08),
  color: isDark ? withAlpha(getReadablePrimaryTextColor(theme, isDark), 0.92) : withAlpha(accentColor, 0.92),
  border: `1px solid ${alpha(accentColor, isDark ? 0.3 : 0.2)}`,
  fontWeight: 600,
  '&:hover': {
    bgcolor: alpha(accentColor, isDark ? 0.24 : 0.12)
  }
});

const getReadableWarningAccentColor = (theme, isDark) =>
  isDark ? withAlpha(theme.palette.warning.light, 0.94) : theme.palette.warning.dark;

const getGitHubChipSx = (theme, isDark) => ({
  bgcolor: isDark ? alpha(theme.palette.common.white, 0.08) : alpha(theme.palette.common.black, 0.04),
  color: isDark ? withAlpha(getReadablePrimaryTextColor(theme, isDark), 0.98) : '#24292f',
  border: `1px solid ${isDark ? alpha(theme.palette.common.white, 0.18) : 'rgba(27, 31, 36, 0.15)'}`,
  fontWeight: 600,
  '&:hover': {
    bgcolor: isDark ? alpha(theme.palette.common.white, 0.12) : alpha(theme.palette.common.black, 0.07)
  },
  '& .MuiChip-icon': {
    color: 'inherit'
  }
});

export const StarReminderCard = () => {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { t } = useTranslation();
  const [starCount, setStarCount] = useState(null);
  const supportAccent = theme.palette.warning.main;
  const supportAccentReadable = getReadableWarningAccentColor(theme, isDark);
  const supportAccentSoft = alpha(supportAccent, isDark ? 0.18 : 0.12);
  const supportAccentBorder = alpha(supportAccent, isDark ? 0.32 : 0.2);

  useEffect(() => {
    const fetchStarCount = async () => {
      try {
        const response = await fetch('https://api.github.com/repos/ZeroDeng01/sublinkPro');
        if (response.ok) {
          const data = await response.json();
          setStarCount(data.stargazers_count);
        }
      } catch (error) {
        console.error('获取Star数量失败:', error);
      }
    };
    fetchStarCount();
  }, []);

  const handleStar = () => {
    window.open('https://github.com/ZeroDeng01/sublinkPro', '_blank');
  };

  const handleFeedback = () => {
    window.open('https://github.com/ZeroDeng01/sublinkPro/issues', '_blank');
  };

  return (
    <Card
      sx={{
        ...getCalmSurface(theme, supportAccent, isDark),
        mb: 3,
        borderRadius: 3,
        position: 'relative',
        overflow: 'hidden',
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          height: 3,
          backgroundColor: supportAccent
        }
      }}
    >
      <CardContent sx={{ py: 2.5, px: 3, position: 'relative' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 2 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, flex: 1, minWidth: { xs: '100%', sm: 280 } }}>
            <Box
              sx={{
                width: 48,
                height: 48,
                borderRadius: 2,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                backgroundColor: supportAccentSoft,
                border: `1px solid ${supportAccentBorder}`,
                flexShrink: 0
              }}
            >
              <StarIcon sx={{ fontSize: 28, color: supportAccentReadable }} />
            </Box>
            <Box>
              <Typography
                variant="subtitle1"
                sx={{ fontWeight: 600, color: supportAccentReadable, display: 'flex', alignItems: 'center', gap: 0.5 }}
              >
                {t('dashboard.default.star.like')}
                <FavoriteIcon sx={{ fontSize: 16, color: 'error.main' }} />
              </Typography>
              <Typography variant="body2" sx={{ color: getReadableSecondaryTextColor(theme, isDark) }}>
                {t('dashboard.default.star.desc')}
              </Typography>
            </Box>
          </Box>

          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: { xs: 'flex-start', sm: 'flex-end' },
              gap: 1.5,
              flexShrink: 0,
              width: { xs: '100%', sm: 'auto' },
              mt: { xs: 1.5, sm: 0 }
            }}
          >
            <Box
              sx={{ display: 'flex', alignItems: 'center', gap: 1, flexWrap: 'wrap', justifyContent: { xs: 'flex-start', sm: 'flex-end' } }}
            >
              {donationConfig.links.map((item, index) => (
                <Chip
                  key={item.id || index}
                  icon={item.icon}
                  label={t(`donation.links.${item.id}`, item.title)}
                  component="a"
                  href={item.url}
                  target="_blank"
                  clickable
                  sx={{
                    fontWeight: 600,
                    px: 0.5,
                    height: 36,
                    borderRadius: 2,
                    bgcolor: isDark ? alpha(theme.palette[item.color].main, 0.15) : alpha(theme.palette[item.color].light, 0.5),
                    color: isDark ? theme.palette[item.color].light : theme.palette[item.color].dark,
                    border: `1px solid ${isDark ? alpha(theme.palette[item.color].main, 0.3) : alpha(theme.palette[item.color].main, 0.2)}`,
                    transition: 'background-color 0.2s ease, border-color 0.2s ease',
                    '&:hover': {
                      bgcolor: isDark ? alpha(theme.palette[item.color].main, 0.22) : alpha(theme.palette[item.color].light, 0.7)
                    },
                    '& .MuiChip-icon': {
                      color: 'inherit',
                      fontSize: 18,
                      ml: 1
                    }
                  }}
                />
              ))}
            </Box>

            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
                gap: 1.5,
                flexWrap: 'wrap',
                justifyContent: { xs: 'flex-start', sm: 'flex-end' }
              }}
            >
              <Tooltip title={t('dashboard.default.star.feedback')} arrow>
                <IconButton
                  onClick={handleFeedback}
                  size="small"
                  sx={{
                    bgcolor: supportAccentSoft,
                    color: supportAccentReadable,
                    width: 36,
                    height: 36,
                    borderRadius: 2,
                    border: '1px solid',
                    borderColor: supportAccentBorder,
                    '&:hover': {
                      bgcolor: alpha(supportAccent, isDark ? 0.24 : 0.18)
                    }
                  }}
                >
                  <BugReportIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              <Chip
                icon={<GitHubIcon sx={{ fontSize: 18, color: 'inherit !important' }} />}
                label={
                  starCount !== null
                    ? t('dashboard.default.star.starCount', { count: starCount >= 1000 ? `${(starCount / 1000).toFixed(1)}k` : starCount })
                    : 'Star'
                }
                onClick={handleStar}
                sx={{
                  fontWeight: 600,
                  px: 1,
                  height: 36,
                  borderRadius: 2,
                  ...getGitHubChipSx(theme, isDark),
                  cursor: 'pointer',
                  '& .MuiChip-icon': {
                    color: 'inherit'
                  }
                }}
              />
            </Box>
          </Box>
        </Box>

        {affiliateRecommendationConfig && affiliateRecommendationConfig.items.length > 0 && (
          <Box sx={{ mt: 2.5, pt: 2, borderTop: `1px dashed ${alpha(supportAccentBorder, 0.5)}` }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
              <Typography
                variant="subtitle2"
                sx={{ fontWeight: 600, color: supportAccentReadable, display: 'flex', alignItems: 'center', gap: 0.5 }}
              >
                {t(affiliateRecommendationConfig.titleKey || 'affiliate.title')}
              </Typography>
              <Typography
                variant="caption"
                sx={{
                  color: 'text.secondary',
                  display: { xs: 'none', sm: 'block' },
                  flex: 1,
                  textOverflow: 'ellipsis',
                  overflow: 'hidden',
                  whiteSpace: 'nowrap'
                }}
              >
                {t(affiliateRecommendationConfig.disclaimerKey || 'affiliate.disclaimer')}
              </Typography>
            </Box>
            <Typography variant="caption" sx={{ color: 'text.secondary', display: { xs: 'block', sm: 'none' }, mb: 1.5 }}>
              {t(affiliateRecommendationConfig.disclaimerKey || 'affiliate.disclaimer')}
            </Typography>
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: 'repeat(auto-fit, minmax(280px, 1fr))' }, gap: 1.5 }}>
              {affiliateRecommendationConfig.items.map((item, index) => (
                <Box
                  key={item.id || index}
                  component="a"
                  href={item.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  sx={{
                    display: 'flex',
                    alignItems: 'flex-start',
                    gap: 1.5,
                    p: 1.5,
                    borderRadius: 2,
                    textDecoration: 'none',
                    bgcolor: isDark
                      ? alpha(theme.palette[item.color || 'primary'].main, 0.08)
                      : alpha(theme.palette[item.color || 'primary'].light, 0.3),
                    border: '1px solid',
                    borderColor: isDark
                      ? alpha(theme.palette[item.color || 'primary'].main, 0.2)
                      : alpha(theme.palette[item.color || 'primary'].main, 0.1),
                    transition: 'all 0.2s',
                    '&:hover': {
                      bgcolor: isDark
                        ? alpha(theme.palette[item.color || 'primary'].main, 0.15)
                        : alpha(theme.palette[item.color || 'primary'].light, 0.6),
                      transform: 'translateY(-2px)'
                    }
                  }}
                >
                  <Box
                    sx={{
                      width: 36,
                      height: 36,
                      borderRadius: 1.5,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      bgcolor: isDark
                        ? alpha(theme.palette[item.color || 'primary'].main, 0.2)
                        : alpha(theme.palette[item.color || 'primary'].main, 0.15),
                      color: isDark ? theme.palette[item.color || 'primary'].light : theme.palette[item.color || 'primary'].dark,
                      flexShrink: 0
                    }}
                  >
                    {item.icon}
                  </Box>
                  <Box sx={{ flex: 1, minWidth: 0 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
                      <Typography
                        variant="subtitle2"
                        sx={{ fontWeight: 600, color: 'text.primary', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
                      >
                        {t(`affiliate.items.${item.id}.title`, item.title)}
                      </Typography>
                      {item.tag && (
                        <Chip
                          label={t(`affiliate.items.${item.id}.tag`, item.tag)}
                          size="small"
                          sx={{
                            height: 18,
                            fontSize: '0.65rem',
                            bgcolor: isDark
                              ? alpha(theme.palette[item.color || 'primary'].main, 0.2)
                              : alpha(theme.palette[item.color || 'primary'].light, 0.8),
                            color: isDark ? theme.palette[item.color || 'primary'].light : theme.palette[item.color || 'primary'].dark,
                            borderRadius: 1
                          }}
                        />
                      )}
                    </Box>
                    <Typography
                      variant="caption"
                      sx={{
                        color: 'text.secondary',
                        display: '-webkit-box',
                        WebkitLineClamp: 2,
                        WebkitBoxOrient: 'vertical',
                        overflow: 'hidden'
                      }}
                    >
                      {t(`affiliate.items.${item.id}.description`, item.description)}
                    </Typography>
                    {item.highlights?.length > 0 && (
                      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5, mt: 1 }}>
                        {item.highlights.map((highlight, hIndex) => (
                          <Chip
                            key={hIndex}
                            label={t(`affiliate.items.${item.id}.highlights.${hIndex}`, highlight)}
                            size="small"
                            sx={{
                              height: 20,
                              fontSize: '0.65rem',
                              bgcolor: 'transparent',
                              color: 'text.secondary',
                              border: '1px solid',
                              borderColor: isDark
                                ? alpha(theme.palette[item.color || 'primary'].main, 0.24)
                                : alpha(theme.palette[item.color || 'primary'].main, 0.18)
                            }}
                          />
                        ))}
                      </Box>
                    )}
                    <Typography
                      variant="caption"
                      sx={{
                        color: isDark ? theme.palette[item.color || 'primary'].light : theme.palette[item.color || 'primary'].dark,
                        fontWeight: 600
                      }}
                    >
                      {t(`affiliate.items.${item.id}.ctaLabel`, item.ctaLabel)} →
                    </Typography>
                  </Box>
                </Box>
              ))}
            </Box>
          </Box>
        )}
      </CardContent>
    </Card>
  );
};

const ReleaseCard = ({ release }) => {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { t, i18n } = useTranslation();

  return (
    <Card
      sx={{
        ...getCalmSurface(theme, theme.palette.primary.main, isDark),
        mb: 2.5,
        borderRadius: 3,
        transition: 'border-color 0.2s ease, box-shadow 0.2s ease',
        '&:hover': {
          boxShadow: isDark
            ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.06)}`
            : `0 4px 12px ${alpha(theme.palette.common.black, 0.08)}`,
          borderColor: theme.palette.primary.main
        }
      }}
    >
      <CardContent sx={{ p: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
          <Chip
            label={release.tag_name}
            size="small"
            sx={{ fontWeight: 700, ...getAccentChipSx(theme, theme.palette.primary.main, isDark), borderRadius: 2, px: 0.5 }}
          />
          <Typography variant="subtitle1" sx={{ fontWeight: 600, flex: 1, color: getReadablePrimaryTextColor(theme, isDark) }}>
            {release.name}
          </Typography>
          <Chip
            label={formatDateTime(release.published_at, i18n.resolvedLanguage || i18n.language, { month: 'short', day: 'numeric' })}
            size="small"
            variant="outlined"
            sx={{ borderRadius: 2 }}
          />
          <Tooltip title={t('dashboard.default.releases.viewGithub')} arrow>
            <IconButton
              size="small"
              component="a"
              href={release.html_url}
              target="_blank"
              rel="noopener noreferrer"
              sx={{
                color: theme.palette.primary.main,
                '&:hover': {
                  background: alpha(theme.palette.primary.main, 0.1)
                }
              }}
            >
              <OpenInNewIcon fontSize="small" />
            </IconButton>
          </Tooltip>
        </Box>
        <Divider sx={{ mb: 2, opacity: 0.5 }} />
        <Box
          sx={{
            '& h1, & h2, & h3': {
              fontSize: '1rem',
              fontWeight: 600,
              mt: 1.5,
              mb: 0.5,
              color: getReadablePrimaryTextColor(theme, isDark)
            },
            '& p': {
              mb: 1,
              fontSize: '0.875rem',
              lineHeight: 1.7,
              color: getReadableSecondaryTextColor(theme, isDark)
            },
            '& ul, & ol': {
              pl: 2.5,
              mb: 1
            },
            '& li': {
              fontSize: '0.875rem',
              mb: 0.5,
              color: getReadableSecondaryTextColor(theme, isDark),
              '&::marker': {
                color: theme.palette.primary.main
              }
            },
            '& code': {
              backgroundColor: isDark ? alpha(theme.palette.primary.main, 0.14) : alpha(theme.palette.primary.main, 0.1),
              color: isDark ? theme.palette.primary.light : theme.palette.primary.main,
              padding: '2px 8px',
              borderRadius: 6,
              fontSize: '0.8rem',
              fontFamily: '"JetBrains Mono", monospace'
            },
            '& pre': {
              backgroundColor: isDark ? alpha(theme.palette.background.paper, 0.52) : alpha(theme.palette.background.default, 0.78),
              padding: 2,
              borderRadius: 2,
              overflow: 'auto',
              border: `1px solid ${isDark ? alpha(theme.palette.divider, 0.9) : alpha(theme.palette.divider, 0.72)}`,
              color: getReadableSecondaryTextColor(theme, isDark),
              '& code': {
                backgroundColor: 'transparent',
                padding: 0,
                color: isDark ? alpha(theme.palette.primary.light, 0.94) : theme.palette.primary.main
              }
            },
            '& a': {
              color: theme.palette.primary.main,
              textDecoration: 'none',
              fontWeight: 500,
              '&:hover': {
                textDecoration: 'underline'
              }
            }
          }}
        >
          <ReactMarkdown>{release.body || t('dashboard.default.releases.noDescription')}</ReactMarkdown>
        </Box>
      </CardContent>
    </Card>
  );
};

export const ReleaseLogPanel = () => {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const [releases, setReleases] = useState([]);
  const [loadingReleases, setLoadingReleases] = useState(true);

  const fetchReleases = async () => {
    try {
      setLoadingReleases(true);
      const response = await fetch('https://api.github.com/repos/ZeroDeng01/sublinkPro/releases?per_page=5');
      if (!response.ok) throw new Error('Failed to fetch releases');
      const data = await response.json();
      setReleases(data);
    } catch (error) {
      console.error('获取发布日志失败:', error);
      setReleases([]);
    } finally {
      setLoadingReleases(false);
    }
  };

  useEffect(() => {
    fetchReleases();
  }, []);

  return (
    <MainCard
      title={
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
          <Box
            sx={{
              width: 36,
              height: 36,
              borderRadius: 2,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              backgroundColor: alpha(theme.palette.secondary.main, isDark ? 0.18 : 0.1),
              border: `1px solid ${alpha(theme.palette.secondary.main, isDark ? 0.32 : 0.18)}`
            }}
          >
            <Typography sx={{ fontSize: '1.2rem' }}>📝</Typography>
          </Box>
          <Typography variant="h4" sx={{ fontWeight: 600 }}>
            {t('dashboard.default.releases.title')}
          </Typography>
        </Box>
      }
      secondary={
        <Tooltip title={t('dashboard.default.releases.refresh')} arrow>
          <Box component="span" sx={{ display: 'inline-block' }}>
            <IconButton
              onClick={fetchReleases}
              disabled={loadingReleases}
              sx={{
                '&:hover': {
                  background: alpha(theme.palette.primary.main, 0.1)
                }
              }}
            >
              <RefreshIcon />
            </IconButton>
          </Box>
        </Tooltip>
      }
      sx={{
        ...getCalmSurface(theme, theme.palette.secondary.main, isDark),
        borderRadius: 4,
        overflow: 'hidden',
        '& .MuiCardHeader-root': {
          borderBottom: `1px solid ${isDark ? alpha(theme.palette.divider, 0.9) : alpha(theme.palette.divider, 0.72)}`
        }
      }}
    >
      {loadingReleases ? (
        <Box>
          {[1, 2, 3].map((i) => (
            <Box key={i} sx={{ mb: 2.5 }}>
              <Skeleton
                variant="rectangular"
                height={140}
                sx={{
                  borderRadius: 3,
                  bgcolor: isDark ? alpha(theme.palette.background.paper, 0.44) : alpha(theme.palette.background.default, 0.56)
                }}
              />
            </Box>
          ))}
        </Box>
      ) : releases.length > 0 ? (
        releases.map((release) => <ReleaseCard key={release.id} release={release} />)
      ) : (
        <Box sx={{ textAlign: 'center', py: 8, px: 3 }}>
          <Typography sx={{ fontSize: '3rem', mb: 2 }}>📭</Typography>
          <Typography variant="h6" color="text.secondary" sx={{ fontWeight: 500 }}>
            {t('dashboard.default.releases.noReleases')}
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
            {t('dashboard.default.releases.networkError')}
          </Typography>
        </Box>
      )}
    </MainCard>
  );
};
