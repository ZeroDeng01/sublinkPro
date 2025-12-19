import { useState, useEffect, useMemo } from 'react';
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';

// material-ui
import Box from '@mui/material/Box';

import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import Stack from '@mui/material/Stack';
import Paper from '@mui/material/Paper';
import Tooltip from '@mui/material/Tooltip';
import CircularProgress from '@mui/material/CircularProgress';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import TableRow from '@mui/material/TableRow';
import Checkbox from '@mui/material/Checkbox';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import Chip from '@mui/material/Chip';

// icons
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import RefreshIcon from '@mui/icons-material/Refresh';
import SearchIcon from '@mui/icons-material/Search';
import ClearIcon from '@mui/icons-material/Clear';
import TableViewIcon from '@mui/icons-material/TableView';
import CodeIcon from '@mui/icons-material/Code';
import SaveIcon from '@mui/icons-material/Save';
import UndoIcon from '@mui/icons-material/Undo';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import SettingsIcon from '@mui/icons-material/Settings';
import PushPinIcon from '@mui/icons-material/PushPin';
import PushPinOutlinedIcon from '@mui/icons-material/PushPinOutlined';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import { getHosts, addHost, updateHost, deleteHost, batchDeleteHosts, exportHosts, syncHosts, pinHost } from 'api/hosts';

import HostSettingsDialog from './component/HostSettingsDialog';
import ConfirmDialog from 'components/ConfirmDialog';
import IPDetailsDialog from 'components/IPDetailsDialog';

// ==============================|| HOST MANAGEMENT ||============================== //

