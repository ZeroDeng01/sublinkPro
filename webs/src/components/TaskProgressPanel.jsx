import { useMemo, useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
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
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
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

const formatTime = (ms, t) => {
  if (ms < 0) return '--';
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return t('tasks.time.seconds', { count: seconds });
  const minutes = Math.floor(seconds / 60);
  const secs = seconds % 60;
  if (minutes < 60) return t('tasks.time.minutesSeconds', { minutes, seconds: secs });
  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;
  return t('tasks.time.hoursMinutes', { hours, minutes: mins });
};

// ==============================|| TASK PROGRESS ITEM ||============================== //

const TaskProgressItem = ({ task, currentTime, onStopTask, isStopping }) => {
  const { t } = useTranslation();
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
    const meta = getTaskTypeMeta(task.taskType, t);
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
  }, [task.taskType, t]);

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

    const elapsedStr = formatTime(elapsed, t);

    // Estimated remaining time (only show when progress > 2%)
    let remainingStr = null;
    if (progressRatio > 0.02 && progressRatio < 1) {
      const remaining = (elapsed / progressRatio) * (1 - progressRatio);
      remainingStr = formatTime(remaining, t);
    }

    return { elapsedStr, remainingStr };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [task.startTime, task.current, task.total, currentTime, isCompleted, isError, isCancelled, t]);

  // Format result display
  const resultDisplay = useMemo(() => {
    if (!task.result) return null;

    const unlockText = getUnlockTaskResultText(task.result, 1);

    if (task.taskType === 'speed_test' && task.result.speed !== undefined) {
      const speed = task.result.speed;
      const latency = task.result.latency;
      if (speed === -1) {
        return unlockText ? `${t('tasks.result.speedTestFailed')} · ${unlockText}` : t('tasks.result.speedTestFailed');
      }
      if (speed === 0) {
        if (latency > 0) {
          return unlockText ? `${t('tasks.result.latency', { latency })} · ${unlockText}` : t('tasks.result.latency', { latency });
        }
        return unlockText;
      }
      return unlockText ? `${speed.toFixed(2)} MB/s | ${latency}ms · ${unlockText}` : `${speed.toFixed(2)} MB/s | ${latency}ms`;
    }

    if (task.taskType === 'sub_update') {
      const { added, exists, deleted } = task.result;
      const parts = [];
      if (added !== undefined) parts.push(t('tasks.result.added', { count: added }));
      if (exists !== undefined) parts.push(t('tasks.result.exists', { count: exists }));
      if (deleted !== undefined) parts.push(t('tasks.result.deleted', { count: deleted }));
      return parts.length > 0 ? parts.join(' · ') : null;
    }

    if (task.taskType === 'tag_rule') {
      const { matchedCount, totalCount } = task.result;
      if (matchedCount !== undefined && totalCount !== undefined) {
        return t('tasks.result.matchedNodes', { matched: matchedCount, total: totalCount });
      }
    }

    if (task.taskType === 'db_migration') {
      const imported = task.result.imported || {};
      const importedKinds = Object.values(imported).filter((count) => Number(count) > 0).length;
      const warnings = task.result.warnings?.length || 0;
      if (importedKinds > 0) {
        return warnings > 0
          ? `${t('tasks.result.importedKinds', { count: importedKinds })} · ${t('tasks.result.warnings', { count: warnings })}`
          : t('tasks.result.importedKinds', { count: importedKinds });
      }
      if (warnings > 0) {
        return t('tasks.result.warnings', { count: warnings });
      }
    }

    return unlockText;
  }, [task.result, task.taskType, t]);

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
                      label={task.result.phase === 'latency' ? t('tasks.phase.latency') : t('tasks.phase.speed')}
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
                    {isCompleted
                      ? t('tasks.shortStatus.completed')
                      : isError
                        ? t('tasks.shortStatus.error')
                        : isCancelled
                          ? t('tasks.shortStatus.cancelled')
                          : isCancelling
                            ? t('tasks.shortStatus.cancelling')
                            : `${progress}%`}
                  </Typography>
                  {isActive && taskConfig.canStop && onStopTask && (
                    <Tooltip title={isCancelling ? t('tasks.actions.stopping') : t('tasks.actions.stopTask')} arrow>
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
                  {t('tasks.processingItem', { item: task.currentItem })}
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
                          {t('tasks.remainingApprox', { time: timeInfo.remainingStr })}
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
  const { t } = useTranslation();
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const tokens = getTaskCenterTokens(theme, isDark);
  const { primaryText: primaryTextColor, secondaryText: secondaryTextColor } = tokens;
  const { taskList, hasActiveTasks, stopTask, isTaskStopping } = useTaskProgress();
  const [currentTime, setCurrentTime] = useState(Date.now());

  // Manage expanded/collapsed state with localStorage persistence
  const [isExpanded, setIsExpanded] = useState(() => {
    try {
      const saved = localStorage.getItem('taskProgressPanelExpanded');
      return saved ? JSON.parse(saved) : true;
    } catch {
      return true;
    }
  });

  // Toggle expanded state and persist to localStorage
  const handleToggleExpand = () => {
    const newState = !isExpanded;
    setIsExpanded(newState);
    try {
      localStorage.setItem('taskProgressPanelExpanded', JSON.stringify(newState));
    } catch {
      // Ignore localStorage errors
    }
  };

  // Calculate task summary statistics
  const taskSummary = useMemo(() => {
    const active = taskList.filter((t) => t.status !== 'completed' && t.status !== 'error' && t.status !== 'cancelled').length;
    const completed = taskList.filter((t) => t.status === 'completed').length;
    const failed = taskList.filter((t) => t.status === 'error').length;
    const cancelled = taskList.filter((t) => t.status === 'cancelled').length;
    return { active, completed, failed, cancelled };
  }, [taskList]);

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
        <CardContent sx={{ p: 2.5, pb: isExpanded ? 2.5 : 2 }}>
          {/* Collapsible Header */}
          <Box
            onClick={handleToggleExpand}
            sx={{
              display: 'flex',
              alignItems: 'center',
              gap: 1.5,
              mb: isExpanded ? 2 : 0,
              cursor: 'pointer',
              borderRadius: 2,
              p: 1,
              mx: -1,
              transition: 'background-color 0.2s',
              '&:hover': {
                backgroundColor: alpha(theme.palette.primary.main, isDark ? 0.08 : 0.04)
              }
            }}
          >
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
              {t('tasks.progressTitle')}
            </Typography>
            <Chip
              label={t('tasks.taskCount', { count: taskList.length })}
              size="small"
              sx={{
                height: 22,
                fontSize: '0.7rem',
                fontWeight: 500,
                ...getTaskChipSx(theme, tokens, TASK_CLUSTER_ACCENT)
              }}
            />

            {/* Task Summary when collapsed */}
            {!isExpanded && (
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flexWrap: 'wrap', flex: 1 }}>
                {taskSummary.active > 0 && (
                  <Typography
                    variant="caption"
                    sx={{
                      color: theme.palette.info.main,
                      fontSize: '0.75rem',
                      fontWeight: 500
                    }}
                  >
                    {t('tasks.summary.active', { count: taskSummary.active })}
                  </Typography>
                )}
                {taskSummary.completed > 0 && (
                  <>
                    {taskSummary.active > 0 && (
                      <Typography variant="caption" sx={{ color: secondaryTextColor }}>
                        ·
                      </Typography>
                    )}
                    <Typography
                      variant="caption"
                      sx={{
                        color: theme.palette.success.main,
                        fontSize: '0.75rem',
                        fontWeight: 500
                      }}
                    >
                      {t('tasks.summary.completed', { count: taskSummary.completed })}
                    </Typography>
                  </>
                )}
                {taskSummary.failed > 0 && (
                  <>
                    {(taskSummary.active > 0 || taskSummary.completed > 0) && (
                      <Typography variant="caption" sx={{ color: secondaryTextColor }}>
                        ·
                      </Typography>
                    )}
                    <Typography
                      variant="caption"
                      sx={{
                        color: theme.palette.error.main,
                        fontSize: '0.75rem',
                        fontWeight: 500
                      }}
                    >
                      {t('tasks.summary.failed', { count: taskSummary.failed })}
                    </Typography>
                  </>
                )}
                {taskSummary.cancelled > 0 && (
                  <>
                    {(taskSummary.active > 0 || taskSummary.completed > 0 || taskSummary.failed > 0) && (
                      <Typography variant="caption" sx={{ color: secondaryTextColor }}>
                        ·
                      </Typography>
                    )}
                    <Typography
                      variant="caption"
                      sx={{
                        color: theme.palette.warning.main,
                        fontSize: '0.75rem',
                        fontWeight: 500
                      }}
                    >
                      {t('tasks.summary.cancelled', { count: taskSummary.cancelled })}
                    </Typography>
                  </>
                )}
              </Box>
            )}

            {/* Expand/Collapse Icon */}
            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
                ml: 'auto',
                transition: 'transform 0.2s'
              }}
            >
              {isExpanded ? (
                <ExpandMoreIcon sx={{ color: secondaryTextColor, fontSize: 24 }} />
              ) : (
                <ChevronRightIcon sx={{ color: secondaryTextColor, fontSize: 24 }} />
              )}
            </Box>
          </Box>

          {/* Task List */}
          <Collapse in={isExpanded} timeout={300}>
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
          </Collapse>
        </CardContent>
      </Card>
    </Collapse>
  );
};

export default TaskProgressPanel;
