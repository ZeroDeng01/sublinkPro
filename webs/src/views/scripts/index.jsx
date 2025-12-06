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

// project imports
import MainCard from 'ui-component/cards/MainCard';
import { getScripts, addScript, updateScript, deleteScript } from 'api/scripts';

// Monaco Editor
import Editor from '@monaco-editor/react';

const DEFAULT_SCRIPT = `//修改节点列表
/**
 * @param {Node[]} nodes
 * @param {string} clientType
 */
function filterNode(nodes, clientType) {
    // nodes: 节点列表
    // clientType: 客户端类型
    // 返回值: 修改后节点列表
    return nodes;
}

//修改订阅文件
/**
 * @param {string} input
 * @param {string} clientType
 */
function subMod(input, clientType) {
    // input: 原始输入内容
    // clientType: 客户端类型
    // 返回值: 修改后的内容字符串
    return input;
}`;

// ==============================|| 脚本管理 ||============================== //

export default function ScriptList() {
  const theme = useTheme();
  const matchDownMd = useMediaQuery(theme.breakpoints.down('md'));

  const [scripts, setScripts] = useState([]);
  const [loading, setLoading] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [isEdit, setIsEdit] = useState(false);
  const [currentScript, setCurrentScript] = useState(null);
  const [formData, setFormData] = useState({ name: '', version: '0.0.0', content: DEFAULT_SCRIPT });
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);

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

  const fetchScripts = async () => {
    setLoading(true);
    try {
      const response = await getScripts();
      setScripts(response.data || []);
    } catch (error) {
      showMessage('获取脚本列表失败', 'error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchScripts();
  }, []);

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const handleAdd = () => {
    setIsEdit(false);
    setCurrentScript(null);
    setFormData({ name: '', version: '0.0.0', content: DEFAULT_SCRIPT });
    setDialogOpen(true);
  };

  const handleEdit = (script) => {
    setIsEdit(true);
    setCurrentScript(script);
    setFormData({ name: script.name, version: script.version, content: script.content });
    setDialogOpen(true);
  };

  const handleDelete = async (script) => {
    openConfirm('删除脚本', `确定要删除脚本 "${script.name}" 吗？`, async () => {
      try {
        await deleteScript(script);
        showMessage('删除成功');
        fetchScripts();
      } catch (error) {
        showMessage('删除失败', 'error');
      }
    });
  };

  const handleSubmit = async () => {
    try {
      if (isEdit) {
        await updateScript({ ...formData, id: currentScript.id });
        showMessage('更新成功');
      } else {
        await addScript(formData);
        showMessage('添加成功');
      }
      setDialogOpen(false);
      fetchScripts();
    } catch (error) {
      showMessage(isEdit ? '更新失败' : '添加失败', 'error');
    }
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN');
  };

  return (
    <MainCard
      title="脚本管理"
      secondary={
        matchDownMd ? (
          <Button variant="contained" size="small" startIcon={<AddIcon />} onClick={handleAdd}>
            添加
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
              使用说明
            </Link>
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
              添加脚本
            </Button>
            <IconButton onClick={fetchScripts} disabled={loading}>
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
            使用说明
          </Link>
          <IconButton onClick={fetchScripts} disabled={loading} size="small">
            <RefreshIcon />
          </IconButton>
        </Stack>
      )}

      {matchDownMd ? (
        <Stack spacing={2}>
          {scripts.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map((script) => (
            <MainCard key={script.id} content={false} border shadow={theme.shadows[1]}>
              <Box p={2}>
                <Stack direction="row" justifyContent="space-between" alignItems="center" mb={1}>
                  <Chip label={script.name} color="success" size="small" />
                  <Typography variant="caption" color="textSecondary">v{script.version}</Typography>
                </Stack>

                <Typography variant="caption" color="textSecondary" display="block">
                  更新于: {formatDate(script.updated_at)}
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
                <TableCell>脚本名称</TableCell>
                <TableCell>版本</TableCell>
                <TableCell>创建时间</TableCell>
                <TableCell>更新时间</TableCell>
                <TableCell align="right">操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {scripts.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map((script) => (
                <TableRow key={script.id} hover>
                  <TableCell>
                    <Chip label={script.name} color="success" size="small" />
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

      <TablePagination
        component="div"
        count={scripts.length}
        page={page}
        onPageChange={(e, newPage) => setPage(newPage)}
        rowsPerPage={rowsPerPage}
        onRowsPerPageChange={(e) => {
          setRowsPerPage(parseInt(e.target.value, 10));
          setPage(0);
        }}
        labelRowsPerPage="每页行数:"
      />

      {/* 添加/编辑对话框 */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="lg" fullWidth>
        <DialogTitle>{isEdit ? '编辑脚本' : '添加脚本'}</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <Stack direction="row" spacing={2}>
              <TextField
                fullWidth
                label="脚本名称"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
              <TextField
                label="版本"
                value={formData.version}
                onChange={(e) => setFormData({ ...formData, version: e.target.value })}
                placeholder="0.0.0"
                sx={{ width: 150 }}
              />
            </Stack>
            <Editor
              height="400px"
              language="javascript"
              value={formData.content}
              onChange={(value) => setFormData({ ...formData, content: value || '' })}
              theme="vs-dark"
              options={{
                minimap: { enabled: false },
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
          <DialogContentText id="alert-dialog-description">
            {confirmInfo.content}
          </DialogContentText>
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
