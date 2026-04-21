import { useState, useEffect } from 'react';

import Alert from '@mui/material/Alert';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Divider from '@mui/material/Divider';
import FormControlLabel from '@mui/material/FormControlLabel';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import InputAdornment from '@mui/material/InputAdornment';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import TextField from '@mui/material/TextField';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import useMediaQuery from '@mui/material/useMediaQuery';
import Autocomplete from '@mui/material/Autocomplete';
import { useTheme } from '@mui/material/styles';

import CachedIcon from '@mui/icons-material/Cached';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import LockIcon from '@mui/icons-material/Lock';
import PsychologyIcon from '@mui/icons-material/Psychology';
import PersonIcon from '@mui/icons-material/Person';
import SaveIcon from '@mui/icons-material/Save';
import ScienceIcon from '@mui/icons-material/Science';
import SecurityIcon from '@mui/icons-material/Security';
import SettingsSuggestIcon from '@mui/icons-material/SettingsSuggest';
import ShieldOutlinedIcon from '@mui/icons-material/ShieldOutlined';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';

import { useAuth } from 'contexts/AuthContext';
import { changePassword, getAISettings, listAIModels, testAISettings, updateAISettings, updateProfile } from 'api/user';
import { QRCodeSVG } from 'qrcode.react';
import { confirmTotpSetup, disableTotp, getTotpStatus, regenerateRecoveryCodes, setupTotp } from 'api/auth';

const dedupeModelOptions = (models = [], currentModel = '') => {
  const seen = new Set();
  const options = [];

  const appendOption = (value) => {
    const nextValue = typeof value === 'string' ? value.trim() : '';
    if (!nextValue || seen.has(nextValue)) {
      return;
    }
    seen.add(nextValue);
    options.push(nextValue);
  };

  models.forEach(appendOption);
  appendOption(currentModel);
  return options;
};

