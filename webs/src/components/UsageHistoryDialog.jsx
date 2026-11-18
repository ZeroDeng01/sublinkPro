import PropTypes from 'prop-types';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import * as echarts from 'echarts';

import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import IconButton from '@mui/material/IconButton';
import Stack from '@mui/material/Stack';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

import CloseIcon from '@mui/icons-material/Close';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import InsightsIcon from '@mui/icons-material/Insights';
import RestartAltIcon from '@mui/icons-material/RestartAlt';

import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';
import { getAirportUsageHistory } from 'api/usageHistory';

// 时间范围预设（0 表示全部保留期，由后端归一化为保留天数）
const RANGE_PRESETS = [1, 3, 7, 30, 90, 0];
// 默认只看最近一天：数据刚接入或刷新间隔较长时，长窗口会让曲线挤在末端。
const DEFAULT_RANGE_DAYS = 1;
const GRANULARITY_OPTIONS = ['auto', 'hour', 'day'];

const formatBytes = (bytes, decimals = 2) => {
  const value = Number(bytes);
  if (!Number.isFinite(value) || value <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  const index = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1);
  return `${parseFloat((value / 1024 ** index).toFixed(decimals))} ${units[index]}`;
};

const formatExpireDate = (timestampSeconds) => {
  const seconds = Number(timestampSeconds);
  if (!Number.isFinite(seconds) || seconds <= 0) return '';
  return new Date(seconds * 1000).toLocaleDateString();
};

const formatPointTime = (timestamp, granularity) => {
  const date = new Date(Number(timestamp));
  if (Number.isNaN(date.getTime())) return '';
  const pad = (value) => String(value).padStart(2, '0');
  const datePart = `${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
  if (granularity === 'refresh') {
    return `${datePart} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
  }
  if (granularity === 'hour') {
    return `${datePart} ${pad(date.getHours())}:00`;
  }
  return `${date.getFullYear()}-${datePart}`;
};

const formatFullTime = (timestamp) => {
  const date = new Date(Number(timestamp));
  if (Number.isNaN(date.getTime())) return '';
  const pad = (value) => String(value).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
};

const escapeHtml = (value) =>
  String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');

const usagePercentColor = (percent, palette) => {
  if (percent >= 85) return palette.error.main;
  if (percent >= 60) return palette.warning.main;
  return palette.success.main;
};

const findClosestPoint = (points, timestamp) => {
  if (!points.length) return null;
  let closest = points[0];
  let closestDistance = Math.abs(closest.timestamp - timestamp);
  for (let index = 1; index < points.length; index += 1) {
    const distance = Math.abs(points[index].timestamp - timestamp);
    if (distance < closestDistance) {
      closest = points[index];
      closestDistance = distance;
    }
  }
  return closest;
};

/**
 * 机场用量历史图表弹窗。
 */
