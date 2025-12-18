import PropTypes from 'prop-types';

// material-ui
import Autocomplete from '@mui/material/Autocomplete';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

/**
 * 批量修改来源对话框
 */
export default function BatchSourceDialog({ open, selectedCount, value, setValue, sourceOptions, onClose, onSubmit }) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>批量修改来源</DialogTitle>
      <DialogContent>
        <Typography variant="body2" color="textSecondary" sx={{ mb: 2 }}>
          将为选中的 {selectedCount} 个节点设置相同的来源
        </Typography>
        <Autocomplete
          freeSolo
          options={sourceOptions.filter((s) => s && s !== 'manual')}
          value={value}
          onChange={(e, newValue) => setValue(newValue || '')}
          onInputChange={(e, newInputValue) => setValue(newInputValue)}
          renderInput={(params) => <TextField {...params} label="来源名称" placeholder="输入或选择来源名称" fullWidth />}
        />
        <Typography variant="caption" color="textSecondary" sx={{ mt: 1, display: 'block' }}>
          提示：留空将设置为手动添加(manual)
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>取消</Button>
        <Button variant="contained" onClick={onSubmit}>
          确认修改
        </Button>
      </DialogActions>
    </Dialog>
  );
}

BatchSourceDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  selectedCount: PropTypes.number.isRequired,
  value: PropTypes.string.isRequired,
  setValue: PropTypes.func.isRequired,
  sourceOptions: PropTypes.array.isRequired,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired
};
