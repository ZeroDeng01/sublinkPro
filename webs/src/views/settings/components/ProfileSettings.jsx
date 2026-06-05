import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

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
  const { t } = useTranslation();
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
      showMessage(t('settings.profilePanel.messages.copied', { label }));
    } catch {
      showMessage(t('settings.profilePanel.messages.copyFailed', { label }), 'warning');
    }
  };

  const startTotpSetup = async () => {
    if (!totpPassword.trim()) {
      showMessage(t('settings.profilePanel.messages.totpPasswordRequired'), 'warning');
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
      showMessage(t('settings.profilePanel.messages.scanTotp'));
    } catch (error) {
      showMessage(
        t('settings.profilePanel.messages.totpSetupFailed', { message: error.response?.data?.message || error.message }),
        'error'
      );
    } finally {
      setLoading(false);
    }
  };

  const handleConfirmTotpSetup = async () => {
    if (!totpCode.trim()) {
      showMessage(t('settings.profilePanel.messages.totpCodeRequired'), 'warning');
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
      showMessage(t('settings.profilePanel.messages.totpEnabled'));
      fetchTotpStatus();
    } catch (error) {
      showMessage(
        t('settings.profilePanel.messages.totpEnableFailed', { message: error.response?.data?.message || error.message }),
        'error'
      );
    } finally {
      setLoading(false);
    }
  };

  const handleDisableTotp = async () => {
    if (!disablePassword.trim()) {
      showMessage(t('settings.profilePanel.messages.currentPasswordRequired'), 'warning');
      return;
    }

    if (!disableVerificationCode.trim()) {
      showMessage(t('settings.profilePanel.messages.currentTotpRequired'), 'warning');
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
      showMessage(t('settings.profilePanel.messages.totpDisabled'));
      fetchTotpStatus();
    } catch (error) {
      showMessage(
        t('settings.profilePanel.messages.totpDisableFailed', { message: error.response?.data?.message || error.message }),
        'error'
      );
    } finally {
      setLoading(false);
    }
  };

  const handleRegenerateRecoveryCodes = async () => {
    if (!disablePassword.trim()) {
      showMessage(t('settings.profilePanel.messages.regenPasswordRequired'), 'warning');
      return;
    }

    if (!disableVerificationCode.trim()) {
      showMessage(t('settings.profilePanel.messages.regenTotpRequired'), 'warning');
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
      showMessage(t('settings.profilePanel.messages.recoveryCodesRegenerated'));
    } catch (error) {
      showMessage(
        t('settings.profilePanel.messages.regenerateFailed', { message: error.response?.data?.message || error.message }),
        'error'
      );
    } finally {
      setLoading(false);
    }
  };

  const visibleRecoveryCodes = totpStatus.recoveryCodes?.length ? totpStatus.recoveryCodes : totpEnrollment.recoveryCodes;

  const handleUpdateProfile = async () => {
    if (!profileForm.username.trim()) {
      showMessage(t('settings.profilePanel.messages.usernameRequired'), 'warning');
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
      showMessage(t('settings.profilePanel.messages.profileUpdated'));
      setProfilePassword('');
      setProfileCode('');

      if (usernameChanged) {
        showMessage(t('settings.profilePanel.messages.usernameChanged'), 'warning');
        setTimeout(() => {
          logout();
        }, 2000);
      }
    } catch (error) {
      showMessage(
        t('settings.profilePanel.messages.updateFailed', { message: error.response?.data?.message || t('common.unknown') }),
        'error'
      );
    } finally {
      setLoading(false);
    }
  };

  const handleChangePassword = async () => {
    if (!passwordForm.oldPassword) {
      showMessage(t('settings.profilePanel.messages.oldPasswordRequired'), 'warning');
      return;
    }
    if (!passwordForm.newPassword) {
      showMessage(t('settings.profilePanel.messages.newPasswordRequired'), 'warning');
      return;
    }
    if (passwordForm.newPassword.length < 6) {
      showMessage(t('settings.profilePanel.messages.passwordTooShort'), 'warning');
      return;
    }
    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      showMessage(t('settings.profilePanel.messages.passwordMismatch'), 'warning');
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
        throw new Error(res.msg || t('settings.profilePanel.messages.changePasswordFailed'));
      }
      showMessage(t('settings.profilePanel.messages.passwordChanged'), 'success');
      resetPasswordForm();
      setPasswordDialogOpen(false);
      setTimeout(() => {
        logout();
      }, 2000);
    } catch (error) {
      const errorMsg = error.response?.data?.message || error.message || '';
      if (errorMsg.includes('password') || errorMsg.includes('密码')) {
        showMessage(t('settings.profilePanel.messages.oldPasswordIncorrect'), 'error');
      } else {
        showMessage(t('settings.profilePanel.messages.changeFailed', { message: errorMsg }), 'error');
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
                    {user?.username || t('settings.profilePanel.userFallback')}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    {t('settings.profilePanel.heroDescription')}
                  </Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ wordBreak: 'break-word' }}>
                    {t('settings.profilePanel.currentNickname', { nickname: user?.nickname || t('settings.profilePanel.notSet') })}
                  </Typography>
                  <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                    <Chip
                      label={
                        totpStatus.enabled ? t('settings.profilePanel.status.totpEnabled') : t('settings.profilePanel.status.totpDisabled')
                      }
                      color={totpStatus.enabled ? 'success' : 'default'}
                      size="small"
                      variant="outlined"
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
                  {t('settings.profilePanel.actions.changePassword')}
                </Button>
              </Stack>
            </Grid>
          </Grid>
        </CardContent>
      </Card>

      <Card>
        <CardHeader title={t('settings.profilePanel.title')} subheader={t('settings.profilePanel.subheader')} />
        <CardContent sx={{ pt: 0 }}>
          <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
            <Tabs
              value={settingsSection}
              onChange={(_event, value) => setSettingsSection(value)}
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
              <Tab icon={<SettingsSuggestIcon sx={{ mr: 1 }} />} iconPosition="start" label={t('settings.profilePanel.tabs.profile')} />
              <Tab icon={<SecurityIcon sx={{ mr: 1 }} />} iconPosition="start" label={t('settings.profilePanel.tabs.security')} />
            </Tabs>
          </Box>

          {settingsSection === 0 && (
            <Grid container spacing={3}>
              <Grid item xs={12} lg={7}>
                <Card variant="outlined">
                  <CardHeader
                    title={t('settings.profilePanel.basic.title')}
                    subheader={t('settings.profilePanel.basic.subheader')}
                    avatar={<PersonIcon color="primary" />}
                  />
                  <CardContent>
                    <Stack spacing={2.5}>
                      <TextField
                        fullWidth
                        label={t('settings.profilePanel.fields.username')}
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
                        label={t('settings.profilePanel.fields.nickname')}
                        value={profileForm.nickname}
                        onChange={(e) => setProfileForm({ ...profileForm, nickname: e.target.value })}
                        placeholder={t('settings.profilePanel.fields.optional')}
                      />

                      <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 2 }}>
                        <Stack spacing={2}>
                          <Typography variant="subtitle2">{t('settings.profilePanel.identity.title')}</Typography>
                          <TextField
                            fullWidth
                            type="password"
                            label={t('settings.profilePanel.fields.currentPassword')}
                            value={profilePassword}
                            onChange={(e) => setProfilePassword(e.target.value)}
                            autoComplete="current-password"
                            helperText={t('settings.profilePanel.fields.currentPasswordHelper')}
                          />
                          <TextField
                            fullWidth
                            label={t('settings.profilePanel.fields.currentTotpOptional')}
                            value={profileCode}
                            onChange={(e) => setProfileCode(e.target.value.replace(/\s+/g, '').slice(0, 8))}
                            inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                            helperText={t('settings.profilePanel.fields.currentTotpHelper')}
                          />
                        </Stack>
                      </Box>

                      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
                        <Button variant="contained" onClick={handleUpdateProfile} disabled={loading} startIcon={<SaveIcon />}>
                          {t('settings.profilePanel.actions.updateProfile')}
                        </Button>
                      </Stack>
                    </Stack>
                  </CardContent>
                </Card>
              </Grid>

              <Grid item xs={12} lg={5}>
                <Stack spacing={2.5}>
                  <Alert severity="info">{t('settings.profilePanel.basic.usernameChangeNotice')}</Alert>
                  <Card variant="outlined">
                    <CardHeader title={t('settings.profilePanel.theme.title')} subheader={t('settings.profilePanel.theme.subheader')} />
                    <CardContent>
                      <Stack spacing={1}>
                        <Typography variant="body2" color="text.secondary">
                          {t('settings.profilePanel.theme.description')}
                        </Typography>
                        <Typography variant="caption" color="text.secondary">
                          {t('settings.profilePanel.theme.helper')}
                        </Typography>
                      </Stack>
                    </CardContent>
                  </Card>
                  <Card variant="outlined">
                    <CardHeader
                      title={t('settings.profilePanel.saveCheck.title')}
                      subheader={t('settings.profilePanel.saveCheck.subheader')}
                    />
                    <CardContent>
                      <Stack spacing={1.5}>
                        <Box>
                          <Typography variant="subtitle2">{t('settings.profilePanel.saveCheck.requiresTitle')}</Typography>
                          <Typography variant="body2" color="text.secondary">
                            {totpStatus.enabled
                              ? t('settings.profilePanel.saveCheck.requiresPasswordAndTotp')
                              : t('settings.profilePanel.saveCheck.requiresPassword')}
                          </Typography>
                        </Box>
                        <Divider />
                        <Box>
                          <Typography variant="subtitle2">{t('settings.profilePanel.saveCheck.usernameTitle')}</Typography>
                          <Typography variant="body2" color="text.secondary">
                            {t('settings.profilePanel.saveCheck.usernameDescription')}
                          </Typography>
                        </Box>
                        <Divider />
                        <Box>
                          <Typography variant="subtitle2">{t('settings.profilePanel.saveCheck.securityTitle')}</Typography>
                          <Typography variant="body2" color="text.secondary">
                            {totpStatus.enabled
                              ? t('settings.profilePanel.saveCheck.securityEnabled')
                              : t('settings.profilePanel.saveCheck.securityDisabled')}
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
                      title={t('settings.profilePanel.security.title')}
                      subheader={t('settings.profilePanel.security.subheader')}
                      avatar={<ShieldOutlinedIcon color="primary" />}
                      action={
                        <Chip
                          label={totpStatus.enabled ? t('common.enabled') : t('common.disabled')}
                          color={totpStatus.enabled ? 'success' : 'default'}
                          size="small"
                          variant="outlined"
                        />
                      }
                    />
                    <CardContent>
                      <Stack spacing={2}>
                        <Alert severity={totpStatus.enabled ? 'success' : 'info'}>
                          {totpStatus.enabled
                            ? t('settings.profilePanel.security.enabledAlert')
                            : t('settings.profilePanel.security.disabledAlert')}
                        </Alert>
                      </Stack>
                    </CardContent>
                  </Card>

                  {!totpStatus.enabled && !totpEnrollment.qrCodeData && (
                    <Card variant="outlined">
                      <CardHeader
                        title={t('settings.profilePanel.totpSetup.title')}
                        subheader={t('settings.profilePanel.totpSetup.subheader')}
                      />
                      <CardContent>
                        <Stack spacing={2}>
                          <Typography variant="body2" color="text.secondary">
                            {t('settings.profilePanel.totpSetup.description')}
                          </Typography>
                          <TextField
                            fullWidth
                            type="password"
                            label={t('settings.profilePanel.fields.currentPassword')}
                            value={totpPassword}
                            onChange={(e) => setTotpPassword(e.target.value)}
                            autoComplete="current-password"
                            helperText={t('settings.profilePanel.fields.startSetupPasswordHelper')}
                          />
                          <Button variant="contained" onClick={startTotpSetup} disabled={loading} sx={{ alignSelf: 'flex-start' }}>
                            {t('settings.profilePanel.actions.startTotpSetup')}
                          </Button>
                        </Stack>
                      </CardContent>
                    </Card>
                  )}

                  {totpStatus.enabled && (
                    <Card variant="outlined">
                      <CardHeader
                        title={t('settings.profilePanel.sensitive.title')}
                        subheader={t('settings.profilePanel.sensitive.subheader')}
                      />
                      <CardContent>
                        <Stack spacing={2}>
                          <TextField
                            fullWidth
                            type="password"
                            label={t('settings.profilePanel.fields.currentPassword')}
                            value={disablePassword}
                            onChange={(e) => setDisablePassword(e.target.value)}
                            autoComplete="current-password"
                            helperText={t('settings.profilePanel.fields.continuePasswordHelper')}
                          />
                          <TextField
                            fullWidth
                            label={t('settings.profilePanel.fields.currentCode')}
                            value={disableVerificationCode}
                            onChange={(e) => setDisableVerificationCode(e.target.value.replace(/\s+/g, '').slice(0, 8))}
                            inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                            helperText={t('settings.profilePanel.fields.authenticatorCodeHelper')}
                          />
                          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
                            <Button
                              variant="outlined"
                              startIcon={<CachedIcon />}
                              onClick={handleRegenerateRecoveryCodes}
                              disabled={loading}
                            >
                              {t('settings.profilePanel.actions.regenerateRecoveryCodes')}
                            </Button>
                            <Button color="error" variant="outlined" onClick={handleDisableTotp} disabled={loading}>
                              {t('settings.profilePanel.actions.disableTotp')}
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
                      <CardHeader
                        title={t('settings.profilePanel.totpEnrollment.title')}
                        subheader={t('settings.profilePanel.totpEnrollment.subheader')}
                      />
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
                              <Alert severity="info">{t('settings.profilePanel.totpEnrollment.scanHint')}</Alert>
                              <TextField
                                fullWidth
                                label={t('settings.profilePanel.fields.manualEntryKey')}
                                value={totpEnrollment.manualEntryKey}
                                InputProps={{
                                  readOnly: true,
                                  endAdornment: (
                                    <InputAdornment position="end">
                                      <Tooltip title={t('settings.profilePanel.actions.copySecret')}>
                                        <IconButton
                                          onClick={() =>
                                            handleCopy(totpEnrollment.manualEntryKey, t('settings.profilePanel.labels.secret'))
                                          }
                                          edge="end"
                                        >
                                          <ContentCopyIcon />
                                        </IconButton>
                                      </Tooltip>
                                    </InputAdornment>
                                  )
                                }}
                                helperText={t('settings.profilePanel.fields.manualEntryKeyHelper')}
                              />
                              <TextField
                                fullWidth
                                label={t('settings.profilePanel.fields.code')}
                                value={totpCode}
                                onChange={(e) => setTotpCode(e.target.value.replace(/\s+/g, '').slice(0, 8))}
                                inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                                helperText={t('settings.profilePanel.fields.authenticatorCodeHelper')}
                              />
                              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
                                <Button variant="contained" onClick={handleConfirmTotpSetup} disabled={loading}>
                                  {t('settings.profilePanel.actions.confirmEnable')}
                                </Button>
                                <Button variant="outlined" onClick={resetTotpEnrollment} disabled={loading}>
                                  {t('settings.profilePanel.actions.cancelSetup')}
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
                      <CardHeader
                        title={t('settings.profilePanel.recovery.title')}
                        subheader={t('settings.profilePanel.recovery.subheader')}
                      />
                      <CardContent>
                        <Stack spacing={2}>
                          <Alert severity="warning">{t('settings.profilePanel.recovery.warning')}</Alert>

                          {visibleRecoveryCodes?.length > 0 ? (
                            <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 2 }}>
                              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} justifyContent="space-between" sx={{ mb: 1.5 }}>
                                <Typography variant="subtitle1">{t('settings.profilePanel.recovery.listTitle')}</Typography>
                                <Button
                                  variant="text"
                                  startIcon={<ContentCopyIcon />}
                                  onClick={() =>
                                    handleCopy(visibleRecoveryCodes.join('\n'), t('settings.profilePanel.labels.recoveryCodes'))
                                  }
                                >
                                  {t('settings.profilePanel.actions.copyAll')}
                                </Button>
                              </Stack>
                              <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 1.5 }}>
                                {t('settings.profilePanel.recovery.helper')}
                              </Typography>
                              <List dense disablePadding>
                                {visibleRecoveryCodes.map((code) => (
                                  <ListItem
                                    key={code}
                                    disableGutters
                                    secondaryAction={
                                      <IconButton
                                        edge="end"
                                        onClick={() => handleCopy(code, t('settings.profilePanel.labels.recoveryCode'))}
                                      >
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
                            <Alert severity="info">{t('settings.profilePanel.recovery.empty')}</Alert>
                          )}
                        </Stack>
                      </CardContent>
                    </Card>
                  )}

                  {!totpStatus.enabled && !totpEnrollment.qrCodeData && (
                    <Card variant="outlined">
                      <CardHeader title={t('settings.profilePanel.recommendation.title')} />
                      <CardContent>
                        <Stack spacing={1.5}>
                          <Typography variant="body2" color="text.secondary">
                            {t('settings.profilePanel.recommendation.description')}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            {t('settings.profilePanel.recommendation.helper')}
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
        <DialogTitle>{t('settings.profilePanel.passwordDialog.title')}</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2.5} sx={{ pt: 0.5 }}>
            <Alert severity="info">{t('settings.profilePanel.passwordDialog.notice')}</Alert>

            <TextField
              fullWidth
              label={t('settings.profilePanel.fields.oldPassword')}
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
                      aria-label={
                        showOldPassword
                          ? t('settings.profilePanel.passwordDialog.hideOld')
                          : t('settings.profilePanel.passwordDialog.showOld')
                      }
                    >
                      {showOldPassword ? <VisibilityOff /> : <Visibility />}
                    </IconButton>
                  </InputAdornment>
                )
              }}
            />
            <TextField
              fullWidth
              label={t('settings.profilePanel.fields.newPassword')}
              type={showNewPassword ? 'text' : 'password'}
              value={passwordForm.newPassword}
              onChange={(e) => setPasswordForm({ ...passwordForm, newPassword: e.target.value })}
              autoComplete="new-password"
              helperText={t('settings.profilePanel.fields.newPasswordHelper')}
              InputProps={{
                endAdornment: (
                  <InputAdornment position="end">
                    <IconButton
                      onClick={() => setShowNewPassword(!showNewPassword)}
                      edge="end"
                      aria-label={
                        showNewPassword
                          ? t('settings.profilePanel.passwordDialog.hideNew')
                          : t('settings.profilePanel.passwordDialog.showNew')
                      }
                    >
                      {showNewPassword ? <VisibilityOff /> : <Visibility />}
                    </IconButton>
                  </InputAdornment>
                )
              }}
            />
            <TextField
              fullWidth
              label={t('settings.profilePanel.fields.confirmPassword')}
              type={showConfirmPassword ? 'text' : 'password'}
              value={passwordForm.confirmPassword}
              onChange={(e) => setPasswordForm({ ...passwordForm, confirmPassword: e.target.value })}
              autoComplete="new-password"
              error={passwordForm.confirmPassword && passwordForm.newPassword !== passwordForm.confirmPassword}
              helperText={
                passwordForm.confirmPassword && passwordForm.newPassword !== passwordForm.confirmPassword
                  ? t('settings.profilePanel.messages.passwordMismatch')
                  : ' '
              }
              InputProps={{
                endAdornment: (
                  <InputAdornment position="end">
                    <IconButton
                      onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                      edge="end"
                      aria-label={
                        showConfirmPassword
                          ? t('settings.profilePanel.passwordDialog.hideConfirm')
                          : t('settings.profilePanel.passwordDialog.showConfirm')
                      }
                    >
                      {showConfirmPassword ? <VisibilityOff /> : <Visibility />}
                    </IconButton>
                  </InputAdornment>
                )
              }}
            />
            <TextField
              fullWidth
              label={t('settings.profilePanel.fields.currentTotpOptional')}
              value={passwordForm.code}
              onChange={(e) => setPasswordForm({ ...passwordForm, code: e.target.value.replace(/\s+/g, '').slice(0, 8) })}
              inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
              helperText={t('settings.profilePanel.fields.currentTotpHelper')}
            />
          </Stack>
        </DialogContent>
        <DialogActions sx={{ px: 3, py: 2 }}>
          <Button onClick={() => setPasswordDialogOpen(false)} color="inherit" disabled={loading}>
            {t('common.cancel')}
          </Button>
          <Button onClick={resetPasswordForm} variant="outlined" disabled={loading}>
            {t('common.reset')}
          </Button>
          <Button variant="contained" onClick={handleChangePassword} disabled={loading} startIcon={<LockIcon />}>
            {t('settings.profilePanel.actions.changePassword')}
          </Button>
        </DialogActions>
      </Dialog>
    </Stack>
  );
}
