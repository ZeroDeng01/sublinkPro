import { useState, useEffect, useCallback, useRef } from 'react';

// material-ui
import { useTheme, alpha, keyframes } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import CircularProgress from '@mui/material/CircularProgress';
import LinearProgress from '@mui/material/LinearProgress';
import Chip from '@mui/material/Chip';
import Skeleton from '@mui/material/Skeleton';

// icons
import RefreshIcon from '@mui/icons-material/Refresh';
import MemoryIcon from '@mui/icons-material/Memory';
import DeveloperBoardIcon from '@mui/icons-material/DeveloperBoard';
import AccountTreeIcon from '@mui/icons-material/AccountTree';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import AutorenewIcon from '@mui/icons-material/Autorenew';
import ComputerIcon from '@mui/icons-material/Computer';
import StorageIcon from '@mui/icons-material/Storage';
import SettingsSuggestIcon from '@mui/icons-material/SettingsSuggest';
import KeyIcon from '@mui/icons-material/Key';

// project imports
import { getSystemStats } from 'api/monitor';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

// ==============================|| 动画定义 ||============================== //

const rotate = keyframes`
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
`;

const AUTO_REFRESH_INTERVAL_MS = 1000;
const AUTO_REFRESH_INTERVAL_LABEL = `${AUTO_REFRESH_INTERVAL_MS / 1000}秒`;

// ==============================|| 工具函数 ||============================== //

// 格式化字节数
const formatBytes = (bytes, decimals = 2) => {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
};

// 格式化运行时间
const formatUptime = (seconds) => {
  if (!seconds) return '0秒';
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (days > 0) return `${days}天${hours}时${minutes}分`;
  if (hours > 0) return `${hours}时${minutes}分${secs}秒`;
  if (minutes > 0) return `${minutes}分${secs}秒`;
  return `${secs}秒`;
};

const useMonitorThemeTokens = () => {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const palette = theme.vars?.palette || theme.palette;
  const surfaceTokens = getSurfaceTokens(theme, isDark);
  const textTokens = getReadableTextTokens(theme, isDark);

  const cardSurface = isDark ? surfaceTokens.dialogSurface : withAlpha(surfaceTokens.dialogSurface, 0.98);
  const elevatedSurface = isDark ? withAlpha(palette.background.paper, 0.82) : withAlpha(surfaceTokens.dialogSurface, 0.96);
  const mutedPanelSurface = isDark ? surfaceTokens.mutedPanelSurface : withAlpha(surfaceTokens.mutedPanelSurface, 0.72);
  const nestedPanelSurface = isDark ? withAlpha(palette.background.paper, 0.46) : withAlpha(surfaceTokens.nestedPanelSurface, 0.92);
  const panelBorder = isDark ? withAlpha(palette.divider, 0.78) : surfaceTokens.panelBorder;
  const softBorder = isDark ? withAlpha(palette.divider, 0.62) : withAlpha(palette.divider, 0.82);
  const primaryText = isDark ? withAlpha(textTokens.primaryText, 0.98) : textTokens.primaryText;
  const secondaryText = isDark ? withAlpha(textTokens.secondaryText, 0.96) : textTokens.secondaryText;
  const tertiaryText = isDark ? withAlpha(textTokens.secondaryText, 0.86) : withAlpha(textTokens.primaryText, 0.62);

  return {
    theme,
    isDark,
    palette,
    cardSurface,
    elevatedSurface,
    mutedPanelSurface,
    nestedPanelSurface,
    panelBorder,
    softBorder,
    primaryText,
    secondaryText,
    tertiaryText,
    topCardBackground: isDark ? `linear-gradient(180deg, ${elevatedSurface} 0%, ${cardSurface} 100%)` : surfaceTokens.dialogSurfaceGradient,
    insetPanelBackground: `linear-gradient(180deg, ${nestedPanelSurface} 0%, ${mutedPanelSurface} 100%)`,
    subtleSurface: isDark ? withAlpha(palette.background.paper, 0.38) : withAlpha(palette.background.paper, 0.88),
    iconButtonSurface: isDark ? withAlpha(palette.background.paper, 0.5) : withAlpha(palette.background.default, 0.88),
    iconButtonHover: isDark ? withAlpha(palette.background.paper, 0.72) : withAlpha(palette.background.paper, 0.98),
    skeletonBase: isDark ? withAlpha(palette.background.paper, 0.52) : withAlpha(palette.background.default, 0.88),
    skeletonHighlight: isDark ? withAlpha(palette.background.paper, 0.74) : withAlpha(palette.background.paper, 0.98)
  };
};

