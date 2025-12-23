import PropTypes from 'prop-types';
import { useState, useEffect } from 'react';

// material-ui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Alert from '@mui/material/Alert';
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Collapse from '@mui/material/Collapse';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Divider from '@mui/material/Divider';
import FormControl from '@mui/material/FormControl';
import FormControlLabel from '@mui/material/FormControlLabel';
import InputAdornment from '@mui/material/InputAdornment';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Paper from '@mui/material/Paper';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Grid from '@mui/material/Grid';

// icons
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import SpeedIcon from '@mui/icons-material/Speed';
import TimerIcon from '@mui/icons-material/Timer';
import TuneIcon from '@mui/icons-material/Tune';
import DataUsageIcon from '@mui/icons-material/DataUsage';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';

// project imports
import CronExpressionGenerator from 'components/CronExpressionGenerator';

// api
import { createNodeCheckProfile, updateNodeCheckProfile } from 'api/nodeCheck';

// constants
import { SPEED_TEST_TCP_OPTIONS, SPEED_TEST_MIHOMO_OPTIONS, LATENCY_TEST_URL_OPTIONS, LANDING_IP_URL_OPTIONS } from '../utils';

/**
 * 可折叠配置区块
 */
function ConfigSection({ title, icon, children, defaultExpanded = true, helperText }) {
  const [expanded, setExpanded] = useState(defaultExpanded);
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  return (
    <Paper
      elevation={0}
      sx={{
        border: `1px solid ${isDark ? 'rgba(255,255,255,0.12)' : 'rgba(0,0,0,0.12)'}`,
        borderRadius: 2,
        overflow: 'hidden',
        mb: 2
      }}
    >
      <Box
        onClick={() => setExpanded(!expanded)}
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          p: 1.5,
          cursor: 'pointer',
          backgroundColor: isDark ? 'rgba(255,255,255,0.03)' : 'rgba(0,0,0,0.02)',
          '&:hover': {
            backgroundColor: isDark ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.04)'
          }
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          {icon}
          <Typography variant="subtitle2" fontWeight={600}>
            {title}
          </Typography>
        </Box>
        {expanded ? <ExpandLessIcon fontSize="small" /> : <ExpandMoreIcon fontSize="small" />}
      </Box>
      <Collapse in={expanded}>
        <Divider />
        <Box sx={{ p: 2 }}>
          {children}
          {helperText && (
            <Typography variant="caption" color="textSecondary" sx={{ mt: 1.5, display: 'block' }}>
              {helperText}
            </Typography>
          )}
        </Box>
      </Collapse>
    </Paper>
  );
}

ConfigSection.propTypes = {
  title: PropTypes.string.isRequired,
  icon: PropTypes.node,
  children: PropTypes.node.isRequired,
  defaultExpanded: PropTypes.bool,
  helperText: PropTypes.string
};

/**
 * 节点检测策略编辑表单对话框
 */
