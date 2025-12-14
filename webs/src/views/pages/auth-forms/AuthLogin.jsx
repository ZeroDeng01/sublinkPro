import { useState, useEffect } from 'react';
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
import { useAuth } from 'contexts/AuthContext';
import { getCaptcha } from 'api/auth';

// assets
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import RefreshIcon from '@mui/icons-material/Refresh';

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

  // 获取验证码
  const fetchCaptcha = async () => {
    try {
      const response = await getCaptcha();
      setCaptchaKey(response.data.captchaKey);
      setCaptchaBase64(response.data.captchaBase64);
    } catch (err) {
      console.error('获取验证码失败:', err);
    }
  };

  useEffect(() => {
    fetchCaptcha();
  }, []);

  const handleClickShowPassword = () => {
    setShowPassword(!showPassword);
  };

  const handleMouseDownPassword = (event) => {
    event.preventDefault();
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

    if (!captchaCode.trim()) {
      setError('请输入验证码');
      return;
    }

    setLoading(true);

    try {
      // 传递 rememberMe 参数给 login 函数，由后端生成安全令牌
      const result = await login(username, password, captchaKey, captchaCode, rememberMe);
      if (result.success) {
        navigate('/dashboard');
      } else {
        setError(result.message || '登录失败');
        // 登录失败时刷新验证码
        fetchCaptcha();
        setCaptchaCode('');
      }
    } catch {
      setError('登录失败，请稍后重试');
      fetchCaptcha();
      setCaptchaCode('');
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
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
