import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

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
import { formatDateTime } from 'i18n/locales';

// ==============================|| API 密钥管理 ||============================== //

export default function ApiKeyList() {
  const { t, i18n } = useTranslation();
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
      showMessage(error.message || t('accessKeys.messages.loadFailed'), 'error');
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
    openConfirm(t('accessKeys.delete.title'), t('accessKeys.delete.confirm'), async () => {
      try {
        await deleteAccessKey(accessKey.ID);
        showMessage(t('accessKeys.messages.deleteSuccess'));
        fetchAccessKeys(page, rowsPerPage);
      } catch (error) {
        showMessage(error.message || t('accessKeys.messages.deleteFailed'), 'error');
      }
    });
  };

  const handleSubmit = async () => {
    if (!formData.description) {
      showMessage(t('accessKeys.messages.descriptionRequired'), 'warning');
      return;
    }

    try {
      const params = {
        description: formData.description,
        username: user?.username
      };

      if (formData.expirationOption === 'custom') {
        if (!formData.expiredAt) {
          showMessage(t('accessKeys.messages.expiredAtRequired'), 'warning');
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
      showMessage(error.message || t('accessKeys.messages.createFailed'), 'error');
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    showMessage(t('accessKeys.messages.copied'));
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return formatDateTime(date, i18n.resolvedLanguage || i18n.language);
  };

  const getExpirationStatus = (apiKey) => {
    if (!apiKey.ExpiredAt) {
      return <Chip label={t('accessKeys.expiration.never')} color="success" size="small" variant="outlined" />;
    }
    const expireDate = new Date(apiKey.ExpiredAt);
    const now = new Date();
    if (expireDate < now) {
      return <Chip label={t('accessKeys.expiration.expired')} color="error" size="small" variant="outlined" />;
    }
    const diffDays = Math.ceil((expireDate - now) / (1000 * 60 * 60 * 24));
    if (diffDays <= 7) {
      return <Chip label={t('accessKeys.expiration.expiresInDays', { count: diffDays })} color="warning" size="small" variant="outlined" />;
    }
    return <Chip label={formatDate(apiKey.ExpiredAt)} color="info" size="small" variant="outlined" />;
  };

  return (
    <MainCard
      title={t('accessKeys.title')}
      secondary={
        matchDownMd ? (
          <Button variant="contained" size="small" startIcon={<AddIcon />} onClick={handleAdd}>
            {t('accessKeys.actions.create')}
          </Button>
        ) : (
          <Stack direction="row" spacing={1}>
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
              {t('accessKeys.actions.createKey')}
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
                  <Typography variant="subtitle1">{accessKey.Description || t('accessKeys.noDescription')}</Typography>
                  {getExpirationStatus(accessKey)}
                </Stack>
                <Typography variant="caption" color="textSecondary" display="block" gutterBottom>
                  ID: {accessKey.ID}
                </Typography>
                <Typography variant="caption" color="textSecondary" display="block" gutterBottom>
                  {t('accessKeys.fields.createdAt')}: {formatDate(accessKey.CreatedAt)}
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
                <TableCell>{t('accessKeys.fields.description')}</TableCell>
                <TableCell>{t('accessKeys.fields.createdAt')}</TableCell>
                <TableCell>{t('accessKeys.fields.expiration')}</TableCell>
                <TableCell align="right">{t('accessKeys.fields.actions')}</TableCell>
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
          <Typography color="textSecondary">{t('accessKeys.empty')}</Typography>
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
        <DialogTitle>{t('accessKeys.createDialog.title')}</DialogTitle>
        <DialogContent>
          <Stack spacing={2}>
            <TextField
              fullWidth
              id="description"
              name="description"
              label={t('accessKeys.fields.description')}
              autoComplete="off"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              inputProps={{ maxLength: 10 }}
              helperText={t('accessKeys.createDialog.descriptionHelper')}
            />
            <FormControl component="fieldset">
              <FormLabel component="legend">{t('accessKeys.createDialog.expirationSettings')}</FormLabel>
              <RadioGroup
                row
                value={formData.expirationOption}
                onChange={(e) => setFormData({ ...formData, expirationOption: e.target.value })}
              >
                <FormControlLabel value="never" control={<Radio />} label={t('accessKeys.expiration.never')} />
                <FormControlLabel value="custom" control={<Radio />} label={t('accessKeys.expiration.custom')} />
              </RadioGroup>
            </FormControl>
            {formData.expirationOption === 'custom' && (
              <TextField
                fullWidth
                id="expiredAt"
                name="expiredAt"
                label={t('accessKeys.fields.expiredAt')}
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
          <Button onClick={() => setDialogOpen(false)}>{t('common.cancel')}</Button>
          <Button variant="contained" onClick={handleSubmit}>
            {t('accessKeys.actions.create')}
          </Button>
        </DialogActions>
      </Dialog>

      {/* 显示新密钥对话框 */}
      <Dialog open={showKeyDialog} onClose={() => setShowKeyDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{t('accessKeys.createdDialog.title')}</DialogTitle>
        <DialogContent>
          <Alert severity="warning" sx={{ mb: 2 }}>
            {t('accessKeys.createdDialog.warning')}
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
            {t('accessKeys.createdDialog.saved')}
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
          <Button onClick={handleConfirmClose}>{t('common.cancel')}</Button>
          <Button onClick={handleConfirmAction} color="primary" autoFocus>
            {t('common.confirm')}
          </Button>
        </DialogActions>
      </Dialog>
    </MainCard>
  );
}
