import { useMemo, useState, useEffect, useRef, useCallback } from 'react';
import { useTheme, alpha, keyframes } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Fab from '@mui/material/Fab';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import LinearProgress from '@mui/material/LinearProgress';
import Chip from '@mui/material/Chip';
import Fade from '@mui/material/Fade';
import Grow from '@mui/material/Grow';
import ClickAwayListener from '@mui/material/ClickAwayListener';
import SpeedIcon from '@mui/icons-material/Speed';
import CloudSyncIcon from '@mui/icons-material/CloudSync';
import LocalOfferIcon from '@mui/icons-material/LocalOffer';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import CloseIcon from '@mui/icons-material/Close';
import { useTaskProgress } from 'contexts/TaskProgressContext';

// ==============================|| CONSTANTS ||============================== //

const STORAGE_KEY = 'task_progress_fab_position';
const DEFAULT_POSITION = { right: 24, bottom: 24 };
const DRAG_THRESHOLD = 5; // pixels to distinguish click from drag
const FAB_SIZE = 56;
const MOBILE_FAB_SIZE = 52;

// ==============================|| ANIMATIONS ||============================== //

const spin = keyframes`
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
`;

const pulse = keyframes`
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.8;
    transform: scale(0.95);
  }
`;

const slideIn = keyframes`
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
`;

const glowPulse = keyframes`
  0%, 100% {
    box-shadow: 0 0 20px 4px var(--glow-color, rgba(33, 150, 243, 0.4));
  }
  50% {
    box-shadow: 0 0 30px 8px var(--glow-color, rgba(33, 150, 243, 0.6));
  }
`;

// ==============================|| TIME FORMATTING HELPER ||============================== //

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

// ==============================|| POSITION STORAGE HELPERS ||============================== //

const loadPosition = () => {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved) {
      const pos = JSON.parse(saved);
      if (typeof pos.right === 'number' && typeof pos.bottom === 'number') {
        return pos;
      }
    }
  } catch {
    // ignore
  }
  return DEFAULT_POSITION;
};

const savePosition = (position) => {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(position));
  } catch {
    // ignore
  }
};

// ==============================|| FAB TASK ITEM ||============================== //

