import PropTypes from 'prop-types';
import { createContext, useContext, useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';

// API imports
import { login as loginApi, logout as logoutApi, getUserInfo, verifyMfaLogin } from 'api/auth';

const REMEMBERED_USERNAME_KEY = 'sublink_remembered_username';

// ==============================|| AUTH CONTEXT ||============================== //

const AuthContext = createContext(null);

// SSE 连接管理
let eventSource = null;
let heartbeatTimeout = null;
let reconnectTimeout = null;

function getPayloadData(payload) {
  return payload?.data || payload || {};
}

function normalizeTokenPayload(payload) {
  const data = getPayloadData(payload);
  const tokenType = data.tokenType || data.token_type || data.authType || data.auth_type;
  const accessToken = data.accessToken || data.access_token || data.token;

  if (!accessToken) {
    return null;
  }

  return {
    tokenType: tokenType || 'Bearer',
    accessToken
  };
}

function normalizeMfaChallenge(payload) {
  const data = getPayloadData(payload);
  const nestedMfa = data.mfa || data.challenge || data.auth || {};
  const challengeData = nestedMfa.challenge || nestedMfa;
  const errorType = data.errorType || challengeData.errorType || nestedMfa.errorType || '';
  const methods = data.methods || challengeData.methods || nestedMfa.methods || [];
  const challengeToken =
    data.challengeToken ||
    data.mfaToken ||
    data.ticket ||
    data.sessionToken ||
    challengeData.challengeToken ||
    challengeData.mfaToken ||
    challengeData.ticket ||
    challengeData.sessionToken ||
    nestedMfa.challengeToken ||
    nestedMfa.mfaToken ||
    nestedMfa.ticket ||
    nestedMfa.sessionToken ||
    '';

  const isMfaRequired = Boolean(
    data.mfaRequired ||
      data.requiresMfa ||
      nestedMfa.required ||
      nestedMfa.mfaRequired ||
      challengeData.required ||
      challengeToken ||
      methods.length > 0 ||
      String(errorType).toUpperCase().includes('MFA') ||
      String(errorType).toUpperCase().includes('TOTP')
  );

  if (!isMfaRequired) {
    return null;
  }

  return {
    challengeToken,
    challengeType: data.challengeType || challengeData.challengeType || nestedMfa.challengeType || 'totp',
    methods,
    availableMethods: methods,
    hint:
      data.challengeHint ||
      data.maskedAccount ||
      challengeData.challengeHint ||
      challengeData.maskedAccount ||
      nestedMfa.challengeHint ||
      nestedMfa.maskedAccount ||
      '',
    recoveryAvailable: Boolean(
      data.recoveryAvailable ??
        data.recoveryCodesAvailable ??
        challengeData.recoveryAvailable ??
        challengeData.recoveryCodesAvailable ??
        nestedMfa.recoveryAvailable ??
        nestedMfa.recoveryCodesAvailable ??
        methods.includes('recovery_code')
    ),
    message: data.msg || data.message || nestedMfa.message || '需要进行二次验证'
  };
}

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

    const appendNotification = (data) => {
      const parsedTimestamp = data.time ? new Date(data.time.replace(' ', 'T')) : new Date();
      const timestamp = Number.isNaN(parsedTimestamp.getTime()) ? new Date() : parsedTimestamp;
      const notification = {
        id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
        type: data.severity || data.type || 'info',
        title: data.title || data.eventName || '通知',
        message: data.message || '',
        timestamp,
        eventKey: data.event || '',
        eventName: data.eventName || '',
        category: data.category || '',
        categoryName: data.categoryName || ''
      };
      setNotifications((prev) => [notification, ...prev].slice(0, 50));
    };

    eventSource.addEventListener('notification', (event) => {
      resetHeartbeat();
      try {
        appendNotification(JSON.parse(event.data));
      } catch (e) {
        console.error('解析 SSE notification 消息失败', e);
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
          timestamp: new Date(),
          eventKey: data.event || '',
          eventName: data.eventName || '',
          category: data.category || '',
          categoryName: data.categoryName || ''
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

  const finalizeAuth = useCallback(
    async ({ accessToken, tokenType, rememberMe = false, username = '' }) => {
      localStorage.setItem('accessToken', `${tokenType} ${accessToken}`);

      if (rememberMe) {
        localStorage.setItem(REMEMBERED_USERNAME_KEY, username);
      } else {
        localStorage.removeItem(REMEMBERED_USERNAME_KEY);
      }

      const userResponse = await getUserInfo();
      setUser(userResponse.data);
      setIsAuthenticated(true);
      connectSSE();
    },
    [connectSSE]
  );

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
        }
      }
      setIsInitialized(true);
    };

    initAuth();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [connectSSE]);

  // 登录 - 支持验证码和记住密码
  const login = async (username, password, captchaKey, captchaCode, rememberMe = false, turnstileToken = '') => {
    try {
      const response = await loginApi({ username, password, captchaKey, captchaCode, rememberMe, turnstileToken });

      const tokenPayload = normalizeTokenPayload(response);
      if (tokenPayload) {
        await finalizeAuth({ ...tokenPayload, rememberMe, username });

        return { success: true };
      }

      const challenge = normalizeMfaChallenge(response);
      if (challenge) {
        localStorage.removeItem('accessToken');
        disconnectSSE();
        setUser(null);
        setIsAuthenticated(false);

        return {
          success: false,
          mfaRequired: true,
          challenge,
          message: challenge.message
        };
      }

      return {
        success: false,
        message: '登录响应缺少访问令牌，请稍后重试'
      };
    } catch (error) {
      console.error('登录失败:', error);
      const challenge = normalizeMfaChallenge(error.data);
      if (challenge) {
        localStorage.removeItem('accessToken');
        disconnectSSE();
        setUser(null);
        setIsAuthenticated(false);

        return {
          success: false,
          mfaRequired: true,
          challenge,
          message: challenge.message
        };
      }

      // 业务错误通过 error.message (来自后端 msg) 和 error.data 获取
      return {
        success: false,
        message: error.message || '登录失败，请检查用户名、密码和验证码',
        errorType: error.data?.data?.errorType || null
      };
    }
  };

  const verifyMfa = async ({ challengeToken, code, recoveryCode, type = 'totp', rememberMe = false, username = '' }) => {
    try {
      const response = await verifyMfaLogin({
        challengeToken,
        code,
        recoveryCode,
        type
      });

      const tokenPayload = normalizeTokenPayload(response);
      if (!tokenPayload) {
        return {
          success: false,
          message: '验证成功，但未收到访问令牌'
        };
      }

      await finalizeAuth({ ...tokenPayload, rememberMe, username });
      return { success: true };
    } catch (error) {
      console.error('MFA 验证失败:', error);
      return {
        success: false,
        message: error.message || '验证失败，请检查验证码或恢复码后重试',
        errorType: error.data?.data?.errorType || null
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
      verifyMfa,
      logout,
      rememberedUsernameKey: REMEMBERED_USERNAME_KEY,
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
