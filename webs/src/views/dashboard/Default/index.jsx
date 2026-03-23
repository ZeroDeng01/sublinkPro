import { useState, useEffect, useMemo } from 'react';
import ReactMarkdown from 'react-markdown';

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

const getCalmSurface = (theme, accentColor) => {
  const isDark = theme.palette.mode === 'dark';

  return {
    backgroundColor: isDark ? alpha(theme.palette.background.paper, 0.92) : theme.palette.background.paper,
    border: `1px solid ${isDark ? alpha(theme.palette.common.white, 0.08) : alpha(accentColor, 0.12)}`,
    boxShadow: isDark ? 'none' : '0 1px 3px rgba(15, 23, 42, 0.06)',
    transition: 'border-color 0.2s ease, box-shadow 0.2s ease',
    '&:hover': {
      borderColor: isDark ? alpha(accentColor, 0.24) : alpha(accentColor, 0.2),
      boxShadow: isDark ? 'none' : '0 4px 12px rgba(15, 23, 42, 0.08)'
    }
  };
};

const getAccentIconBox = (theme, accentColor) => ({
  width: 40,
  height: 40,
  borderRadius: 2,
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  backgroundColor: alpha(accentColor, theme.palette.mode === 'dark' ? 0.18 : 0.12),
  border: `1px solid ${alpha(accentColor, theme.palette.mode === 'dark' ? 0.32 : 0.18)}`,
  color: accentColor,
  flexShrink: 0
});

const getAccentChipSx = (theme, accentColor) => ({
  bgcolor: alpha(accentColor, theme.palette.mode === 'dark' ? 0.18 : 0.1),
  color: theme.palette.mode === 'dark' ? alpha('#fff', 0.92) : accentColor,
  border: `1px solid ${alpha(accentColor, theme.palette.mode === 'dark' ? 0.3 : 0.18)}`,
  fontWeight: 600,
  '&:hover': {
    bgcolor: alpha(accentColor, theme.palette.mode === 'dark' ? 0.24 : 0.14)
  }
});

const COUNTRY_FALLBACK_EMOJI = '🌐';

