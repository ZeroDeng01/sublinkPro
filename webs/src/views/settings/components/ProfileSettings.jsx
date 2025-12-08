import { useState, useEffect } from 'react';

// material-ui
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import Avatar from '@mui/material/Avatar';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Grid from '@mui/material/Grid';
import Divider from '@mui/material/Divider';
import InputAdornment from '@mui/material/InputAdornment';
import IconButton from '@mui/material/IconButton';

// icons
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import PersonIcon from '@mui/icons-material/Person';
import LockIcon from '@mui/icons-material/Lock';
import SaveIcon from '@mui/icons-material/Save';

// project imports
import { useAuth } from 'contexts/AuthContext';
import { changePassword, updateProfile } from 'api/user';

// ==============================|| 个人设置组件 ||============================== //

export default function ProfileSettings({ showMessage, loading, setLoading }) {
  const { user, logout } = useAuth();

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

  useEffect(() => {
    if (user) {
      setProfileForm({
        username: user.username || '',
        nickname: user.nickname || ''
      });
    }
  }, [user]);

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

  return (
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
                bgcolor: 'primary',
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

      {/* 右侧：密码修改 */}
      <Grid item xs={12} md={8}>
        <Card>
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
      </Grid>
    </Grid>
  );
}
