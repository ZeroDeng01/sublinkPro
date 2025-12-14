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

// ==============================|| Telegram 设置组件 ||============================== //

export default function TelegramSettings({ showMessage, loading, setLoading }) {
  const [form, setForm] = useState({
    enabled: false,
    botToken: '',
    chatId: '',
    useProxy: false,
    proxyLink: '',
    systemDomain: '' // 新增: 系统域名
  });
  const [showToken, setShowToken] = useState(false);
  const [status, setStatus] = useState({ connected: false, error: '', botUsername: '', botId: 0 });

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
        // 如果后端返回的 systemDomain 为空，则自动设置为当前页面域名
        let domain = response.data.systemDomain || '';
        if (!domain) {
          domain = window.location.origin;
        }

        setForm({
          enabled: response.data.enabled || false,
          botToken: response.data.botToken || '',
          chatId: response.data.chatId ? String(response.data.chatId) : '',
          useProxy: response.data.useProxy || false,
          proxyLink: response.data.proxyLink || '',
          systemDomain: domain
        });
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
      setForm({ ...form, proxyLink: link });
    } else {
      setForm({ ...form, proxyLink: '' });
    }
  };

  const handleSave = async () => {
    if (form.enabled && !form.botToken) {
      showMessage('请输入 Bot Token', 'warning');
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
        systemDomain: form.systemDomain // 保存系统域名
      });
      showMessage('保存成功');
      fetchConfig();
    } catch (error) {
      showMessage('保存失败: ' + (error.response?.data?.message || error.message), 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleTest = async () => {
    if (!form.botToken) {
      showMessage('请输入 Bot Token', 'warning');
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
        showMessage('连接成功，测试消息已发送');
      } else {
        showMessage('连接成功');
      }
    } catch (error) {
      showMessage('测试失败: ' + (error.response?.data?.message || error.message), 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleReconnect = async () => {
    setLoading(true);
    try {
      await reconnectTelegram();
      showMessage('重连成功');
      fetchConfig();
    } catch (error) {
      showMessage('重连失败: ' + (error.response?.data?.message || error.message), 'error');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader
        title="Telegram 机器人"
        avatar={<TelegramIcon sx={{ color: '#0088cc' }} />}
        action={
          <Stack direction="row" spacing={1} alignItems="center">
            {form.enabled && (
              <Chip
                icon={status.connected ? <CheckCircleIcon /> : <ErrorIcon />}
                label={status.connected ? '已连接' : '未连接'}
                color={status.connected ? 'success' : 'error'}
                size="small"
                variant="outlined"
              />
            )}
            <FormControlLabel
              control={<Switch checked={form.enabled} onChange={(e) => setForm({ ...form, enabled: e.target.checked })} />}
              label={form.enabled ? '启用' : '禁用'}
            />
          </Stack>
        }
      />
      <CardContent>
        <Stack spacing={2.5}>
          <Alert severity="info">
            <Typography variant="body2">
              使用{' '}
              <a
                href="https://t.me/BotFather"
                target="_blank"
                rel="noopener noreferrer"
                style={{ color: '#0088cc', fontWeight: 600, textDecoration: 'none' }}
              >
                @BotFather
              </a>{' '}
              创建机器人并获取 Token。 首次发送 /start 命令后，机器人会自动绑定您的 Chat ID。
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
                  机器人已连接:
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
                <Tooltip title="使用手机扫描二维码打开机器人" placement="top">
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
                    扫描二维码打开机器人
                  </Typography>
                  <Typography variant="caption" color="textSecondary">
                    或点击上方机器人名称直接跳转
                  </Typography>
                </Box>
              </Box>
            </Box>
          )}

          <TextField
            fullWidth
            label="远程访问域名 (用于生成订阅链接)"
            value={form.systemDomain}
            onChange={(e) => setForm({ ...form, systemDomain: e.target.value })}
            placeholder="例如: https://your-domain.com"
            helperText="Telegram Bot 返回的订阅链接将使用此域名作为前缀，默认为当前访问域名。"
          />

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
            placeholder="发送 /start 后自动获取"
            helperText="可留空，首次发送 /start 后自动绑定"
          />

          <Divider />

          <FormControlLabel
            control={<Switch checked={form.useProxy} onChange={(e) => setForm({ ...form, useProxy: e.target.checked })} />}
            label="使用代理连接 Telegram"
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
                label="选择代理节点"
                placeholder="留空则自动选择最佳节点"
                helperText="如果未选择具体代理，系统将自动选择延迟最低且速度最快的节点作为下载代理，你也可以输入外部代理地址"
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

          <Stack direction="row" spacing={2} flexWrap="wrap">
            <Button variant="outlined" color="success" onClick={handleTest} disabled={loading || !form.botToken} startIcon={<SendIcon />}>
              测试连接
            </Button>
            {form.enabled && (
              <Button variant="outlined" onClick={handleReconnect} disabled={loading} startIcon={<RefreshIcon />}>
                重新连接
              </Button>
            )}
            <Button variant="contained" onClick={handleSave} disabled={loading} startIcon={<SaveIcon />}>
              保存设置
            </Button>
          </Stack>

          <Divider />

          <Box>
            <Typography variant="subtitle2" gutterBottom>
              支持的命令
            </Typography>
            <Typography variant="body2" color="textSecondary" component="div">
              <ul style={{ margin: 0, paddingLeft: 20 }}>
                <li>/start - 开始使用</li>
                <li>/stats - 查看仪表盘统计</li>
                <li>/monitor - 查看系统监控</li>
                <li>/speedtest - 开始节点测速</li>
                <li>/subscriptions - 管理订阅</li>
                <li>/nodes - 查看节点信息</li>
                <li>/tags - 执行标签规则</li>
                <li>/tasks - 管理任务</li>
              </ul>
            </Typography>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
}
