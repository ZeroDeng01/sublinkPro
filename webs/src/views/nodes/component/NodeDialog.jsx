import PropTypes from 'prop-types';

// material-ui
import { useTheme } from '@mui/material/styles';
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

// project imports
import SearchableNodeSelect from 'components/SearchableNodeSelect';
import { getNodeColorChipSx, getNodeDialogPaperSx, getNodeFieldControlSx, getNodeThemeTokens } from '../nodeTheme';

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
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);
  const fieldControlSx = getNodeFieldControlSx(tokens);

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth PaperProps={{ sx: getNodeDialogPaperSx(theme, tokens) }}>
      <DialogTitle
        sx={{
          color: tokens.primaryText,
          bgcolor: tokens.mutedPanelSurface,
          borderBottom: '1px solid',
          borderColor: tokens.panelBorder
        }}
      >
        {isEdit ? '编辑节点' : '添加节点'}
      </DialogTitle>
      <DialogContent dividers sx={{ bgcolor: 'transparent', borderColor: tokens.panelBorder }}>
        <Stack spacing={2} sx={{ mt: 1 }}>
          <TextField
            fullWidth
            multiline
            rows={4}
            label="节点链接"
            value={nodeForm.link}
            onChange={(e) => setNodeForm({ ...nodeForm, link: e.target.value })}
            placeholder="支持输入：各类代理链接（vmess://、vless://、wireguard:// 等）、WireGuard 标准配置文件（包含 [Interface] 和 [Peer]）、Clash YAML 配置（包含 proxies 字段）、Base64 订阅链接。多行使用回车分隔"
            sx={fieldControlSx}
          />
          {isEdit && (
            <TextField
              fullWidth
              label="备注"
              value={nodeForm.name}
              onChange={(e) => setNodeForm({ ...nodeForm, name: e.target.value })}
              sx={fieldControlSx}
            />
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
            sx={fieldControlSx}
          />
          <Autocomplete
            freeSolo
            options={groupOptions}
            value={nodeForm.group}
            onChange={(e, newValue) => setNodeForm({ ...nodeForm, group: newValue || '' })}
            onInputChange={(e, newValue) => setNodeForm({ ...nodeForm, group: newValue || '' })}
            sx={fieldControlSx}
            renderInput={(params) => <TextField {...params} label="分组" placeholder="请选择或输入分组名称" sx={fieldControlSx} />}
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
                      backgroundColor: option.color || tokens.palette.primary.main,
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
                    sx={getNodeColorChipSx(theme, tokens, option.color || theme.palette.primary.main, { deletable: true })}
                    {...tagProps}
                  />
                );
              })
            }
            sx={fieldControlSx}
            renderInput={(params) => <TextField {...params} label="标签" placeholder="选择要设置的标签" sx={fieldControlSx} />}
          />
        </Stack>
      </DialogContent>
      <DialogActions sx={{ bgcolor: tokens.mutedPanelSurface, borderTop: '1px solid', borderColor: tokens.panelBorder }}>
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
