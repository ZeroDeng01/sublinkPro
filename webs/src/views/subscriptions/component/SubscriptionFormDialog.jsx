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

import NodeRenameBuilder from './NodeRenameBuilder';
import NodeNamePreprocessor from './NodeNamePreprocessor';
import NodeNameFilter from './NodeNameFilter';
import NodeTagFilter from './NodeTagFilter';
import NodeProtocolFilter from './NodeProtocolFilter';
import NodeTransferBox from './NodeTransferBox';
import DeduplicationConfig from './DeduplicationConfig';
import FilterAltIcon from '@mui/icons-material/FilterAlt';

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

// 预览节点名称
const previewNodeName = (rule) => {
  if (!rule) return '';
  return rule
    .replace(/\$Name/g, '香港节点-备注')
    .replace(/\$Flag/g, '🇭🇰')
    .replace(/\$LinkName/g, '香港01')
    .replace(/\$LinkCountry/g, 'HK')
    .replace(/\$Speed/g, '1.50MB/s')
    .replace(/\$Delay/g, '125ms')
    .replace(/\$Group/g, 'Premium')
    .replace(/\$Source/g, '机场A')
    .replace(/\$Index/g, '1')
    .replace(/\$Protocol/g, 'VMess')
    .replace(/\$Tags/g, '速度优秀|香港节点')
    .replace(/\$Tag/g, '速度优秀');
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
  allNodes,
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
  const matchDownMd = useMediaQuery(theme.breakpoints.down('md'));

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

  // 按分组统计节点数量
  const groupNodeCounts = useMemo(() => {
    const counts = {};
    allNodes.forEach((node) => {
      const group = node.Group || '未分组';
      counts[group] = (counts[group] || 0) + 1;
    });
    return counts;
  }, [allNodes]);

  // 按类别筛选模板
  const clashTemplates = useMemo(() => {
    return templates.filter((t) => !t.category || t.category === 'clash');
  }, [templates]);

  const surgeTemplates = useMemo(() => {
    return templates.filter((t) => t.category === 'surge');
  }, [templates]);

  // 过滤后的节点列表
  const filteredNodes = useMemo(() => {
    return allNodes.filter((node) => {
      if (nodeGroupFilter !== 'all' && node.Group !== nodeGroupFilter) return false;
      if (nodeSourceFilter !== 'all' && node.Source !== nodeSourceFilter) return false;
      if (nodeSearchQuery) {
        const query = nodeSearchQuery.toLowerCase();
        if (!node.Name?.toLowerCase().includes(query) && !node.Group?.toLowerCase().includes(query)) {
          return false;
        }
      }
      if (nodeCountryFilter.length > 0) {
        if (!node.LinkCountry || !nodeCountryFilter.includes(node.LinkCountry)) {
          return false;
        }
      }
      return true;
    });
  }, [allNodes, nodeGroupFilter, nodeSourceFilter, nodeSearchQuery, nodeCountryFilter]);

  // 可选节点（排除已选，使用 ID 匹配）
  const availableNodes = useMemo(() => {
    return filteredNodes.filter((node) => !formData.selectedNodes.includes(node.ID));
  }, [filteredNodes, formData.selectedNodes]);

  // 已选节点（使用 ID 匹配）
  const selectedNodesList = useMemo(() => {
    return allNodes.filter((node) => formData.selectedNodes.includes(node.ID));
  }, [allNodes, formData.selectedNodes]);

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
    return count;
  }, [formData]);

  // 计算高级设置数量
  const advancedSettingsCount = useMemo(() => {
    let count = 0;
    if (formData.selectedScripts?.length > 0) count++;
    if (formData.IPWhitelist) count++;
    if (formData.IPBlacklist) count++;
    return count;
  }, [formData]);

  // 面板样式
  const accordionSx = {
    mb: 1.5,
    '&:before': { display: 'none' },
    boxShadow: theme.shadows[1],
    borderRadius: '12px !important',
    overflow: 'hidden',
    '&.Mui-expanded': {
      margin: '0 0 12px 0'
    }
  };

  const accordionSummarySx = {
    minHeight: 56,
    background: `linear-gradient(145deg, ${theme.palette.mode === 'dark' ? '#1a2027' : '#f8f9fa'} 0%, ${theme.palette.mode === 'dark' ? '#121417' : '#ffffff'} 100%)`,
    '&.Mui-expanded': {
      minHeight: 56
    },
    '& .MuiAccordionSummary-content': {
      alignItems: 'center',
      gap: 1.5,
      '&.Mui-expanded': {
        margin: '12px 0'
      }
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="lg" fullWidth>
      <DialogTitle>{isEdit ? '编辑订阅' : '添加订阅'}</DialogTitle>
      <DialogContent>
        <Box sx={{ mt: 1 }}>
          {/* ========== 基础设置 ========== */}
          <Accordion expanded={expandedPanels.basic} onChange={handlePanelChange('basic')} sx={accordionSx}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={accordionSummarySx}>
              <SettingsIcon color="primary" />
              <Typography variant="subtitle1" fontWeight={600}>
                基础设置
              </Typography>
            </AccordionSummary>
            <AccordionDetails>
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
            <AccordionDetails>
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
                  <Typography variant="caption" color="textSecondary">
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
                            <MenuItem value="all">全部分组 ({allNodes.length})</MenuItem>
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
                      allNodes={allNodes}
                      checkedAvailable={checkedAvailable}
                      checkedSelected={checkedSelected}
                      selectedNodeSearch={selectedNodeSearch}
                      onSelectedNodeSearchChange={setSelectedNodeSearch}
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
            <AccordionDetails>
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
                </Grid>

                {/* 落地IP国家过滤 */}
                <Grid container spacing={2}>
                  <Grid item xs={12} sm={6}>
                    <Autocomplete
                      multiple
                      options={countryOptions}
                      value={formData.CountryWhitelist}
                      onChange={(e, newValue) => setFormData({ ...formData, CountryWhitelist: newValue })}
                      getOptionLabel={(option) => formatCountry(option)}
                      renderInput={(params) => (
                        <TextField {...params} label="落地IP国家白名单" helperText="只保留这些国家的节点，不选则不限制" />
                      )}
                      renderOption={(props, option) => <li {...props}>{formatCountry(option)}</li>}
                    />
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <Autocomplete
                      multiple
                      options={countryOptions}
                      value={formData.CountryBlacklist}
                      onChange={(e, newValue) => setFormData({ ...formData, CountryBlacklist: newValue })}
                      getOptionLabel={(option) => formatCountry(option)}
                      renderInput={(params) => (
                        <TextField {...params} label="落地IP国家黑名单" helperText="排除这些国家的节点（优先级高于白名单）" />
                      )}
                      renderOption={(props, option) => <li {...props}>{formatCountry(option)}</li>}
                    />
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
            <AccordionDetails>
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
              {!expandedPanels.naming && (formData.nodeNamePreprocess || formData.nodeNameRule) && (
                <Chip size="small" label="已配置" color="info" variant="outlined" sx={{ ml: 1 }} />
              )}
            </AccordionSummary>
            <AccordionDetails>
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
                      <Box sx={{ mt: 1, p: 1.5, bgcolor: 'action.hover', borderRadius: 1 }}>
                        <Typography variant="caption" color="textSecondary" component="div">
                          <strong>可用变量：</strong>
                          <br />• <code>$Name</code> - 系统备注名称 &nbsp;&nbsp; • <code>$LinkName</code> - 原始节点名称
                          <br />• <code>$LinkCountry</code> - 落地IP国家代码 &nbsp;&nbsp; • <code>$Speed</code> - 下载速度
                          <br />• <code>$Delay</code> - 延迟 &nbsp;&nbsp; • <code>$Group</code> - 分组名称
                          <br />• <code>$Source</code> - 来源 &nbsp;&nbsp; • <code>$Index</code> - 序号 &nbsp;&nbsp; •{' '}
                          <code>$Protocol</code> - 协议类型
                          <br />• <code>$Tags</code> - 所有标签(逗号分隔) &nbsp;&nbsp; • <code>$Tag</code> - 第一个标签
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
            <AccordionDetails>
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
                        <Typography variant="caption" color="textSecondary">
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
      <DialogActions sx={{ borderTop: '1px solid', borderColor: 'divider' }}>
        <Stack direction="row" spacing={2} sx={{ width: '100%', justifyContent: 'space-between' }}>
          <Button
            variant="outlined"
            startIcon={<VisibilityIcon />}
            onClick={onPreview}
            disabled={previewLoading || (formData.selectedNodes.length === 0 && formData.selectedGroups.length === 0)}
          >
            {previewLoading ? '加载中...' : '预览节点'}
            <Chip size="small" label="Beta" color="error" variant="outlined" sx={{ ml: 1 }} />
          </Button>
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
