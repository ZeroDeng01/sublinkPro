import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

// material-ui
import { useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Skeleton from '@mui/material/Skeleton';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

// project imports
import { getAirportUsageDaily } from 'api/usageHistory';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { withAlpha } from 'utils/colorUtils';
import { formatBytes } from 'views/airports/utils';

// assets
import { IconActivityHeartbeat } from '@tabler/icons-react';

// 用量只会在订阅刷新时变化，轮询保持在较低频率即可
const REFRESH_INTERVAL = 120000;

// ==============================|| TODAY TRAFFIC BADGE ||============================== //

/**
 * 头部今日流量徽标：展示所有机场当天累计消耗的流量。
 */
export default function TodayTrafficBadge() {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();

  const [loading, setLoading] = useState(true);
  const [todayUsed, setTodayUsed] = useState(0);
  const [failed, setFailed] = useState(false);

  const load = useCallback(async () => {
    try {
      const response = await getAirportUsageDaily({ days: 1 });
      const used = Number(response?.data?.todayUsed);
      setTodayUsed(Number.isFinite(used) && used > 0 ? used : 0);
      setFailed(false);
    } catch {
      setFailed(true);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    let cancelled = false;
    const run = async () => {
      if (document.visibilityState === 'hidden') return;
      if (cancelled) return;
      await load();
    };

    run();
    const timer = setInterval(run, REFRESH_INTERVAL);
    // 页面重新可见时立即刷新，避免切回标签页后看到过期数值
    const handleVisible = () => {
      if (document.visibilityState === 'visible') load();
    };
    document.addEventListener('visibilitychange', handleVisible);

    return () => {
      cancelled = true;
      clearInterval(timer);
      document.removeEventListener('visibilitychange', handleVisible);
    };
  }, [load]);

  const accent = theme.palette.info.main;
  const textColor = isDark ? withAlpha(theme.palette.text.primary, 0.94) : withAlpha(accent, 0.96);

  const tooltip = failed ? t('todayTraffic.loadFailed') : t('todayTraffic.tooltip', { used: formatBytes(todayUsed) });

  return (
    <Tooltip arrow title={tooltip}>
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          gap: 0.6,
          height: 28,
          px: 1,
          borderRadius: 1.5,
          bgcolor: withAlpha(accent, isDark ? 0.18 : 0.09),
          border: `1px solid ${withAlpha(accent, isDark ? 0.32 : 0.18)}`,
          color: textColor,
          cursor: 'default',
          userSelect: 'none',
          flexShrink: 0
        }}
      >
        <IconActivityHeartbeat size={14} stroke={1.8} />
        {loading ? (
          <Skeleton variant="text" width={34} height={14} />
        ) : (
          <Typography variant="caption" sx={{ fontWeight: 700, lineHeight: 1, whiteSpace: 'nowrap', color: 'inherit' }}>
            {failed ? '—' : formatBytes(todayUsed, 1)}
          </Typography>
        )}
      </Box>
    </Tooltip>
  );
}
