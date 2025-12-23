import PropTypes from 'prop-types';
import { useState, useEffect, useCallback } from 'react';

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
import Divider from '@mui/material/Divider';
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

/**
 * 节点检测策略选择对话框
 * 用于手动测速时选择检测策略
 */
export default function ProfileSelectDialog({ open, onClose, nodeIds, onSuccess, onOpenSettings }) {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

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
          backgroundColor: isDark ? 'rgba(30,30,30,0.95)' : 'background.paper'
        }
      }}
    >
      <DialogTitle sx={{ pb: 1 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <SpeedIcon color="primary" />
            <span>选择检测策略</span>
          </Box>
          <IconButton size="small" onClick={onClose}>
            <CloseIcon fontSize="small" />
          </IconButton>
        </Box>
        {nodeIds?.length > 0 && (
          <Typography variant="caption" color="text.secondary">
            已选择 {nodeIds.length} 个节点
          </Typography>
        )}
      </DialogTitle>

      <Divider />

      <DialogContent sx={{ p: 0 }}>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress size={32} />
          </Box>
        ) : profiles.length === 0 ? (
          <Box sx={{ textAlign: 'center', py: 4 }}>
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
                  borderBottom:
                    index < profiles.length - 1 ? `1px solid ${isDark ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.08)'}` : 'none',
                  '&.Mui-selected': {
                    backgroundColor: isDark ? 'rgba(33, 150, 243, 0.16)' : 'rgba(33, 150, 243, 0.08)',
                    '&:hover': {
                      backgroundColor: isDark ? 'rgba(33, 150, 243, 0.24)' : 'rgba(33, 150, 243, 0.12)'
                    }
                  }
                }}
              >
                <ListItemIcon sx={{ minWidth: 36 }}>
                  {selectedProfileId === profile.id ? (
                    <CheckCircleIcon color="primary" fontSize="small" />
                  ) : (
                    <SpeedIcon fontSize="small" sx={{ opacity: 0.5 }} />
                  )}
                </ListItemIcon>
                <ListItemText
                  primary={
                    <Box component="span" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <span>{profile.name}</span>
                      <Chip
                        label={profile.mode === 'mihomo' ? '延迟+速度' : '仅延迟'}
                        size="small"
                        sx={{
                          height: 20,
                          fontSize: '0.7rem',
                          backgroundColor:
                            profile.mode === 'mihomo'
                              ? isDark
                                ? 'rgba(76, 175, 80, 0.2)'
                                : 'rgba(76, 175, 80, 0.1)'
                              : isDark
                                ? 'rgba(33, 150, 243, 0.2)'
                                : 'rgba(33, 150, 243, 0.1)',
                          color: profile.mode === 'mihomo' ? 'success.main' : 'primary.main'
                        }}
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

      <Divider />

      <DialogActions sx={{ px: 2, py: 1.5, justifyContent: 'space-between' }}>
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
            background: 'linear-gradient(135deg, #4caf50 0%, #2e7d32 100%)',
            fontStyle: { color: '#ffffff' },
            '&:hover': {
              background: 'linear-gradient(135deg, #66bb6a 0%, #388e3c 100%)'
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
