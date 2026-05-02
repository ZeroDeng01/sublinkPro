import PropTypes from 'prop-types';

import { useTheme } from '@mui/material/styles';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Collapse from '@mui/material/Collapse';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Divider from '@mui/material/Divider';
import IconButton from '@mui/material/IconButton';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Autocomplete from '@mui/material/Autocomplete';
import Typography from '@mui/material/Typography';

import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';

import SearchableNodeSelect from 'components/SearchableNodeSelect';
import CronExpressionGenerator from 'components/CronExpressionGenerator';
import LogoPicker from 'components/LogoPicker';
import NodeNameFilter from 'components/NodeNameFilter';
import NodeNamePreprocessor from 'components/NodeNamePreprocessor';
import NodeProtocolFilter from 'components/NodeProtocolFilter';
import NodeNameUniquifyConfig from 'components/NodeNameUniquifyConfig';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';
import AirportDeduplicationConfig from './AirportDeduplicationConfig';
import AirportDialogSection from './AirportDialogSection';

import { USER_AGENT_OPTIONS } from '../utils';

const createEmptyRequestHeader = () => ({ key: '', value: '' });

const getRequestHeaderRowError = (requestHeader) => {
  const key = `${requestHeader?.key ?? ''}`.trim();
  const value = `${requestHeader?.value ?? ''}`.trim();

  if (!key && !value) {
    return '';
  }

  if (!key && value) {
    return '请输入请求头键名';
  }

  if (key.toLowerCase() === 'user-agent') {
    return 'User-Agent 请使用上方专用字段设置';
  }

  return '';
};

