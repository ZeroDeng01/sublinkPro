import { useCallback, useEffect, useMemo, useState } from 'react';
import { alpha, useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';

import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Grid from '@mui/material/Grid';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Alert from '@mui/material/Alert';
import TextField from '@mui/material/TextField';
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import CircularProgress from '@mui/material/CircularProgress';
import Snackbar from '@mui/material/Snackbar';

import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import RefreshIcon from '@mui/icons-material/Refresh';
import SaveIcon from '@mui/icons-material/Save';
import SendIcon from '@mui/icons-material/Send';
import WebhookIcon from '@mui/icons-material/Webhook';

import MainCard from 'ui-component/cards/MainCard';
import NotificationEventSelector from 'views/settings/components/NotificationEventSelector';
import { createWebhook, deleteWebhook, getWebhooks, testWebhookById, updateWebhook } from 'api/settings';

const createDefaultForm = () => ({
  id: null,
  name: '',
  url: '',
  method: 'POST',
  contentType: 'application/json',
  headers: '',
  body: '',
  enabled: true,
  eventKeys: []
});

const pickValue = (...values) => values.find((value) => value !== undefined && value !== null);

const normalizeWebhook = (item = {}) => ({
  id: item.id,
  name: item.name || item.title || item.remark || '',
  url: item.webhookUrl || item.url || '',
  method: item.webhookMethod || item.method || 'POST',
  contentType: item.webhookContentType || item.contentType || 'application/json',
  headers: item.webhookHeaders || item.headers || '',
  body: item.webhookBody || item.body || '',
  enabled: Boolean(pickValue(item.webhookEnabled, item.enabled, false)),
  eventKeys: Array.isArray(item.eventKeys) ? item.eventKeys : [],
  createdAt: item.createdAt,
  updatedAt: item.updatedAt,
  lastTestAt: item.lastTestAt
});

const toWebhookPayload = (form) => ({
  name: form.name.trim(),
  webhookUrl: form.url.trim(),
  webhookMethod: form.method,
  webhookContentType: form.contentType,
  webhookHeaders: form.headers,
  webhookBody: form.body,
  webhookEnabled: form.enabled,
  eventKeys: form.eventKeys
});

const formatDateTime = (value) => {
  if (!value) return '未记录';

  try {
    return new Date(value).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  } catch {
    return value;
  }
};

const validateJsonText = (value) => {
  if (!value.trim()) return true;

  try {
    JSON.parse(value);
    return true;
  } catch {
    return false;
  }
};

export default function WebhookManagementPage() {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  const [items, setItems] = useState([]);
  const [eventOptions, setEventOptions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [testingId, setTestingId] = useState(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState(null);
  const [form, setForm] = useState(createDefaultForm());
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  const showMessage = useCallback((message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  }, []);

  const fetchWebhooks = useCallback(async () => {
    setLoading(true);
    try {
      const response = await getWebhooks();
      const data = response.data || {};
      setItems((data.items || []).map(normalizeWebhook));
      setEventOptions(data.eventOptions || []);
    } catch (error) {
      showMessage(error.response?.data?.message || error.message || '获取 Webhook 列表失败', 'error');
    } finally {
      setLoading(false);
    }
  }, [showMessage]);

  useEffect(() => {
    fetchWebhooks();
  }, [fetchWebhooks]);

  const enabledCount = useMemo(() => items.filter((item) => item.enabled).length, [items]);

  const openCreateDialog = () => {
    setForm(createDefaultForm());
    setDialogOpen(true);
  };

  const openEditDialog = (item) => {
    setForm(normalizeWebhook(item));
    setDialogOpen(true);
  };

  const closeDialog = () => {
    if (submitting) return;
    setDialogOpen(false);
    setForm(createDefaultForm());
  };

  const handleSave = async () => {
    if (!form.url.trim()) {
      showMessage('请输入 Webhook URL', 'warning');
      return;
    }

    if (!validateJsonText(form.headers)) {
      showMessage('请求头必须是合法的 JSON', 'warning');
      return;
    }

    if (form.contentType === 'application/json' && form.body.trim() && !validateJsonText(form.body)) {
      showMessage('JSON 请求体格式不正确', 'warning');
      return;
    }

    setSubmitting(true);
    try {
      const payload = toWebhookPayload(form);
      if (form.id) {
        await updateWebhook(form.id, payload);
        showMessage('Webhook 已更新');
      } else {
        await createWebhook(payload);
        showMessage('Webhook 已创建');
      }
      setDialogOpen(false);
      setForm(createDefaultForm());
      await fetchWebhooks();
    } catch (error) {
      showMessage(error.response?.data?.message || error.message || '保存 Webhook 失败', 'error');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget?.id) return;

    setSubmitting(true);
    try {
      await deleteWebhook(deleteTarget.id);
      showMessage('Webhook 已删除');
      setDeleteTarget(null);
      await fetchWebhooks();
    } catch (error) {
      showMessage(error.response?.data?.message || error.message || '删除 Webhook 失败', 'error');
    } finally {
      setSubmitting(false);
    }
  };

  const handleToggleEnabled = async (item, enabled) => {
    try {
      await updateWebhook(item.id, toWebhookPayload({ ...normalizeWebhook(item), enabled }));
      setItems((prev) => prev.map((current) => (current.id === item.id ? { ...current, enabled } : current)));
      showMessage(enabled ? 'Webhook 已启用' : 'Webhook 已停用');
    } catch (error) {
      showMessage(error.response?.data?.message || error.message || '更新启用状态失败', 'error');
    }
  };

  const handleTest = async (item) => {
    setTestingId(item.id);
    try {
      const response = await testWebhookById(item.id);
      showMessage(response.data?.message || 'Webhook 测试发送成功');
      await fetchWebhooks();
    } catch (error) {
      showMessage(error.response?.data?.message || error.message || 'Webhook 测试失败', 'error');
    } finally {
      setTestingId(null);
    }
  };

  return (
    <MainCard title="Webhook 管理">
      <Stack spacing={3}>
        <Card variant="outlined" sx={{ borderRadius: 3 }}>
          <CardContent>
            <Stack spacing={2.5}>
              <Stack
                direction={{ xs: 'column', md: 'row' }}
                spacing={2}
                alignItems={{ md: 'center' }}
                justifyContent="space-between"
              >
                <Box>
                  <Typography variant="h4" sx={{ mb: 0.75 }}>
                    系统 Webhook 通知
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    集中管理多个通知目标。每个 Webhook 都可以单独设置请求方式、事件范围，并在保存后独立测试。
                  </Typography>
                </Box>

                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.25} sx={{ width: { xs: '100%', md: 'auto' } }}>
                  <Button
                    variant="outlined"
                    onClick={fetchWebhooks}
                    disabled={loading}
                    startIcon={loading ? <CircularProgress size={16} /> : <RefreshIcon />}
                    sx={{ width: { xs: '100%', sm: 'auto' } }}
                  >
                    刷新列表
                  </Button>
                  <Button variant="contained" onClick={openCreateDialog} startIcon={<AddIcon />} sx={{ width: { xs: '100%', sm: 'auto' } }}>
                    新增 Webhook
                  </Button>
                </Stack>
              </Stack>

              <Alert severity="info" variant="outlined">
                支持为不同通知渠道拆分不同的事件范围，例如错误类通知、订阅更新通知或安全类通知。测试发送不会受事件勾选限制。
              </Alert>

              <Grid container spacing={2}>
                <Grid size={{ xs: 12, sm: 6, md: 4 }}>
                  <Box
                    sx={{
                      p: 2,
                      borderRadius: 2,
                      border: '1px solid',
                      borderColor: 'divider',
                      backgroundColor: alpha(theme.palette.primary.main, 0.04)
                    }}
                  >
                    <Typography variant="caption" color="text.secondary">
                      Webhook 总数
                    </Typography>
                    <Typography variant="h3" sx={{ mt: 0.5 }}>
                      {items.length}
                    </Typography>
                  </Box>
                </Grid>
                <Grid size={{ xs: 12, sm: 6, md: 4 }}>
                  <Box
                    sx={{
                      p: 2,
                      borderRadius: 2,
                      border: '1px solid',
                      borderColor: 'divider',
                      backgroundColor: alpha(theme.palette.success.main, 0.05)
                    }}
                  >
                    <Typography variant="caption" color="text.secondary">
                      已启用
                    </Typography>
                    <Typography variant="h3" sx={{ mt: 0.5 }}>
                      {enabledCount}
                    </Typography>
                  </Box>
                </Grid>
                <Grid size={{ xs: 12, md: 4 }}>
                  <Box
                    sx={{
                      p: 2,
                      borderRadius: 2,
                      border: '1px solid',
                      borderColor: 'divider',
                      backgroundColor: alpha(theme.palette.warning.main, 0.05)
                    }}
                  >
                    <Typography variant="caption" color="text.secondary">
                      当前可选事件
                    </Typography>
                    <Typography variant="h3" sx={{ mt: 0.5 }}>
                      {eventOptions.length}
                    </Typography>
                  </Box>
                </Grid>
              </Grid>
            </Stack>
          </CardContent>
        </Card>

        {loading ? (
          <Box sx={{ py: 6, display: 'flex', justifyContent: 'center' }}>
            <CircularProgress />
          </Box>
        ) : items.length === 0 ? (
          <Alert
            severity="info"
            variant="outlined"
            action={
              <Button color="inherit" size="small" onClick={openCreateDialog} startIcon={<AddIcon />}>
                立即创建
              </Button>
            }
          >
            还没有配置任何 Webhook。建议按通知用途拆分多个接收地址，后续维护会更轻松。
          </Alert>
        ) : (
          <Grid container spacing={2.5}>
            {items.map((item, index) => {
              const title = item.name || `Webhook ${index + 1}`;
              const hasEvents = item.eventKeys.length > 0;

              return (
                <Grid size={{ xs: 12, lg: 6 }} key={item.id || `${item.url}-${index}`}>
                  <Card
                    variant="outlined"
                    sx={{
                      height: '100%',
                      borderRadius: 3,
                      transition: 'all 0.2s ease',
                      '&:hover': {
                        borderColor: theme.palette.primary.main,
                        boxShadow: theme.shadows[3]
                      }
                    }}
                  >
                    <CardHeader
                      avatar={
                        <Box
                          sx={{
                            width: 40,
                            height: 40,
                            borderRadius: 2,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            backgroundColor: alpha(theme.palette.primary.main, 0.1),
                            color: 'primary.main'
                          }}
                        >
                          <WebhookIcon fontSize="small" />
                        </Box>
                      }
                      title={
                        <Typography variant="h5" sx={{ wordBreak: 'break-word' }}>
                          {title}
                        </Typography>
                      }
                      subheader={
                        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5, wordBreak: 'break-all' }}>
                          {item.url || '未配置 URL'}
                        </Typography>
                      }
                      sx={{ pb: 1.5 }}
                    />

                    <CardContent sx={{ pt: 0 }}>
                      <Stack spacing={2}>
                        <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                          <Chip size="small" color={item.enabled ? 'success' : 'default'} label={item.enabled ? '已启用' : '已停用'} />
                          <Chip size="small" variant="outlined" label={item.method} />
                          <Chip size="small" variant="outlined" label={item.contentType} />
                          <Chip size="small" variant="outlined" color={hasEvents ? 'primary' : 'default'} label={`${item.eventKeys.length} 个事件`} />
                        </Stack>

                        {!hasEvents && (
                          <Alert severity="warning" variant="outlined">
                            当前未勾选任何自动触发事件，保存后此 Webhook 不会自动接收业务通知。
                          </Alert>
                        )}

                        <Box>
                          <Typography variant="body2" color="text.secondary">
                            请求头：{item.headers?.trim() ? '已配置' : '未配置'}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            请求体：{item.body?.trim() ? '已配置模板' : '使用默认空内容'}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            最近更新：{formatDateTime(item.updatedAt || item.createdAt)}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            最近测试：{formatDateTime(item.lastTestAt)}
                          </Typography>
                        </Box>

                        <FormControlLabel
                          sx={{ ml: 0 }}
                          control={
                            <Switch
                              checked={item.enabled}
                              onChange={(event) => handleToggleEnabled(item, event.target.checked)}
                              disabled={submitting || testingId === item.id}
                            />
                          }
                          label={item.enabled ? '通知已启用' : '通知已停用'}
                        />

                        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.25}>
                          <Button variant="outlined" startIcon={<EditIcon />} onClick={() => openEditDialog(item)} sx={{ width: { xs: '100%', sm: 'auto' } }}>
                            编辑
                          </Button>
                          <Button
                            variant="outlined"
                            color="success"
                            startIcon={testingId === item.id ? <CircularProgress size={16} /> : <SendIcon />}
                            onClick={() => handleTest(item)}
                            disabled={submitting || testingId === item.id || !item.id}
                            sx={{ width: { xs: '100%', sm: 'auto' } }}
                          >
                            发送测试
                          </Button>
                          <Button
                            variant="outlined"
                            color="error"
                            startIcon={<DeleteIcon />}
                            onClick={() => setDeleteTarget(item)}
                            disabled={submitting || testingId === item.id}
                            sx={{ width: { xs: '100%', sm: 'auto' } }}
                          >
                            删除
                          </Button>
                        </Stack>
                      </Stack>
                    </CardContent>
                  </Card>
                </Grid>
              );
            })}
          </Grid>
        )}
      </Stack>

      <Dialog open={dialogOpen} onClose={closeDialog} maxWidth="md" fullWidth fullScreen={isMobile}>
        <DialogTitle>{form.id ? '编辑 Webhook' : '新增 Webhook'}</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2.5} sx={{ pt: 1 }}>
            <Alert severity="info" variant="outlined">
              使用独立名称区分不同通知渠道会更清晰，例如“错误告警”、“安全通知”或“订阅更新”。保存成功后可回到列表里单独发送测试消息。
            </Alert>

            <TextField
              fullWidth
              label="Webhook 名称"
              value={form.name}
              onChange={(event) => setForm((prev) => ({ ...prev, name: event.target.value }))}
              placeholder="例如：错误告警"
              helperText="可选，未填写时列表会显示默认名称。"
            />

            <FormControlLabel
              control={
                <Switch
                  checked={form.enabled}
                  onChange={(event) => setForm((prev) => ({ ...prev, enabled: event.target.checked }))}
                />
              }
              label={form.enabled ? '保存后立即启用' : '先保存为停用状态'}
            />

            <TextField
              fullWidth
              required
              label="Webhook URL"
              value={form.url}
              onChange={(event) => setForm((prev) => ({ ...prev, url: event.target.value }))}
              placeholder="https://example.com/webhook"
              helperText="支持在 URL 中使用模板变量，例如 {{title}}、{{message}}、{{event}}。"
            />

            <Grid container spacing={2}>
              <Grid size={{ xs: 12, sm: 4 }}>
                <FormControl fullWidth>
                  <InputLabel>请求方法</InputLabel>
                  <Select value={form.method} label="请求方法" onChange={(event) => setForm((prev) => ({ ...prev, method: event.target.value }))}>
                    <MenuItem value="POST">POST</MenuItem>
                    <MenuItem value="GET">GET</MenuItem>
                    <MenuItem value="PUT">PUT</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
              <Grid size={{ xs: 12, sm: 8 }}>
                <FormControl fullWidth>
                  <InputLabel>Content-Type</InputLabel>
                  <Select
                    value={form.contentType}
                    label="Content-Type"
                    onChange={(event) => setForm((prev) => ({ ...prev, contentType: event.target.value }))}
                  >
                    <MenuItem value="application/json">application/json</MenuItem>
                    <MenuItem value="application/x-www-form-urlencoded">application/x-www-form-urlencoded</MenuItem>
                    <MenuItem value="text/plain">text/plain</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
            </Grid>

            <TextField
              fullWidth
              multiline
              minRows={4}
              label="请求头（JSON，可选）"
              value={form.headers}
              onChange={(event) => setForm((prev) => ({ ...prev, headers: event.target.value }))}
              placeholder={'{\n  "Authorization": "Bearer token"\n}'}
              helperText="填写 JSON 请求头，适合放鉴权信息或渠道标识。"
            />

            <TextField
              fullWidth
              multiline
              minRows={6}
              label="请求体模板"
              value={form.body}
              onChange={(event) => setForm((prev) => ({ ...prev, body: event.target.value }))}
              placeholder={
                form.contentType === 'application/json'
                  ? '{\n  "title": "{{title}}",\n  "message": "{{message}}"\n}'
                  : '{{title}} - {{message}}'
              }
              helperText="当 Content-Type 为 JSON 时，这里建议填写合法 JSON。"
            />

            <Alert severity="info" variant="outlined">
              支持变量：{'{{title}}'}、{'{{message}}'}、{'{{event}}'}、{'{{eventName}}'}、{'{{category}}'}、{'{{severity}}'}、{'{{time}}'}、{'{{json .}}'}
            </Alert>

            <NotificationEventSelector
              value={form.eventKeys}
              eventOptions={eventOptions}
              disabled={submitting}
              description="为当前 Webhook 勾选自动触发的业务事件。测试发送不会受这里的选择影响。"
              onChange={(eventKeys) => setForm((prev) => ({ ...prev, eventKeys }))}
            />
          </Stack>
        </DialogContent>
        <DialogActions sx={{ px: 3, py: 2 }}>
          <Button onClick={closeDialog} disabled={submitting} color="inherit">
            取消
          </Button>
          <Button variant="contained" onClick={handleSave} disabled={submitting} startIcon={submitting ? <CircularProgress size={16} /> : <SaveIcon />}>
            {form.id ? '保存修改' : '创建 Webhook'}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={Boolean(deleteTarget)} onClose={() => { if (!submitting) setDeleteTarget(null); }} maxWidth="xs" fullWidth>
        <DialogTitle>删除 Webhook</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="text.secondary">
            确定删除 {deleteTarget?.name || deleteTarget?.url || '这个 Webhook'} 吗？删除后无法恢复。
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteTarget(null)} disabled={submitting} color="inherit">
            取消
          </Button>
          <Button variant="contained" color="error" onClick={handleDelete} disabled={submitting} startIcon={<DeleteIcon />}>
            确认删除
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar((prev) => ({ ...prev, open: false }))}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity} variant="filled" onClose={() => setSnackbar((prev) => ({ ...prev, open: false }))}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </MainCard>
  );
}
