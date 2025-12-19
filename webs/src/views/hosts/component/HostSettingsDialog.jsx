import { useState, useEffect } from 'react';
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';
import Typography from '@mui/material/Typography';
import TextField from '@mui/material/TextField';
import Divider from '@mui/material/Divider';
import Autocomplete from '@mui/material/Autocomplete';
import Collapse from '@mui/material/Collapse';
import Paper from '@mui/material/Paper';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import ToggleButton from '@mui/material/ToggleButton';
import CircularProgress from '@mui/material/CircularProgress';
import Alert from '@mui/material/Alert';

// project imports
import { getHostSettings, updateHostSettings } from 'api/hosts';
import { getNodes } from 'api/nodes';
import SearchableNodeSelect from '../../../components/SearchableNodeSelect'; // Ensure this relative path is correct from the new location

export default function HostSettingsDialog({ open, onClose }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  // Local state for settings form
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [settings, setSettings] = useState({
    persist_host: false,
    dns_server: '',
    dns_presets: [],
    expire_hours: 0,
    dns_use_proxy: false,
    dns_proxy_strategy: 'auto',
    dns_proxy_node_id: 0
  });

  // Proxy nodes state
  const [proxyNodes, setProxyNodes] = useState([]);
  const [loadingNodes, setLoadingNodes] = useState(false);
  const [error, setError] = useState('');

  // Initial fetch when opened
  useEffect(() => {
    if (open) {
      loadSettings();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  const loadSettings = async () => {
    setLoading(true);
    setError('');
    try {
      const res = await getHostSettings();
      if (res.code === 200) {
        const data = res.data;
        setSettings({
          persist_host: data.persist_host || false,
          dns_server: data.dns_server || '',
          dns_presets: data.dns_presets || [],
          expire_hours: data.expire_hours || 0,
          dns_use_proxy: data.dns_use_proxy || false,
          dns_proxy_strategy: data.dns_proxy_strategy || 'auto',
          dns_proxy_node_id: data.dns_proxy_node_id || 0
        });

        // Pre-fetch nodes if manual strategy is already selected
        if (data.dns_proxy_strategy === 'manual') {
          fetchNodes();
        }
      } else {
        setError(res.msg || '加载设置失败');
      }
    } catch (err) {
      console.error(err);
      setError('加载设置失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchNodes = async () => {
    if (proxyNodes.length > 0) return;
    setLoadingNodes(true);
    try {
      const res = await getNodes();
      if (res.code === 200) {
        setProxyNodes(res.data || []);
      }
    } catch (err) {
      console.error(err);
      // Not blocking flow, just can't select
    } finally {
      setLoadingNodes(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    setError('');
    try {
      // Construct payload
      const payload = {
        persist_host: settings.persist_host,
        dns_server: settings.dns_server,
        dns_use_proxy: settings.dns_use_proxy,
        dns_proxy_strategy: settings.dns_proxy_strategy,
        dns_proxy_node_id: settings.dns_proxy_node_id,
        expire_hours: settings.expire_hours
      };

      const res = await updateHostSettings(payload);
      if (res.code === 200) {
        onClose(); // Close on success
      } else {
        setError(res.msg || '保存失败');
      }
    } catch (err) {
      console.error(err);
      setError('保存失败');
    } finally {
      setSaving(false);
    }
  };

  // Helper to update specific field
  const updateField = (field, value) => {
    setSettings((prev) => ({ ...prev, [field]: value }));
  };

  return (
    <Dialog open={open} onClose={saving ? null : onClose} maxWidth="sm" fullWidth>
      <DialogTitle>模块设置</DialogTitle>
      <DialogContent dividers>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
            <CircularProgress />
          </Box>
        ) : (
          <Stack spacing={3}>
            {error && <Alert severity="error">{error}</Alert>}

            {/* 1. Persistence & Expiration */}
            <Box>
              <Stack direction={isMobile ? 'column' : 'row'} spacing={2} alignItems={isMobile ? 'start' : 'center'}>
                <FormControlLabel
                  control={
                    <Switch checked={settings.persist_host} onChange={(e) => updateField('persist_host', e.target.checked)} size="small" />
                  }
                  label={
                    <Typography variant="body2">
                      持久化节点Host
                      <Typography component="span" variant="caption" color="textSecondary" sx={{ ml: 0.5 }}>
                        (测速成功时保存域名→IP)
                      </Typography>
                    </Typography>
                  }
                />

                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <TextField
                    label="有效期"
                    type="number"
                    size="small"
                    value={settings.expire_hours}
                    disabled={!settings.persist_host}
                    onChange={(e) => updateField('expire_hours', Math.max(0, parseInt(e.target.value) || 0))}
                    sx={{ width: isMobile ? 120 : 100 }}
                    InputProps={{ inputProps: { min: 0 } }}
                  />
                  <Typography variant="body2" color="textSecondary">
                    小时 {settings.expire_hours === 0 && '(永不过期)'}
                  </Typography>
                </Box>
              </Stack>
            </Box>

            <Divider />

            {/* 2. DNS Settings */}
            <Box>
              <Typography variant="subtitle2" color="primary" sx={{ mb: 2 }}>
                DNS 解析设置
              </Typography>

              <Stack spacing={2}>
                <Autocomplete
                  freeSolo
                  fullWidth
                  size="small"
                  options={settings.dns_presets || []}
                  getOptionLabel={(option) => {
                    if (typeof option === 'string') return option;
                    return option.label ? `${option.label} (${option.value})` : option.value || '';
                  }}
                  value={settings.dns_presets?.find((p) => p.value === settings.dns_server) || settings.dns_server}
                  onInputChange={(event, newInputValue, reason) => {
                    if (reason === 'input') {
                      updateField('dns_server', newInputValue);
                    }
                    if (reason === 'clear') {
                      updateField('dns_server', '');
                    }
                  }}
                  onChange={(event, newValue) => {
                    const val = typeof newValue === 'string' ? newValue : newValue?.value || '';
                    updateField('dns_server', val);
                  }}
                  renderInput={(params) => (
                    <TextField
                      {...params}
                      label="DNS 解析服务器"
                      placeholder="留空（使用系统 DNS）"
                      helperText="支持 DoH (https://...) 或 UDP (IP)。IP格式（UDP协议）将始终直连不走代理。"
                    />
                  )}
                />

                <Box>
                  <FormControlLabel
                    control={
                      <Switch
                        checked={settings.dns_use_proxy}
                        disabled={!settings.dns_server}
                        onChange={(e) => {
                          const val = e.target.checked;
                          updateField('dns_use_proxy', val);
                          if (val && settings.dns_proxy_strategy === 'manual') {
                            fetchNodes();
                          }
                        }}
                        size="small"
                      />
                    }
                    label="使用代理连接 DNS (仅支持 DoH)"
                  />

                  <Collapse in={settings.dns_use_proxy && !!settings.dns_server}>
                    <Paper
                      variant="outlined"
                      sx={{ mt: 1, p: 2, bgcolor: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.02)' : 'rgba(0,0,0,0.01)' }}
                    >
                      <Stack spacing={2}>
                        <Stack direction="row" alignItems="center" spacing={2} flexWrap="wrap" gap={1}>
                          <Typography variant="body2" sx={{ minWidth: 60 }}>
                            代理策略:
                          </Typography>
                          <ToggleButtonGroup
                            value={settings.dns_proxy_strategy}
                            exclusive
                            onChange={(e, val) => {
                              if (!val) return;
                              updateField('dns_proxy_strategy', val);
                              if (val === 'manual') fetchNodes();
                            }}
                            size="small"
                            color="primary"
                            sx={{ height: 32 }}
                          >
                            <ToggleButton value="auto">自动选择</ToggleButton>
                            <ToggleButton value="manual">手动指定</ToggleButton>
                          </ToggleButtonGroup>
                        </Stack>

                        {settings.dns_proxy_strategy === 'manual' && (
                          <SearchableNodeSelect
                            nodes={proxyNodes}
                            loading={loadingNodes}
                            value={proxyNodes.find((n) => n.ID === settings.dns_proxy_node_id) || null}
                            onChange={(newValue) => {
                              updateField('dns_proxy_node_id', newValue ? newValue.ID : 0);
                            }}
                            label="选择代理节点"
                            displayField="Name"
                            valueField="ID"
                          />
                        )}

                        <Typography variant="caption" color="textSecondary">
                          {settings.dns_proxy_strategy === 'auto'
                            ? '系统将自动选择最近一次测速中延迟最低且无丢包的节点。'
                            : '指定一个节点用于建立 DNS 连接。注意：仅 DoH (https) 协议支持通过代理连接。'}
                        </Typography>
                      </Stack>
                    </Paper>
                  </Collapse>
                </Box>
              </Stack>
            </Box>
          </Stack>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} color="inherit" disabled={saving}>
          取消
        </Button>
        <Button onClick={handleSave} variant="contained" disabled={loading || saving}>
          {saving ? '保存中...' : '确定'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
