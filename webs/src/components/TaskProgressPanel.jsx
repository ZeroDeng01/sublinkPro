import { useMemo, useState, useEffect } from 'react';
import { useTheme, alpha } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import LinearProgress from '@mui/material/LinearProgress';
import Chip from '@mui/material/Chip';
import Collapse from '@mui/material/Collapse';
import IconButton from '@mui/material/IconButton';
import CircularProgress from '@mui/material/CircularProgress';
import Tooltip from '@mui/material/Tooltip';
import SpeedIcon from '@mui/icons-material/Speed';
import CloudSyncIcon from '@mui/icons-material/CloudSync';
import LocalOfferIcon from '@mui/icons-material/LocalOffer';
import StorageIcon from '@mui/icons-material/Storage';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import StopIcon from '@mui/icons-material/Stop';
import CancelIcon from '@mui/icons-material/Cancel';
import { useTaskProgress } from 'contexts/TaskProgressContext';

const formatTime = (ms) => {
  if (ms < 0) return '--';
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `${seconds}秒`;
  const minutes = Math.floor(seconds / 60);
  const secs = seconds % 60;
  if (minutes < 60) return `${minutes}分${secs}秒`;
  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;
  return `${hours}时${mins}分`;
};

const getPanelSurface = (theme, accentColor) => ({
  backgroundColor: theme.palette.mode === 'dark' ? alpha(theme.palette.background.paper, 0.94) : theme.palette.background.paper,
  border: `1px solid ${theme.palette.mode === 'dark' ? alpha(theme.palette.common.white, 0.08) : alpha(accentColor, 0.12)}`,
  boxShadow: theme.palette.mode === 'dark' ? 'none' : '0 1px 3px rgba(15, 23, 42, 0.06)',
  transition: 'border-color 0.2s ease, box-shadow 0.2s ease',
  '&:hover': {
    borderColor: theme.palette.mode === 'dark' ? alpha(accentColor, 0.24) : alpha(accentColor, 0.2),
    boxShadow: theme.palette.mode === 'dark' ? 'none' : '0 4px 12px rgba(15, 23, 42, 0.08)'
  }
});

const getAccentChipSx = (theme, accentColor) => ({
  bgcolor: alpha(accentColor, theme.palette.mode === 'dark' ? 0.18 : 0.1),
  color: theme.palette.mode === 'dark' ? alpha('#fff', 0.92) : accentColor,
  border: `1px solid ${alpha(accentColor, theme.palette.mode === 'dark' ? 0.3 : 0.18)}`,
  '&:hover': {
    bgcolor: alpha(accentColor, theme.palette.mode === 'dark' ? 0.24 : 0.14)
  }
});

// ==============================|| TASK PROGRESS ITEM ||============================== //

