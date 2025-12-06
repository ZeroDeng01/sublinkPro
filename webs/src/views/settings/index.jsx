import { useState, useEffect } from 'react';

// material-ui
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import Avatar from '@mui/material/Avatar';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Grid from '@mui/material/Grid';
import Divider from '@mui/material/Divider';
import InputAdornment from '@mui/material/InputAdornment';
import IconButton from '@mui/material/IconButton';
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
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import PersonIcon from '@mui/icons-material/Person';
import LockIcon from '@mui/icons-material/Lock';
import SaveIcon from '@mui/icons-material/Save';
import WebhookIcon from '@mui/icons-material/Webhook';
import SendIcon from '@mui/icons-material/Send';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import { useAuth } from 'contexts/AuthContext';
import { changePassword, updateProfile } from 'api/user';
import { getWebhookConfig, updateWebhookConfig, testWebhook } from 'api/settings';

// ==============================|| 用户中心 ||============================== //

export default function UserSettings() {
  const { user, logout } = useAuth();
  const [loading, setLoading] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  // 用户资料表单
  const [profileForm, setProfileForm] = useState({
    username: '',
    nickname: ''
  });

  // 密码表单
  const [passwordForm, setPasswordForm] = useState({
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
  });
  const [showOldPassword, setShowOldPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

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
    if (user) {
      setProfileForm({
        username: user.username || '',
        nickname: user.nickname || ''
      });
    }
    fetchWebhookConfig();
  }, [user]);

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

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  // === 更新资料 ===
  const handleUpdateProfile = async () => {
    if (!profileForm.username.trim()) {
      showMessage('用户名不能为空', 'warning');
      return;
    }

    const usernameChanged = user?.username !== profileForm.username;

    setLoading(true);
    try {
      await updateProfile({
        username: profileForm.username.trim(),
        nickname: profileForm.nickname.trim()
      });
      showMessage('资料更新成功');

      if (usernameChanged) {
        showMessage('用户名已修改，需要重新登录...', 'warning');
        setTimeout(() => {
          logout();
        }, 2000);
      }
    } catch (error) {
      showMessage('更新失败: ' + (error.response?.data?.message || '未知错误'), 'error');
    } finally {
      setLoading(false);
    }
  };

  // === 修改密码 ===
  const handleChangePassword = async () => {
    if (!passwordForm.oldPassword) {
      showMessage('请输入旧密码', 'warning');
      return;
    }
    if (!passwordForm.newPassword) {
      showMessage('请输入新密码', 'warning');
      return;
    }
    if (passwordForm.newPassword.length < 6) {
      showMessage('新密码长度至少6位', 'warning');
      return;
    }
    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      showMessage('两次输入的密码不一致', 'warning');
      return;
    }

    setLoading(true);
    try {
      const res = await changePassword({
        oldPassword: passwordForm.oldPassword,
        newPassword: passwordForm.newPassword,
        confirmPassword: passwordForm.confirmPassword
      });

      if (res.code !== 200) {
        throw new Error(res.msg || '修改失败');
      }
      showMessage('密码修改成功，即将重新登录...', 'success');
      setPasswordForm({ oldPassword: '', newPassword: '', confirmPassword: '' });
      setTimeout(() => {
        logout();
      }, 2000);
    } catch (error) {
      const errorMsg = error.response?.data?.message || error.message || '';
      if (errorMsg.includes('password') || errorMsg.includes('密码')) {
        showMessage('旧密码不正确', 'error');
      } else {
        showMessage('修改失败: ' + errorMsg, 'error');
      }
    } finally {
      setLoading(false);
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
    <MainCard title="用户中心">
      <Grid container spacing={3}>
        {/* 左侧：用户资料 */}
        <Grid item xs={12} md={4}>
          <Card elevation={0} sx={{ bgcolor: 'grey.50' }}>
            <CardHeader title="个人资料" />
            <CardContent sx={{ textAlign: 'center' }}>
              <Avatar
                src={user?.avatar}
                sx={{
                  width: 120,
                  height: 120,
                  mx: 'auto',
                  mb: 2,
                  bgcolor: 'primary.main',
                  fontSize: '3rem'
                }}
              >
                {user?.username?.charAt(0)?.toUpperCase() || 'U'}
              </Avatar>
              <Typography variant="h4" gutterBottom>
                {user?.username || '用户'}
              </Typography>
              {user?.nickname && (
                <Typography variant="body2" color="textSecondary" gutterBottom>
                  {user.nickname}
                </Typography>
              )}

              <Divider sx={{ my: 2 }} />

              <Stack spacing={2}>
                <TextField
                  fullWidth
                  label="用户名"
                  value={profileForm.username}
                  onChange={(e) => setProfileForm({ ...profileForm, username: e.target.value })}
                  InputProps={{
                    startAdornment: (
                      <InputAdornment position="start">
                        <PersonIcon color="action" />
                      </InputAdornment>
                    )
                  }}
                />
                <TextField
                  fullWidth
                  label="昵称"
                  value={profileForm.nickname}
                  onChange={(e) => setProfileForm({ ...profileForm, nickname: e.target.value })}
                  placeholder="可选"
                />
                <Button variant="contained" fullWidth onClick={handleUpdateProfile} disabled={loading} startIcon={<SaveIcon />}>
                  更新资料
                </Button>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        {/* 右侧：密码修改 + Webhook */}
        <Grid item xs={12} md={8}>
          {/* 密码修改 */}
          <Card sx={{ mb: 3 }}>
            <CardHeader title="修改密码" avatar={<LockIcon color="primary" />} />
            <CardContent>
              <Stack spacing={2} sx={{ maxWidth: 500 }}>
                <TextField
                  fullWidth
                  label="旧密码"
                  type={showOldPassword ? 'text' : 'password'}
                  value={passwordForm.oldPassword}
                  onChange={(e) => setPasswordForm({ ...passwordForm, oldPassword: e.target.value })}
                  autoComplete="current-password"
                  InputProps={{
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton onClick={() => setShowOldPassword(!showOldPassword)} edge="end">
                          {showOldPassword ? <VisibilityOff /> : <Visibility />}
                        </IconButton>
                      </InputAdornment>
                    )
                  }}
                />
                <TextField
                  fullWidth
                  label="新密码"
                  type={showNewPassword ? 'text' : 'password'}
                  value={passwordForm.newPassword}
                  onChange={(e) => setPasswordForm({ ...passwordForm, newPassword: e.target.value })}
                  autoComplete="new-password"
                  helperText="密码长度至少6位"
                  InputProps={{
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton onClick={() => setShowNewPassword(!showNewPassword)} edge="end">
                          {showNewPassword ? <VisibilityOff /> : <Visibility />}
                        </IconButton>
                      </InputAdornment>
                    )
                  }}
                />
                <TextField
                  fullWidth
                  label="确认新密码"
                  type={showConfirmPassword ? 'text' : 'password'}
                  value={passwordForm.confirmPassword}
                  onChange={(e) => setPasswordForm({ ...passwordForm, confirmPassword: e.target.value })}
                  autoComplete="new-password"
                  error={passwordForm.confirmPassword && passwordForm.newPassword !== passwordForm.confirmPassword}
                  helperText={
                    passwordForm.confirmPassword && passwordForm.newPassword !== passwordForm.confirmPassword ? '两次输入的密码不一致' : ''
                  }
                  InputProps={{
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton onClick={() => setShowConfirmPassword(!showConfirmPassword)} edge="end">
                          {showConfirmPassword ? <VisibilityOff /> : <Visibility />}
                        </IconButton>
                      </InputAdornment>
                    )
                  }}
                />
                <Stack direction="row" spacing={2}>
                  <Button variant="contained" onClick={handleChangePassword} disabled={loading}>
                    修改密码
                  </Button>
                  <Button variant="outlined" onClick={() => setPasswordForm({ oldPassword: '', newPassword: '', confirmPassword: '' })}>
                    重置
                  </Button>
                </Stack>
              </Stack>
            </CardContent>
          </Card>

          {/* Webhook 设置 */}
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
                      支持变量: {'{{ title }}'} 消息标题, {'{{ message }}'} 消息内容, {'{{ event }}'} 事件类型, {'{{ time }}'} 事件时间
                      <br />
                      例如 Bark URL: https://api.day.app/key/{'{{ title }}'}/{'{{ message }}'}
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
        </Grid>
      </Grid>

      {/* 提示消息 */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity}>{snackbar.message}</Alert>
      </Snackbar>
    </MainCard>
  );
}
