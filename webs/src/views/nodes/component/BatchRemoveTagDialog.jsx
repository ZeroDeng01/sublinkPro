import PropTypes from 'prop-types';

// material-ui
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

/**
 * 批量移除标签对话框
 */
export default function BatchRemoveTagDialog({ open, selectedCount, value, setValue, tagOptions, onClose, onSubmit }) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>批量移除标签</DialogTitle>
      <DialogContent>
        <Typography variant="body2" color="textSecondary" sx={{ mb: 2 }}>
          将从选中的 {selectedCount} 个节点移除以下标签（保留其他标签）
        </Typography>
        <Autocomplete
          multiple
          options={tagOptions}
          value={value}
          onChange={(e, newValue) => setValue(newValue)}
          getOptionLabel={(option) => option.name || option}
          isOptionEqualToValue={(option, val) => option.name === (val.name || val)}
          renderOption={(props, option) => {
            const { key, ...otherProps } = props;
            return (
              <li key={key} {...otherProps}>
                <Box
                  sx={{
                    width: 12,
                    height: 12,
                    borderRadius: '50%',
                    backgroundColor: option.color || '#1976d2',
                    mr: 1,
                    flexShrink: 0
                  }}
                />
                {option.name}
              </li>
            );
          }}
          renderTags={(val, getTagProps) =>
            val.map((option, index) => {
              const { key, ...tagProps } = getTagProps({ index });
              return (
                <Chip
                  key={key}
                  label={option.name || option}
                  size="small"
                  sx={{
                    backgroundColor: option.color || '#1976d2',
                    color: '#fff',
                    '& .MuiChip-deleteIcon': { color: 'rgba(255,255,255,0.7)' }
                  }}
                  {...tagProps}
                />
              );
            })
          }
          renderInput={(params) => <TextField {...params} label="选择要移除的标签" placeholder="选择标签" fullWidth />}
        />
        <Typography variant="caption" color="text.main" sx={{ mt: 1, display: 'block' }}>
          提示：只会移除选中的标签，节点的其他标签将保留
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>取消</Button>
        <Button variant="contained" color="warning" onClick={onSubmit} disabled={value.length === 0}>
          确认移除
        </Button>
      </DialogActions>
    </Dialog>
  );
}

BatchRemoveTagDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  selectedCount: PropTypes.number.isRequired,
  value: PropTypes.array.isRequired,
  setValue: PropTypes.func.isRequired,
  tagOptions: PropTypes.array.isRequired,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired
};
