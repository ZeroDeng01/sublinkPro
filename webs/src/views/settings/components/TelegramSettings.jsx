import { useState, useEffect } from 'react';

// material-ui
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import Alert from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import InputAdornment from '@mui/material/InputAdornment';
import Collapse from '@mui/material/Collapse';
import Divider from '@mui/material/Divider';
import Tooltip from '@mui/material/Tooltip';
import { Trans, useTranslation } from 'react-i18next';

// qrcode
import { QRCodeSVG } from 'qrcode.react';

// icons
import TelegramIcon from '@mui/icons-material/Telegram';
import SaveIcon from '@mui/icons-material/Save';
import SendIcon from '@mui/icons-material/Send';
import RefreshIcon from '@mui/icons-material/Refresh';
import VisibilityIcon from '@mui/icons-material/Visibility';
import VisibilityOffIcon from '@mui/icons-material/VisibilityOff';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';

// project imports
import { getTelegramConfig, saveTelegramConfig, testTelegramConnection, reconnectTelegram } from 'api/telegram';
import { getNodes } from 'api/nodes';
import SearchableNodeSelect from 'components/SearchableNodeSelect';
import NotificationEventSelector from './NotificationEventSelector';

// ==============================|| Telegram 设置组件 ||============================== //

export default function TelegramSettings({ showMessage, loading, setLoading }) {
  const { t } = useTranslation();
  const [form, setForm] = useState({
    enabled: false,
    botToken: '',
    chatId: '',
    useProxy: false,
    proxyLink: '',
    eventKeys: []
  });
  const [showToken, setShowToken] = useState(false);
  const [status, setStatus] = useState({ connected: false, error: '', botUsername: '', botId: 0 });
  const [eventOptions, setEventOptions] = useState([]);

  // 代理节点选择
  const [proxyNodes, setProxyNodes] = useState([]);
  const [loadingNodes, setLoadingNodes] = useState(false);
  const [selectedNode, setSelectedNode] = useState(null);

  useEffect(() => {
    fetchConfig();
  }, []);

  // 当启用代理时加载节点列表
  useEffect(() => {
    if (form.useProxy && proxyNodes.length === 0) {
      fetchProxyNodes();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [form.useProxy]);

  const fetchConfig = async () => {
    try {
      const response = await getTelegramConfig();
      if (response.data) {
        setForm({
          enabled: response.data.enabled || false,
          botToken: response.data.botToken || '',
          chatId: response.data.chatId ? String(response.data.chatId) : '',
          useProxy: response.data.useProxy || false,
          proxyLink: response.data.proxyLink || '',
          eventKeys: response.data.eventKeys || []
        });
        setEventOptions(response.data.eventOptions || []);
        setStatus({
          connected: response.data.connected || false,
          error: response.data.lastError || '',
          botUsername: response.data.botUsername || '',
          botId: response.data.botId || 0
        });

        // 如果有已保存的代理链接，设置为选中值
        if (response.data.proxyLink) {
          setSelectedNode(response.data.proxyLink);
        }
      }
    } catch (error) {
      console.error('获取 Telegram 配置失败:', error);
    }
  };

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
      setForm({ ...form, proxyLink: link });
    } else {
      setForm({ ...form, proxyLink: '' });
    }
  };

  const handleSave = async () => {
    if (form.enabled && !form.botToken) {
      showMessage(t('settings.telegramPanel.messages.botTokenRequired'), 'warning');
      return;
    }

    setLoading(true);
    try {
      await saveTelegramConfig({
        enabled: form.enabled,
        botToken: form.botToken,
        chatId: form.chatId ? parseInt(form.chatId, 10) : 0,
        useProxy: form.useProxy,
        proxyLink: form.proxyLink,
        eventKeys: form.eventKeys
      });
      showMessage(t('settings.telegramPanel.messages.saveSuccess'));
      fetchConfig();
    } catch (error) {
      showMessage(t('settings.telegramPanel.messages.saveFailed', { message: error.response?.data?.message || error.message }), 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleTest = async () => {
    if (!form.botToken) {
      showMessage(t('settings.telegramPanel.messages.botTokenRequired'), 'warning');
      return;
    }

    setLoading(true);
    try {
      const response = await testTelegramConnection({
        botToken: form.botToken,
        chatId: form.chatId ? parseInt(form.chatId, 10) : 0,
        useProxy: form.useProxy,
        proxyLink: form.proxyLink
      });
      if (response.data?.messageSent) {
        showMessage(t('settings.telegramPanel.messages.testSent'));
      } else {
        showMessage(t('settings.telegramPanel.messages.testSuccess'));
      }
    } catch (error) {
      showMessage(t('settings.telegramPanel.messages.testFailed', { message: error.response?.data?.message || error.message }), 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleReconnect = async () => {
    setLoading(true);
    try {
      await reconnectTelegram();
      showMessage(t('settings.telegramPanel.messages.reconnectSuccess'));
      fetchConfig();
    } catch (error) {
      showMessage(
        t('settings.telegramPanel.messages.reconnectFailed', { message: error.response?.data?.message || error.message }),
        'error'
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader
        title={t('settings.telegramPanel.title')}
        avatar={<TelegramIcon sx={{ color: '#0088cc' }} />}
        action={
          <Stack direction="row" spacing={1} alignItems="center">
            {form.enabled && (
              <Chip
                icon={status.connected ? <CheckCircleIcon /> : <ErrorIcon />}
                label={status.connected ? t('settings.telegramPanel.status.connected') : t('settings.telegramPanel.status.disconnected')}
                color={status.connected ? 'success' : 'error'}
                size="small"
                variant="outlined"
              />
            )}
            <FormControlLabel
              control={<Switch checked={form.enabled} onChange={(e) => setForm({ ...form, enabled: e.target.checked })} />}
              label={form.enabled ? t('common.enabled') : t('common.disabled')}
            />
          </Stack>
        }
      />
      <CardContent>
        <Stack spacing={2.5}>
          <Alert severity="info">
            <Typography variant="body2">
              <Trans
                i18nKey="settings.telegramPanel.botFatherHint"
                components={{
                  botFather: (
                    <a
                      href="https://t.me/BotFather"
                      target="_blank"
                      rel="noopener noreferrer"
                      style={{ color: '#0088cc', fontWeight: 600, textDecoration: 'none' }}
                    />
                  )
                }}
              />
            </Typography>
          </Alert>

          {/* 机器人连接信息 */}
          {status.connected && status.botUsername && (
            <Box
              sx={{
                backgroundColor: 'rgba(0, 136, 204, 0.08)',
                borderColor: 'rgba(0, 136, 204, 0.3)',
                border: '1px solid',
                borderRadius: 1,
                p: 2
              }}
            >
              <Stack direction="row" alignItems="center" spacing={2} flexWrap="wrap">
                <TelegramIcon sx={{ color: '#0088cc' }} />
                <Typography variant="body2" sx={{ fontWeight: 500 }}>
                  {t('settings.telegramPanel.connectedAs')}
                </Typography>
                <Chip
                  label={`@${status.botUsername}`}
                  size="small"
                  clickable
                  component="a"
                  href={`https://t.me/${status.botUsername}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  sx={{
                    backgroundColor: '#0088cc',
                    color: 'white',
                    fontWeight: 600,
                    '&:hover': {
                      backgroundColor: '#006699'
                    }
                  }}
                />
                <Typography variant="caption" color="textSecondary">
                  ID: {status.botId}
                </Typography>
              </Stack>

              {/* 二维码区域 */}
              <Box sx={{ mt: 2, display: 'flex', alignItems: 'center', gap: 2, flexWrap: 'wrap' }}>
                <Tooltip title={t('settings.telegramPanel.qrTooltip')} placement="top">
                  <Box
                    component="a"
                    href={`https://t.me/${status.botUsername}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    sx={{
                      display: 'inline-block',
                      p: 1.5,
                      bgcolor: 'white',
                      borderRadius: 1,
                      boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                      transition: 'transform 0.2s, box-shadow 0.2s',
                      '&:hover': {
                        transform: 'scale(1.02)',
                        boxShadow: '0 4px 16px rgba(0,136,204,0.3)'
                      }
                    }}
                  >
                    <QRCodeSVG
                      value={`https://t.me/${status.botUsername}`}
                      size={100}
                      level="M"
                      bgColor="#ffffff"
                      fgColor="#0088cc"
                      imageSettings={{
                        src: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iIzAwODhjYyI+PHBhdGggZD0iTTEyIDJDNi40OCAyIDIgNi40OCAyIDEyczQuNDggMTAgMTAgMTAgMTAtNC40OCAxMC0xMFMxNy41MiAyIDEyIDJ6bTQuNjQgNi44bC0xLjY0IDcuNzNjLS4xMi41NC0uNDQuNjctLjktLjQybC0yLjQ4LTEuODMtMS4yIDEuMTVjLS4xMy4xMy0uMjQuMjQtLjUtLjI0bC0uMjgtMi44LTQuNzItMS41N2MtMS4wMy0uMzItMS4wNS0xLjAzLjIyLTEuNTNsOS40My0zLjY0Yy44Ni0uMzIgMS42LjIxIDEuMzIgMS4zOHoiLz48L3N2Zz4=',
                        height: 24,
                        width: 24,
                        excavate: true
                      }}
                    />
                  </Box>
                </Tooltip>
                <Box>
                  <Typography variant="body2" color="textSecondary" sx={{ mb: 0.5 }}>
                    {t('settings.telegramPanel.qrTitle')}
                  </Typography>
                  <Typography variant="caption" color="textSecondary">
                    {t('settings.telegramPanel.qrHelper')}
                  </Typography>
                </Box>
              </Box>
            </Box>
          )}

          <TextField
            fullWidth
            label="Bot Token"
            type={showToken ? 'text' : 'password'}
            value={form.botToken}
            onChange={(e) => setForm({ ...form, botToken: e.target.value })}
            placeholder="123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
            InputProps={{
              endAdornment: (
                <InputAdornment position="end">
                  <IconButton onClick={() => setShowToken(!showToken)} edge="end">
                    {showToken ? <VisibilityOffIcon /> : <VisibilityIcon />}
                  </IconButton>
                </InputAdornment>
              )
            }}
          />

          <TextField
            fullWidth
            label="Chat ID"
            value={form.chatId}
            onChange={(e) => setForm({ ...form, chatId: e.target.value })}
            placeholder={t('settings.telegramPanel.chatIdPlaceholder')}
            helperText={t('settings.telegramPanel.chatIdHelper')}
          />

          <Divider />

          <FormControlLabel
            control={<Switch checked={form.useProxy} onChange={(e) => setForm({ ...form, useProxy: e.target.checked })} />}
            label={t('settings.telegramPanel.useProxy')}
          />

          <Collapse in={form.useProxy}>
            <Box sx={{ mt: 1 }}>
              <SearchableNodeSelect
                nodes={proxyNodes}
                loading={loadingNodes}
                value={selectedNode}
                onChange={handleNodeChange}
                displayField="Name"
                valueField="Link"
                label={t('settings.telegramPanel.proxyNode')}
                placeholder={t('settings.telegramPanel.proxyPlaceholder')}
                helperText={t('settings.telegramPanel.proxyHelper')}
                freeSolo={true}
                limit={50}
              />
            </Box>
          </Collapse>

          {status.error && (
            <Alert severity="error">
              <Typography variant="body2">{status.error}</Typography>
            </Alert>
          )}

          <Divider />

          <NotificationEventSelector
            value={form.eventKeys}
            eventOptions={eventOptions}
            disabled={loading}
            description={t('settings.telegramPanel.eventsDescription')}
            onChange={(eventKeys) => setForm((prev) => ({ ...prev, eventKeys }))}
          />

          <Stack direction="row" spacing={2} flexWrap="wrap">
            <Button variant="outlined" color="success" onClick={handleTest} disabled={loading || !form.botToken} startIcon={<SendIcon />}>
              {t('settings.telegramPanel.actions.test')}
            </Button>
            {form.enabled && (
              <Button variant="outlined" onClick={handleReconnect} disabled={loading} startIcon={<RefreshIcon />}>
                {t('settings.telegramPanel.actions.reconnect')}
              </Button>
            )}
            <Button variant="contained" onClick={handleSave} disabled={loading} startIcon={<SaveIcon />}>
              {t('settings.telegramPanel.actions.save')}
            </Button>
          </Stack>

          <Divider />

          <Box>
            <Typography variant="subtitle2" gutterBottom>
              {t('settings.telegramPanel.commands.title')}
            </Typography>
            <Typography variant="body2" color="textSecondary" component="div">
              <ul style={{ margin: 0, paddingLeft: 20 }}>
                <li>/start - {t('settings.telegramPanel.commands.start')}</li>
                <li>/stats - {t('settings.telegramPanel.commands.stats')}</li>
                <li>/monitor - {t('settings.telegramPanel.commands.monitor')}</li>
                <li>/speedtest - {t('settings.telegramPanel.commands.speedtest')}</li>
                <li>/subscriptions - {t('settings.telegramPanel.commands.subscriptions')}</li>
                <li>/nodes - {t('settings.telegramPanel.commands.nodes')}</li>
                <li>/tags - {t('settings.telegramPanel.commands.tags')}</li>
                <li>/tasks - {t('settings.telegramPanel.commands.tasks')}</li>
              </ul>
            </Typography>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
}
