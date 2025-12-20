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
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import QrCodeIcon from '@mui/icons-material/QrCode';
import LinkIcon from '@mui/icons-material/Link';
import RefreshIcon from '@mui/icons-material/Refresh';
import HistoryIcon from '@mui/icons-material/History';

import { getShares, createShare, updateShare, deleteShare, getShareLogs, refreshShareToken } from '../../../api/shares';
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

  const [shares, setShares] = useState([]);
  const [loading, setLoading] = useState(false);

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

  // 获取服务器URL
  const getServerUrl = () => {
    return `${window.location.protocol}//${window.location.hostname}${window.location.port ? ':' + window.location.port : ''}`;
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

  // 渲染分享卡片
  const renderShareCard = (share) => {
    const expired = isExpired(share);

    return (
      <Card
        key={share.id}
        variant="outlined"
        sx={{
          mb: 1,
          opacity: expired ? 0.6 : 1,
          borderColor: share.is_legacy ? 'primary.main' : expired ? 'error.main' : 'divider'
        }}
      >
        <CardContent sx={{ py: 1.5 }}>
          <Stack direction="row" alignItems="center" spacing={1}>
            <Box
              onClick={() => handleOpenDetail(share)}
              sx={{
                display: 'flex',
                alignItems: 'center',
                flex: 1,
                minWidth: 0,
                cursor: 'pointer',
                gap: 1,
                '&:hover': { opacity: 0.8 }
              }}
            >
              <LinkIcon color={expired ? 'disabled' : 'primary'} fontSize="small" />
              <Box sx={{ flex: 1, minWidth: 0 }}>
                <Stack direction="row" alignItems="center" spacing={1}>
                  <Typography variant="body2" fontWeight="medium" noWrap>
                    {share.name || '未命名分享'}
                  </Typography>
                  {share.is_legacy && (
                    <Chip label="默认" size="small" sx={{ height: 18, fontSize: '0.65rem', bgcolor: '#1976d2', color: '#fff' }} />
                  )}
                </Stack>
                <Typography variant="caption" color="text.secondary">
                  {getExpireText(share)} · 访问 {share.access_count || 0} 次
                </Typography>
              </Box>
            </Box>
            <Stack direction="row" spacing={0.5}>
              <Tooltip title="访问日志">
                <IconButton size="small" onClick={(e) => handleViewLogs(share, e)}>
                  <HistoryIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              <Tooltip title="编辑">
                <IconButton size="small" onClick={(e) => handleEdit(share, e)}>
                  <EditIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              {share.is_legacy ? (
                <Tooltip title="刷新Token">
                  <IconButton size="small" color="warning" onClick={(e) => handleRefreshToken(share, e)}>
                    <RefreshIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              ) : (
                <Tooltip title="删除">
                  <IconButton size="small" color="error" onClick={(e) => handleDelete(share, e)}>
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

  // 渲染链接详情对话框内容
  const renderDetailContent = () => {
    if (!detailShare) return null;
    const serverUrl = getServerUrl();
    const baseUrl = `${serverUrl}/c/?token=${detailShare.token}`;
    const clients = [
      { name: '自动识别', url: baseUrl },
      { name: 'Clash', url: `${baseUrl}&client=clash` },
      { name: 'Surge', url: `${baseUrl}&client=surge` },
      { name: 'V2ray', url: `${baseUrl}&client=v2ray` }
    ];

    return (
      <Stack spacing={1.5}>
        {clients.map((client) => (
          <Card key={client.name} variant="outlined">
            <CardContent sx={{ py: 1, '&:last-child': { pb: 1 } }}>
              <Stack direction="row" alignItems="center" spacing={1}>
                <Chip label={client.name} size="small" color="primary" sx={{ minWidth: 80 }} />
                <Box sx={{ flex: 1, overflow: 'hidden' }}>
                  <Typography variant="body2" noWrap sx={{ fontSize: '0.75rem', color: 'text.secondary' }}>
                    {client.url}
                  </Typography>
                </Box>
                <Tooltip title="复制链接">
                  <IconButton size="small" onClick={() => copyToClipboard(client.url)}>
                    <ContentCopyIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
                <Tooltip title="显示二维码">
                  <IconButton size="small" onClick={() => handleQrCode(client.url, client.name)}>
                    <QrCodeIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              </Stack>
            </CardContent>
          </Card>
        ))}
      </Stack>
    );
  };

  return (
    <>
      {/* 主对话框 - 分享列表 */}
      <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth fullScreen={isMobile}>
        <DialogTitle>
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Typography variant="h6">分享管理 - {subscription?.Name}</Typography>
            <Stack direction="row" spacing={1}>
              <IconButton size="small" onClick={fetchShares} disabled={loading}>
                <RefreshIcon fontSize="small" />
              </IconButton>
              <Button variant="contained" size="small" startIcon={<AddIcon />} onClick={handleAdd}>
                新增
              </Button>
            </Stack>
          </Stack>
        </DialogTitle>

        <DialogContent dividers>
          {loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
              <CircularProgress />
            </Box>
          ) : shares.length === 0 ? (
            <Alert variant={'standard'} severity="info">
              暂无分享链接，点击"新增"创建第一个分享
            </Alert>
          ) : (
            shares.map((share) => renderShareCard(share))
          )}
        </DialogContent>

        <DialogActions>
          <Button onClick={onClose}>关闭</Button>
        </DialogActions>
      </Dialog>

      {/* 链接详情对话框 */}
      <Dialog open={detailOpen} onClose={() => setDetailOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          <Stack direction="row" alignItems="center" spacing={1}>
            <LinkIcon color="primary" />
            <Typography variant="h6">{detailShare?.name || '分享链接'}</Typography>
            {detailShare?.is_legacy && <Chip label="默认" size="small" sx={{ bgcolor: '#1976d2', color: '#fff' }} />}
          </Stack>
        </DialogTitle>
        <DialogContent dividers>{renderDetailContent()}</DialogContent>
        <DialogActions>
          <Button onClick={() => setDetailOpen(false)}>关闭</Button>
        </DialogActions>
      </Dialog>

      {/* 新增/编辑表单对话框 */}
      <Dialog open={formOpen} onClose={() => setFormOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>{editingShare ? '编辑分享' : '新增分享'}</DialogTitle>
        <DialogContent>
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
        <DialogActions>
          <Button onClick={() => setFormOpen(false)}>取消</Button>
          <Button variant="contained" onClick={handleSave}>
            保存
          </Button>
        </DialogActions>
      </Dialog>

      {/* IP访问日志对话框 */}
      <Dialog open={logsOpen} onClose={() => setLogsOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          <Stack direction="row" alignItems="center" spacing={1}>
            <HistoryIcon color="primary" />
            <Typography variant="h6">访问日志 - {logsShareName}</Typography>
          </Stack>
        </DialogTitle>
        <DialogContent dividers sx={{ p: 0 }}>
          {logsLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
              <CircularProgress />
            </Box>
          ) : logs.length === 0 ? (
            <Box sx={{ p: 2 }}>
              <Alert variant={'standard'} severity="info">
                暂无访问记录
              </Alert>
            </Box>
          ) : (
            <Box sx={{ maxHeight: 400, overflow: 'auto' }}>
              {logs.map((log, idx) => (
                <Box
                  key={idx}
                  sx={{
                    p: 2,
                    borderBottom: idx < logs.length - 1 ? '1px solid' : 'none',
                    borderColor: 'divider',
                    '&:hover': { bgcolor: 'action.hover' }
                  }}
                >
                  <Stack direction="row" alignItems="flex-start" spacing={2}>
                    {/* IP地址 - 使用代码风格显示 */}
                    <Box sx={{ flex: 1, minWidth: 0 }}>
                      <Typography
                        variant="body2"
                        sx={{
                          fontFamily: 'monospace',
                          bgcolor: 'action.selected',
                          px: 1,
                          py: 0.5,
                          borderRadius: 1,
                          display: 'inline-block',
                          wordBreak: 'break-all'
                        }}
                      >
                        {log.IP}
                      </Typography>
                      <Stack direction="row" spacing={2} sx={{ mt: 0.5 }}>
                        <Typography variant="caption" color="text.secondary">
                          📍 {log.Addr || '未知'}
                        </Typography>
                        <Typography variant="caption" color="text.secondary">
                          🕐 {log.Date}
                        </Typography>
                      </Stack>
                    </Box>
                    {/* 访问次数 */}
                    <Chip label={`${log.Count} 次`} size="small" color="primary" variant="outlined" sx={{ minWidth: 60 }} />
                  </Stack>
                </Box>
              ))}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Typography variant="caption" color="text.secondary" sx={{ flex: 1, pl: 2 }}>
            共 {logs.length} 条记录
          </Typography>
          <Button onClick={() => setLogsOpen(false)}>关闭</Button>
        </DialogActions>
      </Dialog>

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