const FabTaskItem = ({ task, currentTime, theme }) => {
  const isDark = theme.palette.mode === 'dark';

  // Calculate progress percentage
  // 使用 task.current 保持组件响应式，不依赖 task 引用
  const progress = useMemo(() => {
    if (!task.total || task.total === 0) return 0;
    return Math.round((task.current / task.total) * 100);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [task.current, task.total]);

  // Get task icon and colors based on type - using theme colors
  const taskConfig = useMemo(() => {
    if (task.taskType === 'speed_test') {
      return {
        icon: SpeedIcon,
        gradientColors: [theme.palette.success.main, theme.palette.success.dark],
        label: '节点测速',
        accentColor: theme.palette.success.main
      };
    }
    if (task.taskType === 'tag_rule') {
      return {
        icon: LocalOfferIcon,
        gradientColors: ['#f59e0b', '#d97706'],
        label: '标签规则',
        accentColor: '#f59e0b'
      };
    }
    return {
      icon: CloudSyncIcon,
      gradientColors: [theme.palette.primary.main, theme.palette.primary.dark],
      label: '订阅更新',
      accentColor: theme.palette.primary.main
    };
  }, [task.taskType, theme.palette]);

  const Icon = taskConfig.icon;
  const isCompleted = task.status === 'completed';
  const isError = task.status === 'error';

  // Calculate time info
  const timeInfo = useMemo(() => {
    if (!task.startTime || isCompleted || isError) return null;

    const elapsed = currentTime - task.startTime;
    const progressRatio = task.total > 0 ? task.current / task.total : 0;

    const elapsedStr = formatTime(elapsed);

    // Estimated remaining time (only show when progress > 5%)
    let remainingStr = null;
    if (progressRatio > 0.05 && progressRatio < 1) {
      const remaining = (elapsed / progressRatio) * (1 - progressRatio);
      remainingStr = formatTime(remaining);
    }

    return { elapsedStr, remainingStr };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [task.startTime, task.current, task.total, currentTime, isCompleted, isError]);

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

    return null;
  }, [task.result, task.taskType]);

  return (
    <Box
      sx={{
        animation: `${slideIn} 0.3s ease-out`,
        mb: 1.5,
        '&:last-child': { mb: 0 }
      }}
    >
      <Card
        sx={{
          borderRadius: 2.5,
          background: isDark
            ? `linear-gradient(145deg, ${alpha(taskConfig.accentColor, 0.12)} 0%, ${alpha(taskConfig.accentColor, 0.05)} 100%)`
            : `linear-gradient(145deg, ${alpha(taskConfig.accentColor, 0.08)} 0%, ${alpha('#fff', 0.95)} 100%)`,
          backdropFilter: 'blur(10px)',
          border: `1px solid ${isDark ? alpha(taskConfig.accentColor, 0.2) : alpha(taskConfig.accentColor, 0.15)}`,
          overflow: 'hidden',
          position: 'relative'
        }}
      >
        {/* Progress bar at top */}
        {!isCompleted && !isError && (
          <LinearProgress
            variant="determinate"
            value={progress}
            sx={{
              height: 2,
              backgroundColor: alpha(taskConfig.accentColor, 0.1),
              '& .MuiLinearProgress-bar': {
                background: `linear-gradient(90deg, ${taskConfig.gradientColors[0]} 0%, ${taskConfig.gradientColors[1]} 100%)`
              }
            }}
          />
        )}

        <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
          <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 1.5 }}>
            {/* Icon */}
            <Box
              sx={{
                width: 32,
                height: 32,
                borderRadius: 1.5,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: isCompleted
                  ? `linear-gradient(135deg, ${theme.palette.success.main} 0%, ${theme.palette.success.dark} 100%)`
                  : isError
                    ? `linear-gradient(135deg, ${theme.palette.error.main} 0%, ${theme.palette.error.dark} 100%)`
                    : `linear-gradient(135deg, ${taskConfig.gradientColors[0]} 0%, ${taskConfig.gradientColors[1]} 100%)`,
                flexShrink: 0,
                animation: !isCompleted && !isError ? `${pulse} 2s ease-in-out infinite` : 'none'
              }}
            >
              {isCompleted ? (
                <CheckCircleIcon sx={{ color: '#fff', fontSize: 18 }} />
              ) : isError ? (
                <ErrorIcon sx={{ color: '#fff', fontSize: 18 }} />
              ) : (
                <Icon sx={{ color: '#fff', fontSize: 18 }} />
              )}
            </Box>

            {/* Content */}
            <Box sx={{ flex: 1, minWidth: 0, overflow: 'hidden' }}>
              {/* Header row */}
              <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 0.5, mb: 0.25 }}>
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    flexWrap: 'wrap',
                    gap: 0.5,
                    minWidth: 0,
                    flex: 1,
                    rowGap: 0.25
                  }}
                >
                  <Typography
                    variant="body2"
                    sx={{
                      fontWeight: 600,
                      fontSize: '0.8rem',
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
                        height: 16,
                        fontSize: '0.6rem',
                        fontWeight: 500,
                        bgcolor: alpha(taskConfig.accentColor, 0.15),
                        color: isDark ? alpha('#fff', 0.9) : taskConfig.accentColor,
                        border: `1px solid ${alpha(taskConfig.accentColor, 0.2)}`,
                        '& .MuiChip-label': { px: 0.5, overflow: 'hidden', textOverflow: 'ellipsis' },
                        maxWidth: 70
                      }}
                    />
                  )}
                </Box>
                <Typography
                  variant="caption"
                  sx={{
                    fontWeight: 600,
                    fontSize: '0.75rem',
                    color: isCompleted ? theme.palette.success.main : isError ? theme.palette.error.main : taskConfig.accentColor,
                    whiteSpace: 'nowrap'
                  }}
                >
                  {isCompleted ? '完成' : isError ? '失败' : `${progress}%`}
                </Typography>
              </Box>

              {/* Current item */}
              {task.currentItem && !isCompleted && (
                <Typography
                  variant="caption"
                  sx={{
                    color: isDark ? alpha('#fff', 0.7) : theme.palette.text.secondary,
                    fontSize: '0.7rem',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                    display: 'block',
                    mb: 0.25
                  }}
                >
                  正在处理: {task.currentItem}
                </Typography>
              )}

              {/* Progress info and time display */}
              <Box
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  flexWrap: 'wrap',
                  gap: 0.5
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', flexWrap: 'wrap', gap: 0.75 }}>
                  <Typography
                    variant="caption"
                    sx={{
                      color: isDark ? alpha('#fff', 0.6) : theme.palette.text.secondary,
                      fontSize: '0.7rem'
                    }}
                  >
                    {task.current || 0} / {task.total || 0}
                  </Typography>

                  {/* Time display */}
                  {timeInfo && (
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                      <Typography
                        variant="caption"
                        sx={{
                          color: isDark ? alpha('#fff', 0.5) : theme.palette.text.secondary,
                          fontSize: '0.65rem',
                          display: 'flex',
                          alignItems: 'center',
                          gap: 0.3
                        }}
                      >
                        <AccessTimeIcon sx={{ fontSize: 10 }} />
                        {timeInfo.elapsedStr}
                      </Typography>
                      {timeInfo.remainingStr && (
                        <Typography
                          variant="caption"
                          sx={{
                            color: isDark ? alpha('#fff', 0.5) : theme.palette.text.secondary,
                            fontSize: '0.65rem'
                          }}
                        >
                          · 剩余 ~{timeInfo.remainingStr}
                        </Typography>
                      )}
                    </Box>
                  )}
                </Box>

                {/* Result display */}
                {resultDisplay && (
                  <Typography
                    variant="caption"
                    sx={{
                      color: isDark ? alpha('#fff', 0.7) : theme.palette.text.secondary,
                      fontSize: '0.7rem',
                      fontWeight: 500
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

// ==============================|| TASK PROGRESS FAB ||============================== //

const TaskProgressFab = () => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const { taskList, hasActiveTasks } = useTaskProgress();
  const [isExpanded, setIsExpanded] = useState(false);
  const [currentTime, setCurrentTime] = useState(Date.now());
  const [position, setPosition] = useState(loadPosition);

  // Drag state refs
  const fabRef = useRef(null);
  const isDragging = useRef(false);
  const hasMoved = useRef(false);
  const startPos = useRef({ x: 0, y: 0 });
  const startRight = useRef(0);
  const startBottom = useRef(0);

  // Update currentTime every second when there are active tasks
  useEffect(() => {
    if (!hasActiveTasks) return;
    const timer = setInterval(() => setCurrentTime(Date.now()), 1000);
    return () => clearInterval(timer);
  }, [hasActiveTasks]);

  // Close panel when no tasks
  useEffect(() => {
    if (!hasActiveTasks) {
      setIsExpanded(false);
    }
  }, [hasActiveTasks]);

  // Calculate overall progress for the FAB icon
  const overallProgress = useMemo(() => {
    if (taskList.length === 0) return 0;
    const totalProgress = taskList.reduce((sum, task) => {
      if (!task.total || task.total === 0) return sum;
      return sum + (task.current / task.total) * 100;
    }, 0);
    return Math.round(totalProgress / taskList.length);
  }, [taskList]);

  // Check if any task is still running
  const hasRunningTask = useMemo(() => {
    return taskList.some((task) => task.status !== 'completed' && task.status !== 'error');
  }, [taskList]);

  // Get accent color based on task type
  const fabColor = useMemo(() => {
    if (!hasRunningTask) {
      return {
        main: theme.palette.success.main,
        dark: theme.palette.success.dark,
        glow: alpha(theme.palette.success.main, 0.5)
      };
    }
    // Check if any speed test is running
    const hasSpeedTest = taskList.some((task) => task.taskType === 'speed_test' && task.status !== 'completed' && task.status !== 'error');
    if (hasSpeedTest) {
      return {
        main: theme.palette.success.main,
        dark: theme.palette.success.dark,
        glow: alpha(theme.palette.success.main, 0.5)
      };
    }
    // Check if any tag rule is running
    const hasTagRule = taskList.some((task) => task.taskType === 'tag_rule' && task.status !== 'completed' && task.status !== 'error');
    if (hasTagRule) {
      return {
        main: '#f59e0b',
        dark: '#d97706',
        glow: alpha('#f59e0b', 0.5)
      };
    }
    return {
      main: theme.palette.primary.main,
      dark: theme.palette.primary.dark,
      glow: alpha(theme.palette.primary.main, 0.5)
    };
  }, [hasRunningTask, taskList, theme.palette]);

  // Drag handlers
  const handleDragStart = useCallback(
    (clientX, clientY) => {
      isDragging.current = true;
      hasMoved.current = false;
      startPos.current = { x: clientX, y: clientY };
      startRight.current = position.right;
      startBottom.current = position.bottom;
    },
    [position]
  );

  const handleDragMove = useCallback((clientX, clientY) => {
    if (!isDragging.current) return;

    const deltaX = startPos.current.x - clientX;
    const deltaY = startPos.current.y - clientY;

    // Check if moved enough to be considered a drag
    if (!hasMoved.current && (Math.abs(deltaX) > DRAG_THRESHOLD || Math.abs(deltaY) > DRAG_THRESHOLD)) {
      hasMoved.current = true;
    }

    if (hasMoved.current) {
      // Calculate new position with boundary constraints
      const isMobile = window.innerWidth < 600;
      const fabSize = isMobile ? MOBILE_FAB_SIZE : FAB_SIZE;
      const margin = isMobile ? 16 : 24;

      const maxRight = window.innerWidth - fabSize - margin;
      const maxBottom = window.innerHeight - fabSize - margin;

      const newRight = Math.max(margin, Math.min(maxRight, startRight.current + deltaX));
      const newBottom = Math.max(margin, Math.min(maxBottom, startBottom.current + deltaY));

      setPosition({ right: newRight, bottom: newBottom });
    }
  }, []);

  const handleDragEnd = useCallback(() => {
    if (isDragging.current && hasMoved.current) {
      // Save position to localStorage
      savePosition(position);
    }
    isDragging.current = false;
  }, [position]);

  // Mouse event handlers
  const handleMouseDown = useCallback(
    (e) => {
      e.preventDefault();
      handleDragStart(e.clientX, e.clientY);

      const handleMouseMove = (e) => handleDragMove(e.clientX, e.clientY);
      const handleMouseUp = () => {
        handleDragEnd();
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
      };

      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
    },
    [handleDragStart, handleDragMove, handleDragEnd]
  );

  // Touch event handlers
  const handleTouchStart = useCallback(
    (e) => {
      const touch = e.touches[0];
      handleDragStart(touch.clientX, touch.clientY);
    },
    [handleDragStart]
  );

  const handleTouchMove = useCallback(
    (e) => {
      const touch = e.touches[0];
      handleDragMove(touch.clientX, touch.clientY);
    },
    [handleDragMove]
  );

  const handleTouchEnd = useCallback(() => {
    handleDragEnd();
  }, [handleDragEnd]);

  const handleFabClick = useCallback(() => {
    // Only toggle if not dragged
    if (!hasMoved.current) {
      setIsExpanded((prev) => !prev);
    }
    hasMoved.current = false;
  }, []);

  const handleClickAway = useCallback(() => {
    setIsExpanded(false);
  }, []);

  if (!hasActiveTasks) return null;

  const isMobile = typeof window !== 'undefined' && window.innerWidth < 600;
  const fabSize = isMobile ? MOBILE_FAB_SIZE : FAB_SIZE;

  return (
    <ClickAwayListener onClickAway={handleClickAway}>
      <Box
        sx={{
          position: 'fixed',
          bottom: position.bottom,
          right: position.right,
          zIndex: theme.zIndex.speedDial,
          touchAction: 'none' // Prevent scroll on touch
        }}
      >
        {/* Expanded Card */}
        <Grow in={isExpanded} style={{ transformOrigin: 'bottom right' }} timeout={300}>
          <Card
            sx={{
              position: 'absolute',
              bottom: fabSize + 12,
              right: 0,
              width: { xs: 'calc(100vw - 32px)', sm: 360 },
              maxWidth: 360,
              maxHeight: { xs: 'calc(100vh - 160px)', sm: 400 },
              overflow: 'hidden',
              borderRadius: 3,
              background: isDark
                ? `linear-gradient(145deg, ${alpha('#1e1e2e', 0.95)} 0%, ${alpha('#1e1e2e', 0.9)} 100%)`
                : `linear-gradient(145deg, ${alpha('#f8fafc', 0.98)} 0%, ${alpha('#fff', 0.95)} 100%)`,
              backdropFilter: 'blur(20px)',
              border: `1px solid ${isDark ? alpha('#fff', 0.1) : alpha('#000', 0.08)}`,
              boxShadow: isDark ? '0 20px 60px -15px rgba(0, 0, 0, 0.5)' : '0 20px 60px -15px rgba(0, 0, 0, 0.2)',
              display: isExpanded ? 'block' : 'none'
            }}
          >
            <CardContent sx={{ p: 2 }}>
              {/* Header */}
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1.5 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <Box
                    sx={{
                      width: 28,
                      height: 28,
                      borderRadius: 1.25,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      background: `linear-gradient(135deg, ${theme.palette.primary.main} 0%, ${theme.palette.primary.dark} 100%)`
                    }}
                  >
                    <Typography sx={{ fontSize: '0.85rem' }}>⏳</Typography>
                  </Box>
                  <Typography variant="subtitle2" sx={{ fontWeight: 600, fontSize: '0.9rem' }}>
                    任务进度
                  </Typography>
                  <Chip
                    label={`${taskList.length} 个任务`}
                    size="small"
                    sx={{
                      height: 20,
                      fontSize: '0.65rem',
                      fontWeight: 500,
                      bgcolor: alpha(theme.palette.primary.main, 0.1),
                      color: isDark ? theme.palette.primary.light : theme.palette.primary.main
                    }}
                  />
                </Box>
                <Box
                  onClick={() => setIsExpanded(false)}
                  sx={{
                    width: 24,
                    height: 24,
                    borderRadius: 1,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    cursor: 'pointer',
                    bgcolor: isDark ? alpha('#fff', 0.05) : alpha('#000', 0.05),
                    transition: 'background-color 0.2s',
                    '&:hover': {
                      bgcolor: isDark ? alpha('#fff', 0.1) : alpha('#000', 0.1)
                    }
                  }}
                >
                  <CloseIcon sx={{ fontSize: 16, color: isDark ? alpha('#fff', 0.6) : theme.palette.text.secondary }} />
                </Box>
              </Box>

              {/* Task list */}
              <Box sx={{ maxHeight: { xs: 'calc(100vh - 240px)', sm: 320 }, overflowY: 'auto', pr: 0.5 }}>
                {taskList.map((task) => (
                  <FabTaskItem key={task.taskId} task={task} currentTime={currentTime} theme={theme} />
                ))}
              </Box>
            </CardContent>
          </Card>
        </Grow>

        {/* FAB Button */}
        <Fade in={hasActiveTasks} timeout={300}>
          <Fab
            ref={fabRef}
            color="primary"
            onMouseDown={handleMouseDown}
            onTouchStart={handleTouchStart}
            onTouchMove={handleTouchMove}
            onTouchEnd={handleTouchEnd}
            onClick={handleFabClick}
            sx={{
              '--glow-color': fabColor.glow,
              width: fabSize,
              height: fabSize,
              background: `linear-gradient(135deg, ${fabColor.main} 0%, ${fabColor.dark} 100%)`,
              boxShadow: `0 8px 32px -8px ${fabColor.glow}`,
              transition: 'transform 0.2s ease, box-shadow 0.2s ease',
              animation: hasRunningTask ? `${glowPulse} 2s ease-in-out infinite` : 'none',
              cursor: isDragging.current ? 'grabbing' : 'grab',
              '&:hover': {
                transform: 'scale(1.05)',
                boxShadow: `0 12px 40px -8px ${fabColor.glow}`
              },
              '&:active': {
                cursor: 'grabbing'
              }
            }}
          >
            {/* Inner content with spinning animation */}
            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                position: 'relative'
              }}
            >
              {/* Spinning ring for running tasks */}
              {hasRunningTask && (
                <Box
                  sx={{
                    position: 'absolute',
                    width: fabSize - 12,
                    height: fabSize - 12,
                    borderRadius: '50%',
                    border: '2px solid transparent',
                    borderTopColor: alpha('#fff', 0.5),
                    animation: `${spin} 1.5s linear infinite`
                  }}
                />
              )}
              {/* Icon */}
              {hasRunningTask ? (
                <CloudSyncIcon sx={{ fontSize: 26, color: '#fff' }} />
              ) : (
                <CheckCircleIcon sx={{ fontSize: 26, color: '#fff' }} />
              )}
            </Box>

            {/* Task count badge */}
            {taskList.length > 1 && (
              <Box
                sx={{
                  position: 'absolute',
                  top: -4,
                  right: -4,
                  minWidth: 20,
                  height: 20,
                  borderRadius: 10,
                  bgcolor: theme.palette.error.main,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  px: 0.5,
                  boxShadow: `0 2px 8px ${alpha(theme.palette.error.main, 0.4)}`
                }}
              >
                <Typography sx={{ fontSize: '0.7rem', fontWeight: 600, color: '#fff' }}>{taskList.length}</Typography>
              </Box>
            )}

            {/* Progress ring */}
            {hasRunningTask && (
              <Box
                sx={{
                  position: 'absolute',
                  inset: -3,
                  borderRadius: '50%',
                  background: `conic-gradient(
                                        ${alpha('#fff', 0.8)} ${overallProgress * 3.6}deg,
                                        transparent ${overallProgress * 3.6}deg
                                    )`,
                  opacity: 0.35,
                  pointerEvents: 'none'
                }}
              />
            )}
          </Fab>
        </Fade>
      </Box>
    </ClickAwayListener>
  );
};

export default TaskProgressFab;
