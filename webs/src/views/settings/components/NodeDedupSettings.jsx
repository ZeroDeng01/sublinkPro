import { useState, useEffect } from 'react';
import { Trans, useTranslation } from 'react-i18next';

// material-ui
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import Typography from '@mui/material/Typography';
import Alert from '@mui/material/Alert';

// icons
import SaveIcon from '@mui/icons-material/Save';
import FilterAltIcon from '@mui/icons-material/FilterAlt';

// project imports
import { getNodeDedupConfig, updateNodeDedupConfig } from 'api/settings';

// ==============================|| 节点去重设置 ||============================== //

export default function NodeDedupSettings({ showMessage }) {
  const { t } = useTranslation();
  const [crossAirportDedupEnabled, setCrossAirportDedupEnabled] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    fetchConfig();
  }, []);

  const fetchConfig = async () => {
    try {
      const res = await getNodeDedupConfig();
      if (res.data) {
        setCrossAirportDedupEnabled(res.data.crossAirportDedupEnabled !== false);
      }
    } catch (error) {
      console.error('获取节点去重配置失败:', error);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await updateNodeDedupConfig({ crossAirportDedupEnabled });
      showMessage(t('nodeDedup.messages.saveSuccess'));
    } catch (error) {
      console.error('保存失败:', error);
      showMessage(error.message || t('nodeDedup.messages.saveFailed'), 'error');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card variant="outlined">
      <CardHeader avatar={<FilterAltIcon color="primary" />} title={t('nodeDedup.title')} subheader={t('nodeDedup.subheader')} />
      <CardContent>
        <Stack spacing={2}>
          <FormControlLabel
            control={<Switch checked={crossAirportDedupEnabled} onChange={(e) => setCrossAirportDedupEnabled(e.target.checked)} />}
            label={t('nodeDedup.enable')}
          />
          <Alert severity={crossAirportDedupEnabled ? 'info' : 'warning'} variant="standard">
            <Typography variant="body2">
              {crossAirportDedupEnabled ? (
                <>
                  <Trans i18nKey="nodeDedup.enabledInfo" components={{ strong: <strong /> }} />
                </>
              ) : (
                <>
                  <Trans i18nKey="nodeDedup.disabledInfo" components={{ strong: <strong /> }} />
                </>
              )}
            </Typography>
          </Alert>
          <Stack direction="row" justifyContent="flex-end">
            <Button variant="contained" startIcon={<SaveIcon />} onClick={handleSave} disabled={saving}>
              {saving ? t('common.saving') : t('common.save')}
            </Button>
          </Stack>
        </Stack>
      </CardContent>
    </Card>
  );
}
