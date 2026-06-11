import { useState, useEffect, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import ReactMarkdown from 'react-markdown';
import { getNodeDisplayName } from 'utils/nodeDisplayName';
import { formatDateTime, formatNumber } from 'i18n/locales';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import Skeleton from '@mui/material/Skeleton';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import LinearProgress from '@mui/material/LinearProgress';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import useMediaQuery from '@mui/material/useMediaQuery';

// icons
import CloudQueueIcon from '@mui/icons-material/CloudQueue';
import RefreshIcon from '@mui/icons-material/Refresh';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import AutoAwesomeIcon from '@mui/icons-material/AutoAwesome';
import SpeedIcon from '@mui/icons-material/Speed';
import TimerIcon from '@mui/icons-material/Timer';
import StarIcon from '@mui/icons-material/Star';
import GitHubIcon from '@mui/icons-material/GitHub';
import BugReportIcon from '@mui/icons-material/BugReport';
import FavoriteIcon from '@mui/icons-material/Favorite';
import FlightTakeoffIcon from '@mui/icons-material/FlightTakeoff';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';
import EventIcon from '@mui/icons-material/Event';

// icons for protocols
import PublicIcon from '@mui/icons-material/Public';
import FolderIcon from '@mui/icons-material/Folder';
import SourceIcon from '@mui/icons-material/Input';
import LabelIcon from '@mui/icons-material/Label';
import SecurityIcon from '@mui/icons-material/Security';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import TaskProgressPanel from 'components/TaskProgressPanel';
import {
  getNodeTotal,
  getFastestSpeedNode,
  getLowestDelayNode,
  getDashboardCountryStats,
  getDashboardGroupedStats,
  getQualityStats
} from 'api/total';
import { getAirports } from 'api/airports';
import { formatBytes, formatExpireTime, getUsageColor } from 'views/airports/utils';
import { getQualityStatusMeta } from 'utils/fraudScore';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';
import { COUNTRY_FALLBACK_EMOJI, isoToFlag } from 'utils/countryDisplay';

const getDashboardReadableTextTokens = (theme, isDark) => {
  const readableTextTokens = getReadableTextTokens(theme, isDark);

  return {
    primaryText: readableTextTokens.primaryText,
    secondaryText: isDark ? withAlpha(readableTextTokens.primaryText, 0.84) : withAlpha(readableTextTokens.primaryText, 0.76),
    tertiaryText: isDark ? withAlpha(readableTextTokens.primaryText, 0.72) : readableTextTokens.tertiaryText,
    statValueText: isDark ? withAlpha(readableTextTokens.primaryText, 0.98) : readableTextTokens.primaryText,
    statPercentText: isDark ? withAlpha(readableTextTokens.primaryText, 0.92) : withAlpha(readableTextTokens.primaryText, 0.8)
  };
};

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

const getAccentIconBox = (accentColor, isDark) => ({
  width: 40,
  height: 40,
  borderRadius: 2,
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  backgroundColor: alpha(accentColor, isDark ? 0.18 : 0.12),
  border: `1px solid ${alpha(accentColor, isDark ? 0.32 : 0.18)}`,
  color: accentColor,
  flexShrink: 0
});

const getAccentChipSx = (theme, accentColor, isDark) => ({
  ...(() => {
    const { primaryText } = getDashboardReadableTextTokens(theme, isDark);

    return {
      bgcolor: alpha(accentColor, isDark ? 0.18 : 0.08),
      color: isDark ? withAlpha(primaryText, 0.92) : withAlpha(accentColor, 0.92),
      border: `1px solid ${alpha(accentColor, isDark ? 0.3 : 0.2)}`,
      fontWeight: 600,
      '&:hover': {
        bgcolor: alpha(accentColor, isDark ? 0.24 : 0.12)
      }
    };
  })()
});

const getReadablePrimaryTextColor = (theme, isDark) => {
  return getDashboardReadableTextTokens(theme, isDark).primaryText;
};

const getReadableSecondaryTextColor = (theme, isDark) => {
  return getDashboardReadableTextTokens(theme, isDark).secondaryText;
};

const getReadableTertiaryTextColor = (theme, isDark) => {
  return getDashboardReadableTextTokens(theme, isDark).tertiaryText;
};

const getReadableStatValueColor = (theme, isDark) => {
  return getDashboardReadableTextTokens(theme, isDark).statValueText;
};

const getReadableStatPercentColor = (theme, isDark) => {
  return getDashboardReadableTextTokens(theme, isDark).statPercentText;
};

const getReadableWarningAccentColor = (theme, isDark) =>
  isDark ? withAlpha(theme.palette.warning.light, 0.94) : theme.palette.warning.dark;

const getGitHubChipSx = (theme, isDark) => {
  const { primaryText } = getDashboardReadableTextTokens(theme, isDark);

  return {
    bgcolor: isDark ? alpha(theme.palette.common.white, 0.08) : alpha(theme.palette.common.black, 0.04),
    color: isDark ? withAlpha(primaryText, 0.98) : '#24292f',
    border: `1px solid ${isDark ? alpha(theme.palette.common.white, 0.18) : 'rgba(27, 31, 36, 0.15)'}`,
    fontWeight: 600,
    '&:hover': {
      bgcolor: isDark ? alpha(theme.palette.common.white, 0.12) : alpha(theme.palette.common.black, 0.07)
    },
    '& .MuiChip-icon': {
      color: 'inherit'
    }
  };
};

const getInsetPanelSurface = (theme, accentColor, isDark) => {
  const { palette, mutedPanelSurface } = getSurfaceTokens(theme, isDark);
  const darkPanelBase = mutedPanelSurface;
  const darkPanelElevated = withAlpha(palette.background.paper, 0.4);

  return {
    backgroundColor: isDark ? darkPanelBase : withAlpha(palette.background.default, 0.72),
    backgroundImage: isDark ? `linear-gradient(180deg, ${darkPanelElevated} 0%, ${darkPanelBase} 100%)` : 'none',
    border: `1px solid ${isDark ? withAlpha(palette.divider, 0.74) : alpha(accentColor, 0.12)}`,
    boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.025)}` : 'none'
  };
};

const getFlagEmoji = (countryCode) => {
  return isoToFlag(countryCode, COUNTRY_FALLBACK_EMOJI);
};

const protocolColors = {
  Shadowsocks: ['#3b82f6', '#2563eb'],
  ShadowsocksR: ['#6366f1', '#4f46e5'],
  VMess: ['#8b5cf6', '#7c3aed'],
  VLESS: ['#10b981', '#059669'],
  Trojan: ['#ef4444', '#dc2626'],
  Hysteria: ['#06b6d4', '#0891b2'],
  Hysteria2: ['#14b8a6', '#0d9488'],
  TUIC: ['#f59e0b', '#d97706'],
  WireGuard: ['#84cc16', '#65a30d'],
  NaiveProxy: ['#ec4899', '#db2777'],
  SOCKS5: ['#64748b', '#475569'],
  HTTP: ['#94a3b8', '#64748b'],
  HTTPS: ['#22c55e', '#16a34a']
};

const fraudLevelColors = {
  极佳: '#94a3b8',
  优秀: '#22c55e',
  良好: '#eab308',
  中等: '#f97316',
  差: '#ef4444',
  极差: '#111827'
};

const qualityStatusColorMap = {
  success: '#22c55e',
  partial: '#0ea5e9',
  failed: '#ef4444',
  disabled: '#94a3b8',
  untested: '#64748b'
};

const createCountryStatMap = (stats = []) =>
  stats.reduce((accumulator, item) => {
    accumulator[item.country] = item;
    return accumulator;
  }, {});

const TOTAL_COUNT_KEYS = ['total', 'count', 'totalCount', 'nodeCount', 'value'];
const DELAY_PASS_COUNT_KEYS = ['delayPassCount', 'delayPass', 'delayPassedCount', 'delayPassed', 'delayPassTotal'];
const SPEED_PASS_COUNT_KEYS = ['speedPassCount', 'speedPass', 'speedPassedCount', 'speedPassed', 'speedPassTotal'];

const getNumericStatValue = (source, keys = [], fallback = 0) => {
  if (typeof source === 'number') {
    return Number.isFinite(source) ? source : fallback;
  }

  if (typeof source === 'string' && source.trim() !== '') {
    const parsedValue = Number(source);
    return Number.isFinite(parsedValue) ? parsedValue : fallback;
  }

  if (!source || typeof source !== 'object') {
    return fallback;
  }

  for (const key of keys) {
    const value = source[key];

    if (typeof value === 'number' && Number.isFinite(value)) {
      return value;
    }

    if (typeof value === 'string' && value.trim() !== '') {
      const parsedValue = Number(value);
      if (Number.isFinite(parsedValue)) {
        return parsedValue;
      }
    }
  }

  return fallback;
};

const getLabelStatValue = (source, fallbackLabel) => {
  if (!source || typeof source !== 'object') {
    return fallbackLabel;
  }

  return source.label || source.name || source.title || fallbackLabel;
};

const getCountMetric = (source) => getNumericStatValue(source, TOTAL_COUNT_KEYS, 0);
const getDelayPassMetric = (source) => getNumericStatValue(source, DELAY_PASS_COUNT_KEYS, 0);
const getSpeedPassMetric = (source) => getNumericStatValue(source, SPEED_PASS_COUNT_KEYS, 0);

const buildTopItems = (items = [], total = 0, limit = 5, options = {}, t) => {
  const { forceCollapsedKeys = [] } = options;
  const normalizedItems = items.filter((item) => item && item.count > 0);
  const forcedHiddenItems = normalizedItems.filter((item) => forceCollapsedKeys.includes(item.key));
  const eligibleVisibleItems = normalizedItems.filter((item) => !forceCollapsedKeys.includes(item.key));
  const visibleItems = eligibleVisibleItems.slice(0, limit);
  const hiddenItems = [...forcedHiddenItems, ...eligibleVisibleItems.slice(limit)];
  const hiddenCount = hiddenItems.reduce((sum, item) => sum + item.count, 0);
  const hiddenUniqueIpCount = hiddenItems.reduce((sum, item) => sum + (item.uniqueIpCount || 0), 0);
  const hiddenDelayPassCount = hiddenItems.reduce((sum, item) => sum + (item.delayPassCount || 0), 0);
  const hiddenSpeedPassCount = hiddenItems.reduce((sum, item) => sum + (item.speedPassCount || 0), 0);

  if (hiddenCount > 0) {
    visibleItems.push({
      key: 'collapsed-other',
      label: t ? t('dashboard.default.charts.others', { count: hiddenItems.length }) : `Other ${hiddenItems.length} items`,
      count: hiddenCount,
      uniqueIpCount: hiddenUniqueIpCount,
      delayPassCount: hiddenDelayPassCount,
      speedPassCount: hiddenSpeedPassCount,
      color: '#94a3b8',
      tooltip: t
        ? t('dashboard.default.charts.othersDesc', { count: hiddenItems.length, total: hiddenCount })
        : `Includes other ${hiddenItems.length} items, total ${hiddenCount} nodes`,
      isCollapsedOther: true,
      hiddenItems: hiddenItems.map((item) => ({
        ...item,
        percent: total > 0 ? (item.count / total) * 100 : 0
      }))
    });
  }

  return visibleItems.map((item) => ({
    ...item,
    percent: total > 0 ? (item.count / total) * 100 : 0
  }));
};

const normalizeMapStats = ({ entries = [], total, limit, defaultColor, getItemMeta, forceCollapsedKeys = [], t }) => {
  const resolvedTotal = typeof total === 'number' ? total : entries.reduce((sum, [, value]) => sum + getCountMetric(value), 0);
  const normalized = entries
    .map(([key, value], index) => ({
      key,
      label: getLabelStatValue(value, key),
      count: getCountMetric(value),
      delayPassCount: getDelayPassMetric(value),
      speedPassCount: getSpeedPassMetric(value),
      color: defaultColor,
      ...(getItemMeta ? getItemMeta(key, value, index) : {})
    }))
    .sort((a, b) => b.count - a.count);

  return buildTopItems(normalized, resolvedTotal, limit, { forceCollapsedKeys }, t);
};

const normalizeTagStats = ({ tags = [], limit, t }) => {
  const total = tags.reduce((sum, item) => sum + getCountMetric(item), 0);
  const normalized = [...tags]
    .sort((a, b) => getCountMetric(b) - getCountMetric(a))
    .map((tag, index) => ({
      key: tag.key || tag.name || tag.label || `tag-${index}`,
      label: getLabelStatValue(
        tag,
        tag.name || tag.key || (t ? t('dashboard.default.charts.tagItem', { index: index + 1 }) : `Tag ${index + 1}`)
      ),
      count: getCountMetric(tag),
      delayPassCount: getDelayPassMetric(tag),
      speedPassCount: getSpeedPassMetric(tag),
      color: tag.color || '#ec4899'
    }));

  return buildTopItems(normalized, total, limit, {}, t);
};

const getProgressBarSx = (color, isDark, muted = false) => ({
  height: 7,
  borderRadius: 999,
  bgcolor: alpha(color, isDark ? 0.22 : 0.12),
  '& .MuiLinearProgress-bar': {
    borderRadius: 999,
    backgroundColor: muted ? alpha(color, isDark ? 0.7 : 0.62) : color
  }
});

const StatRowsSkeleton = ({ rows = 5 }) => (
  <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
    {Array.from({ length: rows }).map((_, index) => (
      <Box key={index}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.75, gap: 1 }}>
          <Skeleton variant="text" width="42%" height={24} />
          <Skeleton variant="text" width={64} height={24} />
        </Box>
        <Skeleton variant="rounded" height={8} sx={{ borderRadius: 999 }} />
      </Box>
    ))}
  </Box>
);

const StatsChartCard = ({ title, icon: Icon, accentColor, summary, loading, tooltip, children }) => {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();

  return (
    <Card
      sx={{
        ...getCalmSurface(theme, accentColor, isDark),
        borderRadius: 4,
        height: '100%',
        overflow: 'hidden',
        position: 'relative',
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          height: 3,
          backgroundColor: accentColor
        }
      }}
    >
      <CardContent sx={{ p: 3 }}>
        <Box sx={{ display: 'flex', alignItems: { xs: 'flex-start', sm: 'center' }, gap: 1.5, mb: 2.5, flexWrap: 'wrap' }}>
          <Box sx={getAccentIconBox(accentColor, isDark)}>
            <Icon sx={{ fontSize: 22 }} />
          </Box>
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Typography variant="h5" sx={{ fontWeight: 600 }}>
              {title}
            </Typography>
            {tooltip ? (
              <Typography variant="body2" sx={{ mt: 0.5, color: getReadableSecondaryTextColor(theme, isDark) }}>
                {tooltip}
              </Typography>
            ) : null}
          </Box>
          {summary ? (
            <Box
              sx={{
                ml: { xs: 0, sm: 'auto' },
                px: 1.25,
                py: 0.75,
                borderRadius: 2,
                fontSize: '0.75rem',
                fontWeight: 700,
                color: accentColor,
                bgcolor: alpha(accentColor, isDark ? 0.2 : 0.1),
                border: `1px solid ${alpha(accentColor, isDark ? 0.32 : 0.18)}`
              }}
            >
              {summary}
            </Box>
          ) : null}
        </Box>

        {loading ? <StatRowsSkeleton /> : children}
      </CardContent>
    </Card>
  );
};

const RankedStatList = ({
  items = [],
  emptyText,
  percentSuffix = '%',
  valueFormatter,
  labelFormatter,
  mutedKeys = [],
  detailFormatter,
  secondaryMetricsFormatter
}) => {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { t, i18n } = useTranslation();
  const [expandedKeys, setExpandedKeys] = useState({});

  const formatSecondaryMetricValue = (value) => {
    if (typeof value === 'number') {
      return formatNumber(value, i18n.resolvedLanguage || i18n.language);
    }

    if (typeof value === 'string' && value.trim() !== '') {
      return value;
    }

    return '--';
  };

  const renderSecondaryMetrics = (metrics = [], itemColor, itemKey) => {
    if (!metrics.length) return null;

    return (
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          flexWrap: 'wrap',
          gap: 0.625,
          mt: 0.25,
          minWidth: 0
        }}
      >
        {metrics.map((metric, index) => (
          <Box
            key={`${itemKey}-${metric.key || metric.label}`}
            sx={{
              display: 'inline-flex',
              alignItems: 'baseline',
              gap: 0.5,
              minWidth: 0
            }}
          >
            {index > 0 ? (
              <Typography
                component="span"
                variant="caption"
                sx={{
                  color: alpha(itemColor, isDark ? 0.7 : 0.5),
                  fontWeight: 700,
                  lineHeight: 1
                }}
              >
                ·
              </Typography>
            ) : null}
            <Typography variant="caption" sx={{ color: getReadableSecondaryTextColor(theme, isDark), lineHeight: 1.2 }}>
              {metric.label}
            </Typography>
            <Typography
              variant="caption"
              sx={{
                fontWeight: 700,
                color: isDark ? getReadableStatValueColor(theme, isDark) : alpha(itemColor, 0.88),
                lineHeight: 1.2
              }}
            >
              {formatSecondaryMetricValue(metric.value)}
            </Typography>
          </Box>
        ))}
      </Box>
    );
  };

  if (!items.length) {
    return <Typography sx={{ fontSize: '0.875rem', color: getReadableSecondaryTextColor(theme, isDark) }}>{emptyText}</Typography>;
  }

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.75 }}>
      {items.map((item) => {
        const muted = item.isCollapsedOther || mutedKeys.includes(item.key);
        const isExpanded = Boolean(expandedKeys[item.key]);
        const secondaryMetrics = secondaryMetricsFormatter ? secondaryMetricsFormatter(item) : [];
        const toggleExpanded = () => {
          if (!item.isCollapsedOther) return;
          setExpandedKeys((prev) => ({ ...prev, [item.key]: !prev[item.key] }));
        };
        const content = (
          <Box key={item.key}>
            <Box
              onClick={toggleExpanded}
              sx={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                gap: 1.5,
                mb: 0.75,
                cursor: item.isCollapsedOther ? 'pointer' : 'default'
              }}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.25, minWidth: 0, flex: 1 }}>
                {item.marker ? (
                  <Typography sx={{ fontSize: '1rem', lineHeight: 1 }}>{item.marker}</Typography>
                ) : (
                  <Box
                    sx={{
                      width: 10,
                      height: 10,
                      borderRadius: '50%',
                      bgcolor: item.color,
                      flexShrink: 0,
                      boxShadow: `0 0 0 4px ${alpha(item.color, isDark ? 0.18 : 0.12)}`
                    }}
                  />
                )}
                <Box sx={{ minWidth: 0, flex: 1 }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75, minWidth: 0 }}>
                    <Typography
                      variant="subtitle2"
                      sx={{
                        fontWeight: 600,
                        color: getReadablePrimaryTextColor(theme, isDark),
                        minWidth: 0,
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap'
                      }}
                    >
                      {labelFormatter ? labelFormatter(item, isExpanded) : item.label}
                    </Typography>
                    {item.isCollapsedOther ? (
                      <Typography variant="caption" sx={{ color: getReadableSecondaryTextColor(theme, isDark), flexShrink: 0 }}>
                        {isExpanded ? t('dashboard.default.charts.collapse') : t('dashboard.default.charts.expand')}
                      </Typography>
                    ) : null}
                  </Box>
                  {renderSecondaryMetrics(secondaryMetrics, item.color, item.key)}
                </Box>
              </Box>
              <Box sx={{ display: 'flex', alignItems: 'baseline', gap: 0.75, flexShrink: 0 }}>
                <Box sx={{ textAlign: 'right' }}>
                  <Typography variant="subtitle2" sx={{ fontWeight: 700, color: getReadableStatValueColor(theme, isDark) }}>
                    {valueFormatter ? valueFormatter(item.count, item) : formatNumber(item.count, i18n.resolvedLanguage || i18n.language)}
                  </Typography>
                  {detailFormatter ? (
                    <Typography variant="caption" sx={{ color: getReadableSecondaryTextColor(theme, isDark), display: 'block' }}>
                      {detailFormatter(item)}
                    </Typography>
                  ) : null}
                </Box>
                <Typography
                  variant="caption"
                  sx={{
                    color: getReadableStatPercentColor(theme, isDark),
                    minWidth: 42,
                    textAlign: 'right',
                    fontWeight: isDark ? 700 : 500
                  }}
                >
                  {item.percent.toFixed(1)}
                  {percentSuffix}
                </Typography>
              </Box>
            </Box>
            <LinearProgress
              variant="determinate"
              value={Math.max(0, Math.min(item.percent, 100))}
              sx={getProgressBarSx(item.color, isDark, muted)}
            />
            {item.isCollapsedOther && isExpanded && item.hiddenItems?.length ? (
              <Box
                sx={{
                  mt: 1.25,
                  ml: { xs: 1.5, sm: 2 },
                  pl: 1.5,
                  borderLeft: `2px dashed ${alpha(item.color, isDark ? 0.4 : 0.3)}`,
                  display: 'flex',
                  flexDirection: 'column',
                  gap: 1.25
                }}
              >
                {item.hiddenItems.map((hiddenItem) => (
                  <Box key={`${item.key}-${hiddenItem.key}`}>
                    <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 1, mb: 0.5 }}>
                      <Box sx={{ minWidth: 0, flex: 1 }}>
                        <Tooltip title={hiddenItem.tooltip || hiddenItem.label} arrow>
                          <Box
                            sx={{
                              display: 'flex',
                              alignItems: 'center',
                              gap: 0.75,
                              minWidth: 0,
                              color: getReadableSecondaryTextColor(theme, isDark)
                            }}
                          >
                            {hiddenItem.marker ? (
                              <Typography sx={{ fontSize: '1rem', lineHeight: 1 }}>{hiddenItem.marker}</Typography>
                            ) : null}
                            <Typography
                              component="div"
                              variant="caption"
                              sx={{
                                color: 'inherit',
                                fontWeight: 600,
                                minWidth: 0,
                                overflow: 'hidden',
                                textOverflow: 'ellipsis',
                                whiteSpace: 'nowrap'
                              }}
                            >
                              {labelFormatter ? labelFormatter(hiddenItem, false) : hiddenItem.label}
                            </Typography>
                          </Box>
                        </Tooltip>
                        {renderSecondaryMetrics(
                          secondaryMetricsFormatter ? secondaryMetricsFormatter(hiddenItem) : [],
                          hiddenItem.color || item.color,
                          hiddenItem.key
                        )}
                      </Box>
                      <Typography variant="caption" sx={{ color: getReadableSecondaryTextColor(theme, isDark), flexShrink: 0 }}>
                        {valueFormatter
                          ? valueFormatter(hiddenItem.count, hiddenItem)
                          : formatNumber(hiddenItem.count, i18n.resolvedLanguage || i18n.language)}
                        {detailFormatter ? ` · ${detailFormatter(hiddenItem)}` : ''} · {hiddenItem.percent.toFixed(1)}
                        {percentSuffix}
                      </Typography>
                    </Box>
                    <LinearProgress
                      variant="determinate"
                      value={Math.max(0, Math.min(hiddenItem.percent, 100))}
                      sx={getProgressBarSx(hiddenItem.color || item.color, isDark, true)}
                    />
                  </Box>
                ))}
              </Box>
            ) : null}
          </Box>
        );

        return item.tooltip ? (
          <Tooltip title={item.tooltip} arrow key={item.key}>
            <Box>{content}</Box>
          </Tooltip>
        ) : (
          content
        );
      })}
    </Box>
  );
};

const QualityMetricRow = ({ label, count, percent, color, tooltip }) => {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { i18n } = useTranslation();
  const row = (
    <Box>
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1.5, mb: 0.75 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, minWidth: 0 }}>
          <Box
            sx={{
              width: 10,
              height: 10,
              borderRadius: '50%',
              bgcolor: color,
              flexShrink: 0,
              boxShadow: `0 0 0 4px ${alpha(color, isDark ? 0.18 : 0.12)}`
            }}
          />
          <Typography variant="subtitle2" sx={{ fontWeight: 600, minWidth: 0, color: getReadablePrimaryTextColor(theme, isDark) }}>
            {label}
          </Typography>
        </Box>
        <Box sx={{ display: 'flex', alignItems: 'baseline', gap: 0.75, flexShrink: 0 }}>
          <Typography variant="subtitle2" sx={{ fontWeight: 700, color: getReadablePrimaryTextColor(theme, isDark) }}>
            {formatNumber(count, i18n.resolvedLanguage || i18n.language)}
          </Typography>
          <Typography variant="caption" sx={{ color: getReadableSecondaryTextColor(theme, isDark), minWidth: 42, textAlign: 'right' }}>
            {percent.toFixed(1)}%
          </Typography>
        </Box>
      </Box>
      <LinearProgress variant="determinate" value={Math.max(0, Math.min(percent, 100))} sx={getProgressBarSx(color, isDark)} />
    </Box>
  );

  return tooltip ? (
    <Tooltip title={tooltip} arrow>
      <Box>{row}</Box>
    </Tooltip>
  ) : (
    row
  );
};

const IPQualityBreakdown = ({ stats, loading }) => {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { t, i18n } = useTranslation();

  if (loading) {
    return <StatRowsSkeleton rows={5} />;
  }

  if (!stats || !Array.isArray(stats.ipStats) || stats.ipStats.length === 0) {
    return (
      <Typography sx={{ fontSize: '0.875rem', color: getReadableSecondaryTextColor(theme, isDark) }}>
        {t('dashboard.default.charts.emptyIp')}
      </Typography>
    );
  }

  const total = stats.total || 0;
  const successTotal = stats.successTotal || 0;
  const findCount = (key) => stats.ipStats.find((item) => item.key === key)?.count || 0;

  const residentialRows = [
    { key: 'housing', label: t('dashboard.default.ipQuality.housing'), count: findCount('housing'), color: '#22c55e' },
    { key: 'datacenter', label: t('dashboard.default.ipQuality.datacenter'), count: findCount('datacenter'), color: '#64748b' }
  ];
  const typeRows = [
    { key: 'native', label: t('dashboard.default.ipQuality.native'), count: findCount('native'), color: '#06b6d4' },
    { key: 'broadcast', label: t('dashboard.default.ipQuality.broadcast'), count: findCount('broadcast'), color: '#f59e0b' }
  ];
  const otherCount = findCount('other');

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2.25 }}>
      <Box>
        <Typography variant="body2" sx={{ color: getReadableSecondaryTextColor(theme, isDark), mb: 1.25, fontWeight: 600 }}>
          {t('dashboard.default.ipQuality.residential')}
        </Typography>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
          {residentialRows.map((item) => (
            <QualityMetricRow
              key={item.key}
              label={item.label}
              count={item.count}
              percent={successTotal > 0 ? (item.count / successTotal) * 100 : 0}
              color={item.color}
            />
          ))}
        </Box>
      </Box>

      <Box
        sx={{
          pt: 2,
          borderTop: `1px solid ${alpha(theme.palette.text.primary, isDark ? 0.08 : 0.06)}`
        }}
      >
        <Typography variant="body2" sx={{ color: getReadableSecondaryTextColor(theme, isDark), mb: 1.25, fontWeight: 600 }}>
          {t('dashboard.default.ipQuality.ipType')}
        </Typography>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
          {typeRows.map((item) => (
            <QualityMetricRow
              key={item.key}
              label={item.label}
              count={item.count}
              percent={successTotal > 0 ? (item.count / successTotal) * 100 : 0}
              color={item.color}
            />
          ))}
        </Box>
      </Box>

      <Box
        sx={{
          p: 1.5,
          borderRadius: 3,
          bgcolor: alpha('#94a3b8', isDark ? 0.12 : 0.08),
          border: `1px solid ${alpha('#94a3b8', isDark ? 0.24 : 0.16)}`
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1.5 }}>
          <Box>
            <Typography variant="subtitle2" sx={{ fontWeight: 700, color: getReadablePrimaryTextColor(theme, isDark) }}>
              {t('dashboard.default.ipQuality.others')}
            </Typography>
            <Typography variant="caption" sx={{ color: getReadableSecondaryTextColor(theme, isDark) }}>
              {t('dashboard.default.ipQuality.othersDesc')}
            </Typography>
          </Box>
          <Box sx={{ textAlign: 'right' }}>
            <Typography variant="subtitle2" sx={{ fontWeight: 700, color: getReadablePrimaryTextColor(theme, isDark) }}>
              {formatNumber(otherCount, i18n.resolvedLanguage || i18n.language)}
            </Typography>
            <Typography variant="caption" sx={{ color: getReadableSecondaryTextColor(theme, isDark) }}>
              {total > 0 ? ((otherCount / total) * 100).toFixed(1) : '0.0'}%
            </Typography>
          </Box>
        </Box>
      </Box>
    </Box>
  );
};

// ==============================|| 问候语计算 ||============================== //

const getGreeting = (t) => {
  const hour = new Date().getHours();
  if (hour >= 5 && hour < 9) {
    return { text: t('dashboard.default.greeting.morning'), emoji: '🌅', subText: t('dashboard.default.greeting.morningSub') };
  } else if (hour >= 9 && hour < 12) {
    return { text: t('dashboard.default.greeting.lateMorning'), emoji: '☀️', subText: t('dashboard.default.greeting.lateMorningSub') };
  } else if (hour >= 12 && hour < 14) {
    return { text: t('dashboard.default.greeting.noon'), emoji: '🌤️', subText: t('dashboard.default.greeting.noonSub') };
  } else if (hour >= 14 && hour < 18) {
    return { text: t('dashboard.default.greeting.afternoon'), emoji: '🌇', subText: t('dashboard.default.greeting.afternoonSub') };
  } else if (hour >= 18 && hour < 23) {
    return { text: t('dashboard.default.greeting.evening'), emoji: '🌙', subText: t('dashboard.default.greeting.eveningSub') };
  } else {
    return { text: t('dashboard.default.greeting.night'), emoji: '✨', subText: t('dashboard.default.greeting.nightSub') };
  }
};

// ==============================|| 高级统计卡片组件 ||============================== //

const PremiumStatCard = ({
  title,
  value,
  subValue,
  loading,
  icon: Icon,
  gradientColors,
  accentColor,
  isNodeStat,
  copyLink,
  onCopy,
  nodePassStats
}) => {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { t, i18n } = useTranslation();
  const surfaceSx = getCalmSurface(theme, accentColor || gradientColors[0], isDark);
  const hasNodePassStats = Boolean(nodePassStats);

  const handleClick = () => {
    if (isNodeStat && copyLink && onCopy) {
      navigator.clipboard
        .writeText(copyLink)
        .then(() => {
          onCopy(true, 'success');
        })
        .catch(() => {
          onCopy(false, 'error');
        });
    }
  };

  return (
    <Card
      onClick={handleClick}
      sx={{
        ...surfaceSx,
        position: 'relative',
        overflow: 'hidden',
        borderRadius: 4,
        height: '100%',
        cursor: isNodeStat && copyLink ? 'pointer' : 'default',
        '&:hover': {
          ...surfaceSx['&:hover'],
          '& .stat-icon': {
            borderColor: alpha(gradientColors[0], isDark ? 0.36 : 0.24),
            backgroundColor: alpha(gradientColors[0], isDark ? 0.2 : 0.14)
          }
        },
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          height: 3,
          backgroundColor: alpha(gradientColors[0], 0.85)
        }
      }}
    >
      <CardContent sx={{ position: 'relative', zIndex: 1, p: 2.5 }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
              <Box
                sx={{
                  width: 6,
                  height: 6,
                  borderRadius: '50%',
                  bgcolor: gradientColors[0]
                }}
              />
              <Typography
                variant="body2"
                sx={{
                  fontWeight: 500,
                  color: getReadableSecondaryTextColor(theme, isDark),
                  textTransform: 'uppercase',
                  letterSpacing: 1,
                  fontSize: '0.7rem'
                }}
              >
                {title}
              </Typography>
            </Box>

            <Typography
              className="stat-value"
              variant="h1"
              sx={{
                fontWeight: 700,
                fontSize: subValue || hasNodePassStats ? '1.75rem' : '2.25rem',
                color: getReadablePrimaryTextColor(theme, isDark),
                lineHeight: 1.2,
                whiteSpace: 'nowrap'
              }}
            >
              {loading ? (
                <Skeleton width={60} sx={{ bgcolor: alpha(gradientColors[0], 0.2) }} />
              ) : typeof value === 'number' ? (
                formatNumber(value, i18n.resolvedLanguage || i18n.language)
              ) : (
                value
              )}
            </Typography>

            <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mt: 1, minHeight: 20 }}>
              {hasNodePassStats ? (
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: { xs: 0.625, sm: 0.875 },
                    flexWrap: 'nowrap',
                    minWidth: 0,
                    width: '100%'
                  }}
                >
                  {[
                    { key: 'delay', label: t('dashboard.default.stats.delayPass'), value: nodePassStats.delayPassCount },
                    { key: 'speed', label: t('dashboard.default.stats.speedPass'), value: nodePassStats.speedPassCount }
                  ].map((metric, index) => (
                    <Box
                      key={metric.key}
                      sx={{
                        display: 'inline-flex',
                        alignItems: 'baseline',
                        gap: 0.375,
                        minWidth: 0,
                        flexShrink: 1
                      }}
                    >
                      {index > 0 ? (
                        <Typography
                          component="span"
                          variant="caption"
                          sx={{
                            color: alpha(gradientColors[0], isDark ? 0.72 : 0.52),
                            fontWeight: 700,
                            lineHeight: 1,
                            mr: 0.125,
                            flexShrink: 0
                          }}
                        >
                          ·
                        </Typography>
                      ) : null}
                      <Typography
                        variant="caption"
                        sx={{
                          color: getReadableTertiaryTextColor(theme, isDark),
                          fontWeight: 500,
                          fontSize: '0.7rem',
                          whiteSpace: 'nowrap',
                          flexShrink: 0
                        }}
                      >
                        {metric.label}
                      </Typography>
                      <Typography
                        variant="caption"
                        sx={{
                          color: gradientColors[0],
                          fontWeight: 700,
                          fontSize: '0.72rem',
                          whiteSpace: 'nowrap'
                        }}
                      >
                        {loading ? '--' : formatNumber(metric.value, i18n.resolvedLanguage || i18n.language)}
                      </Typography>
                    </Box>
                  ))}
                </Box>
              ) : subValue ? (
                <Tooltip title={subValue} arrow placement="bottom">
                  <Typography
                    variant="caption"
                    sx={{
                      color: getReadableTertiaryTextColor(theme, isDark),
                      fontWeight: 500,
                      fontSize: '0.7rem',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                      maxWidth: '100%',
                      display: 'block'
                    }}
                  >
                    {isNodeStat ? `📍 ${subValue}` : subValue}
                  </Typography>
                </Tooltip>
              ) : (
                <>
                  <TrendingUpIcon sx={{ fontSize: 14, color: theme.palette.success.main }} />
                  <Typography
                    variant="caption"
                    sx={{
                      color: theme.palette.success.main,
                      fontWeight: 600,
                      fontSize: '0.7rem'
                    }}
                  >
                    {t('dashboard.default.stats.running')}
                  </Typography>
                </>
              )}
            </Box>
          </Box>

          <Box
            className="stat-icon"
            sx={{
              width: 56,
              height: 56,
              borderRadius: 2.5,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              backgroundColor: alpha(gradientColors[0], isDark ? 0.16 : 0.1),
              border: `1px solid ${alpha(gradientColors[0], isDark ? 0.26 : 0.18)}`,
              transition: 'background-color 0.2s ease, border-color 0.2s ease',
              flexShrink: 0
            }}
          >
            <Icon
              sx={{
                fontSize: 28,
                color: gradientColors[0]
              }}
            />
          </Box>
        </Box>

        <Box sx={{ mt: 2 }}>
          <LinearProgress
            variant="determinate"
            value={loading ? 0 : 100}
            sx={{
              height: 3,
              borderRadius: 1.5,
              bgcolor: alpha(gradientColors[0], isDark ? 0.16 : 0.1),
              '& .MuiLinearProgress-bar': {
                borderRadius: 1.5,
                backgroundColor: gradientColors[0]
              }
            }}
          />
        </Box>
      </CardContent>
    </Card>
  );
};

// ==============================|| Star 提醒卡片组件 ||============================== //

import { donationConfig, affiliateRecommendationConfig } from 'config/donation';

const StarReminderCard = () => {
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
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            flexWrap: 'wrap',
            gap: 2
          }}
        >
          {/* 左侧内容 */}
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
                sx={{
                  fontWeight: 600,
                  color: supportAccentReadable,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 0.5
                }}
              >
                {t('dashboard.default.star.like')}
                <FavoriteIcon sx={{ fontSize: 16, color: 'error.main' }} />
              </Typography>
              <Typography variant="body2" sx={{ color: getReadableSecondaryTextColor(theme, isDark) }}>
                {t('dashboard.default.star.desc')}
              </Typography>
            </Box>
          </Box>

          {/* 右侧按钮 */}
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
            {/* 打赏按钮组 */}
            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
                gap: 1,
                flexWrap: 'wrap',
                justifyContent: { xs: 'flex-start', sm: 'flex-end' }
              }}
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

            {/* 工具与 Star 组 */}
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
            <Box
              sx={{
                display: 'grid',
                gridTemplateColumns: { xs: '1fr', sm: 'repeat(auto-fit, minmax(280px, 1fr))' },
                gap: 1.5
              }}
            >
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

// ==============================|| 机场流量概览卡片组件 ||============================== //

const AirportUsageCard = ({ airports = [], loading }) => {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { t } = useTranslation();
  const usageAccent = theme.palette.info.main;
  const getProgressTrackColor = (percent) => alpha(getUsageColor(percent), isDark ? 0.22 : 0.12);
  const usageSurface = getInsetPanelSurface(theme, usageAccent, isDark);

  // 筛选开启用量获取且有有效数据的机场
  const airportsWithUsage = useMemo(() => {
    return airports.filter((a) => a.fetchUsageInfo && a.usageTotal > 0);
  }, [airports]);

  // 全局流量汇总
  const { totalUsed, totalQuota, globalPercent } = useMemo(() => {
    const used = airportsWithUsage.reduce((sum, a) => sum + (a.usageUpload || 0) + (a.usageDownload || 0), 0);
    const quota = airportsWithUsage.reduce((sum, a) => sum + a.usageTotal, 0);
    const percent = quota > 0 ? Math.min((used / quota) * 100, 100) : 0;
    return { totalUsed: used, totalQuota: quota, globalPercent: percent };
  }, [airportsWithUsage]);

  // 最近到期机场
  const nearestExpireAirport = useMemo(() => {
    const now = Date.now() / 1000;
    return airportsWithUsage.filter((a) => a.usageExpire > now).sort((a, b) => a.usageExpire - b.usageExpire)[0] || null;
  }, [airportsWithUsage]);

  // 低流量机场 (剩余 < 10%)
  const lowUsageAirports = useMemo(() => {
    return airportsWithUsage.filter((a) => {
      const used = (a.usageUpload || 0) + (a.usageDownload || 0);
      const remaining = a.usageTotal - used;
      return remaining / a.usageTotal < 0.1;
    });
  }, [airportsWithUsage]);

  // 如果没有开启用量获取的机场，不显示此卡片
  if (!loading && airportsWithUsage.length === 0) {
    return null;
  }

  return (
    <Card
      sx={{
        ...getCalmSurface(theme, usageAccent, isDark),
        mb: 4,
        borderRadius: 4,
        overflow: 'hidden',
        position: 'relative',
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          height: 3,
          backgroundColor: usageAccent
        }
      }}
    >
      <CardContent sx={{ p: 3, position: 'relative' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 3 }}>
          <Box sx={getAccentIconBox(usageAccent, isDark)}>
            <FlightTakeoffIcon sx={{ fontSize: 22 }} />
          </Box>
          <Typography variant="h5" sx={{ fontWeight: 600 }}>
            {t('dashboard.default.airports.usageOverview')}
          </Typography>
          <Chip
            label={`${airportsWithUsage.length} ${t('dashboard.default.airports.airportCount')}`}
            size="small"
            sx={{
              ml: 'auto',
              ...getAccentChipSx(theme, usageAccent, isDark)
            }}
          />
        </Box>

        {loading ? (
          <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} variant="rounded" width={200} height={80} sx={{ borderRadius: 2 }} />
            ))}
          </Box>
        ) : (
          <Grid container spacing={3}>
            {/* 全局流量汇总 */}
            <Grid size={{ xs: 12, sm: 6, md: 4 }}>
              <Box
                sx={{
                  p: 2.5,
                  borderRadius: 3,
                  height: '100%',
                  ...usageSurface
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
                  <TrendingUpIcon sx={{ fontSize: 18, color: usageAccent }} />
                  <Typography variant="subtitle2" sx={{ color: getReadablePrimaryTextColor(theme, isDark), fontWeight: 500 }}>
                    {t('dashboard.default.airports.globalUsage')}
                  </Typography>
                </Box>
                <Typography variant="h5" sx={{ fontWeight: 700, mb: 1, color: getReadablePrimaryTextColor(theme, isDark) }}>
                  {formatBytes(totalUsed)} / {formatBytes(totalQuota)}
                </Typography>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <Box
                    sx={{
                      flexGrow: 1,
                      height: 8,
                      borderRadius: 4,
                      backgroundColor: getProgressTrackColor(globalPercent),
                      overflow: 'hidden'
                    }}
                  >
                    <Box
                      sx={{
                        width: `${globalPercent}%`,
                        height: '100%',
                        borderRadius: 4,
                        backgroundColor: getUsageColor(globalPercent),
                        transition: 'width 0.3s ease'
                      }}
                    />
                  </Box>
                  <Typography
                    variant="caption"
                    sx={{
                      fontWeight: 700,
                      color: alpha(getUsageColor(globalPercent), isDark ? 0.9 : 1),
                      minWidth: 45
                    }}
                  >
                    {globalPercent.toFixed(1)}%
                  </Typography>
                </Box>
              </Box>
            </Grid>

            {/* 最近到期 */}
            <Grid size={{ xs: 12, sm: 6, md: 4 }}>
              <Box
                sx={{
                  p: 2.5,
                  borderRadius: 3,
                  height: '100%',
                  ...usageSurface
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
                  <EventIcon sx={{ fontSize: 18, color: getReadableWarningAccentColor(theme, isDark) }} />
                  <Typography variant="subtitle2" sx={{ color: getReadablePrimaryTextColor(theme, isDark), fontWeight: 500 }}>
                    {t('dashboard.default.airports.recentlyExpiring')}
                  </Typography>
                </Box>
                {nearestExpireAirport ? (
                  <>
                    <Typography variant="h6" sx={{ fontWeight: 600, mb: 0.5, color: getReadableWarningAccentColor(theme, isDark) }}>
                      {nearestExpireAirport.name}
                    </Typography>
                    <Typography variant="body2" sx={{ color: getReadableSecondaryTextColor(theme, isDark) }}>
                      {formatExpireTime(nearestExpireAirport.usageExpire)}
                    </Typography>
                  </>
                ) : (
                  <Typography variant="body2" sx={{ color: getReadableSecondaryTextColor(theme, isDark) }}>
                    {t('dashboard.default.airports.noExpireInfo')}
                  </Typography>
                )}
              </Box>
            </Grid>

            {/* 低流量警告 */}
            <Grid size={{ xs: 12, sm: 12, md: 4 }}>
              <Box
                sx={{
                  p: 2.5,
                  borderRadius: 3,
                  height: '100%',
                  bgcolor:
                    lowUsageAirports.length > 0
                      ? isDark
                        ? alpha(theme.palette.error.main, 0.12)
                        : alpha(theme.palette.error.light, 0.12)
                      : isDark
                        ? usageSurface.backgroundColor
                        : usageSurface.backgroundColor,
                  backgroundImage:
                    lowUsageAirports.length > 0
                      ? isDark
                        ? `linear-gradient(180deg, ${alpha(theme.palette.error.main, 0.14)} 0%, ${withAlpha((theme.vars?.palette || theme.palette).background.default, 0.78)} 100%)`
                        : 'none'
                      : usageSurface.backgroundImage,
                  border: `1px solid ${
                    lowUsageAirports.length > 0
                      ? alpha(theme.palette.error.main, 0.28)
                      : isDark
                        ? withAlpha((theme.vars?.palette || theme.palette).divider, 0.74)
                        : alpha(usageAccent, 0.12)
                  }`,
                  boxShadow: isDark && lowUsageAirports.length === 0 ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.025)}` : 'none'
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
                  <WarningAmberIcon
                    sx={{
                      fontSize: 18,
                      color: lowUsageAirports.length > 0 ? 'error.main' : 'text.secondary'
                    }}
                  />
                  <Typography variant="subtitle2" sx={{ color: getReadablePrimaryTextColor(theme, isDark), fontWeight: 500 }}>
                    {t('dashboard.default.airports.lowUsageWarning')}
                  </Typography>
                  {lowUsageAirports.length > 0 && (
                    <Chip
                      label={lowUsageAirports.length}
                      size="small"
                      sx={{
                        ml: 'auto',
                        height: 20,
                        minWidth: 20,
                        bgcolor: 'error.main',
                        color: 'error.contrastText',
                        fontWeight: 700,
                        fontSize: '0.7rem'
                      }}
                    />
                  )}
                </Box>
                {lowUsageAirports.length > 0 ? (
                  <Box sx={{ display: 'flex', gap: 0.75, flexWrap: 'wrap' }}>
                    {lowUsageAirports.map((airport) => {
                      const used = (airport.usageUpload || 0) + (airport.usageDownload || 0);
                      const remaining = airport.usageTotal - used;
                      const remainPercent = ((remaining / airport.usageTotal) * 100).toFixed(1);
                      return (
                        <Tooltip
                          key={airport.id}
                          title={`${t('dashboard.default.airports.remain')} ${formatBytes(remaining)} (${remainPercent}%)`}
                          arrow
                        >
                          <Chip
                            label={airport.name}
                            size="small"
                            sx={{
                              bgcolor: alpha(theme.palette.error.main, isDark ? 0.2 : 0.12),
                              color: 'error.main',
                              fontWeight: 600,
                              fontSize: '0.75rem',
                              '&:hover': {
                                bgcolor: alpha(theme.palette.error.main, isDark ? 0.28 : 0.18)
                              }
                            }}
                          />
                        </Tooltip>
                      );
                    })}
                  </Box>
                ) : (
                  <Typography variant="body2" sx={{ color: isDark ? alpha(theme.palette.success.light, 0.9) : 'success.main' }}>
                    ✓ {t('dashboard.default.airports.allSufficient')}
                  </Typography>
                )}
              </Box>
            </Grid>
          </Grid>
        )}
      </CardContent>
    </Card>
  );
};

