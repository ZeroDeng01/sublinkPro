import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';

// material-ui
import { useTheme } from '@mui/material/styles';
import Avatar from '@mui/material/Avatar';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import ClickAwayListener from '@mui/material/ClickAwayListener';
import Divider from '@mui/material/Divider';
import List from '@mui/material/List';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Paper from '@mui/material/Paper';
import Popper from '@mui/material/Popper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import Transitions from 'ui-component/extended/Transitions';
import useConfig from 'hooks/useConfig';
import { useAuth } from 'contexts/AuthContext';

// assets
import { IconLogout, IconUser, IconKey, IconDatabaseExport, IconSettings, IconDatabaseOff, IconWorld } from '@tabler/icons-react';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import Backdrop from '@mui/material/Backdrop';
import CircularProgress from '@mui/material/CircularProgress';
import request from 'api/request';
import GeoIPSettingsDialog from 'views/settings/components/GeoIPSettingsDialog';

// ==============================|| 问候语计算 ||============================== //

const getGreeting = () => {
  const hour = new Date().getHours();
  if (hour >= 5 && hour < 9) {
    return '早上好';
  } else if (hour >= 9 && hour < 12) {
    return '上午好';
  } else if (hour >= 12 && hour < 14) {
    return '中午好';
  } else if (hour >= 14 && hour < 18) {
    return '下午好';
  } else if (hour >= 18 && hour < 23) {
    return '晚上好';
  } else {
    return '夜深了';
  }
};

// ==============================|| PROFILE MENU ||============================== //

