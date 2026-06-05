import PropTypes from 'prop-types';
import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

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
import Checkbox from '@mui/material/Checkbox';

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
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { withAlpha } from 'utils/colorUtils';
import { getNodeCheckStrategyThemeTokens } from '../nodeCheckTheme';

// api
import { createNodeCheckProfile, getNodeCheckMeta, updateNodeCheckProfile } from 'api/nodeCheck';

// constants
import {
  SPEED_TEST_TCP_OPTIONS,
  SPEED_TEST_MIHOMO_OPTIONS,
  LATENCY_TEST_URL_OPTIONS,
  LANDING_IP_URL_OPTIONS,
  QUALITY_CHECK_URL_OPTIONS,
  buildNodeCheckProfilePayload,
  createNodeCheckProfileFormState,
  formatUnlockProviderLabel,
  getUnlockProviderOptions,
  setUnlockMeta
} from '../utils';

/**
 * 可折叠配置区块
 */
function ConfigSection({ title, icon, children, defaultExpanded = true, helperText, themeTokens }) {
  const [expanded, setExpanded] = useState(defaultExpanded);
  const { isDark, palette, panelBorder, sectionSurface, sectionHeaderSurface, sectionHoverSurface } = themeTokens;

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
        <Divider sx={{ borderColor: panelBorder }} />
        <Box sx={{ p: 2 }}>
          {children}
          {helperText && (
            <Typography variant="caption" color="text.secondary" sx={{ mt: 1.5, display: 'block' }}>
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
 * 节点检测策略编辑表单对话框
 */
export default function NodeCheckProfileFormDialog({ open, onClose, profile, groupOptions = [], tagOptions = [], onSuccess }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const isEdit = !!profile;
  const { isDark } = useResolvedColorScheme();
  const themeTokens = getNodeCheckStrategyThemeTokens(theme, isDark);
  const { palette, dialogSurface, dialogSurfaceGradient, headerSurface, actionSurface, panelBorder } = themeTokens;
  const getTokenChipSx = (accent) => ({
    color: accent,
    backgroundColor: withAlpha(accent, isDark ? 0.18 : 0.1),
    border: `1px solid ${withAlpha(accent, isDark ? 0.34 : 0.2)}`,
    boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none',
    '& .MuiChip-deleteIcon': {
      color: withAlpha(accent, isDark ? 0.78 : 0.62),
      '&:hover': {
        color: accent
      }
    }
  });
  const getAlertSx = (accent) => ({
    color: isDark ? accent : palette.text.primary,
    backgroundColor: withAlpha(accent, isDark ? 0.12 : 0.08),
    border: `1px solid ${withAlpha(accent, isDark ? 0.28 : 0.16)}`,
    boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none'
  });

  const autocompleteChipSx = {
    '& .MuiAutocomplete-tag': {
      bgcolor: isDark ? withAlpha(palette.primary.main, 0.16) : undefined,
      color: isDark ? themeTokens.primaryText : undefined,
      border: isDark ? '1px solid' : undefined,
      borderColor: isDark ? withAlpha(palette.primary.main, 0.3) : undefined,
      '& .MuiChip-deleteIcon': {
        color: isDark ? withAlpha(palette.primary.main, 0.7) : undefined,
        transition: 'color 0.2s',
        '&:hover': {
          color: isDark ? palette.primary.main : undefined
        }
      }
    }
  };

  // 表单状态
  const [form, setForm] = useState(() => createNodeCheckProfileFormState());

  const [submitting, setSubmitting] = useState(false);
  const [unlockMetaLoading, setUnlockMetaLoading] = useState(false);
  const [unlockProviderQuery, setUnlockProviderQuery] = useState('');

  // 初始化表单
  useEffect(() => {
    if (open) {
      setForm(createNodeCheckProfileFormState(profile));
    }
  }, [open, profile]);

  useEffect(() => {
    if (!open) return;

    const loadUnlockMeta = async () => {
      setUnlockMetaLoading(true);
      try {
        const response = await getNodeCheckMeta();
        setUnlockMeta(response.data || {});
      } catch (error) {
        console.error('加载解锁元数据失败:', error);
      } finally {
        setUnlockMetaLoading(false);
      }
    };

    loadUnlockMeta();
  }, [open]);

  useEffect(() => {
    if (open && groupOptions && form.groups?.length > 0) {
      const validGroups = form.groups.filter((g) => groupOptions.includes(g));
      if (validGroups.length !== form.groups.length) {
        setForm((prev) => ({ ...prev, groups: validGroups }));
      }
    }
  }, [groupOptions, open, form.groups, setForm]);

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
      const data = buildNodeCheckProfilePayload(form);

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
  const unlockProviderOptions = getUnlockProviderOptions();

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      fullScreen={isMobile}
      PaperProps={{
        sx: {
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
          borderBottom: `1px solid ${panelBorder}`,
          backgroundColor: headerSurface
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <SpeedIcon color="primary" />
          <span>{isEdit ? t('nodes.nodeCheckProfiles.form.editTitle') : t('nodes.nodeCheckProfiles.form.createTitle')}</span>
        </Box>
      </DialogTitle>

      <DialogContent dividers sx={{ backgroundColor: dialogSurface, borderColor: panelBorder }}>
        {/* 策略名称 - 增加上边距避免遮挡 */}
        <TextField
          fullWidth
          label={t('nodes.nodeCheckProfiles.form.name')}
          value={form.name}
          onChange={(e) => updateForm('name', e.target.value)}
          placeholder={t('nodes.nodeCheckProfiles.form.namePlaceholder')}
          sx={{ mb: 2, mt: 1 }}
          required
        />

        {/* ========== 定时检测 ========== */}
        <ConfigSection
          title={t('nodes.nodeCheckProfiles.form.scheduleTitle')}
          icon={<TimerIcon fontSize="small" color="action" />}
          themeTokens={themeTokens}
        >
          <Stack spacing={2}>
            <FormControlLabel
              control={<Switch checked={form.enabled} onChange={(e) => updateForm('enabled', e.target.checked)} />}
              label={t('nodes.nodeCheckProfiles.form.enableSchedule')}
            />
            {form.enabled && (
              <CronExpressionGenerator
                value={form.cronExpr}
                onChange={(value) => updateForm('cronExpr', value)}
                label={t('nodes.nodeCheckProfiles.form.cronExpression')}
              />
            )}
          </Stack>
        </ConfigSection>

        {/* ========== 测速模式 ========== */}
        <ConfigSection
          title={t('nodes.nodeCheckProfiles.form.modeTitle')}
          icon={<SpeedIcon fontSize="small" color="action" />}
          themeTokens={themeTokens}
          helperText={
            form.mode === 'mihomo'
              ? t('nodes.nodeCheckProfiles.form.modeHelper.full')
              : t('nodes.nodeCheckProfiles.form.modeHelper.delayOnly')
          }
        >
          <Stack spacing={2}>
            <FormControl fullWidth size="small">
              <InputLabel>{t('nodes.nodeCheckProfiles.form.mode')}</InputLabel>
              <Select value={form.mode} label={t('nodes.nodeCheckProfiles.form.mode')} onChange={(e) => handleModeChange(e.target.value)}>
                <MenuItem value="tcp">{t('nodes.nodeCheckProfiles.form.modeOptions.tcp')}</MenuItem>
                <MenuItem value="mihomo">{t('nodes.nodeCheckProfiles.form.modeOptions.mihomo')}</MenuItem>
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
                    <Typography variant="caption" color="text.secondary" sx={{ wordBreak: 'break-all' }}>
                      {option.value}
                    </Typography>
                  </Box>
                </Box>
              )}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label={
                    form.mode === 'mihomo'
                      ? t('nodes.nodeCheckProfiles.form.downloadTestUrl')
                      : t('nodes.nodeCheckProfiles.form.delayTestUrl')
                  }
                  placeholder={
                    form.mode === 'mihomo'
                      ? t('nodes.nodeCheckProfiles.form.downloadTestUrlPlaceholder')
                      : t('nodes.nodeCheckProfiles.form.delayTestUrlPlaceholder')
                  }
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
                onInputChange={(_, newValue) => updateForm('latencyUrl', newValue || '')}
                renderOption={(props, option) => (
                  <Box component="li" {...props} key={option.value}>
                    <Box>
                      <Typography variant="body2">{option.label}</Typography>
                      <Typography variant="caption" color="text.secondary" sx={{ wordBreak: 'break-all' }}>
                        {option.value}
                      </Typography>
                    </Box>
                  </Box>
                )}
                renderInput={(params) => (
                  <TextField
                    {...params}
                    label={t('nodes.nodeCheckProfiles.form.latencyStageUrl')}
                    placeholder={t('nodes.nodeCheckProfiles.form.latencyStageUrlPlaceholder')}
                  />
                )}
              />
            )}

            <TextField
              fullWidth
              size="small"
              label={t('nodes.nodeCheckProfiles.form.timeout')}
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
              InputProps={{
                endAdornment: <InputAdornment position="end">{t('nodes.nodeCheckProfiles.form.units.seconds')}</InputAdornment>
              }}
            />

            {/* 速度记录模式 - 仅在Mihomo模式下显示 */}
            {form.mode === 'mihomo' && (
              <>
                <FormControl fullWidth size="small">
                  <InputLabel>{t('nodes.nodeCheckProfiles.form.speedRecordMode')}</InputLabel>
                  <Select
                    value={form.speedRecordMode || 'average'}
                    label={t('nodes.nodeCheckProfiles.form.speedRecordMode')}
                    onChange={(e) => updateForm('speedRecordMode', e.target.value)}
                  >
                    <MenuItem value="average">{t('nodes.nodeCheckProfiles.form.speedRecordOptions.average')}</MenuItem>
                    <MenuItem value="peak">{t('nodes.nodeCheckProfiles.form.speedRecordOptions.peak')}</MenuItem>
                  </Select>
                </FormControl>

                {form.speedRecordMode === 'peak' && (
                  <TextField
                    fullWidth
                    size="small"
                    label={t('nodes.nodeCheckProfiles.form.peakSampleInterval')}
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
                    InputProps={{
                      endAdornment: <InputAdornment position="end">{t('nodes.nodeCheckProfiles.form.units.milliseconds')}</InputAdornment>
                    }}
                    helperText={t('nodes.nodeCheckProfiles.form.peakSampleHelper')}
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
                  {t('nodes.nodeCheckProfiles.form.detectCountry')}
                  <Typography component="span" variant="caption" color="text.secondary" sx={{ ml: 0.5 }}>
                    {t('nodes.nodeCheckProfiles.form.detectCountryHint')}
                  </Typography>
                </Typography>
              }
            />
            {form.detectCountry && (
              <FormControl fullWidth size="small">
                <InputLabel>{t('nodes.nodeCheckProfiles.form.ipQueryUrl')}</InputLabel>
                <Select
                  value={form.landingIpUrl || 'https://api.ipify.org'}
                  label={t('nodes.nodeCheckProfiles.form.ipQueryUrl')}
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

            <FormControlLabel
              control={
                <Switch
                  checked={form.detectQuality}
                  onChange={(e) => {
                    const checked = e.target.checked;
                    updateForm('detectQuality', checked);
                    if (checked && !form.qualityCheckUrl) {
                      updateForm('qualityCheckUrl', 'https://my.123169.xyz/v1/info');
                    }
                  }}
                  size="small"
                />
              }
              label={
                <Typography variant="body2">
                  {t('nodes.nodeCheckProfiles.form.detectQuality')}
                  <Typography component="span" variant="caption" color="text.secondary" sx={{ ml: 0.5 }}>
                    {t('nodes.nodeCheckProfiles.form.detectQualityHint')}
                  </Typography>
                </Typography>
              }
            />
            {form.detectQuality && (
              <Autocomplete
                freeSolo
                size="small"
                options={QUALITY_CHECK_URL_OPTIONS}
                getOptionLabel={(option) => (typeof option === 'string' ? option : option.value)}
                value={form.qualityCheckUrl || ''}
                onChange={(_, newValue) => {
                  const url = typeof newValue === 'string' ? newValue : newValue?.value || '';
                  updateForm('qualityCheckUrl', url);
                }}
                onInputChange={(_, newValue) => updateForm('qualityCheckUrl', newValue || '')}
                renderOption={(props, option) => (
                  <Box component="li" {...props} key={option.value}>
                    <Box>
                      <Typography variant="body2">{option.label}</Typography>
                      <Typography variant="caption" color="text.secondary" sx={{ wordBreak: 'break-all' }}>
                        {option.value}
                      </Typography>
                    </Box>
                  </Box>
                )}
                renderInput={(params) => (
                  <TextField
                    {...params}
                    label={t('nodes.nodeCheckProfiles.form.qualityCheckUrl')}
                    placeholder={t('nodes.nodeCheckProfiles.form.qualityCheckUrlPlaceholder')}
                  />
                )}
              />
            )}

            <FormControlLabel
              control={<Switch checked={form.detectUnlock} onChange={(e) => updateForm('detectUnlock', e.target.checked)} size="small" />}
              label={
                <Typography variant="body2">
                  {t('nodes.nodeCheckProfiles.form.detectUnlock')}
                  <Typography component="span" variant="caption" color="text.secondary" sx={{ ml: 0.5 }}>
                    {t('nodes.nodeCheckProfiles.form.detectUnlockHint')}
                  </Typography>
                </Typography>
              }
            />
            {form.detectUnlock && (
              <Stack spacing={1.5}>
                <Alert severity="warning" variant="outlined" sx={getAlertSx(palette.warning.main)}>
                  {t('nodes.nodeCheckProfiles.form.unlockWarning')}
                </Alert>
                <Autocomplete
                  multiple
                  size="small"
                  disableCloseOnSelect
                  options={unlockProviderOptions}
                  getOptionLabel={(option) => (typeof option === 'string' ? option : option.label || option.value)}
                  value={form.unlockProviders}
                  inputValue={unlockProviderQuery}
                  onInputChange={(_, newValue) => setUnlockProviderQuery(newValue || '')}
                  onChange={(_, newValue) =>
                    updateForm(
                      'unlockProviders',
                      newValue.map((item) => (typeof item === 'string' ? item : item?.value || '')).filter(Boolean)
                    )
                  }
                  isOptionEqualToValue={(option, value) => {
                    const optionValue = typeof option === 'string' ? option : option?.value;
                    const selectedValue = typeof value === 'string' ? value : value?.value;
                    return optionValue === selectedValue;
                  }}
                  renderTags={(value, getTagProps) =>
                    value.map((option, index) => {
                      const { key, ...tagProps } = getTagProps({ index });
                      return (
                        <Chip
                          key={key}
                          label={formatUnlockProviderLabel(option)}
                          size="small"
                          sx={getTokenChipSx(palette.info.main)}
                          {...tagProps}
                        />
                      );
                    })
                  }
                  renderOption={(props, option, { selected }) => (
                    <Box component="li" {...props} key={option.value}>
                      <Checkbox checked={selected} size="small" sx={{ mr: 1 }} />
                      <Box>
                        <Typography variant="body2">{option.label || formatUnlockProviderLabel(option.value)}</Typography>
                        <Typography variant="caption" color="text.secondary" display="block">
                          {option.description || option.value}
                        </Typography>
                      </Box>
                    </Box>
                  )}
                  renderInput={(params) => (
                    <TextField
                      {...params}
                      label={t('nodes.nodeCheckProfiles.form.unlockProvider')}
                      placeholder={t('nodes.nodeCheckProfiles.form.unlockProviderPlaceholder')}
                      helperText={
                        unlockMetaLoading
                          ? t('nodes.nodeCheckProfiles.form.unlockProviderLoading')
                          : t('nodes.nodeCheckProfiles.form.unlockProviderHelper')
                      }
                    />
                  )}
                />
              </Stack>
            )}

            {/* TCP模式专属选项：保留速度测试结果 */}
            {form.mode === 'tcp' && (
              <FormControlLabel
                control={
                  <Switch
                    checked={form.preserveSpeedResult}
                    onChange={(e) => updateForm('preserveSpeedResult', e.target.checked)}
                    size="small"
                  />
                }
                label={
                  <Typography variant="body2">
                    {t('nodes.nodeCheckProfiles.form.preserveSpeedResult')}
                    <Typography component="span" variant="caption" color="text.secondary" sx={{ ml: 0.5 }}>
                      {t('nodes.nodeCheckProfiles.form.preserveSpeedResultHint')}
                    </Typography>
                  </Typography>
                }
              />
            )}
          </Stack>
        </ConfigSection>

        {/* ========== 性能参数 ========== */}
        <ConfigSection
          title={t('nodes.nodeCheckProfiles.form.performanceTitle')}
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
                    {t('nodes.nodeCheckProfiles.form.includeHandshake')}
                  </Typography>
                }
                sx={{ mb: 0.5, ml: 0 }}
              />
              <Typography variant="caption" color="text.secondary" component="div">
                {(form.includeHandshake ?? true) ? (
                  <>
                    <strong>{t('nodes.nodeCheckProfiles.form.handshake.enabledTitle')}</strong>
                    {t('nodes.nodeCheckProfiles.form.handshake.enabledDesc')}
                    <br />
                    {t('nodes.nodeCheckProfiles.form.handshake.enabledDetail')}
                  </>
                ) : (
                  <>
                    <strong>{t('nodes.nodeCheckProfiles.form.handshake.disabledTitle')}</strong>
                    {t('nodes.nodeCheckProfiles.form.handshake.disabledDesc')}
                    <br />
                    {t('nodes.nodeCheckProfiles.form.handshake.disabledDetail')}
                  </>
                )}
              </Typography>
            </Alert>

            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  size="small"
                  label={t('nodes.nodeCheckProfiles.form.latencyConcurrency')}
                  type="text"
                  inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                  value={form.latencyConcurrency || ''}
                  placeholder={t('nodes.nodeCheckProfiles.form.auto')}
                  onChange={(e) => {
                    const val = e.target.value;
                    if (val === '' || /^\d+$/.test(val)) {
                      updateForm('latencyConcurrency', val === '' ? 0 : Number(val));
                    }
                  }}
                  helperText={t('nodes.nodeCheckProfiles.form.dynamicHelper')}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  size="small"
                  label={t('nodes.nodeCheckProfiles.form.speedConcurrency')}
                  type="text"
                  inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                  value={form.speedConcurrency || ''}
                  placeholder={t('nodes.nodeCheckProfiles.form.auto')}
                  onChange={(e) => {
                    const val = e.target.value;
                    if (val === '' || /^\d+$/.test(val)) {
                      updateForm('speedConcurrency', val === '' ? 0 : Number(val));
                    }
                  }}
                  helperText={t('nodes.nodeCheckProfiles.form.dynamicHelper')}
                />
              </Grid>
            </Grid>
          </Stack>
        </ConfigSection>

        {/* ========== 测速范围 ========== */}
        <ConfigSection
          title={t('nodes.nodeCheckProfiles.form.scopeTitle')}
          icon={<DataUsageIcon fontSize="small" color="action" />}
          defaultExpanded={false}
          themeTokens={themeTokens}
          helperText={t('nodes.nodeCheckProfiles.form.scopeHelper')}
        >
          <Stack spacing={2}>
            <Autocomplete
              multiple
              size="small"
              options={groupOptions}
              value={form.groups}
              onChange={(_, newValue) => updateForm('groups', newValue)}
              sx={autocompleteChipSx}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label={t('nodes.nodeCheckProfiles.form.groups')}
                  placeholder={t('nodes.nodeCheckProfiles.form.groupsPlaceholder')}
                />
              )}
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
                      sx={getTokenChipSx(tagObj?.color || palette.primary.main)}
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
                      backgroundColor: option.color || palette.primary.main,
                      mr: 1
                    }}
                  />
                  {option.name}
                </Box>
              )}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label={t('nodes.nodeCheckProfiles.form.tags')}
                  placeholder={t('nodes.nodeCheckProfiles.form.tagsPlaceholder')}
                />
              )}
            />
          </Stack>
        </ConfigSection>

        {/* ========== 流量统计 ========== */}
        <ConfigSection
          title={t('nodes.nodeCheckProfiles.form.trafficTitle')}
          icon={<DataUsageIcon fontSize="small" color="action" />}
          defaultExpanded={false}
          themeTokens={themeTokens}
        >
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
                label={<Typography variant="body2">{t('nodes.nodeCheckProfiles.form.trafficByGroup')}</Typography>}
              />
              <FormControlLabel
                control={
                  <Switch
                    checked={form.trafficBySource ?? true}
                    onChange={(e) => updateForm('trafficBySource', e.target.checked)}
                    size="small"
                  />
                }
                label={<Typography variant="body2">{t('nodes.nodeCheckProfiles.form.trafficBySource')}</Typography>}
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
                    {t('nodes.nodeCheckProfiles.form.trafficByNode')}
                    <Typography component="span" variant="caption" color="error.main" sx={{ ml: 0.5 }}>
                      {t('nodes.nodeCheckProfiles.form.largeData')}
                    </Typography>
                  </Typography>
                }
              />
            </Box>
            {form.trafficByNode && (
              <Typography variant="caption" color="error.main">
                {t('nodes.nodeCheckProfiles.form.trafficByNodeWarning')}
              </Typography>
            )}
          </Stack>
        </ConfigSection>
      </DialogContent>

      <DialogActions
        sx={{
          px: 3,
          py: 2,
          borderTop: `1px solid ${panelBorder}`,
          backgroundColor: actionSurface
        }}
      >
        <Button onClick={onClose}>{t('common.cancel')}</Button>
        <Button variant="contained" onClick={handleSubmit} disabled={!form.name.trim() || submitting}>
          {submitting ? t('common.saving') : t('nodes.nodeCheckProfiles.form.saveSettings')}
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
