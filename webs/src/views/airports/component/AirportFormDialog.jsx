import PropTypes from 'prop-types';

// material-ui
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Collapse from '@mui/material/Collapse';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Divider from '@mui/material/Divider';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Autocomplete from '@mui/material/Autocomplete';
import Typography from '@mui/material/Typography';

// project imports
import SearchableNodeSelect from 'components/SearchableNodeSelect';
import CronExpressionGenerator from 'components/CronExpressionGenerator';
import LogoPicker from 'components/LogoPicker';

// constants
import { USER_AGENT_OPTIONS } from '../utils';

/**
 * 分组标题组件
 */
function SectionTitle({ children }) {
  return (
    <Typography
      variant="subtitle2"
      color="primary"
      sx={{
        fontWeight: 600,
        mb: 1.5,
        display: 'flex',
        alignItems: 'center',
        '&::before': {
          content: '""',
          width: 3,
          height: 16,
          bgcolor: 'primary.main',
          borderRadius: 1,
          mr: 1
        }
      }}
    >
      {children}
    </Typography>
  );
}

SectionTitle.propTypes = {
  children: PropTypes.node.isRequired
};

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
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{
        sx: {
          maxHeight: '90vh'
        }
      }}
    >
      <DialogTitle>{isEdit ? '编辑机场' : '添加机场'}</DialogTitle>
      <DialogContent dividers sx={{ pt: 2.5, pb: 2 }}>
        <Stack spacing={2.5}>
          {/* ===== 基本信息 ===== */}
          <Box>
            <SectionTitle>基本信息</SectionTitle>
            <Stack spacing={2}>
              <TextField
                fullWidth
                size="small"
                label="名称"
                value={airportForm.name}
                helperText="机场名称不能重复，名称将作为节点来源"
                onChange={(e) => setAirportForm({ ...airportForm, name: e.target.value })}
              />
              <Box>
                <Typography variant="body2" color="textSecondary" sx={{ mb: 1 }}>
                  Logo（可选）
                </Typography>
                <LogoPicker
                  value={airportForm.logo || ''}
                  onChange={(logo) => setAirportForm({ ...airportForm, logo })}
                  name={airportForm.name}
                />
              </Box>
              <TextField
                fullWidth
                size="small"
                label="订阅地址"
                value={airportForm.url}
                helperText="支持 Clash YAML 订阅和 V2Ray Base64 订阅"
                onChange={(e) => setAirportForm({ ...airportForm, url: e.target.value })}
              />
              <Autocomplete
                freeSolo
                size="small"
                options={groupOptions}
                value={airportForm.group}
                onChange={(e, newValue) => setAirportForm({ ...airportForm, group: newValue || '' })}
                onInputChange={(e, newValue) => setAirportForm({ ...airportForm, group: newValue || '' })}
                renderInput={(params) => <TextField {...params} label="节点分组" helperText="从此机场导入的节点将自动归属到此分组" />}
              />
              <TextField
                fullWidth
                size="small"
                label="备注"
                value={airportForm.remark}
                placeholder="可选，记录机场的备忘信息"
                helperText="一些备注信息，方便你对机场和订阅进行管理"
                multiline
                minRows={2}
                maxRows={4}
                onChange={(e) => setAirportForm({ ...airportForm, remark: e.target.value })}
              />
            </Stack>
          </Box>

          <Divider />

          {/* ===== 定时更新 ===== */}
          <Box>
            <SectionTitle>定时更新</SectionTitle>
            <Stack spacing={2}>
              {/* 启用定时更新开关 */}
              <Box
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between'
                }}
              >
                <Box>
                  <Typography variant="body2">启用定时更新</Typography>
                  <Typography variant="caption" color="textSecondary">
                    关闭后将停止自动拉取订阅
                  </Typography>
                </Box>
                <Switch checked={airportForm.enabled} onChange={(e) => setAirportForm({ ...airportForm, enabled: e.target.checked })} />
              </Box>
              <Collapse in={airportForm.enabled}>
                <CronExpressionGenerator
                  value={airportForm.cronExpr}
                  onChange={(value) => setAirportForm({ ...airportForm, cronExpr: value })}
                  label=""
                />
              </Collapse>
            </Stack>
          </Box>

          <Divider />

          {/* ===== 请求设置 ===== */}
          <Box>
            <SectionTitle>请求设置</SectionTitle>
            <Stack spacing={2}>
              <Autocomplete
                freeSolo
                size="small"
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
                  <TextField {...params} label="User-Agent" placeholder="选择或输入" helperText="拉取订阅时使用的 User-Agent，可留空" />
                )}
              />

              {/* 使用代理下载 */}
              <Box>
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    mb: airportForm.downloadWithProxy ? 1.5 : 0
                  }}
                >
                  <Box>
                    <Typography variant="body2">使用代理下载</Typography>
                    <Typography variant="caption" color="textSecondary">
                      通过代理节点拉取订阅
                    </Typography>
                  </Box>
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
                </Box>
                <Collapse in={airportForm.downloadWithProxy}>
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
                    label="代理节点"
                    placeholder="留空自动选择最佳节点"
                    helperText="系统将自动选择延迟最低且速度最快的节点"
                    freeSolo={true}
                    limit={50}
                    size="small"
                  />
                </Collapse>
              </Box>
            </Stack>
          </Box>

          <Divider />

          {/* ===== 高级选项 ===== */}
          <Box>
            <SectionTitle>高级选项</SectionTitle>
            <Stack spacing={1}>
              {/* 获取用量信息 */}
              <Box>
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    py: 0.5
                  }}
                >
                  <Box>
                    <Typography variant="body2">获取用量信息</Typography>
                    <Typography variant="caption" color="textSecondary">
                      从订阅响应解析流量使用情况
                    </Typography>
                  </Box>
                  <Switch
                    checked={airportForm.fetchUsageInfo || false}
                    onChange={(e) => setAirportForm({ ...airportForm, fetchUsageInfo: e.target.checked })}
                  />
                </Box>
                <Collapse in={airportForm.fetchUsageInfo}>
                  <Alert severity="info" sx={{ mt: 1 }} icon={false}>
                    <Typography variant="caption">需要机场支持，且 User-Agent 需设置为 Clash 相关</Typography>
                  </Alert>
                </Collapse>
              </Box>

              <Divider sx={{ my: 0.5 }} />

              {/* 忽略证书验证 */}
              <Box>
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    py: 0.5
                  }}
                >
                  <Box>
                    <Typography variant="body2">忽略证书验证</Typography>
                    <Typography variant="caption" color="textSecondary">
                      跳过 TLS 证书检查
                    </Typography>
                  </Box>
                  <Switch
                    checked={airportForm.skipTLSVerify || false}
                    onChange={(e) => setAirportForm({ ...airportForm, skipTLSVerify: e.target.checked })}
                  />
                </Box>
                <Collapse in={airportForm.skipTLSVerify}>
                  <Alert severity="warning" sx={{ mt: 1 }} icon={false}>
                    <Typography variant="caption">会降低安全性，仅在信任订阅源且证书有问题时启用</Typography>
                  </Alert>
                </Collapse>
              </Box>
            </Stack>
          </Box>
        </Stack>
      </DialogContent>
      <DialogActions sx={{ px: 3, py: 2 }}>
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
    fetchUsageInfo: PropTypes.bool,
    skipTLSVerify: PropTypes.bool,
    remark: PropTypes.string,
    logo: PropTypes.string
  }).isRequired,
  setAirportForm: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  proxyNodeOptions: PropTypes.array.isRequired,
  loadingProxyNodes: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onFetchProxyNodes: PropTypes.func.isRequired
};
