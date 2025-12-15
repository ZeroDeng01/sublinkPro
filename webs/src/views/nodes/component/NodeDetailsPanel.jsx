import PropTypes from 'prop-types';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import Drawer from '@mui/material/Drawer';
import IconButton from '@mui/material/IconButton';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import Stack from '@mui/material/Stack';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';

// icons
import CloseIcon from '@mui/icons-material/Close';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import PublicIcon from '@mui/icons-material/Public';
import FolderIcon from '@mui/icons-material/Folder';
import SourceIcon from '@mui/icons-material/Source';
import LinkIcon from '@mui/icons-material/Link';
import RouterIcon from '@mui/icons-material/Router';
import FilterVintageIcon from '@mui/icons-material/FilterVintage';
import VpnLockIcon from '@mui/icons-material/VpnLock';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import SignalCellularAltIcon from '@mui/icons-material/SignalCellularAlt';
import SpeedIcon from '@mui/icons-material/Speed';

// utils
import { formatDateTime, formatCountry, getDelayDisplay, getSpeedDisplay } from '../utils';

/**
 * 解析节点协议类型
 */
const getProtocolInfo = (link) => {
  if (!link) return { name: '未知', color: '#9e9e9e', icon: <FilterVintageIcon /> };

  const protocolMap = {
    'vmess://': { name: 'VMess', color: '#1976d2', icon: 'V' },
    'vless://': { name: 'VLESS', color: '#7b1fa2', icon: 'V' },
    'trojan://': { name: 'Trojan', color: '#d32f2f', icon: 'T' },
    'ss://': { name: 'Shadowsocks', color: '#2e7d32', icon: 'S' }, // Darker green
    'ssr://': { name: 'ShadowsocksR', color: '#e64a19', icon: 'R' },
    'hysteria://': { name: 'Hysteria', color: '#f9a825', icon: 'H' }, // Darker yellow
    'hysteria2://': { name: 'Hysteria2', color: '#ef6c00', icon: 'H' }, // Darker orange
    'hy2://': { name: 'Hysteria2', color: '#ef6c00', icon: 'H' },
    'tuic://': { name: 'TUIC', color: '#0277bd', icon: 'T' },
    'wireguard://': { name: 'WireGuard', color: '#455a64', icon: 'W' },
    'wg://': { name: 'WireGuard', color: '#455a64', icon: 'W' },
    'naive://': { name: 'Naive', color: '#5d4037', icon: 'N' },
    'reality://': { name: 'Reality', color: '#c2185b', icon: 'R' },
    'socks5://': { name: 'Socks5', color: '#116ea4ff', icon: 'S' },
    'socks://': { name: 'Socks', color: '#dd4984ff', icon: 'S' }
  };

  const linkLower = link.toLowerCase();
  for (const [prefix, info] of Object.entries(protocolMap)) {
    if (linkLower.startsWith(prefix)) {
      return info;
    }
  }
  return { name: '其他', color: '#616161', icon: <VpnLockIcon /> };
};

/**
 * 获取状态相关样式配置
 * 增强红黄区分度，避免使用难以辨识的浅色
 */
const getStatusStyles = (theme, colorName) => {
  const mode = theme.palette.mode;

  // 定义高对比度颜色
  const colors = {
    warning: mode === 'dark' ? '#c69800ff' : '#d19a04ff', // 深橙色用于浅色模式，确保不像红色
    error: mode === 'dark' ? '#ef5350' : '#d32f2f', // 鲜艳红
    success: mode === 'dark' ? '#66bb6a' : '#2e7d32', // 深绿
    info: mode === 'dark' ? '#4fc3f7' : '#0277bd', // 深蓝
    default: theme.palette.text.secondary
  };

  // 映射 colorName 到具体颜色
  let mainColor = colors.default;
  if (colorName === 'warning' || colorName === 'yellow') mainColor = colors.warning;
  else if (colorName === 'error') mainColor = colors.error;
  else if (colorName === 'success') mainColor = colors.success;
  else if (colorName === 'info') mainColor = colors.info;

  return {
    color: mainColor,
    bg: alpha(mainColor, 0.1),
    border: alpha(mainColor, 0.3)
  };
};

/**
 * 列表项组件 - 完全自定义布局，避免 MUI ListItemText 的 HTML 嵌套问题
 */
