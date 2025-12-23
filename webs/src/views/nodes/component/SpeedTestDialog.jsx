import PropTypes from 'prop-types';

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
import Fab from '@mui/material/Fab';
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
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import Grid from '@mui/material/Grid';

// icons
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import TimerIcon from '@mui/icons-material/Timer';
import SpeedIcon from '@mui/icons-material/Speed';
import TuneIcon from '@mui/icons-material/Tune';
import DataUsageIcon from '@mui/icons-material/DataUsage';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';

// project imports
import CronExpressionGenerator from 'components/CronExpressionGenerator';

// constants
import { SPEED_TEST_TCP_OPTIONS, SPEED_TEST_MIHOMO_OPTIONS, LATENCY_TEST_URL_OPTIONS, LANDING_IP_URL_OPTIONS } from '../utils';

// hooks
import { useState } from 'react';

/**
 * 配置区块组件 - 可折叠的设置分组
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
          },
          transition: 'background-color 0.2s'
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
 * 测速设置对话框 - 重构优化版
 */
export default function SpeedTestDialog({
  open,
  speedTestForm,
  setSpeedTestForm,
  groupOptions,
  tagOptions,
  onClose,
  onSubmit,
  onRunSpeedTest,
  onModeChange
}) {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <SpeedIcon color="primary" />
          <span>测速设置</span>
        </Box>
        <Tooltip title="使用当前配置立即开始测速" placement="left">
          <Fab
            color="primary"
            size={isMobile ? 'small' : 'medium'}
            onClick={() => {
              onRunSpeedTest();
              onClose();
            }}
            sx={{
              background: isDark
                ? 'linear-gradient(135deg, #4caf50 0%, #2e7d32 100%)'
                : 'linear-gradient(135deg, #66bb6a 0%, #43a047 100%)',
              boxShadow: isDark ? '0 4px 14px rgba(76, 175, 80, 0.4)' : '0 4px 14px rgba(76, 175, 80, 0.3)',
              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
              '&:hover': {
                background: isDark
                  ? 'linear-gradient(135deg, #66bb6a 0%, #388e3c 100%)'
                  : 'linear-gradient(135deg, #81c784 0%, #66bb6a 100%)',
                transform: 'scale(1.08)',
                boxShadow: isDark ? '0 6px 20px rgba(76, 175, 80, 0.5)' : '0 6px 20px rgba(76, 175, 80, 0.4)'
              },
              '&:active': {
                transform: 'scale(0.98)'
              }
            }}
          >
            <PlayArrowIcon />
          </Fab>
        </Tooltip>
      </DialogTitle>
      <DialogContent>
        {/* ========== 定时测速设置 ========== */}
        <ConfigSection title="定时测速" icon={<TimerIcon fontSize="small" color="action" />}>
          <Stack spacing={2}>
            <FormControlLabel
              control={
                <Switch
                  checked={speedTestForm.enabled}
                  onChange={(e) => setSpeedTestForm({ ...speedTestForm, enabled: e.target.checked })}
                />
              }
              label="启用自动测速"
            />
            {speedTestForm.enabled && (
              <CronExpressionGenerator
                value={speedTestForm.cron}
                onChange={(value) => setSpeedTestForm({ ...speedTestForm, cron: value })}
                label="定时测速设置"
              />
            )}
          </Stack>
        </ConfigSection>

        {/* ========== 测速模式与URL ========== */}
        <ConfigSection
          title="测速模式"
          icon={<SpeedIcon fontSize="small" color="action" />}
          helperText={
            speedTestForm.mode === 'mihomo' ? '两阶段测试：先并发测延迟，再低并发测下载速度' : '仅测试延迟，速度更快，适合快速筛选可用节点'
          }
        >
          <Stack spacing={2}>
            <FormControl fullWidth size="small">
              <InputLabel>测速模式</InputLabel>
              <Select variant={'outlined'} value={speedTestForm.mode} label="测速模式" onChange={(e) => onModeChange(e.target.value)}>
                <MenuItem value="tcp">仅延迟测试 (更快)</MenuItem>
                <MenuItem value="mihomo">延迟 + 下载速度测试</MenuItem>
              </Select>
            </FormControl>

            <Autocomplete
              freeSolo
              size="small"
              options={speedTestForm.mode === 'mihomo' ? SPEED_TEST_MIHOMO_OPTIONS : SPEED_TEST_TCP_OPTIONS}
              getOptionLabel={(option) => (typeof option === 'string' ? option : option.value)}
              value={speedTestForm.url}
              onChange={(e, newValue) => {
                const value = typeof newValue === 'string' ? newValue : newValue?.value || '';
                setSpeedTestForm({ ...speedTestForm, url: value });
              }}
              onInputChange={(e, newValue) => setSpeedTestForm({ ...speedTestForm, url: newValue || '' })}
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
                  label={speedTestForm.mode === 'mihomo' ? '下载测速URL' : '延迟测试URL'}
                  placeholder={speedTestForm.mode === 'mihomo' ? '请选择或输入下载测速URL' : '请选择或输入204测速URL'}
                />
              )}
            />

            {/* 延迟测试URL - 仅在Mihomo模式显示 */}
            {speedTestForm.mode === 'mihomo' && (
              <Autocomplete
                freeSolo
                size="small"
                options={LATENCY_TEST_URL_OPTIONS}
                getOptionLabel={(option) => (typeof option === 'string' ? option : option.value)}
                value={speedTestForm.latency_url || ''}
                onChange={(e, newValue) => {
                  const value = typeof newValue === 'string' ? newValue : newValue?.value || '';
                  setSpeedTestForm({ ...speedTestForm, latency_url: value });
                }}
                onInputChange={(e, newValue) => setSpeedTestForm({ ...speedTestForm, latency_url: newValue || '' })}
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
              value={speedTestForm.timeout}
              onChange={(e) => {
                const val = e.target.value;
                // 允许空字符串和纯数字
                if (val === '' || /^\d+$/.test(val)) {
                  setSpeedTestForm({ ...speedTestForm, timeout: val === '' ? '' : Number(val) });
                }
              }}
              onBlur={(e) => {
                // 失焦时确保有合法值，默认5秒
                const val = Number(e.target.value) || 5;
                setSpeedTestForm({ ...speedTestForm, timeout: val });
              }}
              InputProps={{ endAdornment: <InputAdornment position="end">秒</InputAdornment> }}
            />

            {/* 速度记录模式 - 仅在Mihomo模式下显示 */}
            {speedTestForm.mode === 'mihomo' && (
              <>
                <FormControl fullWidth size="small">
                  <InputLabel>速度记录模式</InputLabel>
                  <Select
                    value={speedTestForm.speed_record_mode || 'average'}
                    label="速度记录模式"
                    onChange={(e) => setSpeedTestForm({ ...speedTestForm, speed_record_mode: e.target.value })}
                  >
                    <MenuItem value="average">平均速度 (推荐)</MenuItem>
                    <MenuItem value="peak">峰值速度</MenuItem>
                  </Select>
                </FormControl>

                {speedTestForm.speed_record_mode === 'peak' && (
                  <TextField
                    fullWidth
                    size="small"
                    label="峰值采样间隔"
                    type="text"
                    inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                    value={speedTestForm.peak_sample_interval ?? 100}
                    onChange={(e) => {
                      const val = e.target.value;
                      // 允许空字符串和纯数字
                      if (val === '' || /^\d+$/.test(val)) {
                        setSpeedTestForm({ ...speedTestForm, peak_sample_interval: val === '' ? '' : Number(val) });
                      }
                    }}
                    onBlur={(e) => {
                      // 失焦时强制限制范围50-200，默认100
                      const val = Math.min(200, Math.max(50, Number(e.target.value) || 100));
                      setSpeedTestForm({ ...speedTestForm, peak_sample_interval: val });
                    }}
                    InputProps={{ endAdornment: <InputAdornment position="end">毫秒</InputAdornment> }}
                    helperText="采样间隔范围：50-200毫秒"
                  />
                )}
              </>
            )}

            {/* 落地IP检测 - 测速时的附加功能 */}
            <FormControlLabel
              control={
                <Switch
                  checked={speedTestForm.detect_country}
                  onChange={(e) => {
                    const checked = e.target.checked;
                    setSpeedTestForm({
                      ...speedTestForm,
                      detect_country: checked,
                      // 开启时自动设置默认URL
                      landing_ip_url: checked && !speedTestForm.landing_ip_url ? 'https://api.ipify.org' : speedTestForm.landing_ip_url
                    });
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
            {speedTestForm.detect_country && (
              <FormControl fullWidth size="small">
                <InputLabel>IP查询接口</InputLabel>
                <Select
                  value={speedTestForm.landing_ip_url || 'https://api.ipify.org'}
                  label="IP查询接口"
                  onChange={(e) => setSpeedTestForm({ ...speedTestForm, landing_ip_url: e.target.value })}
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
                    checked={speedTestForm.include_handshake ?? true}
                    onChange={(e) => setSpeedTestForm({ ...speedTestForm, include_handshake: e.target.checked })}
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
                {(speedTestForm.include_handshake ?? true) ? (
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
              <Grid item size={{ xs: 12, sm: 6 }}>
                <TextField
                  fullWidth
                  size="small"
                  label="延迟测试并发"
                  type="text"
                  inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                  value={speedTestForm.latency_concurrency || ''}
                  placeholder="自动"
                  onChange={(e) => {
                    const val = e.target.value;
                    if (val === '' || /^\d+$/.test(val)) {
                      setSpeedTestForm({ ...speedTestForm, latency_concurrency: val === '' ? 0 : Number(val) });
                    }
                  }}
                  onBlur={(e) => {
                    const val = Math.min(1000, Math.max(0, Number(e.target.value) || 0));
                    setSpeedTestForm({ ...speedTestForm, latency_concurrency: val });
                  }}
                  helperText="0=智能动态"
                />
              </Grid>
              <Grid item size={{ xs: 12, sm: 6 }}>
                <TextField
                  fullWidth
                  size="small"
                  label="速度测试并发"
                  type="text"
                  inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                  value={speedTestForm.speed_concurrency || ''}
                  placeholder="自动"
                  onChange={(e) => {
                    const val = e.target.value;
                    if (val === '' || /^\d+$/.test(val)) {
                      setSpeedTestForm({ ...speedTestForm, speed_concurrency: val === '' ? 0 : Number(val) });
                    }
                  }}
                  onBlur={(e) => {
                    const val = Math.min(128, Math.max(0, Number(e.target.value) || 0));
                    setSpeedTestForm({ ...speedTestForm, speed_concurrency: val });
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
              value={speedTestForm.groups || []}
              onChange={(e, newValue) => setSpeedTestForm({ ...speedTestForm, groups: newValue })}
              renderInput={(params) => <TextField {...params} label="测速分组" placeholder="留空则测试全部分组" />}
            />
            <Autocomplete
              multiple
              size="small"
              options={tagOptions || []}
              getOptionLabel={(option) => option.name || option}
              value={speedTestForm.tags || []}
              onChange={(e, newValue) => setSpeedTestForm({ ...speedTestForm, tags: newValue.map((t) => t.name || t) })}
              isOptionEqualToValue={(option, value) => (option.name || option) === value}
              renderTags={(value, getTagProps) =>
                value.map((option, index) => {
                  const tagObj = (tagOptions || []).find((t) => t.name === option);
                  const { key, ...tagProps } = getTagProps({ index });
                  return (
                    <Chip
                      key={key}
                      label={option}
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
                    checked={speedTestForm.traffic_by_group ?? true}
                    onChange={(e) => setSpeedTestForm({ ...speedTestForm, traffic_by_group: e.target.checked })}
                    size="small"
                  />
                }
                label={<Typography variant="body2">按分组统计</Typography>}
              />
              <FormControlLabel
                control={
                  <Switch
                    checked={speedTestForm.traffic_by_source ?? true}
                    onChange={(e) => setSpeedTestForm({ ...speedTestForm, traffic_by_source: e.target.checked })}
                    size="small"
                  />
                }
                label={<Typography variant="body2">按来源统计</Typography>}
              />
              <FormControlLabel
                control={
                  <Switch
                    checked={speedTestForm.traffic_by_node ?? false}
                    onChange={(e) => setSpeedTestForm({ ...speedTestForm, traffic_by_node: e.target.checked })}
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
            {speedTestForm.traffic_by_node && (
              <Typography variant="caption" color="error.main">
                ⚠️ 按节点统计会记录每个节点的流量消耗，节点数量过万时会增加约1-2MB存储空间
              </Typography>
            )}
          </Stack>
        </ConfigSection>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>取消</Button>
        <Button variant="contained" onClick={onSubmit}>
          保存设置
        </Button>
      </DialogActions>
    </Dialog>
  );
}

SpeedTestDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  speedTestForm: PropTypes.shape({
    cron: PropTypes.string,
    enabled: PropTypes.bool,
    mode: PropTypes.string,
    url: PropTypes.string,
    latency_url: PropTypes.string,
    timeout: PropTypes.number,
    groups: PropTypes.array,
    tags: PropTypes.array,
    detect_country: PropTypes.bool,
    latency_concurrency: PropTypes.number,
    speed_concurrency: PropTypes.number,
    traffic_by_group: PropTypes.bool,
    traffic_by_source: PropTypes.bool,
    traffic_by_node: PropTypes.bool,
    include_handshake: PropTypes.bool,
    speed_record_mode: PropTypes.string,
    peak_sample_interval: PropTypes.number,
    landing_ip_url: PropTypes.string
  }).isRequired,

  setSpeedTestForm: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  tagOptions: PropTypes.array,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onRunSpeedTest: PropTypes.func.isRequired,
  onModeChange: PropTypes.func.isRequired
};
