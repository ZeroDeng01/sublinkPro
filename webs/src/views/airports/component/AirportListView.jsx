import { useState } from 'react';
import PropTypes from 'prop-types';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Snackbar from '@mui/material/Snackbar';
import Stack from '@mui/material/Stack';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import LinearProgress from '@mui/material/LinearProgress';

// icons
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import SyncIcon from '@mui/icons-material/Sync';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import SpeedIcon from '@mui/icons-material/Speed';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';

// utils
import { formatDateTime, formatBytes, formatExpireTime, getUsageColor } from '../utils';

// local components
import AirportNodeStatsCard from './AirportNodeStatsCard';
import AirportLogo from './AirportLogo';

/**
 * 机场列表视图组件（桌面端表格模式）
 */
export default function AirportListView({ airports, onEdit, onDelete, onPull, onRefreshUsage }) {
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

  /**
   * 根据延迟值获取颜色
   */
  const getDelayColor = (delay) => {
    if (delay < 150) return theme.palette.success.main;
    if (delay < 300) return '#ec8f04ff';
    return theme.palette.error.main;
  };

  /**
   * 根据速度值获取颜色
   */
  const getSpeedColor = (speed) => {
    if (speed > 10) return theme.palette.success.main;
    if (speed > 5) return '#ec8f04ff';
    return theme.palette.error.main;
  };

  /**
   * 计算行背景色
   */
  const getRowBgColor = (airport) => {
    if (!airport.enabled) {
      return alpha(theme.palette.grey[500], 0.05);
    }

    // 检查用量警告（>= 85%）
    if (airport.fetchUsageInfo && airport.usageTotal > 0) {
      const upload = airport.usageUpload || 0;
      const download = airport.usageDownload || 0;
      const used = upload + download;
      const percent = (used / airport.usageTotal) * 100;
      if (percent >= 85) {
        return alpha(theme.palette.error.main, 0.05);
      }
    }

    // 检查过期警告（7天内）
    if (airport.usageExpire > 0) {
      const now = Math.floor(Date.now() / 1000);
      const daysLeft = (airport.usageExpire - now) / (24 * 60 * 60);
      if (daysLeft <= 7) {
        return alpha(theme.palette.error.main, 0.05);
      }
    }

    return 'transparent';
  };

  /**
   * 渲染用量信息（紧凑版）
   */
  const renderUsageCompact = (airport) => {
    if (!airport.fetchUsageInfo) {
      return (
        <Typography variant="caption" color="text.disabled">
          未开启
        </Typography>
      );
    }

    if (airport.usageTotal === -1) {
      return (
        <Typography variant="caption" color="error.main">
          获取失败
        </Typography>
      );
    }

    if (!airport.usageTotal || airport.usageTotal === 0) {
      return (
        <Typography variant="caption" color="text.disabled">
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

    // 检查过期时间
    let expireInfo = null;
    if (airport.usageExpire > 0) {
      const now = Math.floor(Date.now() / 1000);
      const daysLeft = (airport.usageExpire - now) / (24 * 60 * 60);
      const isUrgent = daysLeft <= 7;

      expireInfo = (
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mt: 0.25 }}>
          {isUrgent && <WarningAmberIcon sx={{ fontSize: 10, color: 'error.main' }} />}
          <Typography
            variant="caption"
            sx={{
              fontSize: '0.65rem',
              color: isUrgent ? 'error.main' : 'text.secondary',
              fontWeight: isUrgent ? 600 : 400
            }}
          >
            {formatExpireTime(airport.usageExpire)}
            {isUrgent && ` (${Math.max(0, Math.ceil(daysLeft))}天)`}
          </Typography>
        </Box>
      );
    }

    return (
      <Box sx={{ minWidth: 120 }}>
        {/* 进度条 + 百分比 */}
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Box sx={{ flex: 1, minWidth: 60 }}>
            <LinearProgress
              variant="determinate"
              value={percent}
              sx={{
                height: 4,
                borderRadius: 2,
                bgcolor: alpha(theme.palette.grey[500], 0.2),
                '& .MuiLinearProgress-bar': {
                  borderRadius: 2,
                  bgcolor: color
                }
              }}
            />
          </Box>
          <Typography variant="caption" sx={{ fontWeight: 600, color, minWidth: 35, textAlign: 'right' }}>
            {percent.toFixed(0)}%
          </Typography>
        </Box>
        {/* 已用/总量 */}
        <Typography variant="caption" sx={{ color: 'text.secondary', fontSize: '0.65rem' }}>
          {formatBytes(used)} / {formatBytes(total)}
        </Typography>
        {/* 过期时间 */}
        {expireInfo}
      </Box>
    );
  };

  /**
   * 渲染测速概览（紧凑版）
   */
  const renderSpeedCompact = (nodeStats, nodeCount) => {
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
          <Stack direction="row" spacing={0.5} alignItems="center">
            <Stack direction="row" spacing={0.25} alignItems="center">
              <AccessTimeIcon sx={{ fontSize: 11, color: 'success.main' }} />
              <Typography variant="caption" fontWeight={600} color="success.main" sx={{ fontSize: '0.7rem' }}>
                {nodeStats.delayPassCount}
              </Typography>
            </Stack>
            <Typography variant="caption" color="text.disabled" sx={{ fontSize: '0.65rem' }}>
              /
            </Typography>
            <Stack direction="row" spacing={0.25} alignItems="center">
              <SpeedIcon sx={{ fontSize: 11, color: 'info.main' }} />
              <Typography variant="caption" fontWeight={600} color="info.main" sx={{ fontSize: '0.7rem' }}>
                {nodeStats.speedPassCount}
              </Typography>
            </Stack>
          </Stack>
          {/* 最低延迟/最高速度 */}
          <Stack direction="row" spacing={1} sx={{ mt: 0.25 }}>
            {nodeStats.lowestDelayTime > 0 && (
              <Typography variant="caption" sx={{ fontSize: '0.6rem', color: getDelayColor(nodeStats.lowestDelayTime) }}>
                {nodeStats.lowestDelayTime}ms
              </Typography>
            )}
            {nodeStats.highestSpeed > 0 && (
              <Typography variant="caption" sx={{ fontSize: '0.6rem', color: getSpeedColor(nodeStats.highestSpeed) }}>
                {nodeStats.highestSpeed?.toFixed(1)}MB/s
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
      <TableContainer
        sx={{
          borderRadius: 2,
          border: `1px solid ${alpha(theme.palette.divider, 0.1)}`,
          overflow: 'hidden'
        }}
      >
        <Table size="small" sx={{ minWidth: 900 }}>
          <TableHead>
            <TableRow
              sx={{
                bgcolor: alpha(theme.palette.primary.main, 0.04),
                '& th': {
                  fontWeight: 600,
                  color: 'text.secondary',
                  fontSize: '0.75rem',
                  py: 1.5,
                  borderBottom: `1px solid ${alpha(theme.palette.divider, 0.1)}`
                }
              }}
            >
              <TableCell sx={{ width: '18%' }}>机场</TableCell>
              <TableCell sx={{ width: '6%' }} align="center">
                状态
              </TableCell>
              <TableCell sx={{ width: '6%' }} align="center">
                节点
              </TableCell>
              <TableCell sx={{ width: '10%' }}>调度</TableCell>
              <TableCell sx={{ width: '14%' }}>运行时间</TableCell>
              <TableCell sx={{ width: '18%' }}>用量</TableCell>
              <TableCell sx={{ width: '12%' }}>测速</TableCell>
              <TableCell sx={{ width: 130 }} align="center">
                操作
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {airports.map((airport) => (
              <TableRow
                key={airport.id}
                sx={{
                  bgcolor: getRowBgColor(airport),
                  '&:hover': {
                    bgcolor: alpha(theme.palette.primary.main, 0.04)
                  },
                  '& td': {
                    py: 1.5,
                    borderBottom: `1px solid ${alpha(theme.palette.divider, 0.06)}`
                  }
                }}
              >
                {/* 机场名称 + Logo + 分组 */}
                <TableCell>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <AirportLogo logo={airport.logo} name={airport.name} size="small" />
                    <Box sx={{ minWidth: 0, flex: 1 }}>
                      <Typography
                        variant="body2"
                        fontWeight={600}
                        sx={{
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                          maxWidth: 140
                        }}
                      >
                        {airport.name}
                      </Typography>
                      {airport.group && (
                        <Chip
                          label={airport.group}
                          variant="outlined"
                          size="small"
                          sx={{ height: 16, fontSize: '0.6rem', mt: 0.25, '& .MuiChip-label': { px: 0.75 } }}
                        />
                      )}
                    </Box>
                  </Box>
                </TableCell>

                {/* 状态 */}
                <TableCell align="center">
                  <Chip
                    label={airport.enabled ? '启用' : '禁用'}
                    color={airport.enabled ? 'success' : 'default'}
                    size="small"
                    sx={{ height: 20, fontSize: '0.65rem', '& .MuiChip-label': { px: 1 } }}
                  />
                </TableCell>

                {/* 节点数 */}
                <TableCell align="center">
                  <Typography variant="body2" fontWeight={500} color="primary.main">
                    {airport.nodeCount || 0}
                  </Typography>
                </TableCell>

                {/* 调度 */}
                <TableCell>
                  <Tooltip title={`Cron: ${airport.cronExpr}`} arrow>
                    <Typography
                      variant="caption"
                      sx={{
                        fontFamily: 'monospace',
                        fontSize: '0.7rem',
                        color: 'text.secondary',
                        display: 'block',
                        maxWidth: 90,
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap'
                      }}
                    >
                      {airport.cronExpr}
                    </Typography>
                  </Tooltip>
                </TableCell>

                {/* 运行时间 */}
                <TableCell>
                  <Box>
                    <Typography variant="caption" sx={{ fontSize: '0.65rem', color: 'text.secondary', display: 'block' }}>
                      上次: {formatDateTime(airport.lastRunTime)}
                    </Typography>
                    <Typography variant="caption" sx={{ fontSize: '0.65rem', color: 'text.secondary', display: 'block' }}>
                      下次: {formatDateTime(airport.nextRunTime)}
                    </Typography>
                  </Box>
                </TableCell>

                {/* 用量 */}
                <TableCell>{renderUsageCompact(airport)}</TableCell>

                {/* 测速 */}
                <TableCell>{renderSpeedCompact(airport.nodeStats, airport.nodeCount || 0)}</TableCell>

                {/* 操作按钮 */}
                <TableCell align="center">
                  <Stack direction="row" spacing={0.5} justifyContent="center">
                    <Tooltip title="复制订阅" arrow>
                      <IconButton
                        size="small"
                        onClick={() => handleCopyUrl(airport)}
                        sx={{
                          width: 26,
                          height: 26,
                          bgcolor: alpha(theme.palette.secondary.main, 0.08),
                          color: theme.palette.secondary.main,
                          '&:hover': { bgcolor: alpha(theme.palette.secondary.main, 0.15) }
                        }}
                      >
                        <ContentCopyIcon sx={{ fontSize: 14 }} />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="立即拉取" arrow>
                      <IconButton
                        size="small"
                        onClick={() => onPull(airport)}
                        sx={{
                          width: 26,
                          height: 26,
                          bgcolor: alpha(theme.palette.primary.main, 0.08),
                          color: theme.palette.primary.main,
                          '&:hover': { bgcolor: alpha(theme.palette.primary.main, 0.15) }
                        }}
                      >
                        <PlayArrowIcon sx={{ fontSize: 14 }} />
                      </IconButton>
                    </Tooltip>
                    {airport.fetchUsageInfo && (
                      <Tooltip title="刷新用量" arrow>
                        <IconButton
                          size="small"
                          onClick={() => onRefreshUsage(airport)}
                          sx={{
                            width: 26,
                            height: 26,
                            bgcolor: alpha(theme.palette.success.main, 0.08),
                            color: theme.palette.success.main,
                            '&:hover': { bgcolor: alpha(theme.palette.success.main, 0.15) }
                          }}
                        >
                          <SyncIcon sx={{ fontSize: 14 }} />
                        </IconButton>
                      </Tooltip>
                    )}
                    <Tooltip title="编辑" arrow>
                      <IconButton
                        size="small"
                        onClick={() => onEdit(airport)}
                        sx={{
                          width: 26,
                          height: 26,
                          bgcolor: alpha(theme.palette.info.main, 0.08),
                          color: theme.palette.info.main,
                          '&:hover': { bgcolor: alpha(theme.palette.info.main, 0.15) }
                        }}
                      >
                        <EditIcon sx={{ fontSize: 14 }} />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="删除" arrow>
                      <IconButton
                        size="small"
                        onClick={() => onDelete(airport)}
                        sx={{
                          width: 26,
                          height: 26,
                          bgcolor: alpha(theme.palette.error.main, 0.08),
                          color: theme.palette.error.main,
                          '&:hover': { bgcolor: alpha(theme.palette.error.main, 0.15) }
                        }}
                      >
                        <DeleteIcon sx={{ fontSize: 14 }} />
                      </IconButton>
                    </Tooltip>
                  </Stack>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

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

AirportListView.propTypes = {
  airports: PropTypes.array.isRequired,
  onEdit: PropTypes.func.isRequired,
  onDelete: PropTypes.func.isRequired,
  onPull: PropTypes.func.isRequired,
  onRefreshUsage: PropTypes.func
};
