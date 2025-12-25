import { useState } from 'react';
import PropTypes from 'prop-types';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Snackbar from '@mui/material/Snackbar';
import Stack from '@mui/material/Stack';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

// icons
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import SyncIcon from '@mui/icons-material/Sync';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import ScheduleIcon from '@mui/icons-material/Schedule';
import UpdateIcon from '@mui/icons-material/Update';
import SpeedIcon from '@mui/icons-material/Speed';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';
import EventIcon from '@mui/icons-material/Event';

// utils
import { formatDateTime, formatBytes, formatExpireTime, getUsageColor } from '../utils';

// local components
import AirportNodeStatsCard from './AirportNodeStatsCard';
import AirportLogo from './AirportLogo';

/**
 * 机场列表卡片网格组件（桌面端）
 */
export default function AirportTable({ airports, onEdit, onDelete, onPull, onRefreshUsage }) {
  const theme = useTheme();

  // 复制提示状态
  const [copyTip, setCopyTip] = useState({ open: false, name: '' });

  // 复制订阅地址
  const handleCopyUrl = async (airport) => {
    try {
      await navigator.clipboard.writeText(airport.url);
      setCopyTip({ open: true, name: airport.name });
      setTimeout(() => setCopyTip({ open: false, name: '' }), 2000);
    } catch (err) {
      console.error('复制失败:', err);
    }
  };

  // 根据使用率计算进度条渐变色
  const getProgressGradient = (percent) => {
    if (percent < 60) return `linear-gradient(90deg, ${theme.palette.success.light}, ${theme.palette.success.main})`;
    if (percent < 85) return `linear-gradient(90deg, ${theme.palette.warning.light}, ${theme.palette.warning.main})`;
    return `linear-gradient(90deg, ${theme.palette.error.light}, ${theme.palette.error.main})`;
  };

  /**
   * 根据延迟值获取颜色
   * <100ms 绿色（优秀）
   * <300ms 橙色（良好）
   * >=300ms 红色（较差）
   */
  const getDelayColor = (delay) => {
    if (delay < 150) return theme.palette.success.main;
    if (delay < 300) return '#ec8f04ff'; // 深橙色
    return theme.palette.error.main;
  };

  /**
   * 根据速度值获取颜色
   * >10MB/s 绿色（优秀）
   * >3MB/s 橙色（良好）
   * <=3MB/s 红色（较差）
   */
  const getSpeedColor = (speed) => {
    if (speed > 10) return theme.palette.success.main;
    if (speed > 5) return '#ec8f04ff'; // 深橙色
    return theme.palette.error.main;
  };

  /**
   * 计算顶部状态条颜色
   * 优先级：禁用(灰色) > 用量警告(红色) > 过期警告(红色) > 启用(绿色)
   */
  const getStatusBarColor = (airport) => {
    // 禁用状态
    if (!airport.enabled) {
      return `linear-gradient(90deg, ${theme.palette.grey[400]}, ${theme.palette.grey[300]})`;
    }

    // 检查用量警告（使用率 >= 85%）
    if (airport.fetchUsageInfo && airport.usageTotal > 0) {
      const upload = airport.usageUpload || 0;
      const download = airport.usageDownload || 0;
      const used = upload + download;
      const percent = (used / airport.usageTotal) * 100;
      if (percent >= 85) {
        return `linear-gradient(90deg, ${theme.palette.error.main}, ${theme.palette.error.light})`;
      }
    }

    // 检查过期警告（7天内过期）
    if (airport.usageExpire > 0) {
      const now = Math.floor(Date.now() / 1000);
      const daysLeft = (airport.usageExpire - now) / (24 * 60 * 60);
      if (daysLeft <= 7) {
        return `linear-gradient(90deg, ${theme.palette.error.main}, ${theme.palette.error.light})`;
      }
    }

    // 正常启用状态
    return `linear-gradient(90deg, ${theme.palette.success.main}, ${theme.palette.success.light})`;
  };

  // 渲染用量信息
  const renderUsageInfo = (airport) => {
    // 未开启获取用量信息
    if (!airport.fetchUsageInfo) {
      return (
        <Typography variant="caption" color="textSecondary">
          未开启用量获取
        </Typography>
      );
    }

    // usageTotal 为 -1 表示获取失败
    if (airport.usageTotal === -1) {
      return (
        <Typography variant="body2" sx={{ color: 'error.main', fontWeight: 500 }}>
          用量获取失败
        </Typography>
      );
    }

    // usageTotal 为 0 或未设置，表示尚未获取
    if (!airport.usageTotal || airport.usageTotal === 0) {
      return (
        <Typography variant="body2" color="textSecondary">
          待获取
        </Typography>
      );
    }

    const upload = airport.usageUpload || 0;
    const download = airport.usageDownload || 0;
    const used = upload + download;
    const total = airport.usageTotal;
    const percent = Math.min((used / total) * 100, 100);
    const color = getUsageColor(percent);

    return (
      <Box>
        {/* 已用/总量 + 百分比 */}
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 0.5 }}>
          <Typography variant="body2" sx={{ fontWeight: 600, color: 'text.primary' }}>
            {formatBytes(used)} / {formatBytes(total)}
          </Typography>
          <Typography variant="body2" sx={{ fontWeight: 700, color: color }}>
            {percent.toFixed(1)}%
          </Typography>
        </Box>

        {/* 进度条 */}
        <Box
          sx={{
            height: 6,
            borderRadius: 3,
            backgroundColor: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.08)',
            overflow: 'hidden',
            mb: 0.5
          }}
        >
          <Box
            sx={{
              width: `${percent}%`,
              height: '100%',
              borderRadius: 3,
              background: getProgressGradient(percent),
              transition: 'width 0.3s ease'
            }}
          />
        </Box>

        {/* 上传/下载 */}
        <Box sx={{ display: 'flex', gap: 2, fontSize: '0.75rem' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <Typography variant="caption" sx={{ color: 'success.main', fontWeight: 500 }}>
              ↑
            </Typography>
            <Typography variant="caption" sx={{ color: 'text.secondary' }}>
              {formatBytes(upload)}
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <Typography variant="caption" sx={{ color: 'info.main', fontWeight: 500 }}>
              ↓
            </Typography>
            <Typography variant="caption" sx={{ color: 'text.secondary' }}>
              {formatBytes(download)}
            </Typography>
          </Box>
        </Box>

        {/* 过期时间 */}
        {airport.usageExpire > 0 &&
          (() => {
            const now = Math.floor(Date.now() / 1000);
            const daysLeft = (airport.usageExpire - now) / (24 * 60 * 60);
            const isUrgent = daysLeft <= 7;
            const isWarning = daysLeft <= 30 && daysLeft > 7;

            return (
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mt: 0.5 }}>
                {isUrgent && <WarningAmberIcon sx={{ fontSize: 12, color: 'error.main' }} />}
                {isWarning && <EventIcon sx={{ fontSize: 12, color: 'info.main' }} />}
                <Typography
                  variant="caption"
                  sx={{
                    color: isUrgent ? 'error.main' : isWarning ? 'info.main' : 'text.secondary',
                    fontWeight: isUrgent || isWarning ? 600 : 400
                  }}
                >
                  到期: {formatExpireTime(airport.usageExpire)}
                  {isUrgent && ` (${Math.max(0, Math.ceil(daysLeft))}天)`}
                </Typography>
              </Box>
            );
          })()}
      </Box>
    );
  };

  /**
   * 渲染测速概览信息（紧凑版）
   */
  const renderSpeedSummary = (nodeStats, nodeCount) => {
    const hasData = nodeStats && (nodeStats.delayPassCount > 0 || nodeStats.speedPassCount > 0);

    if (!hasData) {
      return (
        <Typography variant="caption" color="text.disabled" sx={{ fontStyle: 'italic' }}>
          暂未测试
        </Typography>
      );
    }

    return (
      <Tooltip
        arrow
        placement="top"
        componentsProps={{
          tooltip: {
            sx: {
              bgcolor: theme.palette.mode === 'dark' ? 'grey.800' : 'grey.50',
              border: `1px solid ${alpha(theme.palette.divider, 0.1)}`,
              borderRadius: 2,
              boxShadow: theme.shadows[8],
              p: 1.5,
              maxWidth: 320
            }
          }
        }}
        title={<AirportNodeStatsCard nodeStats={nodeStats} nodeCount={nodeCount} />}
      >
        <Box sx={{ cursor: 'help' }}>
          {/* 通过数量 */}
          <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 0.5 }}>
            <Stack direction="row" spacing={0.25} alignItems="center">
              <AccessTimeIcon sx={{ fontSize: 12, color: 'success.main' }} />
              <Typography variant="caption" fontWeight={600} color="success.main">
                {nodeStats.delayPassCount}
              </Typography>
            </Stack>
            <Typography variant="caption" color="text.disabled">
              /
            </Typography>
            <Stack direction="row" spacing={0.25} alignItems="center">
              <SpeedIcon sx={{ fontSize: 12, color: 'info.main' }} />
              <Typography variant="caption" fontWeight={600} color="info.main">
                {nodeStats.speedPassCount}
              </Typography>
            </Stack>
          </Stack>

          {/* 最低延迟和最高速度简要显示 */}
          <Stack spacing={0.25}>
            {nodeStats.lowestDelayTime > 0 && (
              <Typography variant="caption" sx={{ color: 'text.secondary', fontSize: '0.65rem' }}>
                最低延迟:{' '}
                <span style={{ fontWeight: 600, color: getDelayColor(nodeStats.lowestDelayTime) }}>{nodeStats.lowestDelayTime}ms</span>
              </Typography>
            )}
            {nodeStats.highestSpeed > 0 && (
              <Typography variant="caption" sx={{ color: 'text.secondary', fontSize: '0.65rem' }}>
                最高速度:{' '}
                <span style={{ fontWeight: 600, color: getSpeedColor(nodeStats.highestSpeed) }}>
                  {nodeStats.highestSpeed?.toFixed(1)}MB/s
                </span>
              </Typography>
            )}
          </Stack>
        </Box>
      </Tooltip>
    );
  };

  if (airports.length === 0) {
    return (
      <Box sx={{ py: 6, textAlign: 'center' }}>
        <Typography variant="body2" color="textSecondary">
          暂无机场数据，点击上方"添加机场"按钮添加
        </Typography>
      </Box>
    );
  }

  return (
    <>
      <Box
        sx={{
          display: 'grid',
          gridTemplateColumns: {
            xs: '1fr',
            sm: 'repeat(2, 1fr)',
            md: 'repeat(3, 1fr)',
            lg: 'repeat(4, 1fr)',
            xl: 'repeat(5, 1fr)'
          },
          gap: 2
        }}
      >
        {airports.map((airport) => (
          <Card
            key={airport.id}
            sx={{
              display: 'flex',
              flexDirection: 'column',
              borderRadius: 3,
              border: `1px solid ${alpha(theme.palette.divider, 0.12)}`,
              boxShadow:
                theme.palette.mode === 'dark' ? `0 2px 8px ${alpha('#000', 0.25)}` : `0 2px 8px ${alpha(theme.palette.primary.main, 0.06)}`,
              transition: 'all 0.2s ease',
              overflow: 'hidden',
              position: 'relative',
              '&:hover': {
                transform: 'translateY(-2px)',
                boxShadow:
                  theme.palette.mode === 'dark'
                    ? `0 6px 16px ${alpha('#000', 0.35)}`
                    : `0 6px 16px ${alpha(theme.palette.primary.main, 0.12)}`
              },
              // 顶部状态指示条
              '&::before': {
                content: '""',
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                height: 3,
                background: getStatusBarColor(airport)
              }
            }}
          >
            <CardContent sx={{ p: 2, pt: 2, flex: 1, display: 'flex', flexDirection: 'column', '&:last-child': { pb: 2 } }}>
              {/* 头部：Logo、名称、状态 */}
              <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 1.5, mb: 1.5 }}>
                <AirportLogo logo={airport.logo} name={airport.name} size="medium" />
                <Box sx={{ flex: 1, minWidth: 0 }}>
                  <Typography
                    variant="subtitle2"
                    fontWeight={600}
                    sx={{
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                      lineHeight: 1.3
                    }}
                  >
                    {airport.name}
                  </Typography>
                  <Stack direction="row" spacing={0.5} sx={{ mt: 0.5, flexWrap: 'wrap', gap: 0.5 }}>
                    <Chip
                      label={airport.enabled ? '启用' : '禁用'}
                      color={airport.enabled ? 'success' : 'default'}
                      size="small"
                      sx={{ height: 20, fontSize: '0.7rem' }}
                    />
                    {airport.group && (
                      <Chip label={airport.group} variant="outlined" size="small" sx={{ height: 20, fontSize: '0.7rem' }} />
                    )}
                  </Stack>
                </Box>
              </Box>

              {/* 节点和调度信息 */}
              <Stack direction="row" spacing={1} sx={{ mb: 1.5, flexWrap: 'wrap', gap: 0.5 }}>
                <Chip
                  label={`${airport.nodeCount || 0} 节点`}
                  color="primary"
                  variant="outlined"
                  size="small"
                  sx={{ height: 22, fontSize: '0.7rem' }}
                />
                <Chip
                  icon={<AccessTimeIcon sx={{ fontSize: '12px !important' }} />}
                  label={airport.cronExpr}
                  variant="outlined"
                  size="small"
                  sx={{ height: 22, fontSize: '0.7rem' }}
                />
              </Stack>

              {/* 运行时间 */}
              <Box
                sx={{
                  display: 'grid',
                  gridTemplateColumns: '1fr 1fr',
                  gap: 1,
                  mb: 1.5,
                  p: 1,
                  borderRadius: 1.5,
                  bgcolor: alpha(theme.palette.background.default, 0.5),
                  border: `1px solid ${alpha(theme.palette.divider, 0.06)}`
                }}
              >
                <Box>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mb: 0.25 }}>
                    <ScheduleIcon sx={{ fontSize: 12, color: 'text.secondary' }} />
                    <Typography variant="caption" color="textSecondary" sx={{ fontSize: '0.65rem' }}>
                      上次运行
                    </Typography>
                  </Box>
                  <Typography variant="caption" sx={{ fontWeight: 500, fontSize: '0.7rem' }}>
                    {formatDateTime(airport.lastRunTime)}
                  </Typography>
                </Box>
                <Box>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mb: 0.25 }}>
                    <UpdateIcon sx={{ fontSize: 12, color: 'text.secondary' }} />
                    <Typography variant="caption" color="textSecondary" sx={{ fontSize: '0.65rem' }}>
                      下次运行
                    </Typography>
                  </Box>
                  <Typography variant="caption" sx={{ fontWeight: 500, fontSize: '0.7rem' }}>
                    {formatDateTime(airport.nextRunTime)}
                  </Typography>
                </Box>
              </Box>

              {/* 用量信息 */}
              <Box
                sx={{
                  p: 1,
                  borderRadius: 1.5,
                  bgcolor: alpha(theme.palette.primary.main, 0.03),
                  border: `1px solid ${alpha(theme.palette.divider, 0.06)}`,
                  mb: 1.5
                }}
              >
                <Typography
                  variant="caption"
                  color="textSecondary"
                  sx={{ display: 'block', mb: 0.5, fontWeight: 500, fontSize: '0.65rem' }}
                >
                  用量信息
                </Typography>
                {renderUsageInfo(airport)}
              </Box>

              {/* 节点测试统计 */}
              <Box sx={{ mb: 1.5 }}>
                <Typography
                  variant="caption"
                  color="textSecondary"
                  sx={{ display: 'block', mb: 0.5, fontWeight: 500, fontSize: '0.65rem' }}
                >
                  节点测试
                </Typography>
                {renderSpeedSummary(airport.nodeStats, airport.nodeCount || 0)}
              </Box>

              {/* 操作按钮 - 固定在底部 */}
              <Box sx={{ mt: 'auto', pt: 1, borderTop: `1px solid ${alpha(theme.palette.divider, 0.08)}` }}>
                <Box sx={{ display: 'flex', justifyContent: 'center', gap: 0.5 }}>
                  <Tooltip title="复制订阅地址" arrow>
                    <IconButton
                      size="small"
                      onClick={() => handleCopyUrl(airport)}
                      sx={{
                        bgcolor: alpha(theme.palette.secondary.main, 0.08),
                        color: theme.palette.secondary.main,
                        '&:hover': { bgcolor: alpha(theme.palette.secondary.main, 0.15) }
                      }}
                    >
                      <ContentCopyIcon sx={{ fontSize: 16 }} />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="立即拉取" arrow>
                    <IconButton
                      size="small"
                      onClick={() => onPull(airport)}
                      sx={{
                        bgcolor: alpha(theme.palette.primary.main, 0.08),
                        color: theme.palette.primary.main,
                        '&:hover': { bgcolor: alpha(theme.palette.primary.main, 0.15) }
                      }}
                    >
                      <PlayArrowIcon sx={{ fontSize: 16 }} />
                    </IconButton>
                  </Tooltip>
                  {airport.fetchUsageInfo && (
                    <Tooltip title="刷新用量" arrow>
                      <IconButton
                        size="small"
                        onClick={() => onRefreshUsage(airport)}
                        sx={{
                          bgcolor: alpha(theme.palette.success.main, 0.08),
                          color: theme.palette.success.main,
                          '&:hover': { bgcolor: alpha(theme.palette.success.main, 0.15) }
                        }}
                      >
                        <SyncIcon sx={{ fontSize: 16 }} />
                      </IconButton>
                    </Tooltip>
                  )}
                  <Tooltip title="编辑" arrow>
                    <IconButton
                      size="small"
                      onClick={() => onEdit(airport)}
                      sx={{
                        bgcolor: alpha(theme.palette.info.main, 0.08),
                        color: theme.palette.info.main,
                        '&:hover': { bgcolor: alpha(theme.palette.info.main, 0.15) }
                      }}
                    >
                      <EditIcon sx={{ fontSize: 16 }} />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="删除" arrow>
                    <IconButton
                      size="small"
                      onClick={() => onDelete(airport)}
                      sx={{
                        bgcolor: alpha(theme.palette.error.main, 0.08),
                        color: theme.palette.error.main,
                        '&:hover': { bgcolor: alpha(theme.palette.error.main, 0.15) }
                      }}
                    >
                      <DeleteIcon sx={{ fontSize: 16 }} />
                    </IconButton>
                  </Tooltip>
                </Box>
              </Box>
            </CardContent>
          </Card>
        ))}
      </Box>

      {/* 复制成功提示 */}
      <Snackbar
        open={copyTip.open}
        autoHideDuration={2000}
        onClose={() => setCopyTip({ open: false, name: '' })}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity="success" variant="standard" sx={{ width: '100%' }}>
          已复制「{copyTip.name}」的订阅地址
        </Alert>
      </Snackbar>
    </>
  );
}

AirportTable.propTypes = {
  airports: PropTypes.array.isRequired,
  onEdit: PropTypes.func.isRequired,
  onDelete: PropTypes.func.isRequired,
  onPull: PropTypes.func.isRequired,
  onRefreshUsage: PropTypes.func
};
