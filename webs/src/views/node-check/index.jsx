import { useState, useEffect, useCallback } from 'react';

// material-ui
import { useTheme } from '@mui/material/styles';

import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Divider from '@mui/material/Divider';
import IconButton from '@mui/material/IconButton';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

// icons
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import HistoryIcon from '@mui/icons-material/History';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import RefreshIcon from '@mui/icons-material/Refresh';
import ScheduleIcon from '@mui/icons-material/Schedule';
import SpeedIcon from '@mui/icons-material/Speed';
import TimerIcon from '@mui/icons-material/Timer';
import FilterListIcon from '@mui/icons-material/FilterList';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import TaskProgressPanel from 'components/TaskProgressPanel';

// api
import { getNodeCheckProfiles, updateNodeCheckProfile, deleteNodeCheckProfile, runNodeCheckWithProfile } from 'api/nodeCheck';
import { getNodeGroups } from 'api/nodes';
import { getTags } from 'api/tags';

// local components
import NodeCheckProfileFormDialog from 'views/nodes/component/NodeCheckProfileFormDialog';

// ==============================|| 节点检测策略管理 ||============================== //

export default function NodeCheckList() {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  const [profiles, setProfiles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [formOpen, setFormOpen] = useState(false);
  const [editingProfile, setEditingProfile] = useState(null);
  const [groupOptions, setGroupOptions] = useState([]);
  const [tagOptions, setTagOptions] = useState([]);

  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  // 加载策略列表
  const loadProfiles = useCallback(async () => {
    setLoading(true);
    try {
      const response = await getNodeCheckProfiles();
      setProfiles(response.data || []);
    } catch (error) {
      console.error('加载策略列表失败:', error);
      showMessage('加载策略列表失败', 'error');
    } finally {
      setLoading(false);
    }
  }, []);

  // 加载分组和标签选项
  const loadOptions = useCallback(async () => {
    try {
      const [groupRes, tagRes] = await Promise.all([getNodeGroups(), getTags()]);
      setGroupOptions((groupRes.data || []).sort());
      setTagOptions(tagRes.data || []);
    } catch (error) {
      console.error('加载选项失败:', error);
    }
  }, []);

  useEffect(() => {
    loadProfiles();
    loadOptions();
  }, [loadProfiles, loadOptions]);

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
      showMessage(profile.enabled ? '已禁用定时检测' : '已启用定时检测');
    } catch (error) {
      console.error('切换状态失败:', error);
      showMessage('操作失败', 'error');
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
      showMessage('删除成功');
    } catch (error) {
      console.error('删除失败:', error);
      showMessage(error.message || '删除失败', 'error');
    }
  };

  // 执行检测
  const handleRun = async (profile) => {
    try {
      await runNodeCheckWithProfile(profile.id);
      showMessage('检测任务已启动');
    } catch (error) {
      console.error('执行检测失败:', error);
      showMessage(error.message || '执行失败', 'error');
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
    showMessage(editingProfile ? '更新成功' : '创建成功');
  };

  const formatTime = (timeStr) => {
    if (!timeStr) return '-';
    const date = new Date(timeStr);
    // 检查是否是有效日期
    if (isNaN(date.getTime())) return '-';
    return date.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  return (
    <MainCard
      title={
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <SpeedIcon color="primary" />
          <span>节点检测策略</span>
        </Box>
      }
      secondary={
        <Stack direction="row" spacing={1}>
          <Tooltip title="刷新">
            <IconButton onClick={loadProfiles} disabled={loading}>
              <RefreshIcon />
            </IconButton>
          </Tooltip>
          <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
            新建策略
          </Button>
        </Stack>
      }
    >
      {/* 任务进度面板 */}
      <TaskProgressPanel />

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 6 }}>
          <CircularProgress />
        </Box>
      ) : profiles.length === 0 ? (
        <Box sx={{ textAlign: 'center', py: 8 }}>
          <SpeedIcon sx={{ fontSize: 64, opacity: 0.2, mb: 2 }} />
          <Typography variant="h6" color="text.secondary" gutterBottom>
            暂无检测策略
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
            创建检测策略来自动或手动检测节点延迟和速度
          </Typography>
          <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
            创建第一个策略
          </Button>
        </Box>
      ) : (
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: {
              xs: '1fr',
              sm: 'repeat(2, minmax(0, 1fr))',
              md: 'repeat(2, minmax(0, 1fr))',
              lg: 'repeat(3, minmax(0, 1fr))',
              xl: 'repeat(4, minmax(0, 1fr))'
            },
            gap: 2,
            mt: 2
          }}
        >
          {profiles.map((profile) => (
            <Card
              key={profile.id}
              variant="outlined"
              sx={{
                height: 200,
                display: 'flex',
                flexDirection: 'column',
                overflow: 'hidden',
                borderColor: profile.enabled
                  ? isDark
                    ? 'rgba(76, 175, 80, 0.5)'
                    : 'rgba(76, 175, 80, 0.3)'
                  : isDark
                    ? 'rgba(255,255,255,0.12)'
                    : 'rgba(0,0,0,0.12)',
                backgroundColor: isDark ? 'rgba(255,255,255,0.02)' : 'rgba(0,0,0,0.01)',
                transition: 'all 0.2s',
                '&:hover': {
                  borderColor: 'primary.main',
                  transform: 'translateY(-2px)',
                  boxShadow: isDark ? '0 4px 20px rgba(0,0,0,0.3)' : '0 4px 20px rgba(0,0,0,0.1)'
                }
              }}
            >
              <CardContent
                sx={{
                  flex: 1,
                  display: 'flex',
                  flexDirection: 'column',
                  py: 1.5,
                  px: 2,
                  overflow: 'hidden',
                  '&:last-child': { pb: 1.5 }
                }}
              >
                {/* 标题行 */}
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    mb: 1,
                    flexShrink: 0,
                    minWidth: 0
                  }}
                >
                  <Box
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 1,
                      minWidth: 0,
                      flex: 1,
                      overflow: 'hidden'
                    }}
                  >
                    <Typography
                      variant="subtitle1"
                      fontWeight={600}
                      sx={{
                        minWidth: 0,
                        flexShrink: 1,
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap'
                      }}
                    >
                      {profile.name}
                    </Typography>
                    <Chip
                      label={profile.mode === 'mihomo' ? '延迟+速度' : '仅延迟'}
                      size="small"
                      sx={{
                        height: 20,
                        fontSize: '0.65rem',
                        flexShrink: 0,
                        backgroundColor: profile.mode === 'mihomo' ? 'rgba(76, 175, 80, 0.15)' : 'rgba(33, 150, 243, 0.15)',
                        color: profile.mode === 'mihomo' ? 'primary.main' : 'error.main'
                      }}
                    />
                  </Box>
                  <Switch
                    size="small"
                    checked={profile.enabled}
                    onChange={() => handleToggleEnabled(profile)}
                    sx={{ ml: 1, flexShrink: 0 }}
                  />
                </Box>

                <Divider sx={{ mb: 1, flexShrink: 0 }} />

                {/* 信息区 */}
                <Box
                  sx={{
                    flex: 1,
                    minHeight: 0,
                    overflow: 'hidden',
                    display: 'flex',
                    flexDirection: 'column',
                    gap: 0.5
                  }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, minWidth: 0 }}>
                    <TimerIcon sx={{ fontSize: 14, opacity: 0.6, flexShrink: 0 }} />
                    <Typography variant="caption" color="text.secondary" noWrap>
                      超时: {profile.timeout}s
                    </Typography>
                  </Box>

                  {profile.enabled && profile.cronExpr && (
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, minWidth: 0 }}>
                      <ScheduleIcon sx={{ fontSize: 14, opacity: 0.6, flexShrink: 0 }} />
                      <Typography variant="caption" color="text.secondary" noWrap>
                        下次: {formatTime(profile.nextRunTime)}
                      </Typography>
                    </Box>
                  )}

                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, minWidth: 0 }}>
                    <HistoryIcon sx={{ fontSize: 14, opacity: 0.6, flexShrink: 0 }} />
                    <Typography variant="caption" color="text.secondary" noWrap>
                      上次: {formatTime(profile.lastRunTime)}
                    </Typography>
                  </Box>

                  {(profile.groups || profile.tags) && (
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, minWidth: 0, overflow: 'hidden' }}>
                      <FilterListIcon sx={{ fontSize: 14, opacity: 0.6, flexShrink: 0 }} />
                      <Typography
                        variant="caption"
                        color="text.secondary"
                        sx={{
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                          minWidth: 0
                        }}
                      >
                        范围: {profile.groups || '全部'}
                        {profile.tags ? ` | ${profile.tags}` : ''}
                      </Typography>
                    </Box>
                  )}
                </Box>

                {/* 操作按钮 */}
                <Stack direction="row" spacing={0.5} justifyContent="flex-end" sx={{ mt: 1, flexShrink: 0 }}>
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
              </CardContent>
            </Card>
          ))}
        </Box>
      )}

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

      {/* Snackbar */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert onClose={() => setSnackbar({ ...snackbar, open: false })} severity={snackbar.severity} variant="standard">
          {snackbar.message}
        </Alert>
      </Snackbar>
    </MainCard>
  );
}
