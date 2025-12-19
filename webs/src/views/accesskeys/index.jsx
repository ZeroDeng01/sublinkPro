import { useState, useEffect } from 'react';

// material-ui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import IconButton from '@mui/material/IconButton';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import Radio from '@mui/material/Radio';
import RadioGroup from '@mui/material/RadioGroup';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormControl from '@mui/material/FormControl';
import FormLabel from '@mui/material/FormLabel';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';

// icons
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import RefreshIcon from '@mui/icons-material/Refresh';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';

import MainCard from 'ui-component/cards/MainCard';
import Pagination from 'components/Pagination';
import { getAccessKeys, createAccessKey, deleteAccessKey } from 'api/accesskeys';
import { useAuth } from 'contexts/AuthContext';

// ==============================|| API 密钥管理 ||============================== //

export default function ApiKeyList() {
  const { user } = useAuth();
  const theme = useTheme();
  const matchDownMd = useMediaQuery(theme.breakpoints.down('md'));

  const [accessKeys, setAccessKeys] = useState([]);
  const [loading, setLoading] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [showKeyDialog, setShowKeyDialog] = useState(false);
  const [newKey, setNewKey] = useState('');
  const [formData, setFormData] = useState({
    description: '',
    expirationOption: 'never',
    expiredAt: null
  });
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  // 分页
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(() => {
    const saved = localStorage.getItem('accesskeys_rowsPerPage');
    return saved ? parseInt(saved, 10) : 10;
  });
  const [totalItems, setTotalItems] = useState(0);

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

  const fetchAccessKeys = async (currentPage, currentPageSize) => {
    if (!user?.userId) return;
    setLoading(true);
    try {
      const response = await getAccessKeys(user.userId, { page: currentPage + 1, pageSize: currentPageSize });
      // 处理分页响应
      if (response.data && response.data.items !== undefined) {
        setAccessKeys(response.data.items || []);
        setTotalItems(response.data.total || 0);
      } else {
        // 向后兼容：老格式直接返回数组
        setAccessKeys(response.data || []);
        setTotalItems((response.data || []).length);
      }
    } catch (error) {
      showMessage(error.message || '获取 API 密钥列表失败', 'error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (user?.userId) {
      fetchAccessKeys(0, rowsPerPage);
    }
  }, [user]); // eslint-disable-line react-hooks/exhaustive-deps

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const handleAdd = () => {
    setFormData({ description: '', expirationOption: 'never', expiredAt: null });
    setDialogOpen(true);
  };

  const handleDelete = async (accessKey) => {
    openConfirm('删除 API 密钥', '确定要删除此 API 密钥吗？此操作不可恢复！', async () => {
      try {
        await deleteAccessKey(accessKey.ID);
        showMessage('删除成功');
        fetchAccessKeys(page, rowsPerPage);
      } catch (error) {
        showMessage(error.message || '删除失败', 'error');
      }
    });
  };

  const handleSubmit = async () => {
    if (!formData.description) {
      showMessage('请输入描述', 'warning');
      return;
    }

    try {
      const params = {
        description: formData.description,
        username: user?.username
      };

      if (formData.expirationOption === 'custom') {
        if (!formData.expiredAt) {
          showMessage('请选择过期时间', 'warning');
          return;
        }
        params.expiredAt = formData.expiredAt.toISOString();
      }

      const response = await createAccessKey(params);
      setNewKey(response.data.accessKey);
      setDialogOpen(false);
      setShowKeyDialog(true);
      fetchAccessKeys(page, rowsPerPage);
    } catch (error) {
      showMessage(error.message || '创建失败', 'error');
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    showMessage('已复制到剪贴板');
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN');
  };

  const getExpirationStatus = (apiKey) => {
    if (!apiKey.ExpiredAt) {
      return <Chip label="永不过期" color="success" size="small" variant="outlined" />;
    }
    const expireDate = new Date(apiKey.ExpiredAt);
    const now = new Date();
    if (expireDate < now) {
      return <Chip label="已过期" color="error" size="small" variant="outlined" />;
    }
    const diffDays = Math.ceil((expireDate - now) / (1000 * 60 * 60 * 24));
    if (diffDays <= 7) {
      return <Chip label={`${diffDays}天后过期`} color="warning" size="small" variant="outlined" />;
    }
    return <Chip label={formatDate(apiKey.ExpiredAt)} color="info" size="small" variant="outlined" />;
  };

  return (
    <MainCard
      title="API 密钥管理"
      secondary={
        matchDownMd ? (
          <Button variant="contained" size="small" startIcon={<AddIcon />} onClick={handleAdd}>
            创建
          </Button>
        ) : (
          <Stack direction="row" spacing={1}>
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
              创建密钥
            </Button>
            <IconButton onClick={() => fetchAccessKeys(page, rowsPerPage)} disabled={loading}>
              <RefreshIcon />
            </IconButton>
          </Stack>
        )
      }
    >
      {matchDownMd && (
        <Stack direction="row" justifyContent="flex-end" sx={{ mb: 2 }}>
          <IconButton onClick={() => fetchAccessKeys(page, rowsPerPage)} disabled={loading} size="small">
            <RefreshIcon />
          </IconButton>
        </Stack>
      )}

      {matchDownMd ? (
        <Stack spacing={2}>
          {accessKeys.map((accessKey) => (
            <MainCard key={accessKey.ID} content={false} border shadow={theme.shadows[1]}>
              <Box p={2}>
                <Stack direction="row" justifyContent="space-between" alignItems="center" mb={1}>
                  <Typography variant="subtitle1">{accessKey.Description || '无描述'}</Typography>
                  {getExpirationStatus(accessKey)}
                </Stack>
                <Typography variant="caption" color="textSecondary" display="block" gutterBottom>
                  ID: {accessKey.ID}
                </Typography>
                <Typography variant="caption" color="textSecondary" display="block" gutterBottom>
                  创建时间: {formatDate(accessKey.CreatedAt)}
                </Typography>
                <Divider sx={{ my: 1 }} />
                <Stack direction="row" justifyContent="flex-end">
                  <IconButton size="small" color="error" onClick={() => handleDelete(accessKey)}>
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                </Stack>
              </Box>
            </MainCard>
          ))}
        </Stack>
      ) : (
        <TableContainer component={Paper}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>ID</TableCell>
                <TableCell>描述</TableCell>
                <TableCell>创建时间</TableCell>
                <TableCell>过期状态</TableCell>
                <TableCell align="right">操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {accessKeys.map((accessKey) => (
                <TableRow key={accessKey.ID} hover>
                  <TableCell>{accessKey.ID}</TableCell>
                  <TableCell>{accessKey.Description}</TableCell>
                  <TableCell>{formatDate(accessKey.CreatedAt)}</TableCell>
                  <TableCell>{getExpirationStatus(accessKey)}</TableCell>
                  <TableCell align="right">
                    <IconButton size="small" color="error" onClick={() => handleDelete(accessKey)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {accessKeys.length === 0 && !loading && (
        <Box sx={{ textAlign: 'center', py: 4 }}>
          <Typography color="textSecondary">暂无 API 密钥</Typography>
        </Box>
      )}

      <Pagination
        page={page}
        pageSize={rowsPerPage}
        totalItems={totalItems}
        onPageChange={(e, newPage) => {
          setPage(newPage);
          fetchAccessKeys(newPage, rowsPerPage);
        }}
        onPageSizeChange={(e) => {
          const newValue = parseInt(e.target.value, 10);
          setRowsPerPage(newValue);
          localStorage.setItem('accesskeys_rowsPerPage', newValue);
          setPage(0);
          fetchAccessKeys(0, newValue);
        }}
        pageSizeOptions={[10, 20, 50, 100]}
      />

      {/* 创建对话框 */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>创建 API 密钥</DialogTitle>
        <DialogContent>
          <Stack spacing={2}>
            <TextField
              fullWidth
              id="description"
              name="description"
              label="描述"
              autoComplete="off"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              inputProps={{ maxLength: 10 }}
              helperText="最多10个字符"
            />
            <FormControl component="fieldset">
              <FormLabel component="legend">过期设置</FormLabel>
              <RadioGroup
                row
                value={formData.expirationOption}
                onChange={(e) => setFormData({ ...formData, expirationOption: e.target.value })}
              >
                <FormControlLabel value="never" control={<Radio />} label="永不过期" />
                <FormControlLabel value="custom" control={<Radio />} label="自定义" />
              </RadioGroup>
            </FormControl>
            {formData.expirationOption === 'custom' && (
              <TextField
                fullWidth
                id="expiredAt"
                name="expiredAt"
                label="过期时间"
                type="datetime-local"
                autoComplete="off"
                value={
                  formData.expiredAt
                    ? new Date(formData.expiredAt.getTime() - formData.expiredAt.getTimezoneOffset() * 60000).toISOString().slice(0, 16)
                    : ''
                }
                onChange={(e) => {
                  const val = e.target.value;
                  setFormData({ ...formData, expiredAt: val ? new Date(val) : null });
                }}
                InputLabelProps={{
                  shrink: true
                }}
              />
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>取消</Button>
          <Button variant="contained" onClick={handleSubmit}>
            创建
          </Button>
        </DialogActions>
      </Dialog>

      {/* 显示新密钥对话框 */}
      <Dialog open={showKeyDialog} onClose={() => setShowKeyDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>API 密钥已创建</DialogTitle>
        <DialogContent>
          <Alert severity="warning" sx={{ mb: 2 }}>
            请立即保存此密钥，关闭后将无法再次查看！
          </Alert>
          <TextField
            fullWidth
            value={newKey}
            InputProps={{
              readOnly: true,
              endAdornment: (
                <IconButton onClick={() => copyToClipboard(newKey)}>
                  <ContentCopyIcon />
                </IconButton>
              )
            }}
          />
        </DialogContent>
        <DialogActions>
          <Button variant="contained" onClick={() => setShowKeyDialog(false)}>
            我已保存
          </Button>
        </DialogActions>
      </Dialog>

      {/* 提示消息 */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity}>{snackbar.message}</Alert>
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
    </MainCard>
  );
}
