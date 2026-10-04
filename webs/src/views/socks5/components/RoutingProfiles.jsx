import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import Alert from '@mui/material/Alert';
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
import FormHelperText from '@mui/material/FormHelperText';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import {
  createSocks5RoutingProfile,
  deleteSocks5RoutingProfile,
  getSocks5RoutingProfiles,
  getSocks5Settings,
  updateSocks5RoutingProfile
} from 'api/settings';

import CandidateNodePool from './CandidateNodePool';
import SpecificNodeSelector from './SpecificNodeSelector';

const emptyProfile = {
  id: '',
  name: '',
  enabled: true,
  selection: 'smart',
  nodeId: 0,
  maxAttempts: 3,
  dialTimeoutSeconds: 30,
  failureCooldownSeconds: 0,
  specificFallback: false,
  candidateGroups: [],
  candidateSources: [],
  candidateProtocols: [],
  candidateCountries: [],
  stickySessionEnabled: true,
  stickySessionTtlSeconds: 1800
};

function profileForm(profile) {
  return { ...emptyProfile, ...(profile || {}) };
}

function errorMessage(t, error) {
  const response = error.response?.data;
  if (response?.i18nKey) return t(response.i18nKey, response.i18nParams || {});
  return response?.msg || error.message;
}

export default function RoutingProfiles({ showMessage, onChanged }) {
  const { t } = useTranslation();
  const [profiles, setProfiles] = useState([]);
  const [baseUsername, setBaseUsername] = useState('proxy');
  const [authenticationEnabled, setAuthenticationEnabled] = useState(true);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingID, setEditingID] = useState('');
  const [form, setForm] = useState(emptyProfile);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [profilesResponse, settingsResponse] = await Promise.all([getSocks5RoutingProfiles(), getSocks5Settings()]);
      setProfiles(Array.isArray(profilesResponse.data) ? profilesResponse.data : []);
      setBaseUsername(settingsResponse.data?.username || 'proxy');
      setAuthenticationEnabled(Boolean(settingsResponse.data?.requireAuth));
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
    setForm(profileForm());
    setDialogOpen(true);
  };

  const openEdit = (profile) => {
    setEditingID(profile.id);
    setForm(profileForm(profile));
    setDialogOpen(true);
  };

  const closeDialog = () => {
    if (!saving) setDialogOpen(false);
  };

  const save = async () => {
    setSaving(true);
    const payload = {
      id: form.id.trim().toLowerCase(),
      name: form.name.trim(),
      enabled: Boolean(form.enabled),
      selection: form.selection,
      nodeId: Number(form.nodeId) || 0,
      maxAttempts: Number(form.maxAttempts) || 1,
      dialTimeoutSeconds: Number(form.dialTimeoutSeconds) || 30,
      failureCooldownSeconds: Number(form.failureCooldownSeconds) || 0,
      specificFallback: Boolean(form.specificFallback),
      candidateGroups: form.candidateGroups || [],
      candidateSources: form.candidateSources || [],
      candidateProtocols: form.candidateProtocols || [],
      candidateCountries: form.candidateCountries || [],
      stickySessionEnabled: Boolean(form.stickySessionEnabled),
      stickySessionTtlSeconds: Number(form.stickySessionTtlSeconds) || 1800
    };
    try {
      if (editingID) {
        await updateSocks5RoutingProfile(editingID, payload);
      } else {
        await createSocks5RoutingProfile(payload);
      }
      setDialogOpen(false);
      showMessage(t(editingID ? 'settings.socks5.profiles.messages.updated' : 'settings.socks5.profiles.messages.created'));
      await load();
      onChanged?.();
    } catch (error) {
      showMessage(errorMessage(t, error), 'error');
    } finally {
      setSaving(false);
    }
  };

  const remove = async (profile) => {
    if (!window.confirm(t('settings.socks5.profiles.deleteConfirm', { name: profile.name }))) return;
    try {
      await deleteSocks5RoutingProfile(profile.id);
      showMessage(t('settings.socks5.profiles.messages.deleted'));
      await load();
      onChanged?.();
    } catch (error) {
      showMessage(errorMessage(t, error), 'error');
    }
  };

  const idValid = /^[a-z0-9][a-z0-9_-]{0,31}$/.test(form.id.trim().toLowerCase());
  const invalid =
    (!editingID && !idValid) ||
    !form.name.trim() ||
    (form.selection === 'specific' && !Number(form.nodeId)) ||
    Number(form.maxAttempts) < 1 ||
    Number(form.maxAttempts) > 5 ||
    Number(form.dialTimeoutSeconds) < 1 ||
    Number(form.dialTimeoutSeconds) > 120 ||
    Number(form.failureCooldownSeconds) < 0 ||
    Number(form.failureCooldownSeconds) > 3600 ||
    (form.selection === 'specific' && form.specificFallback && Number(form.maxAttempts) < 2) ||
    (form.stickySessionEnabled && (Number(form.stickySessionTtlSeconds) < 60 || Number(form.stickySessionTtlSeconds) > 604800));

  return (
    <>
      <Card variant="outlined" sx={{ mt: 2 }}>
        <CardHeader
          title={t('settings.socks5.profiles.title')}
          subheader={t('settings.socks5.profiles.subheader')}
          action={
            <Button variant="contained" onClick={openCreate} disabled={loading}>
              {t('settings.socks5.profiles.create')}
            </Button>
          }
        />
        <CardContent>
          <Alert severity={authenticationEnabled ? 'info' : 'warning'} sx={{ mb: 2 }}>
            {t(authenticationEnabled ? 'settings.socks5.profiles.usernameHint' : 'settings.socks5.profiles.authenticationRequired', {
              username: baseUsername
            })}
          </Alert>
          <Stack spacing={1.25}>
            {profiles.map((profile) => {
              const filters = [
                ...(profile.candidateGroups || []),
                ...(profile.candidateSources || []),
                ...(profile.candidateProtocols || []),
                ...(profile.candidateCountries || [])
              ];
              return (
                <Box key={profile.id} sx={{ p: 1.5, border: 1, borderColor: 'divider', borderRadius: 1.5 }}>
                  <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5} alignItems={{ md: 'center' }}>
                    <Box sx={{ flex: 1, minWidth: 0 }}>
                      <Stack direction="row" spacing={1} alignItems="center" useFlexGap flexWrap="wrap">
                        <Typography variant="subtitle1">{profile.name}</Typography>
                        <Chip size="small" label={profile.id} />
                        <Chip
                          size="small"
                          color={profile.enabled ? 'success' : 'default'}
                          label={t(profile.enabled ? 'common.enabled' : 'common.disabled')}
                        />
                        {profile.isDefault && <Chip size="small" color="primary" label={t('settings.socks5.profiles.default')} />}
                      </Stack>
                      <Typography variant="body2" color="text.secondary">
                        {t('settings.socks5.profiles.summary', {
                          selection: t(
                            `settings.socks5.form.selection${profile.selection === 'round_robin' ? 'RoundRobin' : profile.selection.charAt(0).toUpperCase() + profile.selection.slice(1)}`
                          ),
                          attempts: profile.maxAttempts,
                          candidates: profile.candidateCount ?? 0,
                          filters: filters.length
                        })}
                      </Typography>
                      {!profile.isDefault && (
                        <Typography variant="caption" color="text.secondary">
                          {baseUsername}@{profile.id}.account
                        </Typography>
                      )}
                    </Box>
                    {!profile.isDefault && (
                      <Stack direction="row" spacing={1}>
                        <Button size="small" onClick={() => openEdit(profile)}>
                          {t('common.edit')}
                        </Button>
                        <Button size="small" color="error" onClick={() => remove(profile)}>
                          {t('common.delete')}
                        </Button>
                      </Stack>
                    )}
                  </Stack>
                </Box>
              );
            })}
          </Stack>
        </CardContent>
      </Card>

      <Dialog open={dialogOpen} onClose={closeDialog} fullWidth maxWidth="md">
        <DialogTitle>{t(editingID ? 'settings.socks5.profiles.editTitle' : 'settings.socks5.profiles.createTitle')}</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2}>
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <TextField
                fullWidth
                label={t('settings.socks5.profiles.id')}
                value={form.id}
                disabled={Boolean(editingID)}
                onChange={(event) => setForm((previous) => ({ ...previous, id: event.target.value.toLowerCase() }))}
                helperText={t('settings.socks5.profiles.idHelper')}
                error={!editingID && form.id !== '' && !idValid}
              />
              <TextField
                fullWidth
                label={t('settings.socks5.profiles.name')}
                value={form.name}
                onChange={(event) => setForm((previous) => ({ ...previous, name: event.target.value }))}
              />
            </Stack>
            <FormControlLabel
              control={
                <Switch
                  checked={Boolean(form.enabled)}
                  onChange={(event) => setForm((previous) => ({ ...previous, enabled: event.target.checked }))}
                />
              }
              label={t('settings.socks5.profiles.enabled')}
            />
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <FormControl fullWidth>
                <InputLabel>{t('settings.socks5.form.selection')}</InputLabel>
                <Select
                  label={t('settings.socks5.form.selection')}
                  value={form.selection}
                  onChange={(event) => setForm((previous) => ({ ...previous, selection: event.target.value }))}
                >
                  <MenuItem value="best">{t('settings.socks5.form.selectionBest')}</MenuItem>
                  <MenuItem value="random">{t('settings.socks5.form.selectionRandom')}</MenuItem>
                  <MenuItem value="round_robin">{t('settings.socks5.form.selectionRoundRobin')}</MenuItem>
                  <MenuItem value="smart">{t('settings.socks5.form.selectionSmart')}</MenuItem>
                  <MenuItem value="specific">{t('settings.socks5.form.selectionSpecific')}</MenuItem>
                </Select>
              </FormControl>
              <TextField
                fullWidth
                type="number"
                label={t('settings.socks5.form.maxAttempts')}
                value={form.maxAttempts}
                onChange={(event) => setForm((previous) => ({ ...previous, maxAttempts: event.target.value }))}
                slotProps={{ htmlInput: { min: 1, max: 5 } }}
              />
            </Stack>
            {form.selection === 'specific' && (
              <Stack spacing={1.25}>
                <SpecificNodeSelector
                  value={Number(form.nodeId) || 0}
                  onChange={(nodeId) => setForm((previous) => ({ ...previous, nodeId }))}
                  disabled={saving}
                />
                <FormControlLabel
                  control={
                    <Switch
                      checked={Boolean(form.specificFallback)}
                      onChange={(event) => setForm((previous) => ({ ...previous, specificFallback: event.target.checked }))}
                    />
                  }
                  label={t('settings.socks5.form.specificFallback')}
                />
              </Stack>
            )}
            <CandidateNodePool value={form} onChange={(pool) => setForm((previous) => ({ ...previous, ...pool }))} disabled={saving} />
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <TextField
                fullWidth
                type="number"
                label={t('settings.socks5.form.dialTimeoutSeconds')}
                value={form.dialTimeoutSeconds}
                onChange={(event) => setForm((previous) => ({ ...previous, dialTimeoutSeconds: event.target.value }))}
                slotProps={{ htmlInput: { min: 1, max: 120 } }}
              />
              <TextField
                fullWidth
                type="number"
                label={t('settings.socks5.form.failureCooldownSeconds')}
                value={form.failureCooldownSeconds}
                onChange={(event) => setForm((previous) => ({ ...previous, failureCooldownSeconds: event.target.value }))}
                slotProps={{ htmlInput: { min: 0, max: 3600 } }}
              />
            </Stack>
            <Box>
              <FormControlLabel
                control={
                  <Switch
                    checked={Boolean(form.stickySessionEnabled)}
                    disabled={form.selection === 'specific'}
                    onChange={(event) => setForm((previous) => ({ ...previous, stickySessionEnabled: event.target.checked }))}
                  />
                }
                label={t('settings.socks5.profiles.sticky')}
              />
              <FormHelperText>{t('settings.socks5.profiles.stickyHelper')}</FormHelperText>
            </Box>
            {form.stickySessionEnabled && form.selection !== 'specific' && (
              <TextField
                fullWidth
                type="number"
                label={t('settings.socks5.form.stickySessionTtlSeconds')}
                value={form.stickySessionTtlSeconds}
                onChange={(event) => setForm((previous) => ({ ...previous, stickySessionTtlSeconds: event.target.value }))}
                slotProps={{ htmlInput: { min: 60, max: 604800 } }}
              />
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={closeDialog} disabled={saving}>
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
