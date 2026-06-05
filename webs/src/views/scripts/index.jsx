import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

// material-ui
import { alpha, useTheme } from '@mui/material/styles';
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
import Link from '@mui/material/Link';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';

// icons
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import RefreshIcon from '@mui/icons-material/Refresh';
import HelpOutlineIcon from '@mui/icons-material/HelpOutline';
import FullscreenIcon from '@mui/icons-material/Fullscreen';
import FullscreenExitIcon from '@mui/icons-material/FullscreenExit';

import MainCard from 'ui-component/cards/MainCard';
import Pagination from 'components/Pagination';
import { getScripts, addScript, updateScript, deleteScript, getScriptUsage } from 'api/scripts';
import { formatDateTime } from 'i18n/locales';

// Monaco Editor
import Editor from '@monaco-editor/react';

const buildDefaultScript = (t) => `// ${t('scripts.template.modifyNodes')}
/**
 * @param {Node[]} nodes
 * @param {string} clientType
 */
function filterNode(nodes, clientType) {
    // nodes: ${t('scripts.template.nodes')}
    // clientType: ${t('scripts.template.clientType')}
    // ${t('scripts.template.returnNodes')}
    return nodes;
}

// ${t('scripts.template.modifySubscription')}
/**
 * @param {string} input
 * @param {string} clientType
 */
function subMod(input, clientType) {
    // input: ${t('scripts.template.input')}
    // clientType: ${t('scripts.template.clientType')}
    // ${t('scripts.template.returnContent')}
    return input;
}`;

// ==============================|| 脚本管理 ||============================== //

