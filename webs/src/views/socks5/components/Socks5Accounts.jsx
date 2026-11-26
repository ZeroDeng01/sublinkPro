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
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import {
  createSocks5Account,
  deleteSocks5Account,
  getSocks5Accounts,
  getSocks5RoutingProfiles,
  updateSocks5Account
} from 'api/proxyGateway';

const emptyAccount = { id: '', username: '', password: '', profileId: 'default', enabled: true, hasPassword: false };

function errorMessage(t, error) {
  const response = error.response?.data;
  if (response?.i18nKey) return t(response.i18nKey, response.i18nParams || {});
  return response?.msg || error.message;
}

export default function Socks5Accounts({ showMessage, onChanged }) {
  const { t } = useTranslation();
  const [accounts, setAccounts] = useState([]);
  const [profiles, setProfiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [deleting, setDeleting] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingID, setEditingID] = useState('');
  const [form, setForm] = useState(emptyAccount);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setLoadError('');
    try {
      const [accountsResponse, profilesResponse] = await Promise.all([getSocks5Accounts(), getSocks5RoutingProfiles()]);
      setAccounts(Array.isArray(accountsResponse.data) ? accountsResponse.data : []);
      setProfiles(Array.isArray(profilesResponse.data) ? profilesResponse.data : []);
    } catch (error) {
      setLoadError(errorMessage(t, error));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    load();
  }, [load]);

  const openCreate = () => {
    setEditingID('');
    setForm(emptyAccount);
    setDialogOpen(true);
  };

  const openEdit = (account) => {
    setEditingID(account.id);
    setForm({ ...emptyAccount, ...account, password: '' });
    setDialogOpen(true);
  };

  const save = async () => {
    setSaving(true);
    const payload = {
      id: form.id.trim().toLowerCase(),
      username: form.username.trim(),
      password: form.password,
      profileId: form.profileId,
      enabled: Boolean(form.enabled)
    };
    try {
      if (editingID) await updateSocks5Account(editingID, payload);
      else await createSocks5Account(payload);
      setDialogOpen(false);
      showMessage(t(editingID ? 'settings.socks5.accounts.messages.updated' : 'settings.socks5.accounts.messages.created'));
      await load();
      onChanged?.();
    } catch (error) {
      showMessage(errorMessage(t, error), 'error');
    } finally {
      setSaving(false);
    }
  };

  const remove = async (account) => {
    if (!window.confirm(t('settings.socks5.accounts.deleteConfirm', { username: account.username }))) return;
    setDeleting(account.id);
    try {
      await deleteSocks5Account(account.id);
      showMessage(t('settings.socks5.accounts.messages.deleted'));
      await load();
      onChanged?.();
    } catch (error) {
      showMessage(errorMessage(t, error), 'error');
    } finally {
      setDeleting('');
    }
  };

  const idValid = /^[a-z0-9][a-z0-9_-]{0,31}$/.test(form.id.trim().toLowerCase());
  const usernameValid = form.username.trim().length > 0 && !form.username.includes('@') && !form.username.includes('\0');
  const invalid = (!editingID && !idValid) || !usernameValid || (!editingID && !form.password) || !form.profileId;

  return (
    <>
      <Card variant="outlined">
        <CardHeader
          sx={{
            flexWrap: 'wrap',
            gap: 1,
            '& .MuiCardHeader-content': { minWidth: 0, flexBasis: { xs: '100%', sm: 'auto' } },
            '& .MuiCardHeader-action': { m: 0, alignSelf: 'center' }
          }}
          title={t('settings.socks5.accounts.title')}
          subheader={t('settings.socks5.accounts.subheader')}
          action={
            <Button variant="contained" onClick={openCreate} disabled={loading || Boolean(loadError) || Boolean(deleting)}>
              {t('settings.socks5.accounts.create')}
            </Button>
          }
        />
        <CardContent sx={{ pt: 0 }}>
          {loadError && (
            <Alert
              severity="error"
              sx={{ mb: 1.5 }}
              action={
                <Button color="inherit" size="small" onClick={load}>
                  {t('settings.socks5.page.retry')}
                </Button>
              }
            >
              {loadError}
            </Alert>
          )}
          {loading && (
            <Typography role="status" color="text.secondary">
              {t('settings.socks5.page.loading')}
            </Typography>
          )}
          <Stack spacing={1.25}>
            {!loading && !loadError && accounts.length === 0 && (
              <Typography color="text.secondary">{t('settings.socks5.accounts.empty')}</Typography>
            )}
            {accounts.map((account) => {
              const profile = profiles.find((item) => item.id === account.profileId);
              return (
                <Box key={account.id} sx={{ p: 1.5, border: 1, borderColor: 'divider', borderRadius: 1.5 }}>
                  <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5} alignItems={{ md: 'center' }}>
                    <Box sx={{ flex: 1, minWidth: 0 }}>
                      <Stack direction="row" spacing={1} alignItems="center" useFlexGap flexWrap="wrap">
                        <Typography variant="subtitle1">{account.username}</Typography>
                        <Chip size="small" label={account.id} />
                        <Chip
                          size="small"
                          color={account.enabled ? 'success' : 'default'}
                          label={t(account.enabled ? 'common.enabled' : 'common.disabled')}
                        />
                        <Chip
                          size="small"
                          color="primary"
                          variant="outlined"
                          label={
                            account.profileId === 'default' ? t('settings.socks5.profiles.default') : profile?.name || account.profileId
                          }
                        />
                      </Stack>
                      <Typography variant="body2" color="text.secondary">
                        {t(account.hasPassword ? 'settings.socks5.accounts.passwordSaved' : 'settings.socks5.accounts.passwordMissing')}
                      </Typography>
                    </Box>
                    <Stack direction="row" spacing={1}>
                      <Button size="small" disabled={loading || Boolean(deleting)} onClick={() => openEdit(account)}>
                        {t('common.edit')}
                      </Button>
                      <Button size="small" color="error" disabled={Boolean(deleting)} onClick={() => remove(account)}>
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
        <DialogTitle>{t(editingID ? 'settings.socks5.accounts.editTitle' : 'settings.socks5.accounts.createTitle')}</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <Typography variant="body2" color="text.secondary">
              {t('settings.socks5.page.applyImmediately')}
            </Typography>
            <TextField
              label={t('settings.socks5.accounts.id')}
              value={form.id}
              disabled={Boolean(editingID)}
              onChange={(event) => setForm((previous) => ({ ...previous, id: event.target.value.toLowerCase() }))}
              helperText={t('settings.socks5.accounts.idHelper')}
              error={!editingID && form.id !== '' && !idValid}
            />
            <TextField
              label={t('settings.socks5.accounts.username')}
              value={form.username}
              onChange={(event) => setForm((previous) => ({ ...previous, username: event.target.value }))}
              helperText={t(
                form.username.includes(':') ? 'settings.socks5.form.httpUsernameWarning' : 'settings.socks5.accounts.usernameHelper'
              )}
              error={form.username !== '' && !usernameValid}
            />
            <TextField
              type="password"
              label={t('settings.socks5.accounts.password')}
              value={form.password}
              onChange={(event) => setForm((previous) => ({ ...previous, password: event.target.value }))}
              helperText={t(editingID ? 'settings.socks5.accounts.passwordEditHelper' : 'settings.socks5.accounts.passwordCreateHelper')}
            />
            <FormControl fullWidth>
              <InputLabel id="gateway-account-profile-label">{t('settings.socks5.accounts.profile')}</InputLabel>
              <Select
                labelId="gateway-account-profile-label"
                value={form.profileId}
                label={t('settings.socks5.accounts.profile')}
                onChange={(event) => setForm((previous) => ({ ...previous, profileId: event.target.value }))}
              >
                {profiles
                  .filter((profile) => profile.enabled)
                  .map((profile) => (
                    <MenuItem key={profile.id} value={profile.id}>
                      {profile.isDefault ? t('settings.socks5.profiles.default') : profile.name}
                    </MenuItem>
                  ))}
              </Select>
            </FormControl>
            <FormControlLabel
              control={
                <Switch
                  checked={Boolean(form.enabled)}
                  onChange={(event) => setForm((previous) => ({ ...previous, enabled: event.target.checked }))}
                />
              }
              label={t('settings.socks5.accounts.enabled')}
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
