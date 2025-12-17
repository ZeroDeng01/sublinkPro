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
import Grid from '@mui/material/Grid';
import Paper from '@mui/material/Paper';
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';

// Monaco Editor
import Editor from '@monaco-editor/react';

// icons
import WebhookIcon from '@mui/icons-material/Webhook';
import SendIcon from '@mui/icons-material/Send';
import SaveIcon from '@mui/icons-material/Save';

// project imports
import { getWebhookConfig, updateWebhookConfig, testWebhook } from 'api/settings';

// ==============================|| Webhook 设置组件 ||============================== //

export default function WebhookSettings({ showMessage, loading, setLoading }) {
  // Webhook 表单
  const [webhookForm, setWebhookForm] = useState({
    webhookUrl: '',
    webhookMethod: 'POST',
    webhookContentType: 'application/json',
    webhookHeaders: '',
    webhookBody: '',
    webhookEnabled: false
  });

  useEffect(() => {
    fetchWebhookConfig();
  }, []);

  const fetchWebhookConfig = async () => {
    try {
      const response = await getWebhookConfig();
      if (response.data) {
        setWebhookForm({
          webhookUrl: response.data.webhookUrl || '',
          webhookMethod: response.data.webhookMethod || 'POST',
          webhookContentType: response.data.webhookContentType || 'application/json',
          webhookHeaders: response.data.webhookHeaders || '',
          webhookBody: response.data.webhookBody || '',
          webhookEnabled: response.data.webhookEnabled || false
        });
      }
    } catch (error) {
      console.error('获取 Webhook 配置失败:', error);
    }
  };

  // === Webhook 操作 ===
  const handleTestWebhook = async () => {
    if (!webhookForm.webhookUrl) {
      showMessage('请输入 Webhook URL', 'warning');
      return;
    }

    setLoading(true);
    try {
      await testWebhook(webhookForm);
      showMessage('Webhook 测试发送成功');
    } catch (error) {
      showMessage('测试失败: ' + (error.response?.data?.message || error.message), 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateWebhook = async () => {
    if (!webhookForm.webhookUrl && webhookForm.webhookEnabled) {
      showMessage('启用 Webhook 时需填写 URL', 'warning');
      return;
    }

    setLoading(true);
    try {
      await updateWebhookConfig(webhookForm);
      showMessage('Webhook 设置保存成功');
    } catch (error) {
      showMessage('保存失败: ' + (error.response?.data?.message || error.message), 'error');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader
        title="Webhook 设置"
        avatar={<WebhookIcon color="primary" />}
        action={
          <FormControlLabel
            control={
              <Switch
                checked={webhookForm.webhookEnabled}
                onChange={(e) => setWebhookForm({ ...webhookForm, webhookEnabled: e.target.checked })}
              />
            }
            label={webhookForm.webhookEnabled ? '启用' : '禁用'}
          />
        }
      />
      <CardContent>
        <Stack spacing={2}>
          <TextField
            fullWidth
            label="Webhook URL"
            value={webhookForm.webhookUrl}
            onChange={(e) => setWebhookForm({ ...webhookForm, webhookUrl: e.target.value })}
            placeholder="https://example.com/webhook"
          />

          <Grid container spacing={2}>
            <Grid item xs={6}>
              <FormControl fullWidth>
                <InputLabel>请求方法</InputLabel>
                <Select
                  value={webhookForm.webhookMethod}
                  label="请求方法"
                  onChange={(e) => setWebhookForm({ ...webhookForm, webhookMethod: e.target.value })}
                >
                  <MenuItem value="POST">POST</MenuItem>
                  <MenuItem value="GET">GET</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={6}>
              <FormControl fullWidth>
                <InputLabel>Content-Type</InputLabel>
                <Select
                  value={webhookForm.webhookContentType}
                  label="Content-Type"
                  onChange={(e) => setWebhookForm({ ...webhookForm, webhookContentType: e.target.value })}
                >
                  <MenuItem value="application/json">application/json</MenuItem>
                  <MenuItem value="application/x-www-form-urlencoded">application/x-www-form-urlencoded</MenuItem>
                </Select>
              </FormControl>
            </Grid>
          </Grid>

          <Box>
            <Typography variant="subtitle2" gutterBottom>
              Headers (JSON)
            </Typography>
            <Paper variant="outlined" sx={{ overflow: 'hidden' }}>
              <Editor
                height="150px"
                language="json"
                value={webhookForm.webhookHeaders}
                onChange={(value) => setWebhookForm({ ...webhookForm, webhookHeaders: value || '' })}
                theme="vs-dark"
                options={{
                  minimap: { enabled: false },
                  fontSize: 13,
                  lineNumbers: 'on',
                  lineNumbersMinChars: 3,
                  scrollBeyondLastLine: false,
                  automaticLayout: true,
                  wordWrap: 'on',
                  folding: false,
                  glyphMargin: false,
                  padding: { top: 15, bottom: 15 },
                  formatOnPaste: true,
                  formatOnType: true
                }}
              />
            </Paper>
          </Box>

          <Box>
            <Typography variant="subtitle2" gutterBottom>
              Body Template (JSON)
            </Typography>
            <Paper variant="outlined" sx={{ overflow: 'hidden' }}>
              <Editor
                height="200px"
                language="json"
                value={webhookForm.webhookBody}
                onChange={(value) => setWebhookForm({ ...webhookForm, webhookBody: value || '' })}
                theme="vs-dark"
                options={{
                  minimap: { enabled: false },
                  fontSize: 13,
                  lineNumbers: 'on',
                  lineNumbersMinChars: 3,
                  scrollBeyondLastLine: false,
                  automaticLayout: true,
                  wordWrap: 'on',
                  folding: false,
                  glyphMargin: false,
                  padding: { top: 15, bottom: 15 },
                  formatOnPaste: true,
                  formatOnType: true
                }}
              />
            </Paper>
            <Alert severity="info" sx={{ mt: 1 }}>
              <Typography variant="caption">
                支持变量: {'{{title}}'} 消息标题, {'{{message}}'} 消息内容, {'{{event}}'} 事件类型, {'{{time}}'} 事件时间, {'{{json .}}'}{' '}
                原始json数据
                <br />
                例如 Bark URL: https://api.day.app/key/{'{{title}}'}/{'{{message}}'}
              </Typography>
            </Alert>
          </Box>

          <Stack direction="row" spacing={2}>
            <Button variant="outlined" color="success" onClick={handleTestWebhook} disabled={loading} startIcon={<SendIcon />}>
              测试 Webhook
            </Button>
            <Button variant="contained" onClick={handleUpdateWebhook} disabled={loading} startIcon={<SaveIcon />}>
              保存 Webhook 设置
            </Button>
          </Stack>
        </Stack>
      </CardContent>
    </Card>
  );
}
