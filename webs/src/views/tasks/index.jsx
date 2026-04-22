import { useState, useEffect, useCallback, useMemo } from 'react';
import { useTheme, alpha } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TablePagination from '@mui/material/TablePagination';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import LinearProgress from '@mui/material/LinearProgress';
import Alert from '@mui/material/Alert';
import Grid from '@mui/material/Grid';
import Tooltip from '@mui/material/Tooltip';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import FormControl from '@mui/material/FormControl';
import RadioGroup from '@mui/material/RadioGroup';
import FormControlLabel from '@mui/material/FormControlLabel';
import Radio from '@mui/material/Radio';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import InputLabel from '@mui/material/InputLabel';
import RefreshIcon from '@mui/icons-material/Refresh';
import SpeedIcon from '@mui/icons-material/Speed';
import CloudSyncIcon from '@mui/icons-material/CloudSync';
import LocalOfferIcon from '@mui/icons-material/LocalOffer';
import StorageIcon from '@mui/icons-material/Storage';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';
import ScheduleIcon from '@mui/icons-material/Schedule';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import PersonIcon from '@mui/icons-material/Person';
import AutoModeIcon from '@mui/icons-material/AutoMode';
import StopIcon from '@mui/icons-material/Stop';
import CancelIcon from '@mui/icons-material/Cancel';
import DeleteSweepIcon from '@mui/icons-material/DeleteSweep';

import MainCard from 'ui-component/cards/MainCard';
import { getTasks, getTaskStats, stopTask, clearTaskHistory } from 'api/tasks';
import { useTaskProgress } from 'contexts/TaskProgressContext';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

import { extractUnlockSummaryFromTaskResult, formatUnlockProviderLabel } from 'views/nodes/utils';
import {
  getTaskCardSx,
  getTaskCenterTokens,
  getTaskChipSx,
  getTaskDialogPaperSx,
  getTaskIconBoxSx,
  getTaskProgressSx,
  getTaskShellSx,
  getTaskStatusMeta,
  getTaskTriggerMeta,
  getTaskTypeMeta,
  TASK_CLUSTER_ACCENT
} from 'components/taskCenterTheme';

const TASK_STATUS_ICONS = {
  pending: ScheduleIcon,
  running: PlayArrowIcon,
  completed: CheckCircleIcon,
  cancelled: CancelIcon,
  cancelling: CancelIcon,
  error: ErrorIcon
};

const TASK_TYPE_ICONS = {
  speed_test: SpeedIcon,
  sub_update: CloudSyncIcon,
  tag_rule: LocalOfferIcon,
  db_migration: StorageIcon
};

const TASK_TRIGGER_ICONS = {
  manual: PersonIcon,
  scheduled: AutoModeIcon
};

const renderUnlockDetails = (unlockSummary) => (
  <Box>
    {unlockSummary.details.map((item) => (
      <Typography key={`${item.providerLabel}-${item.status}-${item.region}`} variant="caption" display="block">
        {item.providerLabel}
        {item.region ? ` · ${item.region}` : ''}
        {item.status ? ` · ${item.status}` : ''}
        {[item.reason, item.detail].filter(Boolean).length > 0 ? ` · ${[item.reason, item.detail].filter(Boolean).join(' · ')}` : ''}
      </Typography>
    ))}
  </Box>
);

// ==============================|| STAT CARD - COMPACT ||============================== //

const StatCard = ({ title, value, icon: Icon, color, theme, tokens }) => (
  <Box
    sx={{
      ...getTaskCardSx(theme, tokens, color, { compact: true }),
      display: 'flex',
      alignItems: 'center',
      gap: 1.5,
      p: 1.5,
      borderRadius: 2.5
    }}
  >
    <Box sx={getTaskIconBoxSx(theme, tokens, color, { size: 40, radius: 1.75 })}>
      <Icon sx={{ color, fontSize: 22 }} />
    </Box>
    <Box sx={{ minWidth: 0 }}>
      <Typography variant="h5" sx={{ color, fontWeight: 700, lineHeight: 1.2 }}>
        {value}
      </Typography>
      <Typography
        variant="caption"
        sx={{
          color: tokens.secondaryText,
          fontWeight: 500,
          whiteSpace: 'nowrap'
        }}
      >
        {title}
      </Typography>
    </Box>
  </Box>
);

// ==============================|| STATUS CHIP ||============================== //

const StatusChip = ({ status, theme, tokens }) => {
  const { label, color } = getTaskStatusMeta(theme, status);
  const Icon = TASK_STATUS_ICONS[status] || TASK_STATUS_ICONS.pending;

  return (
    <Chip
      icon={<Icon sx={{ fontSize: 16 }} />}
      label={label}
      size="small"
      sx={{
        ...getTaskChipSx(theme, tokens, color, { emphasis: status === 'running' ? 'solid' : 'soft' }),
        height: 24,
        fontSize: '0.75rem',
        fontWeight: 500,
        '& .MuiChip-label': { px: 0.9 }
      }}
    />
  );
};

// ==============================|| TYPE CHIP ||============================== //

