import PropTypes from 'prop-types';

// material-ui
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import { alpha, useTheme } from '@mui/material/styles';
import { getProtocolPresentation } from '../../../utils/protocolPresentation';
import { withAlpha } from '../../../utils/colorUtils';

/**
 * 获取状态颜色
 */
const getStatusColor = (theme, status) => {
  switch (status) {
    case 'success':
      return theme.palette.success.main;
    case 'warning':
      return theme.palette.warning.main;
    case 'error':
      return theme.palette.error.main;
    default:
      return theme.palette.text.secondary;
  }
};

/**
 * 获取延迟显示
 */
const getDelayDisplay = (delayTime, delayStatus) => {
  if (delayStatus === 'timeout' || delayStatus === 2) return { text: '超时', color: 'error' };
  if (delayStatus === 'error' || delayStatus === 3) return { text: '错误', color: 'error' };
  if (!delayTime || delayTime <= 0) return { text: '未测', color: 'default' };
  if (delayTime < 200) return { text: `${delayTime}ms`, color: 'success' };
  if (delayTime < 500) return { text: `${delayTime}ms`, color: 'warning' };
  return { text: `${delayTime}ms`, color: 'error' };
};

/**
 * 获取速度显示
 */
const getSpeedDisplay = (speed, speedStatus) => {
  if (speedStatus === 'timeout' || speedStatus === 2) return { text: '超时', color: 'error' };
  if (speedStatus === 'error' || speedStatus === 3) return { text: '错误', color: 'error' };
  if (!speed || speed <= 0) return { text: '未测', color: 'default' };
  if (speed >= 5) return { text: `${speed.toFixed(1)}MB/s`, color: 'success' };
  if (speed >= 1) return { text: `${speed.toFixed(1)}MB/s`, color: 'warning' };
  return { text: `${speed.toFixed(2)}MB/s`, color: 'error' };
};

/**
 * 节点预览卡片组件 - 固定尺寸紧凑卡片
 */
