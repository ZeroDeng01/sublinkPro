import PropTypes from 'prop-types';

// material-ui
import { useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Tooltip from '@mui/material/Tooltip';

// icons
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import SpeedIcon from '@mui/icons-material/Speed';
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';
import HelpOutlineIcon from '@mui/icons-material/HelpOutline';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

/**
 * 机场节点统计信息展示组件
 * 展示延迟测试通过数、速度测试通过数、最低延迟节点、最高速度节点
 */
export default function AirportNodeStatsCard({ nodeStats, nodeCount, compact = false }) {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);

  const shellSx = {
    minWidth: 220,
    p: 1.5,
    borderRadius: 2.5,
    color: primaryText,
    bgcolor: dialogSurface,
    backgroundImage: dialogSurfaceGradient,
    border: '1px solid',
    borderColor: isDark ? withAlpha(palette.divider, 0.58) : panelBorder,
    boxShadow: isDark
      ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}, 0 8px 18px ${withAlpha(palette.common.black, 0.18)}`
      : 'none'
  };

  const sectionSx = {
    p: 1,
    borderRadius: 2,
    bgcolor: mutedPanelSurface,
    border: '1px solid',
    borderColor: isDark ? withAlpha(palette.divider, 0.42) : panelBorder,
    boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.02)}` : 'none'
  };

  const getAccentPanelSx = (accentColor) => ({
    p: 1.25,
    borderRadius: 1.75,
    bgcolor: nestedPanelSurface,
    backgroundImage: `linear-gradient(180deg, ${withAlpha(accentColor, isDark ? 0.16 : 0.08)} 0%, ${nestedPanelSurface} 100%)`,
    border: `1px solid ${withAlpha(accentColor, isDark ? 0.28 : 0.16)}`,
    boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.035)}` : 'none'
  });

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
      <Box sx={shellSx}>
        <Box
          sx={{
            ...sectionSx,
            px: 1.5,
            py: 2,
            textAlign: 'center'
          }}
        >
          <HelpOutlineIcon sx={{ fontSize: 32, color: tertiaryText, mb: 1 }} />
          <Typography variant="body2" sx={{ color: secondaryText, fontWeight: 500 }}>
            该机场节点尚未进行测速
          </Typography>
          <Typography variant="caption" sx={{ color: tertiaryText }}>
            请先运行延迟或速度测试
          </Typography>
        </Box>
      </Box>
    );
  }

  return (
    <Box sx={shellSx}>
      <Stack spacing={1.25}>
        <Box sx={sectionSx}>
          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: '1fr 1fr',
              gap: 1
            }}
          >
            <Box sx={getAccentPanelSx(palette.success.main)}>
              <Stack direction="row" alignItems="center" spacing={0.5} mb={0.5}>
                <AccessTimeIcon sx={{ fontSize: 14, color: 'success.main' }} />
                <Typography variant="caption" sx={{ color: secondaryText }}>
                  延迟通过
                </Typography>
              </Stack>
              <Typography variant="h6" fontWeight={700} color="success.main">
                {nodeStats.delayPassCount}
                <Typography component="span" variant="caption" sx={{ ml: 0.5, color: tertiaryText }}>
                  / {nodeCount}
                </Typography>
              </Typography>
            </Box>

            <Box sx={getAccentPanelSx(palette.info.main)}>
              <Stack direction="row" alignItems="center" spacing={0.5} mb={0.5}>
                <SpeedIcon sx={{ fontSize: 14, color: 'info.main' }} />
                <Typography variant="caption" sx={{ color: secondaryText }}>
                  速度通过
                </Typography>
              </Stack>
              <Typography variant="h6" fontWeight={700} color="info.main">
                {nodeStats.speedPassCount}
                <Typography component="span" variant="caption" sx={{ ml: 0.5, color: tertiaryText }}>
                  / {nodeCount}
                </Typography>
              </Typography>
            </Box>
          </Box>
        </Box>

        <Box sx={sectionSx}>
          <Stack spacing={1}>
            {nodeStats.lowestDelayNode && (
              <Box sx={getAccentPanelSx(palette.warning.main)}>
                <Stack direction="row" alignItems="center" spacing={0.5} mb={0.5}>
                  <CheckCircleOutlineIcon sx={{ fontSize: 14, color: 'warning.main' }} />
                  <Typography variant="caption" sx={{ color: secondaryText }}>
                    最低延迟
                  </Typography>
                </Stack>
                <Tooltip title={nodeStats.lowestDelayNode} placement="top" arrow>
                  <Typography
                    variant="body2"
                    fontWeight={600}
                    sx={{
                      color: isDark ? palette.warning.light : palette.warning.dark,
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap'
                    }}
                  >
                    {nodeStats.lowestDelayNode}
                  </Typography>
                </Tooltip>
                <Typography variant="caption" sx={{ color: secondaryText }}>
                  {nodeStats.lowestDelayTime}ms · {nodeStats.lowestDelaySpeed?.toFixed(1)}MB/s
                </Typography>
              </Box>
            )}

            {nodeStats.highestSpeedNode && (
              <Box sx={getAccentPanelSx(palette.primary.main)}>
                <Stack direction="row" alignItems="center" spacing={0.5} mb={0.5}>
                  <CheckCircleOutlineIcon sx={{ fontSize: 14, color: 'primary.main' }} />
                  <Typography variant="caption" sx={{ color: secondaryText }}>
                    最高速度
                  </Typography>
                </Stack>
                <Tooltip title={nodeStats.highestSpeedNode} placement="top" arrow>
                  <Typography
                    variant="body2"
                    fontWeight={600}
                    sx={{
                      color: isDark ? palette.primary.light : palette.primary.main,
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap'
                    }}
                  >
                    {nodeStats.highestSpeedNode}
                  </Typography>
                </Tooltip>
                <Typography variant="caption" sx={{ color: secondaryText }}>
                  {nodeStats.highestSpeed?.toFixed(1)}MB/s · {nodeStats.highestSpeedDelay}ms
                </Typography>
              </Box>
            )}
          </Stack>
        </Box>
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
