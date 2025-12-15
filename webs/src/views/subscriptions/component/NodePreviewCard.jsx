import PropTypes from 'prop-types';
import { useMemo } from 'react';

// material-ui
import { alpha } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

/**
 * 协议对应的渐变色主题
 */
const protocolThemes = {
  VMess: { bg: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)', text: '#fff' },
  VLESS: { bg: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)', text: '#fff' },
  Trojan: { bg: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)', text: '#fff' },
  SS: { bg: 'linear-gradient(135deg, #11998e 0%, #38ef7d 100%)', text: '#fff' },
  Shadowsocks: { bg: 'linear-gradient(135deg, #11998e 0%, #38ef7d 100%)', text: '#fff' },
  SSR: { bg: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)', text: '#fff' },
  ShadowsocksR: { bg: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)', text: '#fff' },
  Hysteria: { bg: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)', text: '#fff' },
  Hysteria2: { bg: 'linear-gradient(135deg, #ff9a9e 0%, #fad0c4 100%)', text: '#fff' },
  TUIC: { bg: 'linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)', text: '#fff' },
  WireGuard: { bg: 'linear-gradient(135deg, #88d3ce 0%, #6e45e2 100%)', text: '#fff' },
  Naive: { bg: 'linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%)', text: '#333' },
  NaiveProxy: { bg: 'linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%)', text: '#333' },
  Reality: { bg: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)', text: '#fff' },
  SOCKS5: { bg: 'linear-gradient(135deg, #89f7fe 0%, #66a6ff 100%)', text: '#fff' },
  AnyTLS: { bg: 'linear-gradient(135deg, #96fbc4 0%, #f9f586 100%)', text: '#333' }
};

const defaultTheme = { bg: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)', text: '#fff' };

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
  if (speed >= 5) return { text: `${speed.toFixed(1)}M`, color: 'success' };
  if (speed >= 1) return { text: `${speed.toFixed(1)}M`, color: 'warning' };
  return { text: `${speed.toFixed(2)}M`, color: 'error' };
};

/**
 * 节点预览卡片组件 - 固定尺寸紧凑卡片
 * 国旗移至右下角，避免与节点名称中的国旗重复
 */
export default function NodePreviewCard({ node, onClick }) {
  const theme = useMemo(() => {
    return protocolThemes[node.Protocol] || defaultTheme;
  }, [node.Protocol]);

  const delayDisplay = getDelayDisplay(node.DelayTime, node.DelayStatus);
  const speedDisplay = getSpeedDisplay(node.Speed, node.SpeedStatus);

  // 处理节点名称显示
  const displayName = node.PreviewName || node.Name || node.OriginalName || '未知节点';

  return (
    <Box
      onClick={onClick}
      sx={{
        position: 'relative',
        p: 1,
        borderRadius: 2,
        cursor: 'pointer',
        overflow: 'hidden',
        // 固定高度确保所有卡片大小一致
        height: 88,
        display: 'flex',
        flexDirection: 'column',
        transition: 'all 0.2s ease',
        background: theme.bg,
        boxShadow: '0 3px 10px rgba(0,0,0,0.12)',
        '&:hover': {
          transform: 'translateY(-2px)',
          boxShadow: '0 6px 16px rgba(0,0,0,0.2)'
        },
        '&:active': {
          transform: 'scale(0.98)'
        }
      }}
    >
      {/* 协议标签 - 右上角 */}
      <Box
        sx={{
          position: 'absolute',
          top: 4,
          right: 4,
          px: 0.5,
          py: 0.125,
          borderRadius: 0.5,
          bgcolor: 'rgba(255,255,255,0.2)'
        }}
      >
        <Typography sx={{ color: theme.text, fontSize: 9, fontWeight: 700 }}>{node.Protocol || '?'}</Typography>
      </Box>

      {/* 国旗 + 国家代码 - 右下角 */}
      <Stack
        direction="row"
        alignItems="center"
        spacing={0.25}
        sx={{
          position: 'absolute',
          bottom: 4,
          right: 4,
          opacity: 0.9
        }}
      >
        <Typography sx={{ fontSize: 12, lineHeight: 1 }}>{node.CountryFlag || '🌐'}</Typography>
        {node.LinkCountry && (
          <Typography sx={{ color: alpha(theme.text, 0.8), fontSize: 8, fontWeight: 600 }}>{node.LinkCountry}</Typography>
        )}
      </Stack>

      {/* 节点名称 - 占据主要空间 */}
      <Box sx={{ flex: 1, pr: 4 }}>
        <Tooltip title={displayName} placement="top" arrow>
          <Typography
            sx={{
              color: theme.text,
              fontWeight: 600,
              fontSize: 11,
              lineHeight: 1.3,
              display: '-webkit-box',
              WebkitLineClamp: 2,
              WebkitBoxOrient: 'vertical',
              overflow: 'hidden',
              textShadow: theme.text === '#fff' ? '0 1px 2px rgba(0,0,0,0.2)' : 'none'
            }}
          >
            {displayName}
          </Typography>
        </Tooltip>
        {/* 分组 */}
        {node.Group && (
          <Typography
            sx={{
              color: alpha(theme.text, 0.75),
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

      {/* 底部指标 - 左下角 */}
      <Stack direction="row" spacing={0.5} sx={{ pr: 3 }}>
        <Chip
          label={delayDisplay.text}
          size="small"
          color={delayDisplay.color}
          sx={{ height: 16, fontSize: 9, fontWeight: 600, '& .MuiChip-label': { px: 0.5, py: 0 } }}
        />
        <Chip
          label={speedDisplay.text}
          size="small"
          color={speedDisplay.color}
          sx={{ height: 16, fontSize: 9, fontWeight: 600, '& .MuiChip-label': { px: 0.5, py: 0 } }}
        />
      </Stack>
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
