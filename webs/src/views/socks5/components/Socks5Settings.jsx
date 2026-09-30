import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import CircularProgress from '@mui/material/CircularProgress';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormHelperText from '@mui/material/FormHelperText';

import HubIcon from '@mui/icons-material/Hub';
import RefreshIcon from '@mui/icons-material/Refresh';
import SaveIcon from '@mui/icons-material/Save';
import StopCircleIcon from '@mui/icons-material/StopCircle';

import { getSocks5Settings, stopSocks5, updateSocks5Settings } from 'api/settings';

import CandidateNodePool from './CandidateNodePool';
import SpecificNodeSelector from './SpecificNodeSelector';

const defaultConfig = {
  enabled: false,
  listenAddress: '127.0.0.1',
  port: 1080,
  username: '',
  hasPassword: false,
  maskedPassword: '',
  nodeId: 0,
  selection: 'best',
  requireAuth: true,
  maxAttempts: 1,
  dialTimeoutSeconds: 30,
  failureCooldownSeconds: 0,
  specificFallback: false,
  maxConnections: 256,
  maxConnectionsPerClient: 32,
  idleTimeoutSeconds: 600,
  maxConnectionDurationSeconds: 0,
  stickySessionEnabled: false,
  stickySessionMode: 'client_ip',
  stickySessionTtlSeconds: 1800,
  healthCheckEnabled: false,
  healthCheckIntervalSeconds: 60,
  healthCheckTimeoutSeconds: 5,
  candidateGroups: [],
  candidateSources: [],
  candidateProtocols: [],
  candidateCountries: [],
  running: false,
  boundAddress: ''
};

function messageFromError(t, error, fallbackKey) {
  const response = error.response?.data;
  if (response?.i18nKey) return t(response.i18nKey, response.i18nParams || {});
  return t(fallbackKey, { message: response?.msg || error.message });
}