export default function NodePreviewCard({ node, onClick }) {
  const theme = useTheme();
  const palette = theme.vars?.palette || theme.palette;

  const delayDisplay = getDelayDisplay(node.DelayTime, node.DelayStatus);
  const speedDisplay = getSpeedDisplay(node.Speed, node.SpeedStatus);
  const displayName = node.PreviewName || node.Name || node.OriginalName || '未知节点';
  const protocolInfo = getProtocolPresentation(node.Protocol);
  const protocolColor = protocolInfo.color || theme.palette.primary.main;
  const isDark = theme.palette.mode === 'dark';
  const footerBackground = isDark ? withAlpha(palette.background.default, 0.92) : withAlpha(palette.background.default, 0.88);
  const footerBorderColor = isDark ? withAlpha(palette.divider, 0.82) : withAlpha(palette.divider, 0.9);

  return (
    <Box
      onClick={onClick}
      sx={{
        position: 'relative',
        p: 1,
        pb: 3.5, // 底部预留空间给指标栏
        borderRadius: 2,
        cursor: 'pointer',
        overflow: 'hidden',
        height: 88,
        display: 'flex',
        flexDirection: 'column',
        bgcolor: 'background.paper',
        border: '1px solid',
        borderColor: alpha(theme.palette.divider, 0.9),
        boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.03)}` : theme.shadows[1],
        transition: 'border-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease',
        '&::before': {
          content: '""',
          position: 'absolute',
          inset: 0,
          borderTop: `3px solid ${protocolColor}`,
          pointerEvents: 'none'
        },
        '&:hover': {
          transform: 'translateY(-1px)',
          borderColor: alpha(protocolColor, 0.35),
          boxShadow: theme.palette.mode === 'dark' ? `0 0 0 1px ${alpha(protocolColor, 0.18)}` : theme.shadows[3],
          '& .protocol-badge': {
            transform: 'scale(1.1)'
          },
          '& .country-flag': {
            transform: 'scale(1.1)'
          }
        },
        '&:active': {
          transform: 'scale(0.98)',
          boxShadow: theme.shadows[2]
        }
      }}
    >
      {/* 协议标签 - 右上角 */}
      <Box
        className="protocol-badge"
        sx={{
          position: 'absolute',
          top: 4,
          right: 4,
          px: 0.75,
          py: 0.25,
          borderRadius: 0.75,
          bgcolor: alpha(protocolColor, 0.12),
          color: protocolColor,
          border: '1px solid',
          borderColor: alpha(protocolColor, 0.2),
          transition: 'all 0.3s ease'
        }}
      >
        <Typography sx={{ color: 'inherit', fontSize: 9, fontWeight: 700 }}>{node.Protocol || '?'}</Typography>
      </Box>

      {/* 节点名称区域 */}
      <Box sx={{ flex: 1, pr: 4, overflow: 'hidden' }}>
        <Tooltip title={displayName} placement="top" arrow>
          <Typography
            sx={{
              color: 'text.primary',
              fontWeight: 600,
              fontSize: 11,
              lineHeight: 1.3,
              display: '-webkit-box',
              WebkitLineClamp: 2,
              WebkitBoxOrient: 'vertical',
              overflow: 'hidden'
            }}
          >
            {displayName}
          </Typography>
        </Tooltip>
        {node.Group && (
          <Typography
            sx={{
              color: 'text.secondary',
              fontSize: 9,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
              mt: 0.25
            }}
          >
            {node.Group}
          </Typography>
        )}
      </Box>

      {/* 底部指标栏 - 深色半透明背景确保可读性 */}
      <Box
        sx={{
          position: 'absolute',
          bottom: 0,
          left: 0,
          right: 0,
          px: 1,
          py: 0.5,
          background: footerBackground,
          display: 'flex',
          alignItems: 'center',
          gap: 0.75,
          borderRadius: '0 0 8px 8px',
          borderTop: '1px solid',
          borderColor: footerBorderColor,
          backdropFilter: 'blur(6px)'
        }}
      >
        {/* 延迟指标 */}
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
          <Box
            sx={{
              width: 6,
              height: 6,
              borderRadius: '50%',
              bgcolor: getStatusColor(theme, delayDisplay.color),
              flexShrink: 0
            }}
          />
          <Typography
            sx={{
              color: 'text.primary',
              fontSize: 10,
              fontWeight: 700,
              whiteSpace: 'nowrap'
            }}
          >
            {delayDisplay.text}
          </Typography>
        </Box>

        {/* 速度指标 */}
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
          <Box
            sx={{
              width: 6,
              height: 6,
              borderRadius: '50%',
              bgcolor: getStatusColor(theme, speedDisplay.color),
              flexShrink: 0
            }}
          />
          <Typography
            sx={{
              color: 'text.primary',
              fontSize: 10,
              fontWeight: 700,
              whiteSpace: 'nowrap'
            }}
          >
            {speedDisplay.text}
          </Typography>
        </Box>

        {/* 右侧填充 */}
        <Box sx={{ flex: 1 }} />

        {/* 国旗 + 国家代码 */}
        <Stack direction="row" alignItems="center" spacing={0.25} className="country-flag" sx={{ transition: 'all 0.3s ease' }}>
          <Typography sx={{ fontSize: 12, lineHeight: 1 }}>{node.CountryFlag || '🌐'}</Typography>
          {node.LinkCountry && (
            <Typography
              sx={{
                color: 'text.secondary',
                fontSize: 8,
                fontWeight: 700
              }}
            >
              {node.LinkCountry}
            </Typography>
          )}
        </Stack>
      </Box>
    </Box>
  );
}

NodePreviewCard.propTypes = {
  node: PropTypes.shape({
    OriginalName: PropTypes.string,
    PreviewName: PropTypes.string,
    Name: PropTypes.string,
    PreviewLink: PropTypes.string,
    Protocol: PropTypes.string,
    CountryFlag: PropTypes.string,
    LinkCountry: PropTypes.string,
    DelayTime: PropTypes.number,
    DelayStatus: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
    Speed: PropTypes.number,
    SpeedStatus: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
    Group: PropTypes.string,
    Tags: PropTypes.string,
    Link: PropTypes.string
  }).isRequired,
  onClick: PropTypes.func.isRequired
};
