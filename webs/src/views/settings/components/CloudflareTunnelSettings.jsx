import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Divider from '@mui/material/Divider';
import FormControlLabel from '@mui/material/FormControlLabel';
import IconButton from '@mui/material/IconButton';
import InputAdornment from '@mui/material/InputAdornment';
import Link from '@mui/material/Link';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

import CloudQueueIcon from '@mui/icons-material/CloudQueue';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import RefreshIcon from '@mui/icons-material/Refresh';
import SaveIcon from '@mui/icons-material/Save';
import StopCircleIcon from '@mui/icons-material/StopCircle';
import VpnLockIcon from '@mui/icons-material/VpnLock';

import { getCloudflaredStatus, removeCloudflaredToken, startCloudflared, stopCloudflared, updateCloudflaredConfig } from 'api/settings';

const defaultStatus = {
  installed: false,
  path: '',
  version: '',
  running: false,
  enabled: false,
  hasToken: false,
  maskedToken: '',
  lastMessage: '',
  lastError: '',
  commandLabel: 'cloudflared tunnel --no-autoupdate run'
};

const cloudflareTunnelDocsUrl = 'https://github.com/ZeroDeng01/sublinkPro/blob/main/docs/features/cloudflare-tunnel.md';
const cloudflaredInstallDocsUrl = `${cloudflareTunnelDocsUrl}#安装-cloudflared`;
const cloudflareTunnelTokenDocsUrl = `${cloudflareTunnelDocsUrl}#第二步复制-tunnel-token`;

