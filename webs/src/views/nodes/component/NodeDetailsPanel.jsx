import PropTypes from 'prop-types';
import { useState } from 'react';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Accordion from '@mui/material/Accordion';
import AccordionSummary from '@mui/material/AccordionSummary';
import AccordionDetails from '@mui/material/AccordionDetails';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';

import Drawer from '@mui/material/Drawer';
import IconButton from '@mui/material/IconButton';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import Stack from '@mui/material/Stack';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

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

import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import CodeIcon from '@mui/icons-material/Code';

// dialog
import Dialog from '@mui/material/Dialog';

import Zoom from '@mui/material/Zoom';

// utils
import { formatDateTime, formatCountry, getDelayDisplay, getSpeedDisplay } from '../utils';

// components
import NodeRawInfoEditor from './NodeRawInfoEditor';

/**
 * 解析节点协议类型
 * 支持使用后端协议元数据或本地映射
 */
const getProtocolInfo = (link, protocolMeta) => {
  if (!link) return { name: '未知', color: '#9e9e9e', icon: <FilterVintageIcon /> };

  // 如果有后端协议元数据，优先使用
  if (protocolMeta && protocolMeta.length > 0) {
    const linkLower = link.toLowerCase();
    for (const proto of protocolMeta) {
      if (linkLower.startsWith(proto.name + '://') || (proto.name === 'hysteria2' && linkLower.startsWith('hy2://'))) {
        return {
          name: proto.label,
          color: proto.color || '#616161',
          icon: proto.icon || proto.label.charAt(0).toUpperCase()
        };
      }
    }
  }

  // 后备的本地映射
  const protocolMap = {
    'vmess://': { name: 'VMess', color: '#1976d2', icon: 'V' },
    'vless://': { name: 'VLESS', color: '#7b1fa2', icon: 'V' },
    'trojan://': { name: 'Trojan', color: '#d32f2f', icon: 'T' },
    'ss://': { name: 'Shadowsocks', color: '#2e7d32', icon: 'S' },
    'ssr://': { name: 'ShadowsocksR', color: '#e64a19', icon: 'R' },
    'hysteria://': { name: 'Hysteria', color: '#f9a825', icon: 'H' },
    'hysteria2://': { name: 'Hysteria2', color: '#ef6c00', icon: 'H' },
    'hy2://': { name: 'Hysteria2', color: '#ef6c00', icon: 'H' },
    'tuic://': { name: 'TUIC', color: '#0277bd', icon: 'T' },
    'wireguard://': { name: 'WireGuard', color: '#455a64', icon: 'W' },
    'wg://': { name: 'WireGuard', color: '#455a64', icon: 'W' },
    'naive://': { name: 'Naive', color: '#5d4037', icon: 'N' },
    'reality://': { name: 'Reality', color: '#c2185b', icon: 'R' },
    'socks5://': { name: 'Socks5', color: '#116ea4ff', icon: 'S' },
    'socks://': { name: 'Socks', color: '#dd4984ff', icon: 'S' },
    'anytls://': { name: 'AnyTLS', color: '#20a84c', icon: 'A' }
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
export default function NodeDetailsPanel({
  open,
  node,
  tagColorMap,
  protocolMeta,
  onClose,
  onSpeedTest,
  onCopy,
  onEdit,
  onDelete,
  onIPClick,
  onNodeUpdate,
  showMessage
}) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const [rawInfoExpanded, setRawInfoExpanded] = useState(false);

  if (!node) return null;

  const delayDisplay = getDelayDisplay(node.DelayTime, node.DelayStatus);
  const speedDisplay = getSpeedDisplay(node.Speed, node.SpeedStatus);
  const protocolInfo = getProtocolInfo(node.Link, protocolMeta);

  const delayStyles = getStatusStyles(theme, delayDisplay.color);
  const speedStyles = getStatusStyles(theme, speedDisplay.color);

  // Common content to be reused in both Dialog and Drawer
  const NodeContent = (
    <>
      {/* 顶部区域 */}
      <Box
        sx={{
          position: 'relative',
          background: `linear-gradient(135deg, ${alpha(protocolInfo.color, 0.08)} 0%, ${theme.palette.background.paper} 100%)`,
          pb: 3,
          pt: isMobile ? 2 : 3,
          px: 3,
          borderBottom: '1px solid',
          borderColor: 'divider',
          flexShrink: 0 // Prevent shrinking
        }}
      >
        {/* 关闭按钮 (Only needed if not using DialogTitle/Actions standard close in mobile or custom layout) */}
        {!isMobile && (
          <IconButton onClick={onClose} sx={{ position: 'absolute', right: 16, top: 16, color: 'text.secondary' }}>
            <CloseIcon />
          </IconButton>
        )}

        {/* Mobile Swipe Indicator (Optional visual cue) */}
        {isMobile && (
          <Box
            sx={{
              width: 40,
              height: 4,
              bgcolor: 'divider',
              borderRadius: 2,
              mx: 'auto',
              mb: 2,
              opacity: 0.6
            }}
          />
        )}

        {/* 协议与名称核心展示 */}
        <Box sx={{ textAlign: 'center' }}>
          <Box sx={{ position: 'relative', display: 'inline-block', mb: 2 }}>
            <Avatar
              sx={{
                width: 80,
                height: 80,
                bgcolor: protocolInfo.color,
                color: '#fff',
                fontSize: 36,
                fontWeight: 'bold',
                boxShadow: `0 8px 24px ${alpha(protocolInfo.color, 0.25)}`,
                border: `4px solid ${theme.palette.background.paper}`
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
                bottom: -10,
                left: '50%',
                transform: 'translateX(-50%)',
                bgcolor: 'background.paper',
                color: protocolInfo.color,
                fontWeight: 700,
                fontSize: 11,
                height: 22,
                boxShadow: theme.shadows[2],
                border: '1px solid',
                borderColor: alpha(protocolInfo.color, 0.2),
                maxWidth: 'none',
                '& .MuiChip-label': {
                  px: 1,
                  display: 'block',
                  whiteSpace: 'nowrap',
                  overflow: 'visible'
                }
              }}
            />
          </Box>

          <Typography variant="h5" fontWeight="800" sx={{ mt: 2, mb: 0.5, lineHeight: 1.3, wordBreak: 'break-word' }}>
            {node.Name}
          </Typography>

          {node.Group && (
            <Chip
              label={node.Group}
              size="small"
              variant="outlined"
              sx={{
                color: 'text.secondary',
                borderColor: 'divider',
                height: 20,
                fontSize: 11,
                fontWeight: 500
              }}
            />
          )}

          {/* 性能指标卡片 */}
          <Stack direction="row" spacing={2} sx={{ mt: 3 }}>
            <Box
              sx={{
                flex: 1,
                p: 1.5,
                borderRadius: 4,
                bgcolor: delayStyles.bg,
                border: '1px solid',
                borderColor: delayStyles.border,
                textAlign: 'center',
                position: 'relative',
                overflow: 'hidden'
              }}
            >
              <Typography variant="caption" fontWeight={600} sx={{ color: delayStyles.color, opacity: 0.9, display: 'block', mb: 0.5 }}>
                延迟
              </Typography>
              <Typography variant="h5" fontWeight="800" sx={{ color: delayStyles.color }}>
                {node.DelayTime > 0 ? node.DelayTime : '-'}
                <Typography component="span" variant="caption" sx={{ ml: 0.5, opacity: 0.8 }}>
                  ms
                </Typography>
              </Typography>
            </Box>

            <Box
              sx={{
                flex: 1,
                p: 1.5,
                borderRadius: 4,
                bgcolor: speedStyles.bg,
                border: '1px solid',
                borderColor: speedStyles.border,
                textAlign: 'center',
                position: 'relative',
                overflow: 'hidden'
              }}
            >
              <Typography variant="caption" fontWeight={600} sx={{ color: speedStyles.color, opacity: 0.9, display: 'block', mb: 0.5 }}>
                速度
              </Typography>
              <Typography variant="h5" fontWeight="800" sx={{ color: speedStyles.color }}>
                {node.Speed > 0 ? node.Speed.toFixed(1) : '-'}
                <Typography component="span" variant="caption" sx={{ ml: 0.5, opacity: 0.8 }}>
                  MB/s
                </Typography>
              </Typography>
            </Box>
          </Stack>
        </Box>
      </Box>

      {/* 滚动详情区域 */}
      <Box sx={{ flex: 1, overflowY: 'auto', px: 3, py: 2 }}>
        <List disablePadding sx={{ mb: 3 }}>
          <ListItem disablePadding sx={{ py: 1.5, borderBottom: '1px dashed', borderColor: 'divider', display: 'block' }}>
            <Stack direction="row" alignItems="flex-start" spacing={2} width="100%">
              <Box
                sx={{
                  minWidth: 36,
                  height: 36,
                  borderRadius: 12,
                  bgcolor: (theme) => alpha(theme.palette.primary.main, 0.1),
                  color: 'primary.main',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  mt: 0.5
                }}
              >
                <RouterIcon fontSize="small" />
              </Box>
              <Box sx={{ flex: 1 }}>
                <Typography variant="caption" color="text.secondary" display="block" mb={0.5}>
                  原始名称
                </Typography>
                <Typography variant="body2" sx={{ wordBreak: 'break-word', fontWeight: 500 }}>
                  {node.LinkName || '-'}
                </Typography>
                {node.LinkName === node.Name && (
                  <Typography variant="caption" color="text.secondary" display="block" mt={0.3} sx={{ fontSize: 11 }}>
                    名称与订阅一致
                  </Typography>
                )}
              </Box>
            </Stack>
          </ListItem>

          <DetailItem icon={<SourceIcon fontSize="small" />} label="来源" value={node.Source === 'manual' ? '手动添加' : node.Source} />
          {node.DialerProxyName && <DetailItem icon={<LinkIcon fontSize="small" />} label="前置代理" value={node.DialerProxyName} />}

          {node.Tags && (
            <ListItem disablePadding sx={{ py: 1.5, borderBottom: '1px dashed', borderColor: 'divider', display: 'block' }}>
              <Stack direction="row" alignItems="flex-start" spacing={2} width="100%">
                <Box
                  sx={{
                    minWidth: 36,
                    height: 36,
                    borderRadius: 12,
                    bgcolor: (theme) => alpha(theme.palette.secondary.main, 0.1),
                    color: 'secondary.main',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    mt: 0.5
                  }}
                >
                  <FolderIcon fontSize="small" />
                </Box>
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
                            fontWeight: 600,
                            borderRadius: 1.5
                          }}
                        />
                      ))}
                  </Stack>
                </Box>
              </Stack>
            </ListItem>
          )}
        </List>

        {/* 原始协议信息区域 */}
        <Accordion
          expanded={rawInfoExpanded}
          onChange={() => setRawInfoExpanded(!rawInfoExpanded)}
          disableGutters
          elevation={0}
          sx={{
            bgcolor: 'transparent',
            '&:before': { display: 'none' },
            border: '1px solid',
            borderColor: 'divider',
            borderRadius: 3,
            mb: 3,
            overflow: 'hidden'
          }}
        >
          <AccordionSummary
            expandIcon={<ExpandMoreIcon />}
            sx={{
              minHeight: 48,
              '& .MuiAccordionSummary-content': { my: 1 }
            }}
          >
            <Stack direction="row" alignItems="center" spacing={1}>
              <CodeIcon fontSize="small" color="primary" />
              <Typography variant="subtitle2" fontWeight={600}>
                原始协议信息
              </Typography>
            </Stack>
          </AccordionSummary>
          <AccordionDetails sx={{ pt: 0 }}>
            <NodeRawInfoEditor node={node} protocolMeta={protocolMeta} onUpdate={onNodeUpdate} showMessage={showMessage} />
          </AccordionDetails>
        </Accordion>

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
      </Box>
    </>
  );

  // Common Action Bar
  const ActionBar = (
    <Box
      sx={{
        p: 2,
        pb: isMobile ? 3 : 2, // Extra padding for bottom safe area on mobile
        bgcolor: 'background.paper',
        borderTop: '1px solid',
        borderColor: 'divider',
        display: 'flex',
        alignItems: 'center',
        gap: 1.5,
        zIndex: 10
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
        fullWidth
        sx={{
          borderRadius: 3,
          height: 48,
          fontWeight: 700,
          fontSize: 15,
          boxShadow: theme.shadows[4],
          textTransform: 'none'
        }}
      >
        立即检测
      </Button>

      <Stack direction="row" spacing={1}>
        <Tooltip title="复制链接">
          <IconButton
            onClick={() => onCopy(node.Link)}
            color="primary"
            sx={{
              border: '1px solid',
              borderColor: 'divider',
              borderRadius: 3,
              width: 48,
              height: 48
            }}
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
            sx={{
              border: '1px solid',
              borderColor: 'divider',
              borderRadius: 3,
              width: 48,
              height: 48
            }}
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
            sx={{
              border: '1px solid',
              borderColor: 'divider',
              borderRadius: 3,
              width: 48,
              height: 48
            }}
          >
            <DeleteIcon fontSize="small" />
          </IconButton>
        </Tooltip>
      </Stack>
    </Box>
  );

  return (
    <>
      {isMobile ? (
        <Drawer
          anchor="bottom"
          open={open}
          onClose={onClose}
          PaperProps={{
            sx: {
              borderTopLeftRadius: 24,
              borderTopRightRadius: 24,
              maxHeight: '85vh',
              overflow: 'hidden', // Let children scroll
              display: 'flex',
              flexDirection: 'column'
            }
          }}
        >
          {NodeContent}
          {ActionBar}
        </Drawer>
      ) : (
        <Dialog
          open={open}
          onClose={onClose}
          maxWidth="sm"
          fullWidth
          TransitionComponent={Zoom}
          PaperProps={{
            sx: {
              borderRadius: 4,
              overflow: 'hidden',
              bgcolor: 'background.paper',
              backgroundImage: 'none',
              display: 'flex',
              flexDirection: 'column',
              maxHeight: 'calc(100% - 64px)'
            }
          }}
        >
          {NodeContent}
          {ActionBar}
        </Dialog>
      )}
    </>
  );
}

NodeDetailsPanel.propTypes = {
  open: PropTypes.bool.isRequired,
  node: PropTypes.object,
  tagColorMap: PropTypes.object,
  protocolMeta: PropTypes.array, // 协议元数据列表（从后端获取）
  onClose: PropTypes.func.isRequired,
  onSpeedTest: PropTypes.func.isRequired,
  onCopy: PropTypes.func.isRequired,
  onEdit: PropTypes.func.isRequired,
  onDelete: PropTypes.func.isRequired,
  onIPClick: PropTypes.func,
  onNodeUpdate: PropTypes.func, // 节点更新后的回调
  showMessage: PropTypes.func // 消息提示函数
};
