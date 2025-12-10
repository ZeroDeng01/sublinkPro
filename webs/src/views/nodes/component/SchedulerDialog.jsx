import PropTypes from 'prop-types';

// material-ui
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import IconButton from '@mui/material/IconButton';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Tooltip from '@mui/material/Tooltip';

// icons
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import RefreshIcon from '@mui/icons-material/Refresh';

// utils
import { formatDateTime } from '../utils';

/**
 * 订阅调度器列表对话框
 */
export default function SchedulerDialog({ open, schedulers, loading, onClose, onRefresh, onAdd, onEdit, onDelete, onPull }) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="lg" fullWidth>
      <DialogTitle sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        导入订阅
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Button variant="contained" size="small" startIcon={<AddIcon />} onClick={onAdd}>
            添加订阅
          </Button>
          <IconButton onClick={onRefresh} disabled={loading}>
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
        </Box>
      </DialogTitle>
      <DialogContent>
        <TableContainer>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>名称</TableCell>
                <TableCell>URL</TableCell>
                <TableCell>节点数量</TableCell>
                <TableCell>上次运行</TableCell>
                <TableCell>下次运行</TableCell>
                <TableCell>Cron表达式</TableCell>
                <TableCell>分组</TableCell>
                <TableCell>状态</TableCell>
                <TableCell align="right">操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {schedulers.map((scheduler) => (
                <TableRow key={scheduler.ID}>
                  <TableCell>{scheduler.Name}</TableCell>
                  <TableCell sx={{ maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis' }}>{scheduler.URL}</TableCell>
                  <TableCell>
                    <Chip label={scheduler.NodeCount || 0} color="primary" variant="outlined" size="small" />
                  </TableCell>
                  <TableCell>{formatDateTime(scheduler.LastRunTime)}</TableCell>
                  <TableCell>{formatDateTime(scheduler.NextRunTime)}</TableCell>
                  <TableCell>{scheduler.CronExpr}</TableCell>
                  <TableCell>{scheduler.Group || '-'}</TableCell>
                  <TableCell>
                    <Chip label={scheduler.Enabled ? '启用' : '禁用'} color={scheduler.Enabled ? 'success' : 'default'} size="small" />
                  </TableCell>
                  <TableCell align="right">
                    <Tooltip title="立即拉取">
                      <IconButton size="small" onClick={() => onPull(scheduler)}>
                        <PlayArrowIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <IconButton size="small" onClick={() => onEdit(scheduler)}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" color="error" onClick={() => onDelete(scheduler)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>关闭</Button>
      </DialogActions>
    </Dialog>
  );
}

SchedulerDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  schedulers: PropTypes.array.isRequired,
  loading: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onRefresh: PropTypes.func.isRequired,
  onAdd: PropTypes.func.isRequired,
  onEdit: PropTypes.func.isRequired,
  onDelete: PropTypes.func.isRequired,
  onPull: PropTypes.func.isRequired
};
