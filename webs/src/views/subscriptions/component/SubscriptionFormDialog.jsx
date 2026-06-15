import { useMemo, useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import Alert from '@mui/material/Alert';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';
import Radio from '@mui/material/Radio';
import RadioGroup from '@mui/material/RadioGroup';
import Typography from '@mui/material/Typography';
import Autocomplete from '@mui/material/Autocomplete';
import Tooltip from '@mui/material/Tooltip';
import InputAdornment from '@mui/material/InputAdornment';
import Grid from '@mui/material/Grid';
import ButtonGroup from '@mui/material/ButtonGroup';
import Accordion from '@mui/material/Accordion';
import AccordionSummary from '@mui/material/AccordionSummary';
import AccordionDetails from '@mui/material/AccordionDetails';
import Chip from '@mui/material/Chip';

// icons
import BuildIcon from '@mui/icons-material/Build';
import EditNoteIcon from '@mui/icons-material/EditNote';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import SettingsIcon from '@mui/icons-material/Settings';
import AccountTreeIcon from '@mui/icons-material/AccountTree';
import FilterListIcon from '@mui/icons-material/FilterList';
import TextFieldsIcon from '@mui/icons-material/TextFields';
import SecurityIcon from '@mui/icons-material/Security';
import VisibilityIcon from '@mui/icons-material/Visibility';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';

import NodeRenameBuilder from './NodeRenameBuilder';
import NodeNamePreprocessor from 'components/NodeNamePreprocessor';
import NodeNameFilter from 'components/NodeNameFilter';
import NodeTagFilter from './NodeTagFilter';
import NodeProtocolFilter from 'components/NodeProtocolFilter';
import NodeTransferBox from './NodeTransferBox';
import DeduplicationConfig from './DeduplicationConfig';
import FilterAltIcon from '@mui/icons-material/FilterAlt';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';
import { getFraudScoreIcon, QUALITY_STATUS_OPTIONS } from 'utils/fraudScore';
import { getDelayIcon, getSpeedIcon } from 'utils/nodeMetricIcons';
import { formatCountry } from 'utils/countryDisplay';
import {
  formatUnlockProviderLabel,
  getNodeUnlockSummaryDisplay,
  getUnlockProviderOptions,
  getUnlockRenameVariables,
  getUnlockRuleModeOptions,
  getUnlockStatusOptions,
  createEmptyUnlockRule
} from 'views/nodes/utils';

const normalizeCountryCode = (value) => (typeof value === 'string' ? value.trim().toUpperCase() : '');

const normalizeCountryCodeList = (values) => {
  if (!Array.isArray(values)) return [];
  return Array.from(new Set(values.map((value) => normalizeCountryCode(value)).filter(Boolean)));
};

const previewNodeName = (rule) => {
  if (!rule) return '';
  let result = rule.replace(/\$TagGroup\([^)]+\)/g, 'Fast');
  return result
    .replace(/\$Name/g, 'Hong Kong node remark')
    .replace(/\$Flag/g, '🇭🇰')
    .replace(/\$SpeedIcon/g, getSpeedIcon(1.5, 'success'))
    .replace(/\$DelayIcon/g, getDelayIcon(125, 'success'))
    .replace(/\$IpType/g, 'Native IP')
    .replace(/\$Residential/g, 'Residential IP')
    .replace(/\$FraudScoreIcon/g, getFraudScoreIcon(12, 'success'))
    .replace(/\$FraudScore/g, '12')
    .replace(/\$Unlock\([^)]+\)/g, 'Unlock-US')
    .replace(/\$LinkName/g, 'HongKong01')
    .replace(/\$LinkCountry/g, 'HK')
    .replace(/\$Speed/g, '1.50MB/s')
    .replace(/\$Delay/g, '125ms')
    .replace(/\$Group/g, 'Premium')
    .replace(/\$Source/g, 'Airport A')
    .replace(/\$DuplicateIndex/g, '1')
    .replace(/\$Index/g, '1')
    .replace(/\$Protocol/g, 'VMess')
    .replace(/\$Tags/g, 'Fast|Hong Kong');
};

const hasPreprocessRules = (value) => {
  if (!value) return false;
  try {
    const parsed = JSON.parse(value);
    return Array.isArray(parsed) && parsed.length > 0;
  } catch {
    return false;
  }
};