export default function AirportFormDialog({
  open,
  isEdit,
  airportForm,
  setAirportForm,
  groupOptions,
  proxyNodeOptions,
  loadingProxyNodes,
  protocolOptions,
  onClose,
  onSubmit,
  onFetchProxyNodes
}) {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText } = getReadableTextTokens(theme, isDark);

  const controlRowSx = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 2,
    px: 1.5,
    py: 1.25,
    borderRadius: 2,
    bgcolor: isDark ? withAlpha(palette.background.paper, 0.2) : withAlpha(palette.background.paper, 0.92),
    border: '1px solid',
    borderColor: panelBorder
  };

  const requestHeaders = Array.isArray(airportForm.requestHeaders) ? airportForm.requestHeaders : [];
  const hasRequestHeaderRows = requestHeaders.length > 0;

  const requestHeaderPanelSx = {
    p: 1.5,
    borderRadius: 2,
    bgcolor: isDark ? withAlpha(palette.background.default, 0.88) : withAlpha(palette.background.default, 0.56),
    border: '1px solid',
    borderColor: panelBorder
  };

  const requestHeaderRowSx = {
    p: 1.25,
    borderRadius: 2,
    bgcolor: isDark ? withAlpha(palette.background.paper, 0.22) : withAlpha(palette.background.paper, 0.96),
    border: '1px solid',
    borderColor: panelBorder
  };

  const addRequestHeaderButtonSx = {
    minWidth: 0,
    alignSelf: { xs: 'flex-start', sm: 'center' },
    px: 1,
    py: 0.5,
    borderRadius: 1.5,
    fontSize: '0.8125rem',
    fontWeight: 600,
    lineHeight: 1.2,
    color: isDark ? palette.primary.light : palette.primary.main,
    bgcolor: withAlpha(palette.primary.main, isDark ? 0.08 : 0.04),
    borderColor: withAlpha(palette.primary.main, isDark ? 0.24 : 0.16),
    whiteSpace: 'nowrap',
    flexShrink: 0,
    '& .MuiButton-startIcon': {
      mr: 0.5,
      ml: 0,
      '& > *:nth-of-type(1)': {
        fontSize: '1rem'
      }
    },
    '&:hover': {
      borderColor: withAlpha(palette.primary.main, isDark ? 0.34 : 0.22),
      bgcolor: withAlpha(palette.primary.main, isDark ? 0.14 : 0.08)
    }
  };

  const updateRequestHeaders = (updater) => {
    const nextRequestHeaders = typeof updater === 'function' ? updater(requestHeaders) : updater;
    setAirportForm({ ...airportForm, requestHeaders: nextRequestHeaders });
  };

  const handleRequestHeaderChange = (index, field, value) => {
    updateRequestHeaders((currentHeaders) =>
      currentHeaders.map((header, headerIndex) => (headerIndex === index ? { ...header, [field]: value } : header))
    );
  };

  const handleAddRequestHeader = () => {
    updateRequestHeaders((currentHeaders) => [...currentHeaders, createEmptyRequestHeader()]);
  };

  const handleRemoveRequestHeader = (index) => {
    updateRequestHeaders((currentHeaders) => currentHeaders.filter((_, headerIndex) => headerIndex !== index));
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{
        sx: {
          maxHeight: '90vh',
          borderRadius: 2.5,
          border: '1px solid',
          borderColor: panelBorder,
          bgcolor: dialogSurface,
          backgroundImage: dialogSurfaceGradient
        }
      }}
    >
      <DialogTitle
        sx={{
          pb: 1.5,
          color: primaryText,
          bgcolor: mutedPanelSurface,
          borderBottom: '1px solid',
          borderColor: panelBorder
        }}
      >
        {isEdit ? '编辑机场' : '添加机场'}
      </DialogTitle>
      <DialogContent dividers sx={{ pt: 2.5, pb: 2, bgcolor: 'transparent', borderColor: panelBorder }}>
        <Stack spacing={2.5}>
          <AirportDialogSection title="基本信息" surface={nestedPanelSurface} borderColor={panelBorder} titleColor={primaryText}>
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
                <Typography variant="body2" sx={{ mb: 1, color: secondaryText }}>
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
          </AirportDialogSection>

          <AirportDialogSection title="定时更新" surface={nestedPanelSurface} borderColor={panelBorder} titleColor={primaryText}>
            <Stack spacing={2}>
              <Box sx={controlRowSx}>
                <Box>
                  <Typography variant="body2" sx={{ color: primaryText }}>
                    启用定时更新
                  </Typography>
                  <Typography variant="caption" sx={{ color: secondaryText }}>
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
          </AirportDialogSection>

          <AirportDialogSection title="请求设置" surface={nestedPanelSurface} borderColor={panelBorder} titleColor={primaryText}>
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
                      <Typography variant="body2" sx={{ color: primaryText }}>
                        {option.label}
                      </Typography>
                      <Typography variant="caption" sx={{ color: secondaryText }}>
                        {option.value}
                      </Typography>
                    </Box>
                  </Box>
                )}
                renderInput={(params) => (
                  <TextField {...params} label="User-Agent" placeholder="选择或输入" helperText="拉取订阅时使用的 User-Agent，可留空" />
                )}
              />

              <Box sx={requestHeaderPanelSx}>
                <Stack spacing={1.5}>
                  <Stack
                    direction={{ xs: 'column', sm: 'row' }}
                    spacing={1}
                    justifyContent="space-between"
                    alignItems={{ xs: 'flex-start', sm: 'flex-start' }}
                  >
                    <Box sx={{ flex: 1, minWidth: 0 }}>
                      <Typography variant="body2" sx={{ color: primaryText, fontWeight: 500 }}>
                        自定义请求头
                      </Typography>
                      <Typography variant="caption" sx={{ color: secondaryText }}>
                        可按行添加额外请求头，空白行不会提交，User-Agent 请使用上方专用字段
                      </Typography>
                    </Box>
                    <Button
                      variant="outlined"
                      size="small"
                      startIcon={<AddIcon />}
                      onClick={handleAddRequestHeader}
                      sx={addRequestHeaderButtonSx}
                    >
                      添加请求头
                    </Button>
                  </Stack>

                  {hasRequestHeaderRows ? (
                    <Stack spacing={1}>
                      {requestHeaders.map((header, index) => {
                        const rowError = getRequestHeaderRowError(header);

                        return (
                          <Box key={`request-header-${index}`} sx={requestHeaderRowSx}>
                            <Stack spacing={1}>
                              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} alignItems={{ xs: 'stretch', sm: 'flex-start' }}>
                                <TextField
                                  fullWidth
                                  size="small"
                                  label="Header Key"
                                  placeholder="例如：X-Custom-Token"
                                  value={header.key}
                                  error={Boolean(rowError)}
                                  onChange={(e) => handleRequestHeaderChange(index, 'key', e.target.value)}
                                />
                                <TextField
                                  fullWidth
                                  size="small"
                                  label="Header Value"
                                  placeholder="例如：your-value"
                                  value={header.value}
                                  onChange={(e) => handleRequestHeaderChange(index, 'value', e.target.value)}
                                />
                                <IconButton
                                  aria-label={`删除第 ${index + 1} 行请求头`}
                                  color="error"
                                  onClick={() => handleRemoveRequestHeader(index)}
                                  sx={{
                                    alignSelf: { xs: 'flex-end', sm: 'center' },
                                    border: '1px solid',
                                    borderColor: withAlpha(palette.error.main, isDark ? 0.32 : 0.22),
                                    bgcolor: withAlpha(palette.error.main, isDark ? 0.12 : 0.04),
                                    '&:hover': {
                                      bgcolor: withAlpha(palette.error.main, isDark ? 0.18 : 0.08)
                                    }
                                  }}
                                >
                                  <DeleteIcon fontSize="small" />
                                </IconButton>
                              </Stack>
                              {rowError ? (
                                <Typography variant="caption" color="error">
                                  {rowError}
                                </Typography>
                              ) : (
                                <Typography variant="caption" sx={{ color: secondaryText }}>
                                  留空整行会在保存时自动忽略
                                </Typography>
                              )}
                            </Stack>
                          </Box>
                        );
                      })}
                    </Stack>
                  ) : (
                    <Typography variant="caption" sx={{ color: secondaryText }}>
                      当前未设置额外请求头，可按需添加。
                    </Typography>
                  )}
                </Stack>
              </Box>

              <Box>
                <Box sx={{ ...controlRowSx, mb: airportForm.downloadWithProxy ? 1.5 : 0 }}>
                  <Box>
                    <Typography variant="body2" sx={{ color: primaryText }}>
                      使用代理下载
                    </Typography>
                    <Typography variant="caption" sx={{ color: secondaryText }}>
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
                    onChange={(newValue) =>
                      setAirportForm({ ...airportForm, proxyLink: typeof newValue === 'string' ? newValue : newValue?.Link || '' })
                    }
                    displayField="Name"
                    valueField="Link"
                    label="代理节点"
                    placeholder="留空则自动选择最佳节点"
                    helperText="可选择任意现有节点，也可手动输入外部代理链接；留空时系统会自动选择最佳节点。"
                    freeSolo={true}
                    limit={50}
                    size="small"
                  />
                </Collapse>
              </Box>
            </Stack>
          </AirportDialogSection>

          <AirportDialogSection title="高级选项" surface={nestedPanelSurface} borderColor={panelBorder} titleColor={primaryText}>
            <Stack spacing={1}>
              <Box>
                <Box sx={controlRowSx}>
                  <Box>
                    <Typography variant="body2" sx={{ color: primaryText }}>
                      获取用量信息
                    </Typography>
                    <Typography variant="caption" sx={{ color: secondaryText }}>
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

              <Divider sx={{ my: 0.5, borderColor: panelBorder }} />

              <Box>
                <Box sx={controlRowSx}>
                  <Box>
                    <Typography variant="body2" sx={{ color: primaryText }}>
                      忽略证书验证
                    </Typography>
                    <Typography variant="caption" sx={{ color: secondaryText }}>
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
          </AirportDialogSection>

          <AirportDialogSection
            title="节点处理（拉取时生效）"
            surface={nestedPanelSurface}
            borderColor={panelBorder}
            titleColor={primaryText}
          >
            <Stack spacing={2}>
              <Alert severity="info" icon={false}>
                <Typography variant="caption">以下规则在拉取订阅时立即生效，过滤的节点不会存储到数据库</Typography>
              </Alert>
              <NodeNameFilter
                whitelistValue={airportForm.nodeNameWhitelist || ''}
                blacklistValue={airportForm.nodeNameBlacklist || ''}
                onWhitelistChange={(rules) => setAirportForm({ ...airportForm, nodeNameWhitelist: rules })}
                onBlacklistChange={(rules) => setAirportForm({ ...airportForm, nodeNameBlacklist: rules })}
              />
              <NodeProtocolFilter
                protocolOptions={protocolOptions}
                whitelistValue={airportForm.protocolWhitelist || ''}
                blacklistValue={airportForm.protocolBlacklist || ''}
                onWhitelistChange={(protocols) => setAirportForm({ ...airportForm, protocolWhitelist: protocols })}
                onBlacklistChange={(protocols) => setAirportForm({ ...airportForm, protocolBlacklist: protocols })}
              />
              <AirportDeduplicationConfig
                value={airportForm.deduplicationRule || ''}
                onChange={(rule) => setAirportForm({ ...airportForm, deduplicationRule: rule })}
              />
              <NodeNamePreprocessor
                value={airportForm.nodeNamePreprocess || ''}
                onChange={(rules) => setAirportForm({ ...airportForm, nodeNamePreprocess: rules })}
              />
              <NodeNameUniquifyConfig
                enabled={airportForm.nodeNameUniquify || false}
                prefix={airportForm.nodeNamePrefix || ''}
                airportId={airportForm.id || 0}
                onChange={({ enabled, prefix }) => setAirportForm({ ...airportForm, nodeNameUniquify: enabled, nodeNamePrefix: prefix })}
              />
            </Stack>
          </AirportDialogSection>
        </Stack>
      </DialogContent>
      <DialogActions
        sx={{
          px: 3,
          py: 2,
          bgcolor: mutedPanelSurface,
          borderTop: '1px solid',
          borderColor: panelBorder
        }}
      >
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
    requestHeaders: PropTypes.arrayOf(
      PropTypes.shape({
        key: PropTypes.string,
        value: PropTypes.string
      })
    ),
    fetchUsageInfo: PropTypes.bool,
    skipTLSVerify: PropTypes.bool,
    remark: PropTypes.string,
    logo: PropTypes.string,
    nodeNameWhitelist: PropTypes.string,
    nodeNameBlacklist: PropTypes.string,
    protocolWhitelist: PropTypes.string,
    protocolBlacklist: PropTypes.string,
    nodeNamePreprocess: PropTypes.string,
    deduplicationRule: PropTypes.string,
    nodeNameUniquify: PropTypes.bool,
    nodeNamePrefix: PropTypes.string
  }).isRequired,
  setAirportForm: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  proxyNodeOptions: PropTypes.array.isRequired,
  loadingProxyNodes: PropTypes.bool.isRequired,
  protocolOptions: PropTypes.array,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onFetchProxyNodes: PropTypes.func.isRequired
};
