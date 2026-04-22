import { useMemo, useState } from 'react';
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
import {
  formatUnlockProviderLabel,
  getNodeUnlockSummaryDisplay,
  getUnlockProviderOptions,
  getUnlockRenameVariables,
  getUnlockRuleModeOptions,
  getUnlockStatusOptions,
  createEmptyUnlockRule
} from 'views/nodes/utils';

// ISO国家代码转换为国旗emoji
const isoToFlag = (isoCode) => {
  if (!isoCode || isoCode.length !== 2) return '';
  const code = isoCode.toUpperCase() === 'TW' ? 'CN' : isoCode.toUpperCase();
  const codePoints = code.split('').map((char) => 0x1f1e6 + char.charCodeAt(0) - 65);
  return String.fromCodePoint(...codePoints);
};

// 格式化国家显示
const formatCountry = (linkCountry) => {
  if (!linkCountry) return '';
  const flag = isoToFlag(linkCountry);
  return flag ? `${flag}${linkCountry}` : linkCountry;
};

const normalizeCountryCode = (value) => (typeof value === 'string' ? value.trim().toUpperCase() : '');

const normalizeCountryCodeList = (values) => {
  if (!Array.isArray(values)) return [];
  return Array.from(new Set(values.map((value) => normalizeCountryCode(value)).filter(Boolean)));
};

// 预览节点名称
const previewNodeName = (rule) => {
  if (!rule) return '';
  // 处理 $TagGroup(xxx) 格式
  let result = rule.replace(/\$TagGroup\([^)]+\)/g, '速度优秀');
  return result
    .replace(/\$Name/g, '香港节点-备注')
    .replace(/\$Flag/g, '🇭🇰')
    .replace(/\$SpeedIcon/g, getSpeedIcon(1.5, 'success'))
    .replace(/\$DelayIcon/g, getDelayIcon(125, 'success'))
    .replace(/\$IpType/g, '原生IP')
    .replace(/\$Residential/g, '住宅IP')
    .replace(/\$FraudScoreIcon/g, getFraudScoreIcon(12, 'success'))
    .replace(/\$FraudScore/g, '12')
    .replace(/\$Unlock\([^)]+\)/g, '解锁-US')
    .replace(/\$LinkName/g, '香港01')
    .replace(/\$LinkCountry/g, 'HK')
    .replace(/\$Speed/g, '1.50MB/s')
    .replace(/\$Delay/g, '125ms')
    .replace(/\$Group/g, 'Premium')
    .replace(/\$Source/g, '机场A')
    .replace(/\$Index/g, '1')
    .replace(/\$Protocol/g, 'VMess')
    .replace(/\$Tags/g, '速度优秀|香港节点');
};

// 检查预处理规则是否有实际配置
const hasPreprocessRules = (value) => {
  if (!value) return false;
  try {
    const parsed = JSON.parse(value);
    return Array.isArray(parsed) && parsed.length > 0;
  } catch {
    return false;
  }
};

