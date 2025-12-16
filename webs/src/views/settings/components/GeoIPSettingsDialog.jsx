import { useState, useEffect, useCallback } from 'react';

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

// 默认下载地址
const DEFAULT_DOWNLOAD_URL = 'https://git.io/GeoLite2-City.mmdb';

// ==============================|| GeoIP 设置对话框 ||============================== //

export default function GeoIPSettingsDialog({ open, onClose, showMessage }) {
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
                showMessage?.('下载失败: ' + res.data.error, 'error');
              } else if (res.data.available) {
                showMessage?.('GeoIP 数据库下载成功', 'success');
              }
            }
          }
        } catch (error) {
          console.error('获取下载状态失败:', error);
        }
      }, 1000);
    }
    return () => clearInterval(interval);
  }, [open, status.downloading, showMessage]);

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
      const res = await getNodes({ minSpeed: 0.01, pageSize: 200 });
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
        showMessage?.('配置已保存', 'success');
      } else {
        showMessage?.(res.msg || '保存失败', 'error');
      }
    } catch (error) {
      showMessage?.('保存失败: ' + (error.response?.data?.msg || error.message), 'error');
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
        showMessage?.('开始下载 GeoIP 数据库...', 'info');
        setStatus((prev) => ({ ...prev, downloading: true, progress: 0, error: '' }));
      } else {
        showMessage?.(res.msg || '启动下载失败', 'error');
      }
    } catch (error) {
      showMessage?.('启动下载失败: ' + (error.response?.data?.msg || error.message), 'error');
    }
  };

  const handleRestoreDefault = () => {
    setConfig({ ...config, downloadUrl: DEFAULT_DOWNLOAD_URL });
  };

  const handleStopDownload = async () => {
    try {
      const res = await stopGeoIPDownload();
      if (res.code === 200) {
        showMessage?.('已发送停止信号', 'info');
      } else {
        showMessage?.(res.msg || '停止失败', 'error');
      }
    } catch (error) {
      showMessage?.('停止失败: ' + (error.response?.data?.msg || error.message), 'error');
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>
        <Stack direction="row" alignItems="center" justifyContent="space-between">
          <Stack direction="row" alignItems="center" spacing={1}>
            <PublicIcon sx={{ color: 'primary.main' }} />
            <Typography variant="h4">GeoIP 数据库设置</Typography>
          </Stack>
          <IconButton onClick={onClose} size="small">
            <CloseIcon />
          </IconButton>
        </Stack>
      </DialogTitle>

      <DialogContent dividers>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress />
          </Box>
        ) : (
          <Stack spacing={2.5}>
            {/* 状态区域 */}
            <Box
              sx={{
                p: 2,
                borderRadius: 1,
                backgroundColor: status.available ? 'success.light' : 'warning.light',
                border: '1px solid',
                borderColor: status.available ? 'success.main' : 'warning.main'
              }}
            >
              <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 1 }}>
                {status.available ? <CheckCircleIcon sx={{ color: 'success.main' }} /> : <ErrorIcon sx={{ color: 'warning.main' }} />}
                <Typography variant="subtitle1" fontWeight={600}>
                  {status.available ? '数据库已安装' : '数据库未安装'}
                </Typography>
              </Stack>

              {status.available && (
                <Stack spacing={0.5}>
                  <Typography variant="body2" color="textSecondary">
                    文件大小: {status.sizeFormatted}
                  </Typography>
                  <Typography variant="body2" color="textSecondary">
                    更新时间: {status.modTime || config.lastUpdate || '未知'}
                  </Typography>
                </Stack>
              )}

              {!status.available && (
                <Typography variant="body2" color="textSecondary">
                  GeoIP 数据库用于 IP 地理位置查询和落地检测，请下载安装。
                </Typography>
              )}
            </Box>

            {/* 下载进度 */}
            {status.downloading && (
              <Box>
                <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 1 }}>
                  <CircularProgress size={16} />
                  <Typography variant="body2">
                    {status.source === 'auto' ? '系统自动下载中' : '正在下载'}... {status.progress}%
                  </Typography>
                  <Button size="small" color="error" onClick={handleStopDownload}>
                    停止
                  </Button>
                </Stack>
                <LinearProgress variant="determinate" value={status.progress} />
                {status.source === 'auto' && (
                  <Typography variant="caption" color="textSecondary" sx={{ mt: 0.5, display: 'block' }}>
                    系统启动时检测到 GeoIP 数据库缺失，正在自动下载。您可以停止后配置代理重新下载。
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
              label="下载地址"
              value={config.downloadUrl}
              onChange={(e) => setConfig({ ...config, downloadUrl: e.target.value })}
              placeholder={DEFAULT_DOWNLOAD_URL}
              InputProps={{
                endAdornment: (
                  <IconButton onClick={handleRestoreDefault} size="small" title="恢复默认">
                    <RestoreIcon />
                  </IconButton>
                )
              }}
              helperText="MaxMind GeoLite2-City 数据库下载地址"
            />

            {/* 代理设置 */}
            <FormControlLabel
              control={<Switch checked={config.useProxy} onChange={(e) => setConfig({ ...config, useProxy: e.target.checked })} />}
              label="使用代理下载"
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
                  label="选择代理节点"
                  placeholder="留空则自动选择最佳节点"
                  helperText="如果未选择具体代理，系统将自动选择延迟最低且速度最快的节点"
                  freeSolo={true}
                  limit={50}
                />
              </Box>
            </Collapse>

            <Divider />

            {/* 说明 */}
            <Alert severity="info">
              <Typography variant="body2" sx={{ mb: 1 }}>
                <strong>注意事项：</strong>
              </Typography>
              <Typography variant="body2" component="ul" sx={{ m: 0, pl: 2 }}>
                <li>数据库必须是 MaxMind 的 mmdb 格式且包含 city 数据 (GeoLite2-City)</li>
                <li>如果没有 GeoIP 数据库，IP 地理位置查询和落地检测功能将不可用</li>
                <li>建议定期更新数据库以获得更准确的地理位置信息</li>
              </Typography>
            </Alert>
          </Stack>
        )}
      </DialogContent>

      <DialogActions sx={{ px: 3, py: 2 }}>
        <Button onClick={handleSave} disabled={saving || status.downloading} startIcon={<SaveIcon />}>
          保存配置
        </Button>
        <Button variant="contained" onClick={handleDownload} disabled={status.downloading} startIcon={<DownloadIcon />}>
          {status.available ? '更新数据库' : '下载数据库'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
