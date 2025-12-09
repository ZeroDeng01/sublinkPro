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
import TablePagination from '@mui/material/TablePagination';
import Paper from '@mui/material/Paper';
import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';

// icons
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import RefreshIcon from '@mui/icons-material/Refresh';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import { getTemplates, addTemplate, updateTemplate, deleteTemplate } from 'api/templates';

// Monaco Editor
import Editor from '@monaco-editor/react';

// ==============================|| 模板管理 ||============================== //

export default function TemplateList() {
  const theme = useTheme();
  const matchDownMd = useMediaQuery(theme.breakpoints.down('md'));

  const [templates, setTemplates] = useState([]);
  const [loading, setLoading] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [isEdit, setIsEdit] = useState(false);
  const [currentTemplate, setCurrentTemplate] = useState(null);
  const [formData, setFormData] = useState({ filename: '', text: '' });
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(() => {
    const saved = localStorage.getItem('templates_rowsPerPage');
    return saved ? parseInt(saved, 10) : 10;
  });

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

  const fetchTemplates = async () => {
    setLoading(true);
    try {
      const response = await getTemplates();
      setTemplates(response.data || []);
    } catch (error) {
      console.log(error);
      showMessage('获取模板列表失败', 'error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTemplates();
  }, []);

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const handleAdd = () => {
    setIsEdit(false);
    setCurrentTemplate(null);
    setFormData({ filename: '', text: '' });
    setDialogOpen(true);
  };

  const handleEdit = (template) => {
    setIsEdit(true);
    setCurrentTemplate(template);
    setFormData({ filename: template.file, text: template.text });
    setDialogOpen(true);
  };

  const handleDelete = async (template) => {
    openConfirm('删除模板', `确定要删除模板 "${template.file}" 吗？`, async () => {
      try {
        await deleteTemplate({ filename: template.file });
        showMessage('删除成功');
        fetchTemplates();
      } catch (error) {
        console.log(error);
        showMessage('删除失败', 'error');
      }
    });
  };

  const handleSubmit = async () => {
    try {
      if (isEdit) {
        await updateTemplate({
          oldname: currentTemplate.file,
          filename: formData.filename,
          text: formData.text
        });
        showMessage('更新成功');
      } else {
        await addTemplate({
          filename: formData.filename,
          text: formData.text
        });
        showMessage('添加成功');
      }
      setDialogOpen(false);
      fetchTemplates();
    } catch (error) {
      console.log(error);
      showMessage(isEdit ? '更新失败' : '添加失败', 'error');
    }
  };

  return (
    <MainCard
      title="模板管理"
      secondary={
        matchDownMd ? (
          <Button variant="contained" size="small" startIcon={<AddIcon />} onClick={handleAdd}>
            添加
          </Button>
        ) : (
          <Stack direction="row" spacing={1}>
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
              添加模板
            </Button>
            <IconButton onClick={fetchTemplates} disabled={loading}>
              <RefreshIcon
                sx={
                  loading
                    ? {
                        animation: 'spin 1s linear infinite',
                        '@keyframes spin': { from: { transform: 'rotate(0deg)' }, to: { transform: 'rotate(360deg)' } }
                      }
                    : {}
                }
              />
            </IconButton>
          </Stack>
        )
      }
    >
      {matchDownMd && (
        <Stack direction="row" justifyContent="flex-end" sx={{ mb: 2 }}>
          <IconButton onClick={fetchTemplates} disabled={loading} size="small">
            <RefreshIcon
              sx={
                loading
                  ? {
                      animation: 'spin 1s linear infinite',
                      '@keyframes spin': { from: { transform: 'rotate(0deg)' }, to: { transform: 'rotate(360deg)' } }
                    }
                  : {}
              }
            />
          </IconButton>
        </Stack>
      )}

      {matchDownMd ? (
        <Stack spacing={2}>
          {templates.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map((template) => (
            <MainCard key={template.file} content={false} border shadow={theme.shadows[1]}>
              <Box p={2}>
                <Stack direction="row" justifyContent="space-between" alignItems="center" mb={1}>
                  <Chip label={template.file} color="success" variant="outlined" />
                  <Typography variant="caption" color="textSecondary">
                    {template.create_date || '-'}
                  </Typography>
                </Stack>

                <Divider sx={{ my: 1 }} />

                <Stack direction="row" justifyContent="flex-end" spacing={1}>
                  <IconButton size="small" onClick={() => handleEdit(template)}>
                    <EditIcon fontSize="small" />
                  </IconButton>
                  <IconButton size="small" color="error" onClick={() => handleDelete(template)}>
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
                <TableCell>文件名</TableCell>
                <TableCell>创建时间</TableCell>
                <TableCell align="right">操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {templates.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map((template) => (
                <TableRow key={template.file} hover>
                  <TableCell>
                    <Chip label={template.file} color="success" variant="outlined" size="small" />
                  </TableCell>
                  <TableCell>{template.create_date || '-'}</TableCell>
                  <TableCell align="right">
                    <IconButton size="small" onClick={() => handleEdit(template)}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" color="error" onClick={() => handleDelete(template)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      <TablePagination
        component="div"
        count={templates.length}
        page={page}
        onPageChange={(e, newPage) => setPage(newPage)}
        rowsPerPage={rowsPerPage}
        onRowsPerPageChange={(e) => {
          const newValue = parseInt(e.target.value, 10);
          setRowsPerPage(newValue);
          localStorage.setItem('templates_rowsPerPage', newValue);
          setPage(0);
        }}
        labelRowsPerPage="每页行数:"
      />

      {/* 添加/编辑对话框 */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="lg" fullWidth>
        <DialogTitle>{isEdit ? '编辑模板' : '添加模板'}</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              fullWidth
              label="文件名"
              value={formData.filename}
              onChange={(e) => setFormData({ ...formData, filename: e.target.value })}
              placeholder="例如: clash.yaml"
            />
            <Editor
              height="400px"
              language="yaml"
              value={formData.text}
              onChange={(value) => setFormData({ ...formData, text: value || '' })}
              theme="vs-dark"
              options={{
                minimap: { enabled: true },
                fontSize: 14
              }}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>取消</Button>
          <Button variant="contained" onClick={handleSubmit}>
            确定
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
