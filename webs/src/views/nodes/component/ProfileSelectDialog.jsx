import PropTypes from 'prop-types';
import { useState, useEffect, useCallback } from 'react';

// material-ui
import { useColorScheme, useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import IconButton from '@mui/material/IconButton';
import List from '@mui/material/List';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Typography from '@mui/material/Typography';

// icons
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import CloseIcon from '@mui/icons-material/Close';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import ScheduleIcon from '@mui/icons-material/Schedule';
import SettingsIcon from '@mui/icons-material/Settings';
import SpeedIcon from '@mui/icons-material/Speed';
import TimerIcon from '@mui/icons-material/Timer';

// api
import { getNodeCheckProfiles, runNodeCheck } from 'api/nodeCheck';
import { withAlpha } from 'utils/colorUtils';

/**
 * 节点检测策略选择对话框
 * 用于手动测速时选择检测策略
 */

const getStrategyTitleChipSx = (theme, isDark, tone = 'neutral') => {
  const palette = theme.vars?.palette || theme.palette;
  const toneMap = {
    success: {
      accent: palette.success.main,
      text: isDark ? palette.success.light : palette.success.dark
    },
    info: {
      accent: palette.info.main,
      text: isDark ? palette.info.light : palette.info.dark
    },
    warning: {
      accent: palette.warning.main,
      text: isDark ? palette.warning.light : palette.warning.dark
    },
    neutral: {
      accent: palette.grey[500],
      text: isDark ? withAlpha(palette.common.white, 0.92) : withAlpha(palette.text.primary, 0.82)
    }
  };

  const { accent, text } = toneMap[tone] || toneMap.neutral;

  return {
    height: 22,
    borderRadius: 1.5,
    flexShrink: 0,
    fontSize: '0.72rem',
    fontWeight: 700,
    letterSpacing: '0.01em',
    color: text,
    backgroundColor: withAlpha(accent, isDark ? (tone === 'neutral' ? 0.24 : 0.18) : tone === 'neutral' ? 0.08 : 0.1),
    border: `1px solid ${withAlpha(accent, isDark ? (tone === 'neutral' ? 0.34 : 0.32) : tone === 'neutral' ? 0.16 : 0.2)}`,
    boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.05)}` : 'none',
    '& .MuiChip-label': {
      px: 1,
      py: 0
    },
    '& .MuiChip-icon': {
      color: 'inherit',
      fontSize: '0.82rem',
      ml: 0.75,
      mr: -0.25
    }
  };
};

export default function ProfileSelectDialog({ open, onClose, nodeIds, onSuccess, onOpenSettings }) {
  const theme = useTheme();
  const { mode } = useColorScheme();
  const runtimeColorScheme = typeof document !== 'undefined' ? document.documentElement?.getAttribute('data-color-scheme') : null;
  const isDark = (mode || runtimeColorScheme || theme.palette.mode) === 'dark';
  const palette = theme.vars?.palette || theme.palette;
  const dialogSurface = isDark ? withAlpha(palette.background.default, 0.96) : palette.background.paper;
  const dialogSurfaceGradient = isDark
    ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.16)} 0%, ${dialogSurface} 100%)`
    : 'none';
  const headerSurface = isDark ? withAlpha(palette.background.paper, 0.2) : palette.background.paper;
  const actionSurface = isDark ? withAlpha(palette.background.default, 0.9) : withAlpha(palette.background.default, 0.78);
  const emptyStateSurface = isDark ? withAlpha(palette.background.paper, 0.34) : withAlpha(palette.background.default, 0.72);
  const strategyItemDivider = isDark ? withAlpha(palette.divider, 0.82) : withAlpha(palette.divider, 0.9);
  const strategyItemHoverBackground = isDark
    ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.28)} 0%, ${withAlpha(palette.primary.main, 0.06)} 100%)`
    : withAlpha(palette.primary.main, 0.04);
  const strategyItemSelectedBackground = isDark
    ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.42)} 0%, ${withAlpha(palette.primary.main, 0.12)} 100%)`
    : withAlpha(palette.primary.main, 0.08);
  const strategyItemSelectedHoverBackground = isDark
    ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.48)} 0%, ${withAlpha(palette.primary.main, 0.16)} 100%)`
    : withAlpha(palette.primary.main, 0.11);
  const strategyItemSelectedShadow = isDark
    ? `inset 0 0 0 1px ${withAlpha(palette.primary.main, 0.24)}, inset 0 1px 0 ${withAlpha(palette.common.white, 0.05)}`
    : `inset 0 0 0 1px ${withAlpha(palette.primary.main, 0.14)}`;

  const [profiles, setProfiles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [selectedProfileId, setSelectedProfileId] = useState(null);
  const [executing, setExecuting] = useState(false);

  // 加载策略列表
  const loadProfiles = useCallback(async () => {
    setLoading(true);
    try {
      const response = await getNodeCheckProfiles();
      const data = response.data || [];
      setProfiles(data);
      // 默认选中第一个
      if (data.length > 0) {
        setSelectedProfileId((prev) => prev || data[0].id);
      }
    } catch (error) {
      console.error('加载策略列表失败:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (open) {
      loadProfiles();
    }
  }, [open, loadProfiles]);

  const handleSelect = (profileId) => {
    setSelectedProfileId(profileId);
  };

  const handleExecute = async () => {
    setExecuting(true);
    try {
      await runNodeCheck(selectedProfileId, nodeIds);
      onSuccess?.('检测任务已在后台启动');
      onClose();
    } catch (error) {
      console.error('执行检测失败:', error);
      onSuccess?.(error.message || '执行检测失败', 'error');
    } finally {
      setExecuting(false);
    }
  };

  const formatNextRunTime = (nextRunTime) => {
    if (!nextRunTime) return null;
    const date = new Date(nextRunTime);
    return date.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="xs"
      fullWidth
      PaperProps={{
        sx: {
          borderRadius: 3,
          backgroundColor: dialogSurface,
          backgroundImage: dialogSurfaceGradient,
          border: '1px solid',
          borderColor: strategyItemDivider,
          boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.05)}` : theme.shadows[8]
        }
      }}
    >
      <DialogTitle
        sx={{
          pb: 1,
          borderBottom: `1px solid ${strategyItemDivider}`,
          backgroundColor: headerSurface
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <SpeedIcon color="primary" />
            <span>选择检测策略</span>
          </Box>
          <IconButton
            size="small"
            onClick={onClose}
            sx={{
              '&:hover': {
                backgroundColor: withAlpha(palette.primary.main, isDark ? 0.14 : 0.08)
              }
            }}
          >
            <CloseIcon fontSize="small" />
          </IconButton>
        </Box>
        {nodeIds?.length > 0 && (
          <Typography variant="caption" color="text.secondary">
            已选择 {nodeIds.length} 个节点
          </Typography>
        )}
      </DialogTitle>

      <DialogContent sx={{ p: 0, backgroundColor: dialogSurface }}>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress size={32} />
          </Box>
        ) : profiles.length === 0 ? (
          <Box
            sx={{
              textAlign: 'center',
              py: 4,
              px: 2,
              m: 2,
              borderRadius: 2,
              border: '1px solid',
              borderColor: strategyItemDivider,
              backgroundColor: emptyStateSurface,
              boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none'
            }}
          >
            <Typography color="text.secondary" gutterBottom>
              暂无检测策略
            </Typography>
            <Button
              size="small"
              startIcon={<SettingsIcon />}
              onClick={() => {
                onClose();
                onOpenSettings?.();
              }}
            >
              创建策略
            </Button>
          </Box>
        ) : (
          <List sx={{ py: 0 }}>
            {profiles.map((profile, index) => (
              <ListItemButton
                key={profile.id}
                selected={selectedProfileId === profile.id}
                onClick={() => handleSelect(profile.id)}
                sx={{
                  borderBottom: index < profiles.length - 1 ? `1px solid ${strategyItemDivider}` : 'none',
                  transition: 'background 0.2s ease, box-shadow 0.2s ease',
                  '&:hover': {
                    background: strategyItemHoverBackground
                  },
                  '&.Mui-selected': {
                    background: strategyItemSelectedBackground,
                    boxShadow: strategyItemSelectedShadow,
                    '&:hover': {
                      background: strategyItemSelectedHoverBackground
                    }
                  }
                }}
              >
                <ListItemIcon sx={{ minWidth: 36 }}>
                  {selectedProfileId === profile.id ? (
                    <CheckCircleIcon color="primary" fontSize="small" />
                  ) : (
                    <SpeedIcon fontSize="small" sx={{ color: 'text.secondary', opacity: isDark ? 0.82 : 0.64 }} />
                  )}
                </ListItemIcon>
                <ListItemText
                  primary={
                    <Box component="span" sx={{ display: 'flex', alignItems: 'center', columnGap: 0.75, rowGap: 0.75, flexWrap: 'wrap' }}>
                      <span>{profile.name}</span>
                      <Chip
                        label={profile.mode === 'mihomo' ? '延迟+速度' : '仅延迟'}
                        size="small"
                        sx={getStrategyTitleChipSx(theme, isDark, profile.mode === 'mihomo' ? 'success' : 'info')}
                      />
                    </Box>
                  }
                  secondary={
                    <Box component="span" sx={{ display: 'flex', flexDirection: 'row', gap: 1, mt: 0.5, alignItems: 'center' }}>
                      {profile.enabled && (
                        <>
                          <ScheduleIcon sx={{ fontSize: 14, opacity: 0.6 }} />
                          <Typography component="span" variant="caption" color="text.secondary">
                            {formatNextRunTime(profile.nextRunTime) || '定时启用'}
                          </Typography>
                        </>
                      )}
                      {profile.timeout && (
                        <>
                          <TimerIcon sx={{ fontSize: 14, opacity: 0.6 }} />
                          <Typography component="span" variant="caption" color="text.secondary">
                            {profile.timeout}s
                          </Typography>
                        </>
                      )}
                    </Box>
                  }
                  secondaryTypographyProps={{ component: 'div' }}
                />
              </ListItemButton>
            ))}
          </List>
        )}
      </DialogContent>

      <DialogActions
        sx={{
          px: 2,
          py: 1.5,
          justifyContent: 'space-between',
          borderTop: `1px solid ${strategyItemDivider}`,
          backgroundColor: actionSurface
        }}
      >
        <Button
          size="small"
          startIcon={<SettingsIcon />}
          onClick={() => {
            onClose();
            onOpenSettings?.();
          }}
        >
          管理策略
        </Button>
        <Button
          variant="contained"
          color="success"
          startIcon={executing ? <CircularProgress size={16} color="inherit" /> : <PlayArrowIcon />}
          onClick={handleExecute}
          disabled={!selectedProfileId || executing || profiles.length === 0}
          sx={{
            color: palette.common.white,
            background: `linear-gradient(135deg, ${palette.success.main} 0%, ${palette.success.dark} 100%)`,
            boxShadow: `0 8px 18px ${withAlpha(palette.success.main, isDark ? 0.28 : 0.22)}`,
            '&:hover': {
              background: `linear-gradient(135deg, ${palette.success.light} 0%, ${palette.success.main} 100%)`,
              boxShadow: `0 10px 22px ${withAlpha(palette.success.main, isDark ? 0.34 : 0.26)}`
            },
            '&.Mui-disabled': {
              color: 'text.disabled',
              background: theme.palette.action.disabledBackground,
              boxShadow: 'none'
            }
          }}
        >
          开始检测
        </Button>
      </DialogActions>
    </Dialog>
  );
}

ProfileSelectDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  nodeIds: PropTypes.array, // 可选，指定节点ID列表
  onSuccess: PropTypes.func, // 成功回调 (message, severity)
  onOpenSettings: PropTypes.func // 打开策略管理
};
