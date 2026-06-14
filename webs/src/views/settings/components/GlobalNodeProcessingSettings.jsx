import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

// material-ui
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Stack from '@mui/material/Stack';
import Alert from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';

// icons
import SaveIcon from '@mui/icons-material/Save';
import TuneIcon from '@mui/icons-material/Tune';

// project imports
import { getGlobalNodeProcessingConfig, updateGlobalNodeProcessingConfig } from 'api/settings';
import NodeNameFilter from 'components/NodeNameFilter';
import NodeNamePreprocessor from 'components/NodeNamePreprocessor';
import NodeProtocolFilter from 'components/NodeProtocolFilter';

// ==============================|| 全局节点处理设置 ||============================== //

export default function GlobalNodeProcessingSettings({ showMessage }) {
  const { t } = useTranslation();
  const [config, setConfig] = useState({
    nodeNameWhitelist: '',
    nodeNameBlacklist: '',
    protocolWhitelist: '',
    protocolBlacklist: '',
    nodeNamePreprocess: ''
  });
  const [saving, setSaving] = useState(false);
  const [protocolOptions, setProtocolOptions] = useState([]);

  useEffect(() => {
    fetchConfig();
    // 获取协议选项（可以从 API 获取或使用静态列表）
    setProtocolOptions(['ss', 'ssr', 'vmess', 'vless', 'trojan', 'hysteria', 'hysteria2', 'tuic', 'wireguard']);
  }, []);

  const fetchConfig = async () => {
    try {
      const res = await getGlobalNodeProcessingConfig();
      console.log('获取全局节点处理配置:', res);
      if (res.data) {
        setConfig({
          nodeNameWhitelist: res.data.nodeNameWhitelist || '',
          nodeNameBlacklist: res.data.nodeNameBlacklist || '',
          protocolWhitelist: res.data.protocolWhitelist || '',
          protocolBlacklist: res.data.protocolBlacklist || '',
          nodeNamePreprocess: res.data.nodeNamePreprocess || ''
        });
      }
    } catch (error) {
      console.error('获取全局节点处理配置失败:', error);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      console.log('保存全局节点处理配置:', config);
      await updateGlobalNodeProcessingConfig(config);
      showMessage(t('globalNodeProcessing.messages.saveSuccess'));
      // 保存成功后重新获取配置，确认持久化
      await fetchConfig();
    } catch (error) {
      console.error('保存失败:', error);
      showMessage(error.message || t('globalNodeProcessing.messages.saveFailed'), 'error');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card variant="outlined">
      <CardHeader
        avatar={<TuneIcon color="primary" />}
        title={t('globalNodeProcessing.title')}
        subheader={t('globalNodeProcessing.subheader')}
      />
      <CardContent>
        <Stack spacing={3}>
          {/* 提示信息 */}
          <Alert severity="info">
            <Typography variant="body2" dangerouslySetInnerHTML={{ __html: t('globalNodeProcessing.info') }} />
          </Alert>

          {/* 节点名称过滤 */}
          <Box>
            <NodeNameFilter
              whitelistValue={config.nodeNameWhitelist}
              blacklistValue={config.nodeNameBlacklist}
              onWhitelistChange={(val) => setConfig({ ...config, nodeNameWhitelist: val })}
              onBlacklistChange={(val) => setConfig({ ...config, nodeNameBlacklist: val })}
            />
          </Box>

          {/* 协议过滤 */}
          <Box>
            <NodeProtocolFilter
              protocolOptions={protocolOptions}
              whitelistValue={config.protocolWhitelist}
              blacklistValue={config.protocolBlacklist}
              onWhitelistChange={(val) => setConfig({ ...config, protocolWhitelist: val })}
              onBlacklistChange={(val) => setConfig({ ...config, protocolBlacklist: val })}
            />
          </Box>

          {/* 节点名称预处理 */}
          <Box>
            <NodeNamePreprocessor value={config.nodeNamePreprocess} onChange={(val) => setConfig({ ...config, nodeNamePreprocess: val })} />
          </Box>

          {/* 保存按钮 */}
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