const TypeChip = ({ type, theme, tokens }) => {
  const { label, color } = getTaskTypeMeta(type);
  const Icon = TASK_TYPE_ICONS[type] || TASK_TYPE_ICONS.speed_test;

  return (
    <Chip
      icon={<Icon sx={{ fontSize: 14 }} />}
      label={label}
      size="small"
      sx={{
        ...getTaskChipSx(theme, tokens, color),
        height: 22,
        fontSize: '0.7rem',
        fontWeight: 500,
        '& .MuiChip-label': { px: 0.8 }
      }}
    />
  );
};

// ==============================|| TRIGGER CHIP ||============================== //

const TriggerChip = ({ trigger, theme, tokens }) => {
  const { label, color } = getTaskTriggerMeta(trigger);
  const Icon = TASK_TRIGGER_ICONS[trigger] || TASK_TRIGGER_ICONS.manual;

  return (
    <Chip
      icon={<Icon sx={{ fontSize: 14 }} />}
      label={label}
      size="small"
      sx={{
        ...getTaskChipSx(theme, tokens, color),
        height: 22,
        fontSize: '0.7rem',
        fontWeight: 500,
        '& .MuiChip-label': { px: 0.8 }
      }}
    />
  );
};

// ==============================|| FORMAT DATE ||============================== //

