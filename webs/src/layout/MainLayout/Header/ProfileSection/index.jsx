import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

// material-ui
import { alpha, useTheme } from '@mui/material/styles';
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
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { useAuth } from 'contexts/AuthContext';
import { getHeaderPopoverTokens, getHeaderTriggerTokens } from '../headerPopoverTokens';
import { withAlpha } from 'utils/colorUtils';

// assets
import { IconLogout, IconKey, IconDatabaseExport, IconSettings, IconDatabaseOff, IconWorld } from '@tabler/icons-react';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import Backdrop from '@mui/material/Backdrop';
import CircularProgress from '@mui/material/CircularProgress';
import request from 'api/request';
import GeoIPSettingsDialog from 'views/settings/components/GeoIPSettingsDialog';

const getProfileMenuTokens = (theme, isDark) => {
  const {
    palette,
    primaryText,
    emphasizedText,
    secondaryText,
    mutedText,
    tertiaryText,
    popoverSurface,
    popoverSurfaceAccent,
    popoverBorder,
    headerSurface,
    headerDivider,
    nestedSurface,
    nestedBorder,
    listItemHover
  } = getHeaderPopoverTokens(theme, isDark);

  return {
    palette,
    primaryText,
    emphasizedText,
    secondaryText,
    mutedText,
    tertiaryText,
    popoverSurface,
    popoverSurfaceAccent,
    popoverBorder,
    headerSurface,
    headerDivider,
    nestedProfileSurface: nestedSurface,
    nestedProfileBorder: nestedBorder,
    menuHoverBg: listItemHover
  };
};

// ==============================|| PROFILE MENU ||============================== //

