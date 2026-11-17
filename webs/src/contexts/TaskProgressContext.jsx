import PropTypes from 'prop-types';
import { createContext, useContext, useState, useCallback, useMemo, useEffect, useRef } from 'react';
import { stopTask as stopTaskApi, getTask, getRunningTasks } from 'api/tasks';

// ==============================|| TASK PROGRESS CONTEXT ||============================== //

const TaskProgressContext = createContext(null);

// Auto-cleanup completed tasks after this duration (ms)
const COMPLETED_TASK_CLEANUP_DELAY = 3000;

// 有任务在跑时的兜底对账间隔（ms）：SSE 丢事件时由轮询收尾
const RUNNING_TASK_SYNC_INTERVAL = 15000;

const TERMINAL_STATUSES = ['completed', 'error', 'cancelled'];

// ==============================|| TASK PROGRESS PROVIDER ||============================== //

export function TaskProgressProvider({ children }) {
  // Map of taskId -> taskData
  const [tasks, setTasks] = useState(new Map());

  // Track tasks that are being stopped
  const [stoppingTasks, setStoppingTasks] = useState(new Set());

  // Callbacks to invoke when a task completes
  const onCompleteCallbacksRef = useRef(new Set());

  // 最新的任务快照：轮询对账时读取，避免把 tasks 作为回调依赖反复重建
  const tasksRef = useRef(tasks);
  useEffect(() => {
    tasksRef.current = tasks;
  }, [tasks]);

  // Update or add a task
  const updateTask = useCallback((taskId, taskData) => {
    setTasks((prev) => {
      const newTasks = new Map(prev);
      const existing = newTasks.get(taskId) || {};
      newTasks.set(taskId, {
        ...existing,
        ...taskData,
        // Preserve existing startTime unless new one is provided
        startTime: taskData.startTime || existing.startTime,
        lastUpdate: Date.now()
      });
      return newTasks;
    });
  }, []);

  // Remove a task
  const removeTask = useCallback((taskId) => {
    setTasks((prev) => {
      const newTasks = new Map(prev);
      newTasks.delete(taskId);
      return newTasks;
    });
    // Also remove from stopping set
    setStoppingTasks((prev) => {
      const newSet = new Set(prev);
      newSet.delete(taskId);
      return newSet;
    });
  }, []);

  // Clear all tasks
  const clearAll = useCallback(() => {
    setTasks(new Map());
    setStoppingTasks(new Set());
  }, []);

  // 与后端对账：SSE 只推增量，刷新页面或断线会丢事件，需用 /tasks/running 兜底
  const syncRunningTasks = useCallback(async () => {
    // 未登录（登录页）时不需要对账，避免无意义的 401
    if (!localStorage.getItem('accessToken')) return;
    try {
      const res = await getRunningTasks();
      if (!res || (res.code !== 200 && res.code !== 0)) return;
      const serverTasks = res.data || [];
      const serverIds = new Set(serverTasks.map((task) => task.id));

      // 1. 服务端仍在执行的任务补进本地，刷新页面后也能恢复进度
      serverTasks.forEach((task) => {
        if (!task?.id) return;
        updateTask(task.id, {
          taskId: task.id,
          taskType: task.type,
          taskName: task.name,
          status: task.status || 'running',
          current: task.progress || 0,
          total: task.total || 0,
          currentItem: task.currentItem || '',
          message: task.message || '',
          startTime: task.startedAt ? new Date(task.startedAt).getTime() : undefined
        });
      });

      // 2. 本地仍是进行中但服务端已不存在：拉一次最终状态，避免幽灵进度条
      const staleTasks = Array.from(tasksRef.current.values()).filter(
        (task) => task?.taskId && !serverIds.has(task.taskId) && !TERMINAL_STATUSES.includes(task.status)
      );
      for (const task of staleTasks) {
        let removed = false;
        try {
          const detail = await getTask(task.taskId);
          const finalTask = detail?.data;
          if (finalTask && TERMINAL_STATUSES.includes(finalTask.status)) {
            updateTask(task.taskId, {
              status: finalTask.status,
              current: finalTask.progress ?? task.current,
              total: finalTask.total ?? task.total,
              message: finalTask.message || '',
              result: finalTask.result
            });
          } else {
            removeTask(task.taskId);
            removed = true;
          }
        } catch {
          removeTask(task.taskId);
          removed = true;
        }
        if (!removed) {
          setTimeout(() => removeTask(task.taskId), COMPLETED_TASK_CLEANUP_DELAY);
        }
      }
    } catch (error) {
      console.error('Sync running tasks failed:', error);
    }
  }, [updateTask, removeTask]);

  // Stop a task
  const stopTask = useCallback(
    async (taskId) => {
      // Mark as stopping
      setStoppingTasks((prev) => new Set(prev).add(taskId));

      try {
        await stopTaskApi(taskId);
        // The actual status update will come via SSE
        // Update local state optimistically
        updateTask(taskId, { status: 'cancelling' });
      } catch (error) {
        console.error('Failed to stop task:', error);
        // Remove from stopping set on error
        setStoppingTasks((prev) => {
          const newSet = new Set(prev);
          newSet.delete(taskId);
          return newSet;
        });
        throw error;
      }
    },
    [updateTask]
  );

  // Check if a task is being stopped
  const isTaskStopping = useCallback(
    (taskId) => {
      return stoppingTasks.has(taskId);
    },
    [stoppingTasks]
  );

  // Handle incoming progress event from SSE
  const handleProgressEvent = useCallback(
    (data) => {
      const { taskId, taskType, taskName, status, current, total, currentItem, result, message, startTime, traffic } = data;

      if (!taskId) return;

      if (status === 'completed' || status === 'error' || status === 'cancelled') {
        // Update task with completed/cancelled status, then schedule removal
        updateTask(taskId, {
          taskId,
          taskType,
          taskName,
          status,
          current,
          total,
          currentItem,
          result,
          message,
          traffic
        });

        // Notify all registered callbacks that a task has completed
        onCompleteCallbacksRef.current.forEach((callback) => {
          try {
            callback({ taskId, taskType, taskName, status, result });
          } catch (e) {
            console.error('TaskProgressContext onComplete callback error:', e);
          }
        });

        // Auto-remove completed tasks after delay
        setTimeout(() => {
          removeTask(taskId);
        }, COMPLETED_TASK_CLEANUP_DELAY);
      } else {
        // Update task progress
        updateTask(taskId, {
          taskId,
          taskType,
          taskName,
          status,
          current,
          total,
          currentItem,
          result,
          message,
          // Always pass startTime from event (every progress event now includes it)
          startTime,
          traffic
        });
      }
    },
    [updateTask, removeTask]
  );

  // Listen for task_progress CustomEvent from window (dispatched by AuthContext)
  useEffect(() => {
    const handleWindowEvent = (event) => {
      handleProgressEvent(event.detail);
    };
    window.addEventListener('task_progress', handleWindowEvent);
    return () => {
      window.removeEventListener('task_progress', handleWindowEvent);
    };
  }, [handleProgressEvent]);

  // 首次挂载与 SSE 重连后都对账一次：刷新页面/断线期间的任务状态不会丢
  useEffect(() => {
    syncRunningTasks();
    const handleReconnect = () => {
      syncRunningTasks();
    };
    window.addEventListener('sse_reconnected', handleReconnect);
    return () => {
      window.removeEventListener('sse_reconnected', handleReconnect);
    };
  }, [syncRunningTasks]);

  // Convert tasks Map to array for easier rendering
  const taskList = useMemo(() => Array.from(tasks.values()), [tasks]);

  // Check if there are any active tasks
  const hasActiveTasks = useMemo(() => taskList.length > 0, [taskList]);

  // 有任务在跑时才轮询，SSE 丢事件时也能收尾
  // 必须放在 hasActiveTasks 声明之后，否则依赖数组在渲染期求值会触发 TDZ 报错
  useEffect(() => {
    if (!hasActiveTasks) return undefined;
    const timer = setInterval(() => {
      syncRunningTasks();
    }, RUNNING_TASK_SYNC_INTERVAL);
    return () => clearInterval(timer);
  }, [hasActiveTasks, syncRunningTasks]);

  // Register a callback to be invoked when any task completes
  const registerOnComplete = useCallback((callback) => {
    onCompleteCallbacksRef.current.add(callback);
  }, []);

  // Unregister a previously registered callback
  const unregisterOnComplete = useCallback((callback) => {
    onCompleteCallbacksRef.current.delete(callback);
  }, []);

  const value = useMemo(
    () => ({
      tasks,
      taskList,
      hasActiveTasks,
      updateTask,
      removeTask,
      clearAll,
      handleProgressEvent,
      registerOnComplete,
      unregisterOnComplete,
      stopTask,
      isTaskStopping,
      stoppingTasks,
      syncRunningTasks
    }),
    [
      tasks,
      taskList,
      hasActiveTasks,
      updateTask,
      removeTask,
      clearAll,
      handleProgressEvent,
      registerOnComplete,
      unregisterOnComplete,
      stopTask,
      isTaskStopping,
      stoppingTasks,
      syncRunningTasks
    ]
  );

  return <TaskProgressContext.Provider value={value}>{children}</TaskProgressContext.Provider>;
}

TaskProgressProvider.propTypes = { children: PropTypes.node };

// ==============================|| useTaskProgress Hook ||============================== //

export function useTaskProgress() {
  const context = useContext(TaskProgressContext);
  if (!context) {
    throw new Error('useTaskProgress must be used within a TaskProgressProvider');
  }
  return context;
}

export default TaskProgressContext;
