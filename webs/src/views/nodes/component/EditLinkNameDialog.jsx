import { useState } from 'react';
import PropTypes from 'prop-types';

// material-ui
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Alert from '@mui/material/Alert';
import CircularProgress from '@mui/material/CircularProgress';

// api
import { updateNodeLinkName } from 'api/nodes';

/**
 * 编辑节点原始名称对话框
 * 修改原始名称会同步更新节点连接(Link)
 */
export default function EditLinkNameDialog({ open, node, onClose, onSuccess }) {
  const [newLinkName, setNewLinkName] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // 当对话框打开时，初始化值
  const handleEnter = () => {
    setNewLinkName(node?.LinkName || '');
    setError('');
  };

  const handleSubmit = async () => {
    if (!newLinkName.trim()) {
      setError('名称不能为空');
      return;
    }
    if (newLinkName.trim() === node?.LinkName) {
      setError('新名称与原名称相同');
      return;
    }

    setLoading(true);
    setError('');
    try {
      await updateNodeLinkName(node.ID, newLinkName.trim());
      onSuccess && onSuccess();
      onClose();
    } catch (err) {
      setError(err.response?.data?.msg || '修改失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onClose={loading ? undefined : onClose} maxWidth="sm" fullWidth TransitionProps={{ onEnter: handleEnter }}>
      <DialogTitle>修改原始名称</DialogTitle>
      <DialogContent>
        <Alert severity="info" sx={{ mb: 2 }}>
          修改原始名称会同步更新节点连接(Link)中的名称部分。此操作不可撤销。
        </Alert>

        <Box sx={{ mb: 2 }}>
          <Typography variant="body2" color="textSecondary" gutterBottom>
            当前原始名称
          </Typography>
          <Typography
            variant="body1"
            sx={{
              p: 1.5,
              bgcolor: 'action.hover',
              borderRadius: 1,
              wordBreak: 'break-all'
            }}
          >
            {node?.LinkName || '-'}
          </Typography>
        </Box>

        <TextField
          label="新原始名称"
          value={newLinkName}
          onChange={(e) => setNewLinkName(e.target.value)}
          fullWidth
          autoFocus
          error={!!error}
          helperText={error}
          disabled={loading}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={loading}>
          取消
        </Button>
        <Button variant="contained" onClick={handleSubmit} disabled={loading || !newLinkName.trim()}>
          {loading ? <CircularProgress size={20} /> : '确认修改'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

EditLinkNameDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  node: PropTypes.object,
  onClose: PropTypes.func.isRequired,
  onSuccess: PropTypes.func
};