// ==============================|| 指标卡片组件 ||============================== //

const MetricCard = ({ title, value, subValue, icon: Icon, color, progress, children }) => {
  const { theme, isDark, palette, nestedPanelSurface, mutedPanelSurface, secondaryText } = useMonitorThemeTokens();

  return (
    <Box
      sx={{
        p: 2,
        borderRadius: 3,
        backgroundColor: mutedPanelSurface,
        backgroundImage: `linear-gradient(180deg, ${alpha(color, isDark ? 0.12 : 0.05)} 0%, ${nestedPanelSurface} 100%)`,
        border: `1px solid ${alpha(color, isDark ? 0.24 : 0.14)}`,
        boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.035)}` : 'none',
        height: '100%',
        display: 'flex',
        flexDirection: 'column'
      }}
    >
      {/* 标题行 */}
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
        <Box
          sx={{
            width: 32,
            height: 32,
            borderRadius: 1.5,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: `linear-gradient(135deg, ${color} 0%, ${alpha(color, 0.7)} 100%)`,
            color: palette.common?.white || theme.palette.common.white
          }}
        >
          <Icon sx={{ fontSize: 18, color: 'inherit' }} />
        </Box>
        <Typography
          variant="body2"
          sx={{
            fontWeight: 600,
            color: secondaryText,
            fontSize: '0.75rem',
            textTransform: 'uppercase',
            letterSpacing: 0.5
          }}
        >
          {title}
        </Typography>
      </Box>

      {/* 主数值 */}
      <Typography
        variant="h4"
        sx={{
          fontWeight: 700,
          color: color,
          mb: 0.5,
          fontSize: '1.5rem',
          lineHeight: 1.2
        }}
      >
        {value}
      </Typography>

      {/* 副数值 */}
      {subValue && (
        <Typography
          variant="caption"
          sx={{
            color: secondaryText,
            fontSize: '0.7rem',
            mb: 1
          }}
        >
          {subValue}
        </Typography>
      )}

      {/* 进度条 */}
      {progress !== undefined && (
        <Box sx={{ mt: 'auto' }}>
          <LinearProgress
            variant="determinate"
            value={Math.min(progress, 100)}
            sx={{
              height: 6,
              borderRadius: 3,
              bgcolor: isDark
                ? withAlpha(theme.vars?.palette?.background?.default || theme.palette.background.default, 0.92)
                : alpha(color, 0.1),
              '& .MuiLinearProgress-bar': {
                borderRadius: 3,
                background: `linear-gradient(90deg, ${color} 0%, ${alpha(color, 0.7)} 100%)`
              }
            }}
          />
          <Typography
            variant="caption"
            sx={{
              color: secondaryText,
              fontSize: '0.65rem',
              mt: 0.5,
              display: 'block',
              textAlign: 'right'
            }}
          >
            {progress.toFixed(1)}%
          </Typography>
        </Box>
      )}

      {/* 子内容 */}
      {children}
    </Box>
  );
};

// ==============================|| 圆形进度指示器 ||============================== //

const CircularMetric = ({ value, maxValue, label, color, size = 80 }) => {
  const { isDark, secondaryText } = useMonitorThemeTokens();
  const percentage = maxValue > 0 ? (value / maxValue) * 100 : 0;

  return (
    <Box sx={{ position: 'relative', display: 'inline-flex', flexDirection: 'column', alignItems: 'center' }}>
      <Box sx={{ position: 'relative' }}>
        <CircularProgress variant="determinate" value={100} size={size} thickness={4} sx={{ color: alpha(color, isDark ? 0.18 : 0.12) }} />
        <CircularProgress
          variant="determinate"
          value={Math.min(percentage, 100)}
          size={size}
          thickness={4}
          sx={{
            color: color,
            position: 'absolute',
            left: 0,
            '& .MuiCircularProgress-circle': {
              strokeLinecap: 'round'
            }
          }}
        />
        <Box
          sx={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}
        >
          <Typography
            variant="caption"
            sx={{
              fontWeight: 700,
              color: color,
              fontSize: size * 0.18
            }}
          >
            {percentage.toFixed(0)}%
          </Typography>
        </Box>
      </Box>
      <Typography
        variant="caption"
        sx={{
          mt: 1,
          color: secondaryText,
          fontSize: '0.65rem',
          textAlign: 'center'
        }}
      >
        {label}
      </Typography>
    </Box>
  );
};

const ConfigItemCard = ({ item, masked = false }) => {
  const { theme, isDark, nestedPanelSurface, mutedPanelSurface, panelBorder, primaryText, secondaryText } = useMonitorThemeTokens();
  const maskedTextColor = isDark ? theme.palette.warning.light : theme.palette.warning.dark;

  return (
    <Box
      sx={{
        p: 1.5,
        borderRadius: 2.5,
        height: '100%',
        backgroundColor: mutedPanelSurface,
        backgroundImage: `linear-gradient(180deg, ${nestedPanelSurface} 0%, ${mutedPanelSurface} 100%)`,
        border: `1px solid ${panelBorder}`,
        boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.025)}` : 'none'
      }}
    >
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1, mb: 1 }}>
        <Typography variant="body2" sx={{ fontWeight: 600, minWidth: 0, color: primaryText }}>
          {item.label}
        </Typography>
        {masked && (
          <Chip
            size="small"
            label="已隐藏"
            sx={{
              height: 20,
              fontSize: '0.65rem',
              bgcolor: isDark ? withAlpha(theme.palette.warning.main, 0.16) : withAlpha(theme.palette.warning.main, 0.1),
              color: maskedTextColor,
              border: `1px solid ${withAlpha(maskedTextColor, isDark ? 0.32 : 0.2)}`,
              fontWeight: 700
            }}
          />
        )}
      </Box>

      <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75, flexWrap: 'wrap', mb: 1 }}>
        <Chip
          size="small"
          icon={<KeyIcon sx={{ fontSize: 14 }} />}
          label={item.key}
          sx={{
            height: 22,
            fontSize: '0.68rem',
            bgcolor: withAlpha(theme.palette.primary.main, isDark ? 0.16 : 0.08),
            color: theme.palette.primary.main,
            border: `1px solid ${withAlpha(theme.palette.primary.main, isDark ? 0.28 : 0.16)}`,
            '& .MuiChip-icon': {
              color: 'inherit',
              ml: 0.6
            }
          }}
        />
      </Box>

      <Tooltip title={item.value} arrow placement="top-start">
        <Typography
          variant="body2"
          sx={{
            fontWeight: 600,
            color: masked ? maskedTextColor : primaryText,
            lineHeight: 1.6,
            wordBreak: 'break-all'
          }}
        >
          {item.value}
        </Typography>
      </Tooltip>

      {item.env && (
        <Typography
          variant="caption"
          sx={{
            mt: 1,
            display: 'block',
            color: secondaryText,
            fontSize: '0.68rem',
            lineHeight: 1.45,
            wordBreak: 'break-all'
          }}
        >
          环境变量：{item.env}
        </Typography>
      )}
    </Box>
  );
};

