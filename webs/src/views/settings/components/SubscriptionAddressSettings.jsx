import { useEffect, useState } from 'react';

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
      showMessage('订阅地址设置保存成功');
    } catch (error) {
      showMessage('保存失败: ' + (error.response?.data?.message || error.message), 'error');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card>
      <CardHeader
        title="订阅地址设置"
        subheader="用于生成 Telegram 机器人和网页分享中的订阅访问地址"
        avatar={<LanguageIcon color="primary" />}
      />
      <CardContent>
        <Stack spacing={2} sx={{ maxWidth: 600 }}>
          <Alert severity="info" sx={{ mb: 1 }}>
            配置后，Telegram 机器人和网页分享链接将优先使用这里填写的订阅地址。未配置时，网页仍使用当前访问地址，Telegram 使用本地地址。
          </Alert>
          <TextField
            fullWidth
            label="订阅地址"
            value={subscriptionAddress}
            onChange={(e) => setSubscriptionAddress(e.target.value)}
            placeholder="例如: https://your-domain.com"
            helperText="请填写完整地址并包含协议头（http:// 或 https://），保存后将作为订阅链接的基础地址。"
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
            {saving ? '保存中...' : '保存订阅地址'}
          </Button>
        </Stack>
      </CardContent>
    </Card>
  );
}
