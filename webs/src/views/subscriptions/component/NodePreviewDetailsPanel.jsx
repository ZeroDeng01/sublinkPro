import PropTypes from 'prop-types';
import { useState } from 'react';

// material-ui
import { useTheme } from '@mui/material/styles';
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
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { withAlpha } from '../../../utils/colorUtils';
import { getReadableTextTokens, getSurfaceTokens } from '../../../themes/surfaceTokens';
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

const truncateText = (text, maxLength) => {
  if (!text) return '-';
  return text.length > maxLength ? `${text.substring(0, maxLength)}...` : text;
};

/**
 * 节点预览详情面板 - 居中弹窗
 */
export default function NodePreviewDetailsPanel({ open, node, tagColorMap, onClose, onViewIP }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { isDark } = useResolvedColorScheme();
  const palette = theme.vars?.palette || theme.palette;
  const { dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(theme, isDark);
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);

  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  if (!node) return null;

  const displayName = node.PreviewName || node.Name || node.OriginalName || '未知节点';
  const protocolInfo = getProtocolPresentation(node.Protocol);
  const protocolColor = protocolInfo.color || palette.primary.main;

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
  const subtleMonoColor = isDark ? withAlpha(primaryText, 0.9) : withAlpha(palette.text.primary, 0.66);
  const secondaryLinkColor = isDark ? withAlpha(primaryText, 0.74) : tertiaryText;
  const subtleBorder = isDark ? withAlpha(palette.divider, 0.56) : withAlpha(palette.divider, 0.74);
  const elevatedSurface = isDark ? withAlpha(palette.background.paper, 0.28) : palette.background.paper;
  const accentSoftBorder = withAlpha(protocolColor, isDark ? 0.34 : 0.2);
  const previewPanelBackground = isDark
    ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.18)} 0%, ${mutedPanelSurface} 100%)`
    : 'none';
  const copyButtonSx = {
    fontSize: 10,
    py: 0.25,
    px: 1,
    minWidth: 0,
    flexShrink: 0,
    borderRadius: 1,
    borderColor: subtleBorder,
    bgcolor: nestedPanelSurface,
    color: secondaryText,
    boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none',
    '&:hover': {
      borderColor: withAlpha(palette.primary.main, isDark ? 0.42 : 0.24),
      bgcolor: withAlpha(palette.primary.main, isDark ? 0.14 : 0.06),
      color: primaryText
    }
  };

  return (
    <>
      <Dialog
        open={open}
        onClose={onClose}
        maxWidth="xs"
        fullWidth
        slotProps={{
          paper: {
            sx: {
              borderRadius: 3,
              overflow: 'hidden',
              m: isMobile ? 2 : 3,
              bgcolor: dialogSurface,
              backgroundImage: dialogSurfaceGradient,
              border: '1px solid',
              borderColor: panelBorder
            }
          }
        }}
      >
        <Box
          sx={{
            p: 2,
            position: 'relative',
            bgcolor: mutedPanelSurface,
            borderBottom: '1px solid',
            borderColor: panelBorder,
            boxShadow: `inset 0 -1px 0 ${withAlpha(palette.divider, 0.45)}`
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
              color: secondaryText,
              bgcolor: nestedPanelSurface,
              border: '1px solid',
              borderColor: subtleBorder,
              boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none',
              '&:hover': {
                color: primaryText,
                bgcolor: elevatedSurface,
                borderColor: panelBorder
              }
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
                bgcolor: withAlpha(protocolColor, isDark ? 0.18 : 0.12),
                color: protocolColor,
                border: '1px solid',
                borderColor: accentSoftBorder,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexShrink: 0
              }}
            >
              <Typography sx={{ fontSize: 30, lineHeight: 1 }}>{node.CountryFlag || '🌐'}</Typography>
            </Box>
            <Box sx={{ flex: 1, minWidth: 0, pr: 4 }}>
              <Typography sx={{ color: primaryText, fontWeight: 700, fontSize: 16, lineHeight: 1.3, wordBreak: 'break-word' }}>
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
                    bgcolor: withAlpha(protocolColor, isDark ? 0.18 : 0.12),
                    color: protocolColor,
                    border: '1px solid',
                    borderColor: accentSoftBorder
                  }}
                />
                {node.Group && <Typography sx={{ color: secondaryText, fontSize: 11 }}>{node.Group}</Typography>}
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
                bgcolor: nestedPanelSurface,
                borderColor: subtleBorder,
                boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}` : 'none'
              }}
            >
              <Stack direction="row" alignItems="center" justifyContent="center" spacing={0.5}>
                <SignalCellularAltIcon sx={{ fontSize: 12, color: 'success.main' }} />
                <Typography sx={{ fontSize: 10, color: secondaryText }}>延迟</Typography>
              </Stack>
              <Typography sx={{ fontSize: 18, fontWeight: 700, color: primaryText, mt: 0.25 }}>
                {node.DelayTime > 0 ? `${node.DelayTime}ms` : '-'}
              </Typography>
              {node.LatencyCheckAt && (
                <Typography sx={{ fontSize: 9, color: tertiaryText }}>{formatDateTime(node.LatencyCheckAt)}</Typography>
              )}
            </Paper>
            <Paper
              variant="outlined"
              sx={{
                flex: 1,
                borderRadius: 1.5,
                p: 1,
                textAlign: 'center',
                bgcolor: nestedPanelSurface,
                borderColor: subtleBorder,
                boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}` : 'none'
              }}
            >
              <Stack direction="row" alignItems="center" justifyContent="center" spacing={0.5}>
                <SpeedIcon sx={{ fontSize: 12, color: 'info.main' }} />
                <Typography sx={{ fontSize: 10, color: secondaryText }}>速度</Typography>
              </Stack>
              <Typography sx={{ fontSize: 18, fontWeight: 700, color: primaryText, mt: 0.25 }}>
                {node.Speed > 0 ? `${node.Speed.toFixed(1)}M` : '-'}
              </Typography>
              {node.SpeedCheckAt && <Typography sx={{ fontSize: 9, color: tertiaryText }}>{formatDateTime(node.SpeedCheckAt)}</Typography>}
            </Paper>
          </Stack>
        </Box>

        <DialogContent sx={{ p: 2, bgcolor: dialogSurface }}>
          {/* 名称转换 - 紧凑 */}
          {node.OriginalName && node.OriginalName !== displayName && (
            <Box
              sx={{
                mb: 1.5,
                p: 1,
                bgcolor: withAlpha(palette.primary.main, isDark ? 0.14 : 0.05),
                borderRadius: 1.5,
                border: '1px solid',
                borderColor: withAlpha(palette.primary.main, isDark ? 0.3 : 0.18)
              }}
            >
              <Typography variant="caption" color="primary" fontWeight={600}>
                名称转换
              </Typography>
              <Stack direction="row" alignItems="center" spacing={0.5} mt={0.25}>
                <Typography sx={{ fontSize: 11, color: secondaryText }} noWrap>
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
            sx={{ mb: 1.5, p: 1.25, borderRadius: 1.5, bgcolor: mutedPanelSurface, border: '1px solid', borderColor: panelBorder }}
          >
            <Grid item xs={6}>
              <Typography variant="caption" sx={{ color: secondaryText }}>
                来源
              </Typography>
              <Typography variant="body2" fontWeight={500} sx={{ color: primaryText }}>
                {node.Source === 'manual' ? '手动添加' : node.Source || '-'}
              </Typography>
            </Grid>
            <Grid item xs={6}>
              <Typography variant="caption" sx={{ color: secondaryText }}>
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
                <Typography variant="body2" fontWeight={500} sx={{ color: primaryText }}>
                  -
                </Typography>
              )}
            </Grid>
          </Grid>

          {/* 标签 */}
          {tags.length > 0 && (
            <Box sx={{ mb: 1.5 }}>
              <Typography variant="caption" sx={{ color: secondaryText }}>
                标签
              </Typography>
              <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap sx={{ mt: 0.5 }}>
                {tags.map((tag, idx) => {
                  const tagName = tag.trim();
                  const tagColor = tagColorMap?.[tagName];
                  const chipAccent = tagColor || palette.primary.main;
                  return (
                    <Chip
                      key={idx}
                      label={tagName}
                      size="small"
                      sx={{
                        height: 22,
                        fontSize: 11,
                        bgcolor: withAlpha(chipAccent, isDark ? 0.18 : 0.1),
                        color: tagColor ? chipAccent : primaryText,
                        border: '1px solid',
                        borderColor: withAlpha(chipAccent, isDark ? 0.34 : 0.2),
                        '& .MuiChip-label': {
                          px: 0.9,
                          fontWeight: 600
                        }
                      }}
                    />
                  );
                })}
              </Stack>
            </Box>
          )}

          <Divider sx={{ my: 1.5, borderColor: subtleBorder }} />

          <Box
            sx={{
              p: 1.25,
              borderRadius: 1.5,
              bgcolor: mutedPanelSurface,
              backgroundImage: previewPanelBackground,
              border: '1px solid',
              borderColor: panelBorder,
              boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}` : 'none'
            }}
          >
            <Stack direction="row" alignItems="center" spacing={1}>
              <Typography variant="caption" sx={{ flexShrink: 0, fontWeight: 600, color: secondaryText }}>
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
                {truncateText(previewLink, 50)}
              </Typography>
              <Button
                size="small"
                variant="outlined"
                startIcon={<ContentCopyIcon sx={{ fontSize: 12 }} />}
                onClick={() => copyToClipboard(previewLink, '链接')}
                sx={copyButtonSx}
              >
                复制
              </Button>
            </Stack>

            {node.PreviewLink && node.PreviewLink !== node.Link && (
              <Stack
                direction="row"
                alignItems="center"
                spacing={1}
                sx={{ mt: 1, pt: 1, borderTop: '1px dashed', borderColor: subtleBorder }}
              >
                <Typography variant="caption" sx={{ flexShrink: 0, fontWeight: 600, color: secondaryText }}>
                  原始链接
                </Typography>
                <Typography sx={{ flex: 1, fontSize: 9, color: secondaryLinkColor, fontFamily: 'monospace' }} noWrap>
                  {truncateText(node.Link, 40)}
                </Typography>
                <Button
                  size="small"
                  color="inherit"
                  onClick={() => copyToClipboard(node.Link, '原始链接')}
                  sx={{
                    ...copyButtonSx,
                    fontSize: 9,
                    py: 0,
                    px: 0.75
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