// ==============================|| 系统监控卡片主组件 ||============================== //

const SystemMonitorCard = () => {
  const {
    theme,
    isDark,
    palette,
    topCardBackground,
    insetPanelBackground,
    cardSurface,
    elevatedSurface,
    panelBorder,
    softBorder,
    primaryText,
    secondaryText,
    tertiaryText,
    subtleSurface,
    iconButtonSurface,
    iconButtonHover,
    skeletonBase,
    skeletonHighlight
  } = useMonitorThemeTokens();
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const intervalRef = useRef(null);

  // 获取系统统计数据
  const fetchStats = useCallback(async (isManual = false) => {
    try {
      if (isManual) setRefreshing(true);
      const res = await getSystemStats();
      if (res.data) {
        setStats(res.data);
      }
    } catch (error) {
      console.error('获取系统状态失败:', error);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  // 手动刷新
  const handleRefresh = () => {
    fetchStats(true);
  };

  // 切换自动刷新
  const toggleAutoRefresh = () => {
    setAutoRefresh(!autoRefresh);
  };

  // 初始化和自动刷新
  useEffect(() => {
    fetchStats();

    if (autoRefresh) {
      intervalRef.current = setInterval(() => {
        fetchStats();
      }, AUTO_REFRESH_INTERVAL_MS);
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [fetchStats, autoRefresh]);

  // 颜色定义
  const colors = {
    memory: '#6366f1',
    cpu: '#10b981',
    goroutine: '#f59e0b',
    uptime: '#06b6d4',
    gc: '#ec4899'
  };

  const configItems = stats?.runtime_config
    ? [
        ...(stats.runtime_config.safe_to_show || []).map((item) => ({ ...item, masked: false })),
        ...(stats.runtime_config.masked_summary || []).map((item) => ({ ...item, masked: true }))
      ]
    : [];

  return (
    <Card
      sx={{
        mb: 3,
        borderRadius: 4,
        backgroundColor: cardSurface,
        backgroundImage: topCardBackground,
        backdropFilter: isDark ? 'blur(16px)' : 'none',
        border: `1px solid ${panelBorder}`,
        boxShadow: isDark
          ? `0 16px 36px ${alpha(theme.palette.common.black, 0.2)}, inset 0 1px 0 ${alpha(theme.palette.common.white, 0.035)}`
          : `0 8px 24px ${alpha(theme.palette.common.black, 0.06)}`,
        overflow: 'hidden',
        position: 'relative'
      }}
    >
      {/* 顶部装饰条 */}
      <Box
        sx={{
          height: 3,
          background: `linear-gradient(90deg, ${colors.memory} 0%, ${colors.cpu} 25%, ${colors.goroutine} 50%, ${colors.uptime} 75%, ${colors.gc} 100%)`
        }}
      />

      <CardContent sx={{ p: 3 }}>
        {/* 标题栏 */}
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 3 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
            <Box
              sx={{
                width: 40,
                height: 40,
                borderRadius: 2,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: `linear-gradient(135deg, ${colors.memory} 0%, ${colors.cpu} 100%)`,
                color: palette.common?.white || theme.palette.common.white
              }}
            >
              <ComputerIcon sx={{ color: 'inherit', fontSize: 22 }} />
            </Box>
            <Box>
              <Typography variant="h5" sx={{ fontWeight: 600, color: primaryText }}>
                系统监控
              </Typography>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                {stats && (
                  <Chip
                    size="small"
                    label={`${stats.goos} / ${stats.goarch}`}
                    sx={{
                      height: 20,
                      fontSize: '0.65rem',
                      bgcolor: alpha(colors.cpu, isDark ? 0.18 : 0.1),
                      color: colors.cpu,
                      border: `1px solid ${alpha(colors.cpu, isDark ? 0.3 : 0.2)}`
                    }}
                  />
                )}
                {stats && (
                  <Chip
                    size="small"
                    label={stats.go_version}
                    sx={{
                      height: 20,
                      fontSize: '0.65rem',
                      bgcolor: alpha(colors.memory, isDark ? 0.18 : 0.1),
                      color: colors.memory,
                      border: `1px solid ${alpha(colors.memory, isDark ? 0.3 : 0.2)}`
                    }}
                  />
                )}
              </Box>
            </Box>
          </Box>

          {/* 控制按钮 */}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            {/* 自动刷新指示器 */}
            <Tooltip title={autoRefresh ? '关闭自动刷新' : `开启自动刷新 (${AUTO_REFRESH_INTERVAL_LABEL})`} arrow>
              <IconButton
                onClick={toggleAutoRefresh}
                size="small"
                sx={{
                  bgcolor: autoRefresh ? withAlpha(colors.cpu, isDark ? 0.18 : 0.1) : iconButtonSurface,
                  color: autoRefresh ? colors.cpu : secondaryText,
                  border: '1px solid',
                  borderColor: autoRefresh ? withAlpha(colors.cpu, isDark ? 0.32 : 0.18) : softBorder,
                  transition: 'background-color 0.2s ease, border-color 0.2s ease, transform 0.2s ease, color 0.2s ease',
                  '&:hover': {
                    bgcolor: autoRefresh ? withAlpha(colors.cpu, isDark ? 0.24 : 0.14) : iconButtonHover,
                    borderColor: autoRefresh ? withAlpha(colors.cpu, isDark ? 0.4 : 0.24) : panelBorder
                  },
                  '&:active': {
                    transform: 'scale(0.96)'
                  }
                }}
              >
                <AutorenewIcon
                  sx={{
                    fontSize: 20,
                    animation: autoRefresh ? `${rotate} 3s linear infinite` : 'none'
                  }}
                />
              </IconButton>
            </Tooltip>

            {/* 手动刷新 */}
            <Tooltip title="刷新" arrow>
              <IconButton
                onClick={handleRefresh}
                disabled={refreshing}
                size="small"
                sx={{
                  bgcolor: refreshing ? subtleSurface : withAlpha(theme.palette.primary.main, isDark ? 0.16 : 0.1),
                  color: refreshing ? tertiaryText : theme.palette.primary.main,
                  border: '1px solid',
                  borderColor: refreshing ? softBorder : withAlpha(theme.palette.primary.main, isDark ? 0.28 : 0.16),
                  transition: 'all 0.3s ease',
                  '&:hover': {
                    bgcolor: withAlpha(theme.palette.primary.main, isDark ? 0.22 : 0.14),
                    borderColor: withAlpha(theme.palette.primary.main, isDark ? 0.36 : 0.22),
                    transform: 'rotate(180deg)'
                  },
                  '&:active': {
                    transform: 'scale(0.96)'
                  },
                  '&.Mui-disabled': {
                    bgcolor: subtleSurface,
                    color: tertiaryText,
                    borderColor: softBorder
                  }
                }}
              >
                {refreshing ? (
                  <CircularProgress size={20} sx={{ color: theme.palette.primary.main }} />
                ) : (
                  <RefreshIcon sx={{ fontSize: 20 }} />
                )}
              </IconButton>
            </Tooltip>
          </Box>
        </Box>

        {/* 加载状态 */}
        {loading ? (
          <Grid container spacing={2}>
            {[1, 2, 3, 4].map((i) => (
              <Grid key={i} size={{ xs: 6, sm: 3 }}>
                <Skeleton
                  variant="rounded"
                  height={140}
                  sx={{
                    borderRadius: 3,
                    bgcolor: skeletonBase,
                    '&::after': {
                      background: `linear-gradient(90deg, transparent, ${skeletonHighlight}, transparent)`
                    }
                  }}
                />
              </Grid>
            ))}
          </Grid>
        ) : stats ? (
          <>
            {/* 主要指标 */}
            <Grid container spacing={2} sx={{ mb: 2 }}>
              {/* 内存使用 */}
              <Grid size={{ xs: 6, sm: 6, md: 3 }}>
                <MetricCard
                  title="堆内存"
                  value={formatBytes(stats.heap_inuse)}
                  subValue={`共申请 ${formatBytes(stats.heap_sys)}`}
                  icon={MemoryIcon}
                  color={colors.memory}
                  progress={stats.memory_usage}
                />
              </Grid>

              {/* CPU信息 */}
              <Grid size={{ xs: 6, sm: 6, md: 3 }}>
                <MetricCard
                  title="CPU"
                  value={`${stats.num_cpu} 核`}
                  subValue={`GOMAXPROCS: ${stats.gomaxprocs}`}
                  icon={DeveloperBoardIcon}
                  color={colors.cpu}
                  progress={stats.cpu_usage}
                />
              </Grid>

              {/* Goroutine数量 */}
              <Grid size={{ xs: 6, sm: 6, md: 3 }}>
                <MetricCard
                  title="Goroutines"
                  value={stats.num_goroutine}
                  subValue={`CGO调用: ${stats.num_cgo_call}`}
                  icon={AccountTreeIcon}
                  color={colors.goroutine}
                />
              </Grid>

              {/* 运行时间 */}
              <Grid size={{ xs: 6, sm: 6, md: 3 }}>
                <MetricCard
                  title="运行时间"
                  value={formatUptime(stats.uptime)}
                  subValue={`启动: ${new Date(stats.start_time * 1000).toLocaleString()}`}
                  icon={AccessTimeIcon}
                  color={colors.uptime}
                />
              </Grid>
            </Grid>

            {/* 详细内存统计 */}
            <Box
              sx={{
                p: 2,
                borderRadius: 3,
                backgroundColor: elevatedSurface,
                backgroundImage: insetPanelBackground,
                border: `1px solid ${panelBorder}`,
                boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.025)}` : 'none'
              }}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
                <StorageIcon sx={{ fontSize: 18, color: colors.gc }} />
                <Typography variant="body2" sx={{ fontWeight: 600, fontSize: '0.8rem', color: primaryText }}>
                  运行时详情
                </Typography>
              </Box>

              <Grid container spacing={2}>
                {/* 内存详情 */}
                <Grid size={{ xs: 12, sm: 6, md: 4 }}>
                  <Box sx={{ display: 'flex', justifyContent: 'center', gap: 3, flexWrap: 'wrap' }}>
                    <CircularMetric value={stats.heap_inuse} maxValue={stats.sys} label="堆内存使用" color={colors.memory} size={70} />
                    <CircularMetric value={stats.stack_inuse} maxValue={stats.sys} label="栈内存使用" color={colors.cpu} size={70} />
                  </Box>
                </Grid>

                {/* GC统计 */}
                <Grid size={{ xs: 12, sm: 6, md: 4 }}>
                  <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Typography variant="caption" sx={{ color: secondaryText }}>
                        GC 次数
                      </Typography>
                      <Typography variant="body2" sx={{ fontWeight: 600, color: colors.gc }}>
                        {stats.num_gc}
                      </Typography>
                    </Box>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Typography variant="caption" sx={{ color: secondaryText }}>
                        GC 暂停时间
                      </Typography>
                      <Typography variant="body2" sx={{ fontWeight: 600, color: colors.gc }}>
                        {(stats.pause_total_ns / 1e6).toFixed(2)} ms
                      </Typography>
                    </Box>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Typography variant="caption" sx={{ color: secondaryText }}>
                        GC CPU 占用
                      </Typography>
                      <Typography variant="body2" sx={{ fontWeight: 600, color: colors.gc }}>
                        {(stats.gc_cpu_frac * 100).toFixed(3)}%
                      </Typography>
                    </Box>
                  </Box>
                </Grid>

                {/* 内存分配 */}
                <Grid size={{ xs: 12, sm: 12, md: 4 }}>
                  <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Typography variant="caption" sx={{ color: secondaryText }}>
                        系统内存获取
                      </Typography>
                      <Typography variant="body2" sx={{ fontWeight: 600, color: colors.memory }}>
                        {formatBytes(stats.sys)}
                      </Typography>
                    </Box>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Typography variant="caption" sx={{ color: secondaryText }}>
                        累计分配内存
                      </Typography>
                      <Typography variant="body2" sx={{ fontWeight: 600, color: colors.memory }}>
                        {formatBytes(stats.total_alloc)}
                      </Typography>
                    </Box>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Typography variant="caption" sx={{ color: secondaryText }}>
                        栈内存使用
                      </Typography>
                      <Typography variant="body2" sx={{ fontWeight: 600, color: colors.cpu }}>
                        {formatBytes(stats.stack_inuse)}
                      </Typography>
                    </Box>
                  </Box>
                </Grid>
              </Grid>
            </Box>

            {configItems.length > 0 && (
              <Box
                sx={{
                  mt: 2,
                  p: 2,
                  borderRadius: 3,
                  backgroundColor: elevatedSurface,
                  backgroundImage: insetPanelBackground,
                  border: `1px solid ${panelBorder}`,
                  boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.025)}` : 'none'
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
                  <SettingsSuggestIcon sx={{ fontSize: 18, color: colors.uptime }} />
                  <Typography variant="body2" sx={{ fontWeight: 600, fontSize: '0.8rem', color: primaryText }}>
                    运行配置参数
                  </Typography>
                </Box>
                <Grid container spacing={1.5}>
                  {configItems.map((item) => (
                    <Grid key={`${item.key}-${item.env || 'default'}`} size={{ xs: 12, sm: 6, md: 4 }}>
                      <ConfigItemCard item={item} masked={item.masked} />
                    </Grid>
                  ))}
                </Grid>
              </Box>
            )}
          </>
        ) : (
          <Box sx={{ textAlign: 'center', py: 4 }}>
            <Typography sx={{ color: secondaryText }}>无法获取系统状态</Typography>
          </Box>
        )}
      </CardContent>
    </Card>
  );
};

export default SystemMonitorCard;
