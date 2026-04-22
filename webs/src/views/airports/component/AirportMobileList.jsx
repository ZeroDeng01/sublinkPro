import { useState } from 'react';
import PropTypes from 'prop-types';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Checkbox from '@mui/material/Checkbox';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import Snackbar from '@mui/material/Snackbar';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

// icons
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import LanIcon from '@mui/icons-material/Lan';
import MoreVertIcon from '@mui/icons-material/MoreVert';
import DownloadIcon from '@mui/icons-material/Download';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import SpeedIcon from '@mui/icons-material/Speed';
import AccountBalanceWalletIcon from '@mui/icons-material/AccountBalanceWallet';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';
import EventIcon from '@mui/icons-material/Event';

// utils
import { formatDateTime, formatBytes, formatExpireTime, getUsageColor } from '../utils';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { withAlpha } from 'utils/colorUtils';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';

// local components
import AirportNodeStatsCard from './AirportNodeStatsCard';
import AirportLogo from './AirportLogo';

/**
 * 机场移动端列表组件
 */
export default function AirportMobileList({
  airports,
  selectedIds,
  onToggleSelect,
  onEdit,
  onDelete,
  onPull,
  onOpenNodes,
  onQuickCheck,
  onRefreshUsage
}) {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { palette, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(theme, isDark);
  const { primaryText, secondaryText } = getReadableTextTokens(theme, isDark);
  const cardSurface = isDark ? withAlpha(palette.background.paper, 0.94) : palette.background.paper;
  const selectedCardSurface = isDark
    ? `linear-gradient(180deg, ${withAlpha(palette.primary.main, 0.12)} 0%, ${cardSurface} 100%)`
    : withAlpha(palette.primary.main, 0.04);

  // 复制提示状态
  const [copyTip, setCopyTip] = useState({ open: false, name: '' });
  const [menuAnchor, setMenuAnchor] = useState(null);
  const [menuAirportId, setMenuAirportId] = useState(null);

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

  const handleOpenMenu = (event, airportId) => {
    event.stopPropagation();
    setMenuAnchor(event.currentTarget);
    setMenuAirportId(airportId);
  };

  const handleCloseMenu = () => {
    setMenuAnchor(null);
    setMenuAirportId(null);
  };

  const handleMenuAction = (action, airport) => {
    handleCloseMenu();
    action(airport);
  };

  if (airports.length === 0) {
    return (
      <Box sx={{ py: 6, textAlign: 'center' }}>
        <Typography variant="body2" color="text.secondary">
          暂无机场数据，点击上方"添加"按钮添加
        </Typography>
      </Box>
    );
  }

  /**
   * 计算顶部状态条颜色
   * 优先级：禁用(灰色) > 用量警告(红色) > 过期警告(红色) > 启用(绿色)
   */
  const getStatusAccent = (airport) => {
    // 禁用状态
    if (!airport.enabled) {
      return theme.palette.text.secondary;
    }

    // 检查用量警告（使用率 >= 85%）
    if (airport.fetchUsageInfo && airport.usageTotal > 0) {
      const upload = airport.usageUpload || 0;
      const download = airport.usageDownload || 0;
      const used = upload + download;
      const percent = (used / airport.usageTotal) * 100;
      if (percent >= 85) {
        return theme.palette.error.main;
      }
    }

    // 检查过期警告（7天内过期）
    if (airport.usageExpire > 0) {
      const now = Math.floor(Date.now() / 1000);
      const daysLeft = (airport.usageExpire - now) / (24 * 60 * 60);
      if (daysLeft <= 7) {
        return theme.palette.error.main;
      }
    }

    // 正常启用状态
    return theme.palette.success.main;
  };

  const getProgressTrackColor = (percent) => alpha(getUsageColor(percent), isDark ? 0.22 : 0.12);

  const getStatusChipSx = (enabled) => ({
    height: 20,
    fontSize: '0.72rem',
    fontWeight: 600,
    color: enabled ? (isDark ? palette.success.light : palette.success.dark) : secondaryText,
    bgcolor: enabled ? withAlpha(palette.success.main, isDark ? 0.18 : 0.12) : mutedPanelSurface,
    border: `1px solid ${enabled ? withAlpha(palette.success.main, isDark ? 0.34 : 0.18) : panelBorder}`,
    '& .MuiChip-label': {
      px: 1
    }
  });

  return (
    <>
      <Stack spacing={2.5}>
        {airports.map((airport) => {
          const isSelected = selectedIds.includes(airport.id);
          const statusAccent = getStatusAccent(airport);

          return (
            <Card
              key={airport.id}
              sx={{
                borderRadius: 3,
                border: `1px solid ${isSelected ? withAlpha(palette.primary.main, isDark ? 0.42 : 0.24) : panelBorder}`,
                boxShadow: isDark
                  ? `0 4px 12px ${alpha(theme.palette.common.black, 0.24)}`
                  : `0 4px 12px ${alpha(theme.palette.primary.main, isSelected ? 0.14 : 0.08)}`,
                bgcolor: isSelected ? selectedCardSurface : cardSurface,
                transition: 'all 0.2s ease',
                overflow: 'hidden',
                position: 'relative',
                '&:hover': {
                  transform: 'translateY(-1px)',
                  boxShadow: isDark
                    ? `0 8px 24px ${alpha(theme.palette.common.black, 0.3)}`
                    : `0 8px 24px ${alpha(theme.palette.primary.main, 0.15)}`
                },
                // 顶部状态指示条
                '&::before': {
                  content: '""',
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  right: 0,
                  height: 4,
                  backgroundColor: statusAccent
                }
              }}
            >
              <CardContent sx={{ p: 2, pt: 2.5, '&:last-child': { pb: 2 } }}>
                <Stack direction="row" spacing={1.25} alignItems="flex-start" sx={{ mb: 1.5 }}>
                  <AirportLogo logo={airport.logo} name={airport.name} size="medium" />
                  <Box sx={{ flex: 1, minWidth: 0 }}>
                    <Typography
                      variant="subtitle1"
                      fontWeight={600}
                      sx={{
                        color: primaryText,
                        lineHeight: 1.3,
                        wordBreak: 'break-word'
                      }}
                    >
                      {airport.name}
                    </Typography>
                    <Stack direction="row" spacing={0.75} sx={{ mt: 1, flexWrap: 'wrap', gap: 0.75 }}>
                      <Chip label={airport.enabled ? '启用' : '禁用'} variant="filled" size="small" sx={getStatusChipSx(airport.enabled)} />
                      {airport.group && (
                        <Chip
                          label={airport.group}
                          variant="outlined"
                          size="small"
                          sx={{
                            maxWidth: '100%',
                            bgcolor: mutedPanelSurface,
                            borderColor: panelBorder,
                            '& .MuiChip-label': {
                              overflow: 'hidden',
                              textOverflow: 'ellipsis'
                            }
                          }}
                        />
                      )}
                    </Stack>
                  </Box>
                  <Box
                    sx={{
                      flexShrink: 0,
                      width: 42,
                      height: 42,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      borderRadius: 2.5,
                      border: `1px solid ${isSelected ? withAlpha(palette.primary.main, isDark ? 0.38 : 0.24) : panelBorder}`,
                      backgroundColor: isSelected ? withAlpha(palette.primary.main, isDark ? 0.16 : 0.08) : mutedPanelSurface
                    }}
                  >
                    <Checkbox checked={isSelected} onChange={() => onToggleSelect(airport.id)} size="small" />
                  </Box>
                </Stack>

                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mb: 2 }}>
                  <Chip label={`${airport.nodeCount || 0} 节点`} color="primary" variant="outlined" size="small" />
                  <Chip
                    icon={<AccessTimeIcon sx={{ fontSize: '14px !important' }} />}
                    label={airport.cronExpr}
                    variant="outlined"
                    size="small"
                    sx={{
                      maxWidth: '100%',
                      '& .MuiChip-label': {
                        overflow: 'hidden',
                        textOverflow: 'ellipsis'
                      }
                    }}
                  />
                </Box>

                <Box
                  sx={{
                    display: 'grid',
                    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
                    gap: 1.25,
                    mb: 2,
                    p: 1.25,
                    borderRadius: 2,
                    bgcolor: nestedPanelSurface,
                    border: `1px solid ${panelBorder}`
                  }}
                >
                  <Box sx={{ minWidth: 0 }}>
                    <Typography variant="caption" sx={{ color: secondaryText }}>
                      上次运行
                    </Typography>
                    <Typography variant="body2" sx={{ mt: 0.5, color: primaryText, wordBreak: 'break-word' }}>
                      {formatDateTime(airport.lastRunTime)}
                    </Typography>
                  </Box>
                  <Box sx={{ minWidth: 0 }}>
                    <Typography variant="caption" sx={{ color: secondaryText }}>
                      下次运行
                    </Typography>
                    <Typography variant="body2" sx={{ mt: 0.5, color: primaryText, wordBreak: 'break-word' }}>
                      {formatDateTime(airport.nextRunTime)}
                    </Typography>
                  </Box>
                </Box>

                {airport.fetchUsageInfo && (
                  <Box sx={{ mb: 2 }}>
                    <Typography variant="caption" sx={{ display: 'block', mb: 1, fontWeight: 500, color: secondaryText }}>
                      用量信息
                    </Typography>
                    {airport.usageTotal === -1 ? (
                      <Typography variant="body2" sx={{ color: 'error.main', fontWeight: 500 }}>
                        用量获取失败（机场可能不支持）
                      </Typography>
                    ) : airport.usageTotal > 0 ? (
                      <Box
                        sx={{
                          p: 1.5,
                          borderRadius: 2,
                          bgcolor: nestedPanelSurface,
                          border: `1px solid ${panelBorder}`
                        }}
                      >
                        {(() => {
                          const upload = airport.usageUpload || 0;
                          const download = airport.usageDownload || 0;
                          const used = upload + download;
                          const total = airport.usageTotal;
                          const percent = Math.min((used / total) * 100, 100);
                          const color = getUsageColor(percent);

                          return (
                            <>
                              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1, gap: 1 }}>
                                <Typography variant="body2" sx={{ fontWeight: 600, color: primaryText, wordBreak: 'break-word' }}>
                                  {formatBytes(used)} / {formatBytes(total)}
                                </Typography>
                                <Typography variant="body1" sx={{ fontWeight: 700, color }}>
                                  {percent.toFixed(1)}%
                                </Typography>
                              </Box>

                              <Box
                                sx={{
                                  height: 8,
                                  borderRadius: 4,
                                  backgroundColor: getProgressTrackColor(percent),
                                  overflow: 'hidden',
                                  mb: 1.5
                                }}
                              >
                                <Box
                                  sx={{
                                    width: `${percent}%`,
                                    height: '100%',
                                    borderRadius: 4,
                                    backgroundColor: color,
                                    transition: 'width 0.3s ease'
                                  }}
                                />
                              </Box>

                              <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2, mb: airport.usageExpire > 0 ? 1 : 0 }}>
                                <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                  <Typography variant="body2" sx={{ color: 'success.main', fontWeight: 500 }}>
                                    ↑
                                  </Typography>
                                  <Typography variant="body2" sx={{ color: secondaryText }}>
                                    {formatBytes(upload)}
                                  </Typography>
                                </Box>
                                <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                  <Typography variant="body2" sx={{ color: 'info.main', fontWeight: 500 }}>
                                    ↓
                                  </Typography>
                                  <Typography variant="body2" sx={{ color: secondaryText }}>
                                    {formatBytes(download)}
                                  </Typography>
                                </Box>
                              </Box>

                              {airport.usageExpire > 0 &&
                                (() => {
                                  const now = Math.floor(Date.now() / 1000);
                                  const daysLeft = (airport.usageExpire - now) / (24 * 60 * 60);
                                  const isUrgent = daysLeft <= 7;
                                  const isWarning = daysLeft <= 30 && daysLeft > 7;

                                  return (
                                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, flexWrap: 'wrap' }}>
                                      {isUrgent && <WarningAmberIcon sx={{ fontSize: 14, color: 'error.main' }} />}
                                      {isWarning && <EventIcon sx={{ fontSize: 14, color: 'info.main' }} />}
                                      <Typography
                                        variant="body2"
                                        sx={{
                                          color: isUrgent ? 'error.main' : isWarning ? 'info.main' : 'text.secondary',
                                          fontWeight: isUrgent || isWarning ? 600 : 400
                                        }}
                                      >
                                        到期: {formatExpireTime(airport.usageExpire)}
                                        {isUrgent && ` (${Math.max(0, Math.ceil(daysLeft))}天)`}
                                      </Typography>
                                    </Box>
                                  );
                                })()}
                            </>
                          );
                        })()}
                      </Box>
                    ) : (
                      <Typography variant="body2" sx={{ color: secondaryText }}>
                        待获取
                      </Typography>
                    )}
                  </Box>
                )}

                <Box sx={{ mb: 2 }}>
                  <Typography variant="caption" sx={{ display: 'block', mb: 1, fontWeight: 500, color: secondaryText }}>
                    节点测试
                  </Typography>
                  <AirportNodeStatsCard nodeStats={airport.nodeStats} nodeCount={airport.nodeCount || 0} />
                </Box>

                <Stack
                  direction="row"
                  spacing={1}
                  alignItems="center"
                  justifyContent="space-between"
                  sx={{
                    pt: 1,
                    borderTop: `1px solid ${panelBorder}`,
                    mt: 0.5,
                    p: 1,
                    borderRadius: 2,
                    bgcolor: mutedPanelSurface,
                    border: `1px solid ${panelBorder}`
                  }}
                >
                  <Button
                    variant="contained"
                    color="primary"
                    startIcon={<DownloadIcon />}
                    onClick={() => onPull(airport)}
                    sx={{
                      flex: 1,
                      minWidth: 0,
                      minHeight: 40,
                      borderRadius: 2,
                      boxShadow: 'none',
                      whiteSpace: 'nowrap'
                    }}
                  >
                    立即拉取
                  </Button>
                  <IconButton
                    aria-label="刷新用量"
                    onClick={() => onRefreshUsage(airport)}
                    disabled={!airport.fetchUsageInfo}
                    sx={{
                      flexShrink: 0,
                      width: 40,
                      height: 40,
                      borderRadius: 2,
                      bgcolor: airport.fetchUsageInfo
                        ? withAlpha(palette.success.main, 0.1)
                        : withAlpha(palette.action.disabledBackground, 0.7),
                      color: airport.fetchUsageInfo ? palette.success.main : palette.action.disabled,
                      '&:hover': airport.fetchUsageInfo ? { bgcolor: withAlpha(palette.success.main, 0.2) } : undefined
                    }}
                  >
                    <AccountBalanceWalletIcon />
                  </IconButton>
                  <IconButton
                    aria-label="复制订阅地址"
                    onClick={() => handleCopyUrl(airport)}
                    sx={{
                      flexShrink: 0,
                      width: 40,
                      height: 40,
                      borderRadius: 2,
                      bgcolor: withAlpha(palette.secondary.main, 0.1),
                      color: palette.secondary.main,
                      '&:hover': { bgcolor: withAlpha(palette.secondary.main, 0.2) }
                    }}
                  >
                    <ContentCopyIcon />
                  </IconButton>
                  <IconButton
                    aria-label="更多操作"
                    onClick={(event) => handleOpenMenu(event, airport.id)}
                    sx={{
                      width: 40,
                      height: 40,
                      borderRadius: 2,
                      border: `1px solid ${panelBorder}`,
                      bgcolor: mutedPanelSurface,
                      color: secondaryText,
                      '&:hover': {
                        bgcolor: withAlpha(palette.primary.main, isDark ? 0.12 : 0.05)
                      }
                    }}
                  >
                    <MoreVertIcon />
                  </IconButton>
                </Stack>
              </CardContent>
            </Card>
          );
        })}
      </Stack>

      <Menu
        anchorEl={menuAnchor}
        open={Boolean(menuAnchor)}
        onClose={handleCloseMenu}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
        transformOrigin={{ vertical: 'top', horizontal: 'right' }}
        PaperProps={{
          sx: {
            minWidth: 190,
            borderRadius: 2,
            border: '1px solid',
            borderColor: panelBorder,
            bgcolor: cardSurface
          }
        }}
      >
        <MenuItem
          onClick={() => {
            const airport = airports.find((item) => item.id === menuAirportId);
            if (airport) handleMenuAction(onOpenNodes, airport);
          }}
        >
          <ListItemIcon>
            <LanIcon fontSize="small" color="primary" />
          </ListItemIcon>
          <ListItemText>查看节点</ListItemText>
        </MenuItem>
        <MenuItem
          onClick={() => {
            const airport = airports.find((item) => item.id === menuAirportId);
            if (airport) handleMenuAction(onQuickCheck, airport);
          }}
        >
          <ListItemIcon>
            <SpeedIcon fontSize="small" color="primary" />
          </ListItemIcon>
          <ListItemText>快速检测</ListItemText>
        </MenuItem>
        <MenuItem
          onClick={() => {
            const airport = airports.find((item) => item.id === menuAirportId);
            if (airport) handleMenuAction(onEdit, airport);
          }}
        >
          <ListItemIcon>
            <EditIcon fontSize="small" color="info" />
          </ListItemIcon>
          <ListItemText>编辑</ListItemText>
        </MenuItem>
        <MenuItem
          onClick={() => {
            const airport = airports.find((item) => item.id === menuAirportId);
            if (airport) handleMenuAction(onDelete, airport);
          }}
          sx={{ color: 'error.main' }}
        >
          <ListItemIcon>
            <DeleteIcon fontSize="small" color="error" />
          </ListItemIcon>
          <ListItemText>删除</ListItemText>
        </MenuItem>
      </Menu>

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

AirportMobileList.propTypes = {
  airports: PropTypes.array.isRequired,
  selectedIds: PropTypes.array,
  onToggleSelect: PropTypes.func.isRequired,
  onEdit: PropTypes.func.isRequired,
  onDelete: PropTypes.func.isRequired,
  onPull: PropTypes.func.isRequired,
  onOpenNodes: PropTypes.func.isRequired,
  onQuickCheck: PropTypes.func.isRequired,
  onRefreshUsage: PropTypes.func
};

AirportMobileList.defaultProps = {
  selectedIds: []
};
