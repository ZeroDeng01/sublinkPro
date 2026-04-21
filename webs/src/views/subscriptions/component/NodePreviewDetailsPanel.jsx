import PropTypes from 'prop-types';
import { useState } from 'react';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogContent from '@mui/material/DialogContent';
import Divider from '@mui/material/Divider';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import { getProtocolPresentation } from '../../../utils/protocolPresentation';

// icons
import CloseIcon from '@mui/icons-material/Close';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import SignalCellularAltIcon from '@mui/icons-material/SignalCellularAlt';
import SpeedIcon from '@mui/icons-material/Speed';
import ArrowForwardIcon from '@mui/icons-material/ArrowForward';

/**
 * 格式化时间
 */
const formatDateTime = (dateStr) => {
  if (!dateStr) return '-';
  try {
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' });
  } catch {
    return dateStr;
  }
};

/**
 * 节点预览详情面板 - 居中弹窗
 */
export default function NodePreviewDetailsPanel({ open, node, tagColorMap, onClose, onViewIP }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  if (!node) return null;

  const displayName = node.PreviewName || node.Name || node.OriginalName || '未知节点';
  const protocolInfo = getProtocolPresentation(node.Protocol);
  const protocolColor = protocolInfo.color || theme.palette.primary.main;

  // 复制到剪贴板
  const copyToClipboard = async (text, label) => {
    try {
      await navigator.clipboard.writeText(text);
      setSnackbar({ open: true, message: `${label}已复制`, severity: 'success' });
    } catch {
      setSnackbar({ open: true, message: '复制失败', severity: 'error' });
    }
  };

  // 标签列表
  const tags = node.Tags ? node.Tags.split(',').filter((t) => t.trim()) : [];
  const previewLink = node.PreviewLink || node.Link || '';
  const subtleMonoColor = alpha(theme.palette.text.primary, theme.palette.mode === 'dark' ? 0.86 : 0.66);
  const secondaryLinkColor = alpha(theme.palette.text.primary, theme.palette.mode === 'dark' ? 0.76 : 0.58);

  return (
    <>
      <Dialog
        open={open}
        onClose={onClose}
        maxWidth="xs"
        fullWidth
        PaperProps={{
          sx: {
            borderRadius: 3,
            overflow: 'hidden',
            m: isMobile ? 2 : 3,
            bgcolor: 'background.paper',
            backgroundImage: 'none',
            border: '1px solid',
            borderColor: 'divider'
          }
        }}
      >
        <Box
          sx={{
            p: 2,
            position: 'relative',
            bgcolor: 'background.default',
            borderBottom: '1px solid',
            borderColor: 'divider',
            boxShadow: `inset 0 -1px 0 ${alpha(theme.palette.divider, 0.45)}`
          }}
        >
          {/* 关闭按钮 */}
          <IconButton
            onClick={onClose}
            size="small"
            sx={{
              position: 'absolute',
              top: 8,
              right: 8,
              color: 'text.secondary',
              bgcolor: alpha(theme.palette.background.paper, 0.9),
              border: '1px solid',
              borderColor: alpha(theme.palette.divider, 0.9),
              '&:hover': { bgcolor: 'background.paper' }
            }}
          >
            <CloseIcon fontSize="small" />
          </IconButton>

          {/* 头部信息 */}
          <Stack direction="row" alignItems="center" spacing={1.5}>
            <Box
              sx={{
                width: 52,
                height: 52,
                borderRadius: 2,
                bgcolor: alpha(protocolColor, 0.12),
                color: protocolColor,
                border: '1px solid',
                borderColor: alpha(protocolColor, 0.22),
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexShrink: 0
              }}
            >
              <Typography sx={{ fontSize: 30, lineHeight: 1 }}>{node.CountryFlag || '🌐'}</Typography>
            </Box>
            <Box sx={{ flex: 1, minWidth: 0, pr: 4 }}>
              <Typography sx={{ color: 'text.primary', fontWeight: 700, fontSize: 16, lineHeight: 1.3, wordBreak: 'break-word' }}>
                {displayName}
              </Typography>
              <Stack direction="row" alignItems="center" spacing={0.75} mt={0.5}>
                <Chip
                  label={node.Protocol || '未知'}
                  size="small"
                  sx={{
                    height: 20,
                    fontSize: 10,
                    fontWeight: 600,
                    bgcolor: alpha(protocolColor, 0.12),
                    color: protocolColor,
                    border: '1px solid',
                    borderColor: alpha(protocolColor, 0.18)
                  }}
                />
                {node.Group && <Typography sx={{ color: 'text.secondary', fontSize: 11 }}>{node.Group}</Typography>}
              </Stack>
            </Box>
          </Stack>

          {/* 性能指标 - 紧凑横向 */}
          <Stack direction="row" spacing={1} sx={{ mt: 1.5 }}>
            <Paper
              variant="outlined"
              sx={{
                flex: 1,
                borderRadius: 1.5,
                p: 1,
                textAlign: 'center',
                bgcolor: 'background.paper',
                borderColor: alpha(theme.palette.divider, 0.9)
              }}
            >
              <Stack direction="row" alignItems="center" justifyContent="center" spacing={0.5}>
                <SignalCellularAltIcon sx={{ fontSize: 12, color: 'success.main' }} />
                <Typography sx={{ fontSize: 10, color: 'text.secondary' }}>延迟</Typography>
              </Stack>
              <Typography sx={{ fontSize: 18, fontWeight: 700, color: 'text.primary', mt: 0.25 }}>
                {node.DelayTime > 0 ? `${node.DelayTime}ms` : '-'}
              </Typography>
              {node.LatencyCheckAt && (
                <Typography sx={{ fontSize: 9, color: 'text.secondary' }}>{formatDateTime(node.LatencyCheckAt)}</Typography>
              )}
            </Paper>
            <Paper
              variant="outlined"
              sx={{
                flex: 1,
                borderRadius: 1.5,
                p: 1,
                textAlign: 'center',
                bgcolor: 'background.paper',
                borderColor: alpha(theme.palette.divider, 0.9)
              }}
            >
              <Stack direction="row" alignItems="center" justifyContent="center" spacing={0.5}>
                <SpeedIcon sx={{ fontSize: 12, color: 'info.main' }} />
                <Typography sx={{ fontSize: 10, color: 'text.secondary' }}>速度</Typography>
              </Stack>
              <Typography sx={{ fontSize: 18, fontWeight: 700, color: 'text.primary', mt: 0.25 }}>
                {node.Speed > 0 ? `${node.Speed.toFixed(1)}M` : '-'}
              </Typography>
              {node.SpeedCheckAt && (
                <Typography sx={{ fontSize: 9, color: 'text.secondary' }}>{formatDateTime(node.SpeedCheckAt)}</Typography>
              )}
            </Paper>
          </Stack>
        </Box>

        <DialogContent sx={{ p: 2, bgcolor: 'background.paper' }}>
          {/* 名称转换 - 紧凑 */}
          {node.OriginalName && node.OriginalName !== displayName && (
            <Box
              sx={{
                mb: 1.5,
                p: 1,
                bgcolor: 'background.default',
                borderRadius: 1.5,
                border: '1px solid',
                borderColor: alpha(theme.palette.primary.main, 0.18)
              }}
            >
              <Typography variant="caption" color="primary" fontWeight={600}>
                名称转换
              </Typography>
              <Stack direction="row" alignItems="center" spacing={0.5} mt={0.25}>
                <Typography sx={{ fontSize: 11, color: 'text.secondary' }} noWrap>
                  {node.OriginalName}
                </Typography>
                <ArrowForwardIcon sx={{ fontSize: 12, color: 'primary.main', flexShrink: 0 }} />
                <Typography sx={{ fontSize: 11, color: 'primary.main', fontWeight: 600 }} noWrap>
                  {displayName}
                </Typography>
              </Stack>
            </Box>
          )}

          {/* 基本信息 - 一行显示 */}
          <Grid
            container
            spacing={1}
            sx={{ mb: 1.5, p: 1.25, borderRadius: 1.5, bgcolor: 'background.default', border: '1px solid', borderColor: 'divider' }}
          >
            <Grid item xs={6}>
              <Typography variant="caption" color="text.secondary">
                来源
              </Typography>
              <Typography variant="body2" fontWeight={500}>
                {node.Source === 'manual' ? '手动添加' : node.Source || '-'}
              </Typography>
            </Grid>
            <Grid item xs={6}>
              <Typography variant="caption" color="text.secondary">
                落地IP
              </Typography>
              {node.LandingIP ? (
                <Typography
                  variant="body2"
                  fontWeight={500}
                  onClick={(e) => {
                    e.stopPropagation();
                    onViewIP?.(node.LandingIP);
                  }}
                  sx={{
                    cursor: 'pointer',
                    color: 'primary.main',
                    '&:hover': { textDecoration: 'underline' }
                  }}
                >
                  {node.LandingIP.length > 15 ? node.LandingIP.substring(0, 15) + '...' : node.LandingIP}
                </Typography>
              ) : (
                <Typography variant="body2" fontWeight={500}>
                  -
                </Typography>
              )}
            </Grid>
          </Grid>

          {/* 标签 */}
          {tags.length > 0 && (
            <Box sx={{ mb: 1.5 }}>
              <Typography variant="caption" color="text.secondary">
                标签
              </Typography>
              <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap sx={{ mt: 0.5 }}>
                {tags.map((tag, idx) => {
                  const tagName = tag.trim();
                  const tagColor = tagColorMap?.[tagName];
                  return (
                    <Chip
                      key={idx}
                      label={tagName}
                      size="small"
                      sx={{
                        height: 22,
                        fontSize: 11,
                        bgcolor: tagColor || alpha(theme.palette.primary.main, 0.1),
                        color: tagColor ? theme.palette.getContrastText(tagColor) : 'text.primary'
                      }}
                    />
                  );
                })}
              </Stack>
            </Box>
          )}

          <Divider sx={{ my: 1.5 }} />

          <Box
            sx={{
              p: 1.25,
              borderRadius: 1.5,
              bgcolor: theme.palette.mode === 'dark' ? alpha(theme.palette.background.default, 0.84) : 'background.default',
              border: '1px solid',
              borderColor: alpha(theme.palette.divider, theme.palette.mode === 'dark' ? 0.82 : 0.72)
            }}
          >
            <Stack direction="row" alignItems="center" spacing={1}>
              <Typography variant="caption" color="text.secondary" sx={{ flexShrink: 0, fontWeight: 600 }}>
                预览链接
              </Typography>
              <Typography
                sx={{
                  flex: 1,
                  fontSize: 10,
                  color: subtleMonoColor,
                  fontFamily: 'monospace',
                  fontWeight: 500,
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap'
                }}
              >
                {previewLink.substring(0, 50)}...
              </Typography>
              <Button
                size="small"
                variant="outlined"
                startIcon={<ContentCopyIcon sx={{ fontSize: 12 }} />}
                onClick={() => copyToClipboard(previewLink, '链接')}
                sx={{
                  fontSize: 10,
                  py: 0.25,
                  px: 1,
                  minWidth: 0,
                  flexShrink: 0,
                  borderColor: alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.3 : 0.18),
                  color: theme.palette.mode === 'dark' ? alpha(theme.palette.text.primary, 0.9) : 'primary.main'
                }}
              >
                复制
              </Button>
            </Stack>

            {node.PreviewLink && node.PreviewLink !== node.Link && (
              <Stack
                direction="row"
                alignItems="center"
                spacing={1}
                sx={{ mt: 1, pt: 1, borderTop: '1px dashed', borderColor: alpha(theme.palette.divider, 0.72) }}
              >
                <Typography variant="caption" color="text.secondary" sx={{ flexShrink: 0, fontWeight: 600 }}>
                  原始链接
                </Typography>
                <Typography sx={{ flex: 1, fontSize: 9, color: secondaryLinkColor, fontFamily: 'monospace' }} noWrap>
                  {node.Link?.substring(0, 40)}...
                </Typography>
                <Button
                  size="small"
                  color="inherit"
                  onClick={() => copyToClipboard(node.Link, '原始链接')}
                  sx={{
                    fontSize: 9,
                    py: 0,
                    minWidth: 0,
                    color: theme.palette.mode === 'dark' ? alpha(theme.palette.text.primary, 0.82) : 'text.secondary'
                  }}
                >
                  复制
                </Button>
              </Stack>
            )}
          </Box>
        </DialogContent>
      </Dialog>

      {/* 复制成功提示 */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={1500}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity} variant="filled">
          {snackbar.message}
        </Alert>
      </Snackbar>
    </>
  );
}

NodePreviewDetailsPanel.propTypes = {
  open: PropTypes.bool.isRequired,
  node: PropTypes.object,
  tagColorMap: PropTypes.object,
  onClose: PropTypes.func.isRequired,
  onViewIP: PropTypes.func
};
