import { useState } from 'react';
import PropTypes from 'prop-types';

// material-ui
import { useTheme } from '@mui/material/styles';
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import FormControl from '@mui/material/FormControl';
import InputAdornment from '@mui/material/InputAdornment';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Alert from '@mui/material/Alert';
import IconButton from '@mui/material/IconButton';
import Accordion from '@mui/material/Accordion';
import AccordionSummary from '@mui/material/AccordionSummary';
import AccordionDetails from '@mui/material/AccordionDetails';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

// utils
import {
  createEmptyUnlockRule,
  formatUnlockProviderLabel,
  getUnlockProviderOptions,
  getUnlockRuleModeOptions,
  getUnlockStatusOptions,
  isoToFlag,
  QUALITY_STATUS_OPTIONS,
  STATUS_OPTIONS
} from '../utils';
import { getNodeActionButtonSx, getNodeColorChipSx, getNodeFieldControlSx, getNodeThemeTokens } from '../nodeTheme';

/**
 * 节点过滤器工具栏
 */
export default function NodeFilters({
  searchQuery,
  setSearchQuery,
  groupFilter,
  setGroupFilter,
  sourceFilter,
  setSourceFilter,
  maxDelay,
  setMaxDelay,
  minSpeed,
  setMinSpeed,
  maxFraudScore,
  setMaxFraudScore,
  speedStatusFilter,
  setSpeedStatusFilter,
  delayStatusFilter,
  setDelayStatusFilter,
  residentialType,
  setResidentialType,
  ipType,
  setIpType,
  qualityStatus,
  setQualityStatus,
  unlockRules,
  setUnlockRules,
  unlockRuleMode,
  setUnlockRuleMode,
  countryFilter,
  setCountryFilter,
  tagFilter,
  setTagFilter,
  protocolFilter,
  setProtocolFilter,
  groupOptions,
  sourceOptions,
  countryOptions,
  tagOptions,
  protocolOptions,
  onReset
}) {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);
  const fieldControlSx = getNodeFieldControlSx(tokens);
  const unlockProviderOptions = getUnlockProviderOptions();
  const normalizedUnlockRules = Array.isArray(unlockRules) ? unlockRules : [];
  const [unlockExpanded, setUnlockExpanded] = useState(false);

  const updateUnlockRule = (index, patch) => {
    setUnlockRules(normalizedUnlockRules.map((rule, ruleIndex) => (ruleIndex === index ? { ...rule, ...patch } : rule)));
  };

  const addUnlockRule = () => setUnlockRules([...normalizedUnlockRules, createEmptyUnlockRule()]);

  const removeUnlockRule = (index) => {
    const nextRules = normalizedUnlockRules.filter((_, ruleIndex) => ruleIndex !== index);
    setUnlockRules(nextRules);
  };

  return (
    <Stack
      direction="row"
      spacing={2}
      sx={{
        mb: 2,
        p: 2,
        borderRadius: 2.5,
        bgcolor: tokens.toolbarSurface,
        border: '1px solid',
        borderColor: tokens.softBorder,
        alignItems: 'flex-start'
      }}
      flexWrap="wrap"
      useFlexGap
    >
      <FormControl size="small" sx={{ minWidth: 120, ...fieldControlSx }}>
        <InputLabel>分组</InputLabel>
        <Select value={groupFilter} label="分组" onChange={(e) => setGroupFilter(e.target.value)} variant={'outlined'}>
          <MenuItem value="">全部</MenuItem>
          <MenuItem value="未分组">未分组</MenuItem>
          {groupOptions.map((group) => (
            <MenuItem key={group} value={group}>
              {group}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <TextField
        size="small"
        placeholder="搜索节点备注或链接"
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        sx={{ minWidth: 200, flex: { xs: '1 1 100%', md: '1 1 220px' }, ...fieldControlSx }}
      />
      <FormControl size="small" sx={{ minWidth: 120, ...fieldControlSx }}>
        <InputLabel>来源</InputLabel>
        <Select value={sourceFilter} label="来源" onChange={(e) => setSourceFilter(e.target.value)} variant={'outlined'}>
          <MenuItem value="">全部</MenuItem>
          {sourceOptions.map((source) => (
            <MenuItem key={source} value={source}>
              {source === 'manual' ? '手动添加' : source}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 120, ...fieldControlSx }}>
        <InputLabel>协议</InputLabel>
        <Select value={protocolFilter} label="协议" onChange={(e) => setProtocolFilter(e.target.value)} variant={'outlined'}>
          <MenuItem value="">全部</MenuItem>
          {protocolOptions.map((protocol) => (
            <MenuItem key={protocol} value={protocol}>
              {protocol.toUpperCase()}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 100, ...fieldControlSx }}>
        <InputLabel>延迟状态</InputLabel>
        <Select value={delayStatusFilter} label="延迟状态" onChange={(e) => setDelayStatusFilter(e.target.value)}>
          {STATUS_OPTIONS.map((opt) => (
            <MenuItem key={opt.value} value={opt.value}>
              {opt.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 100, ...fieldControlSx }}>
        <InputLabel>速度状态</InputLabel>
        <Select value={speedStatusFilter} label="速度状态" onChange={(e) => setSpeedStatusFilter(e.target.value)}>
          {STATUS_OPTIONS.map((opt) => (
            <MenuItem key={opt.value} value={opt.value}>
              {opt.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <TextField
        size="small"
        placeholder="最大延迟"
        type="number"
        value={maxDelay}
        onChange={(e) => setMaxDelay(e.target.value)}
        sx={{ width: 150, ...fieldControlSx }}
        InputProps={{ endAdornment: <InputAdornment position="end">ms</InputAdornment> }}
      />
      <TextField
        size="small"
        placeholder="最低速度"
        type="number"
        value={minSpeed}
        onChange={(e) => setMinSpeed(e.target.value)}
        sx={{ width: 150, ...fieldControlSx }}
        InputProps={{ endAdornment: <InputAdornment position="end">MB/s</InputAdornment> }}
      />
      <TextField
        size="small"
        placeholder="最大欺诈评分"
        type="number"
        value={maxFraudScore}
        onChange={(e) => setMaxFraudScore(e.target.value)}
        sx={{ width: 160, ...fieldControlSx }}
      />
      <FormControl size="small" sx={{ minWidth: 140, ...fieldControlSx }}>
        <InputLabel>质量状态</InputLabel>
        <Select value={qualityStatus} label="质量状态" onChange={(e) => setQualityStatus(e.target.value)}>
          {QUALITY_STATUS_OPTIONS.map((opt) => (
            <MenuItem key={opt.value} value={opt.value}>
              {opt.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 120, ...fieldControlSx }}>
        <InputLabel>住宅属性</InputLabel>
        <Select value={residentialType} label="住宅属性" onChange={(e) => setResidentialType(e.target.value)}>
          <MenuItem value="">全部</MenuItem>
          <MenuItem value="residential">住宅IP</MenuItem>
          <MenuItem value="datacenter">机房IP</MenuItem>
          <MenuItem value="untested">未检测</MenuItem>
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 120, ...fieldControlSx }}>
        <InputLabel>IP类型</InputLabel>
        <Select value={ipType} label="IP类型" onChange={(e) => setIpType(e.target.value)}>
          <MenuItem value="">全部</MenuItem>
          <MenuItem value="native">原生IP</MenuItem>
          <MenuItem value="broadcast">广播IP</MenuItem>
          <MenuItem value="untested">未检测</MenuItem>
        </Select>
      </FormControl>
      {countryOptions.length > 0 && (
        <Autocomplete
          multiple
          size="small"
          options={countryOptions}
          value={countryFilter}
          onChange={(e, newValue) => setCountryFilter(newValue)}
          sx={{ minWidth: 150, ...fieldControlSx }}
          getOptionLabel={(option) => `${isoToFlag(option)} ${option}`}
          renderOption={(props, option) => {
            const { key, ...otherProps } = props;
            return (
              <li key={key} {...otherProps}>
                {isoToFlag(option)} {option}
              </li>
            );
          }}
          renderTags={(value, getTagProps) =>
            value.map((option, index) => {
              const { key, ...tagProps } = getTagProps({ index });
              return <Chip key={key} label={`${isoToFlag(option)} ${option}`} size="small" {...tagProps} />;
            })
          }
          renderInput={(params) => <TextField {...params} label="国家代码" placeholder="选择国家" />}
        />
      )}
      {tagOptions && tagOptions.length > 0 && (
        <Autocomplete
          multiple
          size="small"
          options={tagOptions}
          value={tagFilter}
          onChange={(e, newValue) => setTagFilter(newValue)}
          sx={{ minWidth: 150, ...fieldControlSx }}
          getOptionLabel={(option) => option.name || option}
          isOptionEqualToValue={(option, value) => option.name === (value.name || value)}
          renderOption={(props, option) => {
            const { key, ...otherProps } = props;
            return (
              <li key={key} {...otherProps}>
                <Box
                  sx={{
                    width: 12,
                    height: 12,
                    borderRadius: '50%',
                    backgroundColor: option.color || tokens.palette.primary.main,
                    mr: 1,
                    flexShrink: 0
                  }}
                />
                {option.name}
              </li>
            );
          }}
          renderTags={(value, getTagProps) =>
            value.map((option, index) => {
              const { key, ...tagProps } = getTagProps({ index });
              return (
                <Chip
                  key={key}
                  label={option.name || option}
                  size="small"
                  sx={getNodeColorChipSx(theme, tokens, option.color || theme.palette.primary.main, { deletable: true })}
                  {...tagProps}
                />
              );
            })
          }
          renderInput={(params) => <TextField {...params} label="标签" placeholder="选择标签" sx={fieldControlSx} />}
        />
      )}
      <Box sx={{ width: '100%', minWidth: 320 }}>
        <Accordion
          expanded={unlockExpanded}
          onChange={(_, expanded) => setUnlockExpanded(expanded)}
          sx={{
            boxShadow: 'none',
            border: '1px solid',
            borderColor: tokens.softBorder,
            borderRadius: 2,
            bgcolor: tokens.nestedPanelSurface,
            '&:before': { display: 'none' }
          }}
        >
          <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={{ minHeight: 56 }}>
            <Stack spacing={0.25} sx={{ width: '100%' }}>
              <Typography variant="subtitle2">解锁筛选</Typography>
              <Typography variant="caption" color="text.secondary">
                {normalizedUnlockRules.length > 0
                  ? `${normalizedUnlockRules.length} 条规则 · ${unlockRuleMode === 'and' ? '同时满足全部' : '满足任意一条'}`
                  : '默认折叠，未添加规则时不启用解锁筛选'}
              </Typography>
            </Stack>
          </AccordionSummary>
          <AccordionDetails>
            <Stack spacing={1.5}>
              <Alert severity="info" variant="outlined">
                不添加规则时不会启用解锁筛选。你可以按需新增规则，并设置多条规则之间是满足任意一条还是同时满足全部。
              </Alert>
              <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5} alignItems={{ md: 'center' }}>
                <FormControl size="small" sx={{ minWidth: 200, ...fieldControlSx }}>
                  <InputLabel>规则关系</InputLabel>
                  <Select value={unlockRuleMode || 'or'} label="规则关系" onChange={(e) => setUnlockRuleMode(e.target.value)}>
                    {getUnlockRuleModeOptions().map((option) => (
                      <MenuItem key={option.value} value={option.value}>
                        {option.label}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
                <Typography variant="caption" color="text.secondary">
                  {unlockRuleMode === 'and' ? '多条规则需要同时满足。' : '多条规则满足任意一条即可。'}
                </Typography>
              </Stack>
              {normalizedUnlockRules.length > 0 ? (
                normalizedUnlockRules.map((rule, index) => (
                  <Stack
                    key={`node-unlock-rule-${index}`}
                    direction={{ xs: 'column', md: 'row' }}
                    spacing={1.5}
                    alignItems={{ md: 'flex-start' }}
                  >
                    <Autocomplete
                      size="small"
                      options={unlockProviderOptions}
                      value={unlockProviderOptions.find((item) => item.value === rule.provider) || null}
                      onChange={(_, newValue) => updateUnlockRule(index, { provider: newValue?.value || '' })}
                      getOptionLabel={(option) => option?.label || formatUnlockProviderLabel(option?.value || '')}
                      sx={{ minWidth: 220, flex: 1, ...fieldControlSx }}
                      renderOption={(props, option) => (
                        <li {...props} key={option.value}>
                          <Box>
                            <Typography variant="body2">{option.label}</Typography>
                            <Typography variant="caption" color="text.secondary">
                              {option.description || option.value}
                            </Typography>
                          </Box>
                        </li>
                      )}
                      renderInput={(params) => <TextField {...params} label="Provider" sx={fieldControlSx} />}
                    />
                    <FormControl size="small" sx={{ minWidth: 180, ...fieldControlSx }}>
                      <InputLabel>状态</InputLabel>
                      <Select value={rule.status || ''} label="状态" onChange={(e) => updateUnlockRule(index, { status: e.target.value })}>
                        {getUnlockStatusOptions(true).map((opt) => (
                          <MenuItem key={opt.value || 'all'} value={opt.value}>
                            {opt.label}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                    <TextField
                      size="small"
                      label="关键词"
                      value={rule.keyword || ''}
                      onChange={(e) => updateUnlockRule(index, { keyword: e.target.value })}
                      sx={{ minWidth: 220, flex: 1, ...fieldControlSx }}
                    />
                    <IconButton color="error" onClick={() => removeUnlockRule(index)} sx={{ alignSelf: { xs: 'flex-end', md: 'center' } }}>
                      <DeleteOutlineIcon fontSize="small" />
                    </IconButton>
                  </Stack>
                ))
              ) : (
                <Alert severity="info" variant="outlined">
                  当前未启用解锁筛选。点击下方按钮后再添加具体规则。
                </Alert>
              )}
              <Box>
                <Button size="small" startIcon={<AddIcon />} variant="outlined" onClick={addUnlockRule}>
                  新增解锁规则
                </Button>
              </Box>
            </Stack>
          </AccordionDetails>
        </Accordion>
      </Box>
      <Button onClick={onReset} sx={getNodeActionButtonSx(theme, tokens, tokens.palette.text.secondary)}>
        重置
      </Button>
    </Stack>
  );
}

NodeFilters.propTypes = {
  searchQuery: PropTypes.string.isRequired,
  setSearchQuery: PropTypes.func.isRequired,
  groupFilter: PropTypes.string.isRequired,
  setGroupFilter: PropTypes.func.isRequired,
  sourceFilter: PropTypes.string.isRequired,
  setSourceFilter: PropTypes.func.isRequired,
  maxDelay: PropTypes.string.isRequired,
  setMaxDelay: PropTypes.func.isRequired,
  minSpeed: PropTypes.string.isRequired,
  setMinSpeed: PropTypes.func.isRequired,
  maxFraudScore: PropTypes.string.isRequired,
  setMaxFraudScore: PropTypes.func.isRequired,
  speedStatusFilter: PropTypes.string.isRequired,
  setSpeedStatusFilter: PropTypes.func.isRequired,
  delayStatusFilter: PropTypes.string.isRequired,
  setDelayStatusFilter: PropTypes.func.isRequired,
  residentialType: PropTypes.string.isRequired,
  setResidentialType: PropTypes.func.isRequired,
  ipType: PropTypes.string.isRequired,
  setIpType: PropTypes.func.isRequired,
  qualityStatus: PropTypes.string.isRequired,
  setQualityStatus: PropTypes.func.isRequired,
  unlockRules: PropTypes.array.isRequired,
  setUnlockRules: PropTypes.func.isRequired,
  unlockRuleMode: PropTypes.string.isRequired,
  setUnlockRuleMode: PropTypes.func.isRequired,
  countryFilter: PropTypes.array.isRequired,
  setCountryFilter: PropTypes.func.isRequired,
  tagFilter: PropTypes.array.isRequired,
  setTagFilter: PropTypes.func.isRequired,
  protocolFilter: PropTypes.string.isRequired,
  setProtocolFilter: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  sourceOptions: PropTypes.array.isRequired,
  countryOptions: PropTypes.array.isRequired,
  tagOptions: PropTypes.array.isRequired,
  protocolOptions: PropTypes.array.isRequired,
  onReset: PropTypes.func.isRequired
};