/**
 * 订阅表单对话框
 * 使用折叠面板组织功能分组，提升用户体验
 */
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
  // 节点过滤
  nodeGroupFilter,
  setNodeGroupFilter,
  nodeSourceFilter,
  setNodeSourceFilter,
  nodeSearchQuery,
  setNodeSearchQuery,
  nodeCountryFilter,
  setNodeCountryFilter,
  // 穿梭框状态
  checkedAvailable,
  checkedSelected,
  mobileTab,
  setMobileTab,
  selectedNodeSearch,
  setSelectedNodeSearch,
  namingMode,
  setNamingMode,
  // 操作回调
  onClose,
  onSubmit,
  onPreview,
  showPreview = false,
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

  // 折叠面板展开状态（支持多个同时展开）
  const [expandedPanels, setExpandedPanels] = useState({
    basic: true,
    nodes: true,
    filter: false,
    dedup: false,
    naming: false,
    advanced: false
  });

  // 切换面板展开状态
  const handlePanelChange = (panel) => (event, isExpanded) => {
    setExpandedPanels((prev) => ({
      ...prev,
      [panel]: isExpanded
    }));
  };

  // 按类别筛选模板
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

  const normalizedSelectorNodes = useMemo(() => selectorNodes || [], [selectorNodes]);
  const selectorLoadingText = selectorNodesLoading ? '节点列表加载中...' : '';

  // 可选节点（排除已选，使用 ID 匹配）
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

  // 已选节点（使用 ID 匹配）
  // 计算过滤规则数量
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

  // 计算高级设置数量
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

  // 面板样式
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
        {isEdit ? '编辑订阅' : '添加订阅'}
      </DialogTitle>
      <DialogContent sx={{ px: matchDownMd ? 2 : 3, pt: 2.5, pb: 2, bgcolor: 'transparent' }}>
        <Box sx={{ mt: 1 }}>
          {/* ========== 基础设置 ========== */}
          <Accordion expanded={expandedPanels.basic} onChange={handlePanelChange('basic')} sx={accordionSx}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={accordionSummarySx}>
              <SettingsIcon color="primary" />
              <Typography variant="subtitle1" fontWeight={600}>
                基础设置
              </Typography>
            </AccordionSummary>
            <AccordionDetails sx={accordionDetailsSx}>
              <Stack spacing={2.5}>
                <TextField
                  fullWidth
                  label="订阅名称"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                />

                <Grid container spacing={2}>
                  <Grid item xs={12} sm={6}>
                    <FormControl fullWidth>
                      <InputLabel shrink>Clash 模板</InputLabel>
                      <Select
                        variant={'outlined'}
                        value={formData.clash}
                        label="Clash 模板"
                        onChange={(e) => setFormData({ ...formData, clash: e.target.value })}
                        displayEmpty
                      >
                        <MenuItem value="">
                          <Typography color="text.secondary">未选择</Typography>
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
                        <Typography variant="caption">未检测到可用模板，请检查 Clash 模板是否存在</Typography>
                      </Alert>
                    )}
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <FormControl fullWidth>
                      <InputLabel shrink>Surge 模板</InputLabel>
                      <Select
                        value={formData.surge}
                        label="Surge 模板"
                        onChange={(e) => setFormData({ ...formData, surge: e.target.value })}
                        displayEmpty
                      >
                        <MenuItem value="">
                          <Typography color="text.secondary">未选择</Typography>
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
                        <Typography variant="caption">未检测到可用模板，请检查 Surge 模板是否存在</Typography>
                      </Alert>
                    )}
                  </Grid>
                </Grid>

                <Stack direction="row" spacing={2} flexWrap="wrap">
                  <FormControlLabel
                    control={<Checkbox checked={formData.udp} onChange={(e) => setFormData({ ...formData, udp: e.target.checked })} />}
                    label="强制开启 UDP"
                  />
                  <FormControlLabel
                    control={<Checkbox checked={formData.cert} onChange={(e) => setFormData({ ...formData, cert: e.target.checked })} />}
                    label="跳过证书验证"
                  />
                  <Tooltip title="根据系统 Host 配置，将节点服务器地址替换为对应的 IP 地址" placement="top" arrow>
                    <FormControlLabel
                      control={
                        <Checkbox
                          checked={formData.replaceServerWithHost}
                          onChange={(e) => setFormData({ ...formData, replaceServerWithHost: e.target.checked })}
                        />
                      }
                      label="替换服务器地址为 Host"
                    />
                  </Tooltip>
                  <Tooltip
                    title="开启后每次访问订阅链接会实时获取最新用量信息（流量、到期时间等），但会增加响应时间；关闭后使用缓存数据，响应更快"
                    placement="top"
                    arrow
                  >
                    <FormControlLabel
                      control={
                        <Checkbox
                          checked={formData.refreshUsageOnRequest}
                          onChange={(e) => setFormData({ ...formData, refreshUsageOnRequest: e.target.checked })}
                        />
                      }
                      label="实时获取用量信息"
                    />
                  </Tooltip>
                </Stack>
              </Stack>
            </AccordionDetails>
          </Accordion>

          {/* ========== 节点选择 ========== */}
          <Accordion expanded={expandedPanels.nodes} onChange={handlePanelChange('nodes')} sx={accordionSx}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={accordionSummarySx}>
              <AccountTreeIcon color="primary" />
              <Typography variant="subtitle1" fontWeight={600}>
                节点选择
              </Typography>
              {!expandedPanels.nodes && (formData.selectedNodes.length > 0 || formData.selectedGroups.length > 0) && (
                <Chip
                  size="small"
                  label={`${formData.selectedNodes.length} 节点 / ${formData.selectedGroups.length} 分组`}
                  color="primary"
                  variant="outlined"
                  sx={{ ml: 1 }}
                />
              )}
            </AccordionSummary>
            <AccordionDetails sx={accordionDetailsSx}>
              <Stack spacing={2.5}>
                {/* 选择模式 */}
                <Box>
                  <RadioGroup
                    row
                    value={formData.selectionMode}
                    onChange={(e) => setFormData({ ...formData, selectionMode: e.target.value })}
                  >
                    <FormControlLabel value="nodes" control={<Radio />} label="手动选择节点" />
                    <FormControlLabel value="groups" control={<Radio />} label="动态选择分组" />
                    <FormControlLabel value="mixed" control={<Radio />} label="混合模式" />
                  </RadioGroup>
                  <Typography variant="caption" sx={helperCaptionSx}>
                    {formData.selectionMode === 'nodes' && '手动选择具体节点，节点不会随分组变化自动更新'}
                    {formData.selectionMode === 'groups' && '选择分组，自动包含该分组下的所有节点，节点会随分组变化自动更新'}
                    {formData.selectionMode === 'mixed' && '同时支持手动选择节点和动态选择分组'}
                  </Typography>
                </Box>

                {/* 分组选择 */}
                {(formData.selectionMode === 'groups' || formData.selectionMode === 'mixed') && (
                  <Autocomplete
                    multiple
                    options={groupOptions}
                    value={formData.selectedGroups}
                    onChange={(e, newValue) => setFormData({ ...formData, selectedGroups: newValue })}
                    renderInput={(params) => <TextField {...params} label="选择分组（动态）" />}
                    renderOption={(props, option) => (
                      <li {...props}>
                        {option} ({groupNodeCounts[option] || 0} 个节点)
                      </li>
                    )}
                  />
                )}

                {/* 节点选择 */}
                {(formData.selectionMode === 'nodes' || formData.selectionMode === 'mixed') && (
                  <>
                    <Grid container spacing={2}>
                      <Grid item xs={6} sm={3}>
                        <FormControl fullWidth size="small">
                          <InputLabel>分组过滤</InputLabel>
                          <Select value={nodeGroupFilter} label="分组过滤" onChange={(e) => setNodeGroupFilter(e.target.value)}>
                            <MenuItem value="all">全部分组 ({allNodeTotal})</MenuItem>
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
                          <InputLabel>来源过滤</InputLabel>
                          <Select value={nodeSourceFilter} label="来源过滤" onChange={(e) => setNodeSourceFilter(e.target.value)}>
                            <MenuItem value="all">全部来源</MenuItem>
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
                          renderInput={(params) => <TextField {...params} label="国家过滤" />}
                          renderOption={(props, option) => <li {...props}>{formatCountry(option)}</li>}
                          limitTags={2}
                        />
                      </Grid>
                      <Grid item xs={6} sm={3}>
                        <TextField
                          fullWidth
                          size="small"
                          label="搜索节点"
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

          {/* ========== 节点过滤 ========== */}
          <Accordion expanded={expandedPanels.filter} onChange={handlePanelChange('filter')} sx={accordionSx}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={accordionSummarySx}>
              <FilterListIcon color="primary" />
              <Typography variant="subtitle1" fontWeight={600}>
                节点过滤
              </Typography>
              {!expandedPanels.filter && filterRulesCount > 0 && (
                <Chip size="small" label={`已启用 ${filterRulesCount} 项规则`} color="warning" variant="outlined" sx={{ ml: 1 }} />
              )}
            </AccordionSummary>
            <AccordionDetails sx={accordionDetailsSx}>
              <Stack spacing={2.5}>
                {/* 延迟和速度过滤 */}
                <Grid container spacing={2}>
                  <Grid item xs={12} sm={6}>
                    <TextField
                      fullWidth
                      label="最大延迟"
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
                      helperText="设置筛选节点的延迟阈值，0表示不限制"
                    />
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <TextField
                      fullWidth
                      label="最小速度"
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
                      helperText="设置筛选节点的最小下载速度，0表示不限制"
                    />
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <TextField
                      fullWidth
                      label="最大欺诈评分"
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
                      helperText="0表示不限制；需要先执行 IP 质量检测"
                    />
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <FormControl fullWidth>
                      <InputLabel>质量状态</InputLabel>
                      <Select
                        value={formData.QualityStatus || ''}
                        label="质量状态"
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
                      可区分完整结果、信息不全、检测失败、未启用和未检测
                    </Typography>
                  </Grid>
                  <Grid item xs={12}>
                    <Stack spacing={1.5}>
                      <Typography variant="subtitle2">解锁筛选规则</Typography>
                      <Alert severity="info" variant="outlined">
                        不添加规则时不会启用解锁筛选。你可以按需新增规则，并设置多条规则之间是满足任意一条还是同时满足全部。
                      </Alert>
                      <Grid container spacing={1.5} alignItems="center">
                        <Grid item xs={12} md={4}>
                          <FormControl fullWidth size="small">
                            <InputLabel>规则关系</InputLabel>
                            <Select
                              value={formData.UnlockRuleMode || 'or'}
                              label="规则关系"
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
                              ? '多条规则需要同时满足，适合做更严格的筛选。'
                              : '多条规则满足任意一条即可，适合组合多个候选解锁条件。'}
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
                                  <TextField {...params} label="Provider" helperText="例如 Gemini / YouTube Premium" />
                                )}
                              />
                            </Grid>
                            <Grid item xs={12} md={3}>
                              <FormControl fullWidth>
                                <InputLabel>状态</InputLabel>
                                <Select
                                  value={rule.status || ''}
                                  label="状态"
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
                                label="关键词"
                                value={rule.keyword || ''}
                                onChange={(e) => updateUnlockRule(index, { keyword: e.target.value })}
                                helperText="如 US、可用、Claude、地区受限"
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
                                删除
                              </Button>
                            </Grid>
                          </Grid>
                        ))
                      ) : (
                        <Alert severity="info" variant="outlined">
                          当前未启用解锁筛选。点击下方按钮后再添加具体规则。
                        </Alert>
                      )}
                      <Box>
                        <Button startIcon={<AddIcon />} variant="outlined" onClick={addUnlockRule}>
                          新增一条解锁规则
                        </Button>
                      </Box>
                    </Stack>
                  </Grid>
                </Grid>

                {/* 落地IP国家过滤 */}
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
                          label="落地IP国家白名单"
                          helperText="可从建议中选择，或输入后按 Enter 添加；建议使用 HK/US/JP 这类两位国家代码"
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
                          label="落地IP国家黑名单"
                          helperText="可从建议中选择，或输入后按 Enter 添加；建议使用 HK/US/JP 这类两位国家代码"
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
                      <InputLabel>住宅属性</InputLabel>
                      <Select
                        value={formData.ResidentialType || ''}
                        label="住宅属性"
                        onChange={(e) => setFormData({ ...formData, ResidentialType: e.target.value })}
                      >
                        <MenuItem value="">全部</MenuItem>
                        <MenuItem value="residential">住宅IP</MenuItem>
                        <MenuItem value="datacenter">机房IP</MenuItem>
                        <MenuItem value="untested">未检测</MenuItem>
                      </Select>
                    </FormControl>
                    <Typography variant="caption" sx={helperCaptionSx}>
                      仅完整结果才会显示住宅/机房属性；信息不全会保留为独立状态
                    </Typography>
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <FormControl fullWidth>
                      <InputLabel>IP类型</InputLabel>
                      <Select
                        value={formData.IPType || ''}
                        label="IP类型"
                        onChange={(e) => setFormData({ ...formData, IPType: e.target.value })}
                      >
                        <MenuItem value="">全部</MenuItem>
                        <MenuItem value="native">原生IP</MenuItem>
                        <MenuItem value="broadcast">广播IP</MenuItem>
                        <MenuItem value="untested">未检测</MenuItem>
                      </Select>
                    </FormControl>
                    <Typography variant="caption" sx={helperCaptionSx}>
                      仅完整结果才会显示原生/广播属性；信息不全不会被误判为原生或广播
                    </Typography>
                  </Grid>
                </Grid>

                {/* 节点标签过滤 */}
                <NodeTagFilter
                  tagOptions={tagOptions}
                  whitelistValue={formData.tagWhitelist}
                  blacklistValue={formData.tagBlacklist}
                  onWhitelistChange={(tags) => setFormData({ ...formData, tagWhitelist: tags })}
                  onBlacklistChange={(tags) => setFormData({ ...formData, tagBlacklist: tags })}
                />

                {/* 协议类型过滤 */}
                <NodeProtocolFilter
                  protocolOptions={formData.protocolOptions || []}
                  whitelistValue={formData.protocolWhitelist}
                  blacklistValue={formData.protocolBlacklist}
                  onWhitelistChange={(protocols) => setFormData({ ...formData, protocolWhitelist: protocols })}
                  onBlacklistChange={(protocols) => setFormData({ ...formData, protocolBlacklist: protocols })}
                />

                {/* 节点名称过滤 */}
                <NodeNameFilter
                  whitelistValue={formData.nodeNameWhitelist}
                  blacklistValue={formData.nodeNameBlacklist}
                  onWhitelistChange={(rules) => setFormData({ ...formData, nodeNameWhitelist: rules })}
                  onBlacklistChange={(rules) => setFormData({ ...formData, nodeNameBlacklist: rules })}
                />
              </Stack>
            </AccordionDetails>
          </Accordion>

          {/* ========== 节点去重 ========== */}
          <Accordion expanded={expandedPanels.dedup} onChange={handlePanelChange('dedup')} sx={accordionSx}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={accordionSummarySx}>
              <FilterAltIcon color="primary" />
              <Typography variant="subtitle1" fontWeight={600}>
                节点去重
                <Chip size="small" label="Beta" color="error" variant="outlined" sx={{ ml: 1 }} />
              </Typography>
              {!expandedPanels.dedup && formData.deduplicationRule && (
                <Chip size="small" label="已配置" color="success" variant="outlined" sx={{ ml: 1 }} />
              )}
            </AccordionSummary>
            <AccordionDetails sx={accordionDetailsSx}>
              <DeduplicationConfig
                value={formData.deduplicationRule || ''}
                onChange={(rule) => setFormData({ ...formData, deduplicationRule: rule })}
              />
            </AccordionDetails>
          </Accordion>

          {/* ========== 名称处理 ========== */}
          <Accordion expanded={expandedPanels.naming} onChange={handlePanelChange('naming')} sx={accordionSx}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={accordionSummarySx}>
              <TextFieldsIcon color="primary" />
              <Typography variant="subtitle1" fontWeight={600}>
                节点名称处理
              </Typography>
              {!expandedPanels.naming && (hasPreprocessRules(formData.nodeNamePreprocess) || formData.nodeNameRule) && (
                <Chip size="small" label="已配置" color="info" variant="outlined" sx={{ ml: 1 }} />
              )}
            </AccordionSummary>
            <AccordionDetails sx={accordionDetailsSx}>
              <Stack spacing={2.5}>
                {/* 原名预处理 */}
                <NodeNamePreprocessor
                  value={formData.nodeNamePreprocess}
                  onChange={(rules) => setFormData({ ...formData, nodeNamePreprocess: rules })}
                />

                {/* 节点命名规则 */}
                <Box>
                  <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 2 }}>
                    <Typography variant="subtitle1" fontWeight="bold">
                      节点命名规则
                    </Typography>
                    <ButtonGroup size="small" variant="outlined">
                      <Tooltip title="可视化构建器 - 拖拽添加变量">
                        <Button
                          onClick={() => setNamingMode('builder')}
                          variant={namingMode === 'builder' ? 'contained' : 'outlined'}
                          startIcon={<BuildIcon />}
                        >
                          {matchDownMd ? '' : '构建器'}
                        </Button>
                      </Tooltip>
                      <Tooltip title="手动输入模式">
                        <Button
                          onClick={() => setNamingMode('manual')}
                          variant={namingMode === 'manual' ? 'contained' : 'outlined'}
                          startIcon={<EditNoteIcon />}
                        >
                          {matchDownMd ? '' : '手动'}
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
                        label="命名规则模板"
                        value={formData.nodeNameRule}
                        onChange={(e) => setFormData({ ...formData, nodeNameRule: e.target.value })}
                        placeholder="例如: [$Protocol]$LinkCountry-$Name"
                        helperText="留空则使用原始名称，仅在访问订阅链接时生效"
                      />
                      <Box sx={helperPanelSx}>
                        <Typography variant="caption" sx={{ color: tertiaryText }} component="div">
                          <strong>可用变量：</strong>
                          <br />• <code>$Name</code> - 系统备注名称 &nbsp;&nbsp; • <code>$LinkName</code> - 原始节点名称
                          <br />• <code>$LinkCountry</code> - 落地IP国家代码 &nbsp;&nbsp; • <code>$Speed</code> - 下载速度
                          <br />• <code>$SpeedIcon</code> - 速度图标(🟢🟡🔴/❌⏱️⛔️) &nbsp;&nbsp; • <code>$Delay</code> - 延迟
                          <br />• <code>$DelayIcon</code> - 延迟图标(🟢🟡🔴/❌⏱️⛔️) &nbsp;&nbsp; • <code>$Group</code> - 分组名称
                          <br />• <code>$Source</code> - 来源 &nbsp;&nbsp; • <code>$Index</code> - 序号 &nbsp;&nbsp; •{' '}
                          <code>$Protocol</code> - 协议类型
                          <br />• <code>$IpType</code> - IP类型(原生IP/广播IP) &nbsp;&nbsp; • <code>$Residential</code> -
                          住宅属性(住宅IP/机房IP)
                          <br />• <code>$FraudScore</code> - 欺诈评分 &nbsp;&nbsp; • <code>$FraudScoreIcon</code> -
                          欺诈图标(⚪🟢🟡🟠🔴⚫/⛔️)
                          {unlockRenameVariables.length > 0 && (
                            <>
                              <br />•{' '}
                              {unlockRenameVariables.map((item, index) => (
                                <span key={item.key}>
                                  <code>{item.key}</code> - {item.label}
                                  {index < unlockRenameVariables.length - 1 ? '； ' : ''}
                                </span>
                              ))}
                            </>
                          )}
                          <br />• <code>$Tags</code> - 所有标签(竖线分隔) &nbsp;&nbsp; • <code>$TagGroup(组名)</code> - 指定标签组中的标签
                        </Typography>
                      </Box>
                      {formData.nodeNameRule && (
                        <Alert variant={'standard'} severity="info" sx={{ mt: 1 }}>
                          <Typography variant="body2">
                            <strong>预览：</strong> {previewNodeName(formData.nodeNameRule)}
                          </Typography>
                        </Alert>
                      )}
                    </>
                  )}
                </Box>
              </Stack>
            </AccordionDetails>
          </Accordion>

          {/* ========== 高级设置 ========== */}
          <Accordion expanded={expandedPanels.advanced} onChange={handlePanelChange('advanced')} sx={accordionSx}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={accordionSummarySx}>
              <SecurityIcon color="primary" />
              <Typography variant="subtitle1" fontWeight={600}>
                高级设置
              </Typography>
              {!expandedPanels.advanced && advancedSettingsCount > 0 && (
                <Chip size="small" label={`已配置 ${advancedSettingsCount} 项`} color="secondary" variant="outlined" sx={{ ml: 1 }} />
              )}
            </AccordionSummary>
            <AccordionDetails sx={accordionDetailsSx}>
              <Stack spacing={2.5}>
                {/* 脚本选择 */}
                <Autocomplete
                  multiple
                  options={scripts}
                  getOptionLabel={(option) => `${option.name} (${option.version})`}
                  value={scripts.filter((s) => formData.selectedScripts.includes(s.id))}
                  onChange={(e, newValue) => setFormData({ ...formData, selectedScripts: newValue.map((s) => s.id) })}
                  renderInput={(params) => (
                    <TextField {...params} label="数据处理脚本" helperText="脚本将在查询到节点数据后运行，多个脚本按顺序执行" />
                  )}
                  renderOption={(props, option) => (
                    <li {...props}>
                      <Box sx={{ display: 'flex', flexDirection: 'column' }}>
                        <Typography variant="body1">{option.name}</Typography>
                        <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                          版本: {option.version}
                        </Typography>
                      </Box>
                    </li>
                  )}
                />

                {/* IP 白名单/黑名单 */}
                <TextField
                  fullWidth
                  label="IP 黑名单（优先级高于白名单），不允许指定IP访问订阅链接"
                  multiline
                  rows={2}
                  value={formData.IPBlacklist}
                  onChange={(e) => setFormData({ ...formData, IPBlacklist: e.target.value })}
                  helperText="每行一个 IP 或 CIDR"
                />
                <TextField
                  fullWidth
                  label="IP 白名单，只允许指定IP访问订阅链接"
                  multiline
                  rows={2}
                  value={formData.IPWhitelist}
                  onChange={(e) => setFormData({ ...formData, IPWhitelist: e.target.value })}
                  helperText="每行一个 IP 或 CIDR"
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
        <Stack direction="row" spacing={2} sx={{ width: '100%', justifyContent: showPreview ? 'space-between' : 'flex-end' }}>
          {showPreview && (
            <Button
              variant="outlined"
              startIcon={<VisibilityIcon />}
              onClick={onPreview}
              disabled={previewLoading || (formData.selectedNodes.length === 0 && formData.selectedGroups.length === 0)}
            >
              {previewLoading ? '加载中...' : '预览节点'}
              <Chip size="small" label="Beta" color="error" variant="outlined" sx={{ ml: 1 }} />
            </Button>
          )}
          <Stack direction="row" spacing={1}>
            <Button onClick={onClose}>关闭</Button>
            <Button variant="contained" onClick={onSubmit}>
              确定
            </Button>
          </Stack>
        </Stack>
      </DialogActions>
    </Dialog>
  );
}
