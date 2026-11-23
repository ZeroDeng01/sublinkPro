import { Alert, Stack, TextField } from '@mui/material';
import { useTranslation } from 'react-i18next';

export const validDeviceLimit = (value) => value !== '' && Number.isInteger(Number(value)) && Number(value) >= 0 && Number(value) <= 10000;

export default function DevicePolicyFields({ maxDevices, onChange, defaults = false }) {
  const { t } = useTranslation();
  return (
    <Stack spacing={1.5}>
      <TextField
        fullWidth
        type="number"
        label={t(defaults ? 'subscriptions.devices.defaultLimit' : 'subscriptions.devices.limit')}
        value={maxDevices}
        onChange={(e) => onChange({ maxDevices: e.target.value === '' ? '' : Number(e.target.value) })}
        error={!validDeviceLimit(maxDevices)}
        helperText={t('subscriptions.devices.limitHelp')}
        slotProps={{ htmlInput: { min: 0, max: 10000, step: 1 } }}
      />
      {defaults && <Alert severity="info">{t('subscriptions.devices.defaultsHelp')}</Alert>}
      {Number(maxDevices) > 0 && <Alert severity="info">{t('subscriptions.devices.importHelp')}</Alert>}
    </Stack>
  );
}
