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
import LockOpenIcon from '@mui/icons-material/LockOpen';

// api
import {
  getNodeCheckMeta,
  getNodeCheckProfiles,
  updateNodeCheckProfile,
  deleteNodeCheckProfile,
  runNodeCheckWithProfile
} from 'api/nodeCheck';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

// local components
import NodeCheckProfileFormDialog from './NodeCheckProfileFormDialog';
import { withAlpha } from 'utils/colorUtils';
import { getNodeCheckStrategyChipSx, getNodeCheckStrategyThemeTokens } from '../nodeCheckTheme';

import { buildNodeCheckProfilePayload, formatUnlockProvidersSummary, setUnlockMeta } from '../utils';

/**
 * 节点检测策略管理抽屉
 */
export default function NodeCheckProfilesDrawer({ open, onClose, groupOptions, tagOptions, onMessage }) {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const themeTokens = getNodeCheckStrategyThemeTokens(theme, isDark);
  const {
    palette,
    dialogSurface,
    dialogSurfaceGradient,
    headerSurface,
    emptyStateSurface,
    panelBorder,
    listRowHoverBackground,
    closeButtonHoverSurface,
    successActionHoverSurface,
    errorActionHoverSurface
  } = themeTokens;

  const [profiles, setProfiles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [formOpen, setFormOpen] = useState(false);
  const [editingProfile, setEditingProfile] = useState(null);

  // 加载策略列表
  const loadProfiles = useCallback(async () => {
    setLoading(true);
    try {
      const [profilesRes, metaRes] = await Promise.all([getNodeCheckProfiles(), getNodeCheckMeta()]);
      setProfiles(profilesRes.data || []);
      setUnlockMeta(metaRes.data || {});
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
      await updateNodeCheckProfile(profile.id, buildNodeCheckProfilePayload(profile, { enabled: !profile.enabled }));
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
            backgroundColor: dialogSurface,
            backgroundImage: dialogSurfaceGradient,
            borderLeft: `1px solid ${panelBorder}`,
            boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.05)}` : theme.shadows[8],
            display: 'flex',
            flexDirection: 'column'
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
            borderBottom: `1px solid ${panelBorder}`,
            backgroundColor: headerSurface
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
            <IconButton
              onClick={onClose}
              size="small"
              sx={{
                '&:hover': {
                  backgroundColor: closeButtonHoverSurface
                }
              }}
            >
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
            <Box
              sx={{
                textAlign: 'center',
                py: 6,
                px: 3,
                m: 2,
                borderRadius: 2,
                border: '1px solid',
                borderColor: panelBorder,
                backgroundColor: emptyStateSurface,
                boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none'
              }}
            >
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
                      transition: 'background 0.2s ease',
                      '&:hover': {
                        background: listRowHoverBackground
                      }
                    }}
                  >
                    <ListItemText
                      sx={{ minWidth: 0, pr: 10.5 }}
                      disableTypography
                      primary={
                        <Box
                          sx={{
                            display: 'flex',
                            alignItems: 'center',
                            columnGap: 0.75,
                            rowGap: 0.75,
                            flexWrap: 'wrap',
                            mb: 0.5,
                            minWidth: 0
                          }}
                        >
                          <Typography
                            variant="subtitle1"
                            fontWeight={600}
                            sx={{ minWidth: 0, maxWidth: '100%', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
                          >
                            {profile.name}
                          </Typography>
                          <Chip
                            label={profile.mode === 'mihomo' ? '延迟+速度' : '仅延迟'}
                            size="small"
                            sx={getNodeCheckStrategyChipSx(themeTokens, profile.mode === 'mihomo' ? 'success' : 'info')}
                          />
                          {profile.detectCountry && (
                            <Chip label="国家" size="small" sx={getNodeCheckStrategyChipSx(themeTokens, 'neutral')} />
                          )}
                          {profile.detectQuality && (
                            <Chip label="质量" size="small" sx={getNodeCheckStrategyChipSx(themeTokens, 'warning')} />
                          )}
                          {profile.detectUnlock && (
                            <Chip
                              icon={<LockOpenIcon sx={{ fontSize: '12px !important' }} />}
                              label={`解锁${profile.unlockProviders?.length ? ` · ${formatUnlockProvidersSummary(profile.unlockProviders, 1)}` : ''}`}
                              size="small"
                              sx={getNodeCheckStrategyChipSx(themeTokens, 'info')}
                            />
                          )}
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
                                sx={{
                                  ...getNodeCheckStrategyChipSx(themeTokens, 'neutral'),
                                  height: 20,
                                  fontSize: '0.65rem'
                                }}
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
                          {profile.detectUnlock && (
                            <Typography variant="caption" color="text.secondary">
                              解锁 Provider: {formatUnlockProvidersSummary(profile.unlockProviders, 2)}
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
                              '&:hover': { backgroundColor: successActionHoverSurface }
                            }}
                          >
                            <PlayArrowIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="编辑">
                          <IconButton
                            size="small"
                            onClick={() => handleEdit(profile)}
                            sx={{
                              color: 'text.secondary',
                              '&:hover': { backgroundColor: closeButtonHoverSurface }
                            }}
                          >
                            <EditIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="删除">
                          <IconButton
                            size="small"
                            onClick={() => handleDelete(profile)}
                            sx={{
                              color: 'error.main',
                              '&:hover': { backgroundColor: errorActionHoverSurface }
                            }}
                          >
                            <DeleteIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      </Stack>
                    </ListItemSecondaryAction>
                  </ListItem>
                  {index < profiles.length - 1 && <Divider sx={{ borderColor: panelBorder }} />}
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
