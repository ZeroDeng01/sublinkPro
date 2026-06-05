import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';

// material-ui
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import Alert from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import IconButton from '@mui/material/IconButton';
import Collapse from '@mui/material/Collapse';
import LinearProgress from '@mui/material/LinearProgress';
import CircularProgress from '@mui/material/CircularProgress';
import Divider from '@mui/material/Divider';
import { useTheme } from '@mui/material/styles';

// icons
import PublicIcon from '@mui/icons-material/Public';
import DownloadIcon from '@mui/icons-material/Download';
import SaveIcon from '@mui/icons-material/Save';
import RestoreIcon from '@mui/icons-material/Restore';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import CloseIcon from '@mui/icons-material/Close';

// project imports
import { getGeoIPConfig, saveGeoIPConfig, getGeoIPStatus, downloadGeoIP, stopGeoIPDownload } from 'api/geoip';
import { getNodes } from 'api/nodes';
import SearchableNodeSelect from 'components/SearchableNodeSelect';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

// 默认下载地址
const DEFAULT_DOWNLOAD_URL = 'https://git.io/GeoLite2-City.mmdb';

// ==============================|| GeoIP 设置对话框 ||============================== //

export default function GeoIPSettingsDialog({ open, onClose, showMessage }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const { dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(theme, isDark);
  const { primaryText, secondaryText } = getReadableTextTokens(theme, isDark);
  const [config, setConfig] = useState({
    downloadUrl: DEFAULT_DOWNLOAD_URL,
    useProxy: false,
    proxyLink: '',
    lastUpdate: ''
  });
  const [status, setStatus] = useState({
    available: false,
    path: '',
    size: 0,
    sizeFormatted: '',
    modTime: '',
    downloading: false,
    progress: 0,
    error: '',
    source: '' // 'auto' 或 'manual'
  });
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  // 代理节点选择
  const [proxyNodes, setProxyNodes] = useState([]);
  const [loadingNodes, setLoadingNodes] = useState(false);
  const [selectedNode, setSelectedNode] = useState(null);

  // 获取配置和状态
  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [configRes, statusRes] = await Promise.all([getGeoIPConfig(), getGeoIPStatus()]);

      if (configRes.code === 200 && configRes.data) {
        setConfig({
          downloadUrl: configRes.data.downloadUrl || DEFAULT_DOWNLOAD_URL,
          useProxy: configRes.data.useProxy || false,
          proxyLink: configRes.data.proxyLink || '',
          lastUpdate: configRes.data.lastUpdate || ''
        });
        if (configRes.data.proxyLink) {
          setSelectedNode(configRes.data.proxyLink);
        }
      }

      if (statusRes.code === 200 && statusRes.data) {
        setStatus(statusRes.data);
      }
    } catch (error) {
      console.error('获取 GeoIP 信息失败:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  // 轮询下载状态
  useEffect(() => {
    let interval;
    if (open && status.downloading) {
      interval = setInterval(async () => {
        try {
          const res = await getGeoIPStatus();
          if (res.code === 200 && res.data) {
            setStatus(res.data);
            if (!res.data.downloading) {
              // 下载完成或失败
              if (res.data.error) {
                showMessage?.(t('components.geoIpSettings.messages.downloadFailedWithReason', { reason: res.data.error }), 'error');
              } else if (res.data.available) {
                showMessage?.(t('components.geoIpSettings.messages.downloadSuccess'), 'success');
              }
            }
          }
        } catch (error) {
          console.error('获取下载状态失败:', error);
        }
      }, 1000);
    }
    return () => clearInterval(interval);
  }, [open, status.downloading, showMessage, t]);

  // 初始化
  useEffect(() => {
    if (open) {
      fetchData();
    }
  }, [open, fetchData]);

  // 当启用代理时加载节点列表
  useEffect(() => {
    if (config.useProxy && proxyNodes.length === 0) {
      fetchProxyNodes();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [config.useProxy]);

  const fetchProxyNodes = async () => {
    setLoadingNodes(true);
    try {
      const res = await getNodes({ pageSize: 200 });
      if (res.data) {
        const items = res.data.items || res.data || [];
        setProxyNodes(items);
      }
    } catch (error) {
      console.error('获取代理节点失败:', error);
    } finally {
      setLoadingNodes(false);
    }
  };

  const handleNodeChange = (node) => {
    setSelectedNode(node);
    if (node) {
      const link = typeof node === 'string' ? node : node.Link;
      setConfig({ ...config, proxyLink: link });
    } else {
      setConfig({ ...config, proxyLink: '' });
    }
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      const res = await saveGeoIPConfig({
        downloadUrl: config.downloadUrl,
        useProxy: config.useProxy,
        proxyLink: config.proxyLink
      });
      if (res.code === 200) {
        showMessage?.(t('components.geoIpSettings.messages.saveSuccess'), 'success');
      } else {
        showMessage?.(res.msg || t('components.geoIpSettings.messages.saveFailed'), 'error');
      }
    } catch (error) {
      showMessage?.(
        t('components.geoIpSettings.messages.saveFailedWithReason', { reason: error.response?.data?.msg || error.message }),
        'error'
      );
    } finally {
      setSaving(false);
    }
  };

  const handleDownload = async () => {
    // 先保存配置
    await handleSave();

    try {
      const res = await downloadGeoIP();
      if (res.code === 200) {
        showMessage?.(t('components.geoIpSettings.messages.downloadStarted'), 'info');
        setStatus((prev) => ({ ...prev, downloading: true, progress: 0, error: '' }));
      } else {
        showMessage?.(res.msg || t('components.geoIpSettings.messages.startDownloadFailed'), 'error');
      }
    } catch (error) {
      showMessage?.(
        t('components.geoIpSettings.messages.startDownloadFailedWithReason', { reason: error.response?.data?.msg || error.message }),
        'error'
      );
    }
  };

  const handleRestoreDefault = () => {
    setConfig({ ...config, downloadUrl: DEFAULT_DOWNLOAD_URL });
  };

  const handleStopDownload = async () => {
    try {
      const res = await stopGeoIPDownload();
      if (res.code === 200) {
        showMessage?.(t('components.geoIpSettings.messages.stopSignalSent'), 'info');
      } else {
        showMessage?.(res.msg || t('components.geoIpSettings.messages.stopFailed'), 'error');
      }
    } catch (error) {
      showMessage?.(
        t('components.geoIpSettings.messages.stopFailedWithReason', { reason: error.response?.data?.msg || error.message }),
        'error'
      );
    }
  };

  const getStatusPanelSx = (accentColor) => ({
    p: 2,
    borderRadius: 2,
    backgroundColor: isDark ? nestedPanelSurface : withAlpha(accentColor, 0.08),
    backgroundImage: isDark ? `linear-gradient(180deg, ${withAlpha(accentColor, 0.12)} 0%, ${nestedPanelSurface} 100%)` : 'none',
    border: '1px solid',
    borderColor: withAlpha(accentColor, isDark ? 0.3 : 0.18),
    boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(theme.palette.common.white, 0.04)}` : 'none'
  });

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{
        sx: {
          border: '1px solid',
          borderColor: panelBorder,
          bgcolor: dialogSurface,
          backgroundImage: dialogSurfaceGradient,
          boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(theme.palette.common.white, 0.04)}` : undefined
        }
      }}
    >
      <DialogTitle sx={{ bgcolor: mutedPanelSurface, borderBottom: '1px solid', borderColor: panelBorder }}>
        <Stack direction="row" alignItems="center" justifyContent="space-between">
          <Stack direction="row" alignItems="center" spacing={1}>
            <PublicIcon sx={{ color: 'primary.main' }} />
            <Typography variant="h4" sx={{ color: primaryText }}>
              {t('components.geoIpSettings.title')}
            </Typography>
          </Stack>
          <IconButton onClick={onClose} size="small">
            <CloseIcon />
          </IconButton>
        </Stack>
      </DialogTitle>

      <DialogContent dividers sx={{ bgcolor: dialogSurface, borderColor: panelBorder }}>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress />
          </Box>
        ) : (
          <Stack spacing={2.5}>
            {/* 状态区域 */}
            <Box sx={getStatusPanelSx(status.available ? theme.palette.success.main : theme.palette.warning.main)}>
              <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 1 }}>
                {status.available ? <CheckCircleIcon sx={{ color: 'success.main' }} /> : <ErrorIcon sx={{ color: 'warning.main' }} />}
                <Typography variant="subtitle1" fontWeight={600} sx={{ color: primaryText }}>
                  {status.available ? t('components.geoIpSettings.status.installed') : t('components.geoIpSettings.status.notInstalled')}
                </Typography>
              </Stack>

              {status.available && (
                <Stack spacing={0.5}>
                  <Typography variant="body2" sx={{ color: secondaryText }}>
                    {t('components.geoIpSettings.status.fileSize', { value: status.sizeFormatted })}
                  </Typography>
                  <Typography variant="body2" sx={{ color: secondaryText }}>
                    {t('components.geoIpSettings.status.updatedAt', { value: status.modTime || config.lastUpdate || t('common.unknown') })}
                  </Typography>
                </Stack>
              )}

              {!status.available && (
                <Typography variant="body2" sx={{ color: secondaryText }}>
                  {t('components.geoIpSettings.status.missingDescription')}
                </Typography>
              )}
            </Box>

            {/* 下载进度 */}
            {status.downloading && (
              <Box>
                <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 1 }}>
                  <CircularProgress size={16} />
                  <Typography variant="body2">
                    {t(
                      status.source === 'auto'
                        ? 'components.geoIpSettings.download.autoDownloading'
                        : 'components.geoIpSettings.download.downloading',
                      {
                        progress: status.progress
                      }
                    )}
                  </Typography>
                  <Button size="small" color="error" onClick={handleStopDownload}>
                    {t('components.geoIpSettings.actions.stop')}
                  </Button>
                </Stack>
                <LinearProgress variant="determinate" value={status.progress} />
                {status.source === 'auto' && (
                  <Typography variant="caption" sx={{ mt: 0.5, display: 'block', color: secondaryText }}>
                    {t('components.geoIpSettings.download.autoHint')}
                  </Typography>
                )}
              </Box>
            )}

            {/* 下载错误 */}
            {status.error && (
              <Alert severity="error">
                <Typography variant="body2">{status.error}</Typography>
              </Alert>
            )}

            <Divider />

            {/* 下载地址 */}
            <TextField
              fullWidth
              label={t('components.geoIpSettings.fields.downloadUrl')}
              value={config.downloadUrl}
              onChange={(e) => setConfig({ ...config, downloadUrl: e.target.value })}
              placeholder={DEFAULT_DOWNLOAD_URL}
              InputProps={{
                endAdornment: (
                  <IconButton onClick={handleRestoreDefault} size="small" title={t('components.geoIpSettings.actions.restoreDefault')}>
                    <RestoreIcon />
                  </IconButton>
                )
              }}
              helperText={t('components.geoIpSettings.helpers.downloadUrl')}
            />

            {/* 代理设置 */}
            <FormControlLabel
              control={<Switch checked={config.useProxy} onChange={(e) => setConfig({ ...config, useProxy: e.target.checked })} />}
              label={t('components.geoIpSettings.fields.useProxy')}
            />

            <Collapse in={config.useProxy}>
              <Box sx={{ mt: 1 }}>
                <SearchableNodeSelect
                  nodes={proxyNodes}
                  loading={loadingNodes}
                  value={selectedNode}
                  onChange={handleNodeChange}
                  displayField="Name"
                  valueField="Link"
                  label={t('components.geoIpSettings.fields.proxyNode')}
                  placeholder={t('components.geoIpSettings.placeholders.proxyNode')}
                  helperText={t('components.geoIpSettings.helpers.proxyNode')}
                  freeSolo={true}
                  limit={50}
                />
              </Box>
            </Collapse>

            <Divider />

            {/* 说明 */}
            <Alert severity="info">
              <Typography variant="body2" sx={{ mb: 1, color: primaryText }}>
                <strong>{t('components.geoIpSettings.notice.title')}</strong>
              </Typography>
              <Typography variant="body2" component="ul" sx={{ m: 0, pl: 2, color: secondaryText }}>
                <li>{t('components.geoIpSettings.notice.mmdb')}</li>
                <li>{t('components.geoIpSettings.notice.required')}</li>
                <li>{t('components.geoIpSettings.notice.update')}</li>
              </Typography>
            </Alert>
          </Stack>
        )}
      </DialogContent>

      <DialogActions sx={{ px: 3, py: 2, bgcolor: mutedPanelSurface, borderTop: '1px solid', borderColor: panelBorder }}>
        <Button onClick={handleSave} disabled={saving || status.downloading} startIcon={<SaveIcon />}>
          {t('components.geoIpSettings.actions.saveConfig')}
        </Button>
        <Button variant="contained" onClick={handleDownload} disabled={status.downloading} startIcon={<DownloadIcon />}>
          {status.available ? t('components.geoIpSettings.actions.updateDatabase') : t('components.geoIpSettings.actions.downloadDatabase')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
