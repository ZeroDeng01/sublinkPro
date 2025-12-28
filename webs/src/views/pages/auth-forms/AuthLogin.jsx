import { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';

// material-ui
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import CircularProgress from '@mui/material/CircularProgress';
import FormControlLabel from '@mui/material/FormControlLabel';
import IconButton from '@mui/material/IconButton';
import InputAdornment from '@mui/material/InputAdornment';
import InputLabel from '@mui/material/InputLabel';
import OutlinedInput from '@mui/material/OutlinedInput';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';

// project imports
import AnimateButton from 'ui-component/extended/AnimateButton';
import CustomFormControl from 'ui-component/extended/Form/CustomFormControl';
import TurnstileDialog from 'ui-component/TurnstileDialog';
import { useAuth } from 'contexts/AuthContext';
import { getCaptcha } from 'api/auth';

// assets
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import RefreshIcon from '@mui/icons-material/Refresh';

// 验证码模式常量（与后端保持一致）
const CAPTCHA_MODE = {
  DISABLED: 1,
  TRADITIONAL: 2,
  TURNSTILE: 3
};

// ===============================|| 登录表单 ||=============================== //

export default function AuthLogin() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const turnstileDialogRef = useRef(null);

  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [captchaCode, setCaptchaCode] = useState('');
  const [captchaKey, setCaptchaKey] = useState('');
  const [captchaBase64, setCaptchaBase64] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [rememberMe, setRememberMe] = useState(false);

  // 验证码配置状态
  const [captchaMode, setCaptchaMode] = useState(CAPTCHA_MODE.TRADITIONAL);
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('');
  const [captchaDegraded, setCaptchaDegraded] = useState(false);

  // Turnstile 弹窗状态
  const [turnstileDialogOpen, setTurnstileDialogOpen] = useState(false);
  // 存储待提交的登录信息
  const pendingLoginRef = useRef(null);

  // 获取验证码配置
  const fetchCaptcha = useCallback(async () => {
    try {
      const response = await getCaptcha();
      const data = response.data;

      // 设置验证码模式
      setCaptchaMode(data.mode || CAPTCHA_MODE.TRADITIONAL);
      setCaptchaDegraded(data.degraded || false);

      // 根据模式设置相应数据
      if (data.mode === CAPTCHA_MODE.TRADITIONAL) {
        setCaptchaKey(data.captchaKey || '');
        setCaptchaBase64(data.captchaBase64 || '');
      } else if (data.mode === CAPTCHA_MODE.TURNSTILE) {
        setTurnstileSiteKey(data.turnstileSiteKey || '');
      }
    } catch (err) {
      console.error('获取验证码失败:', err);
      // 默认使用传统验证码
      setCaptchaMode(CAPTCHA_MODE.TRADITIONAL);
    }
  }, []);

  useEffect(() => {
    fetchCaptcha();
  }, [fetchCaptcha]);

  const handleClickShowPassword = () => {
    setShowPassword(!showPassword);
  };

  const handleMouseDownPassword = (event) => {
    event.preventDefault();
  };

  // 执行实际登录请求
  const performLogin = async (turnstileToken = '') => {
    setLoading(true);
    setError('');

    try {
      const result = await login(
        username,
        password,
        captchaMode === CAPTCHA_MODE.TRADITIONAL ? captchaKey : '',
        captchaMode === CAPTCHA_MODE.TRADITIONAL ? captchaCode : '',
        rememberMe,
        turnstileToken
      );

      if (result.success) {
        navigate('/dashboard/default');
      } else {
        setError(result.message || '登录失败');
        // 登录失败时刷新验证码或重置 Turnstile
        if (captchaMode === CAPTCHA_MODE.TRADITIONAL) {
          fetchCaptcha();
          setCaptchaCode('');
        } else if (captchaMode === CAPTCHA_MODE.TURNSTILE) {
          // 重置 Turnstile 弹窗，用户需要重新验证
          if (turnstileDialogRef.current) {
            turnstileDialogRef.current.reset();
          }
        }
      }
    } catch {
      setError('登录失败，请稍后重试');
      if (captchaMode === CAPTCHA_MODE.TRADITIONAL) {
        fetchCaptcha();
        setCaptchaCode('');
      }
    } finally {
      setLoading(false);
    }
  };

  // Turnstile 验证成功回调
  const handleTurnstileSuccess = (token) => {
    setTurnstileDialogOpen(false);
    // 验证成功，执行登录
    performLogin(token);
  };

  // Turnstile 弹窗关闭
  const handleTurnstileDialogClose = () => {
    setTurnstileDialogOpen(false);
    pendingLoginRef.current = null;
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');

    // 表单验证
    if (!username.trim()) {
      setError('请输入用户名');
      return;
    }

    if (!password.trim()) {
      setError('请输入密码');
      return;
    }

    if (password.length < 6) {
      setError('密码长度至少6位');
      return;
    }

    // 根据验证码模式处理
    switch (captchaMode) {
      case CAPTCHA_MODE.DISABLED:
        // 验证码已关闭，直接登录
        performLogin('');
        break;

      case CAPTCHA_MODE.TURNSTILE:
        // Turnstile 模式，打开验证弹窗
        setTurnstileDialogOpen(true);
        break;

      default:
        // 传统验证码模式
        if (!captchaCode.trim()) {
          setError('请输入验证码');
          return;
        }
        performLogin('');
        break;
    }
  };

  // 渲染验证码区域（仅传统模式）
  const renderCaptcha = () => {
    if (captchaMode !== CAPTCHA_MODE.TRADITIONAL) {
      return null;
    }

    return (
      <CustomFormControl fullWidth>
        <InputLabel htmlFor="outlined-adornment-captcha-login">验证码</InputLabel>
        <OutlinedInput
          id="outlined-adornment-captcha-login"
          type="text"
          value={captchaCode}
          onChange={(e) => setCaptchaCode(e.target.value)}
          name="captchaCode"
          autoComplete="off"
          onKeyDown={(e) => e.key === 'Enter' && handleSubmit(e)}
          endAdornment={
            <InputAdornment position="end">
              <Stack direction="row" alignItems="center" spacing={0.5}>
                {captchaBase64 && (
                  <Box
                    component="img"
                    src={captchaBase64}
                    alt="验证码"
                    sx={{
                      height: 40,
                      cursor: 'pointer',
                      borderRadius: 1
                    }}
                    onClick={fetchCaptcha}
                  />
                )}
                <IconButton onClick={fetchCaptcha} size="small" title="刷新验证码">
                  <RefreshIcon />
                </IconButton>
              </Stack>
            </InputAdornment>
          }
          label="验证码"
        />
      </CustomFormControl>
    );
  };

  return (
    <>
      <form onSubmit={handleSubmit}>
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        {/* 降级提示 */}
        {captchaDegraded && (
          <Alert severity="info" sx={{ mb: 2 }}>
            Turnstile 配置不完整，已降级为传统验证码
          </Alert>
        )}

        <CustomFormControl fullWidth>
          <InputLabel htmlFor="outlined-adornment-username-login">用户名</InputLabel>
          <OutlinedInput
            id="outlined-adornment-username-login"
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            name="username"
            label="用户名"
            autoComplete="username"
            autoFocus
          />
        </CustomFormControl>

        <CustomFormControl fullWidth>
          <InputLabel htmlFor="outlined-adornment-password-login">密码</InputLabel>
          <OutlinedInput
            id="outlined-adornment-password-login"
            type={showPassword ? 'text' : 'password'}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            name="password"
            autoComplete="current-password"
            endAdornment={
              <InputAdornment position="end">
                <IconButton
                  aria-label="切换密码可见性"
                  onClick={handleClickShowPassword}
                  onMouseDown={handleMouseDownPassword}
                  edge="end"
                  size="large"
                >
                  {showPassword ? <Visibility /> : <VisibilityOff />}
                </IconButton>
              </InputAdornment>
            }
            label="密码"
          />
        </CustomFormControl>

        {/* 验证码区域（仅传统模式） */}
        {renderCaptcha()}

        <FormControlLabel
          control={<Checkbox checked={rememberMe} onChange={(e) => setRememberMe(e.target.checked)} name="rememberMe" color="secondary" />}
          label="记住密码"
          sx={{ mt: 1, mb: 1, ml: 0 }}
        />

        <Box sx={{ mt: 1 }}>
          <AnimateButton>
            <Button
              color="secondary"
              fullWidth
              size="large"
              type="submit"
              variant="contained"
              disabled={loading}
              startIcon={loading ? <CircularProgress size={20} color="inherit" /> : null}
            >
              {loading ? '登录中...' : '登 录'}
            </Button>
          </AnimateButton>
        </Box>
      </form>

      {/* Turnstile 验证弹窗 */}
      {captchaMode === CAPTCHA_MODE.TURNSTILE && turnstileSiteKey && (
        <TurnstileDialog
          ref={turnstileDialogRef}
          open={turnstileDialogOpen}
          onClose={handleTurnstileDialogClose}
          onSuccess={handleTurnstileSuccess}
          siteKey={turnstileSiteKey}
        />
      )}
    </>
  );
}