function UsageHistoryDialog({ open, targetId, title, onClose }) {
  const theme = useTheme();
  const { t, i18n } = useTranslation();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);

  // 图表使用具体色值（canvas 无法解析 CSS 变量）
  const chartPalette = useMemo(() => {
    const scheme = theme.colorSchemes?.[isDark ? 'dark' : 'light']?.palette;
    return scheme || theme.palette;
  }, [theme, isDark]);

  const [rangeDays, setRangeDays] = useState(DEFAULT_RANGE_DAYS);
  const [granularity, setGranularity] = useState('auto');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [data, setData] = useState(null);
  const [chartReady, setChartReady] = useState(false);
  const [chartFailed, setChartFailed] = useState(false);

  const chartInstanceRef = useRef(null);
  const chartOptionRef = useRef(null);
  const chartInitFrameRef = useRef(0);
  const chartResizeObserverRef = useRef(null);

  // 每次打开或切换目标时回到最近一天，避免默认长窗口里数据集中在末端。
  useEffect(() => {
    if (!open) return;
    setRangeDays(DEFAULT_RANGE_DAYS);
    setGranularity('auto');
  }, [open, targetId]);

  useEffect(() => {
    if (!open || !targetId) return undefined;

    let cancelled = false;
    setLoading(true);
    setError('');

    const params = { days: rangeDays, granularity };

    getAirportUsageHistory(targetId, params)
      .then((response) => {
        if (cancelled) return;
        setData(response?.data || null);
      })
      .catch((err) => {
        if (cancelled) return;
        setData(null);
        setError(err?.message || t('usageHistory.loadFailed'));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [open, targetId, rangeDays, granularity, t]);

  // 应用最新图表配置；实例已销毁时静默跳过
  const applyChartOption = useCallback(() => {
    const chart = chartInstanceRef.current;
    if (!chart || chart.isDisposed()) return;
    if (!chartOptionRef.current) {
      chart.clear();
      return;
    }
    chart.setOption(chartOptionRef.current, true);
    chart.resize();
  }, []);

  // 用回调 ref 初始化实例：弹窗内容挂载时机不固定，必须等容器真正有尺寸后再创建，
  // 否则 echarts.init 会拿到 0 尺寸容器并在之后完全无法绘制。
  const attachChart = useCallback(
    (node) => {
      if (chartInitFrameRef.current) {
        cancelAnimationFrame(chartInitFrameRef.current);
        chartInitFrameRef.current = 0;
      }
      if (chartResizeObserverRef.current) {
        chartResizeObserverRef.current.disconnect();
        chartResizeObserverRef.current = null;
      }
      const previous = chartInstanceRef.current;
      if (previous) {
        chartInstanceRef.current = null;
        if (!previous.isDisposed()) previous.dispose();
        setChartReady(false);
      }
      if (!node) return;

      // 同一 DOM 上可能残留旧实例（例如关闭动画未结束时又重新打开），先释放
      const stale = echarts.getInstanceByDom(node);
      if (stale && !stale.isDisposed()) {
        stale.dispose();
      }

      let attempts = 0;
      const createInstance = () => {
        chartInitFrameRef.current = 0;
        if (!node.isConnected) return;
        if (node.clientWidth === 0 || node.clientHeight === 0) {
          attempts += 1;
          if (attempts <= 120) {
            chartInitFrameRef.current = requestAnimationFrame(createInstance);
          }
          return;
        }
        try {
          const chart = echarts.init(node, null, { renderer: 'canvas' });
          chartInstanceRef.current = chart;
          setChartFailed(false);
          setChartReady(true);
          applyChartOption();
          if (typeof ResizeObserver !== 'undefined') {
            const observer = new ResizeObserver(() => {
              const current = chartInstanceRef.current;
              if (current && !current.isDisposed()) current.resize();
            });
            observer.observe(node);
            chartResizeObserverRef.current = observer;
          }
        } catch (initError) {
          console.error('Failed to initialize usage chart', initError);
          setChartFailed(true);
        }
      };
      chartInitFrameRef.current = requestAnimationFrame(createInstance);
    },
    [applyChartOption]
  );

  useEffect(() => {
    const handleResize = () => {
      const chart = chartInstanceRef.current;
      if (chart && !chart.isDisposed()) chart.resize();
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  const points = useMemo(() => data?.points || [], [data]);
  const failedPoints = useMemo(() => points.filter((point) => point.failed), [points]);

  // 数据起点明显晚于窗口起点时，提示实际开始记录的时间
  const historyStartsAt = useMemo(() => {
    if (!data || points.length < 2) return '';
    const windowFrom = Date.parse(data.from);
    const windowTo = Date.parse(data.to);
    if (!Number.isFinite(windowFrom) || !Number.isFinite(windowTo) || windowTo <= windowFrom) return '';
    const firstTimestamp = Number(points[0].timestamp);
    if (!Number.isFinite(firstTimestamp)) return '';
    if (firstTimestamp - windowFrom <= (windowTo - windowFrom) * 0.1) return '';
    return formatFullTime(firstTimestamp);
  }, [data, points]);

  const summary = useMemo(() => {
    if (!data) return null;
    return {
      used: Number(data.currentUsed) || 0,
      total: Number(data.currentTotal) || 0,
      expire: Number(data.currentExpire) || 0
    };
  }, [data]);

  const summaryPercent = summary && summary.total > 0 ? Math.min((summary.used / summary.total) * 100, 100) : 0;

  const rangeOptions = useMemo(
    () =>
      RANGE_PRESETS.map((days) => ({
        value: days,
        label: days === 0 ? t('usageHistory.range.all') : t('usageHistory.range.days', { count: days })
      })),
    [t]
  );

  const granularityOptions = useMemo(
    () =>
      GRANULARITY_OPTIONS.map((value) => ({
        value,
        label: t(`usageHistory.granularity.${value}`)
      })),
    [t]
  );

  // 组装 ECharts 配置
  const chartOption = useMemo(() => {
    if (!data) return null;

    const primary = chartPalette.primary.main;
    const errorColor = chartPalette.error.main;
    const axisTextColor = chartPalette.text?.secondary || chartPalette.text?.primary || primary;
    const dividerColor = chartPalette.divider;
    const paperColor = chartPalette.background?.paper || '#ffffff';
    const pointTimeGranularity = data.granularity || granularity;
    const hasFailures = failedPoints.length > 0;

    // 只有一个采样点时折线不可见，改为绘制数据点，并把时间轴展开到前后各半天。
    const isSinglePoint = points.length === 1;
    const usedValues = points.map((point) => Number(point.used) || 0);
    const minUsed = usedValues.length ? Math.min(...usedValues) : 0;
    const maxUsed = usedValues.length ? Math.max(...usedValues) : 0;
    const usedSpan = maxUsed - minUsed;
    const yPadding = usedSpan > 0 ? usedSpan * 0.08 : Math.max(maxUsed * 0.05, 1024 * 1024);
    // 时间轴固定为所选查询窗口：数据只占据实际存在的时段，
    // 不会被 ECharts 拉伸到铺满整个图表宽度。
    const windowFrom = Date.parse(data.from);
    const windowTo = Date.parse(data.to);
    const hasWindow = Number.isFinite(windowFrom) && Number.isFinite(windowTo) && windowTo > windowFrom;
    let xMin;
    let xMax;
    if (isSinglePoint) {
      const singleTimestamp = Number(points[0].timestamp);
      xMin = singleTimestamp - 12 * 60 * 60 * 1000;
      xMax = singleTimestamp + 12 * 60 * 60 * 1000;
    } else if (hasWindow) {
      xMin = windowFrom;
      xMax = windowTo;
    }

    const series = [
      {
        name: t('usageHistory.series.used'),
        type: 'line',
        showSymbol: isSinglePoint,
        symbol: 'circle',
        symbolSize: isSinglePoint ? 9 : 4,
        smooth: 0.15,
        connectNulls: true,
        lineStyle: { width: 2, color: primary },
        itemStyle: { color: primary },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: withAlpha(primary, 0.3) },
            { offset: 1, color: withAlpha(primary, 0) }
          ])
        },
        emphasis: { focus: 'series' },
        data: points.map((point) => [point.timestamp, Number(point.used) || 0])
      }
    ];

    // 蓝点标记成功刷新；范围越长点越密，用更小的点保持可读性
    const successfulPoints = points.filter((point) => !point.failed);
    const showRefreshDots = !isSinglePoint && successfulPoints.length > 0;
    const refreshDotSize = points.length > 1200 ? 3 : points.length > 400 ? 4 : 5;

    if (showRefreshDots) {
      series.push({
        name: t('usageHistory.series.refresh'),
        type: 'scatter',
        symbol: 'circle',
        symbolSize: refreshDotSize,
        z: 5,
        itemStyle: { color: primary, borderColor: paperColor, borderWidth: 1 },
        data: successfulPoints.map((point) => [point.timestamp, Number(point.used) || 0])
      });
    }

    if (hasFailures) {
      series.push({
        name: t('usageHistory.series.failed'),
        type: 'scatter',
        symbolSize: 8,
        z: 12,
        itemStyle: { color: errorColor, borderColor: paperColor, borderWidth: 1 },
        data: failedPoints.map((point) => [point.timestamp, Number(point.used) || 0])
      });
    }

    const legendNames = [t('usageHistory.series.used')];
    if (showRefreshDots) legendNames.push(t('usageHistory.series.refresh'));
    if (hasFailures) legendNames.push(t('usageHistory.series.failed'));
    const hasLegend = legendNames.length > 1;

    return {
      animationDuration: 280,
      color: [primary, errorColor],
      grid: { left: 4, right: 12, top: hasLegend ? 34 : 16, bottom: 24, containLabel: true },
      legend: hasLegend
        ? {
            top: 0,
            right: 0,
            icon: 'roundRect',
            itemWidth: 10,
            itemHeight: 10,
            textStyle: { color: withAlpha(axisTextColor, 0.85), fontSize: 11 },
            data: legendNames
          }
        : undefined,
      tooltip: {
        trigger: 'axis',
        confine: true,
        axisPointer: { type: 'line', lineStyle: { color: withAlpha(primary, 0.4), width: 1 } },
        backgroundColor: isDark ? withAlpha(paperColor, 0.97) : '#ffffff',
        borderColor: withAlpha(dividerColor, 0.9),
        borderWidth: 1,
        padding: [10, 12],
        textStyle: { color: withAlpha(axisTextColor, 1), fontSize: 12 },
        extraCssText: 'border-radius:10px;box-shadow:0 8px 24px rgba(0,0,0,0.18);',
        formatter: (params) => {
          const list = Array.isArray(params) ? params : [params];
          const rawTimestamp = list[0]?.value?.[0] ?? list[0]?.axisValue;
          const point = findClosestPoint(points, Number(rawTimestamp));
          if (!point) return '';

          const rows = [
            `<div style="font-weight:600;margin-bottom:6px;">${escapeHtml(formatPointTime(point.timestamp, pointTimeGranularity))}</div>`,
            `<div style="display:flex;justify-content:space-between;gap:18px;"><span>${escapeHtml(
              t('usageHistory.tooltip.used')
            )}</span><b>${escapeHtml(formatBytes(point.used))}</b></div>`,
            `<div style="display:flex;justify-content:space-between;gap:18px;"><span>${escapeHtml(
              t('usageHistory.tooltip.upload')
            )}</span><span>${escapeHtml(formatBytes(point.upload))}</span></div>`,
            `<div style="display:flex;justify-content:space-between;gap:18px;"><span>${escapeHtml(
              t('usageHistory.tooltip.download')
            )}</span><span>${escapeHtml(formatBytes(point.download))}</span></div>`
          ];

          if (Number(point.total) > 0) {
            const percent = Math.min((Number(point.used) / Number(point.total)) * 100, 100);
            rows.push(
              `<div style="display:flex;justify-content:space-between;gap:18px;"><span>${escapeHtml(
                t('usageHistory.tooltip.quota')
              )}</span><span>${escapeHtml(formatBytes(point.total))} · ${percent.toFixed(1)}%</span></div>`
            );
          }

          if (Number(point.expire) > 0) {
            rows.push(
              `<div style="display:flex;justify-content:space-between;gap:18px;"><span>${escapeHtml(
                t('usageHistory.tooltip.expire')
              )}</span><span>${escapeHtml(formatExpireDate(point.expire))}</span></div>`
            );
          }

          if (point.failed) {
            rows.push(`<div style="margin-top:6px;color:${escapeHtml(errorColor)};">${escapeHtml(t('usageHistory.tooltip.failed'))}</div>`);
          }

          return rows.join('');
        }
      },
      xAxis: {
        type: 'time',
        min: xMin,
        max: xMax,
        axisLabel: { color: withAlpha(axisTextColor, 0.86), fontSize: 11, hideOverlap: true },
        axisLine: { lineStyle: { color: withAlpha(dividerColor, 0.9) } },
        axisTick: { show: false },
        splitLine: { show: false }
      },
      yAxis: {
        type: 'value',
        splitNumber: 4,
        // 值完全相同时给纵轴上下留白，否则单个点会贴在顶部难以辨认
        min: points.length > 0 && usedSpan === 0 ? Math.max(0, minUsed - yPadding) : undefined,
        max: points.length > 0 && usedSpan === 0 ? maxUsed + yPadding : undefined,
        axisLabel: {
          color: withAlpha(axisTextColor, 0.86),
          fontSize: 11,
          formatter: (value) => formatBytes(value, 1)
        },
        splitLine: { lineStyle: { color: withAlpha(dividerColor, 0.55), type: 'dashed' } }
      },
      dataZoom: [{ type: 'inside', throttle: 60, zoomOnMouseWheel: true, moveOnMouseMove: true, moveOnMouseWheel: false }],
      series
    };
  }, [data, chartPalette, isDark, granularity, points, failedPoints, t]);

  // 应用配置：实例创建完成或配置变化时都重新应用一次
  useEffect(() => {
    chartOptionRef.current = chartOption;
    if (!chartReady) return;
    applyChartOption();
  }, [chartReady, chartOption, i18n.resolvedLanguage, applyChartOption]);

  const handleResetZoom = () => {
    const chart = chartInstanceRef.current;
    if (!chart) return;
    chart.dispatchAction({ type: 'dataZoom', start: 0, end: 100 });
  };

  const dialogPaperSx = {
    borderRadius: isMobile ? 0 : 3,
    overflow: 'hidden',
    bgcolor: dialogSurface,
    backgroundImage: dialogSurfaceGradient,
    border: '1px solid',
    borderColor: panelBorder
  };

  const titleSx = {
    px: 2.5,
    py: 2,
    bgcolor: mutedPanelSurface,
    borderBottom: '1px solid',
    borderColor: panelBorder
  };

  const actionsSx = {
    px: 2.5,
    py: 1.5,
    bgcolor: mutedPanelSurface,
    borderTop: '1px solid',
    borderColor: panelBorder
  };

  const panelSx = {
    p: 1.5,
    borderRadius: 2,
    bgcolor: nestedPanelSurface,
    border: '1px solid',
    borderColor: withAlpha(chartPalette.divider, isDark ? 0.6 : 0.8)
  };

  const overlayVisible = loading || Boolean(error) || points.length === 0 || chartFailed;

  // 空窗口提示：保留期长于当前窗口时，引导用户放宽时间范围。
  const emptyText = useMemo(() => {
    if (error) return error;
    if (chartFailed) return t('usageHistory.chartUnavailable');
    const retention = Number(data?.retentionDays) || 0;
    if (retention > 0 && rangeDays > 0 && rangeDays < retention) return t('usageHistory.emptyWiden');
    return t('usageHistory.empty');
  }, [error, chartFailed, data, rangeDays, t]);

  const summaryChips = [];
  if (summary) {
    summaryChips.push({
      key: 'used',
      label: t('usageHistory.summary.used'),
      value: summary.total > 0 ? `${formatBytes(summary.used)} / ${formatBytes(summary.total)}` : formatBytes(summary.used),
      color: summary.total > 0 ? usagePercentColor(summaryPercent, chartPalette) : chartPalette.primary.main
    });
    if (summary.total > 0) {
      summaryChips.push({
        key: 'percent',
        label: t('usageHistory.summary.percent'),
        value: `${summaryPercent.toFixed(1)}%`,
        color: usagePercentColor(summaryPercent, chartPalette)
      });
    }
    if (summary.expire > 0) {
      summaryChips.push({
        key: 'expire',
        label: t('usageHistory.summary.expire'),
        value: formatExpireDate(summary.expire),
        color: summary.expire * 1000 < Date.now() ? chartPalette.error.main : chartPalette.text.primary
      });
    }
    if (data?.retentionDays) {
      summaryChips.push({
        key: 'retention',
        label: t('usageHistory.summary.retention'),
        value: t('usageHistory.summary.retentionValue', { count: data.retentionDays }),
        color: chartPalette.text.secondary
      });
    }
  }

  return (
    <Dialog open={open} onClose={onClose} maxWidth="lg" fullWidth fullScreen={isMobile} slotProps={{ paper: { sx: dialogPaperSx } }}>
      <DialogTitle sx={titleSx}>
        <Stack direction="row" alignItems="center" spacing={1} sx={{ minWidth: 0 }}>
          <InsightsIcon sx={{ color: palette.primary.main, flexShrink: 0 }} />
          <Box sx={{ minWidth: 0, flex: 1 }}>
            <Typography variant="h6" sx={{ color: primaryText, lineHeight: 1.25 }} noWrap>
              {title || t('usageHistory.title')}
            </Typography>
            <Typography variant="caption" sx={{ color: secondaryText }}>
              {t('usageHistory.subtitle.airport')}
            </Typography>
          </Box>
          <IconButton onClick={onClose} aria-label={t('common.close')} sx={{ color: secondaryText }}>
            <CloseIcon fontSize="small" />
          </IconButton>
        </Stack>
      </DialogTitle>

      <DialogContent sx={{ px: 2.5, py: 2 }}>
        <Stack spacing={1.75}>
          {summaryChips.length > 0 && (
            <Box
              sx={{
                display: 'grid',
                gridTemplateColumns: `repeat(auto-fit, minmax(${isMobile ? '132px' : '160px'}, 1fr))`,
                gap: 1
              }}
            >
              {summaryChips.map((chip) => (
                <Box key={chip.key} sx={{ ...panelSx, p: 1.1 }}>
                  <Typography variant="caption" sx={{ color: tertiaryText, display: 'block', mb: 0.25 }}>
                    {chip.label}
                  </Typography>
                  <Typography variant="body2" sx={{ color: chip.color, fontWeight: 600 }} noWrap>
                    {chip.value}
                  </Typography>
                </Box>
              ))}
            </Box>
          )}

          {failedPoints.length > 0 && (
            <Stack
              direction="row"
              spacing={0.75}
              alignItems="center"
              sx={{
                alignSelf: 'flex-start',
                px: 1.25,
                py: 0.6,
                borderRadius: 1.5,
                border: '1px solid',
                borderColor: withAlpha(chartPalette.error.main, 0.35),
                bgcolor: withAlpha(chartPalette.error.main, isDark ? 0.14 : 0.06)
              }}
            >
              <ErrorOutlineIcon sx={{ fontSize: 16, color: chartPalette.error.main }} />
              <Typography variant="caption" sx={{ color: chartPalette.error.main }}>
                {t('usageHistory.failureNotice', { count: failedPoints.length })}
              </Typography>
            </Stack>
          )}

          <Stack
            direction={{ xs: 'column', md: 'row' }}
            spacing={1}
            justifyContent="space-between"
            alignItems={{ xs: 'stretch', md: 'center' }}
          >
            <ToggleButtonGroup
              exclusive
              size="small"
              value={rangeDays}
              onChange={(_, nextValue) => {
                if (nextValue !== null) setRangeDays(nextValue);
              }}
              sx={{
                alignSelf: { xs: 'flex-start', md: 'center' },
                flexWrap: 'wrap',
                '& .MuiToggleButton-root': {
                  px: 1.1,
                  py: 0.4,
                  fontSize: '0.72rem',
                  color: secondaryText,
                  borderColor: withAlpha(chartPalette.divider, isDark ? 0.6 : 0.8),
                  '&.Mui-selected': {
                    color: palette.primary.main,
                    bgcolor: withAlpha(palette.primary.main, isDark ? 0.2 : 0.1)
                  }
                }
              }}
            >
              {rangeOptions.map((option) => (
                <ToggleButton key={option.value} value={option.value}>
                  {option.label}
                </ToggleButton>
              ))}
            </ToggleButtonGroup>

            <Stack direction="row" spacing={1} alignItems="center">
              <ToggleButtonGroup
                exclusive
                size="small"
                value={granularity}
                onChange={(_, nextValue) => {
                  if (nextValue !== null) setGranularity(nextValue);
                }}
                sx={{
                  '& .MuiToggleButton-root': {
                    px: 1.1,
                    py: 0.4,
                    fontSize: '0.72rem',
                    color: secondaryText,
                    borderColor: withAlpha(chartPalette.divider, isDark ? 0.6 : 0.8),
                    '&.Mui-selected': {
                      color: palette.primary.main,
                      bgcolor: withAlpha(palette.primary.main, isDark ? 0.2 : 0.1)
                    }
                  }
                }}
              >
                {granularityOptions.map((option) => (
                  <ToggleButton key={option.value} value={option.value}>
                    {option.label}
                  </ToggleButton>
                ))}
              </ToggleButtonGroup>
              <Tooltip
                arrow
                title={
                  <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5, py: 0.25 }}>
                    <span>{t('usageHistory.granularityHelp.auto')}</span>
                    <span>{t('usageHistory.granularityHelp.hour')}</span>
                    <span>{t('usageHistory.granularityHelp.day')}</span>
                  </Box>
                }
              >
                <IconButton size="small" aria-label={t('usageHistory.granularityHelp.label')} sx={{ color: secondaryText }}>
                  <InfoOutlinedIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              <Tooltip title={t('usageHistory.resetZoom')} arrow>
                <span>
                  <IconButton
                    size="small"
                    onClick={handleResetZoom}
                    disabled={overlayVisible}
                    aria-label={t('usageHistory.resetZoom')}
                    sx={{ color: secondaryText }}
                  >
                    <RestartAltIcon fontSize="small" />
                  </IconButton>
                </span>
              </Tooltip>
            </Stack>
          </Stack>

          <Box sx={{ ...panelSx, p: { xs: 0.75, md: 1.25 } }}>
            <Box sx={{ position: 'relative' }}>
              <Box ref={attachChart} sx={{ width: '100%', height: { xs: 240, md: 330 } }} />
              {overlayVisible && (
                <Box
                  sx={{
                    position: 'absolute',
                    inset: 0,
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: 1,
                    px: 2,
                    textAlign: 'center',
                    borderRadius: 1.5,
                    bgcolor: withAlpha(chartPalette.background?.paper || '#ffffff', 0.86)
                  }}
                >
                  {loading ? (
                    <>
                      <CircularProgress size={24} />
                      <Typography variant="caption" sx={{ color: secondaryText }}>
                        {t('usageHistory.loading')}
                      </Typography>
                    </>
                  ) : (
                    <>
                      <InsightsIcon sx={{ fontSize: 32, color: tertiaryText, opacity: 0.6 }} />
                      <Typography variant="body2" sx={{ color: secondaryText }}>
                        {emptyText}
                      </Typography>
                    </>
                  )}
                </Box>
              )}
            </Box>
          </Box>

          {historyStartsAt && (
            <Typography variant="caption" sx={{ color: tertiaryText }}>
              {t('usageHistory.historyStarts', { time: historyStartsAt })}
            </Typography>
          )}

          {points.length === 1 && (
            <Typography variant="caption" sx={{ color: tertiaryText }}>
              {t('usageHistory.singleSample')}
            </Typography>
          )}
        </Stack>
      </DialogContent>

      <DialogActions sx={actionsSx}>
        <Button onClick={onClose} variant="contained" size="small">
          {t('common.close')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

UsageHistoryDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  targetId: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
  title: PropTypes.string,
  onClose: PropTypes.func.isRequired
};

export default UsageHistoryDialog;
