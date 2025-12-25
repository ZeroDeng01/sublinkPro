import PropTypes from 'prop-types';
import { useState, useEffect, useCallback } from 'react';

// material-ui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Divider from '@mui/material/Divider';
import Drawer from '@mui/material/Drawer';
import IconButton from '@mui/material/IconButton';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemSecondaryAction from '@mui/material/ListItemSecondaryAction';
import ListItemText from '@mui/material/ListItemText';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

// icons
import AddIcon from '@mui/icons-material/Add';
import CloseIcon from '@mui/icons-material/Close';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import ScheduleIcon from '@mui/icons-material/Schedule';
import SpeedIcon from '@mui/icons-material/Speed';

// api
import { getNodeCheckProfiles, updateNodeCheckProfile, deleteNodeCheckProfile, runNodeCheckWithProfile } from 'api/nodeCheck';

// local components
import NodeCheckProfileFormDialog from './NodeCheckProfileFormDialog';

/**
 * 节点检测策略管理抽屉
 */
export default function NodeCheckProfilesDrawer({ open, onClose, groupOptions, tagOptions, onMessage }) {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  const [profiles, setProfiles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [formOpen, setFormOpen] = useState(false);
  const [editingProfile, setEditingProfile] = useState(null);

  // 加载策略列表
  const loadProfiles = useCallback(async () => {
    setLoading(true);
    try {
      const response = await getNodeCheckProfiles();
      setProfiles(response.data || []);
    } catch (error) {
      console.error('加载策略列表失败:', error);
      onMessage?.('加载策略列表失败', 'error');
    } finally {
      setLoading(false);
    }
  }, [onMessage]);

  useEffect(() => {
    if (open) {
      loadProfiles();
    }
  }, [open, loadProfiles]);

  // 切换启用状态
  const handleToggleEnabled = async (profile) => {
    try {
      // 将 groups 和 tags 字符串转换为数组格式（后端 API 期望数组类型）
      const groups = profile.groups ? profile.groups.split(',').filter(Boolean) : [];
      const tags = profile.tags ? profile.tags.split(',').filter(Boolean) : [];

      await updateNodeCheckProfile(profile.id, {
        name: profile.name,
        enabled: !profile.enabled,
        cronExpr: profile.cronExpr,
        mode: profile.mode,
        testUrl: profile.testUrl,
        latencyUrl: profile.latencyUrl,
        timeout: profile.timeout,
        groups,
        tags,
        latencyConcurrency: profile.latencyConcurrency,
        speedConcurrency: profile.speedConcurrency,
        detectCountry: profile.detectCountry,
        landingIpUrl: profile.landingIpUrl,
        includeHandshake: profile.includeHandshake,
        speedRecordMode: profile.speedRecordMode,
        peakSampleInterval: profile.peakSampleInterval,
        trafficByGroup: profile.trafficByGroup,
        trafficBySource: profile.trafficBySource,
        trafficByNode: profile.trafficByNode
      });
      loadProfiles();
      onMessage?.(profile.enabled ? '已禁用定时检测' : '已启用定时检测');
    } catch (error) {
      console.error('切换状态失败:', error);
      onMessage?.('操作失败', 'error');
    }
  };

  // 删除策略
  const handleDelete = async (profile) => {
    if (!window.confirm(`确定要删除策略 "${profile.name}" 吗？`)) {
      return;
    }
    try {
      await deleteNodeCheckProfile(profile.id);
      loadProfiles();
      onMessage?.('删除成功');
    } catch (error) {
      console.error('删除失败:', error);
      onMessage?.(error.message || '删除失败', 'error');
    }
  };

  // 执行检测
  const handleRun = async (profile) => {
    try {
      await runNodeCheckWithProfile(profile.id);
      onMessage?.('检测任务已启动');
    } catch (error) {
      console.error('执行检测失败:', error);
      onMessage?.(error.message || '执行失败', 'error');
    }
  };

  // 编辑策略
  const handleEdit = (profile) => {
    setEditingProfile(profile);
    setFormOpen(true);
  };

  // 新增策略
  const handleAdd = () => {
    setEditingProfile(null);
    setFormOpen(true);
  };

  // 表单提交成功
  const handleFormSuccess = () => {
    setFormOpen(false);
    setEditingProfile(null);
    loadProfiles();
    onMessage?.(editingProfile ? '更新成功' : '创建成功');
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

  const formatLastRunTime = (lastRunTime) => {
    if (!lastRunTime) return '从未执行';
    const date = new Date(lastRunTime);
    return date.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  return (
    <>
      <Drawer
        anchor="right"
        open={open}
        onClose={onClose}
        PaperProps={{
          sx: {
            width: isMobile ? '100%' : 420,
            backgroundColor: isDark ? 'rgba(18,18,18,0.98)' : 'background.paper'
          }
        }}
      >
        {/* 标题栏 */}
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            p: 2,
            borderBottom: `1px solid ${isDark ? 'rgba(255,255,255,0.12)' : 'rgba(0,0,0,0.12)'}`
          }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <SpeedIcon color="primary" />
            <Typography variant="h6">检测策略管理</Typography>
          </Box>
          <Stack direction="row" spacing={1}>
            <Button size="small" variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
              新建
            </Button>
            <IconButton onClick={onClose} size="small">
              <CloseIcon />
            </IconButton>
          </Stack>
        </Box>

        {/* 策略列表 */}
        <Box sx={{ flex: 1, overflow: 'auto' }}>
          {loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
              <CircularProgress size={32} />
            </Box>
          ) : profiles.length === 0 ? (
            <Box sx={{ textAlign: 'center', py: 6 }}>
              <SpeedIcon sx={{ fontSize: 48, opacity: 0.3, mb: 2 }} />
              <Typography color="text.secondary" gutterBottom>
                暂无检测策略
              </Typography>
              <Button variant="outlined" startIcon={<AddIcon />} onClick={handleAdd} sx={{ mt: 2 }}>
                创建第一个策略
              </Button>
            </Box>
          ) : (
            <List sx={{ py: 0 }}>
              {profiles.map((profile, index) => (
                <Box key={profile.id}>
                  <ListItem
                    sx={{
                      py: 2,
                      '&:hover': {
                        backgroundColor: isDark ? 'rgba(255,255,255,0.04)' : 'rgba(0,0,0,0.02)'
                      }
                    }}
                  >
                    <ListItemText
                      primary={
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
                          <Typography variant="subtitle1" fontWeight={600}>
                            {profile.name}
                          </Typography>
                          <Chip
                            label={profile.mode === 'mihomo' ? '延迟+速度' : '仅延迟'}
                            size="small"
                            sx={{
                              height: 20,
                              fontSize: '0.7rem',
                              backgroundColor: profile.mode === 'mihomo' ? 'rgba(76, 175, 80, 0.15)' : 'rgba(33, 150, 243, 0.15)',
                              color: profile.mode === 'mihomo' ? 'success.main' : 'primary.main'
                            }}
                          />
                        </Box>
                      }
                      secondary={
                        <Stack spacing={0.5} sx={{ mt: 1 }}>
                          {/* 定时状态 */}
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <Switch size="small" checked={profile.enabled} onChange={() => handleToggleEnabled(profile)} />
                            <Typography variant="caption" color="text.secondary">
                              {profile.enabled ? '定时已启用' : '定时未启用'}
                            </Typography>
                            {profile.enabled && profile.nextRunTime && (
                              <Chip
                                icon={<ScheduleIcon sx={{ fontSize: '14px !important' }} />}
                                label={`下次: ${formatNextRunTime(profile.nextRunTime)}`}
                                size="small"
                                sx={{ height: 20, fontSize: '0.65rem' }}
                              />
                            )}
                          </Box>
                          {/* 上次执行时间 */}
                          <Typography variant="caption" color="text.secondary">
                            上次执行: {formatLastRunTime(profile.lastRunTime)}
                          </Typography>
                          {/* 检测范围 */}
                          {(profile.groups || profile.tags) && (
                            <Typography variant="caption" color="text.secondary">
                              范围: {profile.groups || '全部分组'} {profile.tags ? `| 标签: ${profile.tags}` : ''}
                            </Typography>
                          )}
                        </Stack>
                      }
                    />
                    <ListItemSecondaryAction>
                      <Stack direction="row" spacing={0.5}>
                        <Tooltip title="立即执行">
                          <IconButton
                            size="small"
                            onClick={() => handleRun(profile)}
                            sx={{
                              color: 'success.main',
                              '&:hover': { backgroundColor: 'rgba(76, 175, 80, 0.1)' }
                            }}
                          >
                            <PlayArrowIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="编辑">
                          <IconButton size="small" onClick={() => handleEdit(profile)}>
                            <EditIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="删除">
                          <IconButton
                            size="small"
                            onClick={() => handleDelete(profile)}
                            sx={{
                              color: 'error.main',
                              '&:hover': { backgroundColor: 'rgba(244, 67, 54, 0.1)' }
                            }}
                          >
                            <DeleteIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      </Stack>
                    </ListItemSecondaryAction>
                  </ListItem>
                  {index < profiles.length - 1 && <Divider />}
                </Box>
              ))}
            </List>
          )}
        </Box>
      </Drawer>

      {/* 策略编辑对话框 */}
      <NodeCheckProfileFormDialog
        open={formOpen}
        onClose={() => {
          setFormOpen(false);
          setEditingProfile(null);
        }}
        profile={editingProfile}
        groupOptions={groupOptions}
        tagOptions={tagOptions}
        onSuccess={handleFormSuccess}
      />
    </>
  );
}

NodeCheckProfilesDrawer.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  groupOptions: PropTypes.array,
  tagOptions: PropTypes.array,
  onMessage: PropTypes.func
};
