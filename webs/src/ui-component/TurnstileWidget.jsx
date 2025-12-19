import { useEffect, useRef, useState, useCallback } from 'react';
import PropTypes from 'prop-types';
import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import Typography from '@mui/material/Typography';
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';

// Cloudflare Turnstile 脚本 URL
const TURNSTILE_SCRIPT_URL = 'https://challenges.cloudflare.com/turnstile/v0/api.js';

/**
 * Cloudflare Turnstile 组件
 */
export default function TurnstileWidget({ siteKey, onVerify, onError, onExpire }) {
  const widgetIdRef = useRef(null);
  const renderedRef = useRef(false);
  const [containerEl, setContainerEl] = useState(null);
  const [isVerified, setIsVerified] = useState(false);
  const [hasError, setHasError] = useState(false);
  const [scriptReady, setScriptReady] = useState(!!window.turnstile);

  // 使用 ref 存储回调
  const callbacksRef = useRef({ onVerify, onError, onExpire });
  callbacksRef.current = { onVerify, onError, onExpire };

  // 使用 callback ref 捕获容器元素
  const containerRefCallback = useCallback((node) => {
    if (node) {
      setContainerEl(node);
    }
  }, []);

  // 加载脚本
  useEffect(() => {
    if (!siteKey || window.turnstile) {
      if (window.turnstile) setScriptReady(true);
      return;
    }

    const existingScript = document.querySelector(`script[src="${TURNSTILE_SCRIPT_URL}"]`);
    if (existingScript) {
      const checkInterval = setInterval(() => {
        if (window.turnstile) {
          clearInterval(checkInterval);
          setScriptReady(true);
        }
      }, 50);
      setTimeout(() => clearInterval(checkInterval), 10000);
      return;
    }

    const script = document.createElement('script');
    script.src = TURNSTILE_SCRIPT_URL;
    script.async = true;

    script.onload = () => {
      const checkInterval = setInterval(() => {
        if (window.turnstile) {
          clearInterval(checkInterval);
          setScriptReady(true);
        }
      }, 50);
      setTimeout(() => {
        clearInterval(checkInterval);
        if (!window.turnstile) setHasError(true);
      }, 10000);
    };

    script.onerror = () => setHasError(true);
    document.head.appendChild(script);
  }, [siteKey]);

  // 渲染 widget（当脚本加载完成且容器存在时）
  useEffect(() => {
    if (!scriptReady || !containerEl || !siteKey || !window.turnstile || renderedRef.current) {
      return;
    }

    try {
      widgetIdRef.current = window.turnstile.render(containerEl, {
        sitekey: siteKey,
        callback: (token) => {
          setIsVerified(true);
          if (callbacksRef.current.onVerify) callbacksRef.current.onVerify(token);
        },
        'error-callback': () => {
          setHasError(true);
          if (callbacksRef.current.onError) callbacksRef.current.onError();
        },
        'expired-callback': () => {
          setIsVerified(false);
          renderedRef.current = false;
          widgetIdRef.current = null;
          if (callbacksRef.current.onExpire) callbacksRef.current.onExpire();
        },
        theme: 'auto',
        language: 'zh-CN'
      });
      renderedRef.current = true;
    } catch (err) {
      console.error('Turnstile 渲染失败:', err);
      setHasError(true);
    }

    return () => {
      if (widgetIdRef.current && window.turnstile) {
        try {
          window.turnstile.remove(widgetIdRef.current);
          widgetIdRef.current = null;
          renderedRef.current = false;
        } catch {
          // 忽略
        }
      }
    };
  }, [scriptReady, containerEl, siteKey]);

  // 验证通过
  if (isVerified) {
    return (
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', py: 1.5, gap: 0.5 }}>
        <CheckCircleOutlineIcon color="success" fontSize="small" />
        <Typography variant="body2" color="success.main" fontWeight={500}>
          人机验证已通过
        </Typography>
      </Box>
    );
  }

  // 错误状态
  if (hasError) {
    return (
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', py: 2 }}>
        <Typography variant="body2" color="error.main">
          人机验证加载失败，请刷新页面重试
        </Typography>
      </Box>
    );
  }

  // 显示加载中或 widget 容器
  return (
    <Box sx={{ my: 1, minHeight: 65, display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
      {!scriptReady ? (
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <CircularProgress size={20} />
          <Typography variant="body2" color="text.secondary">
            正在加载人机验证...
          </Typography>
        </Box>
      ) : (
        <Box ref={containerRefCallback} />
      )}
    </Box>
  );
}

TurnstileWidget.propTypes = {
  siteKey: PropTypes.string.isRequired,
  onVerify: PropTypes.func,
  onError: PropTypes.func,
  onExpire: PropTypes.func
};
