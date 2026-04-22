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
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

import { getUnlockTaskResultText } from 'views/nodes/utils';
import {
  getTaskActionButtonSx,
  getTaskCardSx,
  getTaskCenterTokens,
  getTaskChipSx,
  getTaskIconBoxSx,
  getTaskProgressSx,
  getTaskShellSx,
  getTaskTypeMeta,
  TASK_CLUSTER_ACCENT
} from 'components/taskCenterTheme';

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

// ==============================|| TASK PROGRESS ITEM ||============================== //

const TaskProgressItem = ({ task, currentTime, onStopTask, isStopping }) => {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const tokens = getTaskCenterTokens(theme, isDark);
  const { primaryText: primaryTextColor, secondaryText: secondaryTextColor, tertiaryText: tertiaryTextColor } = tokens;

  // Calculate progress percentage
  const progress = useMemo(() => {
    if (!task.total || task.total === 0) return 0;
    return Math.round((task.current / task.total) * 100);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [task.current, task.total]);

  // Get task icon and colors based on type
  const taskConfig = useMemo(() => {
    const meta = getTaskTypeMeta(task.taskType);
    const iconMap = {
      speed_test: SpeedIcon,
      sub_update: CloudSyncIcon,
      tag_rule: LocalOfferIcon,
      db_migration: StorageIcon
    };

    return {
      icon: iconMap[task.taskType] || CloudSyncIcon,
      label: meta.label,
      accentColor: meta.color,
      canStop: task.taskType === 'speed_test'
    };
  }, [task.taskType]);

  const Icon = taskConfig.icon;
  const isCompleted = task.status === 'completed';
  const isError = task.status === 'error';
  const isCancelled = task.status === 'cancelled';
  const isCancelling = task.status === 'cancelling' || isStopping;
  const isActive = !isCompleted && !isError && !isCancelled;
  const successColor = theme.palette.success.main;
  const errorColor = theme.palette.error.main;
  const warningColor = theme.palette.warning.main;
  const stateAccentColor = isCompleted
    ? successColor
    : isError
      ? errorColor
      : isCancelled || isCancelling
        ? warningColor
        : taskConfig.accentColor;

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

    const unlockText = getUnlockTaskResultText(task.result, 1);

    if (task.taskType === 'speed_test' && task.result.speed !== undefined) {
      const speed = task.result.speed;
      const latency = task.result.latency;
      if (speed === -1) {
        return unlockText ? `测速失败 · ${unlockText}` : '测速失败';
      }
      if (speed === 0) {
        if (latency > 0) {
          return unlockText ? `延迟 ${latency}ms · ${unlockText}` : `延迟 ${latency}ms`;
        }
        return unlockText;
      }
      return unlockText ? `${speed.toFixed(2)} MB/s | ${latency}ms · ${unlockText}` : `${speed.toFixed(2)} MB/s | ${latency}ms`;
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

    return unlockText;
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
          ...getTaskCardSx(theme, tokens, taskConfig.accentColor),
          borderRadius: 3,
          overflow: 'hidden'
        }}
      >
        {isActive && !isCancelling && (
          <LinearProgress variant="determinate" value={progress} sx={getTaskProgressSx(tokens, taskConfig.accentColor, { height: 4 })} />
        )}
        {isCancelling && <LinearProgress sx={getTaskProgressSx(tokens, warningColor, { height: 4 })} />}

        <CardContent sx={{ py: 2, px: 2.5 }}>
          <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 2 }}>
            <Box
              sx={{
                ...getTaskIconBoxSx(theme, tokens, stateAccentColor)
              }}
            >
              {isCompleted ? (
                <CheckCircleIcon sx={{ color: successColor, fontSize: 22 }} />
              ) : isError ? (
                <ErrorIcon sx={{ color: errorColor, fontSize: 22 }} />
              ) : isCancelled || isCancelling ? (
                <CancelIcon sx={{ color: warningColor, fontSize: 22 }} />
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
                      color: primaryTextColor,
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
                        ...getTaskChipSx(theme, tokens, taskConfig.accentColor),
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
                        ...getTaskChipSx(theme, tokens, task.result.phase === 'latency' ? '#06b6d4' : '#f59e0b'),
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
                        ? successColor
                        : isError
                          ? errorColor
                          : isCancelled
                            ? warningColor
                            : isCancelling
                              ? warningColor
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
                            ...getTaskActionButtonSx(theme, tokens, errorColor),
                            p: 0.5,
                            color: isCancelling ? alpha(warningColor, 0.6) : errorColor,
                            minWidth: 0,
                            borderRadius: 1.5
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
                    color: secondaryTextColor,
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
                      color: secondaryTextColor,
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
                          color: tertiaryTextColor,
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
                            color: tertiaryTextColor,
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
                      color: secondaryTextColor,
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
  const { isDark } = useResolvedColorScheme();
  const tokens = getTaskCenterTokens(theme, isDark);
  const { primaryText: primaryTextColor } = tokens;
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
          ...getTaskShellSx(theme, tokens, TASK_CLUSTER_ACCENT, { interactive: false }),
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
                ...getTaskIconBoxSx(theme, tokens, TASK_CLUSTER_ACCENT),
                width: 32,
                height: 32,
                borderRadius: 1.5
              }}
            >
              <Typography sx={{ fontSize: '1rem' }}>⏳</Typography>
            </Box>
            <Typography variant="subtitle1" sx={{ fontWeight: 600, color: primaryTextColor }}>
              任务进度
            </Typography>
            <Chip
              label={`${taskList.length} 个任务`}
              size="small"
              sx={{
                height: 22,
                fontSize: '0.7rem',
                fontWeight: 500,
                ...getTaskChipSx(theme, tokens, TASK_CLUSTER_ACCENT)
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
