import { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Checkbox from '@mui/material/Checkbox';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Divider from '@mui/material/Divider';
import FormControl from '@mui/material/FormControl';
import FormControlLabel from '@mui/material/FormControlLabel';
import IconButton from '@mui/material/IconButton';
import InputLabel from '@mui/material/InputLabel';
import ListItemText from '@mui/material/ListItemText';
import MenuItem from '@mui/material/MenuItem';
import OutlinedInput from '@mui/material/OutlinedInput';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

import ExtensionIcon from '@mui/icons-material/Extension';
import RefreshIcon from '@mui/icons-material/Refresh';
import SaveIcon from '@mui/icons-material/Save';
import ScienceIcon from '@mui/icons-material/Science';

import { getSubStoreSettings, testSubStoreSettings, updateSubStoreSettings } from 'api/settings';

const defaultSettings = {
  configured: false,
  supportedTargets: [],
  enabled: { value: false, source: 'database' },
  baseUrl: { value: '', source: 'database' },
  timeoutSeconds: { value: 10, source: 'default' },
  allowedTargets: { value: [], source: 'default' },
  maxResponseBytes: { value: 8388608, source: 'default' }
};

const targetLabels = {
  loon: 'Loon',
  egern: 'Egern',
  stash: 'Stash',
  surfboard: 'Surfboard',
  shadowrocket: 'Shadowrocket',
  quanx: 'Quantumult X',
  'sing-box': 'sing-box',
  uri: 'URI',
  json: 'JSON'
};

function localizedApiMessage(t, error, fallbackKey) {
  const response = error.response?.data;
  if (response?.i18nKey) {
    return t(response.i18nKey, response.i18nParams || {});
  }
  return t(fallbackKey, { message: response?.msg || error.message });
}

function fieldValue(field, fallback) {
  return field?.value ?? fallback;
}

export default function SubStoreSettings({ showMessage }) {
  const { t } = useTranslation();
  const [settings, setSettings] = useState(defaultSettings);
  const [form, setForm] = useState({
    enabled: false,
    baseUrl: '',
    timeoutSeconds: 10,
    allowedTargets: [],
    maxResponseBytes: 8388608
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);

  const supportedTargets = useMemo(() => settings.supportedTargets || [], [settings.supportedTargets]);

  const syncForm = useCallback((nextSettings) => {
    setSettings({ ...defaultSettings, ...nextSettings });
    setForm({
      enabled: fieldValue(nextSettings.enabled, false),
      baseUrl: fieldValue(nextSettings.baseUrl, ''),
      timeoutSeconds: fieldValue(nextSettings.timeoutSeconds, 10),
      allowedTargets: fieldValue(nextSettings.allowedTargets, []),
      maxResponseBytes: fieldValue(nextSettings.maxResponseBytes, 8388608)
    });
  }, []);

  const fetchSettings = useCallback(
    async ({ silent = false } = {}) => {
      if (!silent) {
        setLoading(true);
      }
      try {
        const response = await getSubStoreSettings();
        syncForm(response.data || defaultSettings);
      } catch (error) {
        showMessage(localizedApiMessage(t, error, 'settings.subStore.messages.loadFailed'), 'error');
      } finally {
        if (!silent) {
          setLoading(false);
        }
      }
    },
    [showMessage, syncForm, t]
  );

  useEffect(() => {
    fetchSettings();
  }, [fetchSettings]);

  const payload = useCallback(
    () => ({
      enabled: form.enabled,
      baseUrl: form.baseUrl.trim(),
      timeoutSeconds: Number(form.timeoutSeconds) || 0,
      allowedTargets: form.allowedTargets,
      maxResponseBytes: Number(form.maxResponseBytes) || 0
    }),
    [form]
  );

  const handleSave = async (event) => {
    event?.preventDefault();
    setSaving(true);
    try {
      const response = await updateSubStoreSettings(payload());
      syncForm(response.data || defaultSettings);
      showMessage(t('settings.subStore.messages.saved'));
    } catch (error) {
      showMessage(localizedApiMessage(t, error, 'settings.subStore.messages.saveFailed'), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleTest = async (event) => {
    event?.preventDefault();
    setTesting(true);
    try {
      const response = await testSubStoreSettings(payload());
      showMessage(
        t('settings.subStore.messages.testSuccess', {
          target: response.data?.target || form.allowedTargets[0] || 'loon',
          bytes: response.data?.resultBytes ?? 0
        })
      );
    } catch (error) {
      showMessage(localizedApiMessage(t, error, 'settings.subStore.messages.testFailed'), 'error');
    } finally {
      setTesting(false);
    }
  };

  const targetLabel = (target) => targetLabels[target] || target;

  return (
    <Card variant="outlined">
      <CardHeader
        title="Sub-Store"
        subheader={t('settings.subStore.subheader')}
        avatar={<ExtensionIcon color="primary" />}
        action={
          <Tooltip title={t('settings.subStore.actions.refresh')}>
            <IconButton type="button" onClick={() => fetchSettings({ silent: true })} disabled={loading || saving || testing}>
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
          <Stack spacing={2.5} sx={{ maxWidth: 820 }}>
            <Alert
              severity="info"
              sx={{
                alignItems: 'flex-start',
                '& .MuiAlert-message': {
                  overflow: 'visible',
                  overflowWrap: 'anywhere',
                  wordBreak: 'break-word'
                }
              }}
            >
              {t('settings.subStore.alerts.sidecar')}
            </Alert>

            <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 2, bgcolor: 'background.default' }}>
              <Stack spacing={1.5}>
                <Typography variant="subtitle2">{t('settings.subStore.status.title')}</Typography>
                <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                  <Chip
                    size="small"
                    color={settings.configured ? 'success' : 'default'}
                    variant="outlined"
                    label={settings.configured ? t('settings.subStore.status.configured') : t('settings.subStore.status.notConfigured')}
                  />
                  <Chip
                    size="small"
                    color={form.enabled ? 'info' : 'default'}
                    variant="outlined"
                    label={form.enabled ? t('settings.subStore.status.enabled') : t('settings.subStore.status.disabled')}
                  />
                </Stack>
                <Typography variant="body2" color="text.secondary">
                  {t('settings.subStore.status.targets', { count: form.allowedTargets.length })}
                </Typography>
              </Stack>
            </Box>

            <FormControlLabel
              sx={{ mr: 0 }}
              control={
                <Switch checked={form.enabled} onChange={(event) => setForm((prev) => ({ ...prev, enabled: event.target.checked }))} />
              }
              label={form.enabled ? t('settings.subStore.form.enabled') : t('settings.subStore.form.disabled')}
            />

            <TextField
              fullWidth
              label={t('settings.subStore.form.baseUrl')}
              value={form.baseUrl}
              onChange={(event) => setForm((prev) => ({ ...prev, baseUrl: event.target.value }))}
              placeholder="http://sub-store:3000"
              helperText={t('settings.subStore.form.baseUrlHelper')}
            />

            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <TextField
                fullWidth
                type="number"
                label={t('settings.subStore.form.timeoutSeconds')}
                value={form.timeoutSeconds}
                onChange={(event) => setForm((prev) => ({ ...prev, timeoutSeconds: event.target.value }))}
                helperText={t('settings.subStore.form.timeoutHelper')}
                slotProps={{ htmlInput: { min: 1 } }}
              />
              <TextField
                fullWidth
                type="number"
                label={t('settings.subStore.form.maxResponseBytes')}
                value={form.maxResponseBytes}
                onChange={(event) => setForm((prev) => ({ ...prev, maxResponseBytes: event.target.value }))}
                helperText={t('settings.subStore.form.maxResponseHelper')}
                slotProps={{ htmlInput: { min: 1 } }}
              />
            </Stack>

            <FormControl fullWidth>
              <InputLabel id="substore-targets-label">{t('settings.subStore.form.allowedTargets')}</InputLabel>
              <Select
                labelId="substore-targets-label"
                multiple
                value={form.allowedTargets}
                onChange={(event) => setForm((prev) => ({ ...prev, allowedTargets: event.target.value }))}
                input={<OutlinedInput label={t('settings.subStore.form.allowedTargets')} />}
                renderValue={(selected) => selected.map(targetLabel).join(', ')}
              >
                {supportedTargets.map((target) => (
                  <MenuItem key={target} value={target}>
                    <Checkbox checked={form.allowedTargets.includes(target)} />
                    <ListItemText primary={targetLabel(target)} secondary={target} />
                  </MenuItem>
                ))}
              </Select>
              <Typography variant="caption" color="text.secondary" sx={{ mt: 0.75 }}>
                {t('settings.subStore.form.allowedTargetsHelper')}
              </Typography>
            </FormControl>

            <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 2, bgcolor: 'background.paper' }}>
              <Typography variant="body2" color="text.secondary">
                {t('settings.subStore.form.bridgeNote')}
              </Typography>
            </Box>

            <Divider />

            <Stack direction="row" spacing={1.5} useFlexGap flexWrap="wrap">
              <Button type="button" variant="outlined" onClick={handleSave} disabled={saving || testing} startIcon={<SaveIcon />}>
                {saving ? t('settings.subStore.actions.saving') : t('settings.subStore.actions.save')}
              </Button>
              <Button type="button" variant="contained" onClick={handleTest} disabled={saving || testing} startIcon={<ScienceIcon />}>
                {testing ? t('settings.subStore.actions.testing') : t('settings.subStore.actions.test')}
              </Button>
            </Stack>
          </Stack>
        )}
      </CardContent>
    </Card>
  );
}