const TaskProgressItem = ({ task, currentTime, onStopTask, isStopping }) => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  // Calculate progress percentage
  const progress = useMemo(() => {
    if (!task.total || task.total === 0) return 0;
    return Math.round((task.current / task.total) * 100);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [task.current, task.total]);

  // Get task icon and colors based on type
  const taskConfig = useMemo(() => {
    if (task.taskType === 'speed_test') {
      return {
        icon: SpeedIcon,
        gradientColors: ['#10b981', '#059669'],
        label: '节点测速',
        accentColor: '#10b981',
        canStop: true // speed_test can be stopped
      };
    }
    if (task.taskType === 'tag_rule') {
      return {
        icon: LocalOfferIcon,
        gradientColors: ['#f59e0b', '#d97706'],
        label: '标签规则',
        accentColor: '#f59e0b',
        canStop: false
      };
    }
    if (task.taskType === 'db_migration') {
      return {
        icon: StorageIcon,
        gradientColors: ['#0284c7', '#0369a1'],
        label: '数据库迁移',
        accentColor: '#0284c7',
        canStop: false
      };
    }
    return {
      icon: CloudSyncIcon,
      gradientColors: ['#6366f1', '#8b5cf6'],
      label: '订阅更新',
      accentColor: '#6366f1',
      canStop: false
    };
  }, [task.taskType]);

  const Icon = taskConfig.icon;
  const isCompleted = task.status === 'completed';
  const isError = task.status === 'error';
  const isCancelled = task.status === 'cancelled';
  const isCancelling = task.status === 'cancelling' || isStopping;
  const isActive = !isCompleted && !isError && !isCancelled;

  // Calculate time info
  const timeInfo = useMemo(() => {
    if (!task.startTime || isCompleted || isError || isCancelled) return null;

    const elapsed = currentTime - task.startTime;
    const progressRatio = task.total > 0 ? task.current / task.total : 0;

    const elapsedStr = formatTime(elapsed);

    // Estimated remaining time (only show when progress > 2%)
    let remainingStr = null;
    if (progressRatio > 0.02 && progressRatio < 1) {
      const remaining = (elapsed / progressRatio) * (1 - progressRatio);
      remainingStr = formatTime(remaining);
    }

    return { elapsedStr, remainingStr };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [task.startTime, task.current, task.total, currentTime, isCompleted, isError, isCancelled]);

  // Format result display
  const resultDisplay = useMemo(() => {
    if (!task.result) return null;

    if (task.taskType === 'speed_test' && task.result.speed !== undefined) {
      const speed = task.result.speed;
      const latency = task.result.latency;
      if (speed === -1) {
        return '测速失败';
      }
      if (speed === 0) {
        return latency > 0 ? `延迟 ${latency}ms` : null;
      }
      return `${speed.toFixed(2)} MB/s | ${latency}ms`;
    }

    if (task.taskType === 'sub_update') {
      const { added, exists, deleted } = task.result;
      const parts = [];
      if (added !== undefined) parts.push(`新增 ${added}`);
      if (exists !== undefined) parts.push(`已存在 ${exists}`);
      if (deleted !== undefined) parts.push(`删除 ${deleted}`);
      return parts.length > 0 ? parts.join(' · ') : null;
    }

    if (task.taskType === 'tag_rule') {
      const { matchedCount, totalCount } = task.result;
      if (matchedCount !== undefined && totalCount !== undefined) {
        return `匹配 ${matchedCount} / ${totalCount} 节点`;
      }
    }

    if (task.taskType === 'db_migration') {
      const imported = task.result.imported || {};
      const importedKinds = Object.values(imported).filter((count) => Number(count) > 0).length;
      const warnings = task.result.warnings?.length || 0;
      if (importedKinds > 0) {
        return warnings > 0 ? `导入 ${importedKinds} 类数据 · ${warnings} 条警告` : `导入 ${importedKinds} 类数据`;
      }
      if (warnings > 0) {
        return `${warnings} 条警告`;
      }
    }

    return null;
  }, [task.result, task.taskType]);

  return (
    <Box
      sx={{
        mb: 1.5,
        '&:last-child': { mb: 0 }
      }}
    >
      <Card
        sx={{
          ...getPanelSurface(theme, taskConfig.accentColor),
          borderRadius: 3,
          overflow: 'hidden',
          position: 'relative',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            height: 3,
            backgroundColor: isCancelling ? '#f59e0b' : isError ? '#ef4444' : isCompleted ? '#10b981' : taskConfig.accentColor
          }
        }}
      >
        {isActive && !isCancelling && (
          <LinearProgress
            variant="determinate"
            value={progress}
            sx={{
              height: 4,
              backgroundColor: alpha(taskConfig.accentColor, 0.1),
              '& .MuiLinearProgress-bar': {
                backgroundColor: taskConfig.accentColor
              }
            }}
          />
        )}
        {isCancelling && (
          <LinearProgress
            sx={{
              height: 4,
              backgroundColor: alpha('#f59e0b', 0.1),
              '& .MuiLinearProgress-bar': {
                backgroundColor: '#f59e0b'
              }
            }}
          />
        )}

        <CardContent sx={{ py: 2, px: 2.5 }}>
          <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 2 }}>
            <Box
              sx={{
                width: 40,
                height: 40,
                borderRadius: 2,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                backgroundColor: alpha(
                  isCompleted ? '#10b981' : isError ? '#ef4444' : isCancelled || isCancelling ? '#f59e0b' : taskConfig.accentColor,
                  isDark ? 0.18 : 0.1
                ),
                border: `1px solid ${alpha(
                  isCompleted ? '#10b981' : isError ? '#ef4444' : isCancelled || isCancelling ? '#f59e0b' : taskConfig.accentColor,
                  isDark ? 0.3 : 0.18
                )}`,
                flexShrink: 0
              }}
            >
              {isCompleted ? (
                <CheckCircleIcon sx={{ color: '#10b981', fontSize: 22 }} />
              ) : isError ? (
                <ErrorIcon sx={{ color: '#ef4444', fontSize: 22 }} />
              ) : isCancelled || isCancelling ? (
                <CancelIcon sx={{ color: '#f59e0b', fontSize: 22 }} />
              ) : (
                <Icon sx={{ color: taskConfig.accentColor, fontSize: 22 }} />
              )}
            </Box>

            <Box sx={{ flex: 1, minWidth: 0, overflow: 'hidden' }}>
              <Box
                sx={{
                  display: 'flex',
                  alignItems: 'flex-start',
                  justifyContent: 'space-between',
                  gap: { xs: 0.5, sm: 1 },
                  mb: 0.5
                }}
              >
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    flexWrap: 'wrap',
                    gap: 0.5,
                    minWidth: 0,
                    flex: 1,
                    rowGap: 0.5
                  }}
                >
                  <Typography
                    variant="subtitle2"
                    sx={{
                      fontWeight: 600,
                      color: isDark ? '#fff' : theme.palette.text.primary,
                      whiteSpace: 'nowrap',
                      flexShrink: 0
                    }}
                  >
                    {taskConfig.label}
                  </Typography>
                  {task.taskName && (
                    <Chip
                      label={task.taskName}
                      size="small"
                      sx={{
                        height: 18,
                        fontSize: '0.65rem',
                        fontWeight: 500,
                        bgcolor: alpha(taskConfig.accentColor, 0.15),
                        color: isDark ? alpha('#fff', 0.9) : taskConfig.accentColor,
                        border: `1px solid ${alpha(taskConfig.accentColor, 0.2)}`,
                        maxWidth: { xs: 80, sm: 100 },
                        '& .MuiChip-label': {
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          px: 0.75
                        }
                      }}
                    />
                  )}
                  {task.taskType === 'speed_test' && task.result?.phase && isActive && !isCancelling && (
                    <Chip
                      label={task.result.phase === 'latency' ? '延迟测试' : '速度测试'}
                      size="small"
                      sx={{
                        height: 18,
                        fontSize: '0.65rem',
                        fontWeight: 500,
                        flexShrink: 0,
                        ...getAccentChipSx(theme, task.result.phase === 'latency' ? '#06b6d4' : '#f59e0b'),
                        '& .MuiChip-label': { px: 0.75 }
                      }}
                    />
                  )}
                </Box>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                  <Typography
                    variant="caption"
                    sx={{
                      fontWeight: 600,
                      color: isCompleted
                        ? '#10b981'
                        : isError
                          ? '#ef4444'
                          : isCancelled
                            ? '#f59e0b'
                            : isCancelling
                              ? '#f59e0b'
                              : taskConfig.accentColor,
                      whiteSpace: 'nowrap'
                    }}
                  >
                    {isCompleted ? '完成' : isError ? '失败' : isCancelled ? '已取消' : isCancelling ? '停止中...' : `${progress}%`}
                  </Typography>
                  {isActive && taskConfig.canStop && onStopTask && (
                    <Tooltip title={isCancelling ? '正在停止...' : '停止任务'} arrow>
                      <span>
                        <IconButton
                          size="small"
                          onClick={() => onStopTask(task.taskId)}
                          disabled={isCancelling}
                          sx={{
                              p: 0.5,
                              color: isCancelling ? alpha('#f59e0b', 0.5) : '#ef4444',
                              '&:hover': {
                                bgcolor: alpha('#ef4444', 0.08)
                              }
                            }}
                        >
                          {isCancelling ? <CircularProgress size={16} color="inherit" /> : <StopIcon sx={{ fontSize: 18 }} />}
                        </IconButton>
                      </span>
                    </Tooltip>
                  )}
                </Box>
              </Box>

              {task.currentItem && !isCompleted && (
                <Typography
                  variant="body2"
                  sx={{
                    color: isDark ? alpha('#fff', 0.7) : theme.palette.text.secondary,
                    fontSize: '0.8rem',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                    mb: 0.5
                  }}
                >
                  正在处理: {task.currentItem}
                </Typography>
              )}

              <Box
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  flexWrap: 'wrap',
                  gap: { xs: 0.5, sm: 1 },
                  rowGap: 0.5
                }}
              >
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    flexWrap: 'wrap',
                    gap: { xs: 0.5, sm: 1.5 },
                    rowGap: 0.5
                  }}
                >
                  <Typography
                    variant="caption"
                    sx={{
                      color: isDark ? alpha('#fff', 0.6) : theme.palette.text.secondary,
                      fontSize: { xs: '0.7rem', sm: '0.75rem' },
                      whiteSpace: 'nowrap'
                    }}
                  >
                    {task.current || 0} / {task.total || 0}
                  </Typography>

                  {timeInfo && (
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: { xs: 0.5, sm: 1 } }}>
                      <Typography
                        variant="caption"
                        sx={{
                          color: isDark ? alpha('#fff', 0.5) : theme.palette.text.secondary,
                          fontSize: { xs: '0.65rem', sm: '0.7rem' },
                          display: 'flex',
                          alignItems: 'center',
                          gap: 0.3,
                          whiteSpace: 'nowrap'
                        }}
                      >
                        <AccessTimeIcon sx={{ fontSize: { xs: 10, sm: 12 } }} />
                        {timeInfo.elapsedStr}
                      </Typography>
                      {timeInfo.remainingStr && (
                        <Typography
                          variant="caption"
                          sx={{
                            color: isDark ? alpha('#fff', 0.5) : theme.palette.text.secondary,
                            fontSize: { xs: '0.65rem', sm: '0.7rem' },
                            whiteSpace: 'nowrap'
                          }}
                        >
                          · 剩余 ~{timeInfo.remainingStr}
                        </Typography>
                      )}
                    </Box>
                  )}
                </Box>

                {resultDisplay && (
                  <Typography
                    variant="caption"
                    sx={{
                      color: isDark ? alpha('#fff', 0.7) : theme.palette.text.secondary,
                      fontSize: { xs: '0.7rem', sm: '0.75rem' },
                      fontWeight: 500,
                      whiteSpace: 'nowrap'
                    }}
                  >
                    {resultDisplay}
                  </Typography>
                )}
              </Box>
            </Box>
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
};

