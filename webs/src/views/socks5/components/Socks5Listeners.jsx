import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

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
import FormControl from '@mui/material/FormControl';
import FormControlLabel from '@mui/material/FormControlLabel';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import {
  createSocks5Listener,
  deleteSocks5Listener,
  getSocks5Listeners,
  getSocks5RoutingProfiles,
  updateSocks5Listener
} from 'api/settings';

const emptyListener = {
  id: '',
  enabled: true,
  listenAddress: '127.0.0.1',
  port: 1080,
  defaultProfileId: 'default',
  requireAuthMode: 'inherit'
};

function listenerForm(listener) {
  return {
    ...emptyListener,
    ...(listener || {}),
    requireAuthMode: listener?.requireAuth === true ? 'required' : listener?.requireAuth === false ? 'disabled' : 'inherit'
  };
}

function errorMessage(t, error) {
  const response = error.response?.data;
  if (response?.i18nKey) return t(response.i18nKey, response.i18nParams || {});
  return response?.msg || error.message;
}

export default function Socks5Listeners({ showMessage, onChanged }) {
  const { t } = useTranslation();
  const [listeners, setListeners] = useState([]);
  const [profiles, setProfiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingID, setEditingID] = useState('');
  const [form, setForm] = useState(emptyListener);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [listenersResponse, profilesResponse] = await Promise.all([getSocks5Listeners(), getSocks5RoutingProfiles()]);
      setListeners(Array.isArray(listenersResponse.data) ? listenersResponse.data : []);
      setProfiles(Array.isArray(profilesResponse.data) ? profilesResponse.data : []);
    } catch (error) {
      showMessage(errorMessage(t, error), 'error');
    } finally {
      setLoading(false);
    }
  }, [showMessage, t]);

  useEffect(() => {
    load();
  }, [load]);

  const openCreate = () => {
    setEditingID('');
    setForm(listenerForm());
    setDialogOpen(true);
  };
  const openEdit = (listener) => {
    setEditingID(listener.id);
    setForm(listenerForm(listener));
    setDialogOpen(true);
  };

  const save = async () => {
    setSaving(true);
    const payload = {
      id: form.id.trim().toLowerCase(),
      enabled: Boolean(form.enabled),
      listenAddress: form.listenAddress.trim(),
      port: Number(form.port),
      defaultProfileId: form.defaultProfileId,
      requireAuth: form.requireAuthMode === 'inherit' ? null : form.requireAuthMode === 'required'
    };
    try {
      if (editingID) await updateSocks5Listener(editingID, payload);
      else await createSocks5Listener(payload);
      setDialogOpen(false);
      showMessage(t(editingID ? 'settings.socks5.listeners.messages.updated' : 'settings.socks5.listeners.messages.created'));
      await load();
      onChanged?.();
    } catch (error) {
      showMessage(errorMessage(t, error), 'error');
    } finally {
      setSaving(false);
    }
  };

  const remove = async (listener) => {
    if (!window.confirm(t('settings.socks5.listeners.deleteConfirm', { id: listener.id }))) return;
    try {
      await deleteSocks5Listener(listener.id);
      showMessage(t('settings.socks5.listeners.messages.deleted'));
      await load();
      onChanged?.();
    } catch (error) {
      showMessage(errorMessage(t, error), 'error');
    }
  };

  const idValid = /^[a-z0-9][a-z0-9_-]{0,31}$/.test(form.id.trim().toLowerCase());
  const invalid =
    (!editingID && !idValid) || !form.listenAddress.trim() || Number(form.port) < 1 || Number(form.port) > 65535 || !form.defaultProfileId;

  return (
    <>
      <Card variant="outlined" sx={{ mt: 2 }}>
        <CardHeader
          title={t('settings.socks5.listeners.title')}
          subheader={t('settings.socks5.listeners.subheader')}
          action={
            <Button variant="contained" onClick={openCreate} disabled={loading}>
              {t('settings.socks5.listeners.create')}
            </Button>
          }
        />
        <CardContent>
          <Stack spacing={1.25}>
            {listeners.length === 0 && <Typography color="text.secondary">{t('settings.socks5.listeners.empty')}</Typography>}
            {listeners.map((listener) => {
              const profile = profiles.find((item) => item.id === listener.defaultProfileId);
              return (
                <Box key={listener.id} sx={{ p: 1.5, border: 1, borderColor: 'divider', borderRadius: 1.5 }}>
                  <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5} alignItems={{ md: 'center' }}>
                    <Box sx={{ flex: 1, minWidth: 0 }}>
                      <Stack direction="row" spacing={1} alignItems="center" useFlexGap flexWrap="wrap">
                        <Typography variant="subtitle1">{listener.id}</Typography>
                        <Chip
                          size="small"
                          color={listener.running ? 'success' : 'default'}
                          label={t(listener.running ? 'settings.socks5.listeners.running' : 'settings.socks5.listeners.stopped')}
                        />
                        <Chip size="small" label={t(listener.enabled ? 'common.enabled' : 'common.disabled')} />
                        <Chip size="small" color="primary" variant="outlined" label={profile?.name || listener.defaultProfileId} />
                      </Stack>
                      <Typography variant="body2" color="text.secondary">
                        {listener.boundAddress || `${listener.listenAddress}:${listener.port}`} ·{' '}
                        {t(
                          `settings.socks5.listeners.auth.${listener.requireAuth === true ? 'required' : listener.requireAuth === false ? 'disabled' : 'inherit'}`
                        )}
                      </Typography>
                    </Box>
                    <Stack direction="row" spacing={1}>
                      <Button size="small" onClick={() => openEdit(listener)}>
                        {t('common.edit')}
                      </Button>
                      <Button size="small" color="error" onClick={() => remove(listener)}>
                        {t('common.delete')}
                      </Button>
                    </Stack>
                  </Stack>
                </Box>
              );
            })}
          </Stack>
        </CardContent>
      </Card>

      <Dialog open={dialogOpen} onClose={() => !saving && setDialogOpen(false)} fullWidth maxWidth="sm">
        <DialogTitle>{t(editingID ? 'settings.socks5.listeners.editTitle' : 'settings.socks5.listeners.createTitle')}</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2}>
            <TextField
              label={t('settings.socks5.listeners.id')}
              value={form.id}
              disabled={Boolean(editingID)}
              onChange={(event) => setForm((previous) => ({ ...previous, id: event.target.value.toLowerCase() }))}
              helperText={t('settings.socks5.listeners.idHelper')}
              error={!editingID && form.id !== '' && !idValid}
            />
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
              <TextField
                fullWidth
                label={t('settings.socks5.listeners.address')}
                value={form.listenAddress}
                onChange={(event) => setForm((previous) => ({ ...previous, listenAddress: event.target.value }))}
              />
              <TextField
                fullWidth
                type="number"
                label={t('settings.socks5.listeners.port')}
                value={form.port}
                onChange={(event) => setForm((previous) => ({ ...previous, port: event.target.value }))}
                slotProps={{ htmlInput: { min: 1, max: 65535 } }}
              />
            </Stack>
            <FormControl fullWidth>
              <InputLabel>{t('settings.socks5.listeners.profile')}</InputLabel>
              <Select
                value={form.defaultProfileId}
                label={t('settings.socks5.listeners.profile')}
                onChange={(event) => setForm((previous) => ({ ...previous, defaultProfileId: event.target.value }))}
              >
                {profiles
                  .filter((profile) => profile.enabled)
                  .map((profile) => (
                    <MenuItem key={profile.id} value={profile.id}>
                      {profile.name}
                    </MenuItem>
                  ))}
              </Select>
            </FormControl>
            <FormControl fullWidth>
              <InputLabel>{t('settings.socks5.listeners.authPolicy')}</InputLabel>
              <Select
                value={form.requireAuthMode}
                label={t('settings.socks5.listeners.authPolicy')}
                onChange={(event) => setForm((previous) => ({ ...previous, requireAuthMode: event.target.value }))}
              >
                <MenuItem value="inherit">{t('settings.socks5.listeners.auth.inherit')}</MenuItem>
                <MenuItem value="required">{t('settings.socks5.listeners.auth.required')}</MenuItem>
                <MenuItem value="disabled">{t('settings.socks5.listeners.auth.disabled')}</MenuItem>
              </Select>
            </FormControl>
            <FormControlLabel
              control={
                <Switch
                  checked={Boolean(form.enabled)}
                  onChange={(event) => setForm((previous) => ({ ...previous, enabled: event.target.checked }))}
                />
              }
              label={t('settings.socks5.listeners.enabled')}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)} disabled={saving}>
            {t('common.cancel')}
          </Button>
          <Button variant="contained" onClick={save} disabled={saving || invalid}>
            {saving ? t('common.saving') : t('common.save')}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}