const DetailItem = ({ icon, label, value, isLink, onClick, secondary, noBorder }) => (
  <ListItem
    disablePadding
    sx={{
      py: 1.5,
      borderBottom: noBorder ? 'none' : '1px dashed',
      borderColor: 'divider',
      display: 'block' // 确保根元素不是 flex，以便内部 stack 能够控制
    }}
  >
    <Stack direction="row" alignItems="flex-start" spacing={2} width="100%">
      <Avatar
        sx={{
          width: 36,
          height: 36,
          bgcolor: (theme) => alpha(theme.palette.primary.main, 0.08),
          color: 'primary.main',
          borderRadius: 2,
          mt: 0.5 // 对齐微调
        }}
      >
        {icon}
      </Avatar>
      <Box sx={{ flex: 1, minWidth: 0 }}>
        <Typography variant="caption" color="text.secondary" display="block" mb={0.2}>
          {label}
        </Typography>
        <Box>
          {' '}
          {/* 使用 Box 包裹内容，避免 Typography 嵌套问题 */}
          {value ? (
            <Typography
              variant="body2"
              color={isLink ? 'primary' : 'text.primary'}
              fontWeight={500}
              sx={{
                wordBreak: 'break-all',
                cursor: onClick ? 'pointer' : 'default',
                lineHeight: 1.5,
                '&:hover': onClick ? { textDecoration: 'underline' } : {}
              }}
              onClick={onClick}
              component={isLink ? 'span' : 'p'} // 显式指定 component
            >
              {value}
            </Typography>
          ) : (
            <Typography variant="body2" color="text.disabled">
              -
            </Typography>
          )}
        </Box>
        {secondary && (
          <Box mt={0.5}>
            <Typography variant="caption" color="text.secondary" display="block">
              {secondary}
            </Typography>
          </Box>
        )}
      </Box>
    </Stack>
  </ListItem>
);

/**
 * 节点详情面板组件
 */