export default function NodeCheckProfileFormDialog({ open, onClose, profile, groupOptions = [], tagOptions = [], onSuccess }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const isEdit = !!profile;

  // 表单状态
  const [form, setForm] = useState({
    name: '',
    enabled: false,
    cronExpr: '',
    mode: 'tcp',
    testUrl: '',
    latencyUrl: '',
    timeout: 5,
    groups: [],
    tags: [],
    latencyConcurrency: 0,
    speedConcurrency: 0,
    detectCountry: false,
    landingIpUrl: '',
    includeHandshake: true,
    speedRecordMode: 'average',
    peakSampleInterval: 100,
    trafficByGroup: true,
    trafficBySource: true,
    trafficByNode: false
  });

  const [submitting, setSubmitting] = useState(false);

  // 初始化表单
  useEffect(() => {
    if (open) {
      if (profile) {
        // 解析 groups 和 tags 字符串为数组
        const groups = profile.groups ? profile.groups.split(',').filter((g) => g) : [];
        const tags = profile.tags ? profile.tags.split(',').filter((t) => t) : [];

        setForm({
          name: profile.name || '',
          enabled: profile.enabled || false,
          cronExpr: profile.cronExpr || '',
          mode: profile.mode || 'tcp',
          testUrl: profile.testUrl || '',
          latencyUrl: profile.latencyUrl || '',
          timeout: profile.timeout || 5,
          groups: groups,
          tags: tags,
          latencyConcurrency: profile.latencyConcurrency || 0,
          speedConcurrency: profile.speedConcurrency ?? 0,
          detectCountry: profile.detectCountry || false,
          landingIpUrl: profile.landingIpUrl || '',
          includeHandshake: profile.includeHandshake !== false,
          speedRecordMode: profile.speedRecordMode || 'average',
          peakSampleInterval: profile.peakSampleInterval || 100,
          trafficByGroup: profile.trafficByGroup !== false,
          trafficBySource: profile.trafficBySource !== false,
          trafficByNode: profile.trafficByNode || false
        });
      } else {
        // 新建时的默认值
        setForm({
          name: '',
          enabled: false,
          cronExpr: '',
          mode: 'tcp',
          testUrl: SPEED_TEST_TCP_OPTIONS[0]?.value || '',
          latencyUrl: '',
          timeout: 5,
          groups: [],
          tags: [],
          latencyConcurrency: 0,
          speedConcurrency: 0,
          detectCountry: false,
          landingIpUrl: '',
          includeHandshake: true,
          speedRecordMode: 'average',
          peakSampleInterval: 100,
          trafficByGroup: true,
          trafficBySource: true,
          trafficByNode: false
        });
      }
    }
  }, [open, profile]);

  // 更新表单字段
  const updateForm = (field, value) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  // 模式切换时更新默认URL
  const handleModeChange = (mode) => {
    const defaultUrl = mode === 'mihomo' ? SPEED_TEST_MIHOMO_OPTIONS[0]?.value : SPEED_TEST_TCP_OPTIONS[0]?.value;
    setForm((prev) => ({ ...prev, mode, testUrl: defaultUrl }));
  };

  // 提交表单
  const handleSubmit = async () => {
    if (!form.name.trim()) {
      return;
    }

    setSubmitting(true);
    try {
      const data = {
        name: form.name.trim(),
        enabled: form.enabled,
        cronExpr: form.cronExpr,
        mode: form.mode,
        testUrl: form.testUrl,
        latencyUrl: form.latencyUrl,
        timeout: form.timeout,
        groups: form.groups,
        tags: form.tags,
        latencyConcurrency: form.latencyConcurrency,
        speedConcurrency: form.speedConcurrency,
        detectCountry: form.detectCountry,
        landingIpUrl: form.landingIpUrl,
        includeHandshake: form.includeHandshake,
        speedRecordMode: form.speedRecordMode,
        peakSampleInterval: form.peakSampleInterval,
        trafficByGroup: form.trafficByGroup,
        trafficBySource: form.trafficBySource,
        trafficByNode: form.trafficByNode
      };

      if (isEdit) {
        await updateNodeCheckProfile(profile.id, data);
      } else {
        await createNodeCheckProfile(data);
      }

      onSuccess?.();
    } catch (error) {
      console.error('保存失败:', error);
    } finally {
      setSubmitting(false);
    }
  };

  const urlOptions = form.mode === 'mihomo' ? SPEED_TEST_MIHOMO_OPTIONS : SPEED_TEST_TCP_OPTIONS;

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth fullScreen={isMobile}>
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <SpeedIcon color="primary" />
          <span>{isEdit ? '编辑检测策略' : '新建检测策略'}</span>
        </Box>
      </DialogTitle>

      <DialogContent dividers>
        {/* 策略名称 - 增加上边距避免遮挡 */}
        <TextField
          fullWidth
          label="策略名称"
          value={form.name}
          onChange={(e) => updateForm('name', e.target.value)}
          placeholder="例如：每日全量检测"
          sx={{ mb: 2, mt: 1 }}
          required
        />

        {/* ========== 定时检测 ========== */}
        <ConfigSection title="定时检测" icon={<TimerIcon fontSize="small" color="action" />}>
          <Stack spacing={2}>
            <FormControlLabel
              control={<Switch checked={form.enabled} onChange={(e) => updateForm('enabled', e.target.checked)} />}
              label="启用定时检测"
            />
            {form.enabled && (
              <CronExpressionGenerator value={form.cronExpr} onChange={(value) => updateForm('cronExpr', value)} label="定时表达式" />
            )}
          </Stack>
        </ConfigSection>

        {/* ========== 测速模式 ========== */}
        <ConfigSection
          title="测速模式"
          icon={<SpeedIcon fontSize="small" color="action" />}
          helperText={
            form.mode === 'mihomo' ? '两阶段测试：先并发测延迟，再低并发测下载速度' : '仅测试延迟，速度更快，适合快速筛选可用节点'
          }
        >
          <Stack spacing={2}>
            <FormControl fullWidth size="small">
              <InputLabel>测速模式</InputLabel>
              <Select value={form.mode} label="测速模式" onChange={(e) => handleModeChange(e.target.value)}>
                <MenuItem value="tcp">仅延迟测试 (更快)</MenuItem>
                <MenuItem value="mihomo">延迟 + 下载速度测试</MenuItem>
              </Select>
            </FormControl>

            <Autocomplete
              freeSolo
              size="small"
              options={urlOptions}
              getOptionLabel={(option) => (typeof option === 'string' ? option : option.value)}
              value={form.testUrl}
              onChange={(_, newValue) => {
                const url = typeof newValue === 'string' ? newValue : newValue?.value || '';
                updateForm('testUrl', url);
              }}
              onInputChange={(_, newValue) => updateForm('testUrl', newValue || '')}
              renderOption={(props, option) => (
                <Box component="li" {...props} key={option.value}>
                  <Box>
                    <Typography variant="body2">{option.label}</Typography>
                    <Typography variant="caption" color="textSecondary" sx={{ wordBreak: 'break-all' }}>
                      {option.value}
                    </Typography>
                  </Box>
                </Box>
              )}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label={form.mode === 'mihomo' ? '下载测速URL' : '延迟测试URL'}
                  placeholder={form.mode === 'mihomo' ? '请选择或输入下载测速URL' : '请选择或输入204测速URL'}
                />
              )}
            />

            {/* 延迟测试URL - 仅在Mihomo模式显示 */}
            {form.mode === 'mihomo' && (
              <Autocomplete
                freeSolo
                size="small"
                options={LATENCY_TEST_URL_OPTIONS}
                getOptionLabel={(option) => (typeof option === 'string' ? option : option.value)}
                value={form.latencyUrl || ''}
                onChange={(_, newValue) => {
                  const url = typeof newValue === 'string' ? newValue : newValue?.value || '';
                  updateForm('latencyUrl', url);
                }}
                renderOption={(props, option) => (
                  <Box component="li" {...props} key={option.value}>
                    <Box>
                      <Typography variant="body2">{option.label}</Typography>
                      <Typography variant="caption" color="textSecondary" sx={{ wordBreak: 'break-all' }}>
                        {option.value}
                      </Typography>
                    </Box>
                  </Box>
                )}
                renderInput={(params) => <TextField {...params} label="延迟测试URL（阶段一）" placeholder="留空则使用速度测试URL" />}
              />
            )}

            <TextField
              fullWidth
              size="small"
              label="超时时间"
              type="text"
              inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
              value={form.timeout}
              onChange={(e) => {
                const val = e.target.value;
                if (val === '' || /^\d+$/.test(val)) {
                  updateForm('timeout', val === '' ? '' : Number(val));
                }
              }}
              onBlur={(e) => {
                const val = Number(e.target.value) || 5;
                updateForm('timeout', val);
              }}
              InputProps={{ endAdornment: <InputAdornment position="end">秒</InputAdornment> }}
            />

            {/* 速度记录模式 - 仅在Mihomo模式下显示 */}
            {form.mode === 'mihomo' && (
              <>
                <FormControl fullWidth size="small">
                  <InputLabel>速度记录模式</InputLabel>
                  <Select
                    value={form.speedRecordMode || 'average'}
                    label="速度记录模式"
                    onChange={(e) => updateForm('speedRecordMode', e.target.value)}
                  >
                    <MenuItem value="average">平均速度 (推荐)</MenuItem>
                    <MenuItem value="peak">峰值速度</MenuItem>
                  </Select>
                </FormControl>

                {form.speedRecordMode === 'peak' && (
                  <TextField
                    fullWidth
                    size="small"
                    label="峰值采样间隔"
                    type="text"
                    inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                    value={form.peakSampleInterval ?? 100}
                    onChange={(e) => {
                      const val = e.target.value;
                      if (val === '' || /^\d+$/.test(val)) {
                        updateForm('peakSampleInterval', val === '' ? '' : Number(val));
                      }
                    }}
                    onBlur={(e) => {
                      const val = Math.min(200, Math.max(50, Number(e.target.value) || 100));
                      updateForm('peakSampleInterval', val);
                    }}
                    InputProps={{ endAdornment: <InputAdornment position="end">毫秒</InputAdornment> }}
                    helperText="采样间隔范围：50-200毫秒"
                  />
                )}
              </>
            )}

            {/* 落地IP检测 */}
            <FormControlLabel
              control={
                <Switch
                  checked={form.detectCountry}
                  onChange={(e) => {
                    const checked = e.target.checked;
                    updateForm('detectCountry', checked);
                    if (checked && !form.landingIpUrl) {
                      updateForm('landingIpUrl', 'https://api.ipify.org');
                    }
                  }}
                  size="small"
                />
              }
              label={
                <Typography variant="body2">
                  检测落地IP国家
                  <Typography component="span" variant="caption" color="textSecondary" sx={{ ml: 0.5 }}>
                    (测速时顺便获取节点出口国家)
                  </Typography>
                </Typography>
              }
            />
            {form.detectCountry && (
              <FormControl fullWidth size="small">
                <InputLabel>IP查询接口</InputLabel>
                <Select
                  value={form.landingIpUrl || 'https://api.ipify.org'}
                  label="IP查询接口"
                  onChange={(e) => updateForm('landingIpUrl', e.target.value)}
                >
                  {LANDING_IP_URL_OPTIONS.map((opt) => (
                    <MenuItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            )}
          </Stack>
        </ConfigSection>

        {/* ========== 性能参数 ========== */}
        <ConfigSection title="性能参数" icon={<TuneIcon fontSize="small" color="action" />} defaultExpanded={true}>
          <Stack spacing={2}>
            {/* 握手时间设置 - 带详细说明 */}
            <Alert
              severity="info"
              variant="standard"
              icon={<InfoOutlinedIcon fontSize="small" />}
              sx={{
                '& .MuiAlert-message': { width: '100%' },
                py: 0.5
              }}
            >
              <FormControlLabel
                control={
                  <Switch
                    checked={form.includeHandshake ?? true}
                    onChange={(e) => updateForm('includeHandshake', e.target.checked)}
                    size="small"
                  />
                }
                label={
                  <Typography variant="body2" fontWeight={500}>
                    延迟包含握手时间
                  </Typography>
                }
                sx={{ mb: 0.5, ml: 0 }}
              />
              <Typography variant="caption" color="textSecondary" component="div">
                {(form.includeHandshake ?? true) ? (
                  <>
                    <strong>开启（推荐）</strong>：测量完整连接时间，包含TCP/TLS/代理协议握手。
                    <br />
                    反映真实使用体验，每次请求都需要握手。
                  </>
                ) : (
                  <>
                    <strong>关闭</strong>：先预热建立连接，再测量纯传输延迟。
                    <br />
                    适合精确评估网络线路质量（排除握手开销）。若预热失败则判定节点不可用。
                  </>
                )}
              </Typography>
            </Alert>

            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  size="small"
                  label="延迟测试并发"
                  type="text"
                  inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                  value={form.latencyConcurrency || ''}
                  placeholder="自动"
                  onChange={(e) => {
                    const val = e.target.value;
                    if (val === '' || /^\d+$/.test(val)) {
                      updateForm('latencyConcurrency', val === '' ? 0 : Number(val));
                    }
                  }}
                  helperText="0=智能动态"
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  size="small"
                  label="速度测试并发"
                  type="text"
                  inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                  value={form.speedConcurrency || ''}
                  placeholder="自动"
                  onChange={(e) => {
                    const val = e.target.value;
                    if (val === '' || /^\d+$/.test(val)) {
                      updateForm('speedConcurrency', val === '' ? 0 : Number(val));
                    }
                  }}
                  helperText="0=智能动态"
                />
              </Grid>
            </Grid>
          </Stack>
        </ConfigSection>

        {/* ========== 测速范围 ========== */}
        <ConfigSection
          title="测速范围"
          icon={<DataUsageIcon fontSize="small" color="action" />}
          defaultExpanded={false}
          helperText="分组优先级高于标签：选了分组则先按分组筛选，再按标签过滤；只选标签则直接按标签筛选；都不选则测全部"
        >
          <Stack spacing={2}>
            <Autocomplete
              multiple
              freeSolo
              size="small"
              options={groupOptions}
              value={form.groups}
              onChange={(_, newValue) => updateForm('groups', newValue)}
              renderInput={(params) => <TextField {...params} label="测速分组" placeholder="留空则测试全部分组" />}
            />
            <Autocomplete
              multiple
              size="small"
              options={tagOptions || []}
              getOptionLabel={(option) => option.name || option}
              value={form.tags.map((t) => tagOptions.find((tag) => tag.name === t) || { name: t })}
              onChange={(_, newValue) =>
                updateForm(
                  'tags',
                  newValue.map((t) => t.name || t)
                )
              }
              isOptionEqualToValue={(option, value) => (option.name || option) === (value.name || value)}
              renderTags={(value, getTagProps) =>
                value.map((option, index) => {
                  const tagObj = (tagOptions || []).find((t) => t.name === (option.name || option));
                  const { key, ...tagProps } = getTagProps({ index });
                  return (
                    <Chip
                      key={key}
                      label={option.name || option}
                      size="small"
                      sx={{
                        backgroundColor: tagObj?.color || '#1976d2',
                        color: '#fff',
                        '& .MuiChip-deleteIcon': { color: 'rgba(255,255,255,0.7)' }
                      }}
                      {...tagProps}
                    />
                  );
                })
              }
              renderOption={(props, option) => (
                <Box component="li" {...props} key={option.name}>
                  <Box
                    sx={{
                      width: 12,
                      height: 12,
                      borderRadius: '50%',
                      backgroundColor: option.color || '#1976d2',
                      mr: 1
                    }}
                  />
                  {option.name}
                </Box>
              )}
              renderInput={(params) => <TextField {...params} label="测速标签" placeholder="留空则不按标签过滤" />}
            />
          </Stack>
        </ConfigSection>

        {/* ========== 流量统计 ========== */}
        <ConfigSection title="流量统计" icon={<DataUsageIcon fontSize="small" color="action" />} defaultExpanded={false}>
          <Stack spacing={1}>
            <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
              <FormControlLabel
                control={
                  <Switch
                    checked={form.trafficByGroup ?? true}
                    onChange={(e) => updateForm('trafficByGroup', e.target.checked)}
                    size="small"
                  />
                }
                label={<Typography variant="body2">按分组统计</Typography>}
              />
              <FormControlLabel
                control={
                  <Switch
                    checked={form.trafficBySource ?? true}
                    onChange={(e) => updateForm('trafficBySource', e.target.checked)}
                    size="small"
                  />
                }
                label={<Typography variant="body2">按来源统计</Typography>}
              />
              <FormControlLabel
                control={
                  <Switch
                    checked={form.trafficByNode ?? false}
                    onChange={(e) => updateForm('trafficByNode', e.target.checked)}
                    size="small"
                    color="error"
                  />
                }
                label={
                  <Typography variant="body2">
                    按节点统计
                    <Typography component="span" variant="caption" color="error.main" sx={{ ml: 0.5 }}>
                      (大数据量)
                    </Typography>
                  </Typography>
                }
              />
            </Box>
            {form.trafficByNode && (
              <Typography variant="caption" color="error.main">
                ⚠️ 按节点统计会记录每个节点的流量消耗，节点数量过万时会增加约1-2MB存储空间
              </Typography>
            )}
          </Stack>
        </ConfigSection>
      </DialogContent>

      <DialogActions sx={{ px: 3, py: 2 }}>
        <Button onClick={onClose}>取消</Button>
        <Button variant="contained" onClick={handleSubmit} disabled={!form.name.trim() || submitting}>
          {submitting ? '保存中...' : '保存设置'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

NodeCheckProfileFormDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  profile: PropTypes.object,
  groupOptions: PropTypes.array,
  tagOptions: PropTypes.array,
  onSuccess: PropTypes.func
};
