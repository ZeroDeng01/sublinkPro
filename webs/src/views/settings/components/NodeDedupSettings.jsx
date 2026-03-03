import { useState, useEffect } from 'react';

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
      showMessage('节点去重设置保存成功');
    } catch (error) {
      console.error('保存失败:', error);
      showMessage(error.message || '保存失败', 'error');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card variant="outlined">
      <CardHeader
        avatar={<FilterAltIcon color="primary" />}
        title="跨机场节点去重"
        subheader="控制不同机场之间是否进行节点内容去重"
      />
      <CardContent>
        <Stack spacing={2}>
          <FormControlLabel
            control={
              <Switch
                checked={crossAirportDedupEnabled}
                onChange={(e) => setCrossAirportDedupEnabled(e.target.checked)}
              />
            }
            label="启用跨机场去重"
          />
          <Alert severity={crossAirportDedupEnabled ? 'info' : 'warning'} variant="standard">
            <Typography variant="body2">
              {crossAirportDedupEnabled ? (
                <>
                  当前为<strong>开启</strong>状态：不同机场间配置完全相同的节点（ContentHash 一致）只保留最先入库的一份，避免重复。
                </>
              ) : (
                <>
                  当前为<strong>关闭</strong>状态：每个机场独立保留自己的节点，即使不同机场存在配置完全相同的节点也会各自入库。
                  同一机场内的重复节点仍然会被去重。
                </>
              )}
            </Typography>
          </Alert>
          <Stack direction="row" justifyContent="flex-end">
            <Button variant="contained" startIcon={<SaveIcon />} onClick={handleSave} disabled={saving}>
              {saving ? '保存中...' : '保存'}
            </Button>
          </Stack>
        </Stack>
      </CardContent>
    </Card>
  );
}
