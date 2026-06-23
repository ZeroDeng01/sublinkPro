import PropTypes from 'prop-types';
import { useState, useMemo, useCallback, useRef, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Collapse from '@mui/material/Collapse';
import Dialog from '@mui/material/Dialog';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import InputAdornment from '@mui/material/InputAdornment';
import Skeleton from '@mui/material/Skeleton';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Tooltip from '@mui/material/Tooltip';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import { getNodeDisplayName } from 'utils/nodeDisplayName';

// icons
import CloseIcon from '@mui/icons-material/Close';
import SearchIcon from '@mui/icons-material/Search';
import FilterListIcon from '@mui/icons-material/FilterList';
import ArrowForwardIcon from '@mui/icons-material/ArrowForward';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import SpeedIcon from '@mui/icons-material/Speed';
import EmojiEventsIcon from '@mui/icons-material/EmojiEvents';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';

// project imports
import NodePreviewCard from './NodePreviewCard';
import NodePreviewDetailsPanel from './NodePreviewDetailsPanel';
import IPDetailsDialog from 'components/IPDetailsDialog';
import Alert from '@mui/material/Alert';
import { formatDateTime } from 'i18n/locales';
import { getCountryDisplay } from '../../../utils/countryDisplay';

const BATCH_SIZE = 100;

const buildStatCardSx = (theme, color, clickable = false) => ({
  bgcolor: 'background.paper',
  border: '1px solid',
  borderColor: alpha(color, 0.2),
  borderRadius: 2,
  boxShadow: theme.shadows[1],
  cursor: clickable ? 'pointer' : 'default',
  height: '100%',
  display: 'flex',
  flexDirection: 'column',
  transition: 'all 0.2s ease',
  '&:hover': clickable
    ? {
        borderColor: alpha(color, 0.35),
        boxShadow: theme.shadows[4],
        transform: 'translateY(-1px)'
      }
    : {
        borderColor: alpha(color, 0.24)
      }
});

const buildStatCardContentSx = (isMobile) => ({
  p: isMobile ? 1.25 : 1.5,
  height: '100%',
  display: 'flex',
  flexDirection: 'column',
  justifyContent: 'center',
  '&:last-child': { pb: isMobile ? 1.25 : 1.5 }
});

const normalizeCountryKey = (country) => {
  const trimmed = String(country || '').trim();
  return trimmed || 'unknown';
};

const sortDistributionItems = (sortMode) => (a, b) => {
  if (a.key === 'unknown' && b.key !== 'unknown') return 1;
  if (b.key === 'unknown' && a.key !== 'unknown') return -1;
  const primaryKey = sortMode === 'ips' ? 'uniqueIpCount' : 'count';
  const secondaryKey = sortMode === 'ips' ? 'count' : 'uniqueIpCount';
  if (b[primaryKey] !== a[primaryKey]) return b[primaryKey] - a[primaryKey];
  if (b[secondaryKey] !== a[secondaryKey]) return b[secondaryKey] - a[secondaryKey];
  return a.label.localeCompare(b.label);
};

const buildPreviewDistributions = (nodes, t, sortMode) => {
  const countryMap = new Map();
  const ipMap = new Map();
  const totalNodes = nodes.length;

  nodes.forEach((node) => {
    const countryKey = normalizeCountryKey(node.LinkCountry);
    const countryDisplay = getCountryDisplay(countryKey === 'unknown' ? '' : countryKey, { unknownLabel: t('common.unknown') });
    const landingIP = String(node.LandingIP || '').trim();

    if (!countryMap.has(countryKey)) {
      countryMap.set(countryKey, {
        key: countryKey,
        label: countryDisplay.label,
        flag: countryDisplay.flag,
        count: 0,
        ipSet: new Set(),
        uniqueIpCount: 0
      });
    }

    const countryItem = countryMap.get(countryKey);
    countryItem.count += 1;
    if (landingIP) {
      countryItem.ipSet.add(landingIP);
      if (!ipMap.has(landingIP)) {
        ipMap.set(landingIP, { key: landingIP, label: landingIP, count: 0, uniqueIpCount: 0 });
      }
      ipMap.get(landingIP).count += 1;
    }
  });

  const totalIPs = ipMap.size;
  const metricTotal = sortMode === 'ips' ? totalIPs : totalNodes;
  const countryDistribution = Array.from(countryMap.values())
    .map((item) => {
      const uniqueIpCount = item.ipSet.size;
      const metricValue = sortMode === 'ips' ? uniqueIpCount : item.count;
      return {
        ...item,
        uniqueIpCount,
        metricValue,
        percent: metricTotal > 0 ? Math.round((metricValue / metricTotal) * 100) : 0
      };
    })
    .sort(sortDistributionItems(sortMode));

  return {
    countryDistribution,
    countryCount: countryMap.size,
    ipCount: ipMap.size
  };
};

const DistributionPanel = ({ title, summary, items, emptyText, accentColor, renderMeta, maxListHeight }) => {
  const theme = useTheme();

  return (
    <Box
      sx={{
        p: 1.25,
        height: '100%',
        borderRadius: 2,
        border: '1px solid',
        borderColor: alpha(accentColor, 0.2),
        bgcolor: 'background.paper',
        boxShadow: theme.shadows[1]
      }}
    >
      <Stack direction="row" alignItems="center" justifyContent="space-between" spacing={1} mb={1}>
        <Typography variant="caption" fontWeight={700} color="text.primary">
          {title}
        </Typography>
        <Chip
          size="small"
          label={summary}
          variant="outlined"
          sx={{ height: 20, fontSize: 10, color: accentColor, borderColor: alpha(accentColor, 0.35) }}
        />
      </Stack>

      {items.length === 0 ? (
        <Typography variant="caption" color="text.secondary">
          {emptyText}
        </Typography>
      ) : (
        <Stack
          spacing={0.85}
          sx={{
            maxHeight: maxListHeight,
            overflow: 'auto',
            pr: 0.5,
            scrollbarWidth: 'thin',
            '&::-webkit-scrollbar': { width: 6 },
            '&::-webkit-scrollbar-thumb': {
              borderRadius: 999,
              bgcolor: alpha(accentColor, 0.32)
            }
          }}
        >
          {items.map((item) => (
            <Box key={item.key}>
              <Stack direction="row" alignItems="center" spacing={1} sx={{ minWidth: 0 }}>
                {item.flag && <Typography sx={{ fontSize: 14, lineHeight: 1 }}>{item.flag}</Typography>}
                <Tooltip title={item.label} placement="top" arrow>
                  <Typography variant="caption" color="text.primary" fontWeight={600} noWrap sx={{ flex: 1, minWidth: 0 }}>
                    {item.label}
                  </Typography>
                </Tooltip>
                <Typography variant="caption" color="text.secondary" sx={{ flexShrink: 0, fontSize: 10 }}>
                  {renderMeta(item)}
                </Typography>
              </Stack>
              <Box sx={{ mt: 0.4, height: 5, borderRadius: 999, bgcolor: alpha(accentColor, 0.12), overflow: 'hidden' }}>
                <Box
                  sx={{ width: `${Math.min(item.percent, 100)}%`, height: '100%', borderRadius: 999, bgcolor: alpha(accentColor, 0.75) }}
                />
              </Box>
            </Box>
          ))}
        </Stack>
      )}
    </Box>
  );
};

DistributionPanel.propTypes = {
  title: PropTypes.string.isRequired,
  summary: PropTypes.string.isRequired,
  items: PropTypes.arrayOf(
    PropTypes.shape({
      key: PropTypes.string.isRequired,
      label: PropTypes.string.isRequired,
      flag: PropTypes.string,
      count: PropTypes.number.isRequired,
      uniqueIpCount: PropTypes.number,
      metricValue: PropTypes.number,
      percent: PropTypes.number.isRequired
    })
  ).isRequired,
  emptyText: PropTypes.string.isRequired,
  accentColor: PropTypes.string.isRequired,
  renderMeta: PropTypes.func.isRequired,
  maxListHeight: PropTypes.oneOfType([PropTypes.number, PropTypes.object])
};

DistributionPanel.defaultProps = {
  maxListHeight: 220
};

const formatBytes = (bytes) => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const formatExpireDate = (timestamp, language) => {
  if (!timestamp) return '-';
  const date = new Date(timestamp * 1000);
  return formatDateTime(date, language, { year: 'numeric', month: 'numeric', day: 'numeric' });
};

export default function NodePreviewDialog({ open, loading, data, tagColorMap, onClose }) {
  const theme = useTheme();
  const { t, i18n } = useTranslation();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const contentRef = useRef(null);

  const [searchText, setSearchText] = useState('');
  const [detailsOpen, setDetailsOpen] = useState(false);
  const [selectedNode, setSelectedNode] = useState(null);
  const [ipDialogOpen, setIpDialogOpen] = useState(false);
  const [selectedIP, setSelectedIP] = useState('');
  const [displayCount, setDisplayCount] = useState(BATCH_SIZE);
  const [statsExpanded, setStatsExpanded] = useState(false);
  const [distributionExpanded, setDistributionExpanded] = useState(false);
  const [distributionSortMode, setDistributionSortMode] = useState('nodes');

  useEffect(() => {
    if (!open) {
      setDisplayCount(BATCH_SIZE);
      setSearchText('');
    } else {
      setStatsExpanded(false);
      setDistributionExpanded(false);
      setDistributionSortMode('nodes');
    }
  }, [open]);

  useEffect(() => {
    setDisplayCount(BATCH_SIZE);
  }, [searchText]);

  const filteredNodes = useMemo(() => {
    if (!data?.Nodes) return [];
    if (!searchText.trim()) return data.Nodes;

    const lowerSearch = searchText.toLowerCase();
    return data.Nodes.filter((node) => {
      const displayName = getNodeDisplayName(node);
      return (
        displayName.toLowerCase().includes(lowerSearch) ||
        node.OriginalName?.toLowerCase().includes(lowerSearch) ||
        node.Protocol?.toLowerCase().includes(lowerSearch) ||
        node.Group?.toLowerCase().includes(lowerSearch) ||
        node.Tags?.toLowerCase().includes(lowerSearch)
      );
    });
  }, [data?.Nodes, searchText]);

  const displayedNodes = useMemo(() => {
    return filteredNodes.slice(0, displayCount);
  }, [filteredNodes, displayCount]);

  const nodeStats = useMemo(() => {
    if (!data?.Nodes || data.Nodes.length === 0) {
      return {
        delayPassCount: 0,
        speedPassCount: 0,
        lowestDelayNode: null,
        highestSpeedNode: null,
        countryDistribution: [],
        countryCount: 0,
        ipCount: 0
      };
    }

    const nodes = data.Nodes;

    const delayPassNodes = nodes.filter((node) => {
      const status = node.DelayStatus;
      const isError = status === 'timeout' || status === 'error' || status === 2 || status === 3;
      return !isError && node.DelayTime > 0;
    });

    const speedPassNodes = nodes.filter((node) => {
      const status = node.SpeedStatus;
      const isError = status === 'timeout' || status === 'error' || status === 2 || status === 3;
      return !isError && node.Speed > 0;
    });

    const validNodesForDelay = nodes.filter((node) => {
      const delayStatus = node.DelayStatus;
      const speedStatus = node.SpeedStatus;
      const isDelayError = delayStatus === 'timeout' || delayStatus === 'error' || delayStatus === 2 || delayStatus === 3;
      const isSpeedError = speedStatus === 'timeout' || speedStatus === 'error' || speedStatus === 2 || speedStatus === 3;
      return !isDelayError && !isSpeedError && node.DelayTime > 0 && node.Speed > 0;
    });

    let lowestDelayNode = null;
    if (validNodesForDelay.length > 0) {
      lowestDelayNode = validNodesForDelay.reduce((min, node) => (node.DelayTime < min.DelayTime ? node : min));
    }

    let highestSpeedNode = null;
    if (speedPassNodes.length > 0) {
      highestSpeedNode = speedPassNodes.reduce((max, node) => (node.Speed > max.Speed ? node : max));
    }

    const distributions = buildPreviewDistributions(nodes, t, distributionSortMode);

    return {
      delayPassCount: delayPassNodes.length,
      speedPassCount: speedPassNodes.length,
      lowestDelayNode,
      highestSpeedNode,
      ...distributions
    };
  }, [data?.Nodes, distributionSortMode, t]);

  const hasMore = displayCount < filteredNodes.length;

  const loadMore = useCallback(() => {
    setDisplayCount((prev) => Math.min(prev + BATCH_SIZE, filteredNodes.length));
  }, [filteredNodes.length]);

  const handleScroll = useCallback(
    (e) => {
      const { scrollTop, scrollHeight, clientHeight } = e.target;
      if (scrollHeight - scrollTop - clientHeight < 200 && hasMore) {
        loadMore();
      }
    },
    [hasMore, loadMore]
  );

  const handleViewDetails = (node) => {
    setSelectedNode(node);
    setDetailsOpen(true);
  };

  const handleCloseDetails = () => {
    setDetailsOpen(false);
    setSelectedNode(null);
  };

  const handleViewIP = (ip) => {
    setSelectedIP(ip);
    setIpDialogOpen(true);
  };

  const renderSkeletons = () => (
    <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(140px, 1fr))', gap: 1.5 }}>
      {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10].map((i) => (
        <Skeleton
          key={i}
          variant="rounded"
          height={88}
          sx={{
            borderRadius: 2,
            bgcolor: alpha(theme.palette.action.hover, 0.1)
          }}
        />
      ))}
    </Box>
  );

  const renderEmpty = () => (
    <Box sx={{ textAlign: 'center', py: 8, px: 4 }}>
      <FilterListIcon sx={{ fontSize: 64, color: 'text.secondary', mb: 2 }} />
      <Typography variant="h6" color="text.secondary" gutterBottom>
        {t('subscriptions.preview.emptyTitle')}
      </Typography>
      <Typography variant="body2" color="text.secondary">
        {searchText ? t('subscriptions.preview.emptySearch') : t('subscriptions.preview.emptyFiltered')}
      </Typography>
    </Box>
  );

  return (
    <>
      <Dialog
        open={open}
        onClose={onClose}
        fullScreen={isMobile}
        maxWidth="lg"
        fullWidth
        PaperProps={{
          sx: {
            minHeight: isMobile ? '100%' : '80vh',
            maxHeight: isMobile ? '100%' : '90vh',
            borderRadius: isMobile ? 0 : 4,
            border: '1px solid',
            borderColor: 'divider',
            bgcolor: 'background.paper'
          }
        }}
      >
        <DialogTitle
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: '1px solid',
            borderColor: 'divider',
            pb: 2,
            bgcolor: 'background.default'
          }}
        >
          <Stack direction="row" alignItems="center" spacing={2} flexWrap="wrap">
            <Typography variant="h5" fontWeight="bold">
              {t('subscriptions.preview.title')}
              <Chip size="small" label="Beta" color="error" variant="outlined" sx={{ ml: 1 }} />
            </Typography>
            {!loading && data && (
              <Stack direction="row" alignItems="center" spacing={1} flexWrap="wrap">
                <Chip
                  label={t('subscriptions.preview.totalCount', { count: data.TotalCount })}
                  size="small"
                  variant="outlined"
                  sx={{ fontWeight: 600 }}
                />
                {data.TotalCount !== data.FilteredCount && (
                  <>
                    <ArrowForwardIcon sx={{ fontSize: 14, color: 'text.secondary' }} />
                    <Chip
                      label={t('subscriptions.preview.filteredCount', { count: data.FilteredCount })}
                      size="small"
                      color="primary"
                      sx={{ fontWeight: 600 }}
                    />
                  </>
                )}
                {data.UsageTotal > 0 && (
                  <Chip
                    label={`${formatBytes(data.UsageUpload + data.UsageDownload)} / ${formatBytes(data.UsageTotal)}`}
                    size="small"
                    color="info"
                    variant="outlined"
                    sx={{ fontWeight: 600 }}
                  />
                )}
                {data.UsageExpire > 0 && (
                  <Chip
                    label={t('subscriptions.preview.latestExpireTime', {
                      date: formatExpireDate(data.UsageExpire, i18n.resolvedLanguage || i18n.language)
                    })}
                    size="small"
                    color="warning"
                    variant="outlined"
                    sx={{ fontWeight: 600 }}
                  />
                )}
              </Stack>
            )}
          </Stack>
          <IconButton onClick={onClose} size="medium">
            <CloseIcon />
          </IconButton>
        </DialogTitle>

        <DialogContent sx={{ p: 0, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
          <Box
            sx={{
              px: isMobile ? 1.5 : 3,
              py: 1,
              borderBottom: '1px solid',
              borderColor: 'divider',
              bgcolor: 'background.default',
              flexShrink: 0
            }}
          >
            <Alert
              severity="warning"
              variant="outlined"
              sx={{
                py: 0.5,
                alignItems: 'center',
                '& .MuiAlert-icon': { py: 0.25 },
                '& .MuiAlert-message': { py: 0.25 }
              }}
            >
              <Typography component="span" fontWeight={700} sx={{ mr: 1 }}>
                {t('subscriptions.preview.noticeTitle')}
              </Typography>
              {t('subscriptions.preview.noticeText')}
            </Alert>
          </Box>
          <Box
            sx={{
              px: 3,
              py: 2,
              borderBottom: '1px solid',
              borderColor: 'divider',
              bgcolor: 'background.default',
              flexShrink: 0
            }}
          >
            <TextField
              fullWidth
              size="small"
              placeholder={t('subscriptions.preview.searchPlaceholder')}
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon color="action" />
                  </InputAdornment>
                ),
                sx: { borderRadius: 2 }
              }}
            />
            {searchText && filteredNodes.length !== data?.Nodes?.length && (
              <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
                {t('subscriptions.preview.searchResultCount', { count: filteredNodes.length })}
              </Typography>
            )}
          </Box>

          {!loading && data?.Nodes && data.Nodes.length > 0 && (
            <Box
              sx={{
                px: isMobile ? 1.5 : 3,
                py: 1.5,
                borderBottom: '1px solid',
                borderColor: 'divider',
                bgcolor: 'background.default',
                flexShrink: 0,
                boxShadow: `inset 0 -1px 0 ${alpha(theme.palette.divider, 0.5)}`
              }}
            >
              <Stack
                direction="row"
                alignItems="center"
                justifyContent="space-between"
                onClick={() => setStatsExpanded(!statsExpanded)}
                sx={{
                  cursor: 'pointer',
                  userSelect: 'none',
                  mb: statsExpanded ? 1.5 : 0,
                  transition: 'margin 0.2s ease'
                }}
              >
                <Stack direction="row" alignItems="center" spacing={1}>
                  <EmojiEventsIcon sx={{ fontSize: 18, color: theme.palette.primary.main }} />
                  <Typography variant="subtitle2" fontWeight={600} color="text.primary">
                    {t('subscriptions.preview.stats.title')}
                  </Typography>
                  <Chip
                    label={
                      nodeStats.delayPassCount + nodeStats.speedPassCount > 0
                        ? t('subscriptions.preview.stats.available')
                        : t('subscriptions.preview.stats.noData')
                    }
                    size="small"
                    color={nodeStats.delayPassCount + nodeStats.speedPassCount > 0 ? 'success' : 'default'}
                    variant="outlined"
                    sx={{ fontSize: 10, height: 20 }}
                  />
                </Stack>
                <IconButton size="small" sx={{ p: 0.5 }}>
                  {statsExpanded ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
                </IconButton>
              </Stack>

              <Collapse in={statsExpanded} timeout="auto">
                <Grid container spacing={isMobile ? 1 : 1.5}>
                  <Grid item xs={6} sm={3} sx={{ display: 'flex' }}>
                    <Card elevation={0} sx={buildStatCardSx(theme, theme.palette.success.main)}>
                      <CardContent sx={buildStatCardContentSx(isMobile)}>
                        <Stack direction="row" alignItems="center" spacing={1} mb={0.5}>
                          <AccessTimeIcon sx={{ fontSize: 16, color: theme.palette.success.main }} />
                          <Typography variant="caption" color="text.secondary" fontWeight={500} sx={{ fontSize: isMobile ? 10 : 11 }}>
                            {t('subscriptions.preview.stats.delayPassed')}
                          </Typography>
                        </Stack>
                        <Typography variant={isMobile ? 'h6' : 'h5'} fontWeight={700} color="success.main">
                          {nodeStats.delayPassCount}
                          <Typography component="span" variant="caption" color="text.secondary" sx={{ ml: 0.5, fontWeight: 400 }}>
                            / {data.Nodes.length}
                          </Typography>
                        </Typography>
                      </CardContent>
                    </Card>
                  </Grid>

                  <Grid item xs={6} sm={3} sx={{ display: 'flex' }}>
                    <Card elevation={0} sx={buildStatCardSx(theme, theme.palette.info.main)}>
                      <CardContent sx={buildStatCardContentSx(isMobile)}>
                        <Stack direction="row" alignItems="center" spacing={1} mb={0.5}>
                          <SpeedIcon sx={{ fontSize: 16, color: theme.palette.info.main }} />
                          <Typography variant="caption" color="text.secondary" fontWeight={500} sx={{ fontSize: isMobile ? 10 : 11 }}>
                            {t('subscriptions.preview.stats.speedPassed')}
                          </Typography>
                        </Stack>
                        <Typography variant={isMobile ? 'h6' : 'h5'} fontWeight={700} color="info.main">
                          {nodeStats.speedPassCount}
                          <Typography component="span" variant="caption" color="text.secondary" sx={{ ml: 0.5, fontWeight: 400 }}>
                            / {data.Nodes.length}
                          </Typography>
                        </Typography>
                      </CardContent>
                    </Card>
                  </Grid>

                  <Grid item xs={6} sm={3} sx={{ display: 'flex' }}>
                    <Card
                      elevation={0}
                      onClick={() => nodeStats.lowestDelayNode && handleViewDetails(nodeStats.lowestDelayNode)}
                      sx={buildStatCardSx(theme, theme.palette.warning.main, Boolean(nodeStats.lowestDelayNode))}
                    >
                      <CardContent sx={buildStatCardContentSx(isMobile)}>
                        <Stack direction="row" alignItems="center" spacing={1} mb={0.5}>
                          <AccessTimeIcon sx={{ fontSize: 16, color: theme.palette.warning.main }} />
                          <Typography variant="caption" color="text.secondary" fontWeight={500} sx={{ fontSize: isMobile ? 10 : 11 }}>
                            {t('subscriptions.preview.stats.lowestDelay')}
                          </Typography>
                        </Stack>
                        {nodeStats.lowestDelayNode ? (
                          <>
                            <Tooltip title={getNodeDisplayName(nodeStats.lowestDelayNode)} placement="top" arrow>
                              <Typography
                                variant="caption"
                                color="text.primary"
                                sx={{
                                  fontWeight: 500,
                                  maxWidth: '120px',
                                  overflow: 'hidden',
                                  textOverflow: 'ellipsis',
                                  whiteSpace: 'nowrap',
                                  fontSize: isMobile ? 11 : 12
                                }}
                              >
                                {getNodeDisplayName(nodeStats.lowestDelayNode)}
                              </Typography>
                            </Tooltip>
                            <Typography variant="caption" color="text.secondary" sx={{ fontSize: isMobile ? 10 : 11 }}>
                              {nodeStats.lowestDelayNode.DelayTime}ms · {nodeStats.lowestDelayNode.Speed?.toFixed(1)}MB/s
                            </Typography>
                          </>
                        ) : (
                          <Typography variant="body2" color="text.secondary" sx={{ fontSize: isMobile ? 11 : 12 }}>
                            {t('subscriptions.preview.stats.noData')}
                          </Typography>
                        )}
                      </CardContent>
                    </Card>
                  </Grid>

                  <Grid item xs={6} sm={3} sx={{ display: 'flex' }}>
                    <Card
                      elevation={0}
                      onClick={() => nodeStats.highestSpeedNode && handleViewDetails(nodeStats.highestSpeedNode)}
                      sx={buildStatCardSx(theme, theme.palette.primary.main, Boolean(nodeStats.highestSpeedNode))}
                    >
                      <CardContent sx={buildStatCardContentSx(isMobile)}>
                        <Stack direction="row" alignItems="center" spacing={1} mb={0.5}>
                          <SpeedIcon sx={{ fontSize: 16, color: theme.palette.primary.main }} />
                          <Typography variant="caption" color="text.secondary" fontWeight={500} sx={{ fontSize: isMobile ? 10 : 11 }}>
                            {t('subscriptions.preview.stats.highestSpeed')}
                          </Typography>
                        </Stack>
                        {nodeStats.highestSpeedNode ? (
                          <>
                            <Tooltip title={getNodeDisplayName(nodeStats.highestSpeedNode)} placement="top" arrow>
                              <Typography
                                variant="body2"
                                fontWeight={600}
                                color="primary.main"
                                sx={{
                                  overflow: 'hidden',
                                  textOverflow: 'ellipsis',
                                  whiteSpace: 'nowrap',
                                  fontSize: isMobile ? 11 : 12
                                }}
                              >
                                {getNodeDisplayName(nodeStats.highestSpeedNode)}
                              </Typography>
                            </Tooltip>
                            <Typography variant="caption" color="text.secondary" sx={{ fontSize: isMobile ? 10 : 11 }}>
                              {nodeStats.highestSpeedNode.Speed?.toFixed(1)}MB/s · {nodeStats.highestSpeedNode.DelayTime}ms
                            </Typography>
                          </>
                        ) : (
                          <Typography variant="body2" color="text.secondary" sx={{ fontSize: isMobile ? 11 : 12 }}>
                            {t('subscriptions.preview.stats.noData')}
                          </Typography>
                        )}
                      </CardContent>
                    </Card>
                  </Grid>
                </Grid>
              </Collapse>
            </Box>
          )}

          {!loading && data?.Nodes && data.Nodes.length > 0 && (
            <Box
              sx={{
                px: isMobile ? 1.5 : 3,
                py: 1.5,
                borderBottom: '1px solid',
                borderColor: 'divider',
                bgcolor: 'background.default',
                flexShrink: 0,
                boxShadow: `inset 0 -1px 0 ${alpha(theme.palette.divider, 0.45)}`
              }}
            >
              <Stack
                direction="row"
                alignItems="center"
                justifyContent="space-between"
                onClick={() => setDistributionExpanded(!distributionExpanded)}
                sx={{
                  cursor: 'pointer',
                  userSelect: 'none',
                  mb: distributionExpanded ? 1.5 : 0,
                  transition: 'margin 0.2s ease'
                }}
              >
                <Stack direction="row" alignItems="center" spacing={1} sx={{ minWidth: 0, flexWrap: 'wrap' }}>
                  <FilterListIcon sx={{ fontSize: 18, color: theme.palette.secondary.main }} />
                  <Typography variant="subtitle2" fontWeight={600} color="text.primary">
                    {t('subscriptions.preview.stats.distributionTitle')}
                  </Typography>
                  <Chip
                    label={t('subscriptions.preview.stats.distributionSummary', {
                      countries: nodeStats.countryCount,
                      ips: nodeStats.ipCount
                    })}
                    size="small"
                    color="secondary"
                    variant="outlined"
                    sx={{ fontSize: 10, height: 20 }}
                  />
                  {distributionExpanded && (
                    <ToggleButtonGroup
                      exclusive
                      size="small"
                      value={distributionSortMode}
                      onClick={(event) => event.stopPropagation()}
                      onChange={(event, nextMode) => {
                        event.stopPropagation();
                        if (nextMode) setDistributionSortMode(nextMode);
                      }}
                      sx={{
                        ml: 0.5,
                        '& .MuiToggleButton-root': {
                          px: 1,
                          py: 0.15,
                          fontSize: 10,
                          lineHeight: 1.4,
                          color: 'text.secondary',
                          borderColor: alpha(theme.palette.secondary.main, 0.25),
                          '&.Mui-selected': {
                            color: theme.palette.secondary.main,
                            bgcolor: alpha(theme.palette.secondary.main, 0.12)
                          }
                        }
                      }}
                    >
                      <ToggleButton value="nodes">{t('subscriptions.preview.stats.sortByNodes')}</ToggleButton>
                      <ToggleButton value="ips">{t('subscriptions.preview.stats.sortByIps')}</ToggleButton>
                    </ToggleButtonGroup>
                  )}
                </Stack>
                <IconButton size="small" sx={{ p: 0.5 }}>
                  {distributionExpanded ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
                </IconButton>
              </Stack>

              <Collapse in={distributionExpanded} timeout="auto">
                <DistributionPanel
                  title={t('subscriptions.preview.stats.countryDistribution')}
                  summary={t('subscriptions.preview.stats.countryCount', { count: nodeStats.countryCount })}
                  items={nodeStats.countryDistribution}
                  emptyText={t('subscriptions.preview.stats.noCountryData')}
                  accentColor={theme.palette.secondary.main}
                  maxListHeight={{ xs: 180, sm: 220, md: 240 }}
                  renderMeta={(item) =>
                    t(
                      distributionSortMode === 'ips'
                        ? 'subscriptions.preview.stats.countryDistributionMetaByIps'
                        : 'subscriptions.preview.stats.countryDistributionMetaByNodes',
                      {
                        nodes: item.count,
                        ips: item.uniqueIpCount,
                        percent: item.percent
                      }
                    )
                  }
                />
              </Collapse>
            </Box>
          )}

          <Box
            ref={contentRef}
            onScroll={handleScroll}
            sx={{
              flex: 1,
              overflow: 'auto',
              px: 1.5,
              py: 1.5
            }}
          >
            {loading ? (
              renderSkeletons()
            ) : filteredNodes.length === 0 ? (
              renderEmpty()
            ) : (
              <>
                <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(140px, 1fr))', gap: 1 }}>
                  {displayedNodes.map((node, index) => (
                    <NodePreviewCard key={index} node={node} onClick={() => handleViewDetails(node)} />
                  ))}
                </Box>

                {hasMore && (
                  <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
                    <Button variant="outlined" onClick={loadMore} endIcon={<ExpandMoreIcon />} size="small">
                      {t('subscriptions.preview.loadMore', { displayed: displayCount, total: filteredNodes.length })}
                    </Button>
                  </Box>
                )}

                {!hasMore && displayedNodes.length > BATCH_SIZE && (
                  <Typography variant="caption" color="text.secondary" sx={{ display: 'block', textAlign: 'center', mt: 2 }}>
                    {t('subscriptions.preview.allLoaded', { count: filteredNodes.length })}
                  </Typography>
                )}
              </>
            )}
          </Box>

          {!loading && filteredNodes.length > 0 && (
            <Box
              sx={{
                px: 3,
                py: 1.5,
                borderTop: '1px solid',
                borderColor: 'divider',
                bgcolor: 'background.default',
                textAlign: 'center',
                flexShrink: 0,
                boxShadow: `inset 0 1px 0 ${alpha(theme.palette.divider, 0.4)}`
              }}
            >
              <Typography variant="caption" color="text.secondary">
                {t('subscriptions.preview.footerHint', { displayed: displayedNodes.length, total: filteredNodes.length })}
              </Typography>
            </Box>
          )}
        </DialogContent>
      </Dialog>

      <NodePreviewDetailsPanel
        open={detailsOpen}
        node={selectedNode}
        tagColorMap={tagColorMap}
        onClose={handleCloseDetails}
        onViewIP={handleViewIP}
      />

      <IPDetailsDialog open={ipDialogOpen} onClose={() => setIpDialogOpen(false)} ip={selectedIP} />
    </>
  );
}

NodePreviewDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  loading: PropTypes.bool,
  data: PropTypes.shape({
    Nodes: PropTypes.array,
    TotalCount: PropTypes.number,
    FilteredCount: PropTypes.number
  }),
  tagColorMap: PropTypes.object,
  onClose: PropTypes.func.isRequired
};