// ==============================|| TASK PROGRESS PANEL ||============================== //

const TaskProgressPanel = () => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const { taskList, hasActiveTasks, stopTask, isTaskStopping } = useTaskProgress();
  const [currentTime, setCurrentTime] = useState(Date.now());

  // Update currentTime every second when there are active tasks
  useEffect(() => {
    if (!hasActiveTasks) return;
    const timer = setInterval(() => setCurrentTime(Date.now()), 1000);
    return () => clearInterval(timer);
  }, [hasActiveTasks]);

  return (
    <Collapse in={hasActiveTasks} unmountOnExit timeout={300}>
      <Card
        sx={{
          ...getPanelSurface(theme, '#6366f1'),
          mb: 4,
          borderRadius: 4,
          overflow: 'hidden',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            height: 3,
            backgroundColor: '#6366f1'
          }
        }}
      >
        <CardContent sx={{ p: 2.5 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 2 }}>
            <Box
              sx={{
                width: 32,
                height: 32,
                borderRadius: 1.5,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                backgroundColor: alpha('#6366f1', isDark ? 0.18 : 0.1),
                border: `1px solid ${alpha('#6366f1', isDark ? 0.3 : 0.18)}`
              }}
            >
              <Typography sx={{ fontSize: '1rem' }}>⏳</Typography>
            </Box>
            <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
              任务进度
            </Typography>
            <Chip
              label={`${taskList.length} 个任务`}
              size="small"
              sx={{
                height: 22,
                fontSize: '0.7rem',
                fontWeight: 500,
                ...getAccentChipSx(theme, '#6366f1')
              }}
            />
          </Box>

          <Box>
            {taskList.map((task) => (
              <TaskProgressItem
                key={task.taskId}
                task={task}
                currentTime={currentTime}
                onStopTask={stopTask}
                isStopping={isTaskStopping(task.taskId)}
              />
            ))}
          </Box>
        </CardContent>
      </Card>
    </Collapse>
  );
};

export default TaskProgressPanel;