export default function ProfileSettings({ showMessage, loading, setLoading }) {
  const { user, logout } = useAuth();
  const theme = useTheme();
  const fullScreenDialog = useMediaQuery(theme.breakpoints.down('sm'));

  const [profileForm, setProfileForm] = useState({
    username: '',
    nickname: ''
  });

  const [passwordForm, setPasswordForm] = useState({
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
    code: ''
  });
  const [profilePassword, setProfilePassword] = useState('');
  const [profileCode, setProfileCode] = useState('');
  const [showOldPassword, setShowOldPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [settingsSection, setSettingsSection] = useState(0);
  const [passwordDialogOpen, setPasswordDialogOpen] = useState(false);
  const [totpStatus, setTotpStatus] = useState({ enabled: false, recoveryCodes: [] });
  const [totpEnrollment, setTotpEnrollment] = useState({
    loading: false,
    secret: '',
    provisioningUri: '',
    qrCodeData: '',
    manualEntryKey: '',
    recoveryCodes: []
  });
  const [totpCode, setTotpCode] = useState('');
  const [totpPassword, setTotpPassword] = useState('');
  const [totpReauthCode, setTotpReauthCode] = useState('');
  const [disableVerificationCode, setDisableVerificationCode] = useState('');
  const [disablePassword, setDisablePassword] = useState('');
  const [aiSettingsLoading, setAISettingsLoading] = useState(false);
  const [aiAction, setAIAction] = useState('');
  const [aiForm, setAIForm] = useState({
    enabled: false,
    baseUrl: '',
    model: '',
    apiKey: '',
    maskedKey: '',
    hasKey: false,
    configured: false,
    providerType: 'openai_compatible',
    temperature: 0.2,
    maxTokens: 1200,
    password: '',
    code: ''
  });
  const [aiHeadersText, setAIHeadersText] = useState('{}');
  const [aiTestResult, setAITestResult] = useState(null);
  const [aiModelOptions, setAIModelOptions] = useState([]);
  const [aiModelsFetched, setAIModelsFetched] = useState(false);

  useEffect(() => {
    if (user) {
      setProfileForm({
        username: user.username || '',
        nickname: user.nickname || ''
      });
    }
  }, [user]);

  useEffect(() => {
    fetchTotpStatus();
    fetchUserAISettings();
  }, []);

  const setAIField = (field, value) => {
    setAIForm((prev) => ({ ...prev, [field]: value }));

    if (field === 'baseUrl' || field === 'apiKey') {
      setAIModelsFetched(false);
    }
  };

  const fetchUserAISettings = async () => {
    setAISettingsLoading(true);
    try {
      const response = await getAISettings();
      const data = response.data || {};
      setAIForm((prev) => ({
        ...prev,
        enabled: Boolean(data.enabled),
        baseUrl: data.baseUrl || '',
        model: data.model || '',
        apiKey: '',
        maskedKey: data.maskedKey || '',
        hasKey: Boolean(data.hasKey),
        configured: Boolean(data.configured),
        providerType: data.providerType || 'openai_compatible',
        temperature: data.temperature ?? 0.2,
        maxTokens: data.maxTokens ?? 1200,
        password: '',
        code: ''
      }));
      setAIModelOptions((prev) => dedupeModelOptions(prev, data.model || ''));
      setAIHeadersText(data.extraHeaders && Object.keys(data.extraHeaders).length > 0 ? JSON.stringify(data.extraHeaders, null, 2) : '{}');
      setAITestResult(null);
    } catch (error) {
      console.error('获取 AI 设置失败:', error);
    } finally {
      setAISettingsLoading(false);
    }
  };

  const parseAIExtraHeaders = () => {
    const trimmed = aiHeadersText.trim();

    if (!trimmed) {
      return {};
    }

    let parsed;
    try {
      parsed = JSON.parse(trimmed);
    } catch {
      throw new Error('额外请求头必须为 JSON 对象，例如 {"HTTP-Referer":"https://example.com"}');
    }

    if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
      throw new Error('额外请求头必须为 JSON 对象');
    }

    return Object.entries(parsed).reduce((acc, [key, value]) => {
      const headerKey = key.trim();
      if (!headerKey) {
        return acc;
      }
      acc[headerKey] = value == null ? '' : String(value);
      return acc;
    }, {});
  };

  const buildAISettingsPayload = () => ({
    enabled: aiForm.enabled,
    baseUrl: aiForm.baseUrl.trim(),
    model: aiForm.model.trim(),
    apiKey: aiForm.apiKey.trim(),
    temperature: aiForm.temperature === '' ? 0.2 : Number(aiForm.temperature),
    maxTokens: aiForm.maxTokens === '' ? 0 : Number(aiForm.maxTokens),
    extraHeaders: parseAIExtraHeaders()
  });

  const handleFetchAIModels = async () => {
    if (!aiForm.baseUrl.trim() && !aiForm.configured) {
      showMessage('请先填写 AI Base URL', 'warning');
      return;
    }

    if (!aiForm.apiKey.trim() && !aiForm.hasKey) {
      showMessage('请先填写 API Key', 'warning');
      return;
    }

    let extraHeaders;
    try {
      extraHeaders = parseAIExtraHeaders();
    } catch (error) {
      showMessage(error.message, 'warning');
      return;
    }

    setAIAction('models');
    setLoading(true);
    try {
      const response = await listAIModels({
        baseUrl: aiForm.baseUrl.trim(),
        apiKey: aiForm.apiKey.trim(),
        extraHeaders
      });
      const models = Array.isArray(response.data?.models) ? response.data.models : [];
      setAIModelOptions((prev) => dedupeModelOptions([...prev, ...models], aiForm.model));
      setAIModelsFetched(true);
      if (models.length > 0) {
        showMessage(`已获取 ${models.length} 个模型选项`);
      } else {
        showMessage('接口未返回可用模型列表，可继续手动输入模型名称', 'warning');
      }
    } catch (error) {
      showMessage('获取模型列表失败: ' + (error.response?.data?.message || error.message), 'error');
    } finally {
      setLoading(false);
      setAIAction('');
    }
  };

  const resetTotpEnrollment = () => {
    setTotpEnrollment({
      loading: false,
      secret: '',
      provisioningUri: '',
      qrCodeData: '',
      manualEntryKey: '',
      recoveryCodes: []
    });
  };

  const resetPasswordForm = () => {
    setPasswordForm({ oldPassword: '', newPassword: '', confirmPassword: '', code: '' });
    setShowOldPassword(false);
    setShowNewPassword(false);
    setShowConfirmPassword(false);
  };

  const fetchTotpStatus = async () => {
    try {
      const response = await getTotpStatus();
      const data = response.data || {};
      setTotpStatus({
        enabled: Boolean(data.enabled || data.isEnabled),
        pendingEnrollment: Boolean(data.pendingEnrollment),
        recoveryCodes: data.recoveryCodes || [],
        recoveryCodesRemaining: data.recoveryCodesRemaining ?? 0,
        issuer: data.issuer || '',
        accountName: data.accountName || ''
      });
    } catch (error) {
      console.error('获取 TOTP 状态失败:', error);
    }
  };

  const handleCopy = async (value, label) => {
    if (!value) {
      return;
    }

    try {
      await navigator.clipboard.writeText(value);
      showMessage(`${label}已复制`);
    } catch {
      showMessage(`复制${label}失败，请手动复制`, 'warning');
    }
  };

  const startTotpSetup = async () => {
    if (!totpPassword.trim()) {
      showMessage('请输入当前密码以开始设置 TOTP', 'warning');
      return;
    }

    setLoading(true);
    try {
      const response = await setupTotp({
        password: totpPassword.trim(),
        code: totpStatus.enabled ? totpReauthCode.trim() : ''
      });
      const data = response.data || {};
      setTotpEnrollment({
        loading: false,
        secret: data.secret || '',
        provisioningUri: data.provisioningUri || data.provisioningURI || data.otpauthUrl || data.otpauthURL || '',
        qrCodeData: data.provisioningUri || data.provisioningURI || data.otpauthUrl || data.otpauthURL || '',
        manualEntryKey: data.secret || '',
        recoveryCodes: data.recoveryCodes || []
      });
      setTotpCode('');
      setTotpReauthCode('');
      showMessage('请使用身份验证器扫描二维码后输入验证码完成绑定');
    } catch (error) {
      showMessage('获取 TOTP 配置失败: ' + (error.response?.data?.message || error.message), 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleConfirmTotpSetup = async () => {
    if (!totpCode.trim()) {
      showMessage('请输入身份验证器中的 6 位验证码', 'warning');
      return;
    }

    setLoading(true);
    try {
      const response = await confirmTotpSetup({
        code: totpCode.trim()
      });
      const data = response.data || {};
      const recoveryCodes = totpEnrollment.recoveryCodes || [];

      setTotpStatus((prev) => ({
        ...prev,
        enabled: true,
        recoveryCodes,
        recoveryCodesRemaining: data.recoveryCodesRemaining ?? recoveryCodes.length,
        pendingEnrollment: false
      }));
      setTotpEnrollment((prev) => ({ ...prev, recoveryCodes }));
      setTotpCode('');
      showMessage('双重验证已启用，请立即妥善保存恢复码');
      fetchTotpStatus();
    } catch (error) {
      showMessage('启用 TOTP 失败: ' + (error.response?.data?.message || error.message), 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleDisableTotp = async () => {
    if (!disablePassword.trim()) {
      showMessage('请输入当前密码', 'warning');
      return;
    }

    if (!disableVerificationCode.trim()) {
      showMessage('请输入当前身份验证器验证码', 'warning');
      return;
    }

    setLoading(true);
    try {
      await disableTotp({
        password: disablePassword.trim(),
        code: disableVerificationCode.trim()
      });
      setTotpStatus({ enabled: false, recoveryCodes: [], recoveryCodesRemaining: 0, pendingEnrollment: false });
      resetTotpEnrollment();
      setDisableVerificationCode('');
      setDisablePassword('');
      setTotpPassword('');
      setTotpReauthCode('');
      showMessage('双重验证已关闭');
      fetchTotpStatus();
    } catch (error) {
      showMessage('关闭 TOTP 失败: ' + (error.response?.data?.message || error.message), 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleRegenerateRecoveryCodes = async () => {
    if (!disablePassword.trim()) {
      showMessage('请输入当前密码以重新生成恢复码', 'warning');
      return;
    }

    if (!disableVerificationCode.trim()) {
      showMessage('请输入当前身份验证器验证码以重新生成恢复码', 'warning');
      return;
    }

    setLoading(true);
    try {
      const response = await regenerateRecoveryCodes({
        password: disablePassword.trim(),
        code: disableVerificationCode.trim()
      });
      const codes = response.data?.recoveryCodes || [];
      setTotpStatus((prev) => ({ ...prev, recoveryCodes: codes, recoveryCodesRemaining: codes.length }));
      setTotpEnrollment((prev) => ({ ...prev, recoveryCodes: codes }));
      setDisableVerificationCode('');
      showMessage('恢复码已重新生成，请保存新的恢复码');
    } catch (error) {
      showMessage('重新生成恢复码失败: ' + (error.response?.data?.message || error.message), 'error');
    } finally {
      setLoading(false);
    }
  };

  const visibleRecoveryCodes = totpStatus.recoveryCodes?.length ? totpStatus.recoveryCodes : totpEnrollment.recoveryCodes;

  const handleTestUserAISettings = async () => {
    if (!aiForm.baseUrl.trim() && !aiForm.configured) {
      showMessage('请先填写 AI Base URL', 'warning');
      return;
    }

    if (!aiForm.model.trim() && !aiForm.configured) {
      showMessage('请先填写模型名称', 'warning');
      return;
    }

    if (!aiForm.apiKey.trim() && !aiForm.hasKey) {
      showMessage('请先填写 API Key', 'warning');
      return;
    }

    let payload;
    try {
      payload = buildAISettingsPayload();
    } catch (error) {
      showMessage(error.message, 'warning');
      return;
    }

    setAIAction('test');
    setLoading(true);
    try {
      const response = await testAISettings(payload);
      setAITestResult(response.data || null);
      showMessage('AI 连接测试成功');
    } catch (error) {
      setAITestResult(null);
      showMessage('连接测试失败: ' + (error.response?.data?.message || error.message), 'error');
    } finally {
      setLoading(false);
      setAIAction('');
    }
  };

  const handleSaveUserAISettings = async () => {
    if (!aiForm.password.trim()) {
      showMessage('请输入当前密码以保存 AI 设置', 'warning');
      return;
    }

    let payload;
    try {
      payload = buildAISettingsPayload();
    } catch (error) {
      showMessage(error.message, 'warning');
      return;
    }

    if (payload.enabled && !payload.baseUrl) {
      showMessage('启用 AI 助手时必须填写 AI Base URL', 'warning');
      return;
    }

    if (payload.enabled && !payload.model) {
      showMessage('启用 AI 助手时必须填写模型名称', 'warning');
      return;
    }

    if (payload.enabled && !payload.apiKey && !aiForm.hasKey) {
      showMessage('启用 AI 助手时必须提供 API Key', 'warning');
      return;
    }

    setAIAction('save');
    setLoading(true);
    try {
      await updateAISettings({
        ...payload,
        password: aiForm.password,
        code: aiForm.code.trim()
      });
      showMessage('AI 设置保存成功');
      await fetchUserAISettings();
    } catch (error) {
      showMessage('保存 AI 设置失败: ' + (error.response?.data?.message || error.message), 'error');
    } finally {
      setLoading(false);
      setAIAction('');
    }
  };

  const handleUpdateProfile = async () => {
    if (!profileForm.username.trim()) {
      showMessage('用户名不能为空', 'warning');
      return;
    }

    const usernameChanged = user?.username !== profileForm.username;

    setLoading(true);
    try {
      await updateProfile({
        username: profileForm.username.trim(),
        nickname: profileForm.nickname.trim(),
        password: profilePassword,
        code: profileCode.trim()
      });
      showMessage('资料更新成功');
      setProfilePassword('');
      setProfileCode('');

      if (usernameChanged) {
        showMessage('用户名已修改，需要重新登录...', 'warning');
        setTimeout(() => {
          logout();
        }, 2000);
      }
    } catch (error) {
      showMessage('更新失败: ' + (error.response?.data?.message || '未知错误'), 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleChangePassword = async () => {
    if (!passwordForm.oldPassword) {
      showMessage('请输入旧密码', 'warning');
      return;
    }
    if (!passwordForm.newPassword) {
      showMessage('请输入新密码', 'warning');
      return;
    }
    if (passwordForm.newPassword.length < 6) {
      showMessage('新密码长度至少6位', 'warning');
      return;
    }
    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      showMessage('两次输入的密码不一致', 'warning');
      return;
    }

    setLoading(true);
    try {
      const res = await changePassword({
        oldPassword: passwordForm.oldPassword,
        newPassword: passwordForm.newPassword,
        confirmPassword: passwordForm.confirmPassword,
        code: passwordForm.code.trim()
      });

      if (res.code !== 200) {
        throw new Error(res.msg || '修改失败');
      }
      showMessage('密码修改成功，即将重新登录...', 'success');
      resetPasswordForm();
      setPasswordDialogOpen(false);
      setTimeout(() => {
        logout();
      }, 2000);
    } catch (error) {
      const errorMsg = error.response?.data?.message || error.message || '';
      if (errorMsg.includes('password') || errorMsg.includes('密码')) {
        showMessage('旧密码不正确', 'error');
      } else {
        showMessage('修改失败: ' + errorMsg, 'error');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <Stack spacing={3}>
      <Card>
        <CardContent sx={{ p: { xs: 2, sm: 3 } }}>
          <Grid container spacing={3} alignItems="center">
            <Grid item xs={12} md={8}>
              <Stack direction="row" spacing={2} alignItems="center">
                <Avatar
                  src={user?.avatar}
                  sx={{
                    width: { xs: 64, sm: 76 },
                    height: { xs: 64, sm: 76 },
                    color: 'primary.dark',
                    bgcolor: 'primary.200',
                    fontSize: { xs: '1.75rem', sm: '2rem' }
                  }}
                >
                  {user?.username?.charAt(0)?.toUpperCase() || 'U'}
                </Avatar>
                <Stack spacing={0.75} sx={{ minWidth: 0 }}>
                  <Typography variant="h3" sx={{ wordBreak: 'break-word' }}>
                    {user?.username || '用户'}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    管理资料、双重验证、密码与个人 AI 助手配置。
                  </Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ wordBreak: 'break-word' }}>
                    当前昵称：{user?.nickname || '未设置'}
                  </Typography>
                  <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                    <Chip
                      label={totpStatus.enabled ? '双重验证已启用' : '双重验证未启用'}
                      color={totpStatus.enabled ? 'success' : 'default'}
                      size="small"
                    />
                    <Chip
                      label={aiForm.enabled ? 'AI 助手已启用' : aiForm.configured ? 'AI 已配置' : 'AI 未配置'}
                      color={aiForm.enabled ? 'primary' : aiForm.configured ? 'info' : 'default'}
                      size="small"
                      variant={aiForm.enabled ? 'filled' : 'outlined'}
                    />
                  </Stack>
                </Stack>
              </Stack>
            </Grid>

            <Grid item xs={12} md={4}>
              <Stack spacing={1.5} alignItems={{ xs: 'stretch', md: 'flex-end' }} sx={{ width: '100%' }}>
                <Button
                  variant="outlined"
                  startIcon={<LockIcon />}
                  onClick={() => setPasswordDialogOpen(true)}
                  sx={{ alignSelf: { xs: 'stretch', md: 'flex-end' } }}
                >
                  修改密码
                </Button>
              </Stack>
            </Grid>
          </Grid>
        </CardContent>
      </Card>

      <Card>
        <CardHeader title="个人设置" subheader="更新资料并管理账号安全。" />
        <CardContent sx={{ pt: 0 }}>
          <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
            <Tabs
              value={settingsSection}
              onChange={(event, value) => setSettingsSection(value)}
              variant="scrollable"
              scrollButtons="auto"
              allowScrollButtonsMobile
              aria-label="profile settings sections"
              sx={{
                '& .MuiTab-root': {
                  minHeight: 48,
                  textTransform: 'none',
                  fontWeight: 500
                }
              }}
            >
              <Tab icon={<SettingsSuggestIcon sx={{ mr: 1 }} />} iconPosition="start" label="基本资料" />
              <Tab icon={<SecurityIcon sx={{ mr: 1 }} />} iconPosition="start" label="安全设置" />
              <Tab icon={<PsychologyIcon sx={{ mr: 1 }} />} iconPosition="start" label="AI 助手" />
            </Tabs>
          </Box>

          {settingsSection === 0 && (
            <Grid container spacing={3}>
              <Grid item xs={12} lg={7}>
                <Card variant="outlined">
                  <CardHeader
                    title="基础资料"
                    subheader="更新用户名和昵称。保存时需要验证当前身份。"
                    avatar={<PersonIcon color="primary" />}
                  />
                  <CardContent>
                    <Stack spacing={2.5}>
                      <TextField
                        fullWidth
                        label="用户名"
                        value={profileForm.username}
                        onChange={(e) => setProfileForm({ ...profileForm, username: e.target.value })}
                        InputProps={{
                          startAdornment: (
                            <InputAdornment position="start">
                              <PersonIcon color="action" />
                            </InputAdornment>
                          )
                        }}
                      />
                      <TextField
                        fullWidth
                        label="昵称"
                        value={profileForm.nickname}
                        onChange={(e) => setProfileForm({ ...profileForm, nickname: e.target.value })}
                        placeholder="可选"
                      />

                      <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 2 }}>
                        <Stack spacing={2}>
                          <Typography variant="subtitle2">身份确认</Typography>
                          <TextField
                            fullWidth
                            type="password"
                            label="当前密码"
                            value={profilePassword}
                            onChange={(e) => setProfilePassword(e.target.value)}
                            autoComplete="current-password"
                            helperText="保存时需要输入当前密码。"
                          />
                          <TextField
                            fullWidth
                            label="当前 TOTP 验证码（已启用时必填）"
                            value={profileCode}
                            onChange={(e) => setProfileCode(e.target.value.replace(/\s+/g, '').slice(0, 8))}
                            inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                            helperText="已启用双重验证时填写。"
                          />
                        </Stack>
                      </Box>

                      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
                        <Button variant="contained" onClick={handleUpdateProfile} disabled={loading} startIcon={<SaveIcon />}>
                          更新资料
                        </Button>
                      </Stack>
                    </Stack>
                  </CardContent>
                </Card>
              </Grid>

              <Grid item xs={12} lg={5}>
                <Stack spacing={2.5}>
                  <Alert severity="info">修改用户名后需要重新登录。</Alert>
                  <Card variant="outlined">
                    <CardHeader title="主题切换入口" subheader="全局主题模式已移动到右上角 Header，作为整个应用的主入口。" />
                    <CardContent>
                      <Stack spacing={1}>
                        <Typography variant="body2" color="text.secondary">
                          请使用右上角 Header 中的主题按钮切换浅色、深色或跟随系统模式。
                        </Typography>
                        <Typography variant="caption" color="text.secondary">
                          主题选择会自动保存，并在下次登录后继续使用。
                        </Typography>
                      </Stack>
                    </CardContent>
                  </Card>
                  <Card variant="outlined">
                    <CardHeader title="保存前确认" subheader="避免因安全校验导致保存失败。" />
                    <CardContent>
                      <Stack spacing={1.5}>
                        <Box>
                          <Typography variant="subtitle2">本次保存需要</Typography>
                          <Typography variant="body2" color="text.secondary">
                            当前密码{totpStatus.enabled ? '，以及当前 TOTP 验证码。' : '。'}
                          </Typography>
                        </Box>
                        <Divider />
                        <Box>
                          <Typography variant="subtitle2">修改用户名时</Typography>
                          <Typography variant="body2" color="text.secondary">
                            保存成功后会自动退出登录，请使用新用户名重新登录。
                          </Typography>
                        </Box>
                        <Divider />
                        <Box>
                          <Typography variant="subtitle2">当前安全状态</Typography>
                          <Typography variant="body2" color="text.secondary">
                            {totpStatus.enabled ? '双重验证已启用。' : '尚未启用双重验证。'}
                          </Typography>
                        </Box>
                      </Stack>
                    </CardContent>
                  </Card>
                </Stack>
              </Grid>
            </Grid>
          )}

          {settingsSection === 1 && (
            <Grid container spacing={3}>
              <Grid item xs={12} xl={5}>
                <Stack spacing={2.5}>
                  <Card variant="outlined">
                    <CardHeader
                      title="安全总览"
                      subheader="查看双重验证状态与恢复方式。"
                      avatar={<ShieldOutlinedIcon color="primary" />}
                      action={
                        <Chip
                          label={totpStatus.enabled ? '已启用' : '未启用'}
                          color={totpStatus.enabled ? 'success' : 'default'}
                          size="small"
                          variant={totpStatus.enabled ? 'filled' : 'outlined'}
                        />
                      }
                    />
                    <CardContent>
                      <Stack spacing={2}>
                        <Alert severity={totpStatus.enabled ? 'success' : 'info'}>
                          {totpStatus.enabled
                            ? '登录时需要身份验证器验证码；设备不可用时可改用恢复码。'
                            : '启用后，登录将增加一步验证码校验。完成绑定后请立即保存恢复码。'}
                        </Alert>
                      </Stack>
                    </CardContent>
                  </Card>

                  {!totpStatus.enabled && !totpEnrollment.qrCodeData && (
                    <Card variant="outlined">
                      <CardHeader title="启用双重验证" subheader="先验证当前密码，再绑定身份验证器。" />
                      <CardContent>
                        <Stack spacing={2}>
                          <Typography variant="body2" color="text.secondary">
                            支持常见的 TOTP 身份验证器应用。
                          </Typography>
                          <TextField
                            fullWidth
                            type="password"
                            label="当前密码"
                            value={totpPassword}
                            onChange={(e) => setTotpPassword(e.target.value)}
                            autoComplete="current-password"
                            helperText="开始设置前需要验证当前密码。"
                          />
                          <Button variant="contained" onClick={startTotpSetup} disabled={loading} sx={{ alignSelf: 'flex-start' }}>
                            开始设置 TOTP
                          </Button>
                        </Stack>
                      </CardContent>
                    </Card>
                  )}

                  {totpStatus.enabled && (
                    <Card variant="outlined">
                      <CardHeader title="敏感操作验证" subheader="关闭双重验证或重置恢复码前，请先完成验证。" />
                      <CardContent>
                        <Stack spacing={2}>
                          <TextField
                            fullWidth
                            type="password"
                            label="当前密码"
                            value={disablePassword}
                            onChange={(e) => setDisablePassword(e.target.value)}
                            autoComplete="current-password"
                            helperText="继续前需要输入当前密码。"
                          />
                          <TextField
                            fullWidth
                            label="当前验证码"
                            value={disableVerificationCode}
                            onChange={(e) => setDisableVerificationCode(e.target.value.replace(/\s+/g, '').slice(0, 8))}
                            inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                            helperText="输入身份验证器当前显示的验证码。"
                          />
                          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
                            <Button
                              variant="outlined"
                              startIcon={<CachedIcon />}
                              onClick={handleRegenerateRecoveryCodes}
                              disabled={loading}
                            >
                              重新生成恢复码
                            </Button>
                            <Button color="error" variant="outlined" onClick={handleDisableTotp} disabled={loading}>
                              关闭双重验证
                            </Button>
                          </Stack>
                        </Stack>
                      </CardContent>
                    </Card>
                  )}
                </Stack>
              </Grid>

              <Grid item xs={12} xl={7}>
                <Stack spacing={2.5}>
                  {!totpStatus.enabled && totpEnrollment.qrCodeData && (
                    <Card variant="outlined">
                      <CardHeader title="完成 TOTP 绑定" subheader="扫描二维码后，输入验证码完成启用。" />
                      <CardContent>
                        <Stack spacing={2.5}>
                          <Stack direction={{ xs: 'column', md: 'row' }} spacing={3} alignItems={{ xs: 'stretch', md: 'flex-start' }}>
                            <Box
                              sx={{
                                p: 2,
                                border: '1px solid',
                                borderColor: 'divider',
                                borderRadius: 1,
                                bgcolor: 'common.white',
                                width: 'fit-content',
                                mx: { xs: 'auto', md: 0 }
                              }}
                            >
                              <QRCodeSVG value={totpEnrollment.qrCodeData} size={180} />
                            </Box>

                            <Stack spacing={2} sx={{ flex: 1, minWidth: 0 }}>
                              <Alert severity="info">使用身份验证器扫描二维码；无法扫描时可改为手动输入密钥。</Alert>
                              <TextField
                                fullWidth
                                label="手动输入密钥"
                                value={totpEnrollment.manualEntryKey}
                                InputProps={{
                                  readOnly: true,
                                  endAdornment: (
                                    <InputAdornment position="end">
                                      <Tooltip title="复制密钥">
                                        <IconButton onClick={() => handleCopy(totpEnrollment.manualEntryKey, '密钥')} edge="end">
                                          <ContentCopyIcon />
                                        </IconButton>
                                      </Tooltip>
                                    </InputAdornment>
                                  )
                                }}
                                helperText="扫码失败时，可复制此密钥手动添加账户。"
                              />
                              <TextField
                                fullWidth
                                label="验证码"
                                value={totpCode}
                                onChange={(e) => setTotpCode(e.target.value.replace(/\s+/g, '').slice(0, 8))}
                                inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                                helperText="输入身份验证器当前显示的验证码。"
                              />
                              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
                                <Button variant="contained" onClick={handleConfirmTotpSetup} disabled={loading}>
                                  确认启用
                                </Button>
                                <Button variant="outlined" onClick={resetTotpEnrollment} disabled={loading}>
                                  取消设置
                                </Button>
                              </Stack>
                            </Stack>
                          </Stack>
                        </Stack>
                      </CardContent>
                    </Card>
                  )}

                  {totpStatus.enabled && (
                    <Card variant="outlined">
                      <CardHeader title="恢复码" subheader="每个恢复码只能使用一次，请及时离线保存。" />
                      <CardContent>
                        <Stack spacing={2}>
                          <Alert severity="warning">
                            关闭双重验证或重新生成恢复码前，请确认仍可访问当前身份验证器，或已保存可用恢复码。
                          </Alert>

                          {visibleRecoveryCodes?.length > 0 ? (
                            <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 2 }}>
                              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} justifyContent="space-between" sx={{ mb: 1.5 }}>
                                <Typography variant="subtitle1">恢复码列表</Typography>
                                <Button
                                  variant="text"
                                  startIcon={<ContentCopyIcon />}
                                  onClick={() => handleCopy(visibleRecoveryCodes.join('\n'), '恢复码')}
                                >
                                  复制全部
                                </Button>
                              </Stack>
                              <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 1.5 }}>
                                每个恢复码只能使用一次。请离线保存，不要与账号密码放在同一处。
                              </Typography>
                              <List dense disablePadding>
                                {visibleRecoveryCodes.map((code) => (
                                  <ListItem
                                    key={code}
                                    disableGutters
                                    secondaryAction={
                                      <IconButton edge="end" onClick={() => handleCopy(code, '恢复码')}>
                                        <ContentCopyIcon fontSize="small" />
                                      </IconButton>
                                    }
                                  >
                                    <ListItemText primary={code} primaryTypographyProps={{ sx: { fontFamily: 'monospace' } }} />
                                  </ListItem>
                                ))}
                              </List>
                            </Box>
                          ) : (
                            <Alert severity="info">启用后生成的恢复码会展示在这里，请及时保存。</Alert>
                          )}
                        </Stack>
                      </CardContent>
                    </Card>
                  )}

                  {!totpStatus.enabled && !totpEnrollment.qrCodeData && (
                    <Card variant="outlined">
                      <CardHeader title="启用建议" />
                      <CardContent>
                        <Stack spacing={1.5}>
                          <Typography variant="body2" color="text.secondary">
                            双重验证会在密码之外增加一步验证码校验，更适合保护管理后台与安全敏感操作。
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            启用后请立即保存恢复码，并确认常用设备上的身份验证器可正常使用。
                          </Typography>
                        </Stack>
                      </CardContent>
                    </Card>
                  )}
                </Stack>
              </Grid>
            </Grid>
          )}

          {settingsSection === 2 && (
            <Card variant="outlined">
              <CardHeader
                title="AI 助手设置"
                subheader="为当前账号配置独立的 Responses API 模型参数。"
                avatar={<PsychologyIcon color="primary" />}
                action={
                  <Stack direction="row" spacing={1} alignItems="center" useFlexGap flexWrap="wrap" justifyContent="flex-end">
                    <FormControlLabel
                      sx={{ mr: 0 }}
                      control={<Switch checked={aiForm.enabled} onChange={(e) => setAIField('enabled', e.target.checked)} />}
                      label={aiForm.enabled ? '启用' : '禁用'}
                    />
                  </Stack>
                }
              />
              <CardContent>
                {aiSettingsLoading ? (
                  <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
                    <CircularProgress size={28} />
                  </Box>
                ) : (
                  <Stack spacing={2.5}>
                    <Alert severity="info">
                      模板编辑器中的 AI 助手会使用这里的个人配置。当前仅支持提供“/responses endpoint”的 AI
                      服务。保存需要当前密码；如已启用双重验证，还需要输入当前验证码。
                    </Alert>

                    <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 2 }}>
                      <Stack spacing={2}>
                        <Box>
                          <Typography variant="subtitle2">启用个人 AI 助手</Typography>
                          <Typography variant="body2" color="text.secondary">
                            当前接口类型：Responses API（仅支持 `/responses` endpoint）。
                          </Typography>
                        </Box>

                        <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                          <Chip
                            size="small"
                            color={aiForm.configured ? 'info' : 'default'}
                            variant={aiForm.configured ? 'filled' : 'outlined'}
                            label={aiForm.configured ? '已保存连接参数' : '尚未完成配置'}
                          />
                          <Chip
                            size="small"
                            color={aiForm.hasKey ? 'success' : 'default'}
                            variant={aiForm.hasKey ? 'outlined' : 'outlined'}
                            label={aiForm.hasKey ? `已保存 API Key：${aiForm.maskedKey || '已隐藏'}` : '未保存 API Key'}
                          />
                        </Stack>

                        <TextField
                          fullWidth
                          label="AI Base URL"
                          value={aiForm.baseUrl}
                          onChange={(e) => setAIField('baseUrl', e.target.value)}
                          placeholder="https://api.openai.com/v1"
                          helperText="需为可用的 Responses API 根地址，并且服务端必须支持 `/responses` endpoint；保存时会校验是否为可公开访问的 HTTP/HTTPS 地址。"
                        />
                        <TextField
                          fullWidth
                          type="password"
                          label="API Key"
                          value={aiForm.apiKey}
                          onChange={(e) => setAIField('apiKey', e.target.value)}
                          placeholder={aiForm.hasKey ? '留空则继续使用已保存的密钥' : '输入新的 API Key'}
                          autoComplete="new-password"
                          helperText={
                            aiForm.hasKey
                              ? `当前已保存密钥：${aiForm.maskedKey || '已隐藏'}；如不需要更换，请保持留空。`
                              : '当前尚未保存 API Key。'
                          }
                        />
                        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} alignItems={{ xs: 'stretch', sm: 'flex-start' }}>
                          <Autocomplete
                            freeSolo
                            fullWidth
                            options={aiModelOptions}
                            value={aiForm.model}
                            onChange={(event, newValue) => setAIField('model', typeof newValue === 'string' ? newValue : newValue || '')}
                            onInputChange={(event, newValue) => setAIField('model', newValue || '')}
                            renderInput={(params) => (
                              <TextField
                                {...params}
                                label="模型名称"
                                placeholder="gpt-4.1-mini"
                                helperText={
                                  aiModelsFetched
                                    ? '可从已获取列表中选择，也可以继续手动输入支持 `/responses` 的模型 ID。'
                                    : '可手动输入模型 ID，也可先根据当前 Responses 接口配置获取模型列表。'
                                }
                              />
                            )}
                          />
                          <Button
                            variant="outlined"
                            onClick={handleFetchAIModels}
                            disabled={loading || aiSettingsLoading}
                            startIcon={loading && aiAction === 'models' ? <CircularProgress size={18} /> : <CachedIcon />}
                            sx={{ flexShrink: 0, minWidth: { sm: 148 }, height: { sm: 56 } }}
                          >
                            获取模型
                          </Button>
                        </Stack>

                        <Grid container spacing={2}>
                          <Grid item xs={12} sm={6}>
                            <TextField
                              fullWidth
                              type="number"
                              label="Temperature"
                              value={aiForm.temperature}
                              onChange={(e) => setAIField('temperature', e.target.value === '' ? '' : Number(e.target.value))}
                              inputProps={{ min: 0, max: 2, step: 0.1 }}
                              helperText="范围 0 - 2，建议保留 0.2。"
                            />
                          </Grid>
                          <Grid item xs={12} sm={6}>
                            <TextField
                              fullWidth
                              type="number"
                              label="Max Tokens"
                              value={aiForm.maxTokens}
                              onChange={(e) => setAIField('maxTokens', e.target.value === '' ? '' : Number(e.target.value))}
                              inputProps={{ min: 0, step: 100 }}
                              helperText="设置为 0 表示使用服务端默认值。连接测试会改用最小请求，并单独限制很小的返回 token 数量。"
                            />
                          </Grid>
                        </Grid>

                        <TextField
                          fullWidth
                          multiline
                          minRows={5}
                          label="额外请求头（JSON）"
                          value={aiHeadersText}
                          onChange={(e) => {
                            setAIHeadersText(e.target.value);
                            setAIModelsFetched(false);
                          }}
                          helperText={'仅支持 JSON 对象，例如 {"HTTP-Referer":"https://example.com"}。留空或 {} 表示不追加请求头。'}
                        />
                      </Stack>
                    </Box>

                    <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 2 }}>
                      <Stack spacing={2}>
                        <Typography variant="subtitle2">保存前身份确认</Typography>
                        <TextField
                          fullWidth
                          type="password"
                          label="当前密码"
                          value={aiForm.password}
                          onChange={(e) => setAIField('password', e.target.value)}
                          autoComplete="current-password"
                          helperText="保存 AI 设置时必填。"
                        />
                        <TextField
                          fullWidth
                          label="当前 TOTP 验证码（已启用时必填）"
                          value={aiForm.code}
                          onChange={(e) => setAIField('code', e.target.value.replace(/\s+/g, '').slice(0, 8))}
                          inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                          helperText="如果已启用双重验证，请输入当前身份验证器验证码。"
                        />
                      </Stack>
                    </Box>

                    <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
                      <Button
                        variant="outlined"
                        startIcon={loading && aiAction === 'test' ? <CircularProgress size={18} /> : <ScienceIcon />}
                        onClick={handleTestUserAISettings}
                        disabled={loading}
                      >
                        测试连接
                      </Button>
                      <Button
                        variant="contained"
                        startIcon={loading && aiAction === 'save' ? <CircularProgress size={18} /> : <SaveIcon />}
                        onClick={handleSaveUserAISettings}
                        disabled={loading || aiSettingsLoading}
                      >
                        保存 AI 设置
                      </Button>
                    </Stack>

                    {aiTestResult && (
                      <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 2 }}>
                        <Stack spacing={1.75}>
                          <Alert severity="success">{aiTestResult.message || '连接测试成功'}</Alert>
                          <Grid container spacing={2}>
                            <Grid item xs={12} sm={6}>
                              <Box>
                                <Typography variant="subtitle2">模型</Typography>
                                <Typography variant="body2" color="text.secondary" sx={{ wordBreak: 'break-word' }}>
                                  {aiTestResult.model || aiForm.model || '-'}
                                </Typography>
                              </Box>
                            </Grid>
                            <Grid item xs={12} sm={6}>
                              <Box>
                                <Typography variant="subtitle2">延迟</Typography>
                                <Typography variant="body2" color="text.secondary">
                                  {typeof aiTestResult.latencyMs === 'number' ? `${aiTestResult.latencyMs} ms` : '-'}
                                </Typography>
                              </Box>
                            </Grid>
                            <Grid item xs={12}>
                              <Box>
                                <Typography variant="subtitle2">Base URL</Typography>
                                <Typography variant="body2" color="text.secondary" sx={{ wordBreak: 'break-word' }}>
                                  {aiTestResult.baseUrl || aiForm.baseUrl || '-'}
                                </Typography>
                              </Box>
                            </Grid>
                            {aiTestResult.finishReason && (
                              <Grid item xs={12}>
                                <Box>
                                  <Typography variant="subtitle2">完成原因</Typography>
                                  <Typography variant="body2" color="text.secondary">
                                    {aiTestResult.finishReason}
                                  </Typography>
                                </Box>
                              </Grid>
                            )}
                          </Grid>
                          {aiTestResult.usage && Object.keys(aiTestResult.usage).length > 0 && (
                            <TextField
                              fullWidth
                              multiline
                              minRows={3}
                              label="用量信息"
                              value={JSON.stringify(aiTestResult.usage, null, 2)}
                              InputProps={{ readOnly: true }}
                            />
                          )}
                        </Stack>
                      </Box>
                    )}

                    <Alert severity="warning">AI 候选修改仍需人工检查，并在模板编辑器中手动点击应用。</Alert>
                  </Stack>
                )}
              </CardContent>
            </Card>
          )}
        </CardContent>
      </Card>

      <Dialog
        open={passwordDialogOpen}
        onClose={loading ? undefined : () => setPasswordDialogOpen(false)}
        maxWidth="sm"
        fullWidth
        fullScreen={fullScreenDialog}
      >
        <DialogTitle>修改密码</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2.5} sx={{ pt: 0.5 }}>
            <Alert severity="info">修改成功后将重新登录；已启用双重验证时，还需要输入当前验证码。</Alert>

            <TextField
              fullWidth
              label="旧密码"
              type={showOldPassword ? 'text' : 'password'}
              value={passwordForm.oldPassword}
              onChange={(e) => setPasswordForm({ ...passwordForm, oldPassword: e.target.value })}
              autoComplete="current-password"
              InputProps={{
                endAdornment: (
                  <InputAdornment position="end">
                    <IconButton
                      onClick={() => setShowOldPassword(!showOldPassword)}
                      edge="end"
                      aria-label={showOldPassword ? '隐藏旧密码' : '显示旧密码'}
                    >
                      {showOldPassword ? <VisibilityOff /> : <Visibility />}
                    </IconButton>
                  </InputAdornment>
                )
              }}
            />
            <TextField
              fullWidth
              label="新密码"
              type={showNewPassword ? 'text' : 'password'}
              value={passwordForm.newPassword}
              onChange={(e) => setPasswordForm({ ...passwordForm, newPassword: e.target.value })}
              autoComplete="new-password"
              helperText="密码长度至少6位"
              InputProps={{
                endAdornment: (
                  <InputAdornment position="end">
                    <IconButton
                      onClick={() => setShowNewPassword(!showNewPassword)}
                      edge="end"
                      aria-label={showNewPassword ? '隐藏新密码' : '显示新密码'}
                    >
                      {showNewPassword ? <VisibilityOff /> : <Visibility />}
                    </IconButton>
                  </InputAdornment>
                )
              }}
            />
            <TextField
              fullWidth
              label="确认新密码"
              type={showConfirmPassword ? 'text' : 'password'}
              value={passwordForm.confirmPassword}
              onChange={(e) => setPasswordForm({ ...passwordForm, confirmPassword: e.target.value })}
              autoComplete="new-password"
              error={passwordForm.confirmPassword && passwordForm.newPassword !== passwordForm.confirmPassword}
              helperText={
                passwordForm.confirmPassword && passwordForm.newPassword !== passwordForm.confirmPassword ? '两次输入的密码不一致' : ' '
              }
              InputProps={{
                endAdornment: (
                  <InputAdornment position="end">
                    <IconButton
                      onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                      edge="end"
                      aria-label={showConfirmPassword ? '隐藏确认密码' : '显示确认密码'}
                    >
                      {showConfirmPassword ? <VisibilityOff /> : <Visibility />}
                    </IconButton>
                  </InputAdornment>
                )
              }}
            />
            <TextField
              fullWidth
              label="当前 TOTP 验证码（已启用时必填）"
              value={passwordForm.code}
              onChange={(e) => setPasswordForm({ ...passwordForm, code: e.target.value.replace(/\s+/g, '').slice(0, 8) })}
              inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
              helperText="已启用双重验证时填写。"
            />
          </Stack>
        </DialogContent>
        <DialogActions sx={{ px: 3, py: 2 }}>
          <Button onClick={() => setPasswordDialogOpen(false)} color="inherit" disabled={loading}>
            取消
          </Button>
          <Button onClick={resetPasswordForm} variant="outlined" disabled={loading}>
            重置
          </Button>
          <Button variant="contained" onClick={handleChangePassword} disabled={loading} startIcon={<LockIcon />}>
            修改密码
          </Button>
        </DialogActions>
      </Dialog>
    </Stack>
  );
}