export default function ScriptList() {
  const { t, i18n } = useTranslation();
  const theme = useTheme();
  const matchDownMd = useMediaQuery(theme.breakpoints.down('md'));

  const [scripts, setScripts] = useState([]);
  const [loading, setLoading] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [isEdit, setIsEdit] = useState(false);
  const [currentScript, setCurrentScript] = useState(null);
  const [formData, setFormData] = useState(() => ({ name: '', version: '0.0.0', content: buildDefaultScript(t) }));
  const [editorFullscreen, setEditorFullscreen] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });
  const [usageDialog, setUsageDialog] = useState({ open: false, title: '', message: '', subscriptions: [], action: null });
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(() => {
    const saved = localStorage.getItem('scripts_rowsPerPage');
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

  const fetchScripts = async (currentPage, currentPageSize) => {
    setLoading(true);
    try {
      const response = await getScripts({ page: currentPage + 1, pageSize: currentPageSize });
      // 处理分页响应
      if (response.data && response.data.items !== undefined) {
        setScripts(response.data.items || []);
        setTotalItems(response.data.total || 0);
      } else {
        // 向后兼容：老格式直接返回数组
        setScripts(response.data || []);
        setTotalItems((response.data || []).length);
      }
    } catch (error) {
      console.error(error);
      showMessage(error.message || t('scripts.messages.loadFailed'), 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    fetchScripts(page, rowsPerPage);
  };

  useEffect(() => {
    fetchScripts(0, rowsPerPage);
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const handleAdd = () => {
    setIsEdit(false);
    setCurrentScript(null);
    setFormData({ name: '', version: '0.0.0', content: buildDefaultScript(t) });
    setEditorFullscreen(false);
    setDialogOpen(true);
  };

  const handleEdit = (script) => {
    setIsEdit(true);
    setCurrentScript(script);
    setFormData({ name: script.name, version: script.version, content: script.content });
    setEditorFullscreen(false);
    setDialogOpen(true);
  };

  const handleDelete = async (script) => {
    let usedSubscriptions = [];

    try {
      const response = await getScriptUsage({ id: script.id });
      usedSubscriptions = response.data?.subscriptions || [];
    } catch (error) {
      console.error(error);
      showMessage(error.message || t('scripts.messages.usageFailed'), 'error');
      return;
    }

    const deleteAction = async () => {
      try {
        await deleteScript(script);
        showMessage(t('scripts.messages.deleteSuccess'));
        fetchScripts(page, rowsPerPage);
      } catch (error) {
        console.error(error);
        showMessage(error.message || t('scripts.messages.deleteFailed'), 'error');
      }
    };

    if (usedSubscriptions.length > 0) {
      setUsageDialog({
        open: true,
        title: t('scripts.usage.title'),
        message: t('scripts.usage.message', { name: script.name }),
        subscriptions: usedSubscriptions,
        action: deleteAction
      });
      return;
    }

    openConfirm(t('scripts.delete.title'), t('scripts.delete.confirm', { name: script.name }), deleteAction);
  };

  const handleSubmit = async () => {
    try {
      if (isEdit) {
        await updateScript({ ...formData, id: currentScript.id });
        showMessage(t('scripts.messages.updateSuccess'));
      } else {
        await addScript(formData);
        showMessage(t('scripts.messages.createSuccess'));
      }
      setEditorFullscreen(false);
      setDialogOpen(false);
      fetchScripts(page, rowsPerPage);
    } catch (error) {
      console.error(error);
      showMessage(error.message || (isEdit ? t('scripts.messages.updateFailed') : t('scripts.messages.createFailed')), 'error');
    }
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return formatDateTime(date, i18n.resolvedLanguage || i18n.language);
  };

  const handleCloseDialog = () => {
    setDialogOpen(false);
    setEditorFullscreen(false);
  };

  return (
    <MainCard
      title={t('scripts.title')}
      secondary={
        matchDownMd ? (
          <Button variant="contained" size="small" startIcon={<AddIcon />} onClick={handleAdd}>
            {t('scripts.actions.add')}
          </Button>
        ) : (
          <Stack direction="row" spacing={1} alignItems="center">
            <Link
              href="https://github.com/ZeroDeng01/sublinkPro/blob/main/docs/script_support.md"
              target="_blank"
              rel="noopener"
              sx={{ display: 'flex', alignItems: 'center' }}
            >
              <HelpOutlineIcon sx={{ mr: 0.5 }} fontSize="small" />
              {t('scripts.actions.usageGuide')}
            </Link>
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
              {t('scripts.actions.addScript')}
            </Button>
            <IconButton onClick={handleRefresh} disabled={loading}>
              <RefreshIcon />
            </IconButton>
          </Stack>
        )
      }
    >
      {matchDownMd && (
        <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
          <Link
            href="https://github.com/ZeroDeng01/sublinkPro/blob/main/docs/script_support.md"
            target="_blank"
            rel="noopener"
            sx={{ display: 'flex', alignItems: 'center' }}
          >
            <HelpOutlineIcon sx={{ mr: 0.5 }} fontSize="small" />
            {t('scripts.actions.usageGuide')}
          </Link>
          <IconButton onClick={handleRefresh} disabled={loading} size="small">
            <RefreshIcon />
          </IconButton>
        </Stack>
      )}

      {matchDownMd ? (
        <Stack spacing={2}>
          {scripts.map((script) => (
            <MainCard key={script.id} content={false} border shadow={theme.shadows[1]}>
              <Box p={2}>
                <Stack direction="row" justifyContent="space-between" alignItems="center" mb={1}>
                  <Chip label={script.name} color="success" variant="outlined" size="small" />
                  <Typography variant="caption" color="textSecondary">
                    v{script.version}
                  </Typography>
                </Stack>

                <Typography variant="caption" color="textSecondary" display="block">
                  {t('scripts.fields.updatedAt')}: {formatDate(script.updated_at)}
                </Typography>

                <Divider sx={{ my: 1 }} />

                <Stack direction="row" justifyContent="flex-end" spacing={1}>
                  <IconButton size="small" onClick={() => handleEdit(script)}>
                    <EditIcon fontSize="small" />
                  </IconButton>
                  <IconButton size="small" color="error" onClick={() => handleDelete(script)}>
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
                <TableCell>{t('scripts.fields.name')}</TableCell>
                <TableCell>{t('scripts.fields.version')}</TableCell>
                <TableCell>{t('scripts.fields.createdAt')}</TableCell>
                <TableCell>{t('scripts.fields.updatedAt')}</TableCell>
                <TableCell align="right">{t('scripts.fields.actions')}</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {scripts.map((script) => (
                <TableRow key={script.id} hover>
                  <TableCell>
                    <Chip label={script.name} color="success" variant="outlined" size="small" />
                  </TableCell>
                  <TableCell>{script.version}</TableCell>
                  <TableCell>{formatDate(script.created_at)}</TableCell>
                  <TableCell>{formatDate(script.updated_at)}</TableCell>
                  <TableCell align="right">
                    <IconButton size="small" onClick={() => handleEdit(script)}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" color="error" onClick={() => handleDelete(script)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      <Pagination
        page={page}
        pageSize={rowsPerPage}
        totalItems={totalItems}
        onPageChange={(e, newPage) => {
          setPage(newPage);
          fetchScripts(newPage, rowsPerPage);
        }}
        onPageSizeChange={(e) => {
          const newValue = parseInt(e.target.value, 10);
          setRowsPerPage(newValue);
          localStorage.setItem('scripts_rowsPerPage', newValue);
          setPage(0);
          fetchScripts(0, newValue);
        }}
        pageSizeOptions={[10, 20, 50, 100]}
      />

      {/* 添加/编辑对话框 */}
      <Dialog
        open={dialogOpen}
        onClose={handleCloseDialog}
        maxWidth={editorFullscreen ? false : 'lg'}
        fullWidth
        fullScreen={editorFullscreen}
        PaperProps={{
          sx: editorFullscreen
            ? {
                height: '100vh',
                maxHeight: '100vh',
                m: 0
              }
            : undefined
        }}
      >
        <DialogTitle
          sx={
            editorFullscreen
              ? {
                  pb: 1,
                  display: 'flex',
                  flexDirection: { xs: 'column', md: 'row' },
                  alignItems: { xs: 'stretch', md: 'center' },
                  justifyContent: 'space-between',
                  gap: 1.5
                }
              : undefined
          }
        >
          <Stack spacing={0.5}>
            <Typography variant="h4">{isEdit ? t('scripts.dialog.editTitle') : t('scripts.dialog.addTitle')}</Typography>
          </Stack>
          {editorFullscreen && (
            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap justifyContent="flex-end">
              <Button variant="outlined" size="small" startIcon={<FullscreenExitIcon />} onClick={() => setEditorFullscreen(false)}>
                {t('scripts.actions.exitFullscreen')}
              </Button>
            </Stack>
          )}
        </DialogTitle>
        <DialogContent
          sx={
            editorFullscreen
              ? {
                  display: 'flex',
                  flexDirection: 'column',
                  overflow: 'hidden',
                  pt: 1,
                  pb: 2
                }
              : undefined
          }
        >
          <Stack spacing={2} sx={editorFullscreen ? { flex: 1, minHeight: 0 } : { mt: 1 }}>
            {!editorFullscreen && (
              <>
                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
                  <TextField
                    fullWidth
                    label={t('scripts.fields.name')}
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  />
                  <TextField
                    label={t('scripts.fields.version')}
                    value={formData.version}
                    onChange={(e) => setFormData({ ...formData, version: e.target.value })}
                    placeholder="0.0.0"
                    sx={{ width: 150 }}
                  />
                </Stack>
                <Stack direction="row" justifyContent="flex-end">
                  <Button variant="outlined" size="small" startIcon={<FullscreenIcon />} onClick={() => setEditorFullscreen(true)}>
                    {t('scripts.actions.fullscreen')}
                  </Button>
                </Stack>
              </>
            )}
            <Box
              sx={
                editorFullscreen
                  ? {
                      position: 'relative',
                      flex: 1,
                      minHeight: 0,
                      borderRadius: 1,
                      overflow: 'hidden'
                    }
                  : { position: 'relative' }
              }
            >
              <Editor
                height={editorFullscreen ? '100%' : '400px'}
                language="javascript"
                value={formData.content}
                onChange={(value) => setFormData({ ...formData, content: value || '' })}
                theme="vs-dark"
                options={{
                  minimap: { enabled: !matchDownMd },
                  fontSize: matchDownMd ? 12 : 14,
                  automaticLayout: true,
                  scrollBeyondLastLine: false,
                  lineNumbers: matchDownMd ? 'off' : 'on'
                }}
              />
            </Box>
          </Stack>
        </DialogContent>
        {!editorFullscreen && (
          <DialogActions>
            <Button onClick={handleCloseDialog}>{t('common.cancel')}</Button>
            <Button variant="contained" onClick={handleSubmit}>
              {t('common.confirm')}
            </Button>
          </DialogActions>
        )}
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
          <DialogContentText id="alert-dialog-description" sx={{ color: 'text.primary' }}>
            {confirmInfo.content}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleConfirmClose}>{t('common.cancel')}</Button>
          <Button onClick={handleConfirmAction} variant="contained" color="error" autoFocus>
            {t('common.confirm')}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={usageDialog.open}
        onClose={() => setUsageDialog({ ...usageDialog, open: false })}
        aria-labelledby="script-usage-dialog-title"
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle id="script-usage-dialog-title">⚠️ {usageDialog.title}</DialogTitle>
        <DialogContent>
          <Alert
            severity="warning"
            variant="outlined"
            sx={{
              mt: 1,
              alignItems: 'flex-start',
              backgroundColor: alpha(theme.palette.warning.main, 0.08),
              borderColor: alpha(theme.palette.warning.main, 0.28),
              color: 'text.primary',
              '& .MuiAlert-icon': {
                color: 'warning.dark',
                mt: '2px'
              },
              '& .MuiAlert-message': {
                width: '100%'
              }
            }}
          >
            {usageDialog.message}
          </Alert>
          {usageDialog.subscriptions?.length > 0 && (
            <Box sx={{ mt: 2 }}>
              <Typography variant="subtitle2" sx={{ mb: 1 }}>
                {t('scripts.usage.usedSubscriptions')}
              </Typography>
              <Stack spacing={1}>
                {usageDialog.subscriptions.map((subscriptionName) => (
                  <Chip key={subscriptionName} label={subscriptionName} color="warning" variant="outlined" sx={{ width: 'fit-content' }} />
                ))}
              </Stack>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setUsageDialog({ ...usageDialog, open: false, subscriptions: [], action: null })}>
            {t('common.cancel')}
          </Button>
          <Button
            variant="contained"
            color="error"
            onClick={async () => {
              const action = usageDialog.action;
              setUsageDialog({ open: false, title: '', message: '', subscriptions: [], action: null });
              if (action) {
                await action();
              }
            }}
            autoFocus
          >
            {t('scripts.actions.continueDelete')}
          </Button>
        </DialogActions>
      </Dialog>
    </MainCard>
  );
}
