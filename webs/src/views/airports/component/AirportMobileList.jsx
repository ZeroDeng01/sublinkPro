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
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import SyncIcon from '@mui/icons-material/Sync';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';
import EventIcon from '@mui/icons-material/Event';

// utils
import { formatDateTime, formatBytes, formatExpireTime, getUsageColor } from '../utils';

// local components
import AirportNodeStatsCard from './AirportNodeStatsCard';
import AirportLogo from './AirportLogo';

/**
 * 机场移动端列表组件
 */
export default function AirportMobileList({ airports, onEdit, onDelete, onPull, onRefreshUsage }) {
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

  if (airports.length === 0) {
    return (
      <Box sx={{ py: 6, textAlign: 'center' }}>
        <Typography variant="body2" color="textSecondary">
          暂无机场数据，点击上方"添加"按钮添加
        </Typography>
      </Box>
    );
  }

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

  return (
    <>
      <Stack spacing={2.5}>
        {airports.map((airport) => (
          <Card
            key={airport.id}
            sx={{
              borderRadius: 3,
              border: `1px solid ${alpha(theme.palette.divider, 0.15)}`,
              boxShadow:
                theme.palette.mode === 'dark'
                  ? `0 4px 12px ${alpha('#000', 0.3)}`
                  : `0 4px 12px ${alpha(theme.palette.primary.main, 0.08)}`,
              transition: 'all 0.2s ease',
              overflow: 'hidden',
              position: 'relative',
              '&:hover': {
                transform: 'translateY(-2px)',
                boxShadow:
                  theme.palette.mode === 'dark'
                    ? `0 8px 24px ${alpha('#000', 0.4)}`
                    : `0 8px 24px ${alpha(theme.palette.primary.main, 0.15)}`
              },
              // 顶部状态指示条
              '&::before': {
                content: '""',
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                height: 4,
                background: getStatusBarColor(airport)
              }
            }}
          >
            <CardContent sx={{ p: 2, pt: 2.5, '&:last-child': { pb: 2 } }}>
              {/* 顶部：Logo、名称和状态 */}
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 1.5 }}>
                <AirportLogo logo={airport.logo} name={airport.name} size="medium" />
                <Typography variant="subtitle1" fontWeight={600} sx={{ flex: 1 }}>
                  {airport.name}
                </Typography>
                <Chip label={airport.enabled ? '启用' : '禁用'} color={airport.enabled ? 'success' : 'default'} size="small" />
              </Box>

              {/* 信息行 */}
              <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mb: 2 }}>
                <Chip label={`${airport.nodeCount || 0} 节点`} color="primary" variant="outlined" size="small" />
                {airport.group && <Chip label={airport.group} variant="outlined" size="small" />}
                <Chip
                  icon={<AccessTimeIcon sx={{ fontSize: '14px !important' }} />}
                  label={airport.cronExpr}
                  variant="outlined"
                  size="small"
                />
              </Box>

              {/* 时间信息 */}
              <Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
                <Box>
                  <Typography variant="caption" color="textSecondary">
                    上次运行
                  </Typography>
                  <Typography variant="body2">{formatDateTime(airport.lastRunTime)}</Typography>
                </Box>
                <Box>
                  <Typography variant="caption" color="textSecondary">
                    下次运行
                  </Typography>
                  <Typography variant="body2">{formatDateTime(airport.nextRunTime)}</Typography>
                </Box>
              </Box>

              {/* 用量信息 */}
              {airport.fetchUsageInfo && (
                <Box sx={{ mb: 2 }}>
                  <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mb: 1, fontWeight: 500 }}>
                    用量信息
                  </Typography>
                  {airport.usageTotal === -1 ? (
                    <Typography variant="body2" sx={{ color: 'error.main', fontWeight: 500 }}>
                      用量获取失败（机场可能不支持）
                    </Typography>
                  ) : airport.usageTotal > 0 ? (
                    <Box
                      sx={{
                        p: 1.5,
                        borderRadius: 2,
                        backgroundColor: alpha(theme.palette.primary.main, 0.04),
                        border: `1px solid ${alpha(theme.palette.divider, 0.1)}`
                      }}
                    >
                      {(() => {
                        const upload = airport.usageUpload || 0;
                        const download = airport.usageDownload || 0;
                        const used = upload + download;
                        const total = airport.usageTotal;
                        const percent = Math.min((used / total) * 100, 100);
                        const color = getUsageColor(percent);

                        // 根据使用率计算进度条渐变色
                        const getProgressGradient = () => {
                          if (percent < 60) return `linear-gradient(90deg, ${theme.palette.success.light}, ${theme.palette.success.main})`;
                          if (percent < 85) return `linear-gradient(90deg, ${theme.palette.warning.light}, ${theme.palette.warning.main})`;
                          return `linear-gradient(90deg, ${theme.palette.error.light}, ${theme.palette.error.main})`;
                        };

                        return (
                          <>
                            {/* 已用/总量 + 百分比 */}
                            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
                              <Typography variant="body2" sx={{ fontWeight: 600, color: 'text.primary' }}>
                                {formatBytes(used)} / {formatBytes(total)}
                              </Typography>
                              <Typography variant="body1" sx={{ fontWeight: 700, color: color }}>
                                {percent.toFixed(1)}%
                              </Typography>
                            </Box>

                            {/* 进度条 */}
                            <Box
                              sx={{
                                height: 8,
                                borderRadius: 4,
                                backgroundColor: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.08)',
                                overflow: 'hidden',
                                mb: 1.5
                              }}
                            >
                              <Box
                                sx={{
                                  width: `${percent}%`,
                                  height: '100%',
                                  borderRadius: 4,
                                  background: getProgressGradient(),
                                  transition: 'width 0.3s ease'
                                }}
                              />
                            </Box>

                            {/* 上传/下载分项 */}
                            <Box sx={{ display: 'flex', gap: 3, mb: airport.usageExpire > 0 ? 1 : 0 }}>
                              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                <Typography variant="body2" sx={{ color: 'success.main', fontWeight: 500 }}>
                                  ↑
                                </Typography>
                                <Typography variant="body2" sx={{ color: 'text.secondary' }}>
                                  {formatBytes(upload)}
                                </Typography>
                              </Box>
                              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                <Typography variant="body2" sx={{ color: 'info.main', fontWeight: 500 }}>
                                  ↓
                                </Typography>
                                <Typography variant="body2" sx={{ color: 'text.secondary' }}>
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
                                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                    {isUrgent && <WarningAmberIcon sx={{ fontSize: 14, color: 'error.main' }} />}
                                    {isWarning && <EventIcon sx={{ fontSize: 14, color: 'info.main' }} />}
                                    <Typography
                                      variant="body2"
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
                          </>
                        );
                      })()}
                    </Box>
                  ) : (
                    <Typography variant="body2" color="textSecondary">
                      待获取
                    </Typography>
                  )}
                </Box>
              )}

              {/* 节点测试统计 */}
              <Box sx={{ mb: 2 }}>
                <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mb: 1, fontWeight: 500 }}>
                  节点测试
                </Typography>
                <AirportNodeStatsCard nodeStats={airport.nodeStats} nodeCount={airport.nodeCount || 0} />
              </Box>

              {/* 操作按钮 */}
              <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 1 }}>
                <Tooltip title="复制订阅地址" arrow>
                  <IconButton
                    size="small"
                    onClick={() => handleCopyUrl(airport)}
                    sx={{
                      bgcolor: alpha(theme.palette.secondary.main, 0.1),
                      color: theme.palette.secondary.main,
                      '&:hover': { bgcolor: alpha(theme.palette.secondary.main, 0.2) }
                    }}
                  >
                    <ContentCopyIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
                <Tooltip title="立即拉取" arrow>
                  <IconButton
                    size="small"
                    onClick={() => onPull(airport)}
                    sx={{
                      bgcolor: alpha(theme.palette.primary.main, 0.1),
                      color: theme.palette.primary.main,
                      '&:hover': { bgcolor: alpha(theme.palette.primary.main, 0.2) }
                    }}
                  >
                    <PlayArrowIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
                {airport.fetchUsageInfo && (
                  <Tooltip title="刷新用量" arrow>
                    <IconButton
                      size="small"
                      onClick={() => onRefreshUsage(airport)}
                      sx={{
                        bgcolor: alpha(theme.palette.success.main, 0.1),
                        color: theme.palette.success.main,
                        '&:hover': { bgcolor: alpha(theme.palette.success.main, 0.2) }
                      }}
                    >
                      <SyncIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                )}
                <Tooltip title="编辑" arrow>
                  <IconButton
                    size="small"
                    onClick={() => onEdit(airport)}
                    sx={{
                      bgcolor: alpha(theme.palette.info.main, 0.1),
                      color: theme.palette.info.main,
                      '&:hover': { bgcolor: alpha(theme.palette.info.main, 0.2) }
                    }}
                  >
                    <EditIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
                <Tooltip title="删除" arrow>
                  <IconButton
                    size="small"
                    onClick={() => onDelete(airport)}
                    sx={{
                      bgcolor: alpha(theme.palette.error.main, 0.1),
                      color: theme.palette.error.main,
                      '&:hover': { bgcolor: alpha(theme.palette.error.main, 0.2) }
                    }}
                  >
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              </Box>
            </CardContent>
          </Card>
        ))}
      </Stack>

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

AirportMobileList.propTypes = {
  airports: PropTypes.array.isRequired,
  onEdit: PropTypes.func.isRequired,
  onDelete: PropTypes.func.isRequired,
  onPull: PropTypes.func.isRequired,
  onRefreshUsage: PropTypes.func
};
