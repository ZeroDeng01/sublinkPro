import { useMemo } from 'react';
import { useTheme, alpha, keyframes } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import LinearProgress from '@mui/material/LinearProgress';
import Chip from '@mui/material/Chip';
import Collapse from '@mui/material/Collapse';
import SpeedIcon from '@mui/icons-material/Speed';
import CloudSyncIcon from '@mui/icons-material/CloudSync';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import { useTaskProgress } from 'contexts/TaskProgressContext';

// ==============================|| ANIMATIONS ||============================== //

const pulse = keyframes`
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.6;
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

// ==============================|| TASK PROGRESS ITEM ||============================== //

const TaskProgressItem = ({ task }) => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  // Calculate progress percentage
  const progress = useMemo(() => {
    if (!task.total || task.total === 0) return 0;
    return Math.round((task.current / task.total) * 100);
  }, [task.current, task.total]);

  // Get task icon and colors based on type
  const taskConfig = useMemo(() => {
    if (task.taskType === 'speed_test') {
      return {
        icon: SpeedIcon,
        gradientColors: ['#10b981', '#059669'],
        label: '节点测速',
        accentColor: '#10b981'
      };
    }
    return {
      icon: CloudSyncIcon,
      gradientColors: ['#6366f1', '#8b5cf6'],
      label: '订阅更新',
      accentColor: '#6366f1'
    };
  }, [task.taskType]);

  const Icon = taskConfig.icon;
  const isCompleted = task.status === 'completed';
  const isError = task.status === 'error';

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
          borderRadius: 3,
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
              height: 3,
              backgroundColor: alpha(taskConfig.accentColor, 0.1),
              '& .MuiLinearProgress-bar': {
                background: `linear-gradient(90deg, ${taskConfig.gradientColors[0]} 0%, ${taskConfig.gradientColors[1]} 100%)`
              }
            }}
          />
        )}

        <CardContent sx={{ py: 2, px: 2.5 }}>
          <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 2 }}>
            {/* Icon */}
            <Box
              sx={{
                width: 40,
                height: 40,
                borderRadius: 2,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: isCompleted
                  ? 'linear-gradient(135deg, #10b981 0%, #059669 100%)'
                  : isError
                    ? 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)'
                    : `linear-gradient(135deg, ${taskConfig.gradientColors[0]} 0%, ${taskConfig.gradientColors[1]} 100%)`,
                flexShrink: 0,
                animation: !isCompleted && !isError ? `${pulse} 2s ease-in-out infinite` : 'none'
              }}
            >
              {isCompleted ? (
                <CheckCircleIcon sx={{ color: '#fff', fontSize: 22 }} />
              ) : isError ? (
                <ErrorIcon sx={{ color: '#fff', fontSize: 22 }} />
              ) : (
                <Icon sx={{ color: '#fff', fontSize: 22 }} />
              )}
            </Box>

            {/* Content */}
            <Box sx={{ flex: 1, minWidth: 0 }}>
              {/* Header row */}
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1, mb: 0.5 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, minWidth: 0 }}>
                  <Typography
                    variant="subtitle2"
                    sx={{
                      fontWeight: 600,
                      color: isDark ? '#fff' : theme.palette.text.primary,
                      whiteSpace: 'nowrap'
                    }}
                  >
                    {taskConfig.label}
                  </Typography>
                  {task.taskName && task.taskType === 'sub_update' && (
                    <Chip
                      label={task.taskName}
                      size="small"
                      sx={{
                        height: 20,
                        fontSize: '0.7rem',
                        fontWeight: 500,
                        bgcolor: alpha(taskConfig.accentColor, 0.15),
                        color: isDark ? alpha('#fff', 0.9) : taskConfig.accentColor,
                        border: `1px solid ${alpha(taskConfig.accentColor, 0.2)}`
                      }}
                    />
                  )}
                </Box>
                <Typography
                  variant="caption"
                  sx={{
                    fontWeight: 600,
                    color: isCompleted ? '#10b981' : isError ? '#ef4444' : taskConfig.accentColor,
                    whiteSpace: 'nowrap'
                  }}
                >
                  {isCompleted ? '完成' : isError ? '失败' : `${progress}%`}
                </Typography>
              </Box>

              {/* Current item */}
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

              {/* Progress info */}
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1 }}>
                <Typography
                  variant="caption"
                  sx={{
                    color: isDark ? alpha('#fff', 0.6) : theme.palette.text.secondary,
                    fontSize: '0.75rem'
                  }}
                >
                  {task.current || 0} / {task.total || 0}
                </Typography>

                {/* Result display */}
                {resultDisplay && (
                  <Typography
                    variant="caption"
                    sx={{
                      color: isDark ? alpha('#fff', 0.7) : theme.palette.text.secondary,
                      fontSize: '0.75rem',
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

// ==============================|| TASK PROGRESS PANEL ||============================== //

const TaskProgressPanel = () => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const { taskList, hasActiveTasks } = useTaskProgress();

  return (
    <Collapse in={hasActiveTasks} unmountOnExit timeout={300}>
      <Card
        sx={{
          mb: 4,
          borderRadius: 4,
          background: isDark
            ? `linear-gradient(145deg, ${alpha('#1e1e2e', 0.8)} 0%, ${alpha('#1e1e2e', 0.6)} 100%)`
            : `linear-gradient(145deg, ${alpha('#f8fafc', 0.95)} 0%, ${alpha('#fff', 0.9)} 100%)`,
          backdropFilter: 'blur(20px)',
          border: `1px solid ${isDark ? alpha('#fff', 0.08) : alpha('#000', 0.06)}`,
          overflow: 'hidden'
        }}
      >
        <CardContent sx={{ p: 2.5 }}>
          {/* Header */}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 2 }}>
            <Box
              sx={{
                width: 32,
                height: 32,
                borderRadius: 1.5,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)'
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
                bgcolor: alpha('#6366f1', 0.1),
                color: isDark ? '#a5b4fc' : '#6366f1'
              }}
            />
          </Box>

          {/* Task list */}
          <Box>
            {taskList.map((task) => (
              <TaskProgressItem key={task.taskId} task={task} />
            ))}
          </Box>
        </CardContent>
      </Card>
    </Collapse>
  );
};

export default TaskProgressPanel;