export default function Socks5Settings({ showMessage, onChanged }) {
  const { t } = useTranslation();
  const [config, setConfig] = useState(defaultConfig);
  const [form, setForm] = useState({ ...defaultConfig, password: '', clearPassword: false });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [stopping, setStopping] = useState(false);

  const syncConfig = useCallback((next) => {
    const value = { ...defaultConfig, ...(next || {}) };
    setConfig(value);
    setForm({ ...value, password: '', clearPassword: false });
  }, []);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const settingsResponse = await getSocks5Settings();
      syncConfig(settingsResponse.data);
    } catch (error) {
      showMessage(messageFromError(t, error, 'settings.socks5.messages.loadFailed'), 'error');
    } finally {
      setLoading(false);
    }
  }, [showMessage, syncConfig, t]);

  useEffect(() => {
    load();
  }, [load]);

  const handleSave = async () => {
    setSaving(true);
    try {
      const response = await updateSocks5Settings({
        enabled: Boolean(form.enabled),
        listenAddress: form.listenAddress.trim(),
        port: Number(form.port) || 1080,
        username: form.username.trim(),
        password: form.password,
        clearPassword: Boolean(form.clearPassword),
        nodeId: Number(form.nodeId) || 0,
        selection: form.selection,
        requireAuth: Boolean(form.requireAuth),
        maxAttempts: Number(form.maxAttempts) || 1,
        dialTimeoutSeconds: Number(form.dialTimeoutSeconds) || 30,
        failureCooldownSeconds: Number(form.failureCooldownSeconds) || 0,
        specificFallback: Boolean(form.specificFallback),
        maxConnections: Number(form.maxConnections) || 256,
        maxConnectionsPerClient: Number(form.maxConnectionsPerClient) || 32,
        idleTimeoutSeconds: Number(form.idleTimeoutSeconds) || 0,
        maxConnectionDurationSeconds: Number(form.maxConnectionDurationSeconds) || 0,
        stickySessionEnabled: Boolean(form.stickySessionEnabled),
        stickySessionMode: form.stickySessionMode,
        stickySessionTtlSeconds: Number(form.stickySessionTtlSeconds) || 1800,
        healthCheckEnabled: Boolean(form.healthCheckEnabled),
        healthCheckIntervalSeconds: Number(form.healthCheckIntervalSeconds) || 60,
        healthCheckTimeoutSeconds: Number(form.healthCheckTimeoutSeconds) || 5,
        candidateGroups: Array.isArray(form.candidateGroups) ? form.candidateGroups : [],
        candidateSources: Array.isArray(form.candidateSources) ? form.candidateSources : [],
        candidateProtocols: Array.isArray(form.candidateProtocols) ? form.candidateProtocols : [],
        candidateCountries: Array.isArray(form.candidateCountries) ? form.candidateCountries : []
      });
      syncConfig(response.data);
      showMessage(t('settings.socks5.messages.saved'));
      onChanged?.();
    } catch (error) {
      showMessage(messageFromError(t, error, 'settings.socks5.messages.saveFailed'), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleStop = async () => {
    setStopping(true);
    try {
      const response = await stopSocks5();
      syncConfig(response.data);
      showMessage(t('settings.socks5.messages.stopped'));
    } catch (error) {
      showMessage(messageFromError(t, error, 'settings.socks5.messages.stopFailed'), 'error');
    } finally {
      setStopping(false);
    }
  };

  const busy = loading || saving || stopping;
  const specificInvalid = form.selection === 'specific' && !Number(form.nodeId);
  const authInvalid = form.enabled && form.requireAuth && (!form.username.trim() || (!form.password && !config.hasPassword));
  const routingInvalid =
    Number(form.maxAttempts) < 1 ||
    Number(form.maxAttempts) > 5 ||
    Number(form.dialTimeoutSeconds) < 1 ||
    Number(form.dialTimeoutSeconds) > 120 ||
    Number(form.failureCooldownSeconds) < 0 ||
    Number(form.failureCooldownSeconds) > 3600 ||
    (form.selection === 'specific' && form.specificFallback && Number(form.maxAttempts) < 2) ||
    Number(form.maxConnections) < 1 ||
    Number(form.maxConnections) > 10000 ||
    Number(form.maxConnectionsPerClient) < 1 ||
    Number(form.maxConnectionsPerClient) > Number(form.maxConnections) ||
    Number(form.idleTimeoutSeconds) < 0 ||
    Number(form.idleTimeoutSeconds) > 86400 ||
    Number(form.maxConnectionDurationSeconds) < 0 ||
    Number(form.maxConnectionDurationSeconds) > 604800 ||
    (form.stickySessionEnabled &&
      (Number(form.stickySessionTtlSeconds) < 60 ||
        Number(form.stickySessionTtlSeconds) > 604800 ||
        (form.stickySessionMode === 'username' && !form.requireAuth))) ||
    Number(form.healthCheckIntervalSeconds) < 10 ||
    Number(form.healthCheckIntervalSeconds) > 3600 ||
    Number(form.healthCheckTimeoutSeconds) < 1 ||
    Number(form.healthCheckTimeoutSeconds) > 30;

  return (
    <Card variant="outlined">
      <CardHeader avatar={<HubIcon color="primary" />} title={t('settings.socks5.title')} subheader={t('settings.socks5.subheader')} />
      <CardContent>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress size={28} />
          </Box>
        ) : (
          <Stack spacing={2.5}>
            <Alert severity="warning">{t('settings.socks5.alerts.security')}</Alert>
            <Alert severity={config.running ? 'success' : 'info'}>
              {config.running
                ? t('settings.socks5.status.running', { address: config.boundAddress || `${config.listenAddress}:${config.port}` })
                : t('settings.socks5.status.stopped')}
            </Alert>
            <FormControlLabel
              control={
                <Switch
                  checked={Boolean(form.enabled)}
                  onChange={(event) => setForm((prev) => ({ ...prev, enabled: event.target.checked }))}
                />
              }
              label={t('settings.socks5.form.enabled')}
            />
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <TextField
                fullWidth
                label={t('settings.socks5.form.listenAddress')}
                value={form.listenAddress}
                onChange={(event) => setForm((prev) => ({ ...prev, listenAddress: event.target.value }))}
                helperText={t('settings.socks5.form.listenAddressHelper')}
              />
              <TextField
                fullWidth
                type="number"
                label={t('settings.socks5.form.port')}
                value={form.port}
                onChange={(event) => setForm((prev) => ({ ...prev, port: event.target.value }))}
                slotProps={{ htmlInput: { min: 1, max: 65535 } }}
              />
            </Stack>
            <FormControl fullWidth>
              <InputLabel>{t('settings.socks5.form.selection')}</InputLabel>
              <Select
                label={t('settings.socks5.form.selection')}
                value={form.selection}
                onChange={(event) => setForm((prev) => ({ ...prev, selection: event.target.value }))}
              >
                <MenuItem value="best">{t('settings.socks5.form.selectionBest')}</MenuItem>
                <MenuItem value="random">{t('settings.socks5.form.selectionRandom')}</MenuItem>
                <MenuItem value="round_robin">{t('settings.socks5.form.selectionRoundRobin')}</MenuItem>
                <MenuItem value="smart">{t('settings.socks5.form.selectionSmart')}</MenuItem>
                <MenuItem value="specific">{t('settings.socks5.form.selectionSpecific')}</MenuItem>
              </Select>
            </FormControl>
            {form.selection === 'specific' && (
              <SpecificNodeSelector
                value={Number(form.nodeId) || 0}
                onChange={(nodeId) => setForm((prev) => ({ ...prev, nodeId }))}
                disabled={busy}
                error={specificInvalid}
              />
            )}
            {form.selection === 'specific' && (
              <Box>
                <FormControlLabel
                  control={
                    <Switch
                      checked={Boolean(form.specificFallback)}
                      onChange={(event) =>
                        setForm((prev) => ({
                          ...prev,
                          specificFallback: event.target.checked,
                          maxAttempts: event.target.checked ? Math.max(Number(prev.maxAttempts) || 1, 2) : prev.maxAttempts
                        }))
                      }
                    />
                  }
                  label={t('settings.socks5.form.specificFallback')}
                />
                <FormHelperText>{t('settings.socks5.form.specificFallbackHelper')}</FormHelperText>
              </Box>
            )}
            {(form.selection !== 'specific' || form.specificFallback) && (
              <CandidateNodePool value={form} onChange={(pool) => setForm((prev) => ({ ...prev, ...pool }))} disabled={busy} />
            )}
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <TextField
                fullWidth
                type="number"
                label={t('settings.socks5.form.maxAttempts')}
                value={form.maxAttempts}
                onChange={(event) => setForm((prev) => ({ ...prev, maxAttempts: event.target.value }))}
                helperText={t('settings.socks5.form.maxAttemptsHelper')}
                slotProps={{ htmlInput: { min: 1, max: 5 } }}
              />
              <TextField
                fullWidth
                type="number"
                label={t('settings.socks5.form.dialTimeoutSeconds')}
                value={form.dialTimeoutSeconds}
                onChange={(event) => setForm((prev) => ({ ...prev, dialTimeoutSeconds: event.target.value }))}
                helperText={t('settings.socks5.form.dialTimeoutSecondsHelper')}
                slotProps={{ htmlInput: { min: 1, max: 120 } }}
              />
              <TextField
                fullWidth
                type="number"
                label={t('settings.socks5.form.failureCooldownSeconds')}
                value={form.failureCooldownSeconds}
                onChange={(event) => setForm((prev) => ({ ...prev, failureCooldownSeconds: event.target.value }))}
                helperText={t('settings.socks5.form.failureCooldownSecondsHelper')}
                slotProps={{ htmlInput: { min: 0, max: 3600 } }}
              />
            </Stack>
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <TextField
                fullWidth
                type="number"
                label={t('settings.socks5.form.maxConnections')}
                value={form.maxConnections}
                onChange={(event) => setForm((prev) => ({ ...prev, maxConnections: event.target.value }))}
                helperText={t('settings.socks5.form.maxConnectionsHelper')}
                slotProps={{ htmlInput: { min: 1, max: 10000 } }}
              />
              <TextField
                fullWidth
                type="number"
                label={t('settings.socks5.form.maxConnectionsPerClient')}
                value={form.maxConnectionsPerClient}
                onChange={(event) => setForm((prev) => ({ ...prev, maxConnectionsPerClient: event.target.value }))}
                helperText={t('settings.socks5.form.maxConnectionsPerClientHelper')}
                slotProps={{ htmlInput: { min: 1, max: 10000 } }}
              />
              <TextField
                fullWidth
                type="number"
                label={t('settings.socks5.form.idleTimeoutSeconds')}
                value={form.idleTimeoutSeconds}
                onChange={(event) => setForm((prev) => ({ ...prev, idleTimeoutSeconds: event.target.value }))}
                helperText={t('settings.socks5.form.idleTimeoutSecondsHelper')}
                slotProps={{ htmlInput: { min: 0, max: 86400 } }}
              />
              <TextField
                fullWidth
                type="number"
                label={t('settings.socks5.form.maxConnectionDurationSeconds')}
                value={form.maxConnectionDurationSeconds}
                onChange={(event) => setForm((prev) => ({ ...prev, maxConnectionDurationSeconds: event.target.value }))}
                helperText={t('settings.socks5.form.maxConnectionDurationSecondsHelper')}
                slotProps={{ htmlInput: { min: 0, max: 604800 } }}
              />
            </Stack>
            <Box>
              <FormControlLabel
                control={
                  <Switch
                    checked={Boolean(form.stickySessionEnabled)}
                    onChange={(event) => setForm((prev) => ({ ...prev, stickySessionEnabled: event.target.checked }))}
                    disabled={form.selection === 'specific'}
                  />
                }
                label={t('settings.socks5.form.stickySessionEnabled')}
              />
              <FormHelperText>
                {form.selection === 'specific'
                  ? t('settings.socks5.form.stickySessionSpecificHelper')
                  : t('settings.socks5.form.stickySessionEnabledHelper')}
              </FormHelperText>
            </Box>
            {form.stickySessionEnabled && form.selection !== 'specific' && (
              <Stack spacing={1.25}>
                <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                  <FormControl fullWidth>
                    <InputLabel>{t('settings.socks5.form.stickySessionMode')}</InputLabel>
                    <Select
                      label={t('settings.socks5.form.stickySessionMode')}
                      value={form.stickySessionMode}
                      onChange={(event) => setForm((prev) => ({ ...prev, stickySessionMode: event.target.value }))}
                    >
                      <MenuItem value="client_ip">{t('settings.socks5.form.stickySessionModeClientIp')}</MenuItem>
                      <MenuItem value="username">{t('settings.socks5.form.stickySessionModeUsername')}</MenuItem>
                    </Select>
                  </FormControl>
                  <TextField
                    fullWidth
                    type="number"
                    label={t('settings.socks5.form.stickySessionTtlSeconds')}
                    value={form.stickySessionTtlSeconds}
                    onChange={(event) => setForm((prev) => ({ ...prev, stickySessionTtlSeconds: event.target.value }))}
                    helperText={t('settings.socks5.form.stickySessionTtlSecondsHelper')}
                    slotProps={{ htmlInput: { min: 60, max: 604800 } }}
                  />
                </Stack>
                {form.stickySessionMode === 'username' && (
                  <Alert severity={form.requireAuth ? 'info' : 'error'}>
                    {t(
                      form.requireAuth
                        ? 'settings.socks5.form.stickySessionUsernameHelper'
                        : 'settings.socks5.form.stickySessionUsernameAuthRequired'
                    )}
                  </Alert>
                )}
              </Stack>
            )}
            <Box>
              <FormControlLabel
                control={
                  <Switch
                    checked={Boolean(form.healthCheckEnabled)}
                    onChange={(event) => setForm((prev) => ({ ...prev, healthCheckEnabled: event.target.checked }))}
                  />
                }
                label={t('settings.socks5.form.healthCheckEnabled')}
              />
              <FormHelperText>{t('settings.socks5.form.healthCheckEnabledHelper')}</FormHelperText>
            </Box>
            {form.healthCheckEnabled && (
              <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                <TextField
                  fullWidth
                  type="number"
                  label={t('settings.socks5.form.healthCheckIntervalSeconds')}
                  value={form.healthCheckIntervalSeconds}
                  onChange={(event) => setForm((prev) => ({ ...prev, healthCheckIntervalSeconds: event.target.value }))}
                  helperText={t('settings.socks5.form.healthCheckIntervalSecondsHelper')}
                  slotProps={{ htmlInput: { min: 10, max: 3600 } }}
                />
                <TextField
                  fullWidth
                  type="number"
                  label={t('settings.socks5.form.healthCheckTimeoutSeconds')}
                  value={form.healthCheckTimeoutSeconds}
                  onChange={(event) => setForm((prev) => ({ ...prev, healthCheckTimeoutSeconds: event.target.value }))}
                  helperText={t('settings.socks5.form.healthCheckTimeoutSecondsHelper')}
                  slotProps={{ htmlInput: { min: 1, max: 30 } }}
                />
              </Stack>
            )}
            <FormControlLabel
              control={
                <Switch
                  checked={Boolean(form.requireAuth)}
                  onChange={(event) => setForm((prev) => ({ ...prev, requireAuth: event.target.checked }))}
                />
              }
              label={t('settings.socks5.form.requireAuth')}
            />
            {form.requireAuth && (
              <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                <TextField
                  fullWidth
                  label={t('settings.socks5.form.username')}
                  value={form.username}
                  onChange={(event) => setForm((prev) => ({ ...prev, username: event.target.value }))}
                />
                <TextField
                  fullWidth
                  type="password"
                  label={t('settings.socks5.form.password')}
                  value={form.password}
                  onChange={(event) => setForm((prev) => ({ ...prev, password: event.target.value, clearPassword: false }))}
                  placeholder={config.maskedPassword || ''}
                  disabled={form.clearPassword}
                  helperText={config.hasPassword ? t('settings.socks5.form.passwordSaved') : t('settings.socks5.form.passwordHelper')}
                />
              </Stack>
            )}
            {config.hasPassword && form.requireAuth && (
              <FormControlLabel
                control={
                  <Switch
                    checked={form.clearPassword}
                    onChange={(event) => setForm((prev) => ({ ...prev, clearPassword: event.target.checked, password: '' }))}
                  />
                }
                label={t('settings.socks5.form.clearPassword')}
              />
            )}
            {authInvalid && <Alert severity="error">{t('settings.socks5.messages.authRequired')}</Alert>}
            <Stack direction="row" spacing={1.5} useFlexGap flexWrap="wrap">
              <Button
                variant="contained"
                startIcon={<SaveIcon />}
                onClick={handleSave}
                disabled={busy || specificInvalid || authInvalid || routingInvalid}
              >
                {saving ? t('settings.socks5.actions.saving') : t('settings.socks5.actions.save')}
              </Button>
              <Button variant="outlined" startIcon={<RefreshIcon />} onClick={load} disabled={busy}>
                {t('settings.socks5.actions.refresh')}
              </Button>
              <Button
                color="error"
                variant="outlined"
                startIcon={<StopCircleIcon />}
                onClick={handleStop}
                disabled={busy || !config.running}
              >
                {stopping ? t('settings.socks5.actions.stopping') : t('settings.socks5.actions.stop')}
              </Button>
            </Stack>
          </Stack>
        )}
      </CardContent>
    </Card>
  );
}
