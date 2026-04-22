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
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

// constants
import { SPEED_TEST_TCP_OPTIONS, SPEED_TEST_MIHOMO_OPTIONS, LATENCY_TEST_URL_OPTIONS, LANDING_IP_URL_OPTIONS } from '../utils';

// hooks
import { useState } from 'react';

/**
 * 配置区块组件 - 可折叠的设置分组
 */
function getSpeedTestDialogThemeTokens(theme, isDark) {
  const surfaceTokens = getSurfaceTokens(theme, isDark);
  const textTokens = getReadableTextTokens(theme, isDark);
  const { palette, mutedPanelSurface, nestedPanelSurface } = surfaceTokens;

  return {
    ...surfaceTokens,
    ...textTokens,
    headerSurface: isDark ? withAlpha(palette.background.paper, 0.18) : mutedPanelSurface,
    actionSurface: isDark ? withAlpha(palette.background.default, 0.88) : mutedPanelSurface,
    sectionSurface: isDark ? withAlpha(palette.background.paper, 0.36) : nestedPanelSurface,
    sectionHeaderSurface: isDark ? withAlpha(palette.background.default, 0.84) : withAlpha(palette.background.default, 0.72),
    sectionHoverSurface: isDark
      ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.18)} 0%, ${withAlpha(palette.primary.main, 0.08)} 100%)`
      : withAlpha(palette.primary.main, 0.04),
    fieldSurface: isDark ? withAlpha(palette.background.paper, 0.74) : palette.background.paper,
    fieldHoverBorder: withAlpha(palette.primary.main, isDark ? 0.4 : 0.24),
    overlaySurface: isDark ? withAlpha(palette.background.default, 0.98) : palette.background.paper,
    overlayHoverSurface: withAlpha(palette.primary.main, isDark ? 0.12 : 0.06),
    overlaySelectedSurface: withAlpha(palette.primary.main, isDark ? 0.18 : 0.08),
    controlRowSurface: isDark ? withAlpha(palette.background.paper, 0.24) : withAlpha(palette.background.default, 0.92),
    controlRowBorder: isDark ? withAlpha(palette.divider, 0.68) : surfaceTokens.panelBorder,
    infoAlertSurface: withAlpha(palette.info.main, isDark ? 0.12 : 0.06),
    infoAlertBorder: withAlpha(palette.info.main, isDark ? 0.28 : 0.16),
    warningAlertSurface: withAlpha(palette.error.main, isDark ? 0.1 : 0.05),
    warningAlertBorder: withAlpha(palette.error.main, isDark ? 0.24 : 0.16),
    successFabSurface: `linear-gradient(135deg, ${palette.success.light} 0%, ${palette.success.main} 100%)`,
    successFabHoverSurface: `linear-gradient(135deg, ${palette.success.main} 0%, ${palette.success.dark} 100%)`,
    successFabShadow: `0 4px 14px ${withAlpha(palette.success.main, isDark ? 0.42 : 0.26)}`,
    successFabHoverShadow: `0 6px 20px ${withAlpha(palette.success.main, isDark ? 0.5 : 0.34)}`
  };
}

function getOverlayPaperSx(themeTokens, theme) {
  const { isDark, palette, overlaySurface, overlayHoverSurface, overlaySelectedSurface, panelBorder } = themeTokens;

  return {
    mt: 0.5,
    borderRadius: 2,
    backgroundColor: overlaySurface,
    backgroundImage: isDark ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.14)} 0%, ${overlaySurface} 100%)` : 'none',
    border: '1px solid',
    borderColor: panelBorder,
    boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : theme.shadows[8],
    '& .MuiMenuItem-root': {
      mx: 0.75,
      my: 0.25,
      borderRadius: 1.5,
      '&:hover': {
        backgroundColor: overlayHoverSurface
      },
      '&.Mui-selected': {
        backgroundColor: overlaySelectedSurface,
        '&:hover': {
          backgroundColor: withAlpha(palette.primary.main, isDark ? 0.24 : 0.12)
        }
      }
    },
    '& .MuiAutocomplete-option': {
      mx: 0.75,
      my: 0.25,
      borderRadius: 1.5,
      alignItems: 'flex-start',
      '&.Mui-focused': {
        backgroundColor: overlayHoverSurface
      },
      '&[aria-selected="true"]': {
        backgroundColor: overlaySelectedSurface
      },
      '&[aria-selected="true"].Mui-focused': {
        backgroundColor: withAlpha(palette.primary.main, isDark ? 0.24 : 0.12)
      }
    }
  };
}

