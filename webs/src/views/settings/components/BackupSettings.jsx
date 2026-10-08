import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Checkbox from '@mui/material/Checkbox';
import Chip from '@mui/material/Chip';
import Collapse from '@mui/material/Collapse';
import CircularProgress from '@mui/material/CircularProgress';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Divider from '@mui/material/Divider';
import FormControlLabel from '@mui/material/FormControlLabel';
import IconButton from '@mui/material/IconButton';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

import BackupIcon from '@mui/icons-material/Backup';
import CloudDownloadIcon from '@mui/icons-material/CloudDownload';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import RefreshIcon from '@mui/icons-material/Refresh';
import SaveIcon from '@mui/icons-material/Save';
import ScienceIcon from '@mui/icons-material/Science';

import {
  getWebDAVBackupSettings,
  listWebDAVBackups,
  restoreWebDAVBackup,
  testWebDAVBackup,
  updateWebDAVBackupSettings,
  uploadWebDAVBackup
} from 'api/settings';
import CronExpressionGenerator, { validateCronExpression } from 'components/CronExpressionGenerator';
import { useTaskProgress } from 'contexts/TaskProgressContext';

const defaultConfig = {
  configured: false,
  baseUrl: '',
  username: '',
  hasPassword: false,
  maskedPassword: '',
  remotePath: 'SublinkPro',
  timeoutSeconds: 60,
  allowInsecureHttp: false,
  allowPrivateNetwork: false,
  scheduleEnabled: false,
  cronExpr: '0 3 * * *'
};

function messageFromError(t, error, fallbackKey) {
  const response = error.response?.data;
  if (response?.i18nKey) {
    return t(response.i18nKey, response.i18nParams || {});
  }
  return t(fallbackKey, { message: response?.msg || error.message });
}

function formatBytes(size) {
  const value = Number(size) || 0;
  if (value < 1024) return `${value} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KiB`;
  return `${(value / 1024 / 1024).toFixed(2)} MiB`;
}

