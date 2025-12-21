import PropTypes from 'prop-types';

// material-ui
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import FormControlLabel from '@mui/material/FormControlLabel';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Autocomplete from '@mui/material/Autocomplete';
import Typography from '@mui/material/Typography';

// project imports
import SearchableNodeSelect from 'components/SearchableNodeSelect';
import CronExpressionGenerator from 'components/CronExpressionGenerator';

// constants
import { USER_AGENT_OPTIONS } from '../utils';

/**
 * 添加/编辑机场表单对话框
 */
export default function AirportFormDialog({
  open,
  isEdit,
  airportForm,
  setAirportForm,
  groupOptions,
  proxyNodeOptions,
  loadingProxyNodes,
  onClose,
  onSubmit,
  onFetchProxyNodes
}) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{isEdit ? '编辑机场' : '添加机场'}</DialogTitle>
      <DialogContent>
        <Stack spacing={2}>
          <TextField
            fullWidth
            label="名称"
            value={airportForm.name}
            helperText="机场名称不能重复，名称将作为节点来源"
            onChange={(e) => setAirportForm({ ...airportForm, name: e.target.value })}
          />
          <TextField
            fullWidth
            label="订阅地址"
            value={airportForm.url}
            helperText="支持Clash协议的YAML订阅和V2Ray的Base64订阅"
            onChange={(e) => setAirportForm({ ...airportForm, url: e.target.value })}
          />
          <CronExpressionGenerator
            value={airportForm.cronExpr}
            onChange={(value) => setAirportForm({ ...airportForm, cronExpr: value })}
            label="定时更新设置"
          />
          <Autocomplete
            freeSolo
            options={groupOptions}
            value={airportForm.group}
            onChange={(e, newValue) => setAirportForm({ ...airportForm, group: newValue || '' })}
            onInputChange={(e, newValue) => setAirportForm({ ...airportForm, group: newValue || '' })}
            renderInput={(params) => <TextField {...params} label="分组" helperText="从此机场导入的所有节点将自动归属到此分组" />}
          />
          <Autocomplete
            freeSolo
            options={USER_AGENT_OPTIONS}
            getOptionLabel={(option) => (typeof option === 'string' ? option : option.value)}
            value={airportForm.userAgent}
            onChange={(e, newValue) => {
              const value = typeof newValue === 'string' ? newValue : (newValue?.value ?? '');
              setAirportForm({ ...airportForm, userAgent: value });
            }}
            onInputChange={(e, newValue) => setAirportForm({ ...airportForm, userAgent: newValue ?? '' })}
            renderOption={(props, option) => (
              <Box component="li" {...props} key={option.value}>
                <Box>
                  <Typography variant="body2">{option.label}</Typography>
                  <Typography variant="caption" color="textSecondary">
                    {option.value}
                  </Typography>
                </Box>
              </Box>
            )}
            renderInput={(params) => (
              <TextField
                {...params}
                label="User-Agent"
                placeholder="选择或输入 User-Agent"
                helperText="拉取订阅时使用的 User-Agent，可留空"
              />
            )}
          />
          <FormControlLabel
            control={
              <Switch checked={airportForm.enabled} onChange={(e) => setAirportForm({ ...airportForm, enabled: e.target.checked })} />
            }
            label="启用"
          />
          <FormControlLabel
            control={
              <Switch
                checked={airportForm.downloadWithProxy}
                onChange={(e) => {
                  const checked = e.target.checked;
                  setAirportForm({ ...airportForm, downloadWithProxy: checked });
                  if (checked) {
                    onFetchProxyNodes();
                  }
                }}
              />
            }
            label="使用代理下载"
          />
          {airportForm.downloadWithProxy && (
            <Box>
              <SearchableNodeSelect
                nodes={proxyNodeOptions}
                loading={loadingProxyNodes}
                value={
                  proxyNodeOptions.find((n) => n.Link === airportForm.proxyLink) ||
                  (airportForm.proxyLink ? { Link: airportForm.proxyLink, Name: '', ID: 0 } : null)
                }
                onChange={(newValue) => setAirportForm({ ...airportForm, proxyLink: newValue?.Link || '' })}
                displayField="Name"
                valueField="Link"
                label="选择代理节点"
                placeholder="留空则自动选择最佳节点"
                helperText="如果未选择具体代理，系统将自动选择延迟最低且速度最快的节点作为下载代理"
                freeSolo={true}
                limit={50}
              />
            </Box>
          )}
          <FormControlLabel
            control={
              <Switch
                checked={airportForm.fetchUsageInfo || false}
                onChange={(e) => setAirportForm({ ...airportForm, fetchUsageInfo: e.target.checked })}
              />
            }
            label="获取用量信息"
          />
          {airportForm.fetchUsageInfo && (
            <Alert severity="info" sx={{ mt: -1 }}>
              开启后将从订阅响应 Header 解析用量信息。此功能需要机场支持，且 User-Agent 需设置为 Clash 相关。
            </Alert>
          )}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>取消</Button>
        <Button variant="contained" onClick={onSubmit}>
          确定
        </Button>
      </DialogActions>
    </Dialog>
  );
}

AirportFormDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  isEdit: PropTypes.bool.isRequired,
  airportForm: PropTypes.shape({
    id: PropTypes.number,
    name: PropTypes.string,
    url: PropTypes.string,
    cronExpr: PropTypes.string,
    enabled: PropTypes.bool,
    group: PropTypes.string,
    downloadWithProxy: PropTypes.bool,
    proxyLink: PropTypes.string,
    userAgent: PropTypes.string,
    fetchUsageInfo: PropTypes.bool
  }).isRequired,
  setAirportForm: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  proxyNodeOptions: PropTypes.array.isRequired,
  loadingProxyNodes: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onFetchProxyNodes: PropTypes.func.isRequired
};
