import { useEffect, useRef, useState, useCallback } from 'react';

// material-ui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Avatar from '@mui/material/Avatar';
import Badge from '@mui/material/Badge';
import Button from '@mui/material/Button';
import CardActions from '@mui/material/CardActions';
import Chip from '@mui/material/Chip';
import ClickAwayListener from '@mui/material/ClickAwayListener';
import Divider from '@mui/material/Divider';
import Paper from '@mui/material/Paper';
import Popper from '@mui/material/Popper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemAvatar from '@mui/material/ListItemAvatar';
import ListItemText from '@mui/material/ListItemText';
import Tooltip from '@mui/material/Tooltip';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import Transitions from 'ui-component/extended/Transitions';
import { useAuth } from 'contexts/AuthContext';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getHeaderPopoverTokens, getHeaderTriggerTokens } from '../headerPopoverTokens';
import { withAlpha } from 'utils/colorUtils';

// assets
import { IconBell, IconCheck, IconTrash, IconCircleCheck, IconCircleX, IconInfoCircle } from '@tabler/icons-react';

// 获取通知图标
const getNotificationIcon = (type) => {
  switch (type) {
    case 'success':
      return <IconCircleCheck size={20} color="#4caf50" />;
    case 'error':
      return <IconCircleX size={20} color="#f44336" />;
    case 'warning':
      return <IconInfoCircle size={20} color="#ff9800" />;
    default:
      return <IconInfoCircle size={20} color="#2196f3" />;
  }
};

// ==============================|| NOTIFICATION ||============================== //

