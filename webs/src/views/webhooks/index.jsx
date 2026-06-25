import { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { alpha, useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';

import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Grid from '@mui/material/Grid';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Alert from '@mui/material/Alert';
import TextField from '@mui/material/TextField';
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import CircularProgress from '@mui/material/CircularProgress';
import Snackbar from '@mui/material/Snackbar';

import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import RefreshIcon from '@mui/icons-material/Refresh';
import SaveIcon from '@mui/icons-material/Save';
import SendIcon from '@mui/icons-material/Send';
import WebhookIcon from '@mui/icons-material/Webhook';

import MainCard from 'ui-component/cards/MainCard';
import NotificationEventSelector from 'views/settings/components/NotificationEventSelector';
import { createWebhook, deleteWebhook, getWebhooks, testWebhookById, updateWebhook } from 'api/settings';
import { formatDateTime as formatLocalizedDateTime } from 'i18n/locales';

const createDefaultForm = () => ({
  id: null,
  name: '',
  url: '',
  method: 'POST',
  contentType: 'application/json',
  headers: '',
  body: '',
  enabled: true,
  eventKeys: []
});

const pickValue = (...values) => values.find((value) => value !== undefined && value !== null);

const normalizeWebhook = (item = {}) => ({
  id: item.id,
  name: item.name || item.title || item.remark || '',
  url: item.webhookUrl || item.url || '',
  method: item.webhookMethod || item.method || 'POST',
  contentType: item.webhookContentType || item.contentType || 'application/json',
  headers: item.webhookHeaders || item.headers || '',
  body: item.webhookBody || item.body || '',
  enabled: Boolean(pickValue(item.webhookEnabled, item.enabled, false)),
  eventKeys: Array.isArray(item.eventKeys) ? item.eventKeys : [],
  createdAt: item.createdAt,
  updatedAt: item.updatedAt,
  lastTestAt: item.lastTestAt
});

const toWebhookPayload = (form) => ({
  name: form.name.trim(),
  webhookUrl: form.url.trim(),
  webhookMethod: form.method,
  webhookContentType: form.contentType,
  webhookHeaders: form.headers,
  webhookBody: form.body,
  webhookEnabled: form.enabled,
  eventKeys: form.eventKeys
});

const formatWebhookDateTime = (value, language, fallback) => {
  if (!value) return fallback;

  const formatted = formatLocalizedDateTime(value, language, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  });

  return formatted || value;
};

const validateJsonText = (value) => {
  if (!value.trim()) return true;

  try {
    JSON.parse(value);
    return true;
  } catch {
    return false;
  }
};

export default function WebhookManagementPage() {
  const theme = useTheme();
  const { i18n, t } = useTranslation();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  const [items, setItems] = useState([]);
  const [eventOptions, setEventOptions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [testingId, setTestingId] = useState(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState(null);
  const [form, setForm] = useState(createDefaultForm());
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });
  const palette = theme.vars?.palette || theme.palette;
  const isDark = theme.palette.mode === 'dark';

  const showMessage = useCallback((message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  }, []);

  const fetchWebhooks = useCallback(async () => {
    setLoading(true);
    try {
      const response = await getWebhooks();
      const data = response.data || {};
      setItems((data.items || []).map(normalizeWebhook));
      setEventOptions(data.eventOptions || []);
    } catch (error) {
      showMessage(error.response?.data?.message || error.message || t('webhooks.messages.fetchFailed'), 'error');
    } finally {
      setLoading(false);
    }
  }, [showMessage, t]);

  useEffect(() => {
    fetchWebhooks();
  }, [fetchWebhooks]);

  const enabledCount = useMemo(() => items.filter((item) => item.enabled).length, [items]);
  const getStatusChipSx = (enabled) => ({
    height: 20,
    fontSize: '0.72rem',
    fontWeight: 600,
    color: enabled ? (isDark ? palette.success.light : palette.success.dark) : palette.text.secondary,
    bgcolor: enabled
      ? alpha(theme.palette.success.main, isDark ? 0.18 : 0.12)
      : isDark
        ? 'background.default'
        : alpha(theme.palette.grey[500], 0.08),
    border: `1px solid ${enabled ? alpha(theme.palette.success.main, isDark ? 0.34 : 0.18) : alpha(theme.palette.divider, isDark ? 0.56 : 0.24)}`,
    '& .MuiChip-label': {
      px: 1
    }
  });

  const openCreateDialog = () => {
    setForm(createDefaultForm());
    setDialogOpen(true);
  };

  const openEditDialog = (item) => {
    setForm(normalizeWebhook(item));
    setDialogOpen(true);
  };

  const closeDialog = () => {
    if (submitting) return;
    setDialogOpen(false);
    setForm(createDefaultForm());
  };

  const handleSave = async () => {
    if (!form.url.trim()) {
      showMessage(t('webhooks.messages.urlRequired'), 'warning');
      return;
    }

    if (!validateJsonText(form.headers)) {
      showMessage(t('webhooks.messages.headersInvalid'), 'warning');
      return;
    }

    if (form.contentType === 'application/json' && form.body.trim() && !validateJsonText(form.body)) {
      showMessage(t('webhooks.messages.bodyInvalid'), 'warning');
      return;
    }

    setSubmitting(true);
    try {
      const payload = toWebhookPayload(form);
      if (form.id) {
        await updateWebhook(form.id, payload);
        showMessage(t('webhooks.messages.updated'));
      } else {
        await createWebhook(payload);
        showMessage(t('webhooks.messages.created'));
      }
      setDialogOpen(false);
      setForm(createDefaultForm());
      await fetchWebhooks();
    } catch (error) {
      showMessage(error.response?.data?.message || error.message || t('webhooks.messages.saveFailed'), 'error');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget?.id) return;

    setSubmitting(true);
    try {
      await deleteWebhook(deleteTarget.id);
      showMessage(t('webhooks.messages.deleted'));
      setDeleteTarget(null);
      await fetchWebhooks();
    } catch (error) {
      showMessage(error.response?.data?.message || error.message || t('webhooks.messages.deleteFailed'), 'error');
    } finally {
      setSubmitting(false);
    }
  };

  const handleToggleEnabled = async (item, enabled) => {
    try {
      await updateWebhook(item.id, toWebhookPayload({ ...normalizeWebhook(item), enabled }));
      setItems((prev) => prev.map((current) => (current.id === item.id ? { ...current, enabled } : current)));
      showMessage(enabled ? t('webhooks.messages.enabled') : t('webhooks.messages.disabled'));
    } catch (error) {
      showMessage(error.response?.data?.message || error.message || t('webhooks.messages.toggleFailed'), 'error');
    }
  };

  const handleTest = async (item) => {
    setTestingId(item.id);
    try {
      const response = await testWebhookById(item.id);
      showMessage(response.data?.message || t('webhooks.messages.testSuccess'));
      await fetchWebhooks();
    } catch (error) {
      showMessage(error.response?.data?.message || error.message || t('webhooks.messages.testFailed'), 'error');
    } finally {
      setTestingId(null);
    }
  };

  return (
    <MainCard title={t('webhooks.title')}>
      <Stack spacing={3}>
        <Card variant="outlined" sx={{ borderRadius: 3 }}>
          <CardContent>
            <Stack spacing={2.5}>
              <Stack direction={{ xs: 'column', md: 'row' }} spacing={2} alignItems={{ md: 'center' }} justifyContent="space-between">
                <Box>
                  <Typography variant="h4" sx={{ mb: 0.75 }}>
                    {t('webhooks.title')}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    {t('webhooks.description')}
                  </Typography>
                </Box>

                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.25} sx={{ width: { xs: '100%', md: 'auto' } }}>
                  <Button
                    variant="outlined"
                    onClick={fetchWebhooks}
                    disabled={loading}
                    startIcon={loading ? <CircularProgress size={16} /> : <RefreshIcon />}
                    sx={{ width: { xs: '100%', sm: 'auto' } }}
                  >
                    {t('common.refresh')}
                  </Button>
                  <Button variant="contained" onClick={openCreateDialog} startIcon={<AddIcon />} sx={{ width: { xs: '100%', sm: 'auto' } }}>
                    {t('webhooks.createWebhook')}
                  </Button>
                </Stack>
              </Stack>

              <Alert severity="info" variant="outlined">
                {t('webhooks.splitAlert')}
              </Alert>

              <Grid container spacing={2}>
                <Grid size={{ xs: 12, sm: 6, md: 4 }}>
                  <Box
                    sx={{
                      p: 2,
                      borderRadius: 2,
                      border: '1px solid',
                      borderColor: 'divider',
                      backgroundColor: alpha(theme.palette.primary.main, 0.04)
                    }}
                  >
                    <Typography variant="caption" color="text.secondary">
                      {t('webhooks.totalWebhooks')}
                    </Typography>
                    <Typography variant="h3" sx={{ mt: 0.5 }}>
                      {items.length}
                    </Typography>
                  </Box>
                </Grid>
                <Grid size={{ xs: 12, sm: 6, md: 4 }}>
                  <Box
                    sx={{
                      p: 2,
                      borderRadius: 2,
                      border: '1px solid',
                      borderColor: 'divider',
                      backgroundColor: alpha(theme.palette.success.main, 0.05)
                    }}
                  >
                    <Typography variant="caption" color="text.secondary">
                      {t('webhooks.enabledCount')}
                    </Typography>
                    <Typography variant="h3" sx={{ mt: 0.5 }}>
                      {enabledCount}
                    </Typography>
                  </Box>
                </Grid>
                <Grid size={{ xs: 12, md: 4 }}>
                  <Box
                    sx={{
                      p: 2,
                      borderRadius: 2,
                      border: '1px solid',
                      borderColor: 'divider',
                      backgroundColor: alpha(theme.palette.warning.main, 0.05)
                    }}
                  >
                    <Typography variant="caption" color="text.secondary">
                      {t('webhooks.optionalEvents')}
                    </Typography>
                    <Typography variant="h3" sx={{ mt: 0.5 }}>
                      {eventOptions.length}
                    </Typography>
                  </Box>
                </Grid>
              </Grid>
            </Stack>
          </CardContent>
        </Card>

        {loading ? (
          <Box sx={{ py: 6, display: 'flex', justifyContent: 'center' }}>
            <CircularProgress />
          </Box>
        ) : items.length === 0 ? (
          <Alert
            severity="info"
            variant="outlined"
            action={
              <Button color="inherit" size="small" onClick={openCreateDialog} startIcon={<AddIcon />}>
                {t('webhooks.createNow')}
              </Button>
            }
          >
            {t('webhooks.emptyTitle')}
          </Alert>
        ) : (
          <Grid container spacing={2.5}>
            {items.map((item, index) => {
              const title = item.name || `${t('webhooks.defaultName', { index: index + 1 })}`;
              const hasEvents = item.eventKeys.length > 0;

              return (
                <Grid size={{ xs: 12, lg: 6 }} key={item.id || `${item.url}-${index}`}>
                  <Card
                    variant="outlined"
                    sx={{
                      height: '100%',
                      borderRadius: 3,
                      transition: 'all 0.2s ease',
                      '&:hover': {
                        borderColor: theme.palette.primary.main,
                        boxShadow: theme.shadows[3]
                      }
                    }}
                  >
                    <CardHeader
                      avatar={
                        <Box
                          sx={{
                            width: 40,
                            height: 40,
                            borderRadius: 2,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            backgroundColor: alpha(theme.palette.primary.main, 0.1),
                            color: 'primary.main'
                          }}
                        >
                          <WebhookIcon fontSize="small" />
                        </Box>
                      }
                      title={
                        <Typography variant="h5" sx={{ wordBreak: 'break-word' }}>
                          {title}
                        </Typography>
                      }
                      subheader={
                        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5, wordBreak: 'break-all' }}>
                          {item.url || t('webhooks.noUrl')}
                        </Typography>
                      }
                      sx={{ pb: 1.5 }}
                    />

                    <CardContent sx={{ pt: 0 }}>
                      <Stack spacing={2}>
                        <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                          <Chip
                            size="small"
                            variant="filled"
                            label={item.enabled ? t('webhooks.statusEnabled') : t('webhooks.statusDisabled')}
                            sx={getStatusChipSx(item.enabled)}
                          />
                          <Chip size="small" variant="outlined" label={item.method} />
                          <Chip size="small" variant="outlined" label={item.contentType} />
                          <Chip
                            size="small"
                            variant="outlined"
                            color={hasEvents ? 'primary' : 'default'}
                            label={t('webhooks.eventsCount', { count: item.eventKeys.length })}
                          />
                        </Stack>

                        {!hasEvents && (
                          <Alert severity="warning" variant="outlined">
                            {t('webhooks.noEventsAlert')}
                          </Alert>
                        )}

                        <Box>
                          <Typography variant="body2" color="text.secondary">
                            {t('webhooks.headersLabel')}
                            {item.headers?.trim() ? t('webhooks.headersConfigured') : t('webhooks.headersUnconfigured')}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            {t('webhooks.bodyLabel')}
                            {item.body?.trim() ? t('webhooks.bodyConfigured') : t('webhooks.bodyUnconfigured')}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            {t('webhooks.lastUpdated')}
                            {formatWebhookDateTime(
                              item.updatedAt || item.createdAt,
                              i18n.resolvedLanguage || i18n.language,
                              t('webhooks.unrecorded')
                            )}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            {t('webhooks.lastTested')}
                            {formatWebhookDateTime(item.lastTestAt, i18n.resolvedLanguage || i18n.language, t('webhooks.unrecorded'))}
                          </Typography>
                        </Box>

                        <FormControlLabel
                          sx={{ ml: 0 }}
                          control={
                            <Switch
                              checked={item.enabled}
                              onChange={(event) => handleToggleEnabled(item, event.target.checked)}
                              disabled={submitting || testingId === item.id}
                            />
                          }
                          label={item.enabled ? t('webhooks.switchEnabled') : t('webhooks.switchDisabled')}
                        />

                        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.25}>
                          <Button
                            variant="outlined"
                            startIcon={<EditIcon />}
                            onClick={() => openEditDialog(item)}
                            sx={{ width: { xs: '100%', sm: 'auto' } }}
                          >
                            {t('webhooks.edit')}
                          </Button>
                          <Button
                            variant="outlined"
                            color="success"
                            startIcon={testingId === item.id ? <CircularProgress size={16} /> : <SendIcon />}
                            onClick={() => handleTest(item)}
                            disabled={submitting || testingId === item.id || !item.id}
                            sx={{ width: { xs: '100%', sm: 'auto' } }}
                          >
                            {t('webhooks.test')}
                          </Button>
                          <Button
                            variant="outlined"
                            color="error"
                            startIcon={<DeleteIcon />}
                            onClick={() => setDeleteTarget(item)}
                            disabled={submitting || testingId === item.id}
                            sx={{ width: { xs: '100%', sm: 'auto' } }}
                          >
                            {t('webhooks.delete')}
                          </Button>
                        </Stack>
                      </Stack>
                    </CardContent>
                  </Card>
                </Grid>
              );
            })}
          </Grid>
        )}
      </Stack>

      <Dialog open={dialogOpen} onClose={closeDialog} maxWidth="md" fullWidth fullScreen={isMobile}>
        <DialogTitle>{form.id ? t('webhooks.editWebhook') : t('webhooks.createWebhook')}</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2.5} sx={{ pt: 1 }}>
            <Alert severity="info" variant="outlined">
              {t('webhooks.dialog.infoAlert')}
            </Alert>

            <TextField
              fullWidth
              label={t('webhooks.dialog.nameLabel')}
              value={form.name}
              onChange={(event) => setForm((prev) => ({ ...prev, name: event.target.value }))}
              placeholder={t('webhooks.dialog.namePlaceholder')}
              helperText={t('webhooks.dialog.nameHelperText')}
            />

            <FormControlLabel
              control={
                <Switch checked={form.enabled} onChange={(event) => setForm((prev) => ({ ...prev, enabled: event.target.checked }))} />
              }
              label={form.enabled ? t('webhooks.dialog.enableNow') : t('webhooks.dialog.enableLater')}
            />

            <TextField
              fullWidth
              required
              label={t('webhooks.dialog.urlLabel')}
              value={form.url}
              onChange={(event) => setForm((prev) => ({ ...prev, url: event.target.value }))}
              placeholder="https://example.com/webhook"
              helperText={t('webhooks.dialog.urlHelperText')}
            />

            <Grid container spacing={2}>
              <Grid size={{ xs: 12, sm: 4 }}>
                <FormControl fullWidth>
                  <InputLabel>{t('webhooks.dialog.methodLabel')}</InputLabel>
                  <Select
                    value={form.method}
                    label={t('webhooks.dialog.methodLabel')}
                    onChange={(event) => setForm((prev) => ({ ...prev, method: event.target.value }))}
                  >
                    <MenuItem value="POST">POST</MenuItem>
                    <MenuItem value="GET">GET</MenuItem>
                    <MenuItem value="PUT">PUT</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
              <Grid size={{ xs: 12, sm: 8 }}>
                <FormControl fullWidth>
                  <InputLabel>{t('webhooks.dialog.contentTypeLabel')}</InputLabel>
                  <Select
                    value={form.contentType}
                    label={t('webhooks.dialog.contentTypeLabel')}
                    onChange={(event) => setForm((prev) => ({ ...prev, contentType: event.target.value }))}
                  >
                    <MenuItem value="application/json">application/json</MenuItem>
                    <MenuItem value="application/x-www-form-urlencoded">application/x-www-form-urlencoded</MenuItem>
                    <MenuItem value="text/plain">text/plain</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
            </Grid>

            <TextField
              fullWidth
              multiline
              minRows={4}
              label={t('webhooks.dialog.headersInputLabel')}
              value={form.headers}
              onChange={(event) => setForm((prev) => ({ ...prev, headers: event.target.value }))}
              placeholder={'{\n  "Authorization": "Bearer token"\n}'}
              helperText={t('webhooks.dialog.headersHelperText')}
            />

            <TextField
              fullWidth
              multiline
              minRows={6}
              label={t('webhooks.dialog.bodyInputLabel')}
              value={form.body}
              onChange={(event) => setForm((prev) => ({ ...prev, body: event.target.value }))}
              placeholder={
                form.contentType === 'application/json'
                  ? '{\n  "title": "{{title}}",\n  "message": "{{message}}"\n}'
                  : '{{title}} - {{message}}'
              }
              helperText={t('webhooks.dialog.bodyHelperText')}
            />

            <Alert severity="info" variant="outlined">
              {t('webhooks.dialog.variablesAlert', {
                variables: '{{title}}、{{message}}、{{event}}、{{eventName}}、{{category}}、{{severity}}、{{time}}、{{json .}}'
              })}
            </Alert>

            <NotificationEventSelector
              value={form.eventKeys}
              eventOptions={eventOptions}
              disabled={submitting}
              description={t('webhooks.dialog.eventSelectorDesc')}
              onChange={(eventKeys) => setForm((prev) => ({ ...prev, eventKeys }))}
            />
          </Stack>
        </DialogContent>
        <DialogActions sx={{ px: 3, py: 2 }}>
          <Button onClick={closeDialog} disabled={submitting} color="inherit">
            {t('common.cancel')}
          </Button>
          <Button
            variant="contained"
            onClick={handleSave}
            disabled={submitting}
            startIcon={submitting ? <CircularProgress size={16} /> : <SaveIcon />}
          >
            {form.id ? t('webhooks.dialog.save') : t('webhooks.dialog.create')}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={Boolean(deleteTarget)}
        onClose={() => {
          if (!submitting) setDeleteTarget(null);
        }}
        maxWidth="xs"
        fullWidth
      >
        <DialogTitle>{t('webhooks.deleteWebhook')}</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="text.secondary">
            {t('webhooks.deleteConfirm', { name: deleteTarget?.name || deleteTarget?.url || 'Webhook' })}
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteTarget(null)} disabled={submitting} color="inherit">
            {t('common.cancel')}
          </Button>
          <Button variant="contained" color="error" onClick={handleDelete} disabled={submitting} startIcon={<DeleteIcon />}>
            {t('webhooks.confirmDelete')}
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar((prev) => ({ ...prev, open: false }))}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity} variant="filled" onClose={() => setSnackbar((prev) => ({ ...prev, open: false }))}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </MainCard>
  );
}