const getFlagEmoji = (countryCode) => {
  if (!countryCode || countryCode === '未知') return COUNTRY_FALLBACK_EMOJI;
  const normalizedCode = countryCode.toUpperCase() === 'TW' ? 'CN' : countryCode.toUpperCase();
  if (normalizedCode.length !== 2) return COUNTRY_FALLBACK_EMOJI;
  const codePoints = normalizedCode.split('').map((char) => 127397 + char.charCodeAt(0));
  return String.fromCodePoint(...codePoints);
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

const qualityStatusLabelMap = {
  success: '完整结果',
  partial: '信息不全',
  failed: '检测失败',
  disabled: '未启用',
  untested: '未检测'
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

const buildTopItems = (items = [], total = 0, limit = 5, options = {}) => {
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
      label: `其他 ${hiddenItems.length} 项`,
      count: hiddenCount,
      uniqueIpCount: hiddenUniqueIpCount,
      delayPassCount: hiddenDelayPassCount,
      speedPassCount: hiddenSpeedPassCount,
      color: '#94a3b8',
      tooltip: `包含未展示的其余 ${hiddenItems.length} 项，合计 ${hiddenCount} 个节点`,
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

const normalizeMapStats = ({ entries = [], total, limit, defaultColor, getItemMeta, forceCollapsedKeys = [] }) => {
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

  return buildTopItems(normalized, resolvedTotal, limit, { forceCollapsedKeys });
};

const normalizeTagStats = ({ tags = [], limit }) => {
  const total = tags.reduce((sum, item) => sum + getCountMetric(item), 0);
  const normalized = [...tags]
    .sort((a, b) => getCountMetric(b) - getCountMetric(a))
    .map((tag, index) => ({
      key: tag.key || tag.name || tag.label || `tag-${index}`,
      label: getLabelStatValue(tag, tag.name || tag.key || `标签 ${index + 1}`),
      count: getCountMetric(tag),
      delayPassCount: getDelayPassMetric(tag),
      speedPassCount: getSpeedPassMetric(tag),
      color: tag.color || '#ec4899'
    }));

  return buildTopItems(normalized, total, limit);
};

const getProgressBarSx = (theme, color, muted = false) => ({
  height: 7,
  borderRadius: 999,
  bgcolor: alpha(color, theme.palette.mode === 'dark' ? 0.22 : 0.12),
  '& .MuiLinearProgress-bar': {
    borderRadius: 999,
    backgroundColor: muted ? alpha(color, theme.palette.mode === 'dark' ? 0.7 : 0.62) : color
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

  return (
    <Card
      sx={{
        ...getCalmSurface(theme, accentColor),
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
          <Box sx={getAccentIconBox(theme, accentColor)}>
            <Icon sx={{ fontSize: 22 }} />
          </Box>
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Typography variant="h5" sx={{ fontWeight: 600 }}>
              {title}
            </Typography>
            {tooltip ? (
              <Typography variant="body2" sx={{ mt: 0.5, color: 'text.secondary' }}>
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
                bgcolor: alpha(accentColor, theme.palette.mode === 'dark' ? 0.2 : 0.1),
                border: `1px solid ${alpha(accentColor, theme.palette.mode === 'dark' ? 0.32 : 0.18)}`
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
  const [expandedKeys, setExpandedKeys] = useState({});

  const formatSecondaryMetricValue = (value) => {
    if (typeof value === 'number') {
      return value.toLocaleString();
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
                  color: alpha(itemColor, theme.palette.mode === 'dark' ? 0.7 : 0.5),
                  fontWeight: 700,
                  lineHeight: 1
                }}
              >
                ·
              </Typography>
            ) : null}
            <Typography variant="caption" sx={{ color: 'text.secondary', lineHeight: 1.2 }}>
              {metric.label}
            </Typography>
            <Typography
              variant="caption"
              sx={{
                fontWeight: 700,
                color: theme.palette.mode === 'dark' ? alpha('#fff', 0.88) : alpha(itemColor, 0.88),
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
    return (
      <Typography color="text.secondary" sx={{ fontSize: '0.875rem' }}>
        {emptyText}
      </Typography>
    );
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
                      boxShadow: `0 0 0 4px ${alpha(item.color, theme.palette.mode === 'dark' ? 0.18 : 0.12)}`
                    }}
                  />
                )}
                <Box sx={{ minWidth: 0, flex: 1 }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75, minWidth: 0 }}>
                    <Typography
                      variant="subtitle2"
                      sx={{
                        fontWeight: 600,
                        minWidth: 0,
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap'
                      }}
                    >
                      {labelFormatter ? labelFormatter(item, isExpanded) : item.label}
                    </Typography>
                    {item.isCollapsedOther ? (
                      <Typography variant="caption" sx={{ color: 'text.secondary', flexShrink: 0 }}>
                        {isExpanded ? '收起' : '展开'}
                      </Typography>
                    ) : null}
                  </Box>
                  {renderSecondaryMetrics(secondaryMetrics, item.color, item.key)}
                </Box>
              </Box>
              <Box sx={{ display: 'flex', alignItems: 'baseline', gap: 0.75, flexShrink: 0 }}>
                <Box sx={{ textAlign: 'right' }}>
                  <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                    {valueFormatter ? valueFormatter(item.count, item) : item.count.toLocaleString()}
                  </Typography>
                  {detailFormatter ? (
                    <Typography variant="caption" sx={{ color: 'text.secondary', display: 'block' }}>
                      {detailFormatter(item)}
                    </Typography>
                  ) : null}
                </Box>
                <Typography variant="caption" sx={{ color: 'text.secondary', minWidth: 42, textAlign: 'right' }}>
                  {item.percent.toFixed(1)}{percentSuffix}
                </Typography>
              </Box>
            </Box>
            <LinearProgress variant="determinate" value={Math.max(0, Math.min(item.percent, 100))} sx={getProgressBarSx(theme, item.color, muted)} />
            {item.isCollapsedOther && isExpanded && item.hiddenItems?.length ? (
              <Box
                sx={{
                  mt: 1.25,
                  ml: { xs: 1.5, sm: 2 },
                  pl: 1.5,
                  borderLeft: `2px dashed ${alpha(item.color, theme.palette.mode === 'dark' ? 0.4 : 0.3)}`,
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
                              color: 'text.secondary'
                            }}
                          >
                            {hiddenItem.marker ? <Typography sx={{ fontSize: '1rem', lineHeight: 1 }}>{hiddenItem.marker}</Typography> : null}
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
                      <Typography variant="caption" sx={{ color: 'text.secondary', flexShrink: 0 }}>
                        {valueFormatter ? valueFormatter(hiddenItem.count, hiddenItem) : hiddenItem.count.toLocaleString()}
                        {detailFormatter ? ` · ${detailFormatter(hiddenItem)}` : ''} · {hiddenItem.percent.toFixed(1)}{percentSuffix}
                      </Typography>
                    </Box>
                    <LinearProgress
                      variant="determinate"
                      value={Math.max(0, Math.min(hiddenItem.percent, 100))}
                      sx={getProgressBarSx(theme, hiddenItem.color || item.color, true)}
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
              boxShadow: `0 0 0 4px ${alpha(color, theme.palette.mode === 'dark' ? 0.18 : 0.12)}`
            }}
          />
          <Typography variant="subtitle2" sx={{ fontWeight: 600, minWidth: 0 }}>
            {label}
          </Typography>
        </Box>
        <Box sx={{ display: 'flex', alignItems: 'baseline', gap: 0.75, flexShrink: 0 }}>
          <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
            {count.toLocaleString()}
          </Typography>
          <Typography variant="caption" sx={{ color: 'text.secondary', minWidth: 42, textAlign: 'right' }}>
            {percent.toFixed(1)}%
          </Typography>
        </Box>
      </Box>
      <LinearProgress variant="determinate" value={Math.max(0, Math.min(percent, 100))} sx={getProgressBarSx(theme, color)} />
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

  if (loading) {
    return <StatRowsSkeleton rows={5} />;
  }

  if (!stats || !Array.isArray(stats.ipStats) || stats.ipStats.length === 0) {
    return (
      <Typography color="text.secondary" sx={{ fontSize: '0.875rem' }}>
        暂无 IP 质量统计数据
      </Typography>
    );
  }

  const total = stats.total || 0;
  const successTotal = stats.successTotal || 0;
  const findCount = (key) => stats.ipStats.find((item) => item.key === key)?.count || 0;

  const residentialRows = [
    { key: 'housing', label: '住宅IP', count: findCount('housing'), color: '#22c55e' },
    { key: 'datacenter', label: '机房IP', count: findCount('datacenter'), color: '#64748b' }
  ];
  const typeRows = [
    { key: 'native', label: '原生IP', count: findCount('native'), color: '#06b6d4' },
    { key: 'broadcast', label: '广播IP', count: findCount('broadcast'), color: '#f59e0b' }
  ];
  const otherCount = findCount('other');

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2.25 }}>
      <Box>
        <Typography variant="body2" sx={{ color: 'text.secondary', mb: 1.25, fontWeight: 600 }}>
          住宅属性
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
          borderTop: `1px solid ${alpha(theme.palette.text.primary, theme.palette.mode === 'dark' ? 0.08 : 0.06)}`
        }}
      >
        <Typography variant="body2" sx={{ color: 'text.secondary', mb: 1.25, fontWeight: 600 }}>
          IP 类型
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
          bgcolor: alpha('#94a3b8', theme.palette.mode === 'dark' ? 0.12 : 0.08),
          border: `1px solid ${alpha('#94a3b8', theme.palette.mode === 'dark' ? 0.24 : 0.16)}`
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1.5 }}>
          <Box>
            <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
              其他数量
            </Typography>
            <Typography variant="caption" sx={{ color: 'text.secondary' }}>
              非完整结果，未纳入细分 IP 判断
            </Typography>
          </Box>
          <Box sx={{ textAlign: 'right' }}>
            <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
              {otherCount.toLocaleString()}
            </Typography>
            <Typography variant="caption" sx={{ color: 'text.secondary' }}>
              {total > 0 ? ((otherCount / total) * 100).toFixed(1) : '0.0'}%
            </Typography>
          </Box>
        </Box>
      </Box>
    </Box>
  );
};

