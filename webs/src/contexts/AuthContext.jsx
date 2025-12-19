import PropTypes from 'prop-types';
import { createContext, useContext, useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';

// API imports
import { login as loginApi, logout as logoutApi, getUserInfo, rememberLogin as rememberLoginApi } from 'api/auth';

// Remember Token 的 localStorage key
const REMEMBER_TOKEN_KEY = 'sublink_remember_token';

// ==============================|| AUTH CONTEXT ||============================== //

const AuthContext = createContext(null);

// SSE 连接管理
let eventSource = null;
let heartbeatTimeout = null;
let reconnectTimeout = null;

// ==============================|| AUTH PROVIDER ||============================== //

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isInitialized, setIsInitialized] = useState(false);
  // 从 localStorage 初始化通知，处理 Date 反序列化
  const [notifications, setNotifications] = useState(() => {
    try {
      const saved = localStorage.getItem('app_notifications');
      if (saved) {
        const parsed = JSON.parse(saved);
        // 恢复 timestamp 为 Date 对象
        return parsed.map((n) => ({
          ...n,
          timestamp: n.timestamp ? new Date(n.timestamp) : new Date()
        }));
      }
    } catch (e) {
      console.error('Failed to parse saved notifications:', e);
    }
    return [];
  });
  const navigate = useNavigate();

  // 重置心跳计时器
  const resetHeartbeat = useCallback(() => {
    if (heartbeatTimeout) clearTimeout(heartbeatTimeout);
    heartbeatTimeout = setTimeout(() => {
      console.warn('SSE 心跳超时，正在重连...');
      if (eventSource) {
        eventSource.close();
        eventSource = null;
      }
      connectSSE();
    }, 15000); // 15s 超时 (后端每10s发送心跳)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // SSE 连接
  const connectSSE = useCallback(() => {
    if (eventSource?.readyState === 1) return; // 已连接

    const token = localStorage.getItem('accessToken');
    if (!token) return;

    const tokenStr = token.replace('Bearer ', '');
    const url = `/api/sse?token=${tokenStr}`;

    if (eventSource) {
      eventSource.close();
    }

    eventSource = new EventSource(url);

    eventSource.onopen = () => {
      console.log('SSE 已连接');
      resetHeartbeat();
    };

    eventSource.addEventListener('heartbeat', () => {
      console.log('SSE 心跳收到');
      resetHeartbeat();
    });

    eventSource.addEventListener('task_update', (event) => {
      resetHeartbeat();
      try {
        const data = JSON.parse(event.data);
        const status = data.status || data.data?.status;
        const notification = {
          id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
          type: status === 'success' ? 'success' : 'error',
          title: data.title || (status === 'success' ? '成功' : '失败'),
          message: data.message,
          timestamp: new Date()
        };
        setNotifications((prev) => [notification, ...prev].slice(0, 50));
      } catch (e) {
        console.error('解析 SSE 消息失败', e);
      }
    });

    eventSource.addEventListener('sub_update', (event) => {
      resetHeartbeat();
      try {
        const data = JSON.parse(event.data);
        const status = data.status || data.data?.status;
        const notification = {
          id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
          type: status === 'success' ? 'success' : 'error',
          title: data.title || (status === 'success' ? '订阅更新成功' : '订阅更新失败'),
          message: data.message,
          timestamp: new Date()
        };
        setNotifications((prev) => [notification, ...prev].slice(0, 50));
      } catch (e) {
        console.error('解析 SSE sub_update 消息失败', e);
      }
    });

    // 监听任务进度事件 (用于实时进度显示，不产生通知)
    eventSource.addEventListener('task_progress', (event) => {
      resetHeartbeat();
      try {
        const data = JSON.parse(event.data);
        // 使用 CustomEvent 分发进度数据，让 TaskProgressContext 接收
        window.dispatchEvent(new CustomEvent('task_progress', { detail: data }));
      } catch (e) {
        console.error('解析 SSE task_progress 消息失败', e);
      }
    });

    // 监听通用消息
    eventSource.onmessage = (event) => {
      resetHeartbeat();
      try {
        const data = JSON.parse(event.data);
        if (data.type === 'heartbeat' || data.type === 'ping') return;

        const notification = {
          id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
          type: data.type || 'info',
          title: data.title || '通知',
          message: data.message || JSON.stringify(data),
          timestamp: new Date()
        };
        setNotifications((prev) => [notification, ...prev].slice(0, 50));
      } catch (e) {
        console.error(e);
        // 忽略非JSON格式的心跳或其他数据
      }
    };

    eventSource.onerror = (err) => {
      console.error('SSE 错误:', err);
      if (eventSource) {
        eventSource.close();
        eventSource = null;
      }
      if (reconnectTimeout) clearTimeout(reconnectTimeout);
      console.log('5秒后尝试重连 SSE...');
      reconnectTimeout = setTimeout(() => {
        connectSSE();
      }, 5000);
    };
  }, [resetHeartbeat]);

  // 断开 SSE
  const disconnectSSE = useCallback(() => {
    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }
    if (reconnectTimeout) clearTimeout(reconnectTimeout);
    if (heartbeatTimeout) clearTimeout(heartbeatTimeout);
  }, []);

  // 通知变化时保存到 localStorage
  useEffect(() => {
    localStorage.setItem('app_notifications', JSON.stringify(notifications));
  }, [notifications]);

  // 初始化 - 检查 token 并获取用户信息
  useEffect(() => {
    const initAuth = async () => {
      const token = localStorage.getItem('accessToken');
      if (token) {
        try {
          const response = await getUserInfo();
          setUser(response.data);
          setIsAuthenticated(true);
          connectSSE();
        } catch (error) {
          console.error('获取用户信息失败:', error);
          localStorage.removeItem('accessToken');
          setIsAuthenticated(false);

          // 尝试使用记住密码令牌自动登录
          await tryRememberLogin();
        }
      } else {
        // 没有 accessToken，尝试使用记住密码令牌登录
        await tryRememberLogin();
      }
      setIsInitialized(true);
    };

    initAuth();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [connectSSE]);

  // 尝试使用记住密码令牌自动登录
  const tryRememberLogin = async () => {
    const rememberToken = localStorage.getItem(REMEMBER_TOKEN_KEY);
    if (!rememberToken) return false;

    try {
      const response = await rememberLoginApi(rememberToken);
      const { tokenType, accessToken } = response.data;
      localStorage.setItem('accessToken', `${tokenType} ${accessToken}`);

      const userResponse = await getUserInfo();
      setUser(userResponse.data);
      setIsAuthenticated(true);
      connectSSE();
      return true;
    } catch (error) {
      console.error('记住密码自动登录失败:', error);
      // 令牌无效，清除它
      localStorage.removeItem(REMEMBER_TOKEN_KEY);
      return false;
    }
  };

  // 登录 - 支持验证码和记住密码
  const login = async (username, password, captchaKey, captchaCode, rememberMe = false, turnstileToken = '') => {
    try {
      const response = await loginApi({ username, password, captchaKey, captchaCode, rememberMe, turnstileToken });

      // 登录成功（code === 200，否则会被 request.js 拦截器 reject）
      const { tokenType, accessToken, rememberToken } = response.data;
      localStorage.setItem('accessToken', `${tokenType} ${accessToken}`);

      // 如果返回了 rememberToken，保存它
      if (rememberToken) {
        localStorage.setItem(REMEMBER_TOKEN_KEY, rememberToken);
      } else {
        // 没有勾选记住密码，清除旧的令牌
        localStorage.removeItem(REMEMBER_TOKEN_KEY);
      }

      // 获取用户信息
      const userResponse = await getUserInfo();
      setUser(userResponse.data);
      setIsAuthenticated(true);
      connectSSE();

      return { success: true };
    } catch (error) {
      console.error('登录失败:', error);
      // 业务错误通过 error.message (来自后端 msg) 和 error.data 获取
      return {
        success: false,
        message: error.message || '登录失败，请检查用户名、密码和验证码',
        errorType: error.data?.data?.errorType || null
      };
    }
  };

  // 登出
  const logout = async () => {
    // 获取当前的 rememberToken，用于后端删除
    const rememberToken = localStorage.getItem(REMEMBER_TOKEN_KEY);
    try {
      await logoutApi(rememberToken);
    } catch (error) {
      console.error('登出API调用失败:', error);
    } finally {
      localStorage.removeItem('accessToken');
      localStorage.removeItem(REMEMBER_TOKEN_KEY); // 登出时清除记住密码令牌
      setUser(null);
      setIsAuthenticated(false);
      disconnectSSE();
      navigate('/login');
    }
  };

  // 清除通知
  const clearNotification = (id) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id));
  };

  const clearAllNotifications = () => {
    setNotifications([]);
  };

  const value = useMemo(
    () => ({
      user,
      isAuthenticated,
      isInitialized,
      notifications,
      login,
      logout,
      clearNotification,
      clearAllNotifications
    }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [user, isAuthenticated, isInitialized, notifications]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

AuthProvider.propTypes = { children: PropTypes.node };

// ==============================|| useAuth Hook ||============================== //

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth 必须在 AuthProvider 内部使用');
  }
  return context;
}

export default AuthContext;
