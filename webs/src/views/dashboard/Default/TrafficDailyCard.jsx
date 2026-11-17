import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import * as echarts from 'echarts';

// material-ui
import { useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CircularProgress from '@mui/material/CircularProgress';
import IconButton from '@mui/material/IconButton';
import Skeleton from '@mui/material/Skeleton';
import Stack from '@mui/material/Stack';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

// project imports
import { getAirportUsageDaily } from 'api/usageHistory';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';
import { formatBytes } from 'views/airports/utils';

// assets
import RefreshIcon from '@mui/icons-material/Refresh';
import { IconChartAreaLine } from '@tabler/icons-react';

const RANGE_PRESETS = [7, 14, 30, 90];
const DEFAULT_RANGE_DAYS = 14;

const escapeHtml = (value) =>
  String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');

const formatDay = (date) => (typeof date === 'string' ? date.slice(5) : '');

// ==============================|| 每日流量卡片 ||============================== //

/**
 * 每日流量消耗卡片：按天展示订阅刷新时记录的流量增量。
 */
export default function TrafficDailyCard() {
  const theme = useTheme();
  const { t, i18n } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const { dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(theme, isDark);
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);

  const [rangeDays, setRangeDays] = useState(DEFAULT_RANGE_DAYS);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [data, setData] = useState(null);
  const [chartReady, setChartReady] = useState(false);

  const containerRef = useRef(null);
  const instanceRef = useRef(null);
  const optionRef = useRef(null);

  // 图表使用具体色值（canvas 无法解析 CSS 变量）
  const chartPalette = useMemo(() => {
    const scheme = theme.colorSchemes?.[isDark ? 'dark' : 'light']?.palette;
    return scheme || theme.palette;
  }, [theme, isDark]);

  // 强调色会写入 canvas 渐变，必须取具体色值而不是 theme.vars 里的 var(--...)
  const accent = chartPalette.info?.main || theme.palette.info.main;
  const todayAccent = chartPalette.primary?.main || theme.palette.primary.main;

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const response = await getAirportUsageDaily({ days: rangeDays });
      setData(response?.data || null);
    } catch (err) {
      setData(null);
      setError(err?.message || t('dashboard.default.trafficDaily.loadFailed'));
    } finally {
      setLoading(false);
    }
  }, [rangeDays, t]);

  useEffect(() => {
    load();
  }, [load]);

  // 初始化图表实例：卡片始终挂载，容器此时已有尺寸
  useEffect(() => {
    const node = containerRef.current;
    if (!node) return undefined;

    const chart = echarts.init(node, null, { renderer: 'canvas' });
    instanceRef.current = chart;
    setChartReady(true);
    if (optionRef.current) chart.setOption(optionRef.current, true);

    let observer = null;
    if (typeof ResizeObserver !== 'undefined') {
      observer = new ResizeObserver(() => {
        if (!chart.isDisposed()) chart.resize();
      });
      observer.observe(node);
    }

    return () => {
      if (observer) observer.disconnect();
      instanceRef.current = null;
      setChartReady(false);
      if (!chart.isDisposed()) chart.dispose();
    };
  }, []);

  const points = useMemo(() => data?.points || [], [data]);

  const summary = useMemo(() => {
    const todayUsed = Number(data?.todayUsed) || 0;
    const rangeUsed = Number(data?.rangeUsed) || 0;
    return {
      todayUsed,
      rangeUsed,
      dailyAverage: points.length > 0 ? rangeUsed / points.length : 0
    };
  }, [data, points]);

  const option = useMemo(() => {
    if (!points.length) return null;

    const axisTextColor = chartPalette.text?.secondary || chartPalette.text?.primary || accent;
    const dividerColor = chartPalette.divider;
    const paperColor = chartPalette.background?.paper || '#ffffff';
    const lastIndex = points.length - 1;

    const tooltipFormatter = (params) => {
      const list = Array.isArray(params) ? params : [params];
      const point = points[list[0]?.dataIndex];
      if (!point) return '';
      const rows = [
        `<div style="font-weight:600;margin-bottom:6px;">${escapeHtml(point.date)}</div>`,
        `<div style="display:flex;justify-content:space-between;gap:18px;"><span>${escapeHtml(
          t('dashboard.default.trafficDaily.series.used')
        )}</span><b>${escapeHtml(formatBytes(point.used))}</b></div>`,
        `<div style="display:flex;justify-content:space-between;gap:18px;"><span>${escapeHtml(
          t('usageHistory.tooltip.upload')
        )}</span><span>${escapeHtml(formatBytes(point.upload))}</span></div>`,
        `<div style="display:flex;justify-content:space-between;gap:18px;"><span>${escapeHtml(
          t('usageHistory.tooltip.download')
        )}</span><span>${escapeHtml(formatBytes(point.download))}</span></div>`
      ];
      return rows.join('');
    };

    return {
      animationDuration: 260,
      grid: { left: 4, right: 12, top: 16, bottom: 4, containLabel: true },
      tooltip: {
        trigger: 'axis',
        confine: true,
        axisPointer: { type: 'shadow', shadowStyle: { color: withAlpha(accent, isDark ? 0.12 : 0.06) } },
        backgroundColor: isDark ? withAlpha(paperColor, 0.97) : '#ffffff',
        borderColor: withAlpha(dividerColor, 0.9),
        borderWidth: 1,
        padding: [10, 12],
        textStyle: { color: withAlpha(axisTextColor, 1), fontSize: 12 },
        extraCssText: 'border-radius:10px;box-shadow:0 8px 24px rgba(0,0,0,0.18);',
        formatter: tooltipFormatter
      },
      xAxis: {
        type: 'category',
        data: points.map((point) => formatDay(point.date)),
        axisLabel: { color: withAlpha(axisTextColor, 0.86), fontSize: 11, hideOverlap: true },
        axisLine: { lineStyle: { color: withAlpha(dividerColor, 0.9) } },
        axisTick: { show: false }
      },
      yAxis: {
        type: 'value',
        splitNumber: 4,
        axisLabel: {
          color: withAlpha(axisTextColor, 0.86),
          fontSize: 11,
          formatter: (value) => formatBytes(value, 1)
        },
        splitLine: { lineStyle: { color: withAlpha(dividerColor, 0.55), type: 'dashed' } }
      },
      series: [
        {
          name: t('dashboard.default.trafficDaily.series.used'),
          type: 'bar',
          barMaxWidth: 22,
          itemStyle: { borderRadius: [4, 4, 0, 0] },
          emphasis: { focus: 'series' },
          data: points.map((point, index) => {
            const color = index === lastIndex ? todayAccent : accent;
            return {
              value: Number(point.used) || 0,
              itemStyle: {
                color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                  { offset: 0, color: withAlpha(color, 0.95) },
                  { offset: 1, color: withAlpha(color, 0.35) }
                ])
              }
            };
          })
        }
      ]
    };
  }, [points, chartPalette, isDark, accent, todayAccent, t]);

  // 配置或实例就绪时重新应用
  useEffect(() => {
    optionRef.current = option;
    const chart = instanceRef.current;
    if (!chartReady || !chart || chart.isDisposed()) return;
    if (!option) {
      chart.clear();
      return;
    }
    chart.setOption(option, true);
  }, [chartReady, option, i18n.resolvedLanguage]);

  const rangeOptions = RANGE_PRESETS.map((days) => ({ value: days, label: t('dashboard.default.trafficDaily.days', { count: days }) }));

  const summaryChips = [
    { key: 'today', label: t('dashboard.default.trafficDaily.today'), value: formatBytes(summary.todayUsed, 1), color: todayAccent },
    { key: 'range', label: t('dashboard.default.trafficDaily.rangeTotal'), value: formatBytes(summary.rangeUsed, 1), color: accent },
    {
      key: 'average',
      label: t('dashboard.default.trafficDaily.dailyAverage'),
      value: formatBytes(summary.dailyAverage, 1),
      color: primaryText
    }
  ];

  const overlayVisible = loading || Boolean(error) || points.length === 0;
  // 完全没有流量记录时（未开启用量获取）不占用首页空间
  if (!loading && !error && summary.rangeUsed === 0) {
    return null;
  }

  return (
    <Card
      sx={{
        mb: 4,
        borderRadius: 4,
        overflow: 'hidden',
        position: 'relative',
        bgcolor: dialogSurface,
        backgroundImage: dialogSurfaceGradient,
        border: '1px solid',
        borderColor: panelBorder,
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          height: 3,
          backgroundColor: accent
        }
      }}
    >
      <CardContent sx={{ p: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 2.5, flexWrap: 'wrap' }}>
          <Box
            sx={{
              width: 40,
              height: 40,
              borderRadius: 2,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              backgroundColor: withAlpha(accent, isDark ? 0.18 : 0.12),
              border: `1px solid ${withAlpha(accent, isDark ? 0.32 : 0.18)}`,
              color: accent,
              flexShrink: 0
            }}
          >
            <IconChartAreaLine size={22} stroke={1.7} />
          </Box>
          <Box sx={{ minWidth: 0 }}>
            <Typography variant="h5" sx={{ fontWeight: 600, color: primaryText }}>
              {t('dashboard.default.trafficDaily.title')}
            </Typography>
            <Typography variant="caption" sx={{ color: secondaryText }}>
              {t('dashboard.default.trafficDaily.subtitle')}
            </Typography>
          </Box>

          <Stack direction="row" spacing={1} alignItems="center" sx={{ ml: 'auto' }}>
            <ToggleButtonGroup
              exclusive
              size="small"
              value={rangeDays}
              onChange={(_, nextValue) => {
                if (nextValue !== null) setRangeDays(nextValue);
              }}
              sx={{
                '& .MuiToggleButton-root': {
                  px: 1.1,
                  py: 0.35,
                  fontSize: '0.72rem',
                  color: secondaryText,
                  borderColor: withAlpha(chartPalette.divider, isDark ? 0.6 : 0.8),
                  '&.Mui-selected': {
                    color: accent,
                    bgcolor: withAlpha(accent, isDark ? 0.2 : 0.1)
                  }
                }
              }}
            >
              {rangeOptions.map((range) => (
                <ToggleButton key={range.value} value={range.value}>
                  {range.label}
                </ToggleButton>
              ))}
            </ToggleButtonGroup>
            <Tooltip arrow title={t('dashboard.default.trafficDaily.refresh')}>
              <span>
                <IconButton size="small" onClick={load} disabled={loading} sx={{ color: secondaryText }}>
                  <RefreshIcon fontSize="small" />
                </IconButton>
              </span>
            </Tooltip>
          </Stack>
        </Box>

        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1.25, mb: 2 }}>
          {summaryChips.map((chip) => (
            <Box
              key={chip.key}
              sx={{
                flex: '1 1 160px',
                p: 1.25,
                borderRadius: 2.5,
                bgcolor: nestedPanelSurface,
                border: `1px solid ${withAlpha(chartPalette.divider, isDark ? 0.6 : 0.8)}`
              }}
            >
              <Typography variant="caption" sx={{ color: tertiaryText, display: 'block', mb: 0.25 }}>
                {chip.label}
              </Typography>
              {loading ? (
                <Skeleton variant="text" width={70} height={22} />
              ) : (
                <Typography variant="subtitle1" sx={{ fontWeight: 700, color: chip.color, lineHeight: 1.3 }} noWrap>
                  {chip.value}
                </Typography>
              )}
            </Box>
          ))}
        </Box>

        <Box
          sx={{
            position: 'relative',
            borderRadius: 2.5,
            bgcolor: mutedPanelSurface,
            border: `1px solid ${withAlpha(chartPalette.divider, isDark ? 0.55 : 0.7)}`
          }}
        >
          <Box ref={containerRef} sx={{ width: '100%', height: { xs: 240, md: 280 } }} />
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
                bgcolor: withAlpha(chartPalette.background?.paper || '#ffffff', 0.86)
              }}
            >
              {loading ? (
                <>
                  <CircularProgress size={22} />
                  <Typography variant="caption" sx={{ color: secondaryText }}>
                    {t('dashboard.default.trafficDaily.loading')}
                  </Typography>
                </>
              ) : (
                <Typography variant="body2" sx={{ color: secondaryText }}>
                  {error || t('dashboard.default.trafficDaily.empty')}
                </Typography>
              )}
            </Box>
          )}
        </Box>

        <Typography variant="caption" sx={{ display: 'block', mt: 1, color: tertiaryText }}>
          {t('dashboard.default.trafficDaily.hint')}
        </Typography>
      </CardContent>
    </Card>
  );
}
