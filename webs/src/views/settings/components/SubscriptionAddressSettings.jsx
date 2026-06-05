import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import InputAdornment from '@mui/material/InputAdornment';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';

import LanguageIcon from '@mui/icons-material/Language';
import SaveIcon from '@mui/icons-material/Save';

import { getSystemDomain as getSubscriptionAddress, updateSystemDomain as updateSubscriptionAddress } from 'api/settings';

export default function SubscriptionAddressSettings({ showMessage }) {
  const { t } = useTranslation();
  const [subscriptionAddress, setSubscriptionAddress] = useState('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    fetchSubscriptionAddress();
  }, []);

  const fetchSubscriptionAddress = async () => {
    try {
      const res = await getSubscriptionAddress();
      setSubscriptionAddress(res.data?.systemDomain || '');
    } catch (error) {
      console.error('获取订阅地址配置失败:', error);
    }
  };

  const handleSaveSubscriptionAddress = async () => {
    const trimmedSubscriptionAddress = subscriptionAddress.trim();

    setSaving(true);
    try {
      await updateSubscriptionAddress({ systemDomain: trimmedSubscriptionAddress });
      setSubscriptionAddress(trimmedSubscriptionAddress);
      showMessage(t('subscriptionAddress.messages.saveSuccess'));
    } catch (error) {
      showMessage(t('subscriptionAddress.messages.saveFailed', { message: error.response?.data?.message || error.message }), 'error');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card>
      <CardHeader
        title={t('subscriptionAddress.title')}
        subheader={t('subscriptionAddress.subheader')}
        avatar={<LanguageIcon color="primary" />}
      />
      <CardContent>
        <Stack spacing={2} sx={{ maxWidth: 600 }}>
          <Alert severity="info" sx={{ mb: 1 }}>
            {t('subscriptionAddress.info')}
          </Alert>
          <TextField
            fullWidth
            label={t('subscriptionAddress.field.label')}
            value={subscriptionAddress}
            onChange={(e) => setSubscriptionAddress(e.target.value)}
            placeholder={t('subscriptionAddress.field.placeholder')}
            helperText={t('subscriptionAddress.field.helper')}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <LanguageIcon color="action" />
                </InputAdornment>
              )
            }}
          />
          <Button
            variant="contained"
            onClick={handleSaveSubscriptionAddress}
            disabled={saving}
            startIcon={<SaveIcon />}
            sx={{ alignSelf: 'flex-start' }}
          >
            {saving ? t('common.saving') : t('subscriptionAddress.actions.save')}
          </Button>
        </Stack>
      </CardContent>
    </Card>
  );
}