function ConfigSection({ title, icon, children, defaultExpanded = true, helperText, themeTokens }) {
  const [expanded, setExpanded] = useState(defaultExpanded);
  const { isDark, palette, panelBorder, primaryText, secondaryText, sectionSurface, sectionHeaderSurface, sectionHoverSurface } =
    themeTokens;

  return (
    <Paper
      elevation={0}
      sx={{
        border: `1px solid ${panelBorder}`,
        borderRadius: 2,
        overflow: 'hidden',
        mb: 2,
        backgroundColor: sectionSurface,
        boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none'
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
          backgroundColor: sectionHeaderSurface,
          '&:hover': {
            background: sectionHoverSurface
          },
          transition: 'background 0.2s ease'
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          {icon}
          <Typography variant="subtitle2" fontWeight={600} sx={{ color: primaryText }}>
            {title}
          </Typography>
        </Box>
        {expanded ? (
          <ExpandLessIcon fontSize="small" sx={{ color: secondaryText }} />
        ) : (
          <ExpandMoreIcon fontSize="small" sx={{ color: secondaryText }} />
        )}
      </Box>
      <Collapse in={expanded}>
        <Divider sx={{ borderColor: panelBorder }} />
        <Box sx={{ p: 2 }}>
          {children}
          {helperText && (
            <Typography variant="caption" sx={{ mt: 1.5, display: 'block', color: secondaryText }}>
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
  helperText: PropTypes.string,
  themeTokens: PropTypes.object.isRequired
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
  const { isDark } = useResolvedColorScheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const themeTokens = getSpeedTestDialogThemeTokens(theme, isDark);
  const {
    palette,
    dialogSurface,
    dialogSurfaceGradient,
    headerSurface,
    actionSurface,
    panelBorder,
    fieldSurface,
    fieldHoverBorder,
    primaryText,
    secondaryText,
    tertiaryText,
    controlRowSurface,
    controlRowBorder,
    infoAlertSurface,
    infoAlertBorder,
    warningAlertSurface,
    warningAlertBorder,
    successFabSurface,
    successFabHoverSurface,
    successFabShadow,
    successFabHoverShadow
  } = themeTokens;
  const overlayPaperSx = getOverlayPaperSx(themeTokens, theme);
  const selectMenuProps = {
    PaperProps: {
      sx: overlayPaperSx
    },
    MenuListProps: {
      sx: { py: 0.5 }
    }
  };
  const autocompleteSlotProps = {
    paper: {
      sx: overlayPaperSx
    },
    listbox: {
      sx: { py: 0.5 }
    }
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{
        sx: {
          borderRadius: 3,
          backgroundColor: dialogSurface,
          backgroundImage: dialogSurfaceGradient,
          border: '1px solid',
          borderColor: panelBorder,
          boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.05)}` : theme.shadows[8]
        }
      }}
    >
      <DialogTitle
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          borderBottom: `1px solid ${panelBorder}`,
          backgroundColor: headerSurface,
          color: primaryText
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <SpeedIcon color="primary" />
          <Typography variant="h6" sx={{ color: primaryText }}>
            测速设置
          </Typography>
        </Box>
        <Tooltip title="使用当前配置立即开始测速" placement="left">
          <Fab
            size={isMobile ? 'small' : 'medium'}
            onClick={() => {
              onRunSpeedTest();
              onClose();
            }}
            sx={{
              background: successFabSurface,
              boxShadow: successFabShadow,
              color: theme.palette.common.white,
              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
              '&:hover': {
                background: successFabHoverSurface,
                transform: 'scale(1.08)',
                boxShadow: successFabHoverShadow
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
      <DialogContent
        sx={{
          backgroundColor: dialogSurface,
          '& .MuiOutlinedInput-root': {
            backgroundColor: fieldSurface,
            transition: 'background-color 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease',
            '& fieldset': {
              borderColor: panelBorder
            },
            '&:hover fieldset': {
              borderColor: fieldHoverBorder
            }
          },
          '& .MuiFormHelperText-root': {
            color: secondaryText
          },
          '& .MuiFormLabel-root': {
            color: secondaryText
          },
          '& .MuiInputAdornment-root, & .MuiAutocomplete-endAdornment': {
            color: secondaryText
          },
          '& .MuiFormControlLabel-label': {
            color: primaryText
          }
        }}
      >
        {/* ========== 定时测速设置 ========== */}
        <ConfigSection title="定时测速" icon={<TimerIcon fontSize="small" color="action" />} themeTokens={themeTokens}>
          <Stack spacing={2}>
            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                gap: 2,
                p: 1.5,
                borderRadius: 2,
                backgroundColor: controlRowSurface,
                border: `1px solid ${controlRowBorder}`
              }}
            >
              <Box>
                <Typography variant="body2" sx={{ color: primaryText }}>
                  启用自动测速
                </Typography>
                <Typography variant="caption" sx={{ color: tertiaryText }}>
                  关闭后将停止按计划自动执行测速任务
                </Typography>
              </Box>
              <Switch checked={speedTestForm.enabled} onChange={(e) => setSpeedTestForm({ ...speedTestForm, enabled: e.target.checked })} />
            </Box>
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
          themeTokens={themeTokens}
          helperText={
            speedTestForm.mode === 'mihomo' ? '两阶段测试：先并发测延迟，再低并发测下载速度' : '仅测试延迟，速度更快，适合快速筛选可用节点'
          }
        >
          <Stack spacing={2}>
            <FormControl fullWidth size="small">
              <InputLabel>测速模式</InputLabel>
              <Select
                variant="outlined"
                value={speedTestForm.mode}
                label="测速模式"
                onChange={(e) => onModeChange(e.target.value)}
                MenuProps={selectMenuProps}
              >
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
              slotProps={autocompleteSlotProps}
              renderOption={(props, option) => (
                <Box component="li" {...props} key={option.value}>
                  <Box>
                    <Typography variant="body2" sx={{ color: primaryText }}>
                      {option.label}
                    </Typography>
                    <Typography variant="caption" sx={{ wordBreak: 'break-all', color: secondaryText }}>
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
                slotProps={autocompleteSlotProps}
                renderOption={(props, option) => (
                  <Box component="li" {...props} key={option.value}>
                    <Box>
                      <Typography variant="body2" sx={{ color: primaryText }}>
                        {option.label}
                      </Typography>
                      <Typography variant="caption" sx={{ wordBreak: 'break-all', color: secondaryText }}>
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
                    MenuProps={selectMenuProps}
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
                <Typography variant="body2" sx={{ color: primaryText }}>
                  检测落地IP国家
                  <Typography component="span" variant="caption" sx={{ ml: 0.5, color: secondaryText }}>
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
                  MenuProps={selectMenuProps}
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
        <ConfigSection
          title="性能参数"
          icon={<TuneIcon fontSize="small" color="action" />}
          defaultExpanded={true}
          themeTokens={themeTokens}
        >
          <Stack spacing={2}>
            {/* 握手时间设置 - 带详细说明 */}
            <Alert
              severity="info"
              variant="standard"
              icon={<InfoOutlinedIcon fontSize="small" />}
              sx={{
                color: isDark ? palette.info.light : palette.text.primary,
                backgroundColor: infoAlertSurface,
                border: `1px solid ${infoAlertBorder}`,
                boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none',
                '& .MuiAlert-message': { width: '100%' },
                '& .MuiAlert-icon': { color: isDark ? palette.info.light : palette.info.main },
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
                  <Typography variant="body2" fontWeight={500} sx={{ color: primaryText }}>
                    延迟包含握手时间
                  </Typography>
                }
                sx={{ mb: 0.5, ml: 0 }}
              />
              <Typography variant="caption" component="div" sx={{ color: secondaryText }}>
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
          themeTokens={themeTokens}
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
              slotProps={autocompleteSlotProps}
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
              slotProps={autocompleteSlotProps}
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
        <ConfigSection
          title="流量统计"
          icon={<DataUsageIcon fontSize="small" color="action" />}
          defaultExpanded={false}
          themeTokens={themeTokens}
        >
          <Stack spacing={1}>
            <Box
              sx={{
                display: 'flex',
                flexWrap: 'wrap',
                gap: 1,
                p: 1.25,
                borderRadius: 2,
                backgroundColor: controlRowSurface,
                border: `1px solid ${controlRowBorder}`
              }}
            >
              <FormControlLabel
                control={
                  <Switch
                    checked={speedTestForm.traffic_by_group ?? true}
                    onChange={(e) => setSpeedTestForm({ ...speedTestForm, traffic_by_group: e.target.checked })}
                    size="small"
                  />
                }
                label={
                  <Typography variant="body2" sx={{ color: primaryText }}>
                    按分组统计
                  </Typography>
                }
              />
              <FormControlLabel
                control={
                  <Switch
                    checked={speedTestForm.traffic_by_source ?? true}
                    onChange={(e) => setSpeedTestForm({ ...speedTestForm, traffic_by_source: e.target.checked })}
                    size="small"
                  />
                }
                label={
                  <Typography variant="body2" sx={{ color: primaryText }}>
                    按来源统计
                  </Typography>
                }
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
                  <Typography variant="body2" sx={{ color: primaryText }}>
                    按节点统计
                    <Typography component="span" variant="caption" color="error.main" sx={{ ml: 0.5 }}>
                      (大数据量)
                    </Typography>
                  </Typography>
                }
              />
            </Box>
            {speedTestForm.traffic_by_node && (
              <Alert
                severity="error"
                variant="standard"
                sx={{
                  backgroundColor: warningAlertSurface,
                  border: `1px solid ${warningAlertBorder}`,
                  '& .MuiAlert-icon': { color: palette.error.main }
                }}
              >
                <Typography variant="caption" sx={{ color: primaryText }}>
                  按节点统计会记录每个节点的流量消耗，节点数量过万时会增加约1-2MB存储空间
                </Typography>
              </Alert>
            )}
          </Stack>
        </ConfigSection>
      </DialogContent>
      <DialogActions
        sx={{
          backgroundColor: actionSurface,
          borderTop: `1px solid ${panelBorder}`
        }}
      >
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
