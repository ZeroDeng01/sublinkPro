import PropTypes from 'prop-types';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Tooltip from '@mui/material/Tooltip';

// icons
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import SpeedIcon from '@mui/icons-material/Speed';
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';
import HelpOutlineIcon from '@mui/icons-material/HelpOutline';

/**
 * 机场节点统计信息展示组件
 * 展示延迟测试通过数、速度测试通过数、最低延迟节点、最高速度节点
 */
export default function AirportNodeStatsCard({ nodeStats, nodeCount, compact = false }) {
  const theme = useTheme();

  // 检查是否有测试数据
  const hasData = nodeStats && (nodeStats.delayPassCount > 0 || nodeStats.speedPassCount > 0);

  // 紧凑模式（用于表格 Tooltip 触发区域）
  if (compact) {
    if (!hasData) {
      return (
        <Typography variant="caption" color="text.disabled" sx={{ fontStyle: 'italic' }}>
          暂未测试
        </Typography>
      );
    }

    return (
      <Stack direction="row" spacing={0.5} alignItems="center">
        <Tooltip title="延迟通过" arrow placement="top">
          <Stack direction="row" spacing={0.25} alignItems="center" sx={{ cursor: 'help' }}>
            <AccessTimeIcon sx={{ fontSize: 14, color: 'success.main' }} />
            <Typography variant="caption" fontWeight={600} color="success.main">
              {nodeStats.delayPassCount}
            </Typography>
          </Stack>
        </Tooltip>
        <Typography variant="caption" color="text.disabled">
          /
        </Typography>
        <Tooltip title="速度通过" arrow placement="top">
          <Stack direction="row" spacing={0.25} alignItems="center" sx={{ cursor: 'help' }}>
            <SpeedIcon sx={{ fontSize: 14, color: 'info.main' }} />
            <Typography variant="caption" fontWeight={600} color="info.main">
              {nodeStats.speedPassCount}
            </Typography>
          </Stack>
        </Tooltip>
      </Stack>
    );
  }

  // 完整展示模式（用于 Tooltip 内容和移动端卡片）
  if (!hasData) {
    return (
      <Box
        sx={{
          p: 2,
          borderRadius: 2,
          bgcolor: alpha(theme.palette.grey[500], 0.08),
          textAlign: 'center'
        }}
      >
        <HelpOutlineIcon sx={{ fontSize: 32, color: 'text.disabled', mb: 1 }} />
        <Typography variant="body2" color="text.disabled">
          该机场节点尚未进行测速
        </Typography>
        <Typography variant="caption" color="text.disabled">
          请先运行延迟或速度测试
        </Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ minWidth: 220 }}>
      {/* 通过数量统计 */}
      <Box
        sx={{
          display: 'grid',
          gridTemplateColumns: '1fr 1fr',
          gap: 1.5,
          mb: 2
        }}
      >
        {/* 延迟通过 */}
        <Box
          sx={{
            p: 1.5,
            borderRadius: 2,
            bgcolor: alpha(theme.palette.success.main, 0.1),
            border: `1px solid ${alpha(theme.palette.success.main, 0.2)}`
          }}
        >
          <Stack direction="row" alignItems="center" spacing={0.5} mb={0.5}>
            <AccessTimeIcon sx={{ fontSize: 14, color: 'success.main' }} />
            <Typography variant="caption" color="text.secondary">
              延迟通过
            </Typography>
          </Stack>
          <Typography variant="h6" fontWeight={700} color="success.main">
            {nodeStats.delayPassCount}
            <Typography component="span" variant="caption" color="text.disabled" sx={{ ml: 0.5 }}>
              / {nodeCount}
            </Typography>
          </Typography>
        </Box>

        {/* 速度通过 */}
        <Box
          sx={{
            p: 1.5,
            borderRadius: 2,
            bgcolor: alpha(theme.palette.info.main, 0.1),
            border: `1px solid ${alpha(theme.palette.info.main, 0.2)}`
          }}
        >
          <Stack direction="row" alignItems="center" spacing={0.5} mb={0.5}>
            <SpeedIcon sx={{ fontSize: 14, color: 'info.main' }} />
            <Typography variant="caption" color="text.secondary">
              速度通过
            </Typography>
          </Stack>
          <Typography variant="h6" fontWeight={700} color="info.main">
            {nodeStats.speedPassCount}
            <Typography component="span" variant="caption" color="text.disabled" sx={{ ml: 0.5 }}>
              / {nodeCount}
            </Typography>
          </Typography>
        </Box>
      </Box>

      {/* 最优节点信息 */}
      <Stack spacing={1.5}>
        {/* 最低延迟节点 */}
        {nodeStats.lowestDelayNode && (
          <Box
            sx={{
              p: 1.5,
              borderRadius: 2,
              bgcolor: alpha(theme.palette.warning.main, 0.08),
              border: `1px solid ${alpha(theme.palette.warning.main, 0.15)}`
            }}
          >
            <Stack direction="row" alignItems="center" spacing={0.5} mb={0.5}>
              <CheckCircleOutlineIcon sx={{ fontSize: 14, color: 'warning.main' }} />
              <Typography variant="caption" color="text.secondary">
                最低延迟
              </Typography>
            </Stack>
            <Tooltip title={nodeStats.lowestDelayNode} placement="top" arrow>
              <Typography
                variant="body2"
                fontWeight={600}
                color="warning.dark"
                sx={{
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap'
                }}
              >
                {nodeStats.lowestDelayNode}
              </Typography>
            </Tooltip>
            <Typography variant="caption" color="text.secondary">
              {nodeStats.lowestDelayTime}ms · {nodeStats.lowestDelaySpeed?.toFixed(1)}MB/s
            </Typography>
          </Box>
        )}

        {/* 最高速度节点 */}
        {nodeStats.highestSpeedNode && (
          <Box
            sx={{
              p: 1.5,
              borderRadius: 2,
              bgcolor: alpha(theme.palette.primary.main, 0.08),
              border: `1px solid ${alpha(theme.palette.primary.main, 0.15)}`
            }}
          >
            <Stack direction="row" alignItems="center" spacing={0.5} mb={0.5}>
              <CheckCircleOutlineIcon sx={{ fontSize: 14, color: 'primary.main' }} />
              <Typography variant="caption" color="text.secondary">
                最高速度
              </Typography>
            </Stack>
            <Tooltip title={nodeStats.highestSpeedNode} placement="top" arrow>
              <Typography
                variant="body2"
                fontWeight={600}
                color="primary.main"
                sx={{
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap'
                }}
              >
                {nodeStats.highestSpeedNode}
              </Typography>
            </Tooltip>
            <Typography variant="caption" color="text.secondary">
              {nodeStats.highestSpeed?.toFixed(1)}MB/s · {nodeStats.highestSpeedDelay}ms
            </Typography>
          </Box>
        )}
      </Stack>
    </Box>
  );
}

AirportNodeStatsCard.propTypes = {
  nodeStats: PropTypes.shape({
    delayPassCount: PropTypes.number,
    speedPassCount: PropTypes.number,
    lowestDelayNode: PropTypes.string,
    lowestDelayTime: PropTypes.number,
    lowestDelaySpeed: PropTypes.number,
    highestSpeedNode: PropTypes.string,
    highestSpeed: PropTypes.number,
    highestSpeedDelay: PropTypes.number
  }),
  nodeCount: PropTypes.number,
  compact: PropTypes.bool
};
