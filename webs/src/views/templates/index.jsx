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
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Autocomplete from '@mui/material/Autocomplete';

// icons
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import RefreshIcon from '@mui/icons-material/Refresh';
import TransformIcon from '@mui/icons-material/Transform';
import UnfoldMoreIcon from '@mui/icons-material/UnfoldMore';
import CircularProgress from '@mui/material/CircularProgress';

import MainCard from 'ui-component/cards/MainCard';
import Pagination from 'components/Pagination';
import { getTemplates, addTemplate, updateTemplate, deleteTemplate, getACL4SSRPresets, convertRules } from 'api/templates';

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
  const [formData, setFormData] = useState({ filename: '', text: '', category: 'clash', ruleSource: '' });
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });
  const [aclPresets, setAclPresets] = useState([]);
  const [converting, setConverting] = useState(false);
  const [errorDialog, setErrorDialog] = useState({ open: false, title: '', message: '' });
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(() => {
    const saved = localStorage.getItem('templates_rowsPerPage');
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

  const fetchTemplates = async (currentPage, currentPageSize) => {
    setLoading(true);
    try {
      const response = await getTemplates({ page: currentPage + 1, pageSize: currentPageSize });
      // 处理分页响应
      if (response.data && response.data.items !== undefined) {
        setTemplates(response.data.items || []);
        setTotalItems(response.data.total || 0);
      } else {
        // 向后兼容：老格式直接返回数组
        setTemplates(response.data || []);
        setTotalItems((response.data || []).length);
      }
    } catch (error) {
      console.log(error);
      showMessage('获取模板列表失败', 'error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTemplates(0, rowsPerPage);
    // 获取 ACL4SSR 预设列表
    getACL4SSRPresets()
      .then((res) => {
        if (res.data) {
          setAclPresets(res.data);
        }
      })
      .catch((err) => console.log('获取预设列表失败:', err));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const handleAdd = () => {
    setIsEdit(false);
    setCurrentTemplate(null);
    setFormData({ filename: '', text: '', category: 'clash', ruleSource: '' });
    setDialogOpen(true);
  };

  const handleEdit = (template) => {
    setIsEdit(true);
    setCurrentTemplate(template);
    setFormData({
      filename: template.file,
      text: template.text,
      category: template.category || 'clash',
      ruleSource: template.ruleSource || ''
    });
    setDialogOpen(true);
  };

  const handleDelete = async (template) => {
    openConfirm('删除模板', `确定要删除模板 "${template.file}" 吗？`, async () => {
      try {
        await deleteTemplate({ filename: template.file });
        showMessage('删除成功');
        fetchTemplates(page, rowsPerPage);
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
          text: formData.text,
          category: formData.category,
          ruleSource: formData.ruleSource
        });
        showMessage('更新成功');
      } else {
        await addTemplate({
          filename: formData.filename,
          text: formData.text,
          category: formData.category,
          ruleSource: formData.ruleSource
        });
        showMessage('添加成功');
      }
      setDialogOpen(false);
      fetchTemplates(page, rowsPerPage);
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
            <IconButton onClick={() => fetchTemplates(page, rowsPerPage)} disabled={loading}>
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
          <IconButton onClick={() => fetchTemplates(page, rowsPerPage)} disabled={loading} size="small">
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
          {templates.map((template) => (
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
                <TableCell>类别</TableCell>
                <TableCell>规则源</TableCell>
                <TableCell>创建时间</TableCell>
                <TableCell align="right">操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {templates.map((template) => (
                <TableRow key={template.file} hover>
                  <TableCell>
                    <Chip label={template.file} color="success" variant="outlined" size="small" />
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={template.category === 'surge' ? 'Surge' : 'Clash'}
                      color={template.category === 'surge' ? 'secondary' : 'primary'}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" sx={{ maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {template.ruleSource || '-'}
                    </Typography>
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

      <Pagination
        page={page}
        pageSize={rowsPerPage}
        totalItems={totalItems}
        onPageChange={(e, newPage) => {
          setPage(newPage);
          fetchTemplates(newPage, rowsPerPage);
        }}
        onPageSizeChange={(e) => {
          const newValue = parseInt(e.target.value, 10);
          setRowsPerPage(newValue);
          localStorage.setItem('templates_rowsPerPage', newValue);
          setPage(0);
          fetchTemplates(0, newValue);
        }}
        pageSizeOptions={[10, 20, 50, 100]}
      />

      {/* 添加/编辑对话框 */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="lg" fullWidth>
        <DialogTitle>{isEdit ? '编辑模板' : '添加模板'}</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
              <TextField
                fullWidth
                label="文件名"
                value={formData.filename}
                onChange={(e) => setFormData({ ...formData, filename: e.target.value })}
                placeholder="例如: clash.yaml"
              />
              <FormControl sx={{ minWidth: 120 }}>
                <InputLabel>类别</InputLabel>
                <Select value={formData.category} label="类别" onChange={(e) => setFormData({ ...formData, category: e.target.value })}>
                  <MenuItem value="clash">Clash</MenuItem>
                  <MenuItem value="surge">Surge</MenuItem>
                </Select>
              </FormControl>
            </Stack>
            <Autocomplete
              freeSolo
              options={aclPresets}
              getOptionLabel={(option) => {
                if (typeof option === 'string') return option;
                return option.label || option.url || '';
              }}
              value={formData.ruleSource}
              onChange={(event, newValue) => {
                if (typeof newValue === 'string') {
                  setFormData({ ...formData, ruleSource: newValue });
                } else if (newValue && newValue.url) {
                  setFormData({ ...formData, ruleSource: newValue.url });
                } else {
                  setFormData({ ...formData, ruleSource: '' });
                }
              }}
              onInputChange={(event, newInputValue) => {
                setFormData({ ...formData, ruleSource: newInputValue });
              }}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="远程规则地址"
                  placeholder="输入 URL 或选择 ACL4SSR 预设"
                  helperText="可填写远程 ACL 规则配置地址，生成订阅时会动态加载规则"
                />
              )}
              renderOption={(props, option) => (
                <li {...props} key={option.name}>
                  <Stack>
                    <Typography variant="body2">{option.label}</Typography>
                    <Typography variant="caption" color="textSecondary" sx={{ fontSize: '0.7rem' }}>
                      {option.url}
                    </Typography>
                  </Stack>
                </li>
              )}
            />
            <Stack direction="row" spacing={1}>
              <Button
                variant="outlined"
                startIcon={converting ? <CircularProgress size={18} /> : <TransformIcon />}
                disabled={!formData.ruleSource || converting}
                onClick={async () => {
                  setConverting(true);
                  try {
                    const res = await convertRules({
                      ruleSource: formData.ruleSource,
                      category: formData.category,
                      expand: false,
                      template: formData.text
                    });
                    if (res.code === 200 && res.data && res.data.content) {
                      setFormData({ ...formData, text: res.data.content });
                      showMessage('规则生成/转换成功');
                    } else {
                      setErrorDialog({
                        open: true,
                        title: '规则生成/转换失败',
                        message: res.msg || '生成/转换过程中发生错误'
                      });
                    }
                  } catch (error) {
                    console.error(error);
                    const errorMsg = error.response?.data?.msg || error.message || '规则生成/转换失败';
                    setErrorDialog({
                      open: true,
                      title: '规则生成/转换失败',
                      message: errorMsg
                    });
                  } finally {
                    setConverting(false);
                  }
                }}
              >
                规则生成/转换
              </Button>
              <Button
                variant="outlined"
                startIcon={converting ? <CircularProgress size={18} /> : <UnfoldMoreIcon />}
                disabled={!formData.ruleSource || converting}
                onClick={async () => {
                  setConverting(true);
                  try {
                    const res = await convertRules({
                      ruleSource: formData.ruleSource,
                      category: formData.category,
                      expand: true,
                      template: formData.text
                    });
                    if (res.code === 200 && res.data && res.data.content) {
                      setFormData({ ...formData, text: res.data.content });
                      showMessage('规则生成/转换并展开成功');
                    } else {
                      setErrorDialog({
                        open: true,
                        title: '规则生成/转换失败',
                        message: res.msg || '转换过程中发生错误'
                      });
                    }
                  } catch (error) {
                    console.error(error);
                    const errorMsg = error.response?.data?.msg || error.message || '规则生成/转换失败';
                    setErrorDialog({
                      open: true,
                      title: '规则生成/转换失败',
                      message: errorMsg
                    });
                  } finally {
                    setConverting(false);
                  }
                }}
              >
                规则生成/转换（远程规则展开模式）
              </Button>
            </Stack>
            <Box sx={{ position: 'relative' }}>
              {converting && (
                <Box
                  sx={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    right: 0,
                    bottom: 0,
                    bgcolor: 'rgba(0,0,0,0.5)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    zIndex: 10,
                    borderRadius: 1
                  }}
                >
                  <Stack alignItems="center" spacing={1}>
                    <CircularProgress />
                    <Typography color="white">正在转换规则...</Typography>
                  </Stack>
                </Box>
              )}
              <Editor
                height="350px"
                language={formData.category === 'surge' ? 'ini' : 'yaml'}
                value={formData.text}
                onChange={(value) => setFormData({ ...formData, text: value || '' })}
                theme="vs-dark"
                options={{
                  minimap: { enabled: true },
                  fontSize: 14,
                  readOnly: converting
                }}
              />
            </Box>
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

      {/* 错误提示对话框 */}
      <Dialog
        open={errorDialog.open}
        onClose={() => setErrorDialog({ ...errorDialog, open: false })}
        aria-labelledby="error-dialog-title"
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle id="error-dialog-title" sx={{ color: 'error.main' }}>
          ⚠️ {errorDialog.title}
        </DialogTitle>
        <DialogContent>
          <Alert severity="error" sx={{ mt: 1 }}>
            {errorDialog.message}
          </Alert>
        </DialogContent>
        <DialogActions>
          <Button variant="contained" onClick={() => setErrorDialog({ ...errorDialog, open: false })} autoFocus>
            知道了
          </Button>
        </DialogActions>
      </Dialog>
    </MainCard>
  );
}
