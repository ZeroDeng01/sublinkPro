import PropTypes from 'prop-types';

// material-ui
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Typography from '@mui/material/Typography';

// project imports
import SearchableNodeSelect from 'components/SearchableNodeSelect';

/**
 * 批量修改前置代理对话框
 */
export default function BatchDialerProxyDialog({
  open,
  selectedCount,
  value,
  setValue,
  proxyNodeOptions,
  loadingProxyNodes,
  onClose,
  onSubmit
}) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>批量修改前置代理</DialogTitle>
      <DialogContent>
        <Typography variant="body2" color="textSecondary" sx={{ mb: 2 }}>
          将为选中的 {selectedCount} 个节点设置相同的前置代理
        </Typography>
        <SearchableNodeSelect
          nodes={proxyNodeOptions}
          loading={loadingProxyNodes}
          value={value}
          onChange={(newValue) => {
            const name = typeof newValue === 'string' ? newValue : newValue?.Name || '';
            setValue(name);
          }}
          displayField="Name"
          valueField="Name"
          label="前置代理节点"
          placeholder="选择或输入代理节点名称/策略组名称，留空则清空前置代理"
          helperText="提示：前置代理节点用于链式代理，流量将先经过此节点再转发。留空将清除前置代理设置。"
          freeSolo={true}
          limit={50}
        />
        <Alert severity="warning" sx={{ mt: 1 }}>
          前置代理仅 Clash-Meta 内核可用
        </Alert>
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

BatchDialerProxyDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  selectedCount: PropTypes.number.isRequired,
  value: PropTypes.string.isRequired,
  setValue: PropTypes.func.isRequired,
  proxyNodeOptions: PropTypes.array.isRequired,
  loadingProxyNodes: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired
};