export default function ProfileSection() {
  const theme = useTheme();
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const {
    state: { borderRadius }
  } = useConfig();

  const [open, setOpen] = useState(false);
  const [backupLoading, setBackupLoading] = useState(false);
  const [ipCacheCount, setIpCacheCount] = useState(0);
  const [ipCacheLoading, setIpCacheLoading] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });
  const [geoipDialogOpen, setGeoipDialogOpen] = useState(false);
  const anchorRef = useRef(null);
  const greeting = getGreeting();

  // 确认对话框
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [confirmInfo, setConfirmInfo] = useState({
    title: '',
    content: '',
    action: null
  });

  const openConfirm = (title, content, action) => {
    setConfirmInfo({ title, content, action });
    setConfirmOpen(true);
  };

  const handleConfirmClose = () => {
    setConfirmOpen(false);
  };

  const handleConfirmAction = async () => {
    if (confirmInfo.action) {
      await confirmInfo.action();
    }
    setConfirmOpen(false);
  };

  const handleToggle = () => {
    setOpen((prevOpen) => !prevOpen);
  };

  const handleClose = (event) => {
    if (anchorRef.current && anchorRef.current.contains(event.target)) {
      return;
    }
    setOpen(false);
  };

  const handleLogout = async () => {
    try {
      await logout();
      navigate('/pages/login3');
    } catch (error) {
      console.error('退出登录失败:', error);
    }
  };

  const handlePersonalCenter = () => {
    setOpen(false);
    navigate('/settings');
  };

  const handleApiKeys = () => {
    setOpen(false);
    navigate('/accesskey');
  };

  const handleBackup = async () => {
    setOpen(false);

    // 确认备份
    openConfirm('系统备份', '确定备份系统数据吗？数据备份文件存有您的机密信息，请妥善保管。', async () => {
      setBackupLoading(true);
      try {
        const response = await request({
          url: '/v1/backup/download',
          method: 'get',
          responseType: 'blob'
        });

        const data = response.data || response;

        if (!(data instanceof Blob) || data.size === 0) {
          setSnackbar({ open: true, message: '备份失败：服务器未返回有效的备份文件', severity: 'error' });
          return;
        }

        // 提取文件名
        let filename = 'sublink-pro-backup.zip';
        const contentDisposition = response.headers?.['content-disposition'];
        if (contentDisposition) {
          const match = contentDisposition.match(/filename="?([^"]+)"?|filename\*=UTF-8''([^"]+)/);
          if (match && (match[1] || match[2])) {
            filename = decodeURIComponent(match[2] || match[1]);
          }
        }

        // 创建下载链接
        const url = window.URL.createObjectURL(data);
        const link = document.createElement('a');
        link.href = url;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        window.URL.revokeObjectURL(url);

        setSnackbar({ open: true, message: '备份文件已开始下载', severity: 'success' });
      } catch (error) {
        console.error('Backup failed:', error);
        setSnackbar({ open: true, message: '备份请求失败，请检查网络或服务器日志', severity: 'error' });
      } finally {
        setBackupLoading(false);
      }
    });
  };

  // 获取IP缓存统计
  const fetchIPCacheStats = async () => {
    try {
      const { getIPCacheStats } = await import('api/nodes');
      const response = await getIPCacheStats();
      // 成功（code === 200 时返回，否则被拦截器 reject）
      setIpCacheCount(response.data?.count || 0);
    } catch (error) {
      console.error('获取IP缓存统计失败:', error);
    }
  };

  // 清除IP缓存
  const handleClearIPCache = async () => {
    setOpen(false);
    openConfirm('清除IP缓存', `确定要清除所有IP缓存数据吗？当前共有 ${ipCacheCount} 条缓存记录。`, async () => {
      setIpCacheLoading(true);
      try {
        const { clearIPCache } = await import('api/nodes');
        await clearIPCache();
        // 成功（code === 200 时返回，否则被拦截器 reject）
        setSnackbar({ open: true, message: 'IP缓存已清除', severity: 'success' });
        setIpCacheCount(0);
        // 同时清除前端 localStorage 中的IP缓存
        try {
          localStorage.removeItem('sublink_ip_info_cache');
        } catch {
          // 忽略 localStorage 错误
        }
      } catch (error) {
        console.error('清除IP缓存失败:', error);
        setSnackbar({ open: true, message: error.message || '清除失败', severity: 'error' });
      } finally {
        setIpCacheLoading(false);
      }
    });
  };

  const prevOpen = useRef(open);
  useEffect(() => {
    if (prevOpen.current === true && open === false) {
      anchorRef.current.focus();
    }
    prevOpen.current = open;
  }, [open]);

  // 组件挂载时获取IP缓存统计
  useEffect(() => {
    fetchIPCacheStats();
  }, []);

  return (
    <>
      <Chip
        slotProps={{ label: { sx: { lineHeight: 0 } } }}
        sx={{
          ml: 2,
          height: '48px',
          alignItems: 'center',
          borderRadius: '27px',
          transition: 'all 0.2s',
          '&:hover': {
            boxShadow: theme.shadows[3]
          }
        }}
        icon={
          <Avatar
            src={user?.avatar}
            alt={!user?.avatar && (user?.username?.[0] || user?.nickname?.[0] || 'U').toUpperCase()}
            sx={{ typography: 'mediumAvatar', margin: '8px 0 8px 8px !important', cursor: 'pointer' }}
            ref={anchorRef}
            aria-controls={open ? 'menu-list-grow' : undefined}
            aria-haspopup="true"
            color="inherit"
          >
            {!user?.avatar && (user?.username?.[0] || user?.nickname?.[0] || 'U').toUpperCase()}
          </Avatar>
        }
        label={<IconSettings stroke={1.5} size="24px" />}
        ref={anchorRef}
        aria-controls={open ? 'menu-list-grow' : undefined}
        aria-haspopup="true"
        onClick={handleToggle}
        color="primary"
        aria-label="用户账户"
      />
      <Popper
        placement="bottom"
        open={open}
        anchorEl={anchorRef.current}
        role={undefined}
        transition
        disablePortal
        modifiers={[
          {
            name: 'offset',
            options: {
              offset: [0, 14]
            }
          }
        ]}
      >
        {({ TransitionProps }) => (
          <ClickAwayListener onClickAway={handleClose}>
            <Transitions in={open} {...TransitionProps}>
              <Paper>
                {open && (
                  <MainCard border={false} elevation={16} content={false} boxShadow shadow={theme.shadows[16]}>
                    <Box sx={{ p: 2, pb: 0 }}>
                      <Stack>
                        <Stack direction="row" sx={{ alignItems: 'center', gap: 0.5 }}>
                          <Typography variant="h4">{greeting}，</Typography>
                          <Typography component="span" variant="h4" sx={{ fontWeight: 600 }}>
                            {user?.nickname || user?.username || '用户'}
                          </Typography>
                        </Stack>
                        <Typography variant="subtitle2" color="textSecondary">
                          {user?.role === 'admin' ? '管理员' : '普通用户'}
                        </Typography>
                      </Stack>
                      <Divider sx={{ my: 2 }} />
                    </Box>
                    <Box
                      sx={{
                        p: 2,
                        pt: 0,
                        height: '100%',
                        maxHeight: 'calc(100vh - 250px)',
                        overflowX: 'hidden',
                        '&::-webkit-scrollbar': { width: 5 }
                      }}
                    >
                      <Card sx={{ bgcolor: 'primary.light', mb: 2 }}>
                        <CardContent>
                          <Stack direction="row" sx={{ alignItems: 'center', gap: 2 }}>
                            <Avatar src={user?.avatar} alt={(user?.username?.[0] || 'U').toUpperCase()} sx={{ width: 56, height: 56 }}>
                              {!user?.avatar && (user?.username?.[0] || user?.nickname?.[0] || 'U').toUpperCase()}
                            </Avatar>
                            <Box>
                              <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
                                {user?.username || '未知用户'}
                              </Typography>
                              <Typography variant="caption" color="textSecondary">
                                {user?.nickname || user?.username || '用户'}
                              </Typography>
                            </Box>
                          </Stack>
                        </CardContent>
                      </Card>
                      <List
                        component="nav"
                        sx={{
                          width: '100%',
                          maxWidth: 350,
                          minWidth: 300,
                          borderRadius: `${borderRadius}px`,
                          '& .MuiListItemButton-root': { mt: 0.5 }
                        }}
                      >
                        <ListItemButton sx={{ borderRadius: `${borderRadius}px` }} onClick={handleApiKeys}>
                          <ListItemIcon>
                            <IconKey stroke={1.5} size="20px" />
                          </ListItemIcon>
                          <ListItemText
                            primaryTypographyProps={{ component: 'div' }}
                            primary={<Typography variant="body2">管理API密钥</Typography>}
                          />
                        </ListItemButton>
                        <ListItemButton sx={{ borderRadius: `${borderRadius}px` }} onClick={handlePersonalCenter}>
                          <ListItemIcon>
                            <IconUser stroke={1.5} size="20px" />
                          </ListItemIcon>
                          <ListItemText
                            primaryTypographyProps={{ component: 'div' }}
                            primary={<Typography variant="body2">个人中心</Typography>}
                          />
                        </ListItemButton>
                        <ListItemButton sx={{ borderRadius: `${borderRadius}px` }} onClick={handleBackup}>
                          <ListItemIcon>
                            <IconDatabaseExport stroke={1.5} size="20px" />
                          </ListItemIcon>
                          <ListItemText
                            primaryTypographyProps={{ component: 'div' }}
                            primary={<Typography variant="body2">系统备份</Typography>}
                          />
                        </ListItemButton>
                        <ListItemButton sx={{ borderRadius: `${borderRadius}px` }} onClick={handleClearIPCache} disabled={ipCacheLoading}>
                          <ListItemIcon>
                            <IconDatabaseOff stroke={1.5} size="20px" />
                          </ListItemIcon>
                          <ListItemText
                            primaryTypographyProps={{ component: 'div' }}
                            primary={<Typography variant="body2">清除IP缓存</Typography>}
                            secondary={
                              <Typography variant="caption" color="textSecondary">
                                当前缓存 {ipCacheCount} 条
                              </Typography>
                            }
                          />
                        </ListItemButton>
                        <ListItemButton
                          sx={{ borderRadius: `${borderRadius}px` }}
                          onClick={() => {
                            setOpen(false);
                            setGeoipDialogOpen(true);
                          }}
                        >
                          <ListItemIcon>
                            <IconWorld stroke={1.5} size="20px" />
                          </ListItemIcon>
                          <ListItemText
                            primaryTypographyProps={{ component: 'div' }}
                            primary={<Typography variant="body2">GeoIP 数据库</Typography>}
                          />
                        </ListItemButton>
                        <Divider sx={{ my: 1 }} />
                        <ListItemButton
                          sx={{
                            borderRadius: `${borderRadius}px`,
                            color: 'error.main',
                            '&:hover': {
                              bgcolor: 'error.light'
                            }
                          }}
                          onClick={handleLogout}
                        >
                          <ListItemIcon>
                            <IconLogout stroke={1.5} size="20px" color={theme.palette.error.main} />
                          </ListItemIcon>
                          <ListItemText
                            primaryTypographyProps={{ component: 'div' }}
                            primary={
                              <Typography variant="body2" color="error">
                                注销登出
                              </Typography>
                            }
                          />
                        </ListItemButton>
                      </List>
                    </Box>
                  </MainCard>
                )}
              </Paper>
            </Transitions>
          </ClickAwayListener>
        )}
      </Popper>

      {/* 备份加载遮罩 */}
      <Backdrop sx={{ color: '#fff', zIndex: (theme) => theme.zIndex.drawer + 1 }} open={backupLoading}>
        <Stack alignItems="center" spacing={2}>
          <CircularProgress color="inherit" />
          <Typography>正在生成备份文件，请稍候...</Typography>
        </Stack>
      </Backdrop>

      {/* 提示消息 */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={4000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert onClose={() => setSnackbar({ ...snackbar, open: false })} severity={snackbar.severity} sx={{ width: '100%' }}>
          {snackbar.message}
        </Alert>
      </Snackbar>

      {/* 确认对话框 */}
      <Dialog
        open={confirmOpen}
        onClose={handleConfirmClose}
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
      >
        <DialogTitle id="alert-dialog-title">{confirmInfo.title}</DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-description">{confirmInfo.content}</DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleConfirmClose}>取消</Button>
          <Button onClick={handleConfirmAction} color="primary" autoFocus>
            确定
          </Button>
        </DialogActions>
      </Dialog>

      {/* GeoIP 设置对话框 */}
      <GeoIPSettingsDialog
        open={geoipDialogOpen}
        onClose={() => setGeoipDialogOpen(false)}
        showMessage={(msg, severity = 'success') => setSnackbar({ open: true, message: msg, severity })}
      />
    </>
  );
}