export default function NotificationSection() {
  const theme = useTheme();
  const downMD = useMediaQuery(theme.breakpoints.down('md'));
  const { notifications, clearAllNotifications } = useAuth();
  const { isDark } = useResolvedColorScheme();
  const bellAccent = isDark ? theme.palette.warning.main : theme.palette.warning.dark;
  const {
    popoverSurface,
    popoverSurfaceAccent,
    popoverBorder,
    popoverInsetShadow,
    headerSurface,
    headerDivider,
    mutedText,
    listItemHover,
    selectedSurface,
    selectedHoverSurface
  } = getHeaderPopoverTokens(theme, isDark);
  const { triggerColor, triggerSurface, triggerBorder, activeColor, activeSurface, activeBorder } = getHeaderTriggerTokens(
    theme,
    isDark,
    bellAccent,
    {
      lightSurfaceAlpha: 0.14,
      lightHoverAlpha: 0.22
    }
  );

  const [open, setOpen] = useState(false);
  // 从 localStorage 初始化已读状态
  const [readIds, setReadIds] = useState(() => {
    try {
      const saved = localStorage.getItem('notification_read_ids');
      if (saved) {
        const parsed = JSON.parse(saved);
        return new Set(parsed);
      }
    } catch (e) {
      console.error('Failed to parse saved read IDs:', e);
    }
    return new Set();
  });
  const [expandedIds, setExpandedIds] = useState(new Set()); // 追踪展开的通知
  const anchorRef = useRef(null);

  // 保存已读状态到 localStorage
  const saveReadIds = useCallback((ids) => {
    try {
      localStorage.setItem('notification_read_ids', JSON.stringify([...ids]));
    } catch (e) {
      console.error('Failed to save read IDs:', e);
    }
  }, []);

  // 计算未读数
  const unreadCount = notifications.filter((n) => !readIds.has(n.id)).length;

  const handleToggle = () => {
    setOpen((prevOpen) => !prevOpen);
  };

  const handleClose = (event) => {
    if (anchorRef.current && anchorRef.current.contains(event.target)) {
      return;
    }
    setOpen(false);
  };

  const handleMarkAsRead = (id) => {
    setReadIds((prev) => {
      const newSet = new Set([...prev, id]);
      saveReadIds(newSet);
      return newSet;
    });
  };

  const handleMarkAllRead = () => {
    const allIds = notifications.map((n) => n.id);
    const newSet = new Set(allIds);
    setReadIds(newSet);
    saveReadIds(newSet);
  };

  const handleClearAll = () => {
    clearAllNotifications();
    setReadIds(new Set());
    saveReadIds(new Set()); // 清空已读状态
    setExpandedIds(new Set());
    setOpen(false);
  };

  // 展开/收起通知
  const handleToggleExpand = (id, e) => {
    e.stopPropagation();
    setExpandedIds((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(id)) {
        newSet.delete(id);
      } else {
        newSet.add(id);
      }
      return newSet;
    });
  };

  const prevOpen = useRef(open);
  useEffect(() => {
    if (prevOpen.current === true && open === false) {
      anchorRef.current.focus();
    }
    prevOpen.current = open;
  }, [open]);

  return (
    <>
      <Box sx={{ ml: 2 }}>
        <Tooltip title="通知">
          <Badge badgeContent={unreadCount} color="error" max={99}>
            <Avatar
              variant="rounded"
              sx={{
                ...theme.typography.commonAvatar,
                ...theme.typography.mediumAvatar,
                transition: 'all .2s ease-in-out',
                color: triggerColor,
                background: triggerSurface,
                border: '1px solid',
                borderColor: triggerBorder,
                position: 'relative',
                '&:hover, &[aria-controls="menu-list-grow"]': {
                  color: activeColor,
                  background: activeSurface,
                  borderColor: activeBorder
                }
              }}
              ref={anchorRef}
              aria-controls={open ? 'menu-list-grow' : undefined}
              aria-haspopup="true"
              onClick={handleToggle}
            >
              <IconBell stroke={1.5} size="20px" />
            </Avatar>
          </Badge>
        </Tooltip>
      </Box>
      <Popper
        placement={downMD ? 'bottom' : 'bottom-end'}
        open={open}
        anchorEl={anchorRef.current}
        role={undefined}
        transition
        disablePortal
        modifiers={[{ name: 'offset', options: { offset: [downMD ? 5 : 0, 20] } }]}
      >
        {({ TransitionProps }) => (
          <ClickAwayListener onClickAway={handleClose}>
            <Transitions position={downMD ? 'top' : 'top-right'} in={open} {...TransitionProps}>
              <Paper sx={{ bgcolor: 'transparent' }}>
                {open && (
                  <MainCard
                    border={false}
                    elevation={0}
                    content={false}
                    boxShadow
                    shadow={isDark ? 'none' : theme.shadows[12]}
                    sx={{
                      minWidth: 330,
                      maxWidth: 400,
                      bgcolor: popoverSurface,
                      backgroundImage: popoverSurfaceAccent,
                      border: '1px solid',
                      borderColor: popoverBorder,
                      boxShadow: popoverInsetShadow
                    }}
                  >
                    <Stack sx={{ gap: 2 }}>
                      <Stack
                        direction="row"
                        sx={{ alignItems: 'center', justifyContent: 'space-between', pt: 2, px: 2, bgcolor: headerSurface }}
                      >
                        <Stack direction="row" sx={{ gap: 1, alignItems: 'center' }}>
                          <Typography variant="subtitle1">通知消息</Typography>
                          {unreadCount > 0 && (
                            <Chip
                              size="small"
                              label={unreadCount}
                              variant="filled"
                              sx={{
                                color: isDark ? 'warning.main' : 'background.default',
                                bgcolor: isDark ? withAlpha(theme.palette.warning.main, 0.14) : 'warning.dark',
                                border: '1px solid',
                                borderColor: withAlpha(theme.palette.warning.main, isDark ? 0.28 : 0),
                                height: 20,
                                fontSize: '0.75rem'
                              }}
                            />
                          )}
                        </Stack>
                        {unreadCount > 0 && (
                          <Button
                            size="small"
                            startIcon={<IconCheck size={16} />}
                            onClick={handleMarkAllRead}
                            sx={{ textTransform: 'none' }}
                          >
                            全部已读
                          </Button>
                        )}
                      </Stack>
                      <Divider sx={{ borderColor: headerDivider }} />
                      <Box
                        sx={{
                          height: 'auto',
                          maxHeight: 'calc(100vh - 250px)',
                          overflowY: 'auto',
                          overflowX: 'hidden',
                          '&::-webkit-scrollbar': { width: 5 }
                        }}
                      >
                        {notifications.length > 0 ? (
                          <List sx={{ py: 0 }}>
                            {notifications.map((notification) => {
                              const isRead = readIds.has(notification.id);
                              return (
                                <ListItem
                                  key={notification.id}
                                  alignItems="flex-start"
                                  sx={{
                                    py: 1.5,
                                    px: 2,
                                    bgcolor: isRead ? 'transparent' : selectedSurface,
                                    borderBottom: `1px solid ${popoverBorder}`,
                                    cursor: 'pointer',
                                    transition: 'background-color 0.2s',
                                    '&:hover': {
                                      bgcolor: isRead ? listItemHover : selectedHoverSurface
                                    }
                                  }}
                                  onClick={() => handleMarkAsRead(notification.id)}
                                >
                                  <ListItemAvatar sx={{ mt: 0.5, minWidth: 40 }}>{getNotificationIcon(notification.type)}</ListItemAvatar>
                                  <ListItemText
                                    primaryTypographyProps={{ component: 'div' }}
                                    primary={
                                      <Stack direction="row" justifyContent="space-between" alignItems="center">
                                        <Typography
                                          variant="subtitle2"
                                          sx={{
                                            fontWeight: isRead ? 400 : 600,
                                            color:
                                              notification.type === 'error'
                                                ? 'error.main'
                                                : notification.type === 'success'
                                                  ? 'success.main'
                                                  : 'text.primary'
                                          }}
                                        >
                                          {notification.title}
                                        </Typography>
                                        {!isRead && (
                                          <Box
                                            sx={{
                                              width: 8,
                                              height: 8,
                                              borderRadius: '50%',
                                              bgcolor: 'primary.main',
                                              flexShrink: 0,
                                              ml: 1
                                            }}
                                          />
                                        )}
                                      </Stack>
                                    }
                                    secondaryTypographyProps={{ component: 'div' }}
                                    secondary={
                                      <>
                                        <Typography
                                          variant="body2"
                                          color={mutedText}
                                          sx={{
                                            ...(expandedIds.has(notification.id)
                                              ? { whiteSpace: 'pre-wrap', wordBreak: 'break-word' }
                                              : {
                                                  display: '-webkit-box',
                                                  WebkitLineClamp: 2,
                                                  WebkitBoxOrient: 'vertical',
                                                  overflow: 'hidden'
                                                }),
                                            mb: 0.5
                                          }}
                                        >
                                          {notification.message}
                                        </Typography>
                                        {(notification.eventName || notification.categoryName) && (
                                          <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap" sx={{ mb: 0.75 }}>
                                            {notification.categoryName && (
                                              <Chip size="small" variant="outlined" label={notification.categoryName} />
                                            )}
                                            {notification.eventName && (
                                              <Chip size="small" color="primary" variant="outlined" label={notification.eventName} />
                                            )}
                                          </Stack>
                                        )}
                                        <Stack direction="row" justifyContent="space-between" alignItems="center">
                                          <Typography variant="caption" sx={{ color: mutedText }}>
                                            {notification.timestamp?.toLocaleString('zh-CN') || '刚刚'}
                                          </Typography>
                                          {notification.message && notification.message.length > 50 && (
                                            <Typography
                                              variant="caption"
                                              color="primary"
                                              sx={{ cursor: 'pointer', '&:hover': { textDecoration: 'underline' } }}
                                              onClick={(e) => handleToggleExpand(notification.id, e)}
                                            >
                                              {expandedIds.has(notification.id) ? '收起' : '展开'}
                                            </Typography>
                                          )}
                                        </Stack>
                                      </>
                                    }
                                  />
                                </ListItem>
                              );
                            })}
                          </List>
                        ) : (
                          <Box sx={{ py: 4, textAlign: 'center' }}>
                            <Typography variant="body2" sx={{ color: mutedText }}>
                              暂无通知消息
                            </Typography>
                          </Box>
                        )}
                      </Box>
                    </Stack>
                    {notifications.length > 0 && (
                      <CardActions sx={{ p: 1.25, justifyContent: 'center', borderTop: `1px solid ${popoverBorder}` }}>
                        <Button
                          size="small"
                          color="error"
                          startIcon={<IconTrash size={16} />}
                          onClick={handleClearAll}
                          sx={{ textTransform: 'none' }}
                        >
                          清空所有通知
                        </Button>
                      </CardActions>
                    )}
                  </MainCard>
                )}
              </Paper>
            </Transitions>
          </ClickAwayListener>
        )}
      </Popper>
    </>
  );
}