// ==============================|| 问候语计算 ||============================== //

const getGreeting = () => {
  const hour = new Date().getHours();
  if (hour >= 5 && hour < 9) {
    return { text: '早上好', emoji: '🌅', subText: '新的一天开始了' };
  } else if (hour >= 9 && hour < 12) {
    return { text: '上午好', emoji: '☀️', subText: '充满活力的上午' };
  } else if (hour >= 12 && hour < 14) {
    return { text: '中午好', emoji: '🌤️', subText: '记得休息一下' };
  } else if (hour >= 14 && hour < 18) {
    return { text: '下午好', emoji: '🌇', subText: '继续加油' };
  } else if (hour >= 18 && hour < 23) {
    return { text: '晚上好', emoji: '🌙', subText: '辛苦了一天' };
  } else {
    return { text: '夜深了', emoji: '✨', subText: '注意休息' };
  }
};

// ==============================|| 高级统计卡片组件 ||============================== //

const PremiumStatCard = ({ title, value, subValue, loading, icon: Icon, gradientColors, accentColor, isNodeStat, copyLink, onCopy, nodePassStats }) => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const surfaceSx = getCalmSurface(theme, accentColor || gradientColors[0]);
  const hasNodePassStats = Boolean(nodePassStats);

  const handleClick = () => {
    if (isNodeStat && copyLink && onCopy) {
      navigator.clipboard
        .writeText(copyLink)
        .then(() => {
          onCopy('节点链接已复制到剪贴板', 'success');
        })
        .catch(() => {
          onCopy('复制失败', 'error');
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
                  color: isDark ? alpha('#fff', 0.7) : theme.palette.text.secondary,
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
                color: theme.palette.text.primary,
                lineHeight: 1.2,
                whiteSpace: 'nowrap'
              }}
            >
              {loading ? (
                <Skeleton width={60} sx={{ bgcolor: alpha(gradientColors[0], 0.2) }} />
              ) : typeof value === 'number' ? (
                value.toLocaleString()
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
                    { key: 'delay', label: '延迟通过', value: nodePassStats.delayPassCount },
                    { key: 'speed', label: '速度通过', value: nodePassStats.speedPassCount }
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
                          color: isDark ? alpha('#fff', 0.6) : theme.palette.text.secondary,
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
                        {loading ? '--' : metric.value.toLocaleString()}
                      </Typography>
                    </Box>
                  ))}
                </Box>
              ) : subValue ? (
                <Tooltip title={subValue} arrow placement="bottom">
                  <Typography
                    variant="caption"
                    sx={{
                      color: isDark ? alpha('#fff', 0.6) : theme.palette.text.secondary,
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
                    运行中
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
              bgcolor: alpha(gradientColors[0], 0.1),
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

import { donationConfig } from 'config/donation';

const StarReminderCard = () => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const [starCount, setStarCount] = useState(null);

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
        ...getCalmSurface(theme, '#f59e0b'),
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
          backgroundColor: '#f59e0b'
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
                backgroundColor: alpha('#f59e0b', isDark ? 0.18 : 0.12),
                border: `1px solid ${alpha('#f59e0b', isDark ? 0.32 : 0.2)}`,
                flexShrink: 0
              }}
            >
              <StarIcon sx={{ fontSize: 28, color: '#f59e0b' }} />
            </Box>
            <Box>
              <Typography
                variant="subtitle1"
                sx={{
                  fontWeight: 600,
                  color: isDark ? '#fcd34d' : '#b45309',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 0.5
                }}
              >
                喜欢这个项目吗？
                <FavoriteIcon sx={{ fontSize: 16, color: '#ef4444' }} />
              </Typography>
              <Typography variant="body2" sx={{ color: isDark ? alpha('#fff', 0.7) : '#92400e' }}>
                如果觉得不错，请给我们一个 Star 支持一下！如果你是L站佬友，也可以使用LDC支持本项目！你的支持是我们前进的动力。
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
                  key={index}
                  icon={item.icon}
                  label={item.title}
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
              <Tooltip title="问题反馈" arrow>
                <IconButton
                  onClick={handleFeedback}
                  size="small"
                  sx={{
                    bgcolor: isDark ? alpha('#fff', 0.1) : alpha('#f59e0b', 0.15),
                    color: isDark ? '#fcd34d' : '#b45309',
                    width: 36,
                    height: 36,
                    borderRadius: 2,
                    '&:hover': {
                      bgcolor: isDark ? alpha('#fff', 0.15) : alpha('#f59e0b', 0.25)
                    }
                  }}
                >
                  <BugReportIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              <Chip
                icon={<GitHubIcon sx={{ fontSize: 18, color: 'inherit !important' }} />}
                label={starCount !== null ? `Star ${starCount >= 1000 ? `${(starCount / 1000).toFixed(1)}k` : starCount}` : 'Star'}
                onClick={handleStar}
                sx={{
                  fontWeight: 600,
                  px: 1,
                  height: 36,
                  borderRadius: 2,
                  ...getAccentChipSx(theme, '#f59e0b'),
                  cursor: 'pointer',
                  '& .MuiChip-icon': {
                    color: 'inherit'
                  }
                }}
              />
            </Box>
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

// ==============================|| 机场流量概览卡片组件 ||============================== //

const AirportUsageCard = ({ airports = [], loading }) => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

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

  // 根据使用率计算进度条渐变色
  const getProgressGradient = (percent) => {
    if (percent < 60) return `linear-gradient(90deg, ${theme.palette.success.light}, ${theme.palette.success.main})`;
    if (percent < 85) return `linear-gradient(90deg, ${theme.palette.warning.light}, ${theme.palette.warning.main})`;
    return `linear-gradient(90deg, ${theme.palette.error.light}, ${theme.palette.error.main})`;
  };

  return (
    <Card
      sx={{
        ...getCalmSurface(theme, '#06b6d4'),
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
          backgroundColor: '#06b6d4'
        }
      }}
    >
      <CardContent sx={{ p: 3, position: 'relative' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 3 }}>
          <Box sx={getAccentIconBox(theme, '#06b6d4')}>
            <FlightTakeoffIcon sx={{ fontSize: 22 }} />
          </Box>
          <Typography variant="h5" sx={{ fontWeight: 600 }}>
            机场流量概览
          </Typography>
          <Chip
            label={`${airportsWithUsage.length} 个机场`}
            size="small"
            sx={{
              ml: 'auto',
              ...getAccentChipSx(theme, '#06b6d4')
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
                  bgcolor: isDark ? alpha(theme.palette.common.white, 0.03) : alpha(theme.palette.common.white, 0.88),
                  border: `1px solid ${isDark ? alpha(theme.palette.common.white, 0.08) : alpha('#06b6d4', 0.12)}`
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
                  <TrendingUpIcon sx={{ fontSize: 18, color: '#06b6d4' }} />
                  <Typography variant="subtitle2" sx={{ color: 'text.secondary', fontWeight: 500 }}>
                    全局流量使用
                  </Typography>
                </Box>
                <Typography variant="h5" sx={{ fontWeight: 700, mb: 1 }}>
                  {formatBytes(totalUsed)} / {formatBytes(totalQuota)}
                </Typography>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <Box
                    sx={{
                      flexGrow: 1,
                      height: 8,
                      borderRadius: 4,
                      backgroundColor: isDark ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.08)',
                      overflow: 'hidden'
                    }}
                  >
                    <Box
                      sx={{
                        width: `${globalPercent}%`,
                        height: '100%',
                        borderRadius: 4,
                        background: getProgressGradient(globalPercent),
                        transition: 'width 0.3s ease'
                      }}
                    />
                  </Box>
                  <Typography variant="caption" sx={{ fontWeight: 700, color: getUsageColor(globalPercent), minWidth: 45 }}>
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
                  bgcolor: isDark ? alpha(theme.palette.common.white, 0.03) : alpha(theme.palette.common.white, 0.88),
                  border: `1px solid ${isDark ? alpha(theme.palette.common.white, 0.08) : alpha('#06b6d4', 0.12)}`
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
                  <EventIcon sx={{ fontSize: 18, color: '#f59e0b' }} />
                  <Typography variant="subtitle2" sx={{ color: 'text.secondary', fontWeight: 500 }}>
                    最近到期
                  </Typography>
                </Box>
                {nearestExpireAirport ? (
                  <>
                    <Typography variant="h6" sx={{ fontWeight: 600, mb: 0.5, color: isDark ? '#fcd34d' : '#b45309' }}>
                      {nearestExpireAirport.name}
                    </Typography>
                    <Typography variant="body2" sx={{ color: 'text.secondary' }}>
                      {formatExpireTime(nearestExpireAirport.usageExpire)}
                    </Typography>
                  </>
                ) : (
                  <Typography variant="body2" sx={{ color: 'text.secondary' }}>
                    暂无到期信息
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
                        ? alpha('#ef4444', 0.1)
                        : alpha('#fef2f2', 0.92)
                      : isDark
                        ? alpha(theme.palette.common.white, 0.03)
                        : alpha(theme.palette.common.white, 0.88),
                  border: `1px solid ${
                    lowUsageAirports.length > 0 ? alpha('#ef4444', 0.3) : isDark ? alpha('#fff', 0.1) : alpha('#06b6d4', 0.15)
                  }`
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
                  <WarningAmberIcon
                    sx={{
                      fontSize: 18,
                      color: lowUsageAirports.length > 0 ? '#ef4444' : 'text.secondary'
                    }}
                  />
                  <Typography variant="subtitle2" sx={{ color: 'text.secondary', fontWeight: 500 }}>
                    流量不足警告
                  </Typography>
                  {lowUsageAirports.length > 0 && (
                    <Chip
                      label={lowUsageAirports.length}
                      size="small"
                      sx={{
                        ml: 'auto',
                        height: 20,
                        minWidth: 20,
                        bgcolor: '#ef4444',
                        color: '#fff',
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
                        <Tooltip key={airport.id} title={`剩余 ${formatBytes(remaining)} (${remainPercent}%)`} arrow>
                          <Chip
                            label={airport.name}
                            size="small"
                            sx={{
                              bgcolor: isDark ? alpha('#ef4444', 0.2) : alpha('#ef4444', 0.1),
                              color: '#ef4444',
                              fontWeight: 600,
                              fontSize: '0.75rem',
                              '&:hover': {
                                bgcolor: isDark ? alpha('#ef4444', 0.3) : alpha('#ef4444', 0.2)
                              }
                            }}
                          />
                        </Tooltip>
                      );
                    })}
                  </Box>
                ) : (
                  <Typography variant="body2" sx={{ color: isDark ? '#86efac' : '#16a34a' }}>
                    ✓ 所有机场流量充足
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
  const isDark = theme.palette.mode === 'dark';

  return (
    <Card
      sx={{
        mb: 4,
        position: 'relative',
        overflow: 'hidden',
        borderRadius: 4,
        backgroundColor: isDark ? alpha(theme.palette.background.paper, 0.96) : theme.palette.background.paper,
        border: `1px solid ${isDark ? alpha(theme.palette.common.white, 0.08) : alpha('#6366f1', 0.1)}`,
        boxShadow: isDark ? 'none' : '0 1px 3px rgba(15, 23, 42, 0.06)',
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          bottom: 0,
          width: 4,
          backgroundColor: '#6366f1'
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
                  color: theme.palette.text.primary,
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
                color: isDark ? alpha('#fff', 0.7) : theme.palette.text.secondary,
                fontSize: '1.1rem'
              }}
            >
              欢迎使用{' '}
              <Box component="span" sx={{ fontWeight: 700, color: isDark ? '#a5b4fc' : '#6366f1' }}>
                SublinkPro
              </Box>{' '}
              订阅管理系统，{greeting.subText}
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
              backgroundColor: alpha('#6366f1', isDark ? 0.14 : 0.08),
              border: `1px solid ${alpha('#6366f1', isDark ? 0.28 : 0.16)}`
            }}
          >
            <AutoAwesomeIcon sx={{ fontSize: 40, color: isDark ? '#a5b4fc' : '#6366f1' }} />
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

// ==============================|| 发布日志组件 ||============================== //

const ReleaseCard = ({ release }) => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  return (
    <Card
      sx={{
        mb: 2.5,
        borderRadius: 3,
        backgroundColor: isDark ? alpha(theme.palette.background.paper, 0.94) : theme.palette.background.paper,
        border: `1px solid ${isDark ? alpha(theme.palette.common.white, 0.08) : alpha(theme.palette.primary.main, 0.08)}`,
        transition: 'border-color 0.2s ease, box-shadow 0.2s ease',
        '&:hover': {
          boxShadow: isDark ? 'none' : '0 4px 12px rgba(15, 23, 42, 0.08)',
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
              ...getAccentChipSx(theme, theme.palette.primary.main),
              borderRadius: 2,
              px: 0.5
            }}
          />
          <Typography variant="subtitle1" sx={{ fontWeight: 600, flex: 1 }}>
            {release.name}
          </Typography>
          <Chip
            label={new Date(release.published_at).toLocaleDateString('zh-CN', {
              month: 'short',
              day: 'numeric'
            })}
            size="small"
            variant="outlined"
            sx={{ borderRadius: 2 }}
          />
          <Tooltip title="在 GitHub 查看" arrow>
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
              color: theme.palette.text.primary
            },
            '& p': {
              mb: 1,
              fontSize: '0.875rem',
              lineHeight: 1.7,
              color: theme.palette.text.secondary
            },
            '& ul, & ol': {
              pl: 2.5,
              mb: 1
            },
            '& li': {
              fontSize: '0.875rem',
              mb: 0.5,
              color: theme.palette.text.secondary,
              '&::marker': {
                color: theme.palette.primary.main
              }
            },
            '& code': {
              backgroundColor: isDark ? alpha('#fff', 0.1) : alpha('#6366f1', 0.1),
              color: isDark ? '#a5b4fc' : '#6366f1',
              padding: '2px 8px',
              borderRadius: 6,
              fontSize: '0.8rem',
              fontFamily: '"JetBrains Mono", monospace'
            },
            '& pre': {
              backgroundColor: isDark ? alpha('#000', 0.3) : alpha('#f1f5f9', 0.8),
              padding: 2,
              borderRadius: 2,
              overflow: 'auto',
              border: `1px solid ${isDark ? alpha('#fff', 0.1) : alpha('#000', 0.05)}`,
              '& code': {
                backgroundColor: 'transparent',
                padding: 0
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
          <ReactMarkdown>{release.body || '暂无更新说明'}</ReactMarkdown>
        </Box>
      </CardContent>
    </Card>
  );
};

// ==============================|| 仪表盘默认页面 ||============================== //

export default function DashboardDefault() {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
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

  const greeting = useMemo(() => getGreeting(), []);

  // 显示提示消息
  const showSnackbar = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  // 获取统计数据
  const fetchStats = async () => {
    try {
      setLoadingStats(true);
      const [nodeRes, fastestRes, lowestDelayRes, countryRes, groupedStatsRes, qualityRes, airportRes] =
        await Promise.all([
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
      title: '机场总数',
      value: airports.length,
      subValue: `${airports.filter((airport) => airport.fetchUsageInfo).length} 个已启用用量获取`,
      icon: FlightTakeoffIcon,
      gradientColors: ['#6366f1', '#8b5cf6'],
      accentColor: '#6366f1'
    },
    {
      title: '节点统计',
      value: nodeTotal,
      subValue: '总节点',
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
      title: '最快速度',
      value: fastestNode?.Speed ? `${fastestNode.Speed.toFixed(2)} MB/s` : '--',
      subValue: fastestNode?.Name || '暂无数据',
      icon: SpeedIcon,
      gradientColors: ['#10b981', '#059669'],
      accentColor: '#10b981',
      isNodeStat: true,
      copyLink: fastestNode?.Link
    },
    {
      title: '最低延迟',
      value: lowestDelayNode?.DelayTime ? `${lowestDelayNode.DelayTime} ms` : '--',
      subValue: lowestDelayNode?.Name || '暂无数据',
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
        getItemMeta: (country) => ({
          marker: getFlagEmoji(country),
          uniqueIpCount: countryStatsMap[country]?.uniqueIpCount || 0,
          tooltip: `节点 ${countryStatsMap[country]?.nodeCount || 0}，可用IP ${countryStatsMap[country]?.uniqueIpCount || 0}`
        })
      }),
    [countryDistributionSource, countryStatsMap, distributionLimit]
  );

  const protocolDistribution = useMemo(
    () =>
      normalizeMapStats({
        entries: Object.entries(protocolStats),
        limit: distributionLimit,
        defaultColor: '#10b981',
        getItemMeta: (protocolName) => ({
          color: protocolColors[protocolName]?.[0] || '#10b981'
        })
      }),
    [protocolStats, distributionLimit]
  );

  const tagDistribution = useMemo(() => normalizeTagStats({ tags: tagStats, limit: distributionLimit }), [tagStats, distributionLimit]);

  const groupDistribution = useMemo(
    () =>
      normalizeMapStats({
        entries: Object.entries(groupStats),
        limit: distributionLimit,
        defaultColor: '#8b5cf6'
      }),
    [groupStats, distributionLimit]
  );

  const sourceDistribution = useMemo(
    () =>
      normalizeMapStats({
        entries: Object.entries(sourceStats),
        limit: distributionLimit,
        defaultColor: '#f97316'
      }),
    [sourceStats, distributionLimit]
  );

  const qualityStatusDistribution = useMemo(() => {
    const items = (qualityStats?.qualityStatus || [])
      .map((item) => {
        const meta = getQualityStatusMeta(item.key);
        return {
          key: item.key,
          label: qualityStatusLabelMap[item.key] || meta.label || item.label,
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
  }, [qualityStats]);

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
    { key: 'delay-pass', label: '延迟通过', value: item.delayPassCount },
    { key: 'speed-pass', label: '速度通过', value: item.speedPassCount }
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
            title="节点国家分布"
            icon={PublicIcon}
            accentColor="#6366f1"
            summary={`${countryDistributionSource.length} 个地区`}
            loading={loadingStats}
            tooltip="按节点落地国家聚合，仅比较真实地区分布；未知节点数量单独展示。"
          >
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.75 }}>
              <RankedStatList
                items={countryDistribution}
                emptyText="暂无国家统计数据"
                detailFormatter={(item) => `可用IP ${item.uniqueIpCount || 0}`}
                labelFormatter={(item) => {
                  if (item.key === 'collapsed-other') {
                    return '其他地区（可展开查看）';
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
                    bgcolor: alpha('#94a3b8', theme.palette.mode === 'dark' ? 0.12 : 0.08),
                    border: `1px solid ${alpha('#94a3b8', theme.palette.mode === 'dark' ? 0.24 : 0.16)}`
                  }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1.5 }}>
                    <Box>
                      <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                        未知节点
                      </Typography>
                      <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                        未获取到落地国家或节点已失效，不参与地区比较
                      </Typography>
                    </Box>
                    <Box sx={{ textAlign: 'right' }}>
                      <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                        {unknownCountryStat.nodeCount.toLocaleString()}
                      </Typography>
                      <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                        可用IP {unknownCountryStat.uniqueIpCount || 0}
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
            title="节点协议分布"
            icon={SecurityIcon}
            accentColor="#10b981"
            summary={`${Object.keys(protocolStats).length} 种协议`}
            loading={loadingStats}
            tooltip="按协议类型统计，便于快速判断节点结构。"
          >
            <RankedStatList
              items={protocolDistribution}
              emptyText="暂无协议统计数据"
              labelFormatter={(item) => (item.key === 'collapsed-other' ? '其他（可展开查看）' : item.label)}
              secondaryMetricsFormatter={groupedMetricStrip}
            />
          </StatsChartCard>
        </Grid>
      </Grid>

      <Grid container spacing={3} sx={{ mb: 4, alignItems: 'stretch' }}>
        <Grid size={{ xs: 12, md: 4 }}>
          <StatsChartCard
            title="标签统计"
            icon={LabelIcon}
            accentColor="#ec4899"
            summary={`${tagStats.length} 个标签`}
            loading={loadingStats}
            tooltip="展示命中最多的标签，便于识别规则覆盖情况。"
          >
            <RankedStatList
              items={tagDistribution}
              emptyText="暂无标签统计数据"
              labelFormatter={(item) => (item.key === 'collapsed-other' ? '其他（可展开查看）' : item.label)}
              secondaryMetricsFormatter={groupedMetricStrip}
            />
          </StatsChartCard>
        </Grid>

        <Grid size={{ xs: 12, md: 4 }}>
          <StatsChartCard
            title="分组统计"
            icon={FolderIcon}
            accentColor="#8b5cf6"
            summary={`${Object.keys(groupStats).length} 个分组`}
            loading={loadingStats}
            tooltip="按节点分组聚合，方便查看主要组织结构。"
          >
            <RankedStatList
              items={groupDistribution}
              emptyText="暂无分组统计数据"
              labelFormatter={(item) => (item.key === 'collapsed-other' ? '其他（可展开查看）' : item.label)}
              secondaryMetricsFormatter={groupedMetricStrip}
            />
          </StatsChartCard>
        </Grid>

        <Grid size={{ xs: 12, md: 4 }}>
          <StatsChartCard
            title="来源统计"
            icon={SourceIcon}
            accentColor="#f97316"
            summary={`${Object.keys(sourceStats).length} 个来源`}
            loading={loadingStats}
            tooltip="展示节点主要来源，便于识别上游贡献占比。"
          >
            <RankedStatList
              items={sourceDistribution}
              emptyText="暂无来源统计数据"
              labelFormatter={(item) => (item.key === 'collapsed-other' ? '其他（可展开查看）' : item.label)}
              secondaryMetricsFormatter={groupedMetricStrip}
            />
          </StatsChartCard>
        </Grid>
      </Grid>

      <Grid container spacing={3} sx={{ mb: 4, alignItems: 'stretch' }}>
        <Grid size={{ xs: 12, md: 4 }}>
          <StatsChartCard
            title="质量状态统计"
            icon={AutoAwesomeIcon}
            accentColor="#0ea5e9"
            summary={`${qualityStats?.successTotal || 0}/${qualityStats?.total || 0} 可细分`}
            loading={loadingStats}
            tooltip="完整结果可参与 IP 和欺诈评分细分，其余状态用于说明覆盖率。"
          >
            <RankedStatList items={qualityStatusDistribution} emptyText="暂无质量状态统计数据" />
          </StatsChartCard>
        </Grid>

        <Grid size={{ xs: 12, md: 4 }}>
          <StatsChartCard
            title="IP 质量统计"
            icon={CloudQueueIcon}
            accentColor="#06b6d4"
            summary={`完整结果 ${qualityStats?.successTotal || 0}`}
            loading={loadingStats}
            tooltip="住宅/机房与原生/广播基于完整质量检测结果统计，其他数量表示未完成细分的节点。"
          >
            <IPQualityBreakdown stats={qualityStats} loading={loadingStats} />
          </StatsChartCard>
        </Grid>

        <Grid size={{ xs: 12, md: 4 }}>
          <StatsChartCard
            title="欺诈评分分布"
            icon={WarningAmberIcon}
            accentColor="#f59e0b"
            summary={`完整结果 ${qualityStats?.successTotal || 0}`}
            loading={loadingStats}
            tooltip={`按系统现有欺诈评分分级方式统计，仅完整结果节点参与分布；另有 ${qualityStats?.otherTotal || 0} 个节点因质量状态不是完整结果而未参与统计。`}
          >
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.75 }}>
              <RankedStatList items={fraudDistribution} emptyText="暂无欺诈评分统计数据" />
              {!loadingStats && (qualityStats?.otherTotal || 0) > 0 ? (
                <Box
                  sx={{
                    p: 1.5,
                    borderRadius: 3,
                    bgcolor: alpha('#94a3b8', theme.palette.mode === 'dark' ? 0.12 : 0.08),
                    border: `1px solid ${alpha('#94a3b8', theme.palette.mode === 'dark' ? 0.24 : 0.16)}`
                  }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1.5 }}>
                    <Box>
                      <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                        未参与评分统计
                      </Typography>
                      <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                        质量状态不是完整结果
                      </Typography>
                    </Box>
                    <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                      {(qualityStats?.otherTotal || 0).toLocaleString()}
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
                background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)'
              }}
            >
              <Typography sx={{ fontSize: '1.2rem' }}>📝</Typography>
            </Box>
            <Typography variant="h4" sx={{ fontWeight: 600 }}>
              更新日志
            </Typography>
          </Box>
        }
        secondary={
          <Tooltip title="刷新" arrow>
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
          borderRadius: 4,
          overflow: 'hidden',
          '& .MuiCardHeader-root': {
            borderBottom: `1px solid ${isDark ? alpha('#fff', 0.08) : alpha('#000', 0.06)}`
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
                    bgcolor: isDark ? alpha('#fff', 0.05) : alpha('#000', 0.04)
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
            <Typography variant="h6" color="textSecondary" sx={{ fontWeight: 500 }}>
              暂无更新日志
            </Typography>
            <Typography variant="body2" color="textSecondary" sx={{ mt: 1 }}>
              请检查网络连接或稍后重试
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
