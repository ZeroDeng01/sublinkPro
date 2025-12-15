import PropTypes from 'prop-types';
import { useMemo } from 'react';

// material-ui
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

/**
 * 协议对应的渐变色主题
 * 优化原则：使用高饱和度深色渐变，确保白色文字始终清晰可读
 */
const protocolThemes = {
  VMess: { bg: 'linear-gradient(135deg, #5046e5 0%, #7c3aed 100%)' },
  VLESS: { bg: 'linear-gradient(135deg, #db2777 0%, #9333ea 100%)' },
  Trojan: { bg: 'linear-gradient(135deg, #0284c7 0%, #0891b2 100%)' },
  SS: { bg: 'linear-gradient(135deg, #059669 0%, #10b981 100%)' },
  Shadowsocks: { bg: 'linear-gradient(135deg, #059669 0%, #10b981 100%)' },
  SSR: { bg: 'linear-gradient(135deg, #ea580c 0%, #f97316 100%)' },
  ShadowsocksR: { bg: 'linear-gradient(135deg, #ea580c 0%, #f97316 100%)' },
  Hysteria: { bg: 'linear-gradient(135deg, #6d28d9 0%, #8b5cf6 100%)' },
  Hysteria2: { bg: 'linear-gradient(135deg, #be185d 0%, #ec4899 100%)' },
  TUIC: { bg: 'linear-gradient(135deg, #4f46e5 0%, #8b5cf6 100%)' },
  WireGuard: { bg: 'linear-gradient(135deg, #1e40af 0%, #7c3aed 100%)' },
  Naive: { bg: 'linear-gradient(135deg, #b45309 0%, #d97706 100%)' },
  NaiveProxy: { bg: 'linear-gradient(135deg, #b45309 0%, #d97706 100%)' },
  Reality: { bg: 'linear-gradient(135deg, #0f766e 0%, #0d9488 100%)' },
  SOCKS5: { bg: 'linear-gradient(135deg, #1d4ed8 0%, #3b82f6 100%)' },
  AnyTLS: { bg: 'linear-gradient(135deg, #047857 0%, #10b981 100%)' }
};

const defaultTheme = { bg: 'linear-gradient(135deg, #5046e5 0%, #7c3aed 100%)' };

/**
 * 获取状态颜色
 */
const getStatusColor = (status) => {
  switch (status) {
    case 'success':
      return '#4caf50';
    case 'warning':
      return '#ff9800';
    case 'error':
      return '#f44336';
    default:
      return '#9e9e9e';
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
  const theme = useMemo(() => {
    return protocolThemes[node.Protocol] || defaultTheme;
  }, [node.Protocol]);

  const delayDisplay = getDelayDisplay(node.DelayTime, node.DelayStatus);
  const speedDisplay = getSpeedDisplay(node.Speed, node.SpeedStatus);
  const displayName = node.PreviewName || node.Name || node.OriginalName || '未知节点';

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
        background: theme.bg,
        boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
        transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
        '&:hover': {
          transform: 'translateY(-4px) scale(1.02)',
          boxShadow: '0 12px 24px rgba(0,0,0,0.15), 0 4px 8px rgba(0,0,0,0.1)',
          outline: '2px solid rgba(255,255,255,0.3)',
          outlineOffset: '-2px',
          '& .protocol-badge': {
            transform: 'scale(1.1)'
          },
          '& .country-flag': {
            transform: 'scale(1.1)'
          }
        },
        '&:active': {
          transform: 'scale(0.98)',
          boxShadow: '0 2px 8px rgba(0,0,0,0.15)'
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
          bgcolor: 'rgba(0,0,0,0.25)',
          transition: 'all 0.3s ease'
        }}
      >
        <Typography sx={{ color: '#fff', fontSize: 9, fontWeight: 700 }}>{node.Protocol || '?'}</Typography>
      </Box>

      {/* 节点名称区域 */}
      <Box sx={{ flex: 1, pr: 4, overflow: 'hidden' }}>
        <Tooltip title={displayName} placement="top" arrow>
          <Typography
            sx={{
              color: '#fff',
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
              color: 'rgba(255,255,255,0.75)',
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
          bgcolor: 'rgba(0,0,0,0.4)',
          display: 'flex',
          alignItems: 'center',
          gap: 0.75,
          borderRadius: '0 0 8px 8px'
        }}
      >
        {/* 延迟指标 */}
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
          <Box
            sx={{
              width: 6,
              height: 6,
              borderRadius: '50%',
              bgcolor: getStatusColor(delayDisplay.color),
              flexShrink: 0
            }}
          />
          <Typography sx={{ color: '#fff', fontSize: 10, fontWeight: 600, whiteSpace: 'nowrap' }}>{delayDisplay.text}</Typography>
        </Box>

        {/* 速度指标 */}
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
          <Box
            sx={{
              width: 6,
              height: 6,
              borderRadius: '50%',
              bgcolor: getStatusColor(speedDisplay.color),
              flexShrink: 0
            }}
          />
          <Typography sx={{ color: '#fff', fontSize: 10, fontWeight: 600, whiteSpace: 'nowrap' }}>{speedDisplay.text}</Typography>
        </Box>

        {/* 右侧填充 */}
        <Box sx={{ flex: 1 }} />

        {/* 国旗 + 国家代码 */}
        <Stack direction="row" alignItems="center" spacing={0.25} className="country-flag" sx={{ transition: 'all 0.3s ease' }}>
          <Typography sx={{ fontSize: 12, lineHeight: 1 }}>{node.CountryFlag || '🌐'}</Typography>
          {node.LinkCountry && (
            <Typography sx={{ color: 'rgba(255,255,255,0.9)', fontSize: 8, fontWeight: 600 }}>{node.LinkCountry}</Typography>
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