export default function BackupSettings({ showMessage }) {
  const { t, i18n } = useTranslation();
  const { registerOnComplete, unregisterOnComplete } = useTaskProgress();
  const [config, setConfig] = useState(defaultConfig);
  const [form, setForm] = useState({ ...defaultConfig, password: '', clearPassword: false });
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [listing, setListing] = useState(false);
  const [restoreFile, setRestoreFile] = useState(null);
  const [restoring, setRestoring] = useState(false);
  const [restoreTaskId, setRestoreTaskId] = useState('');
  const [includeAccessKeys, setIncludeAccessKeys] = useState(true);
  const [includeSubLogs, setIncludeSubLogs] = useState(false);
  const [confirmOverwrite, setConfirmOverwrite] = useState(false);

  const syncConfig = useCallback((next) => {
    const value = { ...defaultConfig, ...(next || {}) };
    setConfig(value);
    setForm({ ...value, password: '', clearPassword: false });
  }, []);

  const loadSettings = useCallback(async () => {
    setLoading(true);
    try {
      const response = await getWebDAVBackupSettings();
      syncConfig(response.data);
    } catch (error) {
      showMessage(messageFromError(t, error, 'settings.backup.messages.loadFailed'), 'error');
    } finally {
      setLoading(false);
    }
  }, [showMessage, syncConfig, t]);

  const loadFiles = useCallback(
    async ({ silent = false } = {}) => {
      if (!silent) setListing(true);
      try {
        const response = await listWebDAVBackups();
        setFiles(response.data || []);
      } catch (error) {
        showMessage(messageFromError(t, error, 'settings.backup.messages.listFailed'), 'error');
      } finally {
        if (!silent) setListing(false);
      }
    },
    [showMessage, t]
  );

  useEffect(() => {
    loadSettings();
  }, [loadSettings]);

  useEffect(() => {
    if (config.configured) loadFiles({ silent: true });
  }, [config.configured, loadFiles]);

  useEffect(() => {
    const handleComplete = ({ taskId, taskType, status }) => {
      if (taskType !== 'db_migration' || !restoreTaskId || taskId !== restoreTaskId) return;
      if (status === 'completed') {
        showMessage(t('settings.backup.messages.restoreCompleted'));
        setRestoring(false);
        setRestoreTaskId('');
      } else if (status === 'error' || status === 'cancelled') {
        showMessage(t('settings.backup.messages.restoreTaskFailed'), 'error');
        setRestoring(false);
        setRestoreTaskId('');
      }
    };
    registerOnComplete(handleComplete);
    return () => unregisterOnComplete(handleComplete);
  }, [registerOnComplete, restoreTaskId, showMessage, t, unregisterOnComplete]);

  const payload = () => ({
    baseUrl: form.baseUrl.trim(),
    username: form.username.trim(),
    password: form.password,
    clearPassword: Boolean(form.clearPassword),
    remotePath: form.remotePath.trim(),
    timeoutSeconds: Number(form.timeoutSeconds) || 60,
    allowInsecureHttp: Boolean(form.allowInsecureHttp),
    allowPrivateNetwork: Boolean(form.allowPrivateNetwork),
    scheduleEnabled: Boolean(form.scheduleEnabled),
    cronExpr: form.cronExpr || ''
  });

  const handleSave = async () => {
    setSaving(true);
    try {
      const response = await updateWebDAVBackupSettings(payload());
      syncConfig(response.data);
      showMessage(t('settings.backup.messages.saved'));
    } catch (error) {
      showMessage(messageFromError(t, error, 'settings.backup.messages.saveFailed'), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleTest = async () => {
    setTesting(true);
    try {
      const response = await testWebDAVBackup(payload());
      showMessage(t('settings.backup.messages.testSuccess', { latency: response.data?.latencyMs ?? 0 }));
    } catch (error) {
      showMessage(messageFromError(t, error, 'settings.backup.messages.testFailed'), 'error');
    } finally {
      setTesting(false);
    }
  };

  const handleUpload = async () => {
    setUploading(true);
    try {
      const response = await uploadWebDAVBackup();
      showMessage(t('settings.backup.messages.uploadSuccess', { name: response.data?.name || '' }));
      await loadFiles({ silent: true });
    } catch (error) {
      showMessage(messageFromError(t, error, 'settings.backup.messages.uploadFailed'), 'error');
    } finally {
      setUploading(false);
    }
  };

  const openRestore = (file) => {
    setRestoreFile(file);
    setConfirmOverwrite(false);
  };

  const handleRestore = async () => {
    if (!confirmOverwrite || !restoreFile) return;
    setRestoring(true);
    try {
      const response = await restoreWebDAVBackup({ filename: restoreFile.name, includeAccessKeys, includeSubLogs });
      setRestoreTaskId(response.data?.taskId || '');
      showMessage(t('settings.backup.messages.restoreStarted'));
      setRestoreFile(null);
    } catch (error) {
      setRestoring(false);
      showMessage(messageFromError(t, error, 'settings.backup.messages.restoreFailed'), 'error');
    }
  };

  const busy = saving || testing || uploading || restoring;
  const scheduleInvalid = Boolean(form.scheduleEnabled) && !validateCronExpression(form.cronExpr || '');

  return (
    <Card variant="outlined">
      <CardHeader
        avatar={<BackupIcon color="primary" />}
        title={t('settings.backup.title')}
        subheader={t('settings.backup.subheader')}
        action={
          <Tooltip title={t('settings.backup.actions.refreshSettings')}>
            <IconButton onClick={loadSettings} disabled={loading || busy}>
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
          <Stack spacing={2.5}>
            <Alert severity="warning">{t('settings.backup.alerts.sensitive')}</Alert>
            {form.allowPrivateNetwork && <Alert severity="warning">{t('settings.backup.alerts.privateNetwork')}</Alert>}
            {form.allowInsecureHttp && <Alert severity="error">{t('settings.backup.alerts.insecureHttp')}</Alert>}

            <Stack direction="row" spacing={1} alignItems="center" useFlexGap flexWrap="wrap">
              <Typography variant="body2" color="text.secondary">
                {t('settings.backup.status.label')}
              </Typography>
              <Chip
                size="small"
                color={config.configured ? 'success' : 'default'}
                variant="outlined"
                label={config.configured ? t('settings.backup.status.configured') : t('settings.backup.status.notConfigured')}
              />
              {config.hasPassword && <Chip size="small" variant="outlined" label={t('settings.backup.status.passwordSaved')} />}
              {config.scheduleEnabled && (
                <Chip size="small" color="info" variant="outlined" label={t('settings.backup.status.scheduleEnabled')} />
              )}
            </Stack>

            <TextField
              fullWidth
              label={t('settings.backup.form.baseUrl')}
              placeholder="https://dav.example.com/remote.php/dav/files/user"
              value={form.baseUrl}
              onChange={(event) => setForm((prev) => ({ ...prev, baseUrl: event.target.value }))}
              helperText={t('settings.backup.form.baseUrlHelper')}
            />
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <TextField
                fullWidth
                label={t('settings.backup.form.username')}
                value={form.username}
                onChange={(event) => setForm((prev) => ({ ...prev, username: event.target.value }))}
              />
              <TextField
                fullWidth
                type="password"
                label={t('settings.backup.form.password')}
                value={form.password}
                onChange={(event) => setForm((prev) => ({ ...prev, password: event.target.value, clearPassword: false }))}
                placeholder={config.maskedPassword || ''}
                disabled={form.clearPassword}
                helperText={config.hasPassword ? t('settings.backup.form.passwordSavedHelper') : t('settings.backup.form.passwordHelper')}
              />
            </Stack>
            {config.hasPassword && (
              <FormControlLabel
                control={
                  <Checkbox
                    checked={form.clearPassword}
                    onChange={(event) => setForm((prev) => ({ ...prev, clearPassword: event.target.checked, password: '' }))}
                  />
                }
                label={t('settings.backup.form.clearPassword')}
              />
            )}
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <TextField
                fullWidth
                label={t('settings.backup.form.remotePath')}
                value={form.remotePath}
                onChange={(event) => setForm((prev) => ({ ...prev, remotePath: event.target.value }))}
                helperText={t('settings.backup.form.remotePathHelper')}
              />
              <TextField
                fullWidth
                type="number"
                label={t('settings.backup.form.timeoutSeconds')}
                value={form.timeoutSeconds}
                onChange={(event) => setForm((prev) => ({ ...prev, timeoutSeconds: event.target.value }))}
                slotProps={{ htmlInput: { min: 1, max: 600 } }}
              />
            </Stack>
            <FormControlLabel
              control={
                <Switch
                  checked={form.allowPrivateNetwork}
                  onChange={(event) => setForm((prev) => ({ ...prev, allowPrivateNetwork: event.target.checked }))}
                />
              }
              label={t('settings.backup.form.allowPrivateNetwork')}
            />
            <FormControlLabel
              control={
                <Switch
                  checked={form.allowInsecureHttp}
                  onChange={(event) => setForm((prev) => ({ ...prev, allowInsecureHttp: event.target.checked }))}
                />
              }
              label={t('settings.backup.form.allowInsecureHttp')}
            />
            <FormControlLabel
              control={
                <Switch
                  checked={form.scheduleEnabled}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      scheduleEnabled: event.target.checked,
                      cronExpr: event.target.checked && !prev.cronExpr ? '0 3 * * *' : prev.cronExpr
                    }))
                  }
                />
              }
              label={t('settings.backup.form.scheduleEnabled')}
            />
            <Collapse in={Boolean(form.scheduleEnabled)}>
              <Stack spacing={1.5}>
                <Typography variant="body2" color="text.secondary">
                  {t('settings.backup.form.scheduleHelper')}
                </Typography>
                <CronExpressionGenerator
                  value={form.cronExpr || '0 3 * * *'}
                  onChange={(value) => setForm((prev) => ({ ...prev, cronExpr: value }))}
                  label={t('settings.backup.form.cronExpr')}
                  error={scheduleInvalid}
                />
              </Stack>
            </Collapse>

            <Stack direction="row" spacing={1.5} useFlexGap flexWrap="wrap">
              <Button variant="outlined" startIcon={<SaveIcon />} onClick={handleSave} disabled={busy || scheduleInvalid}>
                {saving ? t('settings.backup.actions.saving') : t('settings.backup.actions.save')}
              </Button>
              <Button variant="outlined" startIcon={<ScienceIcon />} onClick={handleTest} disabled={busy || scheduleInvalid}>
                {testing ? t('settings.backup.actions.testing') : t('settings.backup.actions.test')}
              </Button>
              <Button variant="contained" startIcon={<CloudUploadIcon />} onClick={handleUpload} disabled={busy || !config.configured}>
                {uploading ? t('settings.backup.actions.uploading') : t('settings.backup.actions.upload')}
              </Button>
            </Stack>

            <Divider />

            <Stack direction="row" alignItems="center" justifyContent="space-between">
              <Typography variant="h4">{t('settings.backup.files.title')}</Typography>
              <Button size="small" startIcon={<RefreshIcon />} onClick={() => loadFiles()} disabled={listing || busy || !config.configured}>
                {listing ? t('settings.backup.actions.refreshing') : t('settings.backup.actions.refreshFiles')}
              </Button>
            </Stack>
            {files.length === 0 ? (
              <Alert severity="info">{t('settings.backup.files.empty')}</Alert>
            ) : (
              <Stack spacing={1}>
                {files.map((file) => (
                  <Box
                    key={file.name}
                    sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 1.5, bgcolor: 'background.paper' }}
                  >
                    <Stack
                      direction={{ xs: 'column', sm: 'row' }}
                      spacing={1.5}
                      justifyContent="space-between"
                      alignItems={{ sm: 'center' }}
                    >
                      <Box sx={{ minWidth: 0 }}>
                        <Typography variant="body2" fontWeight={600} sx={{ overflowWrap: 'anywhere' }}>
                          {file.name}
                        </Typography>
                        <Typography variant="caption" color="text.secondary">
                          {formatBytes(file.size)} · {file.modifiedAt ? new Date(file.modifiedAt).toLocaleString(i18n.language) : '-'}
                        </Typography>
                      </Box>
                      <Button
                        color="error"
                        size="small"
                        startIcon={<CloudDownloadIcon />}
                        onClick={() => openRestore(file)}
                        disabled={busy}
                      >
                        {t('settings.backup.actions.restore')}
                      </Button>
                    </Stack>
                  </Box>
                ))}
              </Stack>
            )}
          </Stack>
        )}
      </CardContent>

      <Dialog open={Boolean(restoreFile)} onClose={() => !restoring && setRestoreFile(null)} fullWidth maxWidth="sm">
        <DialogTitle>{t('settings.backup.restore.title')}</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <Alert severity="error">{t('settings.backup.restore.warning', { name: restoreFile?.name || '' })}</Alert>
            <FormControlLabel
              control={<Checkbox checked={includeAccessKeys} onChange={(event) => setIncludeAccessKeys(event.target.checked)} />}
              label={t('settings.backup.restore.includeAccessKeys')}
            />
            <FormControlLabel
              control={<Checkbox checked={includeSubLogs} onChange={(event) => setIncludeSubLogs(event.target.checked)} />}
              label={t('settings.backup.restore.includeSubLogs')}
            />
            <FormControlLabel
              control={
                <Checkbox color="error" checked={confirmOverwrite} onChange={(event) => setConfirmOverwrite(event.target.checked)} />
              }
              label={<Typography color="error.main">{t('settings.backup.restore.confirm')}</Typography>}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setRestoreFile(null)} disabled={restoring}>
            {t('settings.backup.actions.cancel')}
          </Button>
          <Button color="error" variant="contained" onClick={handleRestore} disabled={!confirmOverwrite || restoring}>
            {restoring ? t('settings.backup.actions.restoring') : t('settings.backup.actions.confirmRestore')}
          </Button>
        </DialogActions>
      </Dialog>
    </Card>
  );
}