export default function CloudflareTunnelSettings({ showMessage }) {
  const { t } = useTranslation();
  const [status, setStatus] = useState(defaultStatus);
  const [token, setToken] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [action, setAction] = useState('');

  const fetchStatus = useCallback(
    async ({ silent = false } = {}) => {
      if (!silent) {
        setLoading(true);
      }
      try {
        const response = await getCloudflaredStatus();
        setStatus({ ...defaultStatus, ...(response.data || {}) });
      } catch (error) {
        showMessage(t('settings.cloudflareTunnel.messages.loadFailed', { message: error.response?.data?.msg || error.message }), 'error');
      } finally {
        if (!silent) {
          setLoading(false);
        }
      }
    },
    [showMessage, t]
  );

  useEffect(() => {
    fetchStatus();
  }, [fetchStatus]);

  useEffect(() => {
    if (!status.running) {
      return undefined;
    }
    const timer = window.setInterval(() => fetchStatus({ silent: true }), 4000);
    return () => window.clearInterval(timer);
  }, [fetchStatus, status.running]);

  const runAction = async (name, handler) => {
    setAction(name);
    try {
      const response = await handler();
      setStatus({ ...defaultStatus, ...(response.data || {}) });
      return response;
    } finally {
      setAction('');
    }
  };

  const handleSave = async (event) => {
    event?.preventDefault();
    setSaving(true);
    try {
      const response = await updateCloudflaredConfig({ enabled: status.enabled, token: token.trim() });
      setStatus({ ...defaultStatus, ...(response.data || {}) });
      setToken('');
      showMessage(t('settings.cloudflareTunnel.messages.saved'));
    } catch (error) {
      showMessage(t('settings.cloudflareTunnel.messages.saveFailed', { message: error.response?.data?.msg || error.message }), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleStart = async (event) => {
    event?.preventDefault();
    try {
      await runAction('start', () => startCloudflared({ token: token.trim() }));
      setToken('');
      showMessage(t('settings.cloudflareTunnel.messages.started'));
    } catch (error) {
      showMessage(t('settings.cloudflareTunnel.messages.startFailed', { message: error.response?.data?.msg || error.message }), 'error');
    }
  };

  const handleStop = async (event) => {
    event?.preventDefault();
    try {
      await runAction('stop', stopCloudflared);
      showMessage(t('settings.cloudflareTunnel.messages.stopped'));
    } catch (error) {
      showMessage(t('settings.cloudflareTunnel.messages.stopFailed', { message: error.response?.data?.msg || error.message }), 'error');
    }
  };

  const handleRemoveToken = async (event) => {
    event?.preventDefault();
    try {
      await runAction('remove', removeCloudflaredToken);
      setToken('');
      showMessage(t('settings.cloudflareTunnel.messages.tokenRemoved'));
    } catch (error) {
      showMessage(t('settings.cloudflareTunnel.messages.removeFailed', { message: error.response?.data?.msg || error.message }), 'error');
    }
  };

  const handleRefresh = async (event) => {
    event?.preventDefault();
    setAction('refresh');
    try {
      await fetchStatus({ silent: true });
    } finally {
      setAction('');
    }
  };

  const canStart = status.installed && !status.running && (status.hasToken || token.trim());

  return (
    <Card variant="outlined">
      <CardHeader
        title="Cloudflare Tunnel"
        subheader={t('settings.cloudflareTunnel.subheader')}
        avatar={<CloudQueueIcon color="primary" />}
        action={
          <Tooltip title={t('settings.cloudflareTunnel.actions.refreshStatus')}>
            <IconButton type="button" onClick={handleRefresh} disabled={loading || Boolean(action)}>
              <RefreshIcon />
            </IconButton>
          </Tooltip>
        }
      />
      <CardContent>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress size={28} />
          </Box>
        ) : (
          <Stack spacing={2.5} sx={{ maxWidth: 760 }}>
            <Alert severity="info">{t('settings.cloudflareTunnel.alerts.setup')}</Alert>

            {!status.installed && (
              <Alert
                severity="warning"
                sx={{ color: (theme) => (theme.palette.mode === 'dark' ? theme.palette.warning.main : theme.palette.warning.dark) }}
              >
                {t('settings.cloudflareTunnel.alerts.notInstalledPrefix')}{' '}
                <Link href={cloudflaredInstallDocsUrl} target="_blank" rel="noreferrer" color="inherit" underline="always">
                  {t('settings.cloudflareTunnel.alerts.installGuide')}
                </Link>
                {t('settings.cloudflareTunnel.alerts.sentenceEnd')}
              </Alert>
            )}

            <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 2, bgcolor: 'background.default' }}>
              <Stack spacing={1.5}>
                <Typography variant="subtitle2">{t('settings.cloudflareTunnel.status.title')}</Typography>
                <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                  <Chip
                    size="small"
                    color={status.installed ? 'success' : 'default'}
                    variant="outlined"
                    label={
                      status.installed
                        ? t('settings.cloudflareTunnel.status.installed')
                        : t('settings.cloudflareTunnel.status.notInstalled')
                    }
                  />
                  <Chip
                    size="small"
                    color={status.running ? 'success' : 'default'}
                    variant="outlined"
                    label={status.running ? t('settings.cloudflareTunnel.status.running') : t('settings.cloudflareTunnel.status.stopped')}
                  />
                  <Chip
                    size="small"
                    color={status.enabled ? 'info' : 'default'}
                    variant="outlined"
                    label={
                      status.enabled
                        ? t('settings.cloudflareTunnel.status.autoEnabled')
                        : t('settings.cloudflareTunnel.status.autoDisabled')
                    }
                  />
                  <Chip
                    size="small"
                    color={status.hasToken ? 'success' : 'default'}
                    variant="outlined"
                    label={
                      status.hasToken
                        ? t('settings.cloudflareTunnel.status.tokenSaved', { token: status.maskedToken })
                        : t('settings.cloudflareTunnel.status.tokenMissing')
                    }
                  />
                </Stack>
                <Typography variant="body2" color="text.secondary">
                  {t('settings.cloudflareTunnel.status.command', { command: status.commandLabel })}
                </Typography>
                {status.path && (
                  <Typography variant="body2" color="text.secondary">
                    {t('settings.cloudflareTunnel.status.path', { path: status.path })}
                  </Typography>
                )}
                {status.version && (
                  <Typography variant="body2" color="text.secondary">
                    {t('settings.cloudflareTunnel.status.version', { version: status.version })}
                  </Typography>
                )}
              </Stack>
            </Box>

            <Stack spacing={2}>
              <FormControlLabel
                sx={{ mr: 0 }}
                control={
                  <Switch
                    checked={status.enabled}
                    disabled={status.running}
                    onChange={(e) => setStatus((prev) => ({ ...prev, enabled: e.target.checked }))}
                  />
                }
                label={status.enabled ? t('settings.cloudflareTunnel.form.autoConnect') : t('settings.cloudflareTunnel.form.noAutoConnect')}
              />

              <TextField
                fullWidth
                type="password"
                label="Cloudflare Tunnel Token"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                placeholder={
                  status.hasToken
                    ? t('settings.cloudflareTunnel.form.tokenPlaceholderSaved')
                    : t('settings.cloudflareTunnel.form.tokenPlaceholderNew')
                }
                helperText={
                  <span>
                    {t('settings.cloudflareTunnel.form.tokenHelperPrefix')}{' '}
                    <Link href={cloudflareTunnelTokenDocsUrl} target="_blank" rel="noreferrer">
                      {t('settings.cloudflareTunnel.form.tokenGuide')}
                    </Link>
                    {t('settings.cloudflareTunnel.alerts.sentenceEnd')}
                  </span>
                }
                disabled={status.running}
                slotProps={{
                  input: {
                    startAdornment: (
                      <InputAdornment position="start">
                        <VpnLockIcon color="action" />
                      </InputAdornment>
                    )
                  }
                }}
              />
            </Stack>

            {(status.lastMessage || status.lastError) && (
              <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 2, bgcolor: 'background.paper' }}>
                <Stack spacing={1}>
                  {status.lastMessage && (
                    <Typography variant="body2" color="text.secondary">
                      {t('settings.cloudflareTunnel.status.lastMessage', { message: status.lastMessage })}
                    </Typography>
                  )}
                  {status.lastError && (
                    <Typography
                      variant="body2"
                      color="text.secondary"
                      sx={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', fontFamily: 'monospace' }}
                    >
                      {t('settings.cloudflareTunnel.status.lastError', { message: status.lastError })}
                    </Typography>
                  )}
                </Stack>
              </Box>
            )}

            <Divider />

            <Stack direction="row" spacing={1.5} useFlexGap flexWrap="wrap">
              <Button
                type="button"
                variant="outlined"
                onClick={handleSave}
                disabled={saving || Boolean(action) || status.running}
                startIcon={<SaveIcon />}
              >
                {saving ? t('settings.cloudflareTunnel.actions.saving') : t('settings.cloudflareTunnel.actions.save')}
              </Button>
              <Button
                type="button"
                variant="contained"
                onClick={handleStart}
                disabled={!canStart || Boolean(action)}
                startIcon={<PlayArrowIcon />}
              >
                {action === 'start' ? t('settings.cloudflareTunnel.actions.starting') : t('settings.cloudflareTunnel.actions.start')}
              </Button>
              <Button
                color="error"
                type="button"
                variant="outlined"
                onClick={handleStop}
                disabled={!status.running || Boolean(action)}
                startIcon={<StopCircleIcon />}
              >
                {action === 'stop' ? t('settings.cloudflareTunnel.actions.stopping') : t('settings.cloudflareTunnel.actions.stop')}
              </Button>
              <Button
                color="error"
                type="button"
                variant="text"
                onClick={handleRemoveToken}
                disabled={!status.hasToken || status.running || Boolean(action)}
                startIcon={<DeleteOutlineIcon />}
              >
                {t('settings.cloudflareTunnel.actions.removeToken')}
              </Button>
            </Stack>
          </Stack>
        )}
      </CardContent>
    </Card>
  );
}
