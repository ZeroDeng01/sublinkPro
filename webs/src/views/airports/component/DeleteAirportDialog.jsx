import PropTypes from 'prop-types';

// material-ui
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';

/**
 * 删除机场确认对话框
 */
export default function DeleteAirportDialog({ open, airport, withNodes, setWithNodes, onClose, onConfirm }) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm">
      <DialogTitle>确认删除</DialogTitle>
      <DialogContent>
        <DialogContentText>
          确定要删除机场 <strong>&quot;{airport?.name}&quot;</strong> 吗？
        </DialogContentText>
        <FormControlLabel
          control={<Switch checked={withNodes} onChange={(e) => setWithNodes(e.target.checked)} />}
          label="同时删除关联节点"
          sx={{ mt: 2 }}
        />
        {withNodes && (
          <DialogContentText sx={{ mt: 1, color: 'error.main' }}>⚠️ 将同时删除此机场导入的所有节点，此操作不可恢复！</DialogContentText>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>取消</Button>
        <Button variant="contained" color="error" onClick={onConfirm}>
          确认删除
        </Button>
      </DialogActions>
    </Dialog>
  );
}

DeleteAirportDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  airport: PropTypes.shape({
    id: PropTypes.number,
    name: PropTypes.string
  }),
  withNodes: PropTypes.bool.isRequired,
  setWithNodes: PropTypes.func.isRequired,
  onClose: PropTypes.func.isRequired,
  onConfirm: PropTypes.func.isRequired
};
