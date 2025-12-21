import { useState, useEffect, useCallback } from 'react';
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
import TurnstileWidget from 'ui-component/TurnstileWidget';
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
  const [turnstileToken, setTurnstileToken] = useState('');
  const [captchaDegraded, setCaptchaDegraded] = useState(false);

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
        setTurnstileToken(''); // 重置 token
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

  // Turnstile 验证回调
  const handleTurnstileVerify = (token) => {
    setTurnstileToken(token);
    setError(''); // 清除错误
  };

  const handleTurnstileError = () => {
    setError('人机验证加载失败，请刷新页面重试');
  };

  const handleTurnstileExpire = () => {
    setTurnstileToken('');
    setError('人机验证已过期，请重新验证');
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');

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

    // 根据验证码模式验证
    if (captchaMode === CAPTCHA_MODE.TRADITIONAL && !captchaCode.trim()) {
      setError('请输入验证码');
      return;
    }

    if (captchaMode === CAPTCHA_MODE.TURNSTILE && !turnstileToken) {
      setError('请完成人机验证');
      return;
    }

    setLoading(true);

    try {
      // 根据模式传递不同的验证码数据
      const result = await login(
        username,
        password,
        captchaMode === CAPTCHA_MODE.TRADITIONAL ? captchaKey : '',
        captchaMode === CAPTCHA_MODE.TRADITIONAL ? captchaCode : '',
        rememberMe,
        captchaMode === CAPTCHA_MODE.TURNSTILE ? turnstileToken : ''
      );

      if (result.success) {
        navigate('/dashboard/default');
      } else {
        setError(result.message || '登录失败');
        // 登录失败时刷新验证码
        fetchCaptcha();
        setCaptchaCode('');
        setTurnstileToken('');
      }
    } catch {
      setError('登录失败，请稍后重试');
      fetchCaptcha();
      setCaptchaCode('');
      setTurnstileToken('');
    } finally {
      setLoading(false);
    }
  };

  // 渲染验证码区域
  const renderCaptcha = () => {
    switch (captchaMode) {
      case CAPTCHA_MODE.DISABLED:
        // 验证码已关闭，不显示任何内容
        return null;

      case CAPTCHA_MODE.TURNSTILE:
        // Cloudflare Turnstile
        return (
          <Box sx={{ my: 1 }}>
            <TurnstileWidget
              siteKey={turnstileSiteKey}
              onVerify={handleTurnstileVerify}
              onError={handleTurnstileError}
              onExpire={handleTurnstileExpire}
            />
          </Box>
        );

      default:
        // 传统图形验证码
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
    }
  };

  return (
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

      {/* 验证码区域 */}
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
  );
}
