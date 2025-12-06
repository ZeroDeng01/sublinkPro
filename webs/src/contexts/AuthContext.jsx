import PropTypes from 'prop-types';
import { createContext, useContext, useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';

// API imports
import { login as loginApi, logout as logoutApi, getUserInfo } from 'api/auth';

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
  const [notifications, setNotifications] = useState([]);
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
          id: Date.now(),
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
          id: Date.now(),
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

    // 监听通用消息
    eventSource.onmessage = (event) => {
      resetHeartbeat();
      try {
        const data = JSON.parse(event.data);
        if (data.type === 'heartbeat' || data.type === 'ping') return;

        const notification = {
          id: Date.now(),
          type: data.type || 'info',
          title: data.title || '通知',
          message: data.message || JSON.stringify(data),
          timestamp: new Date()
        };
        setNotifications((prev) => [notification, ...prev].slice(0, 50));
      } catch (e) {
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
        }
      }
      setIsInitialized(true);
    };

    initAuth();
  }, [connectSSE]);

  // 登录 - 支持验证码
  const login = async (username, password, captchaKey, captchaCode) => {
    try {
      const response = await loginApi({ username, password, captchaKey, captchaCode });
      const { tokenType, accessToken } = response.data;
      localStorage.setItem('accessToken', `${tokenType} ${accessToken}`);

      // 获取用户信息
      const userResponse = await getUserInfo();
      setUser(userResponse.data);
      setIsAuthenticated(true);
      connectSSE();

      return { success: true };
    } catch (error) {
      console.error('登录失败:', error);
      return {
        success: false,
        message: error.response?.data?.message || error.response?.data?.msg || '登录失败，请检查用户名、密码和验证码'
      };
    }
  };

  // 登出
  const logout = async () => {
    try {
      await logoutApi();
    } catch (error) {
      console.error('登出API调用失败:', error);
    } finally {
      localStorage.removeItem('accessToken');
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