export default function HostManagement() {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  // 模式切换：table（表单模式）或 text（文本编辑模式）
  const [viewMode, setViewMode] = useState('table');
  const [hosts, setHosts] = useState([]);
  const [refreshing, setRefreshing] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  // 表单模式相关状态
  const [search, setSearch] = useState('');
  const [selectedIds, setSelectedIds] = useState([]);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingHost, setEditingHost] = useState(null);
  const [formData, setFormData] = useState({ hostname: '', ip: '', remark: '' });

  // 文本编辑模式相关状态
  const [textContent, setTextContent] = useState('');
  const [originalText, setOriginalText] = useState('');
  const [syncing, setSyncing] = useState(false);

  // 模块设置弹窗
  const [settingsOpen, setSettingsOpen] = useState(false);

  // 确认对话框
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [confirmInfo, setConfirmInfo] = useState({
    title: '',
    content: '',
    onConfirm: null
  });

  const openConfirm = (title, content, onConfirm) => {
    setConfirmInfo({ title, content, onConfirm });
    setConfirmOpen(true);
  };

  // IP详情弹窗
  const [ipDialogOpen, setIpDialogOpen] = useState(false);
  const [selectedIP, setSelectedIP] = useState('');

  const handleIPClick = (ip) => {
    setSelectedIP(ip);
    setIpDialogOpen(true);
  };

  // 获取 Host 列表
  const fetchHosts = async (showRefreshing = false) => {
    if (showRefreshing) setRefreshing(true);
    try {
      const res = await getHosts();
      // 成功（code === 200 时返回，否则被拦截器 reject）
      setHosts(res.data || []);
    } catch (error) {
      showMessage(error.message || '获取 Host 列表失败', 'error');
    } finally {
      if (showRefreshing) setRefreshing(false);
    }
  };

  // 加载文本内容
  const loadTextContent = async () => {
    try {
      const res = await exportHosts();
      // 成功（code === 200 时返回，否则被拦截器 reject）
      const text = res.data?.text || '';
      setTextContent(text);
      setOriginalText(text);
    } catch (error) {
      showMessage(error.message || '加载文本内容失败', 'error');
    }
  };

  useEffect(() => {
    fetchHosts();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 切换到文本模式时加载内容
  useEffect(() => {
    if (viewMode === 'text') {
      loadTextContent();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [viewMode]);

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  // 刷新
  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      if (viewMode === 'table') {
        await fetchHosts();
      } else {
        await loadTextContent();
      }
      showMessage('刷新成功');
    } catch (error) {
      showMessage(error.message || '刷新失败', 'error');
    } finally {
      setRefreshing(false);
    }
  };

  // 过滤后的 hosts
  const filteredHosts = useMemo(() => {
    if (!search) return hosts;
    const s = search.toLowerCase();
    return hosts.filter(
      (h) => h.hostname.toLowerCase().includes(s) || h.ip.toLowerCase().includes(s) || (h.remark && h.remark.toLowerCase().includes(s))
    );
  }, [hosts, search]);

  // ========== 表单模式操作 ==========

  const handleAdd = () => {
    setEditingHost(null);
    setFormData({ hostname: '', ip: '', remark: '' });
    setDialogOpen(true);
  };

  const handleEdit = (host) => {
    setEditingHost(host);
    setFormData({ hostname: host.hostname, ip: host.ip, remark: host.remark || '' });
    setDialogOpen(true);
  };

  const handleDelete = async (host) => {
    openConfirm('确定删除', `确定删除 "${host.hostname}" 吗？`, async () => {
      try {
        await deleteHost(host.id);
        // 成功（code === 200 时返回，否则被拦截器 reject）
        showMessage('删除成功');
        fetchHosts();
      } catch (error) {
        showMessage(error.message || '删除失败', 'error');
      }
    });
  };

  const handleBatchDelete = async () => {
    if (selectedIds.length === 0) return;
    openConfirm('批量删除', `确定删除选中的 ${selectedIds.length} 条记录吗？`, async () => {
      try {
        await batchDeleteHosts(selectedIds);
        // 成功（code === 200 时返回，否则被拦截器 reject）
        showMessage('批量删除成功');
        setSelectedIds([]);
        fetchHosts();
      } catch (error) {
        showMessage(error.message || '批量删除失败', 'error');
      }
    });
  };

  const handleSave = async () => {
    if (!formData.hostname || !formData.ip) {
      showMessage('域名和 IP 不能为空', 'error');
      return;
    }
    try {
      if (editingHost) {
        await updateHost({ id: editingHost.id, ...formData });
      } else {
        await addHost(formData);
      }
      // 成功（code === 200 时返回，否则被拦截器 reject）
      showMessage(editingHost ? '更新成功' : '添加成功');
      setDialogOpen(false);
      fetchHosts();
    } catch (error) {
      showMessage(error.message || '操作失败', 'error');
    }
  };

  const handleSelectAll = (e) => {
    if (e.target.checked) {
      setSelectedIds(filteredHosts.map((h) => h.id));
    } else {
      setSelectedIds([]);
    }
  };

  const handleSelectOne = (id) => {
    if (selectedIds.includes(id)) {
      setSelectedIds(selectedIds.filter((i) => i !== id));
    } else {
      setSelectedIds([...selectedIds, id]);
    }
  };

  // Pin/Unpin Host
  const handlePinHost = async (host) => {
    const newPinned = !host.pinned;
    try {
      await pinHost(host.id, newPinned);
      // 成功（code === 200 时返回，否则被拦截器 reject）
      showMessage(newPinned ? '已固定' : '已取消固定');
      fetchHosts();
    } catch (error) {
      showMessage(error.message || '操作失败', 'error');
    }
  };

  // 格式化过期时间显示
  const formatExpireTime = (host) => {
    if (host.pinned) return '已固定';
    if (!host.expireAt) return '永不过期';
    try {
      const expireDate = new Date(host.expireAt);
      const now = new Date();
      if (expireDate <= now) return '已过期';
      const diffHours = Math.ceil((expireDate - now) / (1000 * 60 * 60));
      if (diffHours <= 24) return `${diffHours}小时后过期`;
      const diffDays = Math.ceil(diffHours / 24);
      return `${diffDays}天后过期`;
    } catch {
      return '-';
    }
  };

  // 检查是否即将过期（24小时内）
  const isExpiringSoon = (host) => {
    if (host.pinned || !host.expireAt) return false;
    try {
      const expireDate = new Date(host.expireAt);
      const now = new Date();
      const diffHours = (expireDate - now) / (1000 * 60 * 60);
      return diffHours > 0 && diffHours <= 24;
    } catch {
      return false;
    }
  };

  // ========== 文本编辑模式操作 ==========

  const handleTextSync = async () => {
    setSyncing(true);
    try {
      const res = await syncHosts(textContent);
      // 成功（code === 200 时返回，否则被拦截器 reject）
      const { added, updated, deleted } = res.data || {};
      showMessage(`同步成功：新增 ${added || 0}，更新 ${updated || 0}，删除 ${deleted || 0}`);
      setOriginalText(textContent);
      // 刷新列表以同步状态
      fetchHosts();
    } catch (error) {
      showMessage(error.message || '同步失败', 'error');
    } finally {
      setSyncing(false);
    }
  };

  const handleTextReset = () => {
    setTextContent(originalText);
    showMessage('已还原');
  };

  const handleCopyText = () => {
    navigator.clipboard.writeText(textContent).then(() => {
      showMessage('已复制到剪贴板');
    });
  };

  const hasTextChanged = textContent !== originalText;

  // 格式化日期
  const formatDate = (dateStr) => {
    if (!dateStr) return '-';
    try {
      const date = new Date(dateStr);
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      });
    } catch {
      return dateStr;
    }
  };

  // 移动端 Host 卡片
  const MobileHostCard = ({ host }) => (
    <MainCard
      content={false}
      border={false}
      sx={{
        mb: 2,
        borderRadius: 3,
        transition: 'all 0.2s ease',
        border: '1px solid',
        borderColor: selectedIds.includes(host.id) ? theme.palette.primary.main : theme.palette.divider,
        borderLeft: selectedIds.includes(host.id)
          ? `4px solid ${theme.palette.primary.main}`
          : host.pinned
            ? `4px solid ${theme.palette.warning.main}`
            : `1px solid ${theme.palette.divider}`,
        '&:hover': {
          boxShadow: theme.shadows[4],
          transform: 'translateY(-2px)'
        }
      }}
    >
      <Box sx={{ p: 2 }}>
        <Stack spacing={1.5}>
          {/* 顶部：勾选 + 域名 + 操作 */}
          <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, minWidth: 0, flex: 1 }}>
              <Checkbox
                size="small"
                checked={selectedIds.includes(host.id)}
                onChange={() => handleSelectOne(host.id)}
                sx={{ p: 0.5, mt: -0.5 }}
              />
              <Box sx={{ minWidth: 0, flex: 1 }}>
                <Typography variant="subtitle1" fontWeight={700} sx={{ wordBreak: 'break-all', lineHeight: 1.3 }}>
                  {host.hostname}
                </Typography>
              </Box>
              {host.pinned && <PushPinIcon fontSize="small" color="warning" sx={{ flexShrink: 0 }} />}
            </Box>

            {/* 操作按钮 */}
            <Stack direction="row" spacing={0} sx={{ mt: -0.5, mr: -1 }}>
              <Tooltip title="编辑">
                <IconButton size="small" onClick={() => handleEdit(host)}>
                  <EditIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              <Tooltip title="删除">
                <IconButton size="small" color="error" onClick={() => handleDelete(host)}>
                  <DeleteIcon fontSize="small" />
                </IconButton>
              </Tooltip>
            </Stack>
          </Box>

          {/* 信息区 */}
          <Box>
            <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, alignItems: 'center', mb: 1.5 }}>
              {/* IP Chip - Clickable */}
              <Chip
                label={host.ip}
                size="small"
                color="primary"
                variant="outlined"
                onClick={(e) => {
                  e.stopPropagation();
                  handleIPClick(host.ip);
                }}
                sx={{ fontFamily: 'monospace', fontWeight: 500, cursor: 'pointer' }}
              />

              {/* Source Chip */}
              {host.source && (
                <Chip
                  label={host.source}
                  size="small"
                  variant="outlined"
                  color={host.source === '手动添加' || host.source === '手动导入' ? 'default' : 'info'}
                  sx={{ height: 24, borderRadius: 1 }}
                />
              )}

              {/* Expire Chip */}
              <Chip
                label={formatExpireTime(host)}
                size="small"
                variant="outlined"
                color={isExpiringSoon(host) ? 'warning' : host.pinned ? 'warning' : 'default'}
                sx={{ height: 24, borderRadius: 1 }}
              />
            </Box>

            {host.remark && (
              <Typography
                variant="body2"
                color="text.secondary"
                sx={{
                  mb: 1.5,
                  bgcolor: theme.palette.mode === 'dark' ? theme.palette.background.default : theme.palette.grey[50],
                  p: 1,
                  borderRadius: 1
                }}
              >
                {host.remark}
              </Typography>
            )}

            <Stack direction="row" justifyContent="space-between" alignItems="center">
              <Box />
              <Typography variant="caption" color="text.disabled">
                更新于 {formatDate(host.updatedAt)}
              </Typography>
            </Stack>
          </Box>
        </Stack>
      </Box>
    </MainCard>
  );

  return (
    <MainCard title="Host 管理">
      {/* 顶部工具栏 */}
      <Box sx={{ mb: 2 }}>
        <Stack direction="row" spacing={1} alignItems="center" justifyContent="space-between" flexWrap="wrap" sx={{ gap: 1 }}>
          {/* 模式切换 */}
          <ToggleButtonGroup
            value={viewMode}
            exclusive
            onChange={(e, v) => v && setViewMode(v)}
            size="small"
            sx={{
              '& .MuiToggleButton-root': {
                px: isMobile ? 1.5 : 2,
                py: 0.75,
                fontSize: isMobile ? '0.8rem' : '0.875rem',
                '&.Mui-selected': {
                  backgroundColor: theme.palette.primary.main,
                  color: '#fff',
                  '&:hover': {
                    backgroundColor: theme.palette.primary.dark
                  }
                }
              }
            }}
          >
            <ToggleButton value="table">
              <TableViewIcon sx={{ mr: 0.5, fontSize: isMobile ? '1rem' : '1.25rem' }} />
              表单模式
            </ToggleButton>
            <ToggleButton value="text">
              <CodeIcon sx={{ mr: 0.5, fontSize: isMobile ? '1rem' : '1.25rem' }} />
              文本模式
            </ToggleButton>
          </ToggleButtonGroup>

          <Stack direction="row" spacing={1} alignItems="center">
            {hosts.length > 0 && (
              <Chip label={`总数: ${hosts.length}`} size="small" variant="outlined" color="primary" sx={{ mr: 1, fontWeight: 500 }} />
            )}
            <Tooltip title="模块设置">
              <IconButton onClick={() => setSettingsOpen(true)} color="primary" size="small">
                <SettingsIcon />
              </IconButton>
            </Tooltip>
            <Tooltip title="刷新">
              <IconButton onClick={handleRefresh} disabled={refreshing} color="primary" size="small">
                {refreshing ? <CircularProgress size={18} /> : <RefreshIcon />}
              </IconButton>
            </Tooltip>
          </Stack>
        </Stack>
      </Box>

      <HostSettingsDialog open={settingsOpen} onClose={() => setSettingsOpen(false)} />

      <ConfirmDialog
        open={confirmOpen}
        title={confirmInfo.title}
        content={confirmInfo.content}
        onClose={() => setConfirmOpen(false)}
        onConfirm={confirmInfo.onConfirm}
      />

      <IPDetailsDialog open={ipDialogOpen} onClose={() => setIpDialogOpen(false)} ip={selectedIP} />

      {/* 表单模式 */}

      {viewMode === 'table' && (
        <Box>
          {/* 搜索和操作栏 */}
          <Box sx={{ mb: 2, display: 'flex', gap: 2, flexWrap: 'wrap', alignItems: 'center', justifyContent: 'space-between' }}>
            <TextField
              placeholder="搜索域名、IP、备注..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              size="small"
              sx={{ minWidth: { xs: '100%', sm: 280 }, flex: { xs: '1 1 100%', sm: '0 1 auto' } }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon color="action" fontSize="small" />
                  </InputAdornment>
                ),
                endAdornment: search && (
                  <InputAdornment position="end">
                    <IconButton size="small" onClick={() => setSearch('')}>
                      <ClearIcon fontSize="small" />
                    </IconButton>
                  </InputAdornment>
                )
              }}
            />
            <Stack direction="row" spacing={1}>
              {selectedIds.length > 0 && (
                <Button variant="outlined" color="error" startIcon={<DeleteIcon />} onClick={handleBatchDelete} size="small">
                  删除 ({selectedIds.length})
                </Button>
              )}
              <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd} size={isMobile ? 'small' : 'medium'}>
                添加
              </Button>
            </Stack>
          </Box>

          {/* 搜索结果统计 */}
          {search && (
            <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
              找到 {filteredHosts.length} 条匹配结果
            </Typography>
          )}

          {/* 移动端卡片 / 桌面端表格 */}
          {isMobile ? (
            <Box>
              {filteredHosts.length > 0 && (
                <Box sx={{ mb: 1, display: 'flex', alignItems: 'center' }}>
                  <Checkbox
                    size="small"
                    checked={selectedIds.length === filteredHosts.length && filteredHosts.length > 0}
                    indeterminate={selectedIds.length > 0 && selectedIds.length < filteredHosts.length}
                    onChange={handleSelectAll}
                  />
                  <Typography variant="body2" color="text.secondary">
                    全选
                  </Typography>
                </Box>
              )}
              {filteredHosts.map((host) => (
                <MobileHostCard key={host.id} host={host} />
              ))}
              {filteredHosts.length === 0 && (
                <Typography color="text.secondary" align="center" sx={{ py: 4 }}>
                  {hosts.length === 0 ? '暂无 Host 记录，点击"添加"创建' : '没有找到匹配的结果'}
                </Typography>
              )}
            </Box>
          ) : (
            <TableContainer component={Paper} sx={{ borderRadius: 2 }}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell padding="checkbox">
                      <Checkbox
                        size="small"
                        checked={selectedIds.length === filteredHosts.length && filteredHosts.length > 0}
                        indeterminate={selectedIds.length > 0 && selectedIds.length < filteredHosts.length}
                        onChange={handleSelectAll}
                      />
                    </TableCell>
                    <TableCell>域名</TableCell>
                    <TableCell>IP 地址</TableCell>
                    <TableCell>来源</TableCell>
                    <TableCell>备注</TableCell>
                    <TableCell>状态</TableCell>
                    <TableCell>更新时间</TableCell>
                    <TableCell align="right">操作</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {filteredHosts.map((host) => (
                    <TableRow
                      key={host.id}
                      hover
                      selected={selectedIds.includes(host.id)}
                      sx={{ '&:hover': { backgroundColor: theme.palette.action.hover } }}
                    >
                      <TableCell padding="checkbox">
                        <Checkbox size="small" checked={selectedIds.includes(host.id)} onChange={() => handleSelectOne(host.id)} />
                      </TableCell>
                      <TableCell sx={{ fontWeight: 500 }}>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                          {host.hostname}
                          {host.pinned && <PushPinIcon fontSize="small" color="warning" sx={{ fontSize: 14 }} />}
                        </Box>
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={host.ip}
                          size="small"
                          variant="outlined"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleIPClick(host.ip);
                          }}
                          sx={{ fontFamily: 'monospace', cursor: 'pointer' }}
                        />
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={host.source || '-'}
                          size="small"
                          variant="outlined"
                          color={host.source === '手动添加' || host.source === '手动导入' ? 'default' : 'info'}
                          sx={{ height: 22, fontSize: '0.75rem' }}
                        />
                      </TableCell>
                      <TableCell>
                        <Tooltip title={host.remark || ''} placement="top-start">
                          <Typography
                            variant="body2"
                            color="text.secondary"
                            sx={{ maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', cursor: 'pointer' }}
                          >
                            {host.remark || '-'}
                          </Typography>
                        </Tooltip>
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={formatExpireTime(host)}
                          size="small"
                          variant="outlined"
                          color={isExpiringSoon(host) ? 'warning' : host.pinned ? 'warning' : 'default'}
                          sx={{ height: 22 }}
                        />
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" color="text.secondary" sx={{ whiteSpace: 'nowrap' }}>
                          {formatDate(host.updatedAt)}
                        </Typography>
                      </TableCell>
                      <TableCell align="right">
                        <Tooltip title={host.pinned ? '取消固定' : '固定 (pin住的host不受过期时间影响)'}>
                          <IconButton size="small" onClick={() => handlePinHost(host)} color={host.pinned ? 'warning' : 'default'}>
                            {host.pinned ? <PushPinIcon fontSize="small" /> : <PushPinOutlinedIcon fontSize="small" />}
                          </IconButton>
                        </Tooltip>
                        <IconButton size="small" onClick={() => handleEdit(host)}>
                          <EditIcon fontSize="small" />
                        </IconButton>
                        <IconButton size="small" color="error" onClick={() => handleDelete(host)}>
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                  ))}
                  {filteredHosts.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={8} align="center" sx={{ py: 4 }}>
                        <Typography color="text.secondary">
                          {hosts.length === 0 ? '暂无 Host 记录，点击"添加"创建' : '没有找到匹配的结果'}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </Box>
      )}

      {/* 文本编辑模式 */}
      {viewMode === 'text' && (
        <Box>
          <Alert severity="info" sx={{ mb: 2 }}>
            每行一条记录，格式：<code>域名 IP</code> 或 <code>域名 IP # 备注</code>。保存后将与数据库同步（新增、修改、删除）。
          </Alert>

          <TextField
            multiline
            fullWidth
            rows={isMobile ? 12 : 18}
            value={textContent}
            onChange={(e) => setTextContent(e.target.value)}
            placeholder={`# 示例格式：\nexample.com 192.168.1.1\napi.example.com 192.168.1.2 # API 服务器`}
            sx={{
              mb: 2,
              '& .MuiInputBase-root': {
                fontFamily: 'monospace',
                fontSize: '0.9rem',
                lineHeight: 1.6,
                backgroundColor: theme.palette.mode === 'dark' ? theme.palette.grey[900] : theme.palette.grey[50]
              }
            }}
          />

          <Stack direction={isMobile ? 'column' : 'row'} spacing={isMobile ? 1.5 : 2} alignItems={isMobile ? 'stretch' : 'center'}>
            <Stack direction="row" spacing={1} sx={{ width: isMobile ? '100%' : 'auto' }}>
              <Button
                variant="contained"
                startIcon={syncing ? <CircularProgress size={16} color="inherit" /> : <SaveIcon />}
                onClick={handleTextSync}
                disabled={!hasTextChanged || syncing}
                fullWidth={isMobile}
                size={isMobile ? 'medium' : 'medium'}
              >
                保存同步
              </Button>
              <Button
                variant="outlined"
                startIcon={<UndoIcon />}
                onClick={handleTextReset}
                disabled={!hasTextChanged}
                size={isMobile ? 'medium' : 'medium'}
              >
                还原
              </Button>
              <Button variant="text" startIcon={<ContentCopyIcon />} onClick={handleCopyText} size={isMobile ? 'medium' : 'medium'}>
                复制
              </Button>
            </Stack>
          </Stack>

          {hasTextChanged && (
            <Typography variant="body2" color="warning.main" sx={{ mt: 1 }}>
              * 内容已修改，请保存同步
            </Typography>
          )}
        </Box>
      )}

      {/* 添加/编辑对话框 */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{editingHost ? '编辑 Host' : '添加 Host'}</DialogTitle>
        <DialogContent>
          <Stack spacing={2}>
            <TextField
              label="域名"
              value={formData.hostname}
              onChange={(e) => setFormData({ ...formData, hostname: e.target.value })}
              fullWidth
              placeholder="example.com"
              disabled={!!editingHost}
              helperText={editingHost ? '域名不可修改' : ''}
            />
            <TextField
              label="IP 地址"
              value={formData.ip}
              onChange={(e) => setFormData({ ...formData, ip: e.target.value })}
              fullWidth
              placeholder="192.168.1.1"
            />
            <TextField
              label="备注"
              value={formData.remark}
              onChange={(e) => setFormData({ ...formData, remark: e.target.value })}
              fullWidth
              placeholder="可选"
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>取消</Button>
          <Button variant="contained" onClick={handleSave}>
            {editingHost ? '保存' : '添加'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Snackbar */}
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
