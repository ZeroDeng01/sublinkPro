import PropTypes from 'prop-types';
import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';

// material-ui
import { useTheme } from '@mui/material/styles';
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
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { withAlpha } from 'utils/colorUtils';
import { formatDateTime } from 'i18n/locales';
import { getNodeCheckStrategyChipSx, getNodeCheckStrategyThemeTokens } from '../nodeCheckTheme';

/**
 * 节点检测策略选择对话框
 * 用于手动测速时选择检测策略
 */

export default function ProfileSelectDialog({ open, onClose, nodeIds, onSuccess, onOpenSettings }) {
  const theme = useTheme();
  const { i18n, t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const themeTokens = getNodeCheckStrategyThemeTokens(theme, isDark);
  const {
    palette,
    dialogSurface,
    dialogSurfaceGradient,
    headerSurface,
    actionSurface,
    emptyStateSurface,
    panelBorder,
    listRowHoverBackground,
    listRowSelectedBackground,
    listRowSelectedHoverBackground,
    listRowSelectedShadow,
    closeButtonHoverSurface,
    secondaryIconOpacity
  } = themeTokens;

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
      onSuccess?.(t('nodes.profileSelect.messages.started'));
      onClose();
    } catch (error) {
      console.error('执行检测失败:', error);
      onSuccess?.(error.message || t('nodes.profileSelect.messages.executeFailed'), 'error');
    } finally {
      setExecuting(false);
    }
  };

  const formatNextRunTime = (nextRunTime) => {
    if (!nextRunTime) return null;
    return formatDateTime(nextRunTime, i18n.resolvedLanguage || i18n.language, {
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
          borderColor: panelBorder,
          boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.05)}` : theme.shadows[8]
        }
      }}
    >
      <DialogTitle
        sx={{
          pb: 1,
          borderBottom: `1px solid ${panelBorder}`,
          backgroundColor: headerSurface
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <SpeedIcon color="primary" />
            <span>{t('nodes.profileSelect.title')}</span>
          </Box>
          <IconButton
            size="small"
            onClick={onClose}
            sx={{
              '&:hover': {
                backgroundColor: closeButtonHoverSurface
              }
            }}
          >
            <CloseIcon fontSize="small" />
          </IconButton>
        </Box>
        {nodeIds?.length > 0 && (
          <Typography variant="caption" color="text.secondary">
            {t('nodes.profileSelect.selectedCount', { count: nodeIds.length })}
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
              borderColor: panelBorder,
              backgroundColor: emptyStateSurface,
              boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none'
            }}
          >
            <Typography color="text.secondary" gutterBottom>
              {t('nodes.profileSelect.empty')}
            </Typography>
            <Button
              size="small"
              startIcon={<SettingsIcon />}
              onClick={() => {
                onClose();
                onOpenSettings?.();
              }}
            >
              {t('nodes.profileSelect.create')}
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
                  borderBottom: index < profiles.length - 1 ? `1px solid ${panelBorder}` : 'none',
                  transition: 'background 0.2s ease, box-shadow 0.2s ease',
                  '&:hover': {
                    background: listRowHoverBackground
                  },
                  '&.Mui-selected': {
                    background: listRowSelectedBackground,
                    boxShadow: listRowSelectedShadow,
                    '&:hover': {
                      background: listRowSelectedHoverBackground
                    }
                  }
                }}
              >
                <ListItemIcon sx={{ minWidth: 36 }}>
                  {selectedProfileId === profile.id ? (
                    <CheckCircleIcon color="primary" fontSize="small" />
                  ) : (
                    <SpeedIcon fontSize="small" sx={{ color: 'text.secondary', opacity: secondaryIconOpacity }} />
                  )}
                </ListItemIcon>
                <ListItemText
                  primary={
                    <Box component="span" sx={{ display: 'flex', alignItems: 'center', columnGap: 0.75, rowGap: 0.75, flexWrap: 'wrap' }}>
                      <span>{profile.name}</span>
                      <Chip
                        label={profile.mode === 'mihomo' ? t('nodes.profileSelect.mode.full') : t('nodes.profileSelect.mode.delayOnly')}
                        size="small"
                        sx={getNodeCheckStrategyChipSx(themeTokens, profile.mode === 'mihomo' ? 'success' : 'info')}
                      />
                    </Box>
                  }
                  secondary={
                    <Box component="span" sx={{ display: 'flex', flexDirection: 'row', gap: 1, mt: 0.5, alignItems: 'center' }}>
                      {profile.enabled && (
                        <>
                          <ScheduleIcon sx={{ fontSize: 14, opacity: 0.6 }} />
                          <Typography component="span" variant="caption" color="text.secondary">
                            {formatNextRunTime(profile.nextRunTime) || t('nodes.profileSelect.scheduled')}
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
          borderTop: `1px solid ${panelBorder}`,
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
          {t('nodes.profileSelect.manage')}
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
          {t('nodes.profileSelect.start')}
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