// ==============================|| 欢迎横幅组件 ||============================== //

const WelcomeBanner = ({ greeting }) => {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { t } = useTranslation();
  const bannerAccent = theme.palette.secondary.main;

  return (
    <Card
      sx={{
        ...getCalmSurface(theme, bannerAccent, isDark),
        mb: 4,
        position: 'relative',
        overflow: 'hidden',
        borderRadius: 4,
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          bottom: 0,
          width: 4,
          backgroundColor: bannerAccent
        }
      }}
    >
      <CardContent sx={{ position: 'relative', zIndex: 1, py: 5, px: 4 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 3 }}>
          <Box>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 1 }}>
              <Typography
                variant="h1"
                sx={{
                  fontWeight: 800,
                  fontSize: { xs: '2rem', sm: '2.5rem', md: '3rem' },
                  color: getReadablePrimaryTextColor(theme, isDark),
                  lineHeight: 1.2
                }}
              >
                {greeting.text}
              </Typography>
              <Typography
                sx={{
                  fontSize: { xs: '2rem', sm: '2.5rem', md: '3rem' }
                }}
              >
                {greeting.emoji}
              </Typography>
            </Box>
            <Typography
              variant="body1"
              sx={{
                color: getReadableSecondaryTextColor(theme, isDark),
                fontSize: '1.1rem'
              }}
            >
              {t('dashboard.default.greeting.welcome')}{' '}
              <Box component="span" sx={{ fontWeight: 700, color: bannerAccent }}>
                SublinkPro
              </Box>{' '}
              {t('dashboard.default.greeting.system')}
              {greeting.subText}
            </Typography>
          </Box>

          <Box
            sx={{
              display: { xs: 'none', md: 'flex' },
              alignItems: 'center',
              justifyContent: 'center',
              width: 80,
              height: 80,
              borderRadius: 3,
              backgroundColor: alpha(bannerAccent, isDark ? 0.16 : 0.08),
              border: `1px solid ${alpha(bannerAccent, isDark ? 0.3 : 0.18)}`
            }}
          >
            <AutoAwesomeIcon sx={{ fontSize: 40, color: bannerAccent }} />
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

