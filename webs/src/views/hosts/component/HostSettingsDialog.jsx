import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
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
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

// project imports
import { getHostSettings, updateHostSettings } from 'api/hosts';
import { getNodes } from 'api/nodes';
import SearchableNodeSelect from '../../../components/SearchableNodeSelect'; // Ensure this relative path is correct from the new location

export default function HostSettingsDialog({ open, onClose }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
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
        setError(res.msg || t('hosts.settings.messages.loadFailed'));
      }
    } catch (err) {
      console.error(err);
      setError(t('hosts.settings.messages.loadFailed'));
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
        setError(res.msg || t('hosts.settings.messages.saveFailed'));
      }
    } catch (err) {
      console.error(err);
      setError(t('hosts.settings.messages.saveFailed'));
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
      <DialogTitle>{t('hosts.settings.title')}</DialogTitle>
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
                      {t('hosts.settings.persistHost')}
                      <Typography component="span" variant="caption" color="textSecondary" sx={{ ml: 0.5 }}>
                        {t('hosts.settings.persistHostHint')}
                      </Typography>
                    </Typography>
                  }
                />

                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <TextField
                    label={t('hosts.settings.expireHours')}
                    type="number"
                    size="small"
                    value={settings.expire_hours}
                    disabled={!settings.persist_host}
                    onChange={(e) => updateField('expire_hours', Math.max(0, parseInt(e.target.value) || 0))}
                    sx={{ width: isMobile ? 120 : 100 }}
                    InputProps={{ inputProps: { min: 0 } }}
                  />
                  <Typography variant="body2" color="textSecondary">
                    {t('hosts.settings.hours')} {settings.expire_hours === 0 && t('hosts.settings.neverExpire')}
                  </Typography>
                </Box>
              </Stack>
            </Box>

            <Divider />

            {/* 2. DNS Settings */}
            <Box>
              <Typography variant="subtitle2" color="primary" sx={{ mb: 2 }}>
                {t('hosts.settings.dnsTitle')}
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
                      label={t('hosts.settings.dnsServer')}
                      placeholder={t('hosts.settings.dnsPlaceholder')}
                      helperText={t('hosts.settings.dnsHelper')}
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
                    label={t('hosts.settings.dnsUseProxy')}
                  />

                  <Collapse in={settings.dns_use_proxy && !!settings.dns_server}>
                    <Paper variant="outlined" sx={{ mt: 1, p: 2, bgcolor: isDark ? 'rgba(255,255,255,0.02)' : 'rgba(0,0,0,0.01)' }}>
                      <Stack spacing={2}>
                        <Stack direction="row" alignItems="center" spacing={2} flexWrap="wrap" gap={1}>
                          <Typography variant="body2" sx={{ minWidth: 60 }}>
                            {t('hosts.settings.proxyStrategy')}
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
                            <ToggleButton value="auto">{t('hosts.settings.strategy.auto')}</ToggleButton>
                            <ToggleButton value="manual">{t('hosts.settings.strategy.manual')}</ToggleButton>
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
                            label={t('hosts.settings.proxyNode')}
                            displayField="Name"
                            valueField="ID"
                          />
                        )}

                        <Typography variant="caption" color="textSecondary">
                          {settings.dns_proxy_strategy === 'auto'
                            ? t('hosts.settings.strategy.autoHint')
                            : t('hosts.settings.strategy.manualHint')}
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
          {t('common.cancel')}
        </Button>
        <Button onClick={handleSave} variant="contained" disabled={loading || saving}>
          {saving ? t('common.saving') : t('common.confirm')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
