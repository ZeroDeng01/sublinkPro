import { useState, useEffect, useCallback } from 'react';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import IconButton from '@mui/material/IconButton';
import Chip from '@mui/material/Chip';
import Typography from '@mui/material/Typography';
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import Alert from '@mui/material/Alert';
import CircularProgress from '@mui/material/CircularProgress';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Tooltip from '@mui/material/Tooltip';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme } from '@mui/material/styles';

import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import LinkIcon from '@mui/icons-material/Link';
import RefreshIcon from '@mui/icons-material/Refresh';
import HistoryIcon from '@mui/icons-material/History';

import { getShares, createShare, updateShare, deleteShare, getShareLogs, refreshShareToken } from '../../../api/shares';
import { getSystemDomain } from '../../../api/settings';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';
import AccessLogsDialog from './AccessLogsDialog';
import ClientUrlsDialog from './ClientUrlsDialog';
import QrCodeDialog from './QrCodeDialog';
import ConfirmDialog from './ConfirmDialog';

// 过期类型常量
const EXPIRE_TYPE_NEVER = 0;
const EXPIRE_TYPE_DAYS = 1;
const EXPIRE_TYPE_DATETIME = 2;

/**
 * 分享管理对话框
 */
export default function ShareManageDialog({ open, subscription, onClose, showMessage }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);

  const [shares, setShares] = useState([]);
  const [loading, setLoading] = useState(false);

  // 系统域名配置
  const [systemDomainConfig, setSystemDomainConfig] = useState('');

  // 链接详情对话框
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailShare, setDetailShare] = useState(null);

  // 新增/编辑表单
  const [formOpen, setFormOpen] = useState(false);
  const [editingShare, setEditingShare] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    token: '',
    expire_type: EXPIRE_TYPE_NEVER,
    expire_days: 30,
    expire_at: '',
    enabled: true
  });

  // 二维码对话框
  const [qrOpen, setQrOpen] = useState(false);
  const [qrUrl, setQrUrl] = useState('');
  const [qrTitle, setQrTitle] = useState('');

  // IP日志对话框
  const [logsOpen, setLogsOpen] = useState(false);
  const [logsLoading, setLogsLoading] = useState(false);
  const [logs, setLogs] = useState([]);
  const [logsShareName, setLogsShareName] = useState('');

  // 确认对话框
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [confirmInfo, setConfirmInfo] = useState({ title: '', content: '', onConfirm: null });

  // 获取服务器URL：优先使用系统域名配置，否则使用当前访问地址
  const getServerUrl = useCallback(() => {
    if (systemDomainConfig) {
      // 确保没有末尾斜杠
      return systemDomainConfig.replace(/\/+$/, '');
    }
    return `${window.location.protocol}//${window.location.hostname}${window.location.port ? ':' + window.location.port : ''}`;
  }, [systemDomainConfig]);

  // 获取系统域名配置
  const fetchSystemDomain = async () => {
    try {
      const res = await getSystemDomain();
      if (res.data?.systemDomain) {
        setSystemDomainConfig(res.data.systemDomain);
      }
    } catch (error) {
      console.error('获取系统域名配置失败:', error);
    }
  };

  // 获取分享列表
  const fetchShares = useCallback(async () => {
    if (!subscription?.ID) return;
    setLoading(true);
    try {
      const res = await getShares(subscription.ID);
      setShares(res.data || []);
    } catch (error) {
      console.error('获取分享列表失败:', error);
    } finally {
      setLoading(false);
    }
  }, [subscription?.ID]);

  useEffect(() => {
    if (open && subscription?.ID) {
      fetchSystemDomain();
      fetchShares();
    }
  }, [open, subscription?.ID, fetchShares]);

  // 复制到剪贴板
  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    showMessage?.('已复制到剪贴板', 'success');
  };

  // 打开链接详情
  const handleOpenDetail = (share) => {
    setDetailShare(share);
    setDetailOpen(true);
  };

  // 打开新增表单
  const handleAdd = () => {
    setEditingShare(null);
    setFormData({
      name: '',
      token: '',
      expire_type: EXPIRE_TYPE_NEVER,
      expire_days: 30,
      expire_at: '',
      enabled: true
    });
    setFormOpen(true);
  };

  // 打开编辑表单
  const handleEdit = (share, e) => {
    e?.stopPropagation();
    setEditingShare(share);
    setFormData({
      name: share.name || '',
      token: share.token || '',
      expire_type: share.expire_type || EXPIRE_TYPE_NEVER,
      expire_days: share.expire_days || 30,
      expire_at: share.expire_at ? share.expire_at.substring(0, 16) : '',
      enabled: share.enabled !== false
    });
    setFormOpen(true);
  };

  // 保存分享
  const handleSave = async () => {
    try {
      const data = {
        ...formData,
        subscription_id: subscription.ID
      };

      if (editingShare) {
        data.id = editingShare.id;
        await updateShare(data);
        showMessage?.('更新成功', 'success');
      } else {
        await createShare(data);
        showMessage?.('创建成功', 'success');
      }
      setFormOpen(false);
      fetchShares();
    } catch (error) {
      console.error('保存失败:', error);
      showMessage?.(error.response?.data?.msg || '保存失败', 'error');
    }
  };

  // 删除分享
  const handleDelete = (share, e) => {
    e?.stopPropagation();
    setConfirmInfo({
      title: '删除分享',
      content: `确定要删除分享"${share.name || share.token}"吗？`,
      onConfirm: async () => {
        try {
          await deleteShare(share.id);
          showMessage?.('删除成功', 'success');
          fetchShares();
          if (detailShare?.id === share.id) {
            setDetailOpen(false);
          }
        } catch (error) {
          console.error('删除失败:', error);
          showMessage?.(error.response?.data?.msg || '删除失败', 'error');
        }
        setConfirmOpen(false);
      }
    });
    setConfirmOpen(true);
  };

  // 刷新Token
  const handleRefreshToken = (share, e) => {
    e?.stopPropagation();
    setConfirmInfo({
      title: '刷新Token',
      content: '刷新Token后，旧链接将失效，确定要刷新吗？',
      onConfirm: async () => {
        try {
          await refreshShareToken(share.id);
          showMessage?.('Token已刷新', 'success');
          fetchShares();
          if (detailShare?.id === share.id) {
            setDetailOpen(false);
          }
        } catch (error) {
          console.error('刷新失败:', error);
          showMessage?.(error.response?.data?.msg || '刷新失败', 'error');
        }
        setConfirmOpen(false);
      }
    });
    setConfirmOpen(true);
  };

  // 查看IP日志
  const handleViewLogs = async (share, e) => {
    e?.stopPropagation();
    setLogsShareName(share.name || '未命名分享');
    setLogsLoading(true);
    setLogsOpen(true);
    try {
      const res = await getShareLogs(share.id);
      setLogs(res.data || []);
    } catch (error) {
      console.error('获取日志失败:', error);
      setLogs([]);
    } finally {
      setLogsLoading(false);
    }
  };

  // 显示二维码
  const handleQrCode = (url, title) => {
    setQrUrl(url);
    setQrTitle(title);
    setQrOpen(true);
  };

  // 获取过期状态文本
  const getExpireText = (share) => {
    if (!share.enabled) return '已禁用';
    switch (share.expire_type) {
      case EXPIRE_TYPE_NEVER:
        return '永不过期';
      case EXPIRE_TYPE_DAYS:
        return `${share.expire_days}天后过期`;
      case EXPIRE_TYPE_DATETIME:
        return share.expire_at ? new Date(share.expire_at).toLocaleString() : '指定时间';
      default:
        return '永不过期';
    }
  };

  // 检查是否过期
  const isExpired = (share) => {
    if (!share.enabled) return true;
    if (share.expire_type === EXPIRE_TYPE_DAYS && share.expire_days > 0) {
      const created = new Date(share.created_at);
      const expireDate = new Date(created.getTime() + share.expire_days * 24 * 60 * 60 * 1000);
      return new Date() > expireDate;
    }
    if (share.expire_type === EXPIRE_TYPE_DATETIME && share.expire_at) {
      return new Date() > new Date(share.expire_at);
    }
    return false;
  };

  const getDialogPaperSx = (fullScreen = false) => ({
    borderRadius: fullScreen ? 0 : 3,
    overflow: 'hidden',
    bgcolor: dialogSurface,
    backgroundImage: dialogSurfaceGradient,
    border: fullScreen ? 'none' : '1px solid',
    borderColor: panelBorder
  });

  const iconButtonBaseSx = {
    color: secondaryText,
    bgcolor: nestedPanelSurface,
    border: '1px solid',
    borderColor: panelBorder,
    boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none',
    transition: 'all 0.2s ease',
    '&:hover': {
      color: primaryText,
      bgcolor: withAlpha(palette.primary.main, isDark ? 0.14 : 0.06),
      borderColor: withAlpha(palette.primary.main, isDark ? 0.34 : 0.2)
    }
  };

  const actionIconButtonSx = {
    ...iconButtonBaseSx,
    width: 32,
    height: 32
  };

  const legacyChipSx = {
    height: 20,
    fontSize: '0.68rem',
    fontWeight: 700,
    bgcolor: withAlpha(palette.primary.main, isDark ? 0.18 : 0.1),
    color: palette.primary.main,
    border: '1px solid',
    borderColor: withAlpha(palette.primary.main, isDark ? 0.38 : 0.22),
    '& .MuiChip-label': {
      px: 0.9
    }
  };

  // 渲染分享卡片
  const renderShareCard = (share) => {
    const expired = isExpired(share);
    const accentColor = share.is_legacy ? palette.primary.main : expired ? palette.error.main : palette.info.main;
    const accentSurface = share.is_legacy
      ? withAlpha(palette.primary.main, isDark ? 0.16 : 0.06)
      : expired
        ? withAlpha(palette.error.main, isDark ? 0.14 : 0.05)
        : nestedPanelSurface;
    const accentBorder = share.is_legacy
      ? withAlpha(palette.primary.main, isDark ? 0.38 : 0.22)
      : expired
        ? withAlpha(palette.error.main, isDark ? 0.34 : 0.2)
        : panelBorder;

    return (
      <Card
        key={share.id}
        sx={{
          borderRadius: 2.5,
          bgcolor: accentSurface,
          backgroundImage: share.is_legacy
            ? `linear-gradient(180deg, ${withAlpha(palette.primary.main, isDark ? 0.1 : 0.04)} 0%, ${accentSurface} 100%)`
            : 'none',
          border: '1px solid',
          borderColor: accentBorder,
          boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none',
          opacity: expired ? 0.72 : 1,
          transition: 'all 0.2s ease',
          '&:hover': {
            borderColor: withAlpha(accentColor, isDark ? 0.48 : 0.28),
            bgcolor: share.is_legacy ? withAlpha(palette.primary.main, isDark ? 0.2 : 0.08) : mutedPanelSurface
          }
        }}
      >
        <CardContent sx={{ px: 2, py: 1.75, '&:last-child': { pb: 1.75 } }}>
          <Stack direction="row" alignItems="center" spacing={1.25}>
            <Box
              onClick={() => handleOpenDetail(share)}
              sx={{
                display: 'flex',
                alignItems: 'center',
                flex: 1,
                minWidth: 0,
                cursor: 'pointer',
                gap: 1,
                '&:hover': { opacity: 0.92 }
              }}
            >
              <Box
                sx={{
                  width: 34,
                  height: 34,
                  borderRadius: 1.75,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  flexShrink: 0,
                  bgcolor: withAlpha(accentColor, isDark ? 0.16 : 0.08),
                  color: expired ? tertiaryText : accentColor,
                  border: '1px solid',
                  borderColor: withAlpha(accentColor, isDark ? 0.32 : 0.18)
                }}
              >
                <LinkIcon fontSize="small" />
              </Box>
              <Box sx={{ flex: 1, minWidth: 0 }}>
                <Stack direction="row" alignItems="center" spacing={0.75} sx={{ mb: 0.35 }}>
                  <Typography variant="body2" fontWeight={600} noWrap sx={{ color: primaryText }}>
                    {share.name || '未命名分享'}
                  </Typography>
                  {share.is_legacy && <Chip label="默认" size="small" sx={legacyChipSx} />}
                </Stack>
                <Typography variant="caption" sx={{ color: expired ? tertiaryText : secondaryText }}>
                  {getExpireText(share)} · 访问 {share.access_count || 0} 次
                </Typography>
              </Box>
            </Box>
            <Stack direction="row" spacing={0.5}>
              <Tooltip title="访问日志">
                <IconButton size="small" onClick={(e) => handleViewLogs(share, e)} sx={actionIconButtonSx}>
                  <HistoryIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              <Tooltip title="编辑">
                <IconButton size="small" onClick={(e) => handleEdit(share, e)} sx={actionIconButtonSx}>
                  <EditIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              {share.is_legacy ? (
                <Tooltip title="刷新Token">
                  <IconButton
                    size="small"
                    onClick={(e) => handleRefreshToken(share, e)}
                    sx={{
                      ...actionIconButtonSx,
                      color: palette.warning.main,
                      '&:hover': {
                        ...actionIconButtonSx['&:hover'],
                        color: palette.warning.main,
                        bgcolor: withAlpha(palette.warning.main, isDark ? 0.16 : 0.08),
                        borderColor: withAlpha(palette.warning.main, isDark ? 0.36 : 0.2)
                      }
                    }}
                  >
                    <RefreshIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              ) : (
                <Tooltip title="删除">
                  <IconButton
                    size="small"
                    onClick={(e) => handleDelete(share, e)}
                    sx={{
                      ...actionIconButtonSx,
                      color: palette.error.main,
                      '&:hover': {
                        ...actionIconButtonSx['&:hover'],
                        color: palette.error.main,
                        bgcolor: withAlpha(palette.error.main, isDark ? 0.16 : 0.08),
                        borderColor: withAlpha(palette.error.main, isDark ? 0.34 : 0.2)
                      }
                    }}
                  >
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              )}
            </Stack>
          </Stack>
        </CardContent>
      </Card>
    );
  };

  const detailClientUrls = detailShare
    ? {
        自动识别: `${getServerUrl()}/c/?token=${detailShare.token}`,
        Clash: `${getServerUrl()}/c/?token=${detailShare.token}&client=clash`,
        Surge: `${getServerUrl()}/c/?token=${detailShare.token}&client=surge`,
        V2ray: `${getServerUrl()}/c/?token=${detailShare.token}&client=v2ray`
      }
    : {};

  return (
    <>
      {/* 主对话框 - 分享列表 */}
      <Dialog
        open={open}
        onClose={onClose}
        maxWidth="sm"
        fullWidth
        fullScreen={isMobile}
        slotProps={{
          paper: {
            sx: getDialogPaperSx(isMobile)
          }
        }}
      >
        <DialogTitle
          sx={{
            px: 2.5,
            py: 1.75,
            bgcolor: mutedPanelSurface,
            borderBottom: '1px solid',
            borderColor: panelBorder,
            boxShadow: `inset 0 -1px 0 ${withAlpha(palette.divider, 0.4)}`
          }}
        >
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Typography variant="h6">分享管理 - {subscription?.Name}</Typography>
            <Stack direction="row" spacing={1}>
              <IconButton size="small" onClick={fetchShares} disabled={loading} sx={iconButtonBaseSx}>
                <RefreshIcon fontSize="small" />
              </IconButton>
              <Button variant="contained" size="small" startIcon={<AddIcon />} onClick={handleAdd}>
                新增
              </Button>
            </Stack>
          </Stack>
        </DialogTitle>

        <DialogContent
          sx={{
            px: 2.5,
            pt: 2.5,
            pb: 2,
            bgcolor: dialogSurface
          }}
        >
          {loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4.25 }}>
              <CircularProgress />
            </Box>
          ) : shares.length === 0 ? (
            <Alert
              variant="outlined"
              severity="info"
              sx={{
                mt: 1.5,
                bgcolor: withAlpha(palette.info.main, isDark ? 0.12 : 0.05),
                borderColor: withAlpha(palette.info.main, isDark ? 0.3 : 0.18)
              }}
            >
              暂无分享链接，点击"新增"创建第一个分享
            </Alert>
          ) : (
            <Stack spacing={1.5} sx={{ mt: 1.5 }}>
              {shares.map((share) => renderShareCard(share))}
            </Stack>
          )}
        </DialogContent>

        <DialogActions sx={{ px: 2.5, py: 1.5, bgcolor: mutedPanelSurface, borderTop: '1px solid', borderColor: panelBorder }}>
          <Button onClick={onClose} variant="outlined">
            关闭
          </Button>
        </DialogActions>
      </Dialog>

      {/* 链接详情对话框 */}
      <ClientUrlsDialog
        open={detailOpen}
        title={detailShare?.name || '分享链接'}
        subtitle="选择需要的客户端地址，可直接复制或生成二维码"
        legacy={Boolean(detailShare?.is_legacy)}
        clientUrls={detailClientUrls}
        onClose={() => setDetailOpen(false)}
        onQrCode={handleQrCode}
        onCopy={copyToClipboard}
      />

      {/* 新增/编辑表单对话框 */}
      <Dialog
        open={formOpen}
        onClose={() => setFormOpen(false)}
        maxWidth="xs"
        fullWidth
        slotProps={{
          paper: {
            sx: getDialogPaperSx(false)
          }
        }}
      >
        <DialogTitle
          sx={{
            px: 2.5,
            py: 2,
            bgcolor: mutedPanelSurface,
            borderBottom: '1px solid',
            borderColor: panelBorder
          }}
        >
          {editingShare ? '编辑分享' : '新增分享'}
        </DialogTitle>
        <DialogContent sx={{ bgcolor: dialogSurface }}>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              label="分享名称"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="例如：朋友使用、临时分享"
              size="small"
              fullWidth
            />

            <TextField
              label="自定义Token（可选）"
              value={formData.token}
              onChange={(e) => setFormData({ ...formData, token: e.target.value })}
              placeholder="留空自动生成随机token"
              size="small"
              fullWidth
              helperText="自定义token便于记忆，留空则自动生成安全的随机token"
            />

            <FormControl size="small" fullWidth>
              <InputLabel>过期策略</InputLabel>
              <Select
                value={formData.expire_type}
                label="过期策略"
                onChange={(e) => setFormData({ ...formData, expire_type: e.target.value })}
              >
                <MenuItem value={EXPIRE_TYPE_NEVER}>永不过期</MenuItem>
                <MenuItem value={EXPIRE_TYPE_DAYS}>按天数过期</MenuItem>
                <MenuItem value={EXPIRE_TYPE_DATETIME}>指定时间过期</MenuItem>
              </Select>
            </FormControl>

            {formData.expire_type === EXPIRE_TYPE_DAYS && (
              <TextField
                label="过期天数"
                type="number"
                value={formData.expire_days}
                onChange={(e) => setFormData({ ...formData, expire_days: parseInt(e.target.value) || 0 })}
                size="small"
                fullWidth
                inputProps={{ min: 1 }}
              />
            )}

            {formData.expire_type === EXPIRE_TYPE_DATETIME && (
              <TextField
                label="过期时间"
                type="datetime-local"
                value={formData.expire_at}
                onChange={(e) => setFormData({ ...formData, expire_at: e.target.value })}
                size="small"
                fullWidth
                InputLabelProps={{ shrink: true }}
              />
            )}

            {editingShare && (
              <FormControlLabel
                control={<Switch checked={formData.enabled} onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })} />}
                label="启用此分享"
              />
            )}
          </Stack>
        </DialogContent>
        <DialogActions sx={{ px: 2.5, py: 1.75, bgcolor: mutedPanelSurface, borderTop: '1px solid', borderColor: panelBorder }}>
          <Button onClick={() => setFormOpen(false)}>取消</Button>
          <Button variant="contained" onClick={handleSave}>
            保存
          </Button>
        </DialogActions>
      </Dialog>

      {/* IP访问日志对话框 */}
      <AccessLogsDialog
        open={logsOpen}
        logs={logs}
        loading={logsLoading}
        title={`访问日志 - ${logsShareName}`}
        onClose={() => setLogsOpen(false)}
      />

      {/* 二维码对话框 */}
      <QrCodeDialog open={qrOpen} title={qrTitle} url={qrUrl} onClose={() => setQrOpen(false)} onCopy={copyToClipboard} />

      {/* 确认对话框 */}
      <ConfirmDialog
        open={confirmOpen}
        title={confirmInfo.title}
        content={confirmInfo.content}
        onClose={() => setConfirmOpen(false)}
        onConfirm={confirmInfo.onConfirm}
      />
    </>
  );
}
