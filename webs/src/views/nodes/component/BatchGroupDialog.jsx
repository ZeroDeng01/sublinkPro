import PropTypes from 'prop-types';

// material-ui
import { useTheme } from '@mui/material/styles';
import Autocomplete from '@mui/material/Autocomplete';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getNodeDialogPaperSx, getNodeFieldControlSx, getNodeThemeTokens } from '../nodeTheme';

/**
 * 批量修改分组对话框
 */
export default function BatchGroupDialog({ open, selectedCount, value, setValue, groupOptions, onClose, onSubmit }) {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);
  const fieldControlSx = getNodeFieldControlSx(tokens);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{ sx: getNodeDialogPaperSx(theme, tokens, tokens.palette.warning.main) }}
    >
      <DialogTitle
        sx={{ color: tokens.primaryText, bgcolor: tokens.mutedPanelSurface, borderBottom: '1px solid', borderColor: tokens.panelBorder }}
      >
        批量修改分组
      </DialogTitle>
      <DialogContent dividers sx={{ bgcolor: 'transparent', borderColor: tokens.panelBorder }}>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          将为选中的 {selectedCount} 个节点设置相同的分组
        </Typography>
        <Autocomplete
          freeSolo
          options={groupOptions}
          value={value}
          onChange={(e, newValue) => setValue(newValue || '')}
          onInputChange={(e, newInputValue) => setValue(newInputValue)}
          sx={fieldControlSx}
          renderInput={(params) => (
            <TextField {...params} label="分组名称" placeholder="输入或选择分组名称，留空则清空分组" fullWidth sx={fieldControlSx} />
          )}
        />
        <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
          提示：留空将清除所选节点的分组设置
        </Typography>
      </DialogContent>
      <DialogActions sx={{ bgcolor: tokens.mutedPanelSurface, borderTop: '1px solid', borderColor: tokens.panelBorder }}>
        <Button onClick={onClose}>取消</Button>
        <Button variant="contained" onClick={onSubmit}>
          确认修改
        </Button>
      </DialogActions>
    </Dialog>
  );
}

BatchGroupDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  selectedCount: PropTypes.number.isRequired,
  value: PropTypes.string.isRequired,
  setValue: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired
};
