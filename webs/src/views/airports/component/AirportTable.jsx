import { useState } from 'react';
import PropTypes from 'prop-types';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Snackbar from '@mui/material/Snackbar';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

// icons
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import SyncIcon from '@mui/icons-material/Sync';

// utils
import { formatDateTime, formatBytes, formatExpireTime, getUsageColor } from '../utils';

// local components
import AirportNodeStatsCard from './AirportNodeStatsCard';
import AirportLogo from './AirportLogo';

/**
 * 机场列表表格组件
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

  // 渲染用量信息
  const renderUsageInfo = (airport) => {
    // 未开启获取用量信息
    // 未开启获取用量信息
    if (!airport.fetchUsageInfo) {
      return (
        <Typography variant="caption" color="textSecondary">
          未开启用量获取
        </Typography>
      );
    }

    // usageTotal 为 -1 表示获取失败（机场不支持或网络错误）
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

    // 根据使用率计算进度条渐变色
    const getProgressGradient = () => {
      if (percent < 60) return `linear-gradient(90deg, ${theme.palette.success.light}, ${theme.palette.success.main})`;
      if (percent < 85) return `linear-gradient(90deg, ${theme.palette.warning.light}, ${theme.palette.warning.main})`;
      return `linear-gradient(90deg, ${theme.palette.error.light}, ${theme.palette.error.main})`;
    };

    return (
      <Tooltip
        arrow
        placement="top"
        componentsProps={{
          tooltip: {
            sx: {
              bgcolor: theme.palette.mode === 'dark' ? 'grey.800' : 'grey.900',
              borderRadius: 2,
              p: 1.5,
              '& .MuiTooltip-arrow': {
                color: theme.palette.mode === 'dark' ? 'grey.800' : 'grey.900'
              }
            }
          }
        }}
        title={
          <Box>
            <Typography variant="subtitle2" sx={{ mb: 1, color: 'common.white', fontWeight: 600 }}>
              流量详情
            </Typography>
            <Box sx={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: 0.5, rowGap: 0.75 }}>
              <Typography variant="caption" sx={{ color: 'success.light' }}>
                ↑已用上传
              </Typography>
              <Typography variant="caption" sx={{ color: 'grey.300', textAlign: 'right' }}>
                {formatBytes(upload)}
              </Typography>
              <Typography variant="caption" sx={{ color: 'info.light' }}>
                ↓已用下载
              </Typography>
              <Typography variant="caption" sx={{ color: 'grey.300', textAlign: 'right' }}>
                {formatBytes(download)}
              </Typography>
              <Typography variant="caption" sx={{ color: 'grey.400' }}>
                ⌛️套餐总量
              </Typography>
              <Typography variant="caption" sx={{ color: 'grey.300', textAlign: 'right' }}>
                {formatBytes(total)}
              </Typography>
              {airport.usageExpire > 0 && (
                <>
                  <Typography variant="caption" sx={{ color: 'grey.400' }}>
                    ⏱️过期时间
                  </Typography>
                  <Typography variant="caption" sx={{ color: 'grey.300', textAlign: 'right' }}>
                    {formatExpireTime(airport.usageExpire)}
                  </Typography>
                </>
              )}
            </Box>
          </Box>
        }
      >
        <Box sx={{ minWidth: 140, cursor: 'pointer' }}>
          {/* 第一行：已用 / 总量 */}
          <Typography variant="body2" sx={{ fontWeight: 500, color: 'text.primary', mb: 0.5, lineHeight: 1.2 }}>
            {formatBytes(used)} / {formatBytes(total)}
          </Typography>

          {/* 第二行：进度条 + 百分比 */}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
            <Box
              sx={{
                flexGrow: 1,
                height: 6,
                borderRadius: 3,
                backgroundColor: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.08)',
                overflow: 'hidden'
              }}
            >
              <Box
                sx={{
                  width: `${percent}%`,
                  height: '100%',
                  borderRadius: 3,
                  background: getProgressGradient(),
                  transition: 'width 0.3s ease'
                }}
              />
            </Box>
            <Typography variant="caption" sx={{ fontWeight: 600, color: color, minWidth: 35 }}>
              {percent.toFixed(1)}%
            </Typography>
          </Box>

          {/* 第三行：过期时间 */}
          {airport.usageExpire > 0 && (
            <Typography variant="caption" sx={{ color: 'text.secondary', display: 'block', lineHeight: 1.2 }}>
              {formatExpireTime(airport.usageExpire)}
            </Typography>
          )}
        </Box>
      </Tooltip>
    );
  };

  return (
    <>
      <TableContainer>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>名称</TableCell>
              <TableCell align="center">节点数量</TableCell>
              <TableCell align="center">节点测试</TableCell>
              <TableCell>用量信息</TableCell>
              <TableCell>上次运行</TableCell>
              <TableCell>下次运行</TableCell>
              <TableCell>分组</TableCell>
              <TableCell align="center">状态</TableCell>
              <TableCell align="right">操作</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {airports.length === 0 ? (
              <TableRow>
                <TableCell colSpan={9} align="center">
                  <Typography variant="body2" color="textSecondary" sx={{ py: 4 }}>
                    暂无机场数据，点击上方"添加机场"按钮添加
                  </Typography>
                </TableCell>
              </TableRow>
            ) : (
              airports.map((airport) => (
                <TableRow key={airport.id} hover>
                  <TableCell>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                      <AirportLogo logo={airport.logo} name={airport.name} size="small" />
                      <Typography variant="body2" fontWeight={500}>
                        {airport.name}
                      </Typography>
                    </Box>
                  </TableCell>
                  <TableCell align="center">
                    <Chip label={airport.nodeCount || 0} color="primary" variant="outlined" size="small" />
                  </TableCell>
                  <TableCell align="center">
                    <Tooltip
                      arrow
                      placement="right"
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
                      title={<AirportNodeStatsCard nodeStats={airport.nodeStats} nodeCount={airport.nodeCount || 0} />}
                    >
                      <Box sx={{ cursor: 'help' }}>
                        <AirportNodeStatsCard nodeStats={airport.nodeStats} nodeCount={airport.nodeCount || 0} compact />
                      </Box>
                    </Tooltip>
                  </TableCell>
                  <TableCell>{renderUsageInfo(airport)}</TableCell>
                  <TableCell>
                    <Typography variant="body2" color="textSecondary">
                      {formatDateTime(airport.lastRunTime)}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" color="textSecondary">
                      {formatDateTime(airport.nextRunTime)}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2">{airport.group || '-'}</Typography>
                  </TableCell>
                  <TableCell align="center">
                    <Chip label={airport.enabled ? '启用' : '禁用'} color={airport.enabled ? 'success' : 'default'} size="small" />
                  </TableCell>
                  <TableCell align="right">
                    <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 0.5 }}>
                      <Tooltip title="复制订阅地址" arrow>
                        <IconButton size="small" onClick={() => handleCopyUrl(airport)} sx={{ color: theme.palette.secondary.main }}>
                          <ContentCopyIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="立即拉取" arrow>
                        <IconButton size="small" onClick={() => onPull(airport)} sx={{ color: theme.palette.primary.main }}>
                          <PlayArrowIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                      {airport.fetchUsageInfo && (
                        <Tooltip title="刷新用量" arrow>
                          <IconButton size="small" onClick={() => onRefreshUsage(airport)} sx={{ color: theme.palette.success.main }}>
                            <SyncIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      )}
                      <Tooltip title="编辑" arrow>
                        <IconButton size="small" onClick={() => onEdit(airport)}>
                          <EditIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="删除" arrow>
                        <IconButton size="small" color="error" onClick={() => onDelete(airport)}>
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </Box>
                  </TableCell>
                </TableRow>
              ))
            )}
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

AirportTable.propTypes = {
  airports: PropTypes.array.isRequired,
  onEdit: PropTypes.func.isRequired,
  onDelete: PropTypes.func.isRequired,
  onPull: PropTypes.func.isRequired,
  onRefreshUsage: PropTypes.func
};
