import { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

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
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

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
  const { t } = useTranslation();
  const { login, verifyMfa, rememberedUsernameKey } = useAuth();
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
  const [mfaCode, setMfaCode] = useState('');
  const [recoveryCode, setRecoveryCode] = useState('');
  const [useRecoveryCode, setUseRecoveryCode] = useState(false);
  const [mfaChallenge, setMfaChallenge] = useState(null);

  // 验证码配置状态
  const [captchaMode, setCaptchaMode] = useState(CAPTCHA_MODE.TRADITIONAL);
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('');
  const [captchaDegraded, setCaptchaDegraded] = useState(false);

  // Turnstile 弹窗状态
  const [turnstileDialogOpen, setTurnstileDialogOpen] = useState(false);
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

  useEffect(() => {
    try {
      const rememberedUsername = localStorage.getItem(rememberedUsernameKey);
      if (!rememberedUsername) {
        return;
      }

      setUsername(rememberedUsername);
      setRememberMe(true);
    } catch (err) {
      console.error('读取记住的用户名失败:', err);
      localStorage.removeItem(rememberedUsernameKey);
    }
  }, [rememberedUsernameKey]);

  const handleClickShowPassword = () => {
    setShowPassword(!showPassword);
  };

  const handleMouseDownPassword = (event) => {
    event.preventDefault();
  };

  const resetMfaState = useCallback(() => {
    setMfaChallenge(null);
    setMfaCode('');
    setRecoveryCode('');
    setUseRecoveryCode(false);
  }, []);

  const challengeMethods = mfaChallenge?.availableMethods || [];
  const canUseRecoveryCode = mfaChallenge?.recoveryAvailable || challengeMethods.includes('recovery_code');
  const maskedAccountHint = mfaChallenge?.hint || username;

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
        resetMfaState();
        navigate('/dashboard/default');
      } else if (result.mfaRequired && result.challenge) {
        setMfaChallenge(result.challenge);
        setUseRecoveryCode(false);
        setMfaCode('');
        setRecoveryCode('');
        setError('');
      } else {
        setError(result.message || t('auth.login.loginFailed'));
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
      setError(t('auth.login.loginFailed'));
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
  };

  const handleMfaSubmit = async (event) => {
    event.preventDefault();
    setError('');

    if (!mfaChallenge?.challengeToken) {
      setError(t('auth.mfa.missingContext'));
      return;
    }

    if (useRecoveryCode) {
      if (!recoveryCode.trim()) {
        setError(t('auth.mfa.missingRecovery'));
        return;
      }
    } else if (!mfaCode.trim()) {
      setError(t('auth.mfa.missingCode'));
      return;
    }

    setLoading(true);
    try {
      const result = await verifyMfa({
        challengeToken: mfaChallenge.challengeToken,
        code: useRecoveryCode ? '' : mfaCode.trim(),
        recoveryCode: useRecoveryCode ? recoveryCode.trim() : '',
        type: useRecoveryCode ? 'recovery_code' : 'totp',
        rememberMe,
        username
      });

      if (result.success) {
        resetMfaState();
        navigate('/dashboard/default');
      } else {
        setError(result.message || t('auth.mfa.verifyFailed'));
        if (useRecoveryCode) {
          setRecoveryCode('');
        } else {
          setMfaCode('');
        }
      }
    } catch {
      setError(t('auth.mfa.verifyFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleBackToCredentials = () => {
    resetMfaState();
    setError('');
    if (captchaMode === CAPTCHA_MODE.TRADITIONAL) {
      fetchCaptcha();
      setCaptchaCode('');
    }
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');

    // 表单验证
    if (!username.trim()) {
      setError(t('auth.login.missingUsername'));
      return;
    }

    if (!password.trim()) {
      setError(t('auth.login.missingPassword'));
      return;
    }

    if (password.length < 6) {
      setError(t('auth.login.passwordMinLength'));
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
          setError(t('auth.login.missingCaptcha'));
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
        <InputLabel htmlFor="outlined-adornment-captcha-login">{t('auth.login.captcha')}</InputLabel>
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
                    alt={t('auth.login.captcha')}
                    sx={{
                      height: 40,
                      cursor: 'pointer',
                      borderRadius: 1
                    }}
                    onClick={fetchCaptcha}
                  />
                )}
                <IconButton onClick={fetchCaptcha} size="small" title={t('auth.login.refreshCaptcha')}>
                  <RefreshIcon />
                </IconButton>
              </Stack>
            </InputAdornment>
          }
          label={t('auth.login.captcha')}
        />
      </CustomFormControl>
    );
  };

  const renderMfaForm = () => (
    <form onSubmit={handleMfaSubmit}>
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <Alert severity="info" sx={{ mb: 2 }}>
        <Stack spacing={0.5}>
          <Typography variant="body2">{t('auth.mfa.info')}</Typography>
          {maskedAccountHint && (
            <Typography variant="caption" color="text.secondary">
              {t('auth.mfa.currentAccount', { account: maskedAccountHint })}
            </Typography>
          )}
        </Stack>
      </Alert>

      <Stack spacing={2}>
        <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
          <Chip
            color={!useRecoveryCode ? 'secondary' : 'default'}
            label={t('auth.mfa.authenticator')}
            size="small"
            variant={!useRecoveryCode ? 'filled' : 'outlined'}
          />
          {canUseRecoveryCode && (
            <Chip
              color={useRecoveryCode ? 'secondary' : 'default'}
              label={t('auth.mfa.recoveryCode')}
              size="small"
              variant={useRecoveryCode ? 'filled' : 'outlined'}
            />
          )}
        </Stack>

        {!useRecoveryCode ? (
          <CustomFormControl fullWidth>
            <InputLabel htmlFor="outlined-adornment-totp-code">{t('auth.mfa.codeLabel')}</InputLabel>
            <OutlinedInput
              id="outlined-adornment-totp-code"
              type="text"
              value={mfaCode}
              onChange={(e) => setMfaCode(e.target.value.replace(/\s+/g, '').slice(0, 8))}
              name="totpCode"
              label={t('auth.mfa.codeLabel')}
              autoComplete="one-time-code"
              autoFocus
              inputProps={{ inputMode: 'numeric', pattern: '[0-9]*', 'aria-describedby': 'mfa-code-helper-text' }}
            />
          </CustomFormControl>
        ) : (
          <CustomFormControl fullWidth>
            <InputLabel htmlFor="outlined-adornment-recovery-code">{t('auth.mfa.recoveryCodeLabel')}</InputLabel>
            <OutlinedInput
              id="outlined-adornment-recovery-code"
              type="text"
              value={recoveryCode}
              onChange={(e) => setRecoveryCode(e.target.value.trimStart())}
              name="recoveryCode"
              label={t('auth.mfa.recoveryCodeLabel')}
              autoComplete="one-time-code"
              autoFocus
            />
          </CustomFormControl>
        )}

        <Typography id="mfa-code-helper-text" variant="caption" color="text.secondary" sx={{ mt: -1 }}>
          {useRecoveryCode ? t('auth.mfa.recoveryCodeHint') : t('auth.mfa.authenticatorHint')}
        </Typography>

        {canUseRecoveryCode && (
          <Button
            color="secondary"
            variant="text"
            onClick={() => setUseRecoveryCode((prev) => !prev)}
            sx={{ alignSelf: 'flex-start', px: 0.5 }}
          >
            {useRecoveryCode ? t('auth.mfa.useAuthenticatorBtn') : t('auth.mfa.useRecoveryBtn')}
          </Button>
        )}

        <Divider />

        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
          <Button variant="outlined" onClick={handleBackToCredentials} disabled={loading}>
            {t('auth.mfa.backBtn')}
          </Button>
          <AnimateButton>
            <Button color="secondary" fullWidth size="large" type="submit" variant="contained" disabled={loading}>
              <Box component="span" sx={{ display: 'inline-flex', alignItems: 'center' }}>
                <Box
                  component="span"
                  sx={{
                    width: 20,
                    height: 20,
                    mr: 1,
                    display: 'inline-flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    visibility: loading ? 'visible' : 'hidden'
                  }}
                >
                  <CircularProgress size={20} color="inherit" />
                </Box>
                <Box component="span">{loading ? t('auth.mfa.verifying') : t('auth.mfa.verifyBtn')}</Box>
              </Box>
            </Button>
          </AnimateButton>
        </Stack>
      </Stack>
    </form>
  );

  const renderPrimaryLoginForm = () => (
    <form onSubmit={handleSubmit}>
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      {captchaDegraded && (
        <Alert severity="info" sx={{ mb: 2 }}>
          {t('auth.login.turnstileDegraded')}
        </Alert>
      )}

      <CustomFormControl fullWidth>
        <InputLabel htmlFor="outlined-adornment-username-login">{t('auth.login.username')}</InputLabel>
        <OutlinedInput
          id="outlined-adornment-username-login"
          type="text"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          name="username"
          label={t('auth.login.username')}
          autoComplete="username"
          autoFocus
        />
      </CustomFormControl>

      <CustomFormControl fullWidth>
        <InputLabel htmlFor="outlined-adornment-password-login">{t('auth.login.password')}</InputLabel>
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
                aria-label={t('auth.login.togglePassword')}
                onClick={handleClickShowPassword}
                onMouseDown={handleMouseDownPassword}
                edge="end"
                size="large"
              >
                {showPassword ? <Visibility /> : <VisibilityOff />}
              </IconButton>
            </InputAdornment>
          }
          label={t('auth.login.password')}
        />
      </CustomFormControl>

      {renderCaptcha()}

      <FormControlLabel
        control={<Checkbox checked={rememberMe} onChange={(e) => setRememberMe(e.target.checked)} name="rememberMe" color="secondary" />}
        label={t('auth.login.rememberMe')}
        sx={{ mt: 1, mb: 1, ml: 0 }}
      />

      <Box sx={{ mt: 1 }}>
        <AnimateButton>
          <Button color="secondary" fullWidth size="large" type="submit" variant="contained" disabled={loading}>
            <Box component="span" sx={{ display: 'inline-flex', alignItems: 'center' }}>
              <Box
                component="span"
                sx={{
                  width: 20,
                  height: 20,
                  mr: 1,
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  visibility: loading ? 'visible' : 'hidden'
                }}
              >
                <CircularProgress size={20} color="inherit" />
              </Box>
              <Box component="span">{loading ? t('auth.login.loggingIn') : t('auth.login.loginBtn')}</Box>
            </Box>
          </Button>
        </AnimateButton>
      </Box>
    </form>
  );

  return (
    <>
      {mfaChallenge ? renderMfaForm() : renderPrimaryLoginForm()}

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
