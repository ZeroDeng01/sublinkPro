import PropTypes from 'prop-types';

// material-ui
import Autocomplete from '@mui/material/Autocomplete';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import FormControlLabel from '@mui/material/FormControlLabel';
import Radio from '@mui/material/Radio';
import RadioGroup from '@mui/material/RadioGroup';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';

// project imports
import SearchableNodeSelect from 'components/SearchableNodeSelect';

/**
 * 添加/编辑节点对话框
 */
export default function NodeDialog({
  open,
  isEdit,
  nodeForm,
  setNodeForm,
  groupOptions,
  proxyNodeOptions,
  loadingProxyNodes,
  onClose,
  onSubmit,
  onFetchProxyNodes
}) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>{isEdit ? '编辑节点' : '添加节点'}</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ mt: 1 }}>
          <TextField
            fullWidth
            multiline
            rows={4}
            label="节点链接"
            value={nodeForm.link}
            onChange={(e) => setNodeForm({ ...nodeForm, link: e.target.value })}
            placeholder="请输入节点，多行使用回车或逗号分开，支持base64格式的url订阅"
          />
          {!isEdit && (
            <RadioGroup row value={nodeForm.mergeMode} onChange={(e) => setNodeForm({ ...nodeForm, mergeMode: e.target.value })}>
              <FormControlLabel value="1" control={<Radio />} label="合并" />
              <FormControlLabel value="2" control={<Radio />} label="分开" />
            </RadioGroup>
          )}
          {(isEdit || nodeForm.mergeMode === '1') && (
            <TextField fullWidth label="备注" value={nodeForm.name} onChange={(e) => setNodeForm({ ...nodeForm, name: e.target.value })} />
          )}
          <SearchableNodeSelect
            nodes={proxyNodeOptions}
            loading={loadingProxyNodes}
            value={nodeForm.dialerProxyName}
            onChange={(newValue) => {
              const name = typeof newValue === 'string' ? newValue : newValue?.Name || '';
              setNodeForm({ ...nodeForm, dialerProxyName: name });
            }}
            displayField="Name"
            valueField="Name"
            label="前置代理节点名称或策略组名称"
            placeholder="选择或输入节点名称/策略组名称"
            helperText="仅Clash-Meta内核可用，留空则不使用前置代理"
            freeSolo={true}
            limit={50}
            onFocus={onFetchProxyNodes}
          />
          <Autocomplete
            freeSolo
            options={groupOptions}
            value={nodeForm.group}
            onChange={(e, newValue) => setNodeForm({ ...nodeForm, group: newValue || '' })}
            onInputChange={(e, newValue) => setNodeForm({ ...nodeForm, group: newValue || '' })}
            renderInput={(params) => <TextField {...params} label="分组" placeholder="请选择或输入分组名称" />}
          />
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>关闭</Button>
        <Button variant="contained" onClick={onSubmit}>
          确定
        </Button>
      </DialogActions>
    </Dialog>
  );
}

NodeDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  isEdit: PropTypes.bool.isRequired,
  nodeForm: PropTypes.shape({
    name: PropTypes.string,
    link: PropTypes.string,
    dialerProxyName: PropTypes.string,
    group: PropTypes.string,
    mergeMode: PropTypes.string
  }).isRequired,
  setNodeForm: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  proxyNodeOptions: PropTypes.array.isRequired,
  loadingProxyNodes: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onFetchProxyNodes: PropTypes.func.isRequired
};