// ==============================|| 发布日志组件 ||============================== //

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
            sx={{
              fontWeight: 700,
              ...getAccentChipSx(theme, theme.palette.primary.main, isDark),
              borderRadius: 2,
              px: 0.5
            }}
          />
          <Typography variant="subtitle1" sx={{ fontWeight: 600, flex: 1, color: getReadablePrimaryTextColor(theme, isDark) }}>
            {release.name}
          </Typography>
          <Chip
            label={formatDateTime(release.published_at, i18n.resolvedLanguage || i18n.language, {
              month: 'short',
              day: 'numeric'
            })}
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

// ==============================|| 仪表盘默认页面 ||============================== //

export default function DashboardDefault() {
  const theme = useTheme();
  const { t, i18n } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const [nodeTotal, setNodeTotal] = useState(0);
  const [nodeDelayPassCount, setNodeDelayPassCount] = useState(0);
  const [nodeSpeedPassCount, setNodeSpeedPassCount] = useState(0);
  const [fastestNode, setFastestNode] = useState(null);
  const [lowestDelayNode, setLowestDelayNode] = useState(null);
  const [countryStats, setCountryStats] = useState([]);
  const [protocolStats, setProtocolStats] = useState({});
  const [tagStats, setTagStats] = useState([]);
  const [groupStats, setGroupStats] = useState({});
  const [sourceStats, setSourceStats] = useState({});
  const [qualityStats, setQualityStats] = useState(null);
  const [releases, setReleases] = useState([]);
  const [airports, setAirports] = useState([]);
  const [loadingStats, setLoadingStats] = useState(true);
  const [loadingReleases, setLoadingReleases] = useState(true);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  const greeting = useMemo(() => getGreeting(t), [t]);

  // 显示提示消息
  const showSnackbar = (success, severity = 'success') => {
    const message = success ? t('dashboard.default.stats.copied') : t('dashboard.default.stats.copyFailed');
    setSnackbar({ open: true, message, severity });
  };

  // 获取统计数据
  const fetchStats = async () => {
    try {
      setLoadingStats(true);
      const [nodeRes, fastestRes, lowestDelayRes, countryRes, groupedStatsRes, qualityRes, airportRes] = await Promise.all([
        getNodeTotal(),
        getFastestSpeedNode(),
        getLowestDelayNode(),
        getDashboardCountryStats(),
        getDashboardGroupedStats(),
        getQualityStats(),
        getAirports()
      ]);
      if (nodeRes.data && typeof nodeRes.data === 'object') {
        setNodeTotal(nodeRes.data.total || 0);
        setNodeDelayPassCount(getDelayPassMetric(nodeRes.data));
        setNodeSpeedPassCount(getSpeedPassMetric(nodeRes.data));
      } else {
        setNodeTotal(nodeRes.data || 0);
        setNodeDelayPassCount(0);
        setNodeSpeedPassCount(0);
      }
      setFastestNode(fastestRes.data || null);
      setLowestDelayNode(lowestDelayRes.data || null);
      setCountryStats(countryRes.data || []);
      setProtocolStats(groupedStatsRes.data?.protocolStats || {});
      setTagStats(groupedStatsRes.data?.tagStats || []);
      setGroupStats(groupedStatsRes.data?.groupStats || {});
      setSourceStats(groupedStatsRes.data?.sourceStats || {});
      setQualityStats(qualityRes.data || null);
      setAirports(airportRes.data?.list || airportRes.data || []);
    } catch (error) {
      console.error('获取统计数据失败:', error);
    } finally {
      setLoadingStats(false);
    }
  };

  // 获取 GitHub 发布日志
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
    fetchStats();
    fetchReleases();
  }, []);

  // 统计卡片配置
  const statsConfig = [
    {
      title: t('dashboard.default.stats.airportCount'),
      value: airports.length,
      subValue: `${airports.filter((airport) => airport.fetchUsageInfo).length} ${t('dashboard.default.stats.airportUsageSub')}`,
      icon: FlightTakeoffIcon,
      gradientColors: ['#6366f1', '#8b5cf6'],
      accentColor: '#6366f1'
    },
    {
      title: t('dashboard.default.stats.nodeTotal'),
      value: nodeTotal,
      subValue: t('dashboard.default.stats.totalNodes'),
      icon: CloudQueueIcon,
      gradientColors: ['#06b6d4', '#0891b2'],
      accentColor: '#06b6d4',
      isNodeStat: true,
      nodePassStats: {
        delayPassCount: nodeDelayPassCount,
        speedPassCount: nodeSpeedPassCount
      }
    },
    {
      title: t('dashboard.default.stats.fastestSpeed'),
      value: fastestNode?.Speed ? `${fastestNode.Speed.toFixed(2)} MB/s` : '--',
      subValue: fastestNode ? getNodeDisplayName(fastestNode) : t('dashboard.default.stats.noData'),
      icon: SpeedIcon,
      gradientColors: ['#10b981', '#059669'],
      accentColor: '#10b981',
      isNodeStat: true,
      copyLink: fastestNode?.Link
    },
    {
      title: t('dashboard.default.stats.lowestDelay'),
      value: lowestDelayNode?.DelayTime ? `${lowestDelayNode.DelayTime} ms` : '--',
      subValue: lowestDelayNode ? getNodeDisplayName(lowestDelayNode) : t('dashboard.default.stats.noData'),
      icon: TimerIcon,
      gradientColors: ['#f59e0b', '#d97706'],
      accentColor: '#f59e0b',
      isNodeStat: true,
      copyLink: lowestDelayNode?.Link
    }
  ];

  const distributionLimit = isMobile ? 4 : 5;
  const countryStatsMap = useMemo(() => createCountryStatMap(countryStats), [countryStats]);
  const unknownCountryStat = countryStatsMap['未知'] || null;
  const countryDistributionSource = useMemo(() => countryStats.filter((item) => item.country !== '未知'), [countryStats]);

  const countryDistribution = useMemo(
    () =>
      normalizeMapStats({
        entries: countryDistributionSource.map((item) => [item.country, item.nodeCount]),
        limit: distributionLimit,
        defaultColor: '#6366f1',
        t,
        getItemMeta: (country) => ({
          marker: getFlagEmoji(country),
          uniqueIpCount: countryStatsMap[country]?.uniqueIpCount || 0,
          tooltip: `${t('dashboard.default.charts.nodes')} ${countryStatsMap[country]?.nodeCount || 0}，${t('dashboard.default.charts.availableIp')} ${countryStatsMap[country]?.uniqueIpCount || 0}`
        })
      }),
    [countryDistributionSource, countryStatsMap, distributionLimit, t]
  );

  const protocolDistribution = useMemo(
    () =>
      normalizeMapStats({
        entries: Object.entries(protocolStats),
        limit: distributionLimit,
        defaultColor: '#10b981',
        t,
        getItemMeta: (protocolName) => ({
          color: protocolColors[protocolName]?.[0] || '#10b981'
        })
      }),
    [protocolStats, distributionLimit, t]
  );

  const tagDistribution = useMemo(
    () => normalizeTagStats({ tags: tagStats, limit: distributionLimit, t }),
    [tagStats, distributionLimit, t]
  );

  const groupDistribution = useMemo(
    () =>
      normalizeMapStats({
        entries: Object.entries(groupStats),
        limit: distributionLimit,
        defaultColor: '#8b5cf6',
        t
      }),
    [groupStats, distributionLimit, t]
  );

  const sourceDistribution = useMemo(
    () =>
      normalizeMapStats({
        entries: Object.entries(sourceStats),
        limit: distributionLimit,
        defaultColor: '#f97316',
        t
      }),
    [sourceStats, distributionLimit, t]
  );

  const qualityStatusDistribution = useMemo(() => {
    const items = (qualityStats?.qualityStatus || [])
      .map((item) => {
        const meta = getQualityStatusMeta(item.key);
        return {
          key: item.key,
          label: t(`dashboard.default.qualityStatus.${item.key}`, meta.label || item.label),
          count: item.count,
          color: qualityStatusColorMap[item.key] || '#64748b',
          tooltip: meta.tooltip
        };
      })
      .filter((item) => item.count > 0)
      .map((item) => ({
        ...item,
        percent: (qualityStats?.total || 0) > 0 ? (item.count / qualityStats.total) * 100 : 0
      }));

    const order = ['success', 'partial', 'failed', 'disabled', 'untested'];
    return order.map((key) => items.find((item) => item.key === key)).filter(Boolean);
  }, [qualityStats, t]);

  const fraudDistribution = useMemo(
    () =>
      (qualityStats?.fraudScoreStats || []).map((item) => ({
        key: item.key,
        label: item.label,
        count: item.count,
        color: fraudLevelColors[item.label.split(' ')[0]] || '#f59e0b',
        percent: (qualityStats?.successTotal || 0) > 0 ? (item.count / qualityStats.successTotal) * 100 : 0
      })),
    [qualityStats]
  );

  const groupedMetricStrip = (item) => [
    { key: 'delay-pass', label: t('dashboard.default.stats.delayPass'), value: item.delayPassCount },
    { key: 'speed-pass', label: t('dashboard.default.stats.speedPass'), value: item.speedPassCount }
  ];

  return (
    <Box sx={{ pb: 3 }}>
      {/* 欢迎横幅 */}
      <WelcomeBanner greeting={greeting} />

      {/* Star 提醒卡片 */}
      <StarReminderCard />

      {/* 任务进度面板 */}
      <TaskProgressPanel />

      {/* 统计卡片 */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        {statsConfig.map((stat, index) => (
          <Grid key={stat.title} size={{ xs: 12, sm: 6, md: 3 }}>
            <PremiumStatCard
              title={stat.title}
              value={stat.value}
              subValue={stat.subValue}
              loading={loadingStats}
              icon={stat.icon}
              gradientColors={stat.gradientColors}
              accentColor={stat.accentColor}
              index={index}
              isNodeStat={stat.isNodeStat}
              copyLink={stat.copyLink}
              onCopy={showSnackbar}
              nodePassStats={stat.nodePassStats}
            />
          </Grid>
        ))}
      </Grid>

      {/* 机场流量概览卡片 */}
      <AirportUsageCard airports={airports} loading={loadingStats} />

      <Grid container spacing={3} sx={{ mb: 4, alignItems: 'stretch' }}>
        <Grid size={{ xs: 12, md: 6 }}>
          <StatsChartCard
            title={t('dashboard.default.charts.countryDistribution')}
            icon={PublicIcon}
            accentColor="#6366f1"
            summary={`${countryDistributionSource.length} ${t('dashboard.default.charts.regions')}`}
            loading={loadingStats}
            tooltip={t('dashboard.default.charts.countryTooltip')}
          >
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.75 }}>
              <RankedStatList
                items={countryDistribution}
                emptyText={t('dashboard.default.charts.emptyCountry')}
                detailFormatter={(item) => `${t('dashboard.default.charts.availableIp')} ${item.uniqueIpCount || 0}`}
                labelFormatter={(item) => {
                  if (item.key === 'collapsed-other') {
                    return t('dashboard.default.charts.otherExpand');
                  }

                  return (
                    <Box component="span" sx={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {item.label}
                    </Box>
                  );
                }}
              />

              {!loadingStats && unknownCountryStat ? (
                <Box
                  sx={{
                    p: 1.5,
                    borderRadius: 3,
                    bgcolor: alpha('#94a3b8', isDark ? 0.12 : 0.08),
                    border: `1px solid ${alpha('#94a3b8', isDark ? 0.24 : 0.16)}`
                  }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1.5 }}>
                    <Box>
                      <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                        {t('dashboard.default.charts.unknownNodes')}
                      </Typography>
                      <Typography variant="caption" sx={{ color: getReadableSecondaryTextColor(theme, isDark) }}>
                        {t('dashboard.default.charts.unknownCountryDesc')}
                      </Typography>
                    </Box>
                    <Box sx={{ textAlign: 'right' }}>
                      <Typography variant="subtitle2" sx={{ fontWeight: 700, color: getReadablePrimaryTextColor(theme, isDark) }}>
                        {formatNumber(unknownCountryStat.nodeCount, i18n.resolvedLanguage || i18n.language)}
                      </Typography>
                      <Typography variant="caption" sx={{ color: getReadableSecondaryTextColor(theme, isDark) }}>
                        {t('dashboard.default.charts.availableIp')} {unknownCountryStat.uniqueIpCount || 0}
                      </Typography>
                    </Box>
                  </Box>
                </Box>
              ) : null}
            </Box>
          </StatsChartCard>
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <StatsChartCard
            title={t('dashboard.default.charts.protocolDistribution')}
            icon={SecurityIcon}
            accentColor="#10b981"
            summary={`${Object.keys(protocolStats).length} ${t('dashboard.default.charts.protocols')}`}
            loading={loadingStats}
            tooltip={t('dashboard.default.charts.protocolTooltip')}
          >
            <RankedStatList
              items={protocolDistribution}
              emptyText={t('dashboard.default.charts.emptyProtocol')}
              labelFormatter={(item) => (item.key === 'collapsed-other' ? t('dashboard.default.charts.otherExpand') : item.label)}
              secondaryMetricsFormatter={groupedMetricStrip}
            />
          </StatsChartCard>
        </Grid>
      </Grid>

      <Grid container spacing={3} sx={{ mb: 4, alignItems: 'stretch' }}>
        <Grid size={{ xs: 12, md: 4 }}>
          <StatsChartCard
            title={t('dashboard.default.charts.tagStats')}
            icon={LabelIcon}
            accentColor="#ec4899"
            summary={`${tagStats.length} ${t('dashboard.default.charts.tags')}`}
            loading={loadingStats}
            tooltip={t('dashboard.default.charts.tagTooltip')}
          >
            <RankedStatList
              items={tagDistribution}
              emptyText={t('dashboard.default.charts.emptyTag')}
              labelFormatter={(item) => (item.key === 'collapsed-other' ? t('dashboard.default.charts.otherExpand') : item.label)}
              secondaryMetricsFormatter={groupedMetricStrip}
            />
          </StatsChartCard>
        </Grid>

        <Grid size={{ xs: 12, md: 4 }}>
          <StatsChartCard
            title={t('dashboard.default.charts.groupStats')}
            icon={FolderIcon}
            accentColor="#8b5cf6"
            summary={`${Object.keys(groupStats).length} ${t('dashboard.default.charts.groups')}`}
            loading={loadingStats}
            tooltip={t('dashboard.default.charts.groupTooltip')}
          >
            <RankedStatList
              items={groupDistribution}
              emptyText={t('dashboard.default.charts.emptyGroup')}
              labelFormatter={(item) => (item.key === 'collapsed-other' ? t('dashboard.default.charts.otherExpand') : item.label)}
              secondaryMetricsFormatter={groupedMetricStrip}
            />
          </StatsChartCard>
        </Grid>

        <Grid size={{ xs: 12, md: 4 }}>
          <StatsChartCard
            title={t('dashboard.default.charts.sourceStats')}
            icon={SourceIcon}
            accentColor="#f97316"
            summary={`${Object.keys(sourceStats).length} ${t('dashboard.default.charts.sources')}`}
            loading={loadingStats}
            tooltip={t('dashboard.default.charts.sourceTooltip')}
          >
            <RankedStatList
              items={sourceDistribution}
              emptyText={t('dashboard.default.charts.emptySource')}
              labelFormatter={(item) => (item.key === 'collapsed-other' ? t('dashboard.default.charts.otherExpand') : item.label)}
              secondaryMetricsFormatter={groupedMetricStrip}
            />
          </StatsChartCard>
        </Grid>
      </Grid>

      <Grid container spacing={3} sx={{ mb: 4, alignItems: 'stretch' }}>
        <Grid size={{ xs: 12, md: 4 }}>
          <StatsChartCard
            title={t('dashboard.default.charts.qualityStats')}
            icon={AutoAwesomeIcon}
            accentColor="#0ea5e9"
            summary={`${qualityStats?.successTotal || 0}/${qualityStats?.total || 0} ${t('dashboard.default.charts.qualityAnalyzable')}`}
            loading={loadingStats}
            tooltip={t('dashboard.default.charts.qualityTooltip')}
          >
            <RankedStatList items={qualityStatusDistribution} emptyText={t('dashboard.default.charts.emptyQuality')} />
          </StatsChartCard>
        </Grid>

        <Grid size={{ xs: 12, md: 4 }}>
          <StatsChartCard
            title={t('dashboard.default.charts.ipQuality')}
            icon={CloudQueueIcon}
            accentColor="#06b6d4"
            summary={`${t('dashboard.default.charts.fullResults')} ${qualityStats?.successTotal || 0}`}
            loading={loadingStats}
            tooltip={t('dashboard.default.charts.ipQualityTooltip')}
          >
            <IPQualityBreakdown stats={qualityStats} loading={loadingStats} />
          </StatsChartCard>
        </Grid>

        <Grid size={{ xs: 12, md: 4 }}>
          <StatsChartCard
            title={t('dashboard.default.charts.fraudDistribution')}
            icon={WarningAmberIcon}
            accentColor="#f59e0b"
            summary={`${t('dashboard.default.charts.fullResults')} ${qualityStats?.successTotal || 0}`}
            loading={loadingStats}
            tooltip={t('dashboard.default.charts.fraudTooltip', { count: qualityStats?.otherTotal || 0 })}
          >
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.75 }}>
              <RankedStatList items={fraudDistribution} emptyText={t('dashboard.default.charts.emptyFraud')} />
              {!loadingStats && (qualityStats?.otherTotal || 0) > 0 ? (
                <Box
                  sx={{
                    p: 1.5,
                    borderRadius: 3,
                    bgcolor: alpha('#94a3b8', isDark ? 0.12 : 0.08),
                    border: `1px solid ${alpha('#94a3b8', isDark ? 0.24 : 0.16)}`
                  }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1.5 }}>
                    <Box>
                      <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                        {t('dashboard.default.charts.noFraudStats')}
                      </Typography>
                      <Typography variant="caption" sx={{ color: getReadableSecondaryTextColor(theme, isDark) }}>
                        {t('dashboard.default.charts.notFullResults')}
                      </Typography>
                    </Box>
                    <Typography variant="subtitle2" sx={{ fontWeight: 700, color: getReadablePrimaryTextColor(theme, isDark) }}>
                      {formatNumber(qualityStats?.otherTotal || 0, i18n.resolvedLanguage || i18n.language)}
                    </Typography>
                  </Box>
                </Box>
              ) : null}
            </Box>
          </StatsChartCard>
        </Grid>
      </Grid>

      {/* 更新日志 */}
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
          <Box
            sx={{
              textAlign: 'center',
              py: 8,
              px: 3
            }}
          >
            <Typography
              sx={{
                fontSize: '3rem',
                mb: 2
              }}
            >
              📭
            </Typography>
            <Typography variant="h6" color="text.secondary" sx={{ fontWeight: 500 }}>
              {t('dashboard.default.releases.noReleases')}
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
              {t('dashboard.default.releases.networkError')}
            </Typography>
          </Box>
        )}
      </MainCard>

      {/* 复制成功提示 */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={2000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert onClose={() => setSnackbar({ ...snackbar, open: false })} severity={snackbar.severity} sx={{ width: '100%' }}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}