export default function NodeDetailsPanel({ open, node, tagColorMap, onClose, onSpeedTest, onCopy, onEdit, onDelete, onIPClick }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  if (!node) return null;

  const delayDisplay = getDelayDisplay(node.DelayTime, node.DelayStatus);
  const speedDisplay = getSpeedDisplay(node.Speed, node.SpeedStatus);
  const protocolInfo = getProtocolInfo(node.Link);

  const delayStyles = getStatusStyles(theme, delayDisplay.color);
  const speedStyles = getStatusStyles(theme, speedDisplay.color);

  return (
    <Drawer
      anchor="right"
      open={open}
      onClose={onClose}
      PaperProps={{
        sx: {
          width: isMobile ? '100%' : 400,
          maxWidth: '100vw',
          bgcolor: 'background.default'
        }
      }}
    >
      {/* 顶部区域 */}
      <Box
        sx={{
          position: 'relative',
          background: `linear-gradient(135deg, ${alpha(protocolInfo.color, 0.08)} 0%, ${theme.palette.background.paper} 100%)`,
          pb: 3,
          pt: 1,
          borderBottom: '1px solid',
          borderColor: 'divider'
        }}
      >
        {/* 关闭按钮 */}
        <Box sx={{ display: 'flex', justifyContent: 'flex-end', px: 1 }}>
          <IconButton onClick={onClose} size="medium">
            <CloseIcon />
          </IconButton>
        </Box>

        {/* 协议与名称核心展示 */}
        <Box sx={{ px: 3, textAlign: 'center' }}>
          <Box sx={{ position: 'relative', display: 'inline-block', mb: 2 }}>
            <Avatar
              sx={{
                width: 72,
                height: 72,
                bgcolor: protocolInfo.color,
                color: '#fff',
                fontSize: 32,
                fontWeight: 'bold',
                boxShadow: `0 8px 16px ${alpha(protocolInfo.color, 0.3)}`
              }}
            >
              {protocolInfo.icon}
            </Avatar>
            <Chip
              icon={<RouterIcon sx={{ fontSize: '12px !important', color: 'inherit !important' }} />}
              label={protocolInfo.name}
              size="small"
              sx={{
                position: 'absolute',
                bottom: -8,
                left: '50%',
                transform: 'translateX(-50%)',
                bgcolor: 'background.paper',
                color: protocolInfo.color,
                fontWeight: 700,
                fontSize: 10,
                height: 20,
                boxShadow: theme.shadows[2],
                border: '1px solid',
                borderColor: alpha(protocolInfo.color, 0.3),
                maxWidth: 'none', // 允许宽度超长
                '& .MuiChip-label': {
                  paddingLeft: 0.5,
                  paddingRight: 0.5,
                  display: 'block',
                  whiteSpace: 'nowrap',
                  overflow: 'visible' // 防止文字截断
                }
              }}
            />
          </Box>

          <Typography variant="h6" fontWeight="800" sx={{ mt: 1.5, lineHeight: 1.3, wordBreak: 'break-word' }}>
            {node.Name}
          </Typography>

          {node.Group && (
            <Typography variant="caption" color="text.secondary" sx={{ mt: 0.5, display: 'block', fontWeight: 500 }}>
              {node.Group}
            </Typography>
          )}

          {/* 性能指标卡片 - 增加色块背景以提升区分度 */}
          <Stack direction="row" spacing={2} sx={{ mt: 3 }}>
            <Box
              sx={{
                flex: 1,
                p: 1.5,
                borderRadius: 3,
                bgcolor: delayStyles.bg,
                border: '1px solid',
                borderColor: delayStyles.border,
                textAlign: 'left',
                position: 'relative',
                overflow: 'hidden'
              }}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mb: 0.5 }}>
                <SignalCellularAltIcon sx={{ fontSize: 16, color: delayStyles.color, opacity: 0.8 }} />
                <Typography variant="caption" fontWeight={600} sx={{ color: delayStyles.color, opacity: 0.8 }}>
                  延迟
                </Typography>
              </Box>
              <Typography variant="h5" fontWeight="800" sx={{ color: delayStyles.color }}>
                {node.DelayTime > 0 ? node.DelayTime : '-'}
                <Typography component="span" variant="caption" sx={{ ml: 0.5, color: delayStyles.color, opacity: 0.8 }}>
                  ms
                </Typography>
              </Typography>
              {node.LatencyCheckAt && (
                <Typography variant="caption" sx={{ color: 'text.secondary', fontSize: 10, display: 'block', mt: 0.5 }}>
                  {formatDateTime(node.LatencyCheckAt)}
                </Typography>
              )}
            </Box>

            <Box
              sx={{
                flex: 1,
                p: 1.5,
                borderRadius: 3,
                bgcolor: speedStyles.bg,
                border: '1px solid',
                borderColor: speedStyles.border,
                textAlign: 'left',
                position: 'relative',
                overflow: 'hidden'
              }}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mb: 0.5 }}>
                <SpeedIcon sx={{ fontSize: 16, color: speedStyles.color, opacity: 0.8 }} />
                <Typography variant="caption" fontWeight={600} sx={{ color: speedStyles.color, opacity: 0.8 }}>
                  速度
                </Typography>
              </Box>
              <Typography variant="h5" fontWeight="800" sx={{ color: speedStyles.color }}>
                {node.Speed > 0 ? node.Speed.toFixed(1) : '-'}
                <Typography component="span" variant="caption" sx={{ ml: 0.5, color: speedStyles.color, opacity: 0.8 }}>
                  MB/s
                </Typography>
              </Typography>
              {node.SpeedCheckAt && (
                <Typography variant="caption" sx={{ color: 'text.secondary', fontSize: 10, display: 'block', mt: 0.5 }}>
                  {formatDateTime(node.SpeedCheckAt)}
                </Typography>
              )}
            </Box>
          </Stack>
        </Box>
      </Box>

      {/* 滚动详情区域 */}
      <Box sx={{ flex: 1, overflowY: 'auto', px: 3, py: 2 }}>
        <List disablePadding sx={{ mb: 3 }}>
          <DetailItem
            icon={<RouterIcon fontSize="small" />}
            label="原始名称"
            value={node.LinkName || '-'}
            secondary={node.LinkName === node.Name ? '通过订阅获取的名称一致' : null}
          />
          <DetailItem icon={<SourceIcon fontSize="small" />} label="来源" value={node.Source === 'manual' ? '手动添加' : node.Source} />
          {node.DialerProxyName && <DetailItem icon={<LinkIcon fontSize="small" />} label="前置代理" value={node.DialerProxyName} />}
          {node.Tags && (
            <ListItem disablePadding sx={{ py: 1.5, borderBottom: '1px dashed', borderColor: 'divider', display: 'block' }}>
              <Stack direction="row" alignItems="flex-start" spacing={2} width="100%">
                <Avatar
                  sx={{
                    width: 36,
                    height: 36,
                    bgcolor: (theme) => alpha(theme.palette.secondary.main, 0.08),
                    color: 'secondary.main',
                    borderRadius: 2,
                    mt: 0.5
                  }}
                >
                  <FolderIcon fontSize="small" />
                </Avatar>
                <Box sx={{ flex: 1 }}>
                  <Typography variant="caption" color="text.secondary" display="block" mb={0.8}>
                    标签
                  </Typography>
                  <Stack direction="row" flexWrap="wrap" gap={0.8}>
                    {node.Tags.split(',')
                      .filter((t) => t.trim())
                      .map((tag, idx) => (
                        <Chip
                          key={idx}
                          label={tag.trim()}
                          size="small"
                          sx={{
                            bgcolor: tagColorMap?.[tag.trim()] || theme.palette.action.selected,
                            color: tagColorMap?.[tag.trim()] ? '#fff' : 'text.primary',
                            fontSize: 11,
                            height: 24,
                            border: 'none',
                            fontWeight: 500
                          }}
                        />
                      ))}
                  </Stack>
                </Box>
              </Stack>
            </ListItem>
          )}
        </List>

        <Typography
          variant="subtitle2"
          color="text.secondary"
          fontWeight={700}
          sx={{ mb: 1, fontSize: 11, textTransform: 'uppercase', letterSpacing: 1 }}
        >
          网络与状态
        </Typography>
        <List disablePadding sx={{ mb: 3 }}>
          <DetailItem
            icon={<PublicIcon fontSize="small" />}
            label="国家/地区"
            value={node.LinkCountry ? formatCountry(node.LinkCountry) : '-'}
          />
          {node.LandingIP && (
            <DetailItem
              icon={<RouterIcon fontSize="small" />}
              label="落地 IP"
              value={node.LandingIP}
              isLink
              onClick={() => onIPClick && onIPClick(node.LandingIP)}
              secondary="点击查看 IP 详细信息"
            />
          )}
          <DetailItem icon={<AccessTimeIcon fontSize="small" />} label="更新时间" value={formatDateTime(node.UpdatedAt)} noBorder />
        </List>

        {/* 占位 */}
        <Box height={80} />
      </Box>

      {/* 底部悬浮操作栏 */}
      <Paper
        elevation={8}
        sx={{
          position: 'absolute',
          bottom: 24,
          left: 24,
          right: 24,
          borderRadius: 4,
          p: 1,
          bgcolor: 'background.paper',
          display: 'flex',
          alignItems: 'center',
          gap: 1,
          border: '1px solid',
          borderColor: alpha(theme.palette.divider, 0.1)
        }}
      >
        <Button
          variant="contained"
          color="primary"
          startIcon={<PlayArrowIcon />}
          onClick={() => {
            onSpeedTest(node);
            onClose();
          }}
          sx={{
            borderRadius: 3,
            flex: 1,
            height: 48,
            fontWeight: 700,
            fontSize: 15,
            boxShadow: theme.shadows[4],
            textTransform: 'none'
          }}
        >
          立即测速
        </Button>

        <Divider orientation="vertical" flexItem sx={{ my: 1.5 }} />

        <Stack direction="row" spacing={0.5}>
          <Tooltip title="复制链接">
            <IconButton
              onClick={() => onCopy(node.Link)}
              color="primary"
              sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 2 }}
            >
              <ContentCopyIcon fontSize="small" />
            </IconButton>
          </Tooltip>

          <Tooltip title="编辑">
            <IconButton
              onClick={() => {
                onEdit(node);
                onClose();
              }}
              color="info"
              sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 2 }}
            >
              <EditIcon fontSize="small" />
            </IconButton>
          </Tooltip>

          <Tooltip title="删除">
            <IconButton
              onClick={() => {
                onDelete(node);
                onClose();
              }}
              color="error"
              sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 2 }}
            >
              <DeleteIcon fontSize="small" />
            </IconButton>
          </Tooltip>
        </Stack>
      </Paper>
    </Drawer>
  );
}

NodeDetailsPanel.propTypes = {
  open: PropTypes.bool.isRequired,
  node: PropTypes.object,
  tagColorMap: PropTypes.object,
  onClose: PropTypes.func.isRequired,
  onSpeedTest: PropTypes.func.isRequired,
  onCopy: PropTypes.func.isRequired,
  onEdit: PropTypes.func.isRequired,
  onDelete: PropTypes.func.isRequired,
  onIPClick: PropTypes.func
};