export default function SubscriptionFormDialog({
  open,
  isEdit,
  formData,
  setFormData,
  templates,
  scripts,
  selectorNodes,
  selectorNodesTotal,
  selectorNodesLoading,
  selectedNodesList,
  groupNodeCounts,
  allNodeTotal,
  groupOptions,
  sourceOptions,
  countryOptions,
  tagOptions,
  nodeGroupFilter,
  setNodeGroupFilter,
  nodeSourceFilter,
  setNodeSourceFilter,
  nodeSearchQuery,
  setNodeSearchQuery,
  nodeCountryFilter,
  setNodeCountryFilter,
  checkedAvailable,
  checkedSelected,
  mobileTab,
  setMobileTab,
  selectedNodeSearch,
  setSelectedNodeSearch,
  namingMode,
  setNamingMode,
  onClose,
  onSubmit,
  onPreview,
  previewLoading,
  onAddNode,
  onRemoveNode,
  onAddAllVisible,
  onRemoveAll,
  onToggleAvailable,
  onToggleSelected,
  onAddChecked,
  onRemoveChecked,
  onToggleAllAvailable,
  onToggleAllSelected
}) {
  const { t } = useTranslation();
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const matchDownMd = useMediaQuery(theme.breakpoints.down('md'));
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);
  const [countryWhitelistInput, setCountryWhitelistInput] = useState('');
  const [countryBlacklistInput, setCountryBlacklistInput] = useState('');

  const [expandedPanels, setExpandedPanels] = useState({
    basic: true,
    nodes: true,
    filter: false,
    dedup: false,
    naming: false,
    advanced: false
  });

  const handlePanelChange = (panel) => (event, isExpanded) => {
    void event;
    setExpandedPanels((prev) => ({
      ...prev,
      [panel]: isExpanded
    }));
  };

  const clashTemplates = useMemo(() => {
    return templates.filter((t) => !t.category || t.category === 'clash');
  }, [templates]);

  const surgeTemplates = useMemo(() => {
    return templates.filter((t) => t.category === 'surge');
  }, [templates]);

  const unlockProviderOptions = getUnlockProviderOptions();
  const unlockRenameVariables = getUnlockRenameVariables();
  const unlockRules = useMemo(() => (Array.isArray(formData.unlockRules) ? formData.unlockRules : []), [formData.unlockRules]);
  const normalizedCountryOptions = useMemo(() => normalizeCountryCodeList(countryOptions), [countryOptions]);
  const normalizedCountryWhitelist = useMemo(() => normalizeCountryCodeList(formData.CountryWhitelist), [formData.CountryWhitelist]);
  const normalizedCountryBlacklist = useMemo(() => normalizeCountryCodeList(formData.CountryBlacklist), [formData.CountryBlacklist]);

  const updateCountryFilterField = (field, values) => {
    setFormData({ ...formData, [field]: normalizeCountryCodeList(values) });
  };

  const handleCountryFilterKeyDown = (event, field, inputValue, setInputValue) => {
    if (event.key !== 'Enter') return;
    const normalizedInput = normalizeCountryCode(inputValue);
    if (!normalizedInput) return;
    event.preventDefault();
    updateCountryFilterField(field, [...normalizeCountryCodeList(formData[field]), normalizedInput]);
    setInputValue('');
  };

  useEffect(() => {
    if (open && groupOptions && formData.selectedGroups?.length > 0) {
      const validGroups = formData.selectedGroups.filter((g) => groupOptions.includes(g));
      if (validGroups.length !== formData.selectedGroups.length) {
        setFormData((prev) => ({ ...prev, selectedGroups: validGroups }));
      }
    }
  }, [groupOptions, open, formData.selectedGroups, setFormData]);

  const normalizedSelectorNodes = useMemo(() => selectorNodes || [], [selectorNodes]);
  const selectorLoadingText = selectorNodesLoading ? t('subscriptions.form.nodes.loading') : '';

  const availableNodes = useMemo(() => {
    return normalizedSelectorNodes.filter((node) => {
      if (nodeGroupFilter !== 'all' && node.Group !== nodeGroupFilter) return false;
      if (nodeSourceFilter !== 'all' && node.Source !== nodeSourceFilter) return false;
      if (nodeSearchQuery) {
        const query = nodeSearchQuery.toLowerCase();
        const unlockSummary = getNodeUnlockSummaryDisplay(node, { limit: 4 });
        const unlockText = unlockSummary?.items
          ?.map((item) => [item.providerLabel, item.statusLabel, item.region, item.reason, item.detail].filter(Boolean).join(' '))
          .join(' ')
          .toLowerCase();
        if (!node.Name?.toLowerCase().includes(query) && !node.Group?.toLowerCase().includes(query) && !unlockText?.includes(query)) {
          return false;
        }
      }
      if (nodeCountryFilter.length > 0) {
        if (!node.LinkCountry || !nodeCountryFilter.includes(node.LinkCountry)) {
          return false;
        }
      }
      return !formData.selectedNodes.includes(node.ID);
    });
  }, [normalizedSelectorNodes, nodeGroupFilter, nodeSourceFilter, nodeSearchQuery, nodeCountryFilter, formData.selectedNodes]);

  const selectorNodesCount = selectorNodesTotal || availableNodes.length;

  const filterRulesCount = useMemo(() => {
    let count = 0;
    if (formData.DelayTime > 0) count++;
    if (formData.MinSpeed > 0) count++;
    if (formData.CountryWhitelist?.length > 0) count++;
    if (formData.CountryBlacklist?.length > 0) count++;
    if (formData.tagWhitelist) count++;
    if (formData.tagBlacklist) count++;
    if (formData.protocolWhitelist) count++;
    if (formData.protocolBlacklist) count++;
    if (formData.nodeNameWhitelist) count++;
    if (formData.nodeNameBlacklist) count++;
    if (formData.MaxFraudScore > 0) count++;
    if (formData.QualityStatus) count++;
    if (formData.ResidentialType) count++;
    if (formData.IPType) count++;
    if (unlockRules.some((rule) => rule.provider || rule.status || rule.keyword)) count++;
    return count;
  }, [formData, unlockRules]);

  const updateUnlockRule = (index, patch) => {
    const nextRules = unlockRules.map((rule, ruleIndex) => (ruleIndex === index ? { ...rule, ...patch } : rule));
    setFormData({ ...formData, unlockRules: nextRules });
  };

  const addUnlockRule = () => {
    setFormData({ ...formData, unlockRules: [...unlockRules, createEmptyUnlockRule()] });
  };

  const removeUnlockRule = (index) => {
    const nextRules = unlockRules.filter((_, ruleIndex) => ruleIndex !== index);
    setFormData({ ...formData, unlockRules: nextRules });
  };

  const advancedSettingsCount = useMemo(() => {
    let count = 0;
    if (formData.selectedScripts?.length > 0) count++;
    if (formData.IPWhitelist) count++;
    if (formData.IPBlacklist) count++;
    return count;
  }, [formData]);

  const helperCaptionSx = {
    display: 'block',
    mt: 1,
    color: secondaryText
  };

  const insetHighlight = isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}` : 'none';
  const accordionHoverBorder = withAlpha(palette.primary.main, isDark ? 0.3 : 0.16);
  const accordionExpandedBorder = withAlpha(palette.primary.main, isDark ? 0.36 : 0.2);
  const accordionSummaryHoverSurface = isDark ? withAlpha(palette.background.paper, 0.22) : withAlpha(palette.primary.main, 0.04);
  const accordionDetailsSurface = isDark
    ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.08)} 0%, ${dialogSurface} 100%)`
    : dialogSurface;
  const helperPanelSx = {
    mt: 1,
    p: 1.5,
    bgcolor: nestedPanelSurface,
    borderRadius: 1.5,
    border: '1px solid',
    borderColor: panelBorder,
    boxShadow: insetHighlight
  };

  const accordionSx = {
    mb: 1.5,
    '&:before': { display: 'none' },
    bgcolor: dialogSurface,
    backgroundImage: isDark ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.12)} 0%, ${dialogSurface} 100%)` : 'none',
    border: '1px solid',
    borderColor: panelBorder,
    boxShadow: insetHighlight,
    borderRadius: '12px !important',
    overflow: 'hidden',
    transition: 'border-color 0.2s ease, box-shadow 0.2s ease',
    '&:hover': {
      borderColor: accordionHoverBorder
    },
    '&.Mui-expanded': {
      margin: '0 0 12px 0',
      borderColor: accordionExpandedBorder
    }
  };

  const accordionSummarySx = {
    minHeight: 56,
    px: 0.5,
    color: primaryText,
    bgcolor: mutedPanelSurface,
    transition: 'background-color 0.2s ease, border-color 0.2s ease',
    '& .MuiAccordionSummary-expandIconWrapper': {
      color: tertiaryText,
      transition: 'color 0.2s ease, transform 0.2s ease'
    },
    '&:hover': {
      bgcolor: accordionSummaryHoverSurface,
      '& .MuiAccordionSummary-expandIconWrapper': {
        color: secondaryText
      }
    },
    '&.Mui-expanded': {
      minHeight: 56,
      bgcolor: nestedPanelSurface,
      borderBottom: '1px solid',
      borderColor: panelBorder,
      '& .MuiAccordionSummary-expandIconWrapper': {
        color: primaryText
      }
    },
    '& .MuiAccordionSummary-content': {
      alignItems: 'center',
      gap: 1.5,
      '&.Mui-expanded': {
        margin: '12px 0'
      }
    }
  };

  const accordionDetailsSx = {
    px: matchDownMd ? 2 : 2.5,
    py: 2.25,
    bgcolor: dialogSurface,
    backgroundImage: accordionDetailsSurface
  };

  const autocompleteChipSx = {
    '& .MuiAutocomplete-tag': {
      bgcolor: isDark ? withAlpha(palette.primary.main, 0.16) : undefined,
      color: isDark ? primaryText : undefined,
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

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="lg"
      fullWidth
      fullScreen={matchDownMd}
      slotProps={{
        paper: {
          sx: matchDownMd
            ? {
                borderRadius: 0,
                border: '1px solid',
                borderColor: panelBorder,
                bgcolor: dialogSurface,
                backgroundImage: dialogSurfaceGradient
              }
            : {
                borderRadius: 2.5,
                border: '1px solid',
                borderColor: panelBorder,
                bgcolor: dialogSurface,
                backgroundImage: dialogSurfaceGradient
              }
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
        {isEdit ? t('subscriptions.form.title.edit') : t('subscriptions.form.title.add')}
      </DialogTitle>
      <DialogContent sx={{ px: matchDownMd ? 2 : 3, pt: 2.5, pb: 2, bgcolor: 'transparent' }}>
        <Box sx={{ mt: 1 }}>
          {/* ========== {t('subscriptions.form.sections.basic')} ========== */}
          <Accordion expanded={expandedPanels.basic} onChange={handlePanelChange('basic')} sx={accordionSx}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={accordionSummarySx}>
              <SettingsIcon color="primary" />
              <Typography variant="subtitle1" fontWeight={600}>
                {t('subscriptions.form.sections.basic')}
              </Typography>
            </AccordionSummary>
            <AccordionDetails sx={accordionDetailsSx}>
              <Stack spacing={2.5}>
                <TextField
                  fullWidth
                  label={t('subscriptions.form.basic.name')}
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                />

                <Grid container spacing={2}>
                  <Grid item xs={12} sm={6}>
                    <FormControl fullWidth>
                      <InputLabel shrink>{t('subscriptions.form.basic.clashTemplate')}</InputLabel>
                      <Select
                        variant={'outlined'}
                        value={formData.clash}
                        label={t('subscriptions.form.basic.clashTemplate')}
                        onChange={(e) => setFormData({ ...formData, clash: e.target.value })}
                        displayEmpty
                      >
                        <MenuItem value="">
                          <Typography color="text.secondary">{t('subscriptions.form.basic.clashTemplateNone')}</Typography>
                        </MenuItem>
                        {clashTemplates.map((t) => (
                          <MenuItem key={t.file} value={`./template/${t.file}`}>
                            {t.file}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                    {clashTemplates.length === 0 && (
                      <Alert severity="warning" sx={{ mt: 1 }}>
                        <Typography variant="caption">{t('subscriptions.form.basic.clashTemplateHelper')}</Typography>
                      </Alert>
                    )}
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <FormControl fullWidth>
                      <InputLabel shrink>{t('subscriptions.form.basic.surgeTemplate')}</InputLabel>
                      <Select
                        value={formData.surge}
                        label={t('subscriptions.form.basic.surgeTemplate')}
                        onChange={(e) => setFormData({ ...formData, surge: e.target.value })}
                        displayEmpty
                      >
                        <MenuItem value="">
                          <Typography color="text.secondary">{t('subscriptions.form.basic.clashTemplateNone')}</Typography>
                        </MenuItem>
                        {surgeTemplates.map((t) => (
                          <MenuItem key={t.file} value={`./template/${t.file}`}>
                            {t.file}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                    {surgeTemplates.length === 0 && (
                      <Alert severity="warning" sx={{ mt: 1 }}>
                        <Typography variant="caption">{t('subscriptions.form.basic.surgeTemplateHelper')}</Typography>
                      </Alert>
                    )}
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <TextField
                      fullWidth
                      label={t('subscriptions.form.basic.updateInterval')}
                      type="text"
                      slotProps={{ htmlInput: { inputMode: 'numeric', pattern: '[0-9]*', max: 8760 } }}
                      value={formData.UpdateInterval}
                      onChange={(e) => {
                        const val = e.target.value;
                        if (val === '' || /^\d+$/.test(val)) {
                          setFormData({ ...formData, UpdateInterval: val === '' ? '' : Number(val) });
                        }
                      }}
                      onBlur={(e) => {
                        const val = Math.min(8760, Math.max(0, Number(e.target.value) || 0));
                        setFormData({ ...formData, UpdateInterval: val });
                      }}
                      helperText={t('subscriptions.form.basic.updateIntervalHelper')}
                    />
                  </Grid>
                </Grid>

                <Stack direction="row" spacing={2} flexWrap="wrap">
                  <FormControlLabel
                    control={<Checkbox checked={formData.udp} onChange={(e) => setFormData({ ...formData, udp: e.target.checked })} />}
                    label={t('subscriptions.form.basic.forceUdp')}
                  />
                  <FormControlLabel
                    control={<Checkbox checked={formData.cert} onChange={(e) => setFormData({ ...formData, cert: e.target.checked })} />}
                    label={t('subscriptions.form.basic.skipCertVerify')}
                  />
                  <Tooltip title={t('subscriptions.form.basic.replaceHostTooltip')} placement="top" arrow>
                    <FormControlLabel
                      control={
                        <Checkbox
                          checked={formData.replaceServerWithHost}
                          onChange={(e) => setFormData({ ...formData, replaceServerWithHost: e.target.checked })}
                        />
                      }
                      label={t('subscriptions.form.basic.replaceHost')}
                    />
                  </Tooltip>
                  <Tooltip title={t('subscriptions.form.basic.realtimeUsageTooltip')} placement="top" arrow>
                    <FormControlLabel
                      control={
                        <Checkbox
                          checked={formData.refreshUsageOnRequest}
                          onChange={(e) => setFormData({ ...formData, refreshUsageOnRequest: e.target.checked })}
                        />
                      }
                      label={t('subscriptions.form.basic.realtimeUsage')}
                    />
                  </Tooltip>
                </Stack>
              </Stack>
            </AccordionDetails>
          </Accordion>

          {/* ========== {t('subscriptions.form.sections.nodeSelection')} ========== */}
          <Accordion expanded={expandedPanels.nodes} onChange={handlePanelChange('nodes')} sx={accordionSx}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={accordionSummarySx}>
              <AccountTreeIcon color="primary" />
              <Typography variant="subtitle1" fontWeight={600}>
                {t('subscriptions.form.sections.nodeSelection')}
              </Typography>
              {!expandedPanels.nodes && (formData.selectedNodes.length > 0 || formData.selectedGroups.length > 0) && (
                <Chip
                  size="small"
                  label={t('subscriptions.form.nodeSelection.summary', {
                    nodeCount: formData.selectedNodes.length,
                    groupCount: formData.selectedGroups.length
                  })}
                  color="primary"
                  variant="outlined"
                  sx={{ ml: 1 }}
                />
              )}
            </AccordionSummary>
            <AccordionDetails sx={accordionDetailsSx}>
              <Stack spacing={2.5}>
                {/* Selection Mode */}
                <Box>
                  <RadioGroup
                    row
                    value={formData.selectionMode}
                    onChange={(e) => setFormData({ ...formData, selectionMode: e.target.value })}
                  >
                    <FormControlLabel value="nodes" control={<Radio />} label={t('subscriptions.form.nodeSelection.modeManual')} />
                    <FormControlLabel value="groups" control={<Radio />} label={t('subscriptions.form.nodeSelection.modeDynamic')} />
                    <FormControlLabel value="mixed" control={<Radio />} label={t('subscriptions.form.nodeSelection.modeMixed')} />
                  </RadioGroup>
                  <Typography variant="caption" sx={helperCaptionSx}>
                    {formData.selectionMode === 'nodes' && t('subscriptions.form.nodeSelection.modeManualHelper')}
                    {formData.selectionMode === 'groups' && t('subscriptions.form.nodeSelection.modeDynamicHelper')}
                    {formData.selectionMode === 'mixed' && t('subscriptions.form.nodeSelection.modeMixedHelper')}
                  </Typography>
                </Box>

                {/* Group Selection */}
                {(formData.selectionMode === 'groups' || formData.selectionMode === 'mixed') && (
                  <Autocomplete
                    multiple
                    options={groupOptions}
                    value={formData.selectedGroups}
                    onChange={(e, newValue) => setFormData({ ...formData, selectedGroups: newValue })}
                    sx={autocompleteChipSx}
                    renderInput={(params) => <TextField {...params} label={t('subscriptions.form.nodeSelection.groupSelect')} />}
                    renderOption={(props, option) => (
                      <li {...props}>
                        {t('subscriptions.form.nodeSelection.groupOption', { name: option, count: groupNodeCounts[option] || 0 })}
                      </li>
                    )}
                  />
                )}

                {/* {t('subscriptions.form.sections.nodeSelection')} */}
                {(formData.selectionMode === 'nodes' || formData.selectionMode === 'mixed') && (
                  <>
                    <Grid container spacing={2}>
                      <Grid item xs={6} sm={3}>
                        <FormControl fullWidth size="small">
                          <InputLabel>{t('subscriptions.form.nodeSelection.nodeGroupFilter')}</InputLabel>
                          <Select
                            value={nodeGroupFilter}
                            label={t('subscriptions.form.nodeSelection.nodeGroupFilter')}
                            onChange={(e) => setNodeGroupFilter(e.target.value)}
                          >
                            <MenuItem value="all">{t('subscriptions.form.nodeSelection.nodeGroupAll', { count: allNodeTotal })}</MenuItem>
                            {groupOptions.map((g) => (
                              <MenuItem key={g} value={g}>
                                {g} ({groupNodeCounts[g] || 0})
                              </MenuItem>
                            ))}
                          </Select>
                        </FormControl>
                      </Grid>
                      <Grid item xs={6} sm={3}>
                        <FormControl fullWidth size="small">
                          <InputLabel>{t('subscriptions.form.nodeSelection.nodeSourceFilter')}</InputLabel>
                          <Select
                            value={nodeSourceFilter}
                            label={t('subscriptions.form.nodeSelection.nodeSourceFilter')}
                            onChange={(e) => setNodeSourceFilter(e.target.value)}
                          >
                            <MenuItem value="all">{t('subscriptions.form.nodeSelection.nodeSourceAll')}</MenuItem>
                            {sourceOptions.map((s) => (
                              <MenuItem key={s} value={s}>
                                {s}
                              </MenuItem>
                            ))}
                          </Select>
                        </FormControl>
                      </Grid>
                      <Grid item xs={6} sm={3}>
                        <Autocomplete
                          multiple
                          size="small"
                          options={countryOptions}
                          value={nodeCountryFilter}
                          onChange={(e, newValue) => setNodeCountryFilter(newValue)}
                          getOptionLabel={(option) => formatCountry(option)}
                          renderInput={(params) => (
                            <TextField
                              {...params}
                              label={t('subscriptions.form.nodeSelection.nodeCountryFilter')}
                              sx={{
                                '& .MuiInputBase-root': {
                                  paddingRight: '65px !important'
                                }
                              }}
                            />
                          )}
                          renderOption={(props, option) => <li {...props}>{formatCountry(option)}</li>}
                          renderTags={(value, getTagProps) =>
                            value.map((option, index) => {
                              const { key, ...tagProps } = getTagProps({ index });
                              return (
                                <Chip
                                  key={key}
                                  label={formatCountry(option)}
                                  size="small"
                                  sx={{
                                    bgcolor: isDark ? withAlpha(palette.primary.main, 0.12) : undefined,
                                    borderColor: isDark ? withAlpha(palette.primary.main, 0.3) : undefined
                                  }}
                                  {...tagProps}
                                />
                              );
                            })
                          }
                          limitTags={2}
                        />
                      </Grid>
                      <Grid item xs={6} sm={3}>
                        <TextField
                          fullWidth
                          size="small"
                          label={t('subscriptions.form.nodeSelection.nodeSearch')}
                          value={nodeSearchQuery}
                          onChange={(e) => setNodeSearchQuery(e.target.value)}
                        />
                      </Grid>
                    </Grid>

                    <NodeTransferBox
                      availableNodes={availableNodes}
                      selectedNodes={formData.selectedNodes}
                      selectedNodesList={selectedNodesList}
                      checkedAvailable={checkedAvailable}
                      checkedSelected={checkedSelected}
                      selectedNodeSearch={selectedNodeSearch}
                      onSelectedNodeSearchChange={setSelectedNodeSearch}
                      selectorNodesTotal={selectorNodesCount}
                      selectorNodesLoading={selectorNodesLoading}
                      mobileTab={mobileTab}
                      onMobileTabChange={setMobileTab}
                      matchDownMd={matchDownMd}
                      onAddNode={onAddNode}
                      onRemoveNode={onRemoveNode}
                      onAddAllVisible={onAddAllVisible}
                      onRemoveAll={onRemoveAll}
                      onToggleAvailable={onToggleAvailable}
                      onToggleSelected={onToggleSelected}
                      onAddChecked={onAddChecked}
                      onRemoveChecked={onRemoveChecked}
                      onToggleAllAvailable={onToggleAllAvailable}
                      onToggleAllSelected={onToggleAllSelected}
                    />
                    {selectorLoadingText && (
                      <Alert severity="info" variant="outlined" sx={{ mt: 1 }}>
                        {selectorLoadingText}
                      </Alert>
                    )}
                  </>
                )}
              </Stack>
            </AccordionDetails>
          </Accordion>

          {/* ========== {t('subscriptions.form.sections.nodeFilter')} ========== */}
          <Accordion expanded={expandedPanels.filter} onChange={handlePanelChange('filter')} sx={accordionSx}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={accordionSummarySx}>
              <FilterListIcon color="primary" />
              <Typography variant="subtitle1" fontWeight={600}>
                {t('subscriptions.form.sections.nodeFilter')}
              </Typography>
              {!expandedPanels.filter && filterRulesCount > 0 && (
                <Chip
                  size="small"
                  label={t('subscriptions.form.nodeFilter.activeRules', { count: filterRulesCount })}
                  color="warning"
                  variant="outlined"
                  sx={{ ml: 1 }}
                />
              )}
            </AccordionSummary>
            <AccordionDetails sx={accordionDetailsSx}>
              <Stack spacing={2.5}>
                {/* Delay and Speed Filter */}
                <Grid container spacing={2}>
                  <Grid item xs={12} sm={6}>
                    <TextField
                      fullWidth
                      label={t('subscriptions.form.nodeFilter.maxDelay')}
                      type="text"
                      inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                      value={formData.DelayTime}
                      onChange={(e) => {
                        const val = e.target.value;
                        if (val === '' || /^\d+$/.test(val)) {
                          setFormData({ ...formData, DelayTime: val === '' ? '' : Number(val) });
                        }
                      }}
                      onBlur={(e) => {
                        const val = Math.max(0, Number(e.target.value) || 0);
                        setFormData({ ...formData, DelayTime: val });
                      }}
                      InputProps={{ endAdornment: <InputAdornment position="end">ms</InputAdornment> }}
                      helperText={t('subscriptions.form.nodeFilter.maxDelayHelper')}
                    />
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <TextField
                      fullWidth
                      label={t('subscriptions.form.nodeFilter.minSpeed')}
                      type="text"
                      inputProps={{ inputMode: 'numeric', pattern: '[0-9]*\\.?[0-9]*' }}
                      value={formData.MinSpeed}
                      onChange={(e) => {
                        const val = e.target.value;
                        if (val === '' || /^\d*\.?\d*$/.test(val)) {
                          setFormData({ ...formData, MinSpeed: val === '' ? '' : val });
                        }
                      }}
                      onBlur={(e) => {
                        const val = Math.max(0, parseFloat(e.target.value) || 0);
                        setFormData({ ...formData, MinSpeed: val });
                      }}
                      InputProps={{ endAdornment: <InputAdornment position="end">MB/s</InputAdornment> }}
                      helperText={t('subscriptions.form.nodeFilter.minSpeedHelper')}
                    />
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <TextField
                      fullWidth
                      label={t('subscriptions.form.nodeFilter.maxFraudScore')}
                      type="text"
                      inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                      value={formData.MaxFraudScore}
                      onChange={(e) => {
                        const val = e.target.value;
                        if (val === '' || /^\d+$/.test(val)) {
                          setFormData({ ...formData, MaxFraudScore: val === '' ? '' : Number(val) });
                        }
                      }}
                      onBlur={(e) => {
                        const val = Math.max(0, Number(e.target.value) || 0);
                        setFormData({ ...formData, MaxFraudScore: val });
                      }}
                      helperText={t('subscriptions.form.nodeFilter.maxFraudScoreHelper')}
                    />
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <FormControl fullWidth>
                      <InputLabel>{t('subscriptions.form.nodeFilter.qualityStatus')}</InputLabel>
                      <Select
                        value={formData.QualityStatus || ''}
                        label={t('subscriptions.form.nodeFilter.qualityStatus')}
                        onChange={(e) => setFormData({ ...formData, QualityStatus: e.target.value })}
                      >
                        {QUALITY_STATUS_OPTIONS.map((option) => (
                          <MenuItem key={option.value} value={option.value}>
                            {option.label}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                    <Typography variant="caption" sx={helperCaptionSx}>
                      {t('subscriptions.form.nodeFilter.qualityStatusHelper')}
                    </Typography>
                  </Grid>
                  <Grid item xs={12}>
                    <Stack spacing={1.5}>
                      <Typography variant="subtitle2">{t('subscriptions.form.nodeFilter.unlockRules')}</Typography>
                      <Alert severity="info" variant="outlined">
                        {t('subscriptions.form.nodeFilter.unlockRulesDesc')}
                      </Alert>
                      <Grid container spacing={1.5} alignItems="center">
                        <Grid item xs={12} md={4}>
                          <FormControl fullWidth size="small">
                            <InputLabel>{t('subscriptions.form.nodeFilter.ruleRelation')}</InputLabel>
                            <Select
                              value={formData.UnlockRuleMode || 'or'}
                              label={t('subscriptions.form.nodeFilter.ruleRelation')}
                              onChange={(e) => setFormData({ ...formData, UnlockRuleMode: e.target.value })}
                            >
                              {getUnlockRuleModeOptions().map((option) => (
                                <MenuItem key={option.value} value={option.value}>
                                  {option.label}
                                </MenuItem>
                              ))}
                            </Select>
                          </FormControl>
                        </Grid>
                        <Grid item xs={12} md={8}>
                          <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                            {formData.UnlockRuleMode === 'and'
                              ? t('subscriptions.form.nodeFilter.ruleRelationAnd')
                              : t('subscriptions.form.nodeFilter.ruleRelationOr')}
                          </Typography>
                        </Grid>
                      </Grid>
                      {unlockRules.length > 0 ? (
                        unlockRules.map((rule, index) => (
                          <Grid container spacing={1.5} key={`unlock-rule-${index}`} alignItems="flex-start">
                            <Grid item xs={12} md={4}>
                              <Autocomplete
                                options={unlockProviderOptions}
                                value={unlockProviderOptions.find((item) => item.value === rule.provider) || null}
                                onChange={(_, newValue) => updateUnlockRule(index, { provider: newValue?.value || '' })}
                                getOptionLabel={(option) => option?.label || formatUnlockProviderLabel(option?.value || '')}
                                renderOption={(props, option) => (
                                  <li {...props} key={option.value}>
                                    <Box>
                                      <Typography variant="body2">{option.label}</Typography>
                                      <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                                        {option.description || option.value}
                                      </Typography>
                                    </Box>
                                  </li>
                                )}
                                renderInput={(params) => (
                                  <TextField {...params} label="Provider" helperText={t('subscriptions.form.nodeFilter.providerHelper')} />
                                )}
                              />
                            </Grid>
                            <Grid item xs={12} md={3}>
                              <FormControl fullWidth>
                                <InputLabel>{t('subscriptions.form.nodeFilter.status')}</InputLabel>
                                <Select
                                  value={rule.status || ''}
                                  label={t('subscriptions.form.nodeFilter.status')}
                                  onChange={(e) => updateUnlockRule(index, { status: e.target.value })}
                                >
                                  {getUnlockStatusOptions(true).map((option) => (
                                    <MenuItem key={option.value || 'all'} value={option.value}>
                                      {option.label}
                                    </MenuItem>
                                  ))}
                                </Select>
                              </FormControl>
                            </Grid>
                            <Grid item xs={12} md={4}>
                              <TextField
                                fullWidth
                                label={t('subscriptions.form.nodeFilter.keyword')}
                                value={rule.keyword || ''}
                                onChange={(e) => updateUnlockRule(index, { keyword: e.target.value })}
                                helperText={t('subscriptions.form.nodeFilter.keywordHelper')}
                              />
                            </Grid>
                            <Grid item xs={12} md={1}>
                              <Button
                                fullWidth
                                color="error"
                                variant="outlined"
                                startIcon={<DeleteOutlineIcon />}
                                onClick={() => removeUnlockRule(index)}
                              >
                                {t('subscriptions.form.nodeFilter.delete')}
                              </Button>
                            </Grid>
                          </Grid>
                        ))
                      ) : (
                        <Alert severity="info" variant="outlined">
                          {t('subscriptions.form.nodeFilter.unlockInactive')}
                        </Alert>
                      )}
                      <Box>
                        <Button startIcon={<AddIcon />} variant="outlined" onClick={addUnlockRule}>
                          {t('subscriptions.form.nodeFilter.addUnlockRule')}
                        </Button>
                      </Box>
                    </Stack>
                  </Grid>
                </Grid>

                {/* Landing IP Country Filter */}
                <Grid container spacing={2}>
                  <Grid item xs={12} sm={6}>
                    <Autocomplete
                      multiple
                      freeSolo
                      options={normalizedCountryOptions}
                      value={normalizedCountryWhitelist}
                      inputValue={countryWhitelistInput}
                      onInputChange={(event, newInputValue) => {
                        void event;
                        setCountryWhitelistInput(newInputValue);
                      }}
                      onChange={(event, newValue) => {
                        void event;
                        updateCountryFilterField('CountryWhitelist', newValue);
                      }}
                      getOptionLabel={(option) => formatCountry(normalizeCountryCode(option))}
                      isOptionEqualToValue={(option, value) => normalizeCountryCode(option) === normalizeCountryCode(value)}
                      renderInput={(params) => (
                        <TextField
                          {...params}
                          label={t('subscriptions.form.nodeFilter.countryWhitelist')}
                          helperText={t('subscriptions.form.nodeFilter.countryWhitelistHelper')}
                          onKeyDown={(event) =>
                            handleCountryFilterKeyDown(event, 'CountryWhitelist', countryWhitelistInput, setCountryWhitelistInput)
                          }
                        />
                      )}
                      renderOption={(props, option) => <li {...props}>{formatCountry(normalizeCountryCode(option))}</li>}
                    />
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <Autocomplete
                      multiple
                      freeSolo
                      options={normalizedCountryOptions}
                      value={normalizedCountryBlacklist}
                      inputValue={countryBlacklistInput}
                      onInputChange={(event, newInputValue) => {
                        void event;
                        setCountryBlacklistInput(newInputValue);
                      }}
                      onChange={(event, newValue) => {
                        void event;
                        updateCountryFilterField('CountryBlacklist', newValue);
                      }}
                      getOptionLabel={(option) => formatCountry(normalizeCountryCode(option))}
                      isOptionEqualToValue={(option, value) => normalizeCountryCode(option) === normalizeCountryCode(value)}
                      renderInput={(params) => (
                        <TextField
                          {...params}
                          label={t('subscriptions.form.nodeFilter.countryBlacklist')}
                          helperText={t('subscriptions.form.nodeFilter.countryWhitelistHelper')}
                          onKeyDown={(event) =>
                            handleCountryFilterKeyDown(event, 'CountryBlacklist', countryBlacklistInput, setCountryBlacklistInput)
                          }
                        />
                      )}
                      renderOption={(props, option) => <li {...props}>{formatCountry(normalizeCountryCode(option))}</li>}
                    />
                  </Grid>
                </Grid>

                <Grid container spacing={2}>
                  <Grid item xs={12} sm={6}>
                    <FormControl fullWidth>
                      <InputLabel>{t('subscriptions.form.nodeFilter.residential')}</InputLabel>
                      <Select
                        value={formData.ResidentialType || ''}
                        label={t('subscriptions.form.nodeFilter.residential')}
                        onChange={(e) => setFormData({ ...formData, ResidentialType: e.target.value })}
                      >
                        <MenuItem value="">{t('subscriptions.form.nodeFilter.all')}</MenuItem>
                        <MenuItem value="residential">{t('subscriptions.form.nodeFilter.residentialIp')}</MenuItem>
                        <MenuItem value="datacenter">{t('subscriptions.form.nodeFilter.datacenterIp')}</MenuItem>
                        <MenuItem value="untested">{t('subscriptions.form.nodeFilter.untested')}</MenuItem>
                      </Select>
                    </FormControl>
                    <Typography variant="caption" sx={helperCaptionSx}>
                      {t('subscriptions.form.nodeFilter.residentialHelper')}
                    </Typography>
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <FormControl fullWidth>
                      <InputLabel>{t('subscriptions.form.nodeFilter.ipType')}</InputLabel>
                      <Select
                        value={formData.IPType || ''}
                        label={t('subscriptions.form.nodeFilter.ipType')}
                        onChange={(e) => setFormData({ ...formData, IPType: e.target.value })}
                      >
                        <MenuItem value="">{t('subscriptions.form.nodeFilter.all')}</MenuItem>
                        <MenuItem value="native">{t('subscriptions.form.nodeFilter.nativeIp')}</MenuItem>
                        <MenuItem value="broadcast">{t('subscriptions.form.nodeFilter.broadcastIp')}</MenuItem>
                        <MenuItem value="untested">{t('subscriptions.form.nodeFilter.untested')}</MenuItem>
                      </Select>
                    </FormControl>
                    <Typography variant="caption" sx={helperCaptionSx}>
                      {t('subscriptions.form.nodeFilter.ipTypeHelper')}
                    </Typography>
                  </Grid>
                </Grid>

                {/* Node Tag Filter */}
                <NodeTagFilter
                  tagOptions={tagOptions}
                  whitelistValue={formData.tagWhitelist}
                  blacklistValue={formData.tagBlacklist}
                  onWhitelistChange={(tags) => setFormData({ ...formData, tagWhitelist: tags })}
                  onBlacklistChange={(tags) => setFormData({ ...formData, tagBlacklist: tags })}
                />

                {/* Protocol Type Filter */}
                <NodeProtocolFilter
                  protocolOptions={formData.protocolOptions || []}
                  whitelistValue={formData.protocolWhitelist}
                  blacklistValue={formData.protocolBlacklist}
                  onWhitelistChange={(protocols) => setFormData({ ...formData, protocolWhitelist: protocols })}
                  onBlacklistChange={(protocols) => setFormData({ ...formData, protocolBlacklist: protocols })}
                />

                {/* Node Name Filter */}
                <NodeNameFilter
                  whitelistValue={formData.nodeNameWhitelist}
                  blacklistValue={formData.nodeNameBlacklist}
                  onWhitelistChange={(rules) => setFormData({ ...formData, nodeNameWhitelist: rules })}
                  onBlacklistChange={(rules) => setFormData({ ...formData, nodeNameBlacklist: rules })}
                />
              </Stack>
            </AccordionDetails>
          </Accordion>

          {/* ========== {t('subscriptions.form.sections.nodeDeduplication')} ========== */}
          <Accordion expanded={expandedPanels.dedup} onChange={handlePanelChange('dedup')} sx={accordionSx}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={accordionSummarySx}>
              <FilterAltIcon color="primary" />
              <Typography variant="subtitle1" fontWeight={600}>
                {t('subscriptions.form.sections.nodeDeduplication')}
                <Chip size="small" label="Beta" color="error" variant="outlined" sx={{ ml: 1 }} />
              </Typography>
              {!expandedPanels.dedup && formData.deduplicationRule && (
                <Chip
                  size="small"
                  label={t('subscriptions.form.nodeDeduplication.configured')}
                  color="success"
                  variant="outlined"
                  sx={{ ml: 1 }}
                />
              )}
            </AccordionSummary>
            <AccordionDetails sx={accordionDetailsSx}>
              <DeduplicationConfig
                value={formData.deduplicationRule || ''}
                onChange={(rule) => setFormData({ ...formData, deduplicationRule: rule })}
              />
            </AccordionDetails>
          </Accordion>

          {/* ========== Name Processing ========== */}
          <Accordion expanded={expandedPanels.naming} onChange={handlePanelChange('naming')} sx={accordionSx}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={accordionSummarySx}>
              <TextFieldsIcon color="primary" />
              <Typography variant="subtitle1" fontWeight={600}>
                {t('subscriptions.form.sections.nameProcessing')}
              </Typography>
              {!expandedPanels.naming && (hasPreprocessRules(formData.nodeNamePreprocess) || formData.nodeNameRule) && (
                <Chip
                  size="small"
                  label={t('subscriptions.form.nodeDeduplication.configured')}
                  color="info"
                  variant="outlined"
                  sx={{ ml: 1 }}
                />
              )}
            </AccordionSummary>
            <AccordionDetails sx={accordionDetailsSx}>
              <Stack spacing={2.5}>
                {/* Original Name Preprocessing */}
                <NodeNamePreprocessor
                  value={formData.nodeNamePreprocess}
                  onChange={(rules) => setFormData({ ...formData, nodeNamePreprocess: rules })}
                />

                {/* {t('subscriptions.form.nameProcessing.namingRule')} */}
                <Box>
                  <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 2 }}>
                    <Typography variant="subtitle1" fontWeight="bold">
                      {t('subscriptions.form.nameProcessing.namingRule')}
                    </Typography>
                    <ButtonGroup size="small" variant="outlined">
                      <Tooltip title={t('subscriptions.form.nameProcessing.builderTooltip')}>
                        <Button
                          onClick={() => setNamingMode('builder')}
                          variant={namingMode === 'builder' ? 'contained' : 'outlined'}
                          startIcon={<BuildIcon />}
                        >
                          {matchDownMd ? '' : t('subscriptions.form.nameProcessing.builder')}
                        </Button>
                      </Tooltip>
                      <Tooltip title={t('subscriptions.form.nameProcessing.manualTooltip')}>
                        <Button
                          onClick={() => setNamingMode('manual')}
                          variant={namingMode === 'manual' ? 'contained' : 'outlined'}
                          startIcon={<EditNoteIcon />}
                        >
                          {matchDownMd ? '' : t('subscriptions.form.nameProcessing.manual')}
                        </Button>
                      </Tooltip>
                    </ButtonGroup>
                  </Stack>

                  {namingMode === 'builder' ? (
                    <NodeRenameBuilder
                      value={formData.nodeNameRule}
                      onChange={(rule) => setFormData({ ...formData, nodeNameRule: rule })}
                    />
                  ) : (
                    <>
                      <TextField
                        fullWidth
                        label={t('subscriptions.form.nameProcessing.templateLabel')}
                        value={formData.nodeNameRule}
                        onChange={(e) => setFormData({ ...formData, nodeNameRule: e.target.value })}
                        placeholder={t('subscriptions.form.nameProcessing.templatePlaceholder')}
                        helperText={t('subscriptions.form.nameProcessing.templateHelper')}
                      />
                      <Box sx={helperPanelSx}>
                        <Typography variant="caption" sx={{ color: tertiaryText }} component="div">
                          <strong>{t('subscriptions.form.nameProcessing.variables')}</strong>
                          <br />• <code>$Name</code> - {t('subscriptions.form.nameProcessing.varName')} &nbsp;&nbsp; •{' '}
                          <code>$LinkName</code> - {t('subscriptions.form.nameProcessing.varLinkName')}
                          <br />• <code>$LinkCountry</code> - {t('subscriptions.form.nameProcessing.varLinkCountry')} &nbsp;&nbsp; •{' '}
                          <code>$Speed</code> - {t('subscriptions.form.nameProcessing.varSpeed')}
                          <br />• <code>$SpeedIcon</code> - {t('subscriptions.form.nameProcessing.varSpeedIcon')} &nbsp;&nbsp; •{' '}
                          <code>$Delay</code> - {t('subscriptions.form.nameProcessing.varDelay')}
                          <br />• <code>$DelayIcon</code> - {t('subscriptions.form.nameProcessing.varDelayIcon')} &nbsp;&nbsp; •{' '}
                          <code>$Group</code> - {t('subscriptions.form.nameProcessing.varGroup')}
                          <br />• <code>$Source</code> - {t('subscriptions.form.nameProcessing.varSource')} &nbsp;&nbsp; •{' '}
                          <code>$Index</code> - {t('subscriptions.form.nameProcessing.varIndex')}
                          <br />• <code>$DuplicateIndex</code> - {t('subscriptions.form.nameProcessing.varDuplicateIndex')} &nbsp;&nbsp; •{' '}
                          <code>$Protocol</code> - {t('subscriptions.form.nameProcessing.varProtocol')}
                          <br />• <code>$IpType</code> - {t('subscriptions.form.nameProcessing.varIpType')} &nbsp;&nbsp; •{' '}
                          <code>$Residential</code> -{t('subscriptions.form.nameProcessing.varResidential')}
                          <br />• <code>$FraudScore</code> - {t('subscriptions.form.nameProcessing.varFraudScore')} &nbsp;&nbsp; •{' '}
                          <code>$FraudScoreIcon</code> -{t('subscriptions.form.nameProcessing.varFraudScoreIcon')}
                          {unlockRenameVariables.length > 0 && (
                            <>
                              <br />•{' '}
                              {unlockRenameVariables.map((item, index) => (
                                <span key={item.key}>
                                  <code>{item.key}</code> - {item.label}
                                  {index < unlockRenameVariables.length - 1 ? '; ' : ''}
                                </span>
                              ))}
                            </>
                          )}
                          <br />• <code>$Tags</code> - {t('subscriptions.form.nameProcessing.varTags')} &nbsp;&nbsp; •{' '}
                          <code>$TagGroup(name)</code> - {t('subscriptions.form.nameProcessing.varTagGroup')}
                        </Typography>
                      </Box>
                      {formData.nodeNameRule && (
                        <Alert variant={'standard'} severity="info" sx={{ mt: 1 }}>
                          <Typography variant="body2">
                            <strong>{t('subscriptions.form.nameProcessing.preview')}</strong> {previewNodeName(formData.nodeNameRule)}
                          </Typography>
                        </Alert>
                      )}
                    </>
                  )}
                </Box>
              </Stack>
            </AccordionDetails>
          </Accordion>

          {/* ========== {t('subscriptions.form.sections.advanced')} ========== */}
          <Accordion expanded={expandedPanels.advanced} onChange={handlePanelChange('advanced')} sx={accordionSx}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={accordionSummarySx}>
              <SecurityIcon color="primary" />
              <Typography variant="subtitle1" fontWeight={600}>
                {t('subscriptions.form.sections.advanced')}
              </Typography>
              {!expandedPanels.advanced && advancedSettingsCount > 0 && (
                <Chip
                  size="small"
                  label={t('subscriptions.form.advanced.configured', { count: advancedSettingsCount })}
                  color="secondary"
                  variant="outlined"
                  sx={{ ml: 1 }}
                />
              )}
            </AccordionSummary>
            <AccordionDetails sx={accordionDetailsSx}>
              <Stack spacing={2.5}>
                {/* Script Selection */}
                <Autocomplete
                  multiple
                  options={scripts}
                  getOptionLabel={(option) => `${option.name} (${option.version})`}
                  value={scripts.filter((s) => formData.selectedScripts.includes(s.id))}
                  onChange={(e, newValue) => setFormData({ ...formData, selectedScripts: newValue.map((s) => s.id) })}
                  sx={autocompleteChipSx}
                  renderInput={(params) => (
                    <TextField
                      {...params}
                      label={t('subscriptions.form.advanced.scriptLabel')}
                      helperText={t('subscriptions.form.advanced.scriptHelper')}
                    />
                  )}
                  renderOption={(props, option) => (
                    <li {...props}>
                      <Box sx={{ display: 'flex', flexDirection: 'column' }}>
                        <Typography variant="body1">{option.name}</Typography>
                        <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                          {t('subscriptions.form.advanced.scriptVersion', { version: option.version })}
                        </Typography>
                      </Box>
                    </li>
                  )}
                />

                {/* IP Whitelist/Blacklist */}
                <TextField
                  fullWidth
                  label={t('subscriptions.form.advanced.ipBlacklist')}
                  multiline
                  rows={2}
                  value={formData.IPBlacklist}
                  onChange={(e) => setFormData({ ...formData, IPBlacklist: e.target.value })}
                  helperText={t('subscriptions.form.advanced.ipBlacklistHelper')}
                />
                <TextField
                  fullWidth
                  label={t('subscriptions.form.advanced.ipWhitelist')}
                  multiline
                  rows={2}
                  value={formData.IPWhitelist}
                  onChange={(e) => setFormData({ ...formData, IPWhitelist: e.target.value })}
                  helperText={t('subscriptions.form.advanced.ipBlacklistHelper')}
                />
              </Stack>
            </AccordionDetails>
          </Accordion>
        </Box>
      </DialogContent>
      <DialogActions
        sx={{
          px: matchDownMd ? 2 : 3,
          py: 1.5,
          bgcolor: mutedPanelSurface,
          borderTop: '1px solid',
          borderColor: panelBorder
        }}
      >
        <Stack direction="row" spacing={2} sx={{ width: '100%', justifyContent: 'space-between' }}>
          <Button
            variant="outlined"
            startIcon={<VisibilityIcon />}
            onClick={onPreview}
            disabled={previewLoading || (formData.selectedNodes.length === 0 && formData.selectedGroups.length === 0)}
          >
            {previewLoading ? t('subscriptions.form.actions.previewLoading') : t('subscriptions.form.actions.previewNode')}
            <Chip size="small" label="Beta" color="error" variant="outlined" sx={{ ml: 1 }} />
          </Button>
          <Stack direction="row" spacing={1}>
            <Button onClick={onClose}>{t('subscriptions.form.actions.close')}</Button>
            <Button variant="contained" onClick={onSubmit}>
              {t('subscriptions.form.actions.confirm')}
            </Button>
          </Stack>
        </Stack>
      </DialogActions>
    </Dialog>
  );
}
