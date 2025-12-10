import PropTypes from 'prop-types';

// material-ui
import Autocomplete from '@mui/material/Autocomplete';
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
import Typography from '@mui/material/Typography';

// project imports
import SearchableNodeSelect from 'components/SearchableNodeSelect';

// constants
import { CRON_OPTIONS, USER_AGENT_OPTIONS } from '../utils';

/**
 * 添加/编辑订阅表单对话框
 */
export default function SchedulerFormDialog({
  open,
  isEdit,
  schedulerForm,
  setSchedulerForm,
  groupOptions,
  proxyNodeOptions,
  loadingProxyNodes,
  onClose,
  onSubmit,
  onFetchProxyNodes
}) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{isEdit ? '编辑订阅' : '添加订阅'}</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ mt: 1 }}>
          <TextField
            fullWidth
            label="名称"
            value={schedulerForm.Name}
            helperText="订阅名称不能重复，名称将作为节点来源"
            onChange={(e) => setSchedulerForm({ ...schedulerForm, Name: e.target.value })}
          />
          <TextField
            fullWidth
            label="URL"
            value={schedulerForm.URL}
            helperText="目前仅支持clash协议的yaml订阅和v2ray的base64以及非base64订阅"
            onChange={(e) => setSchedulerForm({ ...schedulerForm, URL: e.target.value })}
          />
          <Autocomplete
            freeSolo
            options={CRON_OPTIONS}
            getOptionLabel={(option) => (typeof option === 'string' ? option : option.value)}
            value={schedulerForm.CronExpr}
            onChange={(e, newValue) => {
              const value = typeof newValue === 'string' ? newValue : newValue?.value || '';
              setSchedulerForm({ ...schedulerForm, CronExpr: value });
            }}
            onInputChange={(e, newValue) => setSchedulerForm({ ...schedulerForm, CronExpr: newValue || '' })}
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
                label="Cron表达式"
                placeholder="分 时 日 月 周"
                helperText="格式: 分 时 日 月 周，如 0 */6 * * * 表示每6小时"
              />
            )}
          />
          <Autocomplete
            freeSolo
            options={groupOptions}
            value={schedulerForm.Group}
            onChange={(e, newValue) => setSchedulerForm({ ...schedulerForm, Group: newValue || '' })}
            onInputChange={(e, newValue) => setSchedulerForm({ ...schedulerForm, Group: newValue || '' })}
            renderInput={(params) => (
              <TextField {...params} label="分组" helperText="设置分组后，从此订阅导入的所有节点将自动归属到此分组" />
            )}
          />
          <Autocomplete
            freeSolo
            options={USER_AGENT_OPTIONS}
            getOptionLabel={(option) => (typeof option === 'string' ? option : option.value)}
            value={schedulerForm.UserAgent}
            onChange={(e, newValue) => {
              const value = typeof newValue === 'string' ? newValue : newValue?.value || 'Clash';
              setSchedulerForm({ ...schedulerForm, UserAgent: value });
            }}
            onInputChange={(e, newValue) => setSchedulerForm({ ...schedulerForm, UserAgent: newValue || 'Clash' })}
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
                helperText="拉取订阅时使用的 User-Agent，默认为 Clash"
              />
            )}
          />
          <FormControlLabel
            control={
              <Switch checked={schedulerForm.Enabled} onChange={(e) => setSchedulerForm({ ...schedulerForm, Enabled: e.target.checked })} />
            }
            label="启用"
          />
          <FormControlLabel
            control={
              <Switch
                checked={schedulerForm.DownloadWithProxy}
                onChange={(e) => {
                  const checked = e.target.checked;
                  setSchedulerForm({ ...schedulerForm, DownloadWithProxy: checked });
                  if (checked) {
                    onFetchProxyNodes();
                  }
                }}
              />
            }
            label="使用代理下载"
          />
          {schedulerForm.DownloadWithProxy && (
            <Box>
              <SearchableNodeSelect
                nodes={proxyNodeOptions}
                loading={loadingProxyNodes}
                value={
                  proxyNodeOptions.find((n) => n.Link === schedulerForm.ProxyLink) ||
                  (schedulerForm.ProxyLink ? { Link: schedulerForm.ProxyLink, Name: '', ID: 0 } : null)
                }
                onChange={(newValue) => setSchedulerForm({ ...schedulerForm, ProxyLink: newValue?.Link || '' })}
                displayField="Name"
                valueField="Link"
                label="选择代理节点"
                placeholder="留空则自动选择最佳节点"
                helperText="如果未选择具体代理，系统将自动选择延迟最低且速度最快的节点作为下载代理，你也可以输入外部代理地址"
                freeSolo={true}
                limit={50}
              />
            </Box>
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

SchedulerFormDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  isEdit: PropTypes.bool.isRequired,
  schedulerForm: PropTypes.shape({
    Name: PropTypes.string,
    URL: PropTypes.string,
    CronExpr: PropTypes.string,
    Enabled: PropTypes.bool,
    Group: PropTypes.string,
    DownloadWithProxy: PropTypes.bool,
    ProxyLink: PropTypes.string,
    UserAgent: PropTypes.string
  }).isRequired,
  setSchedulerForm: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  proxyNodeOptions: PropTypes.array.isRequired,
  loadingProxyNodes: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onFetchProxyNodes: PropTypes.func.isRequired
};
