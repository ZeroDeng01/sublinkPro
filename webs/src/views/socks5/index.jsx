import { useCallback, useEffect, useRef, useState } from 'react';
import { Navigate, useBlocker, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Paper from '@mui/material/Paper';
import Snackbar from '@mui/material/Snackbar';
import Stack from '@mui/material/Stack';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Typography from '@mui/material/Typography';

import AltRouteOutlined from '@mui/icons-material/AltRouteOutlined';
import LanOutlined from '@mui/icons-material/LanOutlined';
import MonitorHeartOutlined from '@mui/icons-material/MonitorHeartOutlined';
import ShieldOutlined from '@mui/icons-material/ShieldOutlined';
import SpaceDashboardOutlined from '@mui/icons-material/SpaceDashboardOutlined';

import { getSocks5Settings, getSocks5Status, startProxyGateway, stopSocks5, updateSocks5Settings } from 'api/proxyGateway';
import ConfirmDialog from 'components/ConfirmDialog';
import { useAuth } from 'contexts/AuthContext';
import MainCard from 'ui-component/cards/MainCard';
import GatewayMonitor from './components/GatewayMonitor';
import GatewayOverview from './components/GatewayOverview';
import GatewaySettingsFields from './components/GatewaySettingsFields';
import RoutingProfiles from './components/RoutingProfiles';
import Socks5Accounts from './components/Socks5Accounts';
import Socks5Listeners from './components/Socks5Listeners';
import { settingsDraft, settingsPayload, validateSettings } from './gatewaySettings';

const tabIcons = {
  overview: SpaceDashboardOutlined,
  listeners: LanOutlined,
  routing: AltRouteOutlined,
  auth: ShieldOutlined,
  monitor: MonitorHeartOutlined
};
const tabs = Object.keys(tabIcons);

function GatewayWorkspace() {
  const { t } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const tab = tabs.includes(searchParams.get('tab')) ? searchParams.get('tab') : 'overview';
  const [snackbar, setSnackbar] = useState(null);
  const [settings, setSettings] = useState(null);
  const [form, setForm] = useState(null);
  const [status, setStatus] = useState(null);
  const [settingsError, setSettingsError] = useState('');
  const [statusError, setStatusError] = useState('');
  const [saveError, setSaveError] = useState('');
  const [pendingApply, setPendingApply] = useState(false);
  const [invalid, setInvalid] = useState(null);
  const [busy, setBusy] = useState('');
  const [stopDialog, setStopDialog] = useState(false);
  const [profileRevision, setProfileRevision] = useState(0);
  const statusRequest = useRef(null);
  const dirty = form && settings && JSON.stringify(form) !== JSON.stringify(settings);
  const pending = Boolean(dirty || pendingApply);
  const label = (key, options) => t(`settings.socks5.page.${key}`, options);

  const showMessage = useCallback((message, severity = 'success') => {
    setSnackbar({ message, severity });
  }, []);

  const loadSettings = useCallback(async () => {
    setSettingsError('');
    try {
      const response = await getSocks5Settings();
      const saved = settingsDraft(response.data);
      setSettings(saved);
      setForm(saved);
    } catch (failure) {
      setSettingsError(failure.message);
    }
  }, []);

  const refreshStatus = useCallback(async (replace = false) => {
    if (statusRequest.current && !replace) return false;
    statusRequest.current?.abort();
    const controller = new AbortController();
    statusRequest.current = controller;
    try {
      const response = await getSocks5Status(controller.signal);
      if (controller.signal.aborted) return false;
      setStatus(response.data);
      setStatusError('');
      return true;
    } catch (failure) {
      if (!controller.signal.aborted) setStatusError(failure.message);
      return false;
    } finally {
      if (statusRequest.current === controller) statusRequest.current = null;
    }
  }, []);

  useEffect(() => {
    void loadSettings();
    void refreshStatus();
    return () => {
      statusRequest.current?.abort();
      statusRequest.current = null;
    };
  }, [loadSettings, refreshStatus]);

  useEffect(() => {
    if (tab !== 'monitor') return undefined;
    const update = () => {
      if (document.visibilityState === 'visible') void refreshStatus();
    };
    const onVisibility = () => {
      if (document.visibilityState === 'visible') update();
      else statusRequest.current?.abort();
    };
    update();
    const timer = window.setInterval(update, 3000);
    document.addEventListener('visibilitychange', onVisibility);
    return () => {
      window.clearInterval(timer);
      document.removeEventListener('visibilitychange', onVisibility);
      statusRequest.current?.abort();
      statusRequest.current = null;
    };
  }, [tab, refreshStatus]);

  const blocker = useBlocker(({ currentLocation, nextLocation }) => pending && currentLocation.pathname !== nextLocation.pathname);
  useEffect(() => {
    if (!pending) return undefined;
    const warn = (event) => {
      event.preventDefault();
      event.returnValue = true;
    };
    window.addEventListener('beforeunload', warn);
    return () => window.removeEventListener('beforeunload', warn);
  }, [pending]);

  const goToTab = (next) => {
    setSearchParams(
      (previous) => {
        const params = new URLSearchParams(previous);
        if (next === 'overview') params.delete('tab');
        else params.set('tab', next);
        return params;
      },
      { preventScrollReset: true }
    );
  };

  const update = (fields) => {
    setForm((previous) => ({ ...previous, ...fields }));
    setInvalid(null);
  };

  const onChanged = useCallback(() => {
    void refreshStatus(true);
  }, [refreshStatus]);

  const save = async () => {
    const failure = validateSettings(form, status?.listeners);
    if (failure) {
      setInvalid(failure);
      if (tab !== failure.tab) goToTab(failure.tab);
      return;
    }
    setBusy('save');
    setSaveError('');
    try {
      const response = await updateSocks5Settings(settingsPayload(form));
      const saved = settingsDraft(response.data);
      setSettings(saved);
      setForm(saved);
      setPendingApply(false);
      setProfileRevision((value) => value + 1);
      showMessage(t('settings.socks5.messages.saved'));
    } catch (failure) {
      const saved = failure.response?.data?.data?.settings;
      if (saved) {
        const savedConfig = settingsDraft(saved);
        setSettings(savedConfig);
        setForm((previous) => ({ ...previous, hasPassword: savedConfig.hasPassword, maskedPassword: savedConfig.maskedPassword }));
        setPendingApply(true);
      }
      setSaveError(failure.message);
    } finally {
      setBusy('');
      onChanged();
    }
  };

  const discard = () => {
    setForm(settings);
    setInvalid(null);
    setPendingApply(false);
    setSaveError('');
  };

  const changeRuntime = async (start) => {
    setBusy(start ? 'start' : 'stop');
    try {
      const response = await (start ? startProxyGateway() : stopSocks5());
      statusRequest.current?.abort();
      statusRequest.current = null;
      setStatus(response.data);
      setStatusError('');
      showMessage(t(start ? 'settings.socks5.api.started' : 'settings.socks5.messages.stopped'));
    } catch (failure) {
      showMessage(failure.message, 'error');
    } finally {
      setBusy('');
      onChanged();
    }
  };

  const fields = (section) => <GatewaySettingsFields section={section} form={form} update={update} error={invalid} busy={Boolean(busy)} />;
  const panels = form && {
    overview: (
      <GatewayOverview
        status={status}
        settings={settings}
        goToTab={goToTab}
        busy={Boolean(busy)}
        showMessage={showMessage}
        onStart={() => changeRuntime(true)}
        onStop={() => setStopDialog(true)}
      />
    ),
    listeners: (
      <Stack spacing={2}>
        {fields('listeners')}
        <Socks5Listeners showMessage={showMessage} onChanged={onChanged} statusListeners={status?.listeners} />
      </Stack>
    ),
    routing: (
      <Stack spacing={2}>
        {fields('routing')}
        <RoutingProfiles key={profileRevision} showMessage={showMessage} onChanged={onChanged} />
      </Stack>
    ),
    auth: (
      <Stack spacing={2}>
        {fields('auth')}
        <Socks5Accounts showMessage={showMessage} onChanged={onChanged} />
      </Stack>
    ),
    monitor: (
      <GatewayMonitor
        status={status}
        statusError={statusError}
        refreshStatus={refreshStatus}
        showMessage={showMessage}
        probeSettings={fields('monitor')}
        focusNodes={invalid?.tab === 'monitor' ? invalid : null}
      />
    )
  };

  return (
    <MainCard
      component="section"
      aria-label={t('settings.socks5.title')}
      title={label(`tabs.${tab}`)}
      sx={{ minWidth: 0, overflow: 'visible' }}
      contentSX={{ px: { xs: 1.5, sm: 3 } }}
      headerSX={{
        px: { xs: 1.5, sm: 3 },
        flexWrap: 'wrap',
        gap: 1,
        '& .MuiCardHeader-action': { m: 0, alignSelf: 'center', minWidth: 0, maxWidth: '100%' }
      }}
      secondary={
        <Stack direction="row" spacing={0.75} useFlexGap flexWrap="wrap" alignItems="center">
          <Stack direction="row" spacing={0.75} useFlexGap flexWrap="wrap" aria-live="polite">
            <Chip
              size="small"
              color={status?.running ? 'success' : 'default'}
              label={label(status ? (status.running ? 'running' : 'stopped') : statusError ? 'statusUnknown' : 'checking')}
            />
            {status && <Chip size="small" color="default" variant="outlined" label={label(status.enabled ? 'enabled' : 'disabled')} />}
          </Stack>
          <Button size="small" onClick={() => refreshStatus(true)}>
            {label('refreshStatus')}
          </Button>
        </Stack>
      }
    >
      <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <Tabs
          value={tab}
          onChange={(_, next) => goToTab(next)}
          variant="scrollable"
          scrollButtons="auto"
          allowScrollButtonsMobile
          aria-label={t('settings.socks5.title')}
          sx={{
            '& .MuiTab-root': {
              minHeight: 48,
              minWidth: 0,
              px: { xs: 1.25, sm: 2 },
              textTransform: 'none',
              fontSize: '0.95rem',
              fontWeight: 500
            }
          }}
        >
          {tabs.map((name) => {
            const Icon = tabIcons[name];
            return (
              <Tab
                key={name}
                value={name}
                id={`gateway-tab-${name}`}
                aria-controls={`gateway-panel-${name}`}
                icon={<Icon fontSize="small" sx={{ display: { xs: 'none', sm: 'inline-flex' } }} />}
                iconPosition="start"
                label={label(`tabs.${name}`)}
              />
            );
          })}
        </Tabs>
      </Box>

      <Stack spacing={2} sx={{ pt: 3 }}>
        {statusError && (
          <Alert
            severity="warning"
            action={
              <Button color="inherit" size="small" onClick={() => refreshStatus(true)}>
                {label('retry')}
              </Button>
            }
          >
            {label(status ? 'statusStale' : 'statusUnavailable', { message: statusError })}
          </Alert>
        )}
        {settingsError && (
          <Alert
            severity="error"
            action={
              <Button color="inherit" size="small" onClick={loadSettings}>
                {label('retry')}
              </Button>
            }
          >
            {settingsError}
          </Alert>
        )}
        {saveError && <Alert severity="error">{saveError}</Alert>}
        {invalid && <Alert severity="error">{t(`settings.socks5.${invalid.message}`, invalid.params)}</Alert>}
        {!form && !settingsError && (
          <CircularProgress size={24} sx={{ alignSelf: 'center', my: 3 }} aria-label={t('settings.socks5.page.loading')} />
        )}

        {tabs.map((name) => (
          <Box
            key={name}
            role="tabpanel"
            id={`gateway-panel-${name}`}
            hidden={tab !== name}
            aria-labelledby={`gateway-tab-${name}`}
            sx={{ minWidth: 0, '&[hidden]': { display: 'none' } }}
          >
            {tab === name && panels?.[name]}
          </Box>
        ))}

        {pending && (
          <Paper variant="outlined" sx={{ position: 'sticky', bottom: 0, zIndex: 2, p: 1.5, pb: 'max(12px, env(safe-area-inset-bottom))' }}>
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} alignItems={{ sm: 'center' }} justifyContent="space-between">
              <Typography variant="body2" color="text.secondary">
                {label(pendingApply ? 'applicationPending' : 'unsaved')}
              </Typography>
              <Stack direction="row" spacing={1} justifyContent="flex-end">
                <Button size="small" disabled={Boolean(busy)} onClick={discard}>
                  {label('discard')}
                </Button>
                <Button size="small" disabled={Boolean(busy)} variant="contained" onClick={save}>
                  {busy === 'save' ? t('settings.socks5.actions.saving') : label('save')}
                </Button>
              </Stack>
            </Stack>
          </Paper>
        )}
      </Stack>

      <ConfirmDialog
        open={stopDialog}
        title={label('stop')}
        content={label('stopConfirm')}
        onClose={() => setStopDialog(false)}
        onConfirm={() => {
          setStopDialog(false);
          return changeRuntime(false);
        }}
      />
      <ConfirmDialog
        open={blocker.state === 'blocked'}
        title={label('leaveTitle')}
        content={label('leaveConfirm')}
        onClose={() => blocker.state === 'blocked' && blocker.reset()}
        onConfirm={() => blocker.proceed?.()}
      />

      <Snackbar
        open={Boolean(snackbar)}
        autoHideDuration={3500}
        onClose={() => setSnackbar(null)}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        {snackbar ? <Alert severity={snackbar.severity}>{snackbar.message}</Alert> : undefined}
      </Snackbar>
    </MainCard>
  );
}

export default function ProxyGatewayPage() {
  const { user } = useAuth();
  const isAdmin = [user?.role, ...(user?.roles || [])].some((role) => String(role || '').toLowerCase() === 'admin');
  return isAdmin ? <GatewayWorkspace /> : <Navigate to="/system/settings" replace />;
}
