import { useState, useEffect } from 'react';

import Alert from '@mui/material/Alert';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Divider from '@mui/material/Divider';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import InputAdornment from '@mui/material/InputAdornment';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import Stack from '@mui/material/Stack';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import TextField from '@mui/material/TextField';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme } from '@mui/material/styles';

import CachedIcon from '@mui/icons-material/Cached';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import LockIcon from '@mui/icons-material/Lock';
import PersonIcon from '@mui/icons-material/Person';
import SaveIcon from '@mui/icons-material/Save';
import SecurityIcon from '@mui/icons-material/Security';
import SettingsSuggestIcon from '@mui/icons-material/SettingsSuggest';
import ShieldOutlinedIcon from '@mui/icons-material/ShieldOutlined';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';

import { useAuth } from 'contexts/AuthContext';
import { changePassword, updateProfile } from 'api/user';
import { QRCodeSVG } from 'qrcode.react';
import { confirmTotpSetup, disableTotp, getTotpStatus, regenerateRecoveryCodes, setupTotp } from 'api/auth';

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
  }, []);

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
                    管理资料、双重验证与密码。
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