export default function ProfileSection() {
  const theme = useTheme();
  const { t } = useTranslation();
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
  const { isDark } = useResolvedColorScheme();
  const {
    palette,
    popoverSurface,
    popoverSurfaceAccent,
    popoverBorder,
    headerSurface,
    headerDivider,
    nestedProfileSurface,
    nestedProfileBorder,
    primaryText,
    emphasizedText,
    mutedText,
    tertiaryText,
    menuHoverBg
  } = getProfileMenuTokens(theme, isDark);
  const profileAccent = isDark ? theme.palette.info.main : theme.palette.primary.main;
  const { triggerColor, triggerSurface, triggerBorder, activeColor, activeSurface, activeBorder } = getHeaderTriggerTokens(
    theme,
    isDark,
    profileAccent,
    {
      lightSurfaceAlpha: 0.12,
      lightHoverAlpha: 0.2,
      triggerColor: isDark ? primaryText : profileAccent,
      activeColor: isDark ? emphasizedText : profileAccent
    }
  );

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
    navigate('/system/settings');
  };

  const handleApiKeys = () => {
    setOpen(false);
    navigate('/accesskey');
  };

  const handleBackup = async () => {
    setOpen(false);

    // 确认备份
    openConfirm('profile.backup.title', 'profile.backup.confirm', async () => {
      setBackupLoading(true);
      try {
        const response = await request({
          url: '/v1/backup/download',
          method: 'get',
          responseType: 'blob'
        });

        const data = response.data || response;

        if (!(data instanceof Blob) || data.size === 0) {
          setSnackbar({ open: true, message: t('profile.backup.invalidFile'), severity: 'error' });
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

        setSnackbar({ open: true, message: t('profile.backup.downloadStarted'), severity: 'success' });
      } catch (error) {
        console.error('Backup failed:', error);
        setSnackbar({ open: true, message: t('profile.backup.requestFailed'), severity: 'error' });
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
    openConfirm('profile.ipCache.clearTitle', 'profile.ipCache.clearConfirm', async () => {
      setIpCacheLoading(true);
      try {
        const { clearIPCache } = await import('api/nodes');
        await clearIPCache();
        // 成功（code === 200 时返回，否则被拦截器 reject）
        setSnackbar({ open: true, message: t('profile.ipCache.cleared'), severity: 'success' });
        setIpCacheCount(0);
        // 同时清除前端 localStorage 中的IP缓存
        try {
          localStorage.removeItem('sublink_ip_info_cache');
        } catch {
          // 忽略 localStorage 错误
        }
      } catch (error) {
        console.error('清除IP缓存失败:', error);
        setSnackbar({ open: true, message: error.message || t('profile.ipCache.clearFailed'), severity: 'error' });
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
          ml: { xs: 1, md: 2 },
          height: '48px',
          alignItems: 'center',
          borderRadius: '27px',
          color: triggerColor,
          background: triggerSurface,
          border: '1px solid',
          borderColor: triggerBorder,
          transition: 'all 0.2s',
          '&:hover, &[aria-controls="menu-list-grow"]': {
            color: activeColor,
            background: activeSurface,
            borderColor: activeBorder,
            boxShadow: theme.shadows[3]
          }
        }}
        icon={
          <Avatar
            src={user?.avatar}
            alt={!user?.avatar && (user?.username?.[0] || user?.nickname?.[0] || 'U').toUpperCase()}
            sx={{
              typography: 'mediumAvatar',
              margin: '8px 0 8px 8px !important',
              cursor: 'pointer',
              bgcolor: isDark ? withAlpha(palette.background.default, 0.72) : withAlpha(theme.palette.common.white, 0.72),
              color: isDark ? emphasizedText : profileAccent,
              border: '1px solid',
              borderColor: withAlpha(profileAccent, isDark ? 0.24 : 0.18)
            }}
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
        aria-label={t('profile.account')}
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
              <Paper sx={{ bgcolor: 'transparent' }}>
                {open && (
                  <MainCard
                    border={false}
                    elevation={0}
                    content={false}
                    boxShadow
                    shadow={isDark ? 'none' : theme.shadows[12]}
                    sx={{
                      bgcolor: popoverSurface,
                      backgroundImage: popoverSurfaceAccent,
                      border: '1px solid',
                      borderColor: popoverBorder,
                      boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.04)}` : undefined
                    }}
                  >
                    <Box
                      sx={{
                        p: 2,
                        pb: 0,
                        bgcolor: headerSurface,
                        borderBottom: isDark ? `1px solid ${headerDivider}` : 'none'
                      }}
                    >
                      <Stack>
                        <Stack direction="row" sx={{ alignItems: 'center', gap: 0.5 }}>
                          <Typography component="span" variant="h4" sx={{ fontWeight: 600, color: emphasizedText }}>
                            {user?.nickname || user?.username || t('profile.userFallback')}
                          </Typography>
                        </Stack>
                        <Typography variant="subtitle2" sx={{ color: mutedText }}>
                          {user?.role === 'admin' ? t('profile.role.admin') : t('profile.role.user')}
                        </Typography>
                      </Stack>
                      <Divider sx={{ my: 2, borderColor: withAlpha(palette.divider, isDark ? 0.7 : 1) }} />
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
                      <Card
                        sx={{
                          mb: 2,
                          bgcolor: nestedProfileSurface,
                          border: '1px solid',
                          borderColor: nestedProfileBorder,
                          boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.04)}` : 'none'
                        }}
                      >
                        <CardContent sx={{ '&:last-child': { pb: 2 } }}>
                          <Stack direction="row" sx={{ alignItems: 'center', gap: 2 }}>
                            <Avatar
                              src={user?.avatar}
                              alt={(user?.username?.[0] || 'U').toUpperCase()}
                              sx={{
                                width: 56,
                                height: 56,
                                bgcolor: alpha(theme.palette.primary.main, isDark ? 0.18 : 0.1),
                                color: 'primary.main',
                                border: '1px solid',
                                borderColor: alpha(theme.palette.primary.main, isDark ? 0.3 : 0.18)
                              }}
                            >
                              {!user?.avatar && (user?.username?.[0] || user?.nickname?.[0] || 'U').toUpperCase()}
                            </Avatar>
                            <Box sx={{ minWidth: 0 }}>
                              <Typography
                                variant="subtitle1"
                                sx={{
                                  fontWeight: 700,
                                  color: emphasizedText
                                }}
                              >
                                {user?.username || t('profile.unknownUser')}
                              </Typography>
                              <Typography variant="caption" color="text.secondary" sx={{ color: mutedText }}>
                                {user?.nickname || user?.username || t('profile.userFallback')}
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
                          '& .MuiListItemButton-root': {
                            mt: 0.5,
                            color: primaryText,
                            '& .MuiListItemIcon-root': {
                              minWidth: 36,
                              color: tertiaryText
                            },
                            '&:hover': {
                              bgcolor: menuHoverBg
                            }
                          }
                        }}
                      >
                        <ListItemButton sx={{ borderRadius: `${borderRadius}px` }} onClick={handleApiKeys}>
                          <ListItemIcon>
                            <IconKey stroke={1.5} size="20px" />
                          </ListItemIcon>
                          <ListItemText
                            primaryTypographyProps={{ component: 'div' }}
                            primary={<Typography variant="body2">{t('profile.actions.apiKeys')}</Typography>}
                          />
                        </ListItemButton>
                        <ListItemButton sx={{ borderRadius: `${borderRadius}px` }} onClick={handlePersonalCenter}>
                          <ListItemIcon>
                            <IconSettings stroke={1.5} size="20px" />
                          </ListItemIcon>
                          <ListItemText
                            primaryTypographyProps={{ component: 'div' }}
                            primary={<Typography variant="body2">{t('profile.actions.personalCenter')}</Typography>}
                          />
                        </ListItemButton>
                        <ListItemButton sx={{ borderRadius: `${borderRadius}px` }} onClick={handleBackup}>
                          <ListItemIcon>
                            <IconDatabaseExport stroke={1.5} size="20px" />
                          </ListItemIcon>
                          <ListItemText
                            primaryTypographyProps={{ component: 'div' }}
                            primary={<Typography variant="body2">{t('profile.actions.backup')}</Typography>}
                          />
                        </ListItemButton>
                        <ListItemButton sx={{ borderRadius: `${borderRadius}px` }} onClick={handleClearIPCache} disabled={ipCacheLoading}>
                          <ListItemIcon>
                            <IconDatabaseOff stroke={1.5} size="20px" />
                          </ListItemIcon>
                          <ListItemText
                            primaryTypographyProps={{ component: 'div' }}
                            primary={<Typography variant="body2">{t('profile.actions.clearIpCache')}</Typography>}
                            secondary={
                              <Typography variant="caption" sx={{ color: mutedText }}>
                                {t('profile.ipCache.currentCount', { count: ipCacheCount })}
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
                            primary={<Typography variant="body2">{t('profile.actions.geoip')}</Typography>}
                          />
                        </ListItemButton>
                        <Divider sx={{ my: 1 }} />
                        <ListItemButton
                          sx={{
                            borderRadius: `${borderRadius}px`,
                            color: 'error.main',
                            '&:hover': {
                              bgcolor: withAlpha(theme.palette.error.main, isDark ? 0.16 : 0.12)
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
                                {t('profile.actions.logout')}
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
      <Backdrop sx={{ color: theme.palette.common.white, zIndex: (muiTheme) => muiTheme.zIndex.drawer + 1 }} open={backupLoading}>
        <Stack alignItems="center" spacing={2}>
          <CircularProgress color="inherit" />
          <Typography>{t('profile.backup.generating')}</Typography>
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
        <DialogTitle id="alert-dialog-title">{t(confirmInfo.title)}</DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-description">{t(confirmInfo.content, { count: ipCacheCount })}</DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleConfirmClose}>{t('common.cancel')}</Button>
          <Button onClick={handleConfirmAction} color="primary" autoFocus>
            {t('common.confirm')}
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
