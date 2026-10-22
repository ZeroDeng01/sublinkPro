import { useCallback, useEffect, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  TextField,
  Typography,
  useMediaQuery
} from '@mui/material';
import { useTheme } from '@mui/material/styles';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getSurfaceTokens } from 'themes/surfaceTokens';
import { useTranslation } from 'react-i18next';
import { formatDateTime } from 'i18n/locales';
import { getShareDevices, updateShareDevice, resetShareDevices } from 'api/shares';
import ConfirmDialog from './ConfirmDialog';

export default function ShareDevicesDialog({ share, onClose, onChanged }) {
  const { t, i18n } = useTranslation();
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { dialogSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(theme, isDark);
  const mobile = useMediaQuery(theme.breakpoints.down('sm'));
  const timestamp = (value) => formatDateTime(value, i18n.language, { dateStyle: 'medium', timeStyle: 'short' });
  const [devices, setDevices] = useState([]);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [confirm, setConfirm] = useState(null);
  const load = useCallback(async () => {
    if (!share) return;
    setBusy(true);
    setError('');
    try {
      const response = await getShareDevices(share.id);
      setDevices(response.data || []);
    } catch (e) {
      setError(e.response?.data?.msg || t('subscriptions.devices.failed'));
    } finally {
      setBusy(false);
    }
  }, [share, t]);
  useEffect(() => {
    setDevices([]);
    load();
  }, [load]);
  const mutate = async (action) => {
    setBusy(true);
    setError('');
    try {
      await action();
      await load();
      onChanged();
    } catch (e) {
      setError(e.response?.data?.msg || t('subscriptions.devices.failed'));
    } finally {
      setBusy(false);
      setConfirm(null);
    }
  };
  return (
    <Dialog
      open={!!share}
      onClose={busy ? undefined : onClose}
      fullWidth
      fullScreen={mobile}
      maxWidth="sm"
      slotProps={{ paper: { sx: { bgcolor: dialogSurface } } }}
    >
      <DialogTitle>
        {t('subscriptions.devices.manage')} · {share?.name}
      </DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ pt: 1 }}>
          <Alert severity="info">{t('subscriptions.devices.scopeHelp')}</Alert>
          {error && <Alert severity="error">{error}</Alert>}
          {busy && <CircularProgress size={24} />}
          {!busy && devices.length === 0 && <Typography>{t('subscriptions.devices.empty')}</Typography>}
          {devices.map((device) => (
            <Card key={device.id} variant="outlined" sx={{ bgcolor: nestedPanelSurface, borderColor: panelBorder }}>
              <CardContent>
                <Stack spacing={1.5}>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Typography>
                      {device.os || '—'} / {device.model || '—'}
                    </Typography>
                    <Chip size="small" label={t(device.revoked ? 'subscriptions.devices.revoked' : 'subscriptions.devices.active')} />
                  </Stack>
                  <Typography variant="caption" color="text.secondary">
                    {t('subscriptions.devices.firstSeen')}: {timestamp(device.created_at)}
                    <br />
                    {t('subscriptions.devices.lastSeen')}: {timestamp(device.last_access_at)}
                  </Typography>
                  <TextField
                    label={t('subscriptions.devices.name')}
                    value={device.name}
                    disabled={busy}
                    slotProps={{ htmlInput: { maxLength: 100 } }}
                    onChange={(e) => setDevices((prev) => prev.map((d) => (d.id === device.id ? { ...d, name: e.target.value } : d)))}
                  />
                  <Stack direction="row" spacing={1}>
                    <Button disabled={busy} onClick={() => mutate(() => updateShareDevice(share.id, { id: device.id, name: device.name }))}>
                      {t('common.save')}
                    </Button>
                    <Button
                      disabled={busy}
                      color={device.revoked ? 'primary' : 'error'}
                      onClick={() =>
                        setConfirm({
                          content: t(device.revoked ? 'subscriptions.devices.restoreConfirm' : 'subscriptions.devices.revokeConfirm'),
                          action: () => updateShareDevice(share.id, { id: device.id, revoked: !device.revoked })
                        })
                      }
                    >
                      {t(device.revoked ? 'subscriptions.devices.restore' : 'subscriptions.devices.revoke')}
                    </Button>
                  </Stack>
                </Stack>
              </CardContent>
            </Card>
          ))}
        </Stack>
      </DialogContent>
      <DialogActions sx={{ flexWrap: 'wrap' }}>
        <Button
          color="error"
          disabled={busy}
          onClick={() => setConfirm({ content: t('subscriptions.devices.resetConfirm'), action: () => resetShareDevices(share.id) })}
        >
          {t('subscriptions.devices.reset')}
        </Button>
        <Button onClick={onClose} disabled={busy}>
          {t('common.close')}
        </Button>
      </DialogActions>
      <ConfirmDialog
        open={!!confirm}
        title={t('common.confirm')}
        content={confirm?.content || ''}
        onClose={() => !busy && setConfirm(null)}
        onConfirm={() => !busy && mutate(confirm.action)}
      />
    </Dialog>
  );
}