const formatDate = (dateStr) => {
  if (!dateStr) return '-';
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now - date;
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);

  if (diffMins < 1) return '刚刚';
  if (diffMins < 60) return `${diffMins}分钟前`;
  if (diffHours < 24) return `${diffHours}小时前`;

  // 超过24小时显示具体时间
  const isThisYear = date.getFullYear() === now.getFullYear();
  return date.toLocaleString('zh-CN', {
    year: isThisYear ? undefined : 'numeric',
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};

// ==============================|| FORMAT DURATION ||============================== //

const formatDuration = (startedAt, completedAt, status) => {
  if (!startedAt) return '-';

  const startTime = new Date(startedAt).getTime();
  let endTime;

  // 如果任务已完成，使用完成时间；否则使用当前时间
  if (completedAt && ['completed', 'cancelled', 'error'].includes(status)) {
    endTime = new Date(completedAt).getTime();
  } else {
    endTime = Date.now();
  }

  const durationMs = endTime - startTime;
  if (durationMs < 0) return '-';

  const seconds = Math.floor(durationMs / 1000);
  if (seconds < 60) return `${seconds}秒`;

  const minutes = Math.floor(seconds / 60);
  const secs = seconds % 60;
  if (minutes < 60) return `${minutes}分${secs}秒`;

  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;
  return `${hours}时${mins}分`;
};

const parseTaskResult = (result) => {
  if (!result) return null;
  if (typeof result === 'string') {
    try {
      return JSON.parse(result);
    } catch (error) {
      console.error('Failed to parse task result:', error);
      return null;
    }
  }
  return result;
};

const getMigrationWarnings = (task) => {
  if (task?.type !== 'db_migration') return [];
  const parsedResult = parseTaskResult(task.result);
  return Array.isArray(parsedResult?.warnings) ? parsedResult.warnings : [];
};

const getTaskUnlockSummary = (task) => {
  if (task?.type !== 'speed_test') return null;

  const unlockSummary = extractUnlockSummaryFromTaskResult(task.result);
  if (!unlockSummary || !Array.isArray(unlockSummary.providers) || unlockSummary.providers.length === 0) {
    return null;
  }

  const compactProviders = unlockSummary.providers.slice(0, 2).map((item) => {
    const providerLabel = formatUnlockProviderLabel(item.provider);
    const region = item.region ? ` ${item.region}` : '';
    const status = item.status ? ` ${item.status}` : '';
    return `${providerLabel}${region || status}`;
  });

  return {
    text: `解锁 ${compactProviders.join(' · ')}${unlockSummary.providers.length > 2 ? ` +${unlockSummary.providers.length - 2}` : ''}`,
    details: unlockSummary.providers.map((item) => ({
      providerLabel: formatUnlockProviderLabel(item.provider),
      status: item.status || '',
      region: item.region || '',
      reason: item.reason || '',
      detail: item.detail || ''
    }))
  };
};

const getMigrationWarningButtonSx = (theme, tokens) => ({
  mt: 0.75,
  minWidth: 0,
  px: 1.25,
  py: 0.5,
  borderRadius: 1.5,
  textTransform: 'none',
  fontWeight: 700,
  color: 'error.main',
  bgcolor: alpha(theme.palette.error.main, tokens.isDark ? 0.16 : 0.08),
  border: '1px solid',
  borderColor: alpha(theme.palette.error.main, tokens.isDark ? 0.32 : 0.24),
  '&:hover': {
    bgcolor: alpha(theme.palette.error.main, tokens.isDark ? 0.22 : 0.12),
    borderColor: alpha(theme.palette.error.main, 0.36)
  }
});

// ==============================|| CLEAR HISTORY DIALOG ||============================== //

const ClearHistoryDialog = ({ open, onClose, onConfirm }) => {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const tokens = getTaskCenterTokens(theme, isDark);
  const [selectedDays, setSelectedDays] = useState('30');

  const handleConfirm = () => {
    onConfirm(selectedDays === 'all' ? 0 : parseInt(selectedDays, 10));
    onClose();
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="xs"
      fullWidth
      PaperProps={{ sx: getTaskDialogPaperSx(theme, tokens, theme.palette.warning.main) }}
    >
      <DialogTitle sx={{ color: tokens.primaryText }}>清理任务历史</DialogTitle>
      <DialogContent>
        <Typography variant="body2" sx={{ mb: 2, color: tokens.secondaryText }}>
          选择要清理的任务记录范围
        </Typography>
        <FormControl component="fieldset">
          <RadioGroup value={selectedDays} onChange={(e) => setSelectedDays(e.target.value)}>
            <FormControlLabel value="7" control={<Radio size="small" />} label="7 天前的记录" />
            <FormControlLabel value="15" control={<Radio size="small" />} label="15 天前的记录" />
            <FormControlLabel value="30" control={<Radio size="small" />} label="30 天前的记录" />
            <FormControlLabel value="60" control={<Radio size="small" />} label="60 天前的记录" />
            <FormControlLabel value="90" control={<Radio size="small" />} label="90 天前的记录" />
            <FormControlLabel
              value="all"
              control={<Radio size="small" color="error" />}
              label={<Typography color="error">清理全部记录</Typography>}
            />
          </RadioGroup>
        </FormControl>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>取消</Button>
        <Button onClick={handleConfirm} variant="contained" color={selectedDays === 'all' ? 'error' : 'primary'}>
          确认清理
        </Button>
      </DialogActions>
    </Dialog>
  );
};

// ==============================|| TASK MOBILE CARD ||============================== //

const TaskMobileCard = ({ task, onStop, canStop, theme, tokens }) => {
  const migrationWarnings = useMemo(() => getMigrationWarnings(task), [task]);
  const unlockSummary = useMemo(() => getTaskUnlockSummary(task), [task]);
  const taskTypeMeta = getTaskTypeMeta(task.type);
  const durationAccent = task.status === 'running' ? theme.palette.primary.main : tokens.primaryText;

  return (
    <Card
      sx={{
        ...getTaskCardSx(theme, tokens, taskTypeMeta.color, { interactive: false }),
        borderRadius: 2.5,
        mb: 1.5
      }}
    >
      <CardContent sx={{ p: 2, '&:last-child': { pb: 2 } }}>
        {/* Header: Name + Status */}
        <Stack direction="row" justifyContent="space-between" alignItems="flex-start" mb={1.5}>
          <Box sx={{ flex: 1, minWidth: 0, mr: 1 }}>
            <Typography variant="subtitle2" fontWeight={600} noWrap sx={{ color: tokens.primaryText }}>
              {task.name}
            </Typography>
            {task.message && (
              <Tooltip title={task.message} arrow placement="top">
                <Typography variant="caption" noWrap display="block" sx={{ color: tokens.secondaryText }}>
                  {task.message}
                </Typography>
              </Tooltip>
            )}
            {migrationWarnings.length > 0 && (
              <Typography variant="caption" color="warning.main" display="block" sx={{ mt: 0.5, fontWeight: 600 }}>
                包含 {migrationWarnings.length} 条迁移警告，请在任务详情中查看
              </Typography>
            )}
            {unlockSummary && (
              <Tooltip title={renderUnlockDetails(unlockSummary)} arrow placement="top-start">
                <Typography variant="caption" color="info.main" display="block" sx={{ mt: 0.5, fontWeight: 600 }}>
                  {unlockSummary.text}
                </Typography>
              </Tooltip>
            )}
          </Box>
          <StatusChip status={task.status} theme={theme} tokens={tokens} />
        </Stack>

        {/* Chips: Type + Trigger */}
        <Stack direction="row" spacing={1} mb={1.5}>
          <TypeChip type={task.type} theme={theme} tokens={tokens} />
          <TriggerChip trigger={task.trigger} theme={theme} tokens={tokens} />
        </Stack>

        {/* Progress */}
        <Box sx={{ mb: 1.5 }}>
          <Stack direction="row" justifyContent="space-between" mb={0.5}>
            <Typography variant="caption" sx={{ color: tokens.secondaryText }}>
              进度
            </Typography>
            <Stack direction="row" spacing={1} alignItems="center">
              {/* Traffic Display for Speed Test */}
              {task.type === 'speed_test' &&
                (() => {
                  try {
                    const result = typeof task.result === 'string' ? JSON.parse(task.result) : task.result;
                    if (result?.traffic?.totalFormatted) {
                      return (
                        <Typography variant="caption" fontWeight={500} color="primary.main">
                          {result.traffic.totalFormatted}
                        </Typography>
                      );
                    }
                  } catch (e) {
                    console.error(e);
                    // ignore parse error
                  }
                  return null;
                })()}
              <Typography variant="caption" fontWeight={500}>
                {task.progress}/{task.total}
              </Typography>
            </Stack>
          </Stack>
          <LinearProgress
            variant="determinate"
            value={task.total > 0 ? (task.progress / task.total) * 100 : 0}
            sx={getTaskProgressSx(tokens, taskTypeMeta.color)}
          />
        </Box>

        {/* Footer: Time + Action */}
        <Stack direction="row" justifyContent="space-between" alignItems="center">
          <Stack direction="row" spacing={2}>
            <Box>
              <Typography variant="caption" sx={{ color: tokens.secondaryText }} display="block">
                创建时间
              </Typography>
              <Typography variant="caption" fontWeight={500} sx={{ color: tokens.primaryText }}>
                {formatDate(task.createdAt)}
              </Typography>
            </Box>
            <Box>
              <Typography variant="caption" sx={{ color: tokens.secondaryText }} display="block">
                耗时
              </Typography>
              <Typography variant="caption" fontWeight={500} sx={{ color: durationAccent }}>
                {formatDuration(task.startedAt, task.completedAt, task.status)}
              </Typography>
            </Box>
          </Stack>
          {canStop && (
            <IconButton
              size="small"
              color="error"
              onClick={() => onStop(task.id)}
              sx={{
                bgcolor: alpha(theme.palette.error.main, tokens.isDark ? 0.14 : 0.06),
                border: '1px solid',
                borderColor: alpha(theme.palette.error.main, tokens.isDark ? 0.26 : 0.16),
                '&:hover': {
                  bgcolor: alpha(theme.palette.error.main, tokens.isDark ? 0.2 : 0.1)
                }
              }}
            >
              <StopIcon fontSize="small" />
            </IconButton>
          )}
        </Stack>
      </CardContent>
    </Card>
  );
};

import TrafficStatsDialog from './TrafficStatsDialog';

// ==============================|| TASK LIST ||============================== //

export default function TaskList() {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const tokens = getTaskCenterTokens(theme, isDark);
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  const [tasks, setTasks] = useState([]);
  const [stats, setStats] = useState({});
  const [loading, setLoading] = useState(true);
  const [tabValue, setTabValue] = useState(0);
  const [typeFilter, setTypeFilter] = useState('');
  const [triggerFilter, setTriggerFilter] = useState('');
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [total, setTotal] = useState(0);
  const [clearDialogOpen, setClearDialogOpen] = useState(false);

  // Traffic Stats Dialog
  const [trafficDialogOpen, setTrafficDialogOpen] = useState(false);
  const [selectedTask, setSelectedTask] = useState(null);
  const [warningsDialogOpen, setWarningsDialogOpen] = useState(false);
  const [warningsTask, setWarningsTask] = useState(null);

  const { taskList: runningTasks, stopTask: stopRunningTask, isTaskStopping, registerOnComplete, unregisterOnComplete } = useTaskProgress();

  const runningTasksWithUnlock = useMemo(
    () => runningTasks.map((task) => ({ ...task, unlockSummary: getTaskUnlockSummary({ type: task.taskType, result: task.result }) })),
    [runningTasks]
  );
  const migrationWarningButtonSx = useMemo(() => getMigrationWarningButtonSx(theme, tokens), [theme, tokens]);

  // Check if any task has stop action available
  const hasStoppableTasks = useMemo(() => {
    return tasks.some((t) => t.status === 'running' && t.type === 'speed_test');
  }, [tasks]);

  // Get status filter based on tab
  const getStatusFilter = () => {
    const statusMap = ['', 'running', 'completed', 'cancelled', 'error'];
    return statusMap[tabValue] || '';
  };

  // Load tasks
  const loadTasks = useCallback(async () => {
    try {
      setLoading(true);
      const res = await getTasks({
        status: getStatusFilter(),
        type: typeFilter,
        trigger: triggerFilter,
        page: page + 1,
        pageSize: rowsPerPage
      });
      if (res.code === 200 || res.code === 0) {
        setTasks(res.data.items || []);
        setTotal(res.data.total || 0);
      }
    } catch (error) {
      console.error('Failed to load tasks:', error);
    } finally {
      setLoading(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, rowsPerPage, tabValue, typeFilter, triggerFilter]);

  // Load stats
  const loadStats = useCallback(async () => {
    try {
      const res = await getTaskStats();
      if (res.code === 200 || res.code === 0) {
        setStats(res.data || {});
      }
    } catch (error) {
      console.error('Failed to load stats:', error);
    }
  }, []);

  // Initial load
  useEffect(() => {
    loadTasks();
    loadStats();
  }, [loadTasks, loadStats]);

  // Auto-refresh when tasks complete/cancel/fail
  useEffect(() => {
    const handleTaskComplete = (taskInfo) => {
      // Refresh task list and stats when any task completes
      if (['completed', 'cancelled', 'error'].includes(taskInfo.status)) {
        setTimeout(() => {
          loadTasks();
          loadStats();
        }, 500); // Small delay to ensure DB is updated
      }
    };

    registerOnComplete(handleTaskComplete);
    return () => unregisterOnComplete(handleTaskComplete);
  }, [registerOnComplete, unregisterOnComplete, loadTasks, loadStats]);

  // Handle stop task
  const handleStopTask = async (taskId) => {
    try {
      await stopTask(taskId);
      // Small delay to ensure DB is updated before refresh
      setTimeout(() => {
        loadTasks();
        loadStats();
      }, 300);
    } catch (error) {
      console.error('Failed to stop task:', error);
    }
  };

  // Handle clear history
  const handleClearHistory = async (days) => {
    try {
      // days === 0 means clear all
      await clearTaskHistory(days === 0 ? {} : { days });
      loadTasks();
      loadStats();
    } catch (error) {
      console.error('Failed to clear history:', error);
    }
  };

  // Handle tab change
  const handleTabChange = (event, newValue) => {
    setTabValue(newValue);
    setPage(0);
  };

  // Handle type filter change
  const handleTypeFilterChange = (event) => {
    setTypeFilter(event.target.value);
    setPage(0);
  };

  // Handle trigger filter change
  const handleTriggerFilterChange = (event) => {
    setTriggerFilter(event.target.value);
    setPage(0);
  };

  // Handle opening traffic stats
  const handleOpenTrafficStats = (task) => {
    setSelectedTask(task);
    setTrafficDialogOpen(true);
  };

  const handleOpenMigrationWarnings = (task) => {
    setWarningsTask(task);
    setWarningsDialogOpen(true);
  };

  return (
    <MainCard
      title="任务管理"
      sx={{
        ...getTaskShellSx(theme, tokens, TASK_CLUSTER_ACCENT, { interactive: false }),
        borderRadius: 3,
        overflow: 'hidden'
      }}
      secondary={
        <Box sx={{ display: 'flex', gap: 1 }}>
          <Tooltip title="清理历史记录">
            <IconButton
              onClick={() => setClearDialogOpen(true)}
              size="small"
              sx={{
                bgcolor: alpha(theme.palette.warning.main, tokens.isDark ? 0.14 : 0.08),
                border: '1px solid',
                borderColor: alpha(theme.palette.warning.main, tokens.isDark ? 0.24 : 0.14),
                color: theme.palette.warning.main,
                '&:hover': {
                  bgcolor: alpha(theme.palette.warning.main, tokens.isDark ? 0.2 : 0.12)
                }
              }}
            >
              <DeleteSweepIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="刷新">
            <IconButton
              onClick={() => {
                loadTasks();
                loadStats();
              }}
              size="small"
              sx={{
                bgcolor: alpha(theme.palette.primary.main, tokens.isDark ? 0.14 : 0.08),
                border: '1px solid',
                borderColor: alpha(theme.palette.primary.main, tokens.isDark ? 0.24 : 0.14),
                color: theme.palette.primary.main,
                '&:hover': {
                  bgcolor: alpha(theme.palette.primary.main, tokens.isDark ? 0.2 : 0.12)
                }
              }}
            >
              <RefreshIcon />
            </IconButton>
          </Tooltip>
        </Box>
      }
    >
      {/* Stats Cards - Compact Grid */}
      <Grid container spacing={1.5} sx={{ mb: 3 }}>
        <Grid item xs={6} sm={3}>
          <StatCard
            title="运行中"
            value={stats.running || runningTasks.length || 0}
            icon={PlayArrowIcon}
            color={theme.palette.primary.main}
            theme={theme}
            tokens={tokens}
          />
        </Grid>
        <Grid item xs={6} sm={3}>
          <StatCard
            title="等待中"
            value={stats.pending || 0}
            icon={ScheduleIcon}
            color={theme.palette.warning.main}
            theme={theme}
            tokens={tokens}
          />
        </Grid>
        <Grid item xs={6} sm={3}>
          <StatCard
            title="已完成"
            value={stats.completed || 0}
            icon={CheckCircleIcon}
            color={theme.palette.success.main}
            theme={theme}
            tokens={tokens}
          />
        </Grid>
        <Grid item xs={6} sm={3}>
          <StatCard
            title="失败/取消"
            value={(stats.error || 0) + (stats.cancelled || 0)}
            icon={ErrorIcon}
            color={theme.palette.error.main}
            theme={theme}
            tokens={tokens}
          />
        </Grid>
      </Grid>

      {/* Running Tasks from SSE */}
      {runningTasks.length > 0 && (
        <Card
          sx={{
            ...getTaskCardSx(theme, tokens, theme.palette.primary.main, { interactive: false }),
            mb: 3,
            borderRadius: 2.5
          }}
        >
          <CardContent sx={{ py: 2 }}>
            <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 1.5, color: tokens.primaryText }}>
              实时任务进度
            </Typography>
            {runningTasksWithUnlock.map((task) => (
              <Box key={task.taskId} sx={{ mb: 2, '&:last-child': { mb: 0 } }}>
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 0.5 }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <TypeChip type={task.taskType} theme={theme} tokens={tokens} />
                    <Typography variant="body2" noWrap sx={{ maxWidth: isMobile ? 120 : 300, color: tokens.primaryText }}>
                      {task.currentItem || '处理中...'}
                    </Typography>
                  </Box>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Typography variant="caption" sx={{ color: tokens.secondaryText }}>
                      {task.current}/{task.total}
                    </Typography>
                    {task.taskType === 'speed_test' && (
                      <IconButton
                        size="small"
                        onClick={() => stopRunningTask(task.taskId)}
                        disabled={isTaskStopping(task.taskId)}
                        sx={{
                          bgcolor: alpha(theme.palette.error.main, tokens.isDark ? 0.14 : 0.06),
                          border: '1px solid',
                          borderColor: alpha(theme.palette.error.main, tokens.isDark ? 0.24 : 0.16),
                          '&:hover': {
                            bgcolor: alpha(theme.palette.error.main, tokens.isDark ? 0.2 : 0.1)
                          }
                        }}
                      >
                        <StopIcon fontSize="small" color="error" />
                      </IconButton>
                    )}
                  </Box>
                </Box>
                <LinearProgress
                  variant="determinate"
                  value={task.total > 0 ? (task.current / task.total) * 100 : 0}
                  sx={getTaskProgressSx(tokens, getTaskTypeMeta(task.taskType).color)}
                />
                {task.unlockSummary && (
                  <Tooltip title={renderUnlockDetails(task.unlockSummary)} arrow placement="top-start">
                    <Typography variant="caption" color="info.main" sx={{ mt: 0.5, display: 'block', fontWeight: 600 }}>
                      {task.unlockSummary.text}
                    </Typography>
                  </Tooltip>
                )}
                {/* Traffic Display for Running Task */}
                {task.traffic?.totalFormatted && (
                  <Box sx={{ mt: 0.5, display: 'flex', alignItems: 'center', gap: 0.5 }}>
                    <Typography variant="caption" sx={{ color: tokens.secondaryText }}>
                      实时流量:
                    </Typography>
                    <Typography variant="caption" fontWeight={500} color="primary.main">
                      {task.traffic.totalFormatted}
                    </Typography>
                    {/* Note: Running task usually doesn't have breakdown yet or it's partial, so no detail dialog here usually */}
                  </Box>
                )}
              </Box>
            ))}
          </CardContent>
        </Card>
      )}

      {/* Tabs */}
      <Tabs
        value={tabValue}
        onChange={handleTabChange}
        variant={isMobile ? 'scrollable' : 'standard'}
        scrollButtons={isMobile ? 'auto' : false}
        sx={{
          mb: 2,
          borderBottom: '1px solid',
          borderColor: tokens.softBorder,
          '& .MuiTab-root': {
            color: tokens.secondaryText,
            minHeight: 44
          },
          '& .Mui-selected': {
            color: theme.palette.primary.main
          }
        }}
      >
        <Tab label="全部" />
        <Tab label="运行中" />
        <Tab label="已完成" />
        <Tab label="已取消" />
        <Tab label="失败" />
      </Tabs>

      {/* Filters */}
      <Stack direction={isMobile ? 'column' : 'row'} spacing={2} sx={{ mb: 2 }} alignItems={isMobile ? 'stretch' : 'center'}>
        <FormControl size="small" sx={{ minWidth: 140 }}>
          <InputLabel id="type-filter-label">任务类型</InputLabel>
          <Select
            labelId="type-filter-label"
            value={typeFilter}
            label="任务类型"
            onChange={handleTypeFilterChange}
            sx={{
              borderRadius: 2,
              bgcolor: tokens.nestedInteractiveSurface,
              '& .MuiSelect-select': { display: 'flex', alignItems: 'center', gap: 1 }
            }}
          >
            <MenuItem value="">
              <em>全部类型</em>
            </MenuItem>
            <MenuItem value="speed_test">
              <SpeedIcon sx={{ fontSize: 16, mr: 1, color: getTaskTypeMeta('speed_test').color }} />
              节点测速
            </MenuItem>
            <MenuItem value="sub_update">
              <CloudSyncIcon sx={{ fontSize: 16, mr: 1, color: getTaskTypeMeta('sub_update').color }} />
              订阅更新
            </MenuItem>
            <MenuItem value="tag_rule">
              <LocalOfferIcon sx={{ fontSize: 16, mr: 1, color: getTaskTypeMeta('tag_rule').color }} />
              标签规则
            </MenuItem>
            <MenuItem value="db_migration">
              <StorageIcon sx={{ fontSize: 16, mr: 1, color: getTaskTypeMeta('db_migration').color }} />
              数据库迁移
            </MenuItem>
          </Select>
        </FormControl>

        <FormControl size="small" sx={{ minWidth: 140 }}>
          <InputLabel id="trigger-filter-label">触发方式</InputLabel>
          <Select
            labelId="trigger-filter-label"
            value={triggerFilter}
            label="触发方式"
            onChange={handleTriggerFilterChange}
            sx={{
              borderRadius: 2,
              bgcolor: tokens.nestedInteractiveSurface,
              '& .MuiSelect-select': { display: 'flex', alignItems: 'center', gap: 1 }
            }}
          >
            <MenuItem value="">
              <em>全部方式</em>
            </MenuItem>
            <MenuItem value="manual">
              <PersonIcon sx={{ fontSize: 16, mr: 1, color: getTaskTriggerMeta('manual').color }} />
              手动
            </MenuItem>
            <MenuItem value="scheduled">
              <AutoModeIcon sx={{ fontSize: 16, mr: 1, color: getTaskTriggerMeta('scheduled').color }} />
              定时
            </MenuItem>
          </Select>
        </FormControl>

        {/* Show clear filters button when filters are active */}
        {(typeFilter || triggerFilter) && (
          <Button
            size="small"
            variant="outlined"
            onClick={() => {
              setTypeFilter('');
              setTriggerFilter('');
              setPage(0);
            }}
            sx={{ borderRadius: 2, textTransform: 'none' }}
          >
            清除筛选
          </Button>
        )}
      </Stack>

      {/* Loading indicator */}
      {loading && <LinearProgress sx={{ mb: 1 }} />}

      {/* Mobile Card View */}
      {isMobile ? (
        <Box>
          {tasks.length === 0 ? (
            <Box sx={{ py: 4, textAlign: 'center' }}>
              <Typography sx={{ color: tokens.secondaryText }}>暂无任务记录</Typography>
            </Box>
          ) : (
            tasks.map((task) => (
              <Box key={task.id}>
                <TaskMobileCard
                  task={task}
                  onStop={handleStopTask}
                  canStop={task.status === 'running' && task.type === 'speed_test'}
                  theme={theme}
                  tokens={tokens}
                />
                {task.type === 'db_migration' && getMigrationWarnings(task).length > 0 && (
                  <Button
                    size="small"
                    fullWidth
                    startIcon={<WarningAmberIcon />}
                    sx={{
                      ...migrationWarningButtonSx,
                      mt: -1.5,
                      mb: 1.5,
                      borderTopLeftRadius: 0,
                      borderTopRightRadius: 0
                    }}
                    onClick={() => handleOpenMigrationWarnings(task)}
                  >
                    查看 {getMigrationWarnings(task).length} 条迁移警告
                  </Button>
                )}
                {/* Add a invisible click handler or a button to open details if it has traffic */}
                {task.type === 'speed_test' && task.result && task.status === 'completed' && (
                  <Button
                    size="small"
                    fullWidth
                    sx={{
                      mt: -1.5,
                      mb: 1.5,
                      borderTopLeftRadius: 0,
                      borderTopRightRadius: 0,
                      color: theme.palette.primary.main,
                      bgcolor: alpha(theme.palette.primary.main, tokens.isDark ? 0.12 : 0.04),
                      border: '1px solid',
                      borderColor: alpha(theme.palette.primary.main, tokens.isDark ? 0.2 : 0.12),
                      '&:hover': {
                        bgcolor: alpha(theme.palette.primary.main, tokens.isDark ? 0.18 : 0.08)
                      }
                    }}
                    onClick={() => handleOpenTrafficStats(task)}
                  >
                    查看流量详情
                  </Button>
                )}
              </Box>
            ))
          )}
          <TablePagination
            component="div"
            count={total}
            page={page}
            onPageChange={(e, newPage) => setPage(newPage)}
            rowsPerPage={rowsPerPage}
            onRowsPerPageChange={(e) => {
              setRowsPerPage(parseInt(e.target.value, 10));
              setPage(0);
            }}
            labelRowsPerPage=""
            rowsPerPageOptions={[10, 20, 50]}
          />
        </Box>
      ) : (
        /* Desktop Table View */
        <TableContainer
          component={Paper}
          sx={{
            ...getTaskCardSx(theme, tokens, TASK_CLUSTER_ACCENT, { interactive: false }),
            borderRadius: 2.5,
            overflow: 'hidden'
          }}
        >
          <Table size="small">
            <TableHead sx={{ bgcolor: tokens.tableHeaderSurface }}>
              <TableRow>
                <TableCell>任务名称</TableCell>
                <TableCell>类型</TableCell>
                <TableCell>触发方式</TableCell>
                <TableCell>状态</TableCell>
                <TableCell>进度</TableCell>
                <TableCell>流量</TableCell>
                <TableCell>创建时间</TableCell>
                <TableCell>耗时</TableCell>
                {hasStoppableTasks && <TableCell>操作</TableCell>}
              </TableRow>
            </TableHead>
            <TableBody>
              {tasks.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={hasStoppableTasks ? 9 : 8} align="center" sx={{ py: 4 }}>
                    <Typography sx={{ color: tokens.secondaryText }}>暂无任务记录</Typography>
                  </TableCell>
                </TableRow>
              ) : (
                tasks.map((task) => {
                  const taskUnlockSummary = getTaskUnlockSummary(task);

                  return (
                    <TableRow
                      key={task.id}
                      hover
                      sx={{
                        '&:hover': {
                          bgcolor: tokens.rowHoverSurface
                        }
                      }}
                    >
                      <TableCell>
                        <Typography variant="body2" sx={{ fontWeight: 500, color: tokens.primaryText }}>
                          {task.name}
                        </Typography>
                        {task.message && (
                          <Tooltip title={task.message} arrow placement="top-start">
                            <Typography
                              variant="caption"
                              noWrap
                              sx={{ maxWidth: 200, display: 'block', cursor: 'help', color: tokens.secondaryText }}
                            >
                              {task.message}
                            </Typography>
                          </Tooltip>
                        )}
                        {taskUnlockSummary && (
                          <Tooltip title={renderUnlockDetails(taskUnlockSummary)} arrow placement="top-start">
                            <Typography
                              variant="caption"
                              color="info.main"
                              sx={{ mt: 0.5, display: 'block', fontWeight: 600, maxWidth: 220 }}
                              noWrap
                            >
                              {taskUnlockSummary.text}
                            </Typography>
                          </Tooltip>
                        )}
                        {task.type === 'db_migration' && getMigrationWarnings(task).length > 0 && (
                          <Button
                            size="small"
                            startIcon={<WarningAmberIcon sx={{ fontSize: 16 }} />}
                            sx={migrationWarningButtonSx}
                            onClick={() => handleOpenMigrationWarnings(task)}
                          >
                            查看 {getMigrationWarnings(task).length} 条迁移警告
                          </Button>
                        )}
                      </TableCell>
                      <TableCell>
                        <TypeChip type={task.type} theme={theme} tokens={tokens} />
                      </TableCell>
                      <TableCell>
                        <TriggerChip trigger={task.trigger} theme={theme} tokens={tokens} />
                      </TableCell>
                      <TableCell>
                        <StatusChip status={task.status} theme={theme} tokens={tokens} />
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" sx={{ color: tokens.primaryText }}>
                          {task.progress}/{task.total}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        {task.type === 'speed_test'
                          ? (() => {
                              try {
                                const result = typeof task.result === 'string' ? JSON.parse(task.result) : task.result;
                                const hasTraffic = result?.traffic?.totalFormatted;

                                if (!hasTraffic) return '-';

                                return (
                                  <Box
                                    sx={{
                                      cursor: 'pointer',
                                      display: 'inline-block',
                                      color: theme.palette.primary.main,
                                      '&:hover': { opacity: 0.82 }
                                    }}
                                    onClick={() => handleOpenTrafficStats(task)}
                                  >
                                    <Typography
                                      variant="body2"
                                      color="primary.main"
                                      fontWeight={500}
                                      sx={{ textDecoration: 'underline', textUnderlineOffset: 2 }}
                                    >
                                      {result.traffic.totalFormatted}
                                    </Typography>
                                  </Box>
                                );
                              } catch (e) {
                                console.error(e);
                                return '-';
                              }
                            })()
                          : '-'}
                      </TableCell>
                      <TableCell>
                        <Tooltip title={task.createdAt ? new Date(task.createdAt).toLocaleString('zh-CN') : ''}>
                          <Typography variant="caption" sx={{ color: tokens.secondaryText }}>
                            {formatDate(task.createdAt)}
                          </Typography>
                        </Tooltip>
                      </TableCell>
                      <TableCell>
                        <Typography
                          variant="caption"
                          sx={{
                            color: task.status === 'running' ? 'primary.main' : 'text.secondary',
                            fontWeight: task.status === 'running' ? 500 : 400
                          }}
                        >
                          {formatDuration(task.startedAt, task.completedAt, task.status)}
                        </Typography>
                      </TableCell>
                      {hasStoppableTasks && (
                        <TableCell>
                          {task.status === 'running' && task.type === 'speed_test' && (
                            <Tooltip title="停止">
                              <IconButton
                                size="small"
                                onClick={() => handleStopTask(task.id)}
                                sx={{
                                  bgcolor: alpha(theme.palette.error.main, tokens.isDark ? 0.14 : 0.06),
                                  border: '1px solid',
                                  borderColor: alpha(theme.palette.error.main, tokens.isDark ? 0.24 : 0.16),
                                  '&:hover': {
                                    bgcolor: alpha(theme.palette.error.main, tokens.isDark ? 0.2 : 0.1)
                                  }
                                }}
                              >
                                <StopIcon fontSize="small" color="error" />
                              </IconButton>
                            </Tooltip>
                          )}
                        </TableCell>
                      )}
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>
          <TablePagination
            component="div"
            count={total}
            page={page}
            onPageChange={(e, newPage) => setPage(newPage)}
            rowsPerPage={rowsPerPage}
            onRowsPerPageChange={(e) => {
              setRowsPerPage(parseInt(e.target.value, 10));
              setPage(0);
            }}
            labelRowsPerPage="每页行数"
            rowsPerPageOptions={[10, 20, 50]}
          />
        </TableContainer>
      )}

      {/* Clear History Dialog */}
      <ClearHistoryDialog open={clearDialogOpen} onClose={() => setClearDialogOpen(false)} onConfirm={handleClearHistory} />

      {/* Traffic Stats Dialog */}
      <TrafficStatsDialog open={trafficDialogOpen} onClose={() => setTrafficDialogOpen(false)} task={selectedTask} />

      <Dialog
        open={warningsDialogOpen}
        onClose={() => setWarningsDialogOpen(false)}
        maxWidth="md"
        fullWidth
        PaperProps={{ sx: getTaskDialogPaperSx(theme, tokens, theme.palette.warning.main) }}
      >
        <DialogTitle sx={{ color: tokens.primaryText }}>迁移警告详情</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2}>
            {warningsTask?.name && (
              <Typography variant="body2" sx={{ color: tokens.secondaryText }}>
                任务：{warningsTask.name}
              </Typography>
            )}
            {getMigrationWarnings(warningsTask).length > 0 ? (
              <Alert severity="warning">
                <Box component="ul" sx={{ m: 0, pl: 2.5 }}>
                  {getMigrationWarnings(warningsTask).map((warning, index) => (
                    <Box component="li" key={`${warning}-${index}`} sx={{ mb: 1 }}>
                      <Typography variant="body2">{warning}</Typography>
                    </Box>
                  ))}
                </Box>
              </Alert>
            ) : (
              <Typography variant="body2" sx={{ color: tokens.secondaryText }}>
                当前任务没有可展示的迁移警告。
              </Typography>
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setWarningsDialogOpen(false)}>关闭</Button>
        </DialogActions>
      </Dialog>
    </MainCard>
  );
}
