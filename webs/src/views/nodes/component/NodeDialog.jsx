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
  tagOptions,
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
          {/* 标签选择 */}
          <Autocomplete
            multiple
            options={tagOptions || []}
            value={nodeForm.tags || []}
            onChange={(e, newValue) => setNodeForm({ ...nodeForm, tags: newValue })}
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
            renderInput={(params) => <TextField {...params} label="标签" placeholder="选择要设置的标签" />}
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
    mergeMode: PropTypes.string,
    tags: PropTypes.array
  }).isRequired,
  setNodeForm: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  proxyNodeOptions: PropTypes.array.isRequired,
  loadingProxyNodes: PropTypes.bool.isRequired,
  tagOptions: PropTypes.array,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onFetchProxyNodes: PropTypes.func.isRequired
};
